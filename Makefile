.PHONY: help dev stop test clean logs backend-logs redis-cli

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment with hot reload
	docker-compose up --build

dev-detached: ## Start development environment in background
	docker-compose up -d --build

stop: ## Stop all containers
	docker-compose down

test: ## Run tests in Docker
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down

test-watch: ## Run tests in watch mode
	docker-compose -f docker-compose.test.yml up --build

clean: ## Clean up containers, volumes, and networks
	docker-compose down -v
	docker-compose -f docker-compose.test.yml down -v

logs: ## Show logs from all containers
	docker-compose logs -f

backend-logs: ## Show backend logs only
	docker-compose logs -f backend

redis-cli: ## Connect to Redis CLI
	docker exec -it task-tracker-redis redis-cli

mailhog: ## Open MailHog web UI
	@echo "Opening MailHog at http://localhost:8025"
	@open http://localhost:8025 2>/dev/null || xdg-open http://localhost:8025 2>/dev/null || echo "Please open http://localhost:8025 in your browser"

rebuild: ## Rebuild and restart containers
	docker-compose down
	docker-compose up --build