package storage

// Engine defines the storage engine interface
type Engine interface {
	Open(devicePath string) error
	Close() error
	Read(offset, size int64) ([]byte, error)
	Write(offset int64, data []byte) error
	Allocate(size int64) (offset int64, err error)
	Free(offset, size int64) error
	Sync() error
	Stats() Stats
	BlockSize() int
}
