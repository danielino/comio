package object

import (
	"time"

	"github.com/danielino/comio/internal/integrity"
)

// Object represents a stored object
type Object struct {
	Key          string             `json:"key"`
	BucketName   string             `json:"bucket_name"`
	VersionID    string             `json:"version_id"`
	Size         int64              `json:"size"`
	ContentType  string             `json:"content_type"`
	ETag         string             `json:"etag"`
	Checksum     integrity.Checksum `json:"checksum"`
	CreatedAt    time.Time          `json:"created_at"`
	ModifiedAt   time.Time          `json:"modified_at"`
	Metadata     map[string]string  `json:"metadata"`
	StorageClass string             `json:"storage_class"`
	DeleteMarker bool               `json:"delete_marker"`
}
