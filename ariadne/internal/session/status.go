// Package session provides core session domain logic for Ariadne.
package session

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
