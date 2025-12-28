package main

import (
	"fmt"
	"log"

	"tridorian-ztna/internal/infrastructure"
	"tridorian-ztna/internal/models"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	db := infrastructure.SetupDatabase()

	var node models.Node
	if err := db.Where("id = ?", "d54bf644-b116-4de0-ae93-887d12940d39").First(&node).Error; err != nil {
		log.Fatalf("Failed to find node: %v", err)
	}

	fmt.Printf("Node ID: %s\n", node.ID)
	fmt.Printf("Gateway Version: '%s'\n", node.GatewayVersion)
}
