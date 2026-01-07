package validation

import (
	"testing"
)

func TestParseSailsColor(t *testing.T) {
	tests := []struct {
		input string
		want  SailsColor
	}{
		{"WHITE", SailsColorWhite},
		{"white", SailsColorWhite},
		{"White", SailsColorWhite},
		{"GRAY", SailsColorGray},
		{"gray", SailsColorGray},
		{"BLACK", SailsColorBlack},
		{"black", SailsColorBlack},
		{"", ""},
		{"unknown", ""},
		{"green", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseSailsColor(tt.input); got != tt.want {
				t.Errorf("ParseSailsColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidSailsColor(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"WHITE", true},
		{"GRAY", true},
		{"BLACK", true},
		{"white", false}, // case-sensitive
		{"green", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsValidSailsColor(tt.input); got != tt.want {
				t.Errorf("IsValidSailsColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidProofStatus(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"PASS", true},
		{"FAIL", true},
		{"SKIP", true},
		{"UNKNOWN", true},
		{"pass", false}, // case-sensitive
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsValidProofStatus(tt.input); got != tt.want {
				t.Errorf("IsValidProofStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateWhiteSails_ValidJSON(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	validJSON := []byte(`{
		"schema_version": "1.0",
		"session_id": "session-20260105-143000-abc12345",
		"generated_at": "2026-01-05T14:30:00Z",
		"color": "WHITE",
		"computed_base": "WHITE",
		"proofs": {
			"tests": {"status": "PASS"},
			"build": {"status": "PASS"},
			"lint": {"status": "PASS"}
		},
		"open_questions": []
	}`)

	result, err := v.ValidateWhiteSails(validJSON)
	if err != nil {
		t.Fatalf("ValidateWhiteSails() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("ValidateWhiteSails() Valid = false, want true; issues = %v", result.Issues)
	}

	if result.Color != SailsColorWhite {
		t.Errorf("ValidateWhiteSails() Color = %v, want %v", result.Color, SailsColorWhite)
	}

	if result.ComputedBase != SailsColorWhite {
		t.Errorf("ValidateWhiteSails() ComputedBase = %v, want %v", result.ComputedBase, SailsColorWhite)
	}

	if result.SessionID != "session-20260105-143000-abc12345" {
		t.Errorf("ValidateWhiteSails() SessionID = %v, want session-20260105-143000-abc12345", result.SessionID)
	}
}

func TestValidateWhiteSails_ValidYAML(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	validYAML := []byte(`
schema_version: "1.0"
session_id: session-20260105-143000-abc12345
generated_at: "2026-01-05T14:30:00Z"
color: GRAY
computed_base: GRAY
proofs:
  tests:
    status: PASS
  build:
    status: PASS
  lint:
    status: PASS
open_questions:
  - "Does the API handle edge cases?"
`)

	result, err := v.ValidateWhiteSails(validYAML)
	if err != nil {
		t.Fatalf("ValidateWhiteSails() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("ValidateWhiteSails() Valid = false, want true; issues = %v", result.Issues)
	}

	if result.Color != SailsColorGray {
		t.Errorf("ValidateWhiteSails() Color = %v, want %v", result.Color, SailsColorGray)
	}
}

func TestValidateWhiteSails_MissingRequiredFields(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Missing color and proofs
	invalidJSON := []byte(`{
		"schema_version": "1.0",
		"session_id": "session-20260105-143000-abc12345",
		"generated_at": "2026-01-05T14:30:00Z",
		"computed_base": "WHITE",
		"open_questions": []
	}`)

	result, err := v.ValidateWhiteSails(invalidJSON)
	if err != nil {
		t.Fatalf("ValidateWhiteSails() error = %v", err)
	}

	if result.Valid {
		t.Errorf("ValidateWhiteSails() Valid = true, want false for missing required fields")
	}

	if len(result.Issues) == 0 {
		t.Errorf("ValidateWhiteSails() Issues is empty, expected validation errors")
	}
}

func TestValidateWhiteSails_InvalidColor(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	invalidJSON := []byte(`{
		"schema_version": "1.0",
		"session_id": "session-20260105-143000-abc12345",
		"generated_at": "2026-01-05T14:30:00Z",
		"color": "GREEN",
		"computed_base": "WHITE",
		"proofs": {
			"tests": {"status": "PASS"},
			"build": {"status": "PASS"},
			"lint": {"status": "PASS"}
		},
		"open_questions": []
	}`)

	result, err := v.ValidateWhiteSails(invalidJSON)
	if err != nil {
		t.Fatalf("ValidateWhiteSails() error = %v", err)
	}

	if result.Valid {
		t.Errorf("ValidateWhiteSails() Valid = true, want false for invalid color")
	}
}

func TestValidateWhiteSails_InvalidSessionIDFormat(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	invalidJSON := []byte(`{
		"schema_version": "1.0",
		"session_id": "invalid-session-id",
		"generated_at": "2026-01-05T14:30:00Z",
		"color": "WHITE",
		"computed_base": "WHITE",
		"proofs": {
			"tests": {"status": "PASS"},
			"build": {"status": "PASS"},
			"lint": {"status": "PASS"}
		},
		"open_questions": []
	}`)

	result, err := v.ValidateWhiteSails(invalidJSON)
	if err != nil {
		t.Fatalf("ValidateWhiteSails() error = %v", err)
	}

	if result.Valid {
		t.Errorf("ValidateWhiteSails() Valid = true, want false for invalid session_id format")
	}
}

func TestValidateWhiteSails_WithModifiers(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	validJSON := []byte(`{
		"schema_version": "1.0",
		"session_id": "session-20260105-143000-abc12345",
		"generated_at": "2026-01-05T14:30:00Z",
		"color": "GRAY",
		"computed_base": "WHITE",
		"proofs": {
			"tests": {"status": "PASS"},
			"build": {"status": "PASS"},
			"lint": {"status": "PASS"}
		},
		"open_questions": [],
		"modifiers": [
			{
				"type": "DOWNGRADE_TO_GRAY",
				"justification": "Requires additional security review before production",
				"applied_by": "human"
			}
		]
	}`)

	result, err := v.ValidateWhiteSails(validJSON)
	if err != nil {
		t.Fatalf("ValidateWhiteSails() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("ValidateWhiteSails() Valid = false, want true; issues = %v", result.Issues)
	}
}

func TestValidateWhiteSailsMap(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	data := map[string]interface{}{
		"schema_version": "1.0",
		"session_id":     "session-20260105-143000-abc12345",
		"generated_at":   "2026-01-05T14:30:00Z",
		"color":          "WHITE",
		"computed_base":  "WHITE",
		"proofs": map[string]interface{}{
			"tests": map[string]interface{}{"status": "PASS"},
			"build": map[string]interface{}{"status": "PASS"},
			"lint":  map[string]interface{}{"status": "PASS"},
		},
		"open_questions": []interface{}{},
	}

	result, err := v.ValidateWhiteSailsMap(data)
	if err != nil {
		t.Fatalf("ValidateWhiteSailsMap() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("ValidateWhiteSailsMap() Valid = false, want true; issues = %v", result.Issues)
	}
}

func TestValidateSailsFields(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]interface{}
		wantIssues bool
	}{
		{
			name: "valid data",
			data: map[string]interface{}{
				"schema_version": "1.0",
				"session_id":     "session-20260105-143000-abc12345",
				"generated_at":   "2026-01-05T14:30:00Z",
				"color":          "WHITE",
				"computed_base":  "WHITE",
				"proofs": map[string]interface{}{
					"tests": map[string]interface{}{"status": "PASS"},
					"build": map[string]interface{}{"status": "PASS"},
					"lint":  map[string]interface{}{"status": "PASS"},
				},
				"open_questions": []interface{}{},
			},
			wantIssues: false,
		},
		{
			name: "missing required field",
			data: map[string]interface{}{
				"schema_version": "1.0",
				"session_id":     "session-20260105-143000-abc12345",
			},
			wantIssues: true,
		},
		{
			name: "invalid color",
			data: map[string]interface{}{
				"schema_version": "1.0",
				"session_id":     "session-20260105-143000-abc12345",
				"generated_at":   "2026-01-05T14:30:00Z",
				"color":          "GREEN",
				"computed_base":  "WHITE",
				"proofs": map[string]interface{}{
					"tests": map[string]interface{}{"status": "PASS"},
					"build": map[string]interface{}{"status": "PASS"},
					"lint":  map[string]interface{}{"status": "PASS"},
				},
				"open_questions": []interface{}{},
			},
			wantIssues: true,
		},
		{
			name: "invalid session_id format",
			data: map[string]interface{}{
				"schema_version": "1.0",
				"session_id":     "bad-session-id",
				"generated_at":   "2026-01-05T14:30:00Z",
				"color":          "WHITE",
				"computed_base":  "WHITE",
				"proofs": map[string]interface{}{
					"tests": map[string]interface{}{"status": "PASS"},
					"build": map[string]interface{}{"status": "PASS"},
					"lint":  map[string]interface{}{"status": "PASS"},
				},
				"open_questions": []interface{}{},
			},
			wantIssues: true,
		},
		{
			name: "missing required proof",
			data: map[string]interface{}{
				"schema_version": "1.0",
				"session_id":     "session-20260105-143000-abc12345",
				"generated_at":   "2026-01-05T14:30:00Z",
				"color":          "WHITE",
				"computed_base":  "WHITE",
				"proofs": map[string]interface{}{
					"tests": map[string]interface{}{"status": "PASS"},
					// missing build and lint
				},
				"open_questions": []interface{}{},
			},
			wantIssues: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidateSailsFields(tt.data)
			hasIssues := len(issues) > 0
			if hasIssues != tt.wantIssues {
				t.Errorf("ValidateSailsFields() hasIssues = %v, want %v; issues = %v", hasIssues, tt.wantIssues, issues)
			}
		})
	}
}
