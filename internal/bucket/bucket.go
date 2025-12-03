package bucket

import (
	"time"
)

// VersioningStatus defines the versioning state of a bucket
type VersioningStatus string

const (
	VersioningEnabled   VersioningStatus = "Enabled"
	VersioningSuspended VersioningStatus = "Suspended"
	VersioningDisabled  VersioningStatus = "Disabled"
)

// Bucket represents a storage bucket
type Bucket struct {
	Name        string           `json:"name"`
	CreatedAt   time.Time        `json:"created_at"`
	Owner       string           `json:"owner"`
	Versioning  VersioningStatus `json:"versioning"`
	Lifecycle   []LifecycleRule  `json:"lifecycle,omitempty"`
}

// LifecycleRule represents a lifecycle policy rule
type LifecycleRule struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
