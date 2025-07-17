package datetime

import (
	"time"
)

// IsExpired checks if a file has expired based on its expiration timestamp
func IsExpired(expiresAtStr string) bool {
	if expiresAtStr == "" {
		return false // No expiration set
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return false // Invalid timestamp, treat as not expired
	}

	return time.Now().After(expiresAt)
}
