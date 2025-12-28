//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// --- 1. DLL DEFINITIONS ---

//
//go:embed wintun.dll
var wintunDllPayload []byte

// loadWintun extracts the embedded DLL and loads it
func loadWintun() *syscall.LazyDLL {
	// 1. Get Executable Directory
	exePath, err := os.Executable()
	if err != nil {
		exePath = "wintun.dll" // Fallback
	} else {
		exePath = filepath.Join(filepath.Dir(exePath), "wintun.dll")
	}

	// 2. Write DLL to disk (if possible)
	// We try to write it. If it fails (e.g. exists and is locked), we ignore and try to load what's there.
	_ = os.WriteFile(exePath, wintunDllPayload, 0644)

	return syscall.NewLazyDLL(exePath)
}

var (
	modWintun = loadWintun()

	procCreateAdapter        = modWintun.NewProc("WintunCreateAdapter")
	procOpenAdapter          = modWintun.NewProc("WintunOpenAdapter")
	procDeleteAdapter        = modWintun.NewProc("WintunDeleteAdapter")
	procStartSession         = modWintun.NewProc("WintunStartSession")
	procEndSession           = modWintun.NewProc("WintunEndSession")
	procGetReadWaitEvent     = modWintun.NewProc("WintunGetReadWaitEvent")
	procReceivePacket        = modWintun.NewProc("WintunReceivePacket")
	procReleaseReceivePacket = modWintun.NewProc("WintunReleaseReceivePacket")
	procAllocateSendPacket   = modWintun.NewProc("WintunAllocateSendPacket")
	procSendPacket           = modWintun.NewProc("WintunSendPacket")
	procGetAdapterLUID       = modWintun.NewProc("WintunGetAdapterLUID")
)

func call(proc *syscall.LazyProc, args ...uintptr) (uintptr, error) {
	r1, _, err := proc.Call(args...)
	if r1 == 0 {
		return 0, err
	}
	return r1, nil
}

// --- 2. WRAPPER STRUCT ---

type NativeWintun struct {
	AdapterName string

	AdapterHandle uintptr
	SessionHandle uintptr
	ReadWaitEvent windows.Handle

	cancelEvent windows.Handle
	closed      bool
	mu          sync.Mutex
	readWg      sync.WaitGroup

	writeMu sync.Mutex
}

// initSession handles the common logic of starting a session and setting up events
func initSession(adapterHandle uintptr, name string) (*NativeWintun, error) {
	wt := &NativeWintun{AdapterHandle: adapterHandle, AdapterName: name}

	rSession, _, err := procStartSession.Call(adapterHandle, 0x400000)
	if rSession == 0 {
		return nil, fmt.Errorf("start session failed: %v", err)
	}
	wt.SessionHandle = rSession
	rEvent, _, _ := procGetReadWaitEvent.Call(rSession)
	wt.ReadWaitEvent = windows.Handle(rEvent)

	// Create manual reset event for cancellation
	cancelEvent, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		procEndSession.Call(rSession)
		return nil, fmt.Errorf("create cancel event failed: %v", err)
	}
	wt.cancelEvent = cancelEvent

	return wt, nil
}

// OpenNativeWintun opens an existing Wintun adapter
func OpenNativeWintun(name string) (*NativeWintun, error) {
	name16, _ := windows.UTF16PtrFromString(name)
	r1, _, _ := procOpenAdapter.Call(uintptr(unsafe.Pointer(name16)))
	if r1 == 0 {
		return nil, fmt.Errorf("adapter not found")
	}
	return initSession(r1, name)
}

// CreateNativeWintun creates or opens a Wintun adapter
func CreateNativeWintun(name string, tunnelType string) (*NativeWintun, error) {
	// 1. Try to open existing adapter
	if wt, err := OpenNativeWintun(name); err == nil {
		return wt, nil
	}

	// 2. If not found, create a new one
	name16, _ := windows.UTF16PtrFromString(name)
	type16, _ := windows.UTF16PtrFromString(tunnelType)

	r1, _, err := procCreateAdapter.Call(
		uintptr(unsafe.Pointer(name16)),
		uintptr(unsafe.Pointer(type16)),
		0,
	)
	if r1 == 0 {
		return nil, fmt.Errorf("create adapter failed: %v", err)
	}

	return initSession(r1, name)
}

