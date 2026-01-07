package session

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"time"
)

// Session ID format: session-YYYYMMDD-HHMMSS-{8-hex}
var sessionIDPattern = regexp.MustCompile(`^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$`)

// GenerateSessionID generates a new unique session ID.
func GenerateSessionID() string {
	now := time.Now()
	hex := make([]byte, 4)
	rand.Read(hex)
	return fmt.Sprintf("session-%s-%x",
		now.Format("20060102-150405"),
		hex,
	)
}

// IsValidSessionID checks if an ID matches the session ID pattern.
func IsValidSessionID(id string) bool {
	return sessionIDPattern.MatchString(id)
}

// ParseSessionTimestamp extracts the timestamp from a session ID.
// Returns zero time if parsing fails.
func ParseSessionTimestamp(id string) time.Time {
	if len(id) < 24 {
		return time.Time{}
	}
	// Extract "YYYYMMDD-HHMMSS" portion
	dateStr := id[8:23] // "20260104-160414"
	t, err := time.Parse("20060102-150405", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
