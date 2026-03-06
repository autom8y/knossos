package session

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// TestSessionLifecycle_FullCycle tests the complete golden path:
// create -> park -> resume -> archive
func TestSessionLifecycle_FullCycle(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260104-160414-12345678")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// 1. Create session (NONE -> ACTIVE)
	ctx := NewContext("Full lifecycle test", "MODULE", "test-rite")
	if ctx.Status != StatusActive {
		t.Errorf("NewContext status = %v, want %v", ctx.Status, StatusActive)
	}
	if ctx.SessionID == "" {
		t.Error("NewContext should set SessionID")
	}

	// Save initial context
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save context: %v", err)
	}

	// 2. Park session (ACTIVE -> PARKED)
	fsm := NewFSM()
	if err := fsm.ValidateTransition(ctx.Status, StatusParked); err != nil {
		t.Fatalf("ValidateTransition(ACTIVE, PARKED) failed: %v", err)
	}

	now := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &now
	ctx.ParkedReason = "Testing park"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save parked context: %v", err)
	}

	// Reload and verify
	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload context: %v", err)
	}
	if loaded.Status != StatusParked {
		t.Errorf("Status = %v, want %v", loaded.Status, StatusParked)
	}
	if loaded.ParkedReason != "Testing park" {
		t.Errorf("ParkedReason = %q, want %q", loaded.ParkedReason, "Testing park")
	}

	// 3. Resume session (PARKED -> ACTIVE)
	if err := fsm.ValidateTransition(loaded.Status, StatusActive); err != nil {
		t.Fatalf("ValidateTransition(PARKED, ACTIVE) failed: %v", err)
	}

	resumedAt := time.Now().UTC()
	loaded.Status = StatusActive
	loaded.ResumedAt = &resumedAt
	loaded.ParkedAt = nil
	loaded.ParkedReason = ""

	if err := loaded.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save resumed context: %v", err)
	}

	// 4. Wrap session (ACTIVE -> ARCHIVED)
	reloaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload after resume: %v", err)
	}

	if err := fsm.ValidateTransition(reloaded.Status, StatusArchived); err != nil {
		t.Fatalf("ValidateTransition(ACTIVE, ARCHIVED) failed: %v", err)
	}

	archivedAt := time.Now().UTC()
	reloaded.Status = StatusArchived
	reloaded.ArchivedAt = &archivedAt

	if err := reloaded.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save archived context: %v", err)
	}

	// Final verification
	final, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload final context: %v", err)
	}
	if final.Status != StatusArchived {
		t.Errorf("Final status = %v, want %v", final.Status, StatusArchived)
	}
	if !final.Status.IsTerminal() {
		t.Error("Archived status should be terminal")
	}
}

// TestSessionLifecycle_AlternatePath tests PARKED -> ARCHIVED transition
func TestSessionLifecycle_AlternatePath(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, "session-20260104-160415-87654321")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Create and park
	ctx := NewContext("Alternate path test", "PATCH", "test")
	now := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &now
	ctx.ParkedReason = "Immediate park"

	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save context: %v", err)
	}

	// Archive directly from PARKED
	fsm := NewFSM()
	if err := fsm.ValidateTransition(ctx.Status, StatusArchived); err != nil {
		t.Fatalf("ValidateTransition(PARKED, ARCHIVED) failed: %v", err)
	}

	archivedAt := time.Now().UTC()
	ctx.Status = StatusArchived
	ctx.ArchivedAt = &archivedAt

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save archived context: %v", err)
	}

	// Verify
	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}
	if loaded.Status != StatusArchived {
		t.Errorf("Status = %v, want %v", loaded.Status, StatusArchived)
	}
}

