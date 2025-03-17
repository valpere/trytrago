# Makefile

.PHONY: all build clean test test-unit test-integration swagger-setup run help

# Default target
all: help

# Build the application
build:
	@echo "Building TryTraGo..."
	@go build -o bin/trytrago main.go

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@go test -v ./test/unit/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@INTEGRATION_TEST=true go test -v ./test/integration/...

# Run all tests
test: test-unit test-integration

# Set up Swagger UI
swagger-setup:
	@echo "Setting up Swagger UI..."
	@chmod +x scripts/swagger-setup.sh
	@scripts/swagger-setup.sh

# Run the application
run: build
	@echo "Starting TryTraGo..."
	@./bin/trytrago server

# Generate OpenAPI specification
openapi-generate:
	@echo "Generating OpenAPI specification..."
	@mkdir -p docs
	@cp interface/api/rest/docs/openapi.yaml docs/OpenAPISpecification.yaml

# Help
help:
	@echo "Available commands:"
	@echo "  make build              - Build the application"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make test-unit          - Run unit tests"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make test               - Run all tests"
	@echo "  make swagger-setup      - Set up Swagger UI"
	@echo "  make run                - Run the application"
	@echo "  make openapi-generate   - Generate OpenAPI specification"
	@echo "  make help               - Show this help message"
