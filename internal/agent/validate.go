package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/validation"
)

// ValidationMode controls how strict validation is.
type ValidationMode int

const (
	// ValidationModeWarn performs lenient validation suitable for existing agents.
	// Missing optional enhanced fields produce warnings, not errors.
	ValidationModeWarn ValidationMode = iota

	// ValidationModeStrict performs full validation requiring all enhanced fields.
	// Suitable for post-migration agents with complete metadata.
	ValidationModeStrict
)

// AgentValidationResult contains the result of agent validation.
type AgentValidationResult struct {
	// Valid is true if the agent passed validation without errors.
	Valid bool `json:"valid"`

	// Issues contains validation errors (empty if Valid is true).
	Issues []ValidationIssue `json:"issues,omitempty"`

	// Warnings contains non-blocking validation messages.
	Warnings []string `json:"warnings,omitempty"`

	// Frontmatter contains the parsed frontmatter (nil if parsing failed).
	Frontmatter *AgentFrontmatter `json:"frontmatter,omitempty"`
}

// ValidationIssue represents a single validation problem.
type ValidationIssue struct {
	// Field is the path to the problematic field (e.g., "tools", "contract.must_not").
	Field string `json:"field,omitempty"`

	// Message describes the validation problem.
	Message string `json:"message"`

	// Value is the actual value that failed validation (if applicable).
	Value interface{} `json:"value,omitempty"`
}

// AgentValidator validates agent files against the schema and semantic rules.
type AgentValidator struct {
	schemaValidator *validation.Validator
}

// NewAgentValidator creates a new agent validator with embedded schema support.
func NewAgentValidator() (*AgentValidator, error) {
	v, err := validation.NewValidator()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create schema validator", err)
	}
	return &AgentValidator{schemaValidator: v}, nil
}

// ValidateAgentFile validates a single agent file at the given path.
func (av *AgentValidator) ValidateAgentFile(path string, mode ValidationMode) (*AgentValidationResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"agent file not found",
				map[string]interface{}{"path": path})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read agent file", err)
	}

	return av.ValidateAgentFrontmatter(content, mode)
}

// ValidateAgentFrontmatter validates agent frontmatter bytes.
func (av *AgentValidator) ValidateAgentFrontmatter(content []byte, mode ValidationMode) (*AgentValidationResult, error) {
	result := &AgentValidationResult{}

	// Phase 1: Parse frontmatter
	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Message: fmt.Sprintf("frontmatter parse error: %s", err.Error()),
		})
		return result, nil
	}
	result.Frontmatter = fm

	// Phase 2: JSON Schema validation
	schemaIssues, schemaWarnings := av.validateSchema(fm)
	result.Issues = append(result.Issues, schemaIssues...)
	result.Warnings = append(result.Warnings, schemaWarnings...)

	// Phase 3: Semantic validation
	semanticIssues, semanticWarnings := av.validateSemantics(fm, mode)
	result.Issues = append(result.Issues, semanticIssues...)
	result.Warnings = append(result.Warnings, semanticWarnings...)

	result.Valid = len(result.Issues) == 0
	return result, nil
}

// validateSchema validates frontmatter against the JSON Schema.
func (av *AgentValidator) validateSchema(fm *AgentFrontmatter) ([]ValidationIssue, []string) {
	var issues []ValidationIssue
	var warnings []string

	// Convert frontmatter to a map for schema validation.
	// We marshal to JSON then unmarshal to interface{} to match schema validator expectations.
	jsonBytes, err := json.Marshal(fm)
	if err != nil {
		issues = append(issues, ValidationIssue{
			Message: fmt.Sprintf("failed to marshal frontmatter for schema validation: %s", err.Error()),
		})
		return issues, warnings
	}

	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		issues = append(issues, ValidationIssue{
			Message: fmt.Sprintf("failed to prepare data for schema validation: %s", err.Error()),
		})
		return issues, warnings
	}

	if err := av.schemaValidator.ValidateAgentData(data); err != nil {
		// Extract structured issues from the schema validation error
		issues = append(issues, ValidationIssue{
			Message: fmt.Sprintf("schema validation: %s", err.Error()),
		})
	}

	return issues, warnings
}

// validateSemantics performs Go-level semantic validation beyond JSON Schema.
func (av *AgentValidator) validateSemantics(fm *AgentFrontmatter, mode ValidationMode) ([]ValidationIssue, []string) {
	var issues []ValidationIssue
	var warnings []string

	// Required fields check
	if fm.Name == "" {
		issues = append(issues, ValidationIssue{
			Field:   "name",
			Message: "name is required",
		})
	}

	if fm.Description == "" {
		issues = append(issues, ValidationIssue{
			Field:   "description",
			Message: "description is required",
		})
	}

	// Tools: required in STRICT, warning in WARN
	if len(fm.Tools) == 0 {
		if mode == ValidationModeStrict {
			issues = append(issues, ValidationIssue{
				Field:   "tools",
				Message: "tools is required",
			})
		} else {
			warnings = append(warnings, "tools field is empty or missing")
		}
	}

	// Validate each tool reference
	for _, tool := range fm.Tools {
		if err := validateToolReference(tool); err != nil {
			issues = append(issues, ValidationIssue{
				Field:   "tools",
				Message: err.Error(),
				Value:   tool,
			})
		}
	}

	// Validate type if present
	if fm.Type != "" && !validAgentTypes[fm.Type] {
		issues = append(issues, ValidationIssue{
			Field:   "type",
			Message: fmt.Sprintf("unknown agent type %q", fm.Type),
			Value:   fm.Type,
		})
	}

	// Validate model if present
	if fm.Model != "" {
		switch fm.Model {
		case "opus", "sonnet", "haiku":
			// valid
		default:
			issues = append(issues, ValidationIssue{
				Field:   "model",
				Message: fmt.Sprintf("unknown model %q, must be opus, sonnet, or haiku", fm.Model),
				Value:   fm.Model,
			})
		}
	}

	// Strict mode: require enhanced fields
	if mode == ValidationModeStrict {
		if fm.Type == "" {
			issues = append(issues, ValidationIssue{
				Field:   "type",
				Message: "type is required in strict mode",
			})
		}
	}

	// Archetype-specific validation
	issues = append(issues, av.validateArchetypeRules(fm, mode)...)
	warnings = append(warnings, av.archetypeWarnings(fm)...)

	return issues, warnings
}

// validateArchetypeRules checks archetype-specific constraints that are errors.
func (av *AgentValidator) validateArchetypeRules(fm *AgentFrontmatter, mode ValidationMode) []ValidationIssue {
	var issues []ValidationIssue

	if fm.Type == "" {
		return issues
	}

	switch fm.Type {
	case "reviewer":
		// Reviewers should have contract.must_not in strict mode
		if mode == ValidationModeStrict {
			if fm.Contract == nil || len(fm.Contract.MustNot) == 0 {
				issues = append(issues, ValidationIssue{
					Field:   "contract.must_not",
					Message: "reviewer agents require contract.must_not to define review boundaries",
				})
			}
		}
	}

	return issues
}

// archetypeWarnings produces non-blocking warnings for archetype patterns.
func (av *AgentValidator) archetypeWarnings(fm *AgentFrontmatter) []string {
	var warnings []string

	if fm.Type == "" {
		return warnings
	}

	switch fm.Type {
	case "orchestrator":
		// Orchestrators should ideally only use Read
		hasNonReadTool := false
		for _, tool := range fm.Tools {
			if tool != "Read" {
				hasNonReadTool = true
				break
			}
		}
		if hasNonReadTool {
			warnings = append(warnings, "orchestrator agents typically only use Read tool; additional tools may violate consultation pattern")
		}

		// Orchestrators should use opus
		if fm.Model != "" && fm.Model != "opus" {
			warnings = append(warnings, fmt.Sprintf("orchestrator agents should use opus model, found %q", fm.Model))
		}

	case "reviewer":
		// Reviewers should have must_not (warn even in WARN mode)
		if fm.Contract == nil || len(fm.Contract.MustNot) == 0 {
			warnings = append(warnings, "reviewer agents should define contract.must_not for review boundaries")
		}
	}

	return warnings
}

// ValidateToolReferences validates all tool references in an agent's frontmatter.
// Returns issues for unknown tools and warnings for MCP tools without manifest verification.
func ValidateToolReferences(fm *AgentFrontmatter) ([]ValidationIssue, []string) {
	var issues []ValidationIssue
	var warnings []string

	for _, tool := range fm.Tools {
		if err := validateToolReference(tool); err != nil {
			issues = append(issues, ValidationIssue{
				Field:   "tools",
				Message: err.Error(),
				Value:   tool,
			})
			continue
		}

		// Warn about MCP tools (they need manifest cross-reference in Phase 5)
		if strings.HasPrefix(tool, "mcp:") {
			warnings = append(warnings, fmt.Sprintf("MCP tool %q will be cross-referenced with rite manifest in future validation", tool))
		}
	}

	return issues, warnings
}
