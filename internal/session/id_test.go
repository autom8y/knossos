package session

import (
	"regexp"
	"testing"
	"time"
)

func TestGenerateSessionID(t *testing.T) {
	id := GenerateSessionID()

	// Should match pattern
	if !IsValidSessionID(id) {
		t.Errorf("GenerateSessionID() = %q, does not match expected pattern", id)
	}

	// Should be unique
	id2 := GenerateSessionID()
	if id == id2 {
		t.Errorf("GenerateSessionID() produced duplicate IDs: %q", id)
	}
}

func TestIsValidSessionID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		// Valid IDs
		{"session-20260104-160414-563c681e", true},
		{"session-20251231-235959-abcdef12", true},
		{"session-20260101-000000-00000000", true},

		// Invalid IDs
		{"session-2026010-160414-563c681e", false},  // short date
		{"session-20260104-16041-563c681e", false},  // short time
		{"session-20260104-160414-563c681", false},  // short hex
		{"session-20260104-160414-563c681eg", false}, // invalid hex
		{"session-20260104-160414-563C681E", false},  // uppercase hex
		{"SESSION-20260104-160414-563c681e", false},  // uppercase session
		{"20260104-160414-563c681e", false},          // missing session prefix
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := IsValidSessionID(tt.id); got != tt.want {
				t.Errorf("IsValidSessionID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestParseSessionTimestamp(t *testing.T) {
	tests := []struct {
		id       string
		wantZero bool
	}{
		{"session-20260104-160414-563c681e", false},
		{"session-20251231-235959-abcdef12", false},
		{"invalid", true},
		{"short", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := ParseSessionTimestamp(tt.id)
			if got.Equal(time.Time{}) != tt.wantZero {
				t.Errorf("ParseSessionTimestamp(%q) = %v, wantZero = %v", tt.id, got, tt.wantZero)
			}
		})
	}

	// Specific timestamp check
	id := "session-20260104-160414-563c681e"
	ts := ParseSessionTimestamp(id)
	expected := time.Date(2026, 1, 4, 16, 4, 14, 0, time.UTC)
	if !ts.Equal(expected) {
		t.Errorf("ParseSessionTimestamp(%q) = %v, want %v", id, ts, expected)
	}
}

func TestSessionIDPattern(t *testing.T) {
	// Ensure our pattern matches the TDD specification
	pattern := regexp.MustCompile(`^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$`)

	validIDs := []string{
		"session-20260104-160414-563c681e",
		"session-20251231-235959-deadbeef",
	}

	for _, id := range validIDs {
		if !pattern.MatchString(id) {
			t.Errorf("Pattern should match valid ID: %q", id)
		}
	}
}
