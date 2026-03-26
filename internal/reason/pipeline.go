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

// contentLookup returns a function that resolves qualified names to raw .know/
// content via the BM25 index. Returns nil when no search index is available,
// which triageCandidatesToSearchResults handles gracefully.
func (p *Pipeline) contentLookup() func(string) (string, bool) {
	if p.search == nil {
		return nil
	}
	return p.search.LookupContent
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
// Uses the top knowledge-domain result's score as the signal.
// Only considers DomainKnowledge results, which carry RRF-fused scores
// (x1000 scaled, range ~16-73). Non-knowledge results (commands, concepts)
// use structural scoring on a different scale and are excluded.
// Returns 0.0 when no knowledge results are found.
func normalizeRetrievalQuality(results []search.SearchResult) float64 {
	// Find top knowledge-domain result score.
	topScore := -1
	for _, r := range results {
		if r.Domain == search.DomainKnowledge && r.Score > topScore {
			topScore = r.Score
			break // Results are sorted by score descending; first match is top.
		}
	}
	if topScore < 0 {
		return 0.0
	}
	// RRF scores after x1000 scaling (index.go:273) range from ~16 to ~73
	// depending on how many retrieval lists the result appears in.
	// Max theoretical: 3000/(RRFConstK+1) = 73.17 (k=40, 3 lists).
	// Dividing by 100 normalizes to [0.16, 0.73] which maps correctly to
	// the trust model's confidence tiers (LOW < 0.4 < MEDIUM < 0.7 < HIGH).
	score := float64(topScore) / 100.0
	if score > 1.0 {
		return 1.0
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

// TriageCandidateInput holds triage candidate data passed to the pipeline.
// This is a data-only struct -- reason/ does NOT import triage/.
// The handler in slack/ converts triage.TriageCandidate to this type.
type TriageCandidateInput struct {
	QualifiedName       string
	RelevanceScore      float64
	EmbeddingSimilarity float64
	Freshness           float64
	Rationale           string
	DomainType          string
	RelatedDomains      []string
}

// TriageResultInput holds triage results passed to the pipeline.
// This is a data-only struct -- reason/ does NOT import triage/.
type TriageResultInput struct {
	RefinedQuery   string
	Candidates     []TriageCandidateInput
	ModelCallCount int

	// ConversationHistory holds recent conversation turns for multi-turn context.
	// WS-2: When non-empty, these turns are injected into the system prompt as a
	// CONVERSATION HISTORY section before source material. The handler populates
	// this from ConversationManager's thread history.
	ConversationHistory []reasoncontext.ConversationTurn
}

// QueryWithTriage runs the reasoning pipeline using pre-computed triage results.
// This method takes pre-computed triage results instead of doing its own intent
// classification and search. Used by the Slack handler when triage is available.
//
// Keep existing Query() for backwards compatibility (tests, non-Slack callers).
func (p *Pipeline) QueryWithTriage(ctx context.Context, triageInput *TriageResultInput) (*response.ReasoningResponse, error) {
	if p.assembler == nil || p.generator == nil || p.scorer == nil {
		return nil, fmt.Errorf("pipeline has nil dependencies")
	}

	if triageInput == nil {
		// Nil triageInput is a programming error -- caller should use Query() instead.
		return nil, fmt.Errorf("QueryWithTriage called with nil triageInput; use Query() for non-triage path")
	}
	if len(triageInput.Candidates) == 0 {
		// Triage ran but returned no candidates -- fall back with refined query.
		return p.Query(ctx, triageInput.RefinedQuery)
	}

	question := triageInput.RefinedQuery

	// Build search results from triage candidates, loading .know/ content
	// from the BM25 index so the assembler receives real knowledge content.
	searchResults := triageCandidatesToSearchResults(triageInput.Candidates, p.contentLookup())

	// Build provenance chain from triage candidates + catalog.
	linkInputs := buildProvenanceLinkInputs(searchResults, p.catalog)
	now := time.Now()
	decay := p.scorer.Config().Decay
	chain := trust.NewProvenanceChain(linkInputs, &decay, now)

	// BC-07: Compute weighted-mean freshness using triage RelevanceScores.
	candidateInfos := make([]reasoncontext.TriageCandidateInfo, len(triageInput.Candidates))
	for i, c := range triageInput.Candidates {
		candidateInfos[i] = reasoncontext.TriageCandidateInfo{
			QualifiedName:  c.QualifiedName,
			RelevanceScore: c.RelevanceScore,
			Freshness:      c.Freshness,
			DomainType:     c.DomainType,
		}
	}

	freshness := reasoncontext.WeightedMeanFreshness(candidateInfos)
	if freshness == 0 {
		// Triage freshness not available (Tier 1 default) -- fall back to chain.
		freshness = trust.FreshnessFromChain(&chain)
	}

	// Compute confidence score.
	scoreInput := trust.ScoreInput{
		Freshness:        freshness,
		RetrievalQuality: normalizeTriageRelevance(triageInput.Candidates),
		DomainCoverage:   1.0, // Triage already selected relevant domains.
		Chain:            &chain,
		StaleDomains:     findStaleDomains(chain, p.scorer.Config().Thresholds.LowThreshold),
	}
	confidence := p.scorer.Score(scoreInput)

	slog.Info("triage confidence score computed",
		"overall", confidence.Overall,
		"freshness", confidence.Freshness,
		"retrieval", confidence.Retrieval,
		"tier", confidence.Tier.String(),
		"triage_candidates", len(triageInput.Candidates),
	)

	// LOW tier short-circuit.
	if confidence.Tier == trust.TierLow {
		return &response.ReasoningResponse{
			Answer: "insufficient knowledge to answer this question reliably",
			Tier:   trust.TierLow,
			Gap:    confidence.Gap,
		}, nil
	}

	// WS-2: Assemble context window using triage candidates and conversation history.
	assembled := p.assembler.Assemble(searchResults, &chain, confidence, question, p.config.Org, triageInput.ConversationHistory)

	// Build intent summary from triage.
	intentSummary := response.IntentSummary{
		Tier:       "OBSERVE",
		Answerable: true,
	}

	// Generate response.
	return p.generator.Generate(ctx, assembled, confidence, &chain, intentSummary)
}

// QueryStream runs the reasoning pipeline with streaming output using pre-computed triage results.
// BC-03: Uses onChunk callback to stream text chunks. reason/ does NOT import slack/.
//
// Same as QueryWithTriage but uses GenerateStream instead of Generate.
// Streaming prompt: free-form text with inline [org::repo::domain] citation markers
// (NOT tool-forced structured output, which is incompatible with streaming).
func (p *Pipeline) QueryStream(ctx context.Context, triageInput *TriageResultInput, onChunk func(chunk string)) (*response.ReasoningResponse, error) {
	if p.assembler == nil || p.generator == nil || p.scorer == nil {
		return nil, fmt.Errorf("pipeline has nil dependencies")
	}

	if triageInput == nil {
		// Nil triageInput is a programming error -- caller should use Query() instead.
		return nil, fmt.Errorf("QueryStream called with nil triageInput; use Query() for non-triage path")
	}
	if len(triageInput.Candidates) == 0 {
		// Triage ran but returned no candidates -- fall back with refined query (non-streaming).
		return p.Query(ctx, triageInput.RefinedQuery)
	}

	question := triageInput.RefinedQuery

	// Build search results from triage candidates, loading .know/ content.
	searchResults := triageCandidatesToSearchResults(triageInput.Candidates, p.contentLookup())

	// Build provenance chain.
	linkInputs := buildProvenanceLinkInputs(searchResults, p.catalog)
	now := time.Now()
	decay := p.scorer.Config().Decay
	chain := trust.NewProvenanceChain(linkInputs, &decay, now)

	// BC-07: Compute weighted-mean freshness.
	candidateInfos := make([]reasoncontext.TriageCandidateInfo, len(triageInput.Candidates))
	for i, c := range triageInput.Candidates {
		candidateInfos[i] = reasoncontext.TriageCandidateInfo{
			QualifiedName:  c.QualifiedName,
			RelevanceScore: c.RelevanceScore,
			Freshness:      c.Freshness,
			DomainType:     c.DomainType,
		}
	}

	freshness := reasoncontext.WeightedMeanFreshness(candidateInfos)
	if freshness == 0 {
		freshness = trust.FreshnessFromChain(&chain)
	}

	// Compute confidence score.
	scoreInput := trust.ScoreInput{
		Freshness:        freshness,
		RetrievalQuality: normalizeTriageRelevance(triageInput.Candidates),
		DomainCoverage:   1.0,
		Chain:            &chain,
		StaleDomains:     findStaleDomains(chain, p.scorer.Config().Thresholds.LowThreshold),
	}
	confidence := p.scorer.Score(scoreInput)

	slog.Info("streaming triage confidence score computed",
		"overall", confidence.Overall,
		"tier", confidence.Tier.String(),
		"triage_candidates", len(triageInput.Candidates),
	)

	// LOW tier short-circuit (no streaming needed).
	if confidence.Tier == trust.TierLow {
		return &response.ReasoningResponse{
			Answer: "insufficient knowledge to answer this question reliably",
			Tier:   trust.TierLow,
			Gap:    confidence.Gap,
		}, nil
	}

	// WS-2: Assemble context window with conversation history.
	assembled := p.assembler.Assemble(searchResults, &chain, confidence, question, p.config.Org, triageInput.ConversationHistory)

	intentSummary := response.IntentSummary{
		Tier:       "OBSERVE",
		Answerable: true,
	}

	// Generate with streaming.
	return p.generator.GenerateStream(ctx, assembled, confidence, &chain, intentSummary, onChunk)
}

// triageCandidatesToSearchResults converts triage candidates to search results
// so the existing assembler can process them. The contentLookup function
// resolves qualified names to their raw .know/ content. When non-nil, each
// candidate's Description is populated so the assembler receives real content
// (not empty stubs). When nil, behavior is unchanged (backward compatible).
func triageCandidatesToSearchResults(candidates []TriageCandidateInput, contentLookup func(string) (string, bool)) []search.SearchResult {
	var results []search.SearchResult
	for _, c := range candidates {
		score := int(c.RelevanceScore * 1000)
		entry := search.SearchEntry{
			Name:   c.QualifiedName,
			Domain: search.DomainKnowledge,
		}

		// Populate Description from BM25 content store so the assembler
		// receives the actual .know/ content for this domain.
		if contentLookup != nil {
			if content, ok := contentLookup(c.QualifiedName); ok {
				entry.Description = content
			}
		}

		results = append(results, search.SearchResult{
			SearchEntry: entry,
			Score:       score,
			MatchType:   "triage",
		})
	}
	return results
}

// normalizeTriageRelevance returns the top candidate's relevance score as the
// retrieval quality signal.
func normalizeTriageRelevance(candidates []TriageCandidateInput) float64 {
	if len(candidates) == 0 {
		return 0.0
	}
	// Top candidate relevance score is already in [0, 1].
	top := candidates[0].RelevanceScore
	if top > 1.0 {
		return 1.0
	}
	return top
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
