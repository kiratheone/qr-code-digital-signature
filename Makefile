# Digital Signature System Makefile

.PHONY: help build up down logs clean test

# Default target
help:
	@echo "Available commands:"
	@echo "  build    - Build all Docker images"
	@echo "  up       - Start all services"
	@echo "  down     - Stop all services"
	@echo "  logs     - Show logs from all services"
	@echo "  clean    - Clean up Docker resources"
	@echo "  test     - Run tests"
	@echo "  dev      - Start development environment"

# Build Docker images
build:
	docker-compose build

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Show logs
logs:
	docker-compose logs -f

# Clean up Docker resources
clean:
	docker-compose down -v
	docker system prune -f

# Run tests
test:
	cd backend && go test ./...
	cd frontend && npm test

# Development environment
dev:
	@echo "Starting development environment..."
	@echo "Make sure to copy .env.example to .env and configure it first"
	docker-compose up postgres -d
	@echo "Database started. You can now run backend and frontend locally."