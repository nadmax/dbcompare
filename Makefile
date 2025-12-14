.PHONY: help build run clean docker-up docker-down docker-logs docker-build docker-run run-postgres run-surrealdb

TARGET=dbcompare
BUILD_DIR=./bin
CONFIG_FILE=./configs/config.yml

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(TARGET)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(TARGET) cmd/dbcompare/main.go
	@echo "✓ Build complete: $(BUILD_DIR)/$(TARGET)"

run: build ## Build and run the application locally
	@$(BUILD_DIR)/$(TARGET) -config $(CONFIG_FILE)

run-postgres: build ## Run PostgreSQL comparison only
	@$(BUILD_DIR)/$(TARGET) -config $(CONFIG_FILE) -db postgres

run-surrealdb: build ## Run SurrealDB comparison only
	@$(BUILD_DIR)/$(TARGET) -config $(CONFIG_FILE) -db surrealdb

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t dbcompare:latest .
	@echo "✓ Docker image built: dbcompare:latest"

docker-up: ## Start all services including app
	@echo "Starting all services..."
	@docker compose up -d
	@echo "✓ All services started"

docker-up-db: ## Start all database contianers (without app)
	@echo "Starting all databases..."
	@docker compose up -d postgres surrealdb
	@echo "✓ All services started"

docker-down: ## Stop all containers
	@echo "Stopping containers..."
	@docker compose down
	@echo "✓ Containers stopped"

docker-logs: ## Show all container logs
	@docker compose logs -f

docker-logs-app: ## Show app logs
	@docker compose logs -f app

docker-logs-postgres: ## Show PostgreSQL logs
	@docker compose logs -f postgres

docker-logs-surrealdb: ## Show SurrealDB logs
	@docker compose logs -f surrealdb

docker-run: docker-build ## Build and run comparison in Docker
	@echo "Running database comparison in Docker..."
	@docker compose up --build dbcompare-app
	@echo "✓ Comparison complete. Check ./results/ for output"

docker-clean: docker-down ## Remove all containers and volumes
	@echo "Removing containers and volumes..."
	@docker compose down -v
	@docker rmi dbcompare:latest 2>/dev/null || true
	@echo "✓ Cleanup complete"

clean: ## Clean build artifacts and results
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf results/*
	@echo "✓ Clean complete"

setup: docker-up ## Complete setup for local development
	@echo "Creating config file if not exists..."
	@test -f $(CONFIG_FILE) || cp configs/config.example.yml $(CONFIG_FILE)
	@echo "✓ Setup complete! Run 'make run' to start locally or 'make docker-run' for Docker"

