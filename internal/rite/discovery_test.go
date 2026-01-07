package rite

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestRites(t *testing.T) (string, func()) {
	t.Helper()

	tempDir := t.TempDir()

	// Create project structure
	claudeDir := filepath.Join(tempDir, ".claude")
	ritesDir := filepath.Join(tempDir, "rites")

	for _, dir := range []string{claudeDir, ritesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create a valid rite
	rite1Dir := filepath.Join(ritesDir, "test-rite")
	if err := os.MkdirAll(rite1Dir, 0755); err != nil {
		t.Fatalf("Failed to create rite dir: %v", err)
	}

	rite1Manifest := `
schema_version: "1.0"
name: test-rite
display_name: "Test Rite"
description: "A test rite"
form: practitioner
agents:
  - name: agent1
    file: agents/agent1.md
skills:
  - ref: skill1
    path: skills/skill1/
`
	if err := os.WriteFile(filepath.Join(rite1Dir, "manifest.yaml"), []byte(rite1Manifest), 0644); err != nil {
		t.Fatalf("Failed to write rite manifest: %v", err)
	}

	// Create a simple rite
	rite2Dir := filepath.Join(ritesDir, "simple-rite")
	if err := os.MkdirAll(rite2Dir, 0755); err != nil {
		t.Fatalf("Failed to create rite dir: %v", err)
	}

	rite2Manifest := `
schema_version: "1.0"
name: simple-rite
form: simple
skills:
  - ref: docs
    path: skills/docs/
`
	if err := os.WriteFile(filepath.Join(rite2Dir, "manifest.yaml"), []byte(rite2Manifest), 0644); err != nil {
		t.Fatalf("Failed to write rite manifest: %v", err)
	}

	// Set active rite
	if err := os.WriteFile(filepath.Join(claudeDir, "ACTIVE_RITE"), []byte("test-rite"), 0644); err != nil {
		t.Fatalf("Failed to write active rite: %v", err)
	}

	cleanup := func() {
		// temp dir is automatically cleaned up
	}

	return tempDir, cleanup
}

func TestDiscovery_List(t *testing.T) {
	tempDir, cleanup := setupTestRites(t)
	defer cleanup()

	discovery := NewDiscoveryWithPaths(
		filepath.Join(tempDir, "rites"),
		"", // No user rites
		"test-rite",
	)

	rites, err := discovery.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(rites) != 2 {
		t.Errorf("len(rites) = %d, want 2", len(rites))
	}

	// Should be sorted by name
	if rites[0].Name != "simple-rite" {
		t.Errorf("First rite = %q, want 'simple-rite'", rites[0].Name)
	}
	if rites[1].Name != "test-rite" {
		t.Errorf("Second rite = %q, want 'test-rite'", rites[1].Name)
	}

	// test-rite should be marked active
	for _, r := range rites {
		if r.Name == "test-rite" && !r.Active {
			t.Error("test-rite should be marked active")
		}
		if r.Name == "simple-rite" && r.Active {
			t.Error("simple-rite should not be marked active")
		}
	}
}

func TestDiscovery_ListByForm(t *testing.T) {
	tempDir, cleanup := setupTestRites(t)
	defer cleanup()

	discovery := NewDiscoveryWithPaths(
		filepath.Join(tempDir, "rites"),
		"",
		"",
	)

	// Filter by practitioner form
	practitioners, err := discovery.ListByForm(FormPractitioner)
	if err != nil {
		t.Fatalf("ListByForm failed: %v", err)
	}
	if len(practitioners) != 1 {
		t.Errorf("len(practitioners) = %d, want 1", len(practitioners))
	}
	if len(practitioners) > 0 && practitioners[0].Name != "test-rite" {
		t.Errorf("Practitioner rite = %q, want 'test-rite'", practitioners[0].Name)
	}

	// Filter by simple form
	simple, err := discovery.ListByForm(FormSimple)
	if err != nil {
		t.Fatalf("ListByForm failed: %v", err)
	}
	if len(simple) != 1 {
		t.Errorf("len(simple) = %d, want 1", len(simple))
	}
}

func TestDiscovery_Get(t *testing.T) {
	tempDir, cleanup := setupTestRites(t)
	defer cleanup()

	discovery := NewDiscoveryWithPaths(
		filepath.Join(tempDir, "rites"),
		"",
		"test-rite",
	)

	// Get existing rite
	rite, err := discovery.Get("test-rite")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if rite.Name != "test-rite" {
		t.Errorf("rite.Name = %q, want 'test-rite'", rite.Name)
	}
	if rite.Form != FormPractitioner {
		t.Errorf("rite.Form = %q, want 'practitioner'", rite.Form)
	}
	if !rite.Active {
		t.Error("test-rite should be active")
	}

	// Get nonexistent rite
	_, err = discovery.Get("nonexistent")
	if err == nil {
		t.Error("Get(nonexistent) should return error")
	}
}

func TestDiscovery_GetManifest(t *testing.T) {
	tempDir, cleanup := setupTestRites(t)
	defer cleanup()

	discovery := NewDiscoveryWithPaths(
		filepath.Join(tempDir, "rites"),
		"",
		"",
	)

	manifest, err := discovery.GetManifest("test-rite")
	if err != nil {
		t.Fatalf("GetManifest failed: %v", err)
	}
	if manifest.Name != "test-rite" {
		t.Errorf("manifest.Name = %q, want 'test-rite'", manifest.Name)
	}
	if len(manifest.Agents) != 1 {
		t.Errorf("len(Agents) = %d, want 1", len(manifest.Agents))
	}
}

func TestDiscovery_GetActive(t *testing.T) {
	tempDir, cleanup := setupTestRites(t)
	defer cleanup()

	discovery := NewDiscoveryWithPaths(
		filepath.Join(tempDir, "rites"),
		"",
		"test-rite",
	)

	active, err := discovery.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active.Name != "test-rite" {
		t.Errorf("active.Name = %q, want 'test-rite'", active.Name)
	}

	// No active rite
	noActiveDiscovery := NewDiscoveryWithPaths(
		filepath.Join(tempDir, "rites"),
		"",
		"",
	)
	_, err = noActiveDiscovery.GetActive()
	if err == nil {
		t.Error("GetActive() with no active should return error")
	}
}

func TestDiscovery_Exists(t *testing.T) {
	tempDir, cleanup := setupTestRites(t)
	defer cleanup()

	discovery := NewDiscoveryWithPaths(
		filepath.Join(tempDir, "rites"),
		"",
		"",
	)

	if !discovery.Exists("test-rite") {
		t.Error("Exists('test-rite') = false, want true")
	}
	if discovery.Exists("nonexistent") {
		t.Error("Exists('nonexistent') = true, want false")
	}
}

func TestDiscovery_ActiveRiteName(t *testing.T) {
	discovery := NewDiscoveryWithPaths(
		"/unused",
		"",
		"my-active-rite",
	)

	if got := discovery.ActiveRiteName(); got != "my-active-rite" {
		t.Errorf("ActiveRiteName() = %q, want 'my-active-rite'", got)
	}
}

func TestDiscovery_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()
	ritesDir := filepath.Join(tempDir, "rites")
	if err := os.MkdirAll(ritesDir, 0755); err != nil {
		t.Fatalf("Failed to create rites dir: %v", err)
	}

	discovery := NewDiscoveryWithPaths(ritesDir, "", "")

	rites, err := discovery.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(rites) != 0 {
		t.Errorf("len(rites) = %d, want 0", len(rites))
	}
}

func TestDiscovery_InvalidRiteSkipped(t *testing.T) {
	tempDir := t.TempDir()
	ritesDir := filepath.Join(tempDir, "rites")

	// Create a valid rite
	validDir := filepath.Join(ritesDir, "valid-rite")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	validManifest := `schema_version: "1.0"
name: valid-rite
form: simple
skills:
  - ref: s1
    path: skills/s1/
`
	if err := os.WriteFile(filepath.Join(validDir, "manifest.yaml"), []byte(validManifest), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create an invalid rite (no manifest.yaml)
	invalidDir := filepath.Join(ritesDir, "invalid-rite")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	discovery := NewDiscoveryWithPaths(ritesDir, "", "")

	rites, err := discovery.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Should only include the valid rite
	if len(rites) != 1 {
		t.Errorf("len(rites) = %d, want 1 (invalid rite should be skipped)", len(rites))
	}
	if len(rites) > 0 && rites[0].Name != "valid-rite" {
		t.Errorf("rite.Name = %q, want 'valid-rite'", rites[0].Name)
	}
}
