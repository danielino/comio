package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danielino/comio/internal/bucket"
)

// BucketHandler handles bucket operations
type BucketHandler struct {
	service *bucket.Service
}

// NewBucketHandler creates a new bucket handler
func NewBucketHandler(service *bucket.Service) *BucketHandler {
	return &BucketHandler{
		service: service,
	}
}

// ListBuckets lists all buckets
func (h *BucketHandler) ListBuckets(c *gin.Context) {
	// TODO: Get owner from auth context
	owner := "default"
	buckets, err := h.service.ListBuckets(c.Request.Context(), owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buckets)
}

// CreateBucket creates a new bucket
func (h *BucketHandler) CreateBucket(c *gin.Context) {
	bucketName := c.Param("bucket")
	// TODO: Get owner from auth context
	owner := "default"
	
	if err := h.service.CreateBucket(c.Request.Context(), bucketName, owner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"bucket": bucketName, "status": "created"})
}

// DeleteBucket deletes a bucket
func (h *BucketHandler) DeleteBucket(c *gin.Context) {
	bucketName := c.Param("bucket")
	if err := h.service.DeleteBucket(c.Request.Context(), bucketName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// HeadBucket checks if bucket exists
func (h *BucketHandler) HeadBucket(c *gin.Context) {
	bucketName := c.Param("bucket")
	if _, err := h.service.GetBucket(c.Request.Context(), bucketName); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}
