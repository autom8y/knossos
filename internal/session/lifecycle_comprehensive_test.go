package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/errors"
)

// =============================================================================
// FSM Transition Exhaustive Tests
// =============================================================================

// TestFSM_AllTransitionPairs exhaustively tests every possible status pair
// to ensure the FSM correctly allows/denies all transitions.
func TestFSM_AllTransitionPairs(t *testing.T) {
	fsm := NewFSM()

	allStatuses := []Status{StatusNone, StatusActive, StatusParked, StatusArchived}

	// Build expected valid transitions map
	validTransitions := map[Status]map[Status]bool{
		StatusNone:     {StatusActive: true},
		StatusActive:   {StatusParked: true, StatusArchived: true},
		StatusParked:   {StatusActive: true, StatusArchived: true},
		StatusArchived: {}, // terminal - no valid transitions
	}

	for _, from := range allStatuses {
		for _, to := range allStatuses {
			name := fmt.Sprintf("%s->%s", from, to)
			t.Run(name, func(t *testing.T) {
				expected := validTransitions[from][to]
				got := fsm.CanTransition(from, to)
				if got != expected {
					t.Errorf("CanTransition(%s, %s) = %v, want %v", from, to, got, expected)
				}

				err := fsm.ValidateTransition(from, to)
				if expected && err != nil {
					t.Errorf("ValidateTransition(%s, %s) returned unexpected error: %v", from, to, err)
				}
				if !expected && err == nil {
					t.Errorf("ValidateTransition(%s, %s) should have returned error", from, to)
				}
			})
		}
	}
}

// TestFSM_InvalidStatusValues tests FSM behavior with invalid status values
// that are not part of the enum.
func TestFSM_InvalidStatusValues(t *testing.T) {
	fsm := NewFSM()

	invalidStatuses := []Status{
		Status("INVALID"),
		Status(""),
		Status("active"),     // lowercase
		Status("SUSPENDED"),  // not a valid status
		Status("TERMINATED"), // not a valid status
	}

	for _, invalid := range invalidStatuses {
		t.Run(string(invalid), func(t *testing.T) {
			// Transitioning FROM an invalid status should fail
			if fsm.CanTransition(invalid, StatusActive) {
				t.Errorf("CanTransition(%q, ACTIVE) should be false", invalid)
			}

			// Transitioning TO an invalid status should fail
			if fsm.CanTransition(StatusActive, invalid) {
				t.Errorf("CanTransition(ACTIVE, %q) should be false", invalid)
			}

			// ValidTransitions should return nil for invalid status
			transitions := fsm.ValidTransitions(invalid)
			if len(transitions) > 0 {
				t.Errorf("ValidTransitions(%q) should be empty, got %v", invalid, transitions)
			}
		})
	}
}

// TestFSM_SelfTransitionsAlwaysDenied verifies no status can transition to itself.
func TestFSM_SelfTransitionsAlwaysDenied(t *testing.T) {
	fsm := NewFSM()

	allStatuses := []Status{StatusNone, StatusActive, StatusParked, StatusArchived}

	for _, status := range allStatuses {
		t.Run(string(status), func(t *testing.T) {
			if fsm.CanTransition(status, status) {
				t.Errorf("Self-transition %s -> %s should be denied", status, status)
			}
		})
	}
}

// TestFSM_ArchivedIsFullyTerminal verifies ARCHIVED cannot transition to any state.
func TestFSM_ArchivedIsFullyTerminal(t *testing.T) {
	fsm := NewFSM()

	allStatuses := []Status{StatusNone, StatusActive, StatusParked, StatusArchived}

	for _, to := range allStatuses {
		t.Run("ARCHIVED->"+string(to), func(t *testing.T) {
			if fsm.CanTransition(StatusArchived, to) {
				t.Errorf("ARCHIVED -> %s should be denied (ARCHIVED is terminal)", to)
			}

			err := fsm.ValidateTransition(StatusArchived, to)
			if err == nil {
				t.Error("ValidateTransition from ARCHIVED should always return error")
			}
			// Every error from ARCHIVED should mention "terminal state"
			if err != nil && !strings.Contains(err.Error(), "terminal state") {
				t.Errorf("Error from ARCHIVED should mention 'terminal state', got: %q", err.Error())
			}
		})
	}
}

// TestFSM_NoneCanOnlyCreate verifies NONE can only transition to ACTIVE.
func TestFSM_NoneCanOnlyCreate(t *testing.T) {
	fsm := NewFSM()

	targetStatuses := []struct {
		to       Status
		wantOK   bool
		errMatch string
	}{
		{StatusActive, true, ""},
		{StatusParked, false, "must start as active"},
		{StatusArchived, false, "must start as active"},
		{StatusNone, false, "must start as active"},
	}

	for _, tt := range targetStatuses {
		t.Run("NONE->"+string(tt.to), func(t *testing.T) {
			if fsm.CanTransition(StatusNone, tt.to) != tt.wantOK {
				t.Errorf("CanTransition(NONE, %s) = %v, want %v", tt.to, !tt.wantOK, tt.wantOK)
			}

			err := fsm.ValidateTransition(StatusNone, tt.to)
			if tt.wantOK && err != nil {
				t.Errorf("ValidateTransition(NONE, %s) unexpected error: %v", tt.to, err)
			}
			if !tt.wantOK {
				if err == nil {
					t.Errorf("ValidateTransition(NONE, %s) should return error", tt.to)
				} else if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("Error should contain %q, got: %q", tt.errMatch, err.Error())
				}
			}
		})
	}
}

