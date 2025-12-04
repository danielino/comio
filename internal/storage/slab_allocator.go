package storage

import (
	"errors"
	"sync"
)

// SlabAllocator allocates space within large blocks (slabs)
// Objects smaller than slab size are packed together
type SlabAllocator struct {
	slabSize   int64
	totalSize  int64
	slabs      map[int64]*Slab // Key: slab offset
	usedBytes  int64
	nextOffset int64
	mu         sync.Mutex
}

// Slab represents a large block that can contain multiple objects
type Slab struct {
	offset    int64
	size      int64
	used      int64
	fragments []Fragment
}

// Fragment represents a portion of a slab used by an object
type Fragment struct {
	offset int64
	size   int64
}

// NewSlabAllocator creates a new slab-based allocator
func NewSlabAllocator(totalSize int64, slabSize int64) *SlabAllocator {
	return &SlabAllocator{
		slabSize:   slabSize,
		totalSize:  totalSize,
		slabs:      make(map[int64]*Slab),
		nextOffset: 0,
	}
}

// Allocate allocates space for an object
// Small objects are packed into existing slabs
// Large objects (>= slabSize) get dedicated slabs
func (a *SlabAllocator) Allocate(size int64) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if size <= 0 {
		return 0, errors.New("invalid size")
	}

	// For large objects, allocate dedicated slab(s)
	if size >= a.slabSize {
		slabsNeeded := (size + a.slabSize - 1) / a.slabSize
		totalSize := slabsNeeded * a.slabSize

		if a.nextOffset+totalSize > a.totalSize {
			return 0, errors.New("out of space")
		}

		offset := a.nextOffset
		a.slabs[offset] = &Slab{
			offset:    offset,
			size:      totalSize,
			used:      size,
			fragments: []Fragment{{offset: offset, size: size}},
		}
		a.nextOffset += totalSize
		a.usedBytes += size
		return offset, nil
	}

	// For small objects, try to pack into existing slab with available space
	// We need to check slabs in deterministic order
	var slabOffsets []int64
	for off := range a.slabs {
		slabOffsets = append(slabOffsets, off)
	}

	for _, off := range slabOffsets {
		slab := a.slabs[off]
		// Only pack into slabs that were created for small objects (size == slabSize)
		if slab.size == a.slabSize && slab.used+size <= slab.size {
			// Found space in existing slab
			fragmentOffset := slab.offset + slab.used
			slab.fragments = append(slab.fragments, Fragment{
				offset: fragmentOffset,
				size:   size,
			})
			slab.used += size
			a.usedBytes += size
			return fragmentOffset, nil
		}
	}

	// No existing slab has space, allocate new slab
	if a.nextOffset+a.slabSize > a.totalSize {
		return 0, errors.New("out of space")
	}

	offset := a.nextOffset
	slab := &Slab{
		offset:    offset,
		size:      a.slabSize,
		used:      size,
		fragments: []Fragment{{offset: offset, size: size}},
	}
	a.slabs[offset] = slab
	a.nextOffset += a.slabSize
	a.usedBytes += size
	return offset, nil
}

// Free frees allocated space
func (a *SlabAllocator) Free(offset, size int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Find which slab contains this offset
	var targetSlab *Slab

	for _, slab := range a.slabs {
		if offset >= slab.offset && offset < slab.offset+slab.size {
			targetSlab = slab
			break
		}
	}

	if targetSlab == nil {
		return errors.New("offset not found")
	}

	// Remove fragment
	for i, frag := range targetSlab.fragments {
		if frag.offset == offset && frag.size == size {
			targetSlab.fragments = append(targetSlab.fragments[:i], targetSlab.fragments[i+1:]...)
			targetSlab.used -= size
			a.usedBytes -= size

			// Keep empty slabs so they can be reused for small objects
			// Do NOT delete them, as we can't reclaim the space before nextOffset anyway

			return nil
		}
	}

	return errors.New("fragment not found")
}

// Stats returns allocation statistics
func (a *SlabAllocator) Stats() Stats {
	a.mu.Lock()
	defer a.mu.Unlock()

	// The real free space is the space after nextOffset
	// We can't reuse space before nextOffset even if slabs are deleted
	freeSpace := a.totalSize - a.nextOffset

	return Stats{
		TotalBytes: a.totalSize,
		UsedBytes:  a.usedBytes,
		FreeBytes:  freeSpace,
	}
}
