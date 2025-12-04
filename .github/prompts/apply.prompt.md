---
agent: 'agent'
model: 'Gemini 3 Pro (Preview)'
description: 'Generate production-ready golang project with proper structure, Makefile, Dockerfile, CI/CD pipeline, and unit tests based on user requirements.'
---

# ComIO - Community IO Storage Implementation Plan

## Overview

ComIO is a production-ready S3-compliant storage solution in Golang featuring RESTful API, CLI management, storage replication, raw device handling, and authentication.

---

## 1. Project Structure

Create the following directory structure:

```
comio/
├── cmd/
│   └── comio/
│       └── main.go                 # Entry point, minimal logic
├── internal/
│   ├── api/
│   │   ├── server.go               # HTTP server setup
│   │   ├── router.go               # Route definitions
│   │   ├── middleware/
│   │   │   ├── auth.go             # Authentication middleware
│   │   │   ├── logging.go          # Request logging middleware
│   │   │   └── recovery.go         # Panic recovery middleware
│   │   └── handlers/
│   │       ├── bucket.go           # Bucket operations handlers
│   │       ├── object.go           # Object operations handlers
│   │       ├── multipart.go        # Multipart upload handlers
│   │       ├── lifecycle.go        # Lifecycle policy handlers
│   │       └── health.go           # Health check handlers
│   ├── cli/
│   │   ├── root.go                 # Root cobra command
│   │   ├── server.go               # Server start command
│   │   ├── bucket.go               # Bucket management commands
│   │   ├── object.go               # Object management commands
│   │   ├── admin.go                # Admin operation commands
│   │   └── config.go               # Configuration commands
│   ├── storage/
│   │   ├── engine.go               # Storage engine interface
│   │   ├── device.go               # Raw device handling (/dev/sdX)
│   │   ├── partition.go            # Partition handling
│   │   ├── block.go                # Block-level operations
│   │   └── allocator.go            # Space allocation
│   ├── bucket/
│   │   ├── bucket.go               # Bucket domain logic
│   │   ├── repository.go           # Bucket persistence interface
│   │   └── service.go              # Bucket service layer
│   ├── object/
│   │   ├── object.go               # Object domain logic
│   │   ├── version.go              # Object versioning logic
│   │   ├── repository.go           # Object persistence interface
│   │   └── service.go              # Object service layer
│   ├── multipart/
│   │   ├── upload.go               # Multipart upload logic
│   │   ├── part.go                 # Part handling
│   │   └── service.go              # Multipart service layer
│   ├── lifecycle/
│   │   ├── policy.go               # Lifecycle policy definitions
│   │   ├── executor.go             # Policy execution engine
│   │   └── service.go              # Lifecycle service layer
│   ├── replication/
│   │   ├── manager.go              # Replication manager
│   │   ├── node.go                 # Node representation
│   │   ├── sync.go                 # Data synchronization
│   │   └── consensus.go            # Consistency handling
│   ├── auth/
│   │   ├── authenticator.go        # Authentication interface
│   │   ├── hmac.go                 # HMAC signature verification (S3-style)
│   │   ├── user.go                 # User management
│   │   └── policy.go               # Authorization policies
│   ├── integrity/
│   │   ├── checksum.go             # Checksum calculation
│   │   └── validator.go            # Data integrity validation
│   ├── backup/
│   │   ├── backup.go               # Backup operations
│   │   └── restore.go              # Restore operations
│   ├── monitoring/
│   │   ├── metrics.go              # Prometheus metrics
│   │   └── logger.go               # Structured logging setup
│   └── config/
│       ├── config.go               # Configuration struct
│       └── loader.go               # Configuration loading
├── pkg/
│   ├── s3/
│   │   ├── types.go                # S3 API types
│   │   ├── errors.go               # S3 error codes
│   │   └── signature.go            # S3 signature handling
│   └── utils/
│       ├── hash.go                 # Hashing utilities
│       └── time.go                 # Time utilities
├── api/
│   └── openapi.yaml                # OpenAPI specification
├── configs/
│   ├── config.yaml.example         # Example configuration
│   └── config.yaml                 # Default configuration
├── scripts/
│   └── setup-device.sh             # Device setup helper script
├── deployments/
│   ├── docker-compose.yaml         # Docker compose for local dev
│   └── kubernetes/
│       ├── deployment.yaml
│       └── service.yaml
├── docs/
│   ├── README.md                   # Main documentation
│   ├── API.md                      # API documentation
│   ├── CLI.md                      # CLI documentation
│   └── ARCHITECTURE.md             # Architecture documentation
├── test/
│   └── integration/
│       └── api_test.go             # Integration tests
├── .github/
│   └── workflows/
│       ├── ci.yaml                 # CI pipeline
│       └── release.yaml            # Release pipeline
├── Makefile
├── Dockerfile
├── .dockerignore
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## 2. Implementation Tasks

### 2.1 Initialize Project

1. Create `go.mod` with module path `github.com/danielino/comio` and Go version 1.25
2. Create `.gitignore` for Go projects (binaries, vendor, IDE files)
3. Create `.dockerignore` to exclude unnecessary files from Docker builds

### 2.2 Configuration Management

Implement in `internal/config/`:

1. **config.go**: Define configuration struct with fields:
   - `Server`: Host, Port, TLS settings
   - `Storage`: Device paths, allocation settings
   - `Replication`: Node list, consistency level
   - `Auth`: Secret key, token expiration
   - `Logging`: Level, format, output
   - `Metrics`: Enabled, endpoint

2. **loader.go**: Implement configuration loading:
   - Load from YAML file (default: `/etc/comio/config.yaml`)
   - Override with environment variables (prefix: `COMIO_`)
   - Validate configuration on load
   - Use viper for configuration management

### 2.3 Structured Logging

Implement in `internal/monitoring/logger.go`:

1. Use `zap` for structured logging
2. Configure log levels: DEBUG, INFO, WARN, ERROR
3. Include request ID, timestamp, caller info in logs
4. Support JSON and console output formats

### 2.4 Storage Engine

Implement in `internal/storage/`:

1. **engine.go**: Define storage engine interface:
   ```go
   type Engine interface {
       Open(devicePath string) error
       Close() error
       Read(offset, size int64) ([]byte, error)
       Write(offset int64, data []byte) error
       Allocate(size int64) (offset int64, err error)
       Free(offset, size int64) error
       Sync() error
   }
   ```

2. **device.go**: Implement raw device handling:
   - Open device files (`/dev/sdX`) with O_DIRECT flag
   - Handle device metadata (size, block size)
   - Implement sector-aligned read/write operations
   - Support both full disks and partitions

3. **block.go**: Implement block-level operations:
   - Define block size (default: 4KB)
   - Implement block allocation bitmap
   - Handle block read/write with checksums

4. **allocator.go**: Implement space allocation:
   - First-fit allocation strategy
   - Free space tracking with bitmap
   - Defragmentation support (background task)

### 2.5 Bucket Management

Implement in `internal/bucket/`:

1. **bucket.go**: Define bucket domain model:
   ```go
   type Bucket struct {
       Name        string
       CreatedAt   time.Time
       Owner       string
       Versioning  VersioningStatus // Enabled, Suspended, Disabled
       Lifecycle   []LifecycleRule
   }
   ```

2. **repository.go**: Define persistence interface:
   ```go
   type Repository interface {
       Create(ctx context.Context, bucket *Bucket) error
       Get(ctx context.Context, name string) (*Bucket, error)
       List(ctx context.Context, owner string) ([]*Bucket, error)
       Delete(ctx context.Context, name string) error
       Update(ctx context.Context, bucket *Bucket) error
   }
   ```

3. **service.go**: Implement bucket service:
   - Validate bucket names (S3 naming rules)
   - Handle bucket creation with owner assignment
   - Implement bucket listing with pagination
   - Handle bucket deletion (check if empty)

### 2.6 Object Management

Implement in `internal/object/`:

1. **object.go**: Define object domain model:
   ```go
   type Object struct {
       Key          string
       BucketName   string
       VersionID    string
       Size         int64
       ContentType  string
       ETag         string
       Checksum     Checksum
       CreatedAt    time.Time
       ModifiedAt   time.Time
       Metadata     map[string]string
       StorageClass string
       DeleteMarker bool
   }

   type Checksum struct {
       Algorithm string // MD5, SHA256, CRC32
       Value     string
   }
   ```

2. **version.go**: Implement versioning:
   - Generate version IDs (UUID v4)
   - Track version history per object
   - Handle delete markers for versioned objects
   - Support listing object versions

3. **repository.go**: Define persistence interface:
   ```go
   type Repository interface {
       Put(ctx context.Context, obj *Object, data io.Reader) error
       Get(ctx context.Context, bucket, key string, versionID *string) (*Object, io.ReadCloser, error)
       Delete(ctx context.Context, bucket, key string, versionID *string) error
       List(ctx context.Context, bucket, prefix string, opts ListOptions) (*ListResult, error)
       Head(ctx context.Context, bucket, key string, versionID *string) (*Object, error)
   }
   ```

4. **service.go**: Implement object service:
   - Stream object data to storage engine
   - Calculate checksums during upload
   - Handle conditional requests (If-Match, If-None-Match)
   - Implement copy operations

### 2.7 Multipart Upload

Implement in `internal/multipart/`:

1. **upload.go**: Define multipart upload model:
   ```go
   type Upload struct {
       UploadID    string
       BucketName  string
       Key         string
       CreatedAt   time.Time
       Parts       []Part
   }
   ```

2. **part.go**: Define part model:
   ```go
   type Part struct {
       PartNumber int
       ETag       string
       Size       int64
       Checksum   string
   }
   ```

3. **service.go**: Implement multipart service:
   - Initiate multipart upload (generate upload ID)
   - Upload part (validate part number 1-10000)
   - List parts for an upload
   - Complete multipart upload (merge parts)
   - Abort multipart upload (cleanup parts)

### 2.8 Lifecycle Policies

Implement in `internal/lifecycle/`:

1. **policy.go**: Define lifecycle policy:
   ```go
   type Rule struct {
       ID                   string
       Status               string // Enabled, Disabled
       Filter               Filter
       Transitions          []Transition
       Expiration           *Expiration
       NoncurrentVersions   *NoncurrentVersionExpiration
   }
   ```

2. **executor.go**: Implement policy executor:
   - Run as background goroutine
   - Evaluate rules periodically (configurable interval)
   - Handle transitions between storage classes
   - Delete expired objects
   - Clean up old versions

### 2.9 Replication

Implement in `internal/replication/`:

1. **manager.go**: Implement replication manager:
   - Track cluster nodes
   - Handle node join/leave events
   - Distribute data across nodes
   - Monitor replication health

2. **node.go**: Define node representation:
   ```go
   type Node struct {
       ID       string
       Address  string
       Status   NodeStatus
       Capacity int64
       Used     int64
   }
   ```

3. **sync.go**: Implement data synchronization:
   - Replicate writes to configured number of nodes
   - Handle write quorum (configurable)
   - Implement anti-entropy repair
   - Support async replication for performance

4. **consensus.go**: Implement consistency handling:
   - Vector clocks for conflict detection
   - Last-write-wins conflict resolution
   - Read repair on inconsistency detection

### 2.10 Authentication & Authorization

Implement in `internal/auth/`:

1. **authenticator.go**: Define auth interface:
   ```go
   type Authenticator interface {
       Authenticate(ctx context.Context, req *http.Request) (*User, error)
       ValidateSignature(req *http.Request, secretKey string) error
   }
   ```

2. **hmac.go**: Implement S3 signature verification:
   - Support AWS Signature Version 4
   - Parse Authorization header
   - Calculate canonical request
   - Verify HMAC-SHA256 signature

3. **user.go**: Implement user management:
   ```go
   type User struct {
       AccessKeyID     string
       SecretAccessKey string
       Username        string
       Policies        []string
       CreatedAt       time.Time
   }
   ```

4. **policy.go**: Implement authorization:
   - Define IAM-like policies
   - Support bucket-level and object-level permissions
   - Actions: GetObject, PutObject, DeleteObject, ListBucket, etc.

### 2.11 Data Integrity

Implement in `internal/integrity/`:

1. **checksum.go**: Implement checksum calculation:
   - MD5 for ETag compatibility
   - SHA256 for content verification
   - CRC32C for fast integrity checks
   - Calculate checksums during streaming

2. **validator.go**: Implement validation:
   - Verify checksums on read
   - Support client-provided checksums
   - Log and alert on corruption detection

### 2.12 Backup & Restore

Implement in `internal/backup/`:

1. **backup.go**: Implement backup operations:
   - Full backup to external storage
   - Incremental backup based on modification time
   - Compress backup data (gzip)
   - Encrypt backup data (AES-256)

2. **restore.go**: Implement restore operations:
   - Restore from backup file
   - Point-in-time recovery
   - Verify backup integrity before restore

### 2.13 HTTP API Server

Implement in `internal/api/`:

1. **server.go**: Implement HTTP server:
   - Use Gin framework
   - Configure graceful shutdown
   - Support TLS/HTTPS
   - Set appropriate timeouts

2. **router.go**: Define routes (S3-compatible):
   ```
   # Service operations
   GET /                                    # ListBuckets

   # Bucket operations
   PUT /{bucket}                            # CreateBucket
   DELETE /{bucket}                         # DeleteBucket
   GET /{bucket}                            # ListObjects
   HEAD /{bucket}                           # HeadBucket
   GET /{bucket}?versioning                 # GetBucketVersioning
   PUT /{bucket}?versioning                 # PutBucketVersioning
   GET /{bucket}?lifecycle                  # GetBucketLifecycle
   PUT /{bucket}?lifecycle                  # PutBucketLifecycle

   # Object operations
   PUT /{bucket}/{key}                      # PutObject
   GET /{bucket}/{key}                      # GetObject
   DELETE /{bucket}/{key}                   # DeleteObject
   HEAD /{bucket}/{key}                     # HeadObject
   POST /{bucket}/{key}?uploads             # InitiateMultipartUpload
   PUT /{bucket}/{key}?partNumber&uploadId  # UploadPart
   POST /{bucket}/{key}?uploadId            # CompleteMultipartUpload
   DELETE /{bucket}/{key}?uploadId          # AbortMultipartUpload
   GET /{bucket}/{key}?uploadId             # ListParts

   # Admin endpoints (non-S3)
   GET /admin/health                        # Health check
   GET /admin/metrics                       # Prometheus metrics
   GET /admin/nodes                         # List cluster nodes
   POST /admin/users                        # Create user
   ```

3. **middleware/**: Implement middleware:
   - `auth.go`: Validate signatures, extract user
   - `logging.go`: Log request/response with zap
   - `recovery.go`: Recover from panics, return 500

4. **handlers/**: Implement handlers:
   - Parse S3 XML/JSON request bodies
   - Return S3-compatible responses
   - Handle errors with S3 error codes
   - Use service layer for business logic

### 2.14 CLI Implementation

Implement in `internal/cli/`:

1. **root.go**: Root command:
   ```
   comio - Community IO Storage

   Usage:
     comio [command]

   Commands:
     server    Start the ComIO server
     bucket    Bucket management commands
     object    Object management commands
     admin     Administrative commands
     config    Configuration commands
   ```

2. **server.go**: Server command:
   ```
   comio server start [--config path] [--port port]
   comio server status
   comio server stop
   ```

3. **bucket.go**: Bucket commands:
   ```
   comio bucket create <name>
   comio bucket delete <name>
   comio bucket list
   comio bucket info <name>
   ```

4. **object.go**: Object commands:
   ```
   comio object put <bucket> <key> <file>
   comio object get <bucket> <key> [--output file]
   comio object delete <bucket> <key>
   comio object list <bucket> [--prefix prefix]
   ```

5. **admin.go**: Admin commands:
   ```
   comio admin user create <username>
   comio admin user delete <username>
   comio admin user list
   comio admin node list
   comio admin backup create [--output path]
   comio admin backup restore <path>
   ```

### 2.15 Metrics & Monitoring

Implement in `internal/monitoring/`:

1. **metrics.go**: Prometheus metrics:
   - `comio_requests_total` - Counter by method, bucket, status
   - `comio_request_duration_seconds` - Histogram
   - `comio_storage_bytes_total` - Gauge
   - `comio_objects_total` - Gauge per bucket
   - `comio_replication_lag_seconds` - Gauge per node

---

## 3. Build & Deployment

### 3.1 Makefile

Create `Makefile` with targets:

```makefile
# Variables
BINARY_NAME=comio
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Default target
.PHONY: all
all: build