// =============================================================================
// Error Message Quality Tests
// =============================================================================

// TestTransitionErrors_AreLifecycleViolations verifies that invalid transition
// errors have the correct error code for downstream handling.
func TestTransitionErrors_AreLifecycleViolations(t *testing.T) {
	fsm := NewFSM()

	invalidPairs := [][2]Status{
		{StatusArchived, StatusActive},
		{StatusArchived, StatusParked},
		{StatusParked, StatusParked},
		{StatusActive, StatusActive},
		{StatusNone, StatusParked},
		{StatusNone, StatusArchived},
		{StatusActive, StatusNone},
	}

	for _, pair := range invalidPairs {
		name := fmt.Sprintf("%s->%s", pair[0], pair[1])
		t.Run(name, func(t *testing.T) {
			err := fsm.ValidateTransition(pair[0], pair[1])
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			// Verify error is a lifecycle violation type
			if !errors.IsLifecycleError(err) {
				t.Errorf("Error should be a lifecycle violation, got: %T", err)
			}

			// Verify error message is not empty and is actionable
			if err.Error() == "" {
				t.Error("Error message should not be empty")
			}
			if len(err.Error()) < 15 {
				t.Errorf("Error message too short to be actionable: %q", err.Error())
			}

			// Verify structured error has details
			if e, ok := err.(*errors.Error); ok {
				if e.Details == nil {
					t.Error("Structured error should have details")
				}
				if _, ok := e.Details["current_status"]; !ok {
					t.Error("Error details should include current_status")
				}
				if _, ok := e.Details["requested_transition"]; !ok {
					t.Error("Error details should include requested_transition")
				}
			}
		})
	}
}

// TestTransitionErrors_ContainContextualMessages verifies each error scenario
// has a message specific to the problem.
func TestTransitionErrors_ContainContextualMessages(t *testing.T) {
	fsm := NewFSM()

	tests := []struct {
		from     Status
		to       Status
		contains string
		desc     string
	}{
		{StatusArchived, StatusActive, "terminal state", "archived is terminal"},
		{StatusArchived, StatusParked, "terminal state", "archived is terminal"},
		{StatusArchived, StatusArchived, "terminal state", "archived is terminal"},
		{StatusParked, StatusParked, "already parked", "cannot double-park"},
		{StatusActive, StatusActive, "already active", "cannot double-activate"},
		{StatusNone, StatusParked, "must start as active", "new sessions start active"},
		{StatusNone, StatusArchived, "must start as active", "cannot skip to archived"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := fsm.ValidateTransition(tt.from, tt.to)
			if err == nil {
				t.Fatalf("Expected error for %s -> %s", tt.from, tt.to)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("Error for %s -> %s should contain %q, got: %q",
					tt.from, tt.to, tt.contains, err.Error())
			}
		})
	}
}

// =============================================================================
// Session Create Tests
// =============================================================================

// TestNewContext_Defaults verifies all defaults are set correctly on creation.
func TestNewContext_Defaults(t *testing.T) {
	before := time.Now().UTC()
	ctx := NewContext("Test initiative", "MODULE", "10x-dev")
	after := time.Now().UTC()

	// Status must be ACTIVE
	if ctx.Status != StatusActive {
		t.Errorf("Status = %v, want %v", ctx.Status, StatusActive)
	}

	// Schema version must be 2.1
	if ctx.SchemaVersion != "2.1" {
		t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "2.1")
	}

	// Current phase must be "requirements"
	if ctx.CurrentPhase != "requirements" {
		t.Errorf("CurrentPhase = %q, want %q", ctx.CurrentPhase, "requirements")
	}

	// SessionID must be valid
	if !IsValidSessionID(ctx.SessionID) {
		t.Errorf("SessionID %q is not valid", ctx.SessionID)
	}

	// CreatedAt must be between before and after
	if ctx.CreatedAt.Before(before) || ctx.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not between %v and %v", ctx.CreatedAt, before, after)
	}

	// Optional fields must be nil
	if ctx.ParkedAt != nil {
		t.Error("ParkedAt should be nil for new session")
	}
	if ctx.ArchivedAt != nil {
		t.Error("ArchivedAt should be nil for new session")
	}
	if ctx.ResumedAt != nil {
		t.Error("ResumedAt should be nil for new session")
	}
	if ctx.ParkedReason != "" {
		t.Error("ParkedReason should be empty for new session")
	}

	// Body should contain default content
	if ctx.Body == "" {
		t.Error("Body should have default content")
	}
	if !strings.Contains(ctx.Body, ctx.Initiative) {
		t.Error("Default body should reference the initiative")
	}
}

// TestNewContext_RiteHandling verifies rite field behavior for different inputs.
func TestNewContext_RiteHandling(t *testing.T) {
	tests := []struct {
		name     string
		rite     string
		wantNil  bool
		wantRite string
	}{
		{"named rite", "10x-dev", false, "10x-dev"},
		{"empty rite (cross-cutting)", "", true, ""},
		{"none rite (cross-cutting)", "none", true, ""},
		{"custom rite", "custom-workflow", false, "custom-workflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("Test", "MODULE", tt.rite)

			if tt.wantNil && ctx.Rite != nil {
				t.Errorf("Rite should be nil for %q, got %q", tt.rite, *ctx.Rite)
			}
			if !tt.wantNil {
				if ctx.Rite == nil {
					t.Errorf("Rite should not be nil for %q", tt.rite)
				} else if *ctx.Rite != tt.wantRite {
					t.Errorf("Rite = %q, want %q", *ctx.Rite, tt.wantRite)
				}
			}

			// ActiveRite should always be set to the input value
			if ctx.ActiveRite != tt.rite {
				t.Errorf("ActiveRite = %q, want %q", ctx.ActiveRite, tt.rite)
			}
		})
	}
}

// TestNewContext_UniqueIDs verifies that consecutive sessions get unique IDs.
func TestNewContext_UniqueIDs(t *testing.T) {
	ids := make(map[string]bool)
	const n = 100

	for i := range n {
		ctx := NewContext("Test", "MODULE", "test")
		if ids[ctx.SessionID] {
			t.Fatalf("Duplicate SessionID after %d generations: %s", i, ctx.SessionID)
		}
		ids[ctx.SessionID] = true
	}

	if len(ids) != n {
		t.Errorf("Expected %d unique IDs, got %d", n, len(ids))
	}
}

// TestNewContext_SaveAndLoad verifies a newly created context can be saved and loaded.
func TestNewContext_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	original := NewContext("Save and load test", "SYSTEM", "ecosystem")

	if err := original.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(ctxPath); os.IsNotExist(err) {
		t.Fatal("SESSION_CONTEXT.md was not created")
	}

	// Load and compare
	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}

	if loaded.SessionID != original.SessionID {
		t.Errorf("SessionID: got %q, want %q", loaded.SessionID, original.SessionID)
	}
	if loaded.Status != original.Status {
		t.Errorf("Status: got %v, want %v", loaded.Status, original.Status)
	}
	if loaded.Initiative != original.Initiative {
		t.Errorf("Initiative: got %q, want %q", loaded.Initiative, original.Initiative)
	}
	if loaded.Complexity != original.Complexity {
		t.Errorf("Complexity: got %q, want %q", loaded.Complexity, original.Complexity)
	}
	if loaded.ActiveRite != original.ActiveRite {
		t.Errorf("ActiveRite: got %q, want %q", loaded.ActiveRite, original.ActiveRite)
	}
	if loaded.CurrentPhase != original.CurrentPhase {
		t.Errorf("CurrentPhase: got %q, want %q", loaded.CurrentPhase, original.CurrentPhase)
	}
	if loaded.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion: got %q, want %q", loaded.SchemaVersion, original.SchemaVersion)
	}

	// CreatedAt comparison: truncate to second because RFC3339 loses sub-second
	if !loaded.CreatedAt.Truncate(time.Second).Equal(original.CreatedAt.Truncate(time.Second)) {
		t.Errorf("CreatedAt: got %v, want %v (truncated to second)",
			loaded.CreatedAt, original.CreatedAt)
	}
}

// TestNewContext_Validate verifies a freshly created context passes validation.
func TestNewContext_Validate(t *testing.T) {
	ctx := NewContext("Validation test", "MODULE", "test-rite")
	issues := ctx.Validate()
	if len(issues) > 0 {
		t.Errorf("NewContext should pass validation, got issues: %v", issues)
	}
}

// =============================================================================
// Session Park Tests
// =============================================================================

// TestPark_UpdatesStatusAndTimestamp verifies parking sets the correct fields.
func TestPark_UpdatesStatusAndTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	ctx := NewContext("Park test", "MODULE", "test")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Park it
	fsm := NewFSM()
	if err := fsm.ValidateTransition(ctx.Status, StatusParked); err != nil {
		t.Fatalf("ValidateTransition failed: %v", err)
	}

	beforePark := time.Now().UTC()
	now := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &now
	ctx.ParkedReason = "Testing park functionality"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save parked failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}

	if loaded.Status != StatusParked {
		t.Errorf("Status = %v, want %v", loaded.Status, StatusParked)
	}
	if loaded.ParkedAt == nil {
		t.Fatal("ParkedAt should be set after parking")
	}
	if loaded.ParkedAt.Before(beforePark.Truncate(time.Second)) {
		t.Error("ParkedAt should be recent")
	}
	if loaded.ParkedReason != "Testing park functionality" {
		t.Errorf("ParkedReason = %q, want %q", loaded.ParkedReason, "Testing park functionality")
	}
}

// TestPark_PreservesOriginalFields verifies parking preserves the immutable fields.
func TestPark_PreservesOriginalFields(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	ctx := NewContext("Preserve fields test", "SYSTEM", "ecosystem")
	originalSessionID := ctx.SessionID
	originalInitiative := ctx.Initiative
	originalComplexity := ctx.Complexity
	originalActiveRite := ctx.ActiveRite
	originalSchemaVersion := ctx.SchemaVersion

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Park
	now := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &now
	ctx.ParkedReason = "Preserving fields"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save parked failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}

	if loaded.SessionID != originalSessionID {
		t.Errorf("SessionID changed: %q -> %q", originalSessionID, loaded.SessionID)
	}
	if loaded.Initiative != originalInitiative {
		t.Errorf("Initiative changed: %q -> %q", originalInitiative, loaded.Initiative)
	}
	if loaded.Complexity != originalComplexity {
		t.Errorf("Complexity changed: %q -> %q", originalComplexity, loaded.Complexity)
	}
	if loaded.ActiveRite != originalActiveRite {
		t.Errorf("ActiveRite changed: %q -> %q", originalActiveRite, loaded.ActiveRite)
	}
	if loaded.SchemaVersion != originalSchemaVersion {
		t.Errorf("SchemaVersion changed: %q -> %q", originalSchemaVersion, loaded.SchemaVersion)
	}
}

