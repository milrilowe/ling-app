package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword creates a bcrypt hash of the password.
// bcrypt automatically handles salting - each hash includes a unique salt.
func (s *AuthService) HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword:
	// - Adds a random salt automatically
	// - The cost factor (12) means 2^12 iterations
	// - Returns a string containing: algorithm + cost + salt + hash
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword verifies a password against its hash.
// Returns true if the password matches.
func (s *AuthService) CheckPassword(hash, password string) bool {
	// bcrypt.CompareHashAndPassword extracts the salt from the hash
	// and uses it to hash the provided password, then compares
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
