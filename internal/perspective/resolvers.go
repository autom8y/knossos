package perspective

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/provenance"
)

// knownCCTools is the canonical set of CC-native tools. Mirrors agent.knownTools
// (which is unexported). We replicate here to avoid depending on unexported state.
var knownCCTools = map[string]bool{
	"Bash":            true,
	"Read":            true,
	"Write":           true,
	"Edit":            true,
	"Glob":            true,
	"Grep":            true,
	"Task":            true,
	"TodoWrite":       true,
	"TodoRead":        true,
	"WebSearch":       true,
	"WebFetch":        true,
	"Skill":           true,
	"NotebookEdit":    true,
	"AskUserQuestion": true,
}

// resolveIdentity resolves the L1 Identity layer from the parse context.
func resolveIdentity(ctx *ParseContext) *LayerEnvelope {
	fm := ctx.AgentFrontmatter
	raw := ctx.AgentFrontmatterRaw

	data := &IdentityData{
		Name:           fm.Name,
		Description:    fm.Description,
		Model:          fm.Model,
		Color:          fm.Color,
		MaxTurns:       fm.MaxTurns,
		PermissionMode: fm.PermissionMode,
	}

	// Extract knossos-only fields from raw map
	data.Role = stringFromMap(raw, "role")
	data.Type = stringFromMap(raw, "type")
	data.SchemaVersion = stringFromMap(raw, "schema_version")
	data.Aliases = stringSliceFromMap(raw, "aliases")

	// Body metrics
	lines := bytes.Count(ctx.AgentBody, []byte("\n"))
	if len(ctx.AgentBody) > 0 && !bytes.HasSuffix(ctx.AgentBody, []byte("\n")) {
		lines++ // count last line if no trailing newline
	}
	data.SystemPromptLines = lines

	excerpt := string(ctx.AgentBody)
	if len(excerpt) > 500 {
		excerpt = excerpt[:500]
	}
	data.SystemPromptExcerpt = excerpt

	// Archetype detection: check if the agent has a type field suggesting archetype origin.
	// In MVP, we do not trace back to knossos/archetypes/; set nil.
	data.ArchetypeSource = nil

	// Determine status
	var gaps []Gap
	if data.Name == "" {
		gaps = append(gaps, Gap{Field: "name", Reason: "name field is empty", Severity: SeverityMissing})
	}
	if data.Description == "" {
		gaps = append(gaps, Gap{Field: "description", Reason: "description field is empty", Severity: SeverityMissing})
	}

	status := StatusResolved
	if len(gaps) > 0 {
		status = StatusPartial
	}

	return &LayerEnvelope{
		Status: status,
		SourceFiles: []SourceRef{{
			Path:            ctx.AgentSourcePath,
			FieldsExtracted: []string{"name", "description", "role", "type", "model", "color", "aliases", "schema_version", "maxTurns", "permissionMode", "body"},
			ReadFrom:        "source",
		}},
		ResolutionMethod: "Parsed agent source frontmatter and body from " + ctx.AgentSourcePath,
		Gaps:             gaps,
		Data:             data,
	}
}

// resolveCapability resolves the L3 Capability layer from the parse context.
func resolveCapability(ctx *ParseContext) *LayerEnvelope {
	fm := ctx.AgentFrontmatter

	// Determine if agent_defaults contributed tools
	manifestDefaults := getAgentDefaults(ctx.RiteManifest)
	defaultTools := stringSliceFromMapAny(manifestDefaults, "tools")

	// REPLACE semantics: if agent declares tools, defaults don't apply
	var resolvedTools []string
	toolsFromDefaults := false
	if len(fm.Tools) > 0 {
		resolvedTools = []string(fm.Tools)
	} else if len(defaultTools) > 0 {
		resolvedTools = defaultTools
		toolsFromDefaults = true
	}

	// Classify tools
	var ccNative, unknown []string
	var mcpTools []MCPToolRef
	for _, tool := range resolvedTools {
		if knownCCTools[tool] {
			ccNative = append(ccNative, tool)
		} else if strings.HasPrefix(tool, "mcp:") {
			ref := parseMCPToolRef(tool, ctx.RiteManifest)
			mcpTools = append(mcpTools, ref)
		} else {
			unknown = append(unknown, tool)
		}
	}

	// Extract hooks from raw frontmatter
	hooks := extractHookSummaries(ctx.AgentFrontmatterRaw)

	sources := []SourceRef{{
		Path:            ctx.AgentSourcePath,
		FieldsExtracted: []string{"tools", "hooks"},
		ReadFrom:        "source",
	}}
	if toolsFromDefaults {
		sources = append(sources, SourceRef{
			Path:            filepath.Join(ctx.RiteSourcePath, "manifest.yaml"),
			FieldsExtracted: []string{"agent_defaults.tools"},
			ReadFrom:        "manifest",
		})
	}

	data := &CapabilityData{
		Tools:             resolvedTools,
		CCNativeTools:     ccNative,
		MCPTools:          mcpTools,
		UnknownTools:      unknown,
		ToolsFromDefaults: toolsFromDefaults,
		AgentDefaultTools: defaultTools,
		Hooks:             hooks,
	}

	return &LayerEnvelope{
		Status:           StatusResolved,
		SourceFiles:      sources,
		ResolutionMethod: "Parsed tools from agent frontmatter with REPLACE semantics against manifest agent_defaults",
		Data:             data,
	}
}

