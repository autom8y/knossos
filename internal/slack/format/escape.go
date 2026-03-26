package format

import (
	"strings"
)

var slackTokenPrefixes = []string{
	"@",
	"#",
	"!",
	"http://",
	"https://",
	"mailto:",
	"tel:",
	"slack://",
}

func isSlackToken(inner string) bool {
	for _, prefix := range slackTokenPrefixes {
		if strings.HasPrefix(inner, prefix) {
			return true
		}
	}
	return false
}

// Escape applies Slack-safe HTML entity escaping to plain text.
// It converts & -> &amp;, < -> &lt;, > -> &gt; while preserving
// Slack angle-bracket tokens (mentions, links, channels, etc.)
// and avoiding double-escaping of existing entities.
func Escape(text string) string {
	var b strings.Builder
	b.Grow(len(text) + len(text)/10)

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if strings.HasPrefix(line, "> ") {
			b.WriteString("> ")
			b.WriteString(escapeLine(line[2:]))
		} else {
			b.WriteString(escapeLine(line))
		}
	}
	return b.String()
}

func escapeLine(line string) string {
	var b strings.Builder
	b.Grow(len(line) + len(line)/10)

	i := 0
	for i < len(line) {
		ch := line[i]

		if ch == '&' {
			// Check for already-escaped entities to avoid double-escaping.
			if strings.HasPrefix(line[i:], "&amp;") ||
				strings.HasPrefix(line[i:], "&lt;") ||
				strings.HasPrefix(line[i:], "&gt;") {
				// Find the semicolon and pass through the entity.
				semi := strings.IndexByte(line[i:], ';')
				b.WriteString(line[i : i+semi+1])
				i += semi + 1
				continue
			}
			b.WriteString("&amp;")
			i++
			continue
		}

		if ch == '<' {
			// Look for a matching > to see if this is a Slack token.
			closeIdx := strings.IndexByte(line[i:], '>')
			if closeIdx > 1 {
				inner := line[i+1 : i+closeIdx]
				if isSlackToken(inner) {
					b.WriteString(line[i : i+closeIdx+1])
					i += closeIdx + 1
					continue
				}
			}
			b.WriteString("&lt;")
			i++
			continue
		}

		if ch == '>' {
			b.WriteString("&gt;")
			i++
			continue
		}

		b.WriteByte(ch)
		i++
	}
	return b.String()
}
