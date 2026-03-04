package mena

import (
	"regexp"
	"strings"
	"testing"
)

// sourceExtPatternForTest mirrors the lint.go sourceExtPattern used in
// production lint checks. After rewriting, no non-fenced segment should
// contain a match.
var sourceExtPatternForTest = regexp.MustCompile(`\.(lego|dro)\.md`)

// TestRewriteMenaContentPaths verifies all rewrite patterns: markdown link targets,
// backtick code spans, INDEX.lego.md -> SKILL.md rename, fenced block exclusion,
// fragment preservation, and no-op cases.
func TestRewriteMenaContentPaths(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "link_lego_simple",
			input: "[criteria](domains/dromena.lego.md)",
			want:  "[criteria](domains/dromena.md)",
		},
		{
			name:  "link_lego_relative",
			input: "[parent](../INDEX.lego.md)",
			want:  "[parent](../SKILL.md)",
		},
		{
			name:  "link_lego_fragment",
			input: "[ref](../INDEX.lego.md#section)",
			want:  "[ref](../SKILL.md#section)",
		},
		{
			name:  "link_lego_deep_path",
			input: "[grading](../../pinakes/schemas/report-format.lego.md)",
			want:  "[grading](../../pinakes/schemas/report-format.md)",
		},
		{
			name:  "link_lego_same_dir",
			input: "[peer](legomena.lego.md)",
			want:  "[peer](legomena.md)",
		},
		{
			name:  "link_lego_glob",
			input: "[domains](../domains/*.lego.md)",
			want:  "[domains](../domains/*.md)",
		},
		{
			name:  "link_dro_simple",
			input: "[ref](../INDEX.dro.md)",
			want:  "[ref](../INDEX.md)",
		},
		{
			name:  "link_dro_fragment",
			input: "[ref](theoria.dro.md#section)",
			want:  "[ref](theoria.md#section)",
		},
		{
			name:  "backtick_lego_simple",
			input: "`domains/dromena.lego.md`",
			want:  "`domains/dromena.md`",
		},
		{
			name:  "backtick_lego_index",
			input: "`INDEX.lego.md`",
			want:  "`SKILL.md`",
		},
		{
			name:  "backtick_lego_path",
			input: "`mena/pinakes/domains/{domain}.lego.md`",
			want:  "`mena/pinakes/domains/{domain}.md`",
		},
		{
			name:  "backtick_dro_simple",
			input: "`mena/navigation/rite.dro.md`",
			want:  "`mena/navigation/rite.md`",
		},
		{
			name:  "backtick_materialized_index",
			input: "`.claude/skills/*/INDEX.lego.md`",
			want:  "`.claude/skills/*/SKILL.md`",
		},
		{
			name:  "fenced_block_preserved",
			input: "```\n`INDEX.lego.md`\n```",
			want:  "```\n`INDEX.lego.md`\n```",
		},
		{
			name:  "fenced_yaml_preserved",
			input: "```yaml\nINDEX.lego.md\n```",
			want:  "```yaml\nINDEX.lego.md\n```",
		},
		{
			name:  "multiple_on_one_line",
			input: "[a](foo.lego.md) and [b](bar.lego.md)",
			want:  "[a](foo.md) and [b](bar.md)",
		},
		{
			name:  "mixed_link_and_backtick",
			input: "[ref](foo.lego.md) see `bar.lego.md`",
			want:  "[ref](foo.md) see `bar.md`",
		},
		{
			name:  "no_match_prose",
			input: "files use .lego.md extension",
			want:  "files use .lego.md extension",
		},
		{
			name:  "no_match_frontmatter",
			input: "---\nname: foo\n---\n[ref](bar.lego.md)",
			want:  "---\nname: foo\n---\n[ref](bar.md)",
		},
		{
			name:  "empty_input",
			input: "",
			want:  "",
		},
		{
			name:  "no_matches",
			input: "plain text with no refs",
			want:  "plain text with no refs",
		},
		{
			name:  "index_dro_link",
			input: "[ref](../INDEX.dro.md)",
			want:  "[ref](../INDEX.md)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(RewriteMenaContentPaths([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("rewriteMenaContentPaths(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestRewriteCorpus is the regression gate for the mena content rewriter.
// It processes a fixture containing all known pattern categories and asserts
// that zero .lego.md/.dro.md references survive in non-fenced output segments.
//
// This mirrors the sourceExtPattern check from internal/cmd/lint/lint.go and
// ensures the materializer pipeline correctly eliminates stale source extensions
// from all link targets and backtick code spans.
func TestRewriteCorpus(t *testing.T) {
	// corpus is a representative mena document containing all pattern categories
	// drawn from the 262 instances identified in the smell report:
	//   - Markdown link targets (INDEX.lego.md, general .lego.md, .dro.md)
	//   - Fragment-preserving links
	//   - Backtick code spans (INDEX.lego.md, general .lego.md, .dro.md)
	//   - Multiple refs on one line
	//   - Mixed link + backtick on same line
	//   - YAML frontmatter (must be untouched structurally; body refs are rewritten)
	//   - Fenced code blocks (must be completely preserved)
	corpus := `---
name: corpus-test
description: "Corpus fixture for rewriter validation."
---

# Corpus Document

## Link Targets

- See [dromena criteria](domains/dromena.lego.md) for evaluation.
- Parent skill: [INDEX](../INDEX.lego.md)
- Anchor: [section](../INDEX.lego.md#criteria)
- Deep path: [grading](../../pinakes/schemas/grading.lego.md)
- Same dir: [legomena](legomena.lego.md)
- Glob: [all](../domains/*.lego.md)
- Dro link: [commands](INDEX.dro.md)
- Dro fragment: [phase](theoria.dro.md#phase-one)
- Multiple on one line: [a](foo.lego.md) and [b](bar.lego.md)

## Backtick Spans

- Plain: ` + "`" + `domains/dromena.lego.md` + "`" + `
- INDEX: ` + "`" + `INDEX.lego.md` + "`" + `
- Path with template var: ` + "`" + `mena/pinakes/domains/{domain}.lego.md` + "`" + `
- Dro span: ` + "`" + `mena/navigation/rite.dro.md` + "`" + `
- Materialized YAML example: ` + "`" + `.claude/skills/*/INDEX.lego.md` + "`" + `
- Mixed link and backtick: [ref](foo.lego.md) see ` + "`" + `bar.lego.md` + "`" + `

## Fenced Blocks (must be fully preserved)

` + "```" + `
[ref](INDEX.lego.md)
` + "`" + `foo.dro.md` + "`" + `
` + "```" + `

` + "```" + `yaml
skills:
  - .claude/skills/*/INDEX.lego.md
` + "```" + `

## Go Code Example (fenced, must be preserved)

` + "```" + `go
// Files named INDEX.lego.md become SKILL.md after materialization.
const ext = ".lego.md"
` + "```" + `

## After Fenced Blocks

More links after fenced content: [back](../schemas/report.lego.md)
`

	output := string(RewriteMenaContentPaths([]byte(corpus)))

	// Split the output on fence boundaries (same logic as the rewriter).
	// Odd-indexed segments are inside fences and are exempt from the check.
	segments := splitOnFences(output)

	var violations []string
	for i, seg := range segments {
		if i%2 != 0 {
			// Inside fenced block: skip.
			continue
		}
		// Outside fenced block: no .lego.md or .dro.md should survive.
		for lineNum, line := range strings.Split(seg, "\n") {
			if sourceExtPatternForTest.MatchString(line) {
				violations = append(violations,
					"non-fenced segment "+string(rune('0'+i))+
						" line "+string(rune('0'+lineNum))+
						": "+line)
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("stale mena extension refs survived rewriting in non-fenced content (%d violations):", len(violations))
		for _, v := range violations {
			t.Errorf("  %s", v)
		}
	}

	// Sanity check: fenced blocks are still intact in the output.
	if !strings.Contains(output, "```\n[ref](INDEX.lego.md)\n") {
		t.Error("fenced block content was incorrectly modified: expected INDEX.lego.md preserved inside fence")
	}
	if !strings.Contains(output, `const ext = ".lego.md"`) {
		t.Error("fenced Go code block was incorrectly modified: expected .lego.md preserved inside fence")
	}
}
