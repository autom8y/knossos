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

	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, sources, "")

	assert.Contains(t, prompt, "KNOWLEDGE SOURCES", "should include source material section")
	assert.Contains(t, prompt, "architecture", "should include source content")
	assert.NotContains(t, prompt, "NO KNOWLEDGE SOURCES", "should NOT include gap admission")
}

func TestRenderSystemPrompt_ZeroSources_GapAdmission(t *testing.T) {
	// BC-14: Sonnet prompt MUST handle len(SourceBlocks) == 0 explicitly.
	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil, "")

	assert.Contains(t, prompt, "NO KNOWLEDGE SOURCES",
		"BC-14: zero sources must trigger gap admission instructions")
	assert.Contains(t, prompt, "do not have information",
		"BC-14: must instruct model to acknowledge gap")
	assert.Contains(t, prompt, "Do NOT fabricate",
		"BC-14: must instruct model not to hallucinate")
}

func TestRenderSystemPrompt_EmptySliceSources_GapAdmission(t *testing.T) {
	// Empty slice (as opposed to nil) should also trigger gap admission.
	prompt := RenderSystemPrompt("autom8y", trust.TierMedium, []SourceMaterial{}, "")

	assert.Contains(t, prompt, "NO KNOWLEDGE SOURCES",
		"BC-14: empty slice sources must also trigger gap admission")
}

func TestRenderSystemPrompt_AlwaysHasIdentity(t *testing.T) {
	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil, "")

	assert.Contains(t, prompt, "Clew", "should always include identity section")
	assert.Contains(t, prompt, "autom8y", "should include org name")
}

func TestRenderSystemPrompt_AlwaysHasTierBehavior(t *testing.T) {
	tiers := []trust.ConfidenceTier{trust.TierHigh, trust.TierMedium, trust.TierLow}

	for _, tier := range tiers {
		prompt := RenderSystemPrompt("autom8y", tier, nil, "")
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

// ---- WS-2: Conversation history tests ----

func TestRenderSystemPrompt_WithConversationHistory(t *testing.T) {
	history := []ConversationTurn{
		{Role: "user", Content: "How is knossos structured?"},
		{Role: "assistant", Content: "Knossos is organized as a monorepo with two subsystems."},
		{Role: "user", Content: "What about the trust system?"},
	}
	sources := []SourceMaterial{
		{
			QualifiedName:  "autom8y::knossos::architecture",
			Content:        "Architecture description",
			Freshness:      0.9,
			FreshnessLabel: "fresh",
			Repo:           "knossos",
		},
	}

	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, sources, "", history)

	// Must contain CONVERSATION HISTORY section.
	assert.Contains(t, prompt, "CONVERSATION HISTORY",
		"WS-2: must include conversation history section when history is provided")
	assert.Contains(t, prompt, "[User]: How is knossos structured?",
		"WS-2: must format user turns as [User]: content")
	assert.Contains(t, prompt, "[Assistant]: Knossos is organized as",
		"WS-2: must format assistant turns as [Assistant]: content")
	assert.Contains(t, prompt, "[User]: What about the trust system?",
		"WS-2: must include all turns in order")

	// Conversation history must appear BEFORE source material.
	// Use the actual section delimiters (not just the words, which also appear
	// in the citation provenance rule within the tier behavior section).
	historyIdx := strings.Index(prompt, "--- CONVERSATION HISTORY ---")
	sourcesIdx := strings.Index(prompt, "--- KNOWLEDGE SOURCES ---")
	assert.Greater(t, historyIdx, 0, "must find conversation history section header")
	assert.Greater(t, sourcesIdx, 0, "must find knowledge sources section header")
	assert.Less(t, historyIdx, sourcesIdx,
		"WS-2: conversation history must appear before source material")
}

func TestRenderSystemPrompt_BackwardCompatibility_NoHistory(t *testing.T) {
	// Calling without the conversation history parameter must produce identical output
	// to the pre-WS-2 behavior.
	sources := []SourceMaterial{
		{
			QualifiedName:  "autom8y::knossos::architecture",
			Content:        "Architecture description",
			Freshness:      0.9,
			FreshnessLabel: "fresh",
			Repo:           "knossos",
		},
	}

	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, sources, "")

	assert.NotContains(t, prompt, "CONVERSATION HISTORY",
		"WS-2 backward compat: no history param must not include conversation history section")
	assert.Contains(t, prompt, "KNOWLEDGE SOURCES",
		"WS-2 backward compat: source material must still be present")
	assert.Contains(t, prompt, "Clew",
		"WS-2 backward compat: identity must still be present")
}

func TestRenderSystemPrompt_EmptyConversationHistory(t *testing.T) {
	// Empty slice should not produce a conversation history section.
	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil, "", []ConversationTurn{})

	assert.NotContains(t, prompt, "CONVERSATION HISTORY",
		"WS-2: empty conversation history must not render a section")
}

