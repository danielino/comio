// Package pathutil provides path-related utility functions
package pathutil

import "strings"

// SanitizePath sanitizes a path component to be filesystem-safe
// It replaces unsafe characters that could cause path traversal attacks
func SanitizePath(s string) string {
	// Replace unsafe characters with underscores
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "..", "_")
	return s
}
