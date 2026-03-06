// Package perspective assembles a first-person experiential view of an agent's
// context by resolving identity, capability, constraint, memory, and provenance
// layers from source files. It reads rite source data (not materialized output)
// to capture knossos-only fields stripped during materialization.
package perspective

import "time"

// PerspectiveDocument is the top-level container for an agent's perspective view.
type PerspectiveDocument struct {
	Version          string                    `json:"version" yaml:"version"`
	GeneratedAt      time.Time                 `json:"generated_at" yaml:"generated_at"`
	Agent            string                    `json:"agent" yaml:"agent"`
	Rite             string                    `json:"rite" yaml:"rite"`
	SourcePath       string                    `json:"source_path" yaml:"source_path"`
	Mode             string                    `json:"mode" yaml:"mode"`
	Layers           map[string]*LayerEnvelope `json:"layers" yaml:"layers"`
	AssemblyMetadata AssemblyMetadata          `json:"assembly_metadata" yaml:"assembly_metadata"`
	AuditOverlay     *AuditOverlay             `json:"audit,omitempty" yaml:"audit,omitempty"`
	SimulateOverlay  *SimulateOverlay          `json:"simulate,omitempty" yaml:"simulate,omitempty"`
}

// LayerEnvelope is the uniform wrapper for every resolved layer.
type LayerEnvelope struct {
	Status           LayerStatus `json:"status" yaml:"status"`
	SourceFiles      []SourceRef `json:"source_files" yaml:"source_files"`
	ResolutionMethod string      `json:"resolution_method" yaml:"resolution_method"`
	Gaps             []Gap       `json:"gaps,omitempty" yaml:"gaps,omitempty"`
	Data             any         `json:"data" yaml:"data"`
}

// LayerStatus represents the resolution status of a layer.
type LayerStatus string

const (
	StatusResolved LayerStatus = "RESOLVED"
	StatusPartial  LayerStatus = "PARTIAL"
	StatusOpaque   LayerStatus = "OPAQUE"
	StatusFailed   LayerStatus = "FAILED"
)

// AssemblyMetadata captures pipeline-level resolution statistics.
type AssemblyMetadata struct {
	ResolutionTimeMs int `json:"resolution_time_ms" yaml:"resolution_time_ms"`
	LayersResolved   int `json:"layers_resolved" yaml:"layers_resolved"`
	LayersDegraded   int `json:"layers_degraded" yaml:"layers_degraded"`
	LayersFailed     int `json:"layers_failed" yaml:"layers_failed"`
}

// SourceRef records a file read during layer resolution.
type SourceRef struct {
	Path            string   `json:"path" yaml:"path"`
	FieldsExtracted []string `json:"fields_extracted" yaml:"fields_extracted"`
	ReadFrom        string   `json:"read_from" yaml:"read_from"` // source, materialized, manifest, provenance, memory_seed, runtime
}

// Gap records a field that could not be resolved and why.
type Gap struct {
	Field    string      `json:"field" yaml:"field"`
	Reason   string      `json:"reason" yaml:"reason"`
	Severity GapSeverity `json:"severity" yaml:"severity"`
}

// GapSeverity classifies the nature of an unresolved field.
type GapSeverity string

const (
	SeverityOpaque  GapSeverity = "OPAQUE"
	SeverityMissing GapSeverity = "MISSING"
	SeverityStale   GapSeverity = "STALE"
)

// --- Layer-specific data types (MVP: L1, L3, L4, L5, L9) ---

// IdentityData contains the agent's identity fields (L1).
type IdentityData struct {
	Name                string   `json:"name" yaml:"name"`
	Description         string   `json:"description" yaml:"description"`
	Role                string   `json:"role,omitempty" yaml:"role,omitempty"`
	Type                string   `json:"type,omitempty" yaml:"type,omitempty"`
	Model               string   `json:"model,omitempty" yaml:"model,omitempty"`
	Color               string   `json:"color,omitempty" yaml:"color,omitempty"`
	Aliases             []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	SchemaVersion       string   `json:"schema_version,omitempty" yaml:"schema_version,omitempty"`
	MaxTurns            int      `json:"max_turns,omitempty" yaml:"max_turns,omitempty"`
	PermissionMode      string   `json:"permission_mode,omitempty" yaml:"permission_mode,omitempty"`
	SystemPromptExcerpt string   `json:"system_prompt_excerpt,omitempty" yaml:"system_prompt_excerpt,omitempty"`
	SystemPromptLines   int      `json:"system_prompt_lines" yaml:"system_prompt_lines"`
	ArchetypeSource     *string  `json:"archetype_source" yaml:"archetype_source"` // nil if not archetype-based
}

