package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/autom8y/knossos/test/hooks/testutil"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func TestAutoparkOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   AutoparkOutput
		contains string
	}{
		{
			name: "was parked",
			output: AutoparkOutput{
				SessionID: "session-test",
				WasParked: true,
			},
			contains: "Session auto-parked: session-test",
		},
		{
			name: "not parked",
			output: AutoparkOutput{
				WasParked: false,
				Message:   "no active session",
			},
			contains: "no active session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.Text()
			if result != tt.contains {
				t.Errorf("Text() = %q, want %q", result, tt.contains)
			}
		})
	}
}

func TestRunAutopark_EarlyExit_HooksDisabled(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.WasParked {
		t.Error("Expected WasParked=false when no context")
	}
}

func TestRunAutopark_WrongEvent(t *testing.T) {
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart", // Not a Stop event
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.WasParked {
		t.Error("Expected WasParked=false for non-Stop event")
	}
	if result.Message != "not a stop event" {
		t.Errorf("Message = %q, want %q", result.Message, "not a stop event")
	}
}

func TestRunAutopark_NoSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal .sos structure (no session)
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "Stop",
		ProjectDir:  tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: nil,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.WasParked {
		t.Error("Expected WasParked=false when no session")
	}
}

func TestRunAutopark_ActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write SESSION_CONTEXT.md with ACTIVE status (unquoted for scan-based discovery)
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: ACTIVE
created_at: "2026-01-04T22:26:13Z"
initiative: "Test"
complexity: "MODULE"
active_rite: "test"
current_phase: "implementation"
---
`
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(contextFile, []byte(sessionContext), 0644); err != nil {
		t.Fatalf("Failed to write session context: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "Stop",
		ProjectDir:  tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: nil,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.WasParked {
		t.Error("Expected WasParked=true")
	}
	if result.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", result.SessionID, sessionID)
	}
	if result.Status != "PARKED" {
		t.Errorf("Status = %q, want %q", result.Status, "PARKED")
	}
	if result.PreviousStatus != "ACTIVE" {
		t.Errorf("PreviousStatus = %q, want %q", result.PreviousStatus, "ACTIVE")
	}
	if result.AutoParkedAt == "" {
		t.Error("AutoParkedAt should not be empty")
	}

	// Verify the file was actually updated
	updatedContent, err := os.ReadFile(contextFile)
	if err != nil {
		t.Fatalf("Failed to read updated context: %v", err)
	}
	if !bytes.Contains(updatedContent, []byte("status: PARKED")) {
		t.Error("Context file should contain 'status: PARKED'")
	}
	if !bytes.Contains(updatedContent, []byte("park_source: auto")) {
		t.Errorf("Context file should contain 'park_source: auto', got:\n%s", string(updatedContent))
	}
}

func TestRunAutopark_AlreadyParked(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write SESSION_CONTEXT.md with PARKED status
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: PARKED
created_at: "2026-01-04T22:26:13Z"
initiative: "Test"
complexity: "MODULE"
active_rite: "test"
current_phase: "implementation"
parked_at: "2026-01-04T23:00:00Z"
parked_reason: "manual"
---
`
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	os.WriteFile(contextFile, []byte(sessionContext), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "Stop",
		ProjectDir:  tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	sessionIDPtr := sessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: &sessionIDPtr,
		},
	}

	err := runAutoparkCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runAutopark() error = %v", err)
	}

	var result AutoparkOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if result.WasParked {
		t.Error("Expected WasParked=false when already parked")
	}
	if !bytes.Contains([]byte(result.Message), []byte("not active")) {
		t.Errorf("Message should indicate session not active, got: %q", result.Message)
	}
}

// --- dismissZombieSummonedAgents tests ---

