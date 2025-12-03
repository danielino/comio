package storage

import (
	"bytes"
	"os"
	"testing"
)

func TestDevice_ReadWrite(t *testing.T) {
	// Create temp file
	f, err := os.CreateTemp("", "device_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Initialize with 1MB
	if err := f.Truncate(1024 * 1024); err != nil {
		t.Fatalf("Failed to truncate: %v", err)
	}

	// Create device
	dev := NewDevice(f.Name(), 4096)
	if err := dev.Open(); err != nil {
		t.Fatalf("Failed to open device: %v", err)
	}
	defer dev.Close()

	// Test data
	data := []byte("Hello, World!")
	offset := int64(0)

	// Write
	if err := dev.Write(offset, data); err != nil {
		t.Errorf("Write() error = %v", err)
	}

	// Read
	read, err := dev.Read(offset, int64(len(data)))
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}

	if !bytes.Equal(read, data) {
		t.Errorf("Read() = %v, want %v", read, data)
	}
}

func TestDevice_ReadWriteLargeData(t *testing.T) {
	f, err := os.CreateTemp("", "device_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// 10MB file
	size := int64(10 * 1024 * 1024)
	if err := f.Truncate(size); err != nil {
		t.Fatalf("Failed to truncate: %v", err)
	}

	dev := NewDevice(f.Name(), 4096)
	if err := dev.Open(); err != nil {
		t.Fatalf("Failed to open device: %v", err)
	}
	defer dev.Close()

	// Write 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	offset := int64(4 * 1024 * 1024) // 4MB offset

	if err := dev.Write(offset, data); err != nil {
		t.Errorf("Write() error = %v", err)
	}

	read, err := dev.Read(offset, int64(len(data)))
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}

	if !bytes.Equal(read, data) {
		t.Error("Read data does not match written data")
	}
}

func TestDevice_MultipleWrites(t *testing.T) {
	f, err := os.CreateTemp("", "device_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err := f.Truncate(1024 * 1024); err != nil {
		t.Fatalf("Failed to truncate: %v", err)
	}

	dev := NewDevice(f.Name(), 4096)
	if err := dev.Open(); err != nil {
		t.Fatalf("Failed to open device: %v", err)
	}
	defer dev.Close()

	// Write multiple chunks
	chunks := []struct {
		offset int64
		data   []byte
	}{
		{0, []byte("first")},
		{1024, []byte("second")},
		{2048, []byte("third")},
	}

	for _, chunk := range chunks {
		if err := dev.Write(chunk.offset, chunk.data); err != nil {
			t.Errorf("Write() at offset %d error = %v", chunk.offset, err)
		}
	}

	// Verify each chunk
	for _, chunk := range chunks {
		read, err := dev.Read(chunk.offset, int64(len(chunk.data)))
		if err != nil {
			t.Errorf("Read() at offset %d error = %v", chunk.offset, err)
		}
		if !bytes.Equal(read, chunk.data) {
			t.Errorf("Read() at offset %d = %v, want %v", chunk.offset, read, chunk.data)
		}
	}
}

func TestDevice_OpenClose(t *testing.T) {
	f, err := os.CreateTemp("", "device_test_*.dat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	dev := NewDevice(f.Name(), 4096)

	// Open
	if err := dev.Open(); err != nil {
		t.Errorf("Open() error = %v", err)
	}

	// Close
	if err := dev.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Reopen
	if err := dev.Open(); err != nil {
		t.Errorf("Reopen() error = %v", err)
	}
	dev.Close()
}

func BenchmarkDevice_Write(b *testing.B) {
	f, err := os.CreateTemp("", "device_bench_*.dat")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err := f.Truncate(100 * 1024 * 1024); err != nil {
		b.Fatalf("Failed to truncate: %v", err)
	}

	dev := NewDevice(f.Name(), 4096)
	if err := dev.Open(); err != nil {
		b.Fatalf("Failed to open device: %v", err)
	}
	defer dev.Close()

	data := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := int64(i%1000) * 4096
		if err := dev.Write(offset, data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDevice_Read(b *testing.B) {
	f, err := os.CreateTemp("", "device_bench_*.dat")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err := f.Truncate(100 * 1024 * 1024); err != nil {
		b.Fatalf("Failed to truncate: %v", err)
	}

	dev := NewDevice(f.Name(), 4096)
	if err := dev.Open(); err != nil {
		b.Fatalf("Failed to open device: %v", err)
	}
	defer dev.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := int64(i%1000) * 4096
		if _, err := dev.Read(offset, 4096); err != nil {
			b.Fatal(err)
		}
	}
}
