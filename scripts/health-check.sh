#!/bin/bash

# Comprehensive health check script for Digital Signature System
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
TIMEOUT=30
RETRY_COUNT=3
RETRY_DELAY=5

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

# Health check functions
check_service_health() {
    local service=$1
    local url=$2
    local expected_status=${3:-200}
    
    log_debug "Checking health of $service at $url"
    
    for i in $(seq 1 $RETRY_COUNT); do
        if curl -f -s --max-time $TIMEOUT "$url" > /dev/null 2>&1; then
            local status_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT "$url")
            if [ "$status_code" = "$expected_status" ]; then
                log_info "$service health check passed (HTTP $status_code)"
                return 0
            else
                log_warn "$service returned HTTP $status_code, expected $expected_status"
            fi
        else
            log_warn "$service health check failed (attempt $i/$RETRY_COUNT)"
        fi
        
        if [ $i -lt $RETRY_COUNT ]; then
            sleep $RETRY_DELAY
        fi
    done
    
    log_error "$service health check failed after $RETRY_COUNT attempts"
    return 1
}

check_docker_service() {
    local service=$1
    
    log_debug "Checking Docker service status for $service"
    
    local status=$(docker-compose -f "$COMPOSE_FILE" ps -q "$service" 2>/dev/null)
    if [ -z "$status" ]; then
        log_error "$service container is not running"
        return 1
    fi
    
    local health=$(docker inspect --format='{{.State.Health.Status}}' $(docker-compose -f "$COMPOSE_FILE" ps -q "$service") 2>/dev/null || echo "unknown")
    case "$health" in
        "healthy")
            log_info "$service container is healthy"
            return 0
            ;;
        "unhealthy")
            log_error "$service container is unhealthy"
            return 1
            ;;
        "starting")
            log_warn "$service container is still starting"
            return 1
            ;;
        *)
            log_warn "$service container health status is unknown"
            return 1
            ;;
    esac
}

check_database_connection() {
    log_debug "Checking database connection"
    
    local db_check=$(docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U postgres -d digital_signature 2>/dev/null)
    if echo "$db_check" | grep -q "accepting connections"; then
        log_info "Database connection check passed"
        return 0
    else
        log_error "Database connection check failed"
        return 1
    fi
}

check_redis_connection() {
    log_debug "Checking Redis connection"
    
    local redis_check=$(docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli ping 2>/dev/null || echo "FAILED")
    if [ "$redis_check" = "PONG" ]; then
        log_info "Redis connection check passed"
        return 0
    else
        log_error "Redis connection check failed"
        return 1
    fi
}

check_disk_space() {
    log_debug "Checking disk space"
    
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$disk_usage" -lt 90 ]; then
        log_info "Disk space check passed ($disk_usage% used)"
        return 0
    else
        log_error "Disk space check failed ($disk_usage% used, threshold: 90%)"
        return 1
    fi
}

check_memory_usage() {
    log_debug "Checking memory usage"
    
    local mem_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [ "$mem_usage" -lt 90 ]; then
        log_info "Memory usage check passed ($mem_usage% used)"
        return 0
    else
        log_error "Memory usage check failed ($mem_usage% used, threshold: 90%)"
        return 1
    fi
}

check_ssl_certificates() {
    log_debug "Checking SSL certificates"
    
    if [ -f "./nginx/ssl/cert.pem" ]; then
        local cert_expiry=$(openssl x509 -in ./nginx/ssl/cert.pem -noout -enddate 2>/dev/null | cut -d= -f2)
        local cert_expiry_epoch=$(date -d "$cert_expiry" +%s 2>/dev/null || echo 0)
        local current_epoch=$(date +%s)
        local days_until_expiry=$(( (cert_expiry_epoch - current_epoch) / 86400 ))
        
        if [ "$days_until_expiry" -gt 30 ]; then
            log_info "SSL certificate check passed ($days_until_expiry days until expiry)"
            return 0
        elif [ "$days_until_expiry" -gt 0 ]; then
            log_warn "SSL certificate expires in $days_until_expiry days"
            return 0
        else
            log_error "SSL certificate has expired"
            return 1
        fi
    else
        log_warn "SSL certificate not found, skipping check"
        return 0
    fi
}

run_comprehensive_health_check() {
    log_info "Starting comprehensive health check..."
    
    local failed_checks=0
    
    # System resource checks
    check_disk_space || ((failed_checks++))
    check_memory_usage || ((failed_checks++))
    
    # SSL certificate check
    check_ssl_certificates || ((failed_checks++))
    
    # Docker service checks
    local services=("postgres" "redis" "backend" "frontend" "nginx")
    for service in "${services[@]}"; do
        check_docker_service "$service" || ((failed_checks++))
    done
    
    # Database and Redis connectivity
    check_database_connection || ((failed_checks++))
    check_redis_connection || ((failed_checks++))
    
    # HTTP endpoint checks
    check_service_health "Frontend" "http://localhost:3000/api/health" || ((failed_checks++))
    check_service_health "Backend API" "http://localhost:8000/api/health" || ((failed_checks++))
    check_service_health "Nginx Proxy" "http://localhost/health" || ((failed_checks++))
    
    # Application-specific checks
    check_service_health "Document API" "http://localhost:8000/api/documents" 401 || ((failed_checks++))
    check_service_health "Verification API" "http://localhost:8000/api/verify/test" 404 || ((failed_checks++))
    
    # Summary
    if [ "$failed_checks" -eq 0 ]; then
        log_info "All health checks passed successfully!"
        return 0
    else
        log_error "$failed_checks health check(s) failed"
        return 1
    fi
}

show_service_status() {
    log_info "Service Status:"
    docker-compose -f "$COMPOSE_FILE" ps
    
    log_info "Resource Usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
    
    log_info "Service Logs (last 10 lines):"
    local services=("backend" "frontend" "postgres" "redis" "nginx")
    for service in "${services[@]}"; do
        echo -e "\n${BLUE}=== $service ===${NC}"
        docker-compose -f "$COMPOSE_FILE" logs --tail=10 "$service" 2>/dev/null || echo "No logs available"
    done
}

monitor_services() {
    log_info "Starting continuous monitoring (press Ctrl+C to stop)..."
    
    while true; do
        clear
        echo "=== Digital Signature System Health Monitor ==="
        echo "$(date)"
        echo
        
        run_comprehensive_health_check
        
        echo
        show_service_status
        
        sleep 30
    done
}

# Main function
main() {
    case "${1:-check}" in
        "check")
            run_comprehensive_health_check
            ;;
        "status")
            show_service_status
            ;;
        "monitor")
            monitor_services
            ;;
        "quick")
            log_info "Running quick health check..."
            check_service_health "Frontend" "http://localhost:3000/api/health" && \
            check_service_health "Backend" "http://localhost:8000/api/health" && \
            log_info "Quick health check passed"
            ;;
        *)
            echo "Usage: $0 {check|status|monitor|quick}"
            echo "  check   - Run comprehensive health check (default)"
            echo "  status  - Show service status and resource usage"
            echo "  monitor - Continuous monitoring mode"
            echo "  quick   - Quick health check of main services"
            exit 1
            ;;
    esac
}

# Check if Docker and Docker Compose are available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose is not installed or not in PATH"
    exit 1
fi

main "$@"