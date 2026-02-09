// Package manifest - schema validation and detection
package manifest

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// Error codes for manifest domain (re-exported from errors package for convenience)
const (
	CodeParseError     = errors.CodeParseError
	CodeSchemaNotFound = errors.CodeSchemaNotFound
)

//go:embed schemas/*.json
var schemaFS embed.FS

// Schema names for different manifest types
const (
	SchemaManifest     = "manifest"
	SchemaRiteManifest = "rite-manifest"
)

// SchemaValidator provides manifest schema validation.
type SchemaValidator struct {
	compiler *jsonschema.Compiler
	schemas  map[string]*jsonschema.Schema
}

// NewSchemaValidator creates a new schema validator for manifests.
// Note: This uses lightweight structural validation rather than full JSON Schema
// due to issues with the jsonschema library and meta-schema resolution.
func NewSchemaValidator() (*SchemaValidator, error) {
	// Load embedded schemas for reference (used for type detection)
	// Full JSON Schema validation is not currently active
	return &SchemaValidator{
		compiler: nil,
		schemas:  make(map[string]*jsonschema.Schema),
	}, nil
}

// getSchema returns a compiled schema, caching the result.
func (v *SchemaValidator) getSchema(name string) (*jsonschema.Schema, error) {
	if s, ok := v.schemas[name]; ok {
		return s, nil
	}
	return nil, errors.ErrSchemaNotFound(name)
}

// HasSchema checks if a schema exists.
func (v *SchemaValidator) HasSchema(name string) bool {
	_, err := v.getSchema(name)
	return err == nil
}

// ValidationIssue represents a single validation issue.
type ValidationIssue struct {
	Path     string `json:"path"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error" or "warning"
}

// ValidationResult holds the result of schema validation.
type ValidationResult struct {
	Path     string            `json:"path"`
	Schema   string            `json:"schema"`
	Valid    bool              `json:"valid"`
	Issues   []ValidationIssue `json:"issues"`
	Warnings []ValidationIssue `json:"warnings"`
}

// Validate validates a manifest against its schema.
// Uses lightweight structural validation to check required fields.
func (v *SchemaValidator) Validate(m *Manifest, schemaName string, strict bool) (*ValidationResult, error) {
	result := &ValidationResult{
		Path:     m.Path,
		Schema:   schemaName + ".schema.json",
		Valid:    true,
		Issues:   []ValidationIssue{},
		Warnings: []ValidationIssue{},
	}

	// Perform lightweight structural validation based on schema type
	switch schemaName {
	case SchemaManifest:
		validateManifestStructure(m.Content, result)
	case SchemaRiteManifest:
		validateRiteManifestStructure(m.Content, result)
	default:
		// For unknown schemas, just check it's valid JSON (which it is if we loaded it)
	}

	// In strict mode, check for additional properties
	if strict {
		warnings := checkAdditionalProperties(m.Content, schemaName)
		result.Warnings = append(result.Warnings, warnings...)
	}

	result.Valid = len(result.Issues) == 0
	return result, nil
}

// validateManifestStructure checks required fields for project manifest.
func validateManifestStructure(content map[string]interface{}, result *ValidationResult) {
	// Required: version
	if _, ok := content["version"]; !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.version",
			Message:  "missing required field",
			Severity: "error",
		})
	} else if v, ok := content["version"].(string); ok {
		// Check version format
		if !isValidVersion(v) {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$.version",
				Message:  "invalid version format, expected X.Y",
				Severity: "error",
			})
		}
	}
}

// validateRiteManifestStructure checks required fields for rite manifest.
func validateRiteManifestStructure(content map[string]interface{}, result *ValidationResult) {
	required := []string{"version", "name", "workflow", "agents"}
	for _, field := range required {
		if _, ok := content[field]; !ok {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + field,
				Message:  "missing required field",
				Severity: "error",
			})
		}
	}

	// Check workflow structure
	if workflow, ok := content["workflow"].(map[string]interface{}); ok {
		if _, ok := workflow["type"]; !ok {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$.workflow.type",
				Message:  "missing required field",
				Severity: "error",
			})
		}
		if _, ok := workflow["entry_point"]; !ok {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$.workflow.entry_point",
				Message:  "missing required field",
				Severity: "error",
			})
		}
	}
}


// isValidVersion checks if a version string matches X.Y format.
func isValidVersion(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

// ValidateBytes validates raw bytes against a schema.
func (v *SchemaValidator) ValidateBytes(data []byte, format Format, schemaName string) (*ValidationResult, error) {
	result := &ValidationResult{
		Path:     "",
		Schema:   schemaName + ".schema.json",
		Valid:    true,
		Issues:   []ValidationIssue{},
		Warnings: []ValidationIssue{},
	}

	schema, err := v.getSchema(schemaName)
	if err != nil {
		return nil, err
	}

	// Convert to JSON if YAML
	var jsonData []byte
	if format == FormatYAML {
		var content interface{}
		if err := yaml.Unmarshal(data, &content); err != nil {
			result.Valid = false
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$",
				Message:  "Invalid YAML: " + err.Error(),
				Severity: "error",
			})
			return result, nil
		}
		jsonData, _ = json.Marshal(content)
	} else {
		jsonData = data
	}

	var parsed interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$",
			Message:  "Invalid JSON: " + err.Error(),
			Severity: "error",
		})
		return result, nil
	}

	// Validate against schema
	if err := schema.Validate(parsed); err != nil {
		result.Valid = false
		result.Issues = parseValidationErrors(err)
	}

	return result, nil
}

// parseValidationErrors converts jsonschema errors to ValidationIssues.
func parseValidationErrors(err error) []ValidationIssue {
	var issues []ValidationIssue

	if ve, ok := err.(*jsonschema.ValidationError); ok {
		// Use detailed output for better error messages
		output := ve.DetailedOutput()
		collectOutputErrors(output, &issues)

		// If no issues collected, use the main error
		if len(issues) == 0 {
			issues = append(issues, ValidationIssue{
				Path:     instanceLocationSliceToPath(ve.InstanceLocation),
				Message:  ve.Error(),
				Severity: "error",
			})
		}
	} else {
		issues = append(issues, ValidationIssue{
			Path:     "$",
			Message:  err.Error(),
			Severity: "error",
		})
	}

	return issues
}

// instanceLocationSliceToPath converts instance location slice to jQuery-style path.
func instanceLocationSliceToPath(location []string) string {
	if len(location) == 0 {
		return "$"
	}
	return "$." + strings.Join(location, ".")
}

// collectOutputErrors recursively collects errors from OutputUnit.
func collectOutputErrors(output *jsonschema.OutputUnit, issues *[]ValidationIssue) {
	if output == nil {
		return
	}

	// Collect error from this unit
	if output.Error != nil {
		*issues = append(*issues, ValidationIssue{
			Path:     jsonPointerToPath(output.InstanceLocation),
			Message:  output.Error.Kind.LocalizedString(nil),
			Severity: "error",
		})
	}

	// Recurse into nested errors
	for i := range output.Errors {
		collectOutputErrors(&output.Errors[i], issues)
	}
}

// jsonPointerToPath converts JSON pointer string to jQuery-style path.
func jsonPointerToPath(location string) string {
	if location == "" {
		return "$"
	}
	// Location is already in format like "/rites/default"
	// Convert to $.rites.default
	path := strings.ReplaceAll(location, "/", ".")
	if strings.HasPrefix(path, ".") {
		return "$" + path
	}
	return "$." + path
}

// jsonPathToJQuery converts JSON pointer to jQuery-style path ($.foo.bar).
func jsonPathToJQuery(location string) string {
	if location == "" {
		return "$"
	}
	// Replace "/" with "." and prepend "$"
	path := strings.ReplaceAll(location, "/", ".")
	if strings.HasPrefix(path, ".") {
		return "$" + path
	}
	return "$." + path
}

// checkAdditionalProperties checks for properties not in the schema.
func checkAdditionalProperties(content map[string]interface{}, schemaName string) []ValidationIssue {
	// This is a simplified check - a full implementation would
	// load the schema and walk the content tree
	var warnings []ValidationIssue

	// Known fields for manifest schema
	knownFields := map[string]map[string]bool{
		SchemaManifest: {
			"version": true, "project": true, "rites": true,
			"paths": true, "schemas": true, "settings": true,
		},
		SchemaRiteManifest: {
			"version": true, "name": true, "description": true,
			"workflow": true, "agents": true, "skills": true, "hooks": true,
		},
	}

	known, ok := knownFields[schemaName]
	if !ok {
		return warnings
	}

	for key := range content {
		if !known[key] {
			warnings = append(warnings, ValidationIssue{
				Path:     "$." + key,
				Message:  "additional property not in schema",
				Severity: "warning",
			})
		}
	}

	return warnings
}

// DetectSchemaFromPath returns the schema name for a given path.
func DetectSchemaFromPath(path string) (string, error) {
	// Extract filename and path parts
	filename := filepath.Base(path)
	dir := filepath.Dir(path)

	// Pattern matching
	patterns := []struct {
		pattern *regexp.Regexp
		schema  string
	}{
		{regexp.MustCompile(`\.claude[/\\]manifest\.json$`), SchemaManifest},
		{regexp.MustCompile(`rites[/\\][^/\\]+[/\\]manifest\.ya?ml$`), SchemaRiteManifest},
	}

	for _, p := range patterns {
		if p.pattern.MatchString(path) {
			return p.schema, nil
		}
	}

	// Fallback: check filename directly
	switch filename {
	case "manifest.json":
		if strings.Contains(dir, ".claude") {
			return SchemaManifest, nil
		}
	case "manifest.yaml", "manifest.yml":
		if strings.Contains(dir, "rites") {
			return SchemaRiteManifest, nil
		}
	}

	return "", errors.NewWithDetails(errors.CodeSchemaNotFound,
		fmt.Sprintf("no schema detected for path: %s", path),
		map[string]interface{}{"path": path})
}

// GetSchemaInfo returns information about the manifest schema.
func GetSchemaInfo(m *Manifest) (*SchemaInfo, error) {
	schemaName, err := DetectSchemaFromPath(m.Path)
	if err != nil {
		// Try to detect from content
		if m.Content != nil {
			if _, hasAgents := m.Content["agents"]; hasAgents {
				if _, hasWorkflow := m.Content["workflow"]; hasWorkflow {
					schemaName = SchemaRiteManifest
				}
			} else {
				schemaName = SchemaManifest
			}
		}
	}

	version := ""
	if v, ok := m.Content["version"].(string); ok {
		version = v
	}

	return &SchemaInfo{
		Type:    schemaName,
		Version: version,
	}, nil
}

// SchemaInfo holds schema metadata for a manifest.
type SchemaInfo struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Valid   bool   `json:"valid,omitempty"`
}

// embedLoader implements jsonschema.URLLoader for embedded files.
type embedLoader struct{}

func (l *embedLoader) Load(url string) (io.ReadCloser, error) {
	path := strings.TrimPrefix(url, "embed:///")
	data, err := schemaFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %s", path)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}
