package services_test

import (
	"bytes"
	"digital-signature-system/internal/domain/services"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSHA256HashService(t *testing.T) {
	hashService := services.NewSHA256HashService()

	// Test CalculateHash
	document := []byte("This is a test document")
	hash, err := hashService.CalculateHash(document)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	assert.Len(t, hash, 32) // SHA-256 produces 32-byte hashes

	// Test CalculateHashFromReader
	reader := strings.NewReader("This is a test document")
	readerHash, err := hashService.CalculateHashFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, hash, readerHash)

	// Test VerifyHash with correct document
	valid, err := hashService.VerifyHash(document, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// Test VerifyHash with modified document
	modifiedDocument := []byte("This is a modified document")
	valid, err = hashService.VerifyHash(modifiedDocument, hash)
	require.NoError(t, err)
	assert.False(t, valid)

	// Test HashToString and StringToHash
	hashStr := hashService.HashToString(hash)
	require.NotEmpty(t, hashStr)

	decodedHash, err := hashService.StringToHash(hashStr)
	require.NoError(t, err)
	assert.Equal(t, hash, decodedHash)
}

func TestCalculateHashFromReader(t *testing.T) {
	hashService := services.NewSHA256HashService()

	// Test with a large document
	var largeDoc bytes.Buffer
	for i := 0; i < 10000; i++ {
		largeDoc.WriteString("This is line " + string(rune(i)) + " of the document.\n")
	}

	// Calculate hash from bytes
	hash1, err := hashService.CalculateHash(largeDoc.Bytes())
	require.NoError(t, err)

	// Calculate hash from reader
	hash2, err := hashService.CalculateHashFromReader(bytes.NewReader(largeDoc.Bytes()))
	require.NoError(t, err)

	// Hashes should be identical
	assert.Equal(t, hash1, hash2)

	// Test with empty document
	emptyHash1, err := hashService.CalculateHash([]byte{})
	require.NoError(t, err)

	emptyHash2, err := hashService.CalculateHashFromReader(bytes.NewReader([]byte{}))
	require.NoError(t, err)

	assert.Equal(t, emptyHash1, emptyHash2)

	// Test with error reader
	errorReader := &ErrorReader{}
	_, err = hashService.CalculateHashFromReader(errorReader)
	assert.Error(t, err)
}

// ErrorReader is a mock reader that always returns an error
type ErrorReader struct{}

func (r *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}