package bucket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileRepository implements Repository using filesystem metadata files
type FileRepository struct {
	metadataDir string
	mu          sync.RWMutex
}

// NewFileRepository creates a new file-based repository
func NewFileRepository(metadataDir string) (*FileRepository, error) {
	// Create buckets directory
	bucketsDir := filepath.Join(metadataDir, "buckets")
	if err := os.MkdirAll(bucketsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create buckets directory: %w", err)
	}

	return &FileRepository{
		metadataDir: metadataDir,
	}, nil
}

// getBucketMetaPath returns the path to a bucket's metadata file
func (r *FileRepository) getBucketMetaPath(name string) string {
	safeName := sanitizePath(name)
	return filepath.Join(r.metadataDir, "buckets", safeName+".json")
}

// sanitizePath sanitizes a path component to be filesystem-safe
func sanitizePath(s string) string {
	// Replace unsafe characters with underscores
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "..", "_")
	return s
}

func (r *FileRepository) Create(ctx context.Context, bucket *Bucket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metaPath := r.getBucketMetaPath(bucket.Name)

	// Check if bucket already exists
	if _, err := os.Stat(metaPath); err == nil {
		return errors.New("bucket already exists")
	}

	// Marshal bucket metadata to JSON
	metaData, err := json.MarshalIndent(bucket, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write metadata file atomically
	tempPath := metaPath + ".tmp"
	if err := os.WriteFile(tempPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	if err := os.Rename(tempPath, metaPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename metadata file: %w", err)
	}

	return nil
}

func (r *FileRepository) Get(ctx context.Context, name string) (*Bucket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metaPath := r.getBucketMetaPath(name)

	// Read metadata file
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("bucket not found")
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Unmarshal metadata
	var bucket Bucket
	if err := json.Unmarshal(metaData, &bucket); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &bucket, nil
}

func (r *FileRepository) List(ctx context.Context, owner string) ([]*Bucket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bucketsDir := filepath.Join(r.metadataDir, "buckets")

	// Read all bucket metadata files
	entries, err := os.ReadDir(bucketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Bucket{}, nil
		}
		return nil, fmt.Errorf("failed to read buckets directory: %w", err)
	}

	var buckets []*Bucket
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		metaPath := filepath.Join(bucketsDir, entry.Name())
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			continue // Skip files we can't read
		}

		var bucket Bucket
		if err := json.Unmarshal(metaData, &bucket); err != nil {
			continue // Skip invalid metadata
		}

		// Filter by owner if specified
		if owner != "" && bucket.Owner != owner {
			continue
		}

		buckets = append(buckets, &bucket)
	}

	return buckets, nil
}

func (r *FileRepository) Delete(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metaPath := r.getBucketMetaPath(name)

	if err := os.Remove(metaPath); err != nil {
		if os.IsNotExist(err) {
			return errors.New("bucket not found")
		}
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

func (r *FileRepository) Update(ctx context.Context, bucket *Bucket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metaPath := r.getBucketMetaPath(bucket.Name)

	// Check if bucket exists
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return errors.New("bucket not found")
	}

	// Marshal bucket metadata to JSON
	metaData, err := json.MarshalIndent(bucket, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write metadata file atomically
	tempPath := metaPath + ".tmp"
	if err := os.WriteFile(tempPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	if err := os.Rename(tempPath, metaPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename metadata file: %w", err)
	}

	return nil
}
