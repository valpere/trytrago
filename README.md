# TryTraGo Dictionary Server

TryTraGo is a high-performance multilanguage dictionary server designed to support approximately 60 million entries. The system provides dual API support (REST and gRPC) and can connect to multiple database management systems (PostgreSQL, MySQL, SQLite) based on configuration.

## Features

- Multi-language dictionary with 60M+ entries support
- Clean architecture design with separation of concerns
- Dual API support (REST with Gin and gRPC)
- Multiple database backends (PostgreSQL, MySQL, SQLite)
- User management with JWT authentication
- Social features (comments, likes, user feeds)
- High-performance caching with Redis
- Comprehensive error handling
- Structured logging with Uber-zap
- Containerized deployment with Docker

## Project Structure

The project follows clean architecture principles with clear separation between layers:

```
trytrago/
├── application/          # Application layer (use cases, DTOs)
│   ├── dto/              # Data Transfer Objects
│   │   ├── request/      # Request DTOs
│   │   └── response/     # Response DTOs
│   ├── mapper/           # Object mappers
│   └── service/          # Application services
├── domain/               # Domain layer (core business logic)
│   ├── database/         # Database entities and repositories
│   ├── logging/          # Logging infrastructure
│   └── model/            # Domain models
├── infrastructure/       # Infrastructure concerns
│   ├── auth/             # Authentication
│   ├── cache/            # Caching
│   └── migration/        # Database migrations
├── interface/            # Interface layer (API controllers)
│   ├── api/              # API implementations
│   │   ├── grpc/         # gRPC API
│   │   └── rest/         # REST API with Gin
│   └── server/           # Server initialization
└── cmd/                  # Command-line interface
```

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose (for containerized setup)
- PostgreSQL, MySQL, or SQLite (depending on your configuration)

### Running Locally

1. Clone the repository
   ```bash
   git clone https://github.com/valpere/trytrago.git
   cd trytrago
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Build the project
   ```bash
   make build
   ```

4. Run the server
   ```bash
   ./build/trytrago server
   ```

### Docker Setup

To run the entire stack with Docker Compose:

```bash
docker-compose up -d
```

This will start:
- TryTraGo server
- PostgreSQL database
- Redis cache
- pgAdmin for database management

## API Documentation

### REST API

The REST API is available at `http://localhost:8080/api/v1` with the following endpoints:

- **Authentication**
  - `POST /auth/login` - Authenticate a user
  - `POST /auth/refresh` - Refresh an access token

- **Entries**
  - `GET /entries` - List entries
  - `GET /entries/:id` - Get an entry by ID
  - `POST /entries` - Create a new entry
  - `PUT /entries/:id` - Update an entry
  - `DELETE /entries/:id` - Delete an entry

- **Meanings**
  - `GET /entries/:id/meanings` - List meanings for an entry
  - `POST /entries/:id/meanings` - Add a meaning to an entry
  - `PUT /entries/:id/meanings/:meaningId` - Update a meaning
  - `DELETE /entries/:id/meanings/:meaningId` - Delete a meaning

- **Translations**
  - `GET /entries/:id/meanings/:meaningId/translations` - List translations
  - `POST /entries/:id/meanings/:meaningId/translations` - Add a translation
  - `PUT /entries/:id/meanings/:meaningId/translations/:translationId` - Update a translation
  - `DELETE /entries/:id/meanings/:meaningId/translations/:translationId` - Delete a translation

### gRPC API

The gRPC API is available at `localhost:9090` with the following services:

- `DictionaryService` - Operations for dictionary entries, meanings, and translations
- `UserService` - User management and authentication

## Configuration

Configuration can be provided via:
- Environment variables (prefixed with `TRYTRAGO_`)
- Configuration file (yaml or json)
- Command-line flags

Example configuration:

```yaml
environment: development
verbose: true

server:
  http_port: 8080
  grpc_port: 9090
  rate_limit:
    requests_per_second: 100
    burst_size: 20

database:
  type: postgres
  host: localhost
  port: 5432
  name: trytrago
  user: postgres
  password: postgres
  pool_size: 20
  max_idle_conns: 10
  max_open_conns: 100
  conn_timeout: 30s

logging:
  level: debug
  format: json
  file_path: logs/trytrago.log

auth:
  jwt_secret: your_secret_key
  token_expiration: 24h
```

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
