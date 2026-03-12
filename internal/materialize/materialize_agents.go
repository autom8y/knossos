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
	"github.com/autom8y/knossos/internal/materialize/compiler"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"gopkg.in/yaml.v3"
)

// materializeAgents copies rite-scoped agent files to .claude/agents/ (or .gemini/agents/).
// Uses selective write: only knossos-managed agents (from manifest) are replaced.
// User-created agents not in the manifest are preserved.
// Cross-rite agents (pythia, moirai, etc.) are user-scope owned and NOT handled here.
// When comp is non-nil, CompileAgent() is called after transformAgentContent()
// to translate tool names for the target channel. The compiler is channel-aware:
// ClaudeCompiler is a pass-through, GeminiCompiler translates tool names.
func (m *Materializer) materializeAgents(manifest *RiteManifest, ritePath, channelDir string, resolved *ResolvedRite, collector provenance.Collector, writeGuardDefaults *WriteGuardDefaults, skillPolicies []SkillPolicy, modelOverride, channel string, comp compiler.ChannelCompiler) error {
	agentsDir := filepath.Join(channelDir, "agents")

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

		content, err := renderArchetypeAgentForChannel(m.resolver.ProjectRoot(), agent, manifest, m.renderArchetypeResolved, channel)
		if err != nil {
			return fmt.Errorf("archetype render failed for %s: %w", agent.Name, err)
		}

		// Run through the same transform pipeline as source-copied agents.
		// Transform failure is an error, not a warning: knossos-only frontmatter fields
		// (type, upstream, downstream, contract) must never reach CC-visible agent files.
		transformed, tErr := transformAgentContent(content, &TransformContext{AgentName: agent.Name, WriteGuardDefaults: writeGuardDefaults, AgentDefaults: manifest.AgentDefaults, SkillPolicies: skillPolicies, ModelOverride: modelOverride, Channel: channel})
		if tErr != nil {
			return fmt.Errorf("agent transform failed for archetype agent %s: %w", agent.Name, tErr)
		}
		content = transformed

		// Channel compilation: translate tool names for the target channel.
		// The compiler is optional (nil for channels without translation needs).
		if comp != nil {
			compiled, cErr := compileAgentContent(agent.Name, content, comp)
			if cErr != nil {
				return fmt.Errorf("agent compile failed for archetype agent %s: %w", agent.Name, cErr)
			}
			content = compiled
		}

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
			transformed, tErr := transformAgentContent(content, &TransformContext{AgentName: agentName, WriteGuardDefaults: writeGuardDefaults, AgentDefaults: manifest.AgentDefaults, SkillPolicies: skillPolicies, Channel: channel})
			if tErr != nil {
				return fmt.Errorf("agent transform failed for %s: %w", agentName, tErr)
			}
			content = transformed

			// Channel compilation: translate tool names for non-claude channels.
			if comp != nil && channel != "claude" {
				compiled, cErr := compileAgentContent(agentName, content, comp)
				if cErr != nil {
					return fmt.Errorf("agent compile failed for %s: %w", agentName, cErr)
				}
				content = compiled
			}
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
			transformed, tErr := transformAgentContent(content, &TransformContext{AgentName: agentName, WriteGuardDefaults: writeGuardDefaults, AgentDefaults: manifest.AgentDefaults, SkillPolicies: skillPolicies, Channel: channel})
			if tErr != nil {
				return fmt.Errorf("agent transform failed for %s: %w", agentName, tErr)
			}
			content = transformed

			// Channel compilation: translate tool names for non-claude channels.
			if comp != nil && channel != "claude" {
				compiled, cErr := compileAgentContent(agentName, content, comp)
				if cErr != nil {
					return fmt.Errorf("agent compile failed for %s: %w", agentName, cErr)
				}
				content = compiled
			}

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

// compileAgentContent parses the frontmatter from transformed agent content,
// calls comp.CompileAgent() to perform channel-specific translation (e.g. tool
// name mapping for Gemini), and returns the re-serialized content.
//
// This is called after transformAgentContent() has already stripped knossos-only
// fields and normalized defaults. The frontmatter at this point is clean and
// ready for channel translation.
func compileAgentContent(agentName string, content []byte, comp compiler.ChannelCompiler) ([]byte, error) {
	// Parse frontmatter separator
	if len(content) < 4 || string(content[:4]) != "---\n" {
		// No frontmatter — pass through unchanged (compiler has nothing to translate)
		return comp.CompileAgent(agentName, nil, string(content))
	}

	// Find closing ---
	rest := content[4:]
	end := -1
	for i := 0; i < len(rest)-3; i++ {
		if rest[i] == '-' && rest[i+1] == '-' && rest[i+2] == '-' && (i+3 == len(rest) || rest[i+3] == '\n') {
			end = i
			break
		}
	}
	if end == -1 {
		// Malformed frontmatter — pass through unchanged
		return content, nil
	}

	yamlBytes := rest[:end]
	bodyOffset := end + 3
	if bodyOffset < len(rest) && rest[bodyOffset] == '\n' {
		bodyOffset++
	}
	body := string(rest[bodyOffset:])

	var fmMap map[string]any
	if err := yaml.Unmarshal(yamlBytes, &fmMap); err != nil {
		// Invalid YAML — pass through unchanged
		return content, nil
	}
	if fmMap == nil {
		fmMap = make(map[string]any)
	}

	return comp.CompileAgent(agentName, fmMap, body)
}

// NOTE: listCrossRiteAgents and materializeCrossRiteAgents were removed.
// Cross-rite agents (pythia, moirai, context-engineer, theoros) are
// user-scope owned: synced from KNOSSOS_HOME/agents/ to ~/.claude/agents/
// by user-scope sync. They are NOT copied to project .claude/agents/.

// detectOrphans finds agent files that are not in the incoming rite's manifest.
// If a provenance manifest exists, uses manifest-based detection: files with
// owner=user or files not in the provenance manifest are orphans.
// Otherwise, falls back to rite manifest membership check (backward compatible).
func (m *Materializer) detectOrphans(manifest *RiteManifest, channelDir string, resolved *ResolvedRite, channel string) ([]string, error) {
	agentsDir := filepath.Join(channelDir, "agents")
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
		return m.detectOrphansFromProvenance(expectedAgents, channelDir, provenanceManifest)
	}

	// Fallback: rite manifest membership check (backward compatible)
	return m.detectOrphansLegacy(expectedAgents, agentsDir)
}

// detectOrphansFromProvenance detects orphans using the provenance manifest.
// An agent file is an orphan if:
//   - It is NOT in the provenance manifest, OR
//   - It has owner=user in the provenance manifest, OR
//   - It is knossos-owned BUT not in the expected agents set (rite + cross-rite)
func (m *Materializer) detectOrphansFromProvenance(expectedAgents map[string]bool, channelDir string, provenanceManifest *provenance.ProvenanceManifest) ([]string, error) {
	agentsDir := filepath.Join(channelDir, "agents")

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
func (m *Materializer) backupAndRemoveOrphans(orphans []string, channelDir string, knossosDir string) (string, error) {
	if len(orphans) == 0 {
		return "", nil
	}

	agentsDir := filepath.Join(channelDir, "agents")
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
func (m *Materializer) promoteOrphans(orphans []string, channelDir string) error {
	if len(orphans) == 0 {
		return nil
	}

	agentsDir := filepath.Join(channelDir, "agents")
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
