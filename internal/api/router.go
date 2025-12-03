package api

import (
	"github.com/danielino/comio/internal/api/handlers"
	"github.com/danielino/comio/internal/api/middleware"
	"github.com/danielino/comio/internal/bucket"
	"github.com/danielino/comio/internal/monitoring"
	"github.com/danielino/comio/internal/object"
	"github.com/danielino/comio/internal/storage"
	"go.uber.org/zap"
)

// SetupRoutes configures the routes
func (s *Server) SetupRoutes() {
	// Initialize dependencies
	// In a real app, these would be singletons or passed in
	
	// Storage Engine
	// Create a dummy file for storage
	engine, err := storage.NewSimpleEngine("storage.data", 1024*1024*1024, storage.DefaultBlockSize) // 1GB
	if err != nil {
		monitoring.Log.Fatal("Failed to initialize storage engine", zap.Error(err))
	}
	if err := engine.Open("storage.data"); err != nil {
		// Try creating it?
		// For now, just log error
		monitoring.Log.Error("Failed to open storage device", zap.Error(err))
	}
	
	// Repositories
	bucketRepo := bucket.NewMemoryRepository()
	objectRepo := object.NewMemoryRepository()
	
	// Services
	bucketService := bucket.NewService(bucketRepo)
	objectService := object.NewService(objectRepo, engine)

	// Middleware
	s.router.Use(middleware.Recovery())
	s.router.Use(middleware.Logging())
	// Auth middleware should be applied to specific routes or globally if appropriate
	
	// Handlers
	bucketHandler := handlers.NewBucketHandler(bucketService)
	objectHandler := handlers.NewObjectHandler(objectService)
	adminHandler := handlers.NewAdminHandler(engine)
	
	// Service operations
	s.router.GET("/", bucketHandler.ListBuckets)
	
	// Bucket operations
	s.router.PUT("/:bucket", bucketHandler.CreateBucket)
	s.router.DELETE("/:bucket", bucketHandler.DeleteBucket)
	s.router.GET("/:bucket", objectHandler.ListObjects) // This conflicts with bucket operations if not careful, but S3 uses query params or just path
	s.router.HEAD("/:bucket", bucketHandler.HeadBucket)
	
	// Object operations
	s.router.PUT("/:bucket/:key", objectHandler.PutObject)
	s.router.GET("/:bucket/:key", objectHandler.GetObject)
	s.router.DELETE("/:bucket/:key", objectHandler.DeleteObject)
	s.router.HEAD("/:bucket/:key", objectHandler.HeadObject)
	
	// Admin object operations
	s.router.DELETE("/admin/:bucket/objects", objectHandler.DeleteAllObjects)
	
	// Admin endpoints
	admin := s.router.Group("/admin")
	{
		admin.GET("/health", adminHandler.HealthCheck)
		admin.GET("/metrics", adminHandler.Metrics)
	}
}
