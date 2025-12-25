package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"vpn-server/internal/firewall"

	"log"
	"math/big"
	"net"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"sync/atomic"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const publicKeyPem = `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAE4mbxQlXwEXVko5uXa/op8tgCcsGg8InQTjQcX6Lj1U=
-----END PUBLIC KEY-----
`

// --- CONFIG ---
const (
	BindAddr         = "0.0.0.0:6500"
	VPNCIDR          = "10.8.0.0/23" // รองรับ 65k IPs
	VPNInterface     = "tun0"
	AllowedGroup     = "admin.group@onzt.tridorian.com"
	AdminImpersonate = "admin@onzt.tridorian.com"
)

var publicKey crypto.PrivateKey

// --- 1. FAST IPAM (Channel Based - O(1)) ---
type FastIPPool struct {
	CIDR         string
	AvailableIPs chan string
	Used         sync.Map
}

func NewFastIPPool(cidr string) *FastIPPool {

	pool := &FastIPPool{AvailableIPs: make(chan string, 65500), CIDR: cidr}
	go func() {
		ip, ipnet, _ := net.ParseCIDR(cidr)
		prefix, _ := ipnet.Mask.Size()
		ip = ip.To4()

		// Generate IPs 10.8.0.2 - 10.8.255.254
		for i := range 255 {
			for j := 2; j < 254; j++ {
				ip[2], ip[3] = byte(i), byte(j)
				pool.AvailableIPs <- ip.String() + "/" + fmt.Sprint(prefix)
			}
		}
	}()
	return pool
}

func (p *FastIPPool) GetHostIPAddress() string {
	ip, ipnet, _ := net.ParseCIDR(p.CIDR)
	prefix, _ := ipnet.Mask.Size()
	ip = ip.To4()
	ip[3] = byte(1)

	return ip.String() + "/" + fmt.Sprint(prefix)
}

// Assign an IP to an email
func (p *FastIPPool) Assign(email string) (string, error) {
	select {
	case ip := <-p.AvailableIPs:
		p.Used.Store(ip, email)
		return ip, nil
	default:
		return "", errors.New("IP Pool Empty")
	}
}

func (p *FastIPPool) Release(ip string) {
	p.Used.Delete(ip)
	select {
	case p.AvailableIPs <- ip:
	default:
	}
}

// --- 2. ACL ENGINE (Fast Path) ---

// --- 3. GOOGLE GROUP CHECK ---
func CheckGroup(email string) bool {
	ctx := context.Background()
	jsonKey, _ := os.ReadFile("service-account.json")
	config, _ := google.JWTConfigFromJSON(jsonKey, admin.AdminDirectoryGroupReadonlyScope)
	config.Subject = AdminImpersonate
	srv, err := admin.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return true
	} // Fail open for dev
	r, err := srv.Members.HasMember(AllowedGroup, email).Do()
	return err == nil && r.IsMember
}

// --- MAIN SERVER ---
func main() {

	var err error
	publicKey, err = jwt.ParseEdPublicKeyFromPEM([]byte(publicKeyPem))
	if err != nil {
		panic(err)
	}

	// Tuning
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	// Setup System
	exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()

	ipPool := NewFastIPPool(VPNCIDR)

	// Create Multiqueue TUN
	ifces := make([]*water.Interface, numCPU)
	for i := range numCPU {
		config := water.Config{
			DeviceType: water.TUN,
			PlatformSpecificParams: water.PlatformSpecificParams{
				MultiQueue: true,
				Name:       VPNInterface,
				Persist:    false,
			},
		}
		ifces[i], _ = water.New(config)
	}

	// Setup IP & NAT
	exec.Command("ip", "addr", "add", ipPool.GetHostIPAddress(), "dev", VPNInterface).Run()
	exec.Command("ip", "link", "set", "dev", VPNInterface, "up", "mtu", "1420").Run()
	exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", VPNCIDR, "-j", "MASQUERADE").Run()

	// Mock config
	firewall.CurrentConfig = &firewall.AgentExecutionConfig{
		ConditionalAccessDefaultPolicies: firewall.ConditionalAccessDefaultPolicies{
			BlockByDefault: true,
		},
		ConditionalAccessPolicies: []firewall.ConditionalAccessPolicy{
			{
				Name:                  "Allow1",
				Description:           "Allow access 1",
				Priority:              50,
				SourceTagType:         "Identity",
				SourceMatchValue:      "chatdanai.phakaket@tridorian.com",
				DestinationTagType:    "CIDR",
				DestinationMatchValue: "216.239.34.21/32,216.239.32.21/32,216.239.36.21/32,216.239.38.21/32",
				Action:                "ALLOW",
			},
			{
				Name:                  "Allow2",
				Description:           "Allow access 2",
				Priority:              50,
				SourceTagType:         "Identity",
				SourceMatchValue:      "chatdanai.phakaket@tridorian.com",
				DestinationTagType:    "SNI",
				DestinationMatchValue: "tridorian.com",
				Action:                "ALLOW",
			},
			// {
			// 	Name:                   "Allow 172.17.0.0/24",
			// 	Description:            "Allow access to 172.17.0.0/24",
			// 	Priority:               100,
			// 	NetworkTagType:         "SourceCIDR",
			// 	MatchValue:             "0.0.0.0/0",
			// 	DestinationNetworkCIDR: "172.17.0.0/24",
			// 	Action:                 "ALLOW",
			// },
		},
	}

	log.Printf("VPN Server Started on %s with %d Queues", BindAddr, numCPU)

	// init firewall engine
	firewall.NewEngine()

	clientConns := sync.Map{}

	// Start Listeners
	tlsConf := generateTLS()
	listener, _ := quic.ListenAddr(BindAddr, tlsConf, &quic.Config{
		EnableDatagrams: true, MaxIdleTimeout: 30 * time.Second,
	})

	// Start TUN Readers (Parallel)
	for i := range numCPU {
		go func(ifce *water.Interface) {
			buf := make([]byte, 1500)
			for {
				n, err := ifce.Read(buf)
				if err != nil {
					return
				}
				packet := buf[:n]
				if len(packet) >= 20 {
					dstIP := net.IP(packet[16:20]).String()
					if connVal, ok := clientConns.Load(dstIP); ok {
						conn := connVal.(*quic.Conn)
						conn.SendDatagram(packet)
					}
				}
			}
		}(ifces[i])
	}

	// Accept Loop
	var rr uint64
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			continue
		}

		// Round-Robin Select Interface for Writing
		idx := int(atomic.AddUint64(&rr, 1) % uint64(numCPU))
		go handleClient(conn, ifces[idx], ipPool, &clientConns)
	}
}

