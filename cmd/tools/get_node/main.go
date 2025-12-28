package main

import (
	"fmt"
	"log"

	"tridorian-ztna/internal/infrastructure"
	"tridorian-ztna/internal/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	db := infrastructure.SetupDatabase()

	// 1. Get Tenant
	var tenant models.Tenant
	if err := db.First(&tenant).Error; err != nil {
		// Create a dummy tenant
		tenant = models.Tenant{
			BaseModel: models.BaseModel{ID: uuid.New()},
			Name:      "CLI Test Tenant",
			Slug:      "cli-test-" + uuid.New().String()[:8], // Ensure uniqueness
		}
		if err := db.Create(&tenant).Error; err != nil {
			log.Fatalf("Failed to create tenant: %v", err)
		}
	}

	// 2. Get SKU
	var sku models.NodeSku
	if err := db.First(&sku).Error; err != nil {
		// Create dummy sku
		sku = models.NodeSku{
			BaseModel: models.BaseModel{ID: uuid.New()},
			Name:      "Small Gateway",
		}
		if err := db.Create(&sku).Error; err != nil {
			log.Fatalf("Failed to create SKU: %v", err)
		}
	}

	// 3. Get or Create Node
	var node models.Node
	// Try to find one that is NOT registered (AuthToken is null)
	if err := db.Where("auth_token IS NULL").First(&node).Error; err != nil {
		// Create one
		node = models.Node{
			BaseTenant: models.BaseTenant{TenantID: tenant.ID},
			BaseModel:  models.BaseModel{ID: uuid.New()},
			Name:       "CLI Gateway",
			Status:     "PENDING_REGISTRATION",
			IsActive:   true,
			NodeSkuID:  sku.ID,
			ClientCIDR: "10.0.0.0/24",
		}
		if err := db.Create(&node).Error; err != nil {
			log.Fatalf("Failed to create Node: %v", err)
		}
	}

	fmt.Println(node.ID.String())
}