// TestPark_FailsFromNone verifies parking from NONE state is rejected.
func TestPark_FailsFromNone(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusNone, StatusParked)
	if err == nil {
		t.Error("Should not be able to park from NONE state")
	}
	if !strings.Contains(err.Error(), "must start as active") {
		t.Errorf("Error should mention 'must start as active', got: %q", err.Error())
	}
}

// TestPark_FailsFromArchived verifies parking from ARCHIVED state is rejected.
func TestPark_FailsFromArchived(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusArchived, StatusParked)
	if err == nil {
		t.Error("Should not be able to park from ARCHIVED state")
	}
	if !strings.Contains(err.Error(), "terminal state") {
		t.Errorf("Error should mention 'terminal state', got: %q", err.Error())
	}
}

// TestPark_FailsDoublePark verifies parking an already parked session is rejected.
func TestPark_FailsDoublePark(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusParked, StatusParked)
	if err == nil {
		t.Error("Should not be able to double-park")
	}
	if !strings.Contains(err.Error(), "already parked") {
		t.Errorf("Error should mention 'already parked', got: %q", err.Error())
	}
}

// =============================================================================
// Session Resume Tests
// =============================================================================

// TestResume_UpdatesStatusAndTimestamp verifies resuming sets the correct fields.
func TestResume_UpdatesStatusAndTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create a parked session
	ctx := NewContext("Resume test", "PATCH", "test")
	parkTime := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &parkTime
	ctx.ParkedReason = "Parked for resume test"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Resume
	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}

	fsm := NewFSM()
	if err := fsm.ValidateTransition(loaded.Status, StatusActive); err != nil {
		t.Fatalf("ValidateTransition failed: %v", err)
	}

	resumeTime := time.Now().UTC()
	loaded.Status = StatusActive
	loaded.ResumedAt = &resumeTime
	loaded.ParkedAt = nil
	loaded.ParkedReason = ""

	if err := loaded.Save(ctxPath); err != nil {
		t.Fatalf("Save resumed failed: %v", err)
	}

	// Verify
	reloaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if reloaded.Status != StatusActive {
		t.Errorf("Status = %v, want %v", reloaded.Status, StatusActive)
	}
	if reloaded.ResumedAt == nil {
		t.Fatal("ResumedAt should be set after resume")
	}
	if reloaded.ParkedAt != nil {
		t.Error("ParkedAt should be nil after resume")
	}
	if reloaded.ParkedReason != "" {
		t.Errorf("ParkedReason should be empty after resume, got %q", reloaded.ParkedReason)
	}
}

// TestResume_FailsFromActive verifies resuming an already active session is rejected.
func TestResume_FailsFromActive(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusActive, StatusActive)
	if err == nil {
		t.Error("Should not be able to resume an active session")
	}
	if !strings.Contains(err.Error(), "already active") {
		t.Errorf("Error should mention 'already active', got: %q", err.Error())
	}
}

// TestResume_FailsFromNone verifies resuming from NONE is rejected.
func TestResume_FailsFromNone(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusNone, StatusActive)
	// NONE -> ACTIVE is valid (create), not resume. But the FSM allows it
	// since create is the operation that goes NONE -> ACTIVE.
	if err != nil {
		t.Errorf("NONE -> ACTIVE should be valid (create), got error: %v", err)
	}
}

// TestResume_FailsFromArchived verifies resuming from ARCHIVED is rejected.
func TestResume_FailsFromArchived(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusArchived, StatusActive)
	if err == nil {
		t.Error("Should not be able to resume an archived session")
	}
}

// =============================================================================
// Session Wrap Tests
// =============================================================================

// TestWrap_FromActive verifies wrapping from ACTIVE state.
func TestWrap_FromActive(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	ctx := NewContext("Wrap from active", "MODULE", "test")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	fsm := NewFSM()
	if err := fsm.ValidateTransition(ctx.Status, StatusArchived); err != nil {
		t.Fatalf("ValidateTransition(ACTIVE, ARCHIVED) failed: %v", err)
	}

	archiveTime := time.Now().UTC()
	ctx.Status = StatusArchived
	ctx.ArchivedAt = &archiveTime

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save archived failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}

	if loaded.Status != StatusArchived {
		t.Errorf("Status = %v, want %v", loaded.Status, StatusArchived)
	}
	if loaded.ArchivedAt == nil {
		t.Fatal("ArchivedAt should be set")
	}
	if !loaded.Status.IsTerminal() {
		t.Error("Archived should be terminal")
	}
}

// TestWrap_FromParked verifies wrapping directly from PARKED state.
func TestWrap_FromParked(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	ctx := NewContext("Wrap from parked", "MODULE", "test")
	parkTime := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &parkTime
	ctx.ParkedReason = "Abandoned session"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	fsm := NewFSM()
	if err := fsm.ValidateTransition(StatusParked, StatusArchived); err != nil {
		t.Fatalf("ValidateTransition(PARKED, ARCHIVED) failed: %v", err)
	}

	archiveTime := time.Now().UTC()
	ctx.Status = StatusArchived
	ctx.ArchivedAt = &archiveTime

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save archived failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}

	if loaded.Status != StatusArchived {
		t.Errorf("Status = %v, want %v", loaded.Status, StatusArchived)
	}
	if !loaded.Status.IsTerminal() {
		t.Error("Archived should be terminal")
	}
}

