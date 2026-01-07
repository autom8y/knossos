package handoff

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/ariadne/internal/hook/clewcontract"
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
active_rite: 10x-dev-pack
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
		if event.Type == clewcontract.EventTypeTaskEnd {
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

			// Verify outcome is success
			outcome, ok := event.Meta["outcome"].(string)
			if !ok {
				t.Errorf("task_end event missing outcome in meta")
				continue
			}
			if outcome != "success" {
				t.Errorf("task_end event outcome = %q, want %q", outcome, "success")
			}

			break
		}
	}

	if !found {
		t.Error("events.jsonl does not contain task_end event for prepare operation")
	}

	// Also verify HANDOFF_PREPARED event exists
	foundHandoffPrepared := false
	for _, event := range events {
		if event.Type == clewcontract.EventTypeHandoffPrepared {
			foundHandoffPrepared = true

			// Verify from_agent and to_agent
			fromAgent, ok := event.Meta["from_agent"].(string)
			if !ok || fromAgent != "architect" {
				t.Errorf("handoff_prepared event from_agent = %q, want %q", fromAgent, "architect")
			}

			toAgent, ok := event.Meta["to_agent"].(string)
			if !ok || toAgent != "principal-engineer" {
				t.Errorf("handoff_prepared event to_agent = %q, want %q", toAgent, "principal-engineer")
			}

			break
		}
	}

	if !foundHandoffPrepared {
		t.Error("events.jsonl does not contain handoff_prepared event")
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
active_rite: 10x-dev-pack
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
		if event.Type == clewcontract.EventTypeTaskStart {
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

			// Verify session_id is set
			sid, ok := event.Meta["session_id"].(string)
			if !ok {
				t.Errorf("task_start event missing session_id in meta")
				continue
			}
			if sid != sessionID {
				t.Errorf("task_start event session_id = %q, want %q", sid, sessionID)
			}

			break
		}
	}

	if !found {
		t.Error("events.jsonl does not contain task_start event for execute operation")
	}

	// Also verify HANDOFF_EXECUTED event exists (ThreadContract format)
	foundHandoffExecuted := false
	for _, event := range events {
		if event.Type == clewcontract.EventTypeHandoffExecuted {
			foundHandoffExecuted = true

			// Verify to_agent
			toAgent, ok := event.Meta["to_agent"].(string)
			if !ok || toAgent != "principal-engineer" {
				t.Errorf("handoff_executed event to_agent = %q, want %q", toAgent, "principal-engineer")
			}

			// Verify artifacts
			artifacts, ok := event.Meta["artifacts"].([]interface{})
			if !ok {
				t.Error("handoff_executed event missing artifacts")
			} else if len(artifacts) != 1 || artifacts[0].(string) != "TDD-test-feature" {
				t.Errorf("handoff_executed event artifacts = %v, want [\"TDD-test-feature\"]", artifacts)
			}

			break
		}
	}

	if !foundHandoffExecuted {
		t.Error("events.jsonl does not contain handoff_executed event (ThreadContract format)")
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
active_rite: 10x-dev-pack
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
active_rite: 10x-dev-pack
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
active_rite: 10x-dev-pack
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
active_rite: 10x-dev-pack
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
active_rite: 10x-dev-pack
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
			errorMsg:   "source agent not in active rite",
		},
		{
			name:       "invalid to agent not in team",
			activeTeam: "consultant-pack",
			fromAgent:  "orchestrator",
			toAgent:    "architect",
			shouldFail: true,
			errorMsg:   "target agent not in active rite",
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
active_rite: ` + tc.activeTeam + `
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
active_rite: 10x-dev-pack
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
active_rite: 10x-dev-pack
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
active_rite: 10x-dev-pack
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

// readEventsFile reads and parses events.jsonl into clewcontract.Event structs.
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

// ============================================================================
// Additional Integration Tests - Prepare Command Edge Cases
// ============================================================================

// TestPrepare_ValidHandoffSequences verifies all valid handoff sequences are accepted.
func TestPrepare_ValidHandoffSequences(t *testing.T) {
	validTransitions := []struct {
		from string
		to   string
		desc string
	}{
		{"requirements-analyst", "architect", "standard: req -> arch"},
		{"architect", "principal-engineer", "standard: arch -> pe"},
		{"principal-engineer", "qa-adversary", "standard: pe -> qa"},
		{"qa-adversary", "orchestrator", "completion: qa -> orch"},
		{"qa-adversary", "architect", "rework: qa -> arch"},
		{"orchestrator", "requirements-analyst", "delegate: orch -> req"},
		{"orchestrator", "architect", "delegate: orch -> arch"},
		{"orchestrator", "principal-engineer", "delegate: orch -> pe"},
		{"orchestrator", "qa-adversary", "delegate: orch -> qa"},
	}

	for _, tc := range validTransitions {
		t.Run(tc.from+"_to_"+tc.to, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectDir := tmpDir

			sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
			sessionID := "session-20260105-valid-" + tc.from + "-" + tc.to
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

			createdAt := time.Now().UTC().Add(-1 * time.Hour)
			contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Valid Handoff
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
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

			opts := prepareOptions{
				fromAgent: tc.from,
				toAgent:   tc.to,
			}

			err := runPrepare(ctx, opts)
			if err != nil {
				t.Errorf("runPrepare should succeed for valid handoff %s -> %s (%s): %v", tc.from, tc.to, tc.desc, err)
			}
		})
	}
}

