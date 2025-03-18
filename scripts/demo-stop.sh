#!/bin/bash
# scripts/stop-demo.sh
# Stop demo environment

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Stopping demo environment..."

# Stop the services but keep volumes
docker-compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" down

echo "Demo environment stopped. Data volumes have been preserved."
echo "To remove all data volumes, run: docker-compose -f \"$PROJECT_ROOT/docker/docker-compose-demo.yml\" down -v"
