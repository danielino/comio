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

func TestSlabAllocator_PackingSmallObjects(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(64 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	// Allocate multiple small objects that should pack into same slab
	offsets := make([]int64, 0)
	for i := 0; i < 10; i++ {
		offset, err := alloc.Allocate(100 * 1024) // 100KB each
		if err != nil {
			t.Fatalf("Failed to allocate object %d: %v", i, err)
		}
		offsets = append(offsets, offset)
	}

	// Verify they're all in first slab
	firstSlab := offsets[0] / slabSize
	for i, offset := range offsets {
		slab := offset / slabSize
		if slab != firstSlab {
			t.Errorf("Object %d at different slab %d, expected %d", i, slab, firstSlab)
		}
	}

	// Stats should show usage
	stats := alloc.Stats()
	expected := int64(10 * 100 * 1024)
	if stats.UsedBytes != expected {
		t.Errorf("UsedBytes = %d, want %d", stats.UsedBytes, expected)
	}
}

func TestSlabAllocator_LargeObjectMultipleSlabs(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(64 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	// Allocate object that needs 3 slabs (10MB)
	size := int64(10 * 1024 * 1024)
	offset, err := alloc.Allocate(size)
	if err != nil {
		t.Fatalf("Failed to allocate large object: %v", err)
	}

	// Should be aligned to slab boundary
	if offset%slabSize != 0 {
		t.Errorf("Offset %d not aligned to slab size %d", offset, slabSize)
	}

	// Stats should reflect large allocation
	stats := alloc.Stats()
	if stats.UsedBytes != size {
		t.Errorf("UsedBytes = %d, want %d", stats.UsedBytes, size)
	}
}

func TestSlabAllocator_FreeInvalidOffset(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(64 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	// Try to free unallocated space
	err := alloc.Free(0, 1024)
	if err == nil {
		t.Error("Free() expected error for unallocated space, got nil")
	}
}

func TestSlabAllocator_MultipleAllocationsAndFrees(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(64 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	// Allocate and free in a loop
	for i := 0; i < 5; i++ {
		offset, err := alloc.Allocate(1024 * 1024) // 1MB
		if err != nil {
			t.Fatalf("Iteration %d: Allocate() error = %v", i, err)
		}

		err = alloc.Free(offset, 1024*1024)
		if err != nil {
			t.Errorf("Iteration %d: Free() error = %v", i, err)
		}
	}

	// Should be back to zero usage
	stats := alloc.Stats()
	if stats.UsedBytes != 0 {
		t.Errorf("After all frees UsedBytes = %d, want 0", stats.UsedBytes)
	}
}

func TestSlabAllocator_EdgeCases(t *testing.T) {
	slabSize := int64(4 * 1024 * 1024)
	totalSize := int64(64 * 1024 * 1024)
	alloc := NewSlabAllocator(totalSize, slabSize)

	// Zero size allocation
	_, err := alloc.Allocate(0)
	if err == nil {
		t.Error("Allocate(0) expected error, got nil")
	}

	// Negative size allocation
	_, err = alloc.Allocate(-1)
	if err == nil {
		t.Error("Allocate(-1) expected error, got nil")
	}
}
