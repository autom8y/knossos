package rite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	// Create a temp directory
	tempDir := t.TempDir()

	// Write a valid manifest
	validYAML := `
schema_version: "1.0"
name: test-rite
display_name: "Test Rite"
description: "A test rite"
form: practitioner

agents:
  - name: test-agent
    file: agents/test-agent.md
    role: "Test role"
    produces: output

skills:
  - ref: test-skill
    path: skills/test-skill/

budget:
  estimated_tokens: 5000
  agents_cost: 3000
  skills_cost: 2000
`
	manifestPath := filepath.Join(tempDir, "rite.yaml")
	if err := os.WriteFile(manifestPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to write test manifest: %v", err)
	}

	// Test loading
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// Verify fields
	if manifest.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", manifest.SchemaVersion, "1.0")
	}
	if manifest.Name != "test-rite" {
		t.Errorf("Name = %q, want %q", manifest.Name, "test-rite")
	}
	if manifest.DisplayName != "Test Rite" {
		t.Errorf("DisplayName = %q, want %q", manifest.DisplayName, "Test Rite")
	}
	if manifest.Form != FormPractitioner {
		t.Errorf("Form = %q, want %q", manifest.Form, FormPractitioner)
	}
	if len(manifest.Agents) != 1 {
		t.Errorf("len(Agents) = %d, want 1", len(manifest.Agents))
	}
	if manifest.Agents[0].Name != "test-agent" {
		t.Errorf("Agents[0].Name = %q, want %q", manifest.Agents[0].Name, "test-agent")
	}
	if len(manifest.Skills) != 1 {
		t.Errorf("len(Skills) = %d, want 1", len(manifest.Skills))
	}
	if manifest.Budget.EstimatedTokens != 5000 {
		t.Errorf("Budget.EstimatedTokens = %d, want 5000", manifest.Budget.EstimatedTokens)
	}
}

func TestLoadManifest_NotFound(t *testing.T) {
	_, err := LoadManifest("/nonexistent/path/rite.yaml")
	if err == nil {
		t.Error("LoadManifest should fail for nonexistent file")
	}
}

func TestLoadManifest_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	invalidPath := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadManifest(invalidPath)
	if err == nil {
		t.Error("LoadManifest should fail for invalid YAML")
	}
}

func TestRiteManifest_Validate(t *testing.T) {
	tests := []struct {
		name       string
		manifest   RiteManifest
		wantIssues int
	}{
		{
			name: "valid practitioner",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "valid-rite",
				Form:          FormPractitioner,
				Agents: []AgentRef{
					{Name: "agent1", File: "agents/agent1.md"},
				},
			},
			wantIssues: 0,
		},
		{
			name: "valid simple",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "simple-rite",
				Form:          FormSimple,
				Skills: []SkillRef{
					{Ref: "skill1", Path: "skills/skill1/"},
				},
			},
			wantIssues: 0,
		},
		{
			name: "missing required fields",
			manifest: RiteManifest{
				// Empty - in new format only name is required
			},
			wantIssues: 1, // name
		},
		{
			name: "invalid name format",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "Invalid Name", // Not kebab-case
				Form:          FormSimple,
			},
			wantIssues: 1,
		},
		{
			name: "invalid form",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "test-rite",
				Form:          "invalid-form",
			},
			wantIssues: 1,
		},
		{
			name: "simple with agents (invalid)",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "simple-with-agents",
				Form:          FormSimple,
				Agents: []AgentRef{
					{Name: "agent1", File: "agents/agent1.md"},
				},
			},
			wantIssues: 1, // simple form should not have agents
		},
		{
			name: "practitioner without agents (invalid)",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "practitioner-no-agents",
				Form:          FormPractitioner,
			},
			wantIssues: 1, // practitioner requires agents
		},
		{
			name: "agent missing name",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "agent-no-name",
				Form:          FormPractitioner,
				Agents: []AgentRef{
					{File: "agents/agent1.md"},
				},
			},
			wantIssues: 1, // agents[0].name is required
		},
		{
			name: "external skill with path (invalid)",
			manifest: RiteManifest{
				SchemaVersion: "1.0",
				Name:          "skill-conflict",
				Form:          FormSimple,
				Skills: []SkillRef{
					{Ref: "skill1", Path: "skills/skill1/", External: true},
				},
			},
			wantIssues: 1, // external skills should not have path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := tt.manifest.Validate()
			if len(issues) != tt.wantIssues {
				t.Errorf("Validate() returned %d issues, want %d\nIssues: %v",
					len(issues), tt.wantIssues, issues)
			}
		})
	}
}

func TestIsValidForm(t *testing.T) {
	validForms := []RiteForm{FormSimple, FormPractitioner, FormProcedural, FormFull}
	for _, form := range validForms {
		if !IsValidForm(form) {
			t.Errorf("IsValidForm(%q) = false, want true", form)
		}
	}

	invalidForms := []RiteForm{"invalid", "SIMPLE", ""}
	for _, form := range invalidForms {
		if IsValidForm(form) {
			t.Errorf("IsValidForm(%q) = true, want false", form)
		}
	}
}

func TestIsKebabCase(t *testing.T) {
	valid := []string{
		"simple",
		"kebab-case",
		"multi-word-name",
		"with123numbers",
		"a",
		"a-b",
	}
	for _, s := range valid {
		if !isKebabCase(s) {
			t.Errorf("isKebabCase(%q) = false, want true", s)
		}
	}

	invalid := []string{
		"",
		"-starts-with-hyphen",
		"ends-with-hyphen-",
		"has--double-hyphen",
		"Has Spaces",
		"UPPERCASE",
		"camelCase",
		"under_score",
	}
	for _, s := range invalid {
		if isKebabCase(s) {
			t.Errorf("isKebabCase(%q) = true, want false", s)
		}
	}
}

func TestRiteManifest_AgentNames(t *testing.T) {
	m := &RiteManifest{
		Agents: []AgentRef{
			{Name: "agent1"},
			{Name: "agent2"},
			{Name: "agent3"},
		},
	}

	names := m.AgentNames()
	if len(names) != 3 {
		t.Errorf("len(AgentNames()) = %d, want 3", len(names))
	}
	expected := []string{"agent1", "agent2", "agent3"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("AgentNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestRiteManifest_SkillRefs(t *testing.T) {
	m := &RiteManifest{
		Skills: []SkillRef{
			{Ref: "skill1"},
			{Ref: "skill2"},
		},
	}

	refs := m.SkillRefs()
	if len(refs) != 2 {
		t.Errorf("len(SkillRefs()) = %d, want 2", len(refs))
	}
	expected := []string{"skill1", "skill2"}
	for i, ref := range refs {
		if ref != expected[i] {
			t.Errorf("SkillRefs()[%d] = %q, want %q", i, ref, expected[i])
		}
	}
}

func TestRiteManifest_GetAgent(t *testing.T) {
	m := &RiteManifest{
		Agents: []AgentRef{
			{Name: "agent1", Role: "Role 1"},
			{Name: "agent2", Role: "Role 2"},
		},
	}

	agent := m.GetAgent("agent1")
	if agent == nil {
		t.Fatal("GetAgent(\"agent1\") returned nil")
	}
	if agent.Role != "Role 1" {
		t.Errorf("GetAgent(\"agent1\").Role = %q, want %q", agent.Role, "Role 1")
	}

	missing := m.GetAgent("nonexistent")
	if missing != nil {
		t.Error("GetAgent(\"nonexistent\") should return nil")
	}
}

func TestRiteManifest_GetEstimatedTokens(t *testing.T) {
	// With explicit estimate
	m1 := &RiteManifest{
		Budget: &BudgetInfo{
			EstimatedTokens: 10000,
			AgentsCost:      5000,
			SkillsCost:      3000,
			WorkflowCost:    2000,
		},
	}
	if got := m1.GetEstimatedTokens(); got != 10000 {
		t.Errorf("GetEstimatedTokens() with explicit = %d, want 10000", got)
	}

	// Without explicit estimate (should sum components)
	m2 := &RiteManifest{
		Budget: &BudgetInfo{
			AgentsCost:   5000,
			SkillsCost:   3000,
			WorkflowCost: 2000,
		},
	}
	if got := m2.GetEstimatedTokens(); got != 10000 {
		t.Errorf("GetEstimatedTokens() from components = %d, want 10000", got)
	}

	// No budget info
	m3 := &RiteManifest{}
	if got := m3.GetEstimatedTokens(); got != 0 {
		t.Errorf("GetEstimatedTokens() with no budget = %d, want 0", got)
	}
}

func TestRiteManifest_HasMethods(t *testing.T) {
	m := &RiteManifest{
		Agents: []AgentRef{{Name: "a"}},
		Skills: []SkillRef{{Ref: "s"}},
		Workflow: &WorkflowConfig{
			Type: "sequential",
		},
	}

	if !m.HasAgents() {
		t.Error("HasAgents() = false, want true")
	}
	if !m.HasSkills() {
		t.Error("HasSkills() = false, want true")
	}
	if !m.HasWorkflow() {
		t.Error("HasWorkflow() = false, want true")
	}

	empty := &RiteManifest{}
	if empty.HasAgents() {
		t.Error("empty.HasAgents() = true, want false")
	}
	if empty.HasSkills() {
		t.Error("empty.HasSkills() = true, want false")
	}
	if empty.HasWorkflow() {
		t.Error("empty.HasWorkflow() = true, want false")
	}
}
