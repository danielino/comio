package api

import (
	"context"
	"testing"
	"time"

	"github.com/danielino/comio/internal/bucket"
	"github.com/danielino/comio/internal/config"
	"github.com/danielino/comio/internal/monitoring"
	"github.com/danielino/comio/internal/object"
	"github.com/danielino/comio/internal/storage"
)

func init() {
	monitoring.InitLogger("info", "json", "stdout")
}

// mockEngine is a minimal mock implementation of storage.Engine for testing
type mockEngine struct{}

func (m *mockEngine) Open(devicePath string) error                  { return nil }
func (m *mockEngine) Close() error                                  { return nil }
func (m *mockEngine) Read(offset, size int64) ([]byte, error)       { return nil, nil }
func (m *mockEngine) Write(offset int64, data []byte) error         { return nil }
func (m *mockEngine) Allocate(size int64) (offset int64, err error) { return 0, nil }
func (m *mockEngine) Free(offset, size int64) error                 { return nil }
func (m *mockEngine) Sync() error                                   { return nil }
func (m *mockEngine) Stats() storage.Stats                          { return storage.Stats{} }
func (m *mockEngine) BlockSize() int                                { return 4096 }

// createTestContainer creates a minimal service container for testing
func createTestContainer(cfg *config.Config) *ServiceContainer {
	// Use in-memory implementations for testing
	bucketRepo := bucket.NewMemoryRepository()
	objectRepo := object.NewMemoryRepository()

	// Create a minimal mock engine
	engine := &mockEngine{}

	// Create services
	bucketService := bucket.NewService(bucketRepo)
	objectService := object.NewService(objectRepo, engine)

	return &ServiceContainer{
		Config:        cfg,
		Engine:        engine,
		BucketRepo:    bucketRepo,
		ObjectRepo:    objectRepo,
		BucketService: bucketService,
		ObjectService: objectService,
	}
}

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	container := createTestContainer(cfg)
	server := NewServer(cfg, container)
	if server == nil {
		t.Error("NewServer() returned nil")
	}
	if server.router == nil {
		t.Error("Server router is nil")
	}
	if server.cfg != cfg {
		t.Error("Server config not set correctly")
	}
	if server.container != container {
		t.Error("Server container not set correctly")
	}
}

func TestServer_StartStop(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         18080,
			ReadTimeout:  "30s",
			WriteTimeout: "30s",
		},
	}

	container := createTestContainer(cfg)
	server := NewServer(cfg, container)

	go func() {
		server.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestParseDuration(t *testing.T) {
	duration := parseDuration("30s")
	if duration != 30*time.Second {
		t.Errorf("parseDuration('30s') = %v, want 30s", duration)
	}

	duration = parseDuration("invalid")
	if duration != 30*time.Second {
		t.Errorf("parseDuration('invalid') should return default 30s, got %v", duration)
	}
}
