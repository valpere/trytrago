#!/bin/bash

# Reset the development database
set -e

echo "Resetting trytrago_dev database..."

# Drop and recreate the database
docker exec -it docker-postgres-1 psql -U postgres -c "DROP DATABASE IF EXISTS trytrago_dev;"
docker exec -it docker-postgres-1 psql -U postgres -c "CREATE DATABASE trytrago_dev;"

# Create necessary extensions
docker exec -it docker-postgres-1 psql -U postgres -d trytrago_dev -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"

echo "Database reset complete!"
echo "Run './build/trytrago migrate --apply' to apply migrations."
