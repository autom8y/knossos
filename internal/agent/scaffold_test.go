package agent

import (
	"strings"
	"testing"
)

func TestScaffoldAgent_AllArchetypes(t *testing.T) {
	archetypeNames := ListArchetypes()
	if len(archetypeNames) == 0 {
		t.Fatal("expected at least one archetype")
	}

	for _, name := range archetypeNames {
		t.Run(name, func(t *testing.T) {
			arch, err := GetArchetype(name)
			if err != nil {
				t.Fatalf("GetArchetype(%q) failed: %v", name, err)
			}

			content, err := ScaffoldAgent(arch, "test-agent", "rnd", "A test agent for validation")
			if err != nil {
				t.Fatalf("ScaffoldAgent(%q) failed: %v", name, err)
			}

			if len(content) == 0 {
				t.Fatal("scaffold produced empty content")
			}

			// Parse and validate the output
			fm, err := ParseAgentFrontmatter(content)
			if err != nil {
				t.Fatalf("generated agent failed to parse: %v", err)
			}

			// Verify frontmatter fields
			if fm.Name != "test-agent" {
				t.Errorf("name = %q, want %q", fm.Name, "test-agent")
			}
			if fm.Type != name {
				t.Errorf("type = %q, want %q", fm.Type, name)
			}
			if fm.Model != arch.Defaults.Model {
				t.Errorf("model = %q, want %q", fm.Model, arch.Defaults.Model)
			}

			// Verify tools match archetype defaults
			if len(fm.Tools) != len(arch.Defaults.Tools) {
				t.Errorf("tools count = %d, want %d; got %v", len(fm.Tools), len(arch.Defaults.Tools), fm.Tools)
			}
			for i, tool := range fm.Tools {
				if i < len(arch.Defaults.Tools) && tool != arch.Defaults.Tools[i] {
					t.Errorf("tools[%d] = %q, want %q", i, tool, arch.Defaults.Tools[i])
				}
			}
		})
	}
}

func TestScaffoldAgent_SectionsPresent(t *testing.T) {
	archetypeNames := ListArchetypes()

	for _, name := range archetypeNames {
		t.Run(name, func(t *testing.T) {
			arch, err := GetArchetype(name)
			if err != nil {
				t.Fatalf("GetArchetype(%q) failed: %v", name, err)
			}

			content, err := ScaffoldAgent(arch, "test-agent", "rnd", "A test agent for validation")
			if err != nil {
				t.Fatalf("ScaffoldAgent(%q) failed: %v", name, err)
			}

			text := string(content)

			// Verify all section headings are present
			for _, section := range arch.Sections {
				if !strings.Contains(text, section.Heading) {
					t.Errorf("missing section heading %q in %s scaffold", section.Heading, name)
				}
			}
		})
	}
}

func TestScaffoldAgent_AuthorSectionsHaveTODO(t *testing.T) {
	archetypeNames := ListArchetypes()

	for _, name := range archetypeNames {
		t.Run(name, func(t *testing.T) {
			arch, err := GetArchetype(name)
			if err != nil {
				t.Fatalf("GetArchetype(%q) failed: %v", name, err)
			}

			content, err := ScaffoldAgent(arch, "test-agent", "rnd", "A test agent for validation")
			if err != nil {
				t.Fatalf("ScaffoldAgent(%q) failed: %v", name, err)
			}

			text := string(content)

			// Every author section should have at least one TODO marker
			authorSections := arch.AuthorSections()
			if len(authorSections) == 0 {
				t.Errorf("archetype %q has no author sections", name)
			}

			for _, section := range authorSections {
				// Find the section content by looking between this heading and the next ##
				idx := strings.Index(text, section.Heading)
				if idx == -1 {
					t.Errorf("author section %q not found in scaffold", section.Heading)
					continue
				}

				// Extract content until the next ## heading or end of file
				afterHeading := text[idx+len(section.Heading):]
				nextHeading := strings.Index(afterHeading, "\n## ")
				var sectionContent string
				if nextHeading >= 0 {
					sectionContent = afterHeading[:nextHeading]
				} else {
					sectionContent = afterHeading
				}

				if !strings.Contains(sectionContent, "TODO:") {
					t.Errorf("author section %q in %s scaffold missing TODO marker", section.Heading, name)
				}
			}
		})
	}
}

func TestScaffoldAgent_PlatformSectionsHaveContent(t *testing.T) {
	archetypeNames := ListArchetypes()

	for _, name := range archetypeNames {
		t.Run(name, func(t *testing.T) {
			arch, err := GetArchetype(name)
			if err != nil {
				t.Fatalf("GetArchetype(%q) failed: %v", name, err)
			}

			content, err := ScaffoldAgent(arch, "test-agent", "rnd", "A test agent for validation")
			if err != nil {
				t.Fatalf("ScaffoldAgent(%q) failed: %v", name, err)
			}

			text := string(content)

			// Platform sections should have real content (more than just the heading)
			for _, section := range arch.PlatformSections() {
				idx := strings.Index(text, section.Heading)
				if idx == -1 {
					t.Errorf("platform section %q not found in scaffold", section.Heading)
					continue
				}

				afterHeading := text[idx+len(section.Heading):]
				nextHeading := strings.Index(afterHeading, "\n## ")
				var sectionContent string
				if nextHeading >= 0 {
					sectionContent = afterHeading[:nextHeading]
				} else {
					sectionContent = afterHeading
				}

				// Platform sections should have meaningful content (at least 50 chars)
				trimmed := strings.TrimSpace(sectionContent)
				if len(trimmed) < 50 {
					t.Errorf("platform section %q in %s scaffold has insufficient content (%d chars): %q",
						section.Heading, name, len(trimmed), trimmed)
				}
			}
		})
	}
}

func TestScaffoldAgent_WarnModeValidation(t *testing.T) {
	// Verify that generated agents pass full WARN-mode validation (including schema)
	av := newTestValidator(t)

	archetypeNames := ListArchetypes()
	for _, name := range archetypeNames {
		t.Run(name, func(t *testing.T) {
			arch, err := GetArchetype(name)
			if err != nil {
				t.Fatalf("GetArchetype(%q) failed: %v", name, err)
			}

			content, err := ScaffoldAgent(arch, "test-agent", "rnd", "A test agent for validation")
			if err != nil {
				t.Fatalf("ScaffoldAgent(%q) failed: %v", name, err)
			}

			result, err := av.ValidateAgentFrontmatter(content, ValidationModeWarn)
			if err != nil {
				t.Fatalf("validation returned error: %v", err)
			}

			if !result.Valid {
				for _, issue := range result.Issues {
					t.Errorf("validation issue: [%s] %s", issue.Field, issue.Message)
				}
			}

			for _, w := range result.Warnings {
				t.Logf("validation warning: %s", w)
			}
		})
	}
}

func TestScaffoldAgent_DefaultDescription(t *testing.T) {
	arch, err := GetArchetype("specialist")
	if err != nil {
		t.Fatalf("GetArchetype failed: %v", err)
	}

	// Empty description should get a default
	content, err := ScaffoldAgent(arch, "my-agent", "rnd", "")
	if err != nil {
		t.Fatalf("ScaffoldAgent failed: %v", err)
	}

	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if fm.Description == "" {
		t.Error("expected non-empty description for default")
	}
	if !strings.Contains(fm.Description, "Specialist") || !strings.Contains(fm.Description, "rnd") {
		t.Errorf("default description = %q, expected to mention archetype and rite", fm.Description)
	}
}

func TestScaffoldAgent_NilArchetype(t *testing.T) {
	_, err := ScaffoldAgent(nil, "test", "rnd", "desc")
	if err == nil {
		t.Fatal("expected error for nil archetype")
	}
}

func TestScaffoldAgent_EmptyName(t *testing.T) {
	arch, _ := GetArchetype("specialist")
	_, err := ScaffoldAgent(arch, "", "rnd", "desc")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestScaffoldAgent_EmptyRite(t *testing.T) {
	arch, _ := GetArchetype("specialist")
	_, err := ScaffoldAgent(arch, "test", "", "desc")
	if err == nil {
		t.Fatal("expected error for empty rite")
	}
}

func TestGetArchetype_Valid(t *testing.T) {
	for _, name := range ListArchetypes() {
		t.Run(name, func(t *testing.T) {
			arch, err := GetArchetype(name)
			if err != nil {
				t.Fatalf("GetArchetype(%q) failed: %v", name, err)
			}
			if arch.Name != name {
				t.Errorf("archetype.Name = %q, want %q", arch.Name, name)
			}
			if arch.Description == "" {
				t.Error("archetype.Description is empty")
			}
			if len(arch.Sections) == 0 {
				t.Error("archetype has no sections")
			}
			if arch.Defaults.Model == "" {
				t.Error("archetype has no default model")
			}
			if len(arch.Defaults.Tools) == 0 {
				t.Error("archetype has no default tools")
			}
		})
	}
}

func TestGetArchetype_Invalid(t *testing.T) {
	_, err := GetArchetype("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent archetype")
	}
}

func TestListArchetypes(t *testing.T) {
	names := ListArchetypes()
	if len(names) != 3 {
		t.Errorf("expected 3 archetypes, got %d: %v", len(names), names)
	}

	// Should be sorted
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("archetypes not sorted: %v", names)
			break
		}
	}

	// Should contain all three
	expected := map[string]bool{"orchestrator": false, "reviewer": false, "specialist": false}
	for _, name := range names {
		expected[name] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("missing archetype %q", name)
		}
	}
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"technology-scout", "Technology Scout"},
		{"orchestrator", "Orchestrator"},
		{"security-reviewer", "Security Reviewer"},
		{"my-cool-agent", "My Cool Agent"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toTitleCase(tt.input)
			if result != tt.expected {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
