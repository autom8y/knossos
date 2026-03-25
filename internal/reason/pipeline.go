// Package reason implements the Clew reasoning pipeline.
// It orchestrates intent classification, knowledge retrieval, trust evaluation,
// context assembly, and response generation for Tier 1 (Observe) queries.
//
// The pipeline converges four existing infrastructure packages:
//   - internal/search/ (BM25+RRF retrieval)
//   - internal/trust/ (confidence scoring, provenance, gap admission)
//   - internal/registry/org/ (domain catalog)
//   - internal/tokenizer/ (token counting)
//
// Layer invariants:
//   - reason/ does NOT import internal/serve/ or internal/cmd/
//   - trust/ and search/ do NOT import reason/
package reason

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/trust"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/intent"
	"github.com/autom8y/knossos/internal/reason/response"
)

// Pipeline is the top-level reasoning orchestrator.
// Constructed once per server lifetime; reused across queries.
type Pipeline struct {
	classifier *intent.Classifier
	assembler  *reasoncontext.Assembler
	generator  *response.Generator
	scorer     *trust.Scorer
	search     *search.SearchIndex
	catalog    *registryorg.DomainCatalog
	config     ReasoningConfig
}

// NewPipeline constructs a Pipeline with all dependencies.
func NewPipeline(
	classifier *intent.Classifier,
	assembler *reasoncontext.Assembler,
	generator *response.Generator,
	scorer *trust.Scorer,
	searchIndex *search.SearchIndex,
	catalog *registryorg.DomainCatalog,
	config ReasoningConfig,
) *Pipeline {
	return &Pipeline{
		classifier: classifier,
		assembler:  assembler,
		generator:  generator,
		scorer:     scorer,
		search:     searchIndex,
		catalog:    catalog,
		config:     config,
	}
}

// Query runs the full reasoning pipeline for a single user question.
// Always returns a response (never returns a nil *ReasoningResponse).
// Returns an error only for programming errors (nil dependencies), not for
// Claude API failures (those produce degraded responses).
func (p *Pipeline) Query(ctx context.Context, question string) (*response.ReasoningResponse, error) {
	if p.classifier == nil || p.assembler == nil || p.generator == nil ||
		p.scorer == nil || p.search == nil {
		return nil, fmt.Errorf("pipeline has nil dependencies")
	}

	// Step 1: Classify intent.
	intentResult := p.classifier.Classify(question)

	// Step 2: Short-circuit for unsupported tiers (Record/Act).
	// D-9: Record and Act are not yet supported. Return "not yet supported" response.
	if !intentResult.Answerable {
		return unsupportedResponse(intentResult), nil
	}

	// Step 3: Search for relevant knowledge.
	domains := extractSearchDomains(intentResult.DomainHints)
	searchResults := p.search.Search(question, search.SearchOptions{
		Limit:   p.config.SearchLimit,
		Domains: domains,
	})

	// Step 4: Build provenance chain from search results + catalog.
	linkInputs := buildProvenanceLinkInputs(searchResults, p.catalog)
	now := time.Now()
	decay := p.scorer.Config().Decay
	chain := trust.NewProvenanceChain(linkInputs, &decay, now)

	// Step 5: Compute confidence score.
	freshness := trust.FreshnessFromChain(&chain)
	scoreInput := trust.ScoreInput{
		Freshness:        freshness,
		RetrievalQuality: normalizeRetrievalQuality(searchResults),
		DomainCoverage:   computeDomainCoverage(intentResult.DomainHints, chain),
		Chain:            &chain,
		MissingDomains:   findMissingDomains(intentResult.DomainHints, chain),
		StaleDomains:     findStaleDomains(chain, p.scorer.Config().Thresholds.LowThreshold),
	}
	confidence := p.scorer.Score(scoreInput)

	slog.Info("confidence score computed",
		"overall", confidence.Overall,
		"freshness", confidence.Freshness,
		"retrieval", confidence.Retrieval,
		"coverage", confidence.Coverage,
		"tier", confidence.Tier.String(),
		"source_count", len(chain.Sources),
		"stale_count", len(scoreInput.StaleDomains),
		"missing_count", len(scoreInput.MissingDomains),
	)

	// Step 6: LOW tier short-circuit (D-9 -- skip Claude entirely).
	if confidence.Tier == trust.TierLow {
		return lowConfidenceResponse(confidence, intentResult), nil
	}

	// Step 7: Assemble context window.
	assembled := p.assembler.Assemble(searchResults, &chain, confidence, question, p.config.Org)

	// Step 8: Build intent summary for response.
	intentSummary := buildIntentSummary(intentResult)

	// Step 9: Generate response (HIGH or MEDIUM tier).
	return p.generator.Generate(ctx, assembled, confidence, &chain, intentSummary)
}

// unsupportedResponse creates a response for Record/Act intents (Tier 2/3).
// Returns an informative "not yet supported" message.
func unsupportedResponse(intentResult intent.IntentResult) *response.ReasoningResponse {
	return &response.ReasoningResponse{
		Answer: fmt.Sprintf("I cannot help with that action yet. %s", intentResult.UnsupportedReason),
		Tier:   trust.TierLow,
		Intent: buildIntentSummary(intentResult),
	}
}

