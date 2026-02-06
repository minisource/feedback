.PHONY: build run test clean docker-build docker-up docker-down

# Variables
APP_NAME := feedback
MAIN_PATH := ./cmd
BUILD_DIR := ./bin
DOCKER_COMPOSE := docker-compose

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

# Run the application
run:
	@go run $(MAIN_PATH)

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Format code
fmt:
	@go fmt ./...

# Lint code
lint:
	@golangci-lint run

# Download dependencies
deps:
	@go mod download
	@go mod tidy

# Docker commands
docker-build:
	@$(DOCKER_COMPOSE) build

docker-up:
	@$(DOCKER_COMPOSE) up -d

docker-down:
	@$(DOCKER_COMPOSE) down

docker-logs:
	@$(DOCKER_COMPOSE) logs -f

# Development with hot reload
dev:
	@$(DOCKER_COMPOSE) -f docker-compose.dev.yml up

# Generate swagger docs
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# All: format, lint, test, build
all: fmt lint test build
