package multipart

import (
	"testing"
	"time"
)

func TestUpload_Creation(t *testing.T) {
	upload := &Upload{
		UploadID:   "test-upload-123",
		BucketName: "test-bucket",
		Key:        "test/object.dat",
		CreatedAt:  time.Now(),
		Parts:      []Part{},
	}

	if upload.UploadID != "test-upload-123" {
		t.Errorf("UploadID = %s, want test-upload-123", upload.UploadID)
	}
	if upload.BucketName != "test-bucket" {
		t.Errorf("BucketName = %s, want test-bucket", upload.BucketName)
	}
	if len(upload.Parts) != 0 {
		t.Errorf("Parts count = %d, want 0", len(upload.Parts))
	}
}

func TestUpload_WithParts(t *testing.T) {
	upload := &Upload{
		UploadID:   "test-upload-456",
		BucketName: "test-bucket",
		Key:        "test/object.dat",
		CreatedAt:  time.Now(),
		Parts: []Part{
			{PartNumber: 1, Size: 1024, ETag: "etag1"},
			{PartNumber: 2, Size: 2048, ETag: "etag2"},
		},
	}

	if len(upload.Parts) != 2 {
		t.Errorf("Parts count = %d, want 2", len(upload.Parts))
	}
	if upload.Parts[0].PartNumber != 1 {
		t.Errorf("Part 0 number = %d, want 1", upload.Parts[0].PartNumber)
	}
	if upload.Parts[1].Size != 2048 {
		t.Errorf("Part 1 size = %d, want 2048", upload.Parts[1].Size)
	}
}
