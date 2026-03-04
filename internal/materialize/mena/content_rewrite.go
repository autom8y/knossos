package mena

import (
	"regexp"
	"strings"
)

// Compiled regexes for rewriting mena content path references.
// INDEX-specific patterns MUST be applied before general patterns to avoid
// INDEX.lego.md being rewritten to INDEX.md instead of SKILL.md.

// reLinkIndexLego matches markdown link targets containing INDEX.lego.md.
// Example: ](path/INDEX.lego.md) or ](path/INDEX.lego.md#fragment)
var reLinkIndexLego = regexp.MustCompile(`\]\(([^)]*?)INDEX\.lego\.md(#[^)]*?)?\)`)

// reLinkLego matches markdown link targets containing .lego.md (general case).
// Example: ](path/name.lego.md) or ](path/name.lego.md#section)
var reLinkLego = regexp.MustCompile(`\]\(([^)]*?)\.lego\.md(#[^)]*?)?\)`)

// reLinkDro matches markdown link targets containing .dro.md.
// Example: ](path/name.dro.md) or ](path/name.dro.md#section)
var reLinkDro = regexp.MustCompile(`\]\(([^)]*?)\.dro\.md(#[^)]*?)?\)`)

// reBacktickIndexLego matches backtick code spans containing INDEX.lego.md.
// Example: `path/INDEX.lego.md`
var reBacktickIndexLego = regexp.MustCompile("`([^`]*?)INDEX\\.lego\\.md([^`]*?)`")

// reBacktickLego matches backtick code spans containing .lego.md (general case).
// Example: `path/name.lego.md`
var reBacktickLego = regexp.MustCompile("`([^`]*?)\\.lego\\.md([^`]*?)`")

// reBacktickDro matches backtick code spans containing .dro.md.
// Example: `path/name.dro.md`
var reBacktickDro = regexp.MustCompile("`([^`]*?)\\.dro\\.md([^`]*?)`")

// rewriteMenaContentPaths rewrites stale .lego.md/.dro.md content references
// to their materialized forms. It applies the same extension-stripping logic
// that the materializer applies to filenames, but at the content level.
//
// Transformation rules:
//   - INDEX.lego.md -> SKILL.md (in link targets and backtick spans)
//   - {name}.lego.md -> {name}.md (in link targets and backtick spans)
//   - {name}.dro.md -> {name}.md (in link targets and backtick spans)
//
// Content inside fenced code blocks (``` regions) is never modified.
// YAML frontmatter passes through naturally as it does not contain markdown
// link targets or backtick code spans with mena extensions.
//
// This is a pure function: no side effects, no file I/O.
func rewriteMenaContentPaths(content []byte) []byte {
	if len(content) == 0 {
		return content
	}

	s := string(content)

	// Split on fenced code block boundaries.
	// Lines starting with ``` (with optional language suffix) delimit fenced blocks.
	// Odd-indexed segments are inside fences and pass through unchanged.
	// Even-indexed segments are outside fences and have rewrites applied.
	//
	// We split on "\n```" (newline + triple backtick) to avoid matching triple
	// backticks that appear mid-line (inside inline code). The first segment
	// is handled specially: if the content starts with ```, it is already inside
	// a fence opener.
	segments := splitOnFences(s)

	for i := range segments {
		if i%2 == 0 {
			// Outside fenced block: apply rewrites
			segments[i] = applyRewrites(segments[i])
		}
		// Odd segments (inside fenced blocks): pass through unchanged
	}

	return []byte(strings.Join(segments, ""))
}

// splitOnFences splits content into alternating non-fenced and fenced segments.
// segments[0] is always non-fenced (may be empty if content starts with a fence).
// segments[1] is fenced, segments[2] is non-fenced, etc.
//
// The split is performed on line boundaries: a fence boundary is a line whose
// content starts with ```. We preserve the delimiter as part of the fenced segment
// so that rejoining produces the original content.
func splitOnFences(s string) []string {
	var segments []string
	remaining := s
	inFence := false

	// We accumulate the current segment character by character via line scanning.
	var current strings.Builder

	lines := strings.SplitAfter(remaining, "\n")
	for _, line := range lines {
		// Determine if this line opens or closes a fenced block.
		// A fence marker is a line that starts with ``` (after stripping \r).
		stripped := strings.TrimRight(line, "\r\n")
		isFenceMarker := strings.HasPrefix(stripped, "```")

		if isFenceMarker {
			if !inFence {
				// Transition: non-fenced -> fenced.
				// Flush current non-fenced segment.
				segments = append(segments, current.String())
				current.Reset()
				inFence = true
			} else {
				// Transition: fenced -> non-fenced.
				// Include the closing ``` line in the fenced segment.
				current.WriteString(line)
				segments = append(segments, current.String())
				current.Reset()
				inFence = false
				continue
			}
		}

		current.WriteString(line)
	}

	// Flush final segment.
	segments = append(segments, current.String())

	return segments
}

// applyRewrites applies all mena path rewrite patterns to a non-fenced segment.
// INDEX-specific patterns run before general patterns (ordering constraint).
func applyRewrites(s string) string {
	// Pass 1: INDEX.lego.md -> SKILL.md (must precede general .lego.md pattern)
	s = reLinkIndexLego.ReplaceAllString(s, "](${1}SKILL.md${2})")
	s = reBacktickIndexLego.ReplaceAllString(s, "`${1}SKILL.md${2}`")

	// Pass 2: {name}.lego.md -> {name}.md
	s = reLinkLego.ReplaceAllString(s, "](${1}.md${2})")
	s = reBacktickLego.ReplaceAllString(s, "`${1}.md${2}`")

	// Pass 3: {name}.dro.md -> {name}.md
	s = reLinkDro.ReplaceAllString(s, "](${1}.md${2})")
	s = reBacktickDro.ReplaceAllString(s, "`${1}.md${2}`")

	return s
}
