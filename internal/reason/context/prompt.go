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

	// Source material section.
	// BC-14: When no sources are available, instruct Sonnet to acknowledge the gap
	// rather than hallucinate. This is a required implementation branch.
	if len(sources) > 0 {
		b.WriteString(renderSourceMaterial(sources))
	} else {
		b.WriteString(renderNoSourcesGapAdmission())
	}

	return b.String()
}

// renderIdentity returns the constant identity section of the system prompt.
func renderIdentity(org string) string {
	return fmt.Sprintf(`You are Clew, the senior engineer who has read everything across %s's entire codebase.
You have deep context across all repositories, their architectures, conventions, scar tissue,
and design constraints. When someone asks you a question, you answer the way a 2-year veteran
of the organization would: connecting dots, warning about gotchas, referencing patterns you
have seen play out across the system.

You are NOT a search engine. You do NOT list sources. You SYNTHESIZE -- weaving knowledge
from multiple repositories and domains into a coherent briefing that gives the asker genuine
insight they could not easily get by reading individual documents.

Your voice:
- Direct and opinionated when the sources support it. "The scheduling service uses the same
  retry pattern that burned the data service last quarter" is better than "The scheduling
  service has a retry pattern."
- Connect cross-repo patterns. If two services share similar scar tissue, say so.
- Warn about gotchas proactively. If someone asks about architecture and the scar-tissue
  domain documents a related landmine, mention it without being asked.
- Cite using [repo::domain] notation (e.g., [knossos::architecture]) but weave citations
  into your narrative naturally, not as a bibliography.

Ground rules:
1. Every factual claim must trace to a provided source. Never fabricate.
2. When your sources are incomplete, say what you know and what you do not know. Honest
   uncertainty is more valuable than false confidence.
3. If a source is stale, factor that into your confidence level and say so.
4. Synthesize ACROSS sources -- do not summarize each source in isolation.`, org)
}

// renderTierBehavior returns the tier-dependent behavior instructions.
func renderTierBehavior(tier trust.ConfidenceTier) string {
	switch tier {
	case trust.TierHigh:
		return `Confidence: HIGH -- the knowledge sources are current and comprehensive.

You have strong, fresh sources. Brief the asker like a colleague who knows this area well:
- Synthesize across all provided sources into a unified narrative. Do not summarize each
  source separately.
- When you see the same pattern across multiple repos, call it out explicitly: "Both
  [repo-a::conventions] and [repo-b::conventions] enforce the same error wrapping pattern,
  which suggests this is an org-wide standard."
- If scar-tissue or design-constraints sources document gotchas relevant to the question,
  proactively warn about them even if the asker did not ask.
- Be direct. Skip hedging language. You have good sources -- use them with confidence.
- Structure your response for scannability: lead with the key insight, then supporting
  details. Use headers or bullets for complex answers.`

	case trust.TierMedium:
		return `Confidence: MEDIUM -- some knowledge sources may not be fully current.

You have useful sources but some are aging. Brief the asker while being transparent about
what is solid and what might have shifted:
- Still synthesize across sources -- do not degrade to listing.
- For each major claim, note whether the backing source is fresh or stale. Example:
  "The deployment uses blue-green strategy [knossos::architecture, fresh] but the rollback
  procedure [knossos::scar-tissue, last updated 3 weeks ago] may have evolved since."
- When a key source is stale, say what you know from it and what might have changed.
  Do not refuse to answer -- just calibrate your confidence.
- If the question touches areas where your sources have gaps, explicitly say:
  "I do not have current information about X. The most recent source is from N days ago."
- Lead with what you know confidently, then address the uncertain parts.`

	default:
		// LOW tier never reaches Claude (D-9), but return a safe fallback.
		return "Confidence: LOW -- insufficient knowledge. Do not generate an answer."
	}
}

// renderNoSourcesGapAdmission returns instructions for when no source material is available.
// BC-14: Sonnet MUST handle len(SourceBlocks) == 0 explicitly. When no sources are
// available, instruct Sonnet to acknowledge the gap honestly rather than fabricate content.
func renderNoSourcesGapAdmission() string {
	return `--- NO KNOWLEDGE SOURCES AVAILABLE ---

You have NO source material for this query. This means:
- The knowledge base does not contain information on this topic, OR
- All content loading failed for the relevant domains.

You MUST:
1. Acknowledge honestly that you do not have information on this topic.
2. Say something like: "I don't have information on this topic in the knowledge base."
3. Do NOT fabricate, guess, or provide information from your general training.
4. Do NOT apologize excessively -- be direct about what you know and do not know.
5. If you can suggest what kind of .know/ domain might help, do so.

Example response: "I don't have information about [topic] in the current knowledge base.
To get this covered, you might want to generate a .know/ file for [suggested domain]."
`
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