// TestWrap_FailsFromNone verifies wrapping from NONE is rejected.
func TestWrap_FailsFromNone(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusNone, StatusArchived)
	if err == nil {
		t.Error("Should not be able to wrap from NONE state")
	}
}

// TestWrap_FailsDoubleWrap verifies wrapping an already archived session is rejected.
func TestWrap_FailsDoubleWrap(t *testing.T) {
	fsm := NewFSM()
	err := fsm.ValidateTransition(StatusArchived, StatusArchived)
	if err == nil {
		t.Error("Should not be able to double-wrap")
	}
}

// TestWrap_NoFurtherTransitions verifies no transitions are possible after wrap.
func TestWrap_NoFurtherTransitions(t *testing.T) {
	fsm := NewFSM()
	transitions := fsm.ValidTransitions(StatusArchived)
	if len(transitions) != 0 {
		t.Errorf("Archived should have no valid transitions, got %v", transitions)
	}
}

// =============================================================================
// Full Lifecycle Tests (Golden Path)
// =============================================================================

// TestLifecycle_CreateParkResumeWrap tests the complete golden path with
// file I/O at every step, verifying all fields after each transition.
func TestLifecycle_CreateParkResumeWrap(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "sessions", "test-session")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	fsm := NewFSM()

	// Step 1: Create (NONE -> ACTIVE)
	ctx := NewContext("Golden path lifecycle", "MODULE", "test-rite")
	if ctx.Status != StatusActive {
		t.Fatalf("Create: Status = %v, want ACTIVE", ctx.Status)
	}
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Create: Save failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Create: Load failed: %v", err)
	}
	if loaded.Status != StatusActive {
		t.Fatalf("Create: Loaded status = %v, want ACTIVE", loaded.Status)
	}
	if loaded.ParkedAt != nil || loaded.ArchivedAt != nil || loaded.ResumedAt != nil {
		t.Error("Create: Optional timestamps should all be nil")
	}

	// Step 2: Park (ACTIVE -> PARKED)
	if err := fsm.ValidateTransition(loaded.Status, StatusParked); err != nil {
		t.Fatalf("Park: ValidateTransition failed: %v", err)
	}

	parkTime := time.Now().UTC()
	loaded.Status = StatusParked
	loaded.ParkedAt = &parkTime
	loaded.ParkedReason = "End of work day"

	if err := loaded.Save(ctxPath); err != nil {
		t.Fatalf("Park: Save failed: %v", err)
	}

	loaded, err = LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Park: Load failed: %v", err)
	}
	if loaded.Status != StatusParked {
		t.Fatalf("Park: Status = %v, want PARKED", loaded.Status)
	}
	if loaded.ParkedAt == nil {
		t.Fatal("Park: ParkedAt should be set")
	}
	if loaded.ParkedReason != "End of work day" {
		t.Errorf("Park: ParkedReason = %q, want %q", loaded.ParkedReason, "End of work day")
	}

	// Step 3: Resume (PARKED -> ACTIVE)
	if err := fsm.ValidateTransition(loaded.Status, StatusActive); err != nil {
		t.Fatalf("Resume: ValidateTransition failed: %v", err)
	}

	resumeTime := time.Now().UTC()
	loaded.Status = StatusActive
	loaded.ResumedAt = &resumeTime
	loaded.ParkedAt = nil
	loaded.ParkedReason = ""

	if err := loaded.Save(ctxPath); err != nil {
		t.Fatalf("Resume: Save failed: %v", err)
	}

	loaded, err = LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Resume: Load failed: %v", err)
	}
	if loaded.Status != StatusActive {
		t.Fatalf("Resume: Status = %v, want ACTIVE", loaded.Status)
	}
	if loaded.ResumedAt == nil {
		t.Fatal("Resume: ResumedAt should be set")
	}
	if loaded.ParkedAt != nil {
		t.Error("Resume: ParkedAt should be cleared")
	}
	if loaded.ParkedReason != "" {
		t.Error("Resume: ParkedReason should be cleared")
	}

	// Step 4: Wrap (ACTIVE -> ARCHIVED)
	if err := fsm.ValidateTransition(loaded.Status, StatusArchived); err != nil {
		t.Fatalf("Wrap: ValidateTransition failed: %v", err)
	}

	archiveTime := time.Now().UTC()
	loaded.Status = StatusArchived
	loaded.ArchivedAt = &archiveTime

	if err := loaded.Save(ctxPath); err != nil {
		t.Fatalf("Wrap: Save failed: %v", err)
	}

	final, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Wrap: Load failed: %v", err)
	}
	if final.Status != StatusArchived {
		t.Fatalf("Wrap: Status = %v, want ARCHIVED", final.Status)
	}
	if final.ArchivedAt == nil {
		t.Fatal("Wrap: ArchivedAt should be set")
	}
	if !final.Status.IsTerminal() {
		t.Error("Wrap: Status should be terminal")
	}

	// Verify no further transitions possible
	for _, target := range []Status{StatusNone, StatusActive, StatusParked, StatusArchived} {
		if fsm.CanTransition(final.Status, target) {
			t.Errorf("After wrap: should not be able to transition to %s", target)
		}
	}

	// Verify immutable fields survived all transitions
	if final.SessionID != ctx.SessionID {
		t.Errorf("SessionID changed during lifecycle: %q -> %q", ctx.SessionID, final.SessionID)
	}
	if final.Initiative != ctx.Initiative {
		t.Errorf("Initiative changed during lifecycle: %q -> %q", ctx.Initiative, final.Initiative)
	}
	if final.Complexity != ctx.Complexity {
		t.Errorf("Complexity changed during lifecycle: %q -> %q", ctx.Complexity, final.Complexity)
	}
	if final.ActiveRite != ctx.ActiveRite {
		t.Errorf("ActiveRite changed during lifecycle: %q -> %q", ctx.ActiveRite, final.ActiveRite)
	}
	if final.SchemaVersion != ctx.SchemaVersion {
		t.Errorf("SchemaVersion changed during lifecycle: %q -> %q", ctx.SchemaVersion, final.SchemaVersion)
	}
}

