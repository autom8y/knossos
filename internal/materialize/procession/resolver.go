// Package procession resolves and renders procession templates into mena
// artifacts during ari sync. Each procession template (e.g., security-remediation.yaml)
// produces a named dromena (command) and a per-procession legomena (skill).
package procession

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/procession"
	"github.com/autom8y/knossos/internal/resolution"
)

// ResolvedProcession represents a procession template discovered through
// the multi-tier resolution chain.
type ResolvedProcession struct {
	Name     string               // Template name (e.g., "security-remediation")
	Source   string               // Resolution tier: "project", "user", "org", "platform", "embedded"
	Template *procession.Template // Loaded and validated template
}

// ResolveProcessions discovers procession templates through a cascading
// resolution chain using global config for tier paths. Higher-priority
// tiers shadow lower-priority ones by template name.
//
// Resolution order (lowest to highest priority):
//  1. Embedded FS (compiled-in fallback)
//  2. Platform ($KNOSSOS_HOME/processions/)
//  3. Org ($XDG_DATA_HOME/knossos/orgs/{org}/processions/)
//  4. User (~/.local/share/knossos/processions/)
//  5. Project ({projectRoot}/processions/)
func ResolveProcessions(projectRoot string, embeddedFS fs.FS) ([]ResolvedProcession, error) {
	platformDir := ""
	if knossosHome := config.KnossosHome(); knossosHome != "" {
		platformDir = filepath.Join(knossosHome, "processions")
	}
	orgDir := ""
	if orgName := config.ActiveOrg(); orgName != "" {
		orgDir = filepath.Join(paths.OrgDataDir(orgName), "processions")
	}
	userDir := filepath.Join(paths.DataDir(), "processions")
	projectDir := ""
	if projectRoot != "" {
		projectDir = filepath.Join(projectRoot, "processions")
	}
	return ResolveProcessionsWithDirs(projectDir, userDir, orgDir, platformDir, embeddedFS)
}

// ResolveProcessionsWithDirs discovers procession templates using explicit
// tier directory paths. Empty directories are skipped. This enables test
// injection without global state mutation.
//
// Delegates to resolution.ProcessionChain for multi-tier resolution.
// Resolution order (highest to lowest priority):
//  1. Project (projectDir)
//  2. User (userDir)
//  3. Org (orgDir)
//  4. Platform (platformDir)
//  5. Embedded FS (compiled-in fallback)
func ResolveProcessionsWithDirs(projectDir, userDir, orgDir, platformDir string, embeddedFS fs.FS) ([]ResolvedProcession, error) {
	chain := resolution.ProcessionChain(projectDir, userDir, orgDir, platformDir, embeddedFS)
	items, err := chain.ResolveAll(yamlFileValidator)
	if err != nil {
		return nil, err
	}

	// Load templates from resolved items. Chain.ResolveAll already
	// shadows by entry name (higher-priority tiers overwrite); we
	// re-key by template name for the rare case where filenames
	// differ across tiers but template names collide.
	byName := make(map[string]ResolvedProcession, len(items))
	for _, item := range items {
		tmpl, err := loadProcessionTemplate(item)
		if err != nil {
			continue // Invalid template — skip silently
		}
		byName[tmpl.Name] = ResolvedProcession{
			Name:     tmpl.Name,
			Source:   item.Source,
			Template: tmpl,
		}
	}

	result := make([]ResolvedProcession, 0, len(byName))
	for _, rp := range byName {
		result = append(result, rp)
	}
	return result, nil
}

// ResolveTemplate finds a single procession template by name through the
// cascading resolution chain. Returns the highest-priority match, or an
// error if no template with that name is found.
func ResolveTemplate(name, projectRoot string, embeddedFS fs.FS) (*ResolvedProcession, error) {
	resolved, err := ResolveProcessions(projectRoot, embeddedFS)
	if err != nil {
		return nil, err
	}
	for i := range resolved {
		if resolved[i].Name == name {
			return &resolved[i], nil
		}
	}
	return nil, fmt.Errorf("procession template %q not found in any resolution tier", name)
}

// yamlFileValidator filters resolution chain entries to valid YAML files.
func yamlFileValidator(item resolution.ResolvedItem) bool {
	return strings.HasSuffix(item.Name, ".yaml") && item.Name != ".gitkeep"
}

// loadProcessionTemplate loads a procession template from a resolved item.
// Handles both disk tiers (Fsys == nil) and embedded FS tiers.
func loadProcessionTemplate(item resolution.ResolvedItem) (*procession.Template, error) {
	if item.Fsys != nil {
		return procession.LoadTemplateFromFS(item.Fsys, item.Path)
	}
	return procession.LoadTemplate(item.Path)
}
