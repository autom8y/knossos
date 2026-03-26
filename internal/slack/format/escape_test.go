package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text unchanged",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "ampersand escaped",
			input:    "A & B",
			expected: "A &amp; B",
		},
		{
			name:     "less than escaped",
			input:    "a < b",
			expected: "a &lt; b",
		},
		{
			name:     "greater than escaped",
			input:    "a > b",
			expected: "a &gt; b",
		},
		{
			name:     "all three characters",
			input:    "<a & b>",
			expected: "&lt;a &amp; b&gt;",
		},
		{
			name:     "user mention preserved",
			input:    "Hey <@U123ABC> check this",
			expected: "Hey <@U123ABC> check this",
		},
		{
			name:     "channel ref preserved",
			input:    "See <#C123ABC>",
			expected: "See <#C123ABC>",
		},
		{
			name:     "special mention preserved",
			input:    "<!here> wake up",
			expected: "<!here> wake up",
		},
		{
			name:     "http link preserved",
			input:    "Visit <http://example.com>",
			expected: "Visit <http://example.com>",
		},
		{
			name:     "https link preserved",
			input:    "Visit <https://example.com>",
			expected: "Visit <https://example.com>",
		},
		{
			name:     "mailto preserved",
			input:    "Email <mailto:user@example.com>",
			expected: "Email <mailto:user@example.com>",
		},
		{
			name:     "tel preserved",
			input:    "Call <tel:+1234567890>",
			expected: "Call <tel:+1234567890>",
		},
		{
			name:     "slack deep link preserved",
			input:    "Open <slack://open>",
			expected: "Open <slack://open>",
		},
		{
			name:     "non-slack angle brackets escaped",
			input:    "<div>hello</div>",
			expected: "&lt;div&gt;hello&lt;/div&gt;",
		},
		{
			name:     "no double escape amp",
			input:    "already &amp; escaped",
			expected: "already &amp; escaped",
		},
		{
			name:     "no double escape lt",
			input:    "already &lt; escaped",
			expected: "already &lt; escaped",
		},
		{
			name:     "no double escape gt",
			input:    "already &gt; escaped",
			expected: "already &gt; escaped",
		},
		{
			name:     "blockquote prefix preserved",
			input:    "> some <html> & stuff",
			expected: "> some &lt;html&gt; &amp; stuff",
		},
		{
			name:     "blockquote with slack mention",
			input:    "> hey <@U123> check this",
			expected: "> hey <@U123> check this",
		},
		{
			name:     "multiline with blockquote",
			input:    "line one\n> quoted <b>\nline three",
			expected: "line one\n> quoted &lt;b&gt;\nline three",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "angle bracket without close",
			input:    "a < b and c",
			expected: "a &lt; b and c",
		},
		{
			name:     "mixed entities and raw",
			input:    "&amp; & &lt; < &gt; >",
			expected: "&amp; &amp; &lt; &lt; &gt; &gt;",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Escape(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestEscape_Adversarial probes edge cases that a malicious or confused user
// might produce, or that upstream GFM content could contain.
func TestEscape_Adversarial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// --- Near-miss Slack token attacks (must be escaped, not preserved) ---
		{
			name:     "javascript URI must be escaped",
			input:    "<javascript:alert(1)>",
			expected: "&lt;javascript:alert(1)&gt;",
		},
		{
			name:     "ftp URI must be escaped",
			input:    "<ftp://evil.com/payload>",
			expected: "&lt;ftp://evil.com/payload&gt;",
		},
		{
			name:     "data URI must be escaped",
			input:    "<data:text/html,<script>alert(1)</script>>",
			expected: "&lt;data:text/html,&lt;script&gt;alert(1)&lt;/script&gt;&gt;",
		},
		{
			name:     "file URI must be escaped",
			input:    "<file:///etc/passwd>",
			expected: "&lt;file:///etc/passwd&gt;",
		},
		// --- Double-escape prevention ---
		{
			name:     "pre-escaped amp in mixed context",
			input:    "A &amp; B & C",
			expected: "A &amp; B &amp; C",
		},
		{
			name:     "triple amp entity should not recurse",
			input:    "&amp;amp;",
			expected: "&amp;amp;",
		},
		// --- Empty and degenerate angle brackets ---
		{
			name:     "empty angle brackets must be escaped",
			input:    "<>",
			expected: "&lt;&gt;",
		},
		{
			name:     "single char non-token angle brackets",
			input:    "<x>",
			expected: "&lt;x&gt;",
		},
		{
			name:     "angle bracket with only space inside",
			input:    "< >",
			expected: "&lt; &gt;",
		},
		// --- Multiline inside angle brackets ---
		{
			name:     "newline inside angle bracket pair not treated as token",
			input:    "<\n@user>",
			expected: "&lt;\n@user&gt;",
		},
		// --- Blockquote edge cases ---
		{
			name:     "blockquote with slack token preserves both",
			input:    "> Check <@U123> for details",
			expected: "> Check <@U123> for details",
		},
		{
			name:     "blockquote with only gt no space",
			input:    ">no space after",
			expected: "&gt;no space after",
		},
		{
			name:     "nested blockquote markers",
			input:    "> > double quoted",
			expected: "> &gt; double quoted",
		},
		// --- Consecutive special chars ---
		{
			name:     "multiple ampersands in a row",
			input:    "&&&",
			expected: "&amp;&amp;&amp;",
		},
		{
			name:     "multiple less-thans in a row",
			input:    "<<<",
			expected: "&lt;&lt;&lt;",
		},
		{
			name:     "multiple greater-thans in a row",
			input:    ">>>",
			expected: "&gt;&gt;&gt;",
		},
		// --- Slack token adjacent to special chars ---
		{
			name:     "slack token immediately followed by ampersand",
			input:    "<@U123>&more",
			expected: "<@U123>&amp;more",
		},
		{
			name:     "two slack tokens back to back",
			input:    "<@U123><#C456>",
			expected: "<@U123><#C456>",
		},
		// --- Only special characters ---
		{
			name:     "only ampersand",
			input:    "&",
			expected: "&amp;",
		},
		{
			name:     "only less-than",
			input:    "<",
			expected: "&lt;",
		},
		{
			name:     "only greater-than",
			input:    ">",
			expected: "&gt;",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Escape(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestIsSlackToken_Adversarial probes near-miss token patterns.
func TestIsSlackToken_Adversarial(t *testing.T) {
	tests := []struct {
		name     string
		inner    string
		expected bool
	}{
		{"javascript scheme not a token", "javascript:alert(1)", false},
		{"ftp scheme not a token", "ftp://evil.com", false},
		{"data scheme not a token", "data:text/html,foo", false},
		{"file scheme not a token", "file:///etc/passwd", false},
		{"ws scheme not a token", "ws://example.com", false},
		{"wss scheme not a token", "wss://example.com", false},
		{"single at sign", "@", true},
		{"single hash", "#", true},
		{"single bang", "!", true},
		{"just http colon", "http:", false},
		{"http no slashes", "httpx://foo", false},
		{"https with caps", "HTTPS://example.com", false},
		{"space before prefix", " @U123", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isSlackToken(tc.inner))
		})
	}
}

func TestIsSlackToken(t *testing.T) {
	tests := []struct {
		name     string
		inner    string
		expected bool
	}{
		{"user mention", "@U123ABC", true},
		{"channel ref", "#C123ABC", true},
		{"special here", "!here", true},
		{"special channel", "!channel", true},
		{"http link", "http://example.com", true},
		{"https link", "https://example.com", true},
		{"mailto", "mailto:a@b.com", true},
		{"tel", "tel:+1234", true},
		{"slack deep link", "slack://open", true},
		{"plain text", "div", false},
		{"empty string", "", false},
		{"slash path", "/some/path", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isSlackToken(tc.inner))
		})
	}
}
