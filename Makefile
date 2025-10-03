# Local Development Commands
.PHONY: build run format test clean

build:
	go build -o bin/web-api cmd/main.go

run:
	go run cmd/main.go

format:
	go fmt web-api/...

test:
	go test -v ./...

clean:
	rm -rf bin/

# Docker Commands
.PHONY: docker-build docker-up docker-down docker-logs docker-exec docker-restart

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f api

docker-exec:
	docker-compose exec api sh

docker-restart:
	docker-compose restart api

# Development with Hot Reload
.PHONY: dev-up dev-down dev-logs

dev-up:
	docker-compose -f docker-compose.dev.yml up

dev-down:
	docker-compose -f docker-compose.dev.yml down

dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f api

# Production Commands
.PHONY: prod-build prod-up prod-down prod-logs

prod-build:
	docker-compose -f docker-compose.yml build --no-cache

prod-up:
	docker-compose -f docker-compose.yml up -d

prod-down:
	docker-compose -f docker-compose.yml down

prod-logs:
	docker-compose -f docker-compose.yml logs -f api

# Database Commands
.PHONY: db-connect db-backup db-restore

db-connect:
	docker-compose exec postgres psql -U erp_user -d erp_database

db-backup:
	docker-compose exec postgres pg_dump -U erp_user erp_database > backup_$$(date +%Y%m%d_%H%M%S).sql

db-restore:
	@echo "Usage: make db-restore FILE=backup.sql"
	@if [ -z "$(FILE)" ]; then echo "Error: FILE parameter required"; exit 1; fi
	docker-compose exec -T postgres psql -U erp_user erp_database < $(FILE)

# Cleanup Commands
.PHONY: clean-docker clean-volumes clean-all

clean-docker:
	docker-compose down --remove-orphans

clean-volumes:
	docker-compose down -v

clean-all: clean clean-docker clean-volumes
	docker system prune -f

# Help
.PHONY: help

help:
	@echo "ERP API Makefile Commands:"
	@echo ""
	@echo "Local Development:"
	@echo "  make build          - Build the Go application"
	@echo "  make run            - Run the application locally"
	@echo "  make format         - Format Go code"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Docker (Production):"
	@echo "  make docker-build   - Build Docker images"
	@echo "  make docker-up      - Start all services"
	@echo "  make docker-down    - Stop all services"
	@echo "  make docker-logs    - View API logs"
	@echo "  make docker-exec    - Access API container shell"
	@echo "  make docker-restart - Restart API service"
	@echo ""
	@echo "Docker (Development):"
	@echo "  make dev-up         - Start dev environment with hot reload"
	@echo "  make dev-down       - Stop dev environment"
	@echo "  make dev-logs       - View dev API logs"
	@echo ""
	@echo "Database:"
	@echo "  make db-connect     - Connect to PostgreSQL"
	@echo "  make db-backup      - Backup database"
	@echo "  make db-restore     - Restore database (requires FILE=path)"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean-docker   - Remove Docker containers"
	@echo "  make clean-volumes  - Remove Docker volumes (deletes data!)"
	@echo "  make clean-all      - Clean everything"