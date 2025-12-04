package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/danielino/comio/internal/monitoring"
	"github.com/danielino/comio/internal/object"
)

// ObjectHandler handles object operations
type ObjectHandler struct {
	service *object.Service
}

// NewObjectHandler creates a new object handler
func NewObjectHandler(service *object.Service) *ObjectHandler {
	return &ObjectHandler{
		service: service,
	}
}

// PutObject uploads an object
func (h *ObjectHandler) PutObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	// Get content length
	size := c.Request.ContentLength
	contentType := c.GetHeader("Content-Type")

	obj, err := h.service.PutObject(c.Request.Context(), bucket, key, c.Request.Body, size, contentType)
	if err != nil {
		monitoring.Log.Error("Failed to put object",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Int64("size", size),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, obj)
}

// GetObject retrieves an object
func (h *ObjectHandler) GetObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	obj, data, err := h.service.GetObject(c.Request.Context(), bucket, key, nil)
	if err != nil {
		monitoring.Log.Error("Failed to get object",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer data.Close()

	c.Header("Content-Type", obj.ContentType)
	c.Header("ETag", obj.ETag)
	// Stream data
	// io.Copy(c.Writer, data)
	// Gin has DataFromReader
	c.DataFromReader(http.StatusOK, obj.Size, obj.ContentType, data, map[string]string{
		"ETag": obj.ETag,
	})
}

// DeleteObject deletes an object
func (h *ObjectHandler) DeleteObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	err := h.service.DeleteObject(c.Request.Context(), bucket, key)
	if err != nil {
		monitoring.Log.Error("Failed to delete object",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// HeadObject checks if object exists and returns metadata
func (h *ObjectHandler) HeadObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	obj, err := h.service.GetObjectMetadata(c.Request.Context(), bucket, key)
	if err != nil {
		monitoring.Log.Error("Failed to head object",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Error(err))
		c.Status(http.StatusNotFound)
		return
	}

	// Return metadata as headers
	c.Header("Content-Type", obj.ContentType)
	c.Header("Content-Length", strconv.FormatInt(obj.Size, 10))
	c.Header("ETag", obj.ETag)
	c.Header("Last-Modified", obj.ModifiedAt.Format(http.TimeFormat))
	c.Status(http.StatusOK)
}

// ListObjects lists objects in a bucket
func (h *ObjectHandler) ListObjects(c *gin.Context) {
	bucket := c.Param("bucket")
	prefix := c.Query("prefix")
	delimiter := c.Query("delimiter")
	startAfter := c.Query("start-after")
	maxKeys := object.DefaultMaxKeys

	if maxKeysParam := c.Query("max-keys"); maxKeysParam != "" {
		if mk, err := strconv.Atoi(maxKeysParam); err == nil {
			maxKeys = mk
			if maxKeys > object.MaxKeysLimit {
				maxKeys = object.MaxKeysLimit
			}
		}
	}

	opts := object.ListOptions{
		Prefix:     prefix,
		Delimiter:  delimiter,
		StartAfter: startAfter,
		MaxKeys:    maxKeys,
	}

	result, err := h.service.ListObjects(c.Request.Context(), bucket, prefix, opts)
	if err != nil {
		monitoring.Log.Error("Failed to list objects",
			zap.String("bucket", bucket),
			zap.String("prefix", prefix),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteAllObjects deletes all objects in a bucket
func (h *ObjectHandler) DeleteAllObjects(c *gin.Context) {
	bucket := c.Param("bucket")

	// Check if this is a confirmation (POST) or info request (GET/DELETE without confirm)
	confirm := c.Query("confirm")

	if confirm == "true" {
		// Actually delete
		count, totalSize, err := h.service.DeleteAllObjects(c.Request.Context(), bucket)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"deleted_count": count,
			"freed_size":    totalSize,
		})
	} else {
		// Just get info using efficient count
		count, totalSize, err := h.service.CountObjects(c.Request.Context(), bucket)
		if err != nil {
			monitoring.Log.Error("Failed to count objects",
				zap.String("bucket", bucket),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":      count,
			"total_size": totalSize,
		})
	}
}
