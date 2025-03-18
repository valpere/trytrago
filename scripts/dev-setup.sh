#!/bin/bash

# Setup script for local development

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "Setting up TryTraGo development environment..."

set -e

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go is not installed. Please install Go 1.24 or higher."
  exit 1
fi

# Check if Docker is installed
if ! command -v docker >/dev/null 2>&1; then
  echo "Error: Docker is not installed. Please install Docker."
  exit 1
fi

# # Check if Docker Compose is installed
# if ! command -v docker-compose >/dev/null 2>&1; then
  # echo "Error: Docker Compose is not installed. Please install Docker Compose."
  # exit 1
# fi

# Install development dependencies
echo "Installing development dependencies..."
go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest

# Set up Swagger UI
echo "Setting up Swagger UI..."
chmod +x "${SCRIPT_DIR}/swagger-setup.sh"
"${SCRIPT_DIR}/swagger-setup.sh"

# Set up environment variables for development
cat > "${SCRIPT_DIR}/../.env.dev.pg" << EOF
TRYTRAGO_DATABASE_HOST=localhost
TRYTRAGO_DATABASE_PORT=5432
TRYTRAGO_DATABASE_USER=postgres
TRYTRAGO_DATABASE_PASSWORD=postgres
TRYTRAGO_DATABASE_NAME=trytrago_dev
TRYTRAGO_CACHE_ADDRESS=localhost:6379
TRYTRAGO_ENVIRONMENT=development
TRYTRAGO_LOGGING_LEVEL=debug
TRYTRAGO_LOGGING_FORMAT=console
EOF

echo "Development environment setup complete!"
echo "To start the development services, run: ${SCRIPT_DIR}/start-dev.sh"
