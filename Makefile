# Makefile for Server Management System

.PHONY: build test clean run docker-build docker-up docker-down docker-logs help

# Variables
APP_NAME=server-management-system
DOCKER_COMPOSE=docker-compose
GO_FILES=$(shell find . -name "*.go")

# Default target
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  test        - Run tests"
	@echo "  test-cover  - Run tests with coverage"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run the application locally"
	@echo "  docker-build- Build Docker image"
	@echo "  docker-up   - Start all services with Docker Compose"
	@echo "  docker-down - Stop all services"
	@echo "  docker-logs - View logs from all services"
	@echo "  lint        - Run linter"
	@echo "  deps        - Download dependencies"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run the application locally
run:
	@echo "Running $(APP_NAME)..."
	go run cmd/server/main.go

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME) .

docker-up:
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d

docker-down:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-logs:
	@echo "Viewing logs..."
	$(DOCKER_COMPOSE) logs -f

# Database migrations (if needed)
migrate-up:
	@echo "Running database migrations..."
	# Add migration command here

migrate-down:
	@echo "Rolling back database migrations..."
	# Add rollback command here

# Generate swagger docs
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
