package replication

import "time"

type EventType string

const (
	EventPutObject    EventType = "put_object"
	EventDeleteObject EventType = "delete_object"
	EventPurgeBucket  EventType = "purge_bucket"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Bucket    string                 `json:"bucket"`
	Key       string                 `json:"key"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Data      []byte                 `json:"data,omitempty"` // For small objects
	DataURL   string                 `json:"data_url,omitempty"` // For large objects
}
