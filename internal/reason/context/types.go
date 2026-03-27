// Package context assembles the Claude API context window from search results,
// trust assessments, and token budget constraints.
package context

import (
	"time"

	"github.com/autom8y/knossos/internal/trust"
)

// SourceMaterial is a single piece of .know/ content prepared for Claude's context window.
type SourceMaterial struct {
	// QualifiedName is the canonical cross-repo address (e.g., "autom8y::knossos::architecture").
	QualifiedName string

	// Section is the specific section heading within the domain file.
	// Empty when the entire domain is included.
	Section string

	// Content is the text content to include in the prompt.
	Content string

	// TokenCount is the token count of Content (pre-computed).
	TokenCount int

	// Freshness is the freshness score at query time (0.0-1.0).
	Freshness float64

	// FreshnessLabel is the human-readable freshness annotation
	// ("fresh", "moderately stale", "stale").
	FreshnessLabel string

	// GeneratedAt is the generation timestamp of the source .know/ file.
	GeneratedAt time.Time

	// Domain is the bare domain name.
	Domain string

	// Repo is the repository name.
	Repo string

	// RelevanceScore is the normalized BM25 score (0.0-1.0).
	RelevanceScore float64

	// InclusionScore is the composite score used for packing priority.
	// Computed from RelevanceScore, Freshness, and domain diversity contribution.
	InclusionScore float64
}

// AssembledContext is the complete context window prepared for a Claude API call.
type AssembledContext struct {
	// SystemPrompt is the rendered system prompt (identity + tier behavior + sources).
	SystemPrompt string

	// UserMessage is the user's question (passed through from the query).
	UserMessage string

	// Sources are the included source materials, ordered by inclusion score descending.
	Sources []SourceMaterial

	// Budget is the budget report for this assembly.
	Budget BudgetReport

	// Tier is the confidence tier governing prompt behavior.
	Tier trust.ConfidenceTier

	// StaleDomains lists domains whose freshness is below the staleness threshold.
	// Populated by the pipeline for MEDIUM-tier responses to enable domain-specific
	// freshness caveats in the system prompt.
	StaleDomains []trust.StaleDomainInfo

	// CEDiagnostics tracks contextual-equilibrium mechanism activity.
	// Non-nil when assembly completes with CE mechanisms active.
	CEDiagnostics *CEDiagnostics
}

// CEDiagnostics captures contextual-equilibrium mechanism activity during assembly.
// Populated by the Assembler to verify all five CE mechanisms are observable.
type CEDiagnostics struct {
	// TypeTokenBreakdown maps domain type -> total tokens consumed.
	TypeTokenBreakdown map[string]int

	// DiversityFloorEvents lists floor enforcement actions (WS-1).
	DiversityFloorEvents []DiversityFloorEvent

	// TypeCeilingHits lists per-type budget ceiling events (WS-5).
	TypeCeilingHits []TypeCeilingHit

	// SectionCandidatesPacked is the count of section-type candidates included (WS-2).
	SectionCandidatesPacked int

	// TotalCandidatesPacked is the total count of candidates included.
	TotalCandidatesPacked int

	// SourceBudget is the resolved source budget in tokens.
	SourceBudget int

	// TypeCeiling is the per-type budget ceiling in tokens (0 = disabled).
	TypeCeiling int
}

// DiversityFloorEvent records a single floor enforcement action.
type DiversityFloorEvent struct {
	FloorType     string  // The domain type that was missing
	QualifiedName string  // The candidate that was force-included
	Score         float64 // The candidate's relevance score
	UsedSummary   bool    // Whether summary substitution was used
}

// TypeCeilingHit records when a domain type hit its budget ceiling.
type TypeCeilingHit struct {
	DomainType      string // The type that hit the ceiling
	QualifiedName   string // The candidate that triggered it
	TokensBefore    int    // Tokens already consumed by this type
	CandidateTokens int    // Tokens the candidate would have consumed
	Ceiling         int    // The per-type ceiling
	UsedSummary     bool   // Whether summary substitution was attempted
	Skipped         bool   // Whether the candidate was ultimately skipped
}

// BudgetReport tracks token allocation for a single context assembly.
type BudgetReport struct {
	// SystemPromptTokens is the token count for the system prompt (identity + tier instructions).
	SystemPromptTokens int

	// SourceMaterialTokens is the total token count for included source material.
	SourceMaterialTokens int

	// UserMessageTokens is the token count for the user's question.
	UserMessageTokens int

	// TotalTokens is the total tokens assembled (system + sources + user).
	TotalTokens int

	// BudgetLimit is the configured maximum for source material tokens.
	BudgetLimit int

	// SourcesIncluded is the number of source sections included.
	SourcesIncluded int

	// SourcesSkipped is the number of source sections that did not fit.
	SourcesSkipped int

	// BudgetUtilization is SourceMaterialTokens / BudgetLimit (0.0-1.0).
	BudgetUtilization float64
}
