package storage

import (
	"errors"
	"sync"
)

// BlockManager handles block-level operations
type BlockManager struct {
	blockSize int
	device    *Device
	mu        sync.RWMutex
}

// NewBlockManager creates a new block manager
func NewBlockManager(device *Device, blockSize int) *BlockManager {
	return &BlockManager{
		blockSize: blockSize,
		device:    device,
	}
}

// ReadBlock reads a single block
func (bm *BlockManager) ReadBlock(blockIndex int64) ([]byte, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	offset := blockIndex * int64(bm.blockSize)
	return bm.device.Read(offset, int64(bm.blockSize))
}

// WriteBlock writes a single block
func (bm *BlockManager) WriteBlock(blockIndex int64, data []byte) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if len(data) != bm.blockSize {
		return errors.New("data size does not match block size")
	}

	offset := blockIndex * int64(bm.blockSize)
	return bm.device.Write(offset, data)
}
