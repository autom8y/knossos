package materialize

import (
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/provenance"
)

// materializeMena copies mena files to .claude/commands/ or .claude/skills/
// based on the filename convention (.dro.md for dromena, .lego.md for legomena).
// Sources: mena/, rites/{rite}/mena/, rites/shared/mena/
// Priority order (later sources override earlier): mena < shared < dependencies < current rite
//
// This method builds the source list and delegates to SyncMena() for the
// actual collection, routing, extension stripping, and file copying.
func (m *Materializer) materializeMena(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite, collector provenance.Collector, overwriteDiverged bool) error {
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	isEmbedded := resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.EmbeddedFS != nil

	// Build priority-ordered source list (later sources override earlier)
	var sources []MenaSource

	// 1. Platform-level mena (lowest priority, can be overridden by rite-specific)
	// Resolution order: project-level mena/ → XDG data dir → KnossosHome → embedded fallback.
	// Platform mena IS the product — operations, guidance, session skills ship to all users.
	if menaDir := m.getMenaDir(); menaDir != "" {
		sources = append(sources, MenaSource{Path: menaDir})
	} else if isEmbedded && m.embeddedMena != nil {
		// No filesystem mena found; fall back to embedded platform mena.
		// This is the expected path for users who installed via `go install`
		// and don't have the knossos source tree.
		sources = append(sources, MenaSource{Fsys: m.embeddedMena, FsysPath: "mena", IsEmbedded: true})
	}

	if isEmbedded {
		// For embedded sources, read mena from embedded FS
		embFS := m.sourceResolver.EmbeddedFS

		// 2. Shared rite mena
		sources = append(sources, MenaSource{Fsys: embFS, FsysPath: "rites/shared/mena", IsEmbedded: true})

		// 3. Dependency rite mena (in order)
		for _, dep := range manifest.Dependencies {
			if dep != "shared" {
				sources = append(sources, MenaSource{Fsys: embFS, FsysPath: "rites/" + dep + "/mena", IsEmbedded: true})
			}
		}

		// 4. Current rite mena (highest priority)
		sources = append(sources, MenaSource{Fsys: embFS, FsysPath: "rites/" + manifest.Name + "/mena", IsEmbedded: true})
	} else if resolved != nil {
		// Determine the base directory for shared/dependency rites.
		//
		// For knossos-core rites (SourceKnossos or SourceProject where the project IS
		// knossos), ritesBase derived from resolved.RitePath is correct — shared/ and
		// dependency rites live alongside the active rite in the same rites/ directory.
		//
		// For satellite-local rites (SourceProject where the project is a satellite),
		// resolved.RitePath points to {satellite}/rites/{name}/, so ritesBase would be
		// {satellite}/rites/ — which has no shared/ or dependency rite directories.
		// Shared and dependency rites always live in $KNOSSOS_HOME/rites/, so we must
		// resolve them from knossosHome instead of from the satellite's rites directory.
		//
		// Strategy: use KnossosHome for shared/dependency if it is set and differs from
		// the project root's rites dir. Fall back to ritesBase when they are the same
		// (knossos-core self-hosting case) so there is no regression.
		ritesBase := filepath.Dir(resolved.RitePath)

		// sharedRitesBase is the rites directory containing shared/ and dependency rites.
		// For satellite-local rites this is $KNOSSOS_HOME/rites/; for knossos-core it is
		// the same as ritesBase (derived from the active rite path).
		sharedRitesBase := ritesBase
		if knossosHome := m.sourceResolver.KnossosHome(); knossosHome != "" {
			knossosRitesDir := filepath.Join(knossosHome, "rites")
			// Use the knossos rites dir only when it differs from the satellite's ritesBase.
			// When they match (knossos syncing its own rites) keep ritesBase as-is.
			if knossosRitesDir != ritesBase {
				sharedRitesBase = knossosRitesDir
			}
		}

		// 2. Shared rite mena (resolved from knossos-core, not from satellite rites dir)
		sharedMenaDir := filepath.Join(sharedRitesBase, "shared", "mena")
		sources = append(sources, MenaSource{Path: sharedMenaDir})

		// 3. Dependency rite mena (dependencies are knossos-core rites, same base)
		for _, dep := range manifest.Dependencies {
			if dep != "shared" {
				sources = append(sources, MenaSource{Path: filepath.Join(sharedRitesBase, dep, "mena")})
			}
		}

		// 4. Current rite mena (highest priority — always from the resolved rite path)
		currentRiteMenaDir := filepath.Join(resolved.RitePath, "mena")
		sources = append(sources, MenaSource{Path: currentRiteMenaDir})
	}

	// Delegate to SyncMena with destructive mode and provenance collector
	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		Collector:         collector,
		ProjectRoot:       m.resolver.ProjectRoot(),
		OverwriteDiverged: overwriteDiverged,
		RiteName:          manifest.Name,
	}

	_, err := SyncMena(sources, opts)
	return err
}

// getMenaDir returns the mena directory path.
// Resolution order: project-level → XDG data dir → KnossosHome.
// Returns "" if none found (caller should try embedded fallback).
func (m *Materializer) getMenaDir() string {
	// 1. Check for project-level mena first (.knossos/mena/ satellite overrides)
	projectMena := filepath.Join(m.resolver.ProjectRoot(), ".knossos", "mena")
	if _, err := os.Stat(projectMena); err == nil {
		return projectMena
	}

	// 2. Check XDG data dir (installed user case)
	if xdgMena := xdgMenaPath(); xdgMena != "" {
		if _, err := os.Stat(xdgMena); err == nil {
			return xdgMena
		}
	}

	// 3. Fall back to Knossos platform mena (developer case)
	if m.sourceResolver.KnossosHome() != "" {
		knossosMena := filepath.Join(m.sourceResolver.KnossosHome(), "mena")
		if _, err := os.Stat(knossosMena); err == nil {
			return knossosMena
		}
	}

	return ""
}

// xdgMenaPath returns the XDG data directory path for platform mena.
func xdgMenaPath() string {
	return filepath.Join(config.XDGDataDir(), "mena")
}
