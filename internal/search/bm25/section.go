package bm25

import (
	"strings"
)

// Section represents a document section split at ## boundaries.
type Section struct {
	// Heading is the text after "## " on the heading line.
	Heading string
	// Body is the full text content of the section (excluding the heading line).
	Body string
	// Slug is a URL-friendly version of the heading for use in qualified names.
	Slug string
}

// SplitSections splits markdown content at "## " heading boundaries.
// Each heading starts a new section. Content before the first heading
// is captured as a section with an empty heading.
// Returns nil for empty input.
func SplitSections(content string) []Section {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	var sections []Section
	var currentHeading string
	var currentBody strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			// Save previous section if it has content.
			bodyText := currentBody.String()
			if currentHeading != "" || strings.TrimSpace(bodyText) != "" {
				sections = append(sections, Section{
					Heading: currentHeading,
					Body:    bodyText,
					Slug:    Slugify(currentHeading),
				})
			}
			currentHeading = strings.TrimPrefix(line, "## ")
			currentBody.Reset()
		} else {
			currentBody.WriteString(line)
			currentBody.WriteString("\n")
		}
	}

	// Save last section.
	bodyText := currentBody.String()
	if currentHeading != "" || strings.TrimSpace(bodyText) != "" {
		sections = append(sections, Section{
			Heading: currentHeading,
			Body:    bodyText,
			Slug:    Slugify(currentHeading),
		})
	}

	return sections
}

// Slugify converts a section heading to a URL-friendly slug.
// Lowercases, replaces spaces with hyphens, removes non-alphanumeric characters
// (except hyphens), and truncates to 60 characters.
func Slugify(heading string) string {
	if heading == "" {
		return ""
	}

	s := strings.ToLower(strings.TrimSpace(heading))
	s = strings.ReplaceAll(s, " ", "-")

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if len(result) > 60 {
		result = result[:60]
	}
	return result
}

// SectionQualifiedName constructs a section-level qualified name by appending
// the section slug to the parent document's qualified name.
// Format: "org::repo::domain##section-slug"
func SectionQualifiedName(parentQN, slug string) string {
	if slug == "" {
		return parentQN
	}
	return parentQN + "##" + slug
}
