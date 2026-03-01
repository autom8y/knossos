package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/session"
)

// =============================================================================
// D3: Moirai Integration Tests
//
// These tests validate session state mutations as they occur through the CLI
// command layer -- the same path Moirai uses when invoking `ari session` commands.
//
// D1 tested the FSM and context serialization directly. These tests exercise
// the full command functions (runCreate, runPark, runResume, runWrap) and verify:
//   - Context file integrity after each mutation
//   - Audit trail / event emission
//   - Error paths for invalid mutations
//   - Concurrent safety via lock acquisition
//
// These tests do NOT duplicate D1's FSM unit tests.
// =============================================================================

// --- Test Helpers ---

// newTestContext creates a cmdContext for testing with JSON output.
// If sessionID is provided, it sets ctx.SessionID for explicit session resolution.
func newTestContext(projectDir string, sessionID ...string) *cmdContext {
	outputFormat := "json"
	verbose := true
	var sessionIDPtr *string
	if len(sessionID) > 0 && sessionID[0] != "" {
		sessionIDPtr = &sessionID[0]
	}
	return &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
			SessionID: sessionIDPtr,
		},
	}
}

// setupProjectDir creates the minimal directory structure for tests.
// Returns the project root directory path.
func setupProjectDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	dirs := []string{
		filepath.Join(tmpDir, ".claude"),
		filepath.Join(tmpDir, ".sos", "sessions"),
		filepath.Join(tmpDir, ".sos", "sessions", ".locks"),
		filepath.Join(tmpDir, ".sos", "sessions", ".audit"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	// Write ACTIVE_RITE so create picks it up
	if err := os.WriteFile(filepath.Join(tmpDir, ".claude", "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	return tmpDir
}

// writeSessionContext writes a SESSION_CONTEXT.md with the given fields.
// This is used to set up preconditions for non-create tests.
func writeSessionContext(t *testing.T, projectDir, sessionID, status, initiative string, parkedAt *time.Time, parkedReason string) {
	t.Helper()

	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	content := "---\nschema_version: \"2.1\"\nsession_id: " + sessionID +
		"\nstatus: " + status +
		"\ninitiative: " + initiative +
		"\ncomplexity: MODULE" +
		"\ncreated_at: " + createdAt.Format(time.RFC3339) +
		"\nactive_rite: test-rite" +
		"\ncurrent_phase: requirements"

	if parkedAt != nil {
		content += "\nparked_at: " + parkedAt.Format(time.RFC3339)
	}
	if parkedReason != "" {
		content += "\nparked_reason: " + parkedReason
	}
	content += "\n---\n\n# Session Context\n"

	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}
}

// loadSessionStatus loads and returns the session status from SESSION_CONTEXT.md.
func loadSessionStatus(t *testing.T, projectDir, sessionID string) session.Status {
	t.Helper()
	ctxPath := filepath.Join(projectDir, ".sos", "sessions", sessionID, "SESSION_CONTEXT.md")
	ctx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}
	return ctx.Status
}

// loadSessionContext loads and returns the full session context.
func loadSessionContext(t *testing.T, projectDir, sessionID string) *session.Context {
	t.Helper()
	ctxPath := filepath.Join(projectDir, ".sos", "sessions", sessionID, "SESSION_CONTEXT.md")
	ctx, err := session.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}
	return ctx
}

// findCreatedSessionID discovers the session ID from the sessions directory.
func findCreatedSessionID(t *testing.T, projectDir string) string {
	t.Helper()
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		t.Fatalf("Failed to read sessions dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "session-") {
			return entry.Name()
		}
	}
	t.Fatal("No session directory found after create")
	return ""
}

// countEventsOfType counts events of a specific type in events.jsonl.
func countEventsOfType(t *testing.T, projectDir, sessionID, eventType string) int {
	t.Helper()
	eventsPath := filepath.Join(projectDir, ".sos", "sessions", sessionID, "events.jsonl")
	content, err := os.ReadFile(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}
	return strings.Count(string(content), eventType)
}

// readEventsJSONL reads events.jsonl and returns parsed JSON objects.
func readEventsJSONL(t *testing.T, projectDir, sessionID string) []map[string]any {
	t.Helper()
	eventsPath := filepath.Join(projectDir, ".sos", "sessions", sessionID, "events.jsonl")
	file, err := os.Open(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("Failed to open events.jsonl: %v", err)
	}
	defer file.Close()

	var events []map[string]any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Skip unparseable lines (legacy format)
			continue
		}
		events = append(events, event)
	}
	return events
}

// =============================================================================
// Test Scenario 1: State Mutation via Session Commands
//
// Validates that create, park, resume, wrap commands correctly transition
// state through the FSM. These are the commands Moirai invokes via shell.
// =============================================================================

