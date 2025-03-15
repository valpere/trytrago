#!/bin/bash

echo "Starting test MySQL container..."
docker compose -f ../docker/docker-compose-test-ms.yml down

echo "Test MySQL is down!"
