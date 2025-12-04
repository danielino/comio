package integrity

import (
	"bytes"
	"testing"
)

func TestNewCalculator(t *testing.T) {
	calc := NewCalculator()
	if calc == nil {
		t.Error("NewCalculator() returned nil")
	}
	if calc.md5 == nil {
		t.Error("md5 hash is nil")
	}
	if calc.sha256 == nil {
		t.Error("sha256 hash is nil")
	}
	if calc.crc32 == nil {
		t.Error("crc32 hash is nil")
	}
}

func TestCalculator_Write(t *testing.T) {
	calc := NewCalculator()
	data := []byte("test data")

	n, err := calc.Write(data)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("Write() n = %d, want %d", n, len(data))
	}
}

func TestCalculator_Sums(t *testing.T) {
	calc := NewCalculator()
	data := []byte("test data")
	calc.Write(data)

	sums := calc.Sums()
	if len(sums) != 3 {
		t.Errorf("Sums() count = %d, want 3", len(sums))
	}
	if _, ok := sums["MD5"]; !ok {
		t.Error("MD5 checksum missing")
	}
	if _, ok := sums["SHA256"]; !ok {
		t.Error("SHA256 checksum missing")
	}
	if _, ok := sums["CRC32"]; !ok {
		t.Error("CRC32 checksum missing")
	}
}

func TestCalculateChecksum_MD5(t *testing.T) {
	data := bytes.NewReader([]byte("test data"))
	checksum, err := CalculateChecksum(data, "MD5")
	if err != nil {
		t.Errorf("CalculateChecksum() error = %v", err)
	}
	if checksum == "" {
		t.Error("CalculateChecksum() returned empty string")
	}
}

func TestCalculateChecksum_SHA256(t *testing.T) {
	data := bytes.NewReader([]byte("test data"))
	checksum, err := CalculateChecksum(data, "SHA256")
	if err != nil {
		t.Errorf("CalculateChecksum() error = %v", err)
	}
	if checksum == "" {
		t.Error("CalculateChecksum() returned empty string")
	}
}

func TestCalculateChecksum_CRC32(t *testing.T) {
	data := bytes.NewReader([]byte("test data"))
	checksum, err := CalculateChecksum(data, "CRC32")
	if err != nil {
		t.Errorf("CalculateChecksum() error = %v", err)
	}
	if checksum == "" {
		t.Error("CalculateChecksum() returned empty string")
	}
}

func TestCalculateChecksum_Invalid(t *testing.T) {
	data := bytes.NewReader([]byte("test data"))
	_, err := CalculateChecksum(data, "INVALID")
	if err == nil {
		t.Error("CalculateChecksum() with invalid algorithm should return error")
	}
}

func TestCalculateChecksum_Consistency(t *testing.T) {
	testData := []byte("consistent test data")

	data1 := bytes.NewReader(testData)
	checksum1, _ := CalculateChecksum(data1, "MD5")

	data2 := bytes.NewReader(testData)
	checksum2, _ := CalculateChecksum(data2, "MD5")

	if checksum1 != checksum2 {
		t.Errorf("Checksums not consistent: %s != %s", checksum1, checksum2)
	}
}
