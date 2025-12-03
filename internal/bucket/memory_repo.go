package bucket

import (
	"context"
	"errors"
	"sync"
)

// MemoryRepository implements Repository in memory
type MemoryRepository struct {
	buckets map[string]*Bucket
	mu      sync.RWMutex
}

// NewMemoryRepository creates a new memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		buckets: make(map[string]*Bucket),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, bucket *Bucket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.buckets[bucket.Name]; exists {
		return errors.New("bucket already exists")
	}
	
	r.buckets[bucket.Name] = bucket
	return nil
}

func (r *MemoryRepository) Get(ctx context.Context, name string) (*Bucket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	bucket, exists := r.buckets[name]
	if !exists {
		return nil, errors.New("bucket not found")
	}
	
	return bucket, nil
}

func (r *MemoryRepository) List(ctx context.Context, owner string) ([]*Bucket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var buckets []*Bucket
	for _, b := range r.buckets {
		if b.Owner == owner || owner == "" {
			buckets = append(buckets, b)
		}
	}
	
	return buckets, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.buckets[name]; !exists {
		return errors.New("bucket not found")
	}
	
	delete(r.buckets, name)
	return nil
}

func (r *MemoryRepository) Update(ctx context.Context, bucket *Bucket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.buckets[bucket.Name]; !exists {
		return errors.New("bucket not found")
	}
	
	r.buckets[bucket.Name] = bucket
	return nil
}
