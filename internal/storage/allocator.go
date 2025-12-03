package storage

import (
	"errors"
	"sync"
)

// Allocator handles space allocation
type Allocator struct {
	totalBlocks int64
	usedBlocks  int64
	bitmap      []bool // Simple in-memory bitmap for now. In prod, this should be persisted.
	mu          sync.Mutex
	blockSize   int
}

// Stats holds storage statistics
type Stats struct {
	TotalBytes int64
	UsedBytes  int64
	FreeBytes  int64
}

// NewAllocator creates a new allocator
func NewAllocator(totalSize int64, blockSize int) *Allocator {
	totalBlocks := totalSize / int64(blockSize)
	return &Allocator{
		totalBlocks: totalBlocks,
		bitmap:      make([]bool, totalBlocks),
		blockSize:   blockSize,
	}
}

// Allocate allocates space for the given size
// Returns the offset in bytes
func (a *Allocator) Allocate(size int64) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	blocksNeeded := (size + int64(a.blockSize) - 1) / int64(a.blockSize)
	
	// First-fit strategy
	startBlock := int64(-1)
	consecutive := int64(0)
	
	for i := int64(0); i < a.totalBlocks; i++ {
		if !a.bitmap[i] {
			if startBlock == -1 {
				startBlock = i
			}
			consecutive++
			if consecutive == blocksNeeded {
				// Found space
				for j := startBlock; j < startBlock+blocksNeeded; j++ {
					a.bitmap[j] = true
				}
				a.usedBlocks += blocksNeeded
				return startBlock * int64(a.blockSize), nil
			}
		} else {
			startBlock = -1
			consecutive = 0
		}
	}
	
	return 0, errors.New("out of space")
}

// Free frees space at the given offset
func (a *Allocator) Free(offset, size int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if offset%int64(a.blockSize) != 0 {
		return errors.New("offset not aligned")
	}
	
	startBlock := offset / int64(a.blockSize)
	blocksToFree := (size + int64(a.blockSize) - 1) / int64(a.blockSize)
	
	if startBlock+blocksToFree > a.totalBlocks {
		return errors.New("invalid free range")
	}
	
	for i := startBlock; i < startBlock+blocksToFree; i++ {
		a.bitmap[i] = false
	}
	
	a.usedBlocks -= blocksToFree
	return nil
}

// Stats returns allocation statistics
func (a *Allocator) Stats() Stats {
	a.mu.Lock()
	defer a.mu.Unlock()

	return Stats{
		TotalBytes: a.totalBlocks * int64(a.blockSize),
		UsedBytes:  a.usedBlocks * int64(a.blockSize),
		FreeBytes:  (a.totalBlocks - a.usedBlocks) * int64(a.blockSize),
	}
}
