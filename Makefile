# Project variables
PROJECT_NAME := trytrago
MAIN_PACKAGE := github.com/valpere/$(PROJECT_NAME)
BINARY_NAME := $(PROJECT_NAME)
PROJECT_VERSION := v0.1.0

# Get the current git version and commit
# GIT_VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT_SHA ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS := -ldflags "\
	-X $(MAIN_PACKAGE)/domain.Version=$(PROJECT_VERSION) \
	-X $(MAIN_PACKAGE)/domain.CommitSHA=$(GIT_COMMIT_SHA) \
	-X $(MAIN_PACKAGE)/domain.BuildTime=$(BUILD_TIME)"

# Environment variables
GO ?= go
GOPATH ?= $(shell $(GO) env GOPATH)
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)

# Build directory
BUILD_DIR := build

# All source files, used for formatting and linting
SOURCES := $(shell find . -name "*.go" -not -path "./vendor/*")

.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help message
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z_-]+:.*?## / { \
		printf "  \033[36m%-20s\033[0m %s\n", substr($$1, 1, length($$1)-1), substr($$0, index($$0, "##") + 3) \
	}' $(MAKEFILE_LIST)

.PHONY: build
build: clean ## Build the binary
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned build directory"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v ./...

.PHONY: coverage
coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	$(GO) test -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

.PHONY: fmt
fmt: ## Format code using gofmt
	@echo "Formatting code..."
	@gofmt -s -w $(SOURCES)
	@echo "Code formatting complete"

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "Linting complete"

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

.PHONY: run
run: build ## Run the application
	@echo "Running $(PROJECT_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: install
install: build ## Install the binary
	@echo "Installing $(PROJECT_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed at $(GOPATH)/bin/$(BINARY_NAME)"
