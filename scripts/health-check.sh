#!/bin/bash

# Digital Signature System - Health Check Script
# This script performs basic health checks on all system components

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL="http://localhost:8000"
FRONTEND_URL="http://localhost:3000"
MAX_RESPONSE_TIME=5 # seconds

echo "Digital Signature System - Health Check"
echo "======================================="
echo ""

# Function to check HTTP endpoint
check_endpoint() {
    local name="$1"
    local url="$2"
    local expected_status="$3"
    
    echo -n "Checking $name... "
    
    # Make request with timeout
    if response=$(curl -s -w "%{http_code}:%{time_total}" --max-time $MAX_RESPONSE_TIME "$url" 2>/dev/null); then
        status_code=$(echo "$response" | cut -d':' -f1)
        response_time=$(echo "$response" | cut -d':' -f2)
        
        if [ "$status_code" = "$expected_status" ]; then
            echo -e "${GREEN}✓ OK${NC} (${response_time}s)"
            return 0
        else
            echo -e "${RED}✗ FAIL${NC} (HTTP $status_code)"
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL${NC} (Connection failed)"
        return 1
    fi
}

# Function to check Docker service
check_docker_service() {
    local service_name="$1"
    
    echo -n "Checking Docker service $service_name... "
    
    if docker-compose ps "$service_name" | grep -q "Up"; then
        echo -e "${GREEN}✓ Running${NC}"
        return 0
    else
        echo -e "${RED}✗ Not running${NC}"
        return 1
    fi
}

# Function to check database connectivity
check_database() {
    echo -n "Checking database connectivity... "
    
    if docker-compose exec -T qds-postgres pg_isready -U "${DB_USER:-postgres}" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Connected${NC}"
        return 0
    else
        echo -e "${RED}✗ Connection failed${NC}"
        return 1
    fi
}

# Function to check disk space
check_disk_space() {
    echo -n "Checking disk space... "
    
    # Get disk usage percentage for root filesystem
    disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$disk_usage" -lt 80 ]; then
        echo -e "${GREEN}✓ OK${NC} (${disk_usage}% used)"
        return 0
    elif [ "$disk_usage" -lt 90 ]; then
        echo -e "${YELLOW}⚠ Warning${NC} (${disk_usage}% used)"
        return 1
    else
        echo -e "${RED}✗ Critical${NC} (${disk_usage}% used)"
        return 1
    fi
}

# Function to check log directory size
check_log_size() {
    echo -n "Checking log directory size... "
    
    log_dir="backend/internal/infrastructure/handlers/logs"
    if [ -d "$log_dir" ]; then
        log_size=$(du -sm "$log_dir" | cut -f1)
        
        if [ "$log_size" -lt 100 ]; then
            echo -e "${GREEN}✓ OK${NC} (${log_size}MB)"
            return 0
        elif [ "$log_size" -lt 500 ]; then
            echo -e "${YELLOW}⚠ Warning${NC} (${log_size}MB)"
            return 1
        else
            echo -e "${RED}✗ Critical${NC} (${log_size}MB)"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠ Log directory not found${NC}"
        return 1
    fi
}

# Load environment variables if available
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Initialize counters
total_checks=0
failed_checks=0

# Perform health checks
echo "System Health Checks:"
echo "--------------------"

# Docker services
for service in qds-frontend qds-backend qds-postgres; do
    total_checks=$((total_checks + 1))
    if ! check_docker_service "$service"; then
        failed_checks=$((failed_checks + 1))
    fi
done

# HTTP endpoints
total_checks=$((total_checks + 1))
if ! check_endpoint "Backend API" "$BACKEND_URL/health" "200"; then
    failed_checks=$((failed_checks + 1))
fi

total_checks=$((total_checks + 1))
if ! check_endpoint "Frontend" "$FRONTEND_URL/api/health" "200"; then
    failed_checks=$((failed_checks + 1))
fi

# Database
total_checks=$((total_checks + 1))
if ! check_database; then
    failed_checks=$((failed_checks + 1))
fi

echo ""
echo "System Resource Checks:"
echo "----------------------"

# System resources
total_checks=$((total_checks + 1))
if ! check_disk_space; then
    failed_checks=$((failed_checks + 1))
fi

total_checks=$((total_checks + 1))
if ! check_log_size; then
    failed_checks=$((failed_checks + 1))
fi

# Summary
echo ""
echo "Health Check Summary:"
echo "===================="

passed_checks=$((total_checks - failed_checks))

if [ $failed_checks -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed${NC} ($passed_checks/$total_checks)"
    exit 0
else
    echo -e "${RED}✗ $failed_checks checks failed${NC} ($passed_checks/$total_checks passed)"
    echo ""
    echo "Recommendations:"
    
    if [ $failed_checks -gt 0 ]; then
        echo "- Check service logs: make logs"
        echo "- Restart services if needed: make down && make up"
        echo "- Review system resources and clean up if necessary"
        echo "- Check OPERATIONS.md for troubleshooting steps"
    fi
    
    exit 1
fi