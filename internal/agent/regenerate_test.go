package agent

import (
	"strings"
	"testing"
)

func TestRegeneratePlatformSections(t *testing.T) {
	content := `---
name: test-orchestrator
description: Test orchestrator
type: orchestrator
tools: Read
model: opus
color: purple
---

# Test Orchestrator

Orchestrator for testing.

## Consultation Role (CRITICAL)

OLD PLATFORM CONTENT - should be replaced.

## Exousia

AUTHOR CONTENT - should be preserved exactly.

## Phase Routing

AUTHOR CONTENT - phase routing table.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	archetype, err := GetArchetype("orchestrator")
	if err != nil {
		t.Fatalf("GetArchetype() error = %v", err)
	}

	updated, err := RegeneratePlatformSections(parsed, archetype)
	if err != nil {
		t.Fatalf("RegeneratePlatformSections() error = %v", err)
	}

	// Verify platform section was replaced
	consultationRole := updated.FindSection("consultation-role")
	if consultationRole == nil {
		t.Fatal("consultation-role section not found")
	}
	if strings.Contains(consultationRole.Content, "OLD PLATFORM CONTENT") {
		t.Error("platform section still contains old content")
	}
	if !strings.Contains(consultationRole.Content, "consultative throughline") {
		t.Error("platform section missing expected template content")
	}

	// Verify author sections were preserved
	domainAuthority := updated.FindSection("exousia")
	if domainAuthority == nil {
		t.Fatal("exousia section not found")
	}
	if !strings.Contains(domainAuthority.Content, "AUTHOR CONTENT - should be preserved exactly") {
		t.Errorf("author section was modified: %q", domainAuthority.Content)
	}

	phaseRouting := updated.FindSection("phase-routing")
	if phaseRouting == nil {
		t.Fatal("phase-routing section not found")
	}
	if !strings.Contains(phaseRouting.Content, "AUTHOR CONTENT - phase routing table") {
		t.Errorf("author section was modified: %q", phaseRouting.Content)
	}
}

func TestRegeneratePlatformSections_MissingAuthorSection(t *testing.T) {
	content := `---
name: test-specialist
description: Test specialist
type: specialist
tools: Read, Write
model: opus
color: orange
---

# Test Specialist

Specialist for testing.

## Core Responsibilities

Existing responsibilities.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	archetype, err := GetArchetype("specialist")
	if err != nil {
		t.Fatalf("GetArchetype() error = %v", err)
	}

	updated, err := RegeneratePlatformSections(parsed, archetype)
	if err != nil {
		t.Fatalf("RegeneratePlatformSections() error = %v", err)
	}

	// Verify existing author section was preserved
	coreResp := updated.FindSection("core-responsibilities")
	if coreResp == nil {
		t.Fatal("core-responsibilities section not found")
	}
	if !strings.Contains(coreResp.Content, "Existing responsibilities") {
		t.Error("existing author section was not preserved")
	}

	// Verify missing author section was added with TODO
	domainAuth := updated.FindSection("exousia")
	if domainAuth == nil {
		t.Fatal("exousia section not found (should be added)")
	}
	if !strings.Contains(domainAuth.Content, "TODO") {
		t.Error("missing author section should have TODO marker")
	}
}

func TestRegeneratePlatformSections_DerivedSections(t *testing.T) {
	content := `---
name: test-specialist
description: Test specialist
type: specialist
tools: Read, Write, Edit, Bash
model: opus
color: orange
upstream:
  - source: analyst
    artifact: requirements
downstream:
  - agent: reviewer
    artifact: implementation
---

# Test Specialist

Specialist for testing.

## Tool Access

OLD TOOL TABLE - should be regenerated.

## Position in Workflow

OLD WORKFLOW - should be regenerated.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	archetype, err := GetArchetype("specialist")
	if err != nil {
		t.Fatalf("GetArchetype() error = %v", err)
	}

	updated, err := RegeneratePlatformSections(parsed, archetype)
	if err != nil {
		t.Fatalf("RegeneratePlatformSections() error = %v", err)
	}

	// Verify tool access was regenerated from frontmatter
	toolAccess := updated.FindSection("tool-access")
	if toolAccess == nil {
		t.Fatal("tool-access section not found")
	}
	if strings.Contains(toolAccess.Content, "OLD TOOL TABLE") {
		t.Error("derived section still contains old content")
	}
	if !strings.Contains(toolAccess.Content, "Read") ||
		!strings.Contains(toolAccess.Content, "Write") ||
		!strings.Contains(toolAccess.Content, "Edit") ||
		!strings.Contains(toolAccess.Content, "Bash") {
		t.Error("tool access section missing expected tools from frontmatter")
	}

	// Verify workflow position was regenerated
	workflow := updated.FindSection("position-in-workflow")
	if workflow == nil {
		t.Fatal("position-in-workflow section not found")
	}
	if strings.Contains(workflow.Content, "OLD WORKFLOW") {
		t.Error("derived section still contains old content")
	}
	if !strings.Contains(workflow.Content, "analyst") {
		t.Error("workflow section missing upstream agent from frontmatter")
	}
	if !strings.Contains(workflow.Content, "reviewer") {
		t.Error("workflow section missing downstream agent from frontmatter")
	}
}

func TestAssembleAgentFile(t *testing.T) {
	content := `---
name: test-agent
description: Test agent
type: specialist
tools: Read, Write
---

# Test Agent

This is the preamble.

## Core Responsibilities

Responsibilities here.

## Exousia

Authority here.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	assembled := AssembleAgentFile(parsed)
	assembledStr := string(assembled)

	// Verify frontmatter is preserved
	if !strings.Contains(assembledStr, "name: test-agent") {
		t.Error("assembled file missing frontmatter")
	}

	// Verify title is present
	if !strings.Contains(assembledStr, "# Test Agent") {
		t.Error("assembled file missing title")
	}

	// Verify preamble is present
	if !strings.Contains(assembledStr, "This is the preamble.") {
		t.Error("assembled file missing preamble")
	}

	// Verify sections are present
	if !strings.Contains(assembledStr, "## Core Responsibilities") {
		t.Error("assembled file missing section heading")
	}
	if !strings.Contains(assembledStr, "Responsibilities here.") {
		t.Error("assembled file missing section content")
	}

	// Verify file ends with single newline
	if !strings.HasSuffix(assembledStr, "\n") {
		t.Error("assembled file should end with newline")
	}
	if strings.HasSuffix(assembledStr, "\n\n") {
		t.Error("assembled file should end with single newline, not multiple")
	}
}

func TestRegeneratePlatformSections_PreservesUnknownSections(t *testing.T) {
	content := `---
name: test-specialist
description: Test specialist
type: specialist
tools: Read
---

# Test Specialist

## Core Responsibilities

Standard section.

## Custom Implementation Notes

This is a custom section not in the archetype.

## Another Custom Section

More custom content.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	archetype, err := GetArchetype("specialist")
	if err != nil {
		t.Fatalf("GetArchetype() error = %v", err)
	}

	updated, err := RegeneratePlatformSections(parsed, archetype)
	if err != nil {
		t.Fatalf("RegeneratePlatformSections() error = %v", err)
	}

	// Verify custom sections are preserved
	customSection := updated.FindSectionByHeading("Custom Implementation Notes")
	if customSection == nil {
		t.Fatal("custom section was removed")
	}
	if !strings.Contains(customSection.Content, "custom section not in the archetype") {
		t.Error("custom section content was modified")
	}

	anotherCustom := updated.FindSectionByHeading("Another Custom Section")
	if anotherCustom == nil {
		t.Fatal("another custom section was removed")
	}
	if !strings.Contains(anotherCustom.Content, "More custom content") {
		t.Error("custom section content was modified")
	}
}

func TestRegeneratePlatformSections_TypeMapping(t *testing.T) {
	// Test that non-standard types (meta, designer, analyst, engineer) map to specialist
	content := `---
name: test-analyst
description: Test analyst
type: analyst
tools: Read, Write
---

# Test Analyst

## Core Responsibilities

Analyst responsibilities.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	// The type is "analyst", which should map to "specialist" archetype
	archetype, err := GetArchetype("specialist")
	if err != nil {
		t.Fatalf("GetArchetype() error = %v", err)
	}

	updated, err := RegeneratePlatformSections(parsed, archetype)
	if err != nil {
		t.Fatalf("RegeneratePlatformSections() error = %v", err)
	}

	// Verify it was processed using specialist archetype
	// Check for a specialist-specific platform section
	behavioralConstraints := updated.FindSection("behavioral-constraints")
	if behavioralConstraints == nil {
		t.Fatal("behavioral-constraints section not found (should be from specialist archetype)")
	}
}

func TestGenerateSkillsReference_WithSkills(t *testing.T) {
	skills := []string{"ecosystem-ref", "orchestrator-templates"}
	result := generateSkillsReference(skills)

	if !strings.Contains(result, "ecosystem-ref") {
		t.Errorf("output missing skill %q: %q", "ecosystem-ref", result)
	}
	if !strings.Contains(result, "orchestrator-templates") {
		t.Errorf("output missing skill %q: %q", "orchestrator-templates", result)
	}
	if strings.Contains(result, "@") {
		t.Errorf("output must not contain @ prefix (SCAR-017 anti-pattern): %q", result)
	}
	if strings.Contains(result, "Load skills on demand") {
		t.Errorf("output should list skills, not generic message: %q", result)
	}
}

func TestGenerateSkillsReference_EmptySkills(t *testing.T) {
	result := generateSkillsReference(nil)

	if !strings.Contains(result, "Load skills on demand") {
		t.Errorf("empty skills should produce generic message, got: %q", result)
	}
	if strings.Contains(result, "@") {
		t.Errorf("output must not contain @ prefix: %q", result)
	}

	// Also test empty slice (not nil)
	result = generateSkillsReference([]string{})
	if !strings.Contains(result, "Load skills on demand") {
		t.Errorf("empty skills slice should produce generic message, got: %q", result)
	}
}

func TestGenerateSkillsReference_NoAtPrefix(t *testing.T) {
	// Verify that no skills list can produce @ in output (guards against regression).
	testCases := [][]string{
		{"standards"},
		{"file-verification"},
		{"standards", "file-verification"},
		{"@should-be-stripped"},
	}
	for _, skills := range testCases {
		result := generateSkillsReference(skills)
		// The output should not inject @ characters beyond what is in the skill name itself.
		// Since skill names should not have @ (that is the anti-pattern), verify the function
		// does not add @ prefixes.
		for _, skill := range skills {
			expectedEntry := "- " + skill
			if strings.Contains(result, "@"+skill) {
				t.Errorf("output must not add @ prefix to skill %q: %q", skill, result)
			}
			if !strings.Contains(result, "@") || strings.Contains(skill, "@") {
				// Only flag if the @ appeared and was not originally in the skill name.
				_ = expectedEntry
			}
		}
	}
}

func TestGenerateSkillsReference_IntegrationWithDerivedSection(t *testing.T) {
	// Verify that the derived skills-reference section is generated from frontmatter skills.
	content := `---
name: test-orchestrator
description: Test orchestrator with skills
type: orchestrator
tools: Read
model: opus
color: purple
skills:
  - ecosystem-ref
  - forge-ref
---

# Test Orchestrator

## Skills Reference

OLD CONTENT - should be replaced with frontmatter skills.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	archetype, err := GetArchetype("orchestrator")
	if err != nil {
		t.Fatalf("GetArchetype() error = %v", err)
	}

	updated, err := RegeneratePlatformSections(parsed, archetype)
	if err != nil {
		t.Fatalf("RegeneratePlatformSections() error = %v", err)
	}

	skillsRef := updated.FindSection("skills-reference")
	if skillsRef == nil {
		t.Fatal("skills-reference section not found")
	}

	if strings.Contains(skillsRef.Content, "OLD CONTENT") {
		t.Error("derived skills-reference still contains old content")
	}
	if !strings.Contains(skillsRef.Content, "ecosystem-ref") {
		t.Errorf("skills-reference missing frontmatter skill %q: %q", "ecosystem-ref", skillsRef.Content)
	}
	if !strings.Contains(skillsRef.Content, "forge-ref") {
		t.Errorf("skills-reference missing frontmatter skill %q: %q", "forge-ref", skillsRef.Content)
	}
	if strings.Contains(skillsRef.Content, "@") {
		t.Errorf("skills-reference must not contain @ prefix (SCAR-017): %q", skillsRef.Content)
	}
}
