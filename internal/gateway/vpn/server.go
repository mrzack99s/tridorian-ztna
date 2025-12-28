package vpn

import (
	"context"
	"crypto"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/netip"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"tridorian-ztna/internal/gateway/firewall"

	"github.com/golang-jwt/jwt/v5"
	quic "github.com/quic-go/quic-go"
	"github.com/songgao/water"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/time/rate"
)

const (
	VPNInterface = "tun0"
)

type IPManager interface {
	AssignIP(ctx context.Context, userID, email string) (string, error)
	ReleaseIP(ctx context.Context, ip, email string)
	SyncSessions(ctx context.Context, sessions []SessionInfo)
}

type SessionInfo struct {
	UserID      string
	Email       string
	IPAddress   string
	ConnectedAt int64
}

type Server struct {
	Addr          string
	PublicKey     crypto.PublicKey
	IPManager     IPManager
	ClientConns   sync.Map
	Ifces         []*water.Interface
	Config        *Config
	GlobalLimiter *rate.Limiter
	mu            sync.RWMutex
}

type Config struct {
	VPNCIDR      string
	PublicKeyPEM string
}

type ClientSession struct {
	Conn        *quic.Conn
	UserID      string
	Email       string
	Groups      []string
	OS          string
	SessionKey  []byte
	AEAD        cipher.AEAD
	ConnectedAt int64
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

// BroadcastRouteUpdates iterates over all connected clients and sends them the updated routes
func (s *Server) BroadcastRouteUpdates() {
	log.Println("📢 Broadcasting route updates to connected clients...")
	s.ClientConns.Range(func(key, value interface{}) bool {
		session, ok := value.(*ClientSession)
		if !ok {
			return true
		}

		// Re-calculate routes
		var routes []string
		if firewall.Engine != nil {
			routes = firewall.Engine.GetAllowedCIDRs(session.Email, session.Groups, session.OS)
		}

		// Send Update
		go func(sess *ClientSession, r []string) {
			// Open a unidirectional stream for control message
			stream, err := (*sess.Conn).OpenUniStream()
			if err != nil {
				log.Printf("Failed to open stream to client %s: %v", sess.Email, err)
				return
			}
			defer stream.Close()

			updateMsg := map[string]interface{}{
				"type":   "route_update",
				"routes": r,
			}

			if err := json.NewEncoder(stream).Encode(updateMsg); err != nil {
				log.Printf("Failed to encode route update for %s: %v", sess.Email, err)
			}
			log.Printf("✅ Sent route update to %s: %v", sess.Email, r)
		}(session, routes)

		return true
	})
}

// UpdateConfig updates the server configuration dynamically
func (s *Server) UpdateConfig(cidr string, pubKeyPEM string, maxMbps int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if maxMbps > 0 {
		rateLimit := rate.Limit(maxMbps * 125000) // Convert Mbps to Bytes/sec
		burst := int(rateLimit)                   // 1 sec burst
		if burst < 1500 {
			burst = 1500
		}
		if s.GlobalLimiter == nil {
			s.GlobalLimiter = rate.NewLimiter(rateLimit, burst)
		} else {
			s.GlobalLimiter.SetLimit(rateLimit)
			s.GlobalLimiter.SetBurst(burst)
		}
	} else {
		s.GlobalLimiter = nil
	}

	// Update Public Key
	pubKey, err := jwt.ParseEdPublicKeyFromPEM([]byte(pubKeyPEM))
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}
	s.PublicKey = pubKey

	// If CIDR changes, we might need to recreate IP Pool.
	// For simplicity, we assume CIDR doesn't change often or requires restart for net change.
	// But if it's the first time:
	if s.Config == nil && cidr != "" {
		// Setup IP & NAT (moved from Start)
		if err := setupNetwork(cidr); err != nil {
			log.Printf("⚠️ Failed to setup network for CIDR %s: %v", cidr, err)
		}
	} else if s.Config != nil && s.Config.VPNCIDR != cidr {
		log.Println("⚠️ CIDR change detected. This may require restart or complex re-net implementation. Ignoring net re-setup for now.")
	}

	s.Config = &Config{
		VPNCIDR:      cidr,
		PublicKeyPEM: pubKeyPEM,
	}

	return nil
}

