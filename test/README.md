# TryTraGo Testing Guide

This directory contains tests for the TryTraGo project. The tests are organized into the following categories:

## Test Organization

- **Unit Tests**: Located in `test/unit/` - Tests individual components in isolation with mocked dependencies
- **Integration Tests**: Located in `test/integration/` - Tests components with real database connections
- **Mocks**: Located in `test/mocks/` - Contains mock implementations for testing

## Running Tests

### Unit Tests

To run unit tests:

```bash
# Run all unit tests
make test-unit

# Run a specific unit test file
go test -v ./test/unit/service/user_service_test.go
```

### Integration Tests

Integration tests require database connections and environment variables:

```bash
# Run all integration tests
make test-integration

# Run a specific integration test
INTEGRATION_TEST=true go test -v ./test/integration/repository/postgres_repository_test.go
```

### Setting Up Test Databases

For PostgreSQL integration tests:

```bash
# Create a test database
createdb trytrago_test

# Optionally set environment variables for custom database connections
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_NAME=trytrago_test
```

For SQLite integration tests, a temporary database file will be created automatically during testing.

## Test Coverage

Generate a test coverage report:

```bash
make test-coverage
```

This will create coverage reports in the `build/` directory.

## Creating Tests

### Unit Tests

When creating a new unit test:

1. Place the test file in the appropriate directory under `test/unit/`
2. Use the provided mock implementations in `test/mocks/`
3. Follow the table-driven test pattern
4. Ensure all error cases are tested

### Integration Tests

When creating integration tests:

1. Place the test file in the appropriate directory under `test/integration/`
2. Extend the `BaseRepositoryTestSuite` for repository tests
3. Use helper methods to set up test data
4. Clean up after tests to leave the database in a clean state
5. Skip tests when not running integration mode

## Mocking Guidelines

For consistent mocking:

1. Use `SetupLoggerMock()` for common logger initialization
2. Verify expectations with `AssertExpectations()` and `VerifyLoggerMock()`
3. Use `ExpectDebug()`, `ExpectInfo()`, `ExpectWarn()`, and `ExpectError()` helper methods for readable test code
4. Be specific about the number of calls expected with `.Once()`
