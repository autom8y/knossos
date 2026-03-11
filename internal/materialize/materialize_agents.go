package materialize

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// materializeAgents copies rite-scoped agent files to .claude/agents/.
// Uses selective write: only knossos-managed agents (from manifest) are replaced.
// User-created agents not in the manifest are preserved.
// Cross-rite agents (pythia, moirai, etc.) are user-scope owned and NOT handled here.
func (m *Materializer) materializeAgents(manifest *RiteManifest, ritePath, claudeDir string, resolved *ResolvedRite, collector provenance.Collector, writeGuardDefaults *WriteGuardDefaults, skillPolicies []SkillPolicy, modelOverride, channel string) error {
	agentsDir := filepath.Join(claudeDir, "agents")

	// Ensure agents directory exists (selective — do NOT RemoveAll)
	if err := paths.EnsureDir(agentsDir); err != nil {
		return err
	}

	// Build managed agent set from rite manifest only.
	// Cross-rite agents are user-scope owned (synced to ~/.claude/agents/).
	managedAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		managedAgents[agent.Name+".md"] = true
	}

	// NOTE: We intentionally do NOT pre-delete managed agents before rewriting.
	// fileutil.WriteIfChanged() handles overwrite-if-different atomically. Pre-deletion
	// causes CC's file watcher to see DELETE events for files that are immediately
	// recreated, which crashes/disrupts active Claude Code sessions.

	// Phase 1: Render archetype agents (before source file walk).
	// Archetype agents are rendered from templates in knossos/archetypes/ and do NOT
	// need a source file in the rite's agents/ directory.
	archetypeAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		if agent.Archetype == "" {
			continue
		}
		archetypeAgents[agent.Name+".md"] = true

		content, err := renderArchetypeAgent(m.resolver.ProjectRoot(), agent, manifest, m.renderArchetypeResolved)
		if err != nil {
			return fmt.Errorf("archetype render failed for %s: %w", agent.Name, err)
		}

		// Run through the same transform pipeline as source-copied agents.
		// Transform failure is an error, not a warning: knossos-only frontmatter fields
		// (type, upstream, downstream, contract) must never reach CC-visible agent files.
		transformed, tErr := transformAgentContent(content, &TransformContext{AgentName: agent.Name, WriteGuardDefaults: writeGuardDefaults, AgentDefaults: manifest.AgentDefaults, SkillPolicies: skillPolicies, ModelOverride: modelOverride})
		if tErr != nil {
			return fmt.Errorf("agent transform failed for archetype agent %s: %w", agent.Name, tErr)
		}
		content = transformed

		destPath := filepath.Join(agentsDir, agent.Name+".md")
		written, err := fileutil.WriteIfChanged(destPath, content, 0644)
		if err != nil {
			return err
		}
		if written {
			relPath := "agents/" + agent.Name + ".md"
			sourcePath := "knossos/archetypes/" + agent.Archetype + ".md.tpl"
			collector.Record(relPath, provenance.NewKnossosEntry(
				provenance.ScopeRite,
				sourcePath,
				"archetype",
				checksum.Bytes(content), channel,
			))
		}
	}

	// Phase 2: Copy source agents from rite directory (skip archetype-rendered agents).
	// For embedded sources, use fs.FS to read agent files
	var writeErr error
	if resolved != nil && resolved.Source.Type == SourceEmbedded {
		rFS := m.riteFS(resolved)
		agentsSub, err := fs.Sub(rFS, "agents")
		if err != nil {
			return nil	// No agents sub-dir in embedded FS
		}
		// Check if agents dir exists in embedded FS
		if _, err := fs.Stat(rFS, "agents"); err != nil {
			return nil	// No agents in this rite
		}
		// Copy agents and record provenance
		writeErr = fs.WalkDir(agentsSub, ".", func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return walkErr
			}
			// Skip agents already rendered from archetype
			if archetypeAgents[filepath.Base(path)] {
				return nil
			}
			content, err := fs.ReadFile(agentsSub, path)
			if err != nil {
				return err
			}
			// Project agent source into CC-consumable form (merge defaults, strip knossos metadata, inject name, resolve hooks).
			// Transform failure is an error, not a warning: knossos-only frontmatter fields
			// (type, upstream, downstream, contract) must never reach CC-visible agent files.
			agentName := strings.TrimSuffix(filepath.Base(path), ".md")
			transformed, tErr := transformAgentContent(content, &TransformContext{AgentName: agentName, WriteGuardDefaults: writeGuardDefaults, AgentDefaults: manifest.AgentDefaults, SkillPolicies: skillPolicies})
			if tErr != nil {
				return fmt.Errorf("agent transform failed for %s: %w", agentName, tErr)
			}
			content = transformed
			destPath := filepath.Join(agentsDir, path)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
			written, err := fileutil.WriteIfChanged(destPath, content, 0644)
			if err != nil {
				return err
			}
			if written {
				relPath := "agents/" + path
				sourcePath := resolved.RitePath + "/agents/" + path
				collector.Record(relPath, provenance.NewKnossosEntry(
					provenance.ScopeRite,
					sourcePath,
					string(resolved.Source.Type),
					checksum.Bytes(content), channel,
				))
			}
			return nil
		})
	} else {
		// Filesystem path: use existing os-based copy
		sourceAgentsDir := filepath.Join(ritePath, "agents")
		if _, err := os.Stat(sourceAgentsDir); os.IsNotExist(err) {
			return nil	// No agents in this rite
		}

		writeErr = filepath.WalkDir(sourceAgentsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			// Skip agents already rendered from archetype
			if archetypeAgents[filepath.Base(path)] {
				return nil
			}

			// Read source file
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Project agent source into CC-consumable form (merge defaults, strip knossos metadata, inject name, resolve hooks).
			// Transform failure is an error, not a warning: knossos-only frontmatter fields
			// (type, upstream, downstream, contract) must never reach CC-visible agent files.
			agentName := strings.TrimSuffix(filepath.Base(path), ".md")
			transformed, tErr := transformAgentContent(content, &TransformContext{AgentName: agentName, WriteGuardDefaults: writeGuardDefaults, AgentDefaults: manifest.AgentDefaults, SkillPolicies: skillPolicies})
			if tErr != nil {
				return fmt.Errorf("agent transform failed for %s: %w", agentName, tErr)
			}
			content = transformed

			// Compute relative path
			relPath, err := filepath.Rel(sourceAgentsDir, path)
			if err != nil {
				return err
			}

			// Write to destination (only if changed)
			destPath := filepath.Join(agentsDir, relPath)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			written, err := fileutil.WriteIfChanged(destPath, content, 0644)
			if err != nil {
				return err
			}

			// Record provenance after successful write
			if written {
				srcRelPath, _ := filepath.Rel(m.resolver.ProjectRoot(), path)
				collector.Record("agents/"+relPath, provenance.NewKnossosEntry(
					provenance.ScopeRite,
					srcRelPath,
					string(resolved.Source.Type),
					checksum.Bytes(content), channel,
				))
			}

			return nil
		})
	}

	// Validate: warn about phantom agents (declared in manifest but not found on disk)
	for _, agent := range manifest.Agents {
		agentPath := filepath.Join(agentsDir, agent.Name+".md")
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			slog.Warn("agent declared in rite manifest but no .md file found", "agent", agent.Name, "path", agentPath)
		}
	}

	return writeErr
}

// NOTE: listCrossRiteAgents and materializeCrossRiteAgents were removed.
// Cross-rite agents (pythia, moirai, context-engineer, theoros) are
// user-scope owned: synced from KNOSSOS_HOME/agents/ to ~/.claude/agents/
// by user-scope sync. They are NOT copied to project .claude/agents/.

// detectOrphans finds agent files that are not in the incoming rite's manifest.
// If a provenance manifest exists, uses manifest-based detection: files with
// owner=user or files not in the provenance manifest are orphans.
// Otherwise, falls back to rite manifest membership check (backward compatible).
func (m *Materializer) detectOrphans(manifest *RiteManifest, claudeDir string, resolved *ResolvedRite, channel string) ([]string, error) {
	agentsDir := filepath.Join(claudeDir, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Build expected agent set from rite manifest only.
	// Cross-rite agents (pythia, moirai, etc.) are user-scope owned — they
	// live at ~/.claude/agents/ and are NOT expected at project level.
	expectedAgents := make(map[string]bool)
	for _, agent := range manifest.Agents {
		expectedAgents[agent.Name+".md"] = true
	}

	// Try loading provenance manifest for manifest-based detection.
	// Uses channel-keyed path so gemini reads its own manifest.
	manifestPath := provenance.ManifestPathForChannel(m.getKnossosDir(), channel)
	provenanceManifest, err := provenance.Load(manifestPath)

	// If provenance manifest exists, use manifest-based orphan detection
	if err == nil && provenanceManifest != nil {
		return m.detectOrphansFromProvenance(expectedAgents, claudeDir, provenanceManifest)
	}

	// Fallback: rite manifest membership check (backward compatible)
	return m.detectOrphansLegacy(expectedAgents, agentsDir)
}

// detectOrphansFromProvenance detects orphans using the provenance manifest.
// An agent file is an orphan if:
//   - It is NOT in the provenance manifest, OR
//   - It has owner=user in the provenance manifest, OR
//   - It is knossos-owned BUT not in the expected agents set (rite + cross-rite)
func (m *Materializer) detectOrphansFromProvenance(expectedAgents map[string]bool, claudeDir string, provenanceManifest *provenance.ProvenanceManifest) ([]string, error) {
	agentsDir := filepath.Join(claudeDir, "agents")

	orphans := []string{}
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only consider .md files
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		// Construct relative path within .claude/
		relativePath := filepath.Join("agents", entry.Name())

		// Check if file is in provenance manifest
		provenanceEntry, exists := provenanceManifest.Entries[relativePath]

		// File not in provenance manifest -> orphan
		if !exists {
			orphans = append(orphans, entry.Name())
			continue
		}

		// File with owner=user -> orphan (user-created or previously promoted)
		if provenanceEntry.Owner == provenance.OwnerUser {
			orphans = append(orphans, entry.Name())
			continue
		}

		// Knossos-owned file not in current rite manifest -> orphan
		if provenanceEntry.Owner == provenance.OwnerKnossos && !expectedAgents[entry.Name()] {
			orphans = append(orphans, entry.Name())
			continue
		}

		// Knossos-owned entries still in rite manifest are NOT orphans
	}

	return orphans, nil
}

// detectOrphansLegacy uses manifest membership check.
// This is the fallback when no provenance manifest exists (backward compatible).
func (m *Materializer) detectOrphansLegacy(expectedAgents map[string]bool, agentsDir string) ([]string, error) {
	// Find files that aren't expected
	orphans := []string{}
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !expectedAgents[entry.Name()] {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				orphans = append(orphans, entry.Name())
			}
		}
	}

	return orphans, nil
}

// backupAndRemoveOrphans creates a backup of orphan agents then removes them.
func (m *Materializer) backupAndRemoveOrphans(orphans []string, claudeDir string, knossosDir string) (string, error) {
	if len(orphans) == 0 {
		return "", nil
	}

	agentsDir := filepath.Join(claudeDir, "agents")
	backupDir := filepath.Join(knossosDir, ".orphan-backup", time.Now().Format("20060102-150405"))

	if err := paths.EnsureDir(backupDir); err != nil {
		return "", err
	}

	for _, orphan := range orphans {
		srcPath := filepath.Join(agentsDir, orphan)
		dstPath := filepath.Join(backupDir, orphan)

		// Copy to backup
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return "", err
		}

		// Remove original
		if err := os.Remove(srcPath); err != nil {
			return "", err
		}
	}

	return backupDir, nil
}

// promoteOrphans moves orphan agents to user-level ~/.claude/agents/.
func (m *Materializer) promoteOrphans(orphans []string, claudeDir string) error {
	if len(orphans) == 0 {
		return nil
	}

	agentsDir := filepath.Join(claudeDir, "agents")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	userAgentsDir := filepath.Join(homeDir, ".claude", "agents")

	if err := paths.EnsureDir(userAgentsDir); err != nil {
		return err
	}

	for _, orphan := range orphans {
		srcPath := filepath.Join(agentsDir, orphan)
		dstPath := filepath.Join(userAgentsDir, orphan)

		// Copy to user-level
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return err
		}

		// Remove from project-level
		if err := os.Remove(srcPath); err != nil {
			return err
		}
	}

	return nil
}