// resolveConstraint resolves the L4 Constraint layer from the parse context.
func resolveConstraint(ctx *ParseContext) *LayerEnvelope {
	fm := ctx.AgentFrontmatter
	raw := ctx.AgentFrontmatterRaw

	data := &ConstraintData{
		DisallowedTools: []string(fm.DisallowedTools),
	}
	if data.DisallowedTools == nil {
		data.DisallowedTools = []string{}
	}

	sources := []SourceRef{{
		Path:            ctx.AgentSourcePath,
		FieldsExtracted: []string{"disallowedTools", "write-guard", "contract"},
		ReadFrom:        "source",
	}}

	// Write-guard: replay 3-tier cascade
	data.WriteGuard = resolveWriteGuardCascade(ctx, raw)
	if data.WriteGuard != nil {
		sources = append(sources,
			SourceRef{
				Path:            filepath.Join(ctx.ProjectRoot, "rites", "shared", "manifest.yaml"),
				FieldsExtracted: []string{"hook_defaults.write_guard.allow_paths"},
				ReadFrom:        "manifest",
			},
			SourceRef{
				Path:            filepath.Join(ctx.RiteSourcePath, "manifest.yaml"),
				FieldsExtracted: []string{"hook_defaults.write_guard.extra_paths"},
				ReadFrom:        "manifest",
			},
		)
	}

	// Behavioral contract from source (knossos-only field)
	var gaps []Gap
	if contractRaw, ok := raw["contract"]; ok && contractRaw != nil {
		data.BehavioralContract = parseContractFromRaw(contractRaw)
	}
	if data.BehavioralContract == nil {
		gaps = append(gaps, Gap{
			Field:    "behavioral_contract",
			Reason:   "No contract defined in agent source frontmatter",
			Severity: SeverityMissing,
		})
	}

	status := StatusResolved
	if data.BehavioralContract == nil {
		status = StatusPartial
	}

	return &LayerEnvelope{
		Status:           status,
		SourceFiles:      sources,
		ResolutionMethod: "Parsed disallowedTools + 3-tier write-guard cascade + behavioral contract from source",
		Gaps:             gaps,
		Data:             data,
	}
}

