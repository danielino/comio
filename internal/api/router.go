package api

import (
	"github.com/danielino/comio/internal/api/handlers"
	"github.com/danielino/comio/internal/api/middleware"
)

// SetupRoutes configures the routes using injected dependencies from the container
// All dependencies are now provided via dependency injection, making this method
// testable and decoupled from implementation details
func (s *Server) SetupRoutes() {
	// Apply global middleware
	s.router.Use(middleware.Recovery())
	s.router.Use(middleware.Logging())
	// Auth middleware should be applied to specific routes or globally if appropriate

	// Create handlers using injected services from container
	bucketHandler := handlers.NewBucketHandler(s.container.BucketService)
	objectHandler := handlers.NewObjectHandler(s.container.ObjectService)
	adminHandler := handlers.NewAdminHandler(s.container.Engine)

	// Service operations
	s.router.GET("/", bucketHandler.ListBuckets)

	// Bucket operations - with validation
	bucketRoutes := s.router.Group("/")
	bucketRoutes.Use(middleware.ValidateBucketName())
	{
		bucketRoutes.PUT("/:bucket", bucketHandler.CreateBucket)
		bucketRoutes.DELETE("/:bucket", bucketHandler.DeleteBucket)
		bucketRoutes.GET("/:bucket", objectHandler.ListObjects)
		bucketRoutes.HEAD("/:bucket", bucketHandler.HeadBucket)
	}

	// Object operations - with validation
	objectRoutes := s.router.Group("/")
	objectRoutes.Use(middleware.ValidateBucketName())
	objectRoutes.Use(middleware.ValidateObjectKey())
	objectRoutes.Use(middleware.ValidateContentLength())
	{
		objectRoutes.PUT("/:bucket/:key", objectHandler.PutObject)
		objectRoutes.GET("/:bucket/:key", objectHandler.GetObject)
		objectRoutes.DELETE("/:bucket/:key", objectHandler.DeleteObject)
		objectRoutes.HEAD("/:bucket/:key", objectHandler.HeadObject)
	}

	// Admin object operations
	s.router.DELETE("/admin/:bucket/objects", objectHandler.DeleteAllObjects)

	// Admin endpoints
	admin := s.router.Group("/admin")
	{
		admin.GET("/health", adminHandler.HealthCheck)
		admin.GET("/metrics", adminHandler.Metrics)
	}
}
