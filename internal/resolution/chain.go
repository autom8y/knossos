// Package resolution provides a unified multi-tier resolution chain for
// discovering resources (rites, processions, contexts) across the knossos
// resolution hierarchy: project > user > org > platform > embedded.
//
// This package has ZERO internal imports. All tier paths are injected via
// constructor to avoid import cycles (TENSION-005).
package resolution

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Tier represents a single tier in the resolution chain.
type Tier struct {
	Label string // Human-readable label: "project", "user", "org", "platform", "embedded"
	Dir   string // Filesystem directory path. Empty if FS-only.
	FS    fs.FS  // Optional virtual filesystem (nil for disk-only tiers).
}

// ResolvedItem represents a discovered entity from the resolution chain.
type ResolvedItem struct {
	Name   string // Entity name (directory name or file stem)
	Path   string // Full filesystem path, or FS-relative path for FS tiers
	Source string // Tier label that provided this item
	Fsys   fs.FS  // Non-nil if resolved from an FS tier
}

// Chain is an ordered list of resolution tiers, highest priority first.
// Use Resolve for single-item lookup (top-down early-exit) and ResolveAll
// for collecting all items with higher-priority shadowing.
type Chain struct {
	tiers []Tier
}

// NewChain creates a resolution chain from the given tiers.
// Tiers must be ordered highest-priority first (e.g., project before embedded).
// Tiers with empty Dir and nil FS are silently skipped.
func NewChain(tiers ...Tier) *Chain {
	filtered := make([]Tier, 0, len(tiers))
	for _, t := range tiers {
		if t.Dir != "" || t.FS != nil {
			filtered = append(filtered, t)
		}
	}
	return &Chain{tiers: filtered}
}

// Tiers returns the ordered tier list for introspection.
func (c *Chain) Tiers() []Tier {
	out := make([]Tier, len(c.tiers))
	copy(out, c.tiers)
	return out
}

// Resolve finds the highest-priority item matching name.
// It scans tiers top-down and returns the first item for which validate
// returns true. Returns an error if no tier contains a valid match.
//
// For disk tiers, the candidate path is filepath.Join(tier.Dir, name).
// For FS tiers, the candidate path is path.Join(tier.Dir, name) within the FS.
// The validate callback receives the candidate item and should check whether
// the path contains the expected content (e.g., manifest.yaml exists).
func (c *Chain) Resolve(name string, validate func(ResolvedItem) bool) (*ResolvedItem, error) {
	for _, tier := range c.tiers {
		item := ResolvedItem{
			Name:   name,
			Source: tier.Label,
		}

		if tier.FS != nil {
			item.Path = filepath.Join(tier.Dir, name)
			item.Fsys = tier.FS
		} else {
			item.Path = filepath.Join(tier.Dir, name)
		}

		if validate(item) {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("not found in any resolution tier: %s", name)
}

// ResolveAll collects all named entities across all tiers.
// Higher-priority tiers shadow lower-priority ones by entity name.
// The validate callback filters which directory entries are valid entities.
//
// For each tier, ResolveAll enumerates entries in the tier's directory.
// For disk tiers it reads the directory; for FS tiers it reads from the FS.
// Each entry that passes validate is included, with higher-priority tiers
// overwriting lower-priority ones sharing the same name.
func (c *Chain) ResolveAll(validate func(ResolvedItem) bool) (map[string]ResolvedItem, error) {
	result := make(map[string]ResolvedItem)

	// Iterate bottom-up (lowest priority first) so higher tiers overwrite.
	for i := len(c.tiers) - 1; i >= 0; i-- {
		tier := c.tiers[i]
		entries := c.listEntries(tier)

		for _, name := range entries {
			item := ResolvedItem{
				Name:   name,
				Source: tier.Label,
			}

			if tier.FS != nil {
				item.Path = filepath.Join(tier.Dir, name)
				item.Fsys = tier.FS
			} else {
				item.Path = filepath.Join(tier.Dir, name)
			}

			if validate(item) {
				result[name] = item
			}
		}
	}

	return result, nil
}

// listEntries enumerates entry names from a tier's directory.
func (c *Chain) listEntries(tier Tier) []string {
	if tier.FS != nil {
		return listEntriesFS(tier.FS, tier.Dir)
	}
	return listEntriesDisk(tier.Dir)
}

// listEntriesDisk reads directory entries from the filesystem.
func listEntriesDisk(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// listEntriesFS reads directory entries from an fs.FS.
func listEntriesFS(fsys fs.FS, dir string) []string {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
