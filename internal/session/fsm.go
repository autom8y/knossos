package session

import (
	"slices"

	"github.com/autom8y/knossos/internal/errors"
)

// FSM validates and executes state transitions per TLA+ spec.
type FSM struct {
	transitions map[Status][]Status
}

// NewFSM creates a new session finite state machine.
func NewFSM() *FSM {
	return &FSM{
		transitions: map[Status][]Status{
			StatusNone:   {StatusActive},                 // create
			StatusActive: {StatusParked, StatusArchived}, // park, wrap
			StatusParked: {StatusActive, StatusArchived}, // resume, wrap
			// StatusArchived has no valid transitions (terminal)
		},
	}
}

// CanTransition checks if a transition from one status to another is valid.
func (f *FSM) CanTransition(from, to Status) bool {
	validTargets, ok := f.transitions[from]
	if !ok {
		return false
	}
	return slices.Contains(validTargets, to)
}

// ValidateTransition checks a transition and returns an error if invalid.
func (f *FSM) ValidateTransition(from, to Status) error {
	if !f.CanTransition(from, to) {
		return errors.ErrLifecycleViolation(from.String(), to.String(),
			transitionErrorMessage(from, to))
	}
	return nil
}

// transitionErrorMessage returns a human-readable error for invalid transitions.
func transitionErrorMessage(from, to Status) string {
	switch {
	case from == StatusArchived:
		return "cannot transition from archived session (terminal state)"
	case from == StatusParked && to == StatusParked:
		return "session already parked"
	case from == StatusActive && to == StatusActive:
		return "session already active"
	case from == StatusNone && to != StatusActive:
		return "new session must start as active"
	default:
		return "invalid state transition"
	}
}

// ValidTransitions returns all valid target states from the given state.
func (f *FSM) ValidTransitions(from Status) []Status {
	return f.transitions[from]
}

// Phase represents a workflow phase.
type Phase string

const (
	PhaseRequirements   Phase = "requirements"
	PhaseDesign         Phase = "design"
	PhaseImplementation Phase = "implementation"
	PhaseValidation     Phase = "validation"
	PhaseComplete       Phase = "complete"
)

// IsValidPhase checks if a phase value is valid.
func IsValidPhase(phase string) bool {
	switch Phase(phase) {
	case PhaseRequirements, PhaseDesign, PhaseImplementation, PhaseValidation, PhaseComplete:
		return true
	default:
		return false
	}
}

// PhaseOrder returns the ordinal position of a phase.
func PhaseOrder(phase Phase) int {
	switch phase {
	case PhaseRequirements:
		return 0
	case PhaseDesign:
		return 1
	case PhaseImplementation:
		return 2
	case PhaseValidation:
		return 3
	case PhaseComplete:
		return 4
	default:
		return -1
	}
}

// CanTransitionPhase checks if a phase transition is valid.
// Phases must progress forward (requirements -> design -> implementation -> validation -> complete).
func CanTransitionPhase(from, to Phase) bool {
	fromOrder := PhaseOrder(from)
	toOrder := PhaseOrder(to)
	if fromOrder < 0 || toOrder < 0 {
		return false
	}
	// Can only move strictly forward
	return toOrder > fromOrder
}
