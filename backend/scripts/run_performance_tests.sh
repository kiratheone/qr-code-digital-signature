#!/bin/bash

# Performance Testing Script for Digital Signature System

set -e

echo "🚀 Starting Performance Tests for Digital Signature System"
echo "=========================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL=${API_URL:-"http://localhost:8000"}
DB_URL=${DB_URL:-"postgres://postgres:password@localhost:5432/digital_signature_test?sslmode=disable"}

echo -e "${YELLOW}Configuration:${NC}"
echo "API URL: $API_URL"
echo "DB URL: $DB_URL"
echo ""

# Function to check if service is running
check_service() {
    local url=$1
    local service_name=$2
    
    echo -n "Checking $service_name... "
    if curl -s "$url/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Running${NC}"
        return 0
    else
        echo -e "${RED}✗ Not running${NC}"
        return 1
    fi
}

# Function to run Go benchmarks
run_go_benchmarks() {
    echo -e "${YELLOW}Running Go Benchmarks...${NC}"
    echo "----------------------------------------"
    
    cd backend
    
    # Run all benchmark tests
    echo "Running database connection pool benchmarks..."
    go test -bench=BenchmarkDatabaseConnectionPool -benchmem ./internal/infrastructure/services/
    
    echo ""
    echo "Running PDF hash calculation benchmarks..."
    go test -bench=BenchmarkPDFHashCalculation -benchmem ./internal/infrastructure/services/
    
    echo ""
    echo "Running cache service benchmarks..."
    go test -bench=BenchmarkCacheService -benchmem ./internal/infrastructure/services/
    
    echo ""
    echo "Running signature service benchmarks..."
    go test -bench=BenchmarkSignatureService -benchmem ./internal/infrastructure/services/
    
    cd ..
}

# Function to run load tests
run_load_tests() {
    echo -e "${YELLOW}Running Load Tests...${NC}"
    echo "----------------------------------------"
    
    cd backend
    
    # Build and run load test
    echo "Building load test..."
    go build -o test_load test_load.go
    
    echo "Running load test with 10 concurrent users, 100 requests..."
    API_URL=$API_URL ./test_load
    
    # Cleanup
    rm -f test_load
    
    cd ..
}

# Function to run stress tests
run_stress_tests() {
    echo -e "${YELLOW}Running Stress Tests...${NC}"
    echo "----------------------------------------"
    
    cd backend
    
    echo "Running database connection pool stress test..."
    go test -run TestDatabaseConnectionPoolStress ./internal/infrastructure/services/
    
    echo ""
    echo "Running cache service concurrency test..."
    go test -run TestCacheServiceConcurrency ./internal/infrastructure/services/
    
    cd ..
}

# Function to run frontend performance tests
run_frontend_tests() {
    echo -e "${YELLOW}Running Frontend Performance Tests...${NC}"
    echo "----------------------------------------"
    
    cd frontend
    
    # Check if dependencies are installed
    if [ ! -d "node_modules" ]; then
        echo "Installing frontend dependencies..."
        npm install
    fi
    
    # Run build performance test
    echo "Testing build performance..."
    time npm run build
    
    # Analyze bundle size
    echo ""
    echo "Analyzing bundle size..."
    if command -v du &> /dev/null; then
        echo "Build output size:"
        du -sh .next/
        echo ""
        echo "Static assets size:"
        du -sh .next/static/
    fi
    
    cd ..
}

# Function to generate performance report
generate_report() {
    echo -e "${YELLOW}Generating Performance Report...${NC}"
    echo "----------------------------------------"
    
    local report_file="performance_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "Digital Signature System - Performance Test Report"
        echo "=================================================="
        echo "Date: $(date)"
        echo "API URL: $API_URL"
        echo "DB URL: $DB_URL"
        echo ""
        echo "System Information:"
        echo "OS: $(uname -s)"
        echo "Architecture: $(uname -m)"
        echo "CPU Cores: $(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 'Unknown')"
        echo "Memory: $(free -h 2>/dev/null | grep Mem | awk '{print $2}' || echo 'Unknown')"
        echo ""
        echo "Go Version: $(go version)"
        echo "Node Version: $(node --version 2>/dev/null || echo 'Not installed')"
        echo ""
    } > "$report_file"
    
    echo "Performance report saved to: $report_file"
}

# Main execution
main() {
    echo "Starting performance test suite..."
    echo ""
    
    # Check if services are running (optional for some tests)
    check_service "$API_URL" "API Server" || echo -e "${YELLOW}Warning: API server not running, some tests may fail${NC}"
    
    echo ""
    
    # Run tests
    run_go_benchmarks
    echo ""
    
    if check_service "$API_URL" "API Server" > /dev/null 2>&1; then
        run_load_tests
        echo ""
    else
        echo -e "${YELLOW}Skipping load tests - API server not running${NC}"
        echo ""
    fi
    
    run_stress_tests
    echo ""
    
    run_frontend_tests
    echo ""
    
    generate_report
    
    echo -e "${GREEN}✅ Performance testing completed!${NC}"
}

# Handle script arguments
case "${1:-all}" in
    "benchmarks"|"bench")
        run_go_benchmarks
        ;;
    "load")
        run_load_tests
        ;;
    "stress")
        run_stress_tests
        ;;
    "frontend")
        run_frontend_tests
        ;;
    "all"|*)
        main
        ;;
esac