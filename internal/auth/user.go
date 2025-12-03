package auth

import (
	"time"
)

// User represents an authenticated user
type User struct {
	AccessKeyID     string    `json:"access_key_id"`
	SecretAccessKey string    `json:"secret_access_key"`
	Username        string    `json:"username"`
	Policies        []string  `json:"policies"`
	CreatedAt       time.Time `json:"created_at"`
}