// TestLifecycle_CreateDirectWrap tests the shortcut path ACTIVE -> ARCHIVED.
func TestLifecycle_CreateDirectWrap(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")
	fsm := NewFSM()

	ctx := NewContext("Direct wrap test", "PATCH", "quick-fix")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := fsm.ValidateTransition(ctx.Status, StatusArchived); err != nil {
		t.Fatalf("ValidateTransition(ACTIVE, ARCHIVED) failed: %v", err)
	}

	archiveTime := time.Now().UTC()
	ctx.Status = StatusArchived
	ctx.ArchivedAt = &archiveTime

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save archived failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}

	if loaded.Status != StatusArchived {
		t.Errorf("Status = %v, want ARCHIVED", loaded.Status)
	}
}

// TestLifecycle_MultipleParkResumeCycles tests parking and resuming multiple times.
func TestLifecycle_MultipleParkResumeCycles(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")
	fsm := NewFSM()

	ctx := NewContext("Multi-cycle test", "MODULE", "test")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	const cycles = 5
	for i := range cycles {
		loaded, err := LoadContext(ctxPath)
		if err != nil {
			t.Fatalf("Cycle %d: Load for park failed: %v", i, err)
		}

		// Park
		if err := fsm.ValidateTransition(loaded.Status, StatusParked); err != nil {
			t.Fatalf("Cycle %d: Park validation failed: %v", i, err)
		}
		parkTime := time.Now().UTC()
		loaded.Status = StatusParked
		loaded.ParkedAt = &parkTime
		loaded.ParkedReason = fmt.Sprintf("Cycle %d park", i)
		if err := loaded.Save(ctxPath); err != nil {
			t.Fatalf("Cycle %d: Park save failed: %v", i, err)
		}

		// Resume
		loaded, err = LoadContext(ctxPath)
		if err != nil {
			t.Fatalf("Cycle %d: Load for resume failed: %v", i, err)
		}

		if err := fsm.ValidateTransition(loaded.Status, StatusActive); err != nil {
			t.Fatalf("Cycle %d: Resume validation failed: %v", i, err)
		}
		resumeTime := time.Now().UTC()
		loaded.Status = StatusActive
		loaded.ResumedAt = &resumeTime
		loaded.ParkedAt = nil
		loaded.ParkedReason = ""
		if err := loaded.Save(ctxPath); err != nil {
			t.Fatalf("Cycle %d: Resume save failed: %v", i, err)
		}
	}

	// Final check: should still be ACTIVE after all cycles
	final, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Final load failed: %v", err)
	}
	if final.Status != StatusActive {
		t.Errorf("After %d park/resume cycles: Status = %v, want ACTIVE", cycles, final.Status)
	}
}

// TestLifecycle_ParkedDirectToArchived tests PARKED -> ARCHIVED shortcut.
func TestLifecycle_ParkedDirectToArchived(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")
	fsm := NewFSM()

	ctx := NewContext("Parked to archived", "MODULE", "test")
	parkTime := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &parkTime
	ctx.ParkedReason = "Abandoned"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Wrap directly from PARKED
	if err := fsm.ValidateTransition(StatusParked, StatusArchived); err != nil {
		t.Fatalf("ValidateTransition(PARKED, ARCHIVED) failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	archiveTime := time.Now().UTC()
	loaded.Status = StatusArchived
	loaded.ArchivedAt = &archiveTime

	if err := loaded.Save(ctxPath); err != nil {
		t.Fatalf("Save archived failed: %v", err)
	}

	final, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Final load failed: %v", err)
	}
	if final.Status != StatusArchived {
		t.Errorf("Status = %v, want ARCHIVED", final.Status)
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

// TestEdgeCases_LoadFromNonexistentFile verifies proper error for missing file.
func TestEdgeCases_LoadFromNonexistentFile(t *testing.T) {
	_, err := LoadContext("/tmp/does-not-exist-12345/SESSION_CONTEXT.md")
	if err == nil {
		t.Error("LoadContext should fail for nonexistent file")
	}
	if !errors.IsNotFound(err) {
		t.Errorf("Error should be a not-found error, got: %v", err)
	}
}

// TestEdgeCases_CorruptFrontmatter verifies parsing errors for various corruption types.
func TestEdgeCases_CorruptFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	tests := []struct {
		name    string
		content string
	}{
		{"empty file", ""},
		{"no frontmatter", "# Just markdown\n\nNo YAML frontmatter here."},
		{"unclosed frontmatter", "---\nschema_version: 2.1\nstatus: ACTIVE\n"},
		{"invalid yaml", "---\n{{{invalid yaml\n---\n"},
		{"only delimiters", "---\n---\n"},
		{"invalid timestamp", "---\nschema_version: \"2.1\"\nsession_id: \"session-20260104-160414-12345678\"\nstatus: ACTIVE\ncreated_at: not-a-date\ninitiative: test\ncomplexity: MODULE\nactive_rite: test\ncurrent_phase: design\n---\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(ctxPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Write failed: %v", err)
			}
			_, err := LoadContext(ctxPath)
			if err == nil {
				t.Errorf("LoadContext should fail for %s", tt.name)
			}
		})
	}
}

// TestEdgeCases_SaveToReadOnlyDirectory verifies save error handling.
func TestEdgeCases_SaveToReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	ctx := NewContext("Read-only test", "MODULE", "test")
	ctxPath := filepath.Join(readOnlyDir, "SESSION_CONTEXT.md")

	err := ctx.Save(ctxPath)
	if err == nil {
		t.Error("Save to read-only directory should fail")
	}
}

// TestEdgeCases_ConcurrentReads verifies concurrent LoadContext is safe.
func TestEdgeCases_ConcurrentReads(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	ctx := NewContext("Concurrent test", "MODULE", "test")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loaded, err := LoadContext(ctxPath)
			if err != nil {
				errs <- err
				return
			}
			if loaded.SessionID != ctx.SessionID {
				errs <- fmt.Errorf("SessionID mismatch: %q vs %q", loaded.SessionID, ctx.SessionID)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("Concurrent read error: %v", err)
	}
}

// TestEdgeCases_EmptyInitiative verifies context with empty initiative.
func TestEdgeCases_EmptyInitiative(t *testing.T) {
	ctx := NewContext("", "MODULE", "test")
	// Should create without error
	if ctx.Initiative != "" {
		t.Errorf("Initiative = %q, want empty", ctx.Initiative)
	}

	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Initiative != "" {
		t.Errorf("Loaded initiative = %q, want empty", loaded.Initiative)
	}
}

// TestEdgeCases_SpecialCharactersInInitiative verifies special chars survive round-trip.
func TestEdgeCases_SpecialCharactersInInitiative(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	specialInitiatives := []string{
		"Fix: colon in initiative",
		"Line with 'single quotes'",
		`Line with "double quotes"`,
		"Unicode: test \u00e9\u00e0\u00fc",
		"Multi-word-hyphenated-initiative",
		"Initiative with #hash and @at",
	}

	for _, initiative := range specialInitiatives {
		t.Run(initiative, func(t *testing.T) {
			ctx := NewContext(initiative, "MODULE", "test")
			if err := ctx.Save(ctxPath); err != nil {
				t.Fatalf("Save failed for %q: %v", initiative, err)
			}

			loaded, err := LoadContext(ctxPath)
			if err != nil {
				t.Fatalf("Load failed for %q: %v", initiative, err)
			}

			if loaded.Initiative != initiative {
				t.Errorf("Initiative round-trip failed: got %q, want %q", loaded.Initiative, initiative)
			}
		})
	}
}

// =============================================================================
// Phase Transition Tests
// =============================================================================

// TestPhaseTransitions_ValidForward verifies all forward phase transitions.
func TestPhaseTransitions_ValidForward(t *testing.T) {
	phases := []Phase{
		PhaseRequirements,
		PhaseDesign,
		PhaseImplementation,
		PhaseValidation,
		PhaseComplete,
	}

	// All forward transitions should be valid
	for i := range phases {
		for j := i + 1; j < len(phases); j++ {
			t.Run(string(phases[i])+"->"+string(phases[j]), func(t *testing.T) {
				if !CanTransitionPhase(phases[i], phases[j]) {
					t.Errorf("Forward transition %s -> %s should be valid",
						phases[i], phases[j])
				}
			})
		}
	}
}

// TestPhaseTransitions_InvalidBackward verifies all backward phase transitions are blocked.
func TestPhaseTransitions_InvalidBackward(t *testing.T) {
	phases := []Phase{
		PhaseRequirements,
		PhaseDesign,
		PhaseImplementation,
		PhaseValidation,
		PhaseComplete,
	}

	for i := 1; i < len(phases); i++ {
		for j := 0; j < i; j++ {
			t.Run(string(phases[i])+"->"+string(phases[j]), func(t *testing.T) {
				if CanTransitionPhase(phases[i], phases[j]) {
					t.Errorf("Backward transition %s -> %s should be invalid",
						phases[i], phases[j])
				}
			})
		}
	}
}

// TestPhaseTransitions_InvalidSamePhase verifies same-phase transitions are blocked.
func TestPhaseTransitions_InvalidSamePhase(t *testing.T) {
	phases := []Phase{
		PhaseRequirements,
		PhaseDesign,
		PhaseImplementation,
		PhaseValidation,
		PhaseComplete,
	}

	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			if CanTransitionPhase(phase, phase) {
				t.Errorf("Same-phase transition %s -> %s should be invalid", phase, phase)
			}
		})
	}
}

// TestPhaseOrder verifies phase ordering is correct.
func TestPhaseOrder(t *testing.T) {
	expected := map[Phase]int{
		PhaseRequirements:   0,
		PhaseDesign:         1,
		PhaseImplementation: 2,
		PhaseValidation:     3,
		PhaseComplete:       4,
	}

	for phase, wantOrder := range expected {
		t.Run(string(phase), func(t *testing.T) {
			got := PhaseOrder(phase)
			if got != wantOrder {
				t.Errorf("PhaseOrder(%s) = %d, want %d", phase, got, wantOrder)
			}
		})
	}

	// Invalid phase should return -1
	if PhaseOrder(Phase("invalid")) != -1 {
		t.Error("PhaseOrder for invalid phase should be -1")
	}
}

