package mena

import (
	"testing"
)

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
			got := string(rewriteMenaContentPaths([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("rewriteMenaContentPaths(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.want)
			}
		})
	}
}