// resolveMemory resolves the L5 Memory layer from the parse context.
func resolveMemory(ctx *ParseContext) *LayerEnvelope {
	fm := ctx.AgentFrontmatter
	agentName := ctx.AgentFrontmatter.Name
	if agentName == "" {
		agentName = filepath.Base(strings.TrimSuffix(ctx.AgentSourcePath, ".md"))
	}

	scope := fm.Memory.Scope()
	enabled := fm.Memory.IsEnabled()

	data := &MemoryData{
		Scope:   scope,
		Enabled: enabled,
	}

	sources := []SourceRef{{
		Path:            ctx.AgentSourcePath,
		FieldsExtracted: []string{"memory"},
		ReadFrom:        "source",
	}}

	var gaps []Gap

	// Check seed file
	seedPath := filepath.Join(ctx.ClaudeDir, "agent-memory", agentName, "MEMORY.md")
	seed := &MemorySeed{Path: seedPath}
	if info, err := os.Stat(seedPath); err == nil {
		seed.Exists = true
		lineCount := countFileLines(seedPath)
		seed.LineCount = &lineCount
		modTime := info.ModTime()
		seed.LastModified = &modTime
		sources = append(sources, SourceRef{
			Path:            seedPath,
			FieldsExtracted: []string{"seed_file.exists", "seed_file.line_count"},
			ReadFrom:        "memory_seed",
		})
	}
	data.SeedFile = seed

	// Runtime memory resolution
	runtime := &RuntimeMemory{Scope: scope}
	switch scope {
	case "user":
		homeDir, _ := os.UserHomeDir()
		runtimePath := filepath.Join(homeDir, ".claude", "memory", "MEMORY.md")
		runtime.ResolvedPath = runtimePath
		runtime.PathResolvable = true
		runtime.ContentAccessible = fileExists(runtimePath)
		if runtime.ContentAccessible {
			lc := countFileLines(runtimePath)
			runtime.ContentLineCount = &lc
		}
	case "local":
		localPath := filepath.Join(ctx.ClaudeDir, "agent-memory-local", agentName, "MEMORY.md")
		runtime.ResolvedPath = localPath
		runtime.PathResolvable = true
		runtime.ContentAccessible = fileExists(localPath)
		if runtime.ContentAccessible {
			lc := countFileLines(localPath)
			runtime.ContentLineCount = &lc
		}
	case "project":
		runtime.PathResolvable = false
		runtime.ContentAccessible = false
		gaps = append(gaps, Gap{
			Field:    "runtime_memory.resolved_path",
			Reason:   "Project-scope memory path depends on CC's opaque path hashing algorithm",
			Severity: SeverityOpaque,
		})
	}
	data.RuntimeMemory = runtime

	// Check agent-memory-local directory
	localPath := filepath.Join(ctx.ClaudeDir, "agent-memory-local", agentName, "MEMORY.md")
	local := &AgentMemoryLocal{Path: localPath}
	if fileExists(localPath) {
		local.Exists = true
		lc := countFileLines(localPath)
		local.LineCount = &lc
	}
	data.AgentMemoryLocal = local

	// Determine status
	status := StatusResolved
	if scope == "project" {
		status = StatusPartial // project scope has OPAQUE runtime path
	}

	return &LayerEnvelope{
		Status:           status,
		SourceFiles:      sources,
		ResolutionMethod: "Parsed memory scope from frontmatter, checked seed and runtime paths",
		Gaps:             gaps,
		Data:             data,
	}
}

// resolveProvenance resolves the L9 Provenance layer from the parse context.
func resolveProvenance(ctx *ParseContext) *LayerEnvelope {
	agentName := ctx.AgentFrontmatter.Name
	if agentName == "" {
		agentName = filepath.Base(strings.TrimSuffix(ctx.AgentSourcePath, ".md"))
	}

	manifestPath := provenance.ManifestPath(ctx.KnossosDir)
	entryKey := "agents/" + agentName + ".md"

	sources := []SourceRef{{
		Path:            manifestPath,
		FieldsExtracted: []string{"entries[" + entryKey + "]"},
		ReadFrom:        "provenance",
	}}

	entry, ok := ctx.Provenance.Entries[entryKey]
	if !ok {
		return &LayerEnvelope{
			Status:           StatusPartial,
			SourceFiles:      sources,
			ResolutionMethod: "Looked up provenance manifest entry for " + entryKey,
			Gaps: []Gap{{
				Field:    "entry",
				Reason:   "No provenance entry found for " + entryKey,
				Severity: SeverityMissing,
			}},
			Data: &ProvenanceData{
				ManifestPath: manifestPath,
			},
		}
	}

	// Compute divergence: checksum the materialized agent file and compare
	materializedPath := filepath.Join(ctx.ClaudeDir, "agents", agentName+".md")
	currentChecksum, _ := checksum.File(materializedPath)
	diverged := currentChecksum != "" && currentChecksum != entry.Checksum

	data := &ProvenanceData{
		Owner:        string(entry.Owner),
		Scope:        string(entry.Scope),
		SourcePath:   entry.SourcePath,
		SourceType:   entry.SourceType,
		Checksum:     entry.Checksum,
		LastSynced:   entry.LastSynced,
		Diverged:     diverged,
		ManifestPath: manifestPath,
	}

	return &LayerEnvelope{
		Status:           StatusResolved,
		SourceFiles:      sources,
		ResolutionMethod: "Looked up provenance entry and computed divergence against materialized file",
		Data:             data,
	}
}

