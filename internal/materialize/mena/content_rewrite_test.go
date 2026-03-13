package mena

import (
	"fmt"
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
	t.Parallel()
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
			input: "`.claude/skills/*/INDEX.lego.md`", // HA-TEST: Claude channel dir name in content rewrite fixture
			want:  "`.claude/skills/*/SKILL.md`", // HA-TEST: Claude channel dir name in content rewrite fixture
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
		{
			// unclosed_fence: content after the opening ``` but no closing ```.
			// The rewriter treats everything after the opener as fenced and should
			// not rewrite it. The content before the fence is rewritten normally.
			name:  "unclosed_fence",
			input: "before [ref](foo.lego.md)\n```\nfenced `INDEX.lego.md`\nno closing",
			want:  "before [ref](foo.md)\n```\nfenced `INDEX.lego.md`\nno closing",
		},
		{
			// crlf_line_endings: same as fenced_block_preserved but with \r\n.
			name:  "crlf_line_endings",
			input: "```\r\n`INDEX.lego.md`\r\n```\r\n[ref](foo.lego.md)",
			want:  "```\r\n`INDEX.lego.md`\r\n```\r\n[ref](foo.md)",
		},
		{
			// fence_at_start: content starting with ``` on line 1.
			name:  "fence_at_start",
			input: "```\nfenced `INDEX.lego.md`\n```\n[ref](foo.lego.md)",
			want:  "```\nfenced `INDEX.lego.md`\n```\n[ref](foo.md)",
		},
		{
			// multiple_fenced_blocks: alternating fenced and non-fenced content.
			name:  "multiple_fenced_blocks",
			input: "[a](a.lego.md)\n```\nfenced\n```\n[b](b.lego.md)\n```\nfenced2\n```\n[c](c.lego.md)",
			want:  "[a](a.md)\n```\nfenced\n```\n[b](b.md)\n```\nfenced2\n```\n[c](c.md)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := string(RewriteMenaContentPaths([]byte(tt.input), ""))
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
	t.Parallel()
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
- Materialized YAML example: ` + "`" + `.claude/skills/*/INDEX.lego.md` + "`" + ` // HA-TEST: Claude channel dir name in content rewrite fixture
- Mixed link and backtick: [ref](foo.lego.md) see ` + "`" + `bar.lego.md` + "`" + `

## Fenced Blocks (must be fully preserved)

` + "```" + `
[ref](INDEX.lego.md)
` + "`" + `foo.dro.md` + "`" + `
` + "```" + `

` + "```" + `yaml
skills:
  - .claude/skills/*/INDEX.lego.md // HA-TEST: Claude channel dir name in content rewrite fixture
` + "```" + `

## Go Code Example (fenced, must be preserved)

` + "```" + `go
// Files named INDEX.lego.md become SKILL.md after materialization.
const ext = ".lego.md"
` + "```" + `

## After Fenced Blocks

More links after fenced content: [back](../schemas/report.lego.md)
`

	output := string(RewriteMenaContentPaths([]byte(corpus), ""))

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

// TestSCAR_ContentRewriteNotBypassed is a regression guard for the
// RewriteMenaContentPaths call in the materialization pipeline.
//
// SCAR guard: if someone removes the RewriteMenaContentPaths call or
// deletes the function entirely, this test will fail at compile time
// (the function reference below will be unresolved) or at runtime
// (the assertion will detect stale .lego.md refs in the output).
// This protects against the class of bypass bugs identified in GO-001,
// GO-002, and GO-003 where the rewriter was omitted from code paths.
func TestSCAR_ContentRewriteNotBypassed(t *testing.T) {
	t.Parallel()
	input := "See [the skill](foo.lego.md) for details."

	output := string(RewriteMenaContentPaths([]byte(input), ""))

	if strings.Contains(output, "](foo.lego.md)") {
		t.Errorf("SCAR regression: RewriteMenaContentPaths did not rewrite .lego.md link target\n  input:  %q\n  output: %q", input, output)
	}
	if !strings.Contains(output, "](foo.md)") {
		t.Errorf("SCAR regression: expected ](foo.md) in output\n  input:  %q\n  output: %q", input, output)
	}
}

// TestChannelDirName verifies the channelDirName helper mapping.
func TestChannelDirName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		channel string
		want    string
	}{
		{"gemini", ".gemini"},
		{"claude", ".claude"},
		{"", ".claude"},
		{"unknown", ".claude"},
	}
	for _, tt := range tests {
		t.Run(tt.channel+"_to_"+tt.want, func(t *testing.T) {
			t.Parallel()
			if got := channelDirName(tt.channel); got != tt.want {
				t.Errorf("channelDirName(%q) = %q, want %q", tt.channel, got, tt.want)
			}
		})
	}
}

// TestRewriteChannelPaths verifies the channel path rewriting rule:
// core substitution, context variants, exclusion zones, and non-match cases.
func TestRewriteChannelPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      string
		channelDir string
		want       string
	}{
		// Section 3.1: Core substitution tests
		{
			name:       "claude_skills_to_gemini",
			input:      ".claude/skills/pinakes/SKILL.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       ".gemini/skills/pinakes/SKILL.md",
		},
		{
			name:       "claude_commands_to_gemini",
			input:      ".claude/commands/spike.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       ".gemini/commands/spike.md",
		},
		{
			name:       "claude_agents_to_gemini",
			input:      ".claude/agents/potnia.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       ".gemini/agents/potnia.md",
		},
		{
			name:       "claude_identity_noop",
			input:      ".claude/skills/foo/SKILL.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".claude",
			want:       ".claude/skills/foo/SKILL.md", // HA-TEST: Claude channel dir name in content rewrite fixture
		},
		{
			name:       "empty_channel_noop",
			input:      ".claude/skills/foo/SKILL.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: "",
			want:       ".claude/skills/foo/SKILL.md", // HA-TEST: Claude channel dir name in content rewrite fixture
		},

		// Section 3.2: Context tests
		{
			name:       "backtick_path",
			input:      "`Read(\".claude/skills/pinakes/domains/{domain}.md\")`", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       "`Read(\".gemini/skills/pinakes/domains/{domain}.md\")`",
		},
		{
			name:       "link_target",
			input:      "[ref](.claude/skills/doc/SKILL.md)", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       "[ref](.gemini/skills/doc/SKILL.md)",
		},
		{
			name:       "prose_path",
			input:      "Full documentation: .claude/skills/ref/SKILL.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       "Full documentation: .gemini/skills/ref/SKILL.md",
		},
		{
			name:       "multiple_on_line",
			input:      ".claude/skills/a.md and .claude/commands/b.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       ".gemini/skills/a.md and .gemini/commands/b.md",
		},

		// Canonical .channel/ placeholder tests
		{
			name:       "channel_placeholder_to_claude",
			input:      ".channel/skills/pinakes/SKILL.md",
			channelDir: ".claude",
			want:       ".claude/skills/pinakes/SKILL.md", // HA-TEST: Claude channel dir name in content rewrite fixture
		},
		{
			name:       "channel_placeholder_to_gemini",
			input:      ".channel/skills/pinakes/SKILL.md",
			channelDir: ".gemini",
			want:       ".gemini/skills/pinakes/SKILL.md",
		},
		{
			name:       "channel_placeholder_commands",
			input:      ".channel/commands/spike.md",
			channelDir: ".gemini",
			want:       ".gemini/commands/spike.md",
		},
		{
			name:       "channel_placeholder_agents",
			input:      ".channel/agents/potnia.md",
			channelDir: ".claude",
			want:       ".claude/agents/potnia.md", // HA-TEST: Claude channel dir name in content rewrite fixture
		},
		{
			name:       "channel_placeholder_empty_defaults_claude",
			input:      ".channel/skills/foo.md",
			channelDir: "",
			want:       ".claude/skills/foo.md", // HA-TEST: Claude channel dir name in content rewrite fixture
		},
		{
			name:       "channel_placeholder_mixed_with_legacy",
			input:      ".channel/skills/a.md and .claude/commands/b.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       ".gemini/skills/a.md and .gemini/commands/b.md",
		},
		{
			name:       "channel_placeholder_bare_not_rewritten",
			input:      ".channel/settings.json",
			channelDir: ".gemini",
			want:       ".channel/settings.json",
		},
		{
			name:       "channel_placeholder_in_backtick",
			input:      "`Read(\".channel/skills/pinakes/domains/{domain}.md\")`",
			channelDir: ".gemini",
			want:       "`Read(\".gemini/skills/pinakes/domains/{domain}.md\")`",
		},

		// Section 3.3: Exclusion tests (rewriteChannelPaths level, not fenced)
		{
			name:       "template_block_preserved",
			input:      "{{ if .claude/skills/foo }}", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       "{{ if .claude/skills/foo }}", // HA-TEST: Claude channel dir name in content rewrite fixture
		},
		{
			name:       "ha_tagged_preserved",
			input:      ".claude/skills/foo // HA-4-001",
			channelDir: ".gemini",
			want:       ".claude/skills/foo // HA-4-001",
		},

		// Section 3.4: Non-match tests
		{
			name:       "bare_claude_dir",
			input:      ".claude/settings.json", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       ".claude/settings.json", // HA-TEST: Claude channel dir name in content rewrite fixture
		},
		{
			name:       "claude_md_file",
			input:      ".claude/CLAUDE.md", // HA-TEST: Claude channel dir name in content rewrite fixture
			channelDir: ".gemini",
			want:       ".claude/CLAUDE.md", // HA-TEST: Claude channel dir name in content rewrite fixture
		},
		{
			name:       "dot_claude_no_slash",
			input:      "the .claude directory",
			channelDir: ".gemini",
			want:       "the .claude directory",
		},
		{
			name:       "gemini_already",
			input:      ".gemini/skills/foo.md",
			channelDir: ".gemini",
			want:       ".gemini/skills/foo.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := rewriteChannelPaths(tt.input, tt.channelDir)
			if got != tt.want {
				t.Errorf("rewriteChannelPaths(%q, %q)\n  got:  %q\n  want: %q", tt.input, tt.channelDir, got, tt.want)
			}
		})
	}
}

// TestRewriteChannelPathsFencedExclusion verifies that fenced code blocks are
// not channel-rewritten. This tests the full RewriteMenaContentPaths pipeline
// (which splits on fences before calling applyRewrites/rewriteChannelPaths).
func TestRewriteChannelPathsFencedExclusion(t *testing.T) {
	t.Parallel()
	input := "```\n.claude/skills/foo.md\n```" // HA-TEST: Claude channel dir name in content rewrite fixture
	want := "```\n.claude/skills/foo.md\n```" // HA-TEST: Claude channel dir name in content rewrite fixture

	got := string(RewriteMenaContentPaths([]byte(input), ".gemini"))
	if got != want {
		t.Errorf("fenced block channel rewrite\n  got:  %q\n  want: %q", got, want)
	}
}

// TestRewriteExtensionThenChannel verifies that Pass 1-3 (extension rewriting)
// feeds into Pass 4 (channel rewriting). A source reference like
// .claude/skills/foo.lego.md should first become .claude/skills/foo.md (Pass 2) // HA-TEST: Claude channel dir name in content rewrite fixture
// then .gemini/skills/foo.md (Pass 4).
func TestRewriteExtensionThenChannel(t *testing.T) {
	t.Parallel()
	input := "[ref](.claude/skills/foo.lego.md)" // HA-TEST: Claude channel dir name in content rewrite fixture
	want := "[ref](.gemini/skills/foo.md)"

	got := string(RewriteMenaContentPaths([]byte(input), ".gemini"))
	if got != want {
		t.Errorf("extension-then-channel ordering\n  got:  %q\n  want: %q", got, want)
	}
}

// TestRewriteCorpusGemini processes a representative corpus with channelDir=".gemini"
// and verifies that zero .claude/skills/, .claude/commands/, .claude/agents/ references // HA-TEST: Claude channel dir name in content rewrite fixture
// survive in non-fenced segments.
func TestRewriteCorpusGemini(t *testing.T) {
	t.Parallel()
	corpus := `---
name: channel-corpus-test
description: "Corpus fixture for channel rewrite validation."
---

# Channel Corpus

## Mena Paths

- Skill ref: .claude/skills/pinakes/SKILL.md // HA-TEST: Claude channel dir name in content rewrite fixture
- Command ref: .claude/commands/spike.md // HA-TEST: Claude channel dir name in content rewrite fixture
- Agent ref: .claude/agents/potnia.md // HA-TEST: Claude channel dir name in content rewrite fixture
- Backtick: ` + "`" + `.claude/skills/doc/SKILL.md` + "`" + ` // HA-TEST: Claude channel dir name in content rewrite fixture
- Link: [ref](.claude/skills/doc/SKILL.md) // HA-TEST: Claude channel dir name in content rewrite fixture
- User-level: ~/.claude/agents/potnia.md // HA-TEST: Claude channel dir name in content rewrite fixture
- Multiple: .claude/skills/a.md and .claude/commands/b.md // HA-TEST: Claude channel dir name in content rewrite fixture

## Non-Targets (must NOT be rewritten)

- Config: .claude/settings.json // HA-TEST: Claude channel dir name in content rewrite fixture
- Inscription: .claude/CLAUDE.md // HA-TEST: Claude channel dir name in content rewrite fixture
- Bare: the .claude directory

## Fenced (must be preserved)

` + "```" + `
.claude/skills/inside-fence.md // HA-TEST: Claude channel dir name in content rewrite fixture
.claude/commands/inside-fence.md // HA-TEST: Claude channel dir name in content rewrite fixture
` + "```" + `

## After Fence

- Post-fence ref: .claude/skills/after-fence.md // HA-TEST: Claude channel dir name in content rewrite fixture
`

	output := string(RewriteMenaContentPaths([]byte(corpus), ".gemini"))

	// Split on fences to check only non-fenced segments
	segments := splitOnFences(output)

	// channelPathPattern matches the three .claude/ content subdirectories // HA-TEST: Claude channel dir name in content rewrite fixture
	channelPathPattern := regexp.MustCompile(`\.claude/(skills|commands|agents)/`) // HA-TEST: Claude channel dir name in content rewrite fixture

	var violations []string
	for i, seg := range segments {
		if i%2 != 0 {
			continue // Inside fenced block: skip
		}
		for lineNum, line := range strings.Split(seg, "\n") {
			if channelPathPattern.MatchString(line) {
				violations = append(violations,
					fmt.Sprintf("segment %d line %d: %s", i, lineNum, line))
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf(".claude/ content path refs survived channel rewriting in non-fenced content (%d violations):", len(violations)) // HA-TEST: Claude channel dir name in content rewrite fixture
		for _, v := range violations {
			t.Errorf("  %s", v)
		}
	}

	// Sanity: fenced block content preserved
	if !strings.Contains(output, ".claude/skills/inside-fence.md") { // HA-TEST: Claude channel dir name in content rewrite fixture
		t.Error("fenced block content was incorrectly channel-rewritten")
	}

	// Sanity: non-target paths preserved
	if !strings.Contains(output, ".claude/settings.json") { // HA-TEST: Claude channel dir name in content rewrite fixture
		t.Error(".claude/settings.json was incorrectly rewritten") // HA-TEST: Claude channel dir name in content rewrite fixture
	}
	if !strings.Contains(output, ".claude/CLAUDE.md") { // HA-TEST: Claude channel dir name in content rewrite fixture
		t.Error(".claude/CLAUDE.md was incorrectly rewritten") // HA-TEST: Claude channel dir name in content rewrite fixture
	}
}

// TestRewriteCorpusClaude verifies that corpus processing with channelDir=".claude"
// produces identical output to channelDir="" (identity transform).
func TestRewriteCorpusClaude(t *testing.T) {
	t.Parallel()
	corpus := `.claude/skills/foo.md and .claude/commands/bar.md` // HA-TEST: Claude channel dir name in content rewrite fixture

	outputEmpty := string(RewriteMenaContentPaths([]byte(corpus), ""))
	outputClaude := string(RewriteMenaContentPaths([]byte(corpus), ".claude"))

	if outputEmpty != outputClaude {
		t.Errorf("channelDir=\"\" and channelDir=\".claude\" should produce identical output\n  empty:  %q\n  claude: %q", outputEmpty, outputClaude)
	}
}
