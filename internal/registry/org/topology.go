package org

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// topologyFileName is the canonical name for the topology.yaml config file.
const topologyFileName = "topology.yaml"

// TopologyConfig is the YAML schema for deploy/registry/topology.yaml.
type TopologyConfig struct {
	SchemaVersion string          `yaml:"schema_version"`
	Org           string          `yaml:"org"`
	Groups        []TopologyGroup `yaml:"groups"`
	Edges         []TopologyEdge  `yaml:"edges"`
}

// TopologyGroup is a named grouping of repos (e.g., "Service layer", "Tooling").
type TopologyGroup struct {
	Name  string         `yaml:"name"`
	Repos []TopologyRepo `yaml:"repos"`
}

// TopologyRepo is a repo entry within a topology group.
type TopologyRepo struct {
	Name      string `yaml:"name"`
	Role      string `yaml:"role"`
	Direction string `yaml:"direction,omitempty"` // "upstream" or empty
}

// TopologyEdge is a directional dependency between two repos.
type TopologyEdge struct {
	From  string `yaml:"from"`
	To    string `yaml:"to"`
	Label string `yaml:"label"`
}

// TopologyPath returns the absolute path for the topology file given an org context.
func TopologyPath(ctx OrgContext) string {
	return filepath.Join(ctx.RegistryDir(), topologyFileName)
}

// LoadTopology reads and parses a TopologyConfig from path.
// Returns nil and no error if the file is missing (fail-open).
// Returns an error only if the file exists but cannot be parsed.
func LoadTopology(path string) (*TopologyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // fail-open: missing file is not an error
		}
		return nil, fmt.Errorf("read topology %s: %w", path, err)
	}

	var topo TopologyConfig
	if err := yaml.Unmarshal(data, &topo); err != nil {
		return nil, fmt.Errorf("parse topology %s: %w", path, err)
	}

	return &topo, nil
}

// RenderTopology produces the plain-text topology section for the system prompt.
// domainCounts maps repo name to the number of knowledge domains cataloged.
// scopeInfo optionally provides scope-enriched domain breakdowns per repo;
// repos with scoped domains render as "N domains: X root, Y scoped" instead
// of the plain "N domains" format. Pass nil for backward-compatible behavior.
// Returns an empty string if topo is nil.
func RenderTopology(topo *TopologyConfig, domainCounts map[string]int, scopeInfo map[string]ScopeInfo) string {
	if topo == nil {
		return ""
	}

	var b strings.Builder

	// Compute totals.
	totalRepos := 0
	totalDomains := 0
	topoRepoSet := make(map[string]bool)
	for _, g := range topo.Groups {
		for _, r := range g.Repos {
			totalRepos++
			topoRepoSet[r.Name] = true
			totalDomains += domainCounts[r.Name]
		}
	}

	// Count uncategorized repos (in domainCounts but not in topology groups).
	var uncategorized []string
	for repoName := range domainCounts {
		if !topoRepoSet[repoName] {
			uncategorized = append(uncategorized, repoName)
			totalRepos++
			totalDomains += domainCounts[repoName]
		}
	}

	b.WriteString("--- ORG TOPOLOGY ---\n\n")
	b.WriteString(fmt.Sprintf("Organization: %s (%d repos, ~%d knowledge domains)\n",
		topo.Org, totalRepos, totalDomains))

	// Build edge lookups: inbound edges (to -> []edge) and outbound edges (from -> []edge).
	inbound := make(map[string][]TopologyEdge)
	outbound := make(map[string][]TopologyEdge)
	for _, e := range topo.Edges {
		inbound[e.To] = append(inbound[e.To], e)
		outbound[e.From] = append(outbound[e.From], e)
	}

	// Render each group.
	for _, g := range topo.Groups {
		b.WriteString(fmt.Sprintf("\n%s:\n", g.Name))
		for _, r := range g.Repos {
			b.WriteString(fmt.Sprintf("  %s", r.Name))
			renderDomainCount(&b, r.Name, domainCounts, scopeInfo)
			b.WriteString(fmt.Sprintf(" -- %s\n", r.Role))

			// Inbound edges: other repos that depend on this repo.
			for _, e := range inbound[r.Name] {
				fromDC := domainCounts[e.From]
				b.WriteString(fmt.Sprintf("    <- %s (%d domains) %s\n", e.From, fromDC, e.Label))
			}

			// Outbound edges: repos this repo depends on or triggers.
			for _, e := range outbound[r.Name] {
				toDC := domainCounts[e.To]
				b.WriteString(fmt.Sprintf("    -> %s (%d domains) %s\n", e.To, toDC, e.Label))
			}
		}
	}

	// Render uncategorized repos (in catalog but not in topology.yaml).
	if len(uncategorized) > 0 {
		// Sort for deterministic output.
		sortStrings(uncategorized)
		b.WriteString("\nOther:\n")
		for _, name := range uncategorized {
			b.WriteString(fmt.Sprintf("  %s", name))
			renderDomainCount(&b, name, domainCounts, scopeInfo)
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderDomainCount writes the domain count annotation for a repo.
// Uses scope-enriched format when the repo has scoped domains, otherwise
// falls back to the plain "(N domains)" format.
func renderDomainCount(b *strings.Builder, repoName string, domainCounts map[string]int, scopeInfo map[string]ScopeInfo) {
	if si, ok := scopeInfo[repoName]; ok && si.ScopedCount > 0 {
		b.WriteString(fmt.Sprintf(" (%d domains: %d root, %d scoped)", si.Total, si.RootCount, si.ScopedCount))
	} else {
		dc := domainCounts[repoName]
		b.WriteString(fmt.Sprintf(" (%d domains)", dc))
	}
}

// sortStrings sorts a string slice in place. Avoids importing "sort" for a
// single use -- uses insertion sort which is optimal for the small slices
// expected here (typically 0-3 uncategorized repos).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// DomainCountsFromCatalog builds a map of repo name -> domain count from a DomainCatalog.
func DomainCountsFromCatalog(catalog *DomainCatalog) map[string]int {
	counts := make(map[string]int)
	if catalog == nil {
		return counts
	}
	for _, repo := range catalog.Repos {
		counts[repo.Name] = len(repo.Domains)
	}
	return counts
}

// ScopeInfo holds domain count breakdown for a repo with scope awareness.
type ScopeInfo struct {
	Total       int // Total domain count
	RootCount   int // Domains with empty scope
	ScopedCount int // Domains with non-empty scope
	ScopeCount  int // Number of distinct scopes
}

// DomainInfoFromCatalog builds a map of repo name -> ScopeInfo from a DomainCatalog.
func DomainInfoFromCatalog(catalog *DomainCatalog) map[string]ScopeInfo {
	info := make(map[string]ScopeInfo)
	if catalog == nil {
		return info
	}
	for _, repo := range catalog.Repos {
		si := ScopeInfo{Total: len(repo.Domains)}
		scopes := make(map[string]bool)
		for _, d := range repo.Domains {
			if d.Scope == "" {
				si.RootCount++
			} else {
				si.ScopedCount++
				scopes[d.Scope] = true
			}
		}
		si.ScopeCount = len(scopes)
		info[repo.Name] = si
	}
	return info
}
