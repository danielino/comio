package auth

import (
	"context"
	"net/http"
)

// Authenticator defines the authentication interface
type Authenticator interface {
	Authenticate(ctx context.Context, req *http.Request) (*User, error)
	ValidateSignature(req *http.Request, secretKey string) error
}
