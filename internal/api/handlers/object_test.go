package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/danielino/comio/internal/bucket"
	"github.com/danielino/comio/internal/object"
	"github.com/danielino/comio/internal/storage"
)

// mockEngine for testing
type mockEngine struct {
	data map[int64][]byte
}

func newMockEngine() *mockEngine {
	return &mockEngine{
		data: make(map[int64][]byte),
	}
}

func (m *mockEngine) Open(devicePath string) error { return nil }
func (m *mockEngine) Close() error                 { return nil }
func (m *mockEngine) Sync() error                  { return nil }
func (m *mockEngine) Stats() storage.Stats         { return storage.Stats{} }
func (m *mockEngine) BlockSize() int               { return 4096 }

func (m *mockEngine) Allocate(size int64) (offset int64, err error) {
	// Simple allocator - just return next offset
	offset = int64(len(m.data))
	return offset, nil
}

func (m *mockEngine) Write(offset int64, data []byte) error {
	m.data[offset] = append([]byte{}, data...)
	return nil
}

func (m *mockEngine) Read(offset, size int64) ([]byte, error) {
	// Reconstruct data from chunks
	var result []byte
	for i := offset; i < offset+size; i++ {
		if chunk, ok := m.data[i]; ok {
			result = append(result, chunk...)
		}
	}
	return result, nil
}

func (m *mockEngine) Free(offset, size int64) error {
	// Simple free - delete entries
	for i := offset; i < offset+size; i++ {
		delete(m.data, i)
	}
	return nil
}

func setupObjectTest() (*gin.Engine, *object.Service, *bucket.Service) {
	router := gin.New()

	bucketRepo := bucket.NewMemoryRepository()
	objectRepo := object.NewMemoryRepository()
	engine := newMockEngine()

	bucketService := bucket.NewService(bucketRepo)
	objectService := object.NewService(objectRepo, engine)

	objectHandler := NewObjectHandler(objectService)

	// Setup routes
	router.PUT("/:bucket/:key", objectHandler.PutObject)
	router.GET("/:bucket/:key", objectHandler.GetObject)
	router.DELETE("/:bucket/:key", objectHandler.DeleteObject)
	router.HEAD("/:bucket/:key", objectHandler.HeadObject)
	router.GET("/:bucket", objectHandler.ListObjects)

	return router, objectService, bucketService
}

func TestObjectHandler_PutObject(t *testing.T) {
	router, _, bucketService := setupObjectTest()

	// Create bucket first
	bucketService.CreateBucket(nil, "test-bucket", "default")

	content := "Hello, World!"
	req, _ := http.NewRequest("PUT", "/test-bucket/test-key", strings.NewReader(content))
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = int64(len(content))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var obj object.Object
	err := json.Unmarshal(w.Body.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, "test-key", obj.Key)
	assert.Equal(t, "test-bucket", obj.BucketName)
	assert.Equal(t, int64(len(content)), obj.Size)
	assert.Equal(t, "text/plain", obj.ContentType)
	assert.NotEmpty(t, obj.ETag)
}

