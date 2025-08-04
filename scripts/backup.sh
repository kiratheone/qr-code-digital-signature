#!/bin/bash

# Digital Signature System - Database Backup Script
# This script creates a backup of the PostgreSQL database

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Configuration
BACKUP_DIR="backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/backup_${TIMESTAMP}.sql"
CONTAINER_NAME="qr-code-digital-signature-qds-postgres-1"

# Create backup directory if it doesn't exist
mkdir -p "${BACKUP_DIR}"

echo "Starting database backup..."
echo "Timestamp: ${TIMESTAMP}"
echo "Backup file: ${BACKUP_FILE}"

# Check if container is running
if ! docker ps | grep -q "${CONTAINER_NAME}"; then
    echo "Error: PostgreSQL container is not running"
    echo "Start the services with: make up"
    exit 1
fi

# Create backup
echo "Creating backup..."
docker exec "${CONTAINER_NAME}" pg_dump -U "${DB_USER}" "${DB_NAME}" > "${BACKUP_FILE}"

# Verify backup was created
if [ -f "${BACKUP_FILE}" ] && [ -s "${BACKUP_FILE}" ]; then
    echo "Backup created successfully: ${BACKUP_FILE}"
    echo "Backup size: $(du -h "${BACKUP_FILE}" | cut -f1)"
    
    # Keep only last 7 backups
    echo "Cleaning up old backups (keeping last 7)..."
    ls -t "${BACKUP_DIR}"/backup_*.sql | tail -n +8 | xargs -r rm
    
    echo "Backup completed successfully!"
else
    echo "Error: Backup failed or is empty"
    exit 1
fi