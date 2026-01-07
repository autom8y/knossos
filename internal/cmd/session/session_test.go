package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/session"
)

// =============================================================================
// Create Command Tests (non-seed mode)
// =============================================================================

// TestCreate_BasicCreation verifies basic session creation without --seed.
func TestCreate_BasicCreation(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude directory and ACTIVE_RITE
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "ACTIVE_RITE"), []byte("10x-dev"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	// Create sessions directory
	sessionsDir := filepath.Join(claudeDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	// Create locks directory
	locksDir := filepath.Join(sessionsDir, ".locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	// Run create command
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := createOptions{
		complexity: "MODULE",
		seed:       false,
	}

	err := runCreate(ctx, "Test Initiative", opts)
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	// Find the created session
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		t.Fatalf("Failed to read sessions dir: %v", err)
	}

	var sessionID string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "session-") {
			sessionID = entry.Name()
			break
		}
	}

	if sessionID == "" {
		t.Fatal("No session directory created")
	}

	// Load and verify session context
	ctxPath := filepath.Join(sessionsDir, sessionID, "SESSION_CONTEXT.md")
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}

	// Verify session is ACTIVE (not PARKED like seed mode)
	if sessCtx.Status != session.StatusActive {
		t.Errorf("Session status = %v, want ACTIVE", sessCtx.Status)
	}

	// Verify initiative
	if sessCtx.Initiative != "Test Initiative" {
		t.Errorf("Initiative = %q, want %q", sessCtx.Initiative, "Test Initiative")
	}

	// Verify complexity
	if sessCtx.Complexity != "MODULE" {
		t.Errorf("Complexity = %q, want %q", sessCtx.Complexity, "MODULE")
	}

	// Verify team (Team is a pointer, may be nil for cross-cutting sessions)
	if sessCtx.Rite != nil && *sessCtx.Rite != "10x-dev" {
		t.Errorf("Team = %q, want %q", *sessCtx.Rite, "10x-dev")
	}

	// Verify current-session file was created
	currentSessionFile := filepath.Join(sessionsDir, ".current-session")
	currentID, err := os.ReadFile(currentSessionFile)
	if err != nil {
		t.Fatalf("Failed to read .current-session: %v", err)
	}
	if string(currentID) != sessionID {
		t.Errorf("Current session = %q, want %q", string(currentID), sessionID)
	}
}

// TestCreate_InvalidComplexity verifies that invalid complexity is rejected.
func TestCreate_InvalidComplexity(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create minimal structure
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "ACTIVE_RITE"), []byte("10x-dev"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := createOptions{
		complexity: "INVALID",
		seed:       false,
	}

	err := runCreate(ctx, "Test Initiative", opts)
	if err == nil {
		t.Error("Expected error for invalid complexity, got nil")
	}
}

// TestCreate_BlocksSecondActiveSession verifies that creating a session fails
// when an active session already exists.
func TestCreate_BlocksSecondActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude directory and ACTIVE_RITE
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "ACTIVE_RITE"), []byte("10x-dev"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	// Create sessions directory
	sessionsDir := filepath.Join(claudeDir, "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := createOptions{
		complexity: "MODULE",
		seed:       false,
	}

	// Create first session
	err := runCreate(ctx, "First Session", opts)
	if err != nil {
		t.Fatalf("First runCreate failed: %v", err)
	}

	// Try to create second session - should fail
	err = runCreate(ctx, "Second Session", opts)
	if err == nil {
		t.Error("Expected error when creating second session, got nil")
	}
}

// =============================================================================
// List Command Tests
// =============================================================================

// TestList_NoSessions verifies list behavior with no sessions.
func TestList_NoSessions(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory (empty)
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := listOptions{
		all:   false,
		limit: 20,
	}

	err := runList(ctx, opts)
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}

