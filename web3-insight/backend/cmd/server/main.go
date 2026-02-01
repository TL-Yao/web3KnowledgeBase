package main

import (
	"fmt"
	"log"

	"github.com/user/web3-insight/internal/api"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	// Seed initial data
	if err := database.Seed(db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}
	log.Println("Seed data loaded")

	router := api.NewRouterWithDB(cfg, db)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