# Build binary
.PHONY: build
build:
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/comio

# Run tests
.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
.PHONY: coverage
coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Run linter
.PHONY: lint
lint:
	golangci-lint run ./...

# Format code
.PHONY: fmt
fmt:
	gofmt -s -w .
	goimports -w .

# Vet code
.PHONY: vet
vet:
	go vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/ coverage.out coverage.html

# Build Docker image
.PHONY: docker-build
docker-build:
	docker build -t comio:${VERSION} .

# Run Docker container
.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 comio:${VERSION}

# Generate mocks for testing
.PHONY: mocks
mocks:
	mockgen -source=internal/storage/engine.go -destination=internal/storage/mock_engine.go -package=storage
	mockgen -source=internal/bucket/repository.go -destination=internal/bucket/mock_repository.go -package=bucket
	mockgen -source=internal/object/repository.go -destination=internal/object/mock_repository.go -package=object

# Install development dependencies
.PHONY: deps
deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang/mock/mockgen@latest

# Run the application
.PHONY: run
run: build
	./bin/${BINARY_NAME} server start

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run tests"
	@echo "  coverage    - Generate coverage report"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-build- Build Docker image"
	@echo "  docker-run  - Run Docker container"
	@echo "  mocks       - Generate mocks"
	@echo "  deps        - Install dev dependencies"
	@echo "  run         - Build and run"
