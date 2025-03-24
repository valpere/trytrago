# TryTraGo

[![Go Report Card](https://goreportcard.com/badge/github.com/valpere/trytrago)](https://goreportcard.com/report/github.com/valpere/trytrago)
[![Go Version](https://img.shields.io/github/go-mod/go-version/valpere/trytrago)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

TryTraGo is a high-performance multilanguage dictionary server designed to support approximately 60 million dictionary entries with robust API functionality, social features, and multi-database support.

## Features

- **Comprehensive Dictionary Functionality**
  - Dictionary entries with meanings and translations
  - Support for multiple word types (words, compound words, phrases)
  - Pronunciation information

- **Social Features**
  - Comments on meanings and translations
  - Like/unlike functionality
  - User profiles and contributions tracking

- **Clean Architecture**
  - Domain-driven design with clear separation of concerns
  - Four-layer architecture (domain, application, interface, infrastructure)
  - Highly testable and maintainable codebase

- **Multi-Database Support**
  - PostgreSQL (primary for production)
  - MySQL (alternative)
  - SQLite (development and testing)

- **Technology Stack**
  - Go 1.24+
  - Gin web framework
  - GORM for database access
  - JWT authentication
  - Uber-zap for structured logging
  - Redis for caching

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose (for local development)
- PostgreSQL, MySQL, or SQLite

### Installation

1. Clone the repository

```bash
git clone https://github.com/valpere/trytrago.git
cd trytrago
```

2. Install dependencies

```bash
make setup
```

3. Build the application

```bash
make build
```

4. Run database migrations

```bash
make db-init
```

5. Start the server

```bash
make run
```

### Docker

Run using Docker:

```bash
make docker-run
```

Or with Docker Compose:

```bash
make docker-compose-up
```

## API Usage

TryTraGo provides a RESTful API for all dictionary operations. You can view the API documentation at:

- Swagger UI: `http://localhost:8080/swagger-ui.html`
- OpenAPI Specification: `http://localhost:8080/v3/api-docs`

### Authentication

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

# Login and get a token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### Dictionary Operations

```bash
# List entries
curl http://localhost:8080/api/v1/entries

# Create a new entry (requires authentication)
curl -X POST http://localhost:8080/api/v1/entries \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"word":"example","type":"WORD","pronunciation":"ɪɡˈzæmpəl"}'
```

## Development

### Project Structure

```
trytrago/
├── application/      # Application services and DTOs
├── cmd/              # CLI commands
├── docker/           # Docker configuration
├── docs/             # Documentation
├── domain/           # Core business entities and interfaces
├── infrastructure/   # External details implementations
├── interface/        # HTTP handlers and middleware
├── migrations/       # Database migration files
├── scripts/          # Utility scripts
├── test/             # Test suites
└── main.go           # Application entry point
```

### Make Commands

```bash
# Development workflow
make setup            # Install development dependencies
make build            # Build the application
make run              # Run the application
make test             # Run unit tests
make test-all         # Run all tests (unit, integration, API)
make lint             # Run linter

# Database operations
make migrate          # Run migrations
make db-init          # Initialize database
make db-reset         # Reset database
make migration-create # Create new migration files

# Docker
make docker-build     # Build Docker image
make docker-run       # Run Docker container

# Documentation
make docs             # Generate API documentation
```

## Configuration

TryTraGo uses a YAML configuration file. You can specify the configuration file location using the `--config` flag:

```bash
./build/trytrago server --config=/path/to/config.yaml
```

Example configuration:

```yaml
server:
  port: 8080
  timeout: 30s

database:
  type: postgres
  host: localhost
  port: 5432
  name: trytrago
  user: postgres
  password: postgres

logging:
  level: info
  format: json
```

## Deployment

For detailed deployment instructions, please refer to [README_DEPLOY.md](README_DEPLOY.md).

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
