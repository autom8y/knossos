package handoff

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/ariadne/internal/hook/threadcontract"
)

// TestPrepare_EmitsTaskEnd verifies that handoff prepare emits a task_end event.
func TestPrepare_EmitsTaskEnd(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140000-prep1234"
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
initiative: Test Handoff Prepare
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

	// Run prepare command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := prepareOptions{
		fromAgent: "architect",
		toAgent:   "principal-engineer",
	}

	err := runPrepare(ctx, opts)
	if err != nil {
		t.Fatalf("runPrepare failed: %v", err)
	}

	// Verify events.jsonl contains task_end event
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	// Find task_end event
	found := false
	for _, event := range events {
		if event.Type == threadcontract.EventTypeTaskEnd {
			found = true

			// Verify agent is the source agent
			agent, ok := event.Meta["agent"].(string)
			if !ok {
				t.Errorf("task_end event missing agent in meta")
				continue
			}
			if agent != "architect" {
				t.Errorf("task_end event agent = %q, want %q", agent, "architect")
			}

			// Verify status is success
			status, ok := event.Meta["status"].(string)
			if !ok {
				t.Errorf("task_end event missing status in meta")
				continue
			}
			if status != "success" {
				t.Errorf("task_end event status = %q, want %q", status, "success")
			}

			break
		}
	}

	if !found {
		t.Error("events.jsonl does not contain task_end event for prepare operation")
	}
}

// TestExecute_EmitsTaskStart verifies that handoff execute emits a task_start event.
func TestExecute_EmitsTaskStart(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140001-exec5678"
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
initiative: Test Handoff Execute
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

	// Run execute command
	outputFormat := "json"
	verbose := true
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := executeOptions{
		artifactID: "TDD-test-feature",
		toAgent:    "principal-engineer",
	}

	err := runExecute(ctx, opts)
	if err != nil {
		t.Fatalf("runExecute failed: %v", err)
	}

	// Verify events.jsonl contains task_start event
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	// Find task_start event
	found := false
	for _, event := range events {
		if event.Type == threadcontract.EventTypeTaskStart {
			found = true

			// Verify agent is the target agent
			agent, ok := event.Meta["agent"].(string)
			if !ok {
				t.Errorf("task_start event missing agent in meta")
				continue
			}
			if agent != "principal-engineer" {
				t.Errorf("task_start event agent = %q, want %q", agent, "principal-engineer")
			}

			// Verify parent_session is set
			parentSession, ok := event.Meta["parent_session"].(string)
			if !ok {
				t.Errorf("task_start event missing parent_session in meta")
				continue
			}
			if parentSession != sessionID {
				t.Errorf("task_start event parent_session = %q, want %q", parentSession, sessionID)
			}

			break
		}
	}

	if !found {
		t.Error("events.jsonl does not contain task_start event for execute operation")
	}

	// Also verify HANDOFF_EXECUTED event exists
	handoffFound := false
	content, _ := os.ReadFile(eventsPath)
	if strings.Contains(string(content), "HANDOFF_EXECUTED") {
		handoffFound = true
	}

	if !handoffFound {
		t.Error("events.jsonl does not contain HANDOFF_EXECUTED event")
	}
}

// TestStatus_ReturnsCurrentState verifies that handoff status returns current state.
func TestStatus_ReturnsCurrentState(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140002-stat9012"
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
	createdAt := time.Now().UTC().Add(-30 * time.Minute)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Handoff Status
complexity: FEATURE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
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

	// Run status command
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	err := runStatus(ctx)
	if err != nil {
		t.Fatalf("runStatus failed: %v", err)
	}

	// Status command doesn't return anything testable directly
	// but it should not error on a valid session
}

