// Package suggest generates contextual suggestions from session state.
// All generator functions are pure: struct in, slice out, no I/O.
// This package is consumed by hook commands that attach suggestions to their output.
package suggest

// Kind classifies a suggestion for filtering and presentation.
type Kind string

const (
	KindSessionStart     Kind = "session_start"
	KindPhaseTransition  Kind = "phase_transition"
	KindBudgetWarning    Kind = "budget_warning"
	KindSubagentComplete Kind = "subagent_complete"
)

// Suggestion is a single proactive recommendation.
type Suggestion struct {
	Kind   Kind   `json:"kind"`
	Text   string `json:"text"`
	Action string `json:"action,omitempty"` // suggested command, e.g., "/task", "/park"
	Reason string `json:"reason,omitempty"` // why this suggestion was generated
}

// SessionInput holds the state needed to generate session-start and budget suggestions.
// All fields are optional (fail-open: zero-value fields produce no suggestions).
type SessionInput struct {
	SessionID     string
	Initiative    string
	Phase         string // "requirements", "design", "implementation", "validation"
	Rite          string
	Complexity    string // "TASK", "MODULE", "INITIATIVE"
	ParkSource    string // non-empty if session was resumed from park
	StrandCount   int    // number of active strands
	ToolCount     int    // from budget hook state
	WarnThreshold int    // budget warn threshold
	ParkThreshold int    // budget park threshold (0 = disabled)
}

// SubagentInput holds the state from a SubagentStop event.
type SubagentInput struct {
	AgentName string
	AgentType string
	Phase     string // current session phase
	Rite      string
}

// PhaseTransitionInput holds state for detecting phase boundaries.
type PhaseTransitionInput struct {
	PreviousPhase string
	CurrentPhase  string
	Rite          string
	Complexity    string
}
