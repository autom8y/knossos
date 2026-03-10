// Package procession resolves and renders procession templates into mena
// artifacts during ari sync. Each procession template (e.g., security-remediation.yaml)
// produces a named dromena (command) and a per-procession legomena (skill).
package procession

import (
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
// resolution chain. Higher-priority tiers shadow lower-priority ones
// by template name.
//
// Resolution order (lowest to highest priority):
//  1. Embedded FS (compiled-in fallback)
//  2. Platform ($KNOSSOS_HOME/processions/)
//  3. Org ($XDG_DATA_HOME/knossos/orgs/{org}/processions/)
//  4. User (~/.local/share/knossos/processions/)
//  5. Project ({projectRoot}/processions/)
func ResolveProcessions(projectRoot string, embeddedFS fs.FS) ([]ResolvedProcession, error) {
	// Map of name → resolved template (later tiers override earlier).
	resolved := make(map[string]ResolvedProcession)

	// 1. Embedded fallback (lowest priority)
	if embeddedFS != nil {
		collectFromFS(embeddedFS, "processions", "embedded", resolved)
	}

	// 2. Platform ($KNOSSOS_HOME/processions/)
	if knossosHome := config.KnossosHome(); knossosHome != "" {
		collectFromDisk(filepath.Join(knossosHome, "processions"), "platform", resolved)
	}

	// 3. Org ($XDG_DATA_HOME/knossos/orgs/{org}/processions/)
	if orgName := config.ActiveOrg(); orgName != "" {
		orgDir := filepath.Join(paths.OrgDataDir(orgName), "processions")
		collectFromDisk(orgDir, "org", resolved)
	}

	// 4. User (~/.local/share/knossos/processions/)
	userDir := filepath.Join(paths.DataDir(), "processions")
	collectFromDisk(userDir, "user", resolved)

	// 5. Project (highest priority)
	if projectRoot != "" {
		collectFromDisk(filepath.Join(projectRoot, "processions"), "project", resolved)
	}

	// Convert map to sorted slice
	result := make([]ResolvedProcession, 0, len(resolved))
	for _, rp := range resolved {
		result = append(result, rp)
	}
	return result, nil
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
