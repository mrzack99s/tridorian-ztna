package gateway

import (
	"context"
	pb "tridorian-ztna/internal/proto/gateway/v1"
	"tridorian-ztna/internal/services"

	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedGatewayServiceServer
	nodeService   *services.NodeService
	policyService *services.PolicyService
	publicKeyPEM  string
}

func NewServer(nodeService *services.NodeService, policyService *services.PolicyService, publicKeyPEM string) *Server {
	return &Server{
		nodeService:   nodeService,
		policyService: policyService,
		publicKeyPEM:  publicKeyPEM,
	}
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	nodeID, err := uuid.Parse(req.NodeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node_id format")
	}

	ipAddress := ""
	if p, ok := peer.FromContext(ctx); ok {
		ipAddress = p.Addr.String()
		if host, _, err := net.SplitHostPort(ipAddress); err == nil {
			ipAddress = host
		}

		// Handle IPv6 loopback mapping to IPv4 loopback for consistency if needed
		if ipAddress == "::1" {
			ipAddress = "127.0.0.1"
		}
	}

	// Extract version from metadata
	clientVersion := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if versions := md.Get("x-gateway-version"); len(versions) > 0 {
			clientVersion = versions[0]
		}
	}

	token, err := s.nodeService.RegisterGateway(nodeID, req.Hostname, req.DeviceHash, ipAddress, clientVersion)
	if err != nil {
		if err.Error() == "gateway node not found" {
			return nil, status.Error(codes.NotFound, "gateway not found")
		}
		// Assuming other errors are permission related or conflict
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return &pb.RegisterResponse{
		AuthToken: token,
	}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.AuthToken == "" {
		return nil, status.Error(codes.Unauthenticated, "auth_token is required")
	}

	// 1. Authenticate Node
	node, err := s.nodeService.GetNodeByToken(req.AuthToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth_token")
	}

	// 2. Load Policies
	// 2. Load Policies
	policies, err := s.policyService.ListAccessPoliciesByNodeID(node.TenantID, node.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load policies")
	}

	// 3. Generate Gateway Config & Calculate Hash
	gatewayPolicies := services.GenerateGatewayPolicies(policies)
	currentHash := services.CalculateConfigHash(gatewayPolicies)

	updateAvailable := currentHash != req.ConfigHash

	// 4. Update Node Status/Heartbeat in Valkey
	_ = s.nodeService.UpdateHeartbeat(node.ID)

	// Simple stub for now
	return &pb.HeartbeatResponse{
		Success:               true,
		ConfigUpdateAvailable: updateAvailable,
	}, nil
}

func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	if req.AuthToken == "" {
		return nil, status.Error(codes.Unauthenticated, "auth_token is required")
	}

	// 1. Authenticate Node
	node, err := s.nodeService.GetNodeByToken(req.AuthToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth_token")
	}

	// 2. Load Policies
	// 2. Load Policies
	policies, err := s.policyService.ListAccessPoliciesByNodeID(node.TenantID, node.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load policies")
	}

	// 3. Generate Gateway Config
	gatewayPolicies := services.GenerateGatewayPolicies(policies)
	currentHash := services.CalculateConfigHash(gatewayPolicies)

	// Return config from Node and generated policies
	return &pb.GetConfigResponse{
		VpnCidr:          node.ClientCIDR,
		PublicKeyPem:     s.publicKeyPEM, // Use global/tenant key for verification
		ConfigHash:       currentHash,
		Policies:         gatewayPolicies,
		MaxBandwidthMbps: node.NodeSku.Bandwidth,
	}, nil
}

func (s *Server) GetSessionIP(ctx context.Context, req *pb.GetSessionIPRequest) (*pb.GetSessionIPResponse, error) {
	if req.AuthToken == "" {
		return nil, status.Error(codes.Unauthenticated, "auth_token is required")
	}

	// 1. Authenticate Node
	node, err := s.nodeService.GetNodeByToken(req.AuthToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth_token")
	}

	// 2. Get/Allocate IP
	ip, err := s.nodeService.GetSessionIP(node.ID, req.UserId, req.UserEmail)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetSessionIPResponse{
		IpAddress: ip,
	}, nil
}

func (s *Server) SyncSessions(ctx context.Context, req *pb.SyncSessionsRequest) (*pb.SyncSessionsResponse, error) {
	if req.AuthToken == "" {
		return nil, status.Error(codes.Unauthenticated, "auth_token is required")
	}

	// 1. Authenticate Node
	node, err := s.nodeService.GetNodeByToken(req.AuthToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth_token")
	}

	// 2. Sync
	err = s.nodeService.SyncSessions(node.ID, req.Sessions)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SyncSessionsResponse{
		Success: true,
	}, nil
}
