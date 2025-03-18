#!/bin/bash

# Restore the database to the demo environment

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="$PROJECT_ROOT/backups"

# Check if backup file is provided
if [ -z "$1" ]; then
  echo "Error: No backup file specified."
  echo "Usage: $0 <backup_file>"
  echo ""
  echo "Available backups:"
  ls -1 "$BACKUP_DIR"
  exit 1
fi

BACKUP_FILE="$1"

# If only filename is provided without path, assume it's in the backup directory
if [ ! -f "$BACKUP_FILE" ]; then
  BACKUP_FILE="$BACKUP_DIR/$BACKUP_FILE"
fi

# Check if the backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
  echo "Error: Backup file not found: $BACKUP_FILE"
  exit 1
fi

echo "Restoring PostgreSQL database from $BACKUP_FILE..."

# Check if the database container is running
if ! docker-compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" ps | grep -q postgres; then
  echo "Error: PostgreSQL container is not running."
  exit 1
fi

# Drop and recreate the database
docker-compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" exec -T postgres \
  psql -U postgres -c "DROP DATABASE IF EXISTS trytrago;"

docker-compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" exec -T postgres \
  psql -U postgres -c "CREATE DATABASE trytrago;"

# Restore from the backup
if [[ "$BACKUP_FILE" == *.gz ]]; then
  # Gunzip and pipe to psql
  gzip -dc "$BACKUP_FILE" | docker-compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" exec -T postgres \
    psql -U postgres -d trytrago
else
  # Plain SQL file
  cat "$BACKUP_FILE" | docker-compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" exec -T postgres \
    psql -U postgres -d trytrago
fi

if [ $? -eq 0 ]; then
  echo "Restore completed successfully."
else
  echo "Error: Restore failed."
  exit 1
fi

# Restart the application to pick up the new database
echo "Restarting the TryTraGo application..."
docker-compose -f "$PROJECT_ROOT/docker/docker-compose-demo.yml" restart trytrago

echo "Database restored and application restarted."
