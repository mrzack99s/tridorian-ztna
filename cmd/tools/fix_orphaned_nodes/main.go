package main

import (
	"fmt"
	"log"
	"tridorian-ztna/internal/infrastructure"
	"tridorian-ztna/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	appEnv := utils.GetEnv("APP_ENV", "development")
	if appEnv == "development" {
		godotenv.Load()
	}

	db := infrastructure.SetupDatabase()

	fmt.Println("Cleaning up orphaned access_policy_nodes...")

	// 1. Delete records where access_policy_id does not exist
	result := db.Exec(`
		DELETE FROM access_policy_nodes 
		WHERE access_policy_id NOT IN (SELECT id FROM access_policies)
	`)
	if result.Error != nil {
		log.Fatalf("Failed to clean up orphaned access_policies: %v", result.Error)
	}
	fmt.Printf("Deleted %d orphaned policy associations (invalid policy ID)\n", result.RowsAffected)

	// 2. Delete records where node_id does not exist
	result = db.Exec(`
		DELETE FROM access_policy_nodes 
		WHERE node_id NOT IN (SELECT id FROM nodes)
	`)
	if result.Error != nil {
		log.Fatalf("Failed to clean up orphaned nodes: %v", result.Error)
	}
	fmt.Printf("Deleted %d orphaned node associations (invalid node ID)\n", result.RowsAffected)

	fmt.Println("Cleanup complete. You can now restart the server.")
}
