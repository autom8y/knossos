package agent

import (
	"fmt"
	"sort"

	"github.com/autom8y/knossos/internal/errors"
)

// SectionOwnership indicates who is responsible for a section's content.
type SectionOwnership string

const (
	// OwnerPlatform means the section has default content provided by the archetype.
	// Authors should not need to modify these sections.
	OwnerPlatform SectionOwnership = "platform"

	// OwnerAuthor means the section must be filled in by the agent author.
	// Scaffold generates TODO markers for these sections.
	OwnerAuthor SectionOwnership = "author"

	// OwnerDerived means the section content is generated from frontmatter or context.
	// Future `ari agent update` will regenerate these sections automatically.
	OwnerDerived SectionOwnership = "derived"
)

// SectionDef defines a section within an agent markdown file.
type SectionDef struct {
	// Name is the kebab-case identifier for the section (e.g., "core-responsibilities").
	Name string

	// Heading is the markdown heading (e.g., "## Core Responsibilities").
	Heading string

	// Ownership indicates who is responsible for the section content.
	Ownership SectionOwnership

	// TodoHint provides guidance for author-owned sections (used in TODO markers).
	TodoHint string
}

// Archetype defines a category of agent with default structure and configuration.
type Archetype struct {
	// Name is the archetype identifier (e.g., "orchestrator", "specialist", "reviewer").
	Name string

	// Description summarizes the archetype purpose.
	Description string

	// Defaults contains default frontmatter values for agents of this archetype.
	Defaults ArchetypeDefaults

	// Sections defines the ordered list of sections for this archetype.
	Sections []SectionDef
}

// ArchetypeDefaults contains default frontmatter values for an archetype.
type ArchetypeDefaults struct {
	// Model is the default Claude model (e.g., "opus", "sonnet").
	Model string

	// Tools is the default tool list.
	Tools []string

	// Color is the default color for the agent badge.
	Color string

	// MaxTurns is the default maximum conversation turns.
	MaxTurns int

	// DisallowedTools is the default list of tools the agent must not use.
	DisallowedTools []string
}

