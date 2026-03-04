package userscope

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/provenance"
)

func TestIsExecutableFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"shell script", "/path/to/hook.sh", true},
		{"bash script", "/path/to/hook.bash", true},
		{"zsh script", "/path/to/hook.zsh", true},
		{"python script", "/path/to/hook.py", true},
		{"ruby script", "/path/to/hook.rb", true},
		{"perl script", "/path/to/hook.pl", true},
		{"markdown file", "/path/to/agent.md", false},
		{"yaml file", "/path/to/config.yaml", false},
		{"json file", "/path/to/manifest.json", false},
		{"lib directory file", "/path/to/lib/helper.txt", true},
		{"hook- prefix", "/path/to/hook-pre-commit", true},
		{"pre- prefix", "/path/to/pre-push", true},
		{"post- prefix", "/path/to/post-commit", true},
		{"no extension", "/path/to/plainfile", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExecutableFile(tt.path)
			if got != tt.want {
				t.Errorf("isExecutableFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestResourcePrefixForType(t *testing.T) {
	tests := []struct {
		resource SyncResource
		want     string
	}{
		{ResourceAgents, "agents/"},
		{ResourceHooks, "hooks/"},
		{ResourceMena, ""},
		{SyncResource("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.resource), func(t *testing.T) {
			got := resourcePrefixForType(tt.resource)
			if got != tt.want {
				t.Errorf("resourcePrefixForType(%q) = %q, want %q", tt.resource, got, tt.want)
			}
		})
	}
}

func TestSyncResourceIsValid(t *testing.T) {
	tests := []struct {
		resource SyncResource
		want     bool
	}{
		{ResourceAll, true},
		{ResourceAgents, true},
		{ResourceMena, true},
		{ResourceHooks, true},
		{SyncResource("unknown"), false},
		{SyncResource(""), true}, // ResourceAll is ""
	}

	for _, tt := range tests {
		t.Run(string(tt.resource), func(t *testing.T) {
			got := tt.resource.IsValid()
			if got != tt.want {
				t.Errorf("SyncResource(%q).IsValid() = %v, want %v", tt.resource, got, tt.want)
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("ErrKnossosHomeNotSet", func(t *testing.T) {
		err := ErrKnossosHomeNotSet()
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
		userSyncError := &UserSyncError{}
		if !errors.As(err, &userSyncError) {
			t.Errorf("expected *UserSyncError, got %T", err)
		}
	})

	t.Run("ErrInvalidResourceType", func(t *testing.T) {
		err := ErrInvalidResourceType()
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
	})
}

func TestCopyUserFile(t *testing.T) {
	t.Run("copies content", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcPath := filepath.Join(tmpDir, "source.md")
		dstPath := filepath.Join(tmpDir, "dest.md")

		content := []byte("# Agent Content\nSome instructions.\n")
		os.WriteFile(srcPath, content, 0644)

		if err := copyUserFile(srcPath, dstPath); err != nil {
			t.Fatalf("copyUserFile: %v", err)
		}

		got, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("read dest: %v", err)
		}
		if string(got) != string(content) {
			t.Errorf("content mismatch: got %q, want %q", got, content)
		}
	})

	t.Run("creates parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcPath := filepath.Join(tmpDir, "source.md")
		dstPath := filepath.Join(tmpDir, "nested", "deep", "dest.md")

		os.WriteFile(srcPath, []byte("content"), 0644)

		if err := copyUserFile(srcPath, dstPath); err != nil {
			t.Fatalf("copyUserFile: %v", err)
		}

		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Error("expected dest file to be created in nested directory")
		}
	})

	t.Run("sets executable bit for scripts", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcPath := filepath.Join(tmpDir, "hook.sh")
		dstPath := filepath.Join(tmpDir, "dest", "hook.sh")

		// Source has no executable bit
		os.WriteFile(srcPath, []byte("#!/bin/sh\necho ok"), 0644)

		if err := copyUserFile(srcPath, dstPath); err != nil {
			t.Fatalf("copyUserFile: %v", err)
		}

		info, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("stat dest: %v", err)
		}
		if info.Mode()&0111 == 0 {
			t.Error("expected executable bit to be set for .sh file")
		}
	})
}