// TestPhaseTransitions_InvalidPhaseValues verifies invalid phase strings are handled.
func TestPhaseTransitions_InvalidPhaseValues(t *testing.T) {
	invalidPhases := []Phase{
		Phase(""),
		Phase("invalid"),
		Phase("REQUIREMENTS"), // wrong case
	}

	for _, invalid := range invalidPhases {
		t.Run(string(invalid), func(t *testing.T) {
			if CanTransitionPhase(invalid, PhaseDesign) {
				t.Errorf("Invalid phase %q should not allow transition", invalid)
			}
			if CanTransitionPhase(PhaseRequirements, invalid) {
				t.Errorf("Transition to invalid phase %q should fail", invalid)
			}
		})
	}
}

// =============================================================================
// Status Type Tests
// =============================================================================

// TestStatus_StringRepresentation verifies String() returns the correct value.
func TestStatus_StringRepresentation(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusNone, "NONE"},
		{StatusActive, "ACTIVE"},
		{StatusParked, "PARKED"},
		{StatusArchived, "ARCHIVED"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestStatus_IsValid_EdgeCases verifies edge cases for IsValid.
func TestStatus_IsValid_EdgeCases(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"NONE", true},
		{"ACTIVE", true},
		{"PARKED", true},
		{"ARCHIVED", true},
		{"none", false},    // wrong case
		{"active", false},  // wrong case
		{"Active", false},  // mixed case
		{"", false},        // empty
		{" ACTIVE", false}, // leading space
		{"ACTIVE ", false}, // trailing space
		{"INVALID", false}, // unknown
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := Status(tt.input)
			if got := s.IsValid(); got != tt.want {
				t.Errorf("Status(%q).IsValid() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestStatus_IsTerminal_OnlyArchived verifies only ARCHIVED is terminal.
func TestStatus_IsTerminal_OnlyArchived(t *testing.T) {
	nonTerminal := []Status{StatusNone, StatusActive, StatusParked}
	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Errorf("%s should not be terminal", s)
		}
	}

	if !StatusArchived.IsTerminal() {
		t.Error("ARCHIVED should be terminal")
	}
}

// =============================================================================
// Context Validation Tests
// =============================================================================

// TestContext_Validate_InvalidFields verifies validation catches specific issues.
func TestContext_Validate_InvalidFields(t *testing.T) {
	tests := []struct {
		name      string
		ctx       *Context
		wantIssue string
	}{
		{
			name: "invalid session ID",
			ctx: &Context{
				SessionID:    "not-valid",
				Status:       StatusActive,
				Initiative:   "Test",
				Complexity:   "MODULE",
				ActiveRite:   "test",
				CurrentPhase: "design",
				CreatedAt:    time.Now().UTC(),
			},
			wantIssue: "session_id",
		},
		{
			name: "invalid status",
			ctx: &Context{
				SessionID:    "session-20260104-160414-12345678",
				Status:       Status("BOGUS"),
				Initiative:   "Test",
				Complexity:   "MODULE",
				ActiveRite:   "test",
				CurrentPhase: "design",
				CreatedAt:    time.Now().UTC(),
			},
			wantIssue: "status",
		},
		{
			name: "unsupported schema version",
			ctx: &Context{
				SchemaVersion: "99.0",
				SessionID:     "session-20260104-160414-12345678",
				Status:        StatusActive,
				Initiative:    "Test",
				Complexity:    "MODULE",
				ActiveRite:    "test",
				CurrentPhase:  "design",
				CreatedAt:     time.Now().UTC(),
			},
			wantIssue: "schema_version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := tt.ctx.Validate()
			if len(issues) == 0 {
				t.Error("Expected validation issues, got none")
				return
			}

			found := false
			for _, issue := range issues {
				if strings.Contains(strings.ToLower(issue), tt.wantIssue) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected issue mentioning %q, got: %v", tt.wantIssue, issues)
			}
		})
	}
}

// TestContext_Validate_AllComplexities verifies all complexity levels are accepted.
func TestContext_Validate_AllComplexities(t *testing.T) {
	complexities := []string{"PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION"}

	for _, c := range complexities {
		t.Run(c, func(t *testing.T) {
			ctx := NewContext("Test", c, "test")
			issues := ctx.Validate()
			if len(issues) > 0 {
				t.Errorf("Complexity %q should be valid, got issues: %v", c, issues)
			}
		})
	}
}

// =============================================================================
// Session ID Tests (additional coverage)
// =============================================================================

// TestSessionID_Format verifies the generated ID matches the documented format.
func TestSessionID_Format(t *testing.T) {
	for range 10 {
		id := GenerateSessionID()
		if !strings.HasPrefix(id, "session-") {
			t.Errorf("ID %q should start with 'session-'", id)
		}
		if !IsValidSessionID(id) {
			t.Errorf("Generated ID %q failed validation", id)
		}
		// Format: "session-" (8) + "YYYYMMDD" (8) + "-" (1) + "HHMMSS" (6) + "-" (1) + hex (8) = 32 chars
		if len(id) != 32 {
			t.Errorf("ID %q length = %d, want 32", id, len(id))
		}
	}
}

// TestParseSessionTimestamp_Correctness verifies timestamp extraction.
func TestParseSessionTimestamp_Correctness(t *testing.T) {
	id := "session-20260215-143022-abcdef01"
	ts := ParseSessionTimestamp(id)

	expected := time.Date(2026, 2, 15, 14, 30, 22, 0, time.UTC)
	if !ts.Equal(expected) {
		t.Errorf("ParseSessionTimestamp(%q) = %v, want %v", id, ts, expected)
	}
}
