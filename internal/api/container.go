package api

import (
	"fmt"

	"github.com/danielino/comio/internal/bucket"
	"github.com/danielino/comio/internal/config"
	"github.com/danielino/comio/internal/monitoring"
	"github.com/danielino/comio/internal/object"
	"github.com/danielino/comio/internal/storage"
	"go.uber.org/zap"
)

// ServiceContainer holds all application dependencies
// This enables dependency injection and makes testing possible
type ServiceContainer struct {
	Config *config.Config

	// Storage layer
	Engine storage.Engine

	// Repositories (file-based like MinIO, no external DB)
	BucketRepo bucket.Repository
	ObjectRepo object.Repository

	// Services
	BucketService *bucket.Service
	ObjectService *object.Service
}

// NewServiceContainer creates and wires up all application dependencies
// This is the single place where we construct the entire dependency graph
func NewServiceContainer(cfg *config.Config) (*ServiceContainer, error) {
	container := &ServiceContainer{
		Config: cfg,
	}

	// Initialize storage engine
	if err := container.initStorage(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize file-based repositories (MinIO-style, no external DB)
	if err := container.initRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Initialize services
	container.initServices()

	return container, nil
}

// initStorage initializes the storage engine
func (c *ServiceContainer) initStorage() error {
	// Use storage config from config file, or fall back to defaults
	storagePath := "storage.data"
	storageSize := int64(1024 * 1024 * 1024) // 1GB default
	blockSize := storage.DefaultBlockSize

	// If config has storage devices configured, use the first one
	if len(c.Config.Storage.Devices) > 0 {
		storagePath = c.Config.Storage.Devices[0].Path
	}

	// Override block size if configured
	if c.Config.Storage.BlockSize > 0 {
		blockSize = c.Config.Storage.BlockSize
	}

	engine, err := storage.NewSimpleEngine(storagePath, storageSize, blockSize)
	if err != nil {
		return fmt.Errorf("failed to create storage engine: %w", err)
	}

	// Open the storage device
	if err := engine.Open(storagePath); err != nil {
		monitoring.Log.Warn("Failed to open existing storage device, it may be created on first use",
			zap.String("path", storagePath),
			zap.Error(err))
	}

	c.Engine = engine
	monitoring.Log.Info("Storage engine initialized",
		zap.String("path", storagePath),
		zap.Int("blockSize", blockSize))

	return nil
}

// initRepositories initializes the bucket and object repositories
// Using file-based storage like MinIO (no external database)
func (c *ServiceContainer) initRepositories() error {
	// Metadata directory
	metadataPath := "metadata"

	// Initialize file-based bucket repository
	bucketRepo, err := bucket.NewFileRepository(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create bucket repository: %w", err)
	}
	c.BucketRepo = bucketRepo

	// Initialize file-based object repository
	objectRepo, err := object.NewFileRepository(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create object repository: %w", err)
	}
	c.ObjectRepo = objectRepo

	monitoring.Log.Info("Repositories initialized",
		zap.String("type", "file-based"),
		zap.String("path", metadataPath),
		zap.String("style", "MinIO-like"))

	return nil
}

// initServices initializes the business logic services
func (c *ServiceContainer) initServices() {
	c.BucketService = bucket.NewService(c.BucketRepo)
	c.ObjectService = object.NewService(c.ObjectRepo, c.Engine)

	// Wire up the object counter for bucket emptiness checks
	c.BucketService.SetObjectCounter(c.ObjectRepo)

	monitoring.Log.Info("Services initialized")
}

// Close gracefully shuts down all resources
// Call this during application shutdown to clean up properly
func (c *ServiceContainer) Close() error {
	monitoring.Log.Info("Shutting down service container")

	// Close storage engine if it has a Close method
	if closer, ok := c.Engine.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			monitoring.Log.Error("Failed to close storage engine", zap.Error(err))
			return err
		}
	}

	monitoring.Log.Info("Service container shut down successfully")
	return nil
}
