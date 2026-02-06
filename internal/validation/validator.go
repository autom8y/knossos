// Package validation provides JSON schema validation for Ariadne.
// Schemas are embedded in the binary for consistent validation.
package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed schemas/*.json
var schemaFS embed.FS

// Validator provides schema validation capabilities.
type Validator struct {
	compiler *jsonschema.Compiler
	schemas  map[string]*jsonschema.Schema
}

// embedLoader implements jsonschema.URLLoader for embedded files.
type embedLoader struct{}

// Load loads a JSON schema from the embedded filesystem.
// It decodes the JSON and returns the decoded value.
func (l *embedLoader) Load(url string) (any, error) {
	// URL format: embed:///schemas/name.json
	path := strings.TrimPrefix(url, "embed:///")
	data, err := schemaFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %s", path)
	}

	// Decode JSON
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("invalid JSON in schema %s: %w", path, err)
	}

	return v, nil
}

// NewValidator creates a new schema validator.
func NewValidator() (*Validator, error) {
	compiler := jsonschema.NewCompiler()

	// Register the embed loader for embed:// URLs
	compiler.UseLoader(&embedLoader{})

	// Register all embedded schemas
	entries, err := schemaFS.ReadDir("schemas")
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read embedded schemas", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := schemaFS.ReadFile("schemas/" + entry.Name())
		if err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to read schema: "+entry.Name(), err)
		}

		// Decode JSON for AddResource
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "invalid JSON in schema: "+entry.Name(), err)
		}

		url := "embed:///schemas/" + entry.Name()
		if err := compiler.AddResource(url, v); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to add schema: "+entry.Name(), err)
		}
	}

	return &Validator{
		compiler: compiler,
		schemas:  make(map[string]*jsonschema.Schema),
	}, nil
}

// getSchema returns a compiled schema, caching the result.
func (v *Validator) getSchema(name string) (*jsonschema.Schema, error) {
	if s, ok := v.schemas[name]; ok {
		return s, nil
	}

	url := fmt.Sprintf("embed:///schemas/%s.schema.json", name)
	s, err := v.compiler.Compile(url)
	if err != nil {
		return nil, errors.NewWithDetails(errors.CodeGeneralError, "failed to compile schema: "+name,
			map[string]interface{}{"url": url, "error": err.Error()})
	}

	v.schemas[name] = s
	return s, nil
}

// ValidateSessionContext validates session context data against the schema.
func (v *Validator) ValidateSessionContext(data []byte) error {
	schema, err := v.getSchema("session-context")
	if err != nil {
		return err
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return errors.Wrap(errors.CodeSchemaInvalid, "invalid JSON", err)
	}

	if err := schema.Validate(parsed); err != nil {
		return errors.NewWithDetails(errors.CodeSchemaInvalid,
			"session context validation failed",
			map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// ValidateSessionContextMap validates a session context map.
func (v *Validator) ValidateSessionContextMap(data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(errors.CodeSchemaInvalid, "failed to marshal data", err)
	}
	return v.ValidateSessionContext(jsonData)
}

// ValidateAgentData validates pre-parsed agent data against the agent schema.
// The data parameter should be an interface{} from JSON unmarshal (not raw bytes).
func (v *Validator) ValidateAgentData(data interface{}) error {
	schema, err := v.getSchema("agent")
	if err != nil {
		return err
	}

	if err := schema.Validate(data); err != nil {
		return errors.NewWithDetails(errors.CodeSchemaInvalid,
			"agent schema validation failed",
			map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// --- Lightweight validation without full schema ---

// ValidateSessionFields performs lightweight validation of required session fields.
// This is used when the full schema validator is not available.
func ValidateSessionFields(data map[string]interface{}) []string {
	var issues []string

	// Required fields per TDD
	required := []string{"session_id", "status", "created_at", "initiative", "complexity", "active_rite", "current_phase"}
	for _, field := range required {
		if _, ok := data[field]; !ok {
			issues = append(issues, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate status enum
	if status, ok := data["status"].(string); ok {
		switch status {
		case "ACTIVE", "PARKED", "ARCHIVED":
			// Valid
		default:
			issues = append(issues, fmt.Sprintf("invalid status: %s (must be ACTIVE, PARKED, or ARCHIVED)", status))
		}
	}

	// Validate session_id format
	if sessionID, ok := data["session_id"].(string); ok {
		if !IsValidSessionID(sessionID) {
			issues = append(issues, fmt.Sprintf("invalid session_id format: %s", sessionID))
		}
	}

	// Validate schema_version if present
	if version, ok := data["schema_version"].(string); ok {
		switch version {
		case "1.0", "2.0", "2.1", "2.2":
			// Valid
		default:
			issues = append(issues, fmt.Sprintf("unsupported schema_version: %s", version))
		}
	}

	return issues
}

// SessionID regex pattern
var sessionIDPattern = regexp.MustCompile(`^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$`)

// IsValidSessionID checks if a session ID matches the expected pattern.
func IsValidSessionID(id string) bool {
	return sessionIDPattern.MatchString(id)
}

// ValidateComplexity checks if a complexity value is valid.
func ValidateComplexity(complexity string) bool {
	switch complexity {
	case "PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION":
		return true
	default:
		return false
	}
}

// ValidatePhase checks if a phase value is valid.
func ValidatePhase(phase string) bool {
	switch phase {
	case "requirements", "design", "implementation", "validation", "complete":
		return true
	default:
		return false
	}
}

// ParseYAMLFrontmatter extracts YAML frontmatter from a markdown file.
// Returns the parsed YAML as a map.
func ParseYAMLFrontmatter(content []byte) (map[string]interface{}, error) {
	str := string(content)

	// Find frontmatter delimiters
	if !strings.HasPrefix(str, "---\n") && !strings.HasPrefix(str, "---\r\n") {
		return nil, errors.New(errors.CodeSchemaInvalid, "no YAML frontmatter found")
	}

	// Find closing delimiter
	endIdx := strings.Index(str[4:], "\n---")
	if endIdx == -1 {
		endIdx = strings.Index(str[4:], "\r\n---")
	}
	if endIdx == -1 {
		return nil, errors.New(errors.CodeSchemaInvalid, "unclosed YAML frontmatter")
	}

	// Extract YAML content
	yamlContent := str[4 : endIdx+4]

	// Parse YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &result); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid YAML frontmatter", err)
	}

	return result, nil
}

// BuildYAMLFrontmatter creates YAML frontmatter from a map.
func BuildYAMLFrontmatter(data map[string]interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", errors.Wrap(errors.CodeGeneralError, "failed to marshal YAML", err)
	}
	return fmt.Sprintf("---\n%s---\n", string(yamlBytes)), nil
}
