package triage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/llm"
)

// errEmbeddingsNotAvailable signals that the embedding model is not available.
// This triggers the explicit BM25 fallback path (BC-06).
var errEmbeddingsNotAvailable = errors.New("embedding model not available (Sprint 5 stub)")

// Orchestrator runs the multi-stage triage pipeline (Stages 0-3).
// Constructed once per server lifetime; reused across queries.
type Orchestrator struct {
	llmClient      llm.Client
	searchIndex    SearchIndex
	embeddingModel EmbeddingModel
}

// NewOrchestrator creates a triage Orchestrator with all dependencies.
func NewOrchestrator(llmClient llm.Client, searchIndex SearchIndex, embeddingModel EmbeddingModel) *Orchestrator {
	return &Orchestrator{
		llmClient:      llmClient,
		searchIndex:    searchIndex,
		embeddingModel: embeddingModel,
	}
}

// Assess runs the full triage pipeline and returns ranked candidates.
//
// Fail-open chain:
//   - Stage 3 fails -> use Stage 2 scores only
//   - Stage 2 fails -> BM25 fallback (BC-06: explicit, required)
//   - All stages fail -> return nil (caller falls back to v1)
func (o *Orchestrator) Assess(ctx context.Context, query string, threadHistory []ThreadMessage) (*TriageResult, error) {
	start := time.Now()
	modelCalls := 0

	// Stage 0: Multi-Turn Context Resolution (conditional).
	refinedQuery := query
	isFollowUp := len(threadHistory) > 0

	if isFollowUp && o.llmClient != nil {
		refined, err := o.stage0RefineQuery(ctx, query, threadHistory)
		if err != nil {
			// Fail-open: use original query if refinement fails.
			slog.Warn("stage 0 query refinement failed, using original query",
				"error", err,
			)
		} else {
			refinedQuery = refined
			slog.Info("stage 0 refined query",
				"original", query,
				"refined", refinedQuery,
			)
		}
		modelCalls++
	}

	// Stage 1: Metadata Pre-filter (zero cost, <1ms).
	allDomains := o.searchIndex.ListAllDomains()
	stage1Candidates := o.stage1MetadataFilter(refinedQuery, allDomains)

	if len(stage1Candidates) == 0 {
		slog.Warn("stage 1 metadata filter returned no candidates")
		return nil, nil
	}

	// Stage 2: Embedding Pre-filter OR BM25 Fallback (BC-06).
	stage2Candidates, usedBM25Fallback := o.stage2PreFilter(ctx, refinedQuery, stage1Candidates)

	if len(stage2Candidates) == 0 {
		slog.Warn("stage 2 pre-filter returned no candidates")
		return nil, nil
	}

	// Cap to 20 candidates for Stage 3.
	if len(stage2Candidates) > 20 {
		stage2Candidates = stage2Candidates[:20]
	}

	// Stage 3: Haiku Deep Assessment (single LLM call).
	result, err := o.stage3HaikuAssessment(ctx, refinedQuery, stage2Candidates)
	if err != nil {
		// Fail-open: use Stage 2 scores as final ranking.
		slog.Warn("stage 3 haiku assessment failed, using stage 2 scores",
			"error", err,
			"candidates", len(stage2Candidates),
		)
		result = o.stage2FallbackResult(refinedQuery, stage2Candidates, isFollowUp)
	} else {
		modelCalls++
	}

	result.RefinedQuery = refinedQuery
	result.TriageLatency = time.Since(start)
	result.ModelCallCount = modelCalls
	result.Intent.IsFollowUp = isFollowUp

	slog.Info("triage complete",
		"candidates", len(result.Candidates),
		"latency_ms", result.TriageLatency.Milliseconds(),
		"model_calls", result.ModelCallCount,
		"bm25_fallback", usedBM25Fallback,
	)

	return result, nil
}

// stage0RefineQuery uses Haiku to resolve implicit references in follow-up queries.
func (o *Orchestrator) stage0RefineQuery(ctx context.Context, query string, history []ThreadMessage) (string, error) {
	userMsg := stage0UserMessage(query, history)

	resp, err := o.llmClient.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: stage0SystemPrompt,
		UserMessage:  userMsg,
		MaxTokens:    200,
	})
	if err != nil {
		return "", fmt.Errorf("stage 0 refinement: %w", err)
	}

	refined := strings.TrimSpace(resp)
	if refined == "" {
		return query, nil
	}

	return refined, nil
}

// stage1MetadataFilter applies zero-cost metadata matching.
// Passes ~60-90% of the corpus through to Stage 2.
func (o *Orchestrator) stage1MetadataFilter(query string, domains []DomainMetadata) []DomainMetadata {
	if len(domains) == 0 {
		return nil
	}

	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	var passed []DomainMetadata
	for _, d := range domains {
		// Check 1: Qualified name substring matching.
		qnLower := strings.ToLower(d.QualifiedName)
		domainLower := strings.ToLower(d.DomainType)
		repoLower := strings.ToLower(d.Repo)

		matched := false
		for _, term := range queryTerms {
			if len(term) < 3 {
				continue
			}
			if strings.Contains(qnLower, term) ||
				strings.Contains(domainLower, term) ||
				strings.Contains(repoLower, term) {
				matched = true
				break
			}
		}

		// Check 2: Domain type matching from query intent signals.
		if !matched {
			matched = domainTypeMatchesQuery(domainLower, queryTerms)
		}

		// Check 3: Pass through all candidates when query is broad.
		// A broad query has no specific domain/repo references.
		if !matched && isBroadQuery(queryTerms) {
			matched = true
		}

		// Staleness gating: exclude severely stale domains (freshness < 0.1).
		// 0.0 means freshness is unknown (Tier 1 default), so we pass those through.
		if d.FreshnessScore > 0 && d.FreshnessScore < 0.1 {
			continue
		}

		if matched {
			passed = append(passed, d)
		}
	}

	return passed
}

// domainTypeMatchesQuery checks if any query terms match common domain type signals.
func domainTypeMatchesQuery(domainType string, queryTerms []string) bool {
	// Map query terms to domain types they imply.
	typeSignals := map[string][]string{
		"architecture":      {"architecture", "struct", "design", "layout", "overview"},
		"scar-tissue":       {"bug", "issue", "problem", "broke", "fail", "error", "scar"},
		"conventions":       {"convention", "practice", "pattern", "style", "standard"},
		"release":           {"release", "deploy", "version", "changelog", "change"},
		"test-coverage":     {"test", "coverage", "testing"},
		"design-constraints": {"constraint", "frozen", "limit", "decision"},
	}

	signals, ok := typeSignals[domainType]
	if !ok {
		return false
	}

	for _, term := range queryTerms {
		for _, signal := range signals {
			if strings.Contains(term, signal) || strings.Contains(signal, term) {
				return true
			}
		}
	}
	return false
}

// isBroadQuery detects queries without specific domain/repo references.
func isBroadQuery(queryTerms []string) bool {
	broadIndicators := []string{"how", "what", "tell", "explain", "describe", "show", "about"}
	for _, term := range queryTerms {
		for _, indicator := range broadIndicators {
			if term == indicator {
				return true
			}
		}
	}
	return len(queryTerms) <= 3
}

// stage2PreFilter attempts embedding-based pre-filter, falling back to BM25.
// BC-06: When EmbeddingModel.Embed fails, MUST call SearchByBM25 as replacement.
// Returns the candidates and whether BM25 fallback was used.
func (o *Orchestrator) stage2PreFilter(ctx context.Context, query string, candidates []DomainMetadata) ([]stage2Candidate, bool) {
	// Attempt embedding-based filtering first.
	if o.embeddingModel != nil {
		_, err := o.embeddingModel.Embed(ctx, query)
		if err == nil {
			// Embedding available -- use cosine similarity pre-filter.
			slog.Info("stage 2 using embedding pre-filter")
			return o.stage2EmbeddingFilter(ctx, query, candidates)
		}
		slog.Info("stage 2 embedding failed, falling back to BM25",
			"error", err,
		)
	}

	// BC-06: Explicit BM25 fallback -- required path, not defensive coding.
	return o.stage2BM25Fallback(query, candidates), true
}

// stage2Candidate holds a candidate with its Stage 2 score.
type stage2Candidate struct {
	metadata            DomainMetadata
	embeddingSimilarity float64
	bm25Score           float64
}

// stage2EmbeddingFilter uses cosine similarity for pre-filtering.
func (o *Orchestrator) stage2EmbeddingFilter(ctx context.Context, query string, candidates []DomainMetadata) ([]stage2Candidate, bool) {
	// Placeholder for Sprint 7 embedding implementation.
	// For now, return all candidates with zero similarity (pass-through).
	result := make([]stage2Candidate, len(candidates))
	for i, c := range candidates {
		result[i] = stage2Candidate{
			metadata:            c,
			embeddingSimilarity: 0,
		}
	}
	return result, false
}

// stage2BM25Fallback uses BM25 search as the Stage 2 replacement.
// BC-06: This is an explicit, required fallback -- not optional defensive coding.
// Narrows to top-20 to avoid sending all 128 candidates to Haiku.
func (o *Orchestrator) stage2BM25Fallback(query string, candidates []DomainMetadata) []stage2Candidate {
	bm25Results := o.searchIndex.SearchByBM25(query, 20)

	if len(bm25Results) == 0 {
		// BM25 returned nothing -- pass through all Stage 1 candidates
		// (capped at 20 by caller).
		slog.Warn("BM25 fallback returned no results, passing Stage 1 candidates through")
		result := make([]stage2Candidate, len(candidates))
		for i, c := range candidates {
			result[i] = stage2Candidate{metadata: c}
		}
		return result
	}

	// Build a set of BM25 result qualified names for O(1) lookup.
	bm25Set := make(map[string]float64, len(bm25Results))
	for _, r := range bm25Results {
		bm25Set[r.QualifiedName] = r.Score
	}

	// Match BM25 results against Stage 1 candidates.
	var result []stage2Candidate
	for _, c := range candidates {
		if score, ok := bm25Set[c.QualifiedName]; ok {
			result = append(result, stage2Candidate{
				metadata:  c,
				bm25Score: score,
			})
		}
	}

	// If BM25 found results that were not in Stage 1 candidates,
	// add them directly (BM25 may find domains Stage 1 missed).
	candidateSet := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		candidateSet[c.QualifiedName] = true
	}
	for _, r := range bm25Results {
		if !candidateSet[r.QualifiedName] {
			md, ok := o.searchIndex.GetMetadata(r.QualifiedName)
			if ok {
				result = append(result, stage2Candidate{
					metadata:  *md,
					bm25Score: r.Score,
				})
			}
		}
	}

	// Sort by BM25 score descending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].bm25Score > result[j].bm25Score
	})

	// Cap at 20 for Stage 3.
	if len(result) > 20 {
		result = result[:20]
	}

	return result
}

// stage3HaikuAssessment uses a single Haiku call to rank candidates.
func (o *Orchestrator) stage3HaikuAssessment(ctx context.Context, query string, candidates []stage2Candidate) (*TriageResult, error) {
	if o.llmClient == nil {
		return nil, fmt.Errorf("llm client is nil")
	}

	// Build candidate metadata for Haiku.
	llmCandidates := make([]candidateForLLM, len(candidates))
	for i, c := range candidates {
		repo := c.metadata.Repo
		if repo == "" {
			repo = repoFromQualifiedName(c.metadata.QualifiedName)
		}
		llmCandidates[i] = candidateForLLM{
			QualifiedName:       c.metadata.QualifiedName,
			DomainType:          c.metadata.DomainType,
			Repo:                repo,
			FreshnessScore:      c.metadata.FreshnessScore,
			EmbeddingSimilarity: c.embeddingSimilarity,
		}
	}

	userMsg := stage3UserMessage(query, llmCandidates)

	resp, err := o.llmClient.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: stage3SystemPrompt,
		UserMessage:  userMsg,
		MaxTokens:    800, // G-3: must be >= 700 for triage calls.
	})
	if err != nil {
		return nil, fmt.Errorf("stage 3 haiku call: %w", err)
	}

	// Parse Haiku's JSON response.
	result, err := parseStage3Response(resp, candidates)
	if err != nil {
		return nil, fmt.Errorf("stage 3 response parse: %w", err)
	}

	return result, nil
}

// stage3Response is the expected JSON structure from Haiku Stage 3.
type stage3Response struct {
	Candidates []stage3Candidate `json:"candidates"`
	Intent     stage3Intent      `json:"intent"`
}

type stage3Candidate struct {
	QualifiedName  string  `json:"qualified_name"`
	RelevanceScore float64 `json:"relevance_score"`
	Rationale      string  `json:"rationale"`
	DomainType     string  `json:"domain_type"`
}

type stage3Intent struct {
	Type              string   `json:"type"`
	TargetDomainTypes []string `json:"target_domain_types"`
	Repos             []string `json:"repos"`
}

// parseStage3Response parses Haiku's JSON output into a TriageResult.
// Implements partial JSON recovery (G-3) for truncated responses.
func parseStage3Response(resp string, stage2Candidates []stage2Candidate) (*TriageResult, error) {
	resp = strings.TrimSpace(resp)

	var parsed stage3Response
	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		// Attempt partial JSON recovery: try to extract any completed candidates.
		parsed = attemptPartialJSONRecovery(resp)
		if len(parsed.Candidates) == 0 {
			return nil, fmt.Errorf("failed to parse stage 3 response: %w", err)
		}
		slog.Warn("stage 3 response partially parsed",
			"recovered_candidates", len(parsed.Candidates),
		)
	}

	// Build embedding similarity lookup from Stage 2.
	embeddingMap := make(map[string]float64, len(stage2Candidates))
	freshnessMap := make(map[string]float64, len(stage2Candidates))
	domainTypeMap := make(map[string]string, len(stage2Candidates))
	for _, c := range stage2Candidates {
		embeddingMap[c.metadata.QualifiedName] = c.embeddingSimilarity
		freshnessMap[c.metadata.QualifiedName] = c.metadata.FreshnessScore
		domainTypeMap[c.metadata.QualifiedName] = c.metadata.DomainType
	}

	var triageCandidates []TriageCandidate
	for _, c := range parsed.Candidates {
		domainType := c.DomainType
		if domainType == "" {
			domainType = domainTypeMap[c.QualifiedName]
		}
		tc := TriageCandidate{
			QualifiedName:       c.QualifiedName,
			RelevanceScore:      c.RelevanceScore,
			EmbeddingSimilarity: embeddingMap[c.QualifiedName],
			Freshness:           freshnessMap[c.QualifiedName],
			Rationale:           c.Rationale,
			DomainType:          domainType,
		}
		triageCandidates = append(triageCandidates, tc)
	}

	// Sort by relevance score descending.
	sort.Slice(triageCandidates, func(i, j int) bool {
		return triageCandidates[i].RelevanceScore > triageCandidates[j].RelevanceScore
	})

	// Cap at 5.
	if len(triageCandidates) > 5 {
		triageCandidates = triageCandidates[:5]
	}

	return &TriageResult{
		Candidates: triageCandidates,
		Intent: QueryIntent{
			Type:              parsed.Intent.Type,
			TargetDomainTypes: parsed.Intent.TargetDomainTypes,
			Repos:             parsed.Intent.Repos,
		},
	}, nil
}

// stage2FallbackResult builds a TriageResult from Stage 2 scores when Stage 3 fails.
func (o *Orchestrator) stage2FallbackResult(query string, candidates []stage2Candidate, isFollowUp bool) *TriageResult {
	// Use BM25 or embedding scores as the relevance proxy.
	var triageCandidates []TriageCandidate
	for _, c := range candidates {
		score := c.embeddingSimilarity
		if score == 0 && c.bm25Score > 0 {
			// Normalize BM25 score to [0, 1] range using max normalization.
			score = 0.5 // Default moderate relevance for BM25 fallback.
		}
		triageCandidates = append(triageCandidates, TriageCandidate{
			QualifiedName:       c.metadata.QualifiedName,
			RelevanceScore:      score,
			EmbeddingSimilarity: c.embeddingSimilarity,
			Freshness:           c.metadata.FreshnessScore,
			DomainType:          c.metadata.DomainType,
		})
	}

	// Sort by relevance descending, cap at 5.
	sort.Slice(triageCandidates, func(i, j int) bool {
		return triageCandidates[i].RelevanceScore > triageCandidates[j].RelevanceScore
	})
	if len(triageCandidates) > 5 {
		triageCandidates = triageCandidates[:5]
	}

	return &TriageResult{
		Candidates: triageCandidates,
		Intent: QueryIntent{
			Type:       "exploration",
			IsFollowUp: isFollowUp,
		},
	}
}

// attemptPartialJSONRecovery tries to extract completed candidates from truncated JSON.
// G-3: Haiku may truncate at max_tokens mid-JSON. This recovers whatever was completed.
func attemptPartialJSONRecovery(resp string) stage3Response {
	var result stage3Response

	// Try to find the candidates array start.
	start := strings.Index(resp, `"candidates"`)
	if start < 0 {
		return result
	}

	// Find the array start.
	arrayStart := strings.Index(resp[start:], "[")
	if arrayStart < 0 {
		return result
	}
	arrayStart += start

	// Try to find each complete JSON object within the array.
	depth := 0
	objStart := -1
	for i := arrayStart; i < len(resp); i++ {
		switch resp[i] {
		case '{':
			if depth == 0 {
				objStart = i
			}
			depth++
		case '}':
			depth--
			if depth == 0 && objStart >= 0 {
				var c stage3Candidate
				if err := json.Unmarshal([]byte(resp[objStart:i+1]), &c); err == nil {
					if c.QualifiedName != "" {
						result.Candidates = append(result.Candidates, c)
					}
				}
				objStart = -1
			}
		}
	}

	return result
}

// repoFromQualifiedName extracts the repo component from "org::repo::domain".
func repoFromQualifiedName(qn string) string {
	parts := strings.SplitN(qn, "::", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
