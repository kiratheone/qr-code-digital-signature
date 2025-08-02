package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"digital-signature-system/internal/infrastructure/crypto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: keyrotate <command>")
		fmt.Println("Commands:")
		fmt.Println("  generate - Generate a new key pair")
		fmt.Println("  rotate   - Rotate current keys to new ones")
		fmt.Println("  validate - Validate current keys")
		fmt.Println("  info     - Show current key information")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generateNewKeyPair()
	case "rotate":
		rotateKeys()
	case "validate":
		validateKeys()
	case "info":
		showKeyInfo()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func generateNewKeyPair() {
	fmt.Println("Generating new RSA-2048 key pair...")

	// Create a temporary key manager to generate new keys
	km, err := crypto.NewKeyManager()
	if err != nil {
		log.Fatalf("Failed to create key manager: %v", err)
	}

	// Generate new key pair
	newKeyPair, err := km.GenerateNewKeyPair(2048)
	if err != nil {
		log.Fatalf("Failed to generate new key pair: %v", err)
	}

	fmt.Printf("New key pair generated successfully!\n")
	fmt.Printf("Key ID: %s\n", newKeyPair.KeyID)
	fmt.Printf("Algorithm: %s\n", newKeyPair.Algorithm)
	fmt.Printf("Created: %s\n", newKeyPair.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Encode keys as base64 for environment variables
	privateKeyB64 := base64.StdEncoding.EncodeToString([]byte(newKeyPair.PrivateKey))
	publicKeyB64 := base64.StdEncoding.EncodeToString([]byte(newKeyPair.PublicKey))

	fmt.Println("Environment variables for .env file:")
	fmt.Printf("PRIVATE_KEY=%s\n", privateKeyB64)
	fmt.Printf("PUBLIC_KEY=%s\n", publicKeyB64)
	fmt.Println()

	fmt.Println("Raw PEM format:")
	fmt.Println("Private Key:")
	fmt.Println(newKeyPair.PrivateKey)
	fmt.Println("Public Key:")
	fmt.Println(newKeyPair.PublicKey)
}

func rotateKeys() {
	fmt.Println("Rotating keys...")

	// Load current key manager
	km, err := crypto.NewKeyManager()
	if err != nil {
		log.Fatalf("Failed to load current key manager: %v", err)
	}

	currentKeyID := km.GetKeyID()
	fmt.Printf("Current Key ID: %s\n", currentKeyID)

	// Generate new key pair
	fmt.Println("Generating new key pair...")
	newKeyPair, err := km.GenerateNewKeyPair(2048)
	if err != nil {
		log.Fatalf("Failed to generate new key pair: %v", err)
	}

	// Rotate to new keys
	fmt.Println("Rotating to new keys...")
	err = km.RotateKeys(newKeyPair)
	if err != nil {
		log.Fatalf("Failed to rotate keys: %v", err)
	}

	fmt.Printf("Keys rotated successfully!\n")
	fmt.Printf("Old Key ID: %s\n", currentKeyID)
	fmt.Printf("New Key ID: %s\n", km.GetKeyID())
	fmt.Println()

	// Export new keys for storage
	exportedKeyPair, err := km.ExportKeyPairForStorage()
	if err != nil {
		log.Fatalf("Failed to export new keys: %v", err)
	}

	// Encode keys as base64 for environment variables
	privateKeyB64 := base64.StdEncoding.EncodeToString([]byte(exportedKeyPair.PrivateKey))
	publicKeyB64 := base64.StdEncoding.EncodeToString([]byte(exportedKeyPair.PublicKey))

	fmt.Println("New environment variables for .env file:")
	fmt.Printf("PRIVATE_KEY=%s\n", privateKeyB64)
	fmt.Printf("PUBLIC_KEY=%s\n", publicKeyB64)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Update your .env file with the new keys and restart the application!")
}

func validateKeys() {
	fmt.Println("Validating current keys...")

	km, err := crypto.NewKeyManager()
	if err != nil {
		log.Fatalf("Failed to load key manager: %v", err)
	}

	err = km.ValidateKeys()
	if err != nil {
		fmt.Printf("❌ Key validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Keys are valid!")
	fmt.Printf("Key ID: %s\n", km.GetKeyID())
	fmt.Printf("Loaded at: %s\n", km.GetCreatedAt().Format("2006-01-02 15:04:05"))
}

func showKeyInfo() {
	fmt.Println("Current key information:")

	km, err := crypto.NewKeyManager()
	if err != nil {
		log.Fatalf("Failed to load key manager: %v", err)
	}

	fmt.Printf("Key ID: %s\n", km.GetKeyID())
	fmt.Printf("Loaded at: %s\n", km.GetCreatedAt().Format("2006-01-02 15:04:05"))
	fmt.Printf("Key size: %d bits\n", km.GetKeySize())
	fmt.Printf("Key age: %s\n", km.GetKeyAge().Round(time.Second))

	// Check if key should be rotated (6 months)
	sixMonths := 6 * 30 * 24 * time.Hour
	if km.ShouldRotateKey(sixMonths) {
		fmt.Printf("⚠️  Key rotation recommended (older than 6 months)\n")
	} else {
		fmt.Printf("✅ Key age is acceptable\n")
	}

	// Validate keys
	err = km.ValidateKeys()
	if err != nil {
		fmt.Printf("Status: ❌ Invalid (%v)\n", err)
	} else {
		fmt.Printf("Status: ✅ Valid\n")
	}

	// Show which environment variables are being used
	if os.Getenv("RSA_PRIVATE_KEY") != "" {
		fmt.Println("Source: RSA_PRIVATE_KEY and RSA_PUBLIC_KEY environment variables")
	} else if os.Getenv("PRIVATE_KEY") != "" {
		fmt.Println("Source: PRIVATE_KEY and PUBLIC_KEY environment variables")
	} else {
		fmt.Println("Source: File-based keys")
	}
}