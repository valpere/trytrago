#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "Starting test MySQL container..."
docker compose -f "${SCRIPT_DIR}/../docker/docker-compose-test-ms.yml" down

echo "Test MySQL is down!"
