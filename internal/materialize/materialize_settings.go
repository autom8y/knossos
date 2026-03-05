package materialize

import (
	"encoding/json"
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
	"github.com/autom8y/knossos/internal/sync"
)

// materializeSettingsWithManifest generates or updates settings.local.json.
// If manifest has MCP servers, merges them into existing settings.
// Loads hooks.yaml and merges hook registrations into settings.
// If no manifest or no MCP servers, creates minimal settings if needed.
func (m *Materializer) materializeSettingsWithManifest(claudeDir string, manifest *RiteManifest, collector provenance.Collector) error {
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Load existing settings or create empty map
	existingSettings, err := loadExistingSettings(settingsPath)
	if err != nil {
		return err
	}

	// Load hooks.yaml and merge hook registrations
	if hooksConfig := m.loadHooksConfig(); hooksConfig != nil {
		existingSettings = mergeHooksSettings(existingSettings, hooksConfig)
	} else if existingSettings["hooks"] == nil {
		// No hooks.yaml found — ensure hooks key exists (empty)
		existingSettings["hooks"] = make(map[string]any)
	}

	// If manifest has MCP servers, merge them
	if manifest != nil && len(manifest.MCPServers) > 0 {
		existingSettings = mergeMCPServers(existingSettings, manifest.MCPServers)
	}

	// Write settings (only if content changed, to avoid triggering Claude Code file watcher)
	err = saveSettings(settingsPath, existingSettings)
	if err != nil {
		return err
	}

	// Record provenance after successful write
	hash, err := checksum.File(settingsPath)
	if err == nil && hash != "" {
		collector.Record("settings.local.json", provenance.NewKnossosEntry(
			provenance.ScopeRite,
			"(generated)",
			"template",
			hash,
		))
	}

	return nil
}

// injectElCheapoSettings layers el-cheapo mode on top of settings.local.json.
// Called AFTER normal settings materialization. Injects model override and a
// Stop hook that reverts the override on session exit.
func (m *Materializer) injectElCheapoSettings(claudeDir string) error {
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	existingSettings, err := loadExistingSettings(settingsPath)
	if err != nil {
		return err
	}

	// 1. Set model override
	existingSettings["model"] = "haiku"

	// 2. Inject Stop hook for revert.
	// Command starts with "ari hook" so IsAriManagedGroup() classifies it as
	// ari-managed — normal sync will strip it during hook merge.
	revertHook := map[string]any{
		"hooks": []map[string]any{
			{
				"type":    "command",
				"command": "ari hook cheapo-revert --output json",
				"timeout": 30,
			},
		},
	}

	// Merge into existing hooks, appending to Stop event (preserve autopark)
	hooksMap, ok := existingSettings["hooks"].(map[string]any)
	if !ok {
		hooksMap = make(map[string]any)
	}

	existingStopHooks := []any{}
	if existing, ok := hooksMap["Stop"]; ok {
		if arr, ok := existing.([]any); ok {
			existingStopHooks = arr
		}
	}
	existingStopHooks = append(existingStopHooks, revertHook)
	hooksMap["Stop"] = existingStopHooks
	existingSettings["hooks"] = hooksMap

	// 3. Write marker file for diagnostics and revert detection
	knossosDir := filepath.Join(filepath.Dir(claudeDir), ".knossos")
	_ = os.MkdirAll(knossosDir, 0755)
	markerPath := filepath.Join(knossosDir, ".el-cheapo-active")
	_ = os.WriteFile(markerPath, []byte("haiku\n"), 0644)

	return saveSettings(settingsPath, existingSettings)
}

// trackState updates .knossos/sync/state.json with materialization metadata.
func (m *Materializer) trackState(manifest *RiteManifest, activeRiteName string) error {
	stateManager := sync.NewStateManager(m.resolver)

	// During staged materialization, override the sync dir to target the staging directory.
	if m.claudeDirOverride != "" {
		// During staged materialization, sync state goes alongside the staging directory.
		stagingParent := filepath.Dir(m.claudeDirOverride)
		stateManager.SetSyncDir(filepath.Join(stagingParent, ".knossos", "sync"))
	}

	// Load or initialize state
	state, err := stateManager.Load()
	if err != nil {
		return err
	}

	if state == nil {
		// Initialize new state
		state, err = stateManager.Initialize()
		if err != nil {
			return err
		}
	}

	// Update last sync time. active_rite was removed from state.json (PKG-008):
	// all 18 runtime consumers read from .claude/ACTIVE_RITE file instead.
	state.LastSync = time.Now().UTC()
	err = stateManager.Save(state)
	if err != nil {
		return err
	}

	return nil
}

// clearInvocationState removes INVOCATION_STATE.yaml which becomes stale on rite switch.
// The file tracks borrowed components from the previous rite's invocations.
func (m *Materializer) clearInvocationState(claudeDir string) error {
	knossosDir := filepath.Join(filepath.Dir(claudeDir), ".knossos")
	err := os.Remove(filepath.Join(knossosDir, "INVOCATION_STATE.yaml"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// throughlineIDsFile is the filename for session-scoped throughline agent ID maps.
// Defined locally to avoid importing internal/cmd/hook from materialize.
const throughlineIDsFile = ".throughline-ids.json"

// cleanupThroughlineIDs removes .throughline-ids.json from all session directories.
// Called on rite switch because agent IDs are rite-specific — stale IDs from the
// previous rite cause wasted resume attempts before falling back to fresh invocation.
// Best-effort: errors on individual files are logged, not fatal.
func (m *Materializer) cleanupThroughlineIDs() int {
	sessionsDir := m.resolver.SessionsDir()
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return 0
	}

	cleaned := 0
	for _, entry := range entries {
		if !entry.IsDir() || !paths.IsSessionDir(entry.Name()) {
			continue
		}
		idFile := filepath.Join(sessionsDir, entry.Name(), throughlineIDsFile)
		if err := os.Remove(idFile); err == nil {
			cleaned++
		} else if !os.IsNotExist(err) {
			slog.Warn("failed to remove throughline IDs", "path", idFile, "error", err)
		}
	}
	return cleaned
}

// materializeWorkflow copies workflow.yaml from the rite to .knossos/ACTIVE_WORKFLOW.yaml.
// If the rite has no workflow.yaml, any existing ACTIVE_WORKFLOW.yaml is removed to
// prevent stale workflow data from a previous rite persisting after switch.
func (m *Materializer) materializeWorkflow(knossosDir string, resolved *ResolvedRite, collector provenance.Collector) error {
	dstPath := filepath.Join(knossosDir, "ACTIVE_WORKFLOW.yaml")
	rFS := m.riteFS(resolved)
	content, err := fs.ReadFile(rFS, "workflow.yaml")
	if err != nil {
		// No workflow.yaml in this rite — remove any stale file from previous rite
		if removeErr := os.Remove(dstPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return removeErr
		}
		return nil
	}
	written, err := fileutil.WriteIfChanged(dstPath, content, 0644)
	if err != nil {
		return err
	}

	// Record provenance after successful write
	if written {
		projectRoot := m.resolver.ProjectRoot()
		sourcePath := resolved.RitePath + "/workflow.yaml"
		srcRelPath, _ := filepath.Rel(projectRoot, sourcePath)
		collector.Record("ACTIVE_WORKFLOW.yaml", provenance.NewKnossosEntry(
			provenance.ScopeRite,
			srcRelPath,
			string(resolved.Source.Type),
			checksum.Bytes(content),
		))
	}

	return nil
}

// writeActiveRite writes the ACTIVE_RITE marker file.
func (m *Materializer) writeActiveRite(riteName, claudeDir string) error {
	activeRitePath := m.resolver.ActiveRiteFile()
	content := []byte(riteName + "\n")
	_, err := fileutil.WriteIfChanged(activeRitePath, content, 0644)
	if err != nil {
		return err
	}

	return nil
}

// cleanupStaleBlanketSettings removes .claude/settings.json if it matches a known
// stale fingerprint from the deleted writeDefaultSettings() function AND has no
// provenance entry (indicating it was not created by the current pipeline).
//
// Two stale fingerprints are recognized:
//  1. Blanket-deny agent-guard hook (no --allow-path flags)
//  2. Empty CC-default stub (permissions + empty hooks)
//
// Both gates must pass: no provenance entry AND structural fingerprint match.
// Returns true if the file was removed, false otherwise.
func (m *Materializer) cleanupStaleBlanketSettings(claudeDir string, manifest *provenance.ProvenanceManifest) bool {
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Gate 1: File must exist
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return false
	}

	// Gate 2: No provenance entry for "settings.json".
	// If the pipeline tracks this file, it is managed — do not touch.
	if manifest != nil {
		if _, tracked := manifest.Entries["settings.json"]; tracked {
			return false
		}
	}

	// Gate 3: Parse JSON and check structural fingerprint
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return false
	}

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		// Not valid JSON — could be user content, leave it alone
		return false
	}

	if !matchesStaleSettingsFingerprint(parsed) {
		return false
	}

	// Both gates passed: safe to remove
	if err := os.Remove(settingsPath); err != nil {
		slog.Warn("failed to remove stale settings.json", "path", claudeDir, "error", err)
		return false
	}
	slog.Info("removed stale settings.json", "path", claudeDir)
	return true
}

// matchesStaleSettingsFingerprint returns true if the parsed settings.json content
// matches one of the two known stale structures from the deleted writeDefaultSettings().
//
// Fingerprint 1: Blanket-deny agent-guard hook (no --allow-path flags)
//
//	{"hooks": {"PreToolUse": [{"hooks": [{"command": "ari hook agent-guard ..."}]}]}}
//	Key signal: top-level has only "hooks", command contains "ari hook agent-guard",
//	command does NOT contain "--allow-path".
//
// Fingerprint 2: Empty CC-default stub
//
//	{"permissions": {"allow": [], "additionalDirectories": []}, "hooks": {}}
func matchesStaleSettingsFingerprint(parsed map[string]any) bool {
	// Fingerprint 1: Blanket-deny agent-guard hook
	// Top-level must have only "hooks" key.
	if len(parsed) == 1 {
		hooksRaw, ok := parsed["hooks"]
		if ok {
			hooks, ok := hooksRaw.(map[string]any)
			if ok {
				preToolUseRaw, ok := hooks["PreToolUse"]
				if ok {
					preToolUse, ok := preToolUseRaw.([]any)
					if ok {
						for _, entryRaw := range preToolUse {
							entry, ok := entryRaw.(map[string]any)
							if !ok {
								continue
							}
							innerHooksRaw, ok := entry["hooks"]
							if !ok {
								continue
							}
							innerHooks, ok := innerHooksRaw.([]any)
							if !ok {
								continue
							}
							for _, hookRaw := range innerHooks {
								hook, ok := hookRaw.(map[string]any)
								if !ok {
									continue
								}
								cmdRaw, ok := hook["command"]
								if !ok {
									continue
								}
								cmd, ok := cmdRaw.(string)
								if !ok {
									continue
								}
								if strings.Contains(cmd, "ari hook agent-guard") &&
									!strings.Contains(cmd, "--allow-path") {
									return true
								}
							}
						}
					}
				}
			}
		}
	}

	// Fingerprint 2: Empty CC-default stub
	// {"permissions": {"allow": [], "additionalDirectories": []}, "hooks": {}}
	if len(parsed) == 2 {
		hooksRaw, hasHooks := parsed["hooks"]
		permissionsRaw, hasPermissions := parsed["permissions"]
		if hasHooks && hasPermissions {
			hooks, hooksIsMap := hooksRaw.(map[string]any)
			permissions, permsIsMap := permissionsRaw.(map[string]any)
			if hooksIsMap && permsIsMap && len(hooks) == 0 && len(permissions) == 2 {
				allowRaw, hasAllow := permissions["allow"]
				addDirsRaw, hasAddDirs := permissions["additionalDirectories"]
				if hasAllow && hasAddDirs {
					allow, allowIsSlice := allowRaw.([]any)
					addDirs, addDirsIsSlice := addDirsRaw.([]any)
					if allowIsSlice && addDirsIsSlice && len(allow) == 0 && len(addDirs) == 0 {
						return true
					}
				}
			}
		}
	}

	return false
}

// saveProvenanceManifest merges collector entries with divergence report and previous manifest,
// then writes the final manifest to disk. Delegates to provenance.Merge() for the algorithm.
//
// claudeDir is passed explicitly because Merge uses it to check whether files still exist
// on disk. knossosDir is the sibling .knossos/ directory — some tracked files (e.g.
// ACTIVE_WORKFLOW.yaml) live there and Merge checks both directories.
func (m *Materializer) saveProvenanceManifest(
	manifestPath string,
	claudeDir string,
	activeRite string,
	collector provenance.Collector,
	divergenceReport *provenance.DivergenceReport,
	prevManifest *provenance.ProvenanceManifest,
	overwriteDiverged bool,
) error {
	knossosDir := m.resolver.KnossosDir()
	finalManifest := provenance.Merge(claudeDir, knossosDir, activeRite, collector, divergenceReport, prevManifest, overwriteDiverged)
	return provenance.Save(manifestPath, finalManifest)
}
