// Package pathutil provides path-related utility functions
package pathutil

import (
	"path/filepath"
	"strings"
)

// SanitizePath sanitizes a path component to be filesystem-safe.
// It prevents path traversal attacks by:
// 1. Replacing path separators (/ and \) with underscores
// 2. Replacing parent directory references (..) with underscores
// 3. Cleaning the path to remove redundant separators
func SanitizePath(s string) string {
	// Replace various forms of path traversal
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "..", "_")
	
	// Clean the path to normalize any remaining issues
	s = filepath.Clean(s)
	
	// Ensure we don't return a path that goes above current directory
	if strings.HasPrefix(s, "..") || strings.HasPrefix(s, "/") || strings.HasPrefix(s, "\\") {
		s = "_" + s
	}
	
	return s
}
