package pathutil

import "testing"

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal path",
			input:    "bucket-name",
			expected: "bucket-name",
		},
		{
			name:     "path with slashes",
			input:    "path/to/file",
			expected: "path_to_file",
		},
		{
			name:     "path with backslashes",
			input:    "path\\to\\file",
			expected: "path_to_file",
		},
		{
			name:     "path traversal attempt",
			input:    "../../../etc/passwd",
			expected: "______etc_passwd",
		},
		{
			name:     "mixed unsafe characters",
			input:    "my/path\\..\\test",
			expected: "my_path___test",
		},
		{
			name:     "empty string",
			input:    "",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePath(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
