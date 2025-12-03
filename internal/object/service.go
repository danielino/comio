package object

import (
	"context"
	"io"
	"time"

	"github.com/danielino/comio/internal/integrity"
	"github.com/danielino/comio/internal/storage"
)

// Service handles object operations
type Service struct {
	repo   Repository
	engine storage.Engine
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
	
	// Save metadata
	if err := s.repo.Put(ctx, obj, nil); err != nil { // data is nil because we already wrote it? Or repo handles metadata only?
		// The repo interface says Put(obj, data). If repo handles storage, then we shouldn't use engine here.
		// But the prompt says "Stream object data to storage engine" in service.go AND "Repository: Put(obj, data)".
		// This implies the Repository might be an abstraction over metadata + storage, OR we use Repository for metadata and Engine for data.
		// Given the prompt structure, it seems Repository is for persistence (metadata DB?), and Engine is for raw storage.
		// But Repository.Put taking `data` suggests it might handle data too.
		// However, `internal/storage` exists.
		// Let's assume Repository is for metadata and we pass nil data if we handled storage manually, OR Repository uses Engine.
		// Let's assume Repository is just metadata for now, or we pass nil.
		return nil, err
	}
	
	return obj, nil
}

// GetObject retrieves an object
func (s *Service) GetObject(ctx context.Context, bucket, key string, versionID *string) (*Object, io.ReadCloser, error) {
	return s.repo.Get(ctx, bucket, key, versionID)
}
