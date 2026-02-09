package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/session"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// TestPark_EmitsSessionEnd verifies that parking a session emits a session_end event
// with status=parked to events.jsonl via Thread Contract v2.
func TestPark_EmitsSessionEnd(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-120000-park1234"
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

	// Create SESSION_CONTEXT.md with ACTIVE status
	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Park Session
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev
current_phase: requirements
---

# Session Context

## Session Type
standard
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Run park command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := parkOptions{
		reason: "Test park reason",
	}

	err := runPark(ctx, opts)
	if err != nil {
		t.Fatalf("runPark failed: %v", err)
	}

	// Verify events.jsonl contains session_end event with status=parked
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	// Find session_end event
	found := false
	for _, event := range events {
		if event.Type == clewcontract.EventTypeSessionEnd {
			found = true

			// Verify status is "parked"
			status, ok := event.Meta["status"].(string)
			if !ok {
				t.Errorf("session_end event missing status in meta")
				continue
			}
			if status != "parked" {
				t.Errorf("session_end event status = %q, want %q", status, "parked")
			}

			// Verify session_id is present
			sid, ok := event.Meta["session_id"].(string)
			if !ok {
				t.Errorf("session_end event missing session_id in meta")
				continue
			}
			if sid != sessionID {
				t.Errorf("session_end event session_id = %q, want %q", sid, sessionID)
			}

			// Verify duration_ms is present and positive
			durationMs, ok := event.Meta["duration_ms"].(float64) // JSON numbers are float64
			if !ok {
				t.Errorf("session_end event missing duration_ms in meta")
				continue
			}
			if durationMs <= 0 {
				t.Errorf("session_end event duration_ms = %v, want > 0", durationMs)
			}

			// Verify summary
			expectedSummary := "Session ended: parked"
			if event.Summary != expectedSummary {
				t.Errorf("session_end event summary = %q, want %q", event.Summary, expectedSummary)
			}

			break
		}
	}

	if !found {
		t.Error("events.jsonl does not contain session_end event for park operation")
	}
}

// TestWrap_EmitsSessionEnd verifies that wrapping a session emits a session_end event
// with status=completed to events.jsonl via Thread Contract v2.
func TestWrap_EmitsSessionEnd(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-120001-wrap5678"
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

	// Create SESSION_CONTEXT.md with ACTIVE status
	createdAt := time.Now().UTC().Add(-2 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Wrap Session
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev
current_phase: qa
---

# Session Context

## Session Type
standard
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Run wrap command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := wrapOptions{
		noArchive: true, // Don't move to archive for easier testing
	}

	err := runWrap(ctx, opts)
	if err != nil {
		t.Fatalf("runWrap failed: %v", err)
	}

	// Verify events.jsonl contains session_end event with status=completed
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	// Find session_end event
	found := false
	for _, event := range events {
		if event.Type == clewcontract.EventTypeSessionEnd {
			found = true

			// Verify status is "completed"
			status, ok := event.Meta["status"].(string)
			if !ok {
				t.Errorf("session_end event missing status in meta")
				continue
			}
			if status != "completed" {
				t.Errorf("session_end event status = %q, want %q", status, "completed")
			}

			// Verify session_id is present
			sid, ok := event.Meta["session_id"].(string)
			if !ok {
				t.Errorf("session_end event missing session_id in meta")
				continue
			}
			if sid != sessionID {
				t.Errorf("session_end event session_id = %q, want %q", sid, sessionID)
			}

			// Verify duration_ms is present and positive
			durationMs, ok := event.Meta["duration_ms"].(float64)
			if !ok {
				t.Errorf("session_end event missing duration_ms in meta")
				continue
			}
			if durationMs <= 0 {
				t.Errorf("session_end event duration_ms = %v, want > 0", durationMs)
			}

			// Verify summary
			expectedSummary := "Session ended: completed"
			if event.Summary != expectedSummary {
				t.Errorf("session_end event summary = %q, want %q", event.Summary, expectedSummary)
			}

			break
		}
	}

	if !found {
		t.Error("events.jsonl does not contain session_end event for wrap operation")
	}

	// Also verify session was properly archived (status changed)
	ctx2, err := session.LoadContext(filepath.Join(sessionDir, "SESSION_CONTEXT.md"))
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}
	if ctx2.Status != session.StatusArchived {
		t.Errorf("Session status = %v, want %v", ctx2.Status, session.StatusArchived)
	}
}

