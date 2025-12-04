package replication

import "time"

type EventType string

const (
	EventPutObject    EventType = "put_object"
	EventDeleteObject EventType = "delete_object"
	EventPurgeBucket  EventType = "purge_bucket"
)

// StoragePointer points to object data in storage engine
// This avoids copying large object data into event queue (memory leak fix)
type StoragePointer struct {
	Offset int64 `json:"offset"`
	Size   int64 `json:"size"`
}

type Event struct {
	ID             string                 `json:"id"`
	Type           EventType              `json:"type"`
	Bucket         string                 `json:"bucket"`
	Key            string                 `json:"key"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Data           []byte                 `json:"data,omitempty"`            // For small objects (<1MB) - inline data
	DataURL        string                 `json:"data_url,omitempty"`        // For large objects - external URL
	StoragePointer *StoragePointer        `json:"storage_pointer,omitempty"` // For objects in local storage - avoids memory copy
}
