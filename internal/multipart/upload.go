package multipart

import (
	"time"
)

// Upload represents a multipart upload
type Upload struct {
	UploadID   string    `json:"upload_id"`
	BucketName string    `json:"bucket_name"`
	Key        string    `json:"key"`
	CreatedAt  time.Time `json:"created_at"`
	Parts      []Part    `json:"parts"`
}
