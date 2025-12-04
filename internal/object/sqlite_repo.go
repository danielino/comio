package object

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/danielino/comio/internal/database"
	"github.com/danielino/comio/internal/integrity"
)

// SQLiteRepository implements Repository using SQLite
type SQLiteRepository struct {
	db *database.DB
}

// NewSQLiteRepository creates a new SQLite-based object repository
func NewSQLiteRepository(db *database.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

// Put stores an object metadata (data parameter is ignored - data is in storage engine)
func (r *SQLiteRepository) Put(ctx context.Context, obj *Object, data io.Reader) error {
	// For SQLite repository, we only store metadata
	// The actual data is stored in the storage engine
	// data parameter is ignored - it's for compatibility with the interface

	// Serialize user metadata to JSON (if any)
	var metadataJSON []byte
	if obj.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(obj.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT OR REPLACE INTO objects (
			bucket_name, key, version_id, size, content_type, etag,
			checksum_algorithm, checksum_value, storage_offset,
			created_at, modified_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecWithRetry(ctx, query,
		obj.BucketName,
		obj.Key,
		obj.VersionID,
		obj.Size,
		obj.ContentType,
		obj.ETag,
		obj.Checksum.Algorithm,
		obj.Checksum.Value,
		obj.Offset,
		obj.CreatedAt,
		obj.ModifiedAt,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

// Get retrieves an object metadata (returns nil for data - data is in storage engine)
func (r *SQLiteRepository) Get(ctx context.Context, bucket, key string, versionID *string) (*Object, io.ReadCloser, error) {
	query := `
		SELECT bucket_name, key, version_id, size, content_type, etag,
		       checksum_algorithm, checksum_value, storage_offset,
		       created_at, modified_at, metadata
		FROM objects
		WHERE bucket_name = ? AND key = ?
	`

	args := []interface{}{bucket, key}

	// If version ID specified, filter by it
	if versionID != nil && *versionID != "" {
		query += " AND version_id = ?"
		args = append(args, *versionID)
	} else {
		// Get latest version
		query += " ORDER BY created_at DESC LIMIT 1"
	}

	obj := &Object{}
	var metadataJSON []byte
	var checksumAlg, checksumVal sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&obj.BucketName,
		&obj.Key,
		&obj.VersionID,
		&obj.Size,
		&obj.ContentType,
		&obj.ETag,
		&checksumAlg,
		&checksumVal,
		&obj.Offset,
		&obj.CreatedAt,
		&obj.ModifiedAt,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("object not found")
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Set checksum if present
	if checksumAlg.Valid && checksumVal.Valid {
		obj.Checksum = integrity.Checksum{
			Algorithm: checksumAlg.String,
			Value:     checksumVal.String,
		}
	}

	// Deserialize metadata into object
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &obj.Metadata); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Return nil for data - the actual object data is in the storage engine
	// The service layer will fetch it using obj.Offset and obj.Size
	return obj, nil, nil
}

// List lists objects in a bucket with pagination
func (r *SQLiteRepository) List(ctx context.Context, bucket, prefix string, opts ListOptions) (*ListResult, error) {
	// Base query - get latest version of each object
	query := `
		SELECT o1.bucket_name, o1.key, o1.version_id, o1.size, o1.content_type,
		       o1.etag, o1.checksum_algorithm, o1.checksum_value, o1.storage_offset,
		       o1.created_at, o1.modified_at
		FROM objects o1
		INNER JOIN (
			SELECT bucket_name, key, MAX(created_at) as max_created
			FROM objects
			WHERE bucket_name = ?
	`

	args := []interface{}{bucket}

	// Add prefix filter
	if prefix != "" {
		query += " AND key LIKE ?"
		args = append(args, prefix+"%")
	}

	query += `
			GROUP BY bucket_name, key
		) o2 ON o1.bucket_name = o2.bucket_name
		   AND o1.key = o2.key
		   AND o1.created_at = o2.max_created
	`

	// Add pagination
	if opts.StartAfter != "" {
		query += " AND o1.key > ?"
		args = append(args, opts.StartAfter)
	}

	query += " ORDER BY o1.key"

	// Limit
	maxKeys := opts.MaxKeys
	if maxKeys <= 0 {
		maxKeys = DefaultMaxKeys
	}
	if maxKeys > MaxKeysLimit {
		maxKeys = MaxKeysLimit
	}

	// Fetch one extra to determine if truncated
	query += " LIMIT ?"
	args = append(args, maxKeys+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	defer rows.Close()

	var objects []*Object
	for rows.Next() {
		obj := &Object{}
		var checksumAlg, checksumVal sql.NullString

		err := rows.Scan(
			&obj.BucketName,
			&obj.Key,
			&obj.VersionID,
			&obj.Size,
			&obj.ContentType,
			&obj.ETag,
			&checksumAlg,
			&checksumVal,
			&obj.Offset,
			&obj.CreatedAt,
			&obj.ModifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan object: %w", err)
		}

		// Set checksum if present
		if checksumAlg.Valid && checksumVal.Valid {
			obj.Checksum = integrity.Checksum{
				Algorithm: checksumAlg.String,
				Value:     checksumVal.String,
			}
		}

		objects = append(objects, obj)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating objects: %w", err)
	}

	// Check if truncated
	isTruncated := len(objects) > maxKeys
	if isTruncated {
		objects = objects[:maxKeys]
	}

	result := &ListResult{
		Objects:     objects,
		IsTruncated: isTruncated,
	}

	if isTruncated && len(objects) > 0 {
		result.NextMarker = objects[len(objects)-1].Key
	}

	// Handle common prefixes for delimiter
	if opts.Delimiter != "" {
		result.CommonPrefixes = r.extractCommonPrefixes(objects, prefix, opts.Delimiter)
	}

	return result, nil
}

// extractCommonPrefixes extracts common prefixes based on delimiter
func (r *SQLiteRepository) extractCommonPrefixes(objects []*Object, prefix, delimiter string) []string {
	prefixMap := make(map[string]bool)

	for _, obj := range objects {
		key := obj.Key
		if prefix != "" {
			key = strings.TrimPrefix(key, prefix)
		}

		if idx := strings.Index(key, delimiter); idx >= 0 {
			commonPrefix := prefix + key[:idx+len(delimiter)]
			prefixMap[commonPrefix] = true
		}
	}

	var prefixes []string
	for p := range prefixMap {
		prefixes = append(prefixes, p)
	}

	return prefixes
}

// Delete deletes an object
func (r *SQLiteRepository) Delete(ctx context.Context, bucket, key string, versionID *string) error {
	query := "DELETE FROM objects WHERE bucket_name = ? AND key = ?"
	args := []interface{}{bucket, key}

	if versionID != nil && *versionID != "" {
		query += " AND version_id = ?"
		args = append(args, *versionID)
	}

	result, err := r.db.ExecWithRetry(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("object not found")
	}

	return nil
}

// DeleteAll deletes all objects in a bucket
func (r *SQLiteRepository) DeleteAll(ctx context.Context, bucket string) (int, int64, error) {
	// First get count and total size
	var count int
	var totalSize int64

	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(size), 0) FROM objects WHERE bucket_name = ?",
		bucket).Scan(&count, &totalSize)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to count objects: %w", err)
	}

	// Delete all objects
	_, err = r.db.ExecWithRetry(ctx, "DELETE FROM objects WHERE bucket_name = ?", bucket)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to delete objects: %w", err)
	}

	return count, totalSize, nil
}

// Count returns the number of objects and total size
func (r *SQLiteRepository) Count(ctx context.Context, bucket string) (int, int64, error) {
	var count int
	var totalSize int64

	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(size), 0) FROM objects WHERE bucket_name = ?",
		bucket).Scan(&count, &totalSize)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to count objects: %w", err)
	}

	return count, totalSize, nil
}

// Head retrieves only object metadata (no data)
func (r *SQLiteRepository) Head(ctx context.Context, bucket, key string, versionID *string) (*Object, error) {
	// Head is similar to Get but doesn't return data
	// Since our Get already doesn't return data (only metadata), we can reuse it
	obj, _, err := r.Get(ctx, bucket, key, versionID)
	return obj, err
}
