package format

import (
	"regexp"
	"strings"
)

var reChunkFence = regexp.MustCompile("(?m)^([ \t]*)(```|~~~)(.*)$")

type fenceSpan struct {
	openPos  int
	closePos int // -1 means unclosed (extends to end of text)
	marker   string
	langTag  string
	indent   string
}

// Chunk splits text into pieces no longer than maxLen, respecting code fence
// boundaries. When a break falls inside a fenced code block, the current
// chunk gets a closing fence appended and the next chunk gets the opening
// fence prepended.
func Chunk(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	fences := scanFences(text)

	var chunks []string
	remaining := text
	offset := 0

	for len(remaining) > maxLen {
		cut := findBreak(remaining, maxLen)

		absPos := offset + cut
		fence := fenceContaining(fences, absPos)

		if fence != nil {
			// Break inside a code fence: close and reopen.
			chunk := remaining[:cut]
			chunk = strings.TrimRight(chunk, "\n") + "\n" + fence.indent + fence.marker
			chunks = append(chunks, chunk)

			rest := remaining[cut:]
			rest = strings.TrimLeft(rest, "\n")
			reopenLine := fence.indent + fence.marker + fence.langTag + "\n"
			remaining = reopenLine + rest
			// Offset tracks position in original text, but we've injected
			// synthetic content. Advance offset past the cut point.
			offset += cut
		} else {
			chunk := remaining[:cut]
			chunks = append(chunks, chunk)
			afterCut := remaining[cut:]
			remaining = strings.TrimLeft(afterCut, "\n")
			offset += cut + (len(afterCut) - len(remaining))
		}
	}

	if len(remaining) > 0 {
		chunks = append(chunks, remaining)
	}

	return chunks
}

func scanFences(text string) []fenceSpan {
	matches := reChunkFence.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}

	allMatches := reChunkFence.FindAllStringSubmatch(text, -1)

	var fences []fenceSpan
	var openFence *fenceSpan

	for i, loc := range matches {
		pos := loc[0]
		indent := allMatches[i][1]
		marker := allMatches[i][2]
		rest := strings.TrimSpace(allMatches[i][3])

		if openFence == nil {
			// Opening fence.
			openFence = &fenceSpan{
				openPos: pos,
				marker:  marker,
				langTag: rest,
				indent:  indent,
			}
		} else if marker == openFence.marker && rest == "" {
			// Closing fence (same marker type, no lang tag).
			openFence.closePos = loc[1]
			fences = append(fences, *openFence)
			openFence = nil
		}
		// Otherwise: different marker or has content -> not a close, skip.
	}

	// Unclosed fence extends to end of text.
	if openFence != nil {
		openFence.closePos = len(text)
		fences = append(fences, *openFence)
	}

	return fences
}

func fenceContaining(fences []fenceSpan, pos int) *fenceSpan {
	for i := range fences {
		if pos > fences[i].openPos && pos < fences[i].closePos {
			return &fences[i]
		}
	}
	return nil
}

func findBreak(text string, maxLen int) int {
	if len(text) <= maxLen {
		return len(text)
	}

	window := text[:maxLen]

	// Prefer paragraph break.
	idx := strings.LastIndex(window, "\n\n")
	if idx > 0 {
		return idx
	}

	// Fall back to line break.
	idx = strings.LastIndex(window, "\n")
	if idx > 0 {
		return idx
	}

	// Hard cut.
	return maxLen
}
