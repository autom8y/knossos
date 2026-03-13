package materialize

import (
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// knossosOnlyFields are knossos-internal metadata fields stripped during projection.
// CC does not consume these. Unknown fields pass through for forward compatibility.
// NOTE: `color` is NOT stripped — CC uses it for subagent UI identification.
var knossosOnlyFields = map[string]bool{
	"type":                  true,
	"role":                  true,
	"upstream":              true,
	"downstream":            true,
	"produces":              true,
	"contract":              true,
	"schema_version":        true,
	"write-guard":           true,
	"aliases":               true,
	"skill_policy_exclude":  true,
	"skill_policy_override": true,
}

// TransformContext bundles all policy inputs for agent content transformation.
// Introduced to prevent parameter list growth as new policy types are added.
type TransformContext struct {
	AgentName          string
	WriteGuardDefaults *WriteGuardDefaults
	AgentDefaults      map[string]any
	SkillPolicies      []SkillPolicy
	ModelOverride      string // If set, forces model field in agent frontmatter (el-cheapo mode)
	Channel            string // Target channel ("claude", "gemini", ""). Empty == "claude" behavior.
}

// transformAgentContent projects agent source into CC-consumable form.
//
// Transformation:
//  1. Parse frontmatter YAML
//  2. Merge agent_defaults from manifest (defaults before agent, agent wins)
//  3. Capture write-guard value (needed for hook resolution)
//  4. Strip all knossosOnlyFields from the frontmatter map
//  5. Inject name from agentName parameter
//  6. If write-guard was present, resolve against defaults and merge hooks
//  7. Reserialize frontmatter + body
func transformAgentContent(content []byte, ctx *TransformContext) ([]byte, error) {
	yamlBytes, body, err := frontmatter.Parse(content)
	if err != nil {
		return content, nil // Not valid frontmatter — pass through unchanged
	}

	// Unmarshal into a map to preserve all fields and unknown keys
	var fmMap map[string]any
	if err := yaml.Unmarshal(yamlBytes, &fmMap); err != nil {
		return content, nil // Invalid YAML — pass through unchanged
	}

	// Merge manifest-level agent_defaults before any stripping
	if len(ctx.AgentDefaults) > 0 {
		fmMap = MergeAgentDefaults(ctx.AgentDefaults, fmMap)
	}

	// Apply skill policies (step 3.5 — after tools resolved from agent_defaults)
	if len(ctx.SkillPolicies) > 0 {
		fmMap, body = applySkillPolicies(fmMap, body, ctx.SkillPolicies)
	}

	// Capture write-guard value before stripping
	agentWG, hasWriteGuard := fmMap["write-guard"]

	// Strip all knossos-only fields
	for field := range knossosOnlyFields {
		delete(fmMap, field)
	}

	// Auto-inject name from filename
	fmMap["name"] = ctx.AgentName

	// Write-guard resolution (conditional on original presence)
	if hasWriteGuard {
		resolved := ResolveWriteGuard(ctx.WriteGuardDefaults, ctx.AgentName, agentWG)
		if resolved != nil {
			generatedHooks := GenerateWriteGuardHooks(resolved)
			mergeHooksIntoMap(fmMap, generatedHooks)
		}
	}

	// Model override (el-cheapo mode): force model field after all other transforms.
	// Runs AFTER MergeAgentDefaults so the override is truly blanket — no agent escapes.
	if ctx.ModelOverride != "" {
		fmMap["model"] = ctx.ModelOverride
	}

	// Channel body substitution: replace CC-specific patterns in body text for Gemini.
	// Runs after all frontmatter transforms so only the body is affected.
	if ctx.Channel == "gemini" {
		body = applyGeminiBodySubstitutions(body)
	}

	return reconstructFrontmatter(fmMap, body)
}

// mergeHooksIntoMap merges generated hook entries into the frontmatter map's hooks field.
// Existing non-write-guard hooks are preserved.
func mergeHooksIntoMap(fmMap map[string]any, generatedHooks map[string]any) {
	if generatedHooks == nil {
		return
	}

	existing, hasHooks := fmMap["hooks"]
	if !hasHooks {
		fmMap["hooks"] = generatedHooks
		return
	}

	// Merge: replace PreToolUse write-guard entries, preserve others
	existingMap, ok := existing.(map[string]any)
	if !ok {
		fmMap["hooks"] = generatedHooks
		return
	}

	maps.Copy(existingMap, generatedHooks)
	fmMap["hooks"] = existingMap
}

// reconstructFrontmatter serializes a frontmatter map and body back into markdown content.
func reconstructFrontmatter(fmMap map[string]any, body []byte) ([]byte, error) {
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

// geminiToolBodyRe matches backtick-delimited CC tool names in body text.
// Anchored to backtick delimiters to avoid replacing English words (e.g., "Read" in prose).
// Covers the tools that have Gemini equivalents in channel.CanonicalTool.
var geminiToolBodyRe = regexp.MustCompile("`(Read|Bash|Edit|Write|Glob|Grep)`")

// applyGeminiBodySubstitutions applies targeted text replacements to agent body content
// for the Gemini channel. This handles CC-specific patterns in body text that are not
// captured by frontmatter translation.
//
// Replacements applied:
//  1. Backtick-delimited tool names: `Read` -> `read_file`, etc.
//  2. Path references: .claude/ -> channel dir (only within backtick-delimited strings)
//  3. CC-specific phrases: "Task tool" -> "delegation" (in specialist anti-pattern warnings)
//
// Note: potnia archetype body is adapted at the archetype template level (orchestrator.md.tpl),
// not here. This function handles the ~5 non-archetype agents with scattered CC references.
func applyGeminiBodySubstitutions(body []byte) []byte {
	s := string(body)

	// 1. Translate backtick-delimited tool names to Gemini equivalents.
	s = geminiToolBodyRe.ReplaceAllStringFunc(s, func(match string) string {
		// match is like "`Read`" — extract the tool name
		tool := match[1 : len(match)-1]
		switch tool {
		case "Read":
			return "`read_file`"
		case "Bash":
			return "`run_shell_command`"
		case "Edit":
			return "`replace`"
		case "Write":
			return "`write_file`"
		case "Glob":
			return "`glob`"
		case "Grep":
			return "`grep_search`"
		}
		return match
	})

	// 2. Replace `.claude/` path references in backtick-delimited contexts (channel rewrite).
	// Only replace within backtick strings to avoid modifying prose.
	s = strings.ReplaceAll(s, "`.claude/", "`.gemini/")

	// 3. Replace "Task tool" phrase in specialist anti-pattern sections.
	// This phrase appears in body text as "Using Task tool (you don't have it)" or
	// "Direct delegation: Using Task tool". It's safe to replace as a phrase.
	s = strings.ReplaceAll(s, "Task tool", "delegation")

	return []byte(s)
}

// loadSharedManifest loads and parses the shared rite manifest (rites/shared/manifest.yaml).
// Returns nil and a nil error if the manifest doesn't exist (graceful degradation).
// Returns nil and a non-nil error only on parse failure (caller should log and degrade).
//
// Load-path order:
//  1. Embedded FS (SourceEmbedded): reads from the embedded rites filesystem
//  2. $KNOSSOS_HOME/rites/shared/manifest.yaml (satellite-safe: shared rites are never
//     inside the satellite project directory)
//  3. Project root fallback (knossos-on-knossos: project IS knossos)
func (m *Materializer) loadSharedManifest(resolved *ResolvedRite) (*RiteManifest, error) {
	var data []byte
	var err error

	if resolved != nil && resolved.Source.Type == SourceEmbedded && m.sourceResolver.EmbeddedFS != nil {
		// Embedded FS: look for shared manifest relative to rite
		data, err = fs.ReadFile(m.sourceResolver.EmbeddedFS, "rites/shared/manifest.yaml")
	} else {
		// Filesystem: resolve shared manifest from knossos-core ($KNOSSOS_HOME/rites/),
		// not from the project root. For satellite-local rites the project root is the
		// satellite directory, which has no rites/shared/. Shared rites always live in
		// $KNOSSOS_HOME/rites/ regardless of which project is being synced.
		sharedBase := m.sourceResolver.KnossosHome()
		if sharedBase == "" {
			// No KNOSSOS_HOME configured; fall back to project root (knossos-on-knossos
			// case where the project itself is knossos).
			sharedBase = m.resolver.ProjectRoot()
		}
		sharedPath := filepath.Join(sharedBase, "rites", "shared", "manifest.yaml")
		data, err = os.ReadFile(sharedPath)
	}
	if err != nil {
		return nil, nil // Shared manifest not found — graceful degradation
	}

	var manifest RiteManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// loadSharedHookDefaults loads hook_defaults from the shared rite manifest.
// Returns nil if the shared manifest doesn't exist or has no hook_defaults.
func (m *Materializer) loadSharedHookDefaults(resolved *ResolvedRite) *HookDefaults {
	manifest, err := m.loadSharedManifest(resolved)
	if err != nil {
		slog.Warn("failed to parse shared manifest for hook_defaults", "error", err)
		return nil
	}
	if manifest == nil {
		return nil
	}
	return manifest.HookDefaults
}
