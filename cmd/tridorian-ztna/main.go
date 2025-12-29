package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"tridorian-ztna/internal/api/auth"
	"tridorian-ztna/internal/api/mgmt"
	"tridorian-ztna/internal/grpc/gateway"
	"tridorian-ztna/internal/infrastructure"
	pb "tridorian-ztna/internal/proto/gateway/v1"
	"tridorian-ztna/internal/services"
	"tridorian-ztna/internal/version"
	"tridorian-ztna/pkg/utils"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	appEnv := utils.GetEnv("APP_ENV", "development")
	if appEnv == "development" {
		godotenv.Load()
	}

	db := infrastructure.SetupDatabase()

	// Security: Load EdDSA Keys from Environment Variables
	// Generate keys using: ./scripts/generate-keys.sh
	privPEM := utils.GetEnv("ZTNA_PRIVATE_KEY", "")
	if privPEM == "" {
		log.Fatal("❌ ZTNA_PRIVATE_KEY environment variable is required. Run: ./scripts/generate-keys.sh")
	}

	pubPEM := utils.GetEnv("ZTNA_PUBLIC_KEY", "")
	if pubPEM == "" {
		log.Fatal("❌ ZTNA_PUBLIC_KEY environment variable is required. Run: ./scripts/generate-keys.sh")
	}

	privKey, err := utils.ParseEdPrivateKeyFromPEM(privPEM)
	if err != nil {
		log.Fatalf("failed to parse private key: %v", err)
	}

	pubKey, err := utils.ParseEdPublicKeyFromPEM(pubPEM)
	if err != nil {
		log.Fatalf("failed to parse public key: %v", err)
	}

	log.Printf("🔑 Public Key (PEM):\n%s", pubPEM)

	// Infrastructure
	valkey := infrastructure.SetCache()

	// Services
	nodeService := services.NewNodeService(db, valkey)
	policyService := services.NewPolicyService(db, valkey)

	mgmtRouter := mgmt.NewRouter(db, valkey, privKey, pubKey)
	authRouter := auth.NewRouter(db, valkey, nil, privKey, pubKey)

	// Combine routers
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/", mgmtRouter)
	mainMux.Handle("/auth/", authRouter)
	mainMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mainMux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(version.GetInfo())
	})

	// Start gRPC Server
	grpcPort := utils.GetEnv("GRPC_PORT", "5443")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	gatewayServer := gateway.NewServer(nodeService, policyService, pubPEM)
	pb.RegisterGatewayServiceServer(grpcServer, gatewayServer)

	log.Printf("🔌 gRPC Gateway Server starting on :%s", grpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	port := utils.GetEnv("MGMT_PORT", "8080")
	log.Printf("🚀 Management API starting on :%s", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mainMux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Management API failed: %v", err)
	}
}
