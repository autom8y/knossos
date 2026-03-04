// Package validation provides artifact validation for Ariadne.
package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ArtifactType represents a type of workflow artifact.
type ArtifactType string

const (
	// ArtifactTypePRD is a Product Requirements Document.
	ArtifactTypePRD ArtifactType = "prd"

	// ArtifactTypeTDD is a Technical Design Document.
	ArtifactTypeTDD ArtifactType = "tdd"

	// ArtifactTypeADR is an Architecture Decision Record.
	ArtifactTypeADR ArtifactType = "adr"

	// ArtifactTypeTestPlan is a Test Plan document.
	ArtifactTypeTestPlan ArtifactType = "test-plan"

	// ArtifactTypeUnknown indicates an unknown artifact type.
	ArtifactTypeUnknown ArtifactType = ""
)

// String returns the string representation of the artifact type.
func (t ArtifactType) String() string {
	return string(t)
}

// SchemaName returns the schema filename for this artifact type.
func (t ArtifactType) SchemaName() string {
	switch t {
	case ArtifactTypePRD:
		return "prd"
	case ArtifactTypeTDD:
		return "tdd"
	case ArtifactTypeADR:
		return "adr"
	case ArtifactTypeTestPlan:
		return "test-plan"
	default:
		return ""
	}
}

// ParseArtifactType parses a string into an ArtifactType.
func ParseArtifactType(s string) ArtifactType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "prd":
		return ArtifactTypePRD
	case "tdd":
		return ArtifactTypeTDD
	case "adr":
		return ArtifactTypeADR
	case "test-plan", "testplan", "test_plan", "tp", "test":
		return ArtifactTypeTestPlan
	default:
		return ArtifactTypeUnknown
	}
}

// Filename patterns for artifact type detection
var (
	prdPattern      = regexp.MustCompile(`^PRD-[a-zA-Z0-9-]+\.md$`)
	tddPattern      = regexp.MustCompile(`^TDD-[a-zA-Z0-9-]+\.md$`)
	adrPattern      = regexp.MustCompile(`^ADR-[0-9]+[a-zA-Z0-9-]*\.md$`)
	testPlanPattern = regexp.MustCompile(`^(TEST|TP)-[a-zA-Z0-9-]+\.md$`)
)

// DetectArtifactType determines the artifact type from filename and/or frontmatter.
// Priority:
//  1. Frontmatter "type" field (if present)
//  2. Filename pattern (PRD-*.md, TDD-*.md, etc.)
//  3. Returns ArtifactTypeUnknown if neither works
func DetectArtifactType(filename string, frontmatter map[string]any) ArtifactType {
	// Priority 1: Check frontmatter "type" field
	if frontmatter != nil {
		if typeField, ok := frontmatter["type"].(string); ok {
			if t := ParseArtifactType(typeField); t != ArtifactTypeUnknown {
				return t
			}
		}
	}

	// Priority 2: Check filename pattern
	basename := filepath.Base(filename)

	if prdPattern.MatchString(basename) {
		return ArtifactTypePRD
	}
	if tddPattern.MatchString(basename) {
		return ArtifactTypeTDD
	}
	if adrPattern.MatchString(basename) {
		return ArtifactTypeADR
	}
	if testPlanPattern.MatchString(basename) {
		return ArtifactTypeTestPlan
	}

	return ArtifactTypeUnknown
}

// ValidationIssue represents a single validation problem.
type ValidationIssue struct {
	// Field is the JSON path to the problematic field (e.g., "artifact_id").
	Field string `json:"field,omitempty"`

	// Message describes the validation problem.
	Message string `json:"message"`

	// Value is the actual value that failed validation (if applicable).
	Value any `json:"value,omitempty"`
}

// ArtifactValidationResult contains the result of artifact validation.
type ArtifactValidationResult struct {
	// Valid is true if the artifact passed validation.
	Valid bool `json:"valid"`

	// ArtifactType is the detected or specified artifact type.
	ArtifactType ArtifactType `json:"artifact_type"`

	// FilePath is the path to the validated file.
	FilePath string `json:"file_path"`

	// Issues contains validation problems (empty if Valid is true).
	Issues []ValidationIssue `json:"issues,omitempty"`

	// Frontmatter contains the parsed frontmatter data.
	Frontmatter map[string]any `json:"frontmatter,omitempty"`
}

// ArtifactValidator validates workflow artifacts against their schemas.
type ArtifactValidator struct {
	validator *Validator
}

// NewArtifactValidator creates a new artifact validator.
func NewArtifactValidator() (*ArtifactValidator, error) {
	v, err := NewValidator()
	if err != nil {
		return nil, err
	}
	return &ArtifactValidator{validator: v}, nil
}

// ValidateFile validates an artifact file against its schema.
// If artifactType is empty, it will be auto-detected.
func (av *ArtifactValidator) ValidateFile(filePath string, artifactType ArtifactType) (*ArtifactValidationResult, error) {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeFileNotFound,
				"artifact file not found",
				map[string]any{"path": filePath})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read artifact file", err)
	}

	return av.Validate(content, filePath, artifactType)
}

// Validate validates artifact content against its schema.
// If artifactType is empty, it will be auto-detected.
func (av *ArtifactValidator) Validate(content []byte, filePath string, artifactType ArtifactType) (*ArtifactValidationResult, error) {
	result := &ArtifactValidationResult{
		FilePath: filePath,
	}

	// Extract frontmatter
	fm, err := ExtractFrontmatter(content)
	if err != nil {
		// Return validation result with error as an issue
		if e, ok := err.(*errors.Error); ok {
			result.Issues = []ValidationIssue{
				{Message: e.Message},
			}
		} else {
			result.Issues = []ValidationIssue{
				{Message: err.Error()},
			}
		}
		return result, nil
	}

	result.Frontmatter = fm.Data

	// Detect artifact type if not specified
	if artifactType == ArtifactTypeUnknown {
		artifactType = DetectArtifactType(filePath, fm.Data)
	}
	result.ArtifactType = artifactType

	// If still unknown, return error
	if artifactType == ArtifactTypeUnknown {
		result.Issues = []ValidationIssue{
			{Message: "cannot determine artifact type from filename or frontmatter"},
		}
		return result, nil
	}

	// Validate against schema
	issues, err := av.validateAgainstSchema(fm.Data, artifactType)
	if err != nil {
		return nil, err
	}

	result.Issues = issues
	result.Valid = len(issues) == 0

	return result, nil
}

// validateAgainstSchema validates frontmatter data against an artifact schema.
func (av *ArtifactValidator) validateAgainstSchema(data map[string]any, artifactType ArtifactType) ([]ValidationIssue, error) {
	schemaName := artifactType.SchemaName()
	if schemaName == "" {
		return nil, errors.NewWithDetails(errors.CodeSchemaNotFound,
			"no schema for artifact type",
			map[string]any{"type": string(artifactType)})
	}

	// Get compiled schema
	schema, err := av.validator.getSchema(schemaName)
	if err != nil {
		// Wrap with more context for debugging
		return nil, errors.NewWithDetails(errors.CodeGeneralError,
			"failed to compile schema: "+schemaName,
			map[string]any{"cause": err.Error()})
	}

	// Convert data to JSON for validation
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to marshal frontmatter", err)
	}

	var parsed any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to parse frontmatter JSON", err)
	}

	// Validate
	if err := schema.Validate(parsed); err != nil {
		return extractValidationIssues(err), nil
	}

	return nil, nil
}

// extractValidationIssues converts jsonschema errors into ValidationIssues.
func extractValidationIssues(err error) []ValidationIssue {
	var issues []ValidationIssue

	// Handle different error types from jsonschema library
	switch e := err.(type) {
	case *jsonschema.ValidationError:
		issues = append(issues, extractFromValidationError(e)...)
	default:
		// Generic error
		issues = append(issues, ValidationIssue{
			Message: err.Error(),
		})
	}

	return issues
}

// extractFromValidationError recursively extracts issues from a ValidationError.
func extractFromValidationError(e *jsonschema.ValidationError) []ValidationIssue {
	var issues []ValidationIssue

	// If there are nested causes, extract from each
	if len(e.Causes) > 0 {
		for _, cause := range e.Causes {
			issues = append(issues, extractFromValidationError(cause)...)
		}
	} else {
		// Leaf error - extract the issue
		issue := ValidationIssue{
			Message: e.Error(),
		}

		// Extract field path from instance location (it's a slice of path components)
		if len(e.InstanceLocation) > 0 {
			// Join path components with /
			issue.Field = strings.Join(e.InstanceLocation, "/")
		}

		issues = append(issues, issue)
	}

	return issues
}

// ValidArtifactTypes returns all valid artifact type strings.
func ValidArtifactTypes() []string {
	return []string{
		string(ArtifactTypePRD),
		string(ArtifactTypeTDD),
		string(ArtifactTypeADR),
		string(ArtifactTypeTestPlan),
	}
}
