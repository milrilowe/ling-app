package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSessionToken creates a cryptographically secure random token.
// Returns a 32-byte random value encoded as base64url (43 characters).
func (s *AuthService) GenerateSessionToken() (string, error) {
	// crypto/rand uses the OS's cryptographic random source
	// (e.g., /dev/urandom on Linux)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// base64url encoding is URL-safe (no +, /, or = padding issues)
	return base64.URLEncoding.EncodeToString(bytes), nil
}
