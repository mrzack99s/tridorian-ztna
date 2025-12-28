package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"tridorian-ztna/internal/gateway/firewall"
	"tridorian-ztna/internal/gateway/vpn"
	pb "tridorian-ztna/internal/proto/gateway/v1"
	"tridorian-ztna/internal/version"
	"tridorian-ztna/pkg/utils"
)

func main() {
	log.Printf("🚀 Starting Gateway... %s", version.String())
	godotenv.Load()

	// Flags
	nodeIDFlag := flag.String("node-id", "", "The unique Node ID (UUID) for this gateway")
	controlPlaneFlag := flag.String("control-plane", "", "The address of the Control Plane (e.g., localhost:5443)")
	hostnameFlag := flag.String("hostname", "", "The hostname of this gateway")
	installServiceFlag := flag.Bool("install-service", false, "Install as a systemd service")
	flag.Parse()

	// Install Service Mode
	if *installServiceFlag {
		installService(*nodeIDFlag, *controlPlaneFlag)
		return
	}

	// Configuration Priority: Flag > Env > Default
	controlPlaneAddr := *controlPlaneFlag
	if controlPlaneAddr == "" {
		controlPlaneAddr = utils.GetEnv("CONTROL_PLANE_ADDR", "localhost:5443")
	}

	nodeID := *nodeIDFlag
	if nodeID == "" {
		nodeID = utils.GetEnv("NODE_ID", "")
	}

	hostname := *hostnameFlag
	if hostname == "" {
		hostname = utils.GetEnv("HOSTNAME", "")
	}

	if hostname == "" {
		h, _ := os.Hostname()
		hostname = h
	}

	if nodeID == "" {
		log.Fatal("❌ NODE_ID is required. Use --node-id flag or set NODE_ID environment variable.")
	}

	// Get Device Hash from actual Hardware
	deviceHash := getDeviceHash()
	log.Printf("💻 Device Hash: %s", deviceHash)

	log.Printf("🔌 Connecting to Control Plane at %s...", controlPlaneAddr)

	// Connect to gRPC Server
	conn, err := grpc.Dial(controlPlaneAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGatewayServiceClient(conn)

	// Register
	log.Printf("📝 Registering Gateway [NodeID: %s, Host: %s]", nodeID, hostname)
	// Create context with metadata
	md := metadata.New(map[string]string{
		"x-gateway-version": version.Version,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	regResp, err := client.Register(ctx, &pb.RegisterRequest{
		NodeId:     nodeID,
		Hostname:   hostname,
		DeviceHash: deviceHash,
	})
	if err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}

	token := regResp.AuthToken
	log.Printf("✅ Registration Successful! Token acquired.")

	// Start VPN Server
	vpnPort := utils.GetEnv("VPN_PORT", "6500")
	vpnAddr := ":" + vpnPort

	vpnServer := vpn.NewServer(vpnAddr)
	vpnServer.IPManager = &grpcIPManager{
		client: client,
		token:  token,
	}
	go func() {
		if err := vpnServer.Start(context.Background()); err != nil {
			log.Printf("❌ VPN Server failed: %v", err)
		}
	}()

	// Initial Config fetch
	if err := getAndApplyConfig(client, token, vpnServer); err != nil {
		log.Printf("❌ Failed to get initial config: %v", err)
	}

	// Start Heartbeat Loop
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial Heartbeat
	sendHeartbeat(client, token, vpnServer)

	for range ticker.C {
		sendHeartbeat(client, token, vpnServer)
	}
}

var currentConfigHash = "none"

func sendHeartbeat(client pb.GatewayServiceClient, token string, vpnServer *vpn.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Heartbeat(ctx, &pb.HeartbeatRequest{
		AuthToken:  token,
		Status:     "ONLINE",
		ConfigHash: currentConfigHash,
	})

	if err != nil {
		log.Printf("❌ Heartbeat failed: %v", err)
	} else {
		log.Printf("💓 Heartbeat sent (Hash: %s)", currentConfigHash)

		// 5. Sync active sessions
		sessions := vpnServer.GetActiveSessions()
		var pbSessions []*pb.SyncSessionsRequest_Session
		for _, s := range sessions {
			pbSessions = append(pbSessions, &pb.SyncSessionsRequest_Session{
				UserId:      s.UserID,
				UserEmail:   s.Email,
				IpAddress:   s.IPAddress,
				ConnectedAt: s.ConnectedAt,
			})
		}
		_, err := client.SyncSessions(ctx, &pb.SyncSessionsRequest{
			AuthToken: token,
			Sessions:  pbSessions,
		})
		if err != nil {
			log.Printf("❌ Session sync failed: %v", err)
		}

		if resp.ConfigUpdateAvailable {
			log.Printf("📥 Config update available, pulling...")
			if err := getAndApplyConfig(client, token, vpnServer); err != nil {
				log.Printf("❌ Failed to update config: %v", err)
			}
		}
	}
}

func getAndApplyConfig(client pb.GatewayServiceClient, token string, vpnServer *vpn.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetConfig(ctx, &pb.GetConfigRequest{AuthToken: token})
	if err != nil {
		return err
	}

	log.Printf("📥 Received Config: CIDR=%s, Policies=%d, Hash=%s", resp.VpnCidr, len(resp.Policies), resp.ConfigHash)
	currentConfigHash = resp.ConfigHash

	fmt.Println(resp)

	// 1. Update VPN Server (Key + CIDR)
	if err := vpnServer.UpdateConfig(resp.VpnCidr, resp.PublicKeyPem, resp.MaxBandwidthMbps); err != nil {
		return err
	}

	// 2. Update Firewall Policies
	var policies []firewall.ConditionalAccessPolicy
	for _, p := range resp.Policies {
		policies = append(policies, firewall.ConditionalAccessPolicy{
			Name:                  p.Name,
			Action:                p.Action,
			SourceTagType:         p.SourceTagType,
			SourceMatchValue:      p.SourceMatchValue,
			DestinationTagType:    p.DestinationTagType,
			DestinationMatchValue: p.DestinationMatchValue,
			Priority:              int(p.Priority),
		})
	}

	// Update Global Config (for Firewall Engine)
	firewall.CurrentConfig = &firewall.AgentExecutionConfig{
		ConditionalAccessDefaultPolicies: firewall.ConditionalAccessDefaultPolicies{
			BlockByDefault: true,
		},
		ConditionalAccessPolicies: policies,
	}

	// Re-init Engine
	firewall.NewEngine()

	// 3. Broadcast Config/Route updates to connected clients
	vpnServer.BroadcastRouteUpdates()

	return nil
}

func getDeviceHash() string {
	// Try to gather hardware signatures
	var signatures []string

	// 1. Board Serial
	if content, err := os.ReadFile("/sys/class/dmi/id/board_serial"); err == nil {
		signatures = append(signatures, strings.TrimSpace(string(content)))
	}

	// 2. Product UUID
	if content, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
		signatures = append(signatures, strings.TrimSpace(string(content)))
	}

	// 3. Machine ID (Linux specific)
	if content, err := os.ReadFile("/etc/machine-id"); err == nil {
		signatures = append(signatures, strings.TrimSpace(string(content)))
	}

	if len(signatures) == 0 {
		// Fallback for dev/container environments: hostname + environment specific fallback
		log.Println("⚠️  Could not read DMI/hardware info. Using hostname fallback.")
		signatures = append(signatures, os.Getenv("HOSTNAME"))
	}

	// Combine signatures and hash
	rawData := strings.Join(signatures, "|")
	hash := sha256.Sum256([]byte(rawData))
	return hex.EncodeToString(hash[:])
}

type grpcIPManager struct {
	client pb.GatewayServiceClient
	token  string
}

func (m *grpcIPManager) AssignIP(ctx context.Context, userID, email string) (string, error) {
	resp, err := m.client.GetSessionIP(ctx, &pb.GetSessionIPRequest{
		AuthToken: m.token,
		UserId:    userID,
		UserEmail: email,
	})
	if err != nil {
		return "", err
	}
	return resp.IpAddress, nil
}

func (m *grpcIPManager) ReleaseIP(ctx context.Context, ip, email string) {
	// For now, the control plane handles release after 1h automatically.
	// But we could explicitly notify if needed.
}

func (m *grpcIPManager) SyncSessions(ctx context.Context, sessions []vpn.SessionInfo) {
	// Handled in heartbeat loop
}
