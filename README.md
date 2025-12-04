# ComIO - Community IO Storage

ComIO is a high-performance, S3-compliant object storage solution written in Go. It is designed to be production-ready, featuring raw device handling, asynchronous replication, and a robust CLI for management.

## Features

- **S3-Compatible API**: Standard RESTful API for bucket and object operations (PUT, GET, DELETE, HEAD).
- **High Performance Storage**: Direct I/O on raw block devices for optimal performance.
- **Cross-Site Replication**: Asynchronous, buffered replication for disaster recovery and high availability.
- **Data Integrity**: Built-in checksums (MD5, SHA256, CRC32) to ensure data consistency.
- **Authentication**: Secure access control using HMAC authentication.
- **Lifecycle Management**: Automated policies for object expiration and management.
- **Observability**: Integrated Prometheus metrics and structured logging.
- **CLI Management**: Comprehensive command-line interface for server administration and data manipulation.

## Architecture

ComIO is built with a modular architecture:

- **API Layer**: Handles HTTP requests using the Gin framework, providing S3-compatible endpoints.
- **Storage Engine**: Manages data persistence directly on block devices (`/dev/nvme...`) or file systems, using a custom allocator.
- **Metadata Store**: Uses a file-based storage system for efficient metadata management (buckets, object versions, user policies).
- **Replication Engine**: A dedicated worker pool handles asynchronous replication events to remote ComIO instances.

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Make
- Docker (optional)

### Installation

Clone the repository and build the project:

```bash
git clone https://github.com/danielino/comio.git
cd comio
make build
```

The binary will be available in `bin/comio`.

### Running with Docker

You can also build and run ComIO using Docker:

```bash
make docker-build
make docker-run
```

## Configuration

ComIO is configured via a YAML file. A default configuration looks like this:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"

storage:
  devices:
    - path: "/dev/nvme0n1"
      type: "block"
  block_size: 4096

replication:
  enabled: true
  nodes:
    - address: "https://replica-site.example.com"
  sync_interval: "1s"

auth:
  enabled: true
  admin_access_key: "admin"
  admin_secret_key: "change-me"

logging:
  level: "info"
  format: "json"
```

See `configs/config.yaml.example` for a full example.

## Usage

### Starting the Server

```bash
./bin/comio server --config config.yaml
```

### Using the CLI

ComIO provides a CLI for managing buckets and objects.

**Create a bucket:**
```bash
./bin/comio bucket create my-bucket
```

**Upload an object:**
```bash
./bin/comio object put my-bucket/my-file.txt --file ./local-file.txt
```

**List objects:**
```bash
./bin/comio object list my-bucket
```

## Development

The project includes a `Makefile` to simplify development tasks:

- `make build`: Build the binary.
- `make test`: Run unit tests.
- `make coverage`: Generate code coverage report.
- `make lint`: Run linters.
- `make fmt`: Format code.
- `make mocks`: Generate mock files for testing.

## Documentation

For more detailed information on specific features, check the `docs/` directory:

- [Replication Architecture](docs/replication.md)

## License

[MIT License](LICENSE)
