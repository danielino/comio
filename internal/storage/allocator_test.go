package storage

import (
	"testing"
)

func TestAllocator_Basic(t *testing.T) {
	alloc := NewAllocator(1024, 64)

	// Test allocation
	offset, err := alloc.Allocate(100)
	if err != nil {
		t.Errorf("Allocate() error = %v", err)
	}
	if offset != 0 {
		t.Errorf("First allocation offset = %d, want 0", offset)
	}

	// Test second allocation
	offset2, err := alloc.Allocate(200)
	if err != nil {
		t.Errorf("Allocate() error = %v", err)
	}
	if offset2 != 128 { // 2 blocks * 64 bytes
		t.Errorf("Second allocation offset = %d, want 128", offset2)
	}

	// Test stats
	stats := alloc.Stats()
	if stats.TotalBytes != 1024 {
		t.Errorf("TotalBytes = %d, want 1024", stats.TotalBytes)
	}
	if stats.UsedBytes < 300 {
		t.Errorf("UsedBytes = %d, should be at least 300", stats.UsedBytes)
	}
}

func TestAllocator_Free(t *testing.T) {
	alloc := NewAllocator(1024, 64)

	offset, _ := alloc.Allocate(100)

	// Free the allocation
	err := alloc.Free(offset, 100)
	if err != nil {
		t.Errorf("Free() error = %v", err)
	}

	// Check stats reflect the free
	stats := alloc.Stats()
	if stats.UsedBytes != 0 {
		t.Errorf("UsedBytes after free = %d, want 0", stats.UsedBytes)
	}
}

func TestAllocator_OutOfSpace(t *testing.T) {
	alloc := NewAllocator(256, 64)

	// First allocation uses most space
	_, err := alloc.Allocate(200)
	if err != nil {
		t.Errorf("Allocate() error = %v", err)
	}

	// This should fail
	_, err = alloc.Allocate(100)
	if err == nil {
		t.Error("Allocate() expected out of space error, got nil")
	}
}

func TestAllocator_InvalidFree(t *testing.T) {
	alloc := NewAllocator(1024, 64)

	// Try to free unaligned offset
	err := alloc.Free(50, 100)
	if err == nil {
		t.Error("Free() invalid offset should error, got nil")
	}
}

func TestAllocator_ZeroSize(t *testing.T) {
	alloc := NewAllocator(1024, 64)

	offset, err := alloc.Allocate(1) // Allocate 1 byte (will use 1 block)
	if err != nil {
		t.Errorf("Allocate(1) error = %v", err)
	}
	if offset < 0 {
		t.Errorf("Allocate(1) offset = %d, want >= 0", offset)
	}
}
