package main

import (
	"fmt"
	"log"

	"github.com/user/web3-insight/internal/api"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
)

func main() {
	// Load all configuration (config.yaml, models.yaml, routing.yaml)
	cfg, err := config.LoadAll()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Log any configuration warnings
	if len(cfg.Warnings) > 0 {
		fmt.Println("⚠️  Configuration warnings detected:")
		for _, warning := range cfg.Warnings {
			fmt.Printf("   - %s\n", warning)
		}
	}

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected")

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed")

	router := api.NewRouterWithDB(cfg, db)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
