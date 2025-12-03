package object

import (
	"github.com/google/uuid"
)

// GenerateVersionID generates a new version ID
func GenerateVersionID() string {
	return uuid.New().String()
}
