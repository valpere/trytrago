#!/bin/bash

# Run all tests with proper environment

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Starting test environment..."

"${SCRIPT_DIR}/test_pg-start.sh"

# Set environment variables for tests
if [ -f "$PROJECT_ROOT/.env.test.pg" ]; then
  export $(grep -v '^#' "$PROJECT_ROOT/.env.test.pg" | xargs)
fi

pushd "$PROJECT_ROOT"

# echo "Running unit tests..."
# go test -v -race ./test/unit/...

# echo "Running integration tests..."
# INTEGRATION_TEST=true  go test -v ./test/integration/...

# echo "Running API tests..."
# go test -v ./test/api/...

# echo "Running auth flow tests..."
# go test -v ./test/auth/...

TEST_TYPE="$1"
make ${TEST_TYPE:-"test-all"}

popd

# Clean up test environment
"${SCRIPT_DIR}/test_pg-stop.sh"

echo "All tests completed!"
