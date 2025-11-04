# ───────────────────────────────────────────────
# Variables
# ───────────────────────────────────────────────
DOCKER_COMPOSE = docker compose
APP_SERVICE = app

# ───────────────────────────────────────────────
# Commands
# ───────────────────────────────────────────────

.PHONY: help build up down restart logs shell

help:
	@echo "Available commands:"
	@echo "  make build       - Build the Docker images"
	@echo "  make up          - Start all services (in detached mode)"
	@echo "  make down        - Stop and remove all containers"
	@echo "  make restart     - Restart all services"
	@echo "  make logs        - Show logs from all services"
	@echo "  make shell       - Open a shell inside the app container"
	@echo "  make tidy        - Run go mod tidy inside the app"
	@echo "  make run         - Run the Go app locally (not via Docker)"

# ───────────────────────────────────────────────
# Docker targets
# ───────────────────────────────────────────────

build:
	$(DOCKER_COMPOSE) build --build-arg TARGET=$(SERVICE)

up:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down

restart: down up

logs:
	$(DOCKER_COMPOSE) logs -f $(APP_SERVICE)

shell:
	$(DOCKER_COMPOSE) exec $(APP_SERVICE) sh

dev-api:
	$(MAKE) build SERVICE=api
	$(DOCKER_COMPOSE) run --rm --service-ports app sh -c "go run ./cmd/api/main.go"
	
dev-worker:
	$(MAKE) build SERVICE=worker
	$(DOCKER_COMPOSE) run --rm --service-ports app sh -c "go run ./cmd/worker/main.go"

dev: dev-worker dev-api

# ───────────────────────────────────────────────
# Go-specific helpers (run inside container)
# ───────────────────────────────────────────────

tidy:
	$(DOCKER_COMPOSE) exec $(APP_SERVICE) go mod tidy

run:
	@echo "Running $(SERVICE) inside Docker..."
	$(DOCKER_COMPOSE) run --rm --service-ports app sh -c "go run ./cmd/$(SERVICE)/main.go"
