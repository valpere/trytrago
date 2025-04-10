#!/bin/bash

DBMS="MySQL"

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Stopping development environment..."
docker compose -f "${PROJECT_ROOT}/docker/docker-compose-dev-ms.yml" --profile debug down

echo "Development environment stopped."
echo "To remove all data volumes, run: docker compose -f \"$PROJECT_ROOT/docker/docker-compose-dev-ms.yml\" down -v"