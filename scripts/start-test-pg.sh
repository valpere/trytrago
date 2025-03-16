#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "Starting test PostgreSQL container..."
docker compose -f "${SCRIPT_DIR}/../docker/docker-compose-test-pg.yml" up --detach

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 5

echo "Test PostgreSQL is ready!"
