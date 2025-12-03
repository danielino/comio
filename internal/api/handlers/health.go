package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns health status
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Metrics returns metrics
func Metrics(c *gin.Context) {
	c.Status(http.StatusOK)
}
