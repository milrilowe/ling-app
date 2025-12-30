package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	s := &AuthService{bcryptCost: 10} // Lower cost for faster tests

	t.Run("hashes password successfully", func(t *testing.T) {
		hash, err := s.HashPassword("password123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if hash == "" {
			t.Error("expected non-empty hash")
		}
		if hash == "password123" {
			t.Error("hash should not equal plaintext password")
		}
	})

	t.Run("produces different hashes for same password", func(t *testing.T) {
		hash1, _ := s.HashPassword("password123")
		hash2, _ := s.HashPassword("password123")
		if hash1 == hash2 {
			t.Error("expected different hashes due to random salt")
		}
	})

	t.Run("handles empty password", func(t *testing.T) {
		hash, err := s.HashPassword("")
		if err != nil {
			t.Fatalf("expected no error for empty password, got %v", err)
		}
		if hash == "" {
			t.Error("expected non-empty hash even for empty password")
		}
	})
}

func TestCheckPassword(t *testing.T) {
	s := &AuthService{bcryptCost: 10}

	t.Run("returns true for correct password", func(t *testing.T) {
		hash, _ := s.HashPassword("password123")
		if !s.CheckPassword(hash, "password123") {
			t.Error("expected CheckPassword to return true for correct password")
		}
	})

	t.Run("returns false for incorrect password", func(t *testing.T) {
		hash, _ := s.HashPassword("password123")
		if s.CheckPassword(hash, "wrongpassword") {
			t.Error("expected CheckPassword to return false for incorrect password")
		}
	})

	t.Run("returns false for empty password when hash is not empty", func(t *testing.T) {
		hash, _ := s.HashPassword("password123")
		if s.CheckPassword(hash, "") {
			t.Error("expected CheckPassword to return false for empty password")
		}
	})

	t.Run("handles empty password hash correctly", func(t *testing.T) {
		hash, _ := s.HashPassword("")
		if !s.CheckPassword(hash, "") {
			t.Error("expected CheckPassword to return true for empty password matching empty hash")
		}
	})

	t.Run("returns false for invalid hash format", func(t *testing.T) {
		if s.CheckPassword("not-a-valid-hash", "password123") {
			t.Error("expected CheckPassword to return false for invalid hash")
		}
	})
}
