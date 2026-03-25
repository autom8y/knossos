// Package intent classifies user queries into action tiers and extracts domain hints.
// Uses keyword heuristics only -- no LLM calls.
package intent

// ActionTier identifies the type of action a query requests.
type ActionTier int

const (
	// TierObserve is a read-only knowledge query (Tier 1). Sprint-6 scope.
	TierObserve ActionTier = iota
	// TierRecord requests knowledge creation or update (Tier 2). Classified but not supported.
	TierRecord
	// TierAct requests an action to be executed (Tier 3). Classified but not supported.
	TierAct
)

// String returns the human-readable tier name.
func (at ActionTier) String() string {
	switch at {
	case TierObserve:
		return "OBSERVE"
	case TierRecord:
		return "RECORD"
	case TierAct:
		return "ACT"
	default:
		return "UNKNOWN"
	}
}

// DomainHint is a hint about which .know/ domain the query relates to.
type DomainHint struct {
	// Domain is the bare domain name (e.g., "architecture", "conventions", "feat/materialization").
	Domain string

	// Confidence is how confident the classifier is in this hint.
	// "HIGH" (keyword match) or "LOW" (inferred from context).
	Confidence string
}

// IntentResult is the output of intent classification.
type IntentResult struct {
	// Tier is the classified action tier (Observe, Record, Act).
	Tier ActionTier

	// DomainHints are the extracted domain references, most relevant first.
	DomainHints []DomainHint

	// Answerable indicates whether the query can be answered in the current sprint.
	// false for Record and Act tiers (not yet supported).
	Answerable bool

	// UnsupportedReason is the explanation when Answerable is false.
	// Empty when Answerable is true.
	UnsupportedReason string

	// RawQuery is the original query string (for downstream logging).
	RawQuery string
}
