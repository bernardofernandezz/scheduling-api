package main

import (
	"log"

	"github.com/bernardofernandezz/scheduling-api/internal/api/routes"
	"github.com/bernardofernandezz/scheduling-api/internal/config"
	"github.com/bernardofernandezz/scheduling-api/internal/repository"
)

func main() {
	log.Println("Starting Scheduling API server...")

	// Load application configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := repository.NewDBConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Migrate database schema
	if err := repos.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully")

	// Initialize router
	router := routes.SetupRouter(repos, cfg)

	// Start server
	log.Printf("Server starting on %s in %s mode", cfg.Server.Address, cfg.Server.Mode)
	if err := router.Run(cfg.Server.Address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

