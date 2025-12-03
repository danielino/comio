package integrity

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"hash/crc32"
	"io"
)

// Checksum holds checksum information
type Checksum struct {
	Algorithm string
	Value     string
}

// Calculator handles checksum calculation
type Calculator struct {
	md5    hash.Hash
	sha256 hash.Hash
	crc32  hash.Hash32
}

// NewCalculator creates a new checksum calculator
func NewCalculator() *Calculator {
	return &Calculator{
		md5:    md5.New(),
		sha256: sha256.New(),
		crc32:  crc32.New(crc32.MakeTable(crc32.Castagnoli)),
	}
}

// Write implements io.Writer to update all hashes
func (c *Calculator) Write(p []byte) (n int, err error) {
	n, err = c.md5.Write(p)
	if err != nil {
		return n, err
	}
	_, _ = c.sha256.Write(p)
	_, _ = c.crc32.Write(p)
	return n, nil
}

// Sums returns all calculated checksums
func (c *Calculator) Sums() map[string]string {
	return map[string]string{
		"MD5":    hex.EncodeToString(c.md5.Sum(nil)),
		"SHA256": hex.EncodeToString(c.sha256.Sum(nil)),
		"CRC32":  hex.EncodeToString(c.crc32.Sum(nil)),
	}
}

// CalculateChecksum calculates checksum for a reader
func CalculateChecksum(r io.Reader, algo string) (string, error) {
	var h hash.Hash
	switch algo {
	case "MD5":
		h = md5.New()
	case "SHA256":
		h = sha256.New()
	case "CRC32":
		h = crc32.New(crc32.MakeTable(crc32.Castagnoli))
	default:
		return "", io.ErrUnexpectedEOF
	}

	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