```

### 3.2 Dockerfile

Create multi-stage `Dockerfile`:

```dockerfile
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
```

### 3.3 CI/CD Pipeline

Create `.github/workflows/ci.yaml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install dependencies
        run: go mod download

      - name: Run vet
        run: go vet ./...

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage is below 80%: $COVERAGE%"
            exit 1
          fi

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Build
        run: make build

      - name: Build Docker image
        run: make docker-build
```

Create `.github/workflows/release.yaml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            danielino/comio:${{ github.ref_name }}
            danielino/comio:latest
```

---

## 4. Dependencies

Add these dependencies to `go.mod`:

```go
require (
    github.com/gin-gonic/gin v1.9.1           // HTTP framework
    github.com/spf13/cobra v1.8.0             // CLI framework
    github.com/spf13/viper v1.18.2            // Configuration
    go.uber.org/zap v1.26.0                   // Structured logging
    github.com/prometheus/client_golang v1.18.0  // Metrics
    github.com/google/uuid v1.5.0             // UUID generation
    github.com/stretchr/testify v1.8.4        // Testing assertions
    github.com/golang/mock v1.6.0             // Mock generation
    golang.org/x/sys v0.15.0                  // System calls for device access
)
```

---

## 5. Testing Strategy

### 5.1 Unit Tests

Create tests for each package achieving minimum 80% coverage:

1. **Storage tests** (`internal/storage/*_test.go`):
   - Test block allocation/deallocation
   - Test read/write operations
   - Test checksum verification
   - Mock device for testing

2. **Bucket tests** (`internal/bucket/*_test.go`):
   - Test bucket name validation
   - Test CRUD operations
   - Test listing with pagination

3. **Object tests** (`internal/object/*_test.go`):
   - Test object put/get/delete
   - Test versioning logic
   - Test conditional operations

4. **API tests** (`internal/api/handlers/*_test.go`):
   - Test each handler with httptest
   - Test error responses
   - Test authentication middleware

5. **CLI tests** (`internal/cli/*_test.go`):
   - Test command parsing
   - Test flag validation

### 5.2 Integration Tests

Create integration tests in `test/integration/`:

1. Test full API flows
2. Test multipart upload end-to-end
3. Test replication between nodes
4. Use testcontainers for realistic testing

---

## 6. Configuration Example

Create `configs/config.yaml.example`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

storage:
  devices:
    - path: "/dev/sdb"
      type: "disk"
    - path: "/dev/sdc1"
      type: "partition"
  block_size: 4096
  replication_factor: 3

replication:
  nodes:
    - address: "node1:8080"
    - address: "node2:8080"
    - address: "node3:8080"
  write_quorum: 2
  read_quorum: 1
  sync_interval: 5m

auth:
  enabled: true
  admin_access_key: "admin"
  admin_secret_key: "change-me-in-production"

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  endpoint: "/admin/metrics"

lifecycle:
  evaluation_interval: 24h
```

---

## 7. Implementation Order

Execute tasks in this order for logical dependency flow:

1. **Phase 1 - Foundation**:
   - Initialize project structure and go.mod
   - Implement configuration loading
   - Set up structured logging
   - Create main.go with CLI root command

2. **Phase 2 - Core Storage**:
   - Implement storage engine interface
   - Implement raw device handling
   - Implement block allocation
   - Add integrity/checksum validation

3. **Phase 3 - Domain Logic**:
   - Implement bucket management
   - Implement object management
   - Implement object versioning
   - Add multipart upload support

4. **Phase 4 - API Layer**:
   - Set up HTTP server with Gin
   - Implement authentication middleware
   - Implement S3-compatible handlers
   - Add error handling with S3 error codes

5. **Phase 5 - Advanced Features**:
   - Implement lifecycle policies
   - Implement replication
   - Add backup/restore functionality
   - Set up metrics and monitoring

6. **Phase 6 - CLI & Operations**:
   - Implement all CLI commands
   - Create Makefile
   - Create Dockerfile
   - Set up CI/CD pipelines

7. **Phase 7 - Testing & Documentation**:
   - Write unit tests (80%+ coverage)
   - Write integration tests
   - Create API documentation
   - Create user documentation

---

## 8. S3 Error Codes Reference

Implement these S3-compatible error codes in `pkg/s3/errors.go`:

| Code | HTTP Status | Description |
|------|-------------|-------------|
| AccessDenied | 403 | Access denied |
| BucketAlreadyExists | 409 | Bucket already exists |
| BucketNotEmpty | 409 | Bucket is not empty |
| EntityTooLarge | 400 | Entity too large |
| EntityTooSmall | 400 | Entity too small |
| InternalError | 500 | Internal server error |
| InvalidArgument | 400 | Invalid argument |
| InvalidBucketName | 400 | Invalid bucket name |
| InvalidPart | 400 | Invalid part |
| InvalidPartOrder | 400 | Invalid part order |
| InvalidRange | 416 | Invalid range |
| NoSuchBucket | 404 | Bucket not found |
| NoSuchKey | 404 | Object not found |
| NoSuchUpload | 404 | Upload not found |
| NoSuchVersion | 404 | Version not found |
| PreconditionFailed | 412 | Precondition failed |

---

## 9. Security Considerations

1. **Input Validation**:
   - Validate all bucket/key names
   - Limit object sizes
   - Sanitize metadata values

2. **Authentication**:
   - Use constant-time comparison for signatures
   - Implement rate limiting
   - Log authentication failures

3. **Authorization**:
   - Enforce least privilege principle
   - Validate all resource access

4. **Data Protection**:
   - Support encryption at rest (future)
   - Use TLS for transport
   - Secure secret key storage

---

## 10. Notes for Implementation Agent

- **DO NOT** overengineer - keep implementations simple and pragmatic
- **DO** use interfaces for all dependencies to enable testing
- **DO** handle context cancellation in all long-running operations
- **DO** return descriptive errors with proper wrapping
- **DO** write tests alongside implementation
- **DO** use `gofmt` and `go vet` to ensure code quality
- **DO** follow Go naming conventions (CamelCase for exported, camelCase for unexported)
- **DO** add JSON struct tags for all API types
- **DO** document all exported functions and types
- **DO NOT** use global variables except for configuration
- **DO NOT** ignore errors - handle or propagate them
- **DO NOT** use `panic` except in truly unrecoverable situations
