package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockAuthenticator struct {
	authenticateFunc      func(ctx context.Context, req *http.Request) (*User, error)
	validateSignatureFunc func(req *http.Request, secretKey string) error
}

func (m *MockAuthenticator) Authenticate(ctx context.Context, req *http.Request) (*User, error) {
	if m.authenticateFunc != nil {
		return m.authenticateFunc(ctx, req)
	}
	return &User{AccessKeyID: "test-key"}, nil
}

func (m *MockAuthenticator) ValidateSignature(req *http.Request, secretKey string) error {
	if m.validateSignatureFunc != nil {
		return m.validateSignatureFunc(req, secretKey)
	}
	return nil
}

func TestAuthenticator_Interface(t *testing.T) {
	var _ Authenticator = (*MockAuthenticator)(nil)
}

func TestMockAuthenticator_Authenticate(t *testing.T) {
	mock := &MockAuthenticator{}
	req := httptest.NewRequest("GET", "/", nil)

	user, err := mock.Authenticate(context.Background(), req)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
	}
	if user == nil {
		t.Error("Authenticate() returned nil user")
	}
	if user.AccessKeyID != "test-key" {
		t.Errorf("User.AccessKeyID = %s, want test-key", user.AccessKeyID)
	}
}

func TestMockAuthenticator_ValidateSignature(t *testing.T) {
	mock := &MockAuthenticator{}
	req := httptest.NewRequest("GET", "/", nil)

	err := mock.ValidateSignature(req, "secret")
	if err != nil {
		t.Errorf("ValidateSignature() error = %v", err)
	}
}

func TestHMACAuthenticator_ValidateSignature(t *testing.T) {
	auth := &HMACAuthenticator{}
	req := httptest.NewRequest("GET", "/bucket/key", nil)

	err := auth.ValidateSignature(req, "test-secret")
	if err != nil {
		t.Errorf("ValidateSignature() error = %v", err)
	}
}
