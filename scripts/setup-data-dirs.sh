#!/bin/bash

# Script to set up required data directories for Docker volumes

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

# Get data path from environment or use default
DATA_PATH="${DATA_PATH:-./data}"

log_info "Setting up data directories in: $DATA_PATH"

# Create required data directories
DIRECTORIES=(
    "$DATA_PATH/postgres"
    "$DATA_PATH/redis"
    "$DATA_PATH/nginx"
    "$DATA_PATH/prometheus"
    "$DATA_PATH/grafana"
    "$DATA_PATH/loki"
)

for dir in "${DIRECTORIES[@]}"; do
    if [ ! -d "$dir" ]; then
        log_info "Creating directory: $dir"
        mkdir -p "$dir"
        
        # Set appropriate permissions
        case "$(basename "$dir")" in
            "postgres")
                # PostgreSQL needs specific ownership
                chmod 700 "$dir"
                log_info "Set PostgreSQL directory permissions (700)"
                ;;
            "grafana")
                # Grafana needs specific ownership
                chmod 755 "$dir"
                log_info "Set Grafana directory permissions (755)"
                ;;
            *)
                chmod 755 "$dir"
                log_info "Set directory permissions (755)"
                ;;
        esac
        
        log_success "Created: $dir"
    else
        log_warning "Directory already exists: $dir"
        
        # Check and fix permissions if needed
        case "$(basename "$dir")" in
            "postgres")
                current_perms=$(stat -c "%a" "$dir")
                if [ "$current_perms" != "700" ]; then
                    chmod 700 "$dir"
                    log_info "Fixed PostgreSQL directory permissions"
                fi
                ;;
        esac
    fi
done

# Create .gitkeep files to preserve empty directories in git
for dir in "${DIRECTORIES[@]}"; do
    if [ ! -f "$dir/.gitkeep" ]; then
        touch "$dir/.gitkeep"
        log_info "Created .gitkeep in $dir"
    fi
done

# Display directory structure
log_info "Data directory structure:"
if command -v tree >/dev/null 2>&1; then
    tree "$DATA_PATH" -a
else
    find "$DATA_PATH" -type d | sort | sed 's|[^/]*/|  |g'
fi

log_success "Data directories setup completed!"

# Show disk usage
log_info "Disk usage for data directory:"
du -sh "$DATA_PATH" 2>/dev/null || echo "Unable to calculate disk usage"

echo ""
log_info "Next steps:"
echo "  1. Run the deployment: ./scripts/deploy.sh deploy"
echo "  2. Or start services: docker-compose -f docker-compose.prod.yml up -d"