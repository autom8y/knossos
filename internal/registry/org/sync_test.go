package org

import (
	"fmt"
	"testing"

	"github.com/autom8y/knossos/internal/config"
)

// mockOrgContext implements OrgContext for tests.
type mockOrgContext struct {
	name        string
	registryDir string
	dataDir     string
	repos       []config.RepoConfig
}

func (m *mockOrgContext) Name() string              { return m.name }
func (m *mockOrgContext) RegistryDir() string       { return m.registryDir }
func (m *mockOrgContext) DataDir() string           { return m.dataDir }
func (m *mockOrgContext) Repos() []config.RepoConfig { return m.repos }

// mockGitHubClient implements GitHubClient for tests.
type mockGitHubClient struct {
	repos    []GitHubRepo
	// contents maps "owner/repo/path" -> []GitHubContent
	contents map[string][]GitHubContent
	// files maps "owner/repo/path" -> raw file bytes
	files    map[string][]byte
	// errors maps "owner/repo/path" -> error (for simulating failures)
	errors   map[string]error
}

func (m *mockGitHubClient) ListOrgRepos(org string) ([]GitHubRepo, error) {
	return m.repos, nil
}

func (m *mockGitHubClient) ListDirectoryContents(owner, repo, path string) ([]GitHubContent, error) {
	key := owner + "/" + repo + "/" + path
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	contents, ok := m.contents[key]
	if !ok {
		return nil, fmt.Errorf("not found: %s", key)
	}
	return contents, nil
}

func (m *mockGitHubClient) GetFileContent(owner, repo, filePath string) ([]byte, error) {
	key := owner + "/" + repo + "/" + filePath
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	data, ok := m.files[key]
	if !ok {
		return nil, fmt.Errorf("not found: %s", key)
	}
	return data, nil
}

// architectureMD is sample frontmatter for testing.
const architectureMD = `---
domain: architecture
generated_at: "2026-03-20T10:00:00Z"
expires_after: "7d"
source_hash: "abc123"
confidence: 0.92
format_version: "1.0"
---

# Architecture

Content here.
`

const conventionsMD = `---
domain: conventions
generated_at: "2026-03-20T10:00:00Z"
expires_after: "14d"
source_hash: "def456"
confidence: 0.88
format_version: "1.0"
---

# Conventions
`

const featMD = `---
domain: feat/materialization
generated_at: "2026-03-01T10:00:00Z"
expires_after: "7d"
source_hash: "ghi789"
confidence: 0.80
format_version: "1.0"
---

# Feature: Materialization
`

