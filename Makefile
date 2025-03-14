# TryTraGo Makefile

.PHONY: all build clean test lint docker-build docker-run run help migrate db-init db-reset

# Project variables
BINARY_NAME=trytrago
# VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
VERSION="dev"
COMMIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/valpere/trytrago/domain.Version=$(VERSION) -X github.com/valpere/trytrago/domain.CommitSHA=$(COMMIT_SHA) -X github.com/valpere/trytrago/domain.BuildTime=$(BUILD_TIME)"
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")

# Default target
all: clean lint test build

# Build the application
build:
	@echo "Building..."
	@mkdir -p build
	go build $(LDFLAGS) -o build/$(BINARY_NAME) .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf build/
	@rm -f coverage.out

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Generate test coverage
coverage:
	@echo "Generating test coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run linting
lint:
	@echo "Linting..."
	golangci-lint run --timeout=5m

# Run the application
run: build
	@echo "Running..."
	./build/$(BINARY_NAME) server

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

# Run with Docker
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):$(VERSION)

# Run with Docker Compose
docker-compose-up:
	@echo "Starting Docker Compose services..."
	docker-compose up -d

# Stop Docker Compose services
docker-compose-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Run database migrations
migrate:
	@echo "Running database migrations..."
	./build/$(BINARY_NAME) migrate --apply

# Initialize database schema
db-init: build
	@echo "Initializing database schema..."
	./build/$(BINARY_NAME) migrate --apply

# Reset database (drop and recreate)
db-reset:
	@echo "Resetting database..."
	./build/$(BINARY_NAME) migrate --rollback --all
	./build/$(BINARY_NAME) migrate --apply

# Generate Go code (GORM models, mock interfaces)
generate:
	@echo "Generating code..."
	go generate ./...

# Show help
help:
	@echo "TryTraGo Makefile"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          Build, test, and lint (default)"
	@echo "  build        Build the application"
	@echo "  clean        Clean build artifacts"
	@echo "  test         Run tests"
	@echo "  coverage     Generate test coverage"
	@echo "  lint         Run linting"
	@echo "  run          Run the application"
	@echo "  docker-build Build Docker image"
	@echo "  docker-run   Run with Docker"
	@echo "  docker-compose-up   Start Docker Compose services"
	@echo "  docker-compose-down Stop Docker Compose services"
	@echo "  migrate      Run database migrations"
	@echo "  db-init      Initialize database schema"
	@echo "  db-reset     Reset database"
	@echo "  generate     Generate Go code"
	@echo "  help         Show this help"