func TestRenderSystemPrompt_NilConversationHistory(t *testing.T) {
	// Nil history slice (passed explicitly) should not produce a section.
	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil, "", nil)

	assert.NotContains(t, prompt, "CONVERSATION HISTORY",
		"WS-2: nil conversation history must not render a section")
}

func TestRenderSystemPrompt_SingleTurnHistory(t *testing.T) {
	history := []ConversationTurn{
		{Role: "user", Content: "Hello"},
	}

	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil, "", history)

	assert.Contains(t, prompt, "CONVERSATION HISTORY",
		"WS-2: single-turn history must render section")
	assert.Contains(t, prompt, "[User]: Hello",
		"WS-2: single-turn must be formatted correctly")
}

func TestRenderConversationHistory_Formatting(t *testing.T) {
	turns := []ConversationTurn{
		{Role: "user", Content: "First question"},
		{Role: "assistant", Content: "First answer"},
		{Role: "user", Content: "Follow-up question"},
		{Role: "assistant", Content: "Follow-up answer"},
	}

	result := renderConversationHistory(turns)

	assert.Contains(t, result, "--- CONVERSATION HISTORY ---",
		"must start with section header")
	assert.Contains(t, result, "[User]: First question",
		"must format user turns")
	assert.Contains(t, result, "[Assistant]: First answer",
		"must format assistant turns")

	// Verify ordering: first question appears before follow-up.
	firstIdx := strings.Index(result, "First question")
	followupIdx := strings.Index(result, "Follow-up question")
	assert.Less(t, firstIdx, followupIdx,
		"turns must appear in chronological order")
}

func TestRenderConversationHistory_RoleCapitalization(t *testing.T) {
	turns := []ConversationTurn{
		{Role: "user", Content: "test"},
		{Role: "assistant", Content: "test"},
		{Role: "system", Content: "test"}, // Unexpected role, should pass through.
	}

	result := renderConversationHistory(turns)

	assert.Contains(t, result, "[User]:", "user role must be capitalized")
	assert.Contains(t, result, "[Assistant]:", "assistant role must be capitalized")
	assert.Contains(t, result, "[system]:", "unknown roles pass through unchanged")
}

// ---- ADR-TOPO: Org topology tests ----

func TestRenderSystemPrompt_WithTopology(t *testing.T) {
	topology := "--- ORG TOPOLOGY ---\n\nOrganization: autom8y (10 repos, ~130 knowledge domains)\n"

	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil, topology)

	assert.Contains(t, prompt, "--- ORG TOPOLOGY ---",
		"ADR-TOPO: must include topology section when provided")
	assert.Contains(t, prompt, "autom8y (10 repos",
		"ADR-TOPO: must include topology content")

	// Topology must appear after tier behavior.
	tierIdx := strings.Index(prompt, "Confidence: HIGH")
	topoIdx := strings.Index(prompt, "--- ORG TOPOLOGY ---")
	assert.Less(t, tierIdx, topoIdx,
		"ADR-TOPO-4: topology must appear after tier behavior")
}

func TestRenderSystemPrompt_EmptyTopology(t *testing.T) {
	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, nil, "")

	assert.NotContains(t, prompt, "ORG TOPOLOGY",
		"ADR-TOPO: empty topology string must not render a topology section")
}

func TestRenderSystemPrompt_TopologyOrdering(t *testing.T) {
	topology := "--- ORG TOPOLOGY ---\n\nOrganization: autom8y\n"
	history := []ConversationTurn{
		{Role: "user", Content: "test question"},
	}
	sources := []SourceMaterial{
		{
			QualifiedName:  "autom8y::knossos::arch",
			Content:        "content",
			Freshness:      0.9,
			FreshnessLabel: "fresh",
		},
	}

	prompt := RenderSystemPrompt("autom8y", trust.TierHigh, sources, topology, history)

	// Verify ordering: identity -> tier -> topology -> conversation -> sources.
	identityIdx := strings.Index(prompt, "Clew")
	tierIdx := strings.Index(prompt, "Confidence: HIGH")
	topoIdx := strings.Index(prompt, "--- ORG TOPOLOGY ---")
	historyIdx := strings.Index(prompt, "--- CONVERSATION HISTORY ---")
	sourcesIdx := strings.Index(prompt, "--- KNOWLEDGE SOURCES ---")

	assert.Less(t, identityIdx, tierIdx, "identity before tier")
	assert.Less(t, tierIdx, topoIdx, "tier before topology")
	assert.Less(t, topoIdx, historyIdx, "topology before conversation history")
	assert.Less(t, historyIdx, sourcesIdx, "conversation history before sources")
}
