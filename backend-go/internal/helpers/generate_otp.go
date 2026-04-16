package helpers

import (
	"crypto/rand"
	"io"
)

// GenerateOTP generates a random numeric OTP of the specified length
func GenerateOTP(length int) (string, error) {
	table := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, length)
	n, err := io.ReadAtLeast(rand.Reader, b, length)
	if n != length || err != nil {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}

// GenerateNumericOTP is a convenience function for a 6-digit OTP
func GenerateNumericOTP() (string, error) {
	return GenerateOTP(6)
}