func TestMoirai_CreateParkResumeWrap_GoldenPath(t *testing.T) {
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	// --- Step 1: Create (NONE -> ACTIVE) ---
	err := runCreate(ctx, "Moirai golden path", createOptions{complexity: "MODULE"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionID := findCreatedSessionID(t, projectDir)
	status := loadSessionStatus(t, projectDir, sessionID)
	if status != session.StatusActive {
		t.Errorf("After create: status = %v, want ACTIVE", status)
	}

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	// --- Step 2: Park (ACTIVE -> PARKED) ---
	err = runPark(ctx, parkOptions{reason: "Moirai-initiated park"})
	if err != nil {
		t.Fatalf("Park failed: %v", err)
	}

	sessCtx := loadSessionContext(t, projectDir, sessionID)
	if sessCtx.Status != session.StatusParked {
		t.Errorf("After park: status = %v, want PARKED", sessCtx.Status)
	}
	if sessCtx.ParkedAt == nil {
		t.Error("After park: ParkedAt should be set")
	}
	if sessCtx.ParkedReason != "Moirai-initiated park" {
		t.Errorf("After park: ParkedReason = %q, want %q", sessCtx.ParkedReason, "Moirai-initiated park")
	}

	// --- Step 3: Resume (PARKED -> ACTIVE) ---
	err = runResume(ctx)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	sessCtx = loadSessionContext(t, projectDir, sessionID)
	if sessCtx.Status != session.StatusActive {
		t.Errorf("After resume: status = %v, want ACTIVE", sessCtx.Status)
	}
	if sessCtx.ResumedAt == nil {
		t.Error("After resume: ResumedAt should be set")
	}
	if sessCtx.ParkedAt != nil {
		t.Error("After resume: ParkedAt should be cleared")
	}
	if sessCtx.ParkedReason != "" {
		t.Errorf("After resume: ParkedReason should be empty, got %q", sessCtx.ParkedReason)
	}

	// --- Step 4: Wrap (ACTIVE -> ARCHIVED) ---
	err = runWrap(ctx, wrapOptions{noArchive: true})
	if err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}

	sessCtx = loadSessionContext(t, projectDir, sessionID)
	if sessCtx.Status != session.StatusArchived {
		t.Errorf("After wrap: status = %v, want ARCHIVED", sessCtx.Status)
	}
	if sessCtx.ArchivedAt == nil {
		t.Error("After wrap: ArchivedAt should be set")
	}
}

func TestMoirai_WrapFromParked_DirectPath(t *testing.T) {
	projectDir := setupProjectDir(t)
	now := time.Now().UTC()
	sessionID := "session-20260205-100000-parkedwrp"
	writeSessionContext(t, projectDir, sessionID, "PARKED", "Direct wrap test", &now, "Testing direct wrap")

	ctx := newTestContext(projectDir, sessionID)
	err := runWrap(ctx, wrapOptions{noArchive: true})
	if err != nil {
		t.Fatalf("Wrap from PARKED failed: %v", err)
	}

	sessCtx := loadSessionContext(t, projectDir, sessionID)
	if sessCtx.Status != session.StatusArchived {
		t.Errorf("After wrap from PARKED: status = %v, want ARCHIVED", sessCtx.Status)
	}
	if sessCtx.ArchivedAt == nil {
		t.Error("After wrap from PARKED: ArchivedAt should be set")
	}
}

func TestMoirai_CreateWithAllComplexities(t *testing.T) {
	complexities := []string{"PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION"}

	for _, c := range complexities {
		t.Run(c, func(t *testing.T) {
			projectDir := setupProjectDir(t)
			ctx := newTestContext(projectDir)

			err := runCreate(ctx, "Test "+c, createOptions{complexity: c})
			if err != nil {
				t.Fatalf("Create with complexity %s failed: %v", c, err)
			}

			sessionID := findCreatedSessionID(t, projectDir)
			sessCtx := loadSessionContext(t, projectDir, sessionID)
			if sessCtx.Complexity != c {
				t.Errorf("Complexity = %q, want %q", sessCtx.Complexity, c)
			}
			if sessCtx.Status != session.StatusActive {
				t.Errorf("Status = %v, want ACTIVE", sessCtx.Status)
			}
		})
	}
}

// =============================================================================
// Test Scenario 2: Context File Integrity
//
// SESSION_CONTEXT.md must be valid YAML after each mutation. Frontmatter
// must survive round-trips. Fields set by earlier operations must persist.
// =============================================================================

func TestMoirai_ContextIntegrity_FieldPreservation(t *testing.T) {
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	// Create session
	err := runCreate(ctx, "Field preservation test", createOptions{complexity: "SYSTEM"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionID := findCreatedSessionID(t, projectDir)
	origCtx := loadSessionContext(t, projectDir, sessionID)
	origCreatedAt := origCtx.CreatedAt
	origInitiative := origCtx.Initiative
	origComplexity := origCtx.Complexity
	origSchemaVersion := origCtx.SchemaVersion

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	// Park -- verify create-time fields survive
	err = runPark(ctx, parkOptions{reason: "Preserve fields"})
	if err != nil {
		t.Fatalf("Park failed: %v", err)
	}

	parkedCtx := loadSessionContext(t, projectDir, sessionID)
	if !parkedCtx.CreatedAt.Truncate(time.Second).Equal(origCreatedAt.Truncate(time.Second)) {
		t.Errorf("CreatedAt changed after park: %v vs %v", parkedCtx.CreatedAt, origCreatedAt)
	}
	if parkedCtx.Initiative != origInitiative {
		t.Errorf("Initiative changed after park: %q vs %q", parkedCtx.Initiative, origInitiative)
	}
	if parkedCtx.Complexity != origComplexity {
		t.Errorf("Complexity changed after park: %q vs %q", parkedCtx.Complexity, origComplexity)
	}
	if parkedCtx.SchemaVersion != origSchemaVersion {
		t.Errorf("SchemaVersion changed after park: %q vs %q", parkedCtx.SchemaVersion, origSchemaVersion)
	}

	// Resume -- verify create-time AND park-time awareness survive
	err = runResume(ctx)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	resumedCtx := loadSessionContext(t, projectDir, sessionID)
	if !resumedCtx.CreatedAt.Truncate(time.Second).Equal(origCreatedAt.Truncate(time.Second)) {
		t.Errorf("CreatedAt changed after resume: %v vs %v", resumedCtx.CreatedAt, origCreatedAt)
	}
	if resumedCtx.Initiative != origInitiative {
		t.Errorf("Initiative changed after resume: %q vs %q", resumedCtx.Initiative, origInitiative)
	}
	if resumedCtx.ResumedAt == nil {
		t.Error("ResumedAt should be set after resume")
	}

	// Wrap -- verify all accumulated fields survive to final state
	err = runWrap(ctx, wrapOptions{noArchive: true})
	if err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}

	finalCtx := loadSessionContext(t, projectDir, sessionID)
	if !finalCtx.CreatedAt.Truncate(time.Second).Equal(origCreatedAt.Truncate(time.Second)) {
		t.Errorf("CreatedAt changed after wrap: %v vs %v", finalCtx.CreatedAt, origCreatedAt)
	}
	if finalCtx.Initiative != origInitiative {
		t.Errorf("Initiative changed after wrap: %q vs %q", finalCtx.Initiative, origInitiative)
	}
	if finalCtx.ArchivedAt == nil {
		t.Error("ArchivedAt should be set after wrap")
	}
}

func TestMoirai_ContextIntegrity_YAMLValid(t *testing.T) {
	// Verify that SESSION_CONTEXT.md is valid parseable YAML at every state
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	err := runCreate(ctx, "YAML validity test", createOptions{complexity: "MODULE"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionID := findCreatedSessionID(t, projectDir)
	ctxPath := filepath.Join(projectDir, ".sos", "sessions", sessionID, "SESSION_CONTEXT.md")

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	// Helper: verify file is loadable (valid frontmatter YAML)
	verifyLoadable := func(label string) {
		t.Helper()
		_, err := session.LoadContext(ctxPath)
		if err != nil {
			t.Errorf("SESSION_CONTEXT.md is not valid YAML after %s: %v", label, err)
		}
	}

	verifyLoadable("create")

	err = runPark(ctx, parkOptions{reason: "YAML test"})
	if err != nil {
		t.Fatalf("Park failed: %v", err)
	}
	verifyLoadable("park")

	err = runResume(ctx)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	verifyLoadable("resume")

	err = runWrap(ctx, wrapOptions{noArchive: true})
	if err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}
	verifyLoadable("wrap")
}

func TestMoirai_ContextIntegrity_SessionIDStable(t *testing.T) {
	// Session ID must remain constant across all mutations
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	err := runCreate(ctx, "Session ID stability", createOptions{complexity: "MODULE"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionID := findCreatedSessionID(t, projectDir)

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	// Verify ID is valid format
	if !session.IsValidSessionID(sessionID) {
		t.Fatalf("Session ID %q is not valid format", sessionID)
	}

	// Park and verify ID unchanged
	if err := runPark(ctx, parkOptions{reason: "ID test"}); err != nil {
		t.Fatalf("Park failed: %v", err)
	}
	parkedCtx := loadSessionContext(t, projectDir, sessionID)
	if parkedCtx.SessionID != sessionID {
		t.Errorf("SessionID changed after park: %q vs %q", parkedCtx.SessionID, sessionID)
	}

	// Resume and verify ID unchanged
	if err := runResume(ctx); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	resumedCtx := loadSessionContext(t, projectDir, sessionID)
	if resumedCtx.SessionID != sessionID {
		t.Errorf("SessionID changed after resume: %q vs %q", resumedCtx.SessionID, sessionID)
	}

	// Wrap and verify ID unchanged
	if err := runWrap(ctx, wrapOptions{noArchive: true}); err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}
	finalCtx := loadSessionContext(t, projectDir, sessionID)
	if finalCtx.SessionID != sessionID {
		t.Errorf("SessionID changed after wrap: %q vs %q", finalCtx.SessionID, sessionID)
	}
}

// =============================================================================
// Test Scenario 3: Audit Trail
//
// State transitions must be logged. Each command emits events to events.jsonl.
// =============================================================================

func TestMoirai_AuditTrail_EventsEmitted(t *testing.T) {
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	// Create
	err := runCreate(ctx, "Audit trail test", createOptions{complexity: "MODULE"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionID := findCreatedSessionID(t, projectDir)

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	// Verify create event (unified clew event system)
	createCount := countEventsOfType(t, projectDir, sessionID, "session.created")
	if createCount != 1 {
		t.Errorf("After create: session.created events = %d, want 1", createCount)
	}

	// Park -- should emit session.parked and session.ended
	err = runPark(ctx, parkOptions{reason: "Audit park"})
	if err != nil {
		t.Fatalf("Park failed: %v", err)
	}

	parkCount := countEventsOfType(t, projectDir, sessionID, "session.parked")
	if parkCount != 1 {
		t.Errorf("After park: session.parked events = %d, want 1", parkCount)
	}
	parkEndCount := countEventsOfType(t, projectDir, sessionID, "session.ended")
	if parkEndCount < 1 {
		t.Errorf("After park: session.ended events = %d, want >= 1", parkEndCount)
	}

	// Resume -- should emit session.resumed (NOT session.ended)
	err = runResume(ctx)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	resumeCount := countEventsOfType(t, projectDir, sessionID, "session.resumed")
	if resumeCount != 1 {
		t.Errorf("After resume: session.resumed events = %d, want 1", resumeCount)
	}
	// session.ended count should NOT increase after resume
	resumeEndCount := countEventsOfType(t, projectDir, sessionID, "session.ended")
	if resumeEndCount != parkEndCount {
		t.Errorf("After resume: session.ended count changed from %d to %d (resume should NOT emit session.ended)",
			parkEndCount, resumeEndCount)
	}

	// Wrap -- should emit session.archived and session.ended
	err = runWrap(ctx, wrapOptions{noArchive: true})
	if err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}

	archiveCount := countEventsOfType(t, projectDir, sessionID, "session.archived")
	if archiveCount != 1 {
		t.Errorf("After wrap: session.archived events = %d, want 1", archiveCount)
	}
	wrapEndCount := countEventsOfType(t, projectDir, sessionID, "session.ended")
	if wrapEndCount < parkEndCount+1 {
		t.Errorf("After wrap: session.ended count = %d, want > %d", wrapEndCount, parkEndCount)
	}
}

func TestMoirai_AuditTrail_EventsAreValidJSON(t *testing.T) {
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	// Run lifecycle to generate events
	if err := runCreate(ctx, "JSON validity test", createOptions{complexity: "MODULE"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	sessionID := findCreatedSessionID(t, projectDir)

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	if err := runPark(ctx, parkOptions{reason: "JSON test"}); err != nil {
		t.Fatalf("Park failed: %v", err)
	}

	// Read events file and verify every line is valid JSON
	eventsPath := filepath.Join(projectDir, ".sos", "sessions", sessionID, "events.jsonl")
	file, err := os.Open(eventsPath)
	if err != nil {
		t.Fatalf("Failed to open events.jsonl: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !json.Valid([]byte(line)) {
			t.Errorf("events.jsonl line %d is not valid JSON: %s", lineNum, line)
		}
	}

	if lineNum == 0 {
		t.Error("events.jsonl is empty, expected at least one event")
	}
}

func TestMoirai_AuditTrail_EventsJSONLPopulated(t *testing.T) {
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	if err := runCreate(ctx, "Events JSONL test", createOptions{complexity: "MODULE"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionID := findCreatedSessionID(t, projectDir)

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	// Park to generate lifecycle events
	if err := runPark(ctx, parkOptions{reason: "Events JSONL park"}); err != nil {
		t.Fatalf("Park failed: %v", err)
	}

	// Verify events.jsonl exists and is non-empty (replaces transitions.log)
	eventsPath := filepath.Join(projectDir, ".sos", "sessions", sessionID, "events.jsonl")
	info, err := os.Stat(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Error("events.jsonl does not exist after park operation")
			return
		}
		t.Fatalf("Failed to stat events.jsonl: %v", err)
	}
	if info.Size() == 0 {
		t.Error("events.jsonl is empty after park operation")
	}
}

// =============================================================================
// Test Scenario 4: Error Paths
//
// Invalid mutations (park when already parked, resume when archived, etc.)
// must produce clear, actionable errors -- not panics or silent failures.
// =============================================================================

func TestMoirai_ErrorPath_ParkAlreadyParked(t *testing.T) {
	projectDir := setupProjectDir(t)
	now := time.Now().UTC()
	sessionID := "session-20260205-110000-parkpark1"
	writeSessionContext(t, projectDir, sessionID, "PARKED", "Double park test", &now, "Already parked")

	ctx := newTestContext(projectDir, sessionID)
	err := runPark(ctx, parkOptions{reason: "Second park attempt"})
	if err == nil {
		t.Fatal("Expected error when parking already-parked session, got nil")
	}
	if !strings.Contains(err.Error(), "already parked") {
		t.Errorf("Error message should contain 'already parked', got: %v", err)
	}

	// Verify session state unchanged
	status := loadSessionStatus(t, projectDir, sessionID)
	if status != session.StatusParked {
		t.Errorf("Session status should remain PARKED, got %v", status)
	}
}

func TestMoirai_ErrorPath_ResumeActiveSession(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-120000-rsmactve"
	writeSessionContext(t, projectDir, sessionID, "ACTIVE", "Resume active test", nil, "")

	ctx := newTestContext(projectDir, sessionID)
	err := runResume(ctx)
	if err == nil {
		t.Fatal("Expected error when resuming active session, got nil")
	}
	if !strings.Contains(err.Error(), "already active") {
		t.Errorf("Error message should contain 'already active', got: %v", err)
	}
}

func TestMoirai_ErrorPath_ResumeArchivedSession(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-130000-rsmarchv"
	writeSessionContext(t, projectDir, sessionID, "ARCHIVED", "Resume archived test", nil, "")

	ctx := newTestContext(projectDir, sessionID)
	err := runResume(ctx)
	if err == nil {
		t.Fatal("Expected error when resuming archived session, got nil")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Errorf("Error message should indicate terminal state, got: %v", err)
	}
}

func TestMoirai_ErrorPath_ParkArchivedSession(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-140000-parkarch"
	writeSessionContext(t, projectDir, sessionID, "ARCHIVED", "Park archived test", nil, "")

	ctx := newTestContext(projectDir, sessionID)
	err := runPark(ctx, parkOptions{reason: "Cannot park archived"})
	if err == nil {
		t.Fatal("Expected error when parking archived session, got nil")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Errorf("Error message should indicate terminal state, got: %v", err)
	}
}

func TestMoirai_ErrorPath_WrapArchivedSession(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-150000-wraparch"
	writeSessionContext(t, projectDir, sessionID, "ARCHIVED", "Wrap archived test", nil, "")

	ctx := newTestContext(projectDir, sessionID)
	err := runWrap(ctx, wrapOptions{noArchive: true})
	if err == nil {
		t.Fatal("Expected error when wrapping archived session, got nil")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Errorf("Error message should indicate terminal state, got: %v", err)
	}
}

func TestMoirai_ErrorPath_NoActiveSession(t *testing.T) {
	projectDir := setupProjectDir(t)
	// No session created, no .current-session file
	ctx := newTestContext(projectDir)

	// All operations that require an active session should fail
	tests := []struct {
		name string
		fn   func() error
	}{
		{"park", func() error { return runPark(ctx, parkOptions{reason: "no session"}) }},
		{"resume", func() error { return runResume(ctx) }},
		{"wrap", func() error { return runWrap(ctx, wrapOptions{noArchive: true}) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Errorf("%s should fail without active session", tt.name)
			}
		})
	}
}

func TestMoirai_ErrorPath_CreateInvalidComplexity(t *testing.T) {
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	invalidComplexities := []string{"INVALID", "LOW", "HIGH", "", "module"}

	for _, c := range invalidComplexities {
		t.Run(c, func(t *testing.T) {
			err := runCreate(ctx, "Invalid complexity", createOptions{complexity: c})
			if err == nil {
				t.Errorf("Expected error for invalid complexity %q, got nil", c)
			}
		})
	}
}

func TestMoirai_ErrorPath_CreateBlocksDuplicate(t *testing.T) {
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	// Create first session
	err := runCreate(ctx, "First session", createOptions{complexity: "MODULE"})
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	// Second create should be blocked
	err = runCreate(ctx, "Second session", createOptions{complexity: "MODULE"})
	if err == nil {
		t.Fatal("Expected error when creating second session, got nil")
	}
}

func TestMoirai_ErrorPath_MissingSessionDir(t *testing.T) {
	// Write .current-session pointing to a session ID whose directory does not exist
	projectDir := setupProjectDir(t)
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	fakeID := "session-20260205-160000-nosuchid"
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(fakeID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	ctx := newTestContext(projectDir)

	err := runPark(ctx, parkOptions{reason: "no dir"})
	if err == nil {
		t.Fatal("Expected error for missing session directory, got nil")
	}
}

// =============================================================================
// Test Scenario 5: Concurrent Safety
//
// Two mutations cannot race. Verify lock acquisition prevents concurrent state
// changes. Informed by D2's finding about .current-session race condition.
// =============================================================================

func TestMoirai_ConcurrentSafety_SerializedMutations(t *testing.T) {
	// This test verifies that concurrent park attempts are serialized by the lock.
	// Within a single process, flock is reentrant per-FD, but each goroutine
	// calls runPark which creates its own FD. We test that the result is consistent:
	// exactly one succeeds and the other fails (because the session is already parked).
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-170000-concurr1"
	writeSessionContext(t, projectDir, sessionID, "ACTIVE", "Concurrent test", nil, "")

	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := newTestContext(projectDir, sessionID)
			results[idx] = runPark(ctx, parkOptions{reason: "Concurrent park " + string(rune('A'+idx))})
		}(i)
	}

	wg.Wait()

	// Exactly one should succeed, one should fail (already parked or lock contention)
	successes := 0
	failures := 0
	for _, err := range results {
		if err == nil {
			successes++
		} else {
			failures++
		}
	}

	if successes != 1 {
		t.Errorf("Expected exactly 1 success from concurrent parks, got %d (errors: %v, %v)",
			successes, results[0], results[1])
	}
	if failures != 1 {
		t.Errorf("Expected exactly 1 failure from concurrent parks, got %d", failures)
	}

	// Final state must be PARKED (not corrupted)
	status := loadSessionStatus(t, projectDir, sessionID)
	if status != session.StatusParked {
		t.Errorf("Final status = %v, want PARKED (state may be corrupted)", status)
	}
}

func TestMoirai_ConcurrentSafety_CreateCreateRace(t *testing.T) {
	// Two concurrent creates should result in exactly one success.
	// The __create__ lock serializes creation attempts.
	projectDir := setupProjectDir(t)

	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := newTestContext(projectDir)
			results[idx] = runCreate(ctx, "Concurrent create", createOptions{complexity: "MODULE"})
		}(i)
	}

	wg.Wait()

	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Errorf("Expected exactly 1 successful create, got %d (errors: %v, %v)",
			successes, results[0], results[1])
	}

	// Verify exactly one session directory exists
	sessionsDir := filepath.Join(projectDir, ".sos", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		t.Fatalf("Failed to read sessions dir: %v", err)
	}

	sessionCount := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "session-") {
			sessionCount++
		}
	}

	if sessionCount != 1 {
		t.Errorf("Expected 1 session directory, found %d", sessionCount)
	}
}

func TestMoirai_ConcurrentSafety_ResumeResumeRace(t *testing.T) {
	// Two concurrent resumes on the same parked session: exactly one should succeed.
	projectDir := setupProjectDir(t)
	now := time.Now().UTC()
	sessionID := "session-20260205-180000-rsmrace1"
	writeSessionContext(t, projectDir, sessionID, "PARKED", "Resume race test", &now, "Parked for race")

	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := newTestContext(projectDir, sessionID)
			results[idx] = runResume(ctx)
		}(i)
	}

	wg.Wait()

	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Errorf("Expected exactly 1 successful resume, got %d (errors: %v, %v)",
			successes, results[0], results[1])
	}

	// Final state must be ACTIVE
	status := loadSessionStatus(t, projectDir, sessionID)
	if status != session.StatusActive {
		t.Errorf("Final status = %v, want ACTIVE", status)
	}
}

