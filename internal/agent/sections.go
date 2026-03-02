package agent

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/autom8y/knossos/internal/errors"
)

// ParsedSection represents a section extracted from an agent markdown file.
type ParsedSection struct {
	// Name is the section name from archetype (e.g., "core-responsibilities").
	// Empty string for sections not in the archetype definition.
	Name string

	// Heading is the actual ## heading text from the file.
	Heading string

	// Content is everything between this heading and the next ## heading (or EOF).
	Content string

	// Ownership indicates who owns this section (from archetype, or OwnerAuthor for unknown).
	Ownership SectionOwnership
}

// ParsedAgent represents a parsed agent markdown file with separated components.
type ParsedAgent struct {
	// Frontmatter is the parsed frontmatter structure.
	Frontmatter *AgentFrontmatter

	// RawFrontmatter is the original YAML frontmatter text (between --- delimiters).
	RawFrontmatter string

	// Title is the # Title line (h1).
	Title string

	// Preamble is content between the title and first ## section.
	Preamble string

	// Sections are all ## sections in the order they appear.
	Sections []ParsedSection
}

var (
	// h1Pattern matches a level-1 heading.
	h1Pattern = regexp.MustCompile(`^# (.+)$`)

	// h2Pattern matches a level-2 heading (section boundary).
	h2Pattern = regexp.MustCompile(`^## (.+)$`)
)

// ParseAgentSections parses an agent markdown file into structured components.
func ParseAgentSections(content []byte) (*ParsedAgent, error) {
	// Parse frontmatter first
	fm, err := ParseAgentFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Extract raw frontmatter text
	rawFM, bodyStart, err := extractRawFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Parse the body (title, preamble, sections)
	body := content[bodyStart:]
	lines := bytes.Split(body, []byte("\n"))

	agent := &ParsedAgent{
		Frontmatter:    fm,
		RawFrontmatter: rawFM,
	}

	// State machine for parsing
	var (
		inPreamble       = false
		currentSection   *ParsedSection
		preambleLines    []string
		sectionLines     []string
	)

	for i, line := range lines {
		lineStr := string(line)

		// Match h1 (title)
		if matches := h1Pattern.FindStringSubmatch(lineStr); matches != nil {
			agent.Title = matches[1]
			inPreamble = true
			continue
		}

		// Match h2 (section boundary)
		if matches := h2Pattern.FindStringSubmatch(lineStr); matches != nil {
			// Save previous section if any
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(strings.Join(sectionLines, "\n"))
				agent.Sections = append(agent.Sections, *currentSection)
			}

			// Close preamble if we were in it
			if inPreamble {
				agent.Preamble = strings.TrimSpace(strings.Join(preambleLines, "\n"))
				inPreamble = false
			}

			// Start new section
			heading := matches[1]
			currentSection = &ParsedSection{
				Heading:   heading,
				Ownership: OwnerAuthor, // Default to author-owned
			}
			sectionLines = []string{}
			continue
		}

		// Accumulate content
		if inPreamble {
			preambleLines = append(preambleLines, lineStr)
		} else if currentSection != nil {
			sectionLines = append(sectionLines, lineStr)
		} else if i > 0 && agent.Title == "" {
			// Content before title - skip (shouldn't happen in valid files)
			continue
		}
	}

	// Save final section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(strings.Join(sectionLines, "\n"))
		agent.Sections = append(agent.Sections, *currentSection)
	}

	// Save final preamble if still open
	if inPreamble {
		agent.Preamble = strings.TrimSpace(strings.Join(preambleLines, "\n"))
	}

	// Map sections to archetype definitions
	if err := mapSectionsToArchetype(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// extractRawFrontmatter extracts the raw frontmatter text and returns the body start position.
func extractRawFrontmatter(content []byte) (string, int, error) {
	// Find frontmatter delimiters
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return "", 0, errors.New(errors.CodeParseError, "missing frontmatter delimiter")
	}

	// Find closing delimiter
	var endIndex int
	startOffset := 4 // Length of "---\n"
	if bytes.HasPrefix(content, []byte("---\r\n")) {
		startOffset = 5
	}

	if idx := bytes.Index(content[startOffset:], []byte("\n---\n")); idx != -1 {
		endIndex = startOffset + idx
	} else if idx := bytes.Index(content[startOffset:], []byte("\n---\r\n")); idx != -1 {
		endIndex = startOffset + idx
	} else if idx := bytes.Index(content[startOffset:], []byte("\r\n---\r\n")); idx != -1 {
		endIndex = startOffset + idx
	} else if idx := bytes.Index(content[startOffset:], []byte("\r\n---\n")); idx != -1 {
		endIndex = startOffset + idx
	} else {
		return "", 0, errors.New(errors.CodeParseError, "missing closing frontmatter delimiter")
	}

	// Extract raw frontmatter (without delimiters)
	rawFM := string(content[startOffset:endIndex])

	// Calculate body start position (after closing delimiter)
	bodyStart := endIndex + 5 // Skip "\n---\n"
	if bytes.Index(content[endIndex:endIndex+6], []byte("\r\n---\r\n")) == 0 {
		bodyStart = endIndex + 7
	}

	return rawFM, bodyStart, nil
}

