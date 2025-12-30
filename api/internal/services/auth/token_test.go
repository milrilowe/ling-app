package auth

import (
	"testing"
)

func TestGenerateSessionToken(t *testing.T) {
	s := &AuthService{}

	t.Run("generates non-empty token", func(t *testing.T) {
		token, err := s.GenerateSessionToken()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		token1, _ := s.GenerateSessionToken()
		token2, _ := s.GenerateSessionToken()
		if token1 == token2 {
			t.Error("expected different tokens on each call")
		}
	})

	t.Run("generates base64url encoded token", func(t *testing.T) {
		token, _ := s.GenerateSessionToken()
		// base64url encoding of 32 bytes = 43 characters (no padding)
		// or 44 characters with padding
		if len(token) < 40 || len(token) > 50 {
			t.Errorf("expected token length around 43-44, got %d", len(token))
		}
	})

	t.Run("generates 1000 unique tokens", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			token, err := s.GenerateSessionToken()
			if err != nil {
				t.Fatalf("error generating token %d: %v", i, err)
			}
			if seen[token] {
				t.Fatalf("duplicate token found at iteration %d", i)
			}
			seen[token] = true
		}
	})
}