// TestHistory_QueriesEvents verifies that handoff history queries events.jsonl.
func TestHistory_QueriesEvents(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140003-hist3456"
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
initiative: Test Handoff History
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

	// Seed some events in events.jsonl
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events := []string{
		`{"ts":"2026-01-05T14:00:00.000Z","type":"task_start","summary":"Task delegated to architect","meta":{"agent":"architect","description":"Design TDD","parent_session":"` + sessionID + `"}}`,
		`{"ts":"2026-01-05T14:30:00.000Z","type":"task_end","summary":"Task completed by architect: success","meta":{"agent":"architect","status":"success","throughline":"TDD completed","artifacts":["docs/design/TDD-test.md"],"duration_ms":1800000}}`,
		`{"timestamp":"2026-01-05T14:30:01.000Z","event":"HANDOFF_EXECUTED","to":"principal-engineer","metadata":{"artifact_id":"TDD-test","from_phase":"design","target_phase":"implementation"}}`,
	}
	if err := os.WriteFile(eventsPath, []byte(strings.Join(events, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("Failed to write events.jsonl: %v", err)
	}

	// Run history command
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := historyOptions{
		limit: 0, // unlimited
	}

	err := runHistory(ctx, opts)
	if err != nil {
		t.Fatalf("runHistory failed: %v", err)
	}

	// The command outputs to stdout, which we can't easily capture in this test
	// but it should not error when events exist
}

// TestPrepare_InvalidHandoffSequence verifies that invalid handoff sequences are rejected.
func TestPrepare_InvalidHandoffSequence(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140004-invl7890"
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
initiative: Test Invalid Handoff
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

	// Run prepare command with invalid sequence (architect -> qa-adversary skips principal-engineer)
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := prepareOptions{
		fromAgent: "architect",
		toAgent:   "qa-adversary", // Invalid: should go to principal-engineer first
	}

	err := runPrepare(ctx, opts)
	if err == nil {
		t.Error("runPrepare should fail for invalid handoff sequence")
	}

	// Verify error message indicates lifecycle violation
	if err != nil && !strings.Contains(err.Error(), "invalid handoff sequence") {
		t.Errorf("Expected error about invalid handoff sequence, got: %v", err)
	}
}

// TestPrepare_SelfHandoff verifies that self-handoff is rejected (C3 edge case).
func TestPrepare_SelfHandoff(t *testing.T) {
	agents := []string{"requirements-analyst", "architect", "principal-engineer", "qa-adversary", "orchestrator"}

	for _, agent := range agents {
		t.Run(agent, func(t *testing.T) {
			// Create temporary project structure
			tmpDir := t.TempDir()
			projectDir := tmpDir

			// Create .claude/sessions directory structure
			sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
			sessionID := "session-20260105-self-" + agent
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
initiative: Test Self Handoff
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

			// Run prepare command with self-handoff
			outputFormat := "json"
			verbose := false
			ctx := &cmdContext{
				output:     &outputFormat,
				verbose:    &verbose,
				projectDir: &projectDir,
			}

			opts := prepareOptions{
				fromAgent: agent,
				toAgent:   agent, // Self-handoff
			}

			err := runPrepare(ctx, opts)
			if err == nil {
				t.Errorf("runPrepare should fail for self-handoff %s -> %s", agent, agent)
			}

			// Verify error message indicates self-handoff is not allowed
			if err != nil && !strings.Contains(err.Error(), "self-handoff not allowed") {
				t.Errorf("Expected error about self-handoff, got: %v", err)
			}
		})
	}
}

// TestPrepare_AllInvalidSequences verifies all 11 invalid handoff sequences are rejected (C3 edge case).
func TestPrepare_AllInvalidSequences(t *testing.T) {
	// Define all invalid transitions (11 total as per C3 requirements)
	invalidTransitions := []struct {
		from string
		to   string
		desc string
	}{
		{"requirements-analyst", "principal-engineer", "skip architect"},
		{"requirements-analyst", "qa-adversary", "skip architect and principal-engineer"},
		{"requirements-analyst", "orchestrator", "invalid backward"},
		{"architect", "qa-adversary", "skip principal-engineer"},
		{"architect", "requirements-analyst", "invalid backward"},
		{"architect", "orchestrator", "invalid sideways"},
		{"principal-engineer", "requirements-analyst", "invalid backward"},
		{"principal-engineer", "architect", "invalid backward"},
		{"principal-engineer", "orchestrator", "invalid sideways"},
		{"qa-adversary", "requirements-analyst", "invalid backward (qa can't go to req)"},
		{"qa-adversary", "principal-engineer", "invalid backward"},
	}

	for _, tc := range invalidTransitions {
		t.Run(tc.from+"_to_"+tc.to, func(t *testing.T) {
			// Create temporary project structure
			tmpDir := t.TempDir()
			projectDir := tmpDir

			// Create .claude/sessions directory structure
			sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
			sessionID := "session-20260105-inv-" + tc.from + "-" + tc.to
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
initiative: Test Invalid Handoff Sequence
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

			// Run prepare command with invalid sequence
			outputFormat := "json"
			verbose := false
			ctx := &cmdContext{
				output:     &outputFormat,
				verbose:    &verbose,
				projectDir: &projectDir,
			}

			opts := prepareOptions{
				fromAgent: tc.from,
				toAgent:   tc.to,
			}

			err := runPrepare(ctx, opts)
			if err == nil {
				t.Errorf("runPrepare should fail for invalid handoff %s -> %s (%s)", tc.from, tc.to, tc.desc)
			}

			// Verify error message indicates invalid sequence
			if err != nil && !strings.Contains(err.Error(), "invalid handoff sequence") {
				t.Errorf("Expected error about invalid handoff sequence for %s -> %s, got: %v", tc.from, tc.to, err)
			}
		})
	}
}

// TestPrepare_CrossTeamValidation verifies agents must exist in active_team (C3 edge case).
func TestPrepare_CrossTeamValidation(t *testing.T) {
	testCases := []struct {
		name       string
		activeTeam string
		fromAgent  string
		toAgent    string
		shouldFail bool
		errorMsg   string
	}{
		{
			name:       "valid agents in 10x-dev-pack",
			activeTeam: "10x-dev-pack",
			fromAgent:  "architect",
			toAgent:    "principal-engineer",
			shouldFail: false,
		},
		{
			name:       "invalid from agent not in team",
			activeTeam: "consultant-pack",
			fromAgent:  "architect",
			toAgent:    "orchestrator",
			shouldFail: true,
			errorMsg:   "source agent not in active team",
		},
		{
			name:       "invalid to agent not in team",
			activeTeam: "consultant-pack",
			fromAgent:  "orchestrator",
			toAgent:    "architect",
			shouldFail: true,
			errorMsg:   "target agent not in active team",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary project structure
			tmpDir := t.TempDir()
			projectDir := tmpDir

			// Create .claude/sessions directory structure
			sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
			sessionID := "session-20260105-team-" + tc.name
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

			// Create SESSION_CONTEXT.md with ACTIVE status and specific team
			createdAt := time.Now().UTC().Add(-1 * time.Hour)
			contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Cross-Team Validation
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: ` + tc.activeTeam + `
current_phase: design
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

			// Run prepare command
			outputFormat := "json"
			verbose := false
			ctx := &cmdContext{
				output:     &outputFormat,
				verbose:    &verbose,
				projectDir: &projectDir,
			}

			opts := prepareOptions{
				fromAgent: tc.fromAgent,
				toAgent:   tc.toAgent,
			}

			err := runPrepare(ctx, opts)
			if tc.shouldFail {
				if err == nil {
					t.Errorf("runPrepare should fail for cross-team validation: %s", tc.name)
				}
				if err != nil && !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("runPrepare should succeed for valid team agents: %v", err)
				}
			}
		})
	}
}