func setupNetwork(cidr string) error {
	// 1. System Tuning
	if err := tuneSystem(); err != nil {
		log.Printf("⚠️ System tuning failed: %v", err)
	}

	// 2. Interface Setup
	ipPool := NewFastIPPool(cidr) // Just to get host IP
	hostIP := ipPool.GetHostIPAddress()

	exec.Command("ip", "addr", "add", hostIP, "dev", VPNInterface).Run()
	exec.Command("ip", "link", "set", "dev", VPNInterface, "up", "mtu", "1420").Run()

	// 3. NFTables NAT (Masquerade)
	// Flush legacy iptables to avoid conflicts if possible, or just apply nftables
	// Note: It is often good practice to use one or the other. We assume nftables is present.

	// Create table
	if err := exec.Command("nft", "add", "table", "ip", "tridorian_nat").Run(); err != nil {
		log.Printf("⚠️ Failed to create nft table: %v", err)
	}
	// Create chain
	if err := exec.Command("nft", "add", "chain", "ip", "tridorian_nat", "postrouting", "{ type nat hook postrouting priority 100 ; }").Run(); err != nil {
		log.Printf("⚠️ Failed to create nft chain: %v", err)
	}
	// Add masquerade rule
	if err := exec.Command("nft", "add", "rule", "ip", "tridorian_nat", "postrouting", "ip", "saddr", cidr, "masquerade").Run(); err != nil {
		log.Printf("⚠️ Failed to add nft masquerade rule: %v", err)
	}

	return nil
}

func tuneSystem() error {
	params := map[string]string{
		"net.core.rmem_max":               "67108864",
		"net.core.wmem_max":               "67108864",
		"net.core.rmem_default":           "33554432",
		"net.core.wmem_default":           "33554432",
		"net.netfilter.nf_conntrack_max":  "1048576",
		"net.ipv4.ip_forward":             "1",
		"net.core.default_qdisc":          "fq",
		"net.ipv4.tcp_congestion_control": "bbr",
	}

	for key, value := range params {
		if err := exec.Command("sysctl", "-w", fmt.Sprintf("%s=%s", key, value)).Run(); err != nil {
			log.Printf("⚠️ Failed to set %s: %v", key, err)
		}
	}
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	// Tuning
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	// Create Multiqueue TUN
	s.Ifces = make([]*water.Interface, numCPU)
	for i := range numCPU {
		config := water.Config{
			DeviceType: water.TUN,
			PlatformSpecificParams: water.PlatformSpecificParams{
				MultiQueue: true,
				Name:       VPNInterface,
				Persist:    false,
			},
		}
		var err error
		s.Ifces[i], err = water.New(config)
		if err != nil {
			return fmt.Errorf("failed to create TUN interface %d: %v", i, err)
		}
	}

	// Start Listeners
	tlsConf := generateTLS()
	listener, err := quic.ListenAddr(s.Addr, tlsConf, &quic.Config{
		EnableDatagrams: true,
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.Addr, err)
	}
	log.Printf("🛡️ VPN Server (QUIC) listening on %s", s.Addr)

	// Start TUN Readers (Parallel)
	for i := range numCPU {
		go func(ifce *water.Interface) {
			buf := make([]byte, 1500)
			for {
				n, err := ifce.Read(buf)
				if err != nil {
					log.Printf("TUN Read Error: %v", err)
					return
				}
				packet := buf[:n]
				if len(packet) >= 20 {
					// Extract Dest IP from IPv4 Header (bytes 16-20)
					dstIP := net.IP(packet[16:20]).String()
					if connVal, ok := s.ClientConns.Load(dstIP); ok {
						session, ok := connVal.(*ClientSession)
						if ok {
							// Apply Global Bandwidth Limiter
							if s.GlobalLimiter != nil {
								s.GlobalLimiter.WaitN(context.Background(), len(packet))
							}
							// Encrypt packet with session key
							nonce := make([]byte, session.AEAD.NonceSize())
							rand.Read(nonce)
							encrypted := session.AEAD.Seal(nonce, nonce, packet, nil)

							(*session.Conn).SendDatagram(encrypted)
						}
					}
				}
			}
		}(s.Ifces[i])
	}

	// Accept Loop
	var rr uint64
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("Accept Error: %v", err)
				continue
			}
		}

		// Round-Robin Select Interface for Writing
		idx := int(atomic.AddUint64(&rr, 1) % uint64(numCPU))
		go s.handleClient(conn, s.Ifces[idx])
	}
}