func handleClient(conn *quic.Conn, ifce *water.Interface, pool *FastIPPool, clients *sync.Map) {
	// Handshake
	stream, _ := conn.AcceptStream(context.Background())
	buf := make([]byte, 2048)
	n, _ := stream.Read(buf)

	// Auth

	token, _ := jwt.Parse(string(buf[:n]), func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil

	})
	if token == nil || !token.Valid {
		conn.CloseWithError(1, "Auth Fail")
		return
	}
	email := token.Claims.(jwt.MapClaims)["email"].(string)

	// if !CheckGroup(email) {
	// 	conn.CloseWithError(1, "Group Fail")
	// 	return
	// }

	// Assign IP
	myIP, err := pool.Assign(email)
	if err != nil {
		conn.CloseWithError(1, "IP Full")
		return
	}

	stream.Write([]byte("OK:" + myIP))
	stream.Close()

	ipOnly, _, _ := net.ParseCIDR(myIP)
	clients.Store(ipOnly.String(), conn)
	defer func() { pool.Release(myIP); clients.Delete(ipOnly.String()) }()

	// Data Loop
	for {
		packetData, err := conn.ReceiveDatagram(context.Background())
		if err != nil {
			return
		}

		version := packetData[0] >> 4

		var srcIP, dstIP netip.Addr
		var parseErr error

		if version == 4 {
			srcIP, dstIP, parseErr = parseIPv4Header(packetData)
		} else if version == 6 {
			continue
		}

		if parseErr != nil {
			continue
		}

		if !firewall.Engine.IsAllowed(packetData, firewall.ValType{
			Addr:     srcIP,
			Identity: email,
		}, firewall.ValType{
			Addr: dstIP,
		}) {

			log.Printf("Blocked packet from %s to %s", srcIP, dstIP)
			continue
		}

		// ---------------------------------------------------------
		// 4. เขียนลง TUN Interface (ถ้าผ่าน)
		// ---------------------------------------------------------
		_, err = ifce.Write(packetData)
		if err != nil {
			log.Printf("TUN Write Error: %v", err)
		}
	}
}

func generateTLS() *tls.Config {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{SerialNumber: big.NewInt(1), NotAfter: time.Now().Add(87600 * time.Hour)}
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	cert, _ := tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	return &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"vpn-quic"}}
}

func parseIPv4Header(b []byte) (src, dst netip.Addr, err error) {
	if len(b) < 20 {
		return netip.Addr{}, netip.Addr{}, fmt.Errorf("packet too short")
	}

	// IPv4: Source IP อยู่ byte ที่ 12-15
	//       Dest IP   อยู่ byte ที่ 16-19
	// ใช้ netip.AddrFrom4 เพื่อความเร็ว (Zero alloc)

	srcAddr := [4]byte{b[12], b[13], b[14], b[15]}
	dstAddr := [4]byte{b[16], b[17], b[18], b[19]}

	return netip.AddrFrom4(srcAddr), netip.AddrFrom4(dstAddr), nil
}
