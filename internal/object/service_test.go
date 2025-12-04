package object

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/danielino/comio/internal/storage"
)

func createTestEngine(t *testing.T) storage.Engine {
	f, err := os.CreateTemp("", "object_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.Close()

	engine, err := storage.NewSimpleEngine(f.Name(), 64*1024*1024, 4*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	if err := engine.Open(f.Name()); err != nil {
		t.Fatalf("Failed to open engine: %v", err)
	}

	t.Cleanup(func() { engine.Close() })
	return engine
}

func TestObjectService_PutObject(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	bucket := "test-bucket"
	key := "test-key"
	data := []byte("test data")
	size := int64(len(data))

	obj, err := service.PutObject(ctx, bucket, key, bytes.NewReader(data), size, "text/plain")
	if err != nil {
		t.Errorf("PutObject() error = %v", err)
	}

	if obj.Key != key {
		t.Errorf("PutObject() key = %s, want %s", obj.Key, key)
	}

	if obj.Size != size {
		t.Errorf("PutObject() size = %d, want %d", obj.Size, size)
	}
}

func TestObjectService_GetObject(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	bucket := "test-bucket"
	key := "test-key"
	data := []byte("test data for get")

	_, err := service.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), "text/plain")
	if err != nil {
		t.Fatalf("Failed to put object: %v", err)
	}

	obj, reader, err := service.GetObject(ctx, bucket, key, nil)
	if err != nil {
		t.Errorf("GetObject() error = %v", err)
	}
	defer reader.Close()

	if obj.Key != key {
		t.Errorf("GetObject() key = %s, want %s", obj.Key, key)
	}

	readData, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("Failed to read object data: %v", err)
	}

	if !bytes.Equal(readData, data) {
		t.Errorf("GetObject() data = %v, want %v", readData, data)
	}
}

func TestObjectService_ListObjects(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	bucket := "test-bucket"
	keys := []string{"obj1", "obj2", "obj3"}
	for _, key := range keys {
		data := []byte("data")
		service.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), "text/plain")
	}

	result, err := service.ListObjects(ctx, bucket, "", ListOptions{MaxKeys: 10})
	if err != nil {
		t.Errorf("ListObjects() error = %v", err)
	}

	if len(result.Objects) != len(keys) {
		t.Errorf("ListObjects() returned %d objects, want %d", len(result.Objects), len(keys))
	}
}

func TestObjectService_DeleteAllObjects(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	bucket := "test-bucket"
	for i := 0; i < 3; i++ {
		key := string(rune('a' + i))
		data := []byte("data")
		service.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), "text/plain")
	}

	count, totalSize, err := service.DeleteAllObjects(ctx, bucket)
	if err != nil {
		t.Errorf("DeleteAllObjects() error = %v", err)
	}

	if count != 3 {
		t.Errorf("DeleteAllObjects() count = %d, want 3", count)
	}

	if totalSize == 0 {
		t.Error("DeleteAllObjects() totalSize = 0, want > 0")
	}

	result, err := service.ListObjects(ctx, bucket, "", ListOptions{MaxKeys: 10})
	if err != nil {
		t.Errorf("ListObjects() after delete error = %v", err)
	}

	if len(result.Objects) != 0 {
		t.Errorf("ListObjects() after delete returned %d objects, want 0", len(result.Objects))
	}
}

func TestMemoryRepository_Head(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	bucket := "test-bucket"
	key := "test-key"

	// Head non-existing object
	_, err := repo.Head(ctx, bucket, key, nil)
	if err == nil {
		t.Error("Head() expected error for non-existing object, got nil")
	}

	// Create object
	obj := &Object{
		BucketName:  bucket,
		Key:         key,
		Size:        100,
		ContentType: "text/plain",
	}
	data := bytes.NewReader([]byte("test data"))
	repo.Put(ctx, obj, data)

	// Head existing object
	found, err := repo.Head(ctx, bucket, key, nil)
	if err != nil {
		t.Errorf("Head() error = %v", err)
	}

	if found.Key != key {
		t.Errorf("Head() key = %s, want %s", found.Key, key)
	}
}