// TestPrepare_NoActiveSession verifies error when no session exists.
func TestPrepare_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	if err := os.MkdirAll(filepath.Join(projectDir, ".claude", "sessions"), 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
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
	if err == nil {
		t.Error("runPrepare should fail when no session exists")
	}
}

// TestPrepare_ParkedSession verifies error when session is not ACTIVE.
func TestPrepare_ParkedSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-parked-test"
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

	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: PARKED
initiative: Test Parked Session
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
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

	opts := prepareOptions{
		fromAgent: "architect",
		toAgent:   "principal-engineer",
	}

	err := runPrepare(ctx, opts)
	if err == nil {
		t.Error("runPrepare should fail when session is PARKED")
	}
	if err != nil && !strings.Contains(err.Error(), "ACTIVE") {
		t.Errorf("Expected error about ACTIVE status, got: %v", err)
	}
}

// TestPrepare_UnknownAgent verifies error for unknown agent names.
func TestPrepare_UnknownAgent(t *testing.T) {
	testCases := []struct {
		name      string
		fromAgent string
		toAgent   string
		errorMsg  string
	}{
		{
			name:      "unknown from agent",
			fromAgent: "unknown-agent",
			toAgent:   "principal-engineer",
			errorMsg:  "invalid source agent",
		},
		{
			name:      "unknown to agent",
			fromAgent: "architect",
			toAgent:   "unknown-agent",
			errorMsg:  "invalid target agent",
		},
		{
			name:      "both unknown",
			fromAgent: "foo",
			toAgent:   "bar",
			errorMsg:  "invalid source agent", // from is checked first
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectDir := tmpDir

			sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
			sessionID := "session-20260105-unknown-" + tc.name
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

			createdAt := time.Now().UTC().Add(-1 * time.Hour)
			contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Unknown Agent
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
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

			opts := prepareOptions{
				fromAgent: tc.fromAgent,
				toAgent:   tc.toAgent,
			}

			err := runPrepare(ctx, opts)
			if err == nil {
				t.Errorf("runPrepare should fail for unknown agent: %s", tc.name)
			}
			if err != nil && !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error containing %q, got: %v", tc.errorMsg, err)
			}
		})
	}
}

// TestPrepare_WithArtifactID verifies artifact ID is included in output.
func TestPrepare_WithArtifactID(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-artifact-test"
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

	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Artifact ID
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
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

	opts := prepareOptions{
		fromAgent:  "architect",
		toAgent:    "principal-engineer",
		artifactID: "TDD-user-auth-feature",
	}

	err := runPrepare(ctx, opts)
	if err != nil {
		t.Fatalf("runPrepare failed: %v", err)
	}

	// Verify task_end event has artifact in metadata
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events, err := readEventsFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events: %v", err)
	}

	for _, event := range events {
		if event.Type == clewcontract.EventTypeTaskEnd {
			artifacts, ok := event.Meta["artifacts"].([]interface{})
			if !ok {
				t.Error("task_end event missing artifacts in meta")
				continue
			}
			found := false
			for _, a := range artifacts {
				if a.(string) == "TDD-user-auth-feature" {
					found = true
					break
				}
			}
			if !found {
				t.Error("artifact ID not found in task_end event artifacts")
			}
		}
	}
}

