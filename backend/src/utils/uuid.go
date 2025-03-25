package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateUUID generates a random UUID
func GenerateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		// If random read fails, use timestamp-based fallback
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	
	// Set version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 1
	
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
