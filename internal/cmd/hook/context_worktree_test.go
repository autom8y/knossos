package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/test/hooks/testutil"
	worktreefixture "github.com/autom8y/knossos/test/worktree/testutil"
)

// TestContextHook_InWorktree_ReportsRites is a P1 test validating Fix 3 (W1c):
// the context hook uses SourceResolver.ListAvailableRites() instead of the old
// listAvailableRites(resolver.RitesDir()) so rites are discoverable even when
// .knossos/rites/ is empty or absent (the common worktree case).
//
// This test runs the context hook from a linked worktree directory. It sets
// KNOSSOS_HOME to a temp dir that contains a test rite, verifying that
// AvailableRites is non-empty despite the worktree having no local rites.
func TestContextHook_InWorktree_ReportsRites(t *testing.T) {
	// Isolate KNOSSOS_HOME: point it at a temp dir containing one test rite.
	knossosHome := t.TempDir()
	t.Setenv("KNOSSOS_HOME", knossosHome)
	config.ResetKnossosHome()
	t.Cleanup(config.ResetKnossosHome)

	// Also isolate HOME to avoid user rites from ~/.claude/rites/ contaminating
	// the count assertion.
	t.Setenv("HOME", knossosHome)

	// Create a rite in KNOSSOS_HOME/rites/ so SourceResolver can discover it.
	kRiteDir := filepath.Join(knossosHome, "rites", "platform-rite")
	if err := os.MkdirAll(kRiteDir, 0755); err != nil {
		t.Fatalf("failed to create KNOSSOS_HOME rite dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(kRiteDir, "manifest.yaml"), []byte("name: platform-rite\n"), 0644); err != nil {
		t.Fatalf("failed to write rite manifest: %v", err)
	}

	// Set up a linked worktree fixture. The worktree has no .knossos/rites/.
	fix := worktreefixture.SetupWorktreeTestFixture(t)
	worktreeDir := fix.WorktreeDir

	// Create a session in the worktree so the hook produces meaningful output
	// (runContextCore returns early if no session is found).
	sessionID := "session-20260302-010000-worktree1"
	sessionDir := filepath.Join(worktreeDir, ".sos", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	sessionContext := `---
schema_version: "2.1"
session_id: "session-20260302-010000-worktree1"
status: ACTIVE
created_at: "2026-03-02T01:00:00Z"
initiative: "Worktree Rites Test"
complexity: "simple"
active_rite: "test-rite"
current_phase: "implementation"
---
`
	if err := os.WriteFile(filepath.Join(sessionDir, "SESSION_CONTEXT.md"), []byte(sessionContext), 0644); err != nil {
		t.Fatalf("failed to write session context: %v", err)
	}

	// Create .claude/ in worktree (ensureProjectDirs creates it during sync,
	// but the hook needs it for ACTIVE_RITE resolution).
	worktreeClaudeDir := filepath.Join(worktreeDir, ".claude")
	if err := os.MkdirAll(worktreeClaudeDir, 0755); err != nil {
		t.Fatalf("failed to create worktree .claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreeClaudeDir, "ACTIVE_RITE"), []byte("test-rite\n"), 0644); err != nil {
		t.Fatalf("failed to write worktree ACTIVE_RITE: %v", err)
	}

	// Set up hook environment pointing at the worktree.
	testutil.SetupEnv(t, &testutil.HookEnv{
		Event:      "SessionStart",
		ProjectDir: worktreeDir,
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
				ProjectDir: &worktreeDir,
			},
			SessionID: nil,
		},
	}

	if err := runContextCore(ctx, printer); err != nil {
		t.Fatalf("runContextCore() error = %v", err)
	}

	var result ContextOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse output: %v\nOutput: %s", err, stdout.String())
	}

	// The hook must report a non-empty AvailableRites list.
	// The SourceResolver picks up "platform-rite" from KNOSSOS_HOME/rites/.
	// This validates Fix 3 (W1c): ListAvailableRites() checks the 4-tier chain,
	// not just .knossos/rites/ (which is absent in the worktree).
	if len(result.AvailableRites) == 0 {
		t.Error("AvailableRites should be non-empty in worktree: SourceResolver must discover rites from KNOSSOS_HOME even when .knossos/rites/ is absent")
	}

	found := false
	for _, r := range result.AvailableRites {
		if r == "platform-rite" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AvailableRites should contain 'platform-rite' from KNOSSOS_HOME: got %v", result.AvailableRites)
	}

	// Session must also be resolved correctly.
	if !result.HasSession {
		t.Error("HasSession should be true — session was created in worktree")
	}
}
