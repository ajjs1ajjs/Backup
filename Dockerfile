# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o novabackup-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nova-cli ./cmd/nova-cli

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 novabackup && \
    adduser -u 1000 -G novabackup -s /bin/sh -D novabackup

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/novabackup-server .
COPY --from=builder /app/nova-cli .

# Copy web assets
COPY --from=builder /app/gui/templates ./gui/templates
COPY --from=builder /app/gui/static ./gui/static

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/backups /app/config && \
    chown -R novabackup:novabackup /app

# Switch to non-root user
USER novabackup

# Expose ports
EXPOSE 8080
EXPOSE 8443

# Volume mounts
VOLUME ["/app/data", "/app/logs", "/app/backups", "/app/config"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the server
CMD ["./novabackup-server"]
