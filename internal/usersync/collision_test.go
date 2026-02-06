package usersync

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCollisionChecker_CheckCollision tests collision detection across rites.
func TestCollisionChecker_CheckCollision(t *testing.T) {
	tmpDir := t.TempDir()

	// Create rite directory structure with mena/ subdirs
	for _, rite := range []string{"10x-dev", "forge"} {
		menaDir := filepath.Join(tmpDir, "rites", rite, "mena")
		if err := os.MkdirAll(menaDir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Add a command to 10x-dev rite
	cmdDir := filepath.Join(tmpDir, "rites", "10x-dev", "mena", "build-ref")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "INDEX.lego.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	checker := &CollisionChecker{
		knossosHome:  tmpDir,
		ritesDir:     filepath.Join(tmpDir, "rites"),
		resourceType: ResourceCommands,
		nested:       true,
	}

	// Should find collision for existing resource
	hasCollision, riteName := checker.CheckCollision("build-ref")
	if !hasCollision {
		t.Error("expected collision for build-ref, got none")
	}
	if riteName != "10x-dev" {
		t.Errorf("expected rite 10x-dev, got %s", riteName)
	}

	// Should not find collision for non-existent resource
	hasCollision, _ = checker.CheckCollision("nonexistent")
	if hasCollision {
		t.Error("expected no collision for nonexistent, got collision")
	}
}

// TestCollisionChecker_EmptyKnossosHome returns false when knossosHome is empty.
func TestCollisionChecker_EmptyKnossosHome(t *testing.T) {
	checker := &CollisionChecker{
		knossosHome:  "",
		ritesDir:     "",
		resourceType: ResourceCommands,
		nested:       true,
	}

	hasCollision, _ := checker.CheckCollision("anything")
	if hasCollision {
		t.Error("expected no collision with empty knossosHome")
	}
}

// TestCollisionChecker_NoRitesDir returns false when rites dir doesn't exist.
func TestCollisionChecker_NoRitesDir(t *testing.T) {
	tmpDir := t.TempDir()
	checker := &CollisionChecker{
		knossosHome:  tmpDir,
		ritesDir:     filepath.Join(tmpDir, "nonexistent-rites"),
		resourceType: ResourceCommands,
		nested:       true,
	}

	hasCollision, _ := checker.CheckCollision("anything")
	if hasCollision {
		t.Error("expected no collision with missing rites dir")
	}
}

// TestCollisionChecker_AgentsFlat tests flat (non-nested) collision for agents.
func TestCollisionChecker_AgentsFlat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create agent in forge rite
	agentDir := filepath.Join(tmpDir, "rites", "forge", "agents")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "my-agent.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	checker := &CollisionChecker{
		knossosHome:  tmpDir,
		ritesDir:     filepath.Join(tmpDir, "rites"),
		resourceType: ResourceAgents,
		nested:       false,
	}

	hasCollision, riteName := checker.CheckCollision("my-agent.md")
	if !hasCollision {
		t.Error("expected collision for my-agent.md")
	}
	if riteName != "forge" {
		t.Errorf("expected rite forge, got %s", riteName)
	}
}

// TestCollisionChecker_MultipleRites tests collision across multiple rites.
func TestCollisionChecker_MultipleRites(t *testing.T) {
	tmpDir := t.TempDir()

	// Create same-named resource in two rites
	for _, rite := range []string{"10x-dev", "forge"} {
		cmdDir := filepath.Join(tmpDir, "rites", rite, "mena", "shared-cmd")
		if err := os.MkdirAll(cmdDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cmdDir, "INDEX.dro.md"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	checker := &CollisionChecker{
		knossosHome:  tmpDir,
		ritesDir:     filepath.Join(tmpDir, "rites"),
		resourceType: ResourceCommands,
		nested:       true,
	}

	// Should find collision (returns first match)
	hasCollision, riteName := checker.CheckCollision("shared-cmd")
	if !hasCollision {
		t.Error("expected collision for shared-cmd")
	}
	// Should be one of the two rites
	if riteName != "10x-dev" && riteName != "forge" {
		t.Errorf("expected 10x-dev or forge, got %s", riteName)
	}
}

// TestCollisionChecker_UsesRiteSubDir verifies mena/ is used for commands, not commands/.
func TestCollisionChecker_UsesRiteSubDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create resource under OLD path (commands/) — should NOT be found
	oldPath := filepath.Join(tmpDir, "rites", "test-rite", "commands", "my-cmd")
	if err := os.MkdirAll(oldPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create resource under NEW path (mena/) — should be found
	newPath := filepath.Join(tmpDir, "rites", "test-rite", "mena", "other-cmd")
	if err := os.MkdirAll(newPath, 0755); err != nil {
		t.Fatal(err)
	}

	checker := &CollisionChecker{
		knossosHome:  tmpDir,
		ritesDir:     filepath.Join(tmpDir, "rites"),
		resourceType: ResourceCommands,
		nested:       true,
	}

	// Old path should NOT be found (proves we use mena/, not commands/)
	hasCollision, _ := checker.CheckCollision("my-cmd")
	if hasCollision {
		t.Error("should NOT find resource under old commands/ path")
	}

	// New path should be found
	hasCollision, riteName := checker.CheckCollision("other-cmd")
	if !hasCollision {
		t.Error("should find resource under new mena/ path")
	}
	if riteName != "test-rite" {
		t.Errorf("expected test-rite, got %s", riteName)
	}
}
