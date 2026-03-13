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

// channelDirName returns the dot-prefixed directory name for a channel.
// Empty or "claude" returns ".claude"; "gemini" returns ".gemini".
func channelDirName(channel string) string {
	switch channel {
	case "gemini":
		return ".gemini"
	default:
		return ".claude"
	}
}

// rewriteChannelPaths rewrites channel path placeholders in non-fenced content
// to the target channel directory. Processing is line-by-line to respect
// exclusion zones (template blocks and HA-tagged lines).
//
// Two source patterns are handled:
//   - ".channel/" — the canonical agnostic placeholder in mena source files
//   - ".claude/" — legacy fallback for source files not yet migrated to .channel/ (identity when targeting CC)
//
// Both are rewritten to the target channelDir (e.g., ".claude" or ".gemini").
// When channelDir is ".claude", the .channel/ placeholder is still rewritten
// (to ".claude/"), but legacy ".claude/" references are identity when targeting CC (no-op).
//
// Only the three mena content subdirectories are rewritten:
// {skills,commands,agents}. Other paths (settings.json, CLAUDE.md) are NOT affected.
func rewriteChannelPaths(s string, channelDir string) string {
	if channelDir == "" {
		channelDir = ".claude"
	}

	// Process line-by-line to respect exclusion zones
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		// Skip lines inside template blocks
		if strings.Contains(line, "{{") && strings.Contains(line, "}}") {
			continue
		}
		// Skip HA-tagged lines
		if strings.Contains(line, "// HA-") {
			continue
		}
		// Canonical placeholder: .channel/ → target channel dir
		lines[i] = strings.ReplaceAll(line, ".channel/skills/", channelDir+"/skills/")
		lines[i] = strings.ReplaceAll(lines[i], ".channel/commands/", channelDir+"/commands/")
		lines[i] = strings.ReplaceAll(lines[i], ".channel/agents/", channelDir+"/agents/")
		// Legacy fallback: .claude/ → target channel dir (identity when channelDir is ".claude")
		if channelDir != ".claude" {
			lines[i] = strings.ReplaceAll(lines[i], ".claude/skills/", channelDir+"/skills/")
			lines[i] = strings.ReplaceAll(lines[i], ".claude/commands/", channelDir+"/commands/")
			lines[i] = strings.ReplaceAll(lines[i], ".claude/agents/", channelDir+"/agents/")
		}
	}
	return strings.Join(lines, "\n")
}

// RewriteMenaContentPaths rewrites stale .lego.md/.dro.md content references
// to their materialized forms and substitutes legacy .claude/ path prefixes with the
// target channel directory. It applies the same extension-stripping logic
// that the materializer applies to filenames, but at the content level.
//
// Transformation rules:
//   - INDEX.lego.md -> SKILL.md (in link targets and backtick spans)
//   - {name}.lego.md -> {name}.md (in link targets and backtick spans)
//   - {name}.dro.md -> {name}.md (in link targets and backtick spans)
//   - .channel/{skills,commands,agents}/ -> {channelDir}/{skills,commands,agents}/ (canonical)
//   - .claude/{skills,commands,agents}/ -> {channelDir}/{skills,commands,agents}/ (legacy; identity for CC channel)
//
// Content inside fenced code blocks (``` regions) is never modified.
// Lines containing template blocks ({{ }}) or HA-tags are not channel-rewritten.
// YAML frontmatter passes through naturally as it does not contain markdown
// link targets or backtick code spans with mena extensions.
//
// When channelDir is empty or ".claude", channel rewriting is identity (no-op).
//
// This is a pure function: no side effects, no file I/O.
func RewriteMenaContentPaths(content []byte, channelDir string) []byte {
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
			segments[i] = applyRewrites(segments[i], channelDir)
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
// Channel path rewriting runs last (Pass 4) so extension rewrites that produce
// .claude/skills/*.md paths get channel-rewritten too (identity for CC channel).
func applyRewrites(s string, channelDir string) string {
	// Pass 1: INDEX.lego.md -> SKILL.md (must precede general .lego.md pattern)
	s = reLinkIndexLego.ReplaceAllString(s, "](${1}SKILL.md${2})")
	s = reBacktickIndexLego.ReplaceAllString(s, "`${1}SKILL.md${2}`")

	// Pass 2: {name}.lego.md -> {name}.md
	s = reLinkLego.ReplaceAllString(s, "](${1}.md${2})")
	s = reBacktickLego.ReplaceAllString(s, "`${1}.md${2}`")

	// Pass 3: {name}.dro.md -> {name}.md
	s = reLinkDro.ReplaceAllString(s, "](${1}.md${2})")
	s = reBacktickDro.ReplaceAllString(s, "`${1}.md${2}`")

	// Pass 4: Channel path rewriting
	// legacy .claude/{skills,commands,agents}/ -> {channelDir}/{skills,commands,agents}/
	s = rewriteChannelPaths(s, channelDir)

	return s
}
