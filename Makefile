.PHONY: build clean test lint vet fmt help run docker-build docker-run setup test-unit test-integration test-api test-all

# Build settings
BINARY_NAME=trytrago
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_DIR=./build
LDFLAGS=-ldflags "-X github.com/valpere/trytrago/domain.Version=$(VERSION)"

# Docker settings
DOCKER_IMAGE=trytrago
DOCKER_TAG=latest

# Default target
.DEFAULT_GOAL := help

# Set up dev environment
setup: ## Install development dependencies
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Build the application
build: ## Build the binary
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

# Clean build artifacts
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

# Run the application
run: ## Run the application
	go run $(LDFLAGS) ./main.go server

# Run unit tests
test-unit: ## Run unit tests
	go test -v -race ./test/unit/...

# Run integration tests
test-integration: ## Run integration tests
	INTEGRATION_TEST=true go test -v ./test/integration/...

# Run API tests
test-api: ## Run API endpoint and auth flow tests
	go test -v ./test/api/...
# 	go test -v ./test/auth/...

# Run all tests
test-all: test-unit test-integration test-api ## Run all tests

# Default test command runs unit tests
test: test-unit ## Run unit tests (default)

# Code quality tools
lint: ## Run linter
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Run gofmt
	go fmt ./...

# Docker commands
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Generate documentation
docs: ## Generate API documentation
	swag init -g interface/api/rest/router.go

# Help command
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
