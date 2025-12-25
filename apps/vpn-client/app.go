package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	vpnCancel context.CancelFunc
}

func NewApp() *App                         { return &App{} }
func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// 1. LOGIN PROCESS (เหมือนเดิม)
func (a *App) StartLoginProcess() string {
	tokenChan := make(chan string)
	srv := &http.Server{Addr: ":9876"}
	go func() {
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			tokenChan <- r.URL.Query().Get("token")
			fmt.Fprintf(w, "Done")
		})
		srv.ListenAndServe()
	}()
	// Note: เปลี่ยน localhost:8080 เป็น IP ของ Auth Server จริงเมื่อ Deploy
	wailsRuntime.BrowserOpenURL(a.ctx, "http://localhost:8080/")
	token := <-tokenChan
	srv.Shutdown(context.Background())
	return token
}

// 2. VPN CONNECT LOGIC (อัปเกรดแล้ว)
func (a *App) ConnectVPN(token string, serverAddr string) string {
	if serverAddr == "" {
		serverAddr = "localhost:6500"
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.vpnCancel = cancel

	go func() {
		defer cancel()

		for {

			tlsConf := &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"vpn-quic"}}

			// 1. เชื่อมต่อ QUIC
			conn, err := quic.DialAddr(ctx, serverAddr, tlsConf, &quic.Config{
				EnableDatagrams: true,
				KeepAlivePeriod: 10 * time.Second, // *สำคัญ* ส่ง Ping ทุก 10 วิ กันหลุด
				MaxIdleTimeout:  30 * time.Second,
			})
			if err != nil {
				emitError(a.ctx, "Connection failed: "+err.Error())
				return
			}

			// 2. Handshake & Auth
			stream, err := conn.OpenStreamSync(ctx)
			if err != nil {
				emitError(a.ctx, "Handshake Error")
				return
			}
			stream.Write([]byte(token))
			buf := make([]byte, 1024)
			n, _ := stream.Read(buf)
			resp := string(buf[:n])
			stream.Close()

			if !strings.HasPrefix(resp, "OK:") {
				emitError(a.ctx, "Auth Failed: "+resp)
				return
			}

			fmt.Println(resp)
			myIPCIDR := strings.TrimPrefix(resp, "OK:") // e.g., 10.8.0.5/24
			wailsRuntime.EventsEmit(a.ctx, "vpn-status", "Authorized. Setting up network...")

			// ==========================================
			// ส่วนแยก Logic: Windows (Wintun) vs Others (Water)
			// ==========================================

			if runtime.GOOS == "windows" {
				// --- WINDOWS (Wintun Logic) ---

				myTun, err := CreateNativeWintun("TriVPN Adapter", "Tridorian Tunnel")
				if err != nil {
					emitError(a.ctx, "Native Wintun Failed: "+err.Error())
					return
				}
				defer myTun.Close()
				// 2. ดึง LUID (NativeWintun ของเรา)
				// *หมายเหตุ: ต้อง Cast เป็น LUID type ที่เราเพิ่งประกาศในไฟล์ใหม่
				rawLUID, _ := myTun.GetLUID()
				luid := LUID(rawLUID)

				// 3. Set IP (ใช้ฟังก์ชัน Native ของเรา!)
				// "10.8.0.5/24"
				err = SetAdapterIP(luid, myIPCIDR)
				if err != nil {
					emitError(a.ctx, "Set IP Error: "+err.Error())
					// return
				}

				wailsRuntime.EventsEmit(a.ctx, "vpn-status", "Connected (Wintun): "+myIPCIDR)

				// Start Loop Read
				go func() {
					for {
						packet, err := myTun.ReadPacket() // เรียกฟังก์ชันที่เราเขียนเอง
						if err != nil {
							continue
						}
						conn.SendDatagram(packet)
					}
				}()

				// Start Loop Write
				for {
					msg, err := conn.ReceiveDatagram(ctx)
					if err != nil {
						continue
					}
					myTun.WritePacket(msg) // เรียกฟังก์ชันที่เราเขียนเอง
				}
			} else {
				// --- macOS / Linux (Water Logic) ---

				config := water.Config{DeviceType: water.TUN}
				ifce, err := water.New(config)
				if err != nil {
					emitError(a.ctx, "TUN Failed: "+err.Error())
					return
				}

				// ตั้งค่า IP (เรียกใช้ฟังก์ชันเดิมที่เราเคยเขียน)
				if err := setupPlatformNetwork(ifce.Name(), myIPCIDR); err != nil {
					emitError(a.ctx, "IP Config Failed: "+err.Error())
					return
				}

				wailsRuntime.EventsEmit(a.ctx, "vpn-status", "Connected (Legacy): "+myIPCIDR)

				// Data Loop สำหรับ Water Interface (Read/Write ไม่มี offset)
				go func() {
					b := make([]byte, 1500)
					for {
						n, err := ifce.Read(b)
						if err != nil {
							break
						}
						conn.SendDatagram(b[:n])
					}
				}()

				for {
					msg, err := conn.ReceiveDatagram(ctx)
					if err != nil {
						break
					}
					ifce.Write(msg)
				}
			}

			wailsRuntime.EventsEmit(a.ctx, "vpn-status", "Disconnected")

			select {
			case <-time.After(3 * time.Second):
				continue // วนกลับไปบรรทัดแรก
			case <-ctx.Done():
				return // User กด Stop ระหว่างรอ
			}
		}

	}()

	return "Connecting..."
}

