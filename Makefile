.PHONY: help build run test clean docker-build docker-up docker-down migrate gen-key

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build both bot and worker binaries
	@echo "Building bot..."
	@go build -o bin/bot cmd/bot/main.go
	@echo "Building worker..."
	@go build -o bin/worker cmd/worker/main.go
	@echo "Build complete!"

run-bot: ## Run the Telegram bot
	@go run cmd/bot/main.go

run-worker: ## Run the background worker
	@go run cmd/worker/main.go

test: ## Run tests
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

docker-build: ## Build Docker image
	@docker build -t seller-assistant:latest .

docker-up: ## Start all services with Docker Compose
	@docker-compose up -d
	@echo "Services started! View logs with: make docker-logs"

docker-down: ## Stop all Docker services
	@docker-compose down

docker-logs: ## View Docker logs
	@docker-compose logs -f

docker-restart: ## Restart Docker services
	@docker-compose restart

migrate: ## MongoDB indexes are auto-created on startup
	@echo "MongoDB indexes are automatically created when the application starts."
	@echo "No manual migration needed!"

gen-key: ## Generate a new encryption key
	@echo "Generating 32-byte encryption key (base64):"
	@openssl rand -base64 32

deps: ## Download dependencies
	@go mod download
	@echo "Dependencies downloaded!"

fmt: ## Format code
	@go fmt ./...
	@echo "Code formatted!"

lint: ## Run linter (requires golangci-lint)
	@golangci-lint run

dev: ## Run both bot and worker in development mode
	@echo "Starting development environment..."
	@make -j2 run-bot run-worker

.DEFAULT_GOAL := help