// lowConfidenceResponse creates a response for LOW tier without calling Claude (D-9).
// The GapAdmission IS the response.
func lowConfidenceResponse(confidence trust.ConfidenceScore, intentResult intent.IntentResult) *response.ReasoningResponse {
	answer := "insufficient knowledge to answer this question reliably"
	if confidence.Gap != nil && confidence.Gap.Reason != "" {
		answer = confidence.Gap.Reason
	}

	return &response.ReasoningResponse{
		Answer:     answer,
		Confidence: confidence,
		Gap:        confidence.Gap,
		Tier:       trust.TierLow,
		Intent:     buildIntentSummary(intentResult),
	}
}

// buildIntentSummary converts an IntentResult to an IntentSummary for the response.
func buildIntentSummary(intentResult intent.IntentResult) response.IntentSummary {
	domains := make([]string, len(intentResult.DomainHints))
	for i, h := range intentResult.DomainHints {
		domains[i] = h.Domain
	}
	return response.IntentSummary{
		Tier:       intentResult.Tier.String(),
		Domains:    domains,
		Answerable: intentResult.Answerable,
	}
}

// normalizeRetrievalQuality converts search scores to [0.0, 1.0].
// Uses the top result's score as the signal.
// Returns 0.0 for empty results.
func normalizeRetrievalQuality(results []search.SearchResult) float64 {
	if len(results) == 0 {
		return 0.0
	}
	// Normalize: top score / 1000 (SearchResult.Score is already scaled by 1000 in fusion).
	// Clamp to [0.0, 1.0].
	score := float64(results[0].Score) / 1000.0
	if score > 1.0 {
		return 1.0
	}
	if score < 0.0 {
		return 0.0
	}
	return score
}

// computeDomainCoverage calculates the fraction of requested domains found.
// Returns 1.0 when DomainHints is empty (unfiltered query -- no specific expectation).
func computeDomainCoverage(hints []intent.DomainHint, chain trust.ProvenanceChain) float64 {
	if len(hints) == 0 {
		return 1.0
	}

	chainDomains := make(map[string]bool, len(chain.Sources))
	for _, s := range chain.Sources {
		chainDomains[s.Domain] = true
	}

	found := 0
	for _, h := range hints {
		if chainDomains[h.Domain] {
			found++
		}
	}
	return float64(found) / float64(len(hints))
}

// findMissingDomains identifies domains hinted by the classifier but absent from the chain.
func findMissingDomains(hints []intent.DomainHint, chain trust.ProvenanceChain) []string {
	chainDomains := make(map[string]bool, len(chain.Sources))
	for _, s := range chain.Sources {
		chainDomains[s.Domain] = true
	}

	var missing []string
	for _, h := range hints {
		if !chainDomains[h.Domain] {
			missing = append(missing, h.Domain)
		}
	}
	return missing
}

// findStaleDomains identifies domains in the chain below the freshness threshold.
func findStaleDomains(chain trust.ProvenanceChain, threshold float64) []trust.StaleDomainInfo {
	now := time.Now()
	var stale []trust.StaleDomainInfo
	for _, s := range chain.Sources {
		if s.FreshnessAtQuery < threshold {
			daysSince := 0
			if !s.GeneratedAt.IsZero() {
				daysSince = int(now.Sub(s.GeneratedAt).Hours() / 24)
			}
			stale = append(stale, trust.StaleDomainInfo{
				QualifiedName:      s.QualifiedName,
				Domain:             s.Domain,
				Repo:               s.Repo,
				Freshness:          s.FreshnessAtQuery,
				DaysSinceGenerated: daysSince,
			})
		}
	}
	return stale
}

// buildProvenanceLinkInputs converts search results + catalog into ProvenanceLinkInput[].
func buildProvenanceLinkInputs(
	results []search.SearchResult,
	catalog *registryorg.DomainCatalog,
) []trust.ProvenanceLinkInput {
	if catalog == nil {
		return nil
	}

	var inputs []trust.ProvenanceLinkInput
	seen := make(map[string]bool)

	for _, r := range results {
		if r.Domain != search.DomainKnowledge {
			continue
		}
		qn := r.Name
		if seen[qn] {
			continue
		}
		seen[qn] = true

		entry, ok := catalog.LookupDomain(qn)
		if !ok {
			continue
		}

		inputs = append(inputs, trust.ProvenanceLinkInput{
			QualifiedName: entry.QualifiedName,
			GeneratedAt:   entry.GeneratedAt,
			SourceHash:    entry.SourceHash,
			FilePath:      entry.Path,
			Domain:        entry.Domain,
			Repo:          repoFromQualifiedName(entry.QualifiedName),
		})
	}

	return inputs
}

// repoFromQualifiedName extracts the repo component from a qualified name "org::repo::domain".
// Returns empty string if the format is not as expected.
func repoFromQualifiedName(qn string) string {
	parts := strings.SplitN(qn, "::", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// extractSearchDomains converts DomainHints to search.Domain filter list.
// Returns nil (unfiltered) when hints is empty.
func extractSearchDomains(hints []intent.DomainHint) []search.Domain {
	if len(hints) == 0 {
		return nil
	}
	// For knowledge queries, always include the knowledge domain.
	return []search.Domain{search.DomainKnowledge}
}
