package bucket

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielino/comio/internal/database"
)

// SQLiteRepository implements Repository using SQLite
type SQLiteRepository struct {
	db *database.DB
}

// NewSQLiteRepository creates a new SQLite-based bucket repository
func NewSQLiteRepository(db *database.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

// Create creates a new bucket
func (r *SQLiteRepository) Create(ctx context.Context, bucket *Bucket) error {
	query := `
		INSERT INTO buckets (name, owner, created_at, versioning_enabled)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecWithRetry(ctx, query,
		bucket.Name,
		bucket.Owner,
		bucket.CreatedAt,
		bucket.Versioning,
	)

	if err != nil {
		// Check for unique constraint violation (bucket already exists)
		if isSQLiteConstraintError(err) {
			return fmt.Errorf("bucket '%s' already exists", bucket.Name)
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// Get retrieves a bucket by name
func (r *SQLiteRepository) Get(ctx context.Context, name string) (*Bucket, error) {
	query := `
		SELECT name, owner, created_at, versioning_enabled
		FROM buckets
		WHERE name = ?
	`

	bucket := &Bucket{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&bucket.Name,
		&bucket.Owner,
		&bucket.CreatedAt,
		&bucket.Versioning,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bucket '%s' not found", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	return bucket, nil
}

// List lists all buckets for an owner
func (r *SQLiteRepository) List(ctx context.Context, owner string) ([]*Bucket, error) {
	query := `
		SELECT name, owner, created_at, versioning_enabled
		FROM buckets
		WHERE owner = ?
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}
	defer rows.Close()

	var buckets []*Bucket
	for rows.Next() {
		bucket := &Bucket{}
		err := rows.Scan(
			&bucket.Name,
			&bucket.Owner,
			&bucket.CreatedAt,
			&bucket.Versioning,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bucket: %w", err)
		}
		buckets = append(buckets, bucket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating buckets: %w", err)
	}

	return buckets, nil
}

// Delete deletes a bucket
func (r *SQLiteRepository) Delete(ctx context.Context, name string) error {
	// First check if bucket has objects
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM objects WHERE bucket_name = ?", name).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check bucket objects: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("bucket '%s' is not empty", name)
	}

	// Delete bucket
	result, err := r.db.ExecWithRetry(ctx, "DELETE FROM buckets WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bucket '%s' not found", name)
	}

	return nil
}

// Exists checks if a bucket exists
func (r *SQLiteRepository) Exists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM buckets WHERE name = ?)", name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	return exists, nil
}

// Update updates bucket metadata
func (r *SQLiteRepository) Update(ctx context.Context, bucket *Bucket) error {
	query := `
		UPDATE buckets
		SET versioning_enabled = ?
		WHERE name = ?
	`

	result, err := r.db.ExecWithRetry(ctx, query,
		bucket.Versioning,
		bucket.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to update bucket: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bucket '%s' not found", bucket.Name)
	}

	return nil
}

// isSQLiteConstraintError checks if error is a constraint violation
func isSQLiteConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite constraint errors contain "UNIQUE constraint failed"
	return containsString(err.Error(), "UNIQUE") || containsString(err.Error(), "constraint")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && ((len(s) >= len(substr) && s[:len(substr)] == substr) ||
		(len(s) >= len(substr) && s[len(s)-len(substr):] == substr) ||
		containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
