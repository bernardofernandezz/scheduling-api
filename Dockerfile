# Build stage
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run tests (will be skipped in CI as tests run separately)
RUN if [ ! -f /.dockerenv ]; then \
    go test -v ./...; \
    fi

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
    -X main.version=$(git describe --tags --always) \
    -X main.commit=$(git rev-parse HEAD) \
    -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o scheduling-api \
    ./cmd/api

# Final stage
FROM alpine:3.18

# Install runtime dependencies and security updates
RUN apk update && \
    apk add --no-cache \
    ca-certificates \
    tzdata \
    curl && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D appuser

# Set working directory
WORKDIR /app

# Copy binary and config files
COPY --from=builder /app/scheduling-api .
COPY --from=builder /app/.env.example ./.env

# Set ownership
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default environment variables
ENV GIN_MODE=release \
    SERVER_ADDRESS=:8080 \
    LOG_LEVEL=info \
    RATE_LIMIT_REQUESTS=60 \
    RATE_LIMIT_DURATION=1m \
    CORS_ALLOWED_ORIGINS=* \
    JWT_EXPIRE_HOURS=24

# Run the application
CMD ["./scheduling-api"]

