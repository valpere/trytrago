#!/bin/bash

# Backup the database from the demo environment

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="$PROJECT_ROOT/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/trytrago_backup_$TIMESTAMP.sql.gz"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

echo "Backing up PostgreSQL database to $BACKUP_FILE..."

# Check if the database container is running
if ! docker compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" ps | grep -q postgres; then
  echo "Error: PostgreSQL container is not running."
  exit 1
fi

# Execute pg_dump inside the container and pipe to gzip
docker compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" exec -T postgres \
  pg_dump -U postgres trytrago | gzip > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
  echo "Backup completed successfully: $BACKUP_FILE"
else
  echo "Error: Backup failed."
  exit 1
fi

# List recent backups
echo "Recent backups:"
ls -lh "$BACKUP_DIR" | tail -5
