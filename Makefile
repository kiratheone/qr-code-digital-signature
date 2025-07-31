# Digital Signature System - Development Scripts

.PHONY: help up down build logs clean dev-backend dev-frontend test keygen logs-frontend logs-backend logs-postgres

# Default target
help:
	@echo "Available commands:"
	@echo "  up          - Start all services with Docker Compose"
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