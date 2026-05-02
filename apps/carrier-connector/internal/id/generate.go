package id

import (
	"fmt"
	"time"
)


func GenerateRuleID() string {
	return fmt.Sprintf("rule_%d", GetCurrentTime().UnixNano())
}

// GetCurrentTime returns the current time
func GetCurrentTime() time.Time {
	return time.Now()
}

func GenerateEventID() string {
	// Generate unique event ID (implementation depends on your ID generation strategy)
	return "evt_" + GenerateRandomString(16)
}

func GenerateRandomString(length int) string {
	// Generate random string (implementation depends on your random string generation)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
