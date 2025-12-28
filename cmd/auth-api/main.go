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

	privPEM := `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEILI5SZI/4pD9D6Zz4r4fz6l291myurzDhcHx/KvAWb2X
-----END PRIVATE KEY-----`
	pubPEM := `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAox3VKM3biK8+yLk8O600A/N91BHiPZOfY0Oqalp6zqA=
-----END PUBLIC KEY-----`

	appEnv := utils.GetEnv("APP_ENV", "development")
	if appEnv == "development" {
		godotenv.Load()
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
