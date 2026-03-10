// Package procession resolves and renders procession templates into mena
// artifacts during ari sync. Each procession template (e.g., security-remediation.yaml)
// produces a named dromena (command) and a per-procession legomena (skill).
package procession

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/procession"
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
// Resolution order (lowest to highest priority):
//  1. Embedded FS (compiled-in fallback)
//  2. Platform (platformDir)
//  3. Org (orgDir)
//  4. User (userDir)
//  5. Project (projectDir)
func ResolveProcessionsWithDirs(projectDir, userDir, orgDir, platformDir string, embeddedFS fs.FS) ([]ResolvedProcession, error) {
	resolved := make(map[string]ResolvedProcession)

	// 1. Embedded fallback (lowest priority)
	if embeddedFS != nil {
		collectFromFS(embeddedFS, "processions", "embedded", resolved)
	}

	// 2. Platform
	if platformDir != "" {
		collectFromDisk(platformDir, "platform", resolved)
	}

	// 3. Org
	if orgDir != "" {
		collectFromDisk(orgDir, "org", resolved)
	}

	// 4. User
	if userDir != "" {
		collectFromDisk(userDir, "user", resolved)
	}

	// 5. Project (highest priority)
	if projectDir != "" {
		collectFromDisk(projectDir, "project", resolved)
	}

	// Convert map to sorted slice
	result := make([]ResolvedProcession, 0, len(resolved))
	for _, rp := range resolved {
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

// collectFromDisk scans a directory for *.yaml procession templates.
// Invalid templates are silently skipped (logged at call site if needed).
func collectFromDisk(dir, source string, resolved map[string]ResolvedProcession) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // Directory doesn't exist or unreadable — skip
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		if entry.Name() == ".gitkeep" {
			continue
		}

		tmpl, err := procession.LoadTemplate(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue // Invalid template — skip silently
		}

		resolved[tmpl.Name] = ResolvedProcession{
			Name:     tmpl.Name,
			Source:   source,
			Template: tmpl,
		}
	}
}

// collectFromFS scans an fs.FS for *.yaml procession templates.
func collectFromFS(fsys fs.FS, dir, source string, resolved map[string]ResolvedProcession) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return // Directory doesn't exist in FS — skip
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := dir + "/" + entry.Name()
		tmpl, err := procession.LoadTemplateFromFS(fsys, path)
		if err != nil {
			continue // Invalid template — skip silently
		}

		resolved[tmpl.Name] = ResolvedProcession{
			Name:     tmpl.Name,
			Source:   source,
			Template: tmpl,
		}
	}
}
