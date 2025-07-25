#!/bin/bash

# PostgreSQL backup script for production
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DB_NAME="${DB_NAME:-digital_signature}"
DB_USER="${DB_USER:-postgres}"
COMPOSE_FILE="docker-compose.prod.yml"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/postgres_backup_$TIMESTAMP.sql"
COMPRESSED_BACKUP="$BACKUP_FILE.gz"

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

create_backup() {
    log_info "Creating PostgreSQL backup..."
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    
    # Create database backup
    if docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_dump -U "$DB_USER" -d "$DB_NAME" --verbose --clean --no-owner --no-privileges > "$BACKUP_FILE"; then
        log_info "Database backup created: $BACKUP_FILE"
        
        # Compress backup
        if gzip "$BACKUP_FILE"; then
            log_info "Backup compressed: $COMPRESSED_BACKUP"
            
            # Get backup size
            local backup_size=$(du -h "$COMPRESSED_BACKUP" | cut -f1)
            log_info "Backup size: $backup_size"
            
            return 0
        else
            log_error "Failed to compress backup"
            return 1
        fi
    else
        log_error "Failed to create database backup"
        return 1
    fi
}

verify_backup() {
    log_info "Verifying backup integrity..."
    
    if [ -f "$COMPRESSED_BACKUP" ]; then
        # Test gzip integrity
        if gzip -t "$COMPRESSED_BACKUP"; then
            log_info "Backup file integrity verified"
            
            # Check if backup contains expected content
            if zcat "$COMPRESSED_BACKUP" | head -20 | grep -q "PostgreSQL database dump"; then
                log_info "Backup content verification passed"
                return 0
            else
                log_error "Backup content verification failed"
                return 1
            fi
        else
            log_error "Backup file is corrupted"
            return 1
        fi
    else
        log_error "Backup file not found"
        return 1
    fi
}

cleanup_old_backups() {
    log_info "Cleaning up old backups (retention: $RETENTION_DAYS days)..."
    
    local deleted_count=0
    
    # Find and delete old backups
    find "$BACKUP_DIR" -name "postgres_backup_*.sql.gz" -type f -mtime +$RETENTION_DAYS -print0 | while IFS= read -r -d '' file; do
        log_info "Deleting old backup: $(basename "$file")"
        rm "$file"
        ((deleted_count++))
    done
    
    if [ $deleted_count -gt 0 ]; then
        log_info "Deleted $deleted_count old backup(s)"
    else
        log_info "No old backups to delete"
    fi
}

restore_backup() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "Backup file not specified"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        return 1
    fi
    
    log_warn "This will restore the database from backup and may overwrite existing data!"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Restoring database from backup: $backup_file"
        
        # Stop application services to prevent connections
        log_info "Stopping application services..."
        docker-compose -f "$COMPOSE_FILE" stop backend frontend
        
        # Restore database
        if zcat "$backup_file" | docker-compose -f "$COMPOSE_FILE" exec -T postgres psql -U "$DB_USER" -d "$DB_NAME"; then
            log_info "Database restored successfully"
            
            # Restart services
            log_info "Restarting application services..."
            docker-compose -f "$COMPOSE_FILE" start backend frontend
            
            return 0
        else
            log_error "Database restore failed"
            
            # Restart services anyway
            docker-compose -f "$COMPOSE_FILE" start backend frontend
            return 1
        fi
    else
        log_info "Restore cancelled"
        return 1
    fi
}

list_backups() {
    log_info "Available backups:"
    
    if [ -d "$BACKUP_DIR" ]; then
        local backups=($(find "$BACKUP_DIR" -name "postgres_backup_*.sql.gz" -type f | sort -r))
        
        if [ ${#backups[@]} -eq 0 ]; then
            log_warn "No backups found"
            return 1
        fi
        
        echo "┌─────────────────────────────────────────────────────────────┐"
        echo "│                    Available Backups                       │"
        echo "├─────────────────────────────────────────────────────────────┤"
        
        for backup in "${backups[@]}"; do
            local filename=$(basename "$backup")
            local size=$(du -h "$backup" | cut -f1)
            local date=$(stat -c %y "$backup" | cut -d' ' -f1,2 | cut -d'.' -f1)
            
            printf "│ %-35s │ %-8s │ %-15s │\n" "$filename" "$size" "$date"
        done
        
        echo "└─────────────────────────────────────────────────────────────┘"
    else
        log_warn "Backup directory does not exist: $BACKUP_DIR"
        return 1
    fi
}

show_backup_info() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "Backup file not specified"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        return 1
    fi
    
    log_info "Backup Information:"
    echo "File: $backup_file"
    echo "Size: $(du -h "$backup_file" | cut -f1)"
    echo "Created: $(stat -c %y "$backup_file" | cut -d'.' -f1)"
    echo "MD5: $(md5sum "$backup_file" | cut -d' ' -f1)"
    
    # Show backup content summary
    log_info "Backup Content Summary:"
    zcat "$backup_file" | head -50 | grep -E "(CREATE TABLE|INSERT INTO)" | head -10
}

# Main function
main() {
    case "${1:-backup}" in
        "backup")
            create_backup && verify_backup && cleanup_old_backups
            ;;
        "restore")
            restore_backup "$2"
            ;;
        "list")
            list_backups
            ;;
        "info")
            show_backup_info "$2"
            ;;
        "cleanup")
            cleanup_old_backups
            ;;
        "verify")
            if [ -n "$2" ]; then
                COMPRESSED_BACKUP="$2"
                verify_backup
            else
                log_error "Please specify backup file to verify"
                exit 1
            fi
            ;;
        *)
            echo "Usage: $0 {backup|restore|list|info|cleanup|verify}"
            echo "  backup           - Create new backup (default)"
            echo "  restore <file>   - Restore from backup file"
            echo "  list             - List available backups"
            echo "  info <file>      - Show backup information"
            echo "  cleanup          - Remove old backups"
            echo "  verify <file>    - Verify backup integrity"
            exit 1
            ;;
    esac
}

# Check requirements
if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose is not installed"
    exit 1
fi

if ! command -v gzip &> /dev/null; then
    log_error "gzip is not installed"
    exit 1
fi

main "$@"