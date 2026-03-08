package search

import (
	"strings"
)

// stopWords is the set of common words filtered from query tokenization.
var stopWords = map[string]bool{
	"the": true, "and": true, "or": true, "for": true,
	"with": true, "from": true, "when": true, "that": true,
	"this": true, "into": true, "use": true, "how": true,
	"do": true, "i": true, "my": true, "a": true, "an": true,
	"to": true, "in": true, "is": true, "it": true, "of": true,
}

// Levenshtein computes the Levenshtein edit distance between two strings.
// Standard dynamic programming algorithm, O(m*n) time and space.
// Exported so callers (e.g., cmd/ask) can reuse the implementation.
func Levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use single-row optimization: O(min(m,n)) space.
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = MinOf3(
				curr[j-1]+1,    // insertion
				prev[j]+1,      // deletion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

// MinOf3 returns the minimum of three integers.
func MinOf3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// tokenize splits a query into searchable tokens, filtering stop words.
// Splits on whitespace and normalizes to lowercase.
func tokenize(s string) []string {
	words := strings.Fields(strings.ToLower(s))
	result := make([]string, 0, len(words))
	for _, w := range words {
		if !stopWords[w] && w != "" {
			result = append(result, w)
		}
	}
	return result
}

// extractKeywords parses description text for "Triggers:" and "Use when:" patterns.
// Returns a deduplicated, lowercased slice of keywords.
func extractKeywords(description string) []string {
	seen := make(map[string]bool)
	var keywords []string

	add := func(s string) {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" && !seen[s] {
			seen[s] = true
			keywords = append(keywords, s)
		}
	}

	// Parse "Triggers: word1, word2" and "Use when: phrase1, phrase2" patterns.
	for _, line := range strings.Split(description, "\n") {
		lower := strings.ToLower(strings.TrimSpace(line))
		var rest string
		switch {
		case strings.HasPrefix(lower, "triggers:"):
			rest = strings.TrimPrefix(lower, "triggers:")
		case strings.HasPrefix(lower, "use when:"):
			rest = strings.TrimPrefix(lower, "use when:")
		default:
			continue
		}
		for _, part := range strings.Split(rest, ",") {
			add(strings.TrimSpace(part))
		}
	}

	// Also extract significant words from the full description as secondary keywords.
	for _, word := range strings.Fields(strings.ToLower(description)) {
		// Strip punctuation from word boundaries.
		word = strings.Trim(word, ".,;:!?\"'()")
		if len(word) > 3 && !stopWords[word] {
			add(word)
		}
	}

	return keywords
}

// scoreEntry scores a single entry against a query.
// Returns the score (higher is better) and a match-type label.
func scoreEntry(query string, entry SearchEntry) (int, string) {
	qLower := strings.ToLower(query)
	nameLower := strings.ToLower(entry.Name)

	// Tier 1: Exact match (1000).
	if qLower == nameLower {
		score := 1000
		if entry.Boosted {
			score += 200
		}
		return score, "exact"
	}
	for _, alias := range entry.Aliases {
		if qLower == strings.ToLower(alias) {
			score := 1000
			if entry.Boosted {
				score += 200
			}
			return score, "exact"
		}
	}

	// Tier 2: Prefix match (500).
	if strings.HasPrefix(nameLower, qLower) {
		score := 500
		if entry.Boosted {
			score += 200
		}
		return score, "prefix"
	}

	// Tier 3: Keyword scoring (accumulated).
	tokens := tokenize(query)
	if len(tokens) > 0 {
		keywordScore := 0
		allMatched := true

		// Build a set of entry name words (split on spaces and hyphens).
		nameWords := make(map[string]bool)
		for _, w := range strings.FieldsFunc(nameLower, func(r rune) bool {
			return r == ' ' || r == '-'
		}) {
			nameWords[strings.ToLower(w)] = true
		}

		sumLower := strings.ToLower(entry.Summary)
		descLower := strings.ToLower(entry.Description)

		// Build a keyword set from entry.Keywords for O(1) lookup.
		entryKeywords := make(map[string]bool, len(entry.Keywords))
		for _, k := range entry.Keywords {
			entryKeywords[strings.ToLower(k)] = true
		}

		for _, tok := range tokens {
			tokMatched := false

			// Exact keyword match: +150.
			if entryKeywords[tok] {
				keywordScore += 150
				tokMatched = true
			}

			// Exact word match in entry name: +120.
			if nameWords[tok] {
				keywordScore += 120
				tokMatched = true
			}

			// Contains in summary: +100.
			if strings.Contains(sumLower, tok) {
				keywordScore += 100
				tokMatched = true
			}

			// Contains in description: +50.
			if strings.Contains(descLower, tok) {
				keywordScore += 50
				tokMatched = true
			}

			if !tokMatched {
				allMatched = false
			}
		}

		if keywordScore > 0 {
			if allMatched {
				keywordScore += 100 // all-tokens-matched bonus
			}
			if entry.Boosted {
				keywordScore += 200
			}
			return keywordScore, "keyword"
		}
	}

	// Tier 4: Fuzzy match (300 - 50*distance). Only used as fallback when
	// all other tiers produce score=0.
	dist := Levenshtein(qLower, nameLower)
	if dist <= 3 && len(nameLower) > 0 && dist < len(nameLower)/2 {
		score := 300 - 50*dist
		if entry.Boosted {
			score += 200
		}
		return score, "fuzzy"
	}

	return 0, "none"
}
