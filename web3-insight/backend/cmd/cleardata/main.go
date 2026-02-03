// backend/cmd/cleardata/main.go
package main

import (
	"fmt"
	"log"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("⚠️  WARNING: This will delete ALL data from the database!")
	fmt.Print("Type 'yes' to confirm: ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("Aborted.")
		return
	}

	// Clear tables in order (respecting foreign keys)
	tables := []string{
		"article_versions",
		"chat_messages",
		"articles",
		"categories",
		"tasks",
		"news_items",
		"data_sources",
	}

	for _, table := range tables {
		result := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if result.Error != nil {
			log.Printf("Warning: Failed to truncate %s: %v", table, result.Error)
		} else {
			fmt.Printf("✓ Cleared table: %s\n", table)
		}
	}

	fmt.Println("\n✅ Database cleared successfully!")
}