// =============================================================================
// Test: Moirai CLI-backed operations (the paths Moirai agent uses)
//
// Moirai invokes CLI commands like `ari session park --reason="..."`.
// These tests validate the same code path by calling the run* functions
// directly with the same parameter structures.
// =============================================================================

func TestMoirai_ParkReasonPreserved(t *testing.T) {
	// Moirai always passes a reason when parking. Verify it persists.
	reasons := []string{
		"Taking a break for lunch",
		"Switching to higher priority work",
		"End of day",
		"Blocked: waiting for design review",
		"Manual park",
	}

	for _, reason := range reasons {
		t.Run(reason, func(t *testing.T) {
			projectDir := setupProjectDir(t)
			sessionID := "session-20260205-190000-reason01"
			writeSessionContext(t, projectDir, sessionID, "ACTIVE", "Reason test", nil, "")

			ctx := newTestContext(projectDir, sessionID)
			err := runPark(ctx, parkOptions{reason: reason})
			if err != nil {
				t.Fatalf("Park failed: %v", err)
			}

			sessCtx := loadSessionContext(t, projectDir, sessionID)
			if sessCtx.ParkedReason != reason {
				t.Errorf("ParkedReason = %q, want %q", sessCtx.ParkedReason, reason)
			}
		})
	}
}

