package multipart

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

// Service handles multipart upload operations
type Service struct {
	uploads map[string]*Upload // In-memory for now
}

// NewService creates a new multipart service
func NewService() *Service {
	return &Service{
		uploads: make(map[string]*Upload),
	}
}

// InitiateMultipartUpload initiates a new multipart upload
func (s *Service) InitiateMultipartUpload(ctx context.Context, bucket, key string) (*Upload, error) {
	uploadID := uuid.New().String()
	upload := &Upload{
		UploadID:   uploadID,
		BucketName: bucket,
		Key:        key,
		CreatedAt:  time.Now(),
		Parts:      make([]Part, 0),
	}

	s.uploads[uploadID] = upload
	return upload, nil
}

// UploadPart handles uploading a part
func (s *Service) UploadPart(ctx context.Context, bucket, key, uploadID string, partNumber int, size int64, etag string) (*Part, error) {
	upload, ok := s.uploads[uploadID]
	if !ok {
		return nil, errors.New("upload not found")
	}

	if partNumber < 1 || partNumber > 10000 {
		return nil, errors.New("invalid part number")
	}

	part := Part{
		PartNumber: partNumber,
		ETag:       etag,
		Size:       size,
	}

	// Check if part already exists and replace it
	found := false
	for i, p := range upload.Parts {
		if p.PartNumber == partNumber {
			upload.Parts[i] = part
			found = true
			break
		}
	}

	if !found {
		upload.Parts = append(upload.Parts, part)
	}

	return &part, nil
}

// ListParts lists parts for an upload
func (s *Service) ListParts(ctx context.Context, bucket, key, uploadID string) ([]Part, error) {
	upload, ok := s.uploads[uploadID]
	if !ok {
		return nil, errors.New("upload not found")
	}

	// Sort parts by part number
	sort.Slice(upload.Parts, func(i, j int) bool {
		return upload.Parts[i].PartNumber < upload.Parts[j].PartNumber
	})

	return upload.Parts, nil
}

// CompleteMultipartUpload completes a multipart upload
func (s *Service) CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []Part) error {
	upload, ok := s.uploads[uploadID]
	if !ok {
		return errors.New("upload not found")
	}

	// Verify parts
	if len(parts) != len(upload.Parts) {
		// This is a simple check, real impl should verify each part ETag/Checksum
	}

	// Merge parts (logic omitted for now as it requires storage engine interaction)

	delete(s.uploads, uploadID)
	return nil
}

// AbortMultipartUpload aborts a multipart upload
func (s *Service) AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error {
	if _, ok := s.uploads[uploadID]; !ok {
		return errors.New("upload not found")
	}

	// Cleanup parts (logic omitted)

	delete(s.uploads, uploadID)
	return nil
}
