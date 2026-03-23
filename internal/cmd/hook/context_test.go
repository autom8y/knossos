package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/materialize/source"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
	"github.com/autom8y/knossos/test/hooks/testutil"
)

func TestContextOutput_Text(t *testing.T) {
	tests := []struct {
		name     string
		output   ContextOutput
		contains []string
		absent   []string
	}{
		{
			name: "with session",
			output: ContextOutput{
				SessionID:     "session-20260104-222613-05a12c6b",
				Status:        "ACTIVE",
				Initiative:    "Test Initiative",
				Rite:          "10x-dev",
				CurrentPhase:  "design",
				ExecutionMode: "orchestrated",
				HasSession:    true,
			},
			contains: []string{
				"---\n",
				"# Session Context (injected by ari hook context)\n",
				"session_id: session-20260104-222613-05a12c6b\n",
				"status: ACTIVE\n",
				"initiative: \"Test Initiative\"\n",
				"active_rite: 10x-dev\n",
				"execution_mode: orchestrated\n",
				"current_phase: design\n",
			},
			absent: []string{
				"## Session Context",
				"| Session |",
				"| Status |",
				"| Rite |",
				"| Mode |",
				"No active session",
			},
		},
		{
			name:   "no session",
			output: ContextOutput{HasSession: false},
			contains: []string{
				"---\n",
				"# Session Context (injected by ari hook context)\n",
				"has_session: false\n",
			},
			absent: []string{
				"No active session",
				"session_id:",
				"status:",
			},
		},
		{
			name: "no session with harness_session_id",
			output: ContextOutput{
				HasSession:  false,
				HarnessSessionID: "cc-abc123",
			},
			contains: []string{
				"has_session: false\n",
				"harness_session_id: \"cc-abc123\"\n",
			},
		},
		{
			name: "with git and ecosystem fields",
			output: ContextOutput{
				SessionID:       "session-001",
				Status:          "ACTIVE",
				ExecutionMode:   "orchestrated",
				HasSession:      true,
				GitBranch:       "feat/backtick-migration",
				BaseBranch:      "main",
				AvailableRites:  []string{"10x-dev", "hygiene", "ecosystem"},
				AvailableAgents: []string{"orchestrator", "context-architect"},
			},
			contains: []string{
				"git_branch: feat/backtick-migration\n",
				"base_branch: main\n",
				"available_rites:\n",
				"  - 10x-dev\n",
				"  - hygiene\n",
				"  - ecosystem\n",
				"available_agents:\n",
				"  - orchestrator\n",
				"  - context-architect\n",
			},
			absent: []string{
				"| Available Rites |",
				"| Available Agents |",
			},
		},
		{
			name: "omits empty optional fields",
			output: ContextOutput{
				SessionID:     "session-002",
				Status:        "ACTIVE",
				ExecutionMode: "cross-cutting",
				HasSession:    true,
			},
			contains: []string{
				"---\n",
				"session_id: session-002\n",
			},
			absent: []string{
				"git_branch:",
				"base_branch:",
				"available_rites:",
				"available_agents:",
				"frayed_from:",
				"frame_ref:",
				"park_source:",
				"current_phase:",
			},
		},
		{
			name: "with active procession",
			output: ContextOutput{
				SessionID:     "session-004",
				Status:        "ACTIVE",
				ExecutionMode: "orchestrated",
				HasSession:    true,
				Procession: &session.Procession{
					ID:             "security-remediation-2026-03-10",
					Type:           "security-remediation",
					CurrentStation: "assess",
					CompletedStations: []session.CompletedStation{
						{Station: "audit", Rite: "security", CompletedAt: "2026-03-10T12:00:00Z"},
					},
					NextStation: "plan",
					NextRite:    "debt-triage",
					ArtifactDir: ".sos/wip/security-remediation/",
				},
			},
			contains: []string{
				"procession:\n",
				"  id: security-remediation-2026-03-10\n",
				"  type: security-remediation\n",
				"  current_station: assess\n",
				"  completed_stations:\n",
				"    - station: audit\n",
				"      rite: security\n",
				"      completed_at: \"2026-03-10T12:00:00Z\"\n",
				"  next_station: plan\n",
				"  next_rite: debt-triage\n",
				"  artifact_dir: .sos/wip/security-remediation/\n",
			},
		},
		{
			name: "procession without completed stations or next",
			output: ContextOutput{
				SessionID:     "session-005",
				Status:        "ACTIVE",
				ExecutionMode: "orchestrated",
				HasSession:    true,
				Procession: &session.Procession{
					ID:             "sec-rem-2026-03-10",
					Type:           "security-remediation",
					CurrentStation: "audit",
					ArtifactDir:    ".sos/wip/sec/",
				},
			},
			contains: []string{
				"procession:\n",
				"  id: sec-rem-2026-03-10\n",
				"  current_station: audit\n",
			},
			absent: []string{
				"completed_stations:",
				"next_station:",
				"next_rite:",
			},
		},
		{
			name: "nil procession omitted",
			output: ContextOutput{
				SessionID:     "session-006",
				Status:        "ACTIVE",
				ExecutionMode: "cross-cutting",
				HasSession:    true,
			},
			absent: []string{
				"procession:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.Text()
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("Text() missing expected content: %q\nGot: %s", want, result)
				}
			}
			for _, notWant := range tt.absent {
				if strings.Contains(result, notWant) {
					t.Errorf("Text() contains unexpected content: %q\nGot: %s", notWant, result)
				}
			}
		})
	}
}

func TestContextOutput_Text_OmitsEmptyFields(t *testing.T) {
	// Verify that empty optional fields do not appear in YAML output
	out := ContextOutput{
		SessionID:     "session-003",
		Status:        "ACTIVE",
		ExecutionMode: "cross-cutting",
		HasSession:    true,
		GitBranch:     "",
		BaseBranch:    "",
	}
	text := out.Text()
	if strings.Contains(text, "git_branch:") {
		t.Error("Text() should not contain git_branch when empty")
	}
	if strings.Contains(text, "base_branch:") {
		t.Error("Text() should not contain base_branch when empty")
	}
	if strings.Contains(text, "available_rites:") {
		t.Error("Text() should not contain available_rites when nil")
	}
	if strings.Contains(text, "available_agents:") {
		t.Error("Text() should not contain available_agents when nil")
	}

	// Verify that populated git fields ARE included in Text() (new YAML format)
	out2 := ContextOutput{
		SessionID:     "session-004",
		Status:        "ACTIVE",
		ExecutionMode: "orchestrated",
		HasSession:    true,
		GitBranch:     "feature/something",
		BaseBranch:    "main",
	}
	text2 := out2.Text()
	if !strings.Contains(text2, "git_branch: feature/something") {
		t.Error("Text() should contain git_branch when populated (now in YAML frontmatter)")
	}
	if !strings.Contains(text2, "base_branch: main") {
		t.Error("Text() should contain base_branch when populated (now in YAML frontmatter)")
	}
}

func TestDeriveExecutionMode(t *testing.T) {
	tests := []struct {
		name       string
		status     session.Status
		activeRite string
		want       string
	}{
		{
			name:       "active session with rite is orchestrated",
			status:     session.StatusActive,
			activeRite: "10x-dev",
			want:       "orchestrated",
		},
		{
			name:       "active session with 'none' rite is cross-cutting",
			status:     session.StatusActive,
			activeRite: "none",
			want:       "cross-cutting",
		},
		{
			name:       "active session with empty rite is cross-cutting",
			status:     session.StatusActive,
			activeRite: "",
			want:       "cross-cutting",
		},
		{
			name:       "parked session is cross-cutting",
			status:     session.StatusParked,
			activeRite: "10x-dev",
			want:       "cross-cutting",
		},
		{
			name:       "archived session is native",
			status:     session.StatusArchived,
			activeRite: "10x-dev",
			want:       "native",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.DeriveExecutionMode(tt.status, tt.activeRite)
			if result != tt.want {
				t.Errorf("DeriveExecutionMode() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestRunContext_EarlyExit_HooksDisabled(t *testing.T) {
	// Test with no session context
	testutil.SetupEnv(t, &testutil.HookEnv{})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	// Create a minimal context
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionIDVal := ""
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDVal,
		},
	}

	// Override getPrinter to use our buffer
	err := runContextCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	// Parse output
	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HasSession {
		t.Error("Expected HasSession=false when hooks disabled")
	}
}

func TestRunContext_WithActiveSession(t *testing.T) {
	// Create temp directory with session structure
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	// Create .sos/sessions structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	// Write SESSION_CONTEXT.md
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: ACTIVE
created_at: "2026-01-04T22:26:13Z"
initiative: "Hooks Migration"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "implementation"
---

# Session: Hooks Migration
`
	contextFile := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := os.WriteFile(contextFile, []byte(sessionContext), 0644); err != nil {
		t.Fatalf("Failed to write session context: %v", err)
	}

	// Write ACTIVE_RITE (.knossos/ holds framework state)
	knossosDir := filepath.Join(tmpDir, ".knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("Failed to create .knossos dir: %v", err)
	}
	activeRiteFile := filepath.Join(knossosDir, "ACTIVE_RITE")
	if err := os.WriteFile(activeRiteFile, []byte("10x-dev"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	// Setup environment
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
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

	err := runContextCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.HasSession {
		t.Error("Expected HasSession=true")
	}
	if result.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", result.SessionID, sessionID)
	}
	if result.Status != "ACTIVE" {
		t.Errorf("Status = %q, want %q", result.Status, "ACTIVE")
	}
	if result.Initiative != "Hooks Migration" {
		t.Errorf("Initiative = %q, want %q", result.Initiative, "Hooks Migration")
	}
	if result.ExecutionMode != "orchestrated" {
		t.Errorf("ExecutionMode = %q, want %q", result.ExecutionMode, "orchestrated")
	}
}

func TestRunContext_NoSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal .sos structure (no session)
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:       "SessionStart",
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

	err := runContextCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.HasSession {
		t.Error("Expected HasSession=false when no session exists")
	}
}

func TestRunContext_RehydratesCompactState(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-rehydrate"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	// Write SESSION_CONTEXT.md
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-rehydrate"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "Rehydration Test"
complexity: "MODULE"
active_rite: "ecosystem"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("ecosystem"), 0644)

	// Write COMPACT_STATE.md (simulating PreCompact wrote it)
	checkpointContent := "# Compact State Checkpoint\n\n| Field | Value |\n|-------|-------|\n| session_id | session-20260208-100000-rehydrate |\n| initiative | Rehydration Test |\n"
	os.WriteFile(filepath.Join(sessionDir, CompactCheckpointFile), []byte(checkpointContent), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
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

	err := runContextCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	// Verify compact state was rehydrated
	if result.CompactState == "" {
		t.Error("Expected CompactState to be populated from COMPACT_STATE.md")
	}
	if !strings.Contains(result.CompactState, "session-20260208-100000-rehydrate") {
		t.Error("CompactState should contain session_id")
	}
	if !strings.Contains(result.CompactState, "Rehydration Test") {
		t.Error("CompactState should contain initiative")
	}

	// Verify checkpoint was consumed (renamed)
	checkpointPath := filepath.Join(sessionDir, CompactCheckpointFile)
	if _, err := os.Stat(checkpointPath); err == nil {
		t.Error("COMPACT_STATE.md should have been renamed after consumption")
	}
	consumedPath := filepath.Join(sessionDir, CompactCheckpointConsumed)
	if _, err := os.Stat(consumedPath); os.IsNotExist(err) {
		t.Error("COMPACT_STATE.consumed.md should exist after consumption")
	}
}

func TestRunContext_NoCompactState(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-nochkpnt"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-nochkpnt"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "No Checkpoint"
complexity: "simple"
active_rite: "forge"
current_phase: "requirements"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("forge"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
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

	err := runContextCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runContext() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	// No checkpoint means CompactState should be empty
	if result.CompactState != "" {
		t.Errorf("Expected empty CompactState, got: %s", result.CompactState)
	}

	// Normal fields should still work
	if !result.HasSession {
		t.Error("Expected HasSession=true")
	}
	if result.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", result.SessionID, sessionID)
	}
}

func TestGetGitBranch_InGitRepo(t *testing.T) {
	// This test runs inside the knossos repo, so git branch should be populated
	branch := getGitBranch(".")
	if branch == "" {
		t.Skip("Not running inside a git repository")
	}
	// Branch name should be a non-empty string without newlines
	if strings.Contains(branch, "\n") {
		t.Errorf("getGitBranch() returned string with newline: %q", branch)
	}
}

func TestGetGitBranch_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	branch := getGitBranch(tmpDir)
	if branch != "" {
		t.Errorf("getGitBranch() in non-git dir returned %q, want empty", branch)
	}
}

func TestGetBaseBranch_FallbackToMain(t *testing.T) {
	// In a directory without a git remote, should fall back to "main"
	tmpDir := t.TempDir()
	base := getBaseBranch(tmpDir)
	if base != "main" {
		t.Errorf("getBaseBranch() fallback = %q, want %q", base, "main")
	}
}

func TestListAvailableAgents(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	// Create agent files
	os.WriteFile(filepath.Join(agentsDir, "orchestrator.md"), []byte("# Orchestrator"), 0644)
	os.WriteFile(filepath.Join(agentsDir, "architect.md"), []byte("# Architect"), 0644)

	// Create non-md file (should be excluded)
	os.WriteFile(filepath.Join(agentsDir, "config.yaml"), []byte("key: value"), 0644)

	// Create subdirectory (should be excluded)
	os.MkdirAll(filepath.Join(agentsDir, "subdir"), 0755)

	agents := listAvailableAgents(agentsDir)
	if len(agents) != 2 {
		t.Fatalf("listAvailableAgents() returned %d agents, want 2: %v", len(agents), agents)
	}
	// os.ReadDir returns sorted entries
	if agents[0] != "architect" {
		t.Errorf("listAvailableAgents()[0] = %q, want %q", agents[0], "architect")
	}
	if agents[1] != "orchestrator" {
		t.Errorf("listAvailableAgents()[1] = %q, want %q", agents[1], "orchestrator")
	}
}

func TestListAvailableAgents_NonexistentDir(t *testing.T) {
	agents := listAvailableAgents("/nonexistent/path/agents")
	if agents != nil {
		t.Errorf("listAvailableAgents() on missing dir = %v, want nil", agents)
	}
}

func TestRunContext_WithActiveSession_IncludesRitesAndAgents(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-abcdef01"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-abcdef01"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "Rites Test"
complexity: "MODULE"
active_rite: "alpha"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)

	// Create .knossos/ for framework state (ACTIVE_RITE)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("alpha"), 0644)

	// Create rites in .knossos/rites/ (project-local satellite rites).
	// SourceResolver checks project-local rites first, so these appear in
	// AvailableRites even when KNOSSOS_HOME is empty.
	ritesDir := filepath.Join(tmpDir, ".knossos", "rites")
	os.MkdirAll(filepath.Join(ritesDir, "alpha"), 0755)
	os.WriteFile(filepath.Join(ritesDir, "alpha", "manifest.yaml"), []byte("name: alpha"), 0644)
	os.MkdirAll(filepath.Join(ritesDir, "beta"), 0755)
	os.WriteFile(filepath.Join(ritesDir, "beta", "manifest.yaml"), []byte("name: beta"), 0644)

	// Create agents directory
	agentsDir := filepath.Join(tmpDir, paths.ClaudeChannel{}.DirName(), "agents")
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(filepath.Join(agentsDir, "orchestrator.md"), []byte("# Orch"), 0644)
	os.WriteFile(filepath.Join(agentsDir, "builder.md"), []byte("# Build"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
	})

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	outputFlag := "json"
	verboseFlag := false
	// Inject a source resolver scoped to only the project dir — no platform/user/org tiers.
	// This replaces the isolateKnossosHome anti-pattern with constructor injection.
	srcResolver := source.NewSourceResolverWithPaths(tmpDir, "", "", "")

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &tmpDir,
			},
			SessionID: nil,
		},
		sourceResolver: srcResolver,
	}

	err := runContextCore(nil, ctx, printer)
	if err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	// Verify that the two project-local rites are included.
	// Source resolver is scoped to project only, so exactly 2 rites appear.
	if len(result.AvailableRites) != 2 {
		t.Errorf("AvailableRites length = %d, want 2: %v", len(result.AvailableRites), result.AvailableRites)
	}
	foundAlpha, foundBeta := false, false
	for _, r := range result.AvailableRites {
		if r == "alpha" {
			foundAlpha = true
		}
		if r == "beta" {
			foundBeta = true
		}
	}
	if !foundAlpha {
		t.Errorf("AvailableRites missing 'alpha': %v", result.AvailableRites)
	}
	if !foundBeta {
		t.Errorf("AvailableRites missing 'beta': %v", result.AvailableRites)
	}

	// Verify agents
	if len(result.AvailableAgents) != 2 {
		t.Errorf("AvailableAgents length = %d, want 2: %v", len(result.AvailableAgents), result.AvailableAgents)
	}

	// GitBranch will be empty since tmpDir is not a git repo — that is expected
	// BaseBranch should fall back to "main"
	if result.BaseBranch != "main" {
		t.Errorf("BaseBranch = %q, want %q (fallback)", result.BaseBranch, "main")
	}
}

// BenchmarkContextHook_EarlyExit benchmarks the early exit path (<5ms target).
func BenchmarkContextHook_EarlyExit(b *testing.B) {
	outputFlag := "json"
	verboseFlag := false
	projectDir := ""
	sessionIDVal := ""

	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFlag,
				Verbose:    &verboseFlag,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionIDVal,
		},
	}

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = runContextCore(nil, ctx, printer)
	}

	// Report ns/op
	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(5*time.Millisecond) {
		b.Errorf("Early exit took %.2f ms, target is <5ms", nsPerOp/1e6)
	}
}

// BenchmarkContextHook_FullExecution benchmarks full execution (<100ms target).
func BenchmarkContextHook_FullExecution(b *testing.B) {
	tmpDir := b.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	// Setup session structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: ACTIVE
created_at: "2026-01-04T22:26:13Z"
initiative: "Benchmark"
complexity: "MODULE"
active_rite: "test"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("test"), 0644)

	os.Setenv("CLAUDE_HOOK_EVENT", "SessionStart")
	os.Setenv("CLAUDE_PROJECT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("CLAUDE_HOOK_EVENT")
		os.Unsetenv("CLAUDE_PROJECT_DIR")
	}()

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

	var stdout, stderr bytes.Buffer
	printer := output.NewPrinter(output.FormatJSON, &stdout, &stderr, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		_ = runContextCore(nil, ctx, printer)
	}

	elapsed := b.Elapsed()
	nsPerOp := float64(elapsed.Nanoseconds()) / float64(b.N)
	if nsPerOp > float64(100*time.Millisecond) {
		b.Errorf("Full execution took %.2f ms, target is <100ms", nsPerOp/1e6)
	}
}

func TestContextOutput_Text_ShowsThroughlineIDs(t *testing.T) {
	out := ContextOutput{
		SessionID:     "session-005",
		Status:        "ACTIVE",
		ExecutionMode: "orchestrated",
		HasSession:    true,
		ThroughlineIDs: map[string]string{
			"potnia": "agent-potnia-abc",
			"moirai": "agent-moirai-def",
		},
	}
	text := out.Text()
	// Throughline section appears after closing --- delimiter
	if !strings.Contains(text, "---\n\nThroughline Agents:") {
		t.Errorf("Text() should contain 'Throughline Agents:' section after closing ---\nGot: %s", text)
	}
	if !strings.Contains(text, "potnia: agent-potnia-abc") {
		t.Errorf("Text() should contain potnia ID\nGot: %s", text)
	}
	if !strings.Contains(text, "moirai: agent-moirai-def") {
		t.Errorf("Text() should contain moirai ID\nGot: %s", text)
	}
}

func TestContextOutput_Text_OmitsThroughlineIDsWhenEmpty(t *testing.T) {
	out := ContextOutput{
		SessionID:     "session-006",
		Status:        "ACTIVE",
		ExecutionMode: "orchestrated",
		HasSession:    true,
	}
	text := out.Text()
	if strings.Contains(text, "Throughline Agents") {
		t.Errorf("Text() should not contain 'Throughline Agents' when map is nil\nGot: %s", text)
	}
}

func TestRunContext_IncludesThroughlineIDs(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-throughline"

	// Create session structure
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-throughline"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "Throughline Test"
complexity: "MODULE"
active_rite: "ecosystem"
current_phase: "implementation"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("ecosystem"), 0644)

	// Write .throughline-ids.json as if SubagentStart had fired
	idData := `{"potnia":"agent-potnia-xyz","moirai":"agent-moirai-uvw"}`
	os.WriteFile(filepath.Join(sessionDir, ThroughlineIDsFile), []byte(idData), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
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

	if err := runContextCore(nil, ctx, printer); err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if !result.HasSession {
		t.Error("Expected HasSession=true")
	}
	if len(result.ThroughlineIDs) == 0 {
		t.Error("Expected ThroughlineIDs to be populated")
	}
	if result.ThroughlineIDs["potnia"] != "agent-potnia-xyz" {
		t.Errorf("potnia ID = %q, want %q", result.ThroughlineIDs["potnia"], "agent-potnia-xyz")
	}
	if result.ThroughlineIDs["moirai"] != "agent-moirai-uvw" {
		t.Errorf("moirai ID = %q, want %q", result.ThroughlineIDs["moirai"], "agent-moirai-uvw")
	}

	// Verify Text() output includes the IDs (post-frontmatter section)
	text := result.Text()
	if !strings.Contains(text, "Throughline Agents:") {
		t.Errorf("Text() missing Throughline Agents section\nGot: %s", text)
	}
}

func TestRunContext_NoThroughlineIDsFile(t *testing.T) {
	// When no .throughline-ids.json exists, ThroughlineIDs should be nil (omitted from JSON)
	tmpDir := t.TempDir()
	sessionID := "session-20260208-100000-nothroughline"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260208-100000-nothroughline"
status: ACTIVE
created_at: "2026-02-08T10:00:00Z"
initiative: "No Throughline"
complexity: "simple"
active_rite: "forge"
current_phase: "requirements"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("forge"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
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

	if err := runContextCore(nil, ctx, printer); err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if len(result.ThroughlineIDs) != 0 {
		t.Errorf("Expected empty ThroughlineIDs when no file exists, got: %v", result.ThroughlineIDs)
	}

	// Text() should not mention Throughline Agents
	text := result.Text()
	if strings.Contains(text, "Throughline Agents") {
		t.Errorf("Text() should not include Throughline Agents section when none tracked\nGot: %s", text)
	}
}

// --- S2 Field Widening Tests ---

func TestContextOutput_Text_WithComplexity(t *testing.T) {
	// Verify complexity appears in YAML frontmatter when set, and is absent when empty.
	t.Run("complexity set", func(t *testing.T) {
		out := ContextOutput{
			SessionID:     "session-20260306-000001-aaaabbbb",
			Status:        "ACTIVE",
			ExecutionMode: "orchestrated",
			HasSession:    true,
			Complexity:    "MODULE",
		}
		text := out.Text()
		if !strings.Contains(text, "complexity: MODULE\n") {
			t.Errorf("Text() missing 'complexity: MODULE'\nGot: %s", text)
		}
	})

	t.Run("complexity empty is absent", func(t *testing.T) {
		out := ContextOutput{
			SessionID:     "session-20260306-000002-aaaabbbb",
			Status:        "ACTIVE",
			ExecutionMode: "orchestrated",
			HasSession:    true,
			Complexity:    "",
		}
		text := out.Text()
		if strings.Contains(text, "complexity:") {
			t.Errorf("Text() should not contain 'complexity:' when empty\nGot: %s", text)
		}
	})
}

func TestContextOutput_Text_WithStrands(t *testing.T) {
	t.Run("zero strands omits key", func(t *testing.T) {
		out := ContextOutput{
			SessionID:     "session-20260306-000003-aaaabbbb",
			Status:        "ACTIVE",
			ExecutionMode: "orchestrated",
			HasSession:    true,
			Strands:       nil,
		}
		text := out.Text()
		if strings.Contains(text, "strands:") {
			t.Errorf("Text() should not contain 'strands:' when nil\nGot: %s", text)
		}
	})

	t.Run("single strand minimal", func(t *testing.T) {
		out := ContextOutput{
			SessionID:     "session-20260306-000004-aaaabbbb",
			Status:        "ACTIVE",
			ExecutionMode: "orchestrated",
			HasSession:    true,
			Strands: []StrandOutput{
				{SessionID: "session-xxx", Status: "ACTIVE"},
			},
		}
		text := out.Text()
		wantParts := []string{
			"strands:\n",
			"  - session_id: session-xxx\n",
			"    status: ACTIVE\n",
		}
		for _, want := range wantParts {
			if !strings.Contains(text, want) {
				t.Errorf("Text() missing %q\nGot: %s", want, text)
			}
		}
		if strings.Contains(text, "frame_ref:") {
			t.Errorf("Text() should not contain 'frame_ref:' when empty\nGot: %s", text)
		}
		if strings.Contains(text, "landed_at:") {
			t.Errorf("Text() should not contain 'landed_at:' when empty\nGot: %s", text)
		}
	})

	t.Run("multiple strands with optional fields", func(t *testing.T) {
		out := ContextOutput{
			SessionID:     "session-20260306-000005-aaaabbbb",
			Status:        "ACTIVE",
			ExecutionMode: "orchestrated",
			HasSession:    true,
			Strands: []StrandOutput{
				{SessionID: "session-aaa", Status: "ACTIVE"},
				{
					SessionID: "session-bbb",
					Status:    "LANDED",
					FrameRef:  ".sos/wip/frames/test.md",
					LandedAt:  "2026-03-06T18:30:00Z",
				},
			},
		}
		text := out.Text()
		wantParts := []string{
			"strands:\n",
			"  - session_id: session-aaa\n",
			"    status: ACTIVE\n",
			"  - session_id: session-bbb\n",
			"    status: LANDED\n",
			"    frame_ref: .sos/wip/frames/test.md\n",
			"    landed_at: \"2026-03-06T18:30:00Z\"\n",
		}
		for _, want := range wantParts {
			if !strings.Contains(text, want) {
				t.Errorf("Text() missing %q\nGot: %s", want, text)
			}
		}
	})
}

func TestContextOutput_Text_WithClaimedBy(t *testing.T) {
	t.Run("claimed_by set", func(t *testing.T) {
		out := ContextOutput{
			SessionID:     "session-20260306-000006-aaaabbbb",
			Status:        "ACTIVE",
			ExecutionMode: "orchestrated",
			HasSession:    true,
			ClaimedBy:     "cc-session-abc123",
		}
		text := out.Text()
		if !strings.Contains(text, "claimed_by: cc-session-abc123\n") {
			t.Errorf("Text() missing 'claimed_by: cc-session-abc123'\nGot: %s", text)
		}
	})

	t.Run("claimed_by empty is absent", func(t *testing.T) {
		out := ContextOutput{
			SessionID:     "session-20260306-000007-aaaabbbb",
			Status:        "ACTIVE",
			ExecutionMode: "orchestrated",
			HasSession:    true,
			ClaimedBy:     "",
		}
		text := out.Text()
		if strings.Contains(text, "claimed_by:") {
			t.Errorf("Text() should not contain 'claimed_by:' when empty\nGot: %s", text)
		}
	})
}

func TestConvertStrands(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := convertStrands(nil)
		if result != nil {
			t.Errorf("convertStrands(nil) = %v, want nil", result)
		}
	})

	t.Run("empty slice returns nil", func(t *testing.T) {
		result := convertStrands([]session.Strand{})
		if result != nil {
			t.Errorf("convertStrands([]) = %v, want nil", result)
		}
	})

	t.Run("populated slice converts correctly", func(t *testing.T) {
		input := []session.Strand{
			{
				SessionID: "s1",
				Status:    "ACTIVE",
				FrameRef:  "f1",
				LandedAt:  "t1",
			},
		}
		result := convertStrands(input)
		if len(result) != 1 {
			t.Fatalf("convertStrands() len = %d, want 1", len(result))
		}
		if result[0].SessionID != "s1" {
			t.Errorf("SessionID = %q, want %q", result[0].SessionID, "s1")
		}
		if result[0].Status != "ACTIVE" {
			t.Errorf("Status = %q, want %q", result[0].Status, "ACTIVE")
		}
		if result[0].FrameRef != "f1" {
			t.Errorf("FrameRef = %q, want %q", result[0].FrameRef, "f1")
		}
		if result[0].LandedAt != "t1" {
			t.Errorf("LandedAt = %q, want %q", result[0].LandedAt, "t1")
		}
	})
}

func TestRunContext_WithActiveSession_IncludesComplexity(t *testing.T) {
	// TestRunContext_WithActiveSession already uses complexity:"MODULE" in its fixture.
	// This test verifies that Complexity is populated end-to-end after S2 wiring.
	tmpDir := t.TempDir()
	sessionID := "session-20260104-222613-05a12c6b"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260104-222613-05a12c6b"
status: ACTIVE
created_at: "2026-01-04T22:26:13Z"
initiative: "Hooks Migration"
complexity: "MODULE"
active_rite: "10x-dev"
current_phase: "implementation"
---

# Session: Hooks Migration
`
	os.WriteFile(filepath.Join(sessionsDir, sessionID, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("10x-dev"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
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

	if err := runContextCore(nil, ctx, printer); err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.Complexity != "MODULE" {
		t.Errorf("Complexity = %q, want %q", result.Complexity, "MODULE")
	}
}

func TestRunContext_WithStrands(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260306-100000-stranded0"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.3"
session_id: "session-20260306-100000-stranded0"
status: ACTIVE
created_at: "2026-03-06T10:00:00Z"
initiative: "Strand Test"
complexity: "SYSTEM"
active_rite: "ecosystem"
current_phase: "implementation"
strands:
  - session_id: session-20260306-110000-child001
    status: ACTIVE
  - session_id: session-20260306-120000-child002
    status: LANDED
    frame_ref: ".sos/wip/frames/child2.md"
    landed_at: "2026-03-06T14:00:00Z"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("ecosystem"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
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

	if err := runContextCore(nil, ctx, printer); err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if len(result.Strands) != 2 {
		t.Fatalf("Strands length = %d, want 2: %v", len(result.Strands), result.Strands)
	}
	if result.Strands[0].SessionID != "session-20260306-110000-child001" {
		t.Errorf("Strands[0].SessionID = %q, want %q", result.Strands[0].SessionID, "session-20260306-110000-child001")
	}
	if result.Strands[0].Status != "ACTIVE" {
		t.Errorf("Strands[0].Status = %q, want %q", result.Strands[0].Status, "ACTIVE")
	}
	if result.Strands[1].Status != "LANDED" {
		t.Errorf("Strands[1].Status = %q, want %q", result.Strands[1].Status, "LANDED")
	}
	if result.Strands[1].FrameRef != ".sos/wip/frames/child2.md" {
		t.Errorf("Strands[1].FrameRef = %q, want %q", result.Strands[1].FrameRef, ".sos/wip/frames/child2.md")
	}
	if result.Strands[1].LandedAt != "2026-03-06T14:00:00Z" {
		t.Errorf("Strands[1].LandedAt = %q, want %q", result.Strands[1].LandedAt, "2026-03-06T14:00:00Z")
	}
}

func TestRunContext_WithClaimedBy(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "session-20260306-143052-claimedc"

	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	sessionDir := filepath.Join(sessionsDir, sessionID)
	os.MkdirAll(sessionDir, 0755)

	sessionContext := `---
schema_version: "2.3"
session_id: "session-20260306-143052-claimedc"
status: ACTIVE
created_at: "2026-03-06T14:30:52Z"
initiative: "Fix login bug"
complexity: "PATCH"
active_rite: "10x-dev"
current_phase: "implementation"
claimed_by: "cc-session-xyz"
---
`
	os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644)
	os.MkdirAll(filepath.Join(tmpDir, ".knossos"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".knossos", "ACTIVE_RITE"), []byte("10x-dev"), 0644)

	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: tmpDir,
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

	if err := runContextCore(nil, ctx, printer); err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	if result.ClaimedBy != "cc-session-xyz" {
		t.Errorf("ClaimedBy = %q, want %q", result.ClaimedBy, "cc-session-xyz")
	}
}

