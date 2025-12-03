package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BucketHandler handles bucket operations
type BucketHandler struct {
}

// NewBucketHandler creates a new bucket handler
func NewBucketHandler() *BucketHandler {
	return &BucketHandler{}
}

// ListBuckets lists all buckets
func (h *BucketHandler) ListBuckets(c *gin.Context) {
	c.Status(http.StatusOK)
}

// CreateBucket creates a new bucket
func (h *BucketHandler) CreateBucket(c *gin.Context) {
	bucket := c.Param("bucket")
	// Call service
	c.JSON(http.StatusOK, gin.H{"bucket": bucket, "status": "created"})
}

// DeleteBucket deletes a bucket
func (h *BucketHandler) DeleteBucket(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// HeadBucket checks if bucket exists
func (h *BucketHandler) HeadBucket(c *gin.Context) {
	c.Status(http.StatusOK)
}
