package object

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileRepository implements Repository using filesystem metadata files
// Like MinIO: no global locks, filesystem handles concurrency
type FileRepository struct {
	metadataDir string
	// No global mutex - each file operation is independent
	// Filesystem provides atomic operations (rename) and concurrency
}

// NewFileRepository creates a new file-based repository
func NewFileRepository(metadataDir string) (*FileRepository, error) {
	// Create metadata directory structure
	objectsDir := filepath.Join(metadataDir, "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	return &FileRepository{
		metadataDir: metadataDir,
	}, nil
}

// getObjectMetaPath returns the path to an object's metadata file
func (r *FileRepository) getObjectMetaPath(bucket, key string) string {
	// Sanitize bucket and key for filesystem
	safeBucket := sanitizePath(bucket)
	safeKey := sanitizePath(key)
	return filepath.Join(r.metadataDir, "objects", safeBucket, safeKey+".meta")
}

// getBucketDir returns the directory for a bucket's objects
func (r *FileRepository) getBucketDir(bucket string) string {
	safeBucket := sanitizePath(bucket)
	return filepath.Join(r.metadataDir, "objects", safeBucket)
}

// sanitizePath sanitizes a path component to be filesystem-safe
func sanitizePath(s string) string {
	// Replace unsafe characters with underscores
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "..", "_")
	return s
}

func (r *FileRepository) Put(ctx context.Context, obj *Object, data io.Reader) error {
	// No global lock - filesystem handles concurrency
	metaPath := r.getObjectMetaPath(obj.BucketName, obj.Key)

	// Create bucket directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(metaPath), 0755); err != nil {
		return fmt.Errorf("failed to create bucket directory: %w", err)
	}

	// Marshal object metadata to JSON
	metaData, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write metadata file atomically (write to temp, then rename)
	tempPath := metaPath + ".tmp"
	if err := os.WriteFile(tempPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	if err := os.Rename(tempPath, metaPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename metadata file: %w", err)
	}

	return nil
}

func (r *FileRepository) Get(ctx context.Context, bucket, key string, versionID *string) (*Object, io.ReadCloser, error) {
	metaPath := r.getObjectMetaPath(bucket, key)

	// Read metadata file
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, errors.New("object not found")
		}
		return nil, nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Unmarshal metadata
	var obj Object
	if err := json.Unmarshal(metaData, &obj); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &obj, nil, nil
}

