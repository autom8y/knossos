package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryDir(t *testing.T) {
	// RegistryDir is a pure function — test via XDG_DATA_HOME override.
	t.Setenv("XDG_DATA_HOME", "/tmp/test-xdg")

	got := RegistryDir("autom8y")
	want := "/tmp/test-xdg/knossos/registry/autom8y"
	if got != want {
		t.Errorf("RegistryDir(%q) = %q, want %q", "autom8y", got, want)
	}
}

func TestRegistryDirDifferentOrgs(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/test-xdg")

	got1 := RegistryDir("org-a")
	got2 := RegistryDir("org-b")

	if got1 == got2 {
		t.Error("RegistryDir should return different paths for different orgs")
	}
	if filepath.Base(got1) != "org-a" {
		t.Errorf("RegistryDir should end with org name, got %q", got1)
	}
	if filepath.Base(got2) != "org-b" {
		t.Errorf("RegistryDir should end with org name, got %q", got2)
	}
}

func TestNewOrgContext_EmptyName(t *testing.T) {
	_, err := NewOrgContext("")
	if err == nil {
		t.Error("NewOrgContext(\"\") should return error")
	}
}

func TestNewOrgContext_NoOrgYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	ctx, err := NewOrgContext("autom8y")
	if err != nil {
		t.Fatalf("NewOrgContext unexpected error: %v", err)
	}

	if ctx.Name() != "autom8y" {
		t.Errorf("Name() = %q, want %q", ctx.Name(), "autom8y")
	}

	// DataDir should be inside the XDG data dir
	wantDataDir := filepath.Join(tmpDir, "knossos", "orgs", "autom8y")
	if ctx.DataDir() != wantDataDir {
		t.Errorf("DataDir() = %q, want %q", ctx.DataDir(), wantDataDir)
	}

	// RegistryDir should be separate from DataDir
	wantRegistryDir := filepath.Join(tmpDir, "knossos", "registry", "autom8y")
	if ctx.RegistryDir() != wantRegistryDir {
		t.Errorf("RegistryDir() = %q, want %q", ctx.RegistryDir(), wantRegistryDir)
	}

	// No repos without org.yaml
	if len(ctx.Repos()) != 0 {
		t.Errorf("Repos() = %v, want empty", ctx.Repos())
	}
}

func TestNewOrgContext_WithRepos(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create org directory and org.yaml with repos
	orgDir := filepath.Join(tmpDir, "knossos", "orgs", "autom8y")
	if err := os.MkdirAll(orgDir, 0755); err != nil {
		t.Fatal(err)
	}

	orgYAMLContent := `name: autom8y
repos:
  - name: knossos
    url: https://github.com/autom8y/knossos
    default_branch: main
  - name: payments
    url: https://github.com/autom8y/payments
    default_branch: main
`
	if err := os.WriteFile(filepath.Join(orgDir, "org.yaml"), []byte(orgYAMLContent), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, err := NewOrgContext("autom8y")
	if err != nil {
		t.Fatalf("NewOrgContext unexpected error: %v", err)
	}

	repos := ctx.Repos()
	if len(repos) != 2 {
		t.Fatalf("Repos() count = %d, want 2", len(repos))
	}
	if repos[0].Name != "knossos" {
		t.Errorf("repos[0].Name = %q, want %q", repos[0].Name, "knossos")
	}
	if repos[0].URL != "https://github.com/autom8y/knossos" {
		t.Errorf("repos[0].URL = %q, want URL", repos[0].URL)
	}
	if repos[0].DefaultBranch != "main" {
		t.Errorf("repos[0].DefaultBranch = %q, want %q", repos[0].DefaultBranch, "main")
	}
}

func TestNewOrgContext_MalformedOrgYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create org directory with malformed org.yaml
	orgDir := filepath.Join(tmpDir, "knossos", "orgs", "autom8y")
	if err := os.MkdirAll(orgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orgDir, "org.yaml"), []byte("not: valid: yaml: :::"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should not error — malformed org.yaml degrades gracefully
	ctx, err := NewOrgContext("autom8y")
	if err != nil {
		t.Fatalf("NewOrgContext should not fail on malformed org.yaml, got: %v", err)
	}

	if len(ctx.Repos()) != 0 {
		t.Error("Repos() should be empty when org.yaml is malformed")
	}
}

func TestDefaultOrgContext_NoActiveOrg(t *testing.T) {
	// Ensure no active org is set
	t.Setenv("KNOSSOS_ORG", "")
	// Use a temp dir for config so active-org file doesn't exist
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	_, err := DefaultOrgContext()
	if err == nil {
		t.Error("DefaultOrgContext() should return error when no active org configured")
	}
}

func TestDefaultOrgContext_WithActiveOrg(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("KNOSSOS_ORG", "test-org")

	ctx, err := DefaultOrgContext()
	if err != nil {
		t.Fatalf("DefaultOrgContext unexpected error: %v", err)
	}

	if ctx.Name() != "test-org" {
		t.Errorf("Name() = %q, want %q", ctx.Name(), "test-org")
	}
}
