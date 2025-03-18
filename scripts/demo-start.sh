#!/bin/bash

# Start demo environment with persistent volumes

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Set environment variables for the build
export BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
export COMMIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Starting TryTraGo demo environment..."

# Start the entire stack with Docker Compose
echo "Starting services..."
docker compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" up -d

# Check if the app is healthy
echo "Waiting for TryTraGo to be ready..."
timeout=120
while ! docker compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" exec -T trytrago wget --no-verbose --tries=1 --spider http://localhost:8080/health >/dev/null 2>&1; do
  timeout=$((timeout - 1))
  if [ $timeout -eq 0 ]; then
    echo "Timed out waiting for TryTraGo to be ready."
    exit 1
  fi
  echo "Waiting for TryTraGo... ($timeout seconds remaining)"
  sleep 1
done

echo "Demo environment is running!"
echo "Access the API at: http://localhost:8080"
echo "Access Adminer at: http://localhost:8081"
echo ""
echo "When finished, run 'scripts/stop-demo.sh' to stop the demo environment."
