package object

import (
	"context"
	"io"
)

const (
	// DefaultMaxKeys is the default number of objects returned in a list operation
	DefaultMaxKeys = 50
	// MaxKeysLimit is the maximum number of objects that can be returned in a list operation
	MaxKeysLimit = 50
)

// ListOptions defines options for listing objects
type ListOptions struct {
	MaxKeys    int
	Prefix     string
	Delimiter  string
	StartAfter string
}

// ListResult defines the result of listing objects
type ListResult struct {
	Objects        []*Object
	CommonPrefixes []string
	IsTruncated    bool
	NextMarker     string
}

// Repository defines the object persistence interface
type Repository interface {
	Put(ctx context.Context, obj *Object, data io.Reader) error
	Get(ctx context.Context, bucket, key string, versionID *string) (*Object, io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string, versionID *string) error
	List(ctx context.Context, bucket, prefix string, opts ListOptions) (*ListResult, error)
	Head(ctx context.Context, bucket, key string, versionID *string) (*Object, error)
}
