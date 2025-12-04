package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// DB wraps sql.DB with application-specific methods
type DB struct {
	*sql.DB
	path string
}

// Config holds database configuration
type Config struct {
	Path string // Database file path
}

// Open opens a database connection and runs migrations
func Open(cfg Config) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	// Use modernc.org/sqlite (pure Go, no CGO)
	sqlDB, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite
	// WAL mode supports concurrent readers, but only 1 writer at a time
	// Keep pool small to avoid too many connections trying to write
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		DB:   sqlDB,
		path: cfg.Path,
	}

	// Run migrations
	if err := db.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return err
	}

	// Set busy timeout to 5 seconds (auto-retry on lock)
	// This makes SQLite wait instead of immediately returning SQLITE_BUSY
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		return err
	}

	// Optimize for concurrent writes
	// NORMAL is safe with WAL mode and much faster than FULL
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return err
	}

	// Increase cache size for better performance (default is ~2MB, set to 20MB)
	if _, err := db.Exec("PRAGMA cache_size = -20000"); err != nil {
		return err
	}

	// Store temp tables in memory for speed
	if _, err := db.Exec("PRAGMA temp_store = MEMORY"); err != nil {
		return err
	}

	// Create migrations table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	// Get current version
	var currentVersion int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM migrations").Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Apply migrations
	migrations := []struct {
		version int
		sql     string
	}{
		{
			version: 1,
			sql: `
				-- Buckets table
				CREATE TABLE buckets (
					name TEXT PRIMARY KEY,
					owner TEXT NOT NULL,
					created_at TIMESTAMP NOT NULL,
					versioning_enabled BOOLEAN DEFAULT FALSE
				);
				CREATE INDEX idx_buckets_owner ON buckets(owner);

				-- Objects table
				CREATE TABLE objects (
					bucket_name TEXT NOT NULL,
					key TEXT NOT NULL,
					version_id TEXT NOT NULL,
					size INTEGER NOT NULL,
					content_type TEXT,
					etag TEXT,
					checksum_algorithm TEXT,
					checksum_value TEXT,
					storage_offset INTEGER NOT NULL,
					created_at TIMESTAMP NOT NULL,
					modified_at TIMESTAMP NOT NULL,
					metadata TEXT, -- JSON
					PRIMARY KEY (bucket_name, key, version_id),
					FOREIGN KEY (bucket_name) REFERENCES buckets(name) ON DELETE CASCADE
				);

				CREATE INDEX idx_objects_bucket ON objects(bucket_name);
				CREATE INDEX idx_objects_key ON objects(bucket_name, key);
				CREATE INDEX idx_objects_created ON objects(created_at);
			`,
		},
		{
			version: 2,
			sql: `
				-- Add index for listing objects with prefix
				CREATE INDEX idx_objects_prefix ON objects(bucket_name, key);
			`,
		},
	}

	// Apply pending migrations
	for _, m := range migrations {
		if m.version <= currentVersion {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", m.version, err)
		}

		// Execute migration SQL
		if _, err := tx.Exec(m.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", m.version, err)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO migrations (version) VALUES (?)", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.version, err)
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Path returns the database file path
func (db *DB) Path() string {
	return db.path
}

// Stats returns database statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// ExecWithRetry executes a query with automatic retry on SQLITE_BUSY
func (db *DB) ExecWithRetry(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := db.ExecContext(ctx, query, args...)
		if err == nil {
			return result, nil
		}

		// Check if it's a busy error
		if isSQLiteBusy(err) {
			lastErr = err
			// Exponential backoff: 10ms, 20ms, 40ms
			backoff := time.Duration(10*(1<<uint(attempt))) * time.Millisecond
			time.Sleep(backoff)
			continue
		}

		// Not a busy error, return immediately
		return nil, err
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// QueryRowWithRetry queries a single row with automatic retry on SQLITE_BUSY
func (db *DB) QueryRowWithRetry(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// sql.Row doesn't return error until Scan(), so we can't retry here
	// The busy_timeout pragma should handle this
	return db.QueryRowContext(ctx, query, args...)
}

// isSQLiteBusy checks if an error is SQLITE_BUSY
func isSQLiteBusy(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsAny(errStr, "database is locked", "SQLITE_BUSY", "database locked")
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