// TestFSM_ValidTransitions tests all valid FSM transitions
func TestFSM_ValidTransitions(t *testing.T) {
	fsm := NewFSM()

	tests := []struct {
		name     string
		from     Status
		to       Status
		wantOK   bool
		wantErr  bool
		errMatch string
	}{
		// Valid transitions
		{"create", StatusNone, StatusActive, true, false, ""},
		{"park", StatusActive, StatusParked, true, false, ""},
		{"wrap_active", StatusActive, StatusArchived, true, false, ""},
		{"resume", StatusParked, StatusActive, true, false, ""},
		{"wrap_parked", StatusParked, StatusArchived, true, false, ""},

		// Invalid transitions - from archived (terminal)
		{"archived_to_active", StatusArchived, StatusActive, false, true, "terminal state"},
		{"archived_to_parked", StatusArchived, StatusParked, false, true, "terminal state"},
		{"archived_to_archived", StatusArchived, StatusArchived, false, true, "terminal state"},

		// Invalid transitions - double park/active
		{"double_park", StatusParked, StatusParked, false, true, "already parked"},
		{"double_active", StatusActive, StatusActive, false, true, "already active"},

		// Invalid transitions - backwards to NONE
		{"active_to_none", StatusActive, StatusNone, false, true, "invalid"},
		{"parked_to_none", StatusParked, StatusNone, false, true, "invalid"},

		// Invalid transitions - from NONE
		{"none_to_parked", StatusNone, StatusParked, false, true, "must start as active"},
		{"none_to_archived", StatusNone, StatusArchived, false, true, "must start as active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test CanTransition
			got := fsm.CanTransition(tt.from, tt.to)
			if got != tt.wantOK {
				t.Errorf("CanTransition(%s, %s) = %v, want %v",
					tt.from, tt.to, got, tt.wantOK)
			}

			// Test ValidateTransition
			err := fsm.ValidateTransition(tt.from, tt.to)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateTransition(%s, %s) should return error", tt.from, tt.to)
				} else if tt.errMatch != "" {
					// Check error message contains expected text
					if !containsString(err.Error(), tt.errMatch) {
						t.Errorf("ValidateTransition error = %q, want to contain %q",
							err.Error(), tt.errMatch)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTransition(%s, %s) unexpected error: %v",
						tt.from, tt.to, err)
				}
			}
		})
	}
}

// TestSessionCreate tests session creation
func TestSessionCreate(t *testing.T) {
	tests := []struct {
		name       string
		initiative string
		complexity string
		rite       string
	}{
		{"basic", "Test initiative", "MODULE", "test-rite"},
		{"cross_cutting", "Cross-cutting work", "PATCH", ""},
		{"none_rite", "No rite", "SYSTEM", "none"},
		{"complex", "Migration project", "MIGRATION", "migration-rite"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(tt.initiative, tt.complexity, tt.rite)

			// Verify defaults
			if ctx.Status != StatusActive {
				t.Errorf("Status = %v, want %v", ctx.Status, StatusActive)
			}
			if ctx.SchemaVersion != "2.3" {
				t.Errorf("SchemaVersion = %q, want %q", ctx.SchemaVersion, "2.3")
			}
			if ctx.CurrentPhase != "requirements" {
				t.Errorf("CurrentPhase = %q, want %q", ctx.CurrentPhase, "requirements")
			}
			if ctx.Initiative != tt.initiative {
				t.Errorf("Initiative = %q, want %q", ctx.Initiative, tt.initiative)
			}
			if ctx.Complexity != tt.complexity {
				t.Errorf("Complexity = %q, want %q", ctx.Complexity, tt.complexity)
			}

			// Verify session ID is valid
			if !IsValidSessionID(ctx.SessionID) {
				t.Errorf("Invalid SessionID: %q", ctx.SessionID)
			}

			// Verify timestamps
			if ctx.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}
			if ctx.ParkedAt != nil {
				t.Error("ParkedAt should be nil for new session")
			}
			if ctx.ArchivedAt != nil {
				t.Error("ArchivedAt should be nil for new session")
			}
		})
	}
}

// TestSessionPark tests parking a session
func TestSessionPark(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create active session
	ctx := NewContext("Park test", "MODULE", "test")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Park it
	fsm := NewFSM()
	if err := fsm.ValidateTransition(ctx.Status, StatusParked); err != nil {
		t.Fatalf("ValidateTransition failed: %v", err)
	}

	now := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &now
	ctx.ParkedReason = "End of day"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save parked: %v", err)
	}

	// Reload and verify
	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}

	if loaded.Status != StatusParked {
		t.Errorf("Status = %v, want %v", loaded.Status, StatusParked)
	}
	if loaded.ParkedAt == nil {
		t.Fatal("ParkedAt should be set")
	}
	if loaded.ParkedReason != "End of day" {
		t.Errorf("ParkedReason = %q, want %q", loaded.ParkedReason, "End of day")
	}

	// Verify original context preserved
	if loaded.Initiative != ctx.Initiative {
		t.Error("Initiative should be preserved")
	}
	// Note: CreatedAt comparison uses UTC truncation due to RFC3339 serialization
	if !loaded.CreatedAt.Truncate(time.Second).Equal(ctx.CreatedAt.Truncate(time.Second)) {
		t.Errorf("CreatedAt should be preserved: got %v, want %v", loaded.CreatedAt, ctx.CreatedAt)
	}
}

// TestSessionResume tests resuming a parked session
func TestSessionResume(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	// Create parked session
	ctx := NewContext("Resume test", "PATCH", "test")
	now := time.Now().UTC()
	ctx.Status = StatusParked
	ctx.ParkedAt = &now
	ctx.ParkedReason = "Parked for testing"

	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Resume it
	fsm := NewFSM()
	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if err := fsm.ValidateTransition(loaded.Status, StatusActive); err != nil {
		t.Fatalf("ValidateTransition failed: %v", err)
	}

	resumedAt := time.Now().UTC()
	loaded.Status = StatusActive
	loaded.ResumedAt = &resumedAt
	loaded.ParkedAt = nil
	loaded.ParkedReason = ""

	if err := loaded.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save resumed: %v", err)
	}

	// Verify
	reloaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}

	if reloaded.Status != StatusActive {
		t.Errorf("Status = %v, want %v", reloaded.Status, StatusActive)
	}
	if reloaded.ResumedAt == nil {
		t.Fatal("ResumedAt should be set")
	}
	if reloaded.ParkedAt != nil {
		t.Error("ParkedAt should be cleared")
	}
	if reloaded.ParkedReason != "" {
		t.Error("ParkedReason should be cleared")
	}
}

// TestSessionWrap tests wrapping/archiving a session
func TestSessionWrap(t *testing.T) {
	tests := []struct {
		name         string
		initialState Status
	}{
		{"from_active", StatusActive},
		{"from_parked", StatusParked},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

			ctx := NewContext("Wrap test", "MODULE", "test")
			ctx.Status = tt.initialState
			if tt.initialState == StatusParked {
				now := time.Now().UTC()
				ctx.ParkedAt = &now
				ctx.ParkedReason = "Parked"
			}

			if err := ctx.Save(ctxPath); err != nil {
				t.Fatalf("Failed to save: %v", err)
			}

			// Wrap it
			fsm := NewFSM()
			loaded, err := LoadContext(ctxPath)
			if err != nil {
				t.Fatalf("Failed to load: %v", err)
			}

			if err := fsm.ValidateTransition(loaded.Status, StatusArchived); err != nil {
				t.Fatalf("ValidateTransition failed: %v", err)
			}

			archivedAt := time.Now().UTC()
			loaded.Status = StatusArchived
			loaded.ArchivedAt = &archivedAt

			if err := loaded.Save(ctxPath); err != nil {
				t.Fatalf("Failed to save archived: %v", err)
			}

			// Verify
			final, err := LoadContext(ctxPath)
			if err != nil {
				t.Fatalf("Failed to reload: %v", err)
			}

			if final.Status != StatusArchived {
				t.Errorf("Status = %v, want %v", final.Status, StatusArchived)
			}
			if final.ArchivedAt == nil {
				t.Fatal("ArchivedAt should be set")
			}
			if !final.Status.IsTerminal() {
				t.Error("Archived status should be terminal")
			}

			// Verify no further transitions possible
			if fsm.CanTransition(final.Status, StatusActive) {
				t.Error("Should not be able to transition from ARCHIVED to ACTIVE")
			}
			if fsm.CanTransition(final.Status, StatusParked) {
				t.Error("Should not be able to transition from ARCHIVED to PARKED")
			}
		})
	}
}

// TestEdgeCases_MissingSessionDirectory tests error handling for missing directory
func TestEdgeCases_MissingSessionDirectory(t *testing.T) {
	nonExistentPath := "/tmp/nonexistent-session-dir-12345/SESSION_CONTEXT.md"

	_, err := LoadContext(nonExistentPath)
	if err == nil {
		t.Error("LoadContext should fail for missing file")
	}
}

// TestEdgeCases_CorruptContext tests handling of corrupt SESSION_CONTEXT.md
func TestEdgeCases_CorruptContext(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	tests := []struct {
		name    string
		content string
	}{
		{"no_frontmatter", "Just markdown content"},
		{"unclosed_frontmatter", "---\nschema_version: 2.1\n# never closed"},
		{"invalid_yaml", "---\ninvalid::yaml::here\n---\n"},
		{"empty_file", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(ctxPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			_, err := LoadContext(ctxPath)
			if err == nil {
				t.Errorf("LoadContext should fail for %s", tt.name)
			}
		})
	}
}

