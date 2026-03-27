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
	contents map[string][]GitHubContent
	files    map[string][]byte
	tree     []GitHubTreeEntry
	treeErr  error
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

func (m *mockGitHubClient) GetTree(owner, repo, sha string) ([]GitHubTreeEntry, error) {
	if m.treeErr != nil {
		return nil, m.treeErr
	}
	return m.tree, nil
}

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
		name: "autom8y",
		repos: []config.RepoConfig{
			{Name: "knossos", URL: "https://github.com/autom8y/knossos", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		tree: []GitHubTreeEntry{
			{Path: ".know", Type: "tree"},
		},
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
	if catalog.SchemaVersion != "1.1" {
		t.Errorf("SchemaVersion = %q, want 1.1", catalog.SchemaVersion)
	}
	if catalog.RepoCount() != 1 {
		t.Fatalf("RepoCount = %d, want 1", catalog.RepoCount())
	}
	if catalog.DomainCount() != 2 {
		t.Errorf("DomainCount = %d, want 2", catalog.DomainCount())
	}

	d, ok := catalog.LookupDomain("autom8y::knossos::architecture")
	if !ok {
		t.Error("architecture domain not found")
	} else {
		if d.Confidence != 0.92 {
			t.Errorf("Confidence = %f, want 0.92", d.Confidence)
		}
		if d.Scope != "" {
			t.Errorf("Scope = %q, want empty for root domain", d.Scope)
		}
	}
}

func TestSyncRegistry_WithFeatSubdir(t *testing.T) {
	ctx := &mockOrgContext{
		name: "autom8y",
		repos: []config.RepoConfig{
			{Name: "knossos", URL: "https://github.com/autom8y/knossos", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		tree: []GitHubTreeEntry{
			{Path: ".know", Type: "tree"},
			{Path: ".know/feat", Type: "tree"},
		},
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

	d, ok := catalog.LookupDomain("autom8y::knossos::feat/materialization")
	if !ok {
		t.Error("feat/materialization domain not found")
	} else if d.Domain != "feat/materialization" {
		t.Errorf("Domain = %q, want feat/materialization", d.Domain)
	}
}

func TestSyncRegistry_NestedScopes(t *testing.T) {
	ctx := &mockOrgContext{
		name: "autom8y",
		repos: []config.RepoConfig{
			{Name: "autom8y", URL: "https://github.com/autom8y/autom8y", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		tree: []GitHubTreeEntry{
			{Path: ".know", Type: "tree"},
			{Path: "services/ads/.know", Type: "tree"},
			{Path: "services/auth/.know", Type: "tree"},
			{Path: "sdks/python/autom8y-meta/.know", Type: "tree"},
		},
		contents: map[string][]GitHubContent{
			"autom8y/autom8y/.know": {
				{Name: "architecture.md", Path: ".know/architecture.md", Type: "file"},
			},
			"autom8y/autom8y/services/ads/.know": {
				{Name: "architecture.md", Path: "services/ads/.know/architecture.md", Type: "file"},
				{Name: "conventions.md", Path: "services/ads/.know/conventions.md", Type: "file"},
			},
			"autom8y/autom8y/services/auth/.know": {
				{Name: "architecture.md", Path: "services/auth/.know/architecture.md", Type: "file"},
			},
			"autom8y/autom8y/sdks/python/autom8y-meta/.know": {
				{Name: "conventions.md", Path: "sdks/python/autom8y-meta/.know/conventions.md", Type: "file"},
			},
		},
		files: map[string][]byte{
			"autom8y/autom8y/.know/architecture.md":                     []byte(architectureMD),
			"autom8y/autom8y/services/ads/.know/architecture.md":        []byte(architectureMD),
			"autom8y/autom8y/services/ads/.know/conventions.md":         []byte(conventionsMD),
			"autom8y/autom8y/services/auth/.know/architecture.md":       []byte(architectureMD),
			"autom8y/autom8y/sdks/python/autom8y-meta/.know/conventions.md": []byte(conventionsMD),
		},
	}

	catalog, err := SyncRegistry(ctx, client)
	if err != nil {
		t.Fatalf("SyncRegistry error: %v", err)
	}

	if catalog.DomainCount() != 5 {
		t.Fatalf("DomainCount = %d, want 5", catalog.DomainCount())
	}

	// Root scope domain
	d, ok := catalog.LookupDomain("autom8y::autom8y::architecture")
	if !ok {
		t.Error("root architecture not found")
	} else if d.Scope != "" {
		t.Errorf("root scope = %q, want empty", d.Scope)
	}

	// Service-scoped domain
	d, ok = catalog.LookupDomain("autom8y::autom8y/services/ads::architecture")
	if !ok {
		t.Error("services/ads architecture not found")
	} else {
		if d.Scope != "services/ads" {
			t.Errorf("Scope = %q, want services/ads", d.Scope)
		}
		if d.Path != "services/ads/.know/architecture.md" {
			t.Errorf("Path = %q, want services/ads/.know/architecture.md", d.Path)
		}
	}

	// Multi-level scope
	d, ok = catalog.LookupDomain("autom8y::autom8y/sdks/python/autom8y-meta::conventions")
	if !ok {
		t.Error("sdks/python/autom8y-meta conventions not found")
	} else if d.Scope != "sdks/python/autom8y-meta" {
		t.Errorf("Scope = %q, want sdks/python/autom8y-meta", d.Scope)
	}

	// Sibling scope
	_, ok = catalog.LookupDomain("autom8y::autom8y/services/auth::architecture")
	if !ok {
		t.Error("services/auth architecture not found")
	}
}

func TestSyncRegistry_ExcludedPaths(t *testing.T) {
	ctx := &mockOrgContext{
		name: "autom8y",
		repos: []config.RepoConfig{
			{Name: "repo", URL: "https://github.com/autom8y/repo", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		tree: []GitHubTreeEntry{
			{Path: ".know", Type: "tree"},
			{Path: "vendor/dep/.know", Type: "tree"},
			{Path: "node_modules/pkg/.know", Type: "tree"},
			{Path: ".terraform/modules/foo/.know", Type: "tree"},
			{Path: ".knossos/worktrees/test/.know", Type: "tree"},
		},
		contents: map[string][]GitHubContent{
			"autom8y/repo/.know": {
				{Name: "architecture.md", Path: ".know/architecture.md", Type: "file"},
			},
		},
		files: map[string][]byte{
			"autom8y/repo/.know/architecture.md": []byte(architectureMD),
		},
	}

	catalog, err := SyncRegistry(ctx, client)
	if err != nil {
		t.Fatalf("SyncRegistry error: %v", err)
	}

	// Only root .know/ should be discovered, excluded paths ignored
	if catalog.DomainCount() != 1 {
		t.Errorf("DomainCount = %d, want 1 (excluded paths should be ignored)", catalog.DomainCount())
	}
}

func TestSyncRegistry_DepthLimit(t *testing.T) {
	ctx := &mockOrgContext{
		name: "autom8y",
		repos: []config.RepoConfig{
			{Name: "repo", URL: "https://github.com/autom8y/repo", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		tree: []GitHubTreeEntry{
			{Path: ".know", Type: "tree"},
			// Depth 6 — exceeds MaxScopeDepth of 5
			{Path: "a/b/c/d/e/f/.know", Type: "tree"},
		},
		contents: map[string][]GitHubContent{
			"autom8y/repo/.know": {
				{Name: "architecture.md", Path: ".know/architecture.md", Type: "file"},
			},
		},
		files: map[string][]byte{
			"autom8y/repo/.know/architecture.md": []byte(architectureMD),
		},
	}

	catalog, err := SyncRegistry(ctx, client)
	if err != nil {
		t.Fatalf("SyncRegistry error: %v", err)
	}

	// Only root .know/ should be discovered, deep path exceeds depth limit
	if catalog.DomainCount() != 1 {
		t.Errorf("DomainCount = %d, want 1 (deep .know/ should be excluded by depth limit)", catalog.DomainCount())
	}
}

func TestSyncRegistry_TreeFallback(t *testing.T) {
	ctx := &mockOrgContext{
		name: "autom8y",
		repos: []config.RepoConfig{
			{Name: "knossos", URL: "https://github.com/autom8y/knossos", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		treeErr: fmt.Errorf("tree API unavailable"),
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

	// Should fall back to scanning root .know/ only
	if catalog.DomainCount() != 1 {
		t.Errorf("DomainCount = %d, want 1 (tree fallback should find root .know/)", catalog.DomainCount())
	}
}

func TestSyncRegistry_GitHubDiscovery(t *testing.T) {
	ctx := &mockOrgContext{
		name:  "autom8y",
		repos: nil, // triggers GitHub API discovery
	}

	client := &mockGitHubClient{
		repos: []GitHubRepo{
			{Name: "knossos", HTMLURL: "https://github.com/autom8y/knossos", DefaultBranch: "main"},
			{Name: "archived-repo", HTMLURL: "https://github.com/autom8y/archived-repo", DefaultBranch: "main", Archived: true},
		},
		tree: []GitHubTreeEntry{
			{Path: ".know", Type: "tree"},
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

	if catalog.RepoCount() != 1 {
		t.Errorf("RepoCount = %d, want 1 (archived should be skipped)", catalog.RepoCount())
	}
}

func TestSyncRegistry_MissingKnowDir(t *testing.T) {
	ctx := &mockOrgContext{
		name: "autom8y",
		repos: []config.RepoConfig{
			{Name: "no-know-repo", URL: "https://github.com/autom8y/no-know-repo", DefaultBranch: "main"},
		},
	}

	client := &mockGitHubClient{
		tree:     nil, // empty tree
		treeErr:  fmt.Errorf("not found"),
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
		t.Errorf("DomainCount = %d, want 0", catalog.DomainCount())
	}
}

func TestScopeFromKnowPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{".know", ""},
		{"services/ads/.know", "services/ads"},
		{"sdks/python/autom8y-meta/.know", "sdks/python/autom8y-meta"},
		{"a/b/c/.know", "a/b/c"},
	}
	for _, tc := range tests {
		if got := scopeFromKnowPath(tc.input); got != tc.want {
			t.Errorf("scopeFromKnowPath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDeriveDomainName(t *testing.T) {
	tests := []struct {
		filePath string
		scope    string
		want     string
	}{
		{".know/architecture.md", "", "architecture"},
		{".know/feat/materialization.md", "", "feat/materialization"},
		{"services/ads/.know/architecture.md", "services/ads", "architecture"},
		{"services/ads/.know/feat/materialization.md", "services/ads", "feat/materialization"},
	}
	for _, tc := range tests {
		if got := deriveDomainName(tc.filePath, tc.scope); got != tc.want {
			t.Errorf("deriveDomainName(%q, %q) = %q, want %q", tc.filePath, tc.scope, got, tc.want)
		}
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
