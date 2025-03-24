#!/bin/bash

# Create the development database
set -e

echo "Creating trytrago database..."
docker exec -it docker-postgres-1 psql -U postgres -c "CREATE DATABASE trytrago_dev;"
echo "Database created successfully."

# Optional: Create extensions if needed
docker exec -it docker-postgres-1 psql -U postgres -d trytrago_dev -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"

echo "Database setup complete!"