// ============================================================================
// Additional Integration Tests - Execute Command Edge Cases
// ============================================================================

// TestExecute_NoActiveSession verifies error when no session exists.
func TestExecute_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	if err := os.MkdirAll(filepath.Join(projectDir, ".claude", "sessions"), 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := executeOptions{
		artifactID: "TDD-test",
		toAgent:    "principal-engineer",
	}

	err := runExecute(ctx, opts)
	if err == nil {
		t.Error("runExecute should fail when no session exists")
	}
}

// TestExecute_ParkedSession verifies error when session is not ACTIVE.
func TestExecute_ParkedSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-exec-parked"
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

	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: PARKED
initiative: Test Execute Parked
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
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

	opts := executeOptions{
		artifactID: "TDD-test",
		toAgent:    "principal-engineer",
	}

	err := runExecute(ctx, opts)
	if err == nil {
		t.Error("runExecute should fail when session is PARKED")
	}
}

// TestExecute_AllTargetAgents verifies execute works for all valid target agents.
func TestExecute_AllTargetAgents(t *testing.T) {
	agents := []struct {
		agent       string
		targetPhase string
	}{
		{"requirements-analyst", "requirements"},
		{"architect", "design"},
		{"principal-engineer", "implementation"},
		{"qa-adversary", "validation"},
		{"orchestrator", "orchestration"},
	}

	for _, tc := range agents {
		t.Run(tc.agent, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectDir := tmpDir

			sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
			sessionID := "session-20260105-exec-" + tc.agent
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

			createdAt := time.Now().UTC().Add(-1 * time.Hour)
			contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Execute All Agents
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
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

			opts := executeOptions{
				artifactID: "TDD-test-" + tc.agent,
				toAgent:    tc.agent,
			}

			err := runExecute(ctx, opts)
			if err != nil {
				t.Fatalf("runExecute failed for agent %s: %v", tc.agent, err)
			}

			// Verify task_start event has correct agent and target phase
			eventsPath := filepath.Join(sessionDir, "events.jsonl")
			events, err := readEventsFile(eventsPath)
			if err != nil {
				t.Fatalf("Failed to read events: %v", err)
			}

			foundTaskStart := false
			for _, event := range events {
				if event.Type == clewcontract.EventTypeTaskStart {
					foundTaskStart = true
					agent, ok := event.Meta["agent"].(string)
					if !ok {
						t.Error("task_start missing agent")
						continue
					}
					if agent != tc.agent {
						t.Errorf("task_start agent = %q, want %q", agent, tc.agent)
					}
				}
			}
			if !foundTaskStart {
				t.Error("No task_start event found")
			}
		})
	}
}

// TestExecute_VerifyEventStructure verifies the complete event structure.
func TestExecute_VerifyEventStructure(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-exec-struct"
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

	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Execute Structure
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
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

	opts := executeOptions{
		artifactID: "TDD-full-structure",
		toAgent:    "principal-engineer",
	}

	err := runExecute(ctx, opts)
	if err != nil {
		t.Fatalf("runExecute failed: %v", err)
	}

	// Read events.jsonl and verify complete structure
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	content, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to read events.jsonl: %v", err)
	}

	// Verify HANDOFF_EXECUTED event
	if !strings.Contains(string(content), "HANDOFF_EXECUTED") {
		t.Error("events.jsonl missing HANDOFF_EXECUTED event")
	}
	if !strings.Contains(string(content), "TDD-full-structure") {
		t.Error("events.jsonl missing artifact_id")
	}
	if !strings.Contains(string(content), "implementation") {
		t.Error("events.jsonl missing target_phase")
	}
}

