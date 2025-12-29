package main

import (
	"encoding/json"
	"log"
	"net/http"
	"tridorian-ztna/internal/api/auth"
	"tridorian-ztna/internal/api/mgmt"
	"tridorian-ztna/internal/infrastructure"
	"tridorian-ztna/internal/version"
	"tridorian-ztna/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	appEnv := utils.GetEnv("APP_ENV", "development")
	if appEnv == "development" {
		godotenv.Load()
	}

	log.Println("🚀 Starting Management API Server...")

	// Database
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

	// Cache
	valkey := infrastructure.SetCache()

	// Routers
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
	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Management API is running"))
	})
	mainMux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(version.GetInfo())
	})

	port := utils.GetEnv("MGMT_PORT", "8080")
	log.Printf("🚀 Management API listening on :%s", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mainMux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Management API failed: %v", err)
	}
}