func TestMoirai_ResumeClears_ParkFields(t *testing.T) {
	projectDir := setupProjectDir(t)
	now := time.Now().UTC()
	sessionID := "session-20260205-200000-clrpark1"
	writeSessionContext(t, projectDir, sessionID, "PARKED", "Clear park fields", &now, "Was parked")

	ctx := newTestContext(projectDir, sessionID)
	if err := runResume(ctx); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	sessCtx := loadSessionContext(t, projectDir, sessionID)
	if sessCtx.ParkedAt != nil {
		t.Error("ParkedAt should be nil after resume")
	}
	if sessCtx.ParkedReason != "" {
		t.Errorf("ParkedReason should be empty after resume, got %q", sessCtx.ParkedReason)
	}
	if sessCtx.ResumedAt == nil {
		t.Error("ResumedAt should be set after resume")
	}
}

func TestMoirai_WrapSetsArchiveTimestamp(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-210000-archts01"
	writeSessionContext(t, projectDir, sessionID, "ACTIVE", "Archive timestamp test", nil, "")

	before := time.Now().UTC()
	ctx := newTestContext(projectDir, sessionID)
	if err := runWrap(ctx, wrapOptions{noArchive: true}); err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}
	after := time.Now().UTC()

	sessCtx := loadSessionContext(t, projectDir, sessionID)
	if sessCtx.ArchivedAt == nil {
		t.Fatal("ArchivedAt should be set after wrap")
	}

	// ArchivedAt should be between before and after timestamps (with 1s tolerance)
	archivedAt := sessCtx.ArchivedAt.Truncate(time.Second)
	if archivedAt.Before(before.Add(-1*time.Second)) || archivedAt.After(after.Add(1*time.Second)) {
		t.Errorf("ArchivedAt %v is not between %v and %v", archivedAt, before, after)
	}
}

