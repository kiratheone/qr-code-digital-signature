package main

import (
	"log"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/infrastructure/database"
	"digital-signature-system/internal/infrastructure/handlers"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize HTTP server
	server := handlers.NewServer(cfg, db)
	
	log.Printf("Server starting on port %s", cfg.Port)
	if err := server.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}