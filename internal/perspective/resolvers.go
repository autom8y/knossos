package perspective

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/provenance"
)

// knownChannelTools is the canonical set of harness-native tools. Mirrors agent.knownTools
// (which is unexported). We replicate here to avoid depending on unexported state.
var knownChannelTools = map[string]bool{
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
	var channelNative, unknown []string
	var mcpTools []MCPToolRef
	for _, tool := range resolvedTools {
		switch {
		case knownChannelTools[tool]:
			channelNative = append(channelNative, tool)
		case strings.HasPrefix(tool, "mcp:"):
			ref := parseMCPToolRef(tool, ctx.RiteManifest)
			mcpTools = append(mcpTools, ref)
		default:
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
		ChannelNativeTools: channelNative,
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
	seedPath := filepath.Join(ctx.ChannelDir, "agent-memory", agentName, "MEMORY.md")
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
		localPath := filepath.Join(ctx.ChannelDir, "agent-memory-local", agentName, "MEMORY.md")
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
	localPath := filepath.Join(ctx.ChannelDir, "agent-memory-local", agentName, "MEMORY.md")
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
	materializedPath := filepath.Join(ctx.ChannelDir, "agents", agentName+".md")
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

// resolvePerception resolves the L2 Perception layer from the parse context.
// It replays skill policy evaluation in read-only mode, replicating the semantics
// of internal/materialize/skill_policies.go without importing that package.
func resolvePerception(ctx *ParseContext, capData *CapabilityData, conData *ConstraintData) *LayerEnvelope {
	raw := ctx.AgentFrontmatterRaw

	// --- Step 1: Extract agent's explicit skills from frontmatter ---
	explicitSkills := stringSliceFromMap(raw, "skills")
	if explicitSkills == nil {
		explicitSkills = []string{}
	}

	// --- Step 2: Parse skill policies from shared and rite manifests ---
	sharedPolicies := perceptionParseSkillPolicies(ctx.SharedManifest)
	ritePolicies := perceptionParseSkillPolicies(ctx.RiteManifest)
	mergedPolicies := perceptionMergeSkillPolicies(sharedPolicies, ritePolicies)

	// --- Step 3: Parse agent overrides and excludes ---
	excludeAll, excludeSet := perceptionParseSkillPolicyExclude(raw)
	overrideMap := perceptionParseSkillPolicyOverride(raw)

	// --- Step 4: Build lookup sets for predicate evaluation ---
	// Tools set: O(1) lookup for requires_tools predicate
	toolsSet := perceptionBuildSet(capData.Tools)
	// Disallowed set: O(1) lookup for requires_none predicate and dead reference guard
	disallowedSet := perceptionBuildSet(conData.DisallowedTools)

	// --- Step 5: Evaluate each merged policy ---
	var policyInjected []string
	var policyReferenced []string
	var effectivePolicies []SkillPolicyResult

	for _, policy := range mergedPolicies {
		result := SkillPolicyResult{
			Skill:         policy.skill,
			Mode:          policy.mode,
			EffectiveMode: policy.mode, // default to original mode
		}

		// Step 5a: Exclude check — exclude wins over everything
		if excludeAll || excludeSet[policy.skill] {
			result.Applied = false
			result.Reason = "excluded by skill_policy_exclude"
			effectivePolicies = append(effectivePolicies, result)
			continue
		}

		// Step 5b: Determine effective mode (agent override wins)
		if override, hasOverride := overrideMap[policy.skill]; hasOverride {
			result.EffectiveMode = override
		}

		// Step 5c: requires_tools predicate — skip if agent lacks ANY required tool
		skipped := false
		for _, req := range policy.requiresTools {
			if !toolsSet[req] {
				result.Applied = false
				result.Reason = "missing required tool: " + req
				effectivePolicies = append(effectivePolicies, result)
				skipped = true
				break
			}
		}
		if skipped {
			continue
		}

		// Step 5d: requires_none predicate — skip if agent HAS any of these in disallowedTools
		for _, none := range policy.requiresNone {
			if disallowedSet[none] {
				result.Applied = false
				result.Reason = "blocked by requires_none: " + none + " in disallowedTools"
				effectivePolicies = append(effectivePolicies, result)
				skipped = true
				break
			}
		}
		if skipped {
			continue
		}

		// Step 5e: Mode application
		switch result.EffectiveMode {
		case "inject":
			result.Applied = true
			policyInjected = append(policyInjected, policy.skill)
		case "reference":
			// Dead reference guard: if Skill tool is disallowed, skip
			if disallowedSet["Skill"] {
				result.Applied = false
				result.Reason = "dead reference: Skill tool disallowed"
			} else {
				result.Applied = true
				policyReferenced = append(policyReferenced, policy.skill)
			}
		default:
			// Unknown mode: record as not applied
			result.Applied = false
			result.Reason = "unknown mode: " + result.EffectiveMode
		}

		effectivePolicies = append(effectivePolicies, result)
	}

	// Normalise nil slices to empty slices for consistent output
	if policyInjected == nil {
		policyInjected = []string{}
	}
	if policyReferenced == nil {
		policyReferenced = []string{}
	}

	// --- Step 6: Determine SkillToolAvailable ---
	skillToolAvailable := !disallowedSet["Skill"]

	// --- Step 7: Compute on-demand skills ---
	// On-demand = materialized skills NOT already in explicit, injected, or referenced sets.
	// Only meaningful if the agent can invoke the Skill tool.
	var onDemandSkills []string
	if skillToolAvailable {
		preloadedSet := make(map[string]bool, len(explicitSkills)+len(policyInjected)+len(policyReferenced))
		for _, s := range explicitSkills {
			preloadedSet[s] = true
		}
		for _, s := range policyInjected {
			preloadedSet[s] = true
		}
		for _, s := range policyReferenced {
			preloadedSet[s] = true
		}
		for _, dir := range ctx.MaterializedSkillsDirs {
			if !preloadedSet[dir] {
				onDemandSkills = append(onDemandSkills, dir)
			}
		}
	}
	if onDemandSkills == nil {
		onDemandSkills = []string{}
	}

	// --- Step 8: Compute totals ---
	totalPreloaded := len(explicitSkills) + len(policyInjected)
	totalReachable := totalPreloaded + len(policyReferenced)
	if skillToolAvailable {
		totalReachable += len(onDemandSkills)
	}

	// --- Step 9: Record source refs ---
	sources := []SourceRef{
		{
			Path:            ctx.AgentSourcePath,
			FieldsExtracted: []string{"skills", "skill_policy_exclude", "skill_policy_override"},
			ReadFrom:        "source",
		},
		{
			Path:            filepath.Join(ctx.ProjectRoot, "rites", "shared", "manifest.yaml"),
			FieldsExtracted: []string{"skill_policies"},
			ReadFrom:        "manifest",
		},
		{
			Path:            filepath.Join(ctx.RiteSourcePath, "manifest.yaml"),
			FieldsExtracted: []string{"skill_policies"},
			ReadFrom:        "manifest",
		},
		{
			Path:            filepath.Join(ctx.ChannelDir, "skills"),
			FieldsExtracted: []string{"materialized_skill_dirs"},
			ReadFrom:        "materialized",
		},
	}

	data := &PerceptionData{
		ExplicitSkills:         explicitSkills,
		PolicyInjectedSkills:   policyInjected,
		PolicyReferencedSkills: policyReferenced,
		OnDemandSkills:         onDemandSkills,
		SkillToolAvailable:     skillToolAvailable,
		TotalPreloaded:         totalPreloaded,
		TotalReachable:         totalReachable,
		EffectivePolicies:      effectivePolicies,
	}

	return &LayerEnvelope{
		Status:           StatusResolved,
		SourceFiles:      sources,
		ResolutionMethod: "Replayed skill policy evaluation against L3 tools and L4 disallowedTools; enumerated materialized skills directory",
		Data:             data,
	}
}

// --- Perception resolver helpers (private) ---

// perceptionSkillPolicy is a local representation of a skill policy entry,
// avoiding any import of internal/materialize.
type perceptionSkillPolicy struct {
	skill         string
	mode          string
	requiresTools []string
	requiresNone  []string
}

// perceptionParseSkillPolicies extracts skill_policies from a manifest map.
// Returns nil if the key is absent or the value cannot be parsed.
func perceptionParseSkillPolicies(manifest map[string]any) []perceptionSkillPolicy {
	raw, ok := manifest["skill_policies"]
	if !ok || raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	policies := make([]perceptionSkillPolicy, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		skillName, _ := m["skill"].(string)
		mode, _ := m["mode"].(string)
		if skillName == "" || mode == "" {
			continue
		}
		p := perceptionSkillPolicy{
			skill:         skillName,
			mode:          mode,
			requiresTools: stringSliceFromMapAny(m, "requires_tools"),
			requiresNone:  stringSliceFromMapAny(m, "requires_none"),
		}
		policies = append(policies, p)
	}
	return policies
}

// perceptionMergeSkillPolicies merges shared and rite-level policies.
// Rite policies override shared policies for the same skill name.
// Preserves order: shared-first, rite-appended for non-overridden rite policies.
// This replicates MergeSkillPolicies from internal/materialize/skill_policies.go.
func perceptionMergeSkillPolicies(shared, rite []perceptionSkillPolicy) []perceptionSkillPolicy {
	if len(shared) == 0 && len(rite) == 0 {
		return nil
	}

	// Track shared policy positions for in-place override
	sharedBySkill := make(map[string]int, len(shared))
	result := make([]perceptionSkillPolicy, 0, len(shared)+len(rite))

	for i, p := range shared {
		result = append(result, p)
		sharedBySkill[p.skill] = i
	}

	for _, p := range rite {
		if idx, exists := sharedBySkill[p.skill]; exists {
			// Rite wins: replace in-place to preserve shared ordering
			result[idx] = p
		} else {
			result = append(result, p)
		}
	}

	return result
}

// perceptionParseSkillPolicyExclude reads the skill_policy_exclude field.
// Replicates parseSkillPolicyExclude from internal/materialize/skill_policies.go.
func perceptionParseSkillPolicyExclude(fmMap map[string]any) (excludeAll bool, excludeSet map[string]bool) {
	val, ok := fmMap["skill_policy_exclude"]
	if !ok {
		return false, nil
	}

	// String "all" means exclude all policies
	if s, ok := val.(string); ok {
		if s == "all" {
			return true, nil
		}
		// Single string (not "all"): treat as single-item exclusion set
		return false, map[string]bool{s: true}
	}

	// String slice: build exclusion set
	items := toStringSlice(val)
	if len(items) == 0 {
		return false, nil
	}
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return false, set
}

// perceptionParseSkillPolicyOverride reads the skill_policy_override field.
// Replicates parseSkillPolicyOverride from internal/materialize/skill_policies.go.
func perceptionParseSkillPolicyOverride(fmMap map[string]any) map[string]string {
	val, ok := fmMap["skill_policy_override"]
	if !ok {
		return nil
	}
	items, ok := val.([]any)
	if !ok {
		return nil
	}
	overrides := make(map[string]string, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		skillName, ok1 := entry["skill"].(string)
		modeName, ok2 := entry["mode"].(string)
		if ok1 && ok2 && skillName != "" && modeName != "" {
			overrides[skillName] = modeName
		}
	}
	if len(overrides) == 0 {
		return nil
	}
	return overrides
}

// perceptionBuildSet constructs a boolean set from a string slice for O(1) lookup.
func perceptionBuildSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
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

// resolvePosition resolves the L6 Position layer from the parse context.
// It cross-references workflow.yaml, orchestrator.yaml, and the rite manifest to
// determine where this agent sits in the workflow graph.
func resolvePosition(ctx *ParseContext) *LayerEnvelope {
	// Determine agent name for matching
	agentName := ctx.AgentFrontmatter.Name
	if agentName == "" {
		agentName = filepath.Base(strings.TrimSuffix(ctx.AgentSourcePath, ".md"))
	}

	sources := []SourceRef{
		{Path: ctx.AgentSourcePath, FieldsExtracted: []string{"upstream", "downstream"}, ReadFrom: "source"},
		{Path: filepath.Join(ctx.RiteSourcePath, "workflow.yaml"), FieldsExtracted: []string{"phases", "entry_point", "back_routes", "complexity_levels"}, ReadFrom: "manifest"},
		{Path: filepath.Join(ctx.RiteSourcePath, "orchestrator.yaml"), FieldsExtracted: []string{"handoff_criteria"}, ReadFrom: "manifest"},
		{Path: filepath.Join(ctx.RiteSourcePath, "manifest.yaml"), FieldsExtracted: []string{"entry_agent", "complexity_levels"}, ReadFrom: "manifest"},
	}

	data := &PositionData{}

	// Extract phases list from workflow
	phases := extractPhasesList(ctx.Workflow)
	data.TotalPhases = len(phases)

	// Find this agent's phase by matching the agent field
	matchedPhaseIdx := -1
	matchedPhaseName := ""
	for i, phase := range phases {
		if agentField, ok := phase["agent"].(string); ok && agentField == agentName {
			matchedPhaseIdx = i
			matchedPhaseName = stringFromMap(phase, "name")
			data.WorkflowPhase = matchedPhaseName
			data.PhaseIndex = i
			data.PhaseProduces = stringFromMap(phase, "produces")
			data.PhaseCondition = stringFromMap(phase, "condition")
			data.InWorkflow = true
			break
		}
	}

	// Determine predecessor and successor if the agent was found in the workflow
	if matchedPhaseIdx >= 0 {
		// Predecessor: the agent from the previous phase (if any)
		if matchedPhaseIdx > 0 {
			if prevAgent, ok := phases[matchedPhaseIdx-1]["agent"].(string); ok {
				data.PhasePredecessor = prevAgent
			}
		}

		// Successor: resolve from the next field (phase name), then look up that phase's agent
		nextPhaseName := stringFromMap(phases[matchedPhaseIdx], "next")
		if nextPhaseName != "" && nextPhaseName != "null" {
			for _, phase := range phases {
				if stringFromMap(phase, "name") == nextPhaseName {
					if nextAgent, ok := phase["agent"].(string); ok {
						data.PhaseSuccessor = nextAgent
					}
					break
				}
			}
		}
	}

	// Check if this agent is the workflow entry_point
	if ep, ok := ctx.Workflow["entry_point"].(map[string]any); ok {
		if epAgent, ok := ep["agent"].(string); ok && epAgent == agentName {
			data.IsEntryPoint = true
		}
	}

	// Check if this agent is the rite manifest entry_agent
	if entryAgent, ok := ctx.RiteManifest["entry_agent"].(string); ok && entryAgent == agentName {
		data.IsEntryAgent = true
	}

	// Collect back_routes targeting this agent
	if backRoutes, ok := ctx.Workflow["back_routes"].([]any); ok {
		for _, br := range backRoutes {
			brMap, ok := br.(map[string]any)
			if !ok {
				continue
			}
			targetAgent, _ := brMap["target_agent"].(string)
			if targetAgent != agentName {
				continue
			}
			route := BackRoute{
				SourcePhase: stringFromMap(brMap, "source_phase"),
				Trigger:     stringFromMap(brMap, "trigger"),
				Condition:   stringFromMap(brMap, "condition"),
			}
			// requires_user_confirmation can be bool
			if ruc, ok := brMap["requires_user_confirmation"].(bool); ok {
				route.RequiresUserConfirmation = ruc
			}
			data.BackRoutes = append(data.BackRoutes, route)
		}
	}

	// Collect complexity gates: levels where this agent's phase is included
	complexityLevels := extractComplexityLevels(ctx.Workflow, ctx.RiteManifest)
	if matchedPhaseName != "" {
		for levelName, phasesInLevel := range complexityLevels {
			if slices.Contains(phasesInLevel, matchedPhaseName) {
				data.ComplexityGates = append(data.ComplexityGates, levelName)
			}
		}
	}

	// Collect handoff criteria from orchestrator for this agent's phase
	if matchedPhaseName != "" {
		if criteria, ok := ctx.Orchestrator["handoff_criteria"].(map[string]any); ok {
			if phaseItems, ok := criteria[matchedPhaseName]; ok {
				data.HandoffCriteria = toStringSlice(phaseItems)
			}
		}
	}

	// Determine status
	var gaps []Gap
	status := StatusResolved
	if !data.InWorkflow {
		status = StatusPartial
		gaps = append(gaps, Gap{
			Field:    "workflow_phase",
			Reason:   "Agent " + agentName + " not found in any workflow phase",
			Severity: SeverityMissing,
		})
	}

	return &LayerEnvelope{
		Status:           status,
		SourceFiles:      sources,
		ResolutionMethod: "Cross-referenced workflow.yaml phases, orchestrator.yaml handoff_criteria, and rite manifest entry_agent",
		Gaps:             gaps,
		Data:             data,
	}
}

// resolveSurface resolves the L7 Surface layer from the parse context.
// It captures the agent's I/O surface: dromena, legomena, artifact types, and commands.
func resolveSurface(ctx *ParseContext) *LayerEnvelope {
	// Determine agent name for command matching
	agentName := ctx.AgentFrontmatter.Name
	if agentName == "" {
		agentName = filepath.Base(strings.TrimSuffix(ctx.AgentSourcePath, ".md"))
	}

	sources := []SourceRef{
		{Path: filepath.Join(ctx.RiteSourcePath, "manifest.yaml"), FieldsExtracted: []string{"dromena", "legomena"}, ReadFrom: "manifest"},
		{Path: filepath.Join(ctx.RiteSourcePath, "workflow.yaml"), FieldsExtracted: []string{"phases[agent].produces", "commands"}, ReadFrom: "manifest"},
		{Path: ctx.AgentSourcePath, FieldsExtracted: []string{"contract.must_produce"}, ReadFrom: "source"},
	}

	data := &SurfaceData{}

	// Dromena from rite manifest
	data.DromenaOwned = stringSliceFromMap(ctx.RiteManifest, "dromena")

	// Legomena from rite manifest
	data.LegomenaAvailable = stringSliceFromMap(ctx.RiteManifest, "legomena")

	// Artifact types: extract from the workflow phase this agent owns
	phases := extractPhasesList(ctx.Workflow)
	for _, phase := range phases {
		if agentField, ok := phase["agent"].(string); ok && agentField == agentName {
			if produces := stringFromMap(phase, "produces"); produces != "" {
				data.ArtifactTypes = []string{produces}
			}
			break
		}
	}

	// Contract must_produce from agent frontmatter
	if contractRaw, ok := ctx.AgentFrontmatterRaw["contract"]; ok && contractRaw != nil {
		if contractMap, ok := contractRaw.(map[string]any); ok {
			data.ContractMustProduce = stringSliceFromMapAny(contractMap, "must_produce")
		}
	}

	// Commands from workflow.yaml where primary_agent matches
	if commands, ok := ctx.Workflow["commands"].([]any); ok {
		for _, cmd := range commands {
			cmdMap, ok := cmd.(map[string]any)
			if !ok {
				continue
			}
			primaryAgent, _ := cmdMap["primary_agent"].(string)
			if primaryAgent != agentName {
				continue
			}
			ref := CommandRef{
				Name:        stringFromMap(cmdMap, "name"),
				File:        stringFromMap(cmdMap, "file"),
				Description: stringFromMap(cmdMap, "description"),
			}
			data.Commands = append(data.Commands, ref)
		}
	}

	return &LayerEnvelope{
		Status:           StatusResolved,
		SourceFiles:      sources,
		ResolutionMethod: "Extracted dromena/legomena from manifest, artifact types from workflow phase, commands from workflow commands",
		Data:             data,
	}
}

// resolveHorizon resolves the L8 Horizon layer by computing the inverse/negative
// space across all resolved layers. It answers: "What can this agent NOT do?"
func resolveHorizon(ctx *ParseContext, doc *PerspectiveDocument) *LayerEnvelope {
	agentName := ctx.AgentFrontmatter.Name
	if agentName == "" {
		agentName = filepath.Base(strings.TrimSuffix(ctx.AgentSourcePath, ".md"))
	}

	data := &HorizonData{}

	// Tools not available: all CC-native tools NOT in agent's tools list
	l3 := getLayerData[*CapabilityData](doc, "L3")
	if l3 != nil {
		agentTools := perceptionBuildSet(l3.Tools)
		for tool := range knownChannelTools {
			if !agentTools[tool] {
				data.ToolsNotAvailable = append(data.ToolsNotAvailable, tool)
			}
		}
		sort.Strings(data.ToolsNotAvailable)
	}

	// Disallowed overlap: tools in BOTH L3.Tools AND L4.DisallowedTools
	l4 := getLayerData[*ConstraintData](doc, "L4")
	if l3 != nil && l4 != nil {
		disallowed := perceptionBuildSet(l4.DisallowedTools)
		for _, tool := range l3.Tools {
			if disallowed[tool] {
				data.DisallowedOverlap = append(data.DisallowedOverlap, tool)
			}
		}
	}

	// Skills unreachable: skills in MaterializedSkillsDirs but NOT reachable by agent
	l2 := getLayerData[*PerceptionData](doc, "L2")
	if l2 != nil {
		reachableSet := make(map[string]bool)
		for _, s := range l2.ExplicitSkills {
			reachableSet[s] = true
		}
		for _, s := range l2.PolicyInjectedSkills {
			reachableSet[s] = true
		}
		for _, s := range l2.PolicyReferencedSkills {
			reachableSet[s] = true
		}
		for _, s := range l2.OnDemandSkills {
			reachableSet[s] = true
		}
		for _, dir := range ctx.MaterializedSkillsDirs {
			if !reachableSet[dir] {
				data.SkillsUnreachable = append(data.SkillsUnreachable, dir)
			}
		}
	}

	// Phases not in: workflow phases where agent != this agent
	l6 := getLayerData[*PositionData](doc, "L6")
	phases := extractPhasesList(ctx.Workflow)
	if l6 != nil {
		for _, phase := range phases {
			phaseAgent, _ := phase["agent"].(string)
			phaseName := stringFromMap(phase, "name")
			if phaseAgent != agentName && phaseName != "" {
				data.PhasesNotIn = append(data.PhasesNotIn, phaseName)
			}
		}
	}

	// Memory blind spots
	l5 := getLayerData[*MemoryData](doc, "L5")
	if l5 != nil {
		allScopes := []string{"user", "project", "local"}
		if !l5.Enabled {
			data.MemoryBlindSpots = append(data.MemoryBlindSpots, "memory disabled — all scopes blind")
		} else {
			for _, scope := range allScopes {
				if l5.Scope != scope {
					data.MemoryBlindSpots = append(data.MemoryBlindSpots, scope+" scope not configured")
				}
			}
		}
	}

	// Surface gaps: dromena/legomena in rite that agent doesn't own
	// For now, surface gaps = rite-level dromena/legomena not in agent's surface
	// (agent-level ownership is hard to determine without per-agent dromena assignment)
	// Skip surface gaps for MVP — the data model doesn't assign dromena per-agent

	// Normalize nil slices
	if data.ToolsNotAvailable == nil {
		data.ToolsNotAvailable = []string{}
	}
	if data.DisallowedOverlap == nil {
		data.DisallowedOverlap = []string{}
	}
	if data.SkillsUnreachable == nil {
		data.SkillsUnreachable = []string{}
	}
	if data.PhasesNotIn == nil {
		data.PhasesNotIn = []string{}
	}
	if data.MemoryBlindSpots == nil {
		data.MemoryBlindSpots = []string{}
	}
	if data.SurfaceGaps == nil {
		data.SurfaceGaps = []string{}
	}

	return &LayerEnvelope{
		Status:           StatusResolved,
		SourceFiles:      []SourceRef{{Path: "computed", FieldsExtracted: []string{"inverse of L2-L7"}, ReadFrom: "computed"}},
		ResolutionMethod: "Computed inverse/negative space across L2-L7 resolved data",
		Data:             data,
	}
}

// extractPhasesList returns the phases array from a workflow map as []map[string]any.
func extractPhasesList(workflow map[string]any) []map[string]any {
	phasesRaw, ok := workflow["phases"]
	if !ok {
		return nil
	}
	phasesSlice, ok := phasesRaw.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(phasesSlice))
	for _, p := range phasesSlice {
		if m, ok := p.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

// extractComplexityLevels returns a map of level name → []phase names from
// workflow complexity_levels or rite manifest complexity_levels.
func extractComplexityLevels(workflow, riteManifest map[string]any) map[string][]string {
	result := make(map[string][]string)

	// Try workflow first, then rite manifest
	sources := []map[string]any{workflow, riteManifest}
	for _, src := range sources {
		levelsRaw, ok := src["complexity_levels"]
		if !ok {
			continue
		}
		levelsSlice, ok := levelsRaw.([]any)
		if !ok {
			continue
		}
		for _, lvl := range levelsSlice {
			lvlMap, ok := lvl.(map[string]any)
			if !ok {
				continue
			}
			name := stringFromMap(lvlMap, "name")
			if name == "" {
				continue
			}
			phases := toStringSlice(lvlMap["phases"])
			result[name] = phases
		}
		// Only read from first source that has the field
		if len(result) > 0 {
			break
		}
	}
	return result
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

