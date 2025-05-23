.PHONY: all build run test clean lint deps migrate docker

# Default target
all: clean build

# Build the application
build:
	@echo "Building application..."
	go build -o scheduling-api ./cmd/api

# Run the application
run:
	@echo "Running application..."
	go run ./cmd/api/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f scheduling-api
	rm -f coverage.out

# Run linter
lint:
	@echo "Running linter..."
	go vet ./...
	@if command -v golint > /dev/null; then \
		golint ./...; \
	else \
		echo "golint not installed. Run: go install golang.org/x/lint/golint@latest"; \
	fi

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t scheduling-api .

# Run migrations
migrate:
	@echo "Running database migrations..."
	go run ./cmd/api/main.go --migrate-only

# Development mode with hot reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		echo "Running with hot reload..."; \
		air; \
	else \
		echo "air not installed. Run: go install github.com/cosmtrek/air@latest"; \
		go run ./cmd/api/main.go; \
	fi

# Help information
help:
	@echo "Available targets:"
	@echo "  all           - Clean and build the application"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  lint          - Run linter"
	@echo "  deps          - Install dependencies"
	@echo "  docker        - Build Docker image"
	@echo "  migrate       - Run database migrations"
	@echo "  dev           - Run with hot reload (requires air)"
	@echo "  help          - Show this help information"

