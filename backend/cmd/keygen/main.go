package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func main() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal("Failed to generate private key:", err)
	}

	// Encode private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Encode public key to PEM format
	publicKeyPKIX, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal("Failed to marshal public key:", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyPKIX,
	}

	// Write private key to file
	privateFile, err := os.Create("private_key.pem")
	if err != nil {
		log.Fatal("Failed to create private key file:", err)
	}
	defer privateFile.Close()

	if err := pem.Encode(privateFile, privateKeyPEM); err != nil {
		log.Fatal("Failed to write private key:", err)
	}

	// Write public key to file
	publicFile, err := os.Create("public_key.pem")
	if err != nil {
		log.Fatal("Failed to create public key file:", err)
	}
	defer publicFile.Close()

	if err := pem.Encode(publicFile, publicKeyPEM); err != nil {
		log.Fatal("Failed to write public key:", err)
	}

	fmt.Println("RSA key pair generated successfully:")
	fmt.Println("- private_key.pem")
	fmt.Println("- public_key.pem")
	fmt.Println("\nAdd these to your .env file:")
	fmt.Printf("PRIVATE_KEY=%s\n", string(pem.EncodeToMemory(privateKeyPEM)))
	fmt.Printf("PUBLIC_KEY=%s\n", string(pem.EncodeToMemory(publicKeyPEM)))
}