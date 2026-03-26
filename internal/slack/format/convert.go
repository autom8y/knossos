package format

import (
	"regexp"
	"strings"
)

var (
	reInlineCode = regexp.MustCompile("`[^`]+`")

	// Triple asterisk (bold+italic) must be matched before double.
	reTripleBold     = regexp.MustCompile(`\*\*\*(.+?)\*\*\*`)
	reBoldAsterisk   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reBoldUnderscore = regexp.MustCompile(`__(.+?)__`)

	// Strikethrough: ~~text~~
	reStrike = regexp.MustCompile(`~~(.+?)~~`)

	// ATX headings: # Heading through ###### Heading.
	reHeading = regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)

	// Horizontal rules: ---, ***, ___ (with optional spaces between chars).
	reHR = regexp.MustCompile(`(?m)^[ \t]*([-*_][ \t]*){3,}$`)

	// Images: ![alt](url)
	reImage = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)

	// Links: [label](url)
	reLink = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	// Spoiler syntax: ||text|| (Discord-style) or >!text!< (Reddit-style).
	reSpoilerPipe   = regexp.MustCompile(`\|\|(.+?)\|\|`)
	reSpoilerReddit = regexp.MustCompile(`>!(.+?)!<`)
)

// Convert transforms GitHub Flavored Markdown into Slack mrkdwn.
// Processing order: protect code regions, apply style conversions
// on non-code segments, then escape plain text.
func Convert(markdown string) string {
	var placeholders []phEntry
	nextID := 0

	makeToken := func() string {
		t := "\x00CB" + string(rune(nextID)) + "\x00"
		nextID++
		return t
	}

	result := markdown

	result = extractCodeBlocks(result, &placeholders, &nextID, makeToken)

	result = reInlineCode.ReplaceAllStringFunc(result, func(match string) string {
		tok := makeToken()
		placeholders = append(placeholders, phEntry{token: tok, content: match})
		return tok
	})

	// Phase 2: Apply style conversions on the non-code text.

	// Triple asterisk (bold+italic): ***text*** -> *_text_* in Slack.
	result = reTripleBold.ReplaceAllString(result, "\x01B_${1}_\x01B")

	// Bold: convert to temporary marker \x01B...\x01B so italic pass doesn't consume them.
	result = reBoldAsterisk.ReplaceAllString(result, "\x01B$1\x01B")
	result = reBoldUnderscore.ReplaceAllString(result, "\x01B$1\x01B")

	// Strikethrough (before italic to avoid ~~ interference).
	result = reStrike.ReplaceAllString(result, "~$1~")

	// Italic: GFM *text* -> Slack _text_
	result = convertItalicAsterisk(result)

	// Headings BEFORE bold restore: strip inner temp-bold markers to avoid **double-bold**
	result = reHeading.ReplaceAllStringFunc(result, func(match string) string {
		m := reHeading.FindStringSubmatch(match)
		content := m[1]
		content = strings.ReplaceAll(content, "\x01B", "")
		return "*" + strings.TrimSpace(content) + "*"
	})

	// Restore bold markers.
	result = strings.ReplaceAll(result, "\x01B", "*")

	// Images (before links since images match a superset of the link pattern).
	result = reImage.ReplaceAllString(result, "$1")

	// Links: [label](url) -> <url|label> or bare url.
	result = reLink.ReplaceAllStringFunc(result, func(match string) string {
		m := reLink.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		label, url := m[1], m[2]
		if label == url {
			return url
		}
		return "<" + url + "|" + label + ">"
	})

	// Horizontal rules.
	result = reHR.ReplaceAllString(result, "───\n")

	// Spoilers: drop content.
	result = reSpoilerPipe.ReplaceAllString(result, "$1")
	result = reSpoilerReddit.ReplaceAllString(result, "$1")

	// Phase 3: Escape plain-text segments (everything that isn't a placeholder).
	result = escapeNonPlaceholders(result, placeholders)

	// Phase 4: Restore placeholders.
	for _, p := range placeholders {
		result = strings.Replace(result, p.token, p.content, 1)
	}

	return result
}

