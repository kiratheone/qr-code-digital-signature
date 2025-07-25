package main

import (
	"digital-signature-system/internal/domain/services"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	// Default key size
	keySize := 2048

	// Check if key size is provided as argument
	if len(os.Args) > 1 {
		size, err := strconv.Atoi(os.Args[1])
		if err == nil && size >= 1024 {
			keySize = size
		}
	}

	// Create signature service
	service, err := services.NewSignatureService("", "")
	if err != nil {
		log.Fatalf("Failed to create signature service: %v", err)
	}

	// Generate key pair
	privateKey, publicKey, err := service.GenerateKeyPair(keySize)
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Write private key to file
	err = os.WriteFile("private_key.pem", []byte(privateKey), 0600)
	if err != nil {
		log.Fatalf("Failed to write private key: %v", err)
	}

	// Write public key to file
	err = os.WriteFile("public_key.pem", []byte(publicKey), 0644)
	if err != nil {
		log.Fatalf("Failed to write public key: %v", err)
	}

	fmt.Printf("Generated %d-bit RSA key pair:\n", keySize)
	fmt.Println("- Private key: private_key.pem")
	fmt.Println("- Public key: public_key.pem")
	fmt.Println("\nIMPORTANT: Keep your private key secure!")
}