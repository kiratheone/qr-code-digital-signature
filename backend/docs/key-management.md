# Key Management System

The Digital Signature System includes a comprehensive key management system that provides secure storage, validation, and rotation of RSA keys used for digital signatures.

## Features

### 1. Secure Key Storage
- **Environment Variables**: Keys can be stored as base64-encoded environment variables
- **File-based Storage**: Keys can be loaded from PEM files
- **Multiple Formats**: Supports both PKCS#8 and PKCS#1 private key formats
- **Automatic Validation**: Keys are validated during loading to ensure they match

### 2. Key Loading Priority
The system loads keys in the following priority order:
1. `RSA_PRIVATE_KEY` and `RSA_PUBLIC_KEY` environment variables (preferred)
2. `PRIVATE_KEY` and `PUBLIC_KEY` environment variables (backward compatibility)
3. File-based keys from `PRIVATE_KEY_PATH` and `PUBLIC_KEY_PATH`
4. Default file paths: `private_key.pem` and `public_key.pem`

### 3. Key Validation
- **Automatic Validation**: Keys are validated during KeyManager creation
- **Pair Matching**: Ensures private and public keys match
- **Cryptographic Testing**: Performs actual sign/verify test to validate functionality
- **Error Reporting**: Provides detailed error messages for validation failures

### 4. Key Rotation
- **Generate New Keys**: Create new RSA key pairs with configurable key sizes
- **Safe Rotation**: Validate new keys before replacing current ones
- **Export Functionality**: Export keys in PEM format for secure storage
- **Rollback Safety**: Failed rotations don't affect current keys

## Usage

### Environment Configuration

#### Option 1: Base64 Encoded (Recommended for Docker)
```bash
# Generate base64 encoded keys
PRIVATE_KEY=$(base64 -w 0 private_key.pem)
PUBLIC_KEY=$(base64 -w 0 public_key.pem)
```

#### Option 2: Raw PEM Format
```bash
# Use raw PEM content
PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----"

PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsNGmFu...
-----END PUBLIC KEY-----"
```

### Code Usage

#### Basic Key Manager Creation
```go
import "digital-signature-system/internal/infrastructure/crypto"

// Create key manager (loads from environment or files)
km, err := crypto.NewKeyManager()
if err != nil {
    log.Fatalf("Failed to create key manager: %v", err)
}

// Validate keys
err = km.ValidateKeys()
if err != nil {
    log.Fatalf("Key validation failed: %v", err)
}
```

#### Key Information
```go
// Get key information
keyID := km.GetKeyID()
createdAt := km.GetCreatedAt()
privateKey := km.GetPrivateKey()
publicKey := km.GetPublicKey()

fmt.Printf("Key ID: %s\n", keyID)
fmt.Printf("Loaded at: %s\n", createdAt)
```

#### Key Rotation
```go
// Generate new key pair
newKeyPair, err := km.GenerateNewKeyPair(2048)
if err != nil {
    log.Fatalf("Failed to generate new key pair: %v", err)
}

// Rotate to new keys
err = km.RotateKeys(newKeyPair)
if err != nil {
    log.Fatalf("Failed to rotate keys: %v", err)
}

// Export new keys for storage
exportedKeyPair, err := km.ExportKeyPairForStorage()
if err != nil {
    log.Fatalf("Failed to export keys: %v", err)
}

// Save to environment variables or files
fmt.Printf("New Private Key:\n%s\n", exportedKeyPair.PrivateKey)
fmt.Printf("New Public Key:\n%s\n", exportedKeyPair.PublicKey)
```

### Command Line Utilities

The system includes a key rotation utility for operational tasks:

#### Show Key Information
```bash
go run cmd/keyrotate/main.go info
```

#### Validate Current Keys
```bash
go run cmd/keyrotate/main.go validate
```

#### Generate New Key Pair
```bash
go run cmd/keyrotate/main.go generate
```

#### Rotate Keys
```bash
go run cmd/keyrotate/main.go rotate
```

## Security Features

### 1. Key Validation
- **Pair Matching**: Ensures private and public keys are mathematically related
- **Cryptographic Testing**: Performs actual sign/verify operations to validate functionality
- **Format Validation**: Validates PEM format and key structure

### 2. Secure Storage
- **Environment Variables**: Keys stored in environment variables are not logged
- **Base64 Encoding**: Supports base64 encoding for safe environment variable storage
- **File Permissions**: Recommends proper file permissions for key files (600 for private keys)

### 3. Key Rotation Safety
- **Validation Before Rotation**: New keys are validated before replacing current ones
- **Atomic Operations**: Key rotation is atomic - either succeeds completely or fails safely
- **Rollback Protection**: Failed rotations don't affect current working keys

### 4. Error Handling
- **Detailed Error Messages**: Provides specific error information for troubleshooting
- **Graceful Degradation**: Falls back to alternative key sources if primary source fails
- **Validation Errors**: Clear indication of what validation checks failed

## Best Practices

### 1. Key Storage
- Use environment variables for production deployments
- Store private keys with restricted permissions (600)
- Use base64 encoding for keys containing special characters
- Never commit keys to version control

### 2. Key Rotation
- Rotate keys regularly (recommended: every 6-12 months)
- Test new keys in staging environment before production
- Keep backup of previous keys during rotation period
- Update all dependent systems after key rotation

### 3. Monitoring
- Monitor key validation status
- Log key rotation events
- Alert on key validation failures
- Track key age and rotation schedule

### 4. Security
- Use minimum 2048-bit RSA keys (4096-bit for high security)
- Validate keys after any configuration changes
- Implement proper access controls for key management utilities
- Audit key access and rotation events

## Error Handling

### Common Errors and Solutions

#### "Key validation failed: public key does not match private key"
- **Cause**: Private and public keys don't form a valid pair
- **Solution**: Ensure both keys are from the same key pair generation

#### "Failed to decode PEM block containing private key"
- **Cause**: Invalid PEM format or corrupted key data
- **Solution**: Verify key format and regenerate if necessary

#### "Failed to decode base64 private key"
- **Cause**: Invalid base64 encoding
- **Solution**: Re-encode the key using proper base64 encoding

#### "Key size must be at least 2048 bits"
- **Cause**: Attempting to generate keys smaller than minimum security requirement
- **Solution**: Use 2048-bit or larger keys

## Integration with Signature Service

The KeyManager integrates seamlessly with the SignatureService:

```go
// Create key manager
km, err := crypto.NewKeyManager()
if err != nil {
    return err
}

// Create signature service using the key manager
sigService, err := crypto.NewSignatureServiceFromKeyManager(km)
if err != nil {
    return err
}

// Use signature service for document signing
hash := sigService.CalculateDocumentHash(documentData)
signature, err := sigService.SignDocument(hash)
if err != nil {
    return err
}
```

This integration ensures that key management and cryptographic operations use the same validated key pair, maintaining security and consistency throughout the system.