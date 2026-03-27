// Package response provides the Claude API integration and response generation
// for the Clew reasoning pipeline.
package response

import (
	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/trust"
)

// ReasoningResponse is the complete output of the reasoning pipeline.
// Produced by Pipeline.Query() and consumed by sprint-8 Slack rendering.
type ReasoningResponse struct {
	// Answer is the response text. For HIGH/MEDIUM: Claude-generated synthesis.
	// For LOW: GapAdmission.Reason. For degraded: fallback message.
	Answer string

	// Confidence is the composite trust assessment.
	Confidence trust.ConfidenceScore

	// Provenance is the provenance chain backing this response.
	// Non-nil for HIGH and MEDIUM; nil for LOW.
	Provenance *trust.ProvenanceChain

	// Gap is populated when Confidence.Tier == TierLow or when degraded.
	// Contains refusal explanation and actionable suggestions.
	Gap *trust.GapAdmission

	// Citations are the validated source citations.
	// Populated for HIGH and MEDIUM (post-validation).
	// Empty for LOW and degraded.
	Citations []Citation

	// TokensUsed tracks API token consumption.
	TokensUsed TokenReport

	// Tier is the confidence tier that governed this response's behavior.
	Tier trust.ConfidenceTier

	// Intent is the classified intent from the query.
	Intent IntentSummary

	// Degraded indicates whether this is a degraded response (Claude unavailable).
	Degraded bool

	// DegradedReason explains why the response is degraded, if applicable.
	DegradedReason string

	// CEDiagnostics tracks contextual-equilibrium mechanism activity.
	// Non-nil when assembly completed with CE mechanisms active.
	CEDiagnostics *reasoncontext.CEDiagnostics
}

// IntentSummary is a lightweight copy of intent.IntentResult for inclusion in the response.
// Avoids requiring downstream consumers to import internal/reason/intent/.
type IntentSummary struct {
	Tier       string
	Domains    []string
	Answerable bool
}

// StructuredAnswer is the JSON schema Claude must produce.
// Enforced via the structured outputs API.
type StructuredAnswer struct {
	// Answer is the response text with inline [repo::domain] citations.
	Answer string `json:"answer"`

	// Citations is the list of sources referenced in the answer.
	Citations []Citation `json:"citations"`

	// Caveats are optional warnings about source quality or coverage.
	Caveats []string `json:"caveats,omitempty"`
}

// Citation references a specific knowledge source used in the answer.
type Citation struct {
	// QualifiedName is the canonical cross-repo address (e.g., "autom8y::knossos::architecture").
	QualifiedName string `json:"qualified_name"`

	// Section is the specific section within the domain file, if applicable.
	Section string `json:"section,omitempty"`

	// Excerpt is a brief excerpt from the source that supports the claim.
	Excerpt string `json:"excerpt"`
}

// TokenReport tracks Claude API token consumption for observability.
type TokenReport struct {
	// PromptTokens is the number of input tokens sent to Claude.
	PromptTokens int

	// CompletionTokens is the number of output tokens received from Claude.
	CompletionTokens int

	// TotalTokens is PromptTokens + CompletionTokens.
	TotalTokens int

	// EstimatedCostUSD is the estimated cost in USD based on model pricing.
	EstimatedCostUSD float64
}
