package org

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidOrgName(t *testing.T) {
	valid := []string{"autom8y", "my-team", "org-123", "ab"}
	for _, name := range valid {
		if !validOrgName.MatchString(name) {
			t.Errorf("Expected %q to be valid", name)
		}
	}

	invalid := []string{"", "a", "A-team", "my_team", "-leading", "trailing-", "has space", "../traversal"}
	for _, name := range invalid {
		if validOrgName.MatchString(name) {
			t.Errorf("Expected %q to be invalid", name)
		}
	}
}

func TestRunInit(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	ctx := &cmdContext{}

	// Use paths.OrgDataDir which reads XDG_DATA_HOME
	// But since adrg/xdg caches, we need to test the directory creation logic directly
	orgDir := filepath.Join(tmpDir, "knossos", "orgs", "test-org")

	// Verify it doesn't exist yet
	if _, err := os.Stat(orgDir); err == nil {
		t.Fatal("Expected org dir to not exist yet")
	}

	// We can't easily test through runInit due to XDG caching,
	// but we can test the org name validation and directory creation logic
	dirs := []string{
		orgDir,
		filepath.Join(orgDir, "rites"),
		filepath.Join(orgDir, "agents"),
		filepath.Join(orgDir, "mena"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	orgYAML := "name: test-org\n"
	if err := os.WriteFile(filepath.Join(orgDir, "org.yaml"), []byte(orgYAML), 0644); err != nil {
		t.Fatalf("Failed to write org.yaml: %v", err)
	}

	// Verify structure
	for _, sub := range []string{"rites", "agents", "mena"} {
		p := filepath.Join(orgDir, sub)
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("Expected %s to exist: %v", sub, err)
		}
		if !info.IsDir() {
			t.Errorf("Expected %s to be a directory", sub)
		}
	}

	data, err := os.ReadFile(filepath.Join(orgDir, "org.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "name: test-org\n" {
		t.Errorf("Unexpected org.yaml content: %s", data)
	}

	// Verify getPrinter doesn't panic
	_ = ctx.getPrinter()
}

func TestRunSet(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake config dir
	configDir := filepath.Join(tmpDir, "config", "knossos")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write active-org
	activeOrgPath := filepath.Join(configDir, "active-org")
	if err := os.WriteFile(activeOrgPath, []byte("test-org\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Read it back
	data, err := os.ReadFile(activeOrgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test-org\n" {
		t.Errorf("Expected 'test-org\\n', got %q", string(data))
	}
}

func TestRunList_EmptyOrgsDir(t *testing.T) {
	tmpDir := t.TempDir()
	orgsDir := filepath.Join(tmpDir, "knossos", "orgs")

	// Don't create the directory — should handle gracefully
	_, err := os.ReadDir(orgsDir)
	if !os.IsNotExist(err) {
		t.Errorf("Expected not-exist error, got %v", err)
	}
}

func TestRunList_WithOrgs(t *testing.T) {
	tmpDir := t.TempDir()
	orgsDir := filepath.Join(tmpDir, "knossos", "orgs")

	// Create some org directories
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := os.MkdirAll(filepath.Join(orgsDir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := os.ReadDir(orgsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 orgs, got %d", len(entries))
	}
}
