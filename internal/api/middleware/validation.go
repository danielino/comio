package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// S3 bucket name rules:
	// - Between 3 and 63 characters
	// - Only lowercase letters, numbers, dots, and hyphens
	// - Must start and end with letter or number
	bucketNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

	// S3 object key rules:
	// - Up to 1024 characters
	// - Can contain any UTF-8 character
	// - Some characters should be avoided but are technically allowed
	maxKeyLength = 1024
)

// ValidateBucketName validates bucket name according to S3 naming rules
func ValidateBucketName() gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := c.Param("bucket")
		if bucket == "" {
			c.Next()
			return
		}

		// Check length
		if len(bucket) < 3 || len(bucket) > 63 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bucket name must be between 3 and 63 characters",
			})
			c.Abort()
			return
		}

		// Check format
		if !bucketNameRegex.MatchString(bucket) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid bucket name format",
			})
			c.Abort()
			return
		}

		// Additional checks
		if strings.Contains(bucket, "..") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bucket name cannot contain consecutive dots",
			})
			c.Abort()
			return
		}

		if strings.HasPrefix(bucket, "xn--") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bucket name cannot start with 'xn--'",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateObjectKey validates object key
func ValidateObjectKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.Next()
			return
		}

		// Check length
		if len(key) > maxKeyLength {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "object key exceeds maximum length of 1024 characters",
			})
			c.Abort()
			return
		}

		// Check for empty key
		if strings.TrimSpace(key) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "object key cannot be empty or only whitespace",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateContentLength validates that Content-Length header is present for PUT requests
func ValidateContentLength() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "PUT" {
			if c.Request.ContentLength < 0 {
				c.JSON(http.StatusLengthRequired, gin.H{
					"error": "Content-Length header is required",
				})
				c.Abort()
				return
			}

			// Optional: check for maximum size
			maxSize := int64(5 * 1024 * 1024 * 1024) // 5GB max
			if c.Request.ContentLength > maxSize {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"error": "object size exceeds maximum allowed size",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
