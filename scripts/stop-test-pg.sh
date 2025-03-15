#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "Starting test PostgreSQL container..."
docker compose -f "${SCRIPT_DIR}/../docker/docker-compose-test-pg.yml" down

echo "Test PostgreSQL is down!"
