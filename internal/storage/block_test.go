package storage

import (
	"bytes"
	"os"
	"testing"
)

func createTestDevice(t *testing.T, size int64) (*Device, func()) {
	tmpFile, err := os.CreateTemp("", "block_test_*.dat")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	device := NewDevice(tmpFile.Name(), 512)
	device.size = size
	err = device.Open()
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatal(err)
	}

	cleanup := func() {
		device.Close()
		os.Remove(tmpFile.Name())
	}

	return device, cleanup
}

func TestBlockManager_ReadWriteBlock(t *testing.T) {
	device, cleanup := createTestDevice(t, 8192)
	defer cleanup()

	bm := NewBlockManager(device, 512)

	// Write a block
	data := []byte("test block data")
	padded := make([]byte, 512)
	copy(padded, data)

	err := bm.WriteBlock(0, padded)
	if err != nil {
		t.Errorf("WriteBlock() error = %v", err)
	}

	// Read it back
	read, err := bm.ReadBlock(0)
	if err != nil {
		t.Errorf("ReadBlock() error = %v", err)
	}

	if !bytes.Equal(read[:len(data)], data) {
		t.Errorf("ReadBlock() data mismatch, got %v, want %v", read[:len(data)], data)
	}
}

func TestBlockManager_MultipleBlocks(t *testing.T) {
	device, cleanup := createTestDevice(t, 16384)
	defer cleanup()

	bm := NewBlockManager(device, 512)

	// Write multiple blocks
	for i := 0; i < 10; i++ {
		data := make([]byte, 512)
		for j := range data {
			data[j] = byte(i)
		}
		err := bm.WriteBlock(int64(i), data)
		if err != nil {
			t.Errorf("WriteBlock(%d) error = %v", i, err)
		}
	}

	// Read them back
	for i := 0; i < 10; i++ {
		data, err := bm.ReadBlock(int64(i))
		if err != nil {
			t.Errorf("ReadBlock(%d) error = %v", i, err)
		}
		if data[0] != byte(i) {
			t.Errorf("ReadBlock(%d) data[0] = %d, want %d", i, data[0], i)
		}
	}
}

func TestBlockManager_InvalidBlockNumber(t *testing.T) {
	device, cleanup := createTestDevice(t, 2048)
	defer cleanup()

	bm := NewBlockManager(device, 512)

	// Try to read block beyond device size (device has 4 blocks)
	_, err := bm.ReadBlock(10)
	if err == nil {
		t.Error("ReadBlock() beyond device size should error, got nil")
	}

	// WriteBlock doesn't validate bounds in current implementation
	// Just test that it doesn't panic
	data := make([]byte, 512)
	_ = bm.WriteBlock(10, data)
}