func (s *Server) handleClient(conn *quic.Conn, ifce *water.Interface) {
	// Handshake / Auth
	log.Println("New Client Connection - Accepting Stream")
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		log.Printf("AcceptStream failed: %v", err)
		return
	}
	log.Println("Stream Accepted - Reading Token")
	buf := make([]byte, 8192) // Increased buffer size
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("Stream Read failed: %v", err)
		return
	}
	log.Printf("Read %d bytes from stream", n)

	tokenString := string(buf[:n])
	log.Printf("Received Token (len=%d): %s...", len(tokenString), tokenString[:min(len(tokenString), 50)])

	s.mu.RLock()
	pubKey := s.PublicKey
	s.mu.RUnlock()

	if pubKey == nil {
		log.Println("Server Public Key is nil")
		conn.CloseWithError(1, "Server Not Ready")
		return
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		log.Printf("JWT Parse Error: %v", err)
	}

	if token == nil || !token.Valid {
		conn.CloseWithError(1, "Auth Fail")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		conn.CloseWithError(1, "Invalid Claims")
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		conn.CloseWithError(1, "Email missing in claims")
		return
	}

	// Assign IP
	if s.IPManager == nil {
		conn.CloseWithError(1, "IP Manager Not Ready")
		return
	}

	osInfo, _ := claims["os"].(string)

	var groups []string
	if g, ok := claims["groups"].([]interface{}); ok {
		for _, group := range g {
			if gs, ok := group.(string); ok {
				groups = append(groups, gs)
			}
		}
	}

	// For sticky IP and 1h release, we delegate to Control Plane
	// We use sub as UserID if available, else email
	userID := email
	if sub, ok := claims["sub"].(string); ok {
		userID = sub
	}

	myIP, err := s.IPManager.AssignIP(context.Background(), userID, email)
	if err != nil {
		log.Printf("IP Assignment failed for %s: %v", email, err)
		conn.CloseWithError(1, "IP Full")
		return
	}

	// We append CIDR prefix for client to setup interface
	_, ipNet, _ := net.ParseCIDR(s.Config.VPNCIDR)
	prefix, _ := ipNet.Mask.Size()
	myIPWithCIDR := fmt.Sprintf("%s/%d", myIP, prefix)

	// Prepare JSON Response
	type HandshakeResponse struct {
		AssignedIP string   `json:"assigned_ip"`
		GW_IP      string   `json:"gw_ip"`
		Routes     []string `json:"routes"`
		SessionKey string   `json:"session_key"` // Base64 encoded
	}

	gwIP := s.GetHostIPAddress()

	var routes []string
	if firewall.Engine != nil {
		routes = firewall.Engine.GetAllowedCIDRs(email, groups, osInfo)
	}

	// Generate session key for double encryption
	sessionKey := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(sessionKey); err != nil {
		conn.CloseWithError(1, "Failed to generate session key")
		return
	}

	aead, err := chacha20poly1305.NewX(sessionKey)
	if err != nil {
		conn.CloseWithError(1, "Failed to create cipher")
		return
	}

	resp := HandshakeResponse{
		AssignedIP: myIPWithCIDR,
		GW_IP:      gwIP,
		Routes:     routes,
		SessionKey: fmt.Sprintf("%x", sessionKey), // Send as hex
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		conn.CloseWithError(1, "JSON Marshal Error")
		return
	}

	// Send assigned IP and Config to client
	stream.Write(respBytes)
	stream.Close()

	s.ClientConns.Store(myIP, &ClientSession{
		Conn:        conn,
		UserID:      userID,
		Email:       email,
		Groups:      groups,
		OS:          osInfo,
		SessionKey:  sessionKey,
		AEAD:        aead,
		ConnectedAt: time.Now().Unix(),
	})

	defer func() {
		s.IPManager.ReleaseIP(context.Background(), myIP, email)
		s.ClientConns.Delete(myIP)
	}()

	log.Printf("✅ Client Connected: %s (IP: %s)", email, myIP)

	// Data Loop
	for {
		encryptedData, err := conn.ReceiveDatagram(context.Background())
		if err != nil {
			return
		}

		// Decrypt packet with session key
		var packetData []byte
		if sessionVal, ok := s.ClientConns.Load(myIP); ok {
			if session, ok := sessionVal.(*ClientSession); ok && session.AEAD != nil {
				nonceSize := session.AEAD.NonceSize()
				if len(encryptedData) < nonceSize {
					continue
				}
				nonce := encryptedData[:nonceSize]
				ciphertext := encryptedData[nonceSize:]

				packetData, err = session.AEAD.Open(nil, nonce, ciphertext, nil)
				if err != nil {
					log.Printf("Decryption Error: %v", err)
					continue
				}
			} else {
				packetData = encryptedData
			}
		} else {
			continue
		}

		version := packetData[0] >> 4
		var srcIP, dstIP netip.Addr
		var parseErr error

		if version == 4 {
			srcIP, dstIP, parseErr = parseIPv4Header(packetData)
		} else if version == 6 {
			// IPv6 not supported yet
			continue
		}

		if parseErr != nil {
			continue
		}

		// Rate Limiting (Ingress) - Shared Global Bandwidth Pool
		if s.GlobalLimiter != nil {
			s.GlobalLimiter.WaitN(context.Background(), len(packetData))
		}

		// Conditional Access / Firewall Check
		if firewall.Engine != nil {
			if !firewall.Engine.IsAllowed(packetData, firewall.ValType{
				Addr:     srcIP,
				Identity: email,
				Groups:   groups,
				OS:       osInfo,
			}, firewall.ValType{
				Addr: dstIP,
			}) {
				// log.Printf("Blocked packet from %s to %s", srcIP, dstIP)
				continue
			}
		}

		// Write to TUN Interface
		_, err = ifce.Write(packetData)
		if err != nil {
			log.Printf("TUN Write Error: %v", err)
		}
	}
}

