#!/bin/bash

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Starting test environment..."

# Start test services
docker compose -f "$PROJECT_ROOT/docker/docker-compose-test.yml" up -d

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
timeout=30
while ! docker compose -f "$PROJECT_ROOT/docker/docker-compose-test.yml" exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
  timeout=$((timeout - 1))
  if [ $timeout -eq 0 ]; then
    echo "Timed out waiting for PostgreSQL to be ready."
    exit 1
  fi
  sleep 1
done
