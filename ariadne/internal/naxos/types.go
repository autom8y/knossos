// Package naxos provides cleanup tooling for abandoned sessions.
// Named after the island where Theseus abandoned Ariadne in Greek mythology.
// In Knossos, Naxos represents the cleanup mechanism for abandoned sessions.
package naxos

import "time"

// OrphanReason describes why a session is flagged as orphaned.
type OrphanReason string

const (
	// ReasonInactive - session has been inactive for longer than threshold.
	ReasonInactive OrphanReason = "INACTIVE"
	// ReasonStaleSails - session has gray sails older than threshold.
	ReasonStaleSails OrphanReason = "STALE_SAILS"
	// ReasonIncompleteWrap - session was marked for wrap but never completed.
	ReasonIncompleteWrap OrphanReason = "INCOMPLETE_WRAP"
)

// String returns the string representation.
func (r OrphanReason) String() string {
	return string(r)
}

// Description returns a human-readable description of the reason.
func (r OrphanReason) Description() string {
	switch r {
	case ReasonInactive:
		return "Inactive for too long"
	case ReasonStaleSails:
		return "Gray sails past threshold"
	case ReasonIncompleteWrap:
		return "Wrap started but not completed"
	default:
		return "Unknown reason"
	}
}

// SuggestedAction describes what action to take on an orphaned session.
type SuggestedAction string

const (
	// ActionWrap - session should be wrapped and archived.
	ActionWrap SuggestedAction = "WRAP"
	// ActionResume - session should be resumed to continue work.
	ActionResume SuggestedAction = "RESUME"
	// ActionDelete - session can be safely deleted (no useful state).
	ActionDelete SuggestedAction = "DELETE"
)

// String returns the string representation.
func (a SuggestedAction) String() string {
	return string(a)
}

// Description returns a human-readable description of the action.
func (a SuggestedAction) Description() string {
	switch a {
	case ActionWrap:
		return "ari session wrap"
	case ActionResume:
		return "ari session resume"
	case ActionDelete:
		return "rm -rf <session-dir>"
	default:
		return "Unknown action"
	}
}

// OrphanedSession represents a session that has been flagged for cleanup review.
type OrphanedSession struct {
	// SessionID is the unique identifier for the session.
	SessionID string `json:"session_id"`
	// SessionDir is the full path to the session directory.
	SessionDir string `json:"session_dir"`
	// Status is the current session status (ACTIVE, PARKED, etc.).
	Status string `json:"status"`
	// Initiative is the session's stated goal/initiative.
	Initiative string `json:"initiative"`
	// Reason is why this session was flagged.
	Reason OrphanReason `json:"reason"`
	// SuggestedAction is what action is recommended.
	SuggestedAction SuggestedAction `json:"suggested_action"`
	// Age is how old the session is since creation.
	Age time.Duration `json:"age"`
	// InactiveFor is how long since the session was last active.
	InactiveFor time.Duration `json:"inactive_for"`
	// CreatedAt is when the session was created.
	CreatedAt time.Time `json:"created_at"`
	// LastActivity is the last known activity timestamp.
	LastActivity time.Time `json:"last_activity"`
	// SailsColor is the current sails color (if known).
	SailsColor string `json:"sails_color,omitempty"`
	// AdditionalInfo contains any extra context about why this was flagged.
	AdditionalInfo string `json:"additional_info,omitempty"`
}

// ScanConfig holds configuration for the session scanner.
type ScanConfig struct {
	// InactiveThreshold is how long a session can be inactive before flagging.
	// Default: 24 hours
	InactiveThreshold time.Duration
	// StaleSailsThreshold is how long gray sails can persist before flagging.
	// Default: 7 days
	StaleSailsThreshold time.Duration
	// IncludeArchived controls whether to scan archived sessions.
	// Default: false
	IncludeArchived bool
}

// DefaultConfig returns the default scan configuration.
func DefaultConfig() ScanConfig {
	return ScanConfig{
		InactiveThreshold:   24 * time.Hour,
		StaleSailsThreshold: 7 * 24 * time.Hour,
		IncludeArchived:     false,
	}
}

// ScanResult holds the results of a session scan.
type ScanResult struct {
	// OrphanedSessions is the list of sessions flagged for review.
	OrphanedSessions []OrphanedSession `json:"orphaned_sessions"`
	// TotalScanned is how many sessions were examined.
	TotalScanned int `json:"total_scanned"`
	// TotalOrphaned is how many sessions were flagged.
	TotalOrphaned int `json:"total_orphaned"`
	// ScannedAt is when the scan was performed.
	ScannedAt time.Time `json:"scanned_at"`
	// Config is the configuration used for the scan.
	Config ScanConfig `json:"config"`
	// ByReason breaks down orphaned sessions by reason.
	ByReason map[OrphanReason]int `json:"by_reason"`
}

// NewScanResult creates a new empty scan result.
func NewScanResult(config ScanConfig) *ScanResult {
	return &ScanResult{
		OrphanedSessions: []OrphanedSession{},
		ScannedAt:        time.Now().UTC(),
		Config:           config,
		ByReason:         make(map[OrphanReason]int),
	}
}

// Add adds an orphaned session to the result.
func (r *ScanResult) Add(session OrphanedSession) {
	r.OrphanedSessions = append(r.OrphanedSessions, session)
	r.TotalOrphaned++
	r.ByReason[session.Reason]++
}
