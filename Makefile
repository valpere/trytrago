# TryTraGo Makefile

.PHONY: all build clean test test-unit test-integration test-api test-auth test-all lint vet fmt run docker-build docker-run docker-compose-up docker-compose-down setup help migrate db-init db-reset migration-create swagger-setup openapi-generate docs test-coverage

# Build settings
BINARY_NAME=trytrago
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/valpere/trytrago/domain.Version=$(VERSION) -X github.com/valpere/trytrago/domain.CommitSHA=$(COMMIT_SHA) -X github.com/valpere/trytrago/domain.BuildTime=$(BUILD_TIME)"

# Docker settings
DOCKER_IMAGE=valpere/$(BINARY_NAME)
DOCKER_TAG=$(VERSION)

# Default target
.DEFAULT_GOAL := help

# Set up dev environment
setup: ## Install development dependencies
	@echo "Setting up development environment..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Build the application
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# Clean build artifacts
clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out

# Run the application
run: build ## Run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) server

# Run unit tests
test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test -v -race ./test/unit/...

# Run integration tests
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	INTEGRATION_TEST=true go test -v ./test/integration/...

# Run API tests
test-api: ## Run API endpoint tests
	@echo "Running API tests..."
	go test -v ./test/api/...

# Run authentication flow tests
test-auth: ## Run authentication flow tests
	@echo "Running authentication flow tests..."
	go test -v ./test/auth/...

# Run all tests
test-all: test-unit test-integration test-api test-auth ## Run all tests
	@echo "All tests passed!"

# Default test command
test: test-unit ## Run unit tests (default)

# Generate test coverage
test-coverage: ## Generate test coverage report
	@echo "Generating test coverage report..."
	@mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

# Code quality tools
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

fmt: ## Run gofmt
	@echo "Running gofmt..."
	go fmt ./...

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker Compose commands
docker-compose-up: ## Start Docker Compose services
	@echo "Starting Docker Compose services..."
	docker-compose up -d

docker-compose-down: ## Stop Docker Compose services
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Database operations
migrate: build ## Run database migrations
	@echo "Running database migrations..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate --apply

db-init: build ## Initialize database schema
	@echo "Initializing database schema..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate --apply

db-reset: build ## Reset database (drop and recreate)
	@echo "Resetting database..."
	./$(BUILD_DIR)/$(BINARY_NAME) migrate --rollback --all
	./$(BUILD_DIR)/$(BINARY_NAME) migrate --apply

# Create a new database migration file
migration-create: ## Create a new database migration
	@echo "Creating migration..."
	@read -p "Enter migration name: " name; \
	version=$$(date +%Y%m%d%H%M%S); \
	echo "Creating migration V$${version}__$${name}.sql"; \
	touch migrations/V$${version}__$${name}.sql; \
	touch migrations/R$${version}__rollback_$${name}.sql

# Documentation
swagger-setup: ## Set up Swagger UI
	@echo "Setting up Swagger UI..."
	@chmod +x scripts/swagger-setup.sh
	@scripts/swagger-setup.sh

openapi-generate: ## Generate OpenAPI specification
	@echo "Generating OpenAPI specification..."
	@mkdir -p docs
	@cp interface/api/rest/docs/openapi.yaml docs/OpenAPISpecification.yaml

docs: openapi-generate ## Generate API documentation
	@echo "Generating API documentation..."
	swag init -g interface/api/rest/router.go

# Generate Go code
generate: ## Generate Go code (GORM models, mock interfaces)
	@echo "Generating code..."
	go generate ./...

# Full development workflow
all: clean lint test-unit build ## Build, test and lint (default)

# Help command
help: ## Display this help
	@echo "TryTraGo Makefile"
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
