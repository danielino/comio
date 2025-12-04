package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danielino/comio/internal/auth"
	"github.com/danielino/comio/internal/config"
)

// ContextKeyUser is the key for user in context
const ContextKeyUser = "user"

// Authentication returns an authentication middleware
func Authentication(cfg *config.AuthConfig, authenticator auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if disabled
		if !cfg.Enabled {
			// Set default user for unauthenticated requests
			c.Set(ContextKeyUser, &auth.User{
				AccessKeyID: "anonymous",
				Username:    "default",
			})
			c.Next()
			return
		}

		// Authenticate the request
		user, err := authenticator.Authenticate(c.Request.Context(), c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication failed: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Store user in context
		c.Set(ContextKeyUser, user)
		c.Next()
	}
}

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(c *gin.Context) *auth.User {
	if user, exists := c.Get(ContextKeyUser); exists {
		if u, ok := user.(*auth.User); ok {
			return u
		}
	}
	// Return default user if not found
	return &auth.User{
		AccessKeyID: "anonymous",
		Username:    "default",
	}
}
