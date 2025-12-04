package storage

import (
	"os"
	"testing"
)

func TestSimpleEngine_AllocateFree(t *testing.T) {
	// Create temp file
	f, err := os.CreateTemp("", "engine_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	// Create engine
	totalSize := int64(64 * 1024 * 1024)
	blockSize := 4 * 1024 * 1024
	engine, err := NewSimpleEngine(f.Name(), totalSize, blockSize)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if err := engine.Open(f.Name()); err != nil {
		t.Fatalf("Failed to open engine: %v", err)
	}

	// Allocate
	size := int64(1024 * 1024)
	offset, err := engine.Allocate(size)
	if err != nil {
		t.Errorf("Allocate() error = %v", err)
	}

	if offset < 0 {
		t.Errorf("Allocate() returned negative offset: %d", offset)
	}

	// Free
	if err := engine.Free(offset, size); err != nil {
		t.Errorf("Free() error = %v", err)
	}
}

func TestSimpleEngine_ReadWrite(t *testing.T) {
	f, err := os.CreateTemp("", "engine_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	engine, err := NewSimpleEngine(f.Name(), 64*1024*1024, 4*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if err := engine.Open(f.Name()); err != nil {
		t.Fatalf("Failed to open engine: %v", err)
	}

	// Allocate
	size := int64(1024)
	offset, err := engine.Allocate(size)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}

	// Write
	data := []byte("test data")
	if err := engine.Write(offset, data); err != nil {
		t.Errorf("Write() error = %v", err)
	}

	// Read
	read, err := engine.Read(offset, int64(len(data)))
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}

	if string(read) != string(data) {
		t.Errorf("Read() = %s, want %s", read, data)
	}
}

func TestSimpleEngine_Stats(t *testing.T) {
	f, err := os.CreateTemp("", "engine_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	totalSize := int64(64 * 1024 * 1024)
	engine, err := NewSimpleEngine(f.Name(), totalSize, 4*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if err := engine.Open(f.Name()); err != nil {
		t.Fatalf("Failed to open engine: %v", err)
	}

	stats := engine.Stats()
	if stats.TotalBytes != totalSize {
		t.Errorf("Stats.TotalBytes = %d, want %d", stats.TotalBytes, totalSize)
	}

	if stats.UsedBytes != 0 {
		t.Errorf("Stats.UsedBytes = %d, want 0", stats.UsedBytes)
	}
}

func TestSimpleEngine_BlockSize(t *testing.T) {
	f, err := os.CreateTemp("", "engine_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	blockSize := 4 * 1024 * 1024
	engine, err := NewSimpleEngine(f.Name(), 64*1024*1024, blockSize)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if engine.BlockSize() != blockSize {
		t.Errorf("BlockSize() = %d, want %d", engine.BlockSize(), blockSize)
	}
}

func TestSimpleEngine_Sync(t *testing.T) {
	f, err := os.CreateTemp("", "engine_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	engine, err := NewSimpleEngine(f.Name(), 64*1024*1024, 4*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if err := engine.Open(f.Name()); err != nil {
		t.Fatalf("Failed to open engine: %v", err)
	}

	// Write some data
	offset, err := engine.Allocate(1024)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}

	data := []byte("test data")
	if err := engine.Write(offset, data); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Sync should not error
	if err := engine.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}
