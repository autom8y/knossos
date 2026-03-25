// Package search provides natural language query matching across the knossos CLI surface.
package search

// Domain identifies the data source of a search entry.
type Domain string

const (
	DomainCommand  Domain = "command"
	DomainConcept  Domain = "concept"
	DomainRite     Domain = "rite"
	DomainAgent    Domain = "agent"
	DomainDromena    Domain = "dromena"
	DomainRouting    Domain = "routing"
	DomainSession       Domain = "session"
	DomainProcession    Domain = "procession"
	DomainKnowledge     Domain = "knowledge"
)

// SearchEntry is a single indexed item from any data source.
type SearchEntry struct {
	Name        string   `json:"name"`
	Domain      Domain   `json:"domain"`
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Action      string   `json:"action"`
	Aliases     []string `json:"aliases,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Boosted     bool     `json:"-"`
}

// SearchResult is a scored search entry returned from a query.
type SearchResult struct {
	SearchEntry
	Score     int    `json:"score"`
	MatchType string `json:"match_type"`
}

// SearchOptions controls search behavior.
type SearchOptions struct {
	Limit   int              // Max results (0 = default 5)
	Domains []Domain         // Filter to these domains; empty = all
	Session *SessionSignals  // Session context for scoring modifiers; nil = no session
}

// DefaultLimit is the default number of results returned.
const DefaultLimit = 5