func (r *FileRepository) Delete(ctx context.Context, bucket, key string, versionID *string) error {

	metaPath := r.getObjectMetaPath(bucket, key)

	if err := os.Remove(metaPath); err != nil {
		if os.IsNotExist(err) {
			return errors.New("object not found")
		}
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

func (r *FileRepository) List(ctx context.Context, bucket, prefix string, opts ListOptions) (*ListResult, error) {

	bucketDir := r.getBucketDir(bucket)

	// Check if bucket directory exists
	if _, err := os.Stat(bucketDir); os.IsNotExist(err) {
		return &ListResult{
			Objects:        []*Object{},
			CommonPrefixes: []string{},
			IsTruncated:    false,
		}, nil
	}

	// Read all metadata files in the bucket
	var allObjects []*Object
	err := filepath.Walk(bucketDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".meta") {
			return nil
		}

		// Read metadata
		metaData, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		var obj Object
		if err := json.Unmarshal(metaData, &obj); err != nil {
			return nil // Skip invalid metadata
		}

		// Apply prefix filter
		if prefix != "" && !strings.HasPrefix(obj.Key, prefix) {
			return nil
		}

		allObjects = append(allObjects, &obj)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	// Sort objects by key
	sort.Slice(allObjects, func(i, j int) bool {
		return allObjects[i].Key < allObjects[j].Key
	})

	// Apply StartAfter filter
	if opts.StartAfter != "" {
		filtered := make([]*Object, 0, len(allObjects))
		for _, obj := range allObjects {
			if obj.Key > opts.StartAfter {
				filtered = append(filtered, obj)
			}
		}
		allObjects = filtered
	}

	// Apply MaxKeys limit
	maxKeys := opts.MaxKeys
	if maxKeys <= 0 {
		maxKeys = DefaultMaxKeys
	}
	if maxKeys > MaxKeysLimit {
		maxKeys = MaxKeysLimit
	}

	isTruncated := len(allObjects) > maxKeys
	if isTruncated {
		allObjects = allObjects[:maxKeys]
	}

	var nextMarker string
	if isTruncated && len(allObjects) > 0 {
		nextMarker = allObjects[len(allObjects)-1].Key
	}

	// Handle delimiter (common prefixes)
	var commonPrefixes []string
	if opts.Delimiter != "" {
		prefixMap := make(map[string]bool)
		var filteredObjects []*Object

		for _, obj := range allObjects {
			// Get the part after the prefix
			remainder := obj.Key
			if prefix != "" {
				remainder = strings.TrimPrefix(obj.Key, prefix)
			}

			// Check if there's a delimiter in the remainder
			idx := strings.Index(remainder, opts.Delimiter)
			if idx >= 0 {
				// This is a "directory", add to common prefixes
				commonPrefix := prefix + remainder[:idx+len(opts.Delimiter)]
				prefixMap[commonPrefix] = true
			} else {
				// This is a direct object
				filteredObjects = append(filteredObjects, obj)
			}
		}

		allObjects = filteredObjects
		for p := range prefixMap {
			commonPrefixes = append(commonPrefixes, p)
		}
		sort.Strings(commonPrefixes)
	}

	return &ListResult{
		Objects:        allObjects,
		CommonPrefixes: commonPrefixes,
		IsTruncated:    isTruncated,
		NextMarker:     nextMarker,
	}, nil
}

func (r *FileRepository) Head(ctx context.Context, bucket, key string, versionID *string) (*Object, error) {

	metaPath := r.getObjectMetaPath(bucket, key)

	// Read metadata file
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("object not found")
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Unmarshal metadata
	var obj Object
	if err := json.Unmarshal(metaData, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &obj, nil
}

func (r *FileRepository) Count(ctx context.Context, bucket string) (int, int64, error) {

	bucketDir := r.getBucketDir(bucket)

	// Check if bucket directory exists
	if _, err := os.Stat(bucketDir); os.IsNotExist(err) {
		return 0, 0, nil
	}

	count := 0
	var totalSize int64

	// Walk directory and count .meta files
	err := filepath.Walk(bucketDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() || !strings.HasSuffix(path, ".meta") {
			return nil
		}

		count++

		// Read metadata to get size
		metaData, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		var obj Object
		if err := json.Unmarshal(metaData, &obj); err != nil {
			return nil // Skip invalid metadata
		}

		totalSize += obj.Size
		return nil
	})

	if err != nil {
		return 0, 0, fmt.Errorf("failed to count objects: %w", err)
	}

	return count, totalSize, nil
}

func (r *FileRepository) DeleteAll(ctx context.Context, bucket string) (int, int64, error) {

	bucketDir := r.getBucketDir(bucket)

	// Check if bucket directory exists
	if _, err := os.Stat(bucketDir); os.IsNotExist(err) {
		return 0, 0, nil
	}

	count := 0
	var totalSize int64
	var objects []*Object

	// Read directory entries (faster than Walk)
	entries, err := os.ReadDir(bucketDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read bucket directory: %w", err)
	}

	// Collect all objects first
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".meta") {
			continue
		}

		metaPath := filepath.Join(bucketDir, entry.Name())

		// Read metadata to get offset and size
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			continue // Skip files we can't read
		}

		var obj Object
		if err := json.Unmarshal(metaData, &obj); err != nil {
			continue // Skip invalid metadata
		}

		objects = append(objects, &obj)
		totalSize += obj.Size
	}

	// Now delete all metadata files
	for _, obj := range objects {
		metaPath := r.getObjectMetaPath(bucket, obj.Key)
		if err := os.Remove(metaPath); err == nil {
			count++
		}
	}

	return count, totalSize, nil
}
