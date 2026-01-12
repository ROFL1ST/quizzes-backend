package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateToken membuat string random 64 karakter (hex)
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