func TestObjectService_SetReplicator(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)

	// Test that SetReplicator doesn't panic
	service.SetReplicator(nil)
}

func TestObjectService_ListObjectsWithPagination(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	bucket := "test-bucket"

	// Create 10 objects
	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		data := []byte("data")
		service.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), "text/plain")
	}

	// List with pagination
	result, err := service.ListObjects(ctx, bucket, "", ListOptions{MaxKeys: 5})
	if err != nil {
		t.Errorf("ListObjects() error = %v", err)
	}

	if len(result.Objects) != 5 {
		t.Errorf("ListObjects() returned %d objects, want 5", len(result.Objects))
	}

	if !result.IsTruncated {
		t.Error("ListObjects() IsTruncated = false, want true")
	}

	// List next page
	result2, err := service.ListObjects(ctx, bucket, "", ListOptions{
		MaxKeys:    5,
		StartAfter: result.NextMarker,
	})
	if err != nil {
		t.Errorf("ListObjects() page 2 error = %v", err)
	}

	if len(result2.Objects) != 5 {
		t.Errorf("ListObjects() page 2 returned %d objects, want 5", len(result2.Objects))
	}
}

func TestObjectService_GetObjectNotFound(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	_, _, err := service.GetObject(ctx, "test-bucket", "non-existing", nil)
	if err == nil {
		t.Error("GetObject() expected error for non-existing object, got nil")
	}
}

func TestObjectService_ListObjectsWithPrefix(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	bucket := "test-bucket"

	// Create objects with different prefixes
	objects := []string{"docs/file1.txt", "docs/file2.txt", "images/pic1.jpg", "videos/vid1.mp4"}
	for _, key := range objects {
		data := []byte("data")
		service.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), "text/plain")
	}

	// List all objects
	result, err := service.ListObjects(ctx, bucket, "", ListOptions{MaxKeys: 10})
	if err != nil {
		t.Errorf("ListObjects() error = %v", err)
	}

	if len(result.Objects) != len(objects) {
		t.Errorf("ListObjects() returned %d objects, want %d", len(result.Objects), len(objects))
	}
}

func TestMemoryRepository_Delete(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	bucket := "test-bucket"
	key := "test-key"

	// Create object
	obj := &Object{
		BucketName:  bucket,
		Key:         key,
		Size:        100,
		ContentType: "text/plain",
	}
	data := bytes.NewReader([]byte("test data"))
	repo.Put(ctx, obj, data)

	// Delete it
	err := repo.Delete(ctx, bucket, key, nil)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deleted
	_, _, err = repo.Get(ctx, bucket, key, nil)
	if err == nil {
		t.Error("Get() after Delete() should return error, got nil")
	}
}

func TestObjectService_PutLargeObject(t *testing.T) {
	repo := NewMemoryRepository()
	engine := createTestEngine(t)
	service := NewService(repo, engine)
	ctx := context.Background()

	bucket := "test-bucket"
	key := "large-file"

	// Create 5MB data
	data := make([]byte, 5*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	obj, err := service.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), "application/octet-stream")
	if err != nil {
		t.Errorf("PutObject() large file error = %v", err)
	}

	if obj.Size != int64(len(data)) {
		t.Errorf("PutObject() size = %d, want %d", obj.Size, len(data))
	}

	// Verify we can get it back
	_, reader, err := service.GetObject(ctx, bucket, key, nil)
	if err != nil {
		t.Errorf("GetObject() large file error = %v", err)
	}
	defer reader.Close()

	readData, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("Failed to read large object: %v", err)
	}

	if len(readData) != len(data) {
		t.Errorf("Read data length = %d, want %d", len(readData), len(data))
	}
}