func (a *App) DisconnectVPN() {
	if a.vpnCancel != nil {
		a.vpnCancel()
	}
}

func emitError(ctx context.Context, msg string) {
	log.Println(msg)
	wailsRuntime.EventsEmit(ctx, "vpn-error", msg)
}

// --- CROSS-PLATFORM NETWORK SETUP ---

func setupPlatformNetwork(ifName string, cidr string) error {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	ipStr := ip.String()

	switch runtime.GOOS {
	case "linux":
		// Linux: ip addr add 10.8.0.5/24 dev tun0
		if err := exec.Command("ip", "addr", "add", cidr, "dev", ifName).Run(); err != nil {
			return fmt.Errorf("linux ip add: %v", err)
		}
		if err := exec.Command("ip", "link", "set", "dev", ifName, "up").Run(); err != nil {
			return fmt.Errorf("linux ip up: %v", err)
		}

	case "darwin": // macOS
		// macOS: ifconfig utunX 10.8.0.5 10.8.0.5 up
		// macOS ต้องการ Point-to-Point address (ใส่ IP ตัวเองซ้ำ 2 รอบได้)
		if err := exec.Command("ifconfig", ifName, ipStr, ipStr, "up").Run(); err != nil {
			return fmt.Errorf("macos ifconfig: %v", err)
		}
		// เพิ่ม Route ให้ traffic วิ่งเข้า VPN (ตัวอย่าง: ให้ 10.8.0.0/24 วิ่งเข้า)
		// targetNetwork := "10.8.0.0/24"
		// exec.Command("route", "-n", "add", "-net", targetNetwork, ipStr).Run()

	case "windows":
		// Windows: netsh interface ip set address "InterfaceName" static IP Mask
		// ต้องแปลง CIDR mask (/24) เป็น Dotted Decimal (255.255.255.0)
		mask := net.IP(ipNet.Mask)
		maskStr := fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])

		log.Printf("Configuring Windows IP: %s Mask: %s on %s", ipStr, maskStr, ifName)

		cmd := exec.Command("netsh", "interface", "ip", "set", "address",
			ifName, "static", ipStr, maskStr)

		// ซ่อนหน้าต่าง command prompt ที่เด้งขึ้นมา
		// cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("windows netsh: %v", err)
		}
	}
	return nil
}