func TestSyncRegistry_WithExplicitRepos(t *testing.T) {
	ctx := &mockOrgContext{
		name:        "autom8y",
		registryDir: "/tmp/registry/autom8y",
		dataDir:     "/tmp/data/autom8y",
		repos: []config.RepoConfig{
			{Name: "knossos", URL: "https://github.com/autom8y/knossos", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		contents: map[string][]GitHubContent{
			"autom8y/knossos/.know": {
				{Name: "architecture.md", Path: ".know/architecture.md", Type: "file"},
				{Name: "conventions.md", Path: ".know/conventions.md", Type: "file"},
			},
		},
		files: map[string][]byte{
			"autom8y/knossos/.know/architecture.md": []byte(architectureMD),
			"autom8y/knossos/.know/conventions.md":  []byte(conventionsMD),
		},
	}

	catalog, err := SyncRegistry(ctx, client)
	if err != nil {
		t.Fatalf("SyncRegistry error: %v", err)
	}

	if catalog.Org != "autom8y" {
		t.Errorf("Org = %q, want %q", catalog.Org, "autom8y")
	}
	if catalog.RepoCount() != 1 {
		t.Fatalf("RepoCount = %d, want 1", catalog.RepoCount())
	}
	if catalog.DomainCount() != 2 {
		t.Errorf("DomainCount = %d, want 2", catalog.DomainCount())
	}
	if catalog.SyncedAt == "" {
		t.Error("SyncedAt should be set after sync")
	}

	// Verify qualified names
	d, ok := catalog.LookupDomain("autom8y::knossos::architecture")
	if !ok {
		t.Error("architecture domain not found")
	} else {
		if d.Confidence != 0.92 {
			t.Errorf("Confidence = %f, want 0.92", d.Confidence)
		}
		if d.Path != ".know/architecture.md" {
			t.Errorf("Path = %q, want %q", d.Path, ".know/architecture.md")
		}
	}
}

func TestSyncRegistry_WithFeatSubdir(t *testing.T) {
	ctx := &mockOrgContext{
		name:        "autom8y",
		registryDir: "/tmp/registry/autom8y",
		dataDir:     "/tmp/data/autom8y",
		repos: []config.RepoConfig{
			{Name: "knossos", URL: "https://github.com/autom8y/knossos", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		contents: map[string][]GitHubContent{
			"autom8y/knossos/.know": {
				{Name: "feat", Path: ".know/feat", Type: "dir"},
			},
			"autom8y/knossos/.know/feat": {
				{Name: "materialization.md", Path: ".know/feat/materialization.md", Type: "file"},
			},
		},
		files: map[string][]byte{
			"autom8y/knossos/.know/feat/materialization.md": []byte(featMD),
		},
	}

	catalog, err := SyncRegistry(ctx, client)
	if err != nil {
		t.Fatalf("SyncRegistry error: %v", err)
	}

	// Domain name should include the feat/ prefix
	d, ok := catalog.LookupDomain("autom8y::knossos::feat/materialization")
	if !ok {
		t.Error("feat/materialization domain not found")
	} else {
		if d.Domain != "feat/materialization" {
			t.Errorf("Domain = %q, want feat/materialization", d.Domain)
		}
	}
}

func TestSyncRegistry_GitHubDiscovery(t *testing.T) {
	// No explicit repos in OrgContext — falls back to GitHub API discovery.
	ctx := &mockOrgContext{
		name:        "autom8y",
		registryDir: "/tmp/registry/autom8y",
		dataDir:     "/tmp/data/autom8y",
		repos:       nil, // triggers GitHub API discovery
	}

	client := &mockGitHubClient{
		repos: []GitHubRepo{
			{Name: "knossos", HTMLURL: "https://github.com/autom8y/knossos", DefaultBranch: "main"},
			{Name: "archived-repo", HTMLURL: "https://github.com/autom8y/archived-repo", DefaultBranch: "main", Archived: true},
		},
		contents: map[string][]GitHubContent{
			"autom8y/knossos/.know": {
				{Name: "architecture.md", Path: ".know/architecture.md", Type: "file"},
			},
		},
		files: map[string][]byte{
			"autom8y/knossos/.know/architecture.md": []byte(architectureMD),
		},
	}

	catalog, err := SyncRegistry(ctx, client)
	if err != nil {
		t.Fatalf("SyncRegistry error: %v", err)
	}

	// Should include knossos (non-archived) but skip archived-repo
	if catalog.RepoCount() != 1 {
		t.Errorf("RepoCount = %d, want 1 (archived should be skipped)", catalog.RepoCount())
	}
}

func TestSyncRegistry_MissingKnowDir(t *testing.T) {
	// Repo has no .know/ directory — should not error, just produce empty domains.
	ctx := &mockOrgContext{
		name:        "autom8y",
		registryDir: "/tmp/registry/autom8y",
		dataDir:     "/tmp/data/autom8y",
		repos: []config.RepoConfig{
			{Name: "no-know-repo", URL: "https://github.com/autom8y/no-know-repo", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		// No .know/ directory configured — will return "not found"
		contents: map[string][]GitHubContent{},
		files:    map[string][]byte{},
	}

	catalog, err := SyncRegistry(ctx, client)
	if err != nil {
		t.Fatalf("SyncRegistry should not error for repos without .know/, got: %v", err)
	}

	if catalog.RepoCount() != 1 {
		t.Fatalf("RepoCount = %d, want 1", catalog.RepoCount())
	}
	if catalog.DomainCount() != 0 {
		t.Errorf("DomainCount = %d, want 0 for repo without .know/", catalog.DomainCount())
	}
}

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantDomain string
		wantErr    bool
	}{
		{
			name:       "valid frontmatter",
			input:      architectureMD,
			wantDomain: "architecture",
		},
		{
			name:    "no frontmatter",
			input:   "# Just a heading\n\nNo frontmatter here.",
			wantErr: true,
		},
		{
			name: "frontmatter without domain field",
			input: `---
generated_at: "2026-01-01T00:00:00Z"
expires_after: "7d"
---

# Content
`,
			wantDomain: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fm, err := extractFrontmatter([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fm.Domain != tc.wantDomain {
				t.Errorf("Domain = %q, want %q", fm.Domain, tc.wantDomain)
			}
		})
	}
}