// TestList_WithStatusFilter verifies list filtering by status.
func TestList_WithStatusFilter(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	// Create an ACTIVE session
	activeSessionID := "session-20260105-100000-active123"
	activeSessionDir := filepath.Join(sessionsDir, activeSessionID)
	if err := os.MkdirAll(activeSessionDir, 0755); err != nil {
		t.Fatalf("Failed to create active session dir: %v", err)
	}

	activeContext := `---
schema_version: "2.1"
session_id: ` + activeSessionID + `
status: ACTIVE
initiative: Active Session
complexity: MODULE
created_at: 2026-01-05T10:00:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(activeSessionDir, "SESSION_CONTEXT.md"), []byte(activeContext), 0644); err != nil {
		t.Fatalf("Failed to write active session context: %v", err)
	}

	// Create a PARKED session
	parkedSessionID := "session-20260105-110000-parked456"
	parkedSessionDir := filepath.Join(sessionsDir, parkedSessionID)
	if err := os.MkdirAll(parkedSessionDir, 0755); err != nil {
		t.Fatalf("Failed to create parked session dir: %v", err)
	}

	parkedContext := `---
schema_version: "2.1"
session_id: ` + parkedSessionID + `
status: PARKED
initiative: Parked Session
complexity: MODULE
created_at: 2026-01-05T11:00:00Z
parked_at: 2026-01-05T11:30:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(parkedSessionDir, "SESSION_CONTEXT.md"), []byte(parkedContext), 0644); err != nil {
		t.Fatalf("Failed to write parked session context: %v", err)
	}

	// Set current session
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(activeSessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Test list with ACTIVE filter
	opts := listOptions{
		all:    false,
		status: "ACTIVE",
		limit:  20,
	}

	err := runList(ctx, opts)
	if err != nil {
		t.Fatalf("runList with ACTIVE filter failed: %v", err)
	}

	// Test list with PARKED filter
	opts.status = "PARKED"
	err = runList(ctx, opts)
	if err != nil {
		t.Fatalf("runList with PARKED filter failed: %v", err)
	}

	// Test list without filter (should return both non-archived)
	opts.status = ""
	err = runList(ctx, opts)
	if err != nil {
		t.Fatalf("runList without filter failed: %v", err)
	}
}

// TestList_ExcludesArchivedByDefault verifies archived sessions are excluded without --all.
func TestList_ExcludesArchivedByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	// Create an ACTIVE session
	activeSessionID := "session-20260105-100000-active789"
	activeSessionDir := filepath.Join(sessionsDir, activeSessionID)
	if err := os.MkdirAll(activeSessionDir, 0755); err != nil {
		t.Fatalf("Failed to create active session dir: %v", err)
	}

	activeContext := `---
schema_version: "2.1"
session_id: ` + activeSessionID + `
status: ACTIVE
initiative: Active Session
complexity: MODULE
created_at: 2026-01-05T10:00:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(activeSessionDir, "SESSION_CONTEXT.md"), []byte(activeContext), 0644); err != nil {
		t.Fatalf("Failed to write session context: %v", err)
	}

	// Create an ARCHIVED session
	archivedSessionID := "session-20260105-090000-archived000"
	archivedSessionDir := filepath.Join(sessionsDir, archivedSessionID)
	if err := os.MkdirAll(archivedSessionDir, 0755); err != nil {
		t.Fatalf("Failed to create archived session dir: %v", err)
	}

	archivedContext := `---
