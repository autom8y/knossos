// Package org implements the org-level knowledge registry for cross-repo domain discovery.
// It catalogs .know/ domains from GitHub-hosted repositories and persists them in a
// domains.yaml file under $XDG_DATA_HOME/knossos/registry/{org}/.
//
// This package is distinct from the LEAF package at internal/registry/ (which maps
// stable CLI/agent reference keys). This package is the Clew knowledge address space.
package org

import (
	"fmt"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/config"
)

// OrgContext is the interface consumed by registry operations.
// Matches config.OrgContext — using a local alias lets tests inject mocks
// without importing internal/config (which would create a layer dependency).
type OrgContext interface {
	Name() string
	RegistryDir() string
	DataDir() string
	Repos() []config.RepoConfig
}

// DomainEntry records a single .know/ domain discovered from a repo.
type DomainEntry struct {
	// QualifiedName is the canonical cross-repo address: "org::repo::domain".
	QualifiedName string `yaml:"qualified_name"`
	// Domain is the bare domain name as found in the .know/ file's frontmatter.
	Domain string `yaml:"domain"`
	// Path is the file path within the repo (e.g., ".know/architecture.md").
	Path string `yaml:"path"`
	// GeneratedAt is the RFC3339 timestamp from the domain file's frontmatter.
	GeneratedAt string `yaml:"generated_at"`
	// ExpiresAfter is the duration string from the domain file's frontmatter.
	ExpiresAfter string `yaml:"expires_after"`
	// SourceHash is the git short SHA recorded in the domain file's frontmatter.
	SourceHash string `yaml:"source_hash"`
	// Confidence is the 0.0–1.0 freshness confidence from the frontmatter.
	Confidence float64 `yaml:"confidence"`
	// FormatVersion is the schema version declared in the frontmatter.
	FormatVersion string `yaml:"format_version"`
}

// IsStale returns true if the domain's generated_at + expires_after is in the past.
// Returns true (stale) when either timestamp cannot be parsed — fail-safe default.
func (d DomainEntry) IsStale() bool {
	if d.GeneratedAt == "" || d.ExpiresAfter == "" {
		return true
	}

	generatedAt, err := time.Parse(time.RFC3339, d.GeneratedAt)
	if err != nil {
		return true
	}

	// Parse expires_after — supports "Nd" (days) and standard Go durations.
	var duration time.Duration
	if strings.HasSuffix(d.ExpiresAfter, "d") {
		dayStr := d.ExpiresAfter[:len(d.ExpiresAfter)-1]
		var days int
		if _, err := fmt.Sscanf(dayStr, "%d", &days); err != nil || days < 0 {
			return true
		}
		duration = time.Duration(days) * 24 * time.Hour
	} else {
		var err error
		duration, err = time.ParseDuration(d.ExpiresAfter)
		if err != nil {
			return true
		}
	}

	return !time.Now().UTC().Before(generatedAt.Add(duration))
}

// RepoEntry records all known domains for a single repository.
type RepoEntry struct {
	// Name is the repository name (without org prefix).
	Name string `yaml:"name"`
	// URL is the clone/web URL for the repository.
	URL string `yaml:"url"`
	// DefaultBranch is the default branch name (e.g., "main").
	DefaultBranch string `yaml:"default_branch"`
	// LastSynced is the RFC3339 timestamp of the last successful sync for this repo.
	LastSynced string `yaml:"last_synced"`
	// Domains is the list of .know/ domains discovered in this repo.
	Domains []DomainEntry `yaml:"domains"`
}

// DomainCatalog is the root catalog persisted at domains.yaml.
type DomainCatalog struct {
	// SchemaVersion identifies the catalog schema for forward-compat checks.
	SchemaVersion string `yaml:"schema_version"`
	// Org is the organization name this catalog belongs to.
	Org string `yaml:"org"`
	// SyncedAt is the RFC3339 timestamp of the last full sync operation.
	SyncedAt string `yaml:"synced_at"`
	// Repos contains one entry per discovered repository.
	Repos []RepoEntry `yaml:"repos"`
}

const catalogSchemaVersion = "1.0"

// NewCatalog creates an empty DomainCatalog for the given org.
// The SyncedAt and repo entries are populated by SyncRegistry.
func NewCatalog(ctx OrgContext) *DomainCatalog {
	return &DomainCatalog{
		SchemaVersion: catalogSchemaVersion,
		Org:           ctx.Name(),
		SyncedAt:      "",
		Repos:         nil,
	}
}

// ListDomains returns all domain entries across all repos in the catalog.
func (c *DomainCatalog) ListDomains() []DomainEntry {
	var all []DomainEntry
	for _, repo := range c.Repos {
		all = append(all, repo.Domains...)
	}
	return all
}

// LookupDomain finds a domain entry by its qualified name.
// Returns the entry and true if found; zero value and false otherwise.
func (c *DomainCatalog) LookupDomain(qualifiedName string) (DomainEntry, bool) {
	for _, repo := range c.Repos {
		for _, d := range repo.Domains {
			if d.QualifiedName == qualifiedName {
				return d, true
			}
		}
	}
	return DomainEntry{}, false
}

// DomainCount returns the total number of domain entries across all repos.
func (c *DomainCatalog) DomainCount() int {
	count := 0
	for _, repo := range c.Repos {
		count += len(repo.Domains)
	}
	return count
}

// RepoCount returns the number of repos in the catalog.
func (c *DomainCatalog) RepoCount() int {
	return len(c.Repos)
}

// StaleCount returns the number of domain entries that are stale.
func (c *DomainCatalog) StaleCount() int {
	count := 0
	for _, repo := range c.Repos {
		for _, d := range repo.Domains {
			if d.IsStale() {
				count++
			}
		}
	}
	return count
}