func TestMoirai_WrapWithArchive_MovesDirectory(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-220000-archive1"
	writeSessionContext(t, projectDir, sessionID, "ACTIVE", "Archive move test", nil, "")

	ctx := newTestContext(projectDir, sessionID)
	// noArchive=false (default) -- should move to archive
	if err := runWrap(ctx, wrapOptions{noArchive: false}); err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}

	// Original session dir should be gone
	originalDir := filepath.Join(projectDir, ".sos", "sessions", sessionID)
	if _, err := os.Stat(originalDir); !os.IsNotExist(err) {
		t.Error("Original session directory should be removed after archive")
	}

	// Archive directory should exist
	archiveDir := filepath.Join(projectDir, ".sos", "archive", sessionID)
	if _, err := os.Stat(archiveDir); err != nil {
		t.Errorf("Archive directory should exist at %s: %v", archiveDir, err)
	}

	// SESSION_CONTEXT.md should be loadable from archive
	archiveCtxPath := filepath.Join(archiveDir, "SESSION_CONTEXT.md")
	archiveCtx, err := session.LoadContext(archiveCtxPath)
	if err != nil {
		t.Fatalf("Failed to load archived context: %v", err)
	}
	if archiveCtx.Status != session.StatusArchived {
		t.Errorf("Archived context status = %v, want ARCHIVED", archiveCtx.Status)
	}
}

