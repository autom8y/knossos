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

	"github.com/autom8y/knossos/internal/know"
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
// When opts provides PriorTurnDomains AND threadHistory is non-empty,
// the follow-up enhancement path activates:
//   - Entity extraction from the last assistant message (vocabulary-seeded)
//   - Enhanced Stage 0 with prior entities
//   - BM25-rescored domain carryover after Stage 2
//
// Fail-open chain:
//   - Stage 3 fails -> use Stage 2 scores only
//   - Stage 2 fails -> BM25 fallback (BC-06: explicit, required)
//   - All stages fail -> return nil (caller falls back to v1)
func (o *Orchestrator) Assess(ctx context.Context, query string, threadHistory []ThreadMessage, opts ...AssessOptions) (*TriageResult, error) {
	start := time.Now()
	modelCalls := 0

	// Resolve options.
	var priorDomains []string
	if len(opts) > 0 {
		priorDomains = opts[0].PriorTurnDomains
	}

	// Determine if follow-up enhancement is active.
	isFollowUp := len(threadHistory) > 0
	enhanceFollowUp := isFollowUp && len(priorDomains) > 0

	// Extract entities from the last assistant message for Stage 0 enhancement.
	var priorEntities []string
	if enhanceFollowUp {
		priorEntities = o.extractPriorEntities(threadHistory)
	}

	// Stage 0: Multi-Turn Context Resolution (conditional).
	refinedQuery := query
	if isFollowUp && o.llmClient != nil {
		refined, err := o.stage0RefineQuery(ctx, query, threadHistory, priorEntities)
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

	// FM-3: BM25-rescored domain carryover after Stage 2.
	if enhanceFollowUp {
		stage2Candidates = o.injectPriorDomains(refinedQuery, stage2Candidates, priorDomains)
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

	// WS-3: Post-triage graph injection — surface same_repo adjacent-type domains
	// as supplemental candidates at baseline priority below any triage-scored candidate.
	injected := o.graphInjectionPass(result)

	result.RefinedQuery = refinedQuery
	result.TriageLatency = time.Since(start)
	result.ModelCallCount = modelCalls
	result.Intent.IsFollowUp = isFollowUp

	slog.Info("triage complete",
		"candidates", len(result.Candidates),
		"graph_injected", injected,
		"latency_ms", result.TriageLatency.Milliseconds(),
		"model_calls", result.ModelCallCount,
		"bm25_fallback", usedBM25Fallback,
	)

	return result, nil
}

// stage0RefineQuery uses Haiku to resolve implicit references in follow-up queries.
// P2-5: Uses a 2-second timeout to bound latency for the optional refinement step.
// priorEntities, when non-empty, are injected as a dedicated prompt section before
// conversation history (not as a synthetic message).
func (o *Orchestrator) stage0RefineQuery(ctx context.Context, query string, history []ThreadMessage, priorEntities []string) (string, error) {
	stage0Ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	userMsg := stage0UserMessage(query, history, priorEntities)

	resp, err := o.llmClient.Complete(stage0Ctx, llm.CompletionRequest{
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
	matchType           string // "document" or "section"
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
//
// WS-2: Merges document-level and section-level BM25 results. Section candidates
// are first-class — they compete with document candidates on equal footing.
func (o *Orchestrator) stage2BM25Fallback(query string, candidates []DomainMetadata) []stage2Candidate {
	bm25Results := o.searchIndex.SearchByBM25(query, 20)

	// WS-2: Also search sections and merge with document results.
	sectionResults := o.searchIndex.SearchSectionsByBM25(query, 10)

	allResults := make([]BM25Result, 0, len(bm25Results)+len(sectionResults))
	allResults = append(allResults, bm25Results...)
	allResults = append(allResults, sectionResults...)

	if len(allResults) == 0 {
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
	// For sections, extract the parent domain QN for metadata lookup.
	bm25Set := make(map[string]float64, len(allResults))
	matchTypes := make(map[string]string, len(allResults))
	for _, r := range allResults {
		// Keep the highest score when a QN appears in both doc and section results.
		if existing, ok := bm25Set[r.QualifiedName]; !ok || r.Score > existing {
			bm25Set[r.QualifiedName] = r.Score
			mt := r.MatchType
			if mt == "" {
				mt = "document"
			}
			matchTypes[r.QualifiedName] = mt
		}
	}

	// Match BM25 results against Stage 1 candidates.
	var result []stage2Candidate
	for _, c := range candidates {
		if score, ok := bm25Set[c.QualifiedName]; ok {
			result = append(result, stage2Candidate{
				metadata:  c,
				bm25Score: score,
				matchType: matchTypes[c.QualifiedName],
			})
		}
	}

	// Add BM25 results not in Stage 1 candidates (doc and section).
	candidateSet := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		candidateSet[c.QualifiedName] = true
	}
	for _, r := range allResults {
		if candidateSet[r.QualifiedName] {
			continue
		}
		candidateSet[r.QualifiedName] = true // Prevent duplicates

		// For section candidates, look up parent domain metadata.
		lookupQN := r.QualifiedName
		if strings.Contains(lookupQN, "##") {
			lookupQN = strings.SplitN(lookupQN, "##", 2)[0]
		}

		md, ok := o.searchIndex.GetMetadata(lookupQN)
		if !ok {
			continue
		}

		mt := r.MatchType
		if mt == "" {
			mt = "document"
		}
		result = append(result, stage2Candidate{
			metadata:  *md,
			bm25Score: r.Score,
			matchType: mt,
		})
		// Override QN for section candidates to preserve section address.
		if mt == "section" {
			result[len(result)-1].metadata.QualifiedName = r.QualifiedName
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
			repo = know.RepoFromQualifiedName(c.metadata.QualifiedName)
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
	matchTypeMap := make(map[string]string, len(stage2Candidates))
	for _, c := range stage2Candidates {
		embeddingMap[c.metadata.QualifiedName] = c.embeddingSimilarity
		freshnessMap[c.metadata.QualifiedName] = c.metadata.FreshnessScore
		domainTypeMap[c.metadata.QualifiedName] = c.metadata.DomainType
		if c.matchType != "" {
			matchTypeMap[c.metadata.QualifiedName] = c.matchType
		}
	}

	var triageCandidates []TriageCandidate
	for _, c := range parsed.Candidates {
		domainType := c.DomainType
		if domainType == "" {
			domainType = domainTypeMap[c.QualifiedName]
		}
		mt := matchTypeMap[c.QualifiedName]
		if mt == "" {
			mt = "document"
		}
		tc := TriageCandidate{
			QualifiedName:       c.QualifiedName,
			RelevanceScore:      c.RelevanceScore,
			EmbeddingSimilarity: embeddingMap[c.QualifiedName],
			Freshness:           freshnessMap[c.QualifiedName],
			Rationale:           c.Rationale,
			DomainType:          domainType,
			MatchType:           mt,
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
		mt := c.matchType
		if mt == "" {
			mt = "document"
		}
		triageCandidates = append(triageCandidates, TriageCandidate{
			QualifiedName:       c.metadata.QualifiedName,
			RelevanceScore:      score,
			EmbeddingSimilarity: c.embeddingSimilarity,
			Freshness:           c.metadata.FreshnessScore,
			DomainType:          c.metadata.DomainType,
			MatchType:           mt,
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

// extractPriorEntities extracts entity names from the last assistant message
// in threadHistory using the vocabulary-seeded keyword matcher.
func (o *Orchestrator) extractPriorEntities(history []ThreadMessage) []string {
	// Find the last assistant message.
	var lastAssistant string
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "assistant" {
			lastAssistant = history[i].Content
			break
		}
	}
	if lastAssistant == "" {
		return nil
	}

	// Build vocabulary from the search index's domain catalog.
	vocabulary := buildDomainVocabulary(o.searchIndex.ListAllDomains())
	return extractEntities(lastAssistant, vocabulary)
}

// buildDomainVocabulary builds a keyword list from domain qualified names
// by splitting on "::" and "-". Deduplicates and filters tokens shorter
// than 3 characters to avoid noise.
func buildDomainVocabulary(domains []DomainMetadata) []string {
	seen := make(map[string]bool)
	var vocab []string

	for _, d := range domains {
		// Split qualified name on "::" to get org, repo, domain parts.
		parts := strings.Split(d.QualifiedName, "::")
		for _, part := range parts {
			// Split each part on "-" to get individual tokens.
			tokens := strings.Split(part, "-")
			for _, tok := range tokens {
				tok = strings.ToLower(strings.TrimSpace(tok))
				if len(tok) < 3 {
					continue
				}
				if !seen[tok] {
					seen[tok] = true
					vocab = append(vocab, tok)
				}
			}
		}
	}

	return vocab
}

// extractEntities performs case-insensitive vocabulary keyword matching
// against text. Deduplicates and caps at 5 entities. Minimum token
// length of 3 characters is enforced by the vocabulary builder.
//
// NO regex. NO LLM call. Just vocabulary keyword matching.
func extractEntities(text string, vocabulary []string) []string {
	textLower := strings.ToLower(text)

	seen := make(map[string]bool)
	var entities []string

	for _, token := range vocabulary {
		if seen[token] {
			continue
		}
		if strings.Contains(textLower, token) {
			seen[token] = true
			entities = append(entities, token)
			if len(entities) >= 5 {
				break
			}
		}
	}

	return entities
}

// injectPriorDomains adds prior-turn domains to the Stage 2 candidate list
// using BM25-rescored relevance. Domains already in Stage 2 are skipped.
// Injected domains receive their actual BM25 score against the refined query,
// or a soft floor of 0.1 if they don't appear in BM25 results.
//
// Single-turn carryover ONLY -- priorDomains represents the immediately prior turn.
func (o *Orchestrator) injectPriorDomains(refinedQuery string, candidates []stage2Candidate, priorDomains []string) []stage2Candidate {
	if len(priorDomains) == 0 {
		return candidates
	}

	// Build set of already-present qualified names.
	present := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		present[c.metadata.QualifiedName] = true
	}

	// Filter to prior domains NOT already in Stage 2.
	var missing []string
	for _, qn := range priorDomains {
		if !present[qn] {
			missing = append(missing, qn)
		}
	}
	if len(missing) == 0 {
		return candidates
	}

	// BM25 re-score: search the refined query and build a score lookup.
	bm25Results := o.searchIndex.SearchByBM25(refinedQuery, 20)
	bm25Scores := make(map[string]float64, len(bm25Results))
	for _, r := range bm25Results {
		bm25Scores[r.QualifiedName] = r.Score
	}

	// Inject missing prior domains with BM25 scores (or soft floor).
	for _, qn := range missing {
		md, ok := o.searchIndex.GetMetadata(qn)
		if !ok {
			continue
		}
		score := 0.1 // Soft floor for domains not in BM25 results.
		if bm25Score, found := bm25Scores[qn]; found {
			score = bm25Score
		}
		candidates = append(candidates, stage2Candidate{
			metadata:  *md,
			bm25Score: score,
		})
	}

	// Re-sort by BM25 score descending after injection.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].bm25Score != candidates[j].bm25Score {
			return candidates[i].bm25Score > candidates[j].bm25Score
		}
		return candidates[i].embeddingSimilarity > candidates[j].embeddingSimilarity
	})

	slog.Info("FM-3 injected prior-turn domains",
		"prior_domains", len(priorDomains),
		"injected", len(missing),
		"total_candidates", len(candidates),
	)

	return candidates
}

// graphInjectionPass traverses same_repo edges for each triage candidate and
// injects adjacent-type domains as supplemental candidates at baseline priority.
//
// WS-3: Graph injection surfaces related-but-undiscovered domains at zero LLM cost.
// Injected candidates receive a baseline score below any triage-scored candidate.
// Injection is bounded to maxGraphInjectPerCandidate per primary candidate
// and maxGraphInjectTotal total injections.
//
// Returns the number of injected candidates.
func (o *Orchestrator) graphInjectionPass(result *TriageResult) int {
	const maxGraphInjectPerCandidate = 2
	const maxGraphInjectTotal = 4
	const baselineScore = 0.15 // Below any real triage score (min ~0.3)

	if result == nil || len(result.Candidates) == 0 {
		return 0
	}

	// Build set of already-present QNs (including section parents).
	present := make(map[string]bool, len(result.Candidates))
	for _, c := range result.Candidates {
		present[c.QualifiedName] = true
		// Also mark parent domain for section candidates.
		if idx := strings.Index(c.QualifiedName, "##"); idx > 0 {
			present[c.QualifiedName[:idx]] = true
		}
	}

	// Collect present domain types for diversity checking.
	presentTypes := make(map[string]bool)
	for _, c := range result.Candidates {
		presentTypes[c.DomainType] = true
	}

	var injected []TriageCandidate
	totalInjected := 0

	for _, c := range result.Candidates {
		if totalInjected >= maxGraphInjectTotal {
			break
		}

		// Get graph edges for this candidate.
		lookupQN := c.QualifiedName
		if idx := strings.Index(lookupQN, "##"); idx > 0 {
			lookupQN = lookupQN[:idx]
		}

		edges := o.searchIndex.GetRelationships(lookupQN)
		perCandidateInjected := 0

		for _, edge := range edges {
			if perCandidateInjected >= maxGraphInjectPerCandidate {
				break
			}
			if totalInjected >= maxGraphInjectTotal {
				break
			}

			// Only inject via same_repo edges (cross-type discovery).
			if edge.Type != "same_repo" {
				continue
			}

			// Skip if already present.
			if present[edge.Target] {
				continue
			}

			// Look up metadata for the injection target.
			md, ok := o.searchIndex.GetMetadata(edge.Target)
			if !ok {
				continue
			}

			// Only inject if the target is a different domain type (cross-type).
			if presentTypes[md.DomainType] {
				continue
			}

			injected = append(injected, TriageCandidate{
				QualifiedName:  edge.Target,
				RelevanceScore: baselineScore,
				Freshness:      md.FreshnessScore,
				DomainType:     md.DomainType,
				MatchType:      "document",
				RelatedDomains: []string{lookupQN}, // provenance: injected from this candidate
			})

			present[edge.Target] = true
			presentTypes[md.DomainType] = true
			perCandidateInjected++
			totalInjected++
		}
	}

	if totalInjected > 0 {
		result.Candidates = append(result.Candidates, injected...)
		slog.Info("WS-3 graph injection",
			"injected", totalInjected,
			"total_candidates", len(result.Candidates),
		)
	}

	return totalInjected
}

