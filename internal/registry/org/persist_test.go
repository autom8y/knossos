package org

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveCatalog_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "domains.yaml")

	now := time.Now().UTC().Format(time.RFC3339)
	original := &DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		SyncedAt:      now,
		Repos: []RepoEntry{
			{
				Name:          "knossos",
				URL:           "https://github.com/autom8y/knossos",
				DefaultBranch: "main",
				LastSynced:    now,
				Domains: []DomainEntry{
					{
						QualifiedName: "autom8y::knossos::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
						GeneratedAt:   now,
						ExpiresAfter:  "7d",
						SourceHash:    "abc123",
						Confidence:    0.92,
						FormatVersion: "1.0",
					},
				},
			},
		},
	}

	if err := SaveCatalog(path, original); err != nil {
		t.Fatalf("SaveCatalog error: %v", err)
	}

	// File should exist
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("catalog file not written: %v", err)
	}

	// Round-trip: load should match original
	loaded, err := LoadCatalog(path)
	if err != nil {
		t.Fatalf("LoadCatalog error: %v", err)
	}

	if loaded.SchemaVersion != original.SchemaVersion {
		t.Errorf("SchemaVersion: got %q, want %q", loaded.SchemaVersion, original.SchemaVersion)
	}
	if loaded.Org != original.Org {
		t.Errorf("Org: got %q, want %q", loaded.Org, original.Org)
	}
	if loaded.RepoCount() != original.RepoCount() {
		t.Errorf("RepoCount: got %d, want %d", loaded.RepoCount(), original.RepoCount())
	}
	if loaded.DomainCount() != original.DomainCount() {
		t.Errorf("DomainCount: got %d, want %d", loaded.DomainCount(), original.DomainCount())
	}

	d, ok := loaded.LookupDomain("autom8y::knossos::architecture")
	if !ok {
		t.Fatal("LookupDomain failed after round-trip")
	}
	if d.Confidence != 0.92 {
		t.Errorf("Confidence: got %f, want 0.92", d.Confidence)
	}
}

func TestLoadCatalog_NotFound(t *testing.T) {
	_, err := LoadCatalog("/tmp/nonexistent-knossos-registry-catalog.yaml")
	if err == nil {
		t.Error("LoadCatalog should return error for missing file")
	}
}

func TestSaveCatalog_CreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	// Use a nested path that doesn't exist yet
	path := filepath.Join(tmpDir, "deep", "nested", "dir", "domains.yaml")

	catalog := &DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "test",
		SyncedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	if err := SaveCatalog(path, catalog); err != nil {
		t.Fatalf("SaveCatalog with nested path error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("catalog file should exist at %s, got: %v", path, err)
	}
}

func TestCatalogPath(t *testing.T) {
	ctx := &mockOrgContext{registryDir: "/tmp/registry/autom8y"}
	got := CatalogPath(ctx)
	want := "/tmp/registry/autom8y/domains.yaml"
	if got != want {
		t.Errorf("CatalogPath() = %q, want %q", got, want)
	}
}
