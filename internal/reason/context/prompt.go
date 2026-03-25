package context

import (
	"fmt"
	"strings"

	"github.com/autom8y/knossos/internal/trust"
)

// RenderSystemPrompt assembles the complete system prompt from components.
// Concatenates: identity section + tier behavior section + few-shot examples + source material section.
func RenderSystemPrompt(org string, tier trust.ConfidenceTier, sources []SourceMaterial) string {
	var b strings.Builder

	// Identity section
	b.WriteString(renderIdentity(org))
	b.WriteString("\n\n")

	// Tier behavior section
	b.WriteString(renderTierBehavior(tier))
	b.WriteString("\n\n")

	// Source material section
	if len(sources) > 0 {
		b.WriteString(renderSourceMaterial(sources))
	}

	return b.String()
}

// renderIdentity returns the constant identity section of the system prompt.
func renderIdentity(org string) string {
	return fmt.Sprintf(`You are Clew, an organizational knowledge assistant for the %s organization.

Your purpose is to answer questions using verified knowledge from .know/ files across
the organization's repositories. You are a thread -- the clew of thread that guides
people through the labyrinth of organizational knowledge.

Rules:
1. NEVER fabricate information. Every factual claim must trace to a provided source.
2. When uncertain, say so explicitly. Uncertainty is trustworthy; false confidence is not.
3. Cite sources using [repo::domain] notation (e.g., [knossos::architecture]).
4. If a source is marked as stale, acknowledge this in your answer.
5. Keep answers concise and actionable. Prefer bullet points for multi-part answers.
6. Do not speculate beyond what the sources support.`, org)
}

// renderTierBehavior returns the tier-dependent behavior instructions.
func renderTierBehavior(tier trust.ConfidenceTier) string {
	switch tier {
	case trust.TierHigh:
		return `Confidence: HIGH -- the knowledge sources are current and comprehensive.

Behavior:
- Answer directly and confidently.
- Cite every factual claim with [repo::domain] notation.
- Do not add unnecessary caveats or hedging language.
- If the answer spans multiple sources, synthesize them into a coherent response.`

	case trust.TierMedium:
		return `Confidence: MEDIUM -- some knowledge sources may not be fully current.

Behavior:
- Answer the question, but note that some sources may not reflect the latest changes.
- For each claim, cite the source and note if it is marked as stale.
- Begin your response with: "Based on available knowledge (some sources may not be fully current):"
- If a key source is stale, explicitly note: "Note: The [repo::domain] source was last
  updated N days ago and may not reflect recent changes."`

	default:
		// LOW tier never reaches Claude (D-9), but return a safe fallback.
		return "Confidence: LOW -- insufficient knowledge. Do not generate an answer."
	}
}

// renderSourceMaterial formats the source material for inclusion in the system prompt.
func renderSourceMaterial(sources []SourceMaterial) string {
	var b strings.Builder
	b.WriteString("--- KNOWLEDGE SOURCES ---\n\n")

	for _, s := range sources {
		b.WriteString(fmt.Sprintf("Source: %s\n", s.QualifiedName))
		if s.Section != "" {
			b.WriteString(fmt.Sprintf("Section: %s\n", s.Section))
		}
		if !s.GeneratedAt.IsZero() {
			b.WriteString(fmt.Sprintf("Generated: %s\n", s.GeneratedAt.Format("2006-01-02")))
		}
		b.WriteString(fmt.Sprintf("Freshness: %.2f (%s)\n", s.Freshness, s.FreshnessLabel))
		if s.Repo != "" {
			b.WriteString(fmt.Sprintf("Repository: %s\n", s.Repo))
		}
		b.WriteString("---\n")
		b.WriteString(s.Content)
		b.WriteString("\n\n---\n\n")
	}

	return b.String()
}
