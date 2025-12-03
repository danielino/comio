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
