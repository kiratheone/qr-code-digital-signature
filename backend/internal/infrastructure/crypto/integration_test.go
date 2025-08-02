package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyManagerWithEnvironmentConfig(t *testing.T) {
	// Save original environment
	originalPrivateKey := os.Getenv("PRIVATE_KEY")
	originalPublicKey := os.Getenv("PUBLIC_KEY")
	originalRSAPrivateKey := os.Getenv("RSA_PRIVATE_KEY")
	originalRSAPublicKey := os.Getenv("RSA_PUBLIC_KEY")

	defer func() {
		// Restore original environment
		os.Setenv("PRIVATE_KEY", originalPrivateKey)
		os.Setenv("PUBLIC_KEY", originalPublicKey)
		os.Setenv("RSA_PRIVATE_KEY", originalRSAPrivateKey)
		os.Setenv("RSA_PUBLIC_KEY", originalRSAPublicKey)
	}()

	t.Run("load from PRIVATE_KEY and PUBLIC_KEY environment variables", func(t *testing.T) {
		// Clear RSA_* variables to test fallback
		os.Unsetenv("RSA_PRIVATE_KEY")
		os.Unsetenv("RSA_PUBLIC_KEY")

		// Set PRIVATE_KEY and PUBLIC_KEY (current .env format)
		privateKeyPEM, publicKeyPEM := createTestKeyPairPEM(t)
		os.Setenv("PRIVATE_KEY", privateKeyPEM)
		os.Setenv("PUBLIC_KEY", publicKeyPEM)

		km, err := NewKeyManager()
		require.NoError(t, err)
		assert.NotNil(t, km)
		assert.NoError(t, km.ValidateKeys())
	})

	t.Run("load from RSA_PRIVATE_KEY and RSA_PUBLIC_KEY environment variables", func(t *testing.T) {
		// Clear PRIVATE_KEY and PUBLIC_KEY
		os.Unsetenv("PRIVATE_KEY")
		os.Unsetenv("PUBLIC_KEY")

		// Set RSA_* variables (preferred format)
		privateKeyPEM, publicKeyPEM := createTestKeyPairPEM(t)
		os.Setenv("RSA_PRIVATE_KEY", privateKeyPEM)
		os.Setenv("RSA_PUBLIC_KEY", publicKeyPEM)

		km, err := NewKeyManager()
		require.NoError(t, err)
		assert.NotNil(t, km)
		assert.NoError(t, km.ValidateKeys())
	})

	t.Run("RSA_PRIVATE_KEY takes precedence over PRIVATE_KEY", func(t *testing.T) {
		// Set both sets of variables
		privateKeyPEM1, publicKeyPEM1 := createTestKeyPairPEM(t)
		privateKeyPEM2, publicKeyPEM2 := createTestKeyPairPEM(t)

		os.Setenv("PRIVATE_KEY", privateKeyPEM1)
		os.Setenv("PUBLIC_KEY", publicKeyPEM1)
		os.Setenv("RSA_PRIVATE_KEY", privateKeyPEM2)
		os.Setenv("RSA_PUBLIC_KEY", publicKeyPEM2)

		km, err := NewKeyManager()
		require.NoError(t, err)
		assert.NotNil(t, km)
		assert.NoError(t, km.ValidateKeys())

		// Verify it used the RSA_* keys by creating another key manager with just RSA_* keys
		os.Unsetenv("PRIVATE_KEY")
		os.Unsetenv("PUBLIC_KEY")

		km2, err := NewKeyManager()
		require.NoError(t, err)
		assert.Equal(t, km.GetKeyID(), km2.GetKeyID())
	})
}

func TestKeyManagerSecurityFeatures(t *testing.T) {
	km := createTestKeyManager(t)

	t.Run("key validation prevents mismatched keys", func(t *testing.T) {
		// Create a key manager with mismatched keys
		privateKey1PEM, _ := createTestKeyPairPEM(t)
		_, publicKey2PEM := createTestKeyPairPEM(t)

		invalidKM, err := NewKeyManagerFromEnv(privateKey1PEM, publicKey2PEM)
		assert.Error(t, err)
		assert.Nil(t, invalidKM)
		assert.Contains(t, err.Error(), "public key does not match private key")
	})

	t.Run("key rotation maintains security", func(t *testing.T) {
		// Generate new key pair
		newKeyPair, err := km.GenerateNewKeyPair(2048)
		require.NoError(t, err)

		// Rotate keys
		err = km.RotateKeys(newKeyPair)
		require.NoError(t, err)

		// Verify new keys are valid
		assert.NoError(t, km.ValidateKeys())

		// Verify we can sign and verify with new keys
		testData := []byte("test data for signing")
		hash := km.CalculateDocumentHash(testData)
		
		signature, err := km.SignDocument(hash)
		require.NoError(t, err)

		isValid, err := km.VerifySignature(hash, signature)
		require.NoError(t, err)
		assert.True(t, isValid)
	})

	t.Run("exported keys can be used to create new key manager", func(t *testing.T) {
		// Export current keys
		keyPair, err := km.ExportKeyPairForStorage()
		require.NoError(t, err)

		// Create new key manager from exported keys
		newKM, err := NewKeyManagerFromEnv(keyPair.PrivateKey, keyPair.PublicKey)
		require.NoError(t, err)

		// Verify they have the same key ID
		assert.Equal(t, km.GetKeyID(), newKM.GetKeyID())

		// Verify both can sign and verify the same data
		testData := []byte("test data")
		hash := km.CalculateDocumentHash(testData)

		signature1, err := km.SignDocument(hash)
		require.NoError(t, err)

		signature2, err := newKM.SignDocument(hash)
		require.NoError(t, err)

		// Both signatures should be valid
		isValid1, err := km.VerifySignature(hash, signature1)
		require.NoError(t, err)
		assert.True(t, isValid1)

		isValid2, err := newKM.VerifySignature(hash, signature2)
		require.NoError(t, err)
		assert.True(t, isValid2)

		// Cross-verification should also work
		isValid3, err := km.VerifySignature(hash, signature2)
		require.NoError(t, err)
		assert.True(t, isValid3)

		isValid4, err := newKM.VerifySignature(hash, signature1)
		require.NoError(t, err)
		assert.True(t, isValid4)
	})
}

// Helper methods for KeyManager to make it compatible with SignatureService interface
func (km *KeyManager) CalculateDocumentHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (km *KeyManager) SignDocument(hash []byte) ([]byte, error) {
	signature, err := rsa.SignPSS(rand.Reader, km.privateKey, crypto.SHA256, hash, nil)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (km *KeyManager) VerifySignature(hash []byte, signature []byte) (bool, error) {
	err := rsa.VerifyPSS(km.publicKey, crypto.SHA256, hash, signature, nil)
	return err == nil, nil
}