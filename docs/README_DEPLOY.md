# TryTraGo Deployment Guide

This document outlines the various deployment options for the TryTraGo multilanguage dictionary server, a robust application designed to handle approximately 60 million dictionary entries with comprehensive API functionality, social features, and multi-database support.

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Deployment Variants](#deployment-variants)
3. [Configuration Management](#configuration-management)
4. [Database Setup](#database-setup)
5. [Migration Management](#migration-management)
6. [Monitoring and Logging](#monitoring-and-logging)
7. [Backup and Restore](#backup-and-restore)
8. [Security Considerations](#security-considerations)
9. [Scaling Strategies](#scaling-strategies)
10. [Deployment Checklist](#deployment-checklist)

## System Requirements

### Minimum Requirements

- CPU: 2 cores
- RAM: 4GB
- Storage: 20GB SSD (excluding database storage)
- Operating System: Linux (Ubuntu 20.04+ or similar)
- Go 1.24 or higher (for manual build)
- Docker 20.10+ (for container deployments)
- Database: PostgreSQL 14+, MySQL 8.0+, or SQLite 3.35+ (for development only)
- Redis 6.0+ (for caching)

### Recommended Production Requirements

- CPU: 4+ cores
- RAM: 8GB+ (16GB+ for high traffic deployments)
- Storage: 50GB+ SSD (excluding database storage)
- Database storage: 500GB+ (for 60M entries)
- Network: 1Gbps

## Deployment Variants

### Docker-based Deployment

Docker is the recommended deployment method for most environments due to its consistency and ease of setup.

#### Single-Host Docker Compose Deployment

The project includes ready-to-use Docker Compose configurations for different scenarios:

1. **Development:** `docker/docker-compose-dev-pg.yml` or `docker/docker-compose-dev-ms.yml`
2. **Testing:** `docker/docker-compose-test.yml`
3. **Production/Demo:** `docker/docker-compose-demo.yml`

To deploy using Docker Compose:

```bash
# Clone the repository
git clone https://github.com/valpere/trytrago.git
cd trytrago

# Start the production deployment
docker-compose -f docker/docker-compose-demo.yml up -d
```

This will start:
- The TryTraGo application container
- PostgreSQL database
- Redis cache
- Adminer (for database management)

#### Custom Docker Deployment

You can build and deploy the TryTraGo container separately:

```bash
# Build the Docker image
docker build -t trytrago:latest -f docker/Dockerfile .

# Run the container
docker run -d --name trytrago \
  -p 8080:8080 \
  -v /path/to/config.yaml:/etc/trytrago/config.yaml \
  -e TRYTRAGO_ENVIRONMENT=production \
  -e TRYTRAGO_DATABASE_HOST=your-db-host \
  -e TRYTRAGO_DATABASE_USER=your-db-user \
  -e TRYTRAGO_DATABASE_PASSWORD=your-db-password \
  trytrago:latest
```

### Kubernetes Deployment

For large-scale and high-availability deployments, Kubernetes is recommended. Below is a basic setup for Kubernetes deployment.

#### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- Helm (optional, for managing deployments)

#### Deployment Steps

1. Create Kubernetes configuration files:

Create `k8s/trytrago-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trytrago
  labels:
    app: trytrago
spec:
  replicas: 3
  selector:
    matchLabels:
      app: trytrago
  template:
    metadata:
      labels:
        app: trytrago
    spec:
      containers:
      - name: trytrago
        image: trytrago:latest
        ports:
        - containerPort: 8080
        env:
        - name: TRYTRAGO_ENVIRONMENT
          value: "production"
        - name: TRYTRAGO_DATABASE_HOST
          valueFrom:
            secretKeyRef:
              name: trytrago-db-secrets
              key: host
        - name: TRYTRAGO_DATABASE_USER
          valueFrom:
            secretKeyRef:
              name: trytrago-db-secrets
              key: username
        - name: TRYTRAGO_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: trytrago-db-secrets
              key: password
        - name: TRYTRAGO_DATABASE_NAME
          value: "trytrago"
        - name: TRYTRAGO_CACHE_ADDRESS
          value: "redis-service:6379"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2"
```

Create `k8s/trytrago-service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: trytrago-service
spec:
  selector:
    app: trytrago
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

Create database secrets:

```bash
kubectl create secret generic trytrago-db-secrets \
  --from-literal=host=your-db-host \
  --from-literal=username=your-db-username \
  --from-literal=password=your-db-password
```

2. Apply Kubernetes configurations:

```bash
kubectl apply -f k8s/trytrago-deployment.yaml
kubectl apply -f k8s/trytrago-service.yaml
```

3. Scale as needed:

```bash
kubectl scale deployment trytrago --replicas=5
```

### Manual Deployment

For environments where containers are not an option, manual deployment can be performed.

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/valpere/trytrago.git
cd trytrago

# Install dependencies
go mod download

# Build the binary
CGO_ENABLED=0 GOOS=linux go build -o trytrago main.go

# Run initial migrations
./trytrago migrate --apply

# Start the server
./trytrago server --config=/path/to/config.yaml
```

#### Systemd Service Configuration

For long-running deployment on Linux servers, create a systemd service:

Create `/etc/systemd/system/trytrago.service`:

```ini
[Unit]
Description=TryTraGo Dictionary Server
After=network.target postgresql.service

[Service]
Type=simple
User=trytrago
WorkingDirectory=/opt/trytrago
ExecStart=/opt/trytrago/trytrago server --config=/opt/trytrago/config.yaml
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment=TRYTRAGO_ENVIRONMENT=production

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable trytrago
sudo systemctl start trytrago
```

## Configuration Management

### Configuration File

TryTraGo uses a YAML configuration file located at `/etc/trytrago/config.yaml` by default. You can specify an alternative location using the `--config` flag.

Example configuration:

```yaml
# Server configuration
server:
  port: 8080
  timeout: 30s
  read_timeout: 15s
  write_timeout: 15s

# Database configuration
database:
  type: postgres  # postgres, mysql, or sqlite
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: trytrago
  max_open_conns: 50
  max_idle_conns: 10
  conn_lifetime: 5m

# Logging configuration
logging:
  level: info  # debug, info, warn, error
  format: json  # json or console
  output: stdout
  file_path: ""

# Authentication configuration
auth:
  jwt_secret: your-strong-secret-key-here
  access_token_duration: 1h
  refresh_token_duration: 168h  # 7 days

# Cache configuration
cache:
  type: redis
  address: localhost:6379
  password: ""
  db: 0
  ttl: 10m

# Environment
environment: production
```

### Environment Variables

TryTraGo supports configuration via environment variables, which take precedence over configuration file values. Variables are prefixed with `TRYTRAGO_`.

Example:

```bash
export TRYTRAGO_SERVER_PORT=8080
export TRYTRAGO_DATABASE_HOST=postgres.example.com
export TRYTRAGO_DATABASE_USER=production_user
export TRYTRAGO_DATABASE_PASSWORD=secure_password
export TRYTRAGO_AUTH_JWT_SECRET=your-secure-jwt-secret
export TRYTRAGO_ENVIRONMENT=production
```

## Database Setup

TryTraGo supports multiple database backends. Below is the setup for each supported database.

### PostgreSQL (Recommended for Production)

```sql
-- Create database
CREATE DATABASE trytrago;

-- Create user
CREATE USER trytrago WITH PASSWORD 'your-password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE trytrago TO trytrago;

-- Connect to database
\c trytrago

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### MySQL

```sql
-- Create database
CREATE DATABASE trytrago CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER 'trytrago'@'%' IDENTIFIED BY 'your-password';

-- Grant privileges
GRANT ALL PRIVILEGES ON trytrago.* TO 'trytrago'@'%';
FLUSH PRIVILEGES;
```

### SQLite (Development/Testing Only)

SQLite doesn't require server setup - just ensure the application has write access to the database file location.

## Migration Management

TryTraGo includes a built-in migration system to manage database schema changes.

### Running Migrations

```bash
# Apply all pending migrations
./trytrago migrate --apply

# Check migration status
./trytrago migrate

# Rollback the most recent migration
./trytrago migrate --rollback
```

### Creating New Migrations

```bash
# Create a new migration
./trytrago migrate --create "add_user_preferences_table"
```

This generates two files:
- `migrations/V{timestamp}_add_user_preferences_table.sql` (forward migration)
- `migrations/R{timestamp}_rollback_add_user_preferences_table.sql` (rollback)

## Monitoring and Logging

### Logging

TryTraGo uses structured logging via Uber's zap logger. In production, JSON format is recommended for log aggregation:

```yaml
logging:
  level: info
  format: json
  output: stdout
```

### Health Checks

The application provides a `/health` endpoint that returns current status:

```
GET /health
```

Response:
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

### Metrics

For advanced monitoring, consider implementing Prometheus metrics:

1. Deploy Prometheus Server
2. Configure Prometheus to scrape metrics from TryTraGo
3. Create Grafana dashboards for visualization

## Backup and Restore

### Database Backup

For PostgreSQL:

```bash
# Using the provided script
./scripts/db-backup.sh

# Manual backup
pg_dump -U postgres -d trytrago | gzip > trytrago_backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

For MySQL:

```bash
mysqldump -u root -p --opt trytrago | gzip > trytrago_backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

### Database Restore

For PostgreSQL:

```bash
# Using the provided script
./scripts/db-restore.sh backups/trytrago_backup_20230101_120000.sql.gz

# Manual restore
gunzip -c trytrago_backup.sql.gz | psql -U postgres -d trytrago
```

For MySQL:

```bash
gunzip -c trytrago_backup.sql.gz | mysql -u root -p trytrago
```

## Security Considerations

### Authentication

- Use a strong, randomly generated `jwt_secret` for JWT token signing
- Regularly rotate JWT secrets
- Implement rate limiting for authentication endpoints
- Use HTTPS in production

### Network Security

- Place the application behind a reverse proxy (e.g., Nginx, Traefik)
- Use TLS/SSL for all connections
- Implement proper firewall rules
- Consider using a Web Application Firewall (WAF)

### Database Security

- Use strong passwords
- Restrict database user permissions
- Enable SSL for database connections
- Regularly update database credentials

## Scaling Strategies

### Horizontal Scaling

TryTraGo supports horizontal scaling for the application tier:

1. Deploy multiple instances behind a load balancer
2. Use Redis for distributed caching
3. Ensure database can handle the increased connection load

### Database Scaling

For large dictionaries (approaching 60M entries):

1. **Vertical Scaling**: Increase database server resources (CPU, RAM)
2. **Read Replicas**: Add read replicas for query distribution
3. **Sharding**: Consider database sharding for very large deployments
4. **Indexing**: Optimize indexes based on query patterns

### Caching Strategy

Effectively use Redis caching:

1. Cache frequently accessed entries
2. Cache translation results
3. Implement cache invalidation on updates
4. Configure appropriate TTL values based on data update frequency

## Deployment Checklist

### Pre-deployment

- [ ] Review and update configuration for target environment
- [ ] Set strong JWT secret and database credentials
- [ ] Verify database connection and schema
- [ ] Run and verify all migrations
- [ ] Set up Redis for caching
- [ ] Prepare monitoring and logging
- [ ] Set up proper firewall rules
- [ ] Configure TLS certificates
- [ ] Configure backup strategy

### Deployment

- [ ] Deploy the latest version
- [ ] Verify application health checks
- [ ] Verify database connections
- [ ] Verify Redis connections
- [ ] Check logs for any errors or warnings
- [ ] Test authentication flow
- [ ] Test core API functionality

### Post-deployment

- [ ] Monitor application performance
- [ ] Verify backup processes
- [ ] Document deployment configuration
- [ ] Update API documentation if needed
- [ ] Set up alerts for critical metrics
