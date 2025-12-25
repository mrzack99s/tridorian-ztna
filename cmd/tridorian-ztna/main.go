package main

import (
	"log"
	"net/http"
	"tridorian-ztna/internal/api/auth"
	"tridorian-ztna/internal/api/mgmt"
	"tridorian-ztna/internal/infrastructure"
	"tridorian-ztna/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	appEnv := utils.GetEnv("APP_ENV", "development")
	if appEnv == "development" {
		godotenv.Load("../../.env")
	}

	db := infrastructure.SetupDatabase()
	jwtSecret := utils.GetEnv("JWT_SECRET", "very-secret-key")

	mgmtRouter := mgmt.NewRouter(db, jwtSecret)
	authRouter := auth.NewRouter(db)

	// Combine routers
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/", mgmtRouter)
	mainMux.Handle("/auth/", authRouter)
	mainMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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