schema_version: "2.1"
session_id: ` + archivedSessionID + `
status: ARCHIVED
initiative: Archived Session
complexity: MODULE
created_at: 2026-01-05T09:00:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(archivedSessionDir, "SESSION_CONTEXT.md"), []byte(archivedContext), 0644); err != nil {
		t.Fatalf("Failed to write archived session context: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Without --all, should only return active
	opts := listOptions{
		all:   false,
		limit: 20,
	}

	err := runList(ctx, opts)
	if err != nil {
		t.Fatalf("runList without --all failed: %v", err)
	}

	// With --all, should include archived
	opts.all = true
	err = runList(ctx, opts)
	if err != nil {
		t.Fatalf("runList with --all failed: %v", err)
	}
}

// =============================================================================
// Audit Command Tests
// =============================================================================

// TestAudit_WithEvents verifies audit command reads events correctly.
func TestAudit_WithEvents(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create session directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-120000-audit123"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create SESSION_CONTEXT.md
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Audit Test
complexity: MODULE
created_at: 2026-01-05T12:00:00Z
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create events.jsonl with some events
	events := `{"timestamp":"2026-01-05T12:00:00Z","event":"SESSION_CREATED","metadata":{"initiative":"Audit Test"}}
{"timestamp":"2026-01-05T12:30:00Z","event":"PHASE_TRANSITION","from_phase":"requirements","to_phase":"design"}
{"timestamp":"2026-01-05T13:00:00Z","event":"PHASE_TRANSITION","from_phase":"design","to_phase":"implementation"}
`
	if err := os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatalf("Failed to write events.jsonl: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := auditOptions{
		limit: 50,
	}

	err := runAudit(ctx, opts)
	if err != nil {
		t.Fatalf("runAudit failed: %v", err)
	}
}

// TestAudit_WithEventTypeFilter verifies audit filtering by event type.
func TestAudit_WithEventTypeFilter(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-130000-filter456"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Filter Test
complexity: MODULE
created_at: 2026-01-05T13:00:00Z
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	events := `{"timestamp":"2026-01-05T13:00:00Z","event":"SESSION_CREATED","metadata":{}}
{"timestamp":"2026-01-05T13:30:00Z","event":"PHASE_TRANSITION","from_phase":"requirements","to_phase":"design"}
{"timestamp":"2026-01-05T14:00:00Z","event":"SESSION_PARKED","metadata":{"reason":"break"}}
`
	if err := os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatalf("Failed to write events.jsonl: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Filter by PHASE_TRANSITION
	opts := auditOptions{
		limit:     50,
		eventType: "PHASE_TRANSITION",
	}

	err := runAudit(ctx, opts)
	if err != nil {
		t.Fatalf("runAudit with event type filter failed: %v", err)
	}
}

// TestAudit_NoSession verifies audit fails gracefully without active session.
func TestAudit_NoSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")

	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := auditOptions{
		limit: 50,
	}

	err := runAudit(ctx, opts)
	if err == nil {
		t.Error("Expected error when no session exists, got nil")
	}
}

// =============================================================================
// FSM State Transition Tests (Full Lifecycle)
// =============================================================================

// TestFSM_FullLifecycle tests NONE -> ACTIVE -> PARKED -> ACTIVE -> ARCHIVED.
func TestFSM_FullLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude directory and ACTIVE_RITE
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "ACTIVE_RITE"), []byte("10x-dev"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	sessionsDir := filepath.Join(claudeDir, "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Step 1: NONE -> ACTIVE (create)
	createOpts := createOptions{
		complexity: "MODULE",
		seed:       false,
	}
	err := runCreate(ctx, "Full Lifecycle Test", createOpts)
	if err != nil {
		t.Fatalf("Step 1 (create) failed: %v", err)
	}

	// Get the session ID
	currentID, err := ctx.getCurrentSessionID()
	if err != nil {
		t.Fatalf("Failed to get current session ID: %v", err)
	}

	// Verify ACTIVE
	ctxPath := filepath.Join(sessionsDir, currentID, "SESSION_CONTEXT.md")
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}
	if sessCtx.Status != session.StatusActive {
		t.Errorf("Step 1: Status = %v, want ACTIVE", sessCtx.Status)
	}

	// Step 2: ACTIVE -> PARKED (park)
	parkOpts := parkOptions{
		reason: "Testing lifecycle",
	}
	err = runPark(ctx, parkOpts)
	if err != nil {
		t.Fatalf("Step 2 (park) failed: %v", err)
	}

	sessCtx, err = session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload session context: %v", err)
	}
	if sessCtx.Status != session.StatusParked {
		t.Errorf("Step 2: Status = %v, want PARKED", sessCtx.Status)
	}

	// Step 3: PARKED -> ACTIVE (resume)
	err = runResume(ctx)
	if err != nil {
		t.Fatalf("Step 3 (resume) failed: %v", err)
	}

	sessCtx, err = session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload session context: %v", err)
	}
	if sessCtx.Status != session.StatusActive {
		t.Errorf("Step 3: Status = %v, want ACTIVE", sessCtx.Status)
	}

	// Step 4: ACTIVE -> ARCHIVED (wrap)
	wrapOpts := wrapOptions{
		noArchive: true, // Don't move to archive for easier testing
	}
	err = runWrap(ctx, wrapOpts)
	if err != nil {
		t.Fatalf("Step 4 (wrap) failed: %v", err)
	}

	sessCtx, err = session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload session context: %v", err)
	}
	if sessCtx.Status != session.StatusArchived {
		t.Errorf("Step 4: Status = %v, want ARCHIVED", sessCtx.Status)
	}
}

// TestFSM_CannotResumeArchived verifies ARCHIVED -> ACTIVE is blocked.
func TestFSM_CannotResumeArchived(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140000-archived789"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create an ARCHIVED session
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ARCHIVED
initiative: Archived Test
complexity: MODULE
created_at: 2026-01-05T14:00:00Z
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Try to resume - should fail
	err := runResume(ctx)
	if err == nil {
		t.Error("Expected error when resuming archived session, got nil")
	}
}

// TestFSM_CannotParkParked verifies PARKED -> PARKED is blocked.
func TestFSM_CannotParkParked(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-150000-parkedparkd"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create a PARKED session
	now := time.Now().UTC()
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: PARKED
initiative: Parked Test
complexity: MODULE
created_at: 2026-01-05T15:00:00Z
parked_at: ` + now.Format(time.RFC3339) + `
parked_reason: Already parked
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Try to park again - should fail
	parkOpts := parkOptions{
		reason: "Double park attempt",
	}
	err := runPark(ctx, parkOpts)
	if err == nil {
		t.Error("Expected error when parking already parked session, got nil")
	}
}

