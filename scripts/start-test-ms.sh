#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "Starting test MySQL container..."
docker compose -f "${SCRIPT_DIR}/../docker/docker-compose-test-ms.yml" up --detach

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
sleep 5

echo "Test MySQL is ready!"
