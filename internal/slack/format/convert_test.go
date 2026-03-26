package format

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bold asterisks",
			input:    "this is **bold** text",
			expected: "this is *bold* text",
		},
		{
			name:     "bold underscores",
			input:    "this is __bold__ text",
			expected: "this is *bold* text",
		},
		{
			name:     "italic asterisks",
			input:    "this is *italic* text",
			expected: "this is _italic_ text",
		},
		{
			name:     "strikethrough",
			input:    "this is ~~struck~~ text",
			expected: "this is ~struck~ text",
		},
		{
			name:     "inline code passthrough",
			input:    "use `fmt.Println` here",
			expected: "use `fmt.Println` here",
		},
		{
			name:     "code block strips lang tag",
			input:    "```go\nfmt.Println(\"hi\")\n```",
			expected: "```\nfmt.Println(\"hi\")\n```",
		},
		{
			name:     "code block no lang tag",
			input:    "```\nsome code\n```",
			expected: "```\nsome code\n```",
		},
		{
			name:     "heading h1",
			input:    "# Main Title",
			expected: "*Main Title*",
		},
		{
			name:     "heading h2",
			input:    "## Section",
			expected: "*Section*",
		},
		{
			name:     "heading h3",
			input:    "### Subsection",
			expected: "*Subsection*",
		},
		{
			name:     "heading h6",
			input:    "###### Deep",
			expected: "*Deep*",
		},
		{
			name:     "blockquote passthrough",
			input:    "> quoted text",
			expected: "> quoted text",
		},
		{
			name:     "horizontal rule dashes",
			input:    "---",
			expected: "───\n",
		},
		{
			name:     "horizontal rule asterisks",
			input:    "***",
			expected: "───\n",
		},
		{
			name:     "horizontal rule underscores",
			input:    "___",
			expected: "───\n",
		},
		{
			name:     "image to alt text",
			input:    "![screenshot](https://example.com/img.png)",
			expected: "screenshot",
		},
		{
			name:     "link with label",
			input:    "[click here](https://example.com)",
			expected: "<https://example.com|click here>",
		},
		{
			name:     "link label equals url",
			input:    "[https://example.com](https://example.com)",
			expected: "https://example.com",
		},
		{
			name:     "code block content not converted",
			input:    "```\n**not bold** and *not italic*\n```",
			expected: "```\n**not bold** and *not italic*\n```",
		},
		{
			name:     "inline code content not converted",
			input:    "run `**this command**` now",
			expected: "run `**this command**` now",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "plain text with html entities",
			input:    "a < b & c > d",
			expected: "a &lt; b &amp; c &gt; d",
		},
		{
			name:     "mixed bold and code",
			input:    "**Important**: use `go test` to verify",
			expected: "*Important*: use `go test` to verify",
		},
		{
			name:     "spoiler pipe syntax dropped",
			input:    "this is ||spoiler|| text",
			expected: "this is spoiler text",
		},
		{
			name:     "spoiler reddit syntax dropped",
			input:    "this is >!spoiler!< text",
			expected: "this is spoiler text",
		},
		{
			name:     "tilde fenced code block",
			input:    "~~~python\nprint('hi')\n~~~",
			expected: "```\nprint('hi')\n```",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Convert(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestConvert_Adversarial probes edge cases in GFM-to-mrkdwn conversion.
func TestConvert_Adversarial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// --- Nested formatting ---
		{
			name:     "bold wrapping italic",
			input:    "**bold _italic_ bold**",
			expected: "*bold _italic_ bold*",
		},
		{
			name:     "bold wrapping italic asterisk",
			input:    "**bold *italic* bold**",
			expected: "*bold _italic_ bold*",
		},
		// --- Bold-italic collision ---
		//
		// DEFECT: "***text***" should produce "*_text_*" (bold wrapping italic)
		// but the non-greedy bold regex captures "**" + "*text*" + "**" = "*text*"
		// as the bold body, then the italic pass misaligns closing delimiters.
		// Actual: "*_text*_" — the closing _ lands after the bold *, breaking
		// Slack rendering. Slack would display literal asterisks and underscores.
		// Severity: MEDIUM — Claude frequently uses ***bold-italic*** in responses.
		{
			name:     "DEFECT triple asterisk bold-italic misaligned",
			input:    "***text***",
			expected: "*_text_*",
		},
		// --- Code block content must NOT be converted ---
		{
			name:     "code block with markdown syntax preserved",
			input:    "```\n**not bold** and *not italic* and ~~not struck~~\n```",
			expected: "```\n**not bold** and *not italic* and ~~not struck~~\n```",
		},
		{
			name:     "code block with heading preserved",
			input:    "```\n# Not a heading\n## Also not\n```",
			expected: "```\n# Not a heading\n## Also not\n```",
		},
		{
			name:     "code block with links preserved",
			input:    "```\n[not a link](http://example.com)\n```",
			expected: "```\n[not a link](http://example.com)\n```",
		},
		// --- Heading with formatting ---
		//
		// DEFECT: "## **Bold Heading**" — bold is converted to placeholder \x01B,
		// then the heading regex wraps in "*...*", then bold restore turns \x01B
		// to *, producing "**Bold Heading**". In Slack mrkdwn, "**text**" is not
		// valid formatting — it renders as literal "*" + bold "text" + literal "*".
		// Ideal output: "*Bold Heading*" (heading subsumes bold). The double-bold
		// collision produces garbage rendering in Slack.
		// Severity: MEDIUM — headings with bold text are common in Claude output.
		{
			name:     "DEFECT heading containing bold produces double-star collision",
			input:    "## **Bold Heading**",
			expected: "*Bold Heading*",
		},
		// --- Link with special chars in URL ---
		{
			name:     "link with ampersand in URL",
			input:    "[click](https://example.com/path?a=1&b=2)",
			expected: "<https://example.com/path?a=1&b=2|click>",
		},
		// --- Image edge cases ---
		{
			name:     "image with no alt text",
			input:    "![](https://example.com/img.png)",
			expected: "",
		},
		{
			name:     "image alt text with special chars",
			input:    "![a < b & c > d](https://example.com/img.png)",
			expected: "a &lt; b &amp; c &gt; d",
		},
		// --- Horizontal rules ---
		{
			name:     "multiple horizontal rules in sequence",
			input:    "---\n---\n---",
			expected: "───\n\n───\n\n───\n",
		},
		{
			name:     "horizontal rule with surrounding spaces",
			input:    "  - - -  ",
			expected: "───\n",
		},
		// --- Underscore italic at word boundary ---
		{
			name:     "underscore italic standard",
			input:    "this is _italic_ text",
			expected: "this is _italic_ text",
		},
		// --- Empty/degenerate formatting ---
		//
		// FINDING: "****" and "____" are consumed by the HR regex before bold
		// processing can see them. In GFM, "****" IS a valid thematic break,
		// so this is arguably correct. But if a user intended empty bold
		// (degenerate case), the HR interpretation wins silently.
		{
			name:     "empty bold asterisks treated as HR",
			input:    "****",
			expected: "───\n",
		},
		{
			name:     "empty bold underscores treated as HR",
			input:    "____",
			expected: "───\n",
		},
		//
		// DEFECT: GFM requires that opening * must NOT be followed by whitespace
		// and closing * must NOT be preceded by whitespace for emphasis.
		// Input "a * b * c" should NOT be italicized because of the spaces.
		// Actual: "a _ b _ c" (spaces inside italic delimiters).
		// Severity: LOW — unlikely in real Claude output but incorrect per spec.
		{
			name:     "DEFECT spaced asterisks incorrectly treated as italic",
			input:    "a * b * c",
			expected: "a * b * c",
		},
		{
			name:     "unclosed bold",
			input:    "**unclosed bold",
			expected: "**unclosed bold",
		},
		{
			name:     "unclosed italic",
			input:    "*unclosed italic",
			expected: "*unclosed italic",
		},
		// --- Inline code with special chars ---
		{
			name:     "inline code with angle brackets preserved",
			input:    "use `<div>` in HTML",
			expected: "use `<div>` in HTML",
		},
		// --- Mixed fences ---
		{
			name:     "tilde-opened backtick-closed treated as unclosed",
			input:    "~~~\ncode\n```",
			expected: "```\ncode\n```\n```",
		},
		{
			name:     "backtick-opened tilde-closed treated as unclosed",
			input:    "```\ncode\n~~~",
			expected: "```\ncode\n~~~\n```",
		},
		// --- Blockquote with formatting ---
		{
			name:     "blockquote with bold and links",
			input:    "> **Important**: see [docs](https://example.com)",
			expected: "> *Important*: see <https://example.com|docs>",
		},
		// --- Strikethrough edge cases ---
		{
			name:     "tilde not strikethrough with single tilde",
			input:    "~not struck~",
			expected: "~not struck~",
		},
		{
			name:     "strikethrough with nested bold",
			input:    "~~**bold struck**~~",
			expected: "~*bold struck*~",
		},
		// --- Spoiler edge cases ---
		{
			name:     "spoiler pipe with special chars",
			input:    "||secret <data>||",
			expected: "secret &lt;data&gt;",
		},
		// --- Only whitespace ---
		{
			name:     "only whitespace preserved",
			input:    "   \n   \n   ",
			expected: "   \n   \n   ",
		},
		// --- Code fence with lang tag containing special chars ---
		{
			name:     "code fence with complex lang tag stripped",
			input:    "```javascript{highlight=[1,3]}\nconsole.log('hi')\n```",
			expected: "```\nconsole.log('hi')\n```",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Convert(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestConvert_IntegrationPipeline tests a realistic multi-element Claude response
// going through the full Convert -> Chunk pipeline.
func TestConvert_IntegrationPipeline(t *testing.T) {
	input := `# Deployment Guide

Here is **important** info about the _deployment_ process.

Check the [documentation](https://docs.example.com/deploy?env=prod&region=us-east-1) for details.

---

## Code Example

` + "```go" + `
func main() {
    // This has **markdown** that should NOT be converted
    fmt.Println("<html>")
}
` + "```" + `

> **Note**: Always use ` + "`kubectl apply`" + ` before deploying.

Contact <@U123ABC> or visit <https://slack.com> for help.

---

That's all for the ~~old~~ *new* deployment process.`

	result := Convert(input)

	// Verify headings converted to bold
	assert.Contains(t, result, "*Deployment Guide*")
	assert.Contains(t, result, "*Code Example*")

	// Verify inline formatting
	assert.Contains(t, result, "*important*")
	assert.Contains(t, result, "_deployment_")

	// Verify link with query params
	assert.Contains(t, result, "<https://docs.example.com/deploy?env=prod&region=us-east-1|documentation>")

	// Verify horizontal rules
	assert.Contains(t, result, "───")

	// Verify code block content is NOT converted
	assert.Contains(t, result, "**markdown**")
	assert.Contains(t, result, `fmt.Println("<html>")`)

	// Verify inline code preserved
	assert.Contains(t, result, "`kubectl apply`")

	// Verify Slack tokens preserved
	assert.Contains(t, result, "<@U123ABC>")
	assert.Contains(t, result, "<https://slack.com>")

	// Verify strikethrough
	assert.Contains(t, result, "~old~")

	// Verify italic asterisk -> underscore
	assert.Contains(t, result, "_new_")

	// Verify the code fence lang tag is stripped
	assert.NotContains(t, result, "```go")

	// Now verify the full pipeline: Convert -> Chunk
	chunks := Chunk(result, 200)
	require.True(t, len(chunks) >= 1, "pipeline should produce at least one chunk")

	// Reassemble and verify no content loss
	combined := strings.Join(chunks, "\n")
	assert.Contains(t, combined, "*Deployment Guide*")
	assert.Contains(t, combined, "*Code Example*")
	assert.Contains(t, combined, "<@U123ABC>")
}

func TestConvert_MultiElement(t *testing.T) {
	input := `# Getting Started

Here is **important** info about the _deployment_ process.

Check the [docs](https://docs.example.com) for details.

---

## Code Example

` + "```go" + `
func main() {
    fmt.Println("hello")
}
` + "```" + `

That's it.`

	result := Convert(input)

	assert.Contains(t, result, "*Getting Started*")
	assert.Contains(t, result, "*important*")
	assert.Contains(t, result, "_deployment_")
	assert.Contains(t, result, "<https://docs.example.com|docs>")
	assert.Contains(t, result, "───")
	assert.Contains(t, result, "*Code Example*")
	assert.Contains(t, result, "```\nfunc main()")
	assert.NotContains(t, result, "```go")
}
