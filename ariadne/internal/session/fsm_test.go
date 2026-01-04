package session

import (
	"testing"
)

func TestFSM_CanTransition(t *testing.T) {
	fsm := NewFSM()

	tests := []struct {
		from, to Status
		want     bool
	}{
		// Valid transitions
		{StatusNone, StatusActive, true},     // create
		{StatusActive, StatusParked, true},   // park
		{StatusActive, StatusArchived, true}, // wrap from active
		{StatusParked, StatusActive, true},   // resume
		{StatusParked, StatusArchived, true}, // wrap from parked

		// Invalid transitions
		{StatusArchived, StatusActive, false},  // can't resume archived
		{StatusArchived, StatusParked, false},  // can't park archived
		{StatusActive, StatusNone, false},      // can't go back to none
		{StatusActive, StatusActive, false},    // can't self-transition
		{StatusParked, StatusParked, false},    // can't self-transition
		{StatusNone, StatusParked, false},      // can't create as parked
		{StatusNone, StatusArchived, false},    // can't create as archived
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			got := fsm.CanTransition(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CanTransition(%s, %s) = %v, want %v",
					tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestFSM_ValidateTransition(t *testing.T) {
	fsm := NewFSM()

	// Valid transition should return nil
	if err := fsm.ValidateTransition(StatusActive, StatusParked); err != nil {
		t.Errorf("ValidateTransition(ACTIVE, PARKED) returned error: %v", err)
	}

	// Invalid transition should return error
	if err := fsm.ValidateTransition(StatusArchived, StatusActive); err == nil {
		t.Errorf("ValidateTransition(ARCHIVED, ACTIVE) should return error")
	}
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusNone, true},
		{StatusActive, true},
		{StatusParked, true},
		{StatusArchived, true},
		{Status("INVALID"), false},
		{Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("Status(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusNone, false},
		{StatusActive, false},
		{StatusParked, false},
		{StatusArchived, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.want {
				t.Errorf("Status(%q).IsTerminal() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsValidPhase(t *testing.T) {
	tests := []struct {
		phase string
		want  bool
	}{
		{"requirements", true},
		{"design", true},
		{"implementation", true},
		{"validation", true},
		{"complete", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			if got := IsValidPhase(tt.phase); got != tt.want {
				t.Errorf("IsValidPhase(%q) = %v, want %v", tt.phase, got, tt.want)
			}
		})
	}
}

func TestCanTransitionPhase(t *testing.T) {
	tests := []struct {
		from, to Phase
		want     bool
	}{
		// Forward transitions
		{PhaseRequirements, PhaseDesign, true},
		{PhaseDesign, PhaseImplementation, true},
		{PhaseImplementation, PhaseValidation, true},
		{PhaseValidation, PhaseComplete, true},
		{PhaseRequirements, PhaseComplete, true}, // can skip

		// Backward transitions (invalid)
		{PhaseDesign, PhaseRequirements, false},
		{PhaseImplementation, PhaseDesign, false},
		{PhaseComplete, PhaseValidation, false},

		// Same phase (invalid)
		{PhaseDesign, PhaseDesign, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			if got := CanTransitionPhase(tt.from, tt.to); got != tt.want {
				t.Errorf("CanTransitionPhase(%s, %s) = %v, want %v",
					tt.from, tt.to, got, tt.want)
			}
		})
	}
}