func TestRemoveUserOrphan(t *testing.T) {
	t.Run("removes knossos-owned file", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentPath := filepath.Join(tmpDir, "agents", "orphan.md")
		os.MkdirAll(filepath.Dir(agentPath), 0755)
		os.WriteFile(agentPath, []byte("# Orphan"), 0644)

		manifest := &provenance.ProvenanceManifest{
			Entries: map[string]*provenance.ProvenanceEntry{
				"agents/orphan.md": {
					Owner: provenance.OwnerKnossos,
					Scope: provenance.ScopeUser,
				},
			},
		}

		removeUserOrphan("agents/orphan.md", manifest, tmpDir)

		// File should be deleted
		if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
			t.Error("expected orphan file to be removed")
		}
		// Manifest entry should be deleted
		if _, exists := manifest.Entries["agents/orphan.md"]; exists {
			t.Error("expected manifest entry to be removed")
		}
	})

	t.Run("skips user-owned", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentPath := filepath.Join(tmpDir, "agents", "user-agent.md")
		os.MkdirAll(filepath.Dir(agentPath), 0755)
		os.WriteFile(agentPath, []byte("# User Agent"), 0644)

		manifest := &provenance.ProvenanceManifest{
			Entries: map[string]*provenance.ProvenanceEntry{
				"agents/user-agent.md": {
					Owner: provenance.OwnerUser,
					Scope: provenance.ScopeUser,
				},
			},
		}

		removeUserOrphan("agents/user-agent.md", manifest, tmpDir)

		// File should still exist
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			t.Error("user-owned file should not be removed")
		}
		// Manifest entry should still exist
		if _, exists := manifest.Entries["agents/user-agent.md"]; !exists {
			t.Error("user-owned manifest entry should not be removed")
		}
	})

	t.Run("handles nil entry", func(t *testing.T) {
		tmpDir := t.TempDir()
		manifest := &provenance.ProvenanceManifest{
			Entries: map[string]*provenance.ProvenanceEntry{},
		}
		// Should not panic
		removeUserOrphan("agents/nonexistent.md", manifest, tmpDir)
	})

	t.Run("removes directory entry", func(t *testing.T) {
		tmpDir := t.TempDir()
		dirPath := filepath.Join(tmpDir, "commands", "my-cmd")
		os.MkdirAll(dirPath, 0755)
		os.WriteFile(filepath.Join(dirPath, "INDEX.md"), []byte("# Cmd"), 0644)

		manifest := &provenance.ProvenanceManifest{
			Entries: map[string]*provenance.ProvenanceEntry{
				"commands/my-cmd/": {
					Owner: provenance.OwnerKnossos,
					Scope: provenance.ScopeUser,
				},
			},
		}

		removeUserOrphan("commands/my-cmd/", manifest, tmpDir)

		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			t.Error("expected directory to be removed")
		}
	})
}

func TestFindMenaSource(t *testing.T) {
	tmpDir := t.TempDir()
	menaDir := filepath.Join(tmpDir, "mena")

	// Create source files with mena extensions
	droDir := filepath.Join(menaDir, "my-cmd")
	os.MkdirAll(droDir, 0755)
	os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("# Dro"), 0644)

	legoDir := filepath.Join(menaDir, "my-skill")
	os.MkdirAll(legoDir, 0755)
	os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), []byte("# Lego"), 0644)

	// Create a plain file too
	os.WriteFile(filepath.Join(menaDir, "plain.md"), []byte("# Plain"), 0644)

	t.Run("finds dro variant", func(t *testing.T) {
		src, _ := findMenaSource(filepath.Join(menaDir, "my-cmd"), "INDEX.md")
		if src == "" {
			t.Error("expected to find dro variant of INDEX.md")
		}
	})

	t.Run("finds lego variant", func(t *testing.T) {
		src, _ := findMenaSource(filepath.Join(menaDir, "my-skill"), "INDEX.md")
		if src == "" {
			t.Error("expected to find lego variant of INDEX.md")
		}
	})

	t.Run("finds exact match", func(t *testing.T) {
		src, _ := findMenaSource(menaDir, "plain.md")
		if src == "" {
			t.Error("expected to find exact match for plain.md")
		}
	})

	t.Run("returns empty for missing file", func(t *testing.T) {
		src, _ := findMenaSource(menaDir, "nonexistent.md")
		if src != "" {
			t.Errorf("expected empty path for missing file, got %q", src)
		}
	})
}

func TestCountUserCollisions(t *testing.T) {
	skipped := []UserSkippedEntry{
		{Name: "agents/a.md", Reason: "collision with rite resource"},
		{Name: "agents/b.md", Reason: "user-created"},
		{Name: "agents/c.md", Reason: "collision with rite resource"},
		{Name: "agents/d.md", Reason: "diverged"},
	}

	got := countUserCollisions(skipped)
	if got != 2 {
		t.Errorf("countUserCollisions = %d, want 2", got)
	}
}

func TestMatchesKnossosKey(t *testing.T) {
	keys := map[string]bool{
		"commands/spike/":    true,
		"commands/spike.md":  true,
		"skills/standards/":  true,
		"commands/architect": true,
	}

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"exact match", "commands/spike.md", true},
		{"prefix match", "commands/spike/INDEX.md", true},
		{"deep prefix match", "skills/standards/code.md", true},
		{"no match", "commands/user-custom.md", false},
		{"partial no match", "commands/spike-extra/INDEX.md", false},
		{"exact non-suffix match", "commands/architect", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesKnossosKey(tt.key, keys)
			if got != tt.want {
				t.Errorf("matchesKnossosKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