// mapSectionsToArchetype maps parsed sections to archetype section definitions.
func mapSectionsToArchetype(agent *ParsedAgent) error {
	// Get archetype from frontmatter type field
	archetypeName := agent.Frontmatter.Type
	if archetypeName == "" {
		// No type field - leave sections unmapped (author-owned)
		return nil
	}

	// Map non-standard types to specialist
	if archetypeName != "orchestrator" && archetypeName != "specialist" && archetypeName != "reviewer" {
		archetypeName = "specialist"
	}

	archetype, err := GetArchetype(archetypeName)
	if err != nil {
		// Unknown archetype - leave sections unmapped
		return nil
	}

	// Build a map of heading -> section definition (case-insensitive)
	sectionMap := make(map[string]*SectionDef)
	for i := range archetype.Sections {
		section := &archetype.Sections[i]
		// Extract heading text without the "## " prefix
		heading := strings.TrimPrefix(section.Heading, "## ")
		normalizedHeading := strings.ToLower(strings.TrimSpace(heading))
		sectionMap[normalizedHeading] = section
	}

	// Match parsed sections to archetype definitions
	for i := range agent.Sections {
		parsedSection := &agent.Sections[i]
		normalizedHeading := strings.ToLower(strings.TrimSpace(parsedSection.Heading))

		if sectionDef, found := sectionMap[normalizedHeading]; found {
			// Exact match found — use it directly.
			parsedSection.Name = sectionDef.Name
			parsedSection.Ownership = sectionDef.Ownership
			continue
		}

		// Exact match failed. Try prefix matching: if any archetype heading is a prefix of
		// the parsed heading, treat it as a match. This handles heading variants like
		// "Behavioral Constraints (DO NOT)" matching archetype "Behavioral Constraints", or
		// "Anti-Patterns to Avoid" matching archetype "Anti-Patterns".
		//
		// We require the archetype prefix to be followed by whitespace, a parenthesis, a dash,
		// a colon, or end-of-string — preventing false matches (e.g., "tool access" must not
		// match a heading "tool accessibility").
		for archetypeNorm, sectionDef := range sectionMap {
			if !strings.HasPrefix(normalizedHeading, archetypeNorm) {
				continue
			}
			// Confirm the character immediately after the prefix is a valid separator.
			remainder := normalizedHeading[len(archetypeNorm):]
			if remainder == "" || remainder[0] == ' ' || remainder[0] == '(' ||
				remainder[0] == '-' || remainder[0] == ':' {
				parsedSection.Name = sectionDef.Name
				parsedSection.Ownership = sectionDef.Ownership
				break
			}
		}
		// If still not found, section remains unmapped (Name="", Ownership=OwnerAuthor)
	}

	return nil
}

// FindSection finds a section by name in the parsed agent.
func (p *ParsedAgent) FindSection(name string) *ParsedSection {
	for i := range p.Sections {
		if p.Sections[i].Name == name {
			return &p.Sections[i]
		}
	}
	return nil
}

// FindSectionByHeading finds a section by heading text (case-insensitive).
func (p *ParsedAgent) FindSectionByHeading(heading string) *ParsedSection {
	normalized := strings.ToLower(strings.TrimSpace(heading))
	for i := range p.Sections {
		sectionHeading := strings.ToLower(strings.TrimSpace(p.Sections[i].Heading))
		if sectionHeading == normalized {
			return &p.Sections[i]
		}
	}
	return nil
}
