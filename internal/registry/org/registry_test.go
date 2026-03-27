package org

import (
	"testing"
	"time"
)

// catalogForTest builds a DomainCatalog with known entries for assertions.
func catalogForTest() *DomainCatalog {
	now := time.Now().UTC()
	fresh := now.Format(time.RFC3339)
	staleTime := now.Add(-30 * 24 * time.Hour).Format(time.RFC3339)

	return &DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		SyncedAt:      fresh,
		Repos: []RepoEntry{
			{
				Name:          "knossos",
				URL:           "https://github.com/autom8y/knossos",
				DefaultBranch: "main",
				LastSynced:    fresh,
				Domains: []DomainEntry{
					{
						QualifiedName: "autom8y::knossos::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
						GeneratedAt:   fresh,
						ExpiresAfter:  "7d",
						SourceHash:    "abc123",
						Confidence:    0.92,
						FormatVersion: "1.0",
					},
					{
						QualifiedName: "autom8y::knossos::feat/materialization",
						Domain:        "feat/materialization",
						Path:          ".know/feat/materialization.md",
						GeneratedAt:   staleTime,
						ExpiresAfter:  "7d",
						SourceHash:    "def456",
						Confidence:    0.80,
						FormatVersion: "1.0",
					},
				},
			},
			{
				Name:          "payments",
				URL:           "https://github.com/autom8y/payments",
				DefaultBranch: "main",
				LastSynced:    fresh,
				Domains: []DomainEntry{
					{
						QualifiedName: "autom8y::payments::conventions",
						Domain:        "conventions",
						Path:          ".know/conventions.md",
						GeneratedAt:   fresh,
						ExpiresAfter:  "14d",
						SourceHash:    "ghi789",
						Confidence:    0.88,
						FormatVersion: "1.0",
					},
				},
			},
		},
	}
}

func TestNewCatalog(t *testing.T) {
	ctx := &mockOrgContext{name: "autom8y", registryDir: "/tmp/registry", dataDir: "/tmp/data"}
	c := NewCatalog(ctx)

	if c.Org != "autom8y" {
		t.Errorf("Org = %q, want %q", c.Org, "autom8y")
	}
	if c.SchemaVersion != "1.1" {
		t.Errorf("SchemaVersion = %q, want %q", c.SchemaVersion, "1.1")
	}
	if len(c.Repos) != 0 {
		t.Errorf("Repos should be empty, got %d", len(c.Repos))
	}
}

func TestDomainCount(t *testing.T) {
	c := catalogForTest()
	if got := c.DomainCount(); got != 3 {
		t.Errorf("DomainCount() = %d, want 3", got)
	}
}

func TestRepoCount(t *testing.T) {
	c := catalogForTest()
	if got := c.RepoCount(); got != 2 {
		t.Errorf("RepoCount() = %d, want 2", got)
	}
}

func TestStaleCount(t *testing.T) {
	c := catalogForTest()
	// feat/materialization was generated 30 days ago with 7d expiry -> stale
	if got := c.StaleCount(); got != 1 {
		t.Errorf("StaleCount() = %d, want 1", got)
	}
}

func TestListDomains(t *testing.T) {
	c := catalogForTest()
	domains := c.ListDomains()
	if len(domains) != 3 {
		t.Fatalf("ListDomains() count = %d, want 3", len(domains))
	}
}

func TestLookupDomain_Found(t *testing.T) {
	c := catalogForTest()
	d, ok := c.LookupDomain("autom8y::knossos::architecture")
	if !ok {
		t.Fatal("LookupDomain() returned false for existing domain")
	}
	if d.Domain != "architecture" {
		t.Errorf("Domain = %q, want %q", d.Domain, "architecture")
	}
}

func TestLookupDomain_NotFound(t *testing.T) {
	c := catalogForTest()
	_, ok := c.LookupDomain("autom8y::unknown::nope")
	if ok {
		t.Error("LookupDomain() returned true for non-existent domain")
	}
}

func TestDomainEntry_IsStale(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		entry     DomainEntry
		wantStale bool
	}{
		{
			name: "fresh domain",
			entry: DomainEntry{
				GeneratedAt:  now.Format(time.RFC3339),
				ExpiresAfter: "7d",
			},
			wantStale: false,
		},
		{
			name: "expired domain",
			entry: DomainEntry{
				GeneratedAt:  now.Add(-8 * 24 * time.Hour).Format(time.RFC3339),
				ExpiresAfter: "7d",
			},
			wantStale: true,
		},
		{
			name:      "empty GeneratedAt",
			entry:     DomainEntry{ExpiresAfter: "7d"},
			wantStale: true,
		},
		{
			name:      "empty ExpiresAfter",
			entry:     DomainEntry{GeneratedAt: now.Format(time.RFC3339)},
			wantStale: true,
		},
		{
			name:      "both empty",
			entry:     DomainEntry{},
			wantStale: true,
		},
		{
			name: "invalid GeneratedAt",
			entry: DomainEntry{
				GeneratedAt:  "not-a-date",
				ExpiresAfter: "7d",
			},
			wantStale: true,
		},
		{
			name: "hours duration",
			entry: DomainEntry{
				GeneratedAt:  now.Add(-1 * time.Hour).Format(time.RFC3339),
				ExpiresAfter: "2h",
			},
			wantStale: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.entry.IsStale()
			if got != tc.wantStale {
				t.Errorf("IsStale() = %v, want %v", got, tc.wantStale)
			}
		})
	}
}
