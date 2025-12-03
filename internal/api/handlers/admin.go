package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danielino/comio/internal/storage"
)

// AdminHandler handles admin operations
type AdminHandler struct {
	engine storage.Engine
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(engine storage.Engine) *AdminHandler {
	return &AdminHandler{
		engine: engine,
	}
}

// Metrics returns metrics
func (h *AdminHandler) Metrics(c *gin.Context) {
	stats := h.engine.Stats()
	c.JSON(http.StatusOK, gin.H{
		"storage": stats,
	})
}

// HealthCheck returns health status
func (h *AdminHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
