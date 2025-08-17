package utils

import (
	"crypto/sha256"
	"fmt"
)

// GenerateID generates a unique ID from a place ID using SHA256 hash
func GenerateID(placeID string) string {
	hash := sha256.Sum256([]byte(placeID))
	return fmt.Sprintf("%x", hash)[:12]
}