// TestEdgeCases_ConcurrentAccess tests behavior with concurrent operations
// This is a basic test - real concurrency control is tested via lock package
func TestEdgeCases_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	ctx := NewContext("Concurrent test", "MODULE", "test")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Multiple goroutines reading is safe
	done := make(chan bool, 5)
	for range 5 {
		go func() {
			if _, err := LoadContext(ctxPath); err != nil {
				t.Errorf("Concurrent read failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all reads
	for range 5 {
		<-done
	}
}

// TestErrorMessages_HumanReadable tests that error messages are clear
func TestErrorMessages_HumanReadable(t *testing.T) {
	fsm := NewFSM()

	tests := []struct {
		from     Status
		to       Status
		contains string
	}{
		{StatusArchived, StatusActive, "terminal state"},
		{StatusParked, StatusParked, "already parked"},
		{StatusActive, StatusActive, "already active"},
		{StatusNone, StatusParked, "must start as active"},
		{StatusActive, StatusNone, "invalid"},
	}

	for _, tt := range tests {
		err := fsm.ValidateTransition(tt.from, tt.to)
		if err == nil {
			t.Errorf("Expected error for %s -> %s", tt.from, tt.to)
			continue
		}

		if !containsString(err.Error(), tt.contains) {
			t.Errorf("Error %q should contain %q", err.Error(), tt.contains)
		}

		// Verify error is human-readable (not just a code)
		if len(err.Error()) < 10 {
			t.Errorf("Error message too short to be helpful: %q", err.Error())
		}
	}
}

// TestValidTransitions tests FSM.ValidTransitions
func TestValidTransitions(t *testing.T) {
	fsm := NewFSM()

	tests := []struct {
		status Status
		want   []Status
	}{
		{StatusNone, []Status{StatusActive}},
		{StatusActive, []Status{StatusParked, StatusArchived}},
		{StatusParked, []Status{StatusActive, StatusArchived}},
		{StatusArchived, nil}, // Terminal - no valid transitions
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := fsm.ValidTransitions(tt.status)

			if len(got) != len(tt.want) {
				t.Errorf("ValidTransitions(%s) = %v, want %v", tt.status, got, tt.want)
				return
			}

			// Check each expected transition is present
			for _, wantStatus := range tt.want {
				found := slices.Contains(got, wantStatus)
				if !found {
					t.Errorf("ValidTransitions(%s) missing %s", tt.status, wantStatus)
				}
			}
		})
	}
}

// TestRoundTrip_PreservesAllFields tests that save/load preserves all fields
func TestRoundTrip_PreservesAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	ctxPath := filepath.Join(tmpDir, "SESSION_CONTEXT.md")

	now := time.Now().UTC().Truncate(time.Second)
	parkedAt := now.Add(-1 * time.Hour)
	resumedAt := now.Add(-30 * time.Minute)

	riteName := "test-rite"
	original := &Context{
		SchemaVersion: "2.1",
		SessionID:     "session-20260104-160414-12345678",
		Status:        StatusActive,
		CreatedAt:     now.Add(-2 * time.Hour),
		Initiative:    "Test with all fields",
		Complexity:    "MODULE",
		ActiveRite:    "test-rite",
		Rite:          &riteName,
		CurrentPhase:  "implementation",
		ParkedAt:      &parkedAt,
		ParkedReason:  "Test parking",
		ResumedAt:     &resumedAt,
		Body:          "\n# Custom body\n\nWith content\n",
	}

	// Save
	if err := original.Save(ctxPath); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load
	loaded, err := LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Compare all fields
	if loaded.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion: got %q, want %q", loaded.SchemaVersion, original.SchemaVersion)
	}
	if loaded.SessionID != original.SessionID {
		t.Errorf("SessionID: got %q, want %q", loaded.SessionID, original.SessionID)
	}
	if loaded.Status != original.Status {
		t.Errorf("Status: got %v, want %v", loaded.Status, original.Status)
	}
	if !loaded.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", loaded.CreatedAt, original.CreatedAt)
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
	if loaded.ParkedReason != original.ParkedReason {
		t.Errorf("ParkedReason: got %q, want %q", loaded.ParkedReason, original.ParkedReason)
	}

	// Check optional time fields
	if (loaded.ParkedAt == nil) != (original.ParkedAt == nil) {
		t.Error("ParkedAt nullability mismatch")
	} else if loaded.ParkedAt != nil && !loaded.ParkedAt.Equal(*original.ParkedAt) {
		t.Errorf("ParkedAt: got %v, want %v", *loaded.ParkedAt, *original.ParkedAt)
	}

	if (loaded.ResumedAt == nil) != (original.ResumedAt == nil) {
		t.Error("ResumedAt nullability mismatch")
	} else if loaded.ResumedAt != nil && !loaded.ResumedAt.Equal(*original.ResumedAt) {
		t.Errorf("ResumedAt: got %v, want %v", *loaded.ResumedAt, *original.ResumedAt)
	}

	// Check body - note: serialization adds extra newline after ---
	// This is a known behavior where body parsing includes the newline after closing ---
	expectedBody := "\n" + original.Body
	if loaded.Body != expectedBody {
		t.Errorf("Body: got %q, want %q", loaded.Body, expectedBody)
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
