package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/session"
)

func TestFray_NoActiveSession(t *testing.T) {
	// Setup: empty project dir with no sessions
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	os.MkdirAll(sessionsDir, 0755)

	// Call fraySession with empty session ID - should return error
	_, err := fraySession(tmpDir, "", frayOptions{noWorktree: true})
	if err == nil {
		t.Fatal("expected error when no active session, got nil")
	}
}

func TestFray_ParkParent(t *testing.T) {
	// Setup: create a real session on disk
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	locksDir := filepath.Join(tmpDir, ".claude", "locks")
	os.MkdirAll(sessionsDir, 0755)
	os.MkdirAll(locksDir, 0755)

	// Create parent session
	parent := session.NewContext("Test Initiative", "MODULE", "10x-dev")
	parentDir := filepath.Join(sessionsDir, parent.SessionID)
	os.MkdirAll(parentDir, 0755)
	parentCtxPath := filepath.Join(parentDir, "SESSION_CONTEXT.md")
	if err := parent.Save(parentCtxPath); err != nil {
		t.Fatalf("failed to save parent: %v", err)
	}

	// Write current session cache
	os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(parent.SessionID), 0644)

	// Fray
	result, err := fraySession(tmpDir, parent.SessionID, frayOptions{noWorktree: true})
	if err != nil {
		t.Fatalf("fraySession() error = %v", err)
	}

	// Verify parent is PARKED
	reloadedParent, err := session.LoadContext(parentCtxPath)
	if err != nil {
		t.Fatalf("failed to reload parent: %v", err)
	}
	if reloadedParent.Status != session.StatusParked {
		t.Errorf("parent status = %v, want PARKED", reloadedParent.Status)
	}

	// Verify parent has strand
	if len(reloadedParent.Strands) != 1 {
		t.Fatalf("parent strands = %d, want 1", len(reloadedParent.Strands))
	}
	if reloadedParent.Strands[0] != result.ChildID {
		t.Errorf("parent strand = %q, want %q", reloadedParent.Strands[0], result.ChildID)
	}
}

func TestFray_CreateChild(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	locksDir := filepath.Join(tmpDir, ".claude", "locks")
	os.MkdirAll(sessionsDir, 0755)
	os.MkdirAll(locksDir, 0755)

	parent := session.NewContext("Test Initiative", "MODULE", "10x-dev")
	parent.CurrentPhase = "design"
	parentDir := filepath.Join(sessionsDir, parent.SessionID)
	os.MkdirAll(parentDir, 0755)
	if err := parent.Save(filepath.Join(parentDir, "SESSION_CONTEXT.md")); err != nil {
		t.Fatalf("failed to save parent: %v", err)
	}
	os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(parent.SessionID), 0644)

	// Fray
	result, err := fraySession(tmpDir, parent.SessionID, frayOptions{noWorktree: true})
	if err != nil {
		t.Fatalf("fraySession() error = %v", err)
	}

	// Load child
	childCtxPath := filepath.Join(sessionsDir, result.ChildID, "SESSION_CONTEXT.md")
	child, err := session.LoadContext(childCtxPath)
	if err != nil {
		t.Fatalf("failed to load child: %v", err)
	}

	// Verify child fields
	if child.Status != session.StatusActive {
		t.Errorf("child status = %v, want ACTIVE", child.Status)
	}
	if child.FrayedFrom != parent.SessionID {
		t.Errorf("child FrayedFrom = %q, want %q", child.FrayedFrom, parent.SessionID)
	}
	if child.FrayPoint != "design" {
		t.Errorf("child FrayPoint = %q, want %q", child.FrayPoint, "design")
	}
	if child.SchemaVersion != "2.2" {
		t.Errorf("child SchemaVersion = %q, want %q", child.SchemaVersion, "2.2")
	}
	if child.Initiative != parent.Initiative {
		t.Errorf("child Initiative = %q, want %q", child.Initiative, parent.Initiative)
	}
	if child.Complexity != parent.Complexity {
		t.Errorf("child Complexity = %q, want %q", child.Complexity, parent.Complexity)
	}
	if child.ActiveRite != parent.ActiveRite {
		t.Errorf("child ActiveRite = %q, want %q", child.ActiveRite, parent.ActiveRite)
	}
}

func TestFray_NoWorktreeFlag(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, ".claude", "sessions")
	locksDir := filepath.Join(tmpDir, ".claude", "locks")
	os.MkdirAll(sessionsDir, 0755)
	os.MkdirAll(locksDir, 0755)

	parent := session.NewContext("Test", "PATCH", "none")
	parentDir := filepath.Join(sessionsDir, parent.SessionID)
	os.MkdirAll(parentDir, 0755)
	if err := parent.Save(filepath.Join(parentDir, "SESSION_CONTEXT.md")); err != nil {
		t.Fatalf("failed to save parent: %v", err)
	}
	os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte(parent.SessionID), 0644)

	// Fray with --no-worktree
	result, err := fraySession(tmpDir, parent.SessionID, frayOptions{noWorktree: true})
	if err != nil {
		t.Fatalf("fraySession() error = %v", err)
	}

	// Verify no worktree path in result
	if result.WorktreePath != "" {
		t.Errorf("WorktreePath should be empty with --no-worktree, got %q", result.WorktreePath)
	}
}
