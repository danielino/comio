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