// makeSummonManifest writes a USER_PROVENANCE_MANIFEST.yaml with a summon entry for the
// given agent name into claudeDir. agentFile is the expected path of the agent file
// that will be written alongside (so tests can verify removal).
func makeSummonManifest(t *testing.T, claudeDir, name string) string {
	t.Helper()
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Write a dummy agent file that the dismiss logic will remove.
	agentFile := filepath.Join(agentsDir, name+".md")
	if err := os.WriteFile(agentFile, []byte("# "+name+"\n"), 0644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/" + name + ".md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "summon:" + name,
				SourceType: "summon",
				Checksum:   "sha256:" + strings.Repeat("a", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.UserManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	return agentFile
}

func TestDismissZombieSummonedAgents_NoSummonedAgents(t *testing.T) {
	// Manifest exists but has no summon entries — should be a graceful no-op.
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/rite-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "rites/ecosystem/agents/rite-agent.md",
				SourceType: "project",
				Checksum:   "sha256:" + strings.Repeat("b", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.UserManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	var stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &bytes.Buffer{}, &stderr, false)

	result := dismissZombieSummonedAgents(printer)
	if len(result) != 0 {
		t.Errorf("expected no dismissed agents, got %v", result)
	}
}

func TestDismissZombieSummonedAgents_WithSummonedAgents(t *testing.T) {
	// Manifest has a summon entry — agent file and manifest entry must both be removed.
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	t.Setenv("HOME", fakeHome)

	agentFile := makeSummonManifest(t, claudeDir, "zombie-agent")

	var stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &bytes.Buffer{}, &stderr, false)

	dismissed := dismissZombieSummonedAgents(printer)

	if len(dismissed) != 1 {
		t.Fatalf("expected 1 dismissed agent, got %d: %v", len(dismissed), dismissed)
	}
	if dismissed[0] != "zombie-agent" {
		t.Errorf("dismissed agent name = %q, want %q", dismissed[0], "zombie-agent")
	}

	// Agent file should be gone.
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		t.Errorf("agent file %s should have been removed", agentFile)
	}

	// Manifest entry should be gone.
	manifestPath := provenance.UserManifestPath(claudeDir)
	manifest, err := provenance.Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to reload manifest: %v", err)
	}
	if _, exists := manifest.Entries["agents/zombie-agent.md"]; exists {
		t.Error("manifest entry for zombie-agent should have been removed")
	}
}

func TestDismissZombieSummonedAgents_MissingManifest(t *testing.T) {
	// No manifest file at all — should degrade gracefully and return nil.
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	var stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &bytes.Buffer{}, &stderr, false)

	result := dismissZombieSummonedAgents(printer)
	if result != nil {
		t.Errorf("expected nil dismissed list when manifest is missing, got %v", result)
	}
}

func TestDismissZombieSummonedAgents_AgentFileAlreadyGone(t *testing.T) {
	// The agent file was already deleted externally. The manifest entry should still be removed.
	fakeHome := t.TempDir()
	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// Write manifest with summon entry but no agent file on disk.
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: "2.0",
		LastSync:      time.Now().UTC(),
		Entries: map[string]*provenance.ProvenanceEntry{
			"agents/ghost-agent.md": {
				Owner:      provenance.OwnerKnossos,
				Scope:      provenance.ScopeUser,
				SourcePath: "summon:ghost-agent",
				SourceType: "summon",
				Checksum:   "sha256:" + strings.Repeat("c", 64),
				LastSynced: time.Now().UTC(),
			},
		},
	}
	manifestPath := provenance.UserManifestPath(claudeDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	var stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &bytes.Buffer{}, &stderr, false)

	dismissed := dismissZombieSummonedAgents(printer)

	// Should still report the agent as dismissed (manifest cleaned up).
	if len(dismissed) != 1 || dismissed[0] != "ghost-agent" {
		t.Errorf("expected ['ghost-agent'] dismissed, got %v", dismissed)
	}

	// Manifest entry should be gone.
	loaded, err := provenance.Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to reload manifest: %v", err)
	}
	if _, exists := loaded.Entries["agents/ghost-agent.md"]; exists {
		t.Error("manifest entry for ghost-agent should have been removed")
	}
}

func TestAutoparkOutput_Text_WithDismissedAgents(t *testing.T) {
	// Parked session + dismissed agents → both lines present.
	out := AutoparkOutput{
		SessionID:       "session-xyz",
		WasParked:       true,
		DismissedAgents: []string{"theoros", "naxos"},
	}
	text := out.Text()
	if !strings.Contains(text, "Session auto-parked: session-xyz") {
		t.Errorf("Text() missing park line, got: %q", text)
	}
	if !strings.Contains(text, "Dismissed zombie agents: theoros, naxos") {
		t.Errorf("Text() missing dismiss line, got: %q", text)
	}
}

func TestAutoparkOutput_Text_DismissedWithoutPark(t *testing.T) {
	// No park, but dismissed agents → dismiss line present, message shown.
	out := AutoparkOutput{
		WasParked:       false,
		Message:         "no active session",
		DismissedAgents: []string{"orphan"},
	}
	text := out.Text()
	if !strings.Contains(text, "no active session") {
		t.Errorf("Text() missing message, got: %q", text)
	}
	if !strings.Contains(text, "Dismissed zombie agents: orphan") {
		t.Errorf("Text() missing dismiss line, got: %q", text)
	}
}

// BenchmarkAutoparkHook_EarlyExit benchmarks the early exit path.
func BenchmarkAutoparkHook_EarlyExit(b *testing.B) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionID := ""

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = runAutoparkCore(nil, ctx, printer)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Early exit took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

