package main

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// --- 1. DLL DEFINITIONS ---

// เปลี่ยนชื่อ DLL ตรงนี้ได้ตามใจชอบ!
var (
	modWintun = syscall.NewLazyDLL("wintun.dll")

	procCreateAdapter        = modWintun.NewProc("WintunCreateAdapter")
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

// Helper ในการเรียก DLL
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

	// ใช้ Mutex ป้องกันการเขียนพร้อมกัน
	writeMu sync.Mutex
}

// CreateAdapter: สร้าง Adapter
func CreateNativeWintun(name string, tunnelType string) (*NativeWintun, error) {
	name16, _ := windows.UTF16PtrFromString(name)
	type16, _ := windows.UTF16PtrFromString(tunnelType)

	// GUID เป็น nil ได้ Wintun จะ gen ให้
	r1, _, err := procCreateAdapter.Call(
		uintptr(unsafe.Pointer(name16)),
		uintptr(unsafe.Pointer(type16)),
		0,
	)
	if r1 == 0 {
		return nil, fmt.Errorf("create adapter failed: %v", err)
	}

	wt := &NativeWintun{AdapterHandle: r1, AdapterName: name}

	// เริ่ม Session ทันที (Capacity แนะนำคือ 0x400000 = 4MB Ring Buffer)
	rSession, _, err := procStartSession.Call(r1, 0x400000)
	if rSession == 0 {
		wt.Close()
		return nil, fmt.Errorf("start session failed: %v", err)
	}
	wt.SessionHandle = rSession

	// ดึง Event Handle สำหรับรออ่านข้อมูล (หัวใจสำคัญของ Wintun)
	rEvent, _, _ := procGetReadWaitEvent.Call(rSession)
	wt.ReadWaitEvent = windows.Handle(rEvent)

	return wt, nil
}

func (w *NativeWintun) Close() {
	if w.SessionHandle != 0 {
		procEndSession.Call(w.SessionHandle)
	}
	if w.AdapterHandle != 0 {
		procDeleteAdapter.Call(w.AdapterHandle)
	}
}

func (w *NativeWintun) GetLUID() (uint64, error) {
	var luid uint64

	// เรียก WintunGetAdapterLUID(AdapterHandle, *LUID)
	// ฟังก์ชันนี้ใน C เป็น void (ไม่คืนค่า) แต่จะเขียนค่าลงใน pointer
	_, _, err := procGetAdapterLUID.Call(
		w.AdapterHandle,
		uintptr(unsafe.Pointer(&luid)),
	)

	// เนื่องจากเป็น System Call บางที err จะมีค่า "The operation completed successfully" ติดมาด้วย
	// เราจึงเช็คที่ค่า luid ว่าได้มาจริงไหมแทน
	if luid == 0 {
		return 0, fmt.Errorf("failed to get LUID (result is 0): %v", err)
	}

	return luid, nil
}

// --- 3. READ LOGIC ---

func (w *NativeWintun) ReadPacket() ([]byte, error) {
	for {
		var packetSize uint32

		// 1. ลองดึง Packet จาก Ring Buffer
		// WintunReceivePacket จะคืน Pointer ไปยังข้อมูลใน Memory ของ Wintun โดยตรง
		ptr, _, _ := procReceivePacket.Call(
			w.SessionHandle,
			uintptr(unsafe.Pointer(&packetSize)),
		)

		if ptr != 0 {
			// เจอข้อมูล!
			// แปลง Pointer เป็น Go Slice
			// (ใช้เทคนิค unsafe slice เพื่อความเร็ว ไม่ต้อง copy ถ้าแค่อ่าน)
			data := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), packetSize)

			// สำคัญ: เราต้อง Copy ออกมาเก็บไว้ เพราะเดี๋ยวต้อง Release คืนให้ Wintun
			result := make([]byte, packetSize)
			copy(result, data)

			// 2. แจ้ง Wintun ว่าอ่านเสร็จแล้ว (Release พื้นที่ใน Ring Buffer)
			procReleaseReceivePacket.Call(w.SessionHandle, ptr)

			return result, nil
		}

		// 3. ถ้าไม่มีข้อมูล (ptr == 0) ให้รอสัญญาณ (Sleep)
		// ไม่กิน CPU 100% เพราะรอที่ระดับ OS Kernel
		event, err := windows.WaitForSingleObject(w.ReadWaitEvent, windows.INFINITE)
		if event == windows.WAIT_FAILED {
			return nil, fmt.Errorf("wait failed: %v", err)
		}
	}
}

// --- 4. WRITE LOGIC ---

func (w *NativeWintun) WritePacket(data []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	size := uint32(len(data))

	// 1. ขอจองพื้นที่ใน Ring Buffer (Allocate)
	ptr, _, _ := procAllocateSendPacket.Call(w.SessionHandle, uintptr(size))

	if ptr == 0 {
		return fmt.Errorf("ring buffer full")
	}

	// 2. Copy ข้อมูลจาก Go ไปใส่ในพื้นที่ที่จองไว้ (C Memory)
	dst := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
	copy(dst, data)

	// 3. สั่งส่งข้อมูล (Send)
	procSendPacket.Call(w.SessionHandle, ptr)

	return nil
}
