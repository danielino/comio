package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashSHA256 calculates SHA256 hash of a string
func HashSHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
