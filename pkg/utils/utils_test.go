package utils

import (
"testing"
"time"
)

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"test", 64},
		{"", 64},
		{"hello world", 64},
	}

	for _, tt := range tests {
		got := HashSHA256(tt.input)
		if len(got) != tt.want {
			t.Errorf("HashSHA256(%q) length = %d, want %d", tt.input, len(got), tt.want)
		}
	}

	hash1 := HashSHA256("test")
	hash2 := HashSHA256("test")
	if hash1 != hash2 {
		t.Error("Hash should be consistent")
	}

	hash3 := HashSHA256("different")
	if hash1 == hash3 {
		t.Error("Different inputs should produce different hashes")
	}
}

func TestNow(t *testing.T) {
	now1 := Now()
	time.Sleep(1 * time.Millisecond)
	now2 := Now()

	if now2.Before(now1) {
		t.Error("Now() should return increasing time")
	}

	if now1.Location() != time.UTC {
		t.Error("Now() should return UTC time")
	}
}