// TestStatus_NoSession verifies that status returns error when no session exists.
func TestStatus_NoSession(t *testing.T) {
	// Create temporary project structure with no session
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create minimal .claude structure but no session
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude", "sessions"), 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	// Run status command
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	err := runStatus(ctx)
	if err == nil {
		t.Error("runStatus should fail when no session exists")
	}
}

// TestHistory_EmptyEvents verifies that history handles empty events gracefully.
func TestHistory_EmptyEvents(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140005-empt1234"
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
	createdAt := time.Now().UTC().Add(-15 * time.Minute)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Empty History
complexity: TRIVIAL
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: requirements
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

	// Note: No events.jsonl file - should handle gracefully

	// Run history command
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := historyOptions{
		limit: 0,
	}

	err := runHistory(ctx, opts)
	if err != nil {
		t.Fatalf("runHistory should not fail with empty/missing events: %v", err)
	}
}

// TestExecute_DryRunValidation verifies that execute --dry-run still validates (C3 edge case).
func TestExecute_DryRunValidation(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-dryval-1234"
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
initiative: Test Dry Run Validation
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

	// Test 1: dry-run with invalid agent should fail validation
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := executeOptions{
		artifactID: "TDD-dry-run-test",
		toAgent:    "invalid-agent", // Invalid agent
		dryRun:     true,
	}

	err := runExecute(ctx, opts)
	if err == nil {
		t.Error("runExecute with dry-run should fail validation for invalid agent")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid target agent") {
		t.Errorf("Expected error about invalid agent, got: %v", err)
	}

	// Test 2: dry-run with valid agent should succeed but not emit events
	opts.toAgent = "principal-engineer"
	err = runExecute(ctx, opts)
	if err != nil {
		t.Fatalf("runExecute with dry-run and valid agent should succeed: %v", err)
	}

	// Verify no events were written
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	if _, err := os.Stat(eventsPath); !os.IsNotExist(err) {
		// File exists, check if it's empty
		content, _ := os.ReadFile(eventsPath)
		if len(content) > 0 {
			t.Error("Dry-run should not create events in events.jsonl")
		}
	}
}

// TestExecute_DryRun verifies that execute --dry-run does not emit events.
func TestExecute_DryRun(t *testing.T) {
	// Create temporary project structure
	tmpDir := t.TempDir()
	projectDir := tmpDir

	// Create .claude/sessions directory structure
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-140006-dryr5678"
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
initiative: Test Dry Run
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_team: 10x-dev-pack
current_phase: design
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

	// Run execute command with --dry-run
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := executeOptions{
		artifactID: "TDD-dry-run-test",
		toAgent:    "principal-engineer",
		dryRun:     true,
	}

	err := runExecute(ctx, opts)
	if err != nil {
		t.Fatalf("runExecute with dry-run failed: %v", err)
	}

	// Verify no events were written
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	if _, err := os.Stat(eventsPath); !os.IsNotExist(err) {
		// File exists, check if it's empty
		content, _ := os.ReadFile(eventsPath)
		if len(content) > 0 {
			t.Error("Dry-run should not create events in events.jsonl")
		}
	}
}

// readEventsFile reads and parses events.jsonl into threadcontract.Event structs.
func readEventsFile(path string) ([]threadcontract.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []threadcontract.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var event threadcontract.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Some events may use the old format, skip those
			continue
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