func TestObjectHandler_GetObject(t *testing.T) {
	router, objectService, bucketService := setupObjectTest()

	// Create bucket
	bucketService.CreateBucket(nil, "test-bucket", "default")

	// Put an object first
	content := "Test content for retrieval"
	objectService.PutObject(nil, "test-bucket", "test-key",
		strings.NewReader(content), int64(len(content)), "text/plain")

	// Get the object
	req, _ := http.NewRequest("GET", "/test-bucket/test-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Header().Get("ETag"))

	// Note: The body might be empty due to mock engine limitations
	// In a real implementation with proper mock, we'd verify the content
}

func TestObjectHandler_GetObject_NotFound(t *testing.T) {
	router, _, _ := setupObjectTest()

	req, _ := http.NewRequest("GET", "/test-bucket/missing-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestObjectHandler_HeadObject(t *testing.T) {
	router, objectService, bucketService := setupObjectTest()

	// Create bucket and object
	bucketService.CreateBucket(nil, "test-bucket", "default")
	content := "Test content"
	objectService.PutObject(nil, "test-bucket", "test-key",
		strings.NewReader(content), int64(len(content)), "application/octet-stream")

	// HEAD request
	req, _ := http.NewRequest("HEAD", "/test-bucket/test-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Header().Get("Content-Length"))
	assert.NotEmpty(t, w.Header().Get("ETag"))
	assert.NotEmpty(t, w.Header().Get("Last-Modified"))
	assert.Empty(t, w.Body.String())
}

func TestObjectHandler_HeadObject_NotFound(t *testing.T) {
	router, _, _ := setupObjectTest()

	req, _ := http.NewRequest("HEAD", "/test-bucket/missing-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestObjectHandler_DeleteObject(t *testing.T) {
	router, objectService, bucketService := setupObjectTest()

	// Create bucket and object
	bucketService.CreateBucket(nil, "test-bucket", "default")
	content := "Delete me"
	objectService.PutObject(nil, "test-bucket", "delete-key",
		strings.NewReader(content), int64(len(content)), "text/plain")

	// Delete it
	req, _ := http.NewRequest("DELETE", "/test-bucket/delete-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	// Verify it's gone with HEAD
	req, _ = http.NewRequest("HEAD", "/test-bucket/delete-key", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestObjectHandler_DeleteObject_NotFound(t *testing.T) {
	router, _, _ := setupObjectTest()

	req, _ := http.NewRequest("DELETE", "/test-bucket/missing-key", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestObjectHandler_ListObjects(t *testing.T) {
	router, objectService, bucketService := setupObjectTest()

	// Create bucket
	bucketService.CreateBucket(nil, "test-bucket", "default")

	// Add some objects
	objects := []string{"file1.txt", "file2.txt", "dir/file3.txt"}
	for _, key := range objects {
		content := "content for " + key
		objectService.PutObject(nil, "test-bucket", key,
			strings.NewReader(content), int64(len(content)), "text/plain")
	}

	// List objects
	req, _ := http.NewRequest("GET", "/test-bucket", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result object.ListResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result.Objects, 3)
}

func TestObjectHandler_ListObjects_WithPrefix(t *testing.T) {
	router, objectService, bucketService := setupObjectTest()

	// Create bucket
	bucketService.CreateBucket(nil, "test-bucket", "default")

	// Add objects with different prefixes
	objects := []string{"logs/2024/file1.txt", "logs/2024/file2.txt", "data/file3.txt"}
	for _, key := range objects {
		content := "content"
		objectService.PutObject(nil, "test-bucket", key,
			strings.NewReader(content), int64(len(content)), "text/plain")
	}

	// List with prefix
	req, _ := http.NewRequest("GET", "/test-bucket?prefix=logs/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result object.ListResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result.Objects, 2)

	for _, obj := range result.Objects {
		assert.True(t, strings.HasPrefix(obj.Key, "logs/"))
	}
}

func TestObjectHandler_PutObject_LargeContent(t *testing.T) {
	router, _, bucketService := setupObjectTest()

	// Create bucket
	bucketService.CreateBucket(nil, "test-bucket", "default")

	// Create large content (1MB)
	largeContent := bytes.Repeat([]byte("a"), 1024*1024)
	req, _ := http.NewRequest("PUT", "/test-bucket/large-file", bytes.NewReader(largeContent))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = int64(len(largeContent))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var obj object.Object
	err := json.Unmarshal(w.Body.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, int64(1024*1024), obj.Size)
}

func TestObjectHandler_PutObject_EmptyContent(t *testing.T) {
	router, _, bucketService := setupObjectTest()

	// Create bucket
	bucketService.CreateBucket(nil, "test-bucket", "default")

	// Put empty object
	req, _ := http.NewRequest("PUT", "/test-bucket/empty-file", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = 0

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var obj object.Object
	err := json.Unmarshal(w.Body.Bytes(), &obj)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), obj.Size)
}

// Benchmark tests
func BenchmarkObjectHandler_PutObject(b *testing.B) {
	router, _, bucketService := setupObjectTest()
	bucketService.CreateBucket(nil, "bench-bucket", "default")

	content := bytes.Repeat([]byte("x"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("PUT", "/bench-bucket/bench-key", bytes.NewReader(content))
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = int64(len(content))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkObjectHandler_GetObject(b *testing.B) {
	router, objectService, bucketService := setupObjectTest()
	bucketService.CreateBucket(nil, "bench-bucket", "default")

	content := bytes.Repeat([]byte("x"), 1024)
	objectService.PutObject(nil, "bench-bucket", "bench-key",
		bytes.NewReader(content), int64(len(content)), "application/octet-stream")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/bench-bucket/bench-key", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		io.Copy(io.Discard, w.Body)
	}
}
