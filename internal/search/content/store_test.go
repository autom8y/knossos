package content

import (
	"os"
	"path/filepath"
	"testing"

	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreBakedStore_LoadContent(t *testing.T) {
	tests := []struct {
		name        string
		entry       registryorg.DomainEntry
		fileContent string
		wantBody    string
		wantErr     bool
	}{
		{
			name: "strips frontmatter and returns body",
			entry: registryorg.DomainEntry{
				QualifiedName: "autom8y::knossos::architecture",
				Domain:        "architecture",
				Path:          ".know/architecture.md",
			},
			fileContent: "---\ndomain: architecture\ngenerated_at: \"2026-03-23T18:00:00Z\"\n---\n\n# Architecture\n\nThis is the body.",
			wantBody:    "# Architecture\n\nThis is the body.",
		},
		{
			name: "no frontmatter returns full content",
			entry: registryorg.DomainEntry{
				QualifiedName: "autom8y::knossos::conventions",
				Domain:        "conventions",
				Path:          ".know/conventions.md",
			},
			fileContent: "# Conventions\n\nNo frontmatter here.",
			wantBody:    "# Conventions\n\nNo frontmatter here.",
		},
		{
			name: "nested path within repo",
			entry: registryorg.DomainEntry{
				QualifiedName: "autom8y::knossos::feat/index",
				Domain:        "feat/index",
				Path:          ".know/feat/INDEX.md",
			},
			fileContent: "---\ndomain: feat/index\n---\n\n# Feature Index",
			wantBody:    "# Feature Index",
		},
		{
			name: "missing file returns error",
			entry: registryorg.DomainEntry{
				QualifiedName: "autom8y::knossos::nonexistent",
				Domain:        "nonexistent",
				Path:          ".know/nonexistent.md",
			},
			wantErr: true,
		},
		{
			name: "invalid qualified name returns error",
			entry: registryorg.DomainEntry{
				QualifiedName: "bad-format",
				Domain:        "bad",
				Path:          ".know/bad.md",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			store := NewPreBakedStore(tmpDir)

			// Create the file if content is provided.
			if tt.fileContent != "" {
				repoName := repoFromQualifiedName(tt.entry.QualifiedName)
				filePath := filepath.Join(tmpDir, repoName, tt.entry.Path)
				require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
				require.NoError(t, os.WriteFile(filePath, []byte(tt.fileContent), 0o644))
			}

			got, err := store.LoadContent(tt.entry)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, got)
		})
	}
}

func TestPreBakedStore_HasContent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewPreBakedStore(tmpDir)

	entry := registryorg.DomainEntry{
		QualifiedName: "autom8y::knossos::architecture",
		Domain:        "architecture",
		Path:          ".know/architecture.md",
	}

	// Before file exists.
	assert.False(t, store.HasContent(entry))

	// Create the file.
	filePath := filepath.Join(tmpDir, "knossos", ".know", "architecture.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte("# Arch"), 0o644))

	// After file exists.
	assert.True(t, store.HasContent(entry))

	// Invalid qualified name.
	badEntry := registryorg.DomainEntry{
		QualifiedName: "bad",
		Domain:        "bad",
		Path:          ".know/bad.md",
	}
	assert.False(t, store.HasContent(badEntry))
}

func TestLocalStore_LoadContent(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "knossos")
	knowDir := filepath.Join(repoRoot, ".know")
	require.NoError(t, os.MkdirAll(knowDir, 0o755))

	content := "---\ndomain: architecture\n---\n\n# Architecture\n\nFull body content here."
	require.NoError(t, os.WriteFile(filepath.Join(knowDir, "architecture.md"), []byte(content), 0o644))

	store := NewLocalStore(map[string]string{
		"knossos": repoRoot,
	})

	entry := registryorg.DomainEntry{
		QualifiedName: "autom8y::knossos::architecture",
		Domain:        "architecture",
		Path:          ".know/architecture.md",
	}

	got, err := store.LoadContent(entry)
	require.NoError(t, err)
	assert.Equal(t, "# Architecture\n\nFull body content here.", got)
	assert.True(t, store.HasContent(entry))
}

func TestLocalStore_MissingRepo(t *testing.T) {
	store := NewLocalStore(map[string]string{})

	entry := registryorg.DomainEntry{
		QualifiedName: "autom8y::knossos::architecture",
		Domain:        "architecture",
		Path:          ".know/architecture.md",
	}

	_, err := store.LoadContent(entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no local path for repo")
	assert.False(t, store.HasContent(entry))
}

func TestRepoFromQualifiedName(t *testing.T) {
	tests := []struct {
		qn   string
		want string
	}{
		{"autom8y::knossos::architecture", "knossos"},
		{"autom8y::auth::conventions", "auth"},
		{"org::repo::domain", "repo"},
		{"bad-format", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.qn, func(t *testing.T) {
			assert.Equal(t, tt.want, repoFromQualifiedName(tt.qn))
		})
	}
}

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"with frontmatter", "---\ndomain: arch\n---\n# Title\nBody text", "# Title\nBody text"},
		{"no frontmatter", "# Title\nBody text", "# Title\nBody text"},
		{"empty", "", ""},
		{"only frontmatter", "---\nfoo: bar\n---", ""},
		{"no closing delimiter", "---\nfoo: bar\nstuff", "---\nfoo: bar\nstuff"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stripFrontmatter(tt.input))
		})
	}
}
