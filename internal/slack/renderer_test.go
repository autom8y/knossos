package slack

import (
	"testing"

	slackapi "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/trust"
)

func TestRenderHighConfidence(t *testing.T) {
	resp := &response.ReasoningResponse{
		Answer: "The architecture follows a 3-tier model.",
		Tier:   trust.TierHigh,
		Citations: []response.Citation{
			{QualifiedName: "autom8y::knossos::architecture", Section: "Layer Boundaries", Excerpt: "strict 3-tier layer model"},
			{QualifiedName: "autom8y::knossos::conventions", Section: "", Excerpt: "error handling style"},
		},
	}

	blocks := RenderHighConfidence(resp)

	require.GreaterOrEqual(t, len(blocks), 4, "expected at least 4 blocks: answer, divider, 2 citations, confidence")

	// First block: section with answer text
	assertSectionContains(t, blocks[0], "architecture follows a 3-tier model")

	// Second block: divider
	assert.IsType(t, &slackapi.DividerBlock{}, blocks[1])

	// Citation blocks: human-readable labels
	assertContextContains(t, blocks[2], "Architecture (knossos)")
	assertContextContains(t, blocks[2], "Layer Boundaries")
	assertContextContains(t, blocks[3], "Conventions (knossos)")

	// Last block: confidence indicator
	last := blocks[len(blocks)-1]
	assertContextContains(t, last, "High confidence")
	assertContextContains(t, last, ":large_green_circle:")
}

func TestRenderMediumConfidence(t *testing.T) {
	resp := &response.ReasoningResponse{
		Answer: "Based on available knowledge, the release process involves tagging.",
		Tier:   trust.TierMedium,
		Gap: &trust.GapAdmission{
			StaleDomains: []trust.StaleDomainInfo{
				{QualifiedName: "autom8y::knossos::release", Domain: "release", Freshness: 0.3, DaysSinceGenerated: 14},
			},
		},
		Citations: []response.Citation{
			{QualifiedName: "autom8y::knossos::release", Section: "Process", Excerpt: "tagging process"},
		},
	}

	blocks := RenderMediumConfidence(resp)

	require.GreaterOrEqual(t, len(blocks), 5, "expected at least 5 blocks")

	// First block: answer
	assertSectionContains(t, blocks[0], "release process involves tagging")

	// Second block: staleness warning with human-readable domain name
	assertSectionContains(t, blocks[1], "Some sources may not be current")
	assertSectionContains(t, blocks[1], "Release (knossos)")

	// Divider
	assert.IsType(t, &slackapi.DividerBlock{}, blocks[2])

	// Citation: human-readable
	assertContextContains(t, blocks[3], "Release (knossos)")

	// Confidence indicator
	last := blocks[len(blocks)-1]
	assertContextContains(t, last, "Medium confidence")
	assertContextContains(t, last, ":large_yellow_circle:")
}

func TestRenderLowConfidence(t *testing.T) {
	resp := &response.ReasoningResponse{
		Answer: "insufficient knowledge to answer this question reliably",
		Tier:   trust.TierLow,
		Gap: &trust.GapAdmission{
			Reason:         "no knowledge found for: deployment",
			MissingDomains: []string{"deployment"},
			StaleDomains: []trust.StaleDomainInfo{
				{QualifiedName: "autom8y::knossos::release", Domain: "release", DaysSinceGenerated: 30},
			},
			Suggestions: []string{
				"Knowledge about \"deployment\" has not been generated yet. A developer can add it to the knowledge base.",
				"The release knowledge in knossos was last updated 30 days ago and may need to be refreshed.",
			},
		},
	}

	blocks := RenderLowConfidence(resp)

	require.GreaterOrEqual(t, len(blocks), 4, "expected at least header + reason + missing + stale + suggestions")

	// Header
	assertHeaderContains(t, blocks[0], "Cannot answer reliably")

	// Reason
	assertSectionContains(t, blocks[1], "no knowledge found for: deployment")

	// Missing domains
	assertSectionContains(t, blocks[2], "deployment")

	// Stale domains: human-readable format
	assertSectionContains(t, blocks[3], "Release (knossos)")
	assertSectionContains(t, blocks[3], "last updated 30 days ago")

	// Suggestions: no CLI commands
	assertSectionContains(t, blocks[4], "deployment")
	assertSectionContains(t, blocks[4], "has not been generated yet")
}

func TestRenderLowConfidence_NoGap(t *testing.T) {
	resp := &response.ReasoningResponse{
		Answer: "insufficient knowledge",
		Tier:   trust.TierLow,
		Gap:    nil,
	}

	blocks := RenderLowConfidence(resp)

	require.GreaterOrEqual(t, len(blocks), 2, "expected header + reason")
	assertHeaderContains(t, blocks[0], "Cannot answer reliably")
	assertSectionContains(t, blocks[1], "insufficient knowledge")
}

func TestRenderResponse_Dispatches(t *testing.T) {
	tests := []struct {
		name         string
		tier         trust.ConfidenceTier
		expectMarker string
	}{
		{"high", trust.TierHigh, "High confidence"},
		{"medium", trust.TierMedium, "Medium confidence"},
		{"low", trust.TierLow, "Cannot answer reliably"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &response.ReasoningResponse{
				Answer: "test answer",
				Tier:   tc.tier,
				Gap: &trust.GapAdmission{
					Reason: "test reason",
				},
			}
			blocks := RenderResponse(resp)
			require.NotEmpty(t, blocks)

			// Verify the expected marker appears somewhere in the blocks
			found := blocksContainText(blocks, tc.expectMarker)
			assert.True(t, found, "expected blocks to contain %q", tc.expectMarker)
		})
	}
}

func TestRenderHighConfidence_NoCitations(t *testing.T) {
	resp := &response.ReasoningResponse{
		Answer:    "A direct answer with no citations.",
		Tier:      trust.TierHigh,
		Citations: nil,
	}

	blocks := RenderHighConfidence(resp)

	// Should still have: answer, divider, confidence indicator
	require.GreaterOrEqual(t, len(blocks), 3)
	assertSectionContains(t, blocks[0], "direct answer")
	assert.IsType(t, &slackapi.DividerBlock{}, blocks[1])
	assertContextContains(t, blocks[len(blocks)-1], "High confidence")
}

func TestHumanReadableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard", "autom8y::knossos::architecture", "Architecture (knossos)"},
		{"hyphenated", "autom8y::knossos::test-coverage", "Test Coverage (knossos)"},
		{"feat domain", "autom8y::knossos::feat/materialization", "Feat/Materialization (knossos)"},
		{"not qualified", "just-a-name", "just-a-name"},
		{"two parts", "a::b", "a::b"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := humanReadableName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// --- test helpers ---

func assertSectionContains(t *testing.T, block slackapi.Block, substr string) {
	t.Helper()
	section, ok := block.(*slackapi.SectionBlock)
	require.True(t, ok, "expected SectionBlock, got %T", block)
	require.NotNil(t, section.Text, "section text is nil")
	assert.Contains(t, section.Text.Text, substr)
}

func assertContextContains(t *testing.T, block slackapi.Block, substr string) {
	t.Helper()
	ctx, ok := block.(*slackapi.ContextBlock)
	require.True(t, ok, "expected ContextBlock, got %T", block)
	found := false
	for _, elem := range ctx.ContextElements.Elements {
		if textObj, isText := elem.(*slackapi.TextBlockObject); isText {
			if containsStr(textObj.Text, substr) {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "context block does not contain %q", substr)
}

func assertHeaderContains(t *testing.T, block slackapi.Block, substr string) {
	t.Helper()
	header, ok := block.(*slackapi.HeaderBlock)
	require.True(t, ok, "expected HeaderBlock, got %T", block)
	require.NotNil(t, header.Text)
	assert.Contains(t, header.Text.Text, substr)
}

func blocksContainText(blocks []slackapi.Block, substr string) bool {
	for _, b := range blocks {
		switch bb := b.(type) {
		case *slackapi.SectionBlock:
			if bb.Text != nil && containsStr(bb.Text.Text, substr) {
				return true
			}
		case *slackapi.ContextBlock:
			for _, elem := range bb.ContextElements.Elements {
				if textObj, ok := elem.(*slackapi.TextBlockObject); ok {
					if containsStr(textObj.Text, substr) {
						return true
					}
				}
			}
		case *slackapi.HeaderBlock:
			if bb.Text != nil && containsStr(bb.Text.Text, substr) {
				return true
			}
		}
	}
	return false
}

func containsStr(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 &&
		// simple substring check
		len(haystack) >= len(needle) &&
		indexStr(haystack, needle) >= 0
}

func indexStr(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
