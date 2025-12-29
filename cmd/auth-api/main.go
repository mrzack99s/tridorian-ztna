package main

import (
	"log"
	"net/http"
	"tridorian-ztna/internal/api/auth"
	"tridorian-ztna/internal/infrastructure"
	"tridorian-ztna/pkg/geoip"
	"tridorian-ztna/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {

	// Security: Load EdDSA Keys from Environment Variables
	// Generate keys using: ./scripts/generate-keys.sh

	appEnv := utils.GetEnv("APP_ENV", "development")
	if appEnv == "development" {
		godotenv.Load()
	}

	privPEM := utils.GetEnv("ZTNA_PRIVATE_KEY", "")
	if privPEM == "" {
		log.Fatal("❌ ZTNA_PRIVATE_KEY environment variable is required. Run: ./scripts/generate-keys.sh")
	}

	pubPEM := utils.GetEnv("ZTNA_PUBLIC_KEY", "")
	if pubPEM == "" {
		log.Fatal("❌ ZTNA_PUBLIC_KEY environment variable is required. Run: ./scripts/generate-keys.sh")
	}

	db := infrastructure.SetupDatabase()
	valkey := infrastructure.SetCache()

	geoIP := geoip.New()
	if err := geoIP.Load("ip2country-v4.tsv"); err != nil {
		log.Printf("⚠️ Failed to load ip2country-v4.tsv: %v", err)
		// Proceeding without GeoIP might be dangerous if policies rely on it,
		// but maybe we just log it for now.
	} else {
		log.Printf("🌍 GeoIP loaded with %d ranges", len(geoIP.Ranges))
	}

	privKey, err := utils.ParseEdPrivateKeyFromPEM(privPEM)
	if err != nil {
		log.Fatalf("failed to parse private key: %v", err)
	}

	pubKey, err := utils.ParseEdPublicKeyFromPEM(pubPEM)
	if err != nil {
		log.Fatalf("failed to parse public key: %v", err)
	}

	router := auth.NewRouter(db, valkey, geoIP, privKey, pubKey)

	port := utils.GetEnv("AUTH_PORT", "8081")
	log.Printf("🔐 Authentication API starting on :%s", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Authentication API failed: %v", err)
	}
}
