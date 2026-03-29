package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword returns a bcrypt hash of the plain-text password
func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPassword compares a bcrypt hashed password with its plain-text version
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
