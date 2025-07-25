#!/bin/bash

# Production monitoring script for Digital Signature System
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
LOG_DIR="./logs"
ALERT_EMAIL=""  # Set this to receive email alerts

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

check_service_health() {
    local service=$1
    local url=$2
    
    if curl -f -s "$url" > /dev/null 2>&1; then
        log_info "$service is healthy"
        return 0
    else
        log_error "$service is unhealthy"
        return 1
    fi
}

check_container_status() {
    local service=$1
    
    local status=$(docker-compose -f "$COMPOSE_FILE" ps -q "$service" | xargs docker inspect --format='{{.State.Status}}' 2>/dev/null)
    
    if [ "$status" = "running" ]; then
        log_info "$service container is running"
        return 0
    else
        log_error "$service container is not running (status: $status)"
        return 1
    fi
}

check_resource_usage() {
    log_info "Checking resource usage..."
    
    # Get container stats
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}"
    
    # Check disk usage
    echo ""
    log_info "Disk usage:"
    df -h | grep -E "(Filesystem|/dev/)"
    
    # Check Docker volume usage
    echo ""
    log_info "Docker volume usage:"
    docker system df
}

check_logs_for_errors() {
    log_info "Checking logs for errors..."
    
    # Check for recent errors in logs
    local error_count=$(docker-compose -f "$COMPOSE_FILE" logs --since="1h" 2>&1 | grep -i "error\|exception\|fatal" | wc -l)
    
    if [ "$error_count" -gt 0 ]; then
        log_warn "Found $error_count error entries in the last hour"
        docker-compose -f "$COMPOSE_FILE" logs --since="1h" 2>&1 | grep -i "error\|exception\|fatal" | tail -10
    else
        log_info "No recent errors found in logs"
    fi
}

check_database_connections() {
    log_info "Checking database connections..."
    
    local db_connections=$(docker-compose -f "$COMPOSE_FILE" exec -T postgres psql -U postgres -d digital_signature -c "SELECT count(*) FROM pg_stat_activity;" 2>/dev/null | grep -E "^\s*[0-9]+\s*$" | tr -d ' ')
    
    if [ -n "$db_connections" ]; then
        log_info "Database has $db_connections active connections"
        
        if [ "$db_connections" -gt 50 ]; then
            log_warn "High number of database connections: $db_connections"
        fi
    else
        log_error "Could not check database connections"
    fi
}

check_ssl_certificate() {
    log_info "Checking SSL certificate..."
    
    local cert_file="./nginx/ssl/cert.pem"
    
    if [ -f "$cert_file" ]; then
        local expiry_date=$(openssl x509 -in "$cert_file" -noout -enddate | cut -d= -f2)
        local expiry_timestamp=$(date -d "$expiry_date" +%s)
        local current_timestamp=$(date +%s)
        local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
        
        if [ "$days_until_expiry" -lt 30 ]; then
            log_warn "SSL certificate expires in $days_until_expiry days"
        else
            log_info "SSL certificate is valid for $days_until_expiry more days"
        fi
    else
        log_warn "SSL certificate file not found"
    fi
}

run_comprehensive_check() {
    log_info "Running comprehensive health check..."
    
    local failed_checks=0
    
    # Check container status
    services=("frontend" "backend" "postgres" "redis" "nginx")
    for service in "${services[@]}"; do
        if ! check_container_status "$service"; then
            ((failed_checks++))
        fi
    done
    
    # Check service health endpoints
    if ! check_service_health "Frontend" "http://localhost/health"; then
        ((failed_checks++))
    fi
    
    if ! check_service_health "Backend API" "http://localhost/api/health"; then
        ((failed_checks++))
    fi
    
    # Check database
    check_database_connections
    
    # Check resources
    check_resource_usage
    
    # Check logs
    check_logs_for_errors
    
    # Check SSL
    check_ssl_certificate
    
    # Summary
    echo ""
    if [ "$failed_checks" -eq 0 ]; then
        log_info "All health checks passed ✅"
    else
        log_error "$failed_checks health checks failed ❌"
        return 1
    fi
}

generate_report() {
    local report_file="$LOG_DIR/health-report-$(date +%Y%m%d-%H%M%S).txt"
    
    mkdir -p "$LOG_DIR"
    
    {
        echo "=== Digital Signature System Health Report ==="
        echo "Generated: $(date)"
        echo ""
        
        echo "=== Service Status ==="
        docker-compose -f "$COMPOSE_FILE" ps
        echo ""
        
        echo "=== Resource Usage ==="
        docker stats --no-stream
        echo ""
        
        echo "=== Recent Logs ==="
        docker-compose -f "$COMPOSE_FILE" logs --since="1h" --tail=50
        
    } > "$report_file"
    
    log_info "Health report generated: $report_file"
}

continuous_monitoring() {
    log_info "Starting continuous monitoring (Ctrl+C to stop)..."
    
    while true; do
        echo ""
        echo "=== $(date) ==="
        
        if ! run_comprehensive_check; then
            log_error "Health check failed - generating report"
            generate_report
            
            # Send alert if email is configured
            if [ -n "$ALERT_EMAIL" ]; then
                echo "Health check failed at $(date)" | mail -s "Digital Signature System Alert" "$ALERT_EMAIL"
            fi
        fi
        
        sleep 300  # Check every 5 minutes
    done
}

# Main function
main() {
    case "${1:-check}" in
        "check")
            run_comprehensive_check
            ;;
        "monitor")
            continuous_monitoring
            ;;
        "report")
            generate_report
            ;;
        "resources")
            check_resource_usage
            ;;
        "logs")
            check_logs_for_errors
            ;;
        "ssl")
            check_ssl_certificate
            ;;
        *)
            echo "Usage: $0 {check|monitor|report|resources|logs|ssl}"
            echo "  check     - Run comprehensive health check (default)"
            echo "  monitor   - Start continuous monitoring"
            echo "  report    - Generate detailed health report"
            echo "  resources - Check resource usage"
            echo "  logs      - Check logs for errors"
            echo "  ssl       - Check SSL certificate status"
            exit 1
            ;;
    esac
}

main "$@"