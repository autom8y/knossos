package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	sess "github.com/autom8y/knossos/internal/session"
)

// --- helpers ---

// setupSnapshotTestSession creates a session environment for snapshot command tests.
// Returns (ctx, sessionDir, eventsPath).
func setupSnapshotTestSession(t *testing.T, contextBody string) (*cmdContext, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectDir := tmpDir
	sessionsDir := filepath.Join(projectDir, ".claude", "sessions")
	locksDir := filepath.Join(sessionsDir, ".locks")
	sessionID := "session-20260226-120000-snaptst1"
	sessionDir := filepath.Join(sessionsDir, sessionID)

	for _, dir := range []string{sessionDir, locksDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	if contextBody == "" {
		contextBody = "\n# Session: snapshot test\n\n## Timeline\n\n## Blockers\nNone yet.\n"
	}

	contextContent := "---\n" +
		"schema_version: \"2.1\"\n" +
		"session_id: " + sessionID + "\n" +
		"status: ACTIVE\n" +
		"initiative: snapshot test initiative\n" +
		"complexity: MODULE\n" +
		"active_rite: ecosystem\n" +
		"current_phase: implementation\n" +
		"created_at: 2026-02-26T12:00:00Z\n" +
		"---" + contextBody

	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(ctxPath, []byte(contextContent), 0644); err != nil {
		t.Fatalf("failed to write SESSION_CONTEXT.md: %v", err)
	}

	// Write .current-session marker.
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatalf("failed to write .current-session: %v", err)
	}

	eventsPath := filepath.Join(sessionDir, "events.jsonl")

	outputFormat := "text"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	return ctx, sessionDir, eventsPath
}

// writeSnapshotEvents appends TypedEvents to events.jsonl.
func writeSnapshotEvents(t *testing.T, eventsPath string, events []clewcontract.TypedEvent) {
	t.Helper()
	f, err := os.OpenFile(eventsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("failed to open events.jsonl: %v", err)
	}
	defer f.Close()

	for _, e := range events {
		line, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("failed to marshal event: %v", err)
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			t.Fatalf("failed to write event line: %v", err)
		}
	}
}

// --- Tests: command structure ---

// TestNewContextCmd verifies the context subcommand group is correctly configured.
func TestNewContextCmd(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir
	outputFormat := "text"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	cmd := newContextCmd(ctx)
	if cmd.Use != "context" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "context")
	}

	// Should have snapshot as a subcommand.
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "snapshot" {
			found = true
			break
		}
	}
	if !found {
		t.Error("context cmd should have 'snapshot' subcommand")
	}
}

// TestNewSnapshotCmd verifies flag configuration of the snapshot command.
func TestNewSnapshotCmd(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir
	outputFormat := "text"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	cmd := newSnapshotCmd(ctx)
	if cmd.Use != "snapshot" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "snapshot")
	}

	roleFlag := cmd.Flags().Lookup("role")
	if roleFlag == nil {
		t.Error("snapshot cmd missing --role flag")
	} else if roleFlag.DefValue != "orchestrator" {
		t.Errorf("--role default = %q, want %q", roleFlag.DefValue, "orchestrator")
	}

	agentFlag := cmd.Flags().Lookup("agent")
	if agentFlag == nil {
		t.Error("snapshot cmd missing --agent flag")
	}
}

// --- Tests: runSnapshot ---

// TestRunSnapshot_NoActiveSession verifies error when no session exists.
func TestRunSnapshot_NoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir
	locksDir := filepath.Join(projectDir, ".claude", "sessions", ".locks")
	if err := os.MkdirAll(locksDir, 0755); err != nil {
		t.Fatalf("failed to create locks dir: %v", err)
	}

	outputFormat := "text"
	verbose := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
		},
	}

	err := runSnapshot(ctx, snapshotOptions{role: "orchestrator"})
	if err == nil {
		t.Fatal("expected error when no active session")
	}
}

// TestRunSnapshot_InvalidRole verifies error for unknown role string.
func TestRunSnapshot_InvalidRole(t *testing.T) {
	ctx, _, _ := setupSnapshotTestSession(t, "")

	err := runSnapshot(ctx, snapshotOptions{role: "superagent"})
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
	if !strings.Contains(err.Error(), "invalid role") {
		t.Errorf("error message should mention 'invalid role': %v", err)
	}
}

// TestRunSnapshot_OrchestratorTextOutput verifies orchestrator text output succeeds.
func TestRunSnapshot_OrchestratorTextOutput(t *testing.T) {
	ctx, _, eventsPath := setupSnapshotTestSession(t, "")

	writeSnapshotEvents(t, eventsPath, []clewcontract.TypedEvent{
		clewcontract.NewTypedDecisionRecordedEvent("use CSS variables", "runtime perf", nil),
	})

	if err := runSnapshot(ctx, snapshotOptions{role: "orchestrator", agentName: "pythia"}); err != nil {
		t.Fatalf("runSnapshot orchestrator returned error: %v", err)
	}
}

// TestRunSnapshot_SpecialistTextOutput verifies specialist output succeeds.
func TestRunSnapshot_SpecialistTextOutput(t *testing.T) {
	ctx, _, _ := setupSnapshotTestSession(t, "")

	if err := runSnapshot(ctx, snapshotOptions{role: "specialist", agentName: "integration-engineer"}); err != nil {
		t.Fatalf("runSnapshot specialist returned error: %v", err)
	}
}

// TestRunSnapshot_BackgroundTextOutput verifies background output succeeds.
func TestRunSnapshot_BackgroundTextOutput(t *testing.T) {
	ctx, _, _ := setupSnapshotTestSession(t, "")

	if err := runSnapshot(ctx, snapshotOptions{role: "background", agentName: "linter"}); err != nil {
		t.Fatalf("runSnapshot background returned error: %v", err)
	}
}

// TestRunSnapshot_JSONOutput verifies JSON output format flag accepted.
func TestRunSnapshot_JSONOutput(t *testing.T) {
	ctx, _, _ := setupSnapshotTestSession(t, "")
	jsonFormat := "json"
	ctx.Output = &jsonFormat

	if err := runSnapshot(ctx, snapshotOptions{role: "orchestrator", agentName: "pythia"}); err != nil {
		t.Fatalf("runSnapshot JSON output returned error: %v", err)
	}
}

// TestRunSnapshot_WithBlockers verifies command succeeds with blockers in context.
func TestRunSnapshot_WithBlockers(t *testing.T) {
	body := "\n# Session: snapshot test\n\n## Timeline\n\n## Blockers\n- critical dependency missing\n"
	ctx, _, _ := setupSnapshotTestSession(t, body)

	if err := runSnapshot(ctx, snapshotOptions{role: "orchestrator"}); err != nil {
		t.Fatalf("runSnapshot with blockers returned error: %v", err)
	}
}

// --- Tests: snapshot integration via library ---

// TestRunSnapshot_JSONStructureViaLibrary validates JSON structure via GenerateSnapshot.
func TestRunSnapshot_JSONStructureViaLibrary(t *testing.T) {
	ctx, _, eventsPath := setupSnapshotTestSession(t, "")

	writeSnapshotEvents(t, eventsPath, []clewcontract.TypedEvent{
		clewcontract.NewTypedDecisionRecordedEvent("chose X", "fast", nil),
	})

	resolver := ctx.GetResolver()
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		t.Fatalf("GetSessionID error: %v", err)
	}

	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext error: %v", err)
	}

	snap, err := sess.GenerateSnapshot(sessCtx, eventsPath, sess.SnapshotConfig{
		Role:      sess.RoleOrchestrator,
		AgentName: "pythia",
	})
	if err != nil {
		t.Fatalf("GenerateSnapshot error: %v", err)
	}

	raw, err := snap.RenderJSON()
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if out["role"] != "orchestrator" {
		t.Errorf("role = %v, want orchestrator", out["role"])
	}
	if out["current_phase"] != "implementation" {
		t.Errorf("current_phase = %v, want implementation", out["current_phase"])
	}
	if out["complexity"] != "MODULE" {
		t.Errorf("complexity = %v, want MODULE", out["complexity"])
	}
	if out["active_rite"] != "ecosystem" {
		t.Errorf("active_rite = %v, want ecosystem", out["active_rite"])
	}
	// generated_at should be present.
	if _, ok := out["generated_at"]; !ok {
		t.Error("missing generated_at in orchestrator JSON")
	}
	// decisions should be present (we wrote one decision event).
	if _, ok := out["decisions"]; !ok {
		t.Error("missing decisions in orchestrator JSON")
	}
}

// TestRunSnapshot_MarkdownViaLibrary validates markdown rendering via GenerateSnapshot.
func TestRunSnapshot_MarkdownViaLibrary(t *testing.T) {
	ctx, _, eventsPath := setupSnapshotTestSession(t, "")

	writeSnapshotEvents(t, eventsPath, []clewcontract.TypedEvent{
		clewcontract.NewTypedAgentDelegatedEvent(clewcontract.SourceHook, "context-architect", "specialist", "", ""),
	})

	resolver := ctx.GetResolver()
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		t.Fatalf("GetSessionID error: %v", err)
	}

	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadContext error: %v", err)
	}

	snap, err := sess.GenerateSnapshot(sessCtx, eventsPath, sess.SnapshotConfig{
		Role:      sess.RoleSpecialist,
		AgentName: "context-architect",
	})
	if err != nil {
		t.Fatalf("GenerateSnapshot error: %v", err)
	}

	md := snap.RenderMarkdown()

	// Must have the standard header.
	if !strings.Contains(md, "## Session Context (auto-injected)") {
		t.Errorf("markdown missing header: %q", md)
	}
	// Specialist shows "(recent)" in timeline header, not "(last N)".
	if len(snap.Timeline) > 0 && !strings.Contains(md, "### Timeline (recent)") {
		t.Errorf("specialist timeline header should be '(recent)': %q", md)
	}
	// Specialist should NOT have ### Decisions.
	if strings.Contains(md, "### Decisions") {
		t.Errorf("specialist should not have decisions section: %q", md)
	}
}