// TestFSM_WrapFromParked verifies PARKED -> ARCHIVED is allowed.
func TestFSM_WrapFromParked(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-160000-parkedwrap"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create a PARKED session
	now := time.Now().UTC()
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: PARKED
initiative: Wrap From Parked Test
complexity: MODULE
created_at: 2026-01-05T16:00:00Z
parked_at: ` + now.Format(time.RFC3339) + `
parked_reason: Testing wrap from parked
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Wrap from PARKED should succeed
	wrapOpts := wrapOptions{
		noArchive: true,
	}
	err := runWrap(ctx, wrapOpts)
	if err != nil {
		t.Fatalf("Wrap from PARKED failed: %v", err)
	}

	// Verify ARCHIVED
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload session context: %v", err)
	}
	if sessCtx.Status != session.StatusArchived {
		t.Errorf("Status = %v, want ARCHIVED", sessCtx.Status)
	}
}

// =============================================================================
// Transition Command Tests (Phase Transitions)
// =============================================================================

// TestTransition_RequirementsToDesign verifies phase transition.
func TestTransition_RequirementsToDesign(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-170000-phase123"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create required artifact (PRD)
	prdDir := filepath.Join(projectDir, "docs", "requirements")
	if err := os.MkdirAll(prdDir, 0755); err != nil {
		t.Fatalf("Failed to create PRD dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(prdDir, "PRD-test.md"), []byte("# Test PRD\n"), 0644); err != nil {
		t.Fatalf("Failed to write PRD: %v", err)
	}

	// Create ACTIVE session in requirements phase
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Phase Transition Test
complexity: MODULE
created_at: 2026-01-05T17:00:00Z
current_phase: requirements
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := transitionOptions{
		force: false,
	}

	// Transition to design
	err := runTransition(ctx, "design", opts)
	if err != nil {
		t.Fatalf("Transition to design failed: %v", err)
	}

	// Verify phase changed
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload session context: %v", err)
	}
	if sessCtx.CurrentPhase != "design" {
		t.Errorf("CurrentPhase = %q, want design", sessCtx.CurrentPhase)
	}
}

// TestTransition_InvalidPhase verifies invalid phase is rejected.
func TestTransition_InvalidPhase(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-180000-invalid456"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Invalid Phase Test
complexity: MODULE
created_at: 2026-01-05T18:00:00Z
current_phase: requirements
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := transitionOptions{
		force: false,
	}

	// Try invalid phase
	err := runTransition(ctx, "invalid_phase", opts)
	if err == nil {
		t.Error("Expected error for invalid phase, got nil")
	}
}

// TestTransition_CannotTransitionParkedSession verifies parked sessions cannot transition phases.
func TestTransition_CannotTransitionParkedSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-190000-parkedtrans"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create a PARKED session
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: PARKED
initiative: Parked Transition Test
complexity: MODULE
created_at: 2026-01-05T19:00:00Z
current_phase: requirements
parked_at: 2026-01-05T19:30:00Z
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := transitionOptions{
		force: true, // Even with force, should fail
	}

	// Try to transition - should fail because session is not ACTIVE
	err := runTransition(ctx, "design", opts)
	if err == nil {
		t.Error("Expected error when transitioning parked session, got nil")
	}
}

// TestTransition_ForceSkipsArtifactValidation verifies --force bypasses artifact checks.
func TestTransition_ForceSkipsArtifactValidation(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-200000-force789"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	locksDir := filepath.Join(sessionsDir, ".locks")
	auditDir := filepath.Join(sessionsDir, ".audit")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("Failed to create locks dir: %v", err)
	}
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	// Create ACTIVE session - no PRD artifact exists
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Force Transition Test
complexity: MODULE
created_at: 2026-01-05T20:00:00Z
current_phase: requirements
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	// Without force - should fail (no PRD)
	opts := transitionOptions{
		force: false,
	}
	err := runTransition(ctx, "design", opts)
	if err == nil {
		t.Error("Expected error without force (missing PRD), got nil")
	}

	// With force - should succeed
	opts.force = true
	err = runTransition(ctx, "design", opts)
	if err != nil {
		t.Fatalf("Transition with --force failed: %v", err)
	}

	// Verify phase changed
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to reload session context: %v", err)
	}
	if sessCtx.CurrentPhase != "design" {
		t.Errorf("CurrentPhase = %q, want design", sessCtx.CurrentPhase)
	}
}