type phEntry struct {
	token   string
	content string
}

func extractCodeBlocks(text string, placeholders *[]phEntry, nextID *int, makeToken func() string) string {
	lines := strings.Split(text, "\n")
	var out []string
	var fenceContent []string
	inFence := false
	var fenceMarker string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inFence {
			if isFenceClose(trimmed, fenceMarker) {
				fenceContent = append(fenceContent, "```")
				tok := makeToken()
				*placeholders = append(*placeholders, phEntry{
					token:   tok,
					content: strings.Join(fenceContent, "\n"),
				})
				out = append(out, tok)
				inFence = false
				fenceContent = nil
			} else {
				fenceContent = append(fenceContent, line)
			}
			continue
		}

		if isOpenFence(trimmed) {
			inFence = true
			fenceMarker = extractFenceMarker(trimmed)
			// Strip language tag from opening fence for Slack.
			fenceContent = []string{"```"}
			continue
		}

		out = append(out, line)
	}

	// Unclosed fence: treat accumulated content as a code block.
	if inFence && len(fenceContent) > 0 {
		fenceContent = append(fenceContent, "```")
		tok := makeToken()
		*placeholders = append(*placeholders, phEntry{
			token:   tok,
			content: strings.Join(fenceContent, "\n"),
		})
		out = append(out, tok)
	}

	return strings.Join(out, "\n")
}

func isOpenFence(trimmed string) bool {
	return strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")
}

func extractFenceMarker(trimmed string) string {
	if strings.HasPrefix(trimmed, "```") {
		return "```"
	}
	return "~~~"
}

func isFenceClose(trimmed, marker string) bool {
	// A close fence is exactly the marker (``` or ~~~) with optional trailing whitespace.
	if marker == "```" {
		return trimmed == "```"
	}
	return trimmed == "~~~"
}

// convertItalicAsterisk converts remaining single-asterisk italic (*text*)
// to Slack italic (_text_). At this point, bold ** has already been converted
// to Slack bold *, so we need to be careful not to match Slack bold markers.
func convertItalicAsterisk(text string) string {
	// Simple state-machine approach: find *text* where text doesn't contain *
	// and the surrounding context isn't another *.
	var b strings.Builder
	b.Grow(len(text))
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		if runes[i] == '*' {
			// Check this isn't adjacent to another * (which would be bold).
			prevIsStar := i > 0 && runes[i-1] == '*'
			nextIsStar := i+1 < len(runes) && runes[i+1] == '*'
			if !prevIsStar && !nextIsStar {
				// GFM italic requires no space after opening * and no space before closing *.
				if i+1 < len(runes) && runes[i+1] == ' ' {
					b.WriteRune(runes[i])
					i++
					continue
				}
				// Look for closing *.
				end := -1
				for j := i + 1; j < len(runes); j++ {
					if runes[j] == '\n' {
						break
					}
					if runes[j] == '*' {
						nextJ := j+1 < len(runes) && runes[j+1] == '*'
						prevSpace := j > 0 && runes[j-1] == ' '
						if !nextJ && !prevSpace {
							end = j
							break
						}
					}
				}
				if end > i+1 {
					b.WriteRune('_')
					for k := i + 1; k < end; k++ {
						b.WriteRune(runes[k])
					}
					b.WriteRune('_')
					i = end + 1
					continue
				}
			}
		}
		b.WriteRune(runes[i])
		i++
	}
	return b.String()
}

func escapeNonPlaceholders(text string, placeholders []phEntry) string {
	if len(placeholders) == 0 {
		return Escape(text)
	}

	// Split on placeholder tokens and escape each segment.
	var b strings.Builder
	b.Grow(len(text) + len(text)/10)

	remaining := text
	for _, p := range placeholders {
		idx := strings.Index(remaining, p.token)
		if idx < 0 {
			continue
		}
		b.WriteString(Escape(remaining[:idx]))
		b.WriteString(p.token)
		remaining = remaining[idx+len(p.token):]
	}
	b.WriteString(Escape(remaining))

	return b.String()
}
