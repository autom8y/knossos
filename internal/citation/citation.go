// Package citation provides platform-wide citation marker parsing.
// This is a LEAF package — it imports only stdlib.
package citation

import "regexp"

// citationPattern matches inline citations in free-form text.
// Format: [org::repo::domain] e.g., [autom8y::knossos::architecture]
var citationPattern = regexp.MustCompile(`\[([a-zA-Z0-9_-]+::[a-zA-Z0-9_-]+::[a-zA-Z0-9_-]+)\]`)

// ExtractCitations parses inline citation markers from free-form text.
// Returns deduplicated qualified names in order of first appearance.
func ExtractCitations(text string) []string {
	matches := citationPattern.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var citations []string
	for _, m := range matches {
		if len(m) >= 2 && !seen[m[1]] {
			seen[m[1]] = true
			citations = append(citations, m[1])
		}
	}
	return citations
}
