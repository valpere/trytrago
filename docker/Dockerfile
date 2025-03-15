# Builder stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/valpere/trytrago/domain.Version=${VERSION} \
              -X github.com/valpere/trytrago/domain.CommitSHA=${COMMIT_SHA} \
              -X github.com/valpere/trytrago/domain.BuildTime=${BUILD_TIME}" \
    -o /trytrago .

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set up non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy the binary from builder
COPY --from=builder /trytrago /usr/local/bin/trytrago

# Copy config files
COPY --from=builder /app/config.yaml /etc/trytrago/config.yaml
COPY --from=builder /app/migrations /etc/trytrago/migrations

# Create data directory with appropriate permissions
USER root
RUN mkdir -p /data/trytrago && chown -R appuser:appgroup /data/trytrago
USER appuser

# Set working directory
WORKDIR /home/appuser

# Expose ports
EXPOSE 8080

# Set environment variables
ENV TRYTRAGO_CONFIG=/etc/trytrago/config.yaml
ENV TRYTRAGO_ENVIRONMENT=production

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["trytrago"]
CMD ["server"]