// --- Helper functions ---

// stringFromMap extracts a string value from a map, returning "" if not found.
func stringFromMap(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// stringSliceFromMap extracts a string slice from a map.
func stringSliceFromMap(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	return toStringSlice(v)
}

// stringSliceFromMapAny extracts a string slice from a map[string]any.
func stringSliceFromMapAny(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	return toStringSlice(v)
}

// toStringSlice converts an interface{} (typically []any) to []string.
func toStringSlice(v any) []string {
	switch items := v.(type) {
	case []any:
		result := make([]string, 0, len(items))
		for _, item := range items {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return items
	case string:
		// Handle comma-separated string (FlexibleStringSlice format)
		parts := strings.Split(items, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	default:
		return nil
	}
}

// getAgentDefaults extracts the agent_defaults map from a manifest.
func getAgentDefaults(manifest map[string]any) map[string]any {
	if v, ok := manifest["agent_defaults"]; ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return map[string]any{}
}

// parseMCPToolRef parses an "mcp:server/method" reference and checks wiring.
func parseMCPToolRef(ref string, riteManifest map[string]any) MCPToolRef {
	trimmed := strings.TrimPrefix(ref, "mcp:")
	parts := strings.SplitN(trimmed, "/", 2)
	server := parts[0]
	method := ""
	if len(parts) > 1 {
		method = parts[1]
	}

	// Check if server is wired in manifest mcp_servers
	wired := isMCPServerWired(server, riteManifest)

	return MCPToolRef{
		Reference:   ref,
		Server:      server,
		Method:      method,
		ServerWired: wired,
	}
}

// isMCPServerWired checks if a server name appears in the manifest's mcp_servers list.
func isMCPServerWired(serverName string, manifest map[string]any) bool {
	servers, ok := manifest["mcp_servers"]
	if !ok {
		return false
	}
	serverList, ok := servers.([]any)
	if !ok {
		return false
	}
	for _, s := range serverList {
		if m, ok := s.(map[string]any); ok {
			if name, ok := m["name"].(string); ok && name == serverName {
				return true
			}
		}
	}
	return false
}

// extractHookSummaries extracts hook information from raw frontmatter.
func extractHookSummaries(raw map[string]any) []HookSummary {
	hooksRaw, ok := raw["hooks"]
	if !ok || hooksRaw == nil {
		return nil
	}

	hooksMap, ok := hooksRaw.(map[string]any)
	if !ok {
		return nil
	}

	var summaries []HookSummary
	for event, eventData := range hooksMap {
		items, ok := eventData.([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			summary := HookSummary{Event: event}

			// Drill into hooks array within the matcher group
			if hooks, ok := m["hooks"].([]any); ok {
				for _, h := range hooks {
					hm, ok := h.(map[string]any)
					if !ok {
						continue
					}
					if t, ok := hm["type"].(string); ok {
						summary.Type = t
					}
					if cmd, ok := hm["command"].(string); ok {
						if len(cmd) > 100 {
							cmd = cmd[:100]
						}
						summary.CommandExcerpt = cmd
						summary.IsWriteGuard = strings.Contains(cmd, "agent-guard")
					}
					summaries = append(summaries, summary)
				}
			} else {
				// Simple hook format (no matcher group)
				if t, ok := m["type"].(string); ok {
					summary.Type = t
				}
				if cmd, ok := m["command"].(string); ok {
					if len(cmd) > 100 {
						cmd = cmd[:100]
					}
					summary.CommandExcerpt = cmd
					summary.IsWriteGuard = strings.Contains(cmd, "agent-guard")
				}
				summaries = append(summaries, summary)
			}
		}
	}

	return summaries
}

// resolveWriteGuardCascade replays the 3-tier write-guard merge.
// This replicates read-only logic from materialize/hookdefaults.go.
func resolveWriteGuardCascade(ctx *ParseContext, raw map[string]any) *WriteGuardResolved {
	// Check if agent has write-guard field
	agentWG, hasWG := raw["write-guard"]
	if !hasWG || agentWG == nil {
		return nil
	}

	// Handle opt-out
	if b, ok := agentWG.(bool); ok && !b {
		return nil
	}

	// Tier 1: shared manifest hook_defaults.write_guard.allow_paths
	sharedPaths := getWriteGuardPaths(ctx.SharedManifest, "allow_paths")
	// Tier 2: rite manifest hook_defaults.write_guard.extra_paths
	ritePaths := getWriteGuardPaths(ctx.RiteManifest, "extra_paths")
	riteAllowPaths := getWriteGuardPaths(ctx.RiteManifest, "allow_paths")

	// Tier 3: agent frontmatter write-guard.extra-paths
	var agentPaths []string
	if m, ok := agentWG.(map[string]any); ok {
		agentPaths = stringSliceFromMapAny(m, "extra-paths")
	}

	// Merge all tiers (shared base + rite extras + rite allow + agent extras)
	var merged []string
	merged = append(merged, sharedPaths...)
	merged = append(merged, ritePaths...)
	merged = append(merged, riteAllowPaths...)
	merged = append(merged, agentPaths...)
	merged = dedupStrings(merged)

	// Resolve timeout
	timeout := 3 // default
	if t := getWriteGuardTimeout(ctx.SharedManifest); t > 0 {
		timeout = t
	}
	if t := getWriteGuardTimeout(ctx.RiteManifest); t > 0 {
		timeout = t
	}

	// Build generated command
	agentName := ctx.AgentFrontmatter.Name
	if agentName == "" {
		agentName = filepath.Base(strings.TrimSuffix(ctx.AgentSourcePath, ".md"))
	}
	var cmdParts []string
	cmdParts = append(cmdParts, "ari hook agent-guard")
	cmdParts = append(cmdParts, fmt.Sprintf("--agent %s", agentName))
	for _, p := range merged {
		cmdParts = append(cmdParts, fmt.Sprintf("--allow-path %s", p))
	}
	cmdParts = append(cmdParts, "--output json")

	return &WriteGuardResolved{
		Enabled:          true,
		AllowPaths:       merged,
		SharedBasePaths:  sharedPaths,
		RiteExtraPaths:   append(ritePaths, riteAllowPaths...),
		AgentExtraPaths:  agentPaths,
		Timeout:          timeout,
		GeneratedCommand: strings.Join(cmdParts, " "),
	}
}

// getWriteGuardPaths extracts write-guard paths from a manifest's hook_defaults.
func getWriteGuardPaths(manifest map[string]any, key string) []string {
	hookDefaults, ok := manifest["hook_defaults"]
	if !ok {
		return nil
	}
	hdMap, ok := hookDefaults.(map[string]any)
	if !ok {
		return nil
	}
	wg, ok := hdMap["write_guard"]
	if !ok {
		return nil
	}
	wgMap, ok := wg.(map[string]any)
	if !ok {
		return nil
	}
	return stringSliceFromMapAny(wgMap, key)
}

// getWriteGuardTimeout extracts the write-guard timeout from a manifest.
func getWriteGuardTimeout(manifest map[string]any) int {
	hookDefaults, ok := manifest["hook_defaults"]
	if !ok {
		return 0
	}
	hdMap, ok := hookDefaults.(map[string]any)
	if !ok {
		return 0
	}
	wg, ok := hdMap["write_guard"]
	if !ok {
		return 0
	}
	wgMap, ok := wg.(map[string]any)
	if !ok {
		return 0
	}
	if t, ok := wgMap["timeout"]; ok {
		switch v := t.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

// parseContractFromRaw parses a behavioral contract from the raw frontmatter value.
func parseContractFromRaw(v any) *BehavioralContractData {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}

	contract := &BehavioralContractData{
		Enforcement: "behavioral",
	}
	contract.MustUse = stringSliceFromMapAny(m, "must_use")
	contract.MustProduce = stringSliceFromMapAny(m, "must_produce")
	contract.MustNot = stringSliceFromMapAny(m, "must_not")
	if mt, ok := m["max_turns"]; ok {
		switch v := mt.(type) {
		case int:
			contract.MaxTurns = v
		case float64:
			contract.MaxTurns = int(v)
		}
	}

	return contract
}

// --- File utilities ---

// fileExists returns true if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// countFileLines counts the number of lines in a file.
func countFileLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte("\n"))
	if !bytes.HasSuffix(data, []byte("\n")) {
		lines++
	}
	return lines
}

// dedupStrings removes duplicate strings preserving first-occurrence order.
func dedupStrings(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

