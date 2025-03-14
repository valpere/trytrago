# TryTraGo: Multilanguage Dictionary Server

!!!!!!!!!!!!!!!!!!!
!!! In Progress !!!
!!!!!!!!!!!!!!!!!!!

TryTraGo is a robust, high-performance dictionary server designed to support approximately 60 million dictionary entries across multiple languages. Built with Go and following clean architecture principles, TryTraGo offers a comprehensive API for managing dictionary entries, meanings, translations, and social interactions.

## Features

- **Comprehensive Dictionary Management**: Create, read, update, and delete dictionary entries, meanings, and translations.
- **Multilanguage Support**: Handle translations between any language pairs with ISO language code support.
- **Social Features**: Comment on and like meanings and translations.
- **User Management**: Authentication, authorization, and user profile management.
- **Clean Architecture**: Clear separation of concerns across domain, application, interface, and infrastructure layers.
- **Multiple Database Support**: Works with PostgreSQL (recommended for production), MySQL, and SQLite (for development).
- **High Performance**: Optimized for handling 60 million dictionary entries with proper indexing and caching strategies.
- **REST API**: Comprehensive HTTP API with proper resource modeling.

## Architecture

TryTraGo follows clean architecture principles with clear separation of concerns:

1. **Domain Layer**: Core business entities and interfaces
   - Database models (Entry, Meaning, Translation, etc.)
   - Repository interfaces
   - Domain-specific errors and validation logic

2. **Application Layer**: Use cases and business logic
   - Service implementations for entries, translations, and users
   - DTOs for communication between layers
   - Mapping logic between domain entities and DTOs

3. **Interface Layer**: API implementations
   - REST API using the Gin framework
   - Request handlers and middleware
   - Response formatting and error handling

4. **Infrastructure Layer**: External implementations
   - Database implementations for PostgreSQL, MySQL, and SQLite
   - Authentication with JWT
   - Caching with Redis
   - Migration tooling

## Technology Stack

- **Go (Golang)**: version 1.24 or higher
- **Gin**: Web framework
- **GORM**: ORM with support for multiple databases
- **Uber-zap**: High-performance, structured logging
- **JWT**: For authentication
- **Redis**: For caching
- **Cobra + Viper**: CLI framework with configuration management
- **Docker**: For containerization and local development

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose (for local development)
- Database (PostgreSQL, MySQL, or SQLite)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/valpere/trytrago.git
   cd trytrago
   ```

2. Build the project:

   ```bash
   make build
   ```

3. Run the server:

   ```bash
   ./build/trytrago server
   ```

### Using Docker

For local development with Docker:

```bash
docker-compose up
```

This will start the server with PostgreSQL and Redis.

## API Endpoints

The API follows REST principles for resource management.

### Public Endpoints

- `GET /api/v1/entries`: List dictionary entries
- `GET /api/v1/entries/:id`: Get a specific entry
- `GET /api/v1/entries/:id/meanings`: List meanings for an entry
- `GET /api/v1/entries/:entryId/meanings/:meaningId`: Get a specific meaning
- `GET /api/v1/entries/:entryId/meanings/:meaningId/translations`: List translations for a meaning

### User Authentication

- `POST /api/v1/auth/register`: Create a new user
- `POST /api/v1/auth/login`: Authenticate a user
- `POST /api/v1/auth/refresh`: Refresh an authentication token

### Protected Endpoints

All protected endpoints require authentication:

- `POST /api/v1/entries`: Create a new entry
- `PUT /api/v1/entries/:id`: Update an entry
- `DELETE /api/v1/entries/:id`: Delete an entry
- `POST /api/v1/entries/:entryId/meanings`: Add a meaning to an entry
- `POST /api/v1/entries/:entryId/meanings/:meaningId/translations`: Add a translation to a meaning
- `POST /api/v1/entries/:entryId/meanings/:meaningId/comments`: Comment on a meaning
- `POST /api/v1/entries/:entryId/meanings/:meaningId/likes`: Like a meaning

## Configuration

TryTraGo can be configured via:

1. Configuration file (YAML)
2. Environment variables
3. Command-line flags

Example configuration:

```yaml
server:
  port: 8080
  timeout: 30s

database:
  type: postgres  # postgres, mysql, or sqlite
  host: localhost
  port: 5432
  name: trytrago
  user: postgres
  password: postgres

logging:
  level: info  # debug, info, warn, error
  format: json  # json or console

auth:
  jwt_secret: your-secret-key-change-this-in-production
  access_token_duration: 1h
  refresh_token_duration: 7d
```

## CLI Commands

TryTraGo comes with a powerful CLI:

- `trytrago server`: Start the HTTP server
- `trytrago migrate`: Manage database migrations
- `trytrago backup`: Backup dictionary content
- `trytrago restore`: Restore dictionary content
- `trytrago config`: Display current configuration
- `trytrago version`: Display version information

## Development

### Project Structure

```plaintext
trytrago/
├── cmd/                   # CLI commands
├── domain/                # Domain layer
│   ├── database/          # Database models and repositories
│   ├── logging/           # Logging infrastructure
│   └── model/             # Domain models
├── application/           # Application layer
│   ├── dto/               # Data Transfer Objects
│   ├── mapper/            # Object mappers
│   └── service/           # Application services
├── interface/             # Interface layer
│   ├── api/               # API implementations
│   │   └── rest/          # REST API
│   └── server/            # Server implementations
├── infrastructure/        # Infrastructure layer
│   ├── auth/              # Authentication
│   ├── cache/             # Caching
│   └── migration/         # Database migrations
├── migrations/            # SQL migration files
└── scripts/               # Utility scripts
```

### Testing

Run the tests with:

```bash
make test
```

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- Built with love by the TryTraGo team
- Special thanks to all contributors

---

For detailed documentation, please refer to the [docs](./docs) directory.
