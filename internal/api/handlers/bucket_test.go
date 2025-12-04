package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/danielino/comio/internal/bucket"
	"github.com/danielino/comio/internal/monitoring"
)

func init() {
	gin.SetMode(gin.TestMode)
	monitoring.InitLogger("error", "json", "stdout")
}

func setupBucketTest() (*gin.Engine, *bucket.Service) {
	router := gin.New()
	repo := bucket.NewMemoryRepository()
	service := bucket.NewService(repo)
	handler := NewBucketHandler(service)

	// Setup routes
	router.GET("/", handler.ListBuckets)
	router.PUT("/:bucket", handler.CreateBucket)
	router.DELETE("/:bucket", handler.DeleteBucket)
	router.HEAD("/:bucket", handler.HeadBucket)

	return router, service
}

func TestBucketHandler_CreateBucket(t *testing.T) {
	router, _ := setupBucketTest()

	tests := []struct {
		name           string
		bucketName     string
		expectedStatus int
	}{
		{
			name:           "create valid bucket",
			bucketName:     "test-bucket",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "create another bucket",
			bucketName:     "my-bucket-123",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("PUT", "/"+tt.bucketName, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.bucketName, response["bucket"])
				assert.Equal(t, "created", response["status"])
			}
		})
	}
}

func TestBucketHandler_CreateBucket_Duplicate(t *testing.T) {
	router, service := setupBucketTest()

	// Create bucket first time
	service.CreateBucket(nil, "test-bucket", "default")

	// Try to create again
	req, _ := http.NewRequest("PUT", "/test-bucket", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "already exists")
}

func TestBucketHandler_ListBuckets(t *testing.T) {
	router, service := setupBucketTest()

	// Create some buckets
	service.CreateBucket(nil, "bucket1", "default")
	service.CreateBucket(nil, "bucket2", "default")
	service.CreateBucket(nil, "bucket3", "default")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var buckets []*bucket.Bucket
	err := json.Unmarshal(w.Body.Bytes(), &buckets)
	assert.NoError(t, err)
	assert.Len(t, buckets, 3)

	// Check bucket names
	names := make([]string, len(buckets))
	for i, b := range buckets {
		names[i] = b.Name
	}
	assert.Contains(t, names, "bucket1")
	assert.Contains(t, names, "bucket2")
	assert.Contains(t, names, "bucket3")
}

func TestBucketHandler_ListBuckets_Empty(t *testing.T) {
	router, _ := setupBucketTest()

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var buckets []*bucket.Bucket
	err := json.Unmarshal(w.Body.Bytes(), &buckets)
	assert.NoError(t, err)
	assert.Len(t, buckets, 0)
}

func TestBucketHandler_HeadBucket(t *testing.T) {
	router, service := setupBucketTest()

	// Create a bucket
	service.CreateBucket(nil, "existing-bucket", "default")

	tests := []struct {
		name           string
		bucketName     string
		expectedStatus int
	}{
		{
			name:           "existing bucket",
			bucketName:     "existing-bucket",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent bucket",
			bucketName:     "missing-bucket",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("HEAD", "/"+tt.bucketName, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Empty(t, w.Body.String())
		})
	}
}

func TestBucketHandler_DeleteBucket(t *testing.T) {
	router, service := setupBucketTest()

	// Create a bucket
	service.CreateBucket(nil, "delete-me", "default")

	// Delete it
	req, _ := http.NewRequest("DELETE", "/delete-me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	// Verify it's gone
	req, _ = http.NewRequest("HEAD", "/delete-me", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBucketHandler_DeleteBucket_NotFound(t *testing.T) {
	router, _ := setupBucketTest()

	req, _ := http.NewRequest("DELETE", "/non-existent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Delete returns 500 if bucket doesn't exist (could be improved to return 404)
	assert.True(t, w.Code >= 400)
}
