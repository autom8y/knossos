package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/errors"
)

func TestParseArtifactType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  ArtifactType
	}{
		{"prd", ArtifactTypePRD},
		{"PRD", ArtifactTypePRD},
		{"Prd", ArtifactTypePRD},
		{"tdd", ArtifactTypeTDD},
		{"TDD", ArtifactTypeTDD},
		{"adr", ArtifactTypeADR},
		{"ADR", ArtifactTypeADR},
		{"test-plan", ArtifactTypeTestPlan},
		{"testplan", ArtifactTypeTestPlan},
		{"test_plan", ArtifactTypeTestPlan},
		{"tp", ArtifactTypeTestPlan},
		{"test", ArtifactTypeTestPlan},
		{"", ArtifactTypeUnknown},
		{"unknown", ArtifactTypeUnknown},
		{"something-else", ArtifactTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := ParseArtifactType(tt.input); got != tt.want {
				t.Errorf("ParseArtifactType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectArtifactType_Filename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename string
		want     ArtifactType
	}{
		{"PRD-user-auth.md", ArtifactTypePRD},
		{"PRD-feature-123.md", ArtifactTypePRD},
		{"TDD-user-auth.md", ArtifactTypeTDD},
		{"TDD-api-design.md", ArtifactTypeTDD},
		{"ADR-0001.md", ArtifactTypeADR},
		{"ADR-0010.md", ArtifactTypeADR},
		{"ADR-0001-some-decision.md", ArtifactTypeADR},
		{"TEST-validation.md", ArtifactTypeTestPlan},
		{"TP-integration.md", ArtifactTypeTestPlan},
		{"README.md", ArtifactTypeUnknown},
		{"some-file.md", ArtifactTypeUnknown},
		{"prd-lowercase.md", ArtifactTypeUnknown}, // Must be uppercase
		{"PRD.md", ArtifactTypeUnknown},           // Missing name part
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			got := DetectArtifactType(tt.filename, nil)
			if got != tt.want {
				t.Errorf("DetectArtifactType(%q, nil) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectArtifactType_Frontmatter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		filename    string
		frontmatter map[string]interface{}
		want        ArtifactType
	}{
		{
			name:        "frontmatter type takes priority",
			filename:    "some-file.md",
			frontmatter: map[string]interface{}{"type": "prd"},
			want:        ArtifactTypePRD,
		},
		{
			name:        "frontmatter overrides filename",
			filename:    "TDD-something.md",
			frontmatter: map[string]interface{}{"type": "prd"},
			want:        ArtifactTypePRD,
		},
		{
			name:        "falls back to filename when no type field",
			filename:    "PRD-feature.md",
			frontmatter: map[string]interface{}{"title": "Feature"},
			want:        ArtifactTypePRD,
		},
		{
			name:        "unknown type in frontmatter falls back to filename",
			filename:    "TDD-design.md",
			frontmatter: map[string]interface{}{"type": "invalid"},
			want:        ArtifactTypeTDD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DetectArtifactType(tt.filename, tt.frontmatter)
			if got != tt.want {
				t.Errorf("DetectArtifactType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArtifactType_SchemaName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		aType ArtifactType
		want  string
	}{
		{ArtifactTypePRD, "prd"},
		{ArtifactTypeTDD, "tdd"},
		{ArtifactTypeADR, "adr"},
		{ArtifactTypeTestPlan, "test-plan"},
		{ArtifactTypeUnknown, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.aType), func(t *testing.T) {
			t.Parallel()
			if got := tt.aType.SchemaName(); got != tt.want {
				t.Errorf("SchemaName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArtifactValidator_Validate_ValidPRD(t *testing.T) {
	t.Parallel()
	content := []byte(`---
artifact_id: PRD-test-feature
title: Test Feature PRD
status: draft
created_at: 2025-12-29T20:00:00Z
author: test-user
---
# PRD: Test Feature

This is the content.
`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	result, err := validator.Validate(content, "PRD-test-feature.md", ArtifactTypeUnknown)
	if err != nil {
		// Print details if available
		if e, ok := err.(*errors.Error); ok && e.Details != nil {
			t.Logf("Error details: %v", e.Details)
		}
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Validate() Valid = false, want true")
		for _, issue := range result.Issues {
			t.Logf("  Issue: %s", issue.Message)
		}
	}

	if result.ArtifactType != ArtifactTypePRD {
		t.Errorf("Validate() ArtifactType = %v, want %v", result.ArtifactType, ArtifactTypePRD)
	}
}

func TestArtifactValidator_Validate_ValidTDD(t *testing.T) {
	t.Parallel()
	content := []byte(`---
artifact_id: TDD-feature-design
title: Feature Design
status: approved
prd_ref: PRD-test-feature
---
# TDD: Feature Design

Technical design here.
`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	result, err := validator.Validate(content, "TDD-feature-design.md", ArtifactTypeUnknown)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Validate() Valid = false, want true")
		for _, issue := range result.Issues {
			t.Logf("  Issue: %s", issue.Message)
		}
	}
}

func TestArtifactValidator_Validate_ValidADR(t *testing.T) {
	t.Parallel()
	// Note: YAML parses bare dates as timestamps, so we quote it to keep as string
	content := []byte(`---
artifact_id: ADR-0001
title: Use JSON Schema for Validation
status: accepted
date: "2025-12-29"
---
# ADR-0001: Use JSON Schema for Validation

Decision content.
`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	result, err := validator.Validate(content, "ADR-0001.md", ArtifactTypeUnknown)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Validate() Valid = false, want true")
		for _, issue := range result.Issues {
			t.Logf("  Issue: %s", issue.Message)
		}
	}
}

func TestArtifactValidator_Validate_ValidTestPlan(t *testing.T) {
	t.Parallel()
	content := []byte(`---
artifact_id: TEST-feature-validation
title: Feature Validation Test Plan
status: approved
---
# Test Plan: Feature Validation

Test cases here.
`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	result, err := validator.Validate(content, "TEST-feature-validation.md", ArtifactTypeUnknown)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Validate() Valid = false, want true")
		for _, issue := range result.Issues {
			t.Logf("  Issue: %s", issue.Message)
		}
	}
}

func TestArtifactValidator_Validate_MissingRequiredFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		content      string
		artifactType ArtifactType
		wantField    string
	}{
		{
			name: "PRD missing artifact_id",
			content: `---
title: Test Feature
status: draft
created_at: 2025-12-29T20:00:00Z
author: test-user
---
Content`,
			artifactType: ArtifactTypePRD,
			wantField:    "artifact_id",
		},
		{
			name: "TDD missing prd_ref",
			content: `---
artifact_id: TDD-test
title: Test Design
status: draft
---
Content`,
			artifactType: ArtifactTypeTDD,
			wantField:    "prd_ref",
		},
		{
			name: "ADR missing date",
			content: `---
artifact_id: ADR-0001
title: Test Decision
status: accepted
---
Content`,
			artifactType: ArtifactTypeADR,
			wantField:    "date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator, err := NewArtifactValidator()
			if err != nil {
				t.Fatalf("NewArtifactValidator() error = %v", err)
			}

			result, err := validator.Validate([]byte(tt.content), "test.md", tt.artifactType)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			if result.Valid {
				t.Error("Validate() Valid = true, want false")
				return
			}

			// Check that the expected field is mentioned in issues
			foundField := false
			for _, issue := range result.Issues {
				if issue.Field == tt.wantField || contains(issue.Message, tt.wantField) {
					foundField = true
					break
				}
			}

			if !foundField {
				t.Errorf("Validate() expected issue for field %q, got issues: %v", tt.wantField, result.Issues)
			}
		})
	}
}

func TestArtifactValidator_Validate_InvalidFieldValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "PRD with invalid status enum",
			content: `---
artifact_id: PRD-test
title: Test
status: invalid_status
created_at: 2025-12-29T20:00:00Z
author: test-user
---
Content`,
		},
		{
			name: "PRD with invalid artifact_id pattern",
			content: `---
artifact_id: invalid-id
title: Test
status: draft
created_at: 2025-12-29T20:00:00Z
author: test-user
---
Content`,
		},
		{
			name: "ADR with invalid date format",
			content: `---
artifact_id: ADR-0001
title: Test
status: accepted
date: 12/29/2025
---
Content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator, err := NewArtifactValidator()
			if err != nil {
				t.Fatalf("NewArtifactValidator() error = %v", err)
			}

			result, err := validator.Validate([]byte(tt.content), "test.md", ArtifactTypePRD)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			if result.Valid {
				t.Error("Validate() Valid = true, want false for invalid field values")
			}
		})
	}
}

func TestArtifactValidator_Validate_NoFrontmatter(t *testing.T) {
	t.Parallel()
	content := []byte(`# Document Title

This document has no frontmatter.
`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	result, err := validator.Validate(content, "PRD-test.md", ArtifactTypePRD)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Valid {
		t.Error("Validate() Valid = true, want false for missing frontmatter")
	}

	if len(result.Issues) == 0 {
		t.Error("Validate() should have issues for missing frontmatter")
	}
}

func TestArtifactValidator_Validate_UnknownType(t *testing.T) {
	t.Parallel()
	content := []byte(`---
title: Some Document
---
Content`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	// Use a filename that doesn't match any pattern
	result, err := validator.Validate(content, "README.md", ArtifactTypeUnknown)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Valid {
		t.Error("Validate() Valid = true, want false for unknown type")
	}

	if result.ArtifactType != ArtifactTypeUnknown {
		t.Errorf("Validate() ArtifactType = %v, want Unknown", result.ArtifactType)
	}
}

func TestArtifactValidator_ValidateFile(t *testing.T) {
	t.Parallel()
	// Create a temp file with valid PRD content
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "PRD-test.md")

	content := `---
artifact_id: PRD-test
title: Test Feature
status: draft
created_at: 2025-12-29T20:00:00Z
author: test-user
---
# PRD: Test Feature

Content here.
`

	if err := os.WriteFile(prdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	result, err := validator.ValidateFile(prdFile, ArtifactTypeUnknown)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}

	if !result.Valid {
		t.Error("ValidateFile() Valid = false, want true")
		for _, issue := range result.Issues {
			t.Logf("  Issue: %s", issue.Message)
		}
	}

	if result.FilePath != prdFile {
		t.Errorf("ValidateFile() FilePath = %q, want %q", result.FilePath, prdFile)
	}
}

func TestArtifactValidator_ValidateFile_NotFound(t *testing.T) {
	t.Parallel()
	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	_, err = validator.ValidateFile("/nonexistent/file.md", ArtifactTypePRD)
	if err == nil {
		t.Error("ValidateFile() expected error for nonexistent file")
	}
}

func TestValidArtifactTypes(t *testing.T) {
	t.Parallel()
	types := ValidArtifactTypes()

	expected := []string{"prd", "tdd", "adr", "test-plan"}
	if len(types) != len(expected) {
		t.Errorf("ValidArtifactTypes() length = %d, want %d", len(types), len(expected))
	}

	for _, e := range expected {
		found := false
		for _, t := range types {
			if t == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidArtifactTypes() missing %q", e)
		}
	}
}

func TestArtifactValidator_ExplicitType(t *testing.T) {
	t.Parallel()
	// Content matches TDD pattern in filename but we force PRD validation
	content := []byte(`---
artifact_id: PRD-forced-type
title: Test Feature
status: draft
created_at: 2025-12-29T20:00:00Z
author: test-user
---
Content`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	// Use TDD filename but force PRD type
	result, err := validator.Validate(content, "TDD-something.md", ArtifactTypePRD)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should validate as PRD (the explicit type)
	if result.ArtifactType != ArtifactTypePRD {
		t.Errorf("Validate() ArtifactType = %v, want PRD", result.ArtifactType)
	}
}

func TestArtifactType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		aType ArtifactType
		want  string
	}{
		{ArtifactTypePRD, "prd"},
		{ArtifactTypeTDD, "tdd"},
		{ArtifactTypeADR, "adr"},
		{ArtifactTypeTestPlan, "test-plan"},
		{ArtifactTypeUnknown, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.aType), func(t *testing.T) {
			t.Parallel()
			if got := tt.aType.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewArtifactValidator(t *testing.T) {
	t.Parallel()
	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	if validator == nil {
		t.Fatal("NewArtifactValidator() returned nil")
	}

	if validator.validator == nil {
		t.Error("NewArtifactValidator() validator.validator is nil")
	}
}

func TestArtifactValidator_Validate_AllSchemas(t *testing.T) {
	t.Parallel()
	// Test that all schema types can be compiled and used
	schemaTypes := []ArtifactType{
		ArtifactTypePRD,
		ArtifactTypeTDD,
		ArtifactTypeADR,
		ArtifactTypeTestPlan,
	}

	for _, schemaType := range schemaTypes {
		t.Run(string(schemaType), func(t *testing.T) {
			t.Parallel()
			validator, err := NewArtifactValidator()
			if err != nil {
				t.Fatalf("NewArtifactValidator() error = %v", err)
			}

			// Validate with empty data to verify schema is loadable
			_, err = validator.validateAgainstSchema(map[string]interface{}{}, schemaType)
			// We expect this to fail (missing required fields) but NOT with a schema loading error
			if err != nil {
				t.Errorf("validateAgainstSchema(%s) schema error: %v", schemaType, err)
			}
		})
	}
}

func TestValidationIssue(t *testing.T) {
	t.Parallel()
	issue := ValidationIssue{
		Field:   "artifact_id",
		Message: "missing required field",
		Value:   nil,
	}

	if issue.Field != "artifact_id" {
		t.Errorf("Field = %q, want artifact_id", issue.Field)
	}
	if issue.Message != "missing required field" {
		t.Errorf("Message = %q, want missing required field", issue.Message)
	}
}

func TestDetectArtifactType_PathVariations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want ArtifactType
	}{
		{"/full/path/to/PRD-feature.md", ArtifactTypePRD},
		{"./relative/TDD-design.md", ArtifactTypeTDD},
		{"ADR-0001.md", ArtifactTypeADR},
		{"docs/tests/TEST-integration.md", ArtifactTypeTestPlan},
		{"docs/plans/TP-validation.md", ArtifactTypeTestPlan},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			got := DetectArtifactType(tt.path, nil)
			if got != tt.want {
				t.Errorf("DetectArtifactType(%q, nil) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestArtifactValidator_Validate_FrontmatterContainedInResult(t *testing.T) {
	t.Parallel()
	content := []byte(`---
artifact_id: PRD-test
title: Test Feature
status: draft
created_at: 2025-12-29T20:00:00Z
author: test-user
custom_field: custom_value
---
Content`)

	validator, err := NewArtifactValidator()
	if err != nil {
		t.Fatalf("NewArtifactValidator() error = %v", err)
	}

	result, err := validator.Validate(content, "PRD-test.md", ArtifactTypeUnknown)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Check that frontmatter is included in result
	if result.Frontmatter == nil {
		t.Fatal("Validate() Frontmatter is nil")
	}

	if result.Frontmatter["artifact_id"] != "PRD-test" {
		t.Errorf("Frontmatter[artifact_id] = %v, want PRD-test", result.Frontmatter["artifact_id"])
	}

	if result.Frontmatter["custom_field"] != "custom_value" {
		t.Errorf("Frontmatter[custom_field] = %v, want custom_value", result.Frontmatter["custom_field"])
	}
}
