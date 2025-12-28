package main

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"sync"

	"github.com/pkg/browser"
	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/chacha20poly1305"
)

//go:embed callback.html
var callbackHTML string

// App struct
type App struct {
	ctx         context.Context
	authToken   string
	userEmail   string
	isConnected bool

	// Session Lifecycle
	lifecycleLock sync.Mutex
	sessionCancel context.CancelFunc
	sessionDone   chan struct{}

	currentRoutes []string
	diffLock      sync.Mutex
	sessionAEAD   cipher.AEAD

	winTun  *NativeWintun
	unixTun *water.Interface
}

type HandshakeResponse struct {
	AssignedIP string   `json:"assigned_ip"`
	GW_IP      string   `json:"gw_ip"`
	Routes     []string `json:"routes"`
	SessionKey string   `json:"session_key"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Setup File Logging
	f, err := os.OpenFile("ztna_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(f)
	}

	switch runtime.GOOS {
	case "windows":
		a.winTun, err = CreateNativeWintun("TriVPN Adapter", "Tridorian Tunnel")
		if err != nil {
			log.Printf("Failed to create adapter: %v", err)
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Native Wintun Failed: "+err.Error())
			return
		}
		// When this function exits, Close() is called, which deletes the adapter.
		// This happens inside the defer stack, effectively blocking 'close(newDone)' until finished.
		defer a.winTun.Close()
	case "linux", "darwin":
		config := water.Config{
			DeviceType: water.TUN,
		}

		a.unixTun, err = water.New(config)
		if err != nil {
			log.Printf("Failed to create TUN: %v", err)
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "TUN Error: "+err.Error())
			return
		}

	}
	log.Println("App Started")
}

// Greet -> Login: Trigger Login Flow
func (a *App) Greet(domain string) string {
	// 1. Start Local Listener
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "Error starting local listener: " + err.Error()
	}

	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	// 2. Open Browser to Auth API
	target := domain
	if target == "" {
		target = "localhost:8081"
	}
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}
	authURL := fmt.Sprintf("%s:8081/?desktop_port=%d&os=%s", target, port, runtime.GOOS)

	log.Printf("Opening browser: %s", authURL)
	go browser.OpenURL(authURL)

	// 3. Wait for Callback
	errChan := make(chan error)
	successChan := make(chan string)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		email := r.URL.Query().Get("email")

		if token != "" {
			a.authToken = token
			a.userEmail = email
			// Serve embedded callback HTML page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(callbackHTML))
			successChan <- fmt.Sprintf("Logged in as %s", email)
		} else {
			errChan <- fmt.Errorf("login failed: no token received")
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case msg := <-successChan:
		server.Shutdown(context.Background())
		// Don't auto-connect, wait for user to select gateway
		return msg
	case err := <-errChan:
		return "Error: " + err.Error()
	case <-time.After(2 * time.Minute):
		server.Shutdown(context.Background())
		return "Login timed out"
	}
}

// GetGateways fetches the list of gateways from the Auth API
func (a *App) GetGateways(domain string) []map[string]interface{} {
	if a.authToken == "" {
		return nil
	}

	target := domain
	if target == "" {
		target = "localhost:8081"
	}
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}

	url := fmt.Sprintf("%s:8081/gateways", target)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+a.authToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch gateways: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch gateways: Status %d", resp.StatusCode)
		return nil
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Failed to decode gateways: %v", err)
		return nil
	}

	return result.Data
}

// Connect initiates the VPN connection to a specific gateway
func (a *App) Connect(gatewayAddress string) {
	a.lifecycleLock.Lock()

	// 1. Cancel previous session
	if a.sessionCancel != nil {
		a.sessionCancel()
	} else if a.isConnected {
		a.Disconnect() // Fallback
	}

	// 2. Prepare for new session with synchronization
	prevDone := a.sessionDone // Capture the done channel of the *previous* session
	newDone := make(chan struct{})
	a.sessionDone = newDone

	sessionCtx, cancel := context.WithCancel(context.Background())
	a.sessionCancel = cancel
	a.isConnected = true

	a.lifecycleLock.Unlock()

	go func() {
		// 3. Wait for the previous session to completely exit/cleanup
		if prevDone != nil {
			log.Println("Waiting for previous session cleanup...")
			<-prevDone
			log.Println("Previous session cleanup complete.")
		}

		var connectionFailed bool

		defer func() {
			if r := recover(); r != nil {
				log.Printf("CRITICAL PANIC in Connect: %v", r)
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", fmt.Sprintf("Crash Detect: %v", r))
			}
			// Context might have been cancelled already, but ensure cleanup
			cancel()
			close(newDone) // Signal that THIS session is done cleaning up

			a.lifecycleLock.Lock()
			if a.sessionDone == newDone {
				// Only clear isConnected if we are still the "current" session
				a.isConnected = false
			}
			a.lifecycleLock.Unlock()

			if !connectionFailed {
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Disconnected")
			}
		}()

		if a.authToken == "" {
			connectionFailed = true
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Authentication required first")
			return
		}

		log.Printf("Initiating connection to %s...", gatewayAddress)
		wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Connecting...")

		// 1. Dial Gateway via QUIC
		tlsConf := &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"vpn-quic"},
		}

		dialCtx, dialCancel := context.WithTimeout(sessionCtx, 10*time.Second)
		defer dialCancel()

		conn, err := quic.DialAddr(dialCtx, fmt.Sprintf("%s:6500", gatewayAddress), tlsConf, &quic.Config{
			EnableDatagrams: true,
		})
		if err != nil {
			if sessionCtx.Err() != nil {
				log.Println("Connection cancelled by user during dial")
				return
			}
			log.Printf("QUIC Dial Error: %v", err)
			connectionFailed = true
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Connection Failed: "+err.Error())
			return
		}

		// Ensure connection is closed when we exit this scope
		defer func() {
			conn.CloseWithError(0, "Disconnecting")
		}()

		// 2. Auth Handshake
		stream, err := conn.OpenStreamSync(sessionCtx)
		if err != nil {
			connectionFailed = true
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Stream Error: "+err.Error())
			return
		}
		_, err = stream.Write([]byte(a.authToken))
		if err != nil {
			connectionFailed = true
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Auth Send Error: "+err.Error())
			return
		}
		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			connectionFailed = true
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Auth Read Error: "+err.Error())
			return
		}
		if n == 0 {
			connectionFailed = true
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Auth Read Error: Empty Response")
			return
		}
		resp := string(buf[:n])

		log.Printf("Gateway Response: %s", resp)
		wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Authenticated")

		var response HandshakeResponse
		if err := json.Unmarshal([]byte(resp), &response); err != nil {
			log.Printf("JSON Error: %v", err)
			connectionFailed = true
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "JSON Unmarshal Error: "+err.Error())
			return
		}

		// Initialize encryption cipher from session key
		if response.SessionKey != "" {
			sessionKeyBytes, err := hex.DecodeString(response.SessionKey)
			if err != nil {
				log.Printf("Session Key Decode Error: %v", err)
				connectionFailed = true
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Session Key Error: "+err.Error())
				return
			}

			aead, err := chacha20poly1305.NewX(sessionKeyBytes)
			if err != nil {
				log.Printf("AEAD Creation Error: %v", err)
				connectionFailed = true
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Cipher Error: "+err.Error())
				return
			}

			a.sessionAEAD = aead
			log.Println("✅ Encryption enabled")
		}

		fmt.Println(response)

		log.Println("Creating Wintun Adapter...")
		switch runtime.GOOS {
		case "windows":

			rawLUID, _ := a.winTun.GetLUID()
			luid := LUID(rawLUID)

			a.winTun, err = OpenNativeWintun("TriVPN Adapter")
			if err != nil {
				log.Printf("OpenNativeWintun Error: %v", err)
				connectionFailed = true
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Open Native Wintun Error: "+err.Error())
				return
			}

			log.Printf("Setting IP: %s", response.AssignedIP)
			err = SetAdapterIP(luid, response.AssignedIP)
			if err != nil {
				log.Printf("SetAdapterIP Error: %v", err)
				connectionFailed = true
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Set IP Error: "+err.Error())
				return
			}

			log.Printf("Adding Routes: %v", response.Routes)
			a.diffLock.Lock()
			a.currentRoutes = response.Routes
			a.diffLock.Unlock()

			for _, route := range response.Routes {
				err = AddWinsRoute(luid, route)
				if err != nil {
					log.Printf("AddRoute Error: %v", err)
					connectionFailed = true
					wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Add Route Error: "+err.Error())
					return
				}
			}

			// Check for cancellation before starting loop
			if sessionCtx.Err() != nil {
				log.Println("Session cancelled before starting loop")
				return
			}

			log.Println("Starting VPN Loop...")
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Connected: "+resp)

			// Start Control Stream Handler (Route Updates)
			a.handleControlStreams(sessionCtx, conn, func(newRoutes []string) {
				log.Printf("📥 Received Route Update: %v", newRoutes)

				a.diffLock.Lock()
				defer a.diffLock.Unlock()

				current := make(map[string]bool)
				for _, r := range a.currentRoutes {
					current[r] = true
				}

				target := make(map[string]bool)
				for _, r := range newRoutes {
					target[r] = true
				}

				// toAdd = target - current
				var toAdd []string
				for r := range target {
					if !current[r] {
						toAdd = append(toAdd, r)
					}
				}

				// toRemove = current - target
				var toRemove []string
				for r := range current {
					if !target[r] {
						toRemove = append(toRemove, r)
					}
				}

				log.Printf("Diff: +%v, -%v", toAdd, toRemove)

				// Apply Removals
				for _, route := range toRemove {
					if err := RemoveWinsRoute(luid, route); err != nil {
						log.Printf("Failed to remove route %s: %v", route, err)
					}
				}

				// Apply Additions
				for _, route := range toAdd {
					if err := AddWinsRoute(luid, route); err != nil {
						log.Printf("Failed to add route %s: %v", route, err)
					}
				}

				a.currentRoutes = newRoutes
				fmt.Println(toRemove, toAdd, newRoutes)
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", fmt.Sprintf("Routes Sync: +%d, -%d", len(toAdd), len(toRemove)))
			})

			a.vpnLoopWinTun(sessionCtx, conn, a.winTun)

		case "linux":

			var response HandshakeResponse
			if err := json.Unmarshal([]byte(resp), &response); err != nil {
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", "JSON Error: "+err.Error())
				return
			}

			// Configure Interface
			cmd := exec.Command("ip", "addr", "add", response.AssignedIP, "dev", a.unixTun.Name())
			if output, err := cmd.CombinedOutput(); err != nil {
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", fmt.Sprintf("IP Addr Error: %v, %s", err, string(output)))
				return
			}

			cmd = exec.Command("ip", "link", "set", "dev", a.unixTun.Name(), "up", "mtu", "1420")
			if output, err := cmd.CombinedOutput(); err != nil {
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", fmt.Sprintf("Link Up Error: %v, %s", err, string(output)))
				return
			}

			// Add Routes
			a.diffLock.Lock()
			a.currentRoutes = response.Routes
			a.diffLock.Unlock()

			for _, route := range response.Routes {
				cmd := exec.Command("ip", "route", "add", route, "dev", a.unixTun.Name())
				if output, err := cmd.CombinedOutput(); err != nil {
					log.Printf("Route Error for %s: %v %s", route, err, string(output))
				}
			}

			// Check for cancellation before starting loop
			if sessionCtx.Err() != nil {
				log.Println("Session cancelled before starting loop")
				return
			}

			log.Println("Starting Linux VPN Loop...")
			wailsRuntime.EventsEmit(a.ctx, "vpn_status", "Connected: "+resp)

			// Start Control Stream Handler
			a.handleControlStreams(sessionCtx, conn, func(newRoutes []string) {
				// (Same update logic as above)
				log.Printf("📥 Received Route Update: %v", newRoutes)
				a.diffLock.Lock()
				defer a.diffLock.Unlock()
				current := make(map[string]bool)
				for _, r := range a.currentRoutes {
					current[r] = true
				}
				target := make(map[string]bool)
				for _, r := range newRoutes {
					target[r] = true
				}
				var toAdd, toRemove []string
				for r := range target {
					if !current[r] {
						toAdd = append(toAdd, r)
					}
				}
				for r := range current {
					if !target[r] {
						toRemove = append(toRemove, r)
					}
				}
				// Apply Linux Removals
				for _, route := range toRemove {
					cmd := exec.Command("ip", "route", "del", route, "dev", a.unixTun.Name())
					if output, err := cmd.CombinedOutput(); err != nil {
						log.Printf("Failed to remove route %s: %v %s", route, err, string(output))
					}
				}
				// Apply Linux Additions
				for _, route := range toAdd {
					cmd := exec.Command("ip", "route", "add", route, "dev", a.unixTun.Name())
					if output, err := cmd.CombinedOutput(); err != nil {
						log.Printf("Failed to add route %s: %v %s", route, err, string(output))
					}
				}
				a.currentRoutes = newRoutes
				wailsRuntime.EventsEmit(a.ctx, "vpn_status", fmt.Sprintf("Routes Sync: +%d, -%d", len(toAdd), len(toRemove)))
			})

			a.vpnLoopWater(sessionCtx, conn, a.unixTun)
		}

	}()
}

func (a *App) vpnLoopWinTun(ctx context.Context, conn *quic.Conn, tun *NativeWintun) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in vpnLoopWinTun: %v", r)
		}
		log.Println("Exiting vpnLoopWinTun")
	}()

	// Monitor disconnect signal
	go func() {
		<-ctx.Done()
		conn.CloseWithError(0, "User Disconnect")
	}()

	// Read from TUN -> Send to QUIC
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC in Packet Forwarder: %v", r)
			}
		}()
		for {
			// Check context frequently or rely on tun closing?
			if ctx.Err() != nil {
				return
			}

			packet, err := tun.ReadPacket()
			if err != nil {
				// Stop loop if reading fails (e.g. adapter closed)
				return
			}

			// Encrypt packet before sending
			var dataToSend []byte
			if a.sessionAEAD != nil {
				nonce := make([]byte, a.sessionAEAD.NonceSize())
				rand.Read(nonce)
				dataToSend = a.sessionAEAD.Seal(nonce, nonce, packet, nil)
			} else {
				dataToSend = packet
			}

			err = conn.SendDatagram(dataToSend)
			if err != nil {
				// Connection likely closed
				return
			}
		}
	}()

	// Read from QUIC -> Write to TUN
	// This will block until the connection is closed (by the monitor goroutine) or fails
	for {
		encryptedMsg, err := conn.ReceiveDatagram(ctx)
		if err != nil {
			log.Printf("QUIC Receive Error (likely disconnect): %v", err)
			return
		}

		// Decrypt packet
		var packet []byte
		if a.sessionAEAD != nil {
			nonceSize := a.sessionAEAD.NonceSize()
			if len(encryptedMsg) < nonceSize {
				continue
			}
			nonce := encryptedMsg[:nonceSize]
			ciphertext := encryptedMsg[nonceSize:]

			packet, err = a.sessionAEAD.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				log.Printf("Decryption Error: %v", err)
				continue
			}
		} else {
			packet = encryptedMsg
		}

		tun.WritePacket(packet)
	}
}

func (a *App) vpnLoopWater(ctx context.Context, conn *quic.Conn, tun *water.Interface) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in vpnLoopWater: %v", r)
		}
	}()

	// Monitor disconnect signal
	go func() {
		<-ctx.Done()
		conn.CloseWithError(0, "User Disconnect")
	}()

	// Read from TUN -> Send to QUIC
	go func() {
		buf := make([]byte, 2048)
		for {
			if ctx.Err() != nil {
				return
			}
			n, err := tun.Read(buf)
			if err != nil {
				log.Printf("TUN Read Error: %v", err)
				return
			}
			packet := make([]byte, n)
			copy(packet, buf[:n])

			// Encrypt packet before sending
			var dataToSend []byte
			if a.sessionAEAD != nil {
				nonce := make([]byte, a.sessionAEAD.NonceSize())
				rand.Read(nonce)
				dataToSend = a.sessionAEAD.Seal(nonce, nonce, packet, nil)
			} else {
				dataToSend = packet
			}

			err = conn.SendDatagram(dataToSend)
			if err != nil {
				return
			}
		}
	}()

	// Read from QUIC -> Write to TUN
	for {
		encryptedMsg, err := conn.ReceiveDatagram(ctx)
		if err != nil {
			log.Printf("QUIC Read Error: %v", err)
			return // Connection likely closed
		}

		// Decrypt packet
		var packet []byte
		if a.sessionAEAD != nil {
			nonceSize := a.sessionAEAD.NonceSize()
			if len(encryptedMsg) < nonceSize {
				continue
			}
			nonce := encryptedMsg[:nonceSize]
			ciphertext := encryptedMsg[nonceSize:]

			packet, err = a.sessionAEAD.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				log.Printf("Decryption Error: %v", err)
				continue
			}
		} else {
			packet = encryptedMsg
		}

		_, err = tun.Write(packet)
		if err != nil {
			log.Printf("TUN Write Error: %v", err)
		}
	}
}

func (a *App) Disconnect() string {
	fmt.Println("Disconnect requested...")
	a.lifecycleLock.Lock()
	if a.isConnected || a.sessionCancel != nil {
		log.Println("Disconnect requested...")
		if a.sessionCancel != nil {
			a.sessionCancel()
		}
		a.isConnected = false
	}
	a.lifecycleLock.Unlock()

	switch runtime.GOOS {
	case "windows":
		a.winTun.Close()
	case "linux":
		a.unixTun.Close()
	}

	log.Println("Disconnecting...")
	return "Disconnected"
}

// SignOut disconnects the VPN and clears user credentials
func (a *App) SignOut() string {
	a.Disconnect()

	// Reset State for clean login
	a.authToken = ""
	a.userEmail = ""
	a.currentRoutes = nil

	return "Signed Out"
}

func (a *App) handleControlStreams(ctx context.Context, conn *quic.Conn, updateRoutes func([]string)) {
	go func() {
		log.Println("Starting Control Stream Handler")
		defer log.Println("Exiting Control Stream Handler")
		for {
			if ctx.Err() != nil {
				return
			}
			stream, err := conn.AcceptUniStream(ctx)
			if err != nil {
				log.Printf("Control Stream Accept Error: %v", err)
				return
			}
			go func() {
				var msg struct {
					Type   string   `json:"type"`
					Routes []string `json:"routes"`
				}
				if err := json.NewDecoder(stream).Decode(&msg); err != nil {
					log.Printf("Control Msg Decode Error: %v", err)
					return
				}
				if msg.Type == "route_update" {
					updateRoutes(msg.Routes)
				}
			}()
		}
	}()
}
