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
	
	// Write to storage engine
	// We need to allocate space first? Or does the engine handle it?
	// Engine.Write takes an offset. Engine.Allocate takes size.
	
	offset, err := s.engine.Allocate(size)
	if err != nil {
		return nil, err
	}
	
	// Read all data into memory? No, stream.
	// But Engine.Write takes []byte. We need a streaming write in Engine or read in chunks.
	// The Engine interface has Write(offset, data).
	// We should read from tee in chunks and write to engine.
	
	buf := make([]byte, 4096) // 4KB chunks
	currentOffset := offset
	totalRead := int64(0)
	
	for {
		n, err := tee.Read(buf)
		if n > 0 {
			if wErr := s.engine.Write(currentOffset, buf[:n]); wErr != nil {
				// Cleanup allocated space?
				return nil, wErr
			}
			currentOffset += int64(n)
			totalRead += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
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
		return nil, err
	}

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

		// For small objects (<1MB), include data inline
		if size < replication.InlineDataThreshold {
			// Read buf content for inline replication
			bufData := make([]byte, totalRead)
			copy(bufData, buf[:totalRead])
			event.Data = bufData
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
	// List all objects
	result, err := s.repo.List(ctx, bucket, "", ListOptions{MaxKeys: 0})
	if err != nil {
		return 0, 0, err
	}
	
	count := 0
	var totalSize int64
	
	// Delete each object
	for _, obj := range result.Objects {
		// For slab allocator, we only track actual used bytes
		// The allocator handles the block/slab overhead internally
		allocatedSize := obj.Size
		
		// Free storage space
		if err := s.engine.Free(obj.Offset, obj.Size); err != nil {
			// Log error but continue
			continue
		}
		
		// Delete metadata
		if err := s.repo.Delete(ctx, bucket, obj.Key, nil); err != nil {
			continue
		}
		
		count++
		totalSize += allocatedSize
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
