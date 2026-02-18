// Package session provides core session domain logic for Ariadne.
package session

import "strings"

// Status represents session lifecycle state.
type Status string

const (
	// StatusNone indicates no session exists.
	StatusNone Status = "NONE"
	// StatusActive indicates an active session.
	StatusActive Status = "ACTIVE"
	// StatusParked indicates a suspended session.
	StatusParked Status = "PARKED"
	// StatusArchived indicates a completed session (terminal).
	StatusArchived Status = "ARCHIVED"
)

// String returns the string representation.
func (s Status) String() string {
	return string(s)
}

// IsValid checks if the status is a valid value.
func (s Status) IsValid() bool {
	switch s {
	case StatusNone, StatusActive, StatusParked, StatusArchived:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if this is a terminal state.
func (s Status) IsTerminal() bool {
	return s == StatusArchived
}

// statusAliases maps known non-FSM status names to their canonical equivalents.
// These phantom values appear in real session files written by legacy scripts
// or Moirai prompts that used informal status names.
var statusAliases = map[string]Status{
	"COMPLETE":  StatusArchived,
	"COMPLETED": StatusArchived,
}

// NormalizeStatus maps known status aliases to canonical FSM values.
// Valid FSM statuses pass through unchanged. Unknown values are returned
// as-is for downstream validation to catch.
func NormalizeStatus(raw string) Status {
	s := Status(strings.ToUpper(strings.TrimSpace(raw)))
	if s.IsValid() {
		return s
	}
	if canonical, ok := statusAliases[string(s)]; ok {
		return canonical
	}
	return s
}
