package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

func setupSprintTest(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()

	// Create .claude/sessions directory structure
	claudeDir := filepath.Join(dir, ".claude")
	sessionsDir := filepath.Join(claudeDir, "sessions")
	auditDir := filepath.Join(sessionsDir, ".audit")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a session
	sessionID := "session-20260104-160414-563c681e"
	sessionDir := filepath.Join(sessionsDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write current session
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(sessionID), 0644); err != nil {
		t.Fatal(err)
	}

	// Write active session context
	ctx := &session.Context{
		SchemaVersion: "2.1",
		SessionID:     sessionID,
		Status:        session.StatusActive,
		Initiative:    "Test",
		Complexity:    "MODULE",
		ActiveRite:    "test",
		CurrentPhase:  "design",
	}
	// Use current time for CreatedAt
	ctxPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	if err := ctx.Save(ctxPath); err != nil {
		t.Fatal(err)
	}

	return dir, sessionID
}

func TestSprintCreate_HappyPath(t *testing.T) {
	dir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(dir)

	// Create sprint
	sprintCtx := session.NewSprintContext(sessionID, "Test Sprint", []string{"Task A"})
	sprintDir := resolver.SprintDir(sessionID, sprintCtx.SprintID)
	if err := paths.EnsureDir(sprintDir); err != nil {
		t.Fatal(err)
	}
	ctxPath := resolver.SprintContextFile(sessionID, sprintCtx.SprintID)
	if err := sprintCtx.Save(ctxPath); err != nil {
		t.Fatal(err)
	}

	// Verify sprint was created
	loaded, err := session.LoadSprintContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadSprintContext() error = %v", err)
	}
	if loaded.Goal != "Test Sprint" {
		t.Errorf("Goal = %q, want %q", loaded.Goal, "Test Sprint")
	}
	if loaded.Status != session.SprintStatusActive {
		t.Errorf("Status = %v, want %v", loaded.Status, session.SprintStatusActive)
	}
}

func TestSprintMarkComplete_HappyPath(t *testing.T) {
	dir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(dir)

	// Create sprint first
	sprintCtx := session.NewSprintContext(sessionID, "Test Sprint", nil)
	sprintDir := resolver.SprintDir(sessionID, sprintCtx.SprintID)
	if err := paths.EnsureDir(sprintDir); err != nil {
		t.Fatal(err)
	}
	ctxPath := resolver.SprintContextFile(sessionID, sprintCtx.SprintID)
	if err := sprintCtx.Save(ctxPath); err != nil {
		t.Fatal(err)
	}

	// Mark complete
	sprintCtx.Status = session.SprintStatusCompleted
	if err := sprintCtx.Save(ctxPath); err != nil {
		t.Fatal(err)
	}

	loaded, err := session.LoadSprintContext(ctxPath)
	if err != nil {
		t.Fatalf("LoadSprintContext() error = %v", err)
	}
	if loaded.Status != session.SprintStatusCompleted {
		t.Errorf("Status = %v, want %v", loaded.Status, session.SprintStatusCompleted)
	}
}

func TestSprintDelete_HappyPath(t *testing.T) {
	dir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(dir)

	// Create completed sprint
	sprintCtx := session.NewSprintContext(sessionID, "Test Sprint", nil)
	sprintCtx.Status = session.SprintStatusCompleted
	sprintDir := resolver.SprintDir(sessionID, sprintCtx.SprintID)
	if err := paths.EnsureDir(sprintDir); err != nil {
		t.Fatal(err)
	}
	ctxPath := resolver.SprintContextFile(sessionID, sprintCtx.SprintID)
	if err := sprintCtx.Save(ctxPath); err != nil {
		t.Fatal(err)
	}

	// Delete
	if err := os.RemoveAll(sprintDir); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}

	// Verify deleted
	if _, err := os.Stat(sprintDir); !os.IsNotExist(err) {
		t.Error("Sprint directory should be deleted")
	}
}

func TestFindSprintByStatus(t *testing.T) {
	dir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(dir)

	// Create active sprint
	sprintCtx := session.NewSprintContext(sessionID, "Active Sprint", nil)
	sprintDir := resolver.SprintDir(sessionID, sprintCtx.SprintID)
	if err := paths.EnsureDir(sprintDir); err != nil {
		t.Fatal(err)
	}
	if err := sprintCtx.Save(resolver.SprintContextFile(sessionID, sprintCtx.SprintID)); err != nil {
		t.Fatal(err)
	}

	// Find active
	found, err := findSprintByStatus(resolver, sessionID, session.SprintStatusActive)
	if err != nil {
		t.Fatalf("findSprintByStatus() error = %v", err)
	}
	if found != sprintCtx.SprintID {
		t.Errorf("found = %q, want %q", found, sprintCtx.SprintID)
	}

	// No completed sprint
	found, err = findSprintByStatus(resolver, sessionID, session.SprintStatusCompleted)
	if err != nil {
		t.Fatalf("findSprintByStatus() error = %v", err)
	}
	if found != "" {
		t.Errorf("found = %q, want empty", found)
	}
}

func TestFindSprintByStatus_NoSprintsDir(t *testing.T) {
	dir, sessionID := setupSprintTest(t)
	resolver := paths.NewResolver(dir)

	found, err := findSprintByStatus(resolver, sessionID, session.SprintStatusActive)
	if err != nil {
		t.Fatalf("findSprintByStatus() error = %v", err)
	}
	if found != "" {
		t.Errorf("found = %q, want empty", found)
	}
}
