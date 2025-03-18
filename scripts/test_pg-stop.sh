#!/bin/bash

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Cleaning up test environment..."
docker-compose -f "$PROJECT_ROOT/docker/docker-compose-test.yml" down
