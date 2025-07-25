package services_test

import (
	"digital-signature-system/internal/domain/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRSASignatureService(t *testing.T) {
	// Generate a key pair for testing
	signatureService := &services.RSASignatureService{}
	privateKeyPEM, publicKeyPEM, err := signatureService.GenerateKeyPair(2048)
	require.NoError(t, err)
	require.NotEmpty(t, privateKeyPEM)
	require.NotEmpty(t, publicKeyPEM)

	// Create a new signature service with the generated keys
	service, err := services.NewRSASignatureService(privateKeyPEM, publicKeyPEM)
	require.NoError(t, err)

	// Test document signing and verification
	document := []byte("This is a test document")
	docHash := services.CalculateHash(document)

	// Sign the document
	signature, err := service.SignDocument(docHash)
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// Verify the signature
	valid, err := service.VerifySignature(docHash, signature)
	require.NoError(t, err)
	assert.True(t, valid)

	// Test with modified document
	modifiedDocument := []byte("This is a modified document")
	modifiedHash := services.CalculateHash(modifiedDocument)

	valid, err = service.VerifySignature(modifiedHash, signature)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestSignatureEncoding(t *testing.T) {
	// Generate a key pair for testing
	signatureService := &services.RSASignatureService{}
	privateKeyPEM, publicKeyPEM, err := signatureService.GenerateKeyPair(2048)
	require.NoError(t, err)

	// Create a new signature service with the generated keys
	service, err := services.NewRSASignatureService(privateKeyPEM, publicKeyPEM)
	require.NoError(t, err)

	// Test document signing and encoding/decoding
	document := []byte("This is a test document")
	docHash := services.CalculateHash(document)

	// Sign the document
	signature, err := service.SignDocument(docHash)
	require.NoError(t, err)

	// Encode the signature
	encodedSignature := services.EncodeSignature(signature)
	require.NotEmpty(t, encodedSignature)

	// Decode the signature
	decodedSignature, err := services.DecodeSignature(encodedSignature)
	require.NoError(t, err)
	assert.Equal(t, signature, decodedSignature)

	// Verify the decoded signature
	valid, err := service.VerifySignature(docHash, decodedSignature)
	require.NoError(t, err)
	assert.True(t, valid)
}