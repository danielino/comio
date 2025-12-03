package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MultipartHandler handles multipart upload operations
type MultipartHandler struct {
}

// NewMultipartHandler creates a new multipart handler
func NewMultipartHandler() *MultipartHandler {
	return &MultipartHandler{}
}

// InitiateMultipartUpload initiates a multipart upload
func (h *MultipartHandler) InitiateMultipartUpload(c *gin.Context) {
	c.Status(http.StatusOK)
}

// UploadPart uploads a part
func (h *MultipartHandler) UploadPart(c *gin.Context) {
	c.Status(http.StatusOK)
}

// CompleteMultipartUpload completes a multipart upload
func (h *MultipartHandler) CompleteMultipartUpload(c *gin.Context) {
	c.Status(http.StatusOK)
}

// AbortMultipartUpload aborts a multipart upload
func (h *MultipartHandler) AbortMultipartUpload(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ListParts lists parts
func (h *MultipartHandler) ListParts(c *gin.Context) {
	c.Status(http.StatusOK)
}
