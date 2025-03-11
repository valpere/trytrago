# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc libc-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/build/trytrago /app/trytrago

# Copy configuration (if needed)
COPY --from=builder /app/config.yaml /app/config.yaml

# Expose ports
EXPOSE 8080 9090

# Set environment variables
ENV TRYTRAGO_ENVIRONMENT=production

# Run the application
ENTRYPOINT ["/app/trytrago"]
CMD ["server"]
