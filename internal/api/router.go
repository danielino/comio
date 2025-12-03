package api

import (
	"github.com/danielino/comio/internal/api/handlers"
	"github.com/danielino/comio/internal/api/middleware"
)

// SetupRoutes configures the routes
func (s *Server) SetupRoutes() {
	// Middleware
	s.router.Use(middleware.Recovery())
	s.router.Use(middleware.Logging())
	// Auth middleware should be applied to specific routes or globally if appropriate
	
	// Handlers
	bucketHandler := handlers.NewBucketHandler()
	objectHandler := handlers.NewObjectHandler()
	
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
	
	// Admin endpoints
	admin := s.router.Group("/admin")
	{
		admin.GET("/health", handlers.HealthCheck)
		admin.GET("/metrics", handlers.Metrics)
	}
}
