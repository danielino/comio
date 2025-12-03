package storage

import (
	"testing"
)

func TestSlabAllocator_Allocate(t *testing.T) {
	tests := []struct {
		name      string
		slabSize  int64
		totalSize int64
		allocSize int64
		wantErr   bool
	}{
		{
			name:      "small object in single slab",
			slabSize:  4 * 1024 * 1024,
			totalSize: 64 * 1024 * 1024,
			allocSize: 1024,
			wantErr:   false,
		},
		{
			name:      "large object multiple slabs",
			slabSize:  4 * 1024 * 1024,
			totalSize: 64 * 1024 * 1024,
			allocSize: 10 * 1024 * 1024,
			wantErr:   false,
		},
		{
			name:      "allocation exceeds total size",
			slabSize:  4 * 1024 * 1024,
			totalSize: 8 * 1024 * 1024,
			allocSize: 10 * 1024 * 1024,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alloc := NewSlabAllocator(tt.totalSize, tt.slabSize)
			offset, err := alloc.Allocate(tt.allocSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("Allocate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && offset < 0 {
				t.Errorf("Allocate() returned negative offset: %d", offset)
			}
		})
	}
}

func TestSlabAllocator_Free(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(64 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	offset1, err := alloc.Allocate(1024)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}

	err = alloc.Free(offset1, 1024)
	if err != nil {
		t.Errorf("Free() error = %v", err)
	}

	// After free, stats should reflect freed space
	stats := alloc.Stats()
	if stats.UsedBytes != 0 {
		t.Errorf("After free UsedBytes = %d, want 0", stats.UsedBytes)
	}

	// Should be able to allocate again
	offset2, err := alloc.Allocate(2048)
	if err != nil {
		t.Fatalf("Failed to allocate after free: %v", err)
	}

	if offset2 < 0 {
		t.Errorf("Allocate() after free returned negative offset: %d", offset2)
	}
}

func TestSlabAllocator_Stats(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(64 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	stats := alloc.Stats()
	if stats.UsedBytes != 0 {
		t.Errorf("Initial UsedBytes = %d, want 0", stats.UsedBytes)
	}

	size := int64(1024 * 1024)
	offset, err := alloc.Allocate(size)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}

	stats = alloc.Stats()
	if stats.UsedBytes != size {
		t.Errorf("After allocation UsedBytes = %d, want %d", stats.UsedBytes, size)
	}

	alloc.Free(offset, size)

	stats = alloc.Stats()
	if stats.UsedBytes != 0 {
		t.Errorf("After free UsedBytes = %d, want 0", stats.UsedBytes)
	}
}

func TestSlabAllocator_OutOfSpace(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(8 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	_, err := alloc.Allocate(totalSize)
	if err != nil {
		t.Fatalf("Failed to allocate full size: %v", err)
	}

	_, err = alloc.Allocate(1024)
	if err == nil {
		t.Error("Expected out of space error, got nil")
	}
}
