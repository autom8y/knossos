package streaming

import (
	"regexp"
)

// citationPattern matches inline citations in free-form streaming text.
// Format: [org::repo::domain] e.g., [autom8y::knossos::architecture]
// Novel Discovery 5: Streaming uses post-hoc citation parsing (less reliable
// than tool-forced structured output). Target: 90% extraction success rate.
var citationPattern = regexp.MustCompile(`\[([a-zA-Z0-9_-]+::[a-zA-Z0-9_-]+::[a-zA-Z0-9_-]+)\]`)

// ExtractCitations parses inline citation markers from free-form text.
// Returns deduplicated qualified names in order of first appearance.
//
// This is the streaming path counterpart to the structured output parser.
// When streaming is active, Claude produces free-form markdown with inline
// [org::repo::domain] markers instead of tool-forced structured JSON.
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
