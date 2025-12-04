package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
)

// HMACAuthenticator implements S3-style HMAC authentication.
// Note: This is a simplified implementation for demonstration purposes.
// A production implementation should follow the complete AWS Signature Version 4
// specification including canonical request building, signing key derivation,
// and proper timestamp validation. See:
// https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html
type HMACAuthenticator struct {
	users map[string]*User // accessKeyID -> User
}

// NewHMACAuthenticator creates a new HMAC authenticator
func NewHMACAuthenticator() *HMACAuthenticator {
	return &HMACAuthenticator{
		users: make(map[string]*User),
	}
}

// AddUser adds a user to the authenticator
func (a *HMACAuthenticator) AddUser(user *User) {
	a.users[user.AccessKeyID] = user
}

// Authenticate authenticates a request and returns the user
func (a *HMACAuthenticator) Authenticate(ctx context.Context, req *http.Request) (*User, error) {
	// Get the Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	// Parse AWS4-HMAC-SHA256 authorization header
	// Format: AWS4-HMAC-SHA256 Credential=AKID/date/region/service/aws4_request, SignedHeaders=..., Signature=...
	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
		return nil, errors.New("invalid authorization scheme")
	}

	// Extract Credential to get the access key ID
	credStart := strings.Index(authHeader, "Credential=")
	if credStart == -1 {
		return nil, errors.New("missing Credential in authorization header")
	}
	credEnd := strings.Index(authHeader[credStart:], ",")
	if credEnd == -1 {
		credEnd = len(authHeader) - credStart
	}
	credential := authHeader[credStart+11 : credStart+credEnd]
	
	// Access key ID is the first part before the first slash
	parts := strings.SplitN(credential, "/", 2)
	if len(parts) < 1 {
		return nil, errors.New("invalid Credential format")
	}
	accessKeyID := parts[0]

	// Look up user by access key ID
	user, ok := a.users[accessKeyID]
	if !ok {
		return nil, errors.New("unknown access key")
	}

	// Validate the signature
	if err := a.ValidateSignature(req, user.SecretAccessKey); err != nil {
		return nil, err
	}

	return user, nil
}

// ValidateSignature validates the request signature using AWS Signature V4 style
func (a *HMACAuthenticator) ValidateSignature(req *http.Request, secretKey string) error {
	// Get the Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("missing Authorization header")
	}

	// Parse AWS4-HMAC-SHA256 authorization header
	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
		return errors.New("invalid authorization scheme")
	}

	// Extract signature from header
	// Format: AWS4-HMAC-SHA256 Credential=.../..., SignedHeaders=..., Signature=...
	signatureIdx := strings.Index(authHeader, "Signature=")
	if signatureIdx == -1 {
		return errors.New("missing signature in authorization header")
	}
	providedSignature := authHeader[signatureIdx+10:]

	// Get the string to sign from x-amz-content-sha256 or compute it
	stringToSign := req.Header.Get("X-Amz-Content-Sha256")
	if stringToSign == "" {
		// Compute hash of empty body as fallback
		stringToSign = "UNSIGNED-PAYLOAD"
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(stringToSign))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures in constant time to prevent timing attacks
	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return errors.New("signature mismatch")
	}

	return nil
}

// NewAdminUser creates an admin user with the given credentials
func NewAdminUser(accessKey, secretKey string) *User {
	return &User{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Username:        "admin",
		Policies:        []string{"admin"},
		CreatedAt:       time.Now(),
	}
}
