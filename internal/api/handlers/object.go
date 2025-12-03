package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ObjectHandler handles object operations
type ObjectHandler struct {
}

// NewObjectHandler creates a new object handler
func NewObjectHandler() *ObjectHandler {
	return &ObjectHandler{}
}

// PutObject uploads an object
func (h *ObjectHandler) PutObject(c *gin.Context) {
	c.Status(http.StatusOK)
}

// GetObject retrieves an object
func (h *ObjectHandler) GetObject(c *gin.Context) {
	c.Status(http.StatusOK)
}

// DeleteObject deletes an object
func (h *ObjectHandler) DeleteObject(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// HeadObject checks if object exists
func (h *ObjectHandler) HeadObject(c *gin.Context) {
	c.Status(http.StatusOK)
}

// ListObjects lists objects in a bucket
func (h *ObjectHandler) ListObjects(c *gin.Context) {
	c.Status(http.StatusOK)
}
