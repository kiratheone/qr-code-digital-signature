#!/bin/bash

# Enhanced script to setup secrets for production deployment
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SECRETS_DIR="./secrets"
BACKUP_DIR="./secrets-backup"
KEY_SIZE=2048
CERT_DAYS=365

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

generate_secure_password() {
    local length=${1:-32}
    openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length
}

generate_jwt_secret() {
    openssl rand -base64 64 | tr -d "\n"
}

generate_api_key() {
    openssl rand -hex 32
}

generate_rsa_keys() {
    local private_key_file="$1"
    local public_key_file="$2"
    
    log_debug "Generating RSA key pair ($KEY_SIZE bits)..."
    
    # Generate private key with password protection option
    openssl genrsa -out "$private_key_file" $KEY_SIZE
    
    # Generate public key
    openssl rsa -in "$private_key_file" -pubout -out "$public_key_file"
    
    # Set proper permissions
    chmod 600 "$private_key_file"
    chmod 644 "$public_key_file"
    
    log_debug "RSA keys generated successfully"
}

generate_ssl_certificate() {
    local cert_file="$1"
    local key_file="$2"
    local domain="${3:-localhost}"
    
    log_debug "Generating self-signed SSL certificate for $domain..."
    
    # Create SSL directory if it doesn't exist
    mkdir -p "$(dirname "$cert_file")"
    
    # Generate private key and certificate
    openssl req -x509 -nodes -days $CERT_DAYS -newkey rsa:2048 \
        -keyout "$key_file" \
        -out "$cert_file" \
        -subj "/C=US/ST=State/L=City/O=Digital Signature System/CN=$domain" \
        -addext "subjectAltName=DNS:$domain,DNS:localhost,IP:127.0.0.1"
    
    # Set proper permissions
    chmod 600 "$key_file"
    chmod 644 "$cert_file"
    
    log_debug "SSL certificate generated successfully"
}

backup_existing_secrets() {
    if [ -d "$SECRETS_DIR" ]; then
        log_info "Backing up existing secrets..."
        local timestamp=$(date +%Y%m%d_%H%M%S)
        local backup_path="${BACKUP_DIR}_${timestamp}"
        
        cp -r "$SECRETS_DIR" "$backup_path"
        log_info "Secrets backed up to $backup_path"
    fi
}

setup_secrets() {
    log_info "Setting up secrets for production deployment..."
    
    # Backup existing secrets
    backup_existing_secrets
    
    # Create secrets directory
    mkdir -p "$SECRETS_DIR"
    chmod 700 "$SECRETS_DIR"
    
    # Generate database password
    if [ ! -f "$SECRETS_DIR/db_password.txt" ]; then
        log_info "Generating database password..."
        generate_secure_password 32 > "$SECRETS_DIR/db_password.txt"
        chmod 600 "$SECRETS_DIR/db_password.txt"
    else
        log_warn "Database password already exists, skipping..."
    fi
    
    # Generate Redis password
    if [ ! -f "$SECRETS_DIR/redis_password.txt" ]; then
        log_info "Generating Redis password..."
        generate_secure_password 32 > "$SECRETS_DIR/redis_password.txt"
        chmod 600 "$SECRETS_DIR/redis_password.txt"
    else
        log_warn "Redis password already exists, skipping..."
    fi
    
    # Generate JWT secret
    if [ ! -f "$SECRETS_DIR/jwt_secret.txt" ]; then
        log_info "Generating JWT secret..."
        generate_jwt_secret > "$SECRETS_DIR/jwt_secret.txt"
        chmod 600 "$SECRETS_DIR/jwt_secret.txt"
    else
        log_warn "JWT secret already exists, skipping..."
    fi
    
    # Generate API key for internal services
    if [ ! -f "$SECRETS_DIR/api_key.txt" ]; then
        log_info "Generating API key..."
        generate_api_key > "$SECRETS_DIR/api_key.txt"
        chmod 600 "$SECRETS_DIR/api_key.txt"
    else
        log_warn "API key already exists, skipping..."
    fi
    
    # Generate RSA key pair for digital signatures
    if [ ! -f "$SECRETS_DIR/private_key.pem" ] || [ ! -f "$SECRETS_DIR/public_key.pem" ]; then
        log_info "Generating RSA key pair for digital signatures..."
        generate_rsa_keys "$SECRETS_DIR/private_key.pem" "$SECRETS_DIR/public_key.pem"
    else
        log_warn "RSA keys already exist, skipping..."
    fi
    
    # Generate SSL certificate for HTTPS
    if [ ! -f "./nginx/ssl/cert.pem" ] || [ ! -f "./nginx/ssl/key.pem" ]; then
        log_info "Generating SSL certificate..."
        mkdir -p "./nginx/ssl"
        generate_ssl_certificate "./nginx/ssl/cert.pem" "./nginx/ssl/key.pem" "localhost"
    else
        log_warn "SSL certificate already exists, skipping..."
    fi
    
    # Generate session secret
    if [ ! -f "$SECRETS_DIR/session_secret.txt" ]; then
        log_info "Generating session secret..."
        generate_secure_password 64 > "$SECRETS_DIR/session_secret.txt"
        chmod 600 "$SECRETS_DIR/session_secret.txt"
    else
        log_warn "Session secret already exists, skipping..."
    fi
    
    # Generate encryption key for sensitive data
    if [ ! -f "$SECRETS_DIR/encryption_key.txt" ]; then
        log_info "Generating encryption key..."
        openssl rand -hex 32 > "$SECRETS_DIR/encryption_key.txt"
        chmod 600 "$SECRETS_DIR/encryption_key.txt"
    else
        log_warn "Encryption key already exists, skipping..."
    fi
    
    log_info "Secrets setup completed!"
    log_info "Secret files created in $SECRETS_DIR/"
    log_warn "Please ensure these files are properly secured and backed up!"
    
    # Show summary
    show_secrets_summary
}

