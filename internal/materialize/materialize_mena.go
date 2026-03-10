package materialize

import (
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/config"
	procmena "github.com/autom8y/knossos/internal/materialize/procession"
	"github.com/autom8y/knossos/internal/provenance"
)

// materializeMena copies mena files to .claude/commands/ or .claude/skills/
// based on the filename convention (.dro.md for dromena, .lego.md for legomena).
// Sources: mena/ (platform), rites/shared/mena/ (cross-rite overlay), rites/{rite}/mena/
// Priority order (later sources override earlier): platform < shared < dependencies < current rite
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
	// Resolution order: project-level mena/ → KnossosHome → XDG data dir → embedded fallback.
	// Platform mena IS the product — operations, guidance, session skills ship to all users.
	if menaDir := m.getMenaDir(); menaDir != "" {
		sources = append(sources, MenaSource{Path: menaDir})
	} else if isEmbedded && m.embeddedMena != nil {
		// No filesystem mena found; fall back to embedded platform mena.
		// This is the expected path for users who installed via `go install`
		// and don't have the knossos source tree.
		sources = append(sources, MenaSource{Fsys: m.embeddedMena, FsysPath: "mena", IsEmbedded: true})
	}

	// 1.5. Procession-generated mena (between platform and shared)
	// Each procession template generates a named dromena (rite-filtered) and
	// per-procession skill (universal). Dromena only projected when current rite
	// matches the template's entry rite.
	if procDir, err := m.renderProcessionMena(manifest.Name); err == nil && procDir != "" {
		sources = append(sources, MenaSource{Path: procDir})
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
		KnossosDir:        m.resolver.KnossosDir(),
		OverwriteDiverged: overwriteDiverged,
		RiteName:          manifest.Name,
	}

	_, err := SyncMena(sources, opts)
	return err
}

// materializeMinimalMena projects platform mena and shared rite mena into
// .claude/commands/ and .claude/skills/ without requiring an active rite.
// Called from MaterializeMinimal so that cross-cutting mode still gets core
// features like /know, /radar, /research.
func (m *Materializer) materializeMinimalMena(claudeDir string, collector provenance.Collector, overwriteDiverged bool) error {
	commandsDir := filepath.Join(claudeDir, "commands")
	skillsDir := filepath.Join(claudeDir, "skills")

	var sources []MenaSource

	// 1. Platform-level mena (lowest priority)
	if menaDir := m.getMenaDir(); menaDir != "" {
		sources = append(sources, MenaSource{Path: menaDir})
	} else if m.embeddedMena != nil {
		sources = append(sources, MenaSource{Fsys: m.embeddedMena, FsysPath: "mena", IsEmbedded: true})
	}

	// 1.5. Procession-generated mena (legomena only in minimal mode — no active rite)
	if procDir, err := m.renderProcessionMena(""); err == nil && procDir != "" {
		sources = append(sources, MenaSource{Path: procDir})
	}

	// 2. Shared rite mena — core cross-rite features (/know, /radar, etc.)
	if m.sourceResolver.EmbeddedFS != nil {
		sources = append(sources, MenaSource{Fsys: m.sourceResolver.EmbeddedFS, FsysPath: "rites/shared/mena", IsEmbedded: true})
	} else if knossosHome := m.sourceResolver.KnossosHome(); knossosHome != "" {
		sharedMenaDir := filepath.Join(knossosHome, "rites", "shared", "mena")
		sources = append(sources, MenaSource{Path: sharedMenaDir})
	}

	if len(sources) == 0 {
		return nil // No mena sources available
	}

	opts := MenaProjectionOptions{
		Mode:              MenaProjectionDestructive,
		Filter:            ProjectAll,
		TargetCommandsDir: commandsDir,
		TargetSkillsDir:   skillsDir,
		Collector:         collector,
		ProjectRoot:       m.resolver.ProjectRoot(),
		KnossosDir:        m.resolver.KnossosDir(),
		OverwriteDiverged: overwriteDiverged,
	}

	_, err := SyncMena(sources, opts)
	return err
}

// renderProcessionMena resolves procession templates and renders their
// dromena and legomena into a deterministic directory for mena projection.
// Uses .knossos/procession-mena/ to ensure provenance source paths are stable
// across syncs (idempotency invariant).
//
// currentRite controls dromena filtering:
//   - Non-empty: only render dromena for templates whose entry rite matches
//   - Empty: render legomena only (minimal/cross-cutting mode)
//
// Returns empty string if no templates found or rendering fails (fail-open).
func (m *Materializer) renderProcessionMena(currentRite string) (string, error) {
	procDir := filepath.Join(m.resolver.KnossosDir(), "procession-mena")

	// Clean any prior render to avoid stale entries
	os.RemoveAll(procDir)

	if err := os.MkdirAll(procDir, 0o755); err != nil {
		return "", err
	}

	projectRoot := m.resolver.ProjectRoot()
	count, err := procmena.RenderToDir(projectRoot, procDir, RenderArchetype, currentRite)
	if err != nil {
		os.RemoveAll(procDir)
		return "", err
	}
	if count == 0 {
		os.RemoveAll(procDir)
		return "", nil
	}

	return procDir, nil
}

// getMenaDir returns the mena directory path.
// Resolution order: project-level → KnossosHome → XDG data dir.
// Returns "" if none found (caller should try embedded fallback).
func (m *Materializer) getMenaDir() string {
	// 1. Check for project-level mena first (.knossos/mena/ satellite overrides)
	projectMena := filepath.Join(m.resolver.ProjectRoot(), ".knossos", "mena")
	if _, err := os.Stat(projectMena); err == nil {
		return projectMena
	}

	// 2. Fall back to Knossos platform mena (developer case)
	if m.sourceResolver.KnossosHome() != "" {
		knossosMena := filepath.Join(m.sourceResolver.KnossosHome(), "mena")
		if _, err := os.Stat(knossosMena); err == nil {
			return knossosMena
		}
	}

	// 3. Check XDG data dir (installed user case)
	if xdgMena := xdgMenaPath(); xdgMena != "" {
		if _, err := os.Stat(xdgMena); err == nil {
			return xdgMena
		}
	}

	return ""
}

// xdgMenaPath returns the XDG data directory path for platform mena.
func xdgMenaPath() string {
	return filepath.Join(config.XDGDataDir(), "mena")
}
