package object

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/danielino/comio/internal/integrity"
	"github.com/danielino/comio/internal/replication"
	"github.com/danielino/comio/internal/storage"
)

// Service handles object operations
type Service struct {
	repo       Repository
	engine     storage.Engine
	replicator *replication.Replicator
}

func (s *Service) SetReplicator(replicator *replication.Replicator) {
	s.replicator = replicator
}

// NewService creates a new object service
func NewService(repo Repository, engine storage.Engine) *Service {
	return &Service{
		repo:   repo,
		engine: engine,
	}
}

// PutObject uploads an object
func (s *Service) PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) (*Object, error) {
	// Calculate checksums while streaming?
	// For now, just pass through

	obj := &Object{
		Key:         key,
		BucketName:  bucket,
		Size:        size,
		ContentType: contentType,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		VersionID:   GenerateVersionID(), // Always generate version ID for now
	}

	// In a real impl, we would stream to storage engine here, calculate checksums, then save metadata to repo.
	// The repo.Put might handle the storage engine interaction or we do it here.
	// The prompt says "Stream object data to storage engine" in service.go

	// We need to wrap the reader to calculate checksums
	calc := integrity.NewCalculator()
	tee := io.TeeReader(data, calc)

	// Allocate storage space
	offset, err := s.engine.Allocate(size)
	if err != nil {
		return nil, err
	}

	// Setup cleanup: free allocated space if operation fails
	allocated := true
	defer func() {
		if allocated {
			// Operation failed - free the allocated space
			if freeErr := s.engine.Free(offset, size); freeErr != nil {
				// Log error but don't fail - cleanup is best effort
				// In production, a background process should handle orphaned blocks
			}
		}
	}()

	// Stream data from reader to storage in chunks
	buf := make([]byte, 4096) // 4KB chunks
	currentOffset := offset
	totalRead := int64(0)

	for {
		n, err := tee.Read(buf)
		if n > 0 {
			if wErr := s.engine.Write(currentOffset, buf[:n]); wErr != nil {
				// Write failed - cleanup will happen via defer
				return nil, wErr
			}
			currentOffset += int64(n)
			totalRead += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			// Read failed - cleanup will happen via defer
			return nil, err
		}
	}

	// Update object metadata with checksums
	sums := calc.Sums()
	obj.ETag = sums["MD5"]
	obj.Checksum = integrity.Checksum{Algorithm: "SHA256", Value: sums["SHA256"]}
	obj.Offset = offset // Store offset

	// Save metadata
	if err := s.repo.Put(ctx, obj, nil); err != nil {
		// Metadata save failed - cleanup will happen via defer
		return nil, err
	}

	// Success! Mark as committed so defer doesn't free the space
	allocated = false

	// Queue replication event
	if s.replicator != nil {
		event := replication.Event{
			Type:   replication.EventPutObject,
			Bucket: bucket,
			Key:    key,
			Metadata: map[string]interface{}{
				"content_type": contentType,
				"size":         size,
			},
		}

		// For very small objects (<1KB), include data inline to avoid extra storage reads
		// For larger objects, use storage pointer to avoid memory leak
		if size < 1024 { // 1KB threshold for inline
			// Small objects: read data and include inline
			inlineData, err := s.engine.Read(offset, size)
			if err == nil {
				event.Data = inlineData
			} else {
				// Fallback to pointer if read fails
				event.StoragePointer = &replication.StoragePointer{
					Offset: offset,
					Size:   size,
				}
			}
		} else {
			// Larger objects: use storage pointer (avoids memory leak)
			event.StoragePointer = &replication.StoragePointer{
				Offset: offset,
				Size:   size,
			}
		}

		s.replicator.QueueEvent(event)
	}

	return obj, nil
}

// GetObject retrieves an object
func (s *Service) GetObject(ctx context.Context, bucket, key string, versionID *string) (*Object, io.ReadCloser, error) {
	// Get metadata from repo
	obj, _, err := s.repo.Get(ctx, bucket, key, versionID)
	if err != nil {
		return nil, nil, err
	}

	// Read data from engine
	data, err := s.engine.Read(obj.Offset, obj.Size)
	if err != nil {
		return nil, nil, err
	}

	// Convert []byte to ReadCloser
	// In a real impl, we'd want a stream from the engine, not read all into memory.
	// But Engine.Read returns []byte.
	return obj, io.NopCloser(bytes.NewReader(data)), nil
}

// ListObjects lists objects in a bucket
func (s *Service) ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) (*ListResult, error) {
	return s.repo.List(ctx, bucket, prefix, opts)
}

// DeleteAllObjects deletes all objects in a bucket and returns total size freed
func (s *Service) DeleteAllObjects(ctx context.Context, bucket string) (int, int64, error) {
	// First, list all objects to get their offsets (we need to free storage)
	var allObjects []*Object
	startAfter := ""

	for {
		result, err := s.repo.List(ctx, bucket, "", ListOptions{
			MaxKeys:    1000,
			StartAfter: startAfter,
		})
		if err != nil {
			return 0, 0, err
		}

		if len(result.Objects) == 0 {
			break
		}

		allObjects = append(allObjects, result.Objects...)

		if !result.IsTruncated {
			break
		}
		startAfter = result.NextMarker
	}

	// Free storage for all objects
	for _, obj := range allObjects {
		if err := s.engine.Free(obj.Offset, obj.Size); err != nil {
			// Log but continue
		}
	}

	// Delete all metadata in one shot
	count, totalSize, err := s.repo.DeleteAll(ctx, bucket)
	if err != nil {
		return 0, 0, err
	}

	// Queue replication event
	if s.replicator != nil {
		s.replicator.QueueEvent(replication.Event{
			Type:   replication.EventPurgeBucket,
			Bucket: bucket,
		})
	}

	return count, totalSize, nil
}

// CountObjects returns the number of objects and total size in a bucket
func (s *Service) CountObjects(ctx context.Context, bucket string) (int, int64, error) {
	return s.repo.Count(ctx, bucket)
}

// DeleteObject deletes a single object
func (s *Service) DeleteObject(ctx context.Context, bucket, key string) error {
	// Get object metadata first to find storage location
	obj, _, err := s.repo.Get(ctx, bucket, key, nil)
	if err != nil {
		return err
	}

	// Free storage space
	if err := s.engine.Free(obj.Offset, obj.Size); err != nil {
		// Log but continue with metadata deletion
		// Storage cleanup can be done later by background process
	}

	// Delete metadata
	if err := s.repo.Delete(ctx, bucket, key, nil); err != nil {
		return err
	}

	// Queue replication event
	if s.replicator != nil {
		s.replicator.QueueEvent(replication.Event{
			Type:   replication.EventDeleteObject,
			Bucket: bucket,
			Key:    key,
		})
	}

	return nil
}

// GetObjectMetadata retrieves only object metadata without data
func (s *Service) GetObjectMetadata(ctx context.Context, bucket, key string) (*Object, error) {
	obj, _, err := s.repo.Get(ctx, bucket, key, nil)
	return obj, err
}
