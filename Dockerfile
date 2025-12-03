# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN make build

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' comio

# Copy binary from builder
COPY --from=builder /app/bin/comio /app/comio

# Copy default config
COPY configs/config.yaml.example /etc/comio/config.yaml

# Set ownership
RUN chown -R comio:comio /app /etc/comio

# Switch to non-root user
USER comio

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/admin/health || exit 1

# Run binary
ENTRYPOINT ["/app/comio"]
CMD ["server", "start"]
