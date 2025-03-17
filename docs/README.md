# TryTraGo API Documentation

This directory contains the API documentation for the TryTraGo multilanguage dictionary server.

## Contents

- `APIDocumentation.md` - Comprehensive API documentation in Markdown format
- `OpenAPISpecification.yaml` - OpenAPI 3.0 specification for the REST API

## Interactive API Documentation

When the server is running, you can access the interactive API documentation at:

- Swagger UI: [http://localhost:8080/swagger-ui/](http://localhost:8080/swagger-ui/)
- OpenAPI Specification: [http://localhost:8080/v3/api-docs](http://localhost:8080/v3/api-docs)

## Setting Up Swagger UI

To set up the Swagger UI for local development, run:

```bash
make swagger-setup
```

This command will download and configure the Swagger UI files needed for the interactive documentation.

## API Features

TryTraGo provides a comprehensive REST API for:

- Dictionary entries management
- Meanings and translations
- User authentication and profile management
- Social features (comments, likes)
- Admin operations

For detailed information on all endpoints, request/response formats, and authentication requirements, please see the `APIDocumentation.md` file or use the interactive Swagger UI.

## Authentication

The API uses JWT-based authentication. To access protected endpoints:

1. Register a user or log in to obtain an access token
2. Include the token in the Authorization header: `Authorization: Bearer <token>`

## Rate Limiting

The API implements rate limiting to prevent abuse. Clients are limited to 10 requests per second with a burst capacity of 20 requests.