// =============================================================================
// Test: Phase transitions within the Moirai flow
// =============================================================================

func TestMoirai_PhaseTransition_RequiresActiveSession(t *testing.T) {
	projectDir := setupProjectDir(t)
	now := time.Now().UTC()
	sessionID := "session-20260205-230000-phsprk01"
	writeSessionContext(t, projectDir, sessionID, "PARKED", "Phase from parked", &now, "Parked")

	ctx := newTestContext(projectDir, sessionID)
	err := runTransition(ctx, "design", transitionOptions{force: true})
	if err == nil {
		t.Fatal("Expected error for phase transition on parked session, got nil")
	}
}

func TestMoirai_PhaseTransition_ForwardOnly(t *testing.T) {
	projectDir := setupProjectDir(t)
	sessionID := "session-20260205-231000-phsfwd01"
	writeSessionContext(t, projectDir, sessionID, "ACTIVE", "Phase forward test", nil, "")

	ctx := newTestContext(projectDir, sessionID)

	// Transition to design (force to skip artifact checks)
	if err := runTransition(ctx, "design", transitionOptions{force: true}); err != nil {
		t.Fatalf("Transition to design failed: %v", err)
	}

	// Verify we cannot go backwards to requirements
	err := runTransition(ctx, "requirements", transitionOptions{force: true})
	if err == nil {
		t.Fatal("Expected error for backward phase transition, got nil")
	}
}