show_secrets_info() {
    log_info "Secrets information:"
    
    if [ -f "$SECRETS_DIR/db_password.txt" ]; then
        echo "Database password: $(head -c 8 "$SECRETS_DIR/db_password.txt")..."
    fi
    
    if [ -f "$SECRETS_DIR/redis_password.txt" ]; then
        echo "Redis password: $(head -c 8 "$SECRETS_DIR/redis_password.txt")..."
    fi
    
    if [ -f "$SECRETS_DIR/jwt_secret.txt" ]; then
        echo "JWT secret: $(head -c 20 "$SECRETS_DIR/jwt_secret.txt")..."
    fi
    
    if [ -f "$SECRETS_DIR/api_key.txt" ]; then
        echo "API key: $(head -c 16 "$SECRETS_DIR/api_key.txt")..."
    fi
    
    if [ -f "$SECRETS_DIR/private_key.pem" ]; then
        echo "Private key: $SECRETS_DIR/private_key.pem"
        openssl rsa -in "$SECRETS_DIR/private_key.pem" -noout -text | head -1
    fi
    
    if [ -f "$SECRETS_DIR/public_key.pem" ]; then
        echo "Public key: $SECRETS_DIR/public_key.pem"
    fi
    
    if [ -f "./nginx/ssl/cert.pem" ]; then
        echo "SSL certificate: ./nginx/ssl/cert.pem"
        local expiry=$(openssl x509 -in ./nginx/ssl/cert.pem -noout -enddate 2>/dev/null | cut -d= -f2)
        echo "SSL certificate expires: $expiry"
    fi
}

show_secrets_summary() {
    log_info "Secrets Summary:"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                    Generated Secrets                        │"
    echo "├─────────────────────────────────────────────────────────────┤"
    
    local secrets=(
        "db_password.txt:Database Password"
        "redis_password.txt:Redis Password"
        "jwt_secret.txt:JWT Secret"
        "api_key.txt:API Key"
        "session_secret.txt:Session Secret"
        "encryption_key.txt:Encryption Key"
        "private_key.pem:RSA Private Key"
        "public_key.pem:RSA Public Key"
    )
    
    for secret in "${secrets[@]}"; do
        local file="${secret%%:*}"
        local desc="${secret##*:}"
        if [ -f "$SECRETS_DIR/$file" ]; then
            printf "│ %-25s │ %-30s │\n" "$desc" "✓ Generated"
        else
            printf "│ %-25s │ %-30s │\n" "$desc" "✗ Missing"
        fi
    done
    
    echo "├─────────────────────────────────────────────────────────────┤"
    if [ -f "./nginx/ssl/cert.pem" ]; then
        printf "│ %-25s │ %-30s │\n" "SSL Certificate" "✓ Generated"
    else
        printf "│ %-25s │ %-30s │\n" "SSL Certificate" "✗ Missing"
    fi
    echo "└─────────────────────────────────────────────────────────────┘"
}

validate_secrets() {
    log_info "Validating secrets..."
    
    local errors=0
    
    # Check required secret files
    local required_secrets=(
        "db_password.txt"
        "redis_password.txt"
        "jwt_secret.txt"
        "api_key.txt"
        "session_secret.txt"
        "encryption_key.txt"
        "private_key.pem"
        "public_key.pem"
    )
    
    for secret in "${required_secrets[@]}"; do
        if [ ! -f "$SECRETS_DIR/$secret" ]; then
            log_error "Missing required secret: $secret"
            ((errors++))
        else
            # Check if file is not empty
            if [ ! -s "$SECRETS_DIR/$secret" ]; then
                log_error "Secret file is empty: $secret"
                ((errors++))
            fi
        fi
    done
    
    # Validate RSA private key
    if [ -f "$SECRETS_DIR/private_key.pem" ]; then
        if ! openssl rsa -in "$SECRETS_DIR/private_key.pem" -check -noout >/dev/null 2>&1; then
            log_error "Invalid RSA private key"
            ((errors++))
        fi
    fi
    
    # Validate RSA public key
    if [ -f "$SECRETS_DIR/public_key.pem" ]; then
        if ! openssl rsa -in "$SECRETS_DIR/public_key.pem" -pubin -noout >/dev/null 2>&1; then
            log_error "Invalid RSA public key"
            ((errors++))
        fi
    fi
    
    # Validate that RSA key pair matches (if both exist)
    if [ -f "$SECRETS_DIR/private_key.pem" ] && [ -f "$SECRETS_DIR/public_key.pem" ]; then
        local private_modulus=$(openssl rsa -in "$SECRETS_DIR/private_key.pem" -noout -modulus 2>/dev/null)
        local public_modulus=$(openssl rsa -in "$SECRETS_DIR/public_key.pem" -pubin -noout -modulus 2>/dev/null)
        
        if [ "$private_modulus" != "$public_modulus" ]; then
            log_error "RSA key pair mismatch - private and public keys don't match"
            ((errors++))
        fi
    fi
    
    # Validate SSL certificate
    if [ -f "./nginx/ssl/cert.pem" ]; then
        if ! openssl x509 -in ./nginx/ssl/cert.pem -noout -text >/dev/null 2>&1; then
            log_error "Invalid SSL certificate"
            ((errors++))
        else
            # Check if certificate is expired
            if ! openssl x509 -in ./nginx/ssl/cert.pem -noout -checkend 0 >/dev/null 2>&1; then
                log_warn "SSL certificate has expired"
            fi
        fi
    fi
    
    # Validate SSL private key
    if [ -f "./nginx/ssl/key.pem" ]; then
        if ! openssl rsa -in ./nginx/ssl/key.pem -check -noout >/dev/null 2>&1; then
            log_error "Invalid SSL private key"
            ((errors++))
        fi
    fi
    
    # Validate that SSL certificate and key match (if both exist)
    if [ -f "./nginx/ssl/cert.pem" ] && [ -f "./nginx/ssl/key.pem" ]; then
        local cert_modulus=$(openssl x509 -in ./nginx/ssl/cert.pem -noout -modulus 2>/dev/null)
        local key_modulus=$(openssl rsa -in ./nginx/ssl/key.pem -noout -modulus 2>/dev/null)
        
        if [ "$cert_modulus" != "$key_modulus" ]; then
            log_error "SSL certificate and key mismatch"
            ((errors++))
        fi
    fi
    
    if [ $errors -eq 0 ]; then
        log_info "All secrets are valid"
        return 0
    else
        log_error "$errors validation error(s) found"
        return 1
    fi
}

rotate_secrets() {
    log_warn "This will regenerate all secrets and may break existing deployments!"
    read -p "Are you sure you want to rotate all secrets? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Rotating secrets..."
        
        # Backup current secrets
        backup_existing_secrets
        
        # Remove existing secrets
        rm -f "$SECRETS_DIR"/*.txt "$SECRETS_DIR"/*.pem
        rm -f "./nginx/ssl/cert.pem" "./nginx/ssl/key.pem"
        
        # Generate new secrets
        setup_secrets
        
        log_info "Secret rotation completed"
    else
        log_info "Secret rotation cancelled"
    fi
}

# Main function
main() {
    case "${1:-setup}" in
        "setup")
            setup_secrets
            ;;
        "info")
            show_secrets_info
            ;;
        "summary")
            show_secrets_summary
            ;;
        "validate")
            validate_secrets
            ;;
        "rotate")
            rotate_secrets
            ;;
        "clean")
            log_warn "This will delete all secret files!"
            read -p "Are you sure? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                rm -rf "$SECRETS_DIR" "./nginx/ssl"
                log_info "Secrets cleaned up"
            else
                log_info "Operation cancelled"
            fi
            ;;
        *)
            echo "Usage: $0 {setup|info|summary|validate|rotate|clean}"
            echo "  setup    - Generate and setup secrets (default)"
            echo "  info     - Show secrets information"
            echo "  summary  - Show secrets summary table"
            echo "  validate - Validate existing secrets"
            echo "  rotate   - Rotate all secrets (regenerate)"
            echo "  clean    - Remove all secret files"
            exit 1
            ;;
    esac
}

# Check requirements
if ! command -v openssl &> /dev/null; then
    log_error "OpenSSL is required but not installed"
    exit 1
fi

main "$@"