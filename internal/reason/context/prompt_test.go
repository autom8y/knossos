package context

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/autom8y/knossos/internal/trust"
)

// ---- BC-14: Zero-source gap admission tests ----

func TestRenderSystemPrompt_WithSources(t *testing.T) {
	sources := []SourceMaterial{
		{
			QualifiedName:  "autom8y::knossos::architecture",
			Content:        "Architecture description",
			Freshness:      0.9,
			FreshnessLabel: "fresh",
			Repo:           "knossos",
		},
	}

	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, sources)

	assert.Contains(t, prompt, "KNOWLEDGE SOURCES", "should include source material section")
	assert.Contains(t, prompt, "architecture", "should include source content")
	assert.NotContains(t, prompt, "NO KNOWLEDGE SOURCES", "should NOT include gap admission")
}

func TestRenderSystemPrompt_ZeroSources_GapAdmission(t *testing.T) {
	// BC-14: Sonnet prompt MUST handle len(SourceBlocks) == 0 explicitly.
	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil)

	assert.Contains(t, prompt, "NO KNOWLEDGE SOURCES",
		"BC-14: zero sources must trigger gap admission instructions")
	assert.Contains(t, prompt, "do not have information",
		"BC-14: must instruct model to acknowledge gap")
	assert.Contains(t, prompt, "Do NOT fabricate",
		"BC-14: must instruct model not to hallucinate")
}

func TestRenderSystemPrompt_EmptySliceSources_GapAdmission(t *testing.T) {
	// Empty slice (as opposed to nil) should also trigger gap admission.
	prompt := RenderSystemPrompt("autom8y", trust.TierMedium, []SourceMaterial{})

	assert.Contains(t, prompt, "NO KNOWLEDGE SOURCES",
		"BC-14: empty slice sources must also trigger gap admission")
}

func TestRenderSystemPrompt_AlwaysHasIdentity(t *testing.T) {
	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil)

	assert.Contains(t, prompt, "Clew", "should always include identity section")
	assert.Contains(t, prompt, "autom8y", "should include org name")
}

func TestRenderSystemPrompt_AlwaysHasTierBehavior(t *testing.T) {
	tiers := []trust.ConfidenceTier{trust.TierHigh, trust.TierMedium, trust.TierLow}

	for _, tier := range tiers {
		prompt := RenderSystemPrompt("autom8y", tier, nil)
		assert.True(t, strings.Contains(prompt, "Confidence:"),
			"tier %s: should include tier behavior section", tier.String())
	}
}

func TestRenderNoSourcesGapAdmission_Content(t *testing.T) {
	result := renderNoSourcesGapAdmission()

	assert.Contains(t, result, "NO KNOWLEDGE SOURCES AVAILABLE")
	assert.Contains(t, result, "MUST")
	assert.Contains(t, result, "Do NOT fabricate")
}

// ---- T3: Citation provenance rule tests ----

func TestRenderTierBehavior_HighTier_CitationProvenanceRule(t *testing.T) {
	result := renderTierBehavior(trust.TierHigh)

	assert.Contains(t, result, "CITATION PROVENANCE RULE",
		"HIGH tier must include citation provenance instruction")
	assert.Contains(t, result, "Only cite sources that are explicitly listed",
		"must instruct Claude to only cite listed sources")
	assert.Contains(t, result, "Do NOT cite domains",
		"must warn against citing referenced domains from within content")
}

func TestRenderTierBehavior_MediumTier_CitationProvenanceRule(t *testing.T) {
	result := renderTierBehavior(trust.TierMedium)

	assert.Contains(t, result, "CITATION PROVENANCE RULE",
		"MEDIUM tier must include citation provenance instruction")
	assert.Contains(t, result, "Only cite sources that are explicitly listed",
		"must instruct Claude to only cite listed sources")
}

func TestRenderTierBehavior_LowTier_NoCitationRule(t *testing.T) {
	result := renderTierBehavior(trust.TierLow)

	assert.NotContains(t, result, "CITATION PROVENANCE RULE",
		"LOW tier should not include citation provenance rule (never reaches Claude)")
}
