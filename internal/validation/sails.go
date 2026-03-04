// Package validation provides WHITE_SAILS validation for session confidence signals.
package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
	"gopkg.in/yaml.v3"
)

// SailsColor represents a WHITE_SAILS confidence color.
type SailsColor string

const (
	// SailsColorWhite indicates high confidence - all proofs pass.
	SailsColorWhite SailsColor = "WHITE"

	// SailsColorGray indicates medium confidence - some uncertainty remains.
	SailsColorGray SailsColor = "GRAY"

	// SailsColorBlack indicates low confidence - significant issues.
	SailsColorBlack SailsColor = "BLACK"
)

// String returns the string representation of the color.
func (c SailsColor) String() string {
	return string(c)
}

// ParseSailsColor parses a string into a SailsColor.
func ParseSailsColor(s string) SailsColor {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "WHITE":
		return SailsColorWhite
	case "GRAY":
		return SailsColorGray
	case "BLACK":
		return SailsColorBlack
	default:
		return ""
	}
}

// ValidSailsColors returns all valid color strings.
func ValidSailsColors() []string {
	return []string{
		string(SailsColorWhite),
		string(SailsColorGray),
		string(SailsColorBlack),
	}
}

// ProofStatus represents the status of a verification proof.
type ProofStatus string

const (
	// ProofStatusPass indicates the proof passed.
	ProofStatusPass ProofStatus = "PASS"

	// ProofStatusFail indicates the proof failed.
	ProofStatusFail ProofStatus = "FAIL"

	// ProofStatusSkip indicates the proof was intentionally skipped.
	ProofStatusSkip ProofStatus = "SKIP"

	// ProofStatusUnknown indicates the proof was not collected.
	ProofStatusUnknown ProofStatus = "UNKNOWN"
)

// ValidProofStatuses returns all valid proof status strings.
func ValidProofStatuses() []string {
	return []string{
		string(ProofStatusPass),
		string(ProofStatusFail),
		string(ProofStatusSkip),
		string(ProofStatusUnknown),
	}
}

// SailsValidationResult contains the result of WHITE_SAILS validation.
type SailsValidationResult struct {
	// Valid is true if the sails data passed schema validation.
	Valid bool `json:"valid"`

	// Issues contains validation problems (empty if Valid is true).
	Issues []ValidationIssue `json:"issues,omitempty"`

	// Color is the parsed color value if valid.
	Color SailsColor `json:"color,omitempty"`

	// ComputedBase is the computed base color before modifiers.
	ComputedBase SailsColor `json:"computed_base,omitempty"`

	// SessionID is the parsed session ID if valid.
	SessionID string `json:"session_id,omitempty"`
}

// ValidateWhiteSails validates WHITE_SAILS YAML/JSON data against the schema.
// The data should be the raw bytes of the WHITE_SAILS content (JSON or YAML).
// For YAML input, it is first converted to JSON for schema validation.
func (v *Validator) ValidateWhiteSails(data []byte) (*SailsValidationResult, error) {
	result := &SailsValidationResult{
		Valid: false,
	}

	// Try parsing as JSON first, then YAML
	var parsed any
	var parseErr error

	if err := json.Unmarshal(data, &parsed); err != nil {
		// Try YAML parsing
		yamlParsed, yamlErr := parseYAMLToInterface(data)
		if yamlErr != nil {
			return nil, errors.Wrap(errors.CodeSchemaInvalid,
				"failed to parse WHITE_SAILS data as JSON or YAML", err)
		}
		parsed = yamlParsed
		parseErr = nil
	} else {
		parseErr = nil
	}

	if parseErr != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "failed to parse WHITE_SAILS data", parseErr)
	}

	// Get compiled schema
	schema, err := v.getSchema("white-sails")
	if err != nil {
		return nil, err
	}

	// Validate against schema
	if err := schema.Validate(parsed); err != nil {
		result.Issues = extractValidationIssues(err)
		return result, nil
	}

	// Extract key fields for convenience
	if m, ok := parsed.(map[string]any); ok {
		if color, ok := m["color"].(string); ok {
			result.Color = ParseSailsColor(color)
		}
		if computedBase, ok := m["computed_base"].(string); ok {
			result.ComputedBase = ParseSailsColor(computedBase)
		}
		if sessionID, ok := m["session_id"].(string); ok {
			result.SessionID = sessionID
		}
	}

	result.Valid = true
	return result, nil
}

// ValidateWhiteSailsMap validates a WHITE_SAILS map against the schema.
func (v *Validator) ValidateWhiteSailsMap(data map[string]any) (*SailsValidationResult, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "failed to marshal WHITE_SAILS data", err)
	}
	return v.ValidateWhiteSails(jsonData)
}

// parseYAMLToInterface parses YAML bytes to a generic interface.
// This is used for YAML input that needs JSON schema validation.
func parseYAMLToInterface(data []byte) (any, error) {
	var result any
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	// Convert YAML map[interface{}]interface{} to JSON-compatible map[string]interface{}
	return convertYAMLToJSON(result), nil
}

// convertYAMLToJSON recursively converts YAML structures to JSON-compatible structures.
func convertYAMLToJSON(v any) any {
	switch val := v.(type) {
	case map[any]any:
		result := make(map[string]any)
		for k, v := range val {
			result[fmt.Sprintf("%v", k)] = convertYAMLToJSON(v)
		}
		return result
	case map[string]any:
		result := make(map[string]any)
		for k, v := range val {
			result[k] = convertYAMLToJSON(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = convertYAMLToJSON(v)
		}
		return result
	default:
		return v
	}
}

// --- Lightweight validation without full schema ---

// ValidateSailsFields performs lightweight validation of required WHITE_SAILS fields.
// This is used when the full schema validator is not available.
func ValidateSailsFields(data map[string]any) []string {
	var issues []string

	// Required fields per schema
	required := []string{"schema_version", "session_id", "generated_at", "color", "computed_base", "proofs", "open_questions"}
	for _, field := range required {
		if _, ok := data[field]; !ok {
			issues = append(issues, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate color enum
	if color, ok := data["color"].(string); ok {
		switch color {
		case "WHITE", "GRAY", "BLACK":
			// Valid
		default:
			issues = append(issues, fmt.Sprintf("invalid color: %s (must be WHITE, GRAY, or BLACK)", color))
		}
	}

	// Validate computed_base enum
	if computedBase, ok := data["computed_base"].(string); ok {
		switch computedBase {
		case "WHITE", "GRAY", "BLACK":
			// Valid
		default:
			issues = append(issues, fmt.Sprintf("invalid computed_base: %s (must be WHITE, GRAY, or BLACK)", computedBase))
		}
	}

	// Validate session_id format
	if sessionID, ok := data["session_id"].(string); ok {
		if !IsValidSessionID(sessionID) {
			issues = append(issues, fmt.Sprintf("invalid session_id format: %s", sessionID))
		}
	}

	// Validate proofs structure
	if proofs, ok := data["proofs"].(map[string]any); ok {
		requiredProofs := []string{"tests", "build", "lint"}
		for _, proofType := range requiredProofs {
			if _, ok := proofs[proofType]; !ok {
				issues = append(issues, fmt.Sprintf("missing required proof: proofs.%s", proofType))
			}
		}
	}

	// Validate schema_version format
	if version, ok := data["schema_version"].(string); ok {
		if !sailsVersionPattern.MatchString(version) {
			issues = append(issues, fmt.Sprintf("invalid schema_version format: %s (expected X.Y or X.Y.Z)", version))
		}
	}

	return issues
}

// sailsVersionPattern matches semantic version patterns (e.g., 1.0, 1.0.0)
var sailsVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+(\.[0-9]+)?$`)

// IsValidSailsColor checks if a color value is valid.
func IsValidSailsColor(color string) bool {
	switch color {
	case "WHITE", "GRAY", "BLACK":
		return true
	default:
		return false
	}
}

// IsValidProofStatus checks if a proof status value is valid.
func IsValidProofStatus(status string) bool {
	switch status {
	case "PASS", "FAIL", "SKIP", "UNKNOWN":
		return true
	default:
		return false
	}
}