// archetypes is the registry of known archetypes.
var archetypes = map[string]*Archetype{
	"orchestrator": {
		Name:        "orchestrator",
		Description: "Consultative coordinator that analyzes context, routes work to specialists, and maintains decision consistency across phases. Orchestrators do not execute work directly.",
		Defaults: ArchetypeDefaults{
			Model:           "opus",
			Tools:           []string{"Read"},
			Color:           "purple",
			MaxTurns:        40,
			DisallowedTools: []string{"Bash", "Write", "Edit", "Glob", "Grep", "Task"},
		},
		Sections: []SectionDef{
			{Name: "consultation-role", Heading: "## Consultation Role", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "tool-access", Heading: "## Tool Access", Ownership: OwnerDerived,
				TodoHint: ""},
			{Name: "consultation-protocol", Heading: "## Consultation Protocol", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "position-in-workflow", Heading: "## Position in Workflow", Ownership: OwnerDerived,
				TodoHint: ""},
			{Name: "exousia", Heading: "## Exousia", Ownership: OwnerAuthor,
				TodoHint: "Define You Decide / You Escalate / You Do NOT Decide"},
			{Name: "phase-routing", Heading: "## Phase Routing", Ownership: OwnerAuthor,
				TodoHint: "Define which specialist handles which phase and routing conditions"},
			{Name: "behavioral-constraints", Heading: "## Behavioral Constraints", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "handling-failures", Heading: "## Handling Failures", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "the-acid-test", Heading: "## The Acid Test", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "cross-rite-protocol", Heading: "## Cross-Rite Protocol", Ownership: OwnerAuthor,
				TodoHint: "Define how cross-rite concerns are routed and resolved"},
			{Name: "skills-reference", Heading: "## Skills Reference", Ownership: OwnerDerived,
				TodoHint: ""},
			{Name: "anti-patterns", Heading: "## Anti-Patterns", Ownership: OwnerPlatform,
				TodoHint: ""},
		},
	},
	"specialist": {
		Name:        "specialist",
		Description: "Domain expert that executes focused work within a specific discipline. Specialists receive prompts from orchestrators, produce artifacts, and hand off to downstream agents.",
		Defaults: ArchetypeDefaults{
			Model:           "opus",
			Tools:           []string{"Bash", "Glob", "Grep", "Read", "Edit", "Write", "TodoWrite", "Skill"},
			Color:           "orange",
			MaxTurns:        150,
			DisallowedTools: nil,
		},
		Sections: []SectionDef{
			{Name: "core-responsibilities", Heading: "## Core Responsibilities", Ownership: OwnerAuthor,
				TodoHint: "Define what this agent is responsible for"},
			{Name: "position-in-workflow", Heading: "## Position in Workflow", Ownership: OwnerDerived,
				TodoHint: ""},
			{Name: "exousia", Heading: "## Exousia", Ownership: OwnerAuthor,
				TodoHint: "Define You Decide / You Escalate / You Do NOT Decide"},
			{Name: "tool-access", Heading: "## Tool Access", Ownership: OwnerDerived,
				TodoHint: ""},
			{Name: "what-you-produce", Heading: "## What You Produce", Ownership: OwnerAuthor,
				TodoHint: "Define the artifacts this agent produces with format and audience"},
			{Name: "quality-standards", Heading: "## Quality Standards", Ownership: OwnerAuthor,
				TodoHint: "Define quality criteria and verification requirements"},
			{Name: "handoff-criteria", Heading: "## Handoff Criteria", Ownership: OwnerAuthor,
				TodoHint: "Define the checklist that must be complete before handoff"},
			{Name: "behavioral-constraints", Heading: "## Behavioral Constraints", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "the-acid-test", Heading: "## The Acid Test", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "anti-patterns", Heading: "## Anti-Patterns", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "skills-reference", Heading: "## Skills Reference", Ownership: OwnerDerived,
				TodoHint: ""},
		},
	},
	"reviewer": {
		Name:        "reviewer",
		Description: "Quality gate that reviews work products against domain-specific criteria. Reviewers evaluate, classify findings by severity, and provide clear approve/reject decisions with actionable feedback.",
		Defaults: ArchetypeDefaults{
			Model:           "opus",
			Tools:           []string{"Bash", "Glob", "Grep", "Read", "Edit", "Write", "WebFetch", "WebSearch", "TodoWrite", "Skill"},
			Color:           "red",
			MaxTurns:        100,
			DisallowedTools: []string{"Task"},
		},
		Sections: []SectionDef{
			{Name: "core-purpose", Heading: "## Core Purpose", Ownership: OwnerAuthor,
				TodoHint: "Define what this reviewer catches and why it matters"},
			{Name: "position-in-workflow", Heading: "## Position in Workflow", Ownership: OwnerDerived,
				TodoHint: ""},
			{Name: "exousia", Heading: "## Exousia", Ownership: OwnerAuthor,
				TodoHint: "Define You Decide / You Escalate / You Do NOT Decide"},
			{Name: "quality-standards", Heading: "## Quality Standards", Ownership: OwnerAuthor,
				TodoHint: "Define review focus areas and criteria"},
			{Name: "severity-classification", Heading: "## Severity Classification", Ownership: OwnerAuthor,
				TodoHint: "Define severity levels with definitions and examples"},
			{Name: "what-you-produce", Heading: "## What You Produce", Ownership: OwnerAuthor,
				TodoHint: "Define the review artifacts and signoff format"},
			{Name: "behavioral-constraints", Heading: "## Behavioral Constraints", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "the-acid-test", Heading: "## The Acid Test", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "anti-patterns", Heading: "## Anti-Patterns", Ownership: OwnerPlatform,
				TodoHint: ""},
			{Name: "skills-reference", Heading: "## Skills Reference", Ownership: OwnerDerived,
				TodoHint: ""},
		},
	},
}

// GetArchetype returns the archetype with the given name, or an error if not found.
func GetArchetype(name string) (*Archetype, error) {
	a, ok := archetypes[name]
	if !ok {
		return nil, errors.NewWithDetails(errors.CodeUsageError,
			fmt.Sprintf("unknown archetype %q, must be one of: %s", name, listArchetypeNames()),
			map[string]interface{}{"archetype": name, "available": ListArchetypes()})
	}
	return a, nil
}

// ListArchetypes returns the sorted list of available archetype names.
func ListArchetypes() []string {
	names := make([]string, 0, len(archetypes))
	for name := range archetypes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// listArchetypeNames returns a comma-separated string of archetype names for error messages.
func listArchetypeNames() string {
	names := ListArchetypes()
	result := ""
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}

// AuthorSections returns only the author-owned sections for this archetype.
func (a *Archetype) AuthorSections() []SectionDef {
	var result []SectionDef
	for _, s := range a.Sections {
		if s.Ownership == OwnerAuthor {
			result = append(result, s)
		}
	}
	return result
}

// PlatformSections returns only the platform-owned sections for this archetype.
func (a *Archetype) PlatformSections() []SectionDef {
	var result []SectionDef
	for _, s := range a.Sections {
		if s.Ownership == OwnerPlatform {
			result = append(result, s)
		}
	}
	return result
}

// DerivedSections returns only the derived sections for this archetype.
func (a *Archetype) DerivedSections() []SectionDef {
	var result []SectionDef
	for _, s := range a.Sections {
		if s.Ownership == OwnerDerived {
			result = append(result, s)
		}
	}
	return result
}
