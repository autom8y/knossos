package session

import "testing"

func TestNormalizeStatus_ValidPassthrough(t *testing.T) {
	tests := []struct {
		input string
		want  Status
	}{
		{"ACTIVE", StatusActive},
		{"PARKED", StatusParked},
		{"ARCHIVED", StatusArchived},
		{"NONE", StatusNone},
	}

	for _, tt := range tests {
		got := NormalizeStatus(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeStatus_AliasMapping(t *testing.T) {
	tests := []struct {
		input string
		want  Status
	}{
		{"COMPLETE", StatusArchived},
		{"COMPLETED", StatusArchived},
		{"complete", StatusArchived},   // case-insensitive
		{"Completed", StatusArchived},  // mixed case
	}

	for _, tt := range tests {
		got := NormalizeStatus(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeStatus_UnknownPassthrough(t *testing.T) {
	// Unknown values pass through as-is (uppercased) for downstream validation
	got := NormalizeStatus("BOGUS")
	if got != Status("BOGUS") {
		t.Errorf("NormalizeStatus(\"BOGUS\") = %q, want %q", got, "BOGUS")
	}
}

func TestNormalizeStatus_WhitespaceAndCase(t *testing.T) {
	tests := []struct {
		input string
		want  Status
	}{
		{"  ACTIVE  ", StatusActive},
		{"active", StatusActive},
		{"Active", StatusActive},
		{"  parked ", StatusParked},
	}

	for _, tt := range tests {
		got := NormalizeStatus(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- DEBT-177: Schema Registry Tests ---

// TestSchemaRegistry_AllCanonicalStatuses_Normalize verifies that every canonical
// session status passes through NormalizeStatus unchanged. If a new Status constant
// is added to the FSM without being registered in NormalizeStatus, this test fails.
//
// This is a schema evolution guard: adding a new status value to the FSM without
// updating NormalizeStatus would cause the new status to fall through to the
// unknown-passthrough path, bypassing alias resolution.
func TestSchemaRegistry_AllCanonicalStatuses_Normalize(t *testing.T) {
	// Exhaustive list of all canonical status constants.
	// Adding a new Status constant requires adding it here -- that is the point.
	allCanonical := []Status{
		StatusNone,
		StatusActive,
		StatusParked,
		StatusArchived,
	}

	for _, status := range allCanonical {
		t.Run(string(status), func(t *testing.T) {
			got := NormalizeStatus(string(status))
			if got != status {
				t.Errorf("NormalizeStatus(%q) = %q, want %q (canonical status must pass through)", status, got, status)
			}
			if !got.IsValid() {
				t.Errorf("NormalizeStatus(%q).IsValid() = false, but canonical statuses must be valid", status)
			}
		})
	}
}

// TestSchemaRegistry_AllAliases_ResolveToCanonical verifies that every entry in
// statusAliases maps to a valid canonical status. If an alias is added that points
// to a typo or removed status, this test fails.
func TestSchemaRegistry_AllAliases_ResolveToCanonical(t *testing.T) {
	for alias, canonical := range statusAliases {
		t.Run(alias, func(t *testing.T) {
			if !canonical.IsValid() {
				t.Errorf("statusAliases[%q] = %q, which is not a valid canonical status", alias, canonical)
			}
			// Verify NormalizeStatus resolves this alias correctly
			got := NormalizeStatus(alias)
			if got != canonical {
				t.Errorf("NormalizeStatus(%q) = %q, want %q (must match statusAliases)", alias, got, canonical)
			}
		})
	}
}

// TestSchemaRegistry_FSMTransitions_UseCanonicalStatuses verifies that the FSM
// transition table only references valid canonical statuses. This catches cases
// where the FSM is extended with a new status that is not registered in the
// status constants.
func TestSchemaRegistry_FSMTransitions_UseCanonicalStatuses(t *testing.T) {
	fsm := NewFSM()
	allCanonical := map[Status]bool{
		StatusNone:     true,
		StatusActive:   true,
		StatusParked:   true,
		StatusArchived: true,
	}

	for from, targets := range fsm.transitions {
		if !allCanonical[from] {
			t.Errorf("FSM transition source %q is not a canonical status", from)
		}
		for _, to := range targets {
			if !allCanonical[to] {
				t.Errorf("FSM transition target %q (from %q) is not a canonical status", to, from)
			}
		}
	}
}

// TestSchemaRegistry_AllPhases_AreValid verifies all Phase constants return true
// from IsValidPhase. Similar evolution guard for workflow phases.
func TestSchemaRegistry_AllPhases_AreValid(t *testing.T) {
	allPhases := []Phase{
		PhaseRequirements,
		PhaseDesign,
		PhaseImplementation,
		PhaseValidation,
		PhaseComplete,
	}

	for _, phase := range allPhases {
		t.Run(string(phase), func(t *testing.T) {
			if !IsValidPhase(string(phase)) {
				t.Errorf("IsValidPhase(%q) = false, but this is a declared Phase constant", phase)
			}
			if PhaseOrder(phase) < 0 {
				t.Errorf("PhaseOrder(%q) = -1, but declared phases must have a valid ordinal", phase)
			}
		})
	}
}