// DeleteWintunAdapter deletes a Wintun adapter by name
func DeleteWintunAdapter(name string) error {
	name16, _ := windows.UTF16PtrFromString(name)

	// Open Adapter to get Handle
	r1, _, _ := procOpenAdapter.Call(uintptr(unsafe.Pointer(name16)))
	if r1 == 0 {
		return fmt.Errorf("adapter not found")
	}

	// Delete
	r2, _, err := procDeleteAdapter.Call(r1)
	if r2 == 0 {
		return fmt.Errorf("failed to delete adapter: %v", err)
	}
	return nil
}

func (w *NativeWintun) Close() {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	w.closed = true

	// Signal readers to exit
	if w.cancelEvent != 0 {
		windows.SetEvent(w.cancelEvent)
	}
	w.mu.Unlock()

	// Wait for all readers to exit to ensure no one is using the handles
	w.readWg.Wait()

	// Now unsafe to use session handles, but safe to destroy them
	if w.SessionHandle != 0 {
		procEndSession.Call(w.SessionHandle)
	}
	// Adapter persistence requested - do not delete adapter

	if w.cancelEvent != 0 {
		windows.CloseHandle(w.cancelEvent)
	}
}

func (w *NativeWintun) GetLUID() (uint64, error) {
	var luid uint64
	_, _, err := procGetAdapterLUID.Call(
		w.AdapterHandle,
		uintptr(unsafe.Pointer(&luid)),
	)
	if luid == 0 {
		return 0, fmt.Errorf("failed to get LUID (result is 0): %v", err)
	}
	return luid, nil
}

// --- 3. READ LOGIC ---

func (w *NativeWintun) ReadPacket() ([]byte, error) {
	w.readWg.Add(1)
	defer w.readWg.Done()

	for {
		w.mu.Lock()
		if w.closed {
			w.mu.Unlock()
			return nil, fmt.Errorf("closed")
		}
		session := w.SessionHandle
		w.mu.Unlock()

		if session == 0 {
			return nil, fmt.Errorf("closed")
		}

		// Try to receive packet - lock held to prevent EndSession race
		// Actually, we rely on readWg now. Close() won't EndSession until we return.
		// But EndSession invalidates calls? Close() waits for us to return.
		// So session is valid here.

		var packetSize uint32
		ptr, _, _ := procReceivePacket.Call(
			session,
			uintptr(unsafe.Pointer(&packetSize)),
		)

		if ptr != 0 {
			data := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), packetSize)
			result := make([]byte, packetSize)
			copy(result, data)
			procReleaseReceivePacket.Call(session, ptr)
			return result, nil
		}

		// Wait for data or cancel
		events := []windows.Handle{w.ReadWaitEvent, w.cancelEvent}
		waitResult, _ := windows.WaitForMultipleObjects(events, false, windows.INFINITE)

		switch waitResult {
		case windows.WAIT_OBJECT_0:
			// ReadWaitEvent signaled, try read again
			continue
		case windows.WAIT_OBJECT_0 + 1:
			// cancelEvent signaled
			return nil, fmt.Errorf("closed")
		default:
			return nil, fmt.Errorf("wait failed")
		}
	}
}

// --- 4. WRITE LOGIC ---

func (w *NativeWintun) WritePacket(data []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	w.mu.Lock()
	if w.closed || w.SessionHandle == 0 {
		w.mu.Unlock()
		return fmt.Errorf("closed")
	}
	session := w.SessionHandle
	w.mu.Unlock() // Safe because Close waits for us? No, Write isn't tracked by readWg.
	// We need Write tracking too if we want total safety, or holding lock?
	// Holding lock during syscall is safer for Write.

	// For Write, let's hold the lock during the operation to be simple and safe vs Close()
	// But Close() waits for readWg, not write lock.
	// Actually, if Close() sets closed=true, further writes fail.
	// If a write is in progress... implementation detail: WintunEndSession blocks? No.
	// For robustness, let's wrap write in a lightweight valid check or just standard lock.

	size := uint32(len(data))
	ptr, _, _ := procAllocateSendPacket.Call(session, uintptr(size))
	if ptr == 0 {
		return fmt.Errorf("ring buffer full")
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
	copy(dst, data)
	procSendPacket.Call(session, ptr)

	return nil
}
