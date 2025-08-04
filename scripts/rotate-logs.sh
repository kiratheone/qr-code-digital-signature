#!/bin/bash

# Digital Signature System - Log Rotation Script
# This script rotates application logs to prevent disk space issues

set -e

# Configuration
LOG_DIR="backend/internal/infrastructure/handlers/logs"
MAX_LOG_SIZE="10M"  # 10 MB
KEEP_LOGS=7         # Keep last 7 rotated logs

echo "Starting log rotation..."
echo "Log directory: ${LOG_DIR}"
echo "Max log size: ${MAX_LOG_SIZE}"
echo "Keep logs: ${KEEP_LOGS}"

# Create log directory if it doesn't exist
mkdir -p "${LOG_DIR}"

# Function to rotate a log file
rotate_log() {
    local log_file="$1"
    local log_name=$(basename "$log_file" .log)
    
    if [ -f "$log_file" ]; then
        # Check file size
        local file_size=$(stat -f%z "$log_file" 2>/dev/null || stat -c%s "$log_file" 2>/dev/null || echo "0")
        local max_size_bytes=10485760  # 10MB in bytes
        
        if [ "$file_size" -gt "$max_size_bytes" ]; then
            echo "Rotating $log_file (size: $(du -h "$log_file" | cut -f1))"
            
            # Create timestamp
            local timestamp=$(date +%Y%m%d_%H%M%S)
            local rotated_file="${log_file}.${timestamp}"
            
            # Move current log to rotated file
            mv "$log_file" "$rotated_file"
            
            # Create new empty log file
            touch "$log_file"
            chmod 666 "$log_file"
            
            echo "Created: $rotated_file"
            
            # Clean up old rotated logs
            cleanup_old_logs "$log_file"
        else
            echo "Skipping $log_file (size: $(du -h "$log_file" | cut -f1))"
        fi
    else
        echo "Log file not found: $log_file"
    fi
}

# Function to clean up old rotated logs
cleanup_old_logs() {
    local base_log_file="$1"
    local log_dir=$(dirname "$base_log_file")
    local log_name=$(basename "$base_log_file" .log)
    
    # Find all rotated logs for this log file
    local rotated_logs=$(find "$log_dir" -name "${log_name}.log.*" -type f | sort -r)
    local count=0
    
    for rotated_log in $rotated_logs; do
        count=$((count + 1))
        if [ $count -gt $KEEP_LOGS ]; then
            echo "Removing old log: $rotated_log"
            rm -f "$rotated_log"
        fi
    done
}

# Rotate each log file
rotate_log "${LOG_DIR}/app.log"
rotate_log "${LOG_DIR}/error.log"
rotate_log "${LOG_DIR}/audit.log"

echo "Log rotation completed!"

# Show current log files
echo ""
echo "Current log files:"
ls -lh "${LOG_DIR}"/*.log 2>/dev/null || echo "No log files found"

echo ""
echo "Rotated log files:"
ls -lh "${LOG_DIR}"/*.log.* 2>/dev/null || echo "No rotated log files found"