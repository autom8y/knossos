package materialize

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// knossosOnlyFields are knossos-internal metadata fields stripped during projection.
// CC does not consume these. Unknown fields pass through for forward compatibility.
// NOTE: `color` is NOT stripped — CC uses it for subagent UI identification.
var knossosOnlyFields = map[string]bool{
	"type":           true,
	"role":           true,
	"upstream":       true,
	"downstream":     true,
	"produces":       true,
	"contract":       true,
	"schema_version": true,
	"write-guard":    true,
	"aliases":        true,
}

// transformAgentContent projects agent source into CC-consumable form.
//
// Transformation:
//  1. Parse frontmatter YAML
//  2. Capture write-guard value (needed for hook resolution)
//  3. Strip all knossosOnlyFields from the frontmatter map
//  4. Inject name from agentName parameter
//  5. If write-guard was present, resolve against defaults and merge hooks
//  6. Reserialize frontmatter + body
func transformAgentContent(content []byte, agentName string, defaults *WriteGuardDefaults) ([]byte, error) {
	yamlBytes, body, err := frontmatter.Parse(content)
	if err != nil {
		return content, nil // Not valid frontmatter — pass through unchanged
	}

	// Unmarshal into a map to preserve all fields and unknown keys
	var fmMap map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &fmMap); err != nil {
		return content, nil // Invalid YAML — pass through unchanged
	}

	// Capture write-guard value before stripping
	agentWG, hasWriteGuard := fmMap["write-guard"]

	// Strip all knossos-only fields
	for field := range knossosOnlyFields {
		delete(fmMap, field)
	}

	// Auto-inject name from filename
	fmMap["name"] = agentName

	// Write-guard resolution (conditional on original presence)
	if hasWriteGuard {
		resolved := ResolveWriteGuard(defaults, agentName, agentWG)
		if resolved != nil {
			generatedHooks := GenerateWriteGuardHooks(resolved)
			mergeHooksIntoMap(fmMap, generatedHooks)
		}
	}

	return reconstructFrontmatter(fmMap, body)
}

// mergeHooksIntoMap merges generated hook entries into the frontmatter map's hooks field.
// Existing non-write-guard hooks are preserved.
func mergeHooksIntoMap(fmMap map[string]interface{}, generatedHooks map[string]interface{}) {
	if generatedHooks == nil {
		return
	}

	existing, hasHooks := fmMap["hooks"]
	if !hasHooks {
		fmMap["hooks"] = generatedHooks
		return
	}

	// Merge: replace PreToolUse write-guard entries, preserve others
	existingMap, ok := existing.(map[string]interface{})
	if !ok {
		fmMap["hooks"] = generatedHooks
		return
	}

	for event, entries := range generatedHooks {
		existingMap[event] = entries
	}
	fmMap["hooks"] = existingMap
}

// reconstructFrontmatter serializes a frontmatter map and body back into markdown content.
func reconstructFrontmatter(fmMap map[string]interface{}, body []byte) ([]byte, error) {
	yamlOut, err := yaml.Marshal(fmMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transformed frontmatter: %w", err)
	}

	// Build: ---\n + yaml + ---\n + body
	result := []byte("---\n")
	result = append(result, yamlOut...)
	result = append(result, []byte("---\n")...)
	result = append(result, body...)
	return result, nil
}

// loadSharedHookDefaults loads hook_defaults from the shared rite manifest.
// Returns nil if the shared manifest doesn't exist or has no hook_defaults.
func (m *Materializer) loadSharedHookDefaults(resolved *ResolvedRite) *HookDefaults {
	var data []byte
	var err error

	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.EmbeddedFS != nil {
		// Embedded FS: look for shared manifest relative to rite
		data, err = fs.ReadFile(m.sourceResolver.EmbeddedFS, "rites/shared/manifest.yaml")
	} else {
		// Filesystem: look relative to project root
		sharedPath := filepath.Join(m.resolver.ProjectRoot(), "rites", "shared", "manifest.yaml")
		data, err = os.ReadFile(sharedPath)
	}
	if err != nil {
		return nil // Shared manifest not found — graceful degradation
	}

	var manifest RiteManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		log.Printf("Warning: failed to parse shared manifest for hook_defaults: %v", err)
		return nil
	}

	return manifest.HookDefaults
}
