#!/bin/bash

DBMS="PostgreSQL"

echo "Start development environment with ${DBMS}"

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load environment variables from .env.dev if it exists
ENV_DEV=".env.dev.pg"
if [ -f "$PROJECT_ROOT/${ENV_DEV}" ]; then
  export $(grep -v '^#' "$PROJECT_ROOT/${ENV_DEV}" | xargs)
fi

echo "Starting ${DBMS} and Redis..."
docker compose -f "${PROJECT_ROOT}/docker/docker-compose-dev-pg.yml" --profile debug up --detach

echo "Waiting for ${DBMS} to be ready..."
timeout=60
while ! docker compose -f "$PROJECT_ROOT/docker/docker-compose-dev-pg.yml" exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
  timeout=$((timeout - 1))
  if [ $timeout -eq 0 ]; then
    echo "Timed out waiting for ${DBMS} to be ready."
    exit 1
  fi
  echo "Waiting for ${DBMS}... ($timeout seconds remaining)"
  sleep 1
done

echo "Development ${DBMS} environment is ready!"
echo "Run 'make run' to start the application."
echo "When finished, run 'scripts/debug_pg-stop.sh' to stop the development environment."
