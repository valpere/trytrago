.PHONY: build clean test test-unit test-integration lint docker

# Build settings
BINARY_NAME=trytrago
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/valpere/trytrago/domain.Version=$(VERSION) -X github.com/valpere/trytrago/domain.CommitSHA=$(COMMIT_SHA) -X github.com/valpere/trytrago/domain.BuildTime=$(BUILD_TIME)"

# Default target
all: clean build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Run the application
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./test/unit/...

# Run integration tests (requires database connections)
test-integration:
	@echo "Running integration tests..."
	INTEGRATION_TEST=true go test -v ./test/integration/...

# Run all tests
test: test-unit test-integration

# Run linting checks
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Generate test coverage report
test-coverage:
	@echo "Generating test coverage report..."
	@mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t valpere/$(BINARY_NAME):$(VERSION) .

# Run in Docker
docker-run: docker
	@echo "Running in Docker..."
	docker run --rm -p 8080:8080 valpere/$(BINARY_NAME):$(VERSION)

# Create a new database migration file
migration-create:
	@echo "Creating migration..."
	@read -p "Enter migration name: " name; \
	version=$$(date +%Y%m%d%H%M%S); \
	echo "Creating migration V$${version}__$${name}.sql"; \
	touch migrations/V$${version}__$${name}.sql; \
	touch migrations/R$${version}__rollback_$${name}.sql

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@./$(BUILD_DIR)/$(BINARY_NAME) migrate --apply

# Show help
help:
	@echo "TryTraGo Makefile commands:"
	@echo "  make build           - Build the application"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make run             - Build and run the application"
	@echo "  make test            - Run all tests"
	@echo "  make test-unit       - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-coverage   - Generate test coverage report"
	@echo "  make lint            - Run linting checks"
	@echo "  make docker          - Build Docker image"
	@echo "  make docker-run      - Run application in Docker"
	@echo "  make migration-create - Create a new migration file"
	@echo "  make migrate         - Run database migrations"
