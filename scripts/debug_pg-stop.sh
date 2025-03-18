#!/bin/bash

DBMS="PostgreSQL"

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Stopping development environment..."
docker compose -f "${PROJECT_ROOT}/docker/docker-compose-dev-pg.yml" down

echo "Development environment stopped."
