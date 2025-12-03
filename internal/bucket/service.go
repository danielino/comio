package bucket

import (
	"context"
	"errors"
	"regexp"
	"time"
)

// Service handles bucket operations
type Service struct {
	repo Repository
}

// NewService creates a new bucket service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateBucket creates a new bucket
func (s *Service) CreateBucket(ctx context.Context, name, owner string) error {
	if !isValidBucketName(name) {
		return errors.New("invalid bucket name")
	}

	// Check if exists
	_, err := s.repo.Get(ctx, name)
	if err == nil {
		return errors.New("bucket already exists")
	}

	bucket := &Bucket{
		Name:       name,
		CreatedAt:  time.Now(),
		Owner:      owner,
		Versioning: VersioningDisabled,
	}

	return s.repo.Create(ctx, bucket)
}

// GetBucket retrieves a bucket
func (s *Service) GetBucket(ctx context.Context, name string) (*Bucket, error) {
	return s.repo.Get(ctx, name)
}

// ListBuckets lists buckets for an owner
func (s *Service) ListBuckets(ctx context.Context, owner string) ([]*Bucket, error) {
	return s.repo.List(ctx, owner)
}

// DeleteBucket deletes a bucket
func (s *Service) DeleteBucket(ctx context.Context, name string) error {
	// TODO: Check if empty
	return s.repo.Delete(ctx, name)
}

func isValidBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	// Simple regex for S3 bucket naming
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`, name)
	return matched
}
