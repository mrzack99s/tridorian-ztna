package main

import (
	"log"
	"net/http"
	"tridorian-ztna/internal/api/auth"
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
	router := auth.NewRouter(db)

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
