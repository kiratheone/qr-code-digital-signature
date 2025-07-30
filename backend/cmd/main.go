package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/infrastructure/database"
	"digital-signature-system/internal/infrastructure/di"
	"digital-signature-system/internal/infrastructure/server"
	"digital-signature-system/internal/infrastructure/services"

	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	healthCheck := flag.Bool("health-check", false, "Run health check")
	flag.Parse()

	// Handle health check
	if *healthCheck {
		performHealthCheck()
		return
	}

	// Load environment variables from .env file if it exists (optional)
	if err := godotenv.Load(); err != nil {
		// This is expected in containerized environments where env vars are passed directly
		log.Println("No .env file found - using environment variables from container")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize services
	keyService, err := services.NewKeyService(cfg)
	if err != nil {
		log.Fatal("Failed to initialize key service:", err)
	}

	// Load or generate keys
	privateKey, publicKey, err := services.LoadOrGenerateKeys(keyService)
	if err != nil {
		log.Fatal("Failed to load or generate keys:", err)
	}

	// Update config with loaded keys
	cfg.Security.PrivateKey = privateKey
	cfg.Security.PublicKey = publicKey

	// Check for key rotation
	keyRotationService := services.NewKeyRotationService(cfg, keyService)
	rotated, err := keyRotationService.CheckAndRotateKeys()
	if err != nil {
		log.Println("Warning: Key rotation check failed:", err)
	} else if rotated {
		log.Println("Keys were rotated successfully")
		
		// Reload keys after rotation
		privateKey, publicKey, err = services.LoadOrGenerateKeys(keyService)
		if err != nil {
			log.Fatal("Failed to reload keys after rotation:", err)
		}
		
		// Update config with new keys
		cfg.Security.PrivateKey = privateKey
		cfg.Security.PublicKey = publicKey
	}

	// Initialize dependency injection container
	container := di.NewContainer(cfg, db)
	
	// Initialize and start server
	srv := server.New(cfg, db, container)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// performHealthCheck performs a health check for the application
func performHealthCheck() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/health", port))
	if err != nil {
		log.Printf("Health check failed: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Health check failed with status: %d", resp.StatusCode)
		os.Exit(1)
	}

	log.Println("Health check passed")
	os.Exit(0)
}