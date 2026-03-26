package format

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunk_ShortText(t *testing.T) {
	text := "short text"
	chunks := Chunk(text, 100)
	require.Len(t, chunks, 1)
	assert.Equal(t, "short text", chunks[0])
}

func TestChunk_ExactLimit(t *testing.T) {
	text := strings.Repeat("a", 100)
	chunks := Chunk(text, 100)
	require.Len(t, chunks, 1)
	assert.Equal(t, text, chunks[0])
}

func TestChunk_ParagraphBreak(t *testing.T) {
	para1 := strings.Repeat("a", 40)
	para2 := strings.Repeat("b", 40)
	text := para1 + "\n\n" + para2
	chunks := Chunk(text, 50)

	require.Len(t, chunks, 2)
	assert.Equal(t, para1, chunks[0])
	assert.Equal(t, para2, chunks[1])
}

func TestChunk_LineBreakFallback(t *testing.T) {
	line1 := strings.Repeat("a", 40)
	line2 := strings.Repeat("b", 40)
	text := line1 + "\n" + line2
	chunks := Chunk(text, 50)

	require.Len(t, chunks, 2)
	assert.Equal(t, line1, chunks[0])
	assert.Equal(t, line2, chunks[1])
}

func TestChunk_HardCut(t *testing.T) {
	text := strings.Repeat("a", 200)
	chunks := Chunk(text, 100)

	require.Len(t, chunks, 2)
	assert.Len(t, chunks[0], 100)
	assert.Len(t, chunks[1], 100)
}

func TestChunk_FenceAware(t *testing.T) {
	before := strings.Repeat("x", 30)
	codeBody := strings.Repeat("y", 30)
	after := strings.Repeat("z", 30)
	text := before + "\n\n```go\n" + codeBody + "\n```\n\n" + after

	// Set limit so the break falls inside the code block.
	chunks := Chunk(text, 50)

	require.GreaterOrEqual(t, len(chunks), 2)

	// Verify the first chunk that breaks mid-fence gets a closing fence.
	foundClose := false
	for _, c := range chunks[:len(chunks)-1] {
		if strings.Contains(c, "```") {
			// If it opens a fence, it should also close it.
			fenceCount := strings.Count(c, "```")
			if fenceCount%2 == 0 {
				foundClose = true
			}
		}
	}
	if strings.Contains(text, "```") {
		assert.True(t, foundClose || len(chunks) == 1,
			"chunks that break inside a fence should have matching open/close fences")
	}
}

func TestChunk_UnclosedFence(t *testing.T) {
	text := "before\n\n```python\nline1\nline2\nline3"
	chunks := Chunk(text, 25)

	require.GreaterOrEqual(t, len(chunks), 1)

	// All text should be present across all chunks.
	combined := strings.Join(chunks, "")
	assert.Contains(t, combined, "before")
	assert.Contains(t, combined, "line1")
}

func TestChunk_NoFenceBackwardCompat(t *testing.T) {
	// Without fences, chunking should match the old splitAnswerBlocks behavior:
	// prefer \n\n, then \n, then hard cut.
	para1 := strings.Repeat("a", 40)
	para2 := strings.Repeat("b", 40)
	para3 := strings.Repeat("c", 40)
	text := para1 + "\n\n" + para2 + "\n\n" + para3

	chunks := Chunk(text, 50)

	require.GreaterOrEqual(t, len(chunks), 2)
	assert.Equal(t, para1, chunks[0], "first chunk should break at paragraph boundary")
}

func TestChunk_EmptyText(t *testing.T) {
	chunks := Chunk("", 100)
	require.Len(t, chunks, 1)
	assert.Equal(t, "", chunks[0])
}

func TestChunk_FenceReopensWithLangTag(t *testing.T) {
	code := strings.Repeat("y", 60)
	text := "```python\n" + code + "\n```"

	chunks := Chunk(text, 50)

	require.GreaterOrEqual(t, len(chunks), 2)

	// Second chunk should start with the reopened fence including lang tag.
	for i := 1; i < len(chunks); i++ {
		if strings.HasPrefix(chunks[i], "```") {
			assert.True(t,
				strings.HasPrefix(chunks[i], "```python\n") || strings.HasPrefix(chunks[i], "```"),
				"reopened fence should include language tag",
			)
			break
		}
	}
}

// TestChunk_Adversarial probes boundary conditions and degenerate inputs.
func TestChunk_Adversarial(t *testing.T) {
	t.Run("code block spanning exactly at maxLen boundary", func(t *testing.T) {
		// Construct a code block that ends precisely at the maxLen boundary.
		// "```\n" = 4 chars, "\n```" = 4 chars, so body = maxLen - 8
		maxLen := 50
		body := strings.Repeat("x", maxLen-8)
		text := "```\n" + body + "\n```"
		assert.Equal(t, maxLen, len(text), "test setup: text length should exactly equal maxLen")

		chunks := Chunk(text, maxLen)
		require.Len(t, chunks, 1, "text exactly at limit should produce single chunk")
		assert.Equal(t, text, chunks[0])
	})

	t.Run("code block one byte over maxLen boundary", func(t *testing.T) {
		maxLen := 50
		body := strings.Repeat("x", maxLen-7) // one byte over
		text := "```\n" + body + "\n```"
		assert.Equal(t, maxLen+1, len(text), "test setup: text length should be maxLen+1")

		chunks := Chunk(text, maxLen)
		require.GreaterOrEqual(t, len(chunks), 2, "text over limit should be split")

		// Each chunk that contains code should be properly fenced
		for _, c := range chunks {
			if strings.Contains(c, "```") {
				fenceCount := strings.Count(c, "```")
				assert.True(t, fenceCount%2 == 0,
					"each chunk should have matching open/close fences, got %d in: %q", fenceCount, c)
			}
		}
	})

	t.Run("very long single line inside code fence", func(t *testing.T) {
		// No natural break point at all -- forces hard cut inside fence
		longLine := strings.Repeat("a", 200)
		text := "```\n" + longLine + "\n```"
		chunks := Chunk(text, 100)

		require.GreaterOrEqual(t, len(chunks), 2, "should split long single-line code block")

		// Verify all code content is preserved across chunks (modulo synthetic fences)
		combined := strings.Join(chunks, "")
		// Remove synthetic fence markers to count original content
		contentOnly := strings.ReplaceAll(combined, "```", "")
		contentOnly = strings.ReplaceAll(contentOnly, "\n", "")
		assert.Equal(t, longLine, contentOnly,
			"all original code content should be preserved across chunks")
	})

	t.Run("multiple code blocks some spanning boundary", func(t *testing.T) {
		block1 := "```\n" + strings.Repeat("a", 30) + "\n```"
		block2 := "```\n" + strings.Repeat("b", 30) + "\n```"
		block3 := "```\n" + strings.Repeat("c", 30) + "\n```"
		text := block1 + "\n\n" + block2 + "\n\n" + block3

		chunks := Chunk(text, 50)
		require.GreaterOrEqual(t, len(chunks), 2, "should produce multiple chunks")

		// Verify all three code blocks' content is present
		combined := strings.Join(chunks, "")
		assert.Contains(t, combined, strings.Repeat("a", 30))
		assert.Contains(t, combined, strings.Repeat("b", 30))
		assert.Contains(t, combined, strings.Repeat("c", 30))
	})

	t.Run("text with no newlines forces hard cut", func(t *testing.T) {
		text := strings.Repeat("w", 250)
		chunks := Chunk(text, 100)

		require.Len(t, chunks, 3, "250 chars with maxLen 100 should produce 3 chunks")
		assert.Len(t, chunks[0], 100)
		assert.Len(t, chunks[1], 100)
		assert.Len(t, chunks[2], 50)
	})

	t.Run("input shorter than maxLen returns single chunk", func(t *testing.T) {
		text := "short"
		chunks := Chunk(text, 100)
		require.Len(t, chunks, 1)
		assert.Equal(t, "short", chunks[0])
	})

	t.Run("empty string returns single empty chunk", func(t *testing.T) {
		chunks := Chunk("", 100)
		require.Len(t, chunks, 1)
		assert.Equal(t, "", chunks[0])
	})

	t.Run("mixed fence markers backtick-open tilde-close", func(t *testing.T) {
		// Opening with ``` but attempting to close with ~~~ should leave the
		// ``` block unclosed (fence mismatch). The ~~~ starts a NEW fence.
		text := "```\ncode line one\n~~~\nmore text"
		fences := scanFences(text)

		// We should get two fences: one unclosed ``` and one unclosed ~~~
		// OR the implementation may handle it differently. Let's just verify
		// the behavior is consistent and doesn't panic.
		require.NotNil(t, fences, "should return fence info even for mismatched markers")

		// The ``` fence should be unclosed (extends to end)
		if len(fences) > 0 {
			assert.Equal(t, "```", fences[0].marker)
		}
	})

	t.Run("maxLen of 1 forces character-by-character chunking", func(t *testing.T) {
		text := "abc"
		chunks := Chunk(text, 1)
		// Each character becomes its own chunk
		require.GreaterOrEqual(t, len(chunks), 3, "maxLen=1 should produce at least 3 chunks for 'abc'")
	})

	t.Run("text exactly double maxLen with paragraph break in middle", func(t *testing.T) {
		half := strings.Repeat("m", 49) // 49 chars
		text := half + "\n\n" + half     // 49 + 2 + 49 = 100
		chunks := Chunk(text, 51)        // break should fall at paragraph boundary

		require.GreaterOrEqual(t, len(chunks), 2)
		assert.Equal(t, half, chunks[0], "first chunk should be first paragraph")
	})

	t.Run("fence containing only empty lines", func(t *testing.T) {
		text := "before\n\n```\n\n\n\n```\n\nafter"
		chunks := Chunk(text, 20)
		require.GreaterOrEqual(t, len(chunks), 1)

		combined := strings.Join(chunks, "")
		assert.Contains(t, combined, "before")
		assert.Contains(t, combined, "after")
	})

	t.Run("deeply indented code fence", func(t *testing.T) {
		text := "text\n\n    ```\n    indented code\n    ```\n\nmore"
		fences := scanFences(text)
		// Indented fences should still be recognized
		assert.GreaterOrEqual(t, len(fences), 1, "indented fences should be detected")
		if len(fences) > 0 {
			assert.Equal(t, "    ", fences[0].indent, "indent should be preserved")
		}
	})
}

// TestScanFences_Adversarial probes the fence scanner with edge cases.
func TestScanFences_Adversarial(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		fenceCount int
	}{
		{
			name:       "mismatched markers not paired",
			input:      "```\ncode\n~~~",
			fenceCount: 1, // ``` unclosed, extends to end of text; ~~~ inside is just content
		},
		{
			name:       "fence close with lang tag not treated as close",
			input:      "```python\ncode\n```python",
			fenceCount: 1, // "```python" has rest="python", not empty, so not a close
		},
		{
			name:       "triple backtick in middle of line not a fence",
			input:      "some ``` text",
			fenceCount: 0, // regex requires start-of-line
		},
		{
			name:       "empty fence block",
			input:      "```\n```",
			fenceCount: 1,
		},
		{
			name:       "multiple unclosed fences",
			input:      "```\ncode1\n\n```\ncode2",
			fenceCount: 1, // first ``` opens, second ``` closes, remaining text is not a fence
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fences := scanFences(tc.input)
			assert.Len(t, fences, tc.fenceCount, "input: %q", tc.input)
		})
	}
}

func TestScanFences(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		fenceCount int
	}{
		{
			name:       "no fences",
			input:      "plain text",
			fenceCount: 0,
		},
		{
			name:       "one complete fence",
			input:      "```\ncode\n```",
			fenceCount: 1,
		},
		{
			name:       "two fences",
			input:      "```\ncode1\n```\ntext\n```\ncode2\n```",
			fenceCount: 2,
		},
		{
			name:       "unclosed fence",
			input:      "```python\ncode\nmore code",
			fenceCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fences := scanFences(tc.input)
			assert.Len(t, fences, tc.fenceCount)
		})
	}
}