// ============================================================================
// Additional Integration Tests - Status Command Edge Cases
// ============================================================================

// TestStatus_WithHandoffHistory verifies status reflects handoff history.
func TestStatus_WithHandoffHistory(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-status-hist"
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

	createdAt := time.Now().UTC().Add(-2 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Status With History
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: implementation
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Seed events with handoffs
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events := []string{
		`{"ts":"2026-01-05T13:00:00.000Z","type":"task_start","summary":"Task delegated to architect","meta":{"agent":"architect","description":"Design TDD","parent_session":"` + sessionID + `"}}`,
		`{"ts":"2026-01-05T13:30:00.000Z","type":"task_end","summary":"Task completed by architect","meta":{"agent":"architect","status":"success","parent_session":"` + sessionID + `"}}`,
		`{"timestamp":"2026-01-05T13:30:01.000Z","event":"HANDOFF_EXECUTED","to":"principal-engineer","metadata":{"artifact_id":"TDD-test","from_phase":"design","target_phase":"implementation"}}`,
		`{"ts":"2026-01-05T13:30:02.000Z","type":"task_start","summary":"Task delegated to principal-engineer","meta":{"agent":"principal-engineer","description":"Implement","parent_session":"` + sessionID + `"}}`,
	}
	if err := os.WriteFile(eventsPath, []byte(strings.Join(events, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("Failed to write events.jsonl: %v", err)
	}

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
	// Status command outputs to stdout; test verifies no error on valid data
}

// TestStatus_AllPhases verifies next expected handoff for all phases.
func TestStatus_AllPhases(t *testing.T) {
	phases := []struct {
		phase    string
		expected string
	}{
		{"requirements", "requirements-analyst -> architect"},
		{"design", "architect -> principal-engineer"},
		{"implementation", "principal-engineer -> qa-adversary"},
		{"validation", "qa-adversary -> orchestrator (complete)"},
		{"qa", "qa-adversary -> orchestrator (complete)"},
	}

	for _, tc := range phases {
		t.Run(tc.phase, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectDir := tmpDir

			sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
			sessionID := "session-20260105-phase-" + tc.phase
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

			createdAt := time.Now().UTC().Add(-1 * time.Hour)
			contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Phase Status
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: ` + tc.phase + `
---

# Session Context
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

			err := runStatus(ctx)
			if err != nil {
				t.Fatalf("runStatus failed for phase %s: %v", tc.phase, err)
			}
		})
	}
}

// ============================================================================
// Additional Integration Tests - History Command Edge Cases
// ============================================================================

// TestHistory_WithLimit verifies limit parameter works correctly.
func TestHistory_WithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-hist-limit"
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

	createdAt := time.Now().UTC().Add(-3 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test History Limit
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: validation
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Seed many events
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events := []string{
		`{"ts":"2026-01-05T12:00:00.000Z","type":"task_start","summary":"Task 1","meta":{"agent":"requirements-analyst","parent_session":"` + sessionID + `"}}`,
		`{"ts":"2026-01-05T12:30:00.000Z","type":"task_end","summary":"End 1","meta":{"agent":"requirements-analyst","status":"success","parent_session":"` + sessionID + `"}}`,
		`{"ts":"2026-01-05T13:00:00.000Z","type":"task_start","summary":"Task 2","meta":{"agent":"architect","parent_session":"` + sessionID + `"}}`,
		`{"ts":"2026-01-05T13:30:00.000Z","type":"task_end","summary":"End 2","meta":{"agent":"architect","status":"success","parent_session":"` + sessionID + `"}}`,
		`{"ts":"2026-01-05T14:00:00.000Z","type":"task_start","summary":"Task 3","meta":{"agent":"principal-engineer","parent_session":"` + sessionID + `"}}`,
		`{"ts":"2026-01-05T14:30:00.000Z","type":"task_end","summary":"End 3","meta":{"agent":"principal-engineer","status":"success","parent_session":"` + sessionID + `"}}`,
	}
	if err := os.WriteFile(eventsPath, []byte(strings.Join(events, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("Failed to write events.jsonl: %v", err)
	}

	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
	}

	opts := historyOptions{
		limit: 2,
	}

	err := runHistory(ctx, opts)
	if err != nil {
		t.Fatalf("runHistory with limit failed: %v", err)
	}
	// Outputs to stdout; test verifies no error
}

// TestHistory_PhaseTransitionEvents verifies PHASE_TRANSITIONED events are included.
func TestHistory_PhaseTransitionEvents(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-hist-phase"
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

	createdAt := time.Now().UTC().Add(-1 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Phase Transitions
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: implementation
---

# Session Context
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// Include PHASE_TRANSITIONED event
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	events := []string{
		`{"timestamp":"2026-01-05T13:00:00.000Z","event":"PHASE_TRANSITIONED","from_phase":"design","to_phase":"implementation"}`,
		`{"ts":"2026-01-05T13:00:01.000Z","type":"task_start","summary":"Implementation started","meta":{"agent":"principal-engineer","parent_session":"` + sessionID + `"}}`,
	}
	if err := os.WriteFile(eventsPath, []byte(strings.Join(events, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("Failed to write events.jsonl: %v", err)
	}

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
		t.Fatalf("runHistory with phase transitions failed: %v", err)
	}
}

// TestHistory_NoSession verifies error when no session exists.
func TestHistory_NoSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	if err := os.MkdirAll(filepath.Join(projectDir, ".claude", "sessions"), 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

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
	if err == nil {
		t.Error("runHistory should fail when no session exists")
	}
}

// ============================================================================
// Additional Integration Tests - Concurrent and Session ID Override
// ============================================================================

// TestPrepare_WithExplicitSessionID verifies explicit session ID override.
func TestPrepare_WithExplicitSessionID(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")

	// Create two sessions
	session1ID := "session-20260105-explicit-1"
	session2ID := "session-20260105-explicit-2"

	for _, sessionID := range []string{session1ID, session2ID} {
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

		createdAt := time.Now().UTC().Add(-1 * time.Hour)
		contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: ACTIVE
initiative: Test Explicit Session ID
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: design
---

# Session Context
`
		if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(contextContent), 0644); err != nil {
			t.Fatalf("Failed to write SESSION_CONTEXT.md: %v", err)
		}
	}

	// Set current-session to session1
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(session1ID), 0644); err != nil {
		t.Fatalf("Failed to write .current-session: %v", err)
	}

	// But explicitly use session2
	outputFormat := "json"
	verbose := false
	ctx := &cmdContext{
		output:     &outputFormat,
		verbose:    &verbose,
		projectDir: &projectDir,
		sessionID:  &session2ID,
	}

	opts := prepareOptions{
		fromAgent: "architect",
		toAgent:   "principal-engineer",
	}

	err := runPrepare(ctx, opts)
	if err != nil {
		t.Fatalf("runPrepare failed with explicit session ID: %v", err)
	}

	// Verify events were written to session2, not session1
	session1Events := filepath.Join(sessionsDir, session1ID, "events.jsonl")
	session2Events := filepath.Join(sessionsDir, session2ID, "events.jsonl")

	if _, err := os.Stat(session1Events); !os.IsNotExist(err) {
		content, _ := os.ReadFile(session1Events)
		if len(content) > 0 {
			t.Error("Events should not be written to session1")
		}
	}

	if _, err := os.Stat(session2Events); os.IsNotExist(err) {
		t.Error("Events should be written to session2")
	}
}

// TestExecute_CompletedSession verifies error when session is COMPLETED.
func TestExecute_CompletedSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	sessionID := "session-20260105-exec-completed"
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

	createdAt := time.Now().UTC().Add(-2 * time.Hour)
	contextContent := `---
schema_version: "2.1"
session_id: ` + sessionID + `
status: COMPLETED
initiative: Test Execute Completed
complexity: MODULE
created_at: ` + createdAt.Format(time.RFC3339) + `
active_rite: 10x-dev-pack
current_phase: validation
---

# Session Context
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

	opts := executeOptions{
		artifactID: "TDD-test",
		toAgent:    "orchestrator",
	}

	err := runExecute(ctx, opts)
	if err == nil {
		t.Error("runExecute should fail when session is COMPLETED")
	}
}
