package auth

import (
	"testing"
	"time"
)

func TestUser_Creation(t *testing.T) {
	user := &User{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Username:        "testuser",
		Policies:        []string{"read", "write"},
		CreatedAt:       time.Now(),
	}

	if user.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("AccessKeyID = %s, want AKIAIOSFODNN7EXAMPLE", user.AccessKeyID)
	}
	if user.Username != "testuser" {
		t.Errorf("Username = %s, want testuser", user.Username)
	}
	if len(user.Policies) != 2 {
		t.Errorf("Policies count = %d, want 2", len(user.Policies))
	}
}

func TestUser_EmptyPolicies(t *testing.T) {
	user := &User{
		AccessKeyID: "test",
		Policies:    []string{},
	}

	if user.Policies == nil {
		t.Error("Policies should not be nil")
	}
	if len(user.Policies) != 0 {
		t.Errorf("Policies count = %d, want 0", len(user.Policies))
	}
}
