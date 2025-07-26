#!/bin/bash

# CI Simulation Script
# This script simulates the GitHub Actions CI pipeline locally

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GO_VERSION="1.22"
NODE_VERSION="20"
POSTGRES_VERSION="16"

# Database configuration
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="password"
DB_NAME_TEST="digital_signature_test"
DB_NAME="digital_signature"
DB_SSLMODE="disable"

# Test environment variables
export PRIVATE_KEY="test-private-key"
export PUBLIC_KEY="test-public-key"
export JWT_SECRET="test-jwt-secret"
export PORT="8000"
export NEXT_PUBLIC_API_URL="http://localhost:8000"
export API_URL="http://localhost:8000"

print_step() {
    echo -e "${BLUE}==== $1 ====${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

check_dependencies() {
    print_step "Checking Dependencies"
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go $GO_VERSION"
        exit 1
    fi
    print_success "Go found: $(go version)"
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install Node.js $NODE_VERSION"
        exit 1
    fi
    print_success "Node.js found: $(node --version)"
    
    # Check npm
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed"
        exit 1
    fi
    print_success "npm found: $(npm --version)"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_warning "Docker is not installed. Some tests may fail."
    else
        print_success "Docker found: $(docker --version)"
    fi
    
    # Check PostgreSQL
    if ! command -v psql &> /dev/null; then
        print_warning "PostgreSQL client not found. Will try to use Docker."
    else
        print_success "PostgreSQL client found"
    fi
}

setup_postgres() {
    print_step "Setting up PostgreSQL"
    
    # Check if PostgreSQL is running locally
    if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER 2>/dev/null; then
        print_success "PostgreSQL is already running"
    else
        print_warning "PostgreSQL not running locally, starting with Docker..."
        
        # Stop any existing postgres container
        docker stop postgres-test 2>/dev/null || true
        docker rm postgres-test 2>/dev/null || true
        
        # Start PostgreSQL container
        docker run -d \
            --name postgres-test \
            -e POSTGRES_PASSWORD=$DB_PASSWORD \
            -e POSTGRES_USER=$DB_USER \
            -e POSTGRES_DB=$DB_NAME_TEST \
            -p $DB_PORT:5432 \
            postgres:$POSTGRES_VERSION
        
        # Wait for PostgreSQL to be ready
        echo "Waiting for PostgreSQL to be ready..."
        for i in {1..30}; do
            if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER 2>/dev/null; then
                break
            fi
            sleep 1
        done
        
        if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER 2>/dev/null; then
            print_error "PostgreSQL failed to start"
            exit 1
        fi
        
        print_success "PostgreSQL started successfully"
    fi
    
    # Create test database if it doesn't exist
    PGPASSWORD=$DB_PASSWORD createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME_TEST 2>/dev/null || true
    PGPASSWORD=$DB_PASSWORD createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME 2>/dev/null || true
}

run_backend_tests() {
    print_step "Running Backend Tests"
    
    cd ../backend
    
    # Set environment variables
    export DB_HOST=$DB_HOST
    export DB_PORT=$DB_PORT
    export DB_USER=$DB_USER
    export DB_PASSWORD=$DB_PASSWORD
    export DB_NAME=$DB_NAME_TEST
    export DB_SSLMODE=$DB_SSLMODE
    
    # Download dependencies
    print_step "Installing Go dependencies"
    go mod download
    
    # Run unit tests
    print_step "Running Go unit tests"
    if go test -v -race -coverprofile=coverage.out ./...; then
        print_success "Backend unit tests passed"
        go tool cover -html=coverage.out -o coverage.html
        print_success "Coverage report generated: backend/coverage.html"
    else
        print_error "Backend unit tests failed"
        cd ..
        return 1
    fi
    
    # Run benchmarks
    print_step "Running Go benchmarks"
    if go test -bench=. -benchmem ./internal/infrastructure/services/ 2>/dev/null; then
        print_success "Backend benchmarks completed"
    else
        print_warning "Backend benchmarks failed or no benchmarks found"
    fi
    
    cd ../scripts
}

run_frontend_tests() {
    print_step "Running Frontend Tests"
    
    cd ../frontend
    
    # Install dependencies
    print_step "Installing npm dependencies"
    if npm ci; then
        print_success "Frontend dependencies installed"
    else
        print_error "Failed to install frontend dependencies"
        cd ..
        return 1
    fi
    
    # Run linting
    print_step "Running ESLint"
    if npm run lint; then
        print_success "Frontend linting passed"
    else
        print_error "Frontend linting failed"
        cd ..
        return 1
    fi
    
    # Run type checking
    print_step "Running TypeScript type checking"
    if npm run type-check; then
        print_success "Frontend type checking passed"
    else
        print_error "Frontend type checking failed"
        cd ..
        return 1
    fi
    
    # Run unit tests
    print_step "Running Jest unit tests"
    if npm run test -- --coverage --watchAll=false; then
        print_success "Frontend unit tests passed"
        print_success "Coverage report generated: frontend/coverage/"
    else
        print_error "Frontend unit tests failed"
        cd ..
        return 1
    fi
    
    # Build application
    print_step "Building Next.js application"
    if npm run build; then
        print_success "Frontend build successful"
    else
        print_error "Frontend build failed"
        cd ../scripts
        return 1
    fi
    
    cd ../scripts
}

run_integration_tests() {
    print_step "Running Integration Tests"
    
    cd ../backend
    
    # Set environment variables
    export DB_HOST=$DB_HOST
    export DB_PORT=$DB_PORT
    export DB_USER=$DB_USER
    export DB_PASSWORD=$DB_PASSWORD
    export DB_NAME=$DB_NAME_TEST
    export DB_SSLMODE=$DB_SSLMODE
    
    # Build backend
    if go build -o main .; then
        print_success "Backend build successful"
    else
        print_error "Backend build failed"
        cd ../scripts
        return 1
    fi
    
    # Run integration tests
    if [ -f "./internal/infrastructure/server/integration_test.go" ]; then
        if go test -v ./internal/infrastructure/server/integration_test.go; then
            print_success "Integration tests passed"
        else
            print_error "Integration tests failed"
            cd ../scripts
            return 1
        fi
    else
        print_warning "Integration test file not found"
    fi
    
    cd ../scripts
}

run_e2e_tests() {
    print_step "Running End-to-End Tests"
    
    # Set up database for E2E tests
    export DB_NAME=$DB_NAME
    PGPASSWORD=$DB_PASSWORD createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME 2>/dev/null || true
    
    cd ../backend
    
    # Build and start backend
    print_step "Starting backend server"
    export DB_HOST=$DB_HOST
    export DB_PORT=$DB_PORT
    export DB_USER=$DB_USER
    export DB_PASSWORD=$DB_PASSWORD
    export DB_NAME=$DB_NAME
    export DB_SSLMODE=$DB_SSLMODE
    export PORT=$PORT
    
    if go build -o main .; then
        ./main &
        BACKEND_PID=$!
        sleep 10
        print_success "Backend server started (PID: $BACKEND_PID)"
    else
        print_error "Failed to build backend"
        cd ../scripts
        return 1
    fi
    
    cd ../frontend
    
    # Build and start frontend
    print_step "Starting frontend server"
    export NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
    
    if npm run build; then
        npm start &
        FRONTEND_PID=$!
        sleep 10
        print_success "Frontend server started (PID: $FRONTEND_PID)"
    else
        print_error "Failed to start frontend"
        kill $BACKEND_PID 2>/dev/null || true
        cd ../scripts
        return 1
    fi
    
    cd ../backend
    
    # Run E2E tests
    if [ -f "test_e2e.go" ]; then
        export API_URL=$API_URL
        if go build -o test_e2e test_e2e.go && ./test_e2e; then
            print_success "E2E tests passed"
        else
            print_error "E2E tests failed"
            kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
            cd ../scripts
            return 1
        fi
    else
        print_warning "E2E test file not found"
    fi
    
    # Cleanup
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    cd ../scripts
}

run_docker_tests() {
    print_step "Running Docker Build Tests"
    
    # Build backend Docker image
    print_step "Building backend Docker image"
    if docker build -t digital-signature-backend:test ../backend; then
        print_success "Backend Docker image built successfully"
    else
        print_error "Backend Docker build failed"
        return 1
    fi
    
    # Build frontend Docker image
    print_step "Building frontend Docker image"
    if docker build -t digital-signature-frontend:test ../frontend; then
        print_success "Frontend Docker image built successfully"
    else
        print_error "Frontend Docker build failed"
        return 1
    fi
    
    # Test Docker Compose configuration
    print_step "Testing Docker Compose configuration"
    if docker-compose -f ../docker-compose.yml config > /dev/null; then
        print_success "Docker Compose configuration is valid"
    else
        print_error "Docker Compose configuration is invalid"
        return 1
    fi
}

cleanup() {
    print_step "Cleaning up"
    
    # Kill any running processes
    pkill -f "digital-signature" 2>/dev/null || true
    pkill -f "next" 2>/dev/null || true
    
    # Stop and remove PostgreSQL container if we started it
    docker stop postgres-test 2>/dev/null || true
    docker rm postgres-test 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Main execution
main() {
    echo -e "${BLUE}🚀 Starting CI Pipeline Simulation${NC}"
    echo "This script will simulate your GitHub Actions CI pipeline locally"
    echo ""
    
    # Parse command line arguments
    SKIP_DEPS=false
    SKIP_POSTGRES=false
    ONLY_BACKEND=false
    ONLY_FRONTEND=false
    SKIP_E2E=false
    SKIP_DOCKER=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --skip-postgres)
                SKIP_POSTGRES=true
                shift
                ;;
            --only-backend)
                ONLY_BACKEND=true
                shift
                ;;
            --only-frontend)
                ONLY_FRONTEND=true
                shift
                ;;
            --skip-e2e)
                SKIP_E2E=true
                shift
                ;;
            --skip-docker)
                SKIP_DOCKER=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --skip-deps      Skip dependency checks"
                echo "  --skip-postgres  Skip PostgreSQL setup"
                echo "  --only-backend   Run only backend tests"
                echo "  --only-frontend  Run only frontend tests"
                echo "  --skip-e2e       Skip end-to-end tests"
                echo "  --skip-docker    Skip Docker build tests"
                echo "  --help           Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Set up trap for cleanup
    trap cleanup EXIT
    
    # Run checks and tests
    if [ "$SKIP_DEPS" = false ]; then
        check_dependencies
    fi
    
    if [ "$SKIP_POSTGRES" = false ]; then
        setup_postgres
    fi
    
    # Run tests based on options
    if [ "$ONLY_FRONTEND" = false ]; then
        run_backend_tests || exit 1
    fi
    
    if [ "$ONLY_BACKEND" = false ]; then
        run_frontend_tests || exit 1
    fi
    
    if [ "$ONLY_FRONTEND" = false ] && [ "$ONLY_BACKEND" = false ]; then
        run_integration_tests || exit 1
        
        if [ "$SKIP_E2E" = false ]; then
            run_e2e_tests || exit 1
        fi
        
        if [ "$SKIP_DOCKER" = false ] && command -v docker &> /dev/null; then
            run_docker_tests || exit 1
        fi
    fi
    
    echo ""
    print_success "🎉 All tests completed successfully!"
    echo -e "${GREEN}Your CI pipeline simulation has finished without errors.${NC}"
}

# Run main function
main "$@"