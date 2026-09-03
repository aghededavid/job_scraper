.PHONY: help dev down build migrate seed test lint clean logs shell-backend shell-db

# Default target
help:
	@echo "RemoteScope — Available Commands"
	@echo "================================="
	@echo "  make dev          Start full local dev stack"
	@echo "  make down         Stop all services"
	@echo "  make build        Build all Docker images"
	@echo "  make migrate      Run database migrations"
	@echo "  make seed         Seed demo data"
	@echo "  make test         Run all tests"
	@echo "  make lint         Lint all code"
	@echo "  make logs         Tail all service logs"
	@echo "  make clean        Remove containers and volumes"
	@echo "  make shell-backend  Open shell in backend container"
	@echo "  make shell-db       Open psql shell"

# Copy env if not present
.env:
	cp .env.example .env
	@echo "✅ Created .env — please fill in your secrets!"

# Start full dev stack
dev: .env
	docker compose up --build -d
	@echo "✅ RemoteScope is running!"
	@echo "   Frontend:  http://localhost:3000"
	@echo "   Backend:   http://localhost:8080"
	@echo "   API Docs:  http://localhost:8080/api/docs"
	@echo "   Grafana:   http://localhost:3001"
	@echo "   pgAdmin:   http://localhost:5050"

# Stop all services
down:
	docker compose down

# Stop and remove volumes (destructive)
clean:
	docker compose down -v --remove-orphans
	@echo "⚠️  All containers and volumes removed."

# Build images without starting
build:
	docker compose build

# Run migrations
migrate:
	docker compose exec backend ./remotescope migrate

# Seed demo data
seed:
	docker compose exec backend ./remotescope seed

# Run all tests
test: test-backend test-scraper test-ai test-frontend

test-backend:
	cd backend && go test ./... -race -count=1 -timeout=120s

test-scraper:
	cd scraper && python -m pytest tests/ -v

test-ai:
	cd ai_service && python -m pytest tests/ -v

test-frontend:
	cd frontend && npm run test -- --passWithNoTests

# Lint all code
lint: lint-backend lint-scraper lint-frontend

lint-backend:
	cd backend && golangci-lint run ./...

lint-scraper:
	cd scraper && ruff check . && mypy .

lint-frontend:
	cd frontend && npm run lint

# Tail logs
logs:
	docker compose logs -f

logs-backend:
	docker compose logs -f backend

logs-scraper:
	docker compose logs -f scraper

logs-ai:
	docker compose logs -f ai_service

# Shell access
shell-backend:
	docker compose exec backend sh

shell-db:
	docker compose exec postgres psql -U remotescope -d remotescope

shell-redis:
	docker compose exec redis redis-cli -a changeme_redis

# Trigger manual scraper run (all sources)
scrape:
	docker compose exec scraper python -m scraper.main --run-now

# Trigger single source
scrape-source:
	docker compose exec scraper python -m scraper.main --source $(SOURCE)

# Generate Go swagger docs
swagger:
	cd backend && swag init -g cmd/api/main.go -o docs/

# Database backup
db-backup:
	docker compose exec postgres pg_dump -U remotescope remotescope > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql

# Apply pgvector extension manually
pgvector:
	docker compose exec postgres psql -U remotescope -d remotescope -c "CREATE EXTENSION IF NOT EXISTS vector;"
