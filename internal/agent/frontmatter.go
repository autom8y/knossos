package agent

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// AgentFrontmatter represents the parsed frontmatter of an agent markdown file.
// All 58+ existing agents have at minimum: name, description, tools.
// Enhanced fields (type, upstream, downstream, produces, contract) are optional
// and will be populated during the Phase 3 migration.
type AgentFrontmatter struct {
	// Identity (required)
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Role        string `yaml:"role,omitempty" json:"role,omitempty"`

	// Archetype
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Capabilities
	Tools   FlexibleStringSlice `yaml:"tools" json:"tools,omitempty"`
	Model   string              `yaml:"model,omitempty" json:"model,omitempty"`
	Color   string              `yaml:"color,omitempty" json:"color,omitempty"`
	Aliases []string            `yaml:"aliases,omitempty" json:"aliases,omitempty"`

	// CC-native fields (camelCase matches Claude Code's expected frontmatter schema)
	MaxTurns        int                 `yaml:"maxTurns,omitempty" json:"maxTurns,omitempty"`
	Skills          []string            `yaml:"skills,omitempty" json:"skills,omitempty"`
	DisallowedTools FlexibleStringSlice `yaml:"disallowedTools,omitempty" json:"disallowedTools,omitempty"`
	Memory          MemoryField         `yaml:"memory,omitempty" json:"memory,omitempty"`
	PermissionMode  string              `yaml:"permissionMode,omitempty" json:"permissionMode,omitempty"`
	McpServers      []McpServerConfig   `yaml:"mcpServers,omitempty" json:"mcpServers,omitempty"`
	Hooks           map[string]any      `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	WriteGuard      any                 `yaml:"write-guard,omitempty" json:"write-guard,omitempty"` // Source-only: consumed during materialization

	// Workflow Position
	Upstream   []UpstreamRef   `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	Downstream []DownstreamRef `yaml:"downstream,omitempty" json:"downstream,omitempty"`
	Produces   []ArtifactDecl  `yaml:"produces,omitempty" json:"produces,omitempty"`

	// Behavioral Contract
	Contract *BehavioralContract `yaml:"contract,omitempty" json:"contract,omitempty"`

	// Schema Version
	SchemaVersion string `yaml:"schema_version,omitempty" json:"schema_version,omitempty"`
}

// Known Claude Code tools. This list is used for tool reference validation.
var knownTools = map[string]bool{
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

// Valid agent types (archetypes).
var validAgentTypes = map[string]bool{
	"orchestrator": true,
	"specialist":   true,
	"reviewer":     true,
	"meta":         true,
	"designer":     true,
	"analyst":      true,
	"engineer":     true,
}

// mcpToolPattern matches MCP tool references like "mcp:github" or "mcp:github/create_issue".
var mcpToolPattern = regexp.MustCompile(`^mcp:[a-z0-9-]+(\/[a-z0-9_-]+)?$`)

// ParseAgentFrontmatter extracts and parses frontmatter from an agent markdown file.
// Returns error if frontmatter is missing or has invalid YAML.
func ParseAgentFrontmatter(content []byte) (*AgentFrontmatter, error) {
	yamlBytes, _, err := frontmatter.Parse(content)
	if err != nil {
		return nil, errors.New(errors.CodeParseError, err.Error())
	}

	var fm AgentFrontmatter
	if err := yaml.Unmarshal(yamlBytes, &fm); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "invalid frontmatter YAML", err)
	}

	return &fm, nil
}

// Validate checks that the frontmatter has required fields and valid values.
// This performs Go-level semantic validation beyond what JSON Schema covers.
// It is the canonical source of truth for struct-level validation rules.
func (f *AgentFrontmatter) Validate() error {
	// Core rules shared with the validator pipeline
	if err := f.validateCore(); err != nil {
		return err
	}

	// disallowedTools validation: errors in standalone Validate(), but
	// the AgentValidator pipeline treats these as warnings instead.
	for _, tool := range f.DisallowedTools {
		if err := validateToolReference(tool); err != nil {
			return errors.Wrap(errors.CodeValidationFailed,
				fmt.Sprintf("agent frontmatter: invalid disallowedTools entry %q", tool), err)
		}
	}

	return nil
}