func (s *Server) GetHostIPAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Config == nil || s.Config.VPNCIDR == "" {
		return ""
	}
	ip, ipnet, _ := net.ParseCIDR(s.Config.VPNCIDR)
	prefix, _ := ipnet.Mask.Size()
	ip = ip.To4()
	ip[3] = byte(1)

	return ip.String() + "/" + fmt.Sprint(prefix)
}

func (s *Server) GetActiveSessions() []SessionInfo {
	var sessions []SessionInfo
	s.ClientConns.Range(func(key, value interface{}) bool {
		ip := key.(string)
		sess := value.(*ClientSession)
		sessions = append(sessions, SessionInfo{
			UserID:      sess.UserID,
			Email:       sess.Email,
			IPAddress:   ip,
			ConnectedAt: sess.ConnectedAt,
		})
		return true
	})
	return sessions
}

func generateTLS() *tls.Config {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Tridorian Zero Trust"},
		},
		NotAfter: time.Now().Add(87600 * time.Hour),
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	cert, _ := tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"vpn-quic"},
	}
}

func parseIPv4Header(b []byte) (src, dst netip.Addr, err error) {
	if len(b) < 20 {
		return netip.Addr{}, netip.Addr{}, fmt.Errorf("packet too short")
	}
	srcAddr := [4]byte{b[12], b[13], b[14], b[15]}
	dstAddr := [4]byte{b[16], b[17], b[18], b[19]}
	return netip.AddrFrom4(srcAddr), netip.AddrFrom4(dstAddr), nil
}
