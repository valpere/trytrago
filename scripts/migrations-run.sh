#!/bin/bash

# Run migrations on the development database
set -e

echo "Running database migrations..."
./build/trytrago migrate --apply

echo "Migrations completed successfully!"