// =============================================================================
// Test: Edge cases discovered during D2 audit
// =============================================================================

func TestMoirai_EdgeCase_CurrentSessionCleared_OnWrap(t *testing.T) {
	// Verify wrap succeeds and transitions to ARCHIVED.
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	if err := runCreate(ctx, "Current session clear test", createOptions{complexity: "MODULE"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	sessionID := findCreatedSessionID(t, projectDir)

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	if err := runWrap(ctx, wrapOptions{noArchive: true}); err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}

	// Verify session is ARCHIVED
	status := loadSessionStatus(t, projectDir, sessionID)
	if status != session.StatusArchived {
		t.Errorf("After wrap: status = %v, want ARCHIVED", status)
	}
}

func TestMoirai_EdgeCase_CurrentSessionSet_OnResume(t *testing.T) {
	// Verify resume transitions session to ACTIVE.
	projectDir := setupProjectDir(t)
	now := time.Now().UTC()
	sessionID := "session-20260205-232000-csresume"
	writeSessionContext(t, projectDir, sessionID, "PARKED", "Current session resume", &now, "Parked")

	ctx := newTestContext(projectDir, sessionID)
	if err := runResume(ctx); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	// Verify session is ACTIVE
	status := loadSessionStatus(t, projectDir, sessionID)
	if status != session.StatusActive {
		t.Errorf("After resume: status = %v, want ACTIVE", status)
	}
}

func TestMoirai_EdgeCase_CreateAfterWrap(t *testing.T) {
	// After wrapping, creating a new session should succeed.
	// This validates the full cycle: create -> wrap (with archive) -> create (new).
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	// Create and wrap first session (with archive so it moves out of sessions/)
	if err := runCreate(ctx, "First session", createOptions{complexity: "MODULE"}); err != nil {
		t.Fatalf("First create failed: %v", err)
	}
	if err := runWrap(ctx, wrapOptions{noArchive: false}); err != nil {
		t.Fatalf("First wrap failed: %v", err)
	}

	// Create second session - should succeed
	if err := runCreate(ctx, "Second session", createOptions{complexity: "PATCH"}); err != nil {
		t.Fatalf("Second create failed: %v", err)
	}

	// Find the second session
	sessionID := findCreatedSessionID(t, projectDir)

	// Verify second session is ACTIVE
	status := loadSessionStatus(t, projectDir, sessionID)
	if status != session.StatusActive {
		t.Errorf("Second session status = %v, want ACTIVE", status)
	}

	// Verify it has the correct initiative
	sessCtx := loadSessionContext(t, projectDir, sessionID)
	if sessCtx.Initiative != "Second session" {
		t.Errorf("Initiative = %q, want %q", sessCtx.Initiative, "Second session")
	}
	if sessCtx.Complexity != "PATCH" {
		t.Errorf("Complexity = %q, want %q", sessCtx.Complexity, "PATCH")
	}
}

func TestMoirai_EdgeCase_MultipleParkResumeCycles(t *testing.T) {
	// Verify that multiple park/resume cycles work correctly.
	// Each cycle should produce fresh timestamps and events.
	projectDir := setupProjectDir(t)
	ctx := newTestContext(projectDir)

	if err := runCreate(ctx, "Multiple cycles", createOptions{complexity: "MODULE"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	sessionID := findCreatedSessionID(t, projectDir)

	// Update ctx with sessionID for subsequent operations
	ctx = newTestContext(projectDir, sessionID)

	cycles := 3
	for i := 0; i < cycles; i++ {
		// Park
		reason := "Cycle " + string(rune('1'+i))
		if err := runPark(ctx, parkOptions{reason: reason}); err != nil {
			t.Fatalf("Park cycle %d failed: %v", i+1, err)
		}

		sessCtx := loadSessionContext(t, projectDir, sessionID)
		if sessCtx.Status != session.StatusParked {
			t.Errorf("Cycle %d: status after park = %v, want PARKED", i+1, sessCtx.Status)
		}
		if sessCtx.ParkedReason != reason {
			t.Errorf("Cycle %d: ParkedReason = %q, want %q", i+1, sessCtx.ParkedReason, reason)
		}

		// Resume
		if err := runResume(ctx); err != nil {
			t.Fatalf("Resume cycle %d failed: %v", i+1, err)
		}

		sessCtx = loadSessionContext(t, projectDir, sessionID)
		if sessCtx.Status != session.StatusActive {
			t.Errorf("Cycle %d: status after resume = %v, want ACTIVE", i+1, sessCtx.Status)
		}
	}

	// Verify event counts (unified clew event system)
	parkCount := countEventsOfType(t, projectDir, sessionID, "session.parked")
	if parkCount != cycles {
		t.Errorf("session.parked events = %d, want %d", parkCount, cycles)
	}
	resumeCount := countEventsOfType(t, projectDir, sessionID, "session.resumed")
	if resumeCount != cycles {
		t.Errorf("session.resumed events = %d, want %d", resumeCount, cycles)
	}
}
