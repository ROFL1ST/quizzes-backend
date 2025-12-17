package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken membuat string random 64 karakter (hex)
func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}