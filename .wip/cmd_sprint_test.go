package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// setupSprintTest creates a temporary project with an active session.
func setupSprintTest(t *testing.T) (*cmdContext, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectDir := tmpDir

	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(claudeDir, "ACTIVE_RITE"), []byte("ecosystem"), 0644); err != nil {
		t.Fatalf("Failed to write ACTIVE_RITE: %v", err)
	}

	sessionsDir := filepath.Join(claudeDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions dir: %v", err)
	}

	sessCtx := session.NewContext("Test Initiative", "MODULE", "ecosystem")
	sessionDir := filepath.Join(sessionsDir, sessCtx.SessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("Failed to create session dir: %v", err)
	}

	resolver := paths.NewResolver(projectDir)
	if err := sessCtx.Save(resolver.SessionContextFile(sessCtx.SessionID)); err != nil {
		t.Fatalf("Failed to save session context: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessCtx.SessionID), 0644); err != nil {
		t.Fatalf("Failed to write current session: %v", err)
	}

	if err := os.MkdirAll(resolver.AuditDir(), 0755); err != nil {
		t.Fatalf("Failed to create audit dir: %v", err)
	}

	outputFormat := "json"
	verbose := false
	sessionID := sessCtx.SessionID
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
			SessionID: &sessionID,
		},
	}

	return ctx, projectDir, sessCtx.SessionID
}

func TestSprintCreate_HappyPath(t *testing.T) {
	ctx, projectDir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(projectDir)

	err := runSprintCreate(ctx, "Test Sprint Goal", []string{"Task A", "Task B"})
	if err != nil {
		t.Fatalf("runSprintCreate() error = %v", err)
	}

	sprintsDir := resolver.SprintsDir(sessionID)
	entries, err := os.ReadDir(sprintsDir)
	if err != nil {
		t.Fatalf("Failed to read sprints dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 sprint directory, got %d", len(entries))
	}

	sprintID := entries[0].Name()
	if !session.IsValidSprintID(sprintID) {
		t.Errorf("Sprint directory name %q is not a valid sprint ID", sprintID)
	}

	sprintCtx, err := session.LoadSprintContext(resolver.SprintContextFile(sessionID, sprintID))
	if err != nil {
		t.Fatalf("LoadSprintContext() error = %v", err)
	}

	if sprintCtx.Goal != "Test Sprint Goal" {
		t.Errorf("Goal = %q, want %q", sprintCtx.Goal, "Test Sprint Goal")
	}
	if sprintCtx.Status != session.SprintStatusActive {
		t.Errorf("Status = %q, want %q", sprintCtx.Status, session.SprintStatusActive)
	}
	if sprintCtx.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", sprintCtx.SessionID, sessionID)
	}
	if len(sprintCtx.Tasks) != 2 {
		t.Fatalf("Tasks length = %d, want 2", len(sprintCtx.Tasks))
	}
}

func TestSprintCreate_DuplicateActiveSprintError(t *testing.T) {
	ctx, _, _ := setupSprintTest(t)

	if err := runSprintCreate(ctx, "First Sprint", nil); err != nil {
		t.Fatalf("First runSprintCreate() error = %v", err)
	}

	err := runSprintCreate(ctx, "Second Sprint", nil)
	if err == nil {
		t.Error("runSprintCreate() should fail when active sprint exists")
	}
}

func TestSprintCreate_NoSessionError(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := tmpDir

	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "sessions"), 0755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}

	outputFormat := "json"
	verbose := false
	emptySession := ""
	ctx := &cmdContext{
		SessionContext: common.SessionContext{
			BaseContext: common.BaseContext{
				Output:     &outputFormat,
				Verbose:    &verbose,
				ProjectDir: &projectDir,
			},
			SessionID: &emptySession,
		},
	}

	err := runSprintCreate(ctx, "No Session Sprint", nil)
	if err == nil {
		t.Error("runSprintCreate() should error when no session exists")
	}
}

func TestSprintMarkComplete_HappyPath(t *testing.T) {
	ctx, projectDir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(projectDir)

	if err := runSprintCreate(ctx, "Complete Me", nil); err != nil {
		t.Fatalf("runSprintCreate() error = %v", err)
	}

	entries, _ := os.ReadDir(resolver.SprintsDir(sessionID))
	sprintID := entries[0].Name()

	if err := runSprintMarkComplete(ctx, sprintID); err != nil {
		t.Fatalf("runSprintMarkComplete() error = %v", err)
	}

	sprintCtx, err := session.LoadSprintContext(resolver.SprintContextFile(sessionID, sprintID))
	if err != nil {
		t.Fatalf("LoadSprintContext() error = %v", err)
	}
	if sprintCtx.Status != session.SprintStatusCompleted {
		t.Errorf("Status = %q, want %q", sprintCtx.Status, session.SprintStatusCompleted)
	}
	if sprintCtx.CompletedAt == nil {
		t.Error("CompletedAt should not be nil after completion")
	}
}

func TestSprintMarkComplete_AutoResolve(t *testing.T) {
	ctx, _, _ := setupSprintTest(t)

	if err := runSprintCreate(ctx, "Auto Resolve", nil); err != nil {
		t.Fatalf("runSprintCreate() error = %v", err)
	}

	if err := runSprintMarkComplete(ctx, ""); err != nil {
		t.Fatalf("runSprintMarkComplete() with auto-resolve error = %v", err)
	}
}

func TestSprintMarkComplete_AlreadyCompleted(t *testing.T) {
	ctx, projectDir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(projectDir)

	if err := runSprintCreate(ctx, "Already Done", nil); err != nil {
		t.Fatalf("runSprintCreate() error = %v", err)
	}
	entries, _ := os.ReadDir(resolver.SprintsDir(sessionID))
	sprintID := entries[0].Name()

	if err := runSprintMarkComplete(ctx, sprintID); err != nil {
		t.Fatalf("First runSprintMarkComplete() error = %v", err)
	}

	err := runSprintMarkComplete(ctx, sprintID)
	if err == nil {
		t.Error("runSprintMarkComplete() should error on already completed sprint")
	}
}

func TestSprintDelete_HappyPath(t *testing.T) {
	ctx, projectDir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(projectDir)

	if err := runSprintCreate(ctx, "Delete Me", nil); err != nil {
		t.Fatalf("runSprintCreate() error = %v", err)
	}
	entries, _ := os.ReadDir(resolver.SprintsDir(sessionID))
	sprintID := entries[0].Name()

	if err := runSprintMarkComplete(ctx, sprintID); err != nil {
		t.Fatalf("runSprintMarkComplete() error = %v", err)
	}

	if err := runSprintDelete(ctx, sprintID); err != nil {
		t.Fatalf("runSprintDelete() error = %v", err)
	}

	sprintDir := resolver.SprintDir(sessionID, sprintID)
	if _, err := os.Stat(sprintDir); !os.IsNotExist(err) {
		t.Error("Sprint directory should not exist after deletion")
	}
}

func TestSprintDelete_ActiveSprintError(t *testing.T) {
	ctx, projectDir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(projectDir)

	if err := runSprintCreate(ctx, "Active Sprint", nil); err != nil {
		t.Fatalf("runSprintCreate() error = %v", err)
	}
	entries, _ := os.ReadDir(resolver.SprintsDir(sessionID))
	sprintID := entries[0].Name()

	err := runSprintDelete(ctx, sprintID)
	if err == nil {
		t.Error("runSprintDelete() should error on active sprint")
	}
}
