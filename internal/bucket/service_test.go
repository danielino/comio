package bucket

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRepository_Create(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	bucket := &Bucket{
		Name:      "test-bucket",
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, bucket)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Try to create duplicate
	err = repo.Create(ctx, bucket)
	if err == nil {
		t.Error("Create() expected error for duplicate bucket, got nil")
	}
}

func TestMemoryRepository_Get(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	name := "test-bucket"
	bucket := &Bucket{
		Name:      name,
		CreatedAt: time.Now(),
	}

	repo.Create(ctx, bucket)

	// Get existing bucket
	found, err := repo.Get(ctx, name)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if found.Name != name {
		t.Errorf("Get() returned bucket with name %s, want %s", found.Name, name)
	}

	// Get non-existing bucket
	_, err = repo.Get(ctx, "non-existing")
	if err == nil {
		t.Error("Get() expected error for non-existing bucket, got nil")
	}
}

func TestMemoryRepository_List(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Create multiple buckets
	names := []string{"bucket-a", "bucket-b", "bucket-c"}
	for _, name := range names {
		repo.Create(ctx, &Bucket{Name: name, CreatedAt: time.Now()})
	}

	// List all
	buckets, err := repo.List(ctx, "")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(buckets) != len(names) {
		t.Errorf("List() returned %d buckets, want %d", len(buckets), len(names))
	}
}

func TestMemoryRepository_Delete(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	name := "test-bucket"
	bucket := &Bucket{Name: name, CreatedAt: time.Now()}
	repo.Create(ctx, bucket)

	// Delete
	err := repo.Delete(ctx, name)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deleted
	_, err = repo.Get(ctx, name)
	if err == nil {
		t.Error("Get() after Delete() should return error, got nil")
	}

	// Delete non-existing
	err = repo.Delete(ctx, "non-existing")
	if err == nil {
		t.Error("Delete() expected error for non-existing bucket, got nil")
	}
}

func TestMemoryRepository_Exists(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	name := "test-bucket"

	// Should not exist
	_, err := repo.Get(ctx, name)
	if err == nil {
		t.Error("Get() should return error for non-existing bucket")
	}

	// Create bucket
	repo.Create(ctx, &Bucket{Name: name, CreatedAt: time.Now()})

	// Should exist
	_, err = repo.Get(ctx, name)
	if err != nil {
		t.Errorf("Get() error = %v after creating bucket", err)
	}
}

func TestBucketService_CreateBucket(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	name := "test-bucket"
	err := service.CreateBucket(ctx, name, "owner1")
	if err != nil {
		t.Errorf("CreateBucket() error = %v", err)
	}

	bucket, err := service.GetBucket(ctx, name)
	if err != nil {
		t.Errorf("GetBucket() error = %v", err)
	}

	if bucket.Name != name {
		t.Errorf("CreateBucket() bucket name = %s, want %s", bucket.Name, name)
	}

	// Duplicate should fail
	err = service.CreateBucket(ctx, name, "owner1")
	if err == nil {
		t.Error("CreateBucket() expected error for duplicate, got nil")
	}
}

func TestBucketService_GetBucket(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	name := "test-bucket"
	service.CreateBucket(ctx, name, "owner1")

	bucket, err := service.GetBucket(ctx, name)
	if err != nil {
		t.Errorf("GetBucket() error = %v", err)
	}

	if bucket.Name != name {
		t.Errorf("GetBucket() bucket name = %s, want %s", bucket.Name, name)
	}
}

func TestBucketService_ListBuckets(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	owner := "owner1"
	names := []string{"bucket-1", "bucket-2", "bucket-3"}
	for _, name := range names {
		service.CreateBucket(ctx, name, owner)
	}

	buckets, err := service.ListBuckets(ctx, owner)
	if err != nil {
		t.Errorf("ListBuckets() error = %v", err)
	}

	if len(buckets) != len(names) {
		t.Errorf("ListBuckets() returned %d buckets, want %d", len(buckets), len(names))
	}
}

func TestBucketService_DeleteBucket(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	name := "test-bucket"
	service.CreateBucket(ctx, name, "owner1")

	err := service.DeleteBucket(ctx, name)
	if err != nil {
		t.Errorf("DeleteBucket() error = %v", err)
	}

	_, err = service.GetBucket(ctx, name)
	if err == nil {
		t.Error("GetBucket() after DeleteBucket() should return error, got nil")
	}
}

func TestMemoryRepository_Update(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	name := "test-bucket"
	bucket := &Bucket{
		Name:       name,
		CreatedAt:  time.Now(),
		Versioning: VersioningDisabled,
	}
	repo.Create(ctx, bucket)

	// Update versioning
	bucket.Versioning = VersioningEnabled
	err := repo.Update(ctx, bucket)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify update
	updated, err := repo.Get(ctx, name)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if updated.Versioning != VersioningEnabled {
		t.Errorf("Update() versioning = %s, want %s", updated.Versioning, VersioningEnabled)
	}

	// Update non-existing bucket
	err = repo.Update(ctx, &Bucket{Name: "non-existing"})
	if err == nil {
		t.Error("Update() expected error for non-existing bucket, got nil")
	}
}

func TestBucketService_InvalidNames(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	tests := []struct {
		name    string
		invalid bool
	}{
		{"valid-bucket-name", false},
		{"valid123", false},
		{"a", true},            // too short
		{"ab", true},           // too short
		{"Invalid", true},      // uppercase
		{"invalid_name", true}, // underscore
		{".invalid", true},     // starts with dot
		{"invalid.", true},     // ends with dot
		// Note: The current regex doesn't fully validate all S3 rules
		// (consecutive dots, IP addresses), just basic naming
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateBucket(ctx, tt.name, "owner")
			if tt.invalid && err == nil {
				t.Errorf("CreateBucket(%s) expected error, got nil", tt.name)
			}
			if !tt.invalid && err != nil {
				t.Errorf("CreateBucket(%s) unexpected error = %v", tt.name, err)
			}
		})
	}
}
