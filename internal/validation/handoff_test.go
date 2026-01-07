package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePhase(t *testing.T) {
	tests := []struct {
		input string
		want  Phase
	}{
		{"requirements", PhaseRequirements},
		{"Requirements", PhaseRequirements},
		{"REQUIREMENTS", PhaseRequirements},
		{"design", PhaseDesign},
		{"implementation", PhaseImplementation},
		{"validation", PhaseValidation},
		{"", ""},
		{"unknown", ""},
		{"complete", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParsePhase(tt.input); got != tt.want {
				t.Errorf("ParsePhase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidPhases(t *testing.T) {
	phases := ValidPhases()
	expected := []string{"requirements", "design", "implementation", "validation"}

	if len(phases) != len(expected) {
		t.Errorf("ValidPhases() length = %d, want %d", len(phases), len(expected))
	}

	for _, e := range expected {
		found := false
		for _, p := range phases {
			if p == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidPhases() missing %q", e)
		}
	}
}

func TestNewHandoffValidator(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	if hv == nil {
		t.Fatal("NewHandoffValidator() returned nil")
	}

	if len(hv.criteria) == 0 {
		t.Error("NewHandoffValidator() criteria is empty")
	}
}

func TestHandoffValidator_GetCriteria(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	tests := []struct {
		name         string
		phase        Phase
		artifactType ArtifactType
		wantBlocking bool
		wantErr      bool
	}{
		{
			name:         "requirements PRD",
			phase:        PhaseRequirements,
			artifactType: ArtifactTypePRD,
			wantBlocking: true,
			wantErr:      false,
		},
		{
			name:         "design TDD",
			phase:        PhaseDesign,
			artifactType: ArtifactTypeTDD,
			wantBlocking: true,
			wantErr:      false,
		},
		{
			name:         "design ADR",
			phase:        PhaseDesign,
			artifactType: ArtifactTypeADR,
			wantBlocking: true,
			wantErr:      false,
		},
		{
			name:         "validation test-plan",
			phase:        PhaseValidation,
			artifactType: ArtifactTypeTestPlan,
			wantBlocking: true,
			wantErr:      false,
		},
		{
			name:         "invalid phase",
			phase:        Phase("invalid"),
			artifactType: ArtifactTypePRD,
			wantErr:      true,
		},
		{
			name:         "undefined artifact for phase returns empty",
			phase:        PhaseRequirements,
			artifactType: ArtifactTypeTestPlan,
			wantBlocking: false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria, err := hv.GetCriteria(tt.phase, tt.artifactType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCriteria() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if tt.wantBlocking && len(criteria.Blocking) == 0 {
				t.Error("GetCriteria() expected blocking criteria")
			}
		})
	}
}

func TestHandoffValidator_ListPhases(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	phases := hv.ListPhases()
	if len(phases) == 0 {
		t.Error("ListPhases() returned empty")
	}

	// Should have at least requirements and design
	foundReq := false
	foundDesign := false
	for _, p := range phases {
		if p == PhaseRequirements {
			foundReq = true
		}
		if p == PhaseDesign {
			foundDesign = true
		}
	}

	if !foundReq {
		t.Error("ListPhases() missing requirements")
	}
	if !foundDesign {
		t.Error("ListPhases() missing design")
	}
}

func TestHandoffValidator_ListArtifactTypes(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	types := hv.ListArtifactTypes(PhaseRequirements)
	if len(types) == 0 {
		t.Error("ListArtifactTypes(requirements) returned empty")
	}

	// Should have PRD for requirements phase
	foundPRD := false
	for _, at := range types {
		if at == ArtifactTypePRD {
			foundPRD = true
			break
		}
	}

	if !foundPRD {
		t.Error("ListArtifactTypes(requirements) missing PRD")
	}
}

func TestHandoffValidator_ValidateHandoff_PassingPRD(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	frontmatter := map[string]interface{}{
		"artifact_id":      "PRD-test-feature",
		"title":            "Test Feature",
		"status":           "approved",
		"success_criteria": []string{"Users can authenticate", "Error handling works"},
		"stakeholders":     []string{"product", "engineering"},
		"complexity":       "MODULE",
	}

	result, err := hv.ValidateHandoff(PhaseRequirements, ArtifactTypePRD, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	if !result.Passed {
		t.Error("ValidateHandoff() Passed = false, want true")
		for _, cr := range result.FailedBlocking() {
			t.Logf("  Failed: %s", cr.Message)
		}
	}

	if len(result.Warnings()) > 0 {
		t.Logf("Warnings: %d", len(result.Warnings()))
	}
}

func TestHandoffValidator_ValidateHandoff_FailingPRD_MissingRequired(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	// Missing success_criteria
	frontmatter := map[string]interface{}{
		"artifact_id": "PRD-test-feature",
		"title":       "Test Feature",
		"status":      "approved",
	}

	result, err := hv.ValidateHandoff(PhaseRequirements, ArtifactTypePRD, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	if result.Passed {
		t.Error("ValidateHandoff() Passed = true, want false for missing success_criteria")
	}

	// Check that success_criteria failure is in blocking results
	found := false
	for _, cr := range result.BlockingResults {
		if cr.Criterion.Field == "success_criteria" && !cr.Passed {
			found = true
			break
		}
	}

	if !found {
		t.Error("ValidateHandoff() expected failure for success_criteria field")
	}
}

func TestHandoffValidator_ValidateHandoff_NonBlockingWarnings(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	// Has blocking fields but missing non-blocking
	frontmatter := map[string]interface{}{
		"artifact_id":      "PRD-test-feature",
		"title":            "Test Feature",
		"status":           "approved",
		"success_criteria": []string{"Users can authenticate"},
		// Missing: stakeholders, complexity (non-blocking)
	}

	result, err := hv.ValidateHandoff(PhaseRequirements, ArtifactTypePRD, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	// Should pass (blocking satisfied)
	if !result.Passed {
		t.Error("ValidateHandoff() Passed = false, want true (only non-blocking missing)")
	}

	// Should have warnings
	warnings := result.Warnings()
	if len(warnings) == 0 {
		t.Error("ValidateHandoff() expected warnings for missing non-blocking fields")
	}
}

func TestHandoffValidator_ValidateHandoff_EmptyField(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	// artifact_id is present but empty
	frontmatter := map[string]interface{}{
		"artifact_id":      "",
		"title":            "Test Feature",
		"status":           "approved",
		"success_criteria": []string{"Users can authenticate"},
	}

	result, err := hv.ValidateHandoff(PhaseRequirements, ArtifactTypePRD, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	if result.Passed {
		t.Error("ValidateHandoff() Passed = true, want false for empty artifact_id")
	}
}

func TestHandoffValidator_ValidateHandoff_MinItems(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	// success_criteria is present but empty array
	frontmatter := map[string]interface{}{
		"artifact_id":      "PRD-test-feature",
		"title":            "Test Feature",
		"status":           "approved",
		"success_criteria": []string{},
	}

	result, err := hv.ValidateHandoff(PhaseRequirements, ArtifactTypePRD, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	if result.Passed {
		t.Error("ValidateHandoff() Passed = true, want false for empty success_criteria array")
	}
}

func TestHandoffValidator_ValidateHandoff_DesignTDD(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	frontmatter := map[string]interface{}{
		"artifact_id":         "TDD-feature-design",
		"title":               "Feature Design",
		"status":              "approved",
		"prd_ref":             "PRD-test-feature",
		"implementation_plan": "Step-by-step implementation plan here",
	}

	result, err := hv.ValidateHandoff(PhaseDesign, ArtifactTypeTDD, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	if !result.Passed {
		t.Error("ValidateHandoff() Passed = false, want true")
		for _, cr := range result.FailedBlocking() {
			t.Logf("  Failed: %s", cr.Message)
		}
	}
}

func TestHandoffValidator_ValidateHandoff_DesignADR(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	frontmatter := map[string]interface{}{
		"artifact_id": "ADR-0001",
		"title":       "Use JSON Schema for Validation",
		"status":      "accepted",
		"date":        "2025-12-29",
	}

	result, err := hv.ValidateHandoff(PhaseDesign, ArtifactTypeADR, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	if !result.Passed {
		t.Error("ValidateHandoff() Passed = false, want true")
		for _, cr := range result.FailedBlocking() {
			t.Logf("  Failed: %s", cr.Message)
		}
	}
}

func TestHandoffValidator_ValidateHandoff_ValidationTestPlan(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	frontmatter := map[string]interface{}{
		"artifact_id": "TEST-feature-validation",
		"title":       "Feature Validation Test Plan",
		"status":      "approved",
	}

	result, err := hv.ValidateHandoff(PhaseValidation, ArtifactTypeTestPlan, frontmatter)
	if err != nil {
		t.Fatalf("ValidateHandoff() error = %v", err)
	}

	if !result.Passed {
		t.Error("ValidateHandoff() Passed = false, want true")
		for _, cr := range result.FailedBlocking() {
			t.Logf("  Failed: %s", cr.Message)
		}
	}
}

func TestHandoffValidator_ValidateHandoffFile(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	// Create a temp file with valid PRD content
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "PRD-test.md")

	content := `---
artifact_id: PRD-test
title: Test Feature
status: approved
created_at: 2025-12-29T20:00:00Z
author: test-user
success_criteria:
  - Users can authenticate
  - Error handling works
---
# PRD: Test Feature

Content here.
`

	if err := os.WriteFile(prdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := hv.ValidateHandoffFile(PhaseRequirements, prdFile)
	if err != nil {
		t.Fatalf("ValidateHandoffFile() error = %v", err)
	}

	if !result.Passed {
		t.Error("ValidateHandoffFile() Passed = false, want true")
		for _, cr := range result.FailedBlocking() {
			t.Logf("  Failed: %s", cr.Message)
		}
	}

	if result.FilePath != prdFile {
		t.Errorf("ValidateHandoffFile() FilePath = %q, want %q", result.FilePath, prdFile)
	}
}

func TestHandoffValidator_ValidateHandoffFile_FailingFile(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	// Create a temp file missing required fields
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "PRD-incomplete.md")

	content := `---
artifact_id: PRD-incomplete
title: Incomplete PRD
status: draft
created_at: 2025-12-29T20:00:00Z
author: test-user
---
# PRD: Incomplete

Missing success_criteria.
`

	if err := os.WriteFile(prdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := hv.ValidateHandoffFile(PhaseRequirements, prdFile)
	if err != nil {
		t.Fatalf("ValidateHandoffFile() error = %v", err)
	}

	if result.Passed {
		t.Error("ValidateHandoffFile() Passed = true, want false for missing success_criteria")
	}
}

func TestHandoffValidator_ValidateHandoffFile_NoFrontmatter(t *testing.T) {
	hv, err := NewHandoffValidator()
	if err != nil {
		t.Fatalf("NewHandoffValidator() error = %v", err)
	}

	// Create a temp file without frontmatter
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "PRD-no-frontmatter.md")

	content := `# PRD: No Frontmatter

This document has no frontmatter.
`

	if err := os.WriteFile(prdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := hv.ValidateHandoffFile(PhaseRequirements, prdFile)
	if err != nil {
		t.Fatalf("ValidateHandoffFile() error = %v", err)
	}

	if result.Passed {
		t.Error("ValidateHandoffFile() Passed = true, want false for missing frontmatter")
	}
}

func TestHandoffResult_FailedBlocking(t *testing.T) {
	result := &HandoffResult{
		BlockingResults: []CriterionResult{
			{Criterion: Criterion{Field: "field1"}, Passed: true},
			{Criterion: Criterion{Field: "field2"}, Passed: false, Message: "failed"},
			{Criterion: Criterion{Field: "field3"}, Passed: false, Message: "also failed"},
		},
	}

	failed := result.FailedBlocking()
	if len(failed) != 2 {
		t.Errorf("FailedBlocking() length = %d, want 2", len(failed))
	}
}

func TestHandoffResult_Warnings(t *testing.T) {
	result := &HandoffResult{
		WarningResults: []CriterionResult{
			{Criterion: Criterion{Field: "warn1"}, Passed: false, Message: "warning 1"},
			{Criterion: Criterion{Field: "warn2"}, Passed: false, Message: "warning 2"},
		},
	}

	warnings := result.Warnings()
	if len(warnings) != 2 {
		t.Errorf("Warnings() length = %d, want 2", len(warnings))
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"non-empty string", "hello", false},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"a"}, false},
		{"empty map", map[string]string{}, true},
		{"non-empty map", map[string]string{"a": "b"}, false},
		{"zero int", 0, false},
		{"non-zero int", 42, false},
		{"false bool", false, false},
		{"true bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmpty(tt.value); got != tt.want {
				t.Errorf("isEmpty(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestGetItemCount(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  int
	}{
		{"nil", nil, 0},
		{"empty slice", []string{}, 0},
		{"slice with items", []string{"a", "b", "c"}, 3},
		{"non-array value", "hello", 1},
		{"int value", 42, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getItemCount(tt.value); got != tt.want {
				t.Errorf("getItemCount(%v) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestPhase_String(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseRequirements, "requirements"},
		{PhaseDesign, "design"},
		{PhaseImplementation, "implementation"},
		{PhaseValidation, "validation"},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := tt.phase.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
