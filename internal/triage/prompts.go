package triage

import (
	"fmt"
	"strings"
)

// Stage 0: Multi-turn context resolution prompt.
// Input: current query + thread history.
// Output: refined query with implicit references resolved.
const stage0SystemPrompt = `You are a query refinement assistant. Your job is to resolve
implicit references in a follow-up question by examining the conversation history.

Given the conversation history and the current query, produce a SINGLE refined query
that is self-contained -- it should make sense without any conversation context.

Rules:
1. Resolve pronouns and implicit references ("that", "it", "the same", "compare to")
   using the conversation history.
2. Preserve the user's intent exactly -- do not add information they did not ask about.
3. If the query is already self-contained, return it unchanged.
4. Output ONLY the refined query text. No explanation, no prefix, no formatting.

Example:
History: User asked "How does the scheduling service handle retries?"
Current query: "Now compare that to ads"
Refined: "Compare the scheduling service retry handling to the ads service retry handling"`

// stage0UserMessage formats the Stage 0 user message with thread history.
func stage0UserMessage(currentQuery string, history []ThreadMessage) string {
	var b strings.Builder
	b.WriteString("Conversation history:\n")
	for _, msg := range history {
		b.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}
	b.WriteString(fmt.Sprintf("\nCurrent query: %s", currentQuery))
	return b.String()
}

// Stage 3: Haiku deep assessment prompt.
// Input: refined query + 20 candidate metadata entries.
// Output: JSON array of ranked winners with relevance scores.
const stage3SystemPrompt = `You are a domain relevance assessor for an organizational knowledge system.
Given a user query and a list of knowledge domain candidates with their metadata,
select the 3-5 most relevant domains and rank them by relevance.

For each selected domain, provide:
- qualified_name: the domain's canonical identifier
- relevance_score: a float between 0.0 and 1.0 (1.0 = perfect match)
- rationale: one sentence explaining why this domain is relevant
- domain_type: the domain's type from its metadata

Also classify the query intent:
- type: one of "architecture", "debugging", "comparison", "how-to", "exploration"
- target_domain_types: which domain types are most relevant
- repos: any repositories mentioned or implied

Rules:
1. Domain-type reasoning: "bugs" or "issues" -> scar-tissue. "how is X structured" -> architecture.
   "best practices" -> conventions. "what changed" or "history" -> release.
2. Cross-domain relationships: if domains share a repo or type, they may both be relevant.
3. Freshness matters: prefer fresher domains when relevance is otherwise equal.
4. Select 3-5 domains. Never select more than 5. Select fewer if the query is narrow.
5. Output ONLY valid JSON. No markdown, no explanation outside the JSON.

Output format:
{
  "candidates": [
    {"qualified_name": "...", "relevance_score": 0.95, "rationale": "...", "domain_type": "..."},
    ...
  ],
  "intent": {
    "type": "architecture",
    "target_domain_types": ["architecture", "conventions"],
    "repos": ["knossos"]
  }
}`

// stage3UserMessage formats the Stage 3 user message with query and candidates.
func stage3UserMessage(query string, candidates []candidateForLLM) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Query: %s\n\nCandidates:\n", query))
	for i, c := range candidates {
		b.WriteString(fmt.Sprintf("%d. %s (type: %s, repo: %s, freshness: %.2f)",
			i+1, c.QualifiedName, c.DomainType, c.Repo, c.FreshnessScore))
		if c.EmbeddingSimilarity > 0 {
			b.WriteString(fmt.Sprintf(", embedding_similarity: %.3f", c.EmbeddingSimilarity))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// candidateForLLM is the metadata sent to Haiku for Stage 3 assessment.
type candidateForLLM struct {
	QualifiedName       string
	DomainType          string
	Repo                string
	FreshnessScore      float64
	EmbeddingSimilarity float64
}
