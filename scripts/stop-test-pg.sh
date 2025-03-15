#!/bin/bash

echo "Starting test PostgreSQL container..."
docker compose -f ../docker/docker-compose-test-pg.yml down

echo "Test PostgreSQL is down!"
