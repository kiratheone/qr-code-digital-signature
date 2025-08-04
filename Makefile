# Digital Signature System - Development Scripts

.PHONY: help up down build logs clean dev-backend dev-frontend test keygen logs-frontend logs-backend logs-postgres prod-up prod-down prod-build prod-logs backup restore rotate-logs health

# Default target
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  up          - Start all services with Docker Compose (development)"
	@echo "  down        - Stop all services"
	@echo "  build       - Build all Docker images"
	@echo "  logs        - View logs from all services"
	@echo "  logs-frontend - View logs from qds-frontend service"
	@echo "  logs-backend  - View logs from qds-backend service"
	@echo "  logs-postgres - View logs from qds-postgres service"
	@echo "  clean       - Clean up containers and volumes"
	@echo "  dev-backend - Run backend in development mode"
	@echo "  dev-frontend- Run frontend in development mode"
	@echo "  test        - Run all tests"
	@echo "  keygen      - Generate RSA key pair for signatures"
	@echo ""
	@echo "Production:"
	@echo "  prod-up     - Start all services in production mode"
	@echo "  prod-down   - Stop production services"
	@echo "  prod-build  - Build production Docker images"
	@echo "  prod-logs   - View production logs"
	@echo "  backup      - Create database backup"
	@echo "  restore     - Restore database from backup"
	@echo "  rotate-logs - Rotate application logs"
	@echo "  health      - Run system health checks"

# Docker commands
up:
	docker-compose up -d

down:
	docker-compose down

build:
	docker-compose build

logs:
	docker-compose logs -f

logs-frontend:
	docker-compose logs -f qds-frontend

logs-backend:
	docker-compose logs -f qds-backend

logs-postgres:
	docker-compose logs -f qds-postgres

clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

# Development commands
dev-backend:
	cd backend && go run cmd/main.go

dev-frontend:
	cd frontend && npm run dev

# Testing
test:
	cd backend && go test ./...
	cd frontend && npm test

# Key generation
keygen:
	cd backend && go run cmd/keygen/main.go

# Production commands
prod-up:
	docker-compose -f docker-compose.prod.yml up -d

prod-down:
	docker-compose -f docker-compose.prod.yml down

prod-build:
	docker-compose -f docker-compose.prod.yml build

prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

# Database backup and restore
backup:
	@./scripts/backup.sh

restore:
	@echo "Available backups:"
	@ls -la backups/*.sql 2>/dev/null || echo "No backups found"
	@echo ""
	@echo "To restore a backup, run:"
	@echo "  docker-compose exec -T qds-postgres psql -U \$$DB_USER \$$DB_NAME < backups/your_backup.sql"
	@echo ""
	@echo "Example:"
	@echo "  docker-compose exec -T qds-postgres psql -U \$$DB_USER \$$DB_NAME < backups/backup_20240101_120000.sql"

# Log management
rotate-logs:
	@./scripts/rotate-logs.sh

# Health checks
health:
	@./scripts/health-check.sh