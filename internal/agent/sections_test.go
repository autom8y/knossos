package agent

import (
	"strings"
	"testing"
)

func TestParseAgentSections(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantTitle      string
		wantPreamble   string
		wantNumSections int
		wantSections   map[string]string // section heading -> content excerpt
	}{
		{
			name: "basic agent with all sections",
			content: `---
name: test-agent
description: Test agent
type: specialist
tools: Read, Write
---

# Test Agent

This is the preamble paragraph.

## Core Responsibilities

Agent responsibilities here.

## Position in Workflow

Workflow info here.

## Exousia

Authority info here.
`,
			wantTitle:       "Test Agent",
			wantPreamble:    "This is the preamble paragraph.",
			wantNumSections: 3,
			wantSections: map[string]string{
				"Core Responsibilities":  "Agent responsibilities here.",
				"Position in Workflow":   "Workflow info here.",
				"Exousia":       "Authority info here.",
			},
		},
		{
			name: "agent with no preamble",
			content: `---
name: test-agent
description: Test agent
type: specialist
tools: Read
---

# Test Agent

## Core Responsibilities

Responsibilities here.
`,
			wantTitle:       "Test Agent",
			wantPreamble:    "",
			wantNumSections: 1,
			wantSections: map[string]string{
				"Core Responsibilities": "Responsibilities here.",
			},
		},
		{
			name: "agent with unknown sections",
			content: `---
name: test-agent
description: Test agent
type: specialist
tools: Read
---

# Test Agent

## Core Responsibilities

Responsibilities here.

## Custom Section

Custom content here.

## Another Custom Section

More custom content.
`,
			wantTitle:       "Test Agent",
			wantNumSections: 3,
			wantSections: map[string]string{
				"Core Responsibilities": "Responsibilities here.",
				"Custom Section":        "Custom content here.",
				"Another Custom Section": "More custom content.",
			},
		},
		{
			name: "agent with multiline section content",
			content: `---
name: test-agent
description: Test agent
type: orchestrator
tools: Read
---

# Test Agent

## Consultation Role (CRITICAL)

Line 1
Line 2
Line 3

More content here.

## Tool Access

Tools listed here.
`,
			wantTitle:       "Test Agent",
			wantNumSections: 2,
			wantSections: map[string]string{
				"Consultation Role (CRITICAL)": "Line 1\nLine 2\nLine 3\n\nMore content here.",
				"Tool Access":                  "Tools listed here.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseAgentSections([]byte(tt.content))
			if err != nil {
				t.Fatalf("ParseAgentSections() error = %v", err)
			}

			if parsed.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", parsed.Title, tt.wantTitle)
			}

			if strings.TrimSpace(parsed.Preamble) != tt.wantPreamble {
				t.Errorf("Preamble = %q, want %q", parsed.Preamble, tt.wantPreamble)
			}

			if len(parsed.Sections) != tt.wantNumSections {
				t.Errorf("got %d sections, want %d", len(parsed.Sections), tt.wantNumSections)
			}

			for heading, wantContent := range tt.wantSections {
				section := parsed.FindSectionByHeading(heading)
				if section == nil {
					t.Errorf("section %q not found", heading)
					continue
				}
				if !strings.Contains(section.Content, wantContent) {
					t.Errorf("section %q content = %q, want to contain %q",
						heading, section.Content, wantContent)
				}
			}
		})
	}
}

func TestParseAgentSections_Ownership(t *testing.T) {
	content := `---
name: test-orchestrator
description: Test orchestrator
type: orchestrator
tools: Read
---

# Test Orchestrator

Preamble here.

## Consultation Role (CRITICAL)

Platform content.

## Exousia

Author content.

## Custom Section

Unknown section.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	tests := []struct {
		heading       string
		wantOwnership SectionOwnership
	}{
		{"Consultation Role (CRITICAL)", OwnerPlatform},
		{"Exousia", OwnerAuthor},
		{"Custom Section", OwnerAuthor}, // Unknown sections default to author
	}

	for _, tt := range tests {
		t.Run(tt.heading, func(t *testing.T) {
			section := parsed.FindSectionByHeading(tt.heading)
			if section == nil {
				t.Fatalf("section %q not found", tt.heading)
			}
			if section.Ownership != tt.wantOwnership {
				t.Errorf("section %q ownership = %v, want %v",
					tt.heading, section.Ownership, tt.wantOwnership)
			}
		})
	}
}

func TestParseAgentSections_NoType(t *testing.T) {
	content := `---
name: legacy-agent
description: Legacy agent without type field
tools: Read
---

# Legacy Agent

## Some Section

Content here.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	// Sections should not be mapped to archetype
	section := parsed.FindSectionByHeading("Some Section")
	if section == nil {
		t.Fatal("section not found")
	}

	// Without type, sections remain unmapped (Name="", Ownership=OwnerAuthor)
	if section.Name != "" {
		t.Errorf("section Name = %q, want empty (unmapped)", section.Name)
	}
	if section.Ownership != OwnerAuthor {
		t.Errorf("section Ownership = %v, want OwnerAuthor", section.Ownership)
	}
}