// validateCore performs the core struct-level validation rules shared between
// standalone Validate() and the AgentValidator pipeline. It validates name,
// description, type, model, tools, and maxTurns. disallowedTools is excluded
// because the pipeline treats unknown entries as warnings rather than errors.
func (f *AgentFrontmatter) validateCore() error {
	if f.Name == "" {
		return errors.New(errors.CodeValidationFailed, "agent frontmatter: name is required")
	}
	if f.Description == "" {
		return errors.New(errors.CodeValidationFailed, "agent frontmatter: description is required")
	}

	// Validate type if present
	if f.Type != "" {
		if !validAgentTypes[f.Type] {
			return errors.New(errors.CodeValidationFailed,
				fmt.Sprintf("agent frontmatter: invalid type %q, must be one of: %s",
					f.Type, validAgentTypesList()))
		}
	}

	// Validate model if present
	if f.Model != "" {
		switch f.Model {
		case "opus", "sonnet", "haiku", "inherit":
			// valid
		default:
			return errors.New(errors.CodeValidationFailed,
				fmt.Sprintf("agent frontmatter: invalid model %q, must be opus, sonnet, haiku, or inherit", f.Model))
		}
	}

	// Validate permissionMode if present
	if f.PermissionMode != "" {
		switch f.PermissionMode {
		case "default", "plan", "bypassPermissions":
			// valid
		default:
			return errors.New(errors.CodeValidationFailed,
				fmt.Sprintf("agent frontmatter: invalid permissionMode %q, must be default, plan, or bypassPermissions", f.PermissionMode))
		}
	}

	// Validate memory scope if present
	if f.Memory != "" {
		if !validMemoryScopes[string(f.Memory)] {
			return errors.New(errors.CodeValidationFailed,
				fmt.Sprintf("agent frontmatter: invalid memory scope %q, must be user, project, or local", f.Memory))
		}
	}

	// Validate tool references
	for _, tool := range f.Tools {
		if err := validateToolReference(tool); err != nil {
			return err
		}
	}

	// Validate maxTurns if present
	if f.MaxTurns < 0 {
		return errors.New(errors.CodeValidationFailed,
			fmt.Sprintf("agent frontmatter: maxTurns must be >= 0, got %d", f.MaxTurns))
	}

	return nil
}

// MCPServers extracts unique MCP server names from the tools list.
// For a tool reference "mcp:github/create_issue", returns "github".
// For "mcp:github", returns "github".
func (f *AgentFrontmatter) MCPServers() []string {
	seen := make(map[string]bool)
	var servers []string

	for _, tool := range f.Tools {
		if !strings.HasPrefix(tool, "mcp:") {
			continue
		}
		// Extract server name: "mcp:github/create_issue" -> "github"
		ref := strings.TrimPrefix(tool, "mcp:")
		parts := strings.SplitN(ref, "/", 2)
		server := parts[0]
		if !seen[server] {
			seen[server] = true
			servers = append(servers, server)
		}
	}

	return servers
}

// validateToolReference checks that a tool name is a known Claude Code tool
// or a valid MCP tool reference.
func validateToolReference(tool string) error {
	// Check known tools
	if knownTools[tool] {
		return nil
	}

	// Check MCP pattern
	if mcpToolPattern.MatchString(tool) {
		return nil
	}

	return errors.New(errors.CodeValidationFailed,
		fmt.Sprintf("agent frontmatter: unknown tool %q (expected known tool or mcp:<server>[/<method>])", tool))
}

// validAgentTypesList returns a comma-separated list of valid agent types.
func validAgentTypesList() string {
	types := make([]string, 0, len(validAgentTypes))
	for t := range validAgentTypes {
		types = append(types, t)
	}
	return strings.Join(types, ", ")
}
