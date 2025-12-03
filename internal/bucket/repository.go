package bucket

import (
	"context"
)

// Repository defines the bucket persistence interface
type Repository interface {
	Create(ctx context.Context, bucket *Bucket) error
	Get(ctx context.Context, name string) (*Bucket, error)
	List(ctx context.Context, owner string) ([]*Bucket, error)
	Delete(ctx context.Context, name string) error
	Update(ctx context.Context, bucket *Bucket) error
}
