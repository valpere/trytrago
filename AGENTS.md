# AGENTS.md - TryTraGo Developer Guide

This file provides guidance for agentic coding agents working in this repository.

## Project Overview

TryTraGo is a Go-based dictionary/translation API service using:
- **Go 1.24+** with Gin web framework
- **GORM** for database abstraction (PostgreSQL, MySQL, SQLite)
- **Redis** for caching
- **JWT** for authentication
- **Zap** for structured logging

## Design Principles

The project strictly adheres to the following design principles:

### Core Principles

1. **DRY (Don't Repeat Yourself)**: Avoid code duplication by creating reusable components and functions
2. **YAGNI (You Aren't Gonna Need It)**: Only implement features when they're actually needed
3. **KISS (Keep It Simple, Stupid)**: Prefer simple, straightforward solutions over complex ones
4. **Encapsulation**: Hide implementation details and expose clean interfaces
5. **PoLA (Principle of Least Astonishment)**: Design APIs and behaviors to be intuitive and predictable

### SOLID Principles

- **Single Responsibility**: Each type/function should have one clear purpose
- **Open/Closed**: Open for extension, closed for modification
- **Liskov Substitution**: Derived types must be substitutable for base types
- **Interface Segregation**: Many specific interfaces are better than one general interface
- **Dependency Inversion**: Depend on abstractions, not concretions

### GRASP Patterns

- **Information Expert**: Assign responsibilities to classes with necessary information
- **Creator**: Assign creation responsibility to classes with creation knowledge
- **Low Coupling**: Minimize dependencies between classes
- **High Cohesion**: Keep related responsibilities focused within classes
- **Controller**: Handle system events in dedicated controller classes
- **Pure Fabrication**: Introduce new classes to fulfill responsibilities without violating principles

## Build, Lint, and Test Commands

### Building
```bash
make build           # Build the binary to build/trytrago
make clean          # Remove build artifacts
make run            # Build and run the server
```

### Testing
```bash
make test           # Run unit tests (default)
make test-unit      # Run all unit tests
make test-integration   # Run integration tests (requires database)
make test-api       # Run API endpoint tests
make test-auth      # Run authentication flow tests
make test-all       # Run all tests
make test-coverage  # Generate coverage report at build/coverage.html
```

**Running a single test:**
```bash
# Single test file
go test -v ./test/unit/service/entry_service_test.go

# Single test function
go test -v -run TestCreateEntry ./test/unit/service/

# Single test with race detector
go test -v -race ./test/unit/service/entry_service_test.go

# Integration test (requires INTEGRATION_TEST=true)
INTEGRATION_TEST=true go test -v ./test/integration/repository/postgres_repository_test.go
```

### Code Quality
```bash
make lint           # Run golangci-lint
make vet            # Run go vet
make fmt            # Run gofmt
```

### Database Operations
```bash
make migrate        # Run database migrations
make db-init        # Initialize database schema
make db-reset       # Reset database (drop and recreate)
make migration-create   # Create new migration file
```

## Code Style Guidelines

### Project Structure (Clean Architecture)

```
cmd/           # CLI entry points
domain/        # Business logic, models, errors, interfaces
application/   # Use cases, DTOs, mappers, services
infrastructure/ # External implementations (DB, cache, auth)
interface/     # HTTP handlers, middleware, routing
test/          # Test files (unit, integration, mocks)
migrations/    # Database migration files
```

### Type Organization

Types are organized in a hierarchical structure to avoid conflicts:

1. **Domain Layer Types** (`domain/`): Core business types, models, errors, interfaces
2. **Application Layer Types** (`application/`): DTOs, mappers, service implementations
3. **Interface Layer Types** (`interface/`): HTTP handlers, middleware, routing

When type conflicts occur, keep types in the highest-level appropriate layer and remove duplicates.

### Naming Conventions

- **Packages**: Use descriptive, lowercase names (e.g., `service`, `repository`, `handler`)
- **Types**: PascalCase (e.g., `EntryService`, `CreateEntryRequest`)
- **Interfaces**: Add `er` suffix for interface names (e.g., `Repository`, `Service`)
- **Variables**: camelCase (e.g., `entryService`, `createEntryReq`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Files**: lowercase with underscores for multiple words (e.g., `entry_service.go`)

### Imports

Organize imports in three groups with blank lines between:
1. Standard library
2. Third-party packages
3. Internal packages

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "github.com/valpere/trytrago/application/dto/request"
    "github.com/valpere/trytrago/domain/database"
    "github.com/valpere/trytrago/domain/logging"
)
```

### Types and Interfaces

- Define interfaces in the `domain` layer, implement in `infrastructure` or `application`
- Use struct composition for shared behavior
- Prefer concrete types over interfaces unless polymorphism is needed
- Define DTOs in `application/dto/request` and `application/dto/response`

### Error Handling

- Use sentinel errors from `domain/errors` for known error types
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return typed errors (e.g., `ErrEntryNotFound`, `ErrDuplicateEntry`)
- Log errors at the point of detection with appropriate level
- Use structured logging with key-value pairs

```go
if err := s.repo.GetEntryByID(ctx, id); err != nil {
    s.logger.Error("failed to get entry",
        logging.Error(err),
        logging.String("id", id.String()),
    )
    return nil, fmt.Errorf("failed to get entry: %w", err)
}
```

### Logging

Use structured logging from `domain/logging`:
- `logger.Debug()` for debug information
- `logger.Info()` for normal operations
- `logger.Warn()` for recoverable issues
- `logger.Error()` for errors

Always include relevant context:
```go
s.logger.Info("creating entry",
    logging.String("word", req.Word),
    logging.Int("userID", userID),
)
```

### Testing Guidelines

- Follow table-driven test pattern
- Place tests in `test/unit/` for unit tests, `test/integration/` for integration
- Use `test/mocks/` for mock implementations
- Use `SetupLoggerMock()` for logger initialization in tests
- Test both success and error cases
- Use `testify` assertions: `assert`, `require`

```go
func TestCreateEntry(t *testing.T) {
    testCases := []struct {
        name          string
        setupMocks    func(*mocks.MockRepository)
        expectedError bool
    }{
        {
            name: "Success",
            setupMocks: func(mockRepo *mocks.MockRepository) {
                mockRepo.On("CreateEntry", mock.Anything, mock.Anything).Return(nil).Once()
            },
            expectedError: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Configuration

- Use Viper for configuration (see `cmd/config.go`)
- Define config in `domain/config.go`
- Support config file, environment variables, and flags

### API Design

- RESTful endpoints following standard conventions
- Request/Response DTOs in `application/dto/`
- Use middleware for cross-cutting concerns (auth, logging, validation)
- Return appropriate HTTP status codes

### Dependency Injection

- Pass dependencies as constructor parameters
- Use interfaces for dependencies to enable mocking
- Initialize services in `cmd/server.go` or main entry points

## Repository Interface Guidelines

The `Repository` interface in `domain/database/repository/repository.go` provides direct database access methods. Always prefer direct lookups over list-based searches:

### Direct Lookup Methods (Preferred)

- `GetMeaningByID(ctx, id)` - Direct O(1) lookup for meanings
- `UpdateMeaning(ctx, meaning)` - Direct update for meanings
- `DeleteMeaning(ctx, id)` - Direct delete for meanings
- `GetTranslationByID(ctx, id)` - Direct O(1) lookup for translations
- `UpdateTranslation(ctx, translation)` - Direct update for translations
- `DeleteTranslation(ctx, id)` - Direct delete for translations

### Never Use List-Based Lookups

Do NOT use `ListEntries()` to find meanings or translations - this causes O(N) performance issues. Always use the direct methods above.

### Error Types

Use domain error types from `domain/database`:
- `database.ErrEntryNotFound`
- `database.ErrMeaningNotFound`
- `database.ErrTranslationNotFound`
- `database.ErrDuplicateEntry`

## Caching Guidelines

When implementing caching:
- Use `SCAN` instead of `KEYS` for pattern matching (see `infrastructure/cache/redis.go`)
- Implement targeted cache invalidation rather than broad patterns
- Log cache operations for monitoring
