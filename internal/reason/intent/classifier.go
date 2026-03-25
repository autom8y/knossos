package intent

import (
	"strings"
)

// Classifier classifies user queries into action tiers and extracts domain hints.
// Uses keyword heuristics only -- no LLM calls.
type Classifier struct {
	// recordVerbs are action verbs that indicate Record intent.
	recordVerbs []string

	// actVerbs are action verbs that indicate Act intent.
	actVerbs []string

	// domainKeywords maps domain names to their trigger keywords.
	domainKeywords map[string][]string

	// observeGuards are query prefixes that force TierObserve
	// regardless of verb presence.
	observeGuards []string
}

// NewClassifier creates a Classifier with the default keyword sets.
func NewClassifier() *Classifier {
	return &Classifier{
		recordVerbs: []string{
			"update", "edit", "modify", "change", "add", "create",
			"write", "record", "file", "log", "note", "append",
			"amend", "revise", "correct",
		},
		actVerbs: []string{
			"run", "execute", "deploy", "hotfix", "rollback", "release",
			"revert", "dispatch", "trigger", "start", "stop", "restart",
			"build", "ship", "merge", "push",
		},
		domainKeywords: map[string][]string{
			"architecture":       {"architecture", "structure", "package", "layer", "import", "dependency", "module"},
			"conventions":        {"convention", "pattern", "idiom", "style", "naming", "error handling", "file organization"},
			"design-constraints": {"constraint", "frozen", "tension", "boundary", "invariant", "limitation"},
			"scar-tissue":        {"scar", "bug", "incident", "postmortem", "lesson", "defensive", "regression"},
			"test-coverage":      {"test", "coverage", "testing", "unit test", "integration test", "ci"},
			"feat/":              {"feature", "feat", "capability"},
			"release":            {"release", "version", "changelog", "deployment", "deploy", "ship"},
			"literature":         {"research", "paper", "literature", "study", "survey", "academic"},
		},
		observeGuards: []string{
			"how do i", "what is", "what are", "what does", "what was",
			"what were", "what will", "what would", "what can",
			"explain", "tell me about", "describe", "show me",
			"how does", "why does", "when does", "where does",
			"explain how to",
		},
	}
}

// Classify analyzes a query and returns an IntentResult with action tier and domain hints.
// Classification algorithm (priority order):
//  1. Check guard clauses -- if a guard prefix matches, force TierObserve.
//  2. Check Act verbs -- if any match, classify as TierAct.
//  3. Check Record verbs -- if any match, classify as TierRecord.
//  4. Default to TierObserve.
func (c *Classifier) Classify(query string) IntentResult {
	lower := strings.ToLower(query)
	tokens := tokenize(lower)

	// Step 1: Guard clause check (forces TierObserve regardless of verb presence).
	// This is the primary safety mechanism: "how do I deploy" -> Observe.
	if c.hasObserveGuard(lower) {
		return IntentResult{
			Tier:       TierObserve,
			Answerable: true,
			DomainHints: c.extractDomainHints(lower, tokens),
			RawQuery:   query,
		}
	}

	// Step 2: Act verbs (highest priority -- safety critical direction).
	if c.hasActVerb(tokens) {
		return IntentResult{
			Tier:              TierAct,
			Answerable:        false,
			UnsupportedReason: "Action execution (Tier 3) is not yet supported. Clew can answer knowledge questions about this topic.",
			DomainHints:       c.extractDomainHints(lower, tokens),
			RawQuery:          query,
		}
	}

	// Step 3: Record verbs.
	if c.hasRecordVerb(tokens) {
		return IntentResult{
			Tier:              TierRecord,
			Answerable:        false,
			UnsupportedReason: "Knowledge creation and updates (Tier 2) are not yet supported. Clew can answer knowledge questions about this topic.",
			DomainHints:       c.extractDomainHints(lower, tokens),
			RawQuery:          query,
		}
	}

	// Step 4: Default to Observe.
	return IntentResult{
		Tier:        TierObserve,
		Answerable:  true,
		DomainHints: c.extractDomainHints(lower, tokens),
		RawQuery:    query,
	}
}

// hasObserveGuard returns true if the query starts with or contains an observe guard prefix.
func (c *Classifier) hasObserveGuard(lower string) bool {
	for _, guard := range c.observeGuards {
		// Check as prefix or as the full query (trimmed).
		if strings.HasPrefix(lower, guard) {
			return true
		}
		// Also check if the query is exactly the guard phrase.
		trimmed := strings.TrimSpace(lower)
		if trimmed == guard {
			return true
		}
	}
	return false
}

// hasActVerb returns true if any act verb appears as a word token in the query.
func (c *Classifier) hasActVerb(tokens []string) bool {
	tokenSet := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = true
	}
	for _, verb := range c.actVerbs {
		if tokenSet[verb] {
			return true
		}
	}
	return false
}

// hasRecordVerb returns true if any record verb appears as a word token in the query.
func (c *Classifier) hasRecordVerb(tokens []string) bool {
	tokenSet := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = true
	}
	for _, verb := range c.recordVerbs {
		if tokenSet[verb] {
			return true
		}
	}
	return false
}

// extractDomainHints identifies domain hints from the query by matching domain keywords.
// Returns hints sorted by match count descending (most relevant first).
func (c *Classifier) extractDomainHints(lower string, tokens []string) []DomainHint {
	type domainScore struct {
		domain string
		count  int
	}

	tokenSet := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = true
	}

	var scores []domainScore
	for domain, keywords := range c.domainKeywords {
		count := 0
		for _, kw := range keywords {
			// Support multi-word keywords via substring match on the lowercased query.
			if strings.Contains(kw, " ") {
				if strings.Contains(lower, kw) {
					count++
				}
			} else {
				if tokenSet[kw] {
					count++
				}
			}
		}
		if count > 0 {
			scores = append(scores, domainScore{domain: domain, count: count})
		}
	}

	if len(scores) == 0 {
		return nil
	}

	// Sort by count descending.
	for i := 1; i < len(scores); i++ {
		for j := i; j > 0 && scores[j].count > scores[j-1].count; j-- {
			scores[j], scores[j-1] = scores[j-1], scores[j]
		}
	}

	hints := make([]DomainHint, len(scores))
	for i, s := range scores {
		hints[i] = DomainHint{
			Domain:     s.domain,
			Confidence: "HIGH",
		}
	}
	return hints
}

// tokenize splits a lowercase query into word tokens.
// Removes punctuation and splits on whitespace.
func tokenize(lower string) []string {
	// Replace common punctuation with spaces.
	replacer := strings.NewReplacer(
		".", " ",
		",", " ",
		"?", " ",
		"!", " ",
		";", " ",
		":", " ",
		"\"", " ",
		"'", " ",
		"(", " ",
		")", " ",
		"-", " ",
		"_", " ",
		"/", " ",
	)
	cleaned := replacer.Replace(lower)

	parts := strings.Fields(cleaned)
	var tokens []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tokens = append(tokens, p)
		}
	}
	return tokens
}
