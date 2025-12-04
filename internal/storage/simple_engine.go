package storage

import (
	"sync"
)

const (
	// DefaultBlockSize is the default block size for storage allocation (4MB for performance)
	DefaultBlockSize = 4 * 1024 * 1024 // 4MB
)

// SimpleEngine implements Engine using slab allocation
type SimpleEngine struct {
	device    *Device
	allocator *SlabAllocator
	blockMgr  *BlockManager
	slabSize  int64
	mu        sync.RWMutex // Protects concurrent access to device operations
}

// NewSimpleEngine creates a new simple engine with slab allocation
func NewSimpleEngine(devicePath string, size int64, slabSize int) (*SimpleEngine, error) {
	device := NewDevice(devicePath, slabSize)
	allocator := NewSlabAllocator(size, int64(slabSize))
	blockMgr := NewBlockManager(device, slabSize)

	return &SimpleEngine{
		device:    device,
		allocator: allocator,
		blockMgr:  blockMgr,
		slabSize:  int64(slabSize),
	}, nil
}

func (e *SimpleEngine) Open(devicePath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.device.Open()
}

func (e *SimpleEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.device.Close()
}

func (e *SimpleEngine) Read(offset, size int64) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.device.Read(offset, size)
}

func (e *SimpleEngine) Write(offset int64, data []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.device.Write(offset, data)
}

func (e *SimpleEngine) Allocate(size int64) (int64, error) {
	// SlabAllocator has its own internal mutex for thread safety.
	// Allocation is independent of device I/O operations, so no engine lock needed.
	return e.allocator.Allocate(size)
}

func (e *SimpleEngine) Free(offset, size int64) error {
	// SlabAllocator has its own internal mutex for thread safety.
	// Freeing is independent of device I/O operations, so no engine lock needed.
	return e.allocator.Free(offset, size)
}

func (e *SimpleEngine) Sync() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.device.Sync()
}

func (e *SimpleEngine) Stats() Stats {
	// Allocator has its own lock
	return e.allocator.Stats()
}

func (e *SimpleEngine) BlockSize() int {
	return int(e.slabSize)
}
