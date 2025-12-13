.PHONY: help build run run-postgres run-oracle run-surrealdb test clean docker-up docker-up-db docker-down docker-logs docker-build docker-run setup setup-docker

TARGET=dbcompare
BUILD_DIR=./bin
CONFIG_FILE=./configs/config.yml
DOCKER_CONFIG_FILE=./configs/config.docker.yml

help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build:
	@echo "Building $(TARGET)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(TARGET) cmd/dbcompare/main.go
	@echo "✓ Build complete: $(BUILD_DIR)/$(TARGET)"

run: build
	@$(BUILD_DIR)/$(TARGET) -config $(CONFIG_FILE)

run-postgres: build
	@$(BUILD_DIR)/$(TARGET) -config $(CONFIG_FILE) -db postgres

run-oracle: build
	@$(BUILD_DIR)/$(TARGET) -config $(CONFIG_FILE) -db oracle

run-surrealdb: build
	@$(BUILD_DIR)/$(TARGET) -config $(CONFIG_FILE) -db surrealdb

docker-build:
	@echo "Building Docker image..."
	@docker build -t dbcompare:latest .
	@echo "✓ Docker image built: dbcompare:latest"

docker-up:
	@echo "Starting all services..."
	@docker compose up -d
	@echo "✓ All services started"

docker-up-db:
	@echo "Starting all databases..."
	@docker compose up -d postgres oracle surrealdb
	@echo "✓ All services started"

docker-down:
	@echo "Stopping containers..."
	@docker compose down
	@echo "✓ Containers stopped"

docker-logs:
	@docker compose logs -f

docker-run: docker-build
	@echo "Ensuring config.docker.yml exists..."
	@test -f $(DOCKER_CONFIG_FILE) || cp configs/config.yml $(DOCKER_CONFIG_FILE)
	@echo "Running database comparison in Docker..."
	@docker compose up --build dbcompare-app
	@echo "✓ Comparison complete. Check ./results/ for output"

docker-clean: docker-down
	@echo "Removing containers and volumes..."
	@docker-compose down -v
	@docker rmi dbcompare:latest 2>/dev/null || true
	@echo "✓ Cleanup complete"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf results/*
	@echo "✓ Clean complete"

setup: docker-up
	@echo "Creating config file if not exists..."
	@test -f $(CONFIG_FILE) || cp configs/config.example.yml $(CONFIG_FILE)
	@test -f $(DOCKER_CONFIG_FILE) || cp configs/config.docker.yml $(DOCKER_CONFIG_FILE)
	@echo "✓ Setup complete! Run 'make run' to start locally or 'make docker-run' for Docker"

setup-docker:
	@echo "Creating Docker config if not exists..."
	@test -f $(DOCKER_CONFIG_FILE) || cp configs/config.example.yml $(DOCKER_CONFIG_FILE)
	@make docker-up
	@echo "✓ Docker setup complete! Run 'make docker-run' to start"

