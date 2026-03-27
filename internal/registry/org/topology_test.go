package org

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- LoadTopology tests ----

func TestLoadTopology_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.yaml")
	content := `schema_version: "1.0"
org: autom8y
groups:
  - name: Service layer
    repos:
      - name: autom8y-data
        role: core data layer
        direction: upstream
      - name: autom8y-scheduling
        role: job orchestration
edges:
  - from: autom8y-scheduling
    to: autom8y-data
    label: reads job/campaign data
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	topo, err := LoadTopology(path)
	if err != nil {
		t.Fatalf("LoadTopology() error = %v", err)
	}
	if topo == nil {
		t.Fatal("LoadTopology() returned nil for valid file")
	}

	if topo.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", topo.SchemaVersion, "1.0")
	}
	if topo.Org != "autom8y" {
		t.Errorf("Org = %q, want %q", topo.Org, "autom8y")
	}
	if len(topo.Groups) != 1 {
		t.Fatalf("Groups count = %d, want 1", len(topo.Groups))
	}
	if topo.Groups[0].Name != "Service layer" {
		t.Errorf("Groups[0].Name = %q, want %q", topo.Groups[0].Name, "Service layer")
	}
	if len(topo.Groups[0].Repos) != 2 {
		t.Fatalf("Repos count = %d, want 2", len(topo.Groups[0].Repos))
	}
	if topo.Groups[0].Repos[0].Direction != "upstream" {
		t.Errorf("Repos[0].Direction = %q, want %q", topo.Groups[0].Repos[0].Direction, "upstream")
	}
	if len(topo.Edges) != 1 {
		t.Fatalf("Edges count = %d, want 1", len(topo.Edges))
	}
	if topo.Edges[0].From != "autom8y-scheduling" {
		t.Errorf("Edges[0].From = %q, want %q", topo.Edges[0].From, "autom8y-scheduling")
	}
}

func TestLoadTopology_MissingFile_FailOpen(t *testing.T) {
	topo, err := LoadTopology("/nonexistent/path/topology.yaml")
	if err != nil {
		t.Fatalf("LoadTopology() should not error on missing file, got: %v", err)
	}
	if topo != nil {
		t.Error("LoadTopology() should return nil for missing file")
	}
}

func TestLoadTopology_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	topo, err := LoadTopology(path)
	if err == nil {
		t.Fatal("LoadTopology() should error on malformed YAML")
	}
	if topo != nil {
		t.Error("LoadTopology() should return nil on error")
	}
}

// ---- RenderTopology tests ----

func TestRenderTopology_FullConfig(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		Groups: []TopologyGroup{
			{
				Name: "Service layer",
				Repos: []TopologyRepo{
					{Name: "autom8y-data", Role: "core data layer", Direction: "upstream"},
					{Name: "autom8y-scheduling", Role: "job orchestration"},
				},
			},
			{
				Name: "Tooling",
				Repos: []TopologyRepo{
					{Name: "knossos", Role: "developer platform (this system)"},
				},
			},
		},
		Edges: []TopologyEdge{
			{From: "autom8y-scheduling", To: "autom8y-data", Label: "reads job/campaign data"},
			{From: "autom8y-scheduling", To: "autom8y-sms", Label: "triggers notifications"},
		},
	}

	domainCounts := map[string]int{
		"autom8y-data":       8,
		"autom8y-scheduling": 5,
		"autom8y-sms":        7,
		"knossos":            47,
	}

	result := RenderTopology(topo, domainCounts, nil)

	// Header.
	if !strings.Contains(result, "--- ORG TOPOLOGY ---") {
		t.Error("missing topology header")
	}

	// Org summary line.
	if !strings.Contains(result, "Organization: autom8y") {
		t.Error("missing org name in summary")
	}
	// 3 repos in groups + 1 uncategorized (autom8y-sms is in edges but not groups) = 4 repos.
	if !strings.Contains(result, "4 repos") {
		t.Errorf("missing repo count in summary, got: %s", result)
	}
	// 8 + 5 + 47 + 7 (autom8y-sms uncategorized) = 67.
	if !strings.Contains(result, "~67 knowledge domains") {
		t.Errorf("missing domain count in summary, got: %s", result)
	}

	// Group headers.
	if !strings.Contains(result, "Service layer:") {
		t.Error("missing Service layer group header")
	}
	if !strings.Contains(result, "Tooling:") {
		t.Error("missing Tooling group header")
	}

	// Repo entries with domain counts and roles.
	if !strings.Contains(result, "autom8y-data (8 domains) -- core data layer") {
		t.Error("missing autom8y-data repo line")
	}
	if !strings.Contains(result, "autom8y-scheduling (5 domains) -- job orchestration") {
		t.Error("missing autom8y-scheduling repo line")
	}
	if !strings.Contains(result, "knossos (47 domains) -- developer platform (this system)") {
		t.Error("missing knossos repo line")
	}

	// Inbound edge arrow (autom8y-scheduling depends on autom8y-data).
	if !strings.Contains(result, "<- autom8y-scheduling (5 domains) reads job/campaign data") {
		t.Error("missing inbound edge arrow for autom8y-data")
	}

	// Outbound edge arrow (autom8y-scheduling triggers autom8y-sms).
	if !strings.Contains(result, "-> autom8y-sms (7 domains) triggers notifications") {
		t.Error("missing outbound edge arrow for autom8y-scheduling")
	}
}

func TestRenderTopology_NoEdges(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Groups: []TopologyGroup{
			{
				Name: "Apps",
				Repos: []TopologyRepo{
					{Name: "app-a", Role: "frontend"},
					{Name: "app-b", Role: "backend"},
				},
			},
		},
		Edges: nil,
	}

	domainCounts := map[string]int{
		"app-a": 3,
		"app-b": 5,
	}

	result := RenderTopology(topo, domainCounts, nil)

	if !strings.Contains(result, "app-a (3 domains) -- frontend") {
		t.Error("missing app-a repo line")
	}
	if !strings.Contains(result, "app-b (5 domains) -- backend") {
		t.Error("missing app-b repo line")
	}
	// No arrow notation should appear.
	if strings.Contains(result, "<-") || strings.Contains(result, "->") {
		t.Error("should not contain arrow notation when there are no edges")
	}
}

func TestRenderTopology_UncategorizedRepos(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Groups: []TopologyGroup{
			{
				Name: "Core",
				Repos: []TopologyRepo{
					{Name: "core-api", Role: "API server"},
				},
			},
		},
	}

	// extra-service is in the catalog but NOT in the topology groups.
	domainCounts := map[string]int{
		"core-api":      10,
		"extra-service": 3,
		"another-repo":  1,
	}

	result := RenderTopology(topo, domainCounts, nil)

	// Uncategorized repos should appear under "Other" group.
	if !strings.Contains(result, "Other:") {
		t.Error("missing Other group for uncategorized repos")
	}
	if !strings.Contains(result, "extra-service (3 domains)") {
		t.Error("missing extra-service in Other group")
	}
	if !strings.Contains(result, "another-repo (1 domains)") {
		t.Error("missing another-repo in Other group")
	}

	// Total counts should include uncategorized repos.
	if !strings.Contains(result, "3 repos") {
		t.Error("total repo count should include uncategorized repos")
	}
	if !strings.Contains(result, "~14 knowledge domains") {
		t.Error("total domain count should include uncategorized repos")
	}
}

func TestRenderTopology_NilConfig(t *testing.T) {
	result := RenderTopology(nil, map[string]int{"foo": 5}, nil)
	if result != "" {
		t.Errorf("RenderTopology(nil) should return empty string, got %q", result)
	}
}

func TestRenderTopology_ZeroDomainCounts(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Groups: []TopologyGroup{
			{
				Name: "All",
				Repos: []TopologyRepo{
					{Name: "new-repo", Role: "not yet cataloged"},
				},
			},
		},
	}

	result := RenderTopology(topo, map[string]int{}, nil)

	if !strings.Contains(result, "new-repo (0 domains) -- not yet cataloged") {
		t.Error("repos not in domainCounts should show 0 domains")
	}
}

// ---- DomainCountsFromCatalog tests ----

func TestDomainCountsFromCatalog(t *testing.T) {
	catalog := &DomainCatalog{
		Repos: []RepoEntry{
			{
				Name: "knossos",
				Domains: []DomainEntry{
					{QualifiedName: "autom8y::knossos::architecture"},
					{QualifiedName: "autom8y::knossos::conventions"},
				},
			},
			{
				Name:    "empty-repo",
				Domains: nil,
			},
			{
				Name: "other",
				Domains: []DomainEntry{
					{QualifiedName: "autom8y::other::arch"},
				},
			},
		},
	}

	counts := DomainCountsFromCatalog(catalog)

	if counts["knossos"] != 2 {
		t.Errorf("knossos domain count = %d, want 2", counts["knossos"])
	}
	if counts["empty-repo"] != 0 {
		t.Errorf("empty-repo domain count = %d, want 0", counts["empty-repo"])
	}
	if counts["other"] != 1 {
		t.Errorf("other domain count = %d, want 1", counts["other"])
	}
}

func TestDomainCountsFromCatalog_NilCatalog(t *testing.T) {
	counts := DomainCountsFromCatalog(nil)
	if len(counts) != 0 {
		t.Errorf("nil catalog should return empty map, got %d entries", len(counts))
	}
}

// ---- DomainInfoFromCatalog tests ----

func TestDomainInfoFromCatalog(t *testing.T) {
	catalog := &DomainCatalog{
		Repos: []RepoEntry{
			{
				Name: "monorepo",
				Domains: []DomainEntry{
					{QualifiedName: "org::monorepo::arch", Scope: ""},
					{QualifiedName: "org::monorepo::conventions", Scope: ""},
					{QualifiedName: "org::monorepo/svc-a::arch", Scope: "svc-a"},
					{QualifiedName: "org::monorepo/svc-a::conventions", Scope: "svc-a"},
					{QualifiedName: "org::monorepo/svc-b::arch", Scope: "svc-b"},
				},
			},
			{
				Name: "simple-repo",
				Domains: []DomainEntry{
					{QualifiedName: "org::simple-repo::arch", Scope: ""},
					{QualifiedName: "org::simple-repo::conventions", Scope: ""},
				},
			},
			{
				Name:    "empty-repo",
				Domains: nil,
			},
		},
	}

	info := DomainInfoFromCatalog(catalog)

	tests := []struct {
		name        string
		repo        string
		wantTotal   int
		wantRoot    int
		wantScoped  int
		wantScopes  int
	}{
		{"monorepo with scopes", "monorepo", 5, 2, 3, 2},
		{"simple repo no scopes", "simple-repo", 2, 2, 0, 0},
		{"empty repo", "empty-repo", 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si, ok := info[tt.repo]
			if !ok {
				t.Fatalf("missing ScopeInfo for %q", tt.repo)
			}
			if si.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", si.Total, tt.wantTotal)
			}
			if si.RootCount != tt.wantRoot {
				t.Errorf("RootCount = %d, want %d", si.RootCount, tt.wantRoot)
			}
			if si.ScopedCount != tt.wantScoped {
				t.Errorf("ScopedCount = %d, want %d", si.ScopedCount, tt.wantScoped)
			}
			if si.ScopeCount != tt.wantScopes {
				t.Errorf("ScopeCount = %d, want %d", si.ScopeCount, tt.wantScopes)
			}
		})
	}
}

func TestDomainInfoFromCatalog_NilCatalog(t *testing.T) {
	info := DomainInfoFromCatalog(nil)
	if len(info) != 0 {
		t.Errorf("nil catalog should return empty map, got %d entries", len(info))
	}
}

// ---- Scope-enriched RenderTopology tests ----

func TestRenderTopology_ScopeEnrichedFormat(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Groups: []TopologyGroup{
			{
				Name: "Core",
				Repos: []TopologyRepo{
					{Name: "monorepo", Role: "monorepo"},
					{Name: "simple-repo", Role: "library"},
				},
			},
		},
	}

	domainCounts := map[string]int{
		"monorepo":    91,
		"simple-repo": 8,
	}

	scopeInfo := map[string]ScopeInfo{
		"monorepo": {Total: 91, RootCount: 20, ScopedCount: 71, ScopeCount: 5},
		"simple-repo": {Total: 8, RootCount: 8, ScopedCount: 0, ScopeCount: 0},
	}

	result := RenderTopology(topo, domainCounts, scopeInfo)

	// Repo with scoped domains renders enriched format.
	if !strings.Contains(result, "monorepo (91 domains: 20 root, 71 scoped) -- monorepo") {
		t.Errorf("scope-enriched format missing for monorepo, got: %s", result)
	}

	// Repo without scoped domains renders original format.
	if !strings.Contains(result, "simple-repo (8 domains) -- library") {
		t.Errorf("plain format missing for simple-repo, got: %s", result)
	}
}

func TestRenderTopology_ScopeInfoZeroScoped(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Groups: []TopologyGroup{
			{
				Name: "Apps",
				Repos: []TopologyRepo{
					{Name: "app-x", Role: "service"},
				},
			},
		},
	}

	domainCounts := map[string]int{"app-x": 5}
	scopeInfo := map[string]ScopeInfo{
		"app-x": {Total: 5, RootCount: 5, ScopedCount: 0, ScopeCount: 0},
	}

	result := RenderTopology(topo, domainCounts, scopeInfo)

	// ScopeInfo with zero ScopedCount should render original format.
	if !strings.Contains(result, "app-x (5 domains) -- service") {
		t.Errorf("expected plain format for zero-scoped repo, got: %s", result)
	}
	if strings.Contains(result, "root") || strings.Contains(result, "scoped") {
		t.Error("should not contain scope details when ScopedCount is 0")
	}
}

func TestRenderTopology_NilScopeInfo(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Groups: []TopologyGroup{
			{
				Name: "All",
				Repos: []TopologyRepo{
					{Name: "repo-a", Role: "backend"},
				},
			},
		},
	}

	domainCounts := map[string]int{"repo-a": 12}

	// nil scopeInfo should render original format (backward compat).
	result := RenderTopology(topo, domainCounts, nil)

	if !strings.Contains(result, "repo-a (12 domains) -- backend") {
		t.Errorf("nil scopeInfo should render plain format, got: %s", result)
	}
}

func TestRenderTopology_ScopeEnrichedUncategorized(t *testing.T) {
	topo := &TopologyConfig{
		SchemaVersion: "1.0",
		Org:           "testorg",
		Groups: []TopologyGroup{
			{
				Name: "Core",
				Repos: []TopologyRepo{
					{Name: "core-api", Role: "API server"},
				},
			},
		},
	}

	domainCounts := map[string]int{
		"core-api":     10,
		"scoped-extra": 15,
	}

	scopeInfo := map[string]ScopeInfo{
		"core-api":     {Total: 10, RootCount: 10, ScopedCount: 0, ScopeCount: 0},
		"scoped-extra": {Total: 15, RootCount: 5, ScopedCount: 10, ScopeCount: 3},
	}

	result := RenderTopology(topo, domainCounts, scopeInfo)

	// Uncategorized repo with scopes should also render enriched format.
	if !strings.Contains(result, "scoped-extra (15 domains: 5 root, 10 scoped)") {
		t.Errorf("uncategorized repo should use scope-enriched format, got: %s", result)
	}
}

// ---- TopologyPath tests ----

func TestTopologyPath(t *testing.T) {
	ctx := &mockOrgContext{registryDir: "/tmp/registry/autom8y"}
	got := TopologyPath(ctx)
	want := "/tmp/registry/autom8y/topology.yaml"
	if got != want {
		t.Errorf("TopologyPath() = %q, want %q", got, want)
	}
}