// TestResume_NoSessionEnd verifies that resuming a session does NOT emit a session_end event.
// Resume should only emit SESSION_RESUMED, not session_end.
func TestResume_NoSessionEnd(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-120002-rsum9012"
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

	// Create SESSION_CONTEXT.md with ACTIVE status first
	createdAt := time.Now().UTC().Add(-3 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Resume Session
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev
current_phase: implementation
---

# Session Context

## Session Type
standard
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// First, park the session
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	parkOpts := parkOptions{
		reason: "Pausing for resume test",
	}

	err := runPark(ctx, parkOpts)
	if err != nil {
		t.Fatalf("runPark failed: %v", err)
	}

	// Read events after park to get baseline count
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsAfterPark, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl after park: %v", err)
	}

	// Count session_end events after park
	sessionEndCountAfterPark := 0
	for _, event := range eventsAfterPark {
		if event.Type == clewcontract.EventTypeSessionEnd {
			sessionEndCountAfterPark++
		}
	}

	// Verify we have exactly 1 session_end from the park operation
	if sessionEndCountAfterPark != 1 {
		t.Fatalf("Expected 1 session_end event after park, got %d", sessionEndCountAfterPark)
	}

	// Now resume the session
	err = runResume(ctx)
	if err != nil {
		t.Fatalf("runResume failed: %v", err)
	}

	// Read events after resume
	eventsAfterResume, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl after resume: %v", err)
	}

	// Count session_end events after resume
	sessionEndCountAfterResume := 0
	for _, event := range eventsAfterResume {
		if event.Type == clewcontract.EventTypeSessionEnd {
			sessionEndCountAfterResume++
		}
	}

	// Verify session_end count did NOT increase after resume
	if sessionEndCountAfterResume != sessionEndCountAfterPark {
		t.Errorf("session_end count changed after resume: before=%d, after=%d. Resume should NOT emit session_end",
			sessionEndCountAfterPark, sessionEndCountAfterResume)
	}

	// Also verify that SESSION_RESUMED event WAS emitted (to the session events file)
	// This uses the session.EventEmitter, not threadcontract
	sessionEventsPath := filepath.Join(sessionDir, "events.jsonl")
	content, err := os.ReadFile(sessionEventsPath)
	if err != nil {
		t.Fatalf("Failed to read session events: %v", err)
	}

	// Verify SESSION_RESUMED exists in the file
	if !strings.Contains(string(content), "SESSION_RESUMED") {
		t.Error("events.jsonl should contain SESSION_RESUMED event from resume operation")
	}

	// Verify session status is ACTIVE after resume
	ctx2, err := session.LoadContext(filepath.Join(sessionDir, "SESSION_CONTEXT.md"))
	if err != nil {
		t.Fatalf("Failed to load session context: %v", err)
	}
	if ctx2.Status != session.StatusActive {
		t.Errorf("Session status = %v, want %v", ctx2.Status, session.StatusActive)
	}
}

// TestPark_EmitsSessionParkedAndSessionEnd verifies both event types are emitted.
// SESSION_PARKED comes from session.EventEmitter, session_end comes from clewcontract.
func TestPark_EmitsBothEventTypes(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-120003-both3456"
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

	// Create SESSION_CONTEXT.md with ACTIVE status
	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Both Events
complexity: FEATURE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev
current_phase: implementation
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Create current-session file
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Run park command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	opts := parkOptions{
		reason: "Testing both event types",
	}

	err := runPark(ctx, opts)
	if err != nil {
		t.Fatalf("runPark failed: %v", err)
	}

	// Read events file
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	content, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	// Verify SESSION_PARKED event exists (from session.EventEmitter)
	if !strings.Contains(string(content), "SESSION_PARKED") {
		t.Error("events.jsonl should contain SESSION_PARKED event")
	}

	// Verify session_end event exists (from threadcontract)
	if !strings.Contains(string(content), `"type":"session.ended"`) {
		t.Error("events.jsonl should contain session.ended event")
	}

	// Verify both events have correct structure by parsing
	events, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to parse events: %v", err)
	}

	hasSessionEnd := false
	for _, event := range events {
		if event.Type == clewcontract.EventTypeSessionEnd {
			hasSessionEnd = true
			if event.Meta["status"] != "parked" {
				t.Errorf("session_end status = %v, want parked", event.Meta["status"])
			}
		}
	}

	if !hasSessionEnd {
		t.Error("Parsed events do not contain session_end")
	}
}

// readEventsFile reads and parses events.jsonl into clewcontract.Event structs.
// This handles the JSONL format where each line is a separate JSON object.
func readEventsFile(path string) ([]clewcontract.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []clewcontract.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var event clewcontract.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Some events may use the old format (session.Event), skip those
			continue
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
