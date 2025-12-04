package object

import (
	"context"
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
)

// MemoryRepository implements Repository in memory
type MemoryRepository struct {
	objects map[string]*Object // Key: bucket/key
	mu      sync.RWMutex
}

// NewMemoryRepository creates a new memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		objects: make(map[string]*Object),
	}
}

func (r *MemoryRepository) Put(ctx context.Context, obj *Object, data io.Reader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := obj.BucketName + "/" + obj.Key
	r.objects[key] = obj
	return nil
}

func (r *MemoryRepository) Get(ctx context.Context, bucket, key string, versionID *string) (*Object, io.ReadCloser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	objKey := bucket + "/" + key
	obj, exists := r.objects[objKey]
	if !exists {
		return nil, nil, errors.New("object not found")
	}

	return obj, nil, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, bucket, key string, versionID *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	objKey := bucket + "/" + key
	delete(r.objects, objKey)
	return nil
}

func (r *MemoryRepository) List(ctx context.Context, bucket, prefix string, opts ListOptions) (*ListResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect matching objects
	var allObjects []*Object
	for _, obj := range r.objects {
		if obj.BucketName != bucket {
			continue
		}

		// Filter by prefix
		if opts.Prefix != "" && !strings.HasPrefix(obj.Key, opts.Prefix) {
			continue
		}

		// Filter by StartAfter
		if opts.StartAfter != "" && obj.Key <= opts.StartAfter {
			continue
		}

		allObjects = append(allObjects, obj)
	}

	// Sort by key
	sort.Slice(allObjects, func(i, j int) bool {
		return allObjects[i].Key < allObjects[j].Key
	})

	// Apply pagination
	maxKeys := opts.MaxKeys
	if maxKeys <= 0 {
		maxKeys = DefaultMaxKeys
	}
	if maxKeys > MaxKeysLimit {
		maxKeys = MaxKeysLimit
	}

	isTruncated := false
	nextMarker := ""
	var objects []*Object

	if len(allObjects) > maxKeys {
		isTruncated = true
		objects = allObjects[:maxKeys]
		nextMarker = objects[maxKeys-1].Key
	} else {
		objects = allObjects
	}

	return &ListResult{
		Objects:     objects,
		IsTruncated: isTruncated,
		NextMarker:  nextMarker,
	}, nil
}

func (r *MemoryRepository) Head(ctx context.Context, bucket, key string, versionID *string) (*Object, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	objKey := bucket + "/" + key
	obj, exists := r.objects[objKey]
	if !exists {
		return nil, errors.New("object not found")
	}

	return obj, nil
}

func (r *MemoryRepository) Count(ctx context.Context, bucket string) (int, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	var totalSize int64

	for _, obj := range r.objects {
		if obj.BucketName == bucket {
			count++
			totalSize += obj.Size
		}
	}

	return count, totalSize, nil
}

func (r *MemoryRepository) DeleteAll(ctx context.Context, bucket string) (int, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	var totalSize int64

	// Collect keys to delete
	var keysToDelete []string
	for key, obj := range r.objects {
		if obj.BucketName == bucket {
			keysToDelete = append(keysToDelete, key)
			count++
			totalSize += obj.Size
		}
	}

	// Delete all collected keys
	for _, key := range keysToDelete {
		delete(r.objects, key)
	}

	return count, totalSize, nil
}
