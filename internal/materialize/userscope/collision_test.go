package userscope

import (
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/provenance"
)

func TestCollisionChecker_ExactMatch(t *testing.T) {
	c := &CollisionChecker{
		manifestLoaded: true,
		riteEntries: map[string]bool{
			"agents/consultant.md": true,
			"commands/go/":         true,
			"skills/guidance/":     true,
		},
	}

	collision, reason := c.CheckCollision("agents/consultant.md")
	if !collision {
		t.Error("Expected exact match collision for agents/consultant.md")
	}
	if reason != "(from manifest)" {
		t.Errorf("Expected reason '(from manifest)', got %q", reason)
	}
}

func TestCollisionChecker_PrefixContainment(t *testing.T) {
	c := &CollisionChecker{
		manifestLoaded: true,
		riteEntries: map[string]bool{
			"skills/guidance/standards/": true,
			"commands/go/":              true,
		},
	}

	// User file inside rite-owned directory should collide
	collision, reason := c.CheckCollision("skills/guidance/standards/code-conventions.md")
	if !collision {
		t.Error("Expected prefix containment collision for file inside rite directory")
	}
	if reason != "(inside rite directory)" {
		t.Errorf("Expected reason '(inside rite directory)', got %q", reason)
	}
}

func TestCollisionChecker_NoFalsePositive(t *testing.T) {
	c := &CollisionChecker{
		manifestLoaded: true,
		riteEntries: map[string]bool{
			"commands/go/": true,
		},
	}

	// Different directory: commands/navigation/go.md should NOT collide with commands/go/
	collision, _ := c.CheckCollision("commands/navigation/go.md")
	if collision {
		t.Error("Expected no collision for commands/navigation/go.md with commands/go/")
	}

	// File that starts with same prefix but is a different path
	collision, _ = c.CheckCollision("commands/go-elsewhere/INDEX.md")
	if collision {
		t.Error("Expected no collision for commands/go-elsewhere/ with commands/go/")
	}
}

func TestCollisionChecker_AgentExactMatch(t *testing.T) {
	c := &CollisionChecker{
		manifestLoaded: true,
		riteEntries: map[string]bool{
			"agents/consultant.md":      true,
			"agents/context-engineer.md": true,
		},
	}

	collision, _ := c.CheckCollision("agents/consultant.md")
	if !collision {
		t.Error("Expected exact match collision for agents/consultant.md")
	}

	collision, _ = c.CheckCollision("agents/context-engineer.md")
	if !collision {
		t.Error("Expected exact match collision for agents/context-engineer.md")
	}

	// Different agent should not collide
	collision, _ = c.CheckCollision("agents/my-custom-agent.md")
	if collision {
		t.Error("Expected no collision for agents/my-custom-agent.md")
	}
}

func TestCollisionChecker_NoManifest(t *testing.T) {
	c := &CollisionChecker{
		manifestLoaded: false,
	}

	collision, _ := c.CheckCollision("agents/anything.md")
	if collision {
		t.Error("Expected no collision when manifest not loaded")
	}
}

func TestCollisionChecker_EmptyEntries(t *testing.T) {
	c := &CollisionChecker{
		manifestLoaded: true,
		riteEntries:    map[string]bool{},
	}

	collision, _ := c.CheckCollision("agents/anything.md")
	if collision {
		t.Error("Expected no collision with empty rite entries")
	}
}

// TestCollisionChecker_MissingManifest_ReportsIneffective verifies that when
// no PROVENANCE_MANIFEST.yaml exists (e.g. a fresh worktree), IsEffective()
// returns false so callers can fail-closed and skip user-scope writes.
func TestCollisionChecker_MissingManifest_ReportsIneffective(t *testing.T) {
	tmpDir := t.TempDir()
	// No manifest file written — .claude/ is empty.
	c := NewCollisionChecker(tmpDir)

	if c.IsEffective() {
		t.Error("IsEffective() should return false when no provenance manifest exists")
	}
	// CheckCollision should return false (no collision) because riteEntries is empty,
	// but the caller should not rely on this path when checker is ineffective.
	collision, _ := c.CheckCollision("agents/anything.md")
	if collision {
		t.Error("CheckCollision should return false when manifest is missing (empty entries)")
	}
}

// TestCollisionChecker_EmptyEntries_ReportsIneffective verifies that an empty
// manifest (no rite-scope entries) still reports IsEffective()=true — the
// manifest loaded successfully, there are simply no rite-owned resources.
// This is distinct from a missing manifest.
func TestCollisionChecker_EmptyEntries_ReportsIneffective(t *testing.T) {
	// Write a valid PROVENANCE_MANIFEST.yaml with no entries using provenance.Save().
	tmpDir := t.TempDir()
	manifest := &provenance.ProvenanceManifest{
		SchemaVersion: provenance.CurrentSchemaVersion,
		LastSync:      time.Now().UTC(),
		Entries:       make(map[string]*provenance.ProvenanceEntry),
	}
	manifestPath := provenance.ManifestPath(tmpDir)
	if err := provenance.Save(manifestPath, manifest); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	c := NewCollisionChecker(tmpDir)

	// Manifest loaded successfully (even if empty), so checker IS effective.
	if !c.IsEffective() {
		t.Error("IsEffective() should return true when a valid (but empty) manifest exists")
	}
	// No rite entries means no collisions.
	collision, _ := c.CheckCollision("agents/anything.md")
	if collision {
		t.Error("Expected no collision when manifest has no rite-scope entries")
	}
}
