.PHONY: help build up down restart logs clean ps

# Variables
DOCKER_COMPOSE = docker-compose -f docker_compose.yaml

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all Docker images
	$(DOCKER_COMPOSE) build

up: ## Start all services
	$(DOCKER_COMPOSE) up -d

down: ## Stop all services
	$(DOCKER_COMPOSE) down

restart: ## Restart all services
	$(DOCKER_COMPOSE) restart

logs: ## Show logs from all services
	$(DOCKER_COMPOSE) logs -f

logs-app: ## Show logs from app service only
	$(DOCKER_COMPOSE) logs -f app

logs-postgres: ## Show logs from postgres service only
	$(DOCKER_COMPOSE) logs -f postgres

logs-mongodb: ## Show logs from mongodb service only
	$(DOCKER_COMPOSE) logs -f mongodb

ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

clean: ## Stop and remove all containers, networks, and volumes
	$(DOCKER_COMPOSE) down -v --remove-orphans

clean-all: clean ## Clean everything including images
	docker system prune -af --volumes

rebuild: down build up ## Rebuild and restart all services

dev-up: ## Start services for local development (only databases)
	$(DOCKER_COMPOSE) up -d postgres mongodb

dev-down: ## Stop development services
	$(DOCKER_COMPOSE) stop postgres mongodb

swagger: ## Generate swagger documentation
	@echo "Generating Swagger docs..."
	@swag init -g cmd/main.go -o docs
	@if grep -q "LeftDelim:" docs/docs.go; then \
		echo "Fixing compatibility..."; \
		sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go; \
	fi
	@echo "✅ Swagger docs generated!"

swagger-rebuild: swagger rebuild ## Regenerate swagger and rebuild app
