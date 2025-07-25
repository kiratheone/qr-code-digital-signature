#!/bin/bash

# Comprehensive maintenance script for Digital Signature System
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
LOG_FILE="${LOG_FILE:-/var/log/digital-signature-maintenance.log}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
MAINTENANCE_MODE="${MAINTENANCE_MODE:-false}"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [INFO] $1" >> "$LOG_FILE" 2>/dev/null || true
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [WARN] $1" >> "$LOG_FILE" 2>/dev/null || true
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [ERROR] $1" >> "$LOG_FILE" 2>/dev/null || true
}

log_success() {
    echo -e "${CYAN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [SUCCESS] $1" >> "$LOG_FILE" 2>/dev/null || true
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [DEBUG] $1" >> "$LOG_FILE" 2>/dev/null || true
}

# Function to enable maintenance mode
enable_maintenance_mode() {
    log_info "Enabling maintenance mode..."
    
    # Create maintenance page
    cat > nginx/maintenance.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>System Maintenance</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-bottom: 20px; }
        p { color: #666; line-height: 1.6; }
        .icon { font-size: 48px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🔧</div>
        <h1>System Maintenance</h1>
        <p>The Digital Signature System is currently undergoing scheduled maintenance.</p>
        <p>We apologize for any inconvenience. The system will be back online shortly.</p>
        <p>Estimated completion time: <span id="eta">Unknown</span></p>
    </div>
    <script>
        // Update ETA if provided
        const urlParams = new URLSearchParams(window.location.search);
        const eta = urlParams.get('eta');
        if (eta) {
            document.getElementById('eta').textContent = eta;
        }
    </script>
</body>
</html>
EOF
    
    # Update nginx configuration to serve maintenance page
    if [ -f "nginx/nginx.conf" ]; then
        cp nginx/nginx.conf nginx/nginx.conf.backup
        
        # Add maintenance mode configuration
        sed -i '/location \/ {/i\
        # Maintenance mode\
        if (-f /etc/nginx/maintenance.html) {\
            return 503;\
        }\
        \
        error_page 503 @maintenance;\
        location @maintenance {\
            root /etc/nginx;\
            rewrite ^(.*)$ /maintenance.html break;\
        }' nginx/nginx.conf
        
        # Restart nginx
        docker-compose -f "$COMPOSE_FILE" restart nginx
        
        log_success "Maintenance mode enabled"
    else
        log_warn "Nginx configuration not found, maintenance mode not enabled"
    fi
}

# Function to disable maintenance mode
disable_maintenance_mode() {
    log_info "Disabling maintenance mode..."
    
    # Restore original nginx configuration
    if [ -f "nginx/nginx.conf.backup" ]; then
        mv nginx/nginx.conf.backup nginx/nginx.conf
        docker-compose -f "$COMPOSE_FILE" restart nginx
        log_success "Maintenance mode disabled"
    else
        log_warn "Backup nginx configuration not found"
    fi
    
    # Remove maintenance page
    rm -f nginx/maintenance.html
}

# Function to clean up Docker resources
cleanup_docker() {
    log_info "Cleaning up Docker resources..."
    
    # Remove unused containers
    local removed_containers=$(docker container prune -f 2>/dev/null | grep "Total reclaimed space" | awk '{print $4, $5}' || echo "0B")
    log_info "Removed unused containers: $removed_containers"
    
    # Remove unused images
    local removed_images=$(docker image prune -f 2>/dev/null | grep "Total reclaimed space" | awk '{print $4, $5}' || echo "0B")
    log_info "Removed unused images: $removed_images"
    
    # Remove unused networks
    local removed_networks=$(docker network prune -f 2>/dev/null | grep "Total reclaimed space" | awk '{print $4, $5}' || echo "0B")
    log_info "Removed unused networks: $removed_networks"
    
    # Remove unused volumes (be careful with this)
    if [ "$AGGRESSIVE_CLEANUP" = "true" ]; then
        log_warn "Performing aggressive cleanup - removing unused volumes"
        local removed_volumes=$(docker volume prune -f 2>/dev/null | grep "Total reclaimed space" | awk '{print $4, $5}' || echo "0B")
        log_info "Removed unused volumes: $removed_volumes"
    fi
    
    log_success "Docker cleanup completed"
}

# Function to optimize database
optimize_database() {
    log_info "Optimizing database..."
    
    # Run VACUUM and ANALYZE on PostgreSQL
    if docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
        log_info "Running database maintenance commands..."
        
        # VACUUM
        docker-compose -f "$COMPOSE_FILE" exec -T postgres psql -U postgres -d digital_signature -c "VACUUM VERBOSE;" 2>/dev/null || log_warn "VACUUM failed"
        
        # ANALYZE
        docker-compose -f "$COMPOSE_FILE" exec -T postgres psql -U postgres -d digital_signature -c "ANALYZE VERBOSE;" 2>/dev/null || log_warn "ANALYZE failed"
        
        # REINDEX
        docker-compose -f "$COMPOSE_FILE" exec -T postgres psql -U postgres -d digital_signature -c "REINDEX DATABASE digital_signature;" 2>/dev/null || log_warn "REINDEX failed"
        
        # Update statistics
        docker-compose -f "$COMPOSE_FILE" exec -T postgres psql -U postgres -d digital_signature -c "SELECT pg_stat_reset();" 2>/dev/null || log_warn "Statistics reset failed"
        
        log_success "Database optimization completed"
    else
        log_error "Database is not accessible"
    fi
}

# Function to rotate logs
rotate_logs() {
    log_info "Rotating application logs..."
    
    # Rotate Docker logs
    if command -v logrotate >/dev/null 2>&1; then
        # Create logrotate configuration
        cat > /tmp/docker-logrotate.conf << EOF
/var/lib/docker/containers/*/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
    postrotate
        docker kill -s USR1 \$(docker ps -q) 2>/dev/null || true
    endscript
}
EOF
        
        logrotate -f /tmp/docker-logrotate.conf 2>/dev/null || log_warn "Log rotation failed"
        rm -f /tmp/docker-logrotate.conf
    fi
    
    # Rotate application logs
    if [ -d "./logs" ]; then
        find ./logs -name "*.log" -type f -mtime +7 -exec gzip {} \;
        find ./logs -name "*.log.gz" -type f -mtime +30 -delete
        log_info "Application logs rotated"
    fi
    
    log_success "Log rotation completed"
}

# Function to update system packages
update_system() {
    log_info "Updating system packages..."
    
    if command -v apt-get >/dev/null 2>&1; then
        # Ubuntu/Debian
        apt-get update >/dev/null 2>&1 || log_warn "Package list update failed"
        
        # Check for security updates
        local security_updates=$(apt list --upgradable 2>/dev/null | grep -i security | wc -l)
        if [ "$security_updates" -gt 0 ]; then
            log_warn "$security_updates security updates available"
            
            if [ "$AUTO_UPDATE" = "true" ]; then
                log_info "Installing security updates..."
                apt-get upgrade -y >/dev/null 2>&1 || log_warn "Security updates failed"
            fi
        fi
        
    elif command -v yum >/dev/null 2>&1; then
        # CentOS/RHEL
        yum check-update >/dev/null 2>&1 || true
        
        if [ "$AUTO_UPDATE" = "true" ]; then
            yum update -y >/dev/null 2>&1 || log_warn "System update failed"
        fi
    fi
    
    log_success "System update check completed"
}

# Function to check and update Docker images
update_docker_images() {
    log_info "Checking for Docker image updates..."
    
    # Pull latest images
    if docker-compose -f "$COMPOSE_FILE" pull 2>/dev/null; then
        log_info "Docker images updated"
        
        # Check if restart is needed
        local outdated_services=$(docker-compose -f "$COMPOSE_FILE" ps --services | while read service; do
            local running_image=$(docker-compose -f "$COMPOSE_FILE" ps -q "$service" | xargs docker inspect --format='{{.Image}}' 2>/dev/null)
            local latest_image=$(docker-compose -f "$COMPOSE_FILE" config | grep -A 5 "$service:" | grep "image:" | awk '{print $2}' | xargs docker images --format='{{.ID}}' 2>/dev/null | head -1)
            
            if [ "$running_image" != "$latest_image" ] && [ -n "$latest_image" ]; then
                echo "$service"
            fi
        done)
        
        if [ -n "$outdated_services" ]; then
            log_warn "Services with outdated images: $outdated_services"
            log_info "Consider restarting these services to use updated images"
        fi
    else
        log_warn "Failed to pull Docker images"
    fi
}

# Function to check SSL certificate expiry
check_ssl_expiry() {
    log_info "Checking SSL certificate expiry..."
    
    local cert_file="./nginx/ssl/cert.pem"
    
    if [ -f "$cert_file" ]; then
        if command -v openssl >/dev/null 2>&1; then
            local expiry_date=$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | cut -d= -f2)
            local expiry_timestamp=$(date -d "$expiry_date" +%s 2>/dev/null || echo "0")
            local current_timestamp=$(date +%s)
            local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
            
            if [ "$days_until_expiry" -lt 30 ]; then
                log_warn "SSL certificate expires in $days_until_expiry days - renewal recommended"
            elif [ "$days_until_expiry" -lt 0 ]; then
                log_error "SSL certificate has expired!"
            else
                log_info "SSL certificate expires in $days_until_expiry days"
            fi
        fi
    else
        log_warn "SSL certificate file not found"
    fi
}

# Function to backup configuration
backup_configuration() {
    log_info "Backing up configuration files..."
    
    local config_backup_dir="$BACKUP_DIR/config_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$config_backup_dir"
    
    # Backup important configuration files
    local config_files=(
        ".env.prod"
        "docker-compose.prod.yml"
        "nginx/nginx.conf"
        "monitoring/prometheus.yml"
        "monitoring/grafana"
        "secrets"
    )
    
    for file in "${config_files[@]}"; do
        if [ -e "$file" ]; then
            cp -r "$file" "$config_backup_dir/" 2>/dev/null || log_warn "Failed to backup $file"
        fi
    done
    
    # Create archive
    tar -czf "${config_backup_dir}.tar.gz" -C "$BACKUP_DIR" "$(basename "$config_backup_dir")" 2>/dev/null
    rm -rf "$config_backup_dir"
    
    log_success "Configuration backup created: ${config_backup_dir}.tar.gz"
}

# Function to generate maintenance report
generate_maintenance_report() {
    log_info "Generating maintenance report..."
    
    local report_file="$BACKUP_DIR/maintenance_report_$(date +%Y%m%d_%H%M%S).txt"
    mkdir -p "$BACKUP_DIR"
    
    {
        echo "=== Digital Signature System Maintenance Report ==="
        echo "Generated: $(date)"
        echo "Hostname: $(hostname)"
        echo ""
        
        echo "=== System Information ==="
        uname -a
        echo ""
        
        echo "=== Disk Usage ==="
        df -h
        echo ""
        
        echo "=== Memory Usage ==="
        free -h
        echo ""
        
        echo "=== Docker System Information ==="
        docker system df
        echo ""
        
        echo "=== Service Status ==="
        docker-compose -f "$COMPOSE_FILE" ps
        echo ""
        
        echo "=== Container Resource Usage ==="
        docker stats --no-stream
        echo ""
        
        echo "=== Recent Logs (Last 100 lines) ==="
        docker-compose -f "$COMPOSE_FILE" logs --tail=100
        
    } > "$report_file"
    
    log_success "Maintenance report generated: $report_file"
}

# Function to perform health check
health_check() {
    log_info "Performing post-maintenance health check..."
    
    local failed_checks=0
    
    # Check services
    local services=("frontend" "backend" "postgres" "redis" "nginx")
    for service in "${services[@]}"; do
        if docker-compose -f "$COMPOSE_FILE" ps "$service" | grep -q "Up"; then
            log_success "$service: Running"
        else
            log_error "$service: Not running"
            ((failed_checks++))
        fi
    done
    
    # Check health endpoints
    sleep 10  # Wait for services to be ready
    
    if curl -f -s --max-time 10 "http://localhost/api/health" >/dev/null 2>&1; then
        log_success "Backend health check: OK"
    else
        log_error "Backend health check: FAILED"
        ((failed_checks++))
    fi
    
    if curl -f -s --max-time 10 "http://localhost:3000/api/health" >/dev/null 2>&1; then
        log_success "Frontend health check: OK"
    else
        log_error "Frontend health check: FAILED"
        ((failed_checks++))
    fi
    
    if [ "$failed_checks" -eq 0 ]; then
        log_success "All health checks passed"
        return 0
    else
        log_error "$failed_checks health checks failed"
        return 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo
    echo "Commands:"
    echo "  full        - Run full maintenance (default)"
    echo "  docker      - Clean up Docker resources only"
    echo "  database    - Optimize database only"
    echo "  logs        - Rotate logs only"
    echo "  system      - Update system packages only"
    echo "  images      - Update Docker images only"
    echo "  ssl         - Check SSL certificate only"
    echo "  backup      - Backup configuration only"
    echo "  report      - Generate maintenance report only"
    echo "  health      - Run health check only"
    echo
    echo "Options:"
    echo "  -m, --maintenance-mode    Enable maintenance mode during operation"
    echo "  -a, --aggressive          Perform aggressive cleanup (removes unused volumes)"
    echo "  -u, --auto-update         Automatically install system updates"
    echo "  -f, --compose-file FILE   Docker compose file (default: docker-compose.prod.yml)"
    echo "  -l, --log-file FILE       Log file path"
    echo "  -b, --backup-dir DIR      Backup directory"
    echo "  -h, --help               Show this help message"
    echo
    echo "Examples:"
    echo "  $0                       Run full maintenance"
    echo "  $0 -m full              Run full maintenance with maintenance mode"
    echo "  $0 docker               Clean up Docker resources only"
    echo "  $0 -a docker            Aggressive Docker cleanup (includes volumes)"
}

# Parse command line arguments
COMMAND="full"
AGGRESSIVE_CLEANUP=false
AUTO_UPDATE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--maintenance-mode)
            MAINTENANCE_MODE=true
            shift
            ;;
        -a|--aggressive)
            AGGRESSIVE_CLEANUP=true
            shift
            ;;
        -u|--auto-update)
            AUTO_UPDATE=true
            shift
            ;;
        -f|--compose-file)
            COMPOSE_FILE="$2"
            shift 2
            ;;
        -l|--log-file)
            LOG_FILE="$2"
            shift 2
            ;;
        -b|--backup-dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        full|docker|database|logs|system|images|ssl|backup|report|health)
            COMMAND="$1"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main maintenance function
main() {
    log_info "Starting maintenance operation: $COMMAND"
    
    # Create log directory
    mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null || true
    
    # Enable maintenance mode if requested
    if [ "$MAINTENANCE_MODE" = "true" ] && [ "$COMMAND" = "full" ]; then
        enable_maintenance_mode
    fi
    
    case "$COMMAND" in
        "full")
            log_info "Running full maintenance..."
            cleanup_docker
            optimize_database
            rotate_logs
            update_system
            update_docker_images
            check_ssl_expiry
            backup_configuration
            generate_maintenance_report
            health_check
            ;;
        "docker")
            cleanup_docker
            ;;
        "database")
            optimize_database
            ;;
        "logs")
            rotate_logs
            ;;
        "system")
            update_system
            ;;
        "images")
            update_docker_images
            ;;
        "ssl")
            check_ssl_expiry
            ;;
        "backup")
            backup_configuration
            ;;
        "report")
            generate_maintenance_report
            ;;
        "health")
            health_check
            ;;
        *)
            log_error "Unknown command: $COMMAND"
            show_usage
            exit 1
            ;;
    esac
    
    # Disable maintenance mode if it was enabled
    if [ "$MAINTENANCE_MODE" = "true" ] && [ "$COMMAND" = "full" ]; then
        disable_maintenance_mode
    fi
    
    log_success "Maintenance operation completed: $COMMAND"
}

# Trap signals for graceful shutdown
trap 'log_info "Maintenance interrupted by user"; disable_maintenance_mode 2>/dev/null || true; exit 1' INT TERM

main