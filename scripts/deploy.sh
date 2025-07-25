#!/bin/bash

# Production deployment script for Digital Signature System
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env.prod"
SECRETS_DIR="./secrets"

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

check_requirements() {
    log_info "Checking deployment requirements..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    # Check if environment file exists
    if [ ! -f "$ENV_FILE" ]; then
        log_error "Environment file $ENV_FILE not found"
        log_info "Please create $ENV_FILE with required environment variables"
        exit 1
    fi
    
    # Check if secrets directory exists
    if [ ! -d "$SECRETS_DIR" ]; then
        log_error "Secrets directory $SECRETS_DIR not found"
        log_info "Please create secrets directory with required secret files"
        exit 1
    fi
    
    log_info "Requirements check passed"
}

setup_secrets() {
    log_info "Setting up secrets..."
    
    # Create secrets directory if it doesn't exist
    mkdir -p "$SECRETS_DIR"
    
    # Check required secret files
    REQUIRED_SECRETS=("db_password.txt" "jwt_secret.txt" "private_key.pem" "public_key.pem")
    
    for secret in "${REQUIRED_SECRETS[@]}"; do
        if [ ! -f "$SECRETS_DIR/$secret" ]; then
            log_error "Secret file $SECRETS_DIR/$secret not found"
            exit 1
        fi
    done
    
    # Set proper permissions for secret files
    chmod 600 "$SECRETS_DIR"/*
    
    log_info "Secrets setup completed"
}

build_images() {
    log_info "Building Docker images..."
    
    # Build images with no cache for production
    docker-compose -f "$COMPOSE_FILE" build --no-cache --parallel
    
    log_info "Docker images built successfully"
}

deploy() {
    log_info "Deploying application..."
    
    # Load environment variables
    export $(cat "$ENV_FILE" | xargs)
    
    # Stop existing containers
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans
    
    # Start services
    docker-compose -f "$COMPOSE_FILE" up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 30
    
    # Check service health
    if docker-compose -f "$COMPOSE_FILE" ps | grep -q "unhealthy\|Exit"; then
        log_error "Some services are not healthy"
        docker-compose -f "$COMPOSE_FILE" logs
        exit 1
    fi
    
    log_info "Application deployed successfully"
}

run_health_checks() {
    log_info "Running health checks..."
    
    # Check if services are responding
    if curl -f http://localhost/health > /dev/null 2>&1; then
        log_info "Frontend health check passed"
    else
        log_error "Frontend health check failed"
        exit 1
    fi
    
    if curl -f http://localhost/api/health > /dev/null 2>&1; then
        log_info "Backend health check passed"
    else
        log_error "Backend health check failed"
        exit 1
    fi
    
    log_info "All health checks passed"
}

cleanup() {
    log_info "Cleaning up unused Docker resources..."
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes (be careful with this in production)
    # docker volume prune -f
    
    log_info "Cleanup completed"
}

show_status() {
    log_info "Application status:"
    docker-compose -f "$COMPOSE_FILE" ps
    
    log_info "Resource usage:"
    docker stats --no-stream
}

# Main deployment process
main() {
    log_info "Starting production deployment..."
    
    check_requirements
    setup_secrets
    build_images
    deploy
    run_health_checks
    cleanup
    show_status
    
    log_info "Deployment completed successfully!"
    log_info "Application is available at: https://localhost"
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "status")
        show_status
        ;;
    "logs")
        docker-compose -f "$COMPOSE_FILE" logs -f "${2:-}"
        ;;
    "stop")
        log_info "Stopping application..."
        docker-compose -f "$COMPOSE_FILE" down
        ;;
    "restart")
        log_info "Restarting application..."
        docker-compose -f "$COMPOSE_FILE" restart "${2:-}"
        ;;
    "health")
        run_health_checks
        ;;
    *)
        echo "Usage: $0 {deploy|status|logs|stop|restart|health}"
        echo "  deploy  - Deploy the application (default)"
        echo "  status  - Show application status"
        echo "  logs    - Show application logs"
        echo "  stop    - Stop the application"
        echo "  restart - Restart the application"
        echo "  health  - Run health checks"
        exit 1
        ;;
esac