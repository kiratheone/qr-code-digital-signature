#!/bin/bash

# Pre-deployment validation script
# Checks all prerequisites before running deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

ERRORS=0
WARNINGS=0

check_file() {
    local file="$1"
    local description="$2"
    
    if [ -f "$file" ]; then
        log_success "$description: $file"
    else
        log_error "$description missing: $file"
        ((ERRORS++))
    fi
}

check_directory() {
    local dir="$1"
    local description="$2"
    local expected_perms="$3"
    
    if [ -d "$dir" ]; then
        if [ -n "$expected_perms" ]; then
            actual_perms=$(stat -c "%a" "$dir")
            if [ "$actual_perms" = "$expected_perms" ]; then
                log_success "$description: $dir (permissions: $actual_perms)"
            else
                log_warning "$description: $dir (permissions: $actual_perms, expected: $expected_perms)"
                ((WARNINGS++))
            fi
        else
            log_success "$description: $dir"
        fi
    else
        log_error "$description missing: $dir"
        ((ERRORS++))
    fi
}

check_command() {
    local cmd="$1"
    local description="$2"
    
    if command -v "$cmd" >/dev/null 2>&1; then
        local version=$($cmd --version 2>/dev/null | head -1 || echo "unknown")
        log_success "$description: $cmd ($version)"
    else
        log_error "$description missing: $cmd"
        ((ERRORS++))
    fi
}

echo "🔍 Pre-deployment validation check"
echo "=================================="

# Check required commands
log_info "Checking required commands..."
check_command "docker" "Docker"
check_command "docker-compose" "Docker Compose"

# Check configuration files
log_info "Checking configuration files..."
check_file ".env.prod" "Environment file"
check_file "docker-compose.prod.yml" "Docker Compose file"

# Check data directories
log_info "Checking data directories..."
check_directory "data" "Data root directory"
check_directory "data/postgres" "PostgreSQL data directory" "700"
check_directory "data/redis" "Redis data directory" "755"
check_directory "data/nginx" "Nginx data directory" "755"
check_directory "data/prometheus" "Prometheus data directory" "755"
check_directory "data/grafana" "Grafana data directory" "755"
check_directory "data/loki" "Loki data directory" "755"

# Check secret files
log_info "Checking secret files..."
check_directory "secrets" "Secrets directory"
check_file "secrets/db_password.txt" "Database password"
check_file "secrets/jwt_secret.txt" "JWT secret"
check_file "secrets/private_key.pem" "Private key"
check_file "secrets/public_key.pem" "Public key"
check_file "secrets/redis_password.txt" "Redis password"

# Check Docker daemon
log_info "Checking Docker daemon..."
if docker info >/dev/null 2>&1; then
    log_success "Docker daemon is running"
else
    log_error "Docker daemon is not running or not accessible"
    ((ERRORS++))
fi

# Check network conflicts
log_info "Checking network conflicts..."
if bash scripts/check-network-conflicts.sh >/dev/null 2>&1; then
    log_success "No network conflicts detected"
else
    log_warning "Network conflicts may exist (run scripts/check-network-conflicts.sh for details)"
    ((WARNINGS++))
fi

# Check environment variables
log_info "Checking environment variables..."
if bash scripts/test-env-loading.sh >/dev/null 2>&1; then
    log_success "Environment variables can be loaded successfully"
else
    log_error "Environment variable loading failed"
    ((ERRORS++))
fi

# Check disk space
log_info "Checking disk space..."
AVAILABLE_SPACE=$(df . | tail -1 | awk '{print $4}')
REQUIRED_SPACE=1048576  # 1GB in KB

if [ "$AVAILABLE_SPACE" -gt "$REQUIRED_SPACE" ]; then
    log_success "Sufficient disk space available ($(($AVAILABLE_SPACE / 1024))MB free)"
else
    log_warning "Low disk space ($(($AVAILABLE_SPACE / 1024))MB free, recommend at least 1GB)"
    ((WARNINGS++))
fi

# Summary
echo ""
echo "📊 Validation Summary"
echo "===================="

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    log_success "All checks passed! ✅"
    echo ""
    log_info "Ready for deployment. Run: ./scripts/deploy.sh deploy"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    log_warning "$WARNINGS warning(s) found, but deployment should work"
    echo ""
    log_info "You can proceed with deployment: ./scripts/deploy.sh deploy"
    exit 0
else
    log_error "$ERRORS error(s) and $WARNINGS warning(s) found"
    echo ""
    log_error "Please fix the errors before deployment"
    echo ""
    log_info "Common fixes:"
    echo "  - Run: ./scripts/setup-data-dirs.sh"
    echo "  - Run: ./scripts/setup-secrets.sh"
    echo "  - Check: ./docs/DEPLOYMENT_TROUBLESHOOTING.md"
    exit 1
fi