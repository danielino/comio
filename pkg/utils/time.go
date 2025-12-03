package utils

import (
	"time"
)

// Now returns current time in UTC
func Now() time.Time {
	return time.Now().UTC()
}