func TestFindSection(t *testing.T) {
	content := `---
name: test-agent
description: Test
type: specialist
tools: Read
---

# Test Agent

## Core Responsibilities

Responsibilities.

## Position in Workflow

Workflow.
`

	parsed, err := ParseAgentSections([]byte(content))
	if err != nil {
		t.Fatalf("ParseAgentSections() error = %v", err)
	}

	// Test FindSection by name
	section := parsed.FindSection("core-responsibilities")
	if section == nil {
		t.Fatal("FindSection(core-responsibilities) returned nil")
	}
	if section.Heading != "Core Responsibilities" {
		t.Errorf("section Heading = %q, want %q", section.Heading, "Core Responsibilities")
	}

	// Test FindSectionByHeading
	section = parsed.FindSectionByHeading("Position in Workflow")
	if section == nil {
		t.Fatal("FindSectionByHeading returned nil")
	}
	if section.Name != "position-in-workflow" {
		t.Errorf("section Name = %q, want %q", section.Name, "position-in-workflow")
	}

	// Test not found
	section = parsed.FindSection("nonexistent")
	if section != nil {
		t.Errorf("FindSection(nonexistent) returned %v, want nil", section)
	}
}

// TestMapSectionsToArchetype_PrefixMatching verifies that heading variants are matched to their
// archetype section via prefix matching when exact matching fails.
func TestMapSectionsToArchetype_PrefixMatching(t *testing.T) {
	tests := []struct {
		name          string
		heading       string
		wantName      string
		wantOwnership SectionOwnership
	}{
		{
			name:          "exact match still works (regression guard)",
			heading:       "Behavioral Constraints",
			wantName:      "behavioral-constraints",
			wantOwnership: OwnerPlatform,
		},
		{
			name:          "behavioral constraints with DO NOT suffix",
			heading:       "Behavioral Constraints (DO NOT)",
			wantName:      "behavioral-constraints",
			wantOwnership: OwnerPlatform,
		},
		{
			name:          "anti-patterns exact match (regression guard)",
			heading:       "Anti-Patterns",
			wantName:      "anti-patterns",
			wantOwnership: OwnerPlatform,
		},
		{
			name:          "anti-patterns with to Avoid suffix",
			heading:       "Anti-Patterns to Avoid",
			wantName:      "anti-patterns",
			wantOwnership: OwnerPlatform,
		},
		{
			name:          "anti-patterns with colon suffix",
			heading:       "Anti-Patterns: Key Warnings",
			wantName:      "anti-patterns",
			wantOwnership: OwnerPlatform,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := "---\nname: test\ndescription: test\ntype: orchestrator\ntools: Read\n---\n\n# Test\n\n## " + tt.heading + "\n\nContent.\n"
			parsed, err := ParseAgentSections([]byte(content))
			if err != nil {
				t.Fatalf("ParseAgentSections() error = %v", err)
			}

			section := parsed.FindSectionByHeading(tt.heading)
			if section == nil {
				t.Fatalf("section %q not found", tt.heading)
			}

			if section.Name != tt.wantName {
				t.Errorf("section %q Name = %q, want %q", tt.heading, section.Name, tt.wantName)
			}
			if section.Ownership != tt.wantOwnership {
				t.Errorf("section %q Ownership = %v, want %v", tt.heading, section.Ownership, tt.wantOwnership)
			}
		})
	}
}

// TestMapSectionsToArchetype_NoPrefixFalseMatch verifies that the prefix matching is conservative
// and does not produce false positives for headings that merely share a prefix.
func TestMapSectionsToArchetype_NoPrefixFalseMatch(t *testing.T) {
	tests := []struct {
		name    string
		heading string
	}{
		{
			// "Tool Access" must not match a heading "Tool Accessibility".
			name:    "tool accessibility not matched as tool access",
			heading: "Tool Accessibility",
		},
		{
			// "Anti-Patterns" must not match "Anti-Pattern Counter-Examples" with no separator
			// — actually "anti-pattern" is a prefix of "anti-patterns", so test something
			// that truly has no separator: a concatenated variant.
			name:    "unrelated heading stays unmapped",
			heading: "Completely Unrelated Heading",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := "---\nname: test\ndescription: test\ntype: specialist\ntools: Read\n---\n\n# Test\n\n## " + tt.heading + "\n\nContent.\n"
			parsed, err := ParseAgentSections([]byte(content))
			if err != nil {
				t.Fatalf("ParseAgentSections() error = %v", err)
			}

			section := parsed.FindSectionByHeading(tt.heading)
			if section == nil {
				t.Fatalf("section %q not found", tt.heading)
			}

			// Should remain unmapped (no archetype section matches).
			if section.Name != "" {
				t.Errorf("section %q should be unmapped but got Name=%q", tt.heading, section.Name)
			}
			if section.Ownership != OwnerAuthor {
				t.Errorf("section %q should be OwnerAuthor but got %v", tt.heading, section.Ownership)
			}
		})
	}
}