// CapabilityData contains the agent's tool and hook configuration (L3).
type CapabilityData struct {
	Tools             []string      `json:"tools" yaml:"tools"`
	CCNativeTools     []string      `json:"cc_native_tools" yaml:"cc_native_tools"`
	MCPTools          []MCPToolRef  `json:"mcp_tools" yaml:"mcp_tools"`
	UnknownTools      []string      `json:"unknown_tools,omitempty" yaml:"unknown_tools,omitempty"`
	ToolsFromDefaults bool          `json:"tools_from_defaults" yaml:"tools_from_defaults"`
	AgentDefaultTools []string      `json:"agent_defaults_tools,omitempty" yaml:"agent_defaults_tools,omitempty"`
	Hooks             []HookSummary `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

// MCPToolRef describes a single MCP tool reference.
type MCPToolRef struct {
	Reference   string `json:"reference" yaml:"reference"`
	Server      string `json:"server" yaml:"server"`
	Method      string `json:"method,omitempty" yaml:"method,omitempty"`
	ServerWired bool   `json:"server_wired" yaml:"server_wired"`
}

// HookSummary describes a hook declared in agent frontmatter.
type HookSummary struct {
	Event          string `json:"event" yaml:"event"`
	Type           string `json:"type,omitempty" yaml:"type,omitempty"`
	CommandExcerpt string `json:"command_excerpt,omitempty" yaml:"command_excerpt,omitempty"`
	IsWriteGuard   bool   `json:"is_write_guard" yaml:"is_write_guard"`
}

// ConstraintData contains the agent's restriction configuration (L4).
type ConstraintData struct {
	DisallowedTools    []string                `json:"disallowed_tools" yaml:"disallowed_tools"`
	WriteGuard         *WriteGuardResolved     `json:"write_guard" yaml:"write_guard"`
	BehavioralContract *BehavioralContractData `json:"behavioral_contract" yaml:"behavioral_contract"`
}

// WriteGuardResolved contains the fully merged write-guard configuration.
type WriteGuardResolved struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowPaths       []string `json:"allow_paths" yaml:"allow_paths"`
	SharedBasePaths  []string `json:"shared_base_paths,omitempty" yaml:"shared_base_paths,omitempty"`
	RiteExtraPaths   []string `json:"rite_extra_paths,omitempty" yaml:"rite_extra_paths,omitempty"`
	AgentExtraPaths  []string `json:"agent_extra_paths,omitempty" yaml:"agent_extra_paths,omitempty"`
	Timeout          int      `json:"timeout" yaml:"timeout"`
	GeneratedCommand string   `json:"generated_command,omitempty" yaml:"generated_command,omitempty"`
}

// BehavioralContractData contains behavioral constraints from source frontmatter.
type BehavioralContractData struct {
	MustUse     []string `json:"must_use,omitempty" yaml:"must_use,omitempty"`
	MustProduce []string `json:"must_produce,omitempty" yaml:"must_produce,omitempty"`
	MustNot     []string `json:"must_not,omitempty" yaml:"must_not,omitempty"`
	MaxTurns    int      `json:"max_turns,omitempty" yaml:"max_turns,omitempty"`
	Enforcement string   `json:"enforcement" yaml:"enforcement"` // always "behavioral"
}

// MemoryData contains the agent's memory configuration (L5).
type MemoryData struct {
	Scope            string            `json:"scope" yaml:"scope"` // user, project, local, or "" (disabled)
	Enabled          bool              `json:"enabled" yaml:"enabled"`
	SeedFile         *MemorySeed       `json:"seed_file" yaml:"seed_file"`
	RuntimeMemory    *RuntimeMemory    `json:"runtime_memory" yaml:"runtime_memory"`
	AgentMemoryLocal *AgentMemoryLocal `json:"agent_memory_local" yaml:"agent_memory_local"`
}

// MemorySeed describes the knossos-managed seed file.
type MemorySeed struct {
	Path         string     `json:"path" yaml:"path"`
	Exists       bool       `json:"exists" yaml:"exists"`
	LineCount    *int       `json:"line_count" yaml:"line_count"`
	LastModified *time.Time `json:"last_modified" yaml:"last_modified"`
}

// RuntimeMemory describes where CC stores the agent's runtime memory.
type RuntimeMemory struct {
	Scope             string `json:"scope" yaml:"scope"`
	ResolvedPath      string `json:"resolved_path" yaml:"resolved_path"`       // empty for project scope
	PathResolvable    bool   `json:"path_resolvable" yaml:"path_resolvable"`   // false for project scope
	ContentAccessible bool   `json:"content_accessible" yaml:"content_accessible"`
	ContentLineCount  *int   `json:"content_line_count" yaml:"content_line_count"`
}

// AgentMemoryLocal describes the local-scope memory file.
type AgentMemoryLocal struct {
	Path      string `json:"path" yaml:"path"`
	Exists    bool   `json:"exists" yaml:"exists"`
	LineCount *int   `json:"line_count" yaml:"line_count"`
}

// ProvenanceData contains the agent's provenance tracking information (L9).
type ProvenanceData struct {
	Owner        string    `json:"owner" yaml:"owner"`
	Scope        string    `json:"scope" yaml:"scope"`
	SourcePath   string    `json:"source_path" yaml:"source_path"`
	SourceType   string    `json:"source_type" yaml:"source_type"`
	Checksum     string    `json:"checksum" yaml:"checksum"`
	LastSynced   time.Time `json:"last_synced" yaml:"last_synced"`
	Diverged     bool      `json:"diverged" yaml:"diverged"`
	ManifestPath string    `json:"manifest_path" yaml:"manifest_path"`
}

// PerceptionData contains the agent's skill awareness configuration (L2).
type PerceptionData struct {
	ExplicitSkills         []string            `json:"explicit_skills" yaml:"explicit_skills"`
	PolicyInjectedSkills   []string            `json:"policy_injected_skills" yaml:"policy_injected_skills"`
	PolicyReferencedSkills []string            `json:"policy_referenced_skills" yaml:"policy_referenced_skills"`
	OnDemandSkills         []string            `json:"on_demand_skills" yaml:"on_demand_skills"`
	SkillToolAvailable     bool                `json:"skill_tool_available" yaml:"skill_tool_available"`
	TotalPreloaded         int                 `json:"total_preloaded" yaml:"total_preloaded"`
	TotalReachable         int                 `json:"total_reachable" yaml:"total_reachable"`
	EffectivePolicies      []SkillPolicyResult `json:"effective_policies,omitempty" yaml:"effective_policies,omitempty"`
}

// SkillPolicyResult describes the resolved skill policy for a single skill.
type SkillPolicyResult struct {
	Skill         string `json:"skill" yaml:"skill"`
	Mode          string `json:"mode" yaml:"mode"`
	EffectiveMode string `json:"effective_mode" yaml:"effective_mode"`
	Applied       bool   `json:"applied" yaml:"applied"`
	Reason        string `json:"reason" yaml:"reason"`
}

// PositionData contains the agent's position in the workflow (L6).
type PositionData struct {
	WorkflowPhase    string      `json:"workflow_phase" yaml:"workflow_phase"`
	PhaseIndex       int         `json:"phase_index" yaml:"phase_index"`
	TotalPhases      int         `json:"total_phases" yaml:"total_phases"`
	IsEntryPoint     bool        `json:"is_entry_point" yaml:"is_entry_point"`
	IsEntryAgent     bool        `json:"is_entry_agent" yaml:"is_entry_agent"`
	InWorkflow       bool        `json:"in_workflow" yaml:"in_workflow"`
	PhasePredecessor string      `json:"phase_predecessor,omitempty" yaml:"phase_predecessor,omitempty"`
	PhaseSuccessor   string      `json:"phase_successor,omitempty" yaml:"phase_successor,omitempty"`
	PhaseCondition   string      `json:"phase_condition,omitempty" yaml:"phase_condition,omitempty"`
	PhaseProduces    string      `json:"phase_produces,omitempty" yaml:"phase_produces,omitempty"`
	BackRoutes       []BackRoute `json:"back_routes,omitempty" yaml:"back_routes,omitempty"`
	ComplexityGates  []string    `json:"complexity_gates,omitempty" yaml:"complexity_gates,omitempty"`
	HandoffCriteria  []string    `json:"handoff_criteria,omitempty" yaml:"handoff_criteria,omitempty"`
}

// BackRoute describes a reverse workflow routing path targeting this agent.
type BackRoute struct {
	SourcePhase              string `json:"source_phase" yaml:"source_phase"`
	Trigger                  string `json:"trigger" yaml:"trigger"`
	Condition                string `json:"condition,omitempty" yaml:"condition,omitempty"`
	RequiresUserConfirmation bool   `json:"requires_user_confirmation" yaml:"requires_user_confirmation"`
}

// SurfaceData contains the agent's I/O surface (L7).
type SurfaceData struct {
	DromenaOwned        []string     `json:"dromena_owned" yaml:"dromena_owned"`
	LegomenaAvailable   []string     `json:"legomena_available" yaml:"legomena_available"`
	ArtifactTypes       []string     `json:"artifact_types" yaml:"artifact_types"`
	ContractMustProduce []string     `json:"contract_must_produce,omitempty" yaml:"contract_must_produce,omitempty"`
	Commands            []CommandRef `json:"commands,omitempty" yaml:"commands,omitempty"`
}

// CommandRef describes a rite command associated with this agent.
type CommandRef struct {
	Name        string `json:"name" yaml:"name"`
	File        string `json:"file,omitempty" yaml:"file,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// HorizonData contains the agent's negative space — what it cannot do (L8).
type HorizonData struct {
	ToolsNotAvailable []string `json:"tools_not_available" yaml:"tools_not_available"`
	DisallowedOverlap []string `json:"disallowed_overlap" yaml:"disallowed_overlap"`
	SkillsUnreachable []string `json:"skills_unreachable" yaml:"skills_unreachable"`
	PhasesNotIn       []string `json:"phases_not_in" yaml:"phases_not_in"`
	MemoryBlindSpots  []string `json:"memory_blind_spots" yaml:"memory_blind_spots"`
	SurfaceGaps       []string `json:"surface_gaps" yaml:"surface_gaps"`
}

// SimulateOverlay holds the results of simulate mode analysis.
type SimulateOverlay struct {
	Prompt         string          `json:"prompt" yaml:"prompt"`
	ToolMatches    []SimulateMatch `json:"tool_matches" yaml:"tool_matches"`
	SkillMatches   []SimulateMatch `json:"skill_matches" yaml:"skill_matches"`
	ConstraintHits []SimulateMatch `json:"constraint_hits" yaml:"constraint_hits"`
	CanAttempt     []string        `json:"can_attempt" yaml:"can_attempt"`
	CannotAttempt  []string        `json:"cannot_attempt" yaml:"cannot_attempt"`
	HandoffNeeded  []string        `json:"handoff_needed" yaml:"handoff_needed"`
}

// SimulateMatch describes a keyword match in simulation analysis.
type SimulateMatch struct {
	Name      string `json:"name" yaml:"name"`
	MatchType string `json:"match_type" yaml:"match_type"` // "exact", "keyword", "partial"
	Relevance string `json:"relevance" yaml:"relevance"`   // "high", "medium", "low"
}

// --- Audit types ---

// AuditOverlay contains the results of audit mode analysis.
type AuditOverlay struct {
	Findings        []AuditFinding  `json:"findings" yaml:"findings"`
	SeveritySummary SeveritySummary `json:"severity_summary" yaml:"severity_summary"`
}

// AuditFinding is a single audit check result.
type AuditFinding struct {
	ID             string        `json:"id" yaml:"id"`
	Severity       AuditSeverity `json:"severity" yaml:"severity"`
	Category       AuditCategory `json:"category" yaml:"category"`
	LayersAffected []string      `json:"layers_affected" yaml:"layers_affected"`
	Title          string        `json:"title" yaml:"title"`
	Description    string        `json:"description" yaml:"description"`
	Evidence       string        `json:"evidence,omitempty" yaml:"evidence,omitempty"`
	Recommendation string        `json:"recommendation,omitempty" yaml:"recommendation,omitempty"`
}

// AuditSeverity classifies the impact of an audit finding.
type AuditSeverity string

const (
	SeverityCritical AuditSeverity = "CRITICAL"
	SeverityWarning  AuditSeverity = "WARNING"
	SeverityInfo     AuditSeverity = "INFO"
)

// AuditCategory classifies the nature of an audit finding.
type AuditCategory string

const (
	CategoryGap           AuditCategory = "GAP"
	CategoryInconsistency AuditCategory = "INCONSISTENCY"
	CategoryDegradation   AuditCategory = "DEGRADATION"
)

// SeveritySummary counts findings by severity level.
type SeveritySummary struct {
	Critical int `json:"critical" yaml:"critical"`
	Warning  int `json:"warning" yaml:"warning"`
	Info     int `json:"info" yaml:"info"`
}
