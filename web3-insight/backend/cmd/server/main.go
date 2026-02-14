package main

import (
	"fmt"
	"log"

	"github.com/user/web3-insight/internal/api"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
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

	// Sync themes from prompts.yaml to database
	if err := service.SyncThemesFromConfig(db, cfg.Prompts); err != nil {
		log.Printf("Warning: theme sync failed: %v", err)
	}
	log.Println("Theme sync completed")

	// Sync tags from tags.yaml to database
	if err := service.SyncTagsFromConfig(db, cfg.Tags); err != nil {
		log.Printf("Warning: tag sync failed: %v", err)
	}
	log.Println("Tag sync completed")

	// Initialize API key provider (DB-backed with caching)
	configRepo := repository.NewConfigRepository(db)
	keyProvider := service.NewKeyProvider(configRepo)

	router := api.NewRouterWithDB(cfg, db, keyProvider)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
