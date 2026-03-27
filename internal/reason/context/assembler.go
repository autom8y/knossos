package context

import (
	"sort"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/trust"
)

// TokenCounter abstracts token counting for testability.
// Production implementation wraps internal/tokenizer.Counter.
type TokenCounter interface {
	Count(text string) int
}

// AssemblerConfig controls context assembly behavior.
type AssemblerConfig struct {
	// SourceBudgetTokens is the maximum tokens for source material.
	// Default: 8000. Used as fallback when triage domain count is unavailable.
	SourceBudgetTokens int

	// RelevanceWeight controls the influence of BM25 score on inclusion priority.
	// Default: 0.50.
	RelevanceWeight float64

	// FreshnessWeight controls the influence of freshness on inclusion priority.
	// Default: 0.30.
	FreshnessWeight float64

	// DiversityWeight controls the influence of domain diversity on inclusion priority.
	// Default: 0.20.
	DiversityWeight float64

	// DiversityFloorTypes lists domain types that should be represented in the
	// assembled context when available. After greedy packing, if a floor type
	// has no representation and a candidate of that type scores above
	// DiversityFloorThreshold, the assembler includes it.
	// WS-1: Driven by configuration, not hardcoded domain name lists (AP-4).
	DiversityFloorTypes []string

	// DiversityFloorThreshold is the minimum relevance score for a floor-type
	// candidate to be force-included. Floor enforcement is skipped when the
	// best candidate of the required type scores below this threshold (R-6).
	// Default: 0.10.
	DiversityFloorThreshold float64

	// MaxTypeFraction is the maximum fraction of SourceBudgetTokens that any
	// single domain type may consume. When a candidate would exceed the ceiling
	// for its type, the assembler substitutes summary content if available.
	// WS-5: Prevents architecture monoculture at the budget level.
	// Default: 0.50 (no single type may consume more than half the budget).
	// Set to 0 to disable per-type budget ceilings.
	MaxTypeFraction float64

	// TriageDomainCount is the number of triage candidates, used to dynamically
	// resolve the source budget via resolveSourceBudget(). When zero, the static
	// SourceBudgetTokens is used as the fallback.
	// FM-3: Set by the caller (pipeline) when triage results are available.
	TriageDomainCount int

	// SummaryLookup returns a summary for a given qualified name.
	// When non-nil and a summary is found, candidates at position 4+ use
	// summary content instead of full content. Fail-open: when nil or when
	// the lookup returns empty, full content is used.
	// FM-3: Typically backed by summary.Store.GetSummary.
	SummaryLookup func(qualifiedName string) (string, bool)

	// OrgTopology is the pre-rendered org topology section for the system prompt.
	// ADR-TOPO-2: Populated at startup from topology.yaml + domain catalog.
	// Empty string = omit topology section (fail-open).
	OrgTopology string
}

// DefaultAssemblerConfig returns production default configuration.
func DefaultAssemblerConfig() AssemblerConfig {
	return AssemblerConfig{
		// Source budget expanded from 4,000 to 8,000 tokens to accommodate
		// full .know/ content (typically 3 domains x 2,500 tokens each).
		// The previous 4,000-token budget was designed for 200-char stubs.
		SourceBudgetTokens: 8000,
		RelevanceWeight:    0.50,
		FreshnessWeight:    0.30,
		DiversityWeight:    0.20,
		// WS-1: Domain types that should be represented when available.
		// Driven by configuration (AP-4), not hardcoded domain name lists.
		DiversityFloorTypes: []string{
			"conventions", "scar-tissue", "design-constraints", "test-coverage",
		},
		DiversityFloorThreshold: 0.10,
		// WS-5: Conservative ceiling — no single type consumes more than half.
		MaxTypeFraction: 0.50,
	}
}

// Assembler builds the Claude API context window from search results and trust data.
type Assembler struct {
	counter TokenCounter
	config  AssemblerConfig
}

// NewAssembler creates an Assembler with the given token counter and config.
func NewAssembler(counter TokenCounter, config AssemblerConfig) *Assembler {
	return &Assembler{
		counter: counter,
		config:  config,
	}
}

// resolveSourceBudget returns a dynamic source budget based on triage domain count.
// When triageDomainCount is zero, the fallback (configured default) is used.
//
// FM-3 Progressive Context Disclosure:
//   - 1-2 domains: fallback (8000) -- narrow query, top domains only
//   - 3-4 domains: 12000 -- medium cross-domain synthesis
//   - 5+ domains: 16000 -- broad org intelligence query
func resolveSourceBudget(triageDomainCount int, fallback int) int {
	if triageDomainCount <= 0 {
		return fallback
	}
	switch {
	case triageDomainCount <= 2:
		return fallback
	case triageDomainCount <= 4:
		return 12000
	default:
		return 16000
	}
}

// candidate is an intermediate struct used during greedy packing scoring.
type candidate struct {
	source         SourceMaterial
	inclusionScore float64
}

// Assemble builds an AssembledContext from search results, trust data, and the user's question.
// Uses metadata-weighted greedy packing (Approach B) per the TDD spec:
//
//	inclusionScore = (relevanceWeight * relevanceScore)
//	              + (freshnessWeight * freshness)
//	              + (diversityWeight * diversityBonus)
//
// diversityBonus: 1.0 for first source from domain, 0.3 for second, 0.0 for third+.
// Candidates are sorted descending by inclusionScore, then greedily packed until
// SourceBudgetTokens is exhausted.
//
// WS-2: conversationHistory is optional. When provided, it is passed through to
// RenderSystemPrompt which inserts a CONVERSATION HISTORY section before source material.
func (a *Assembler) Assemble(
	results []search.SearchResult,
	chain *trust.ProvenanceChain,
	score trust.ConfidenceScore,
	question string,
	org string,
	conversationHistory ...[]ConversationTurn,
) *AssembledContext {
	if len(results) == 0 {
		// No results: return minimal context with empty sources.
		systemPrompt := RenderSystemPrompt(org, score.Tier, nil, a.config.OrgTopology, conversationHistory...)
		budgetMgr := NewBudgetManager(a.config.SourceBudgetTokens)
		report := budgetMgr.Report()
		report.SystemPromptTokens = a.counter.Count(systemPrompt)
		report.UserMessageTokens = a.counter.Count(question)
		report.TotalTokens = report.SystemPromptTokens + report.UserMessageTokens
		return &AssembledContext{
			SystemPrompt: systemPrompt,
			UserMessage:  question,
			Sources:      nil,
			Budget:       report,
			Tier:         score.Tier,
		}
	}

	// Build lookup map from ProvenanceChain for freshness data.
	freshnessByQN := make(map[string]float64)
	generatedAtByQN := make(map[string]time.Time)
	domainByQN := make(map[string]string)
	repoByQN := make(map[string]string)
	if chain != nil {
		for _, s := range chain.Sources {
			freshnessByQN[s.QualifiedName] = s.FreshnessAtQuery
			generatedAtByQN[s.QualifiedName] = s.GeneratedAt
			domainByQN[s.QualifiedName] = s.Domain
			repoByQN[s.QualifiedName] = s.Repo
		}
	}

	// Find max BM25 score for normalization.
	maxScore := 0
	for _, r := range results {
		if r.Score > maxScore {
			maxScore = r.Score
		}
	}

	// Build candidate list with relevance scores.
	// Only include knowledge domain results.
	var candidates []candidate
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

		content := r.Summary
		if r.Description != "" {
			content = r.Description
		}
		if content == "" {
			content = r.Summary
		}

		tokenCount := a.counter.Count(content)
		if tokenCount == 0 {
			// Skip zero-token candidates -- they contribute nothing.
			continue
		}

		// Normalize relevance score to [0.0, 1.0].
		relevanceScore := 0.0
		if maxScore > 0 {
			relevanceScore = float64(r.Score) / float64(maxScore)
			if relevanceScore > 1.0 {
				relevanceScore = 1.0
			}
		}

		// Look up freshness from ProvenanceChain.
		freshness, hasFreshness := freshnessByQN[qn]
		if !hasFreshness {
			// Not in provenance chain -- treat as zero freshness (unknown).
			freshness = 0.0
		}

		// Resolve domain and repo. For section candidates (QN contains "##"),
		// fall back to the parent document QN for provenance chain lookups.
		domain := domainByQN[qn]
		repo := repoByQN[qn]
		generatedAt := generatedAtByQN[qn]
		if domain == "" {
			if idx := strings.Index(qn, "##"); idx > 0 {
				parentQN := qn[:idx]
				domain = domainByQN[parentQN]
				repo = repoByQN[parentQN]
				if generatedAt.IsZero() {
					generatedAt = generatedAtByQN[parentQN]
				}
				if !hasFreshness {
					freshness = freshnessByQN[parentQN]
				}
			}
		}

		src := SourceMaterial{
			QualifiedName:  qn,
			Content:        content,
			TokenCount:     tokenCount,
			Freshness:      freshness,
			FreshnessLabel: freshnessLabel(freshness),
			GeneratedAt:    generatedAt,
			Domain:         domain,
			Repo:           repo,
			RelevanceScore: relevanceScore,
		}

		candidates = append(candidates, candidate{
			source:         src,
			inclusionScore: 0, // computed after domain tracking
		})
	}

	if len(candidates) == 0 {
		systemPrompt := RenderSystemPrompt(org, score.Tier, nil, a.config.OrgTopology, conversationHistory...)
		budgetMgr := NewBudgetManager(a.config.SourceBudgetTokens)
		report := budgetMgr.Report()
		report.SystemPromptTokens = a.counter.Count(systemPrompt)
		report.UserMessageTokens = a.counter.Count(question)
		report.TotalTokens = report.SystemPromptTokens + report.UserMessageTokens
		return &AssembledContext{
			SystemPrompt: systemPrompt,
			UserMessage:  question,
			Sources:      nil,
			Budget:       report,
			Tier:         score.Tier,
		}
	}

	// Compute diversity bonuses: 1.0 for first from domain, 0.3 for second, 0.0 for third+.
	// We need two passes: first sort by relevance to determine "natural" domain order,
	// then apply diversity bonuses based on that order.
	//
	// Per the TDD: diversity bonus is applied based on domain occurrence order
	// to break ties in favor of cross-domain coverage.
	domainCount := make(map[string]int)
	for i := range candidates {
		domain := candidates[i].source.Domain
		domainCount[domain]++
		var diversityBonus float64
		switch domainCount[domain] {
		case 1:
			diversityBonus = 1.0
		case 2:
			diversityBonus = 0.3
		default:
			diversityBonus = 0.0
		}

		candidates[i].inclusionScore = (a.config.RelevanceWeight * candidates[i].source.RelevanceScore) +
			(a.config.FreshnessWeight * candidates[i].source.Freshness) +
			(a.config.DiversityWeight * diversityBonus)

		// Scope relevance: boost candidates whose scope matches query terms.
		candidates[i].inclusionScore += scopeRelevance(candidates[i].source.QualifiedName, question)
		candidates[i].source.InclusionScore = candidates[i].inclusionScore
	}

	// Sort by inclusion score descending.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].inclusionScore > candidates[j].inclusionScore
	})

	// FM-3: Resolve source budget dynamically based on triage domain count.
	sourceBudget := resolveSourceBudget(a.config.TriageDomainCount, a.config.SourceBudgetTokens)

	// Greedy packing: include candidates until token budget is exhausted.
	// FM-3: Candidates at position 4+ (0-indexed 3+) use summary content
	// when a SummaryLookup is available.
	// WS-5: Per-source-type budget ceilings prevent any single type from
	// dominating the context window.
	budgetMgr := NewBudgetManager(sourceBudget)
	typeTokens := make(map[string]int) // domain type -> tokens consumed
	typeCeiling := 0
	if a.config.MaxTypeFraction > 0 {
		typeCeiling = int(float64(sourceBudget) * a.config.MaxTypeFraction)
	}
	var included []SourceMaterial

	// CE diagnostic tracking.
	diag := &CEDiagnostics{
		TypeTokenBreakdown: make(map[string]int),
		SourceBudget:       sourceBudget,
		TypeCeiling:        typeCeiling,
	}

	for i, c := range candidates {
		src := c.source

		// FM-3: Summary tier for positions 4+ (0-indexed >= 3).
		if i >= 3 && a.config.SummaryLookup != nil {
			if summary, ok := a.config.SummaryLookup(src.QualifiedName); ok && summary != "" {
				src.Content = summary
				src.TokenCount = a.counter.Count(summary)
			}
			// Fail-open: if summary not found, use full content.
		}

		// WS-5: Per-type budget ceiling check.
		// The first candidate of any type is always allowed (up to overall budget).
		// Ceiling enforcement starts with the second candidate of the same type.
		if typeCeiling > 0 && typeTokens[src.Domain] > 0 &&
			typeTokens[src.Domain]+src.TokenCount > typeCeiling {
			ceilingHit := TypeCeilingHit{
				DomainType:      src.Domain,
				QualifiedName:   src.QualifiedName,
				TokensBefore:    typeTokens[src.Domain],
				CandidateTokens: src.TokenCount,
				Ceiling:         typeCeiling,
			}
			// Type ceiling would be exceeded. Try summary substitution.
			if a.config.SummaryLookup != nil {
				if summary, ok := a.config.SummaryLookup(src.QualifiedName); ok && summary != "" {
					src.Content = summary
					src.TokenCount = a.counter.Count(summary)
					ceilingHit.UsedSummary = true
				}
			}
			// Re-check after substitution.
			if typeTokens[src.Domain]+src.TokenCount > typeCeiling {
				// Still exceeds ceiling — skip this candidate.
				ceilingHit.Skipped = true
				diag.TypeCeilingHits = append(diag.TypeCeilingHits, ceilingHit)
				continue
			}
			diag.TypeCeilingHits = append(diag.TypeCeilingHits, ceilingHit)
		}

		// Consume() returns true if the item fits and increments included.
		// It returns false and increments skipped if it doesn't fit.
		if budgetMgr.Consume(src.TokenCount) {
			included = append(included, src)
			typeTokens[src.Domain] += src.TokenCount
		}
		// Bin-packing heuristic: continue trying subsequent candidates
		// even after a skip -- a smaller candidate may still fit.
	}

	// WS-1: Diversity floor post-pass — ensure floor types are represented.
	included, floorEvents := a.diversityFloorPass(included, candidates, budgetMgr)
	diag.DiversityFloorEvents = floorEvents

	// Compute final CE diagnostics from included sources.
	diag.TotalCandidatesPacked = len(included)
	for _, src := range included {
		diag.TypeTokenBreakdown[src.Domain] += src.TokenCount
		if strings.Contains(src.QualifiedName, "##") {
			diag.SectionCandidatesPacked++
		}
	}

	// Render system prompt with included sources and conversation history.
	systemPrompt := RenderSystemPrompt(org, score.Tier, included, a.config.OrgTopology, conversationHistory...)

	// Compute final budget report.
	report := budgetMgr.Report()
	report.SystemPromptTokens = a.counter.Count(systemPrompt)
	report.UserMessageTokens = a.counter.Count(question)
	report.TotalTokens = report.SystemPromptTokens + report.SourceMaterialTokens + report.UserMessageTokens

	return &AssembledContext{
		SystemPrompt:  systemPrompt,
		UserMessage:   question,
		Sources:       included,
		Budget:        report,
		Tier:          score.Tier,
		CEDiagnostics: diag,
	}
}

// TriageCandidateInfo holds triage candidate data for the assembler.
// This is a data-only struct passed by value -- reason/ does NOT import triage/.
type TriageCandidateInfo struct {
	QualifiedName  string
	RelevanceScore float64
	Freshness      float64
	DomainType     string
}

// WeightedMeanFreshness computes the relevance-weighted mean freshness from triage candidates.
// BC-07: Formula: sum(RelevanceScore_i * FreshnessScore_i) / sum(RelevanceScore_i)
// This replaces the position-weighted model when triage candidates are available.
// Returns 0.0 when candidates is empty or all relevance scores are zero.
func WeightedMeanFreshness(candidates []TriageCandidateInfo) float64 {
	if len(candidates) == 0 {
		return 0.0
	}

	var weightedSum, weightSum float64
	for _, c := range candidates {
		weightedSum += c.RelevanceScore * c.Freshness
		weightSum += c.RelevanceScore
	}

	if weightSum == 0 {
		return 0.0
	}

	return weightedSum / weightSum
}

// scopeRelevance returns an additive bonus when the candidate's scope path
// contains a segment that appears in the query. Root-scope domains (no scope)
// return 0.0 -- they are neither boosted nor penalized.
//
// Sprint 5: Scope-aware packing. Segments shorter than 3 characters are
// skipped to avoid false positives on short path components like "a" or "db".
func scopeRelevance(qualifiedName, query string) float64 {
	qdn, err := know.Parse(qualifiedName)
	if err != nil || qdn.Scope == "" {
		return 0.0
	}

	lowerQuery := strings.ToLower(query)
	segments := strings.Split(strings.ToLower(qdn.Scope), "/")
	for _, seg := range segments {
		if len(seg) >= 3 && strings.Contains(lowerQuery, seg) {
			return 0.15
		}
	}
	return 0.0
}

// diversityFloorPass checks whether each configured floor type is represented
// in the included sources. If a floor type is missing and a candidate of that
// type exists above the relevance threshold, it is force-included using summary
// content if available, or full content if it fits.
//
// WS-1: Post-greedy-pass correction. Floor types are configurable (AP-4).
// Floor enforcement is gated by relevance threshold (R-6 mitigation).
func (a *Assembler) diversityFloorPass(included []SourceMaterial, candidates []candidate, budgetMgr *BudgetManager) ([]SourceMaterial, []DiversityFloorEvent) {
	if len(a.config.DiversityFloorTypes) == 0 {
		return included, nil
	}

	var events []DiversityFloorEvent

	// Build set of domain types already represented.
	representedTypes := make(map[string]bool)
	for _, src := range included {
		representedTypes[src.Domain] = true
	}

	// Build set of included QNs to avoid duplicates.
	includedQNs := make(map[string]bool, len(included))
	for _, src := range included {
		includedQNs[src.QualifiedName] = true
	}

	for _, floorType := range a.config.DiversityFloorTypes {
		if representedTypes[floorType] {
			continue // Already represented.
		}

		// Find the highest-scoring candidate of this type that wasn't included.
		var best *candidate
		for i := range candidates {
			if candidates[i].source.Domain == floorType && !includedQNs[candidates[i].source.QualifiedName] {
				if best == nil || candidates[i].inclusionScore > best.inclusionScore {
					best = &candidates[i]
				}
			}
		}

		if best == nil {
			continue // No candidate of this type available.
		}

		// R-6: Skip if below relevance threshold.
		if best.source.RelevanceScore < a.config.DiversityFloorThreshold {
			continue
		}

		src := best.source
		usedSummary := false

		// Try summary substitution first (lower token cost).
		if a.config.SummaryLookup != nil {
			if summary, ok := a.config.SummaryLookup(src.QualifiedName); ok && summary != "" {
				src.Content = summary
				src.TokenCount = a.counter.Count(summary)
				usedSummary = true
			}
		}

		// Attempt to include if budget allows.
		// Use CanFit first to avoid incrementing the skip counter.
		if budgetMgr.CanFit(src.TokenCount) {
			budgetMgr.Consume(src.TokenCount)
			included = append(included, src)
			representedTypes[floorType] = true
			events = append(events, DiversityFloorEvent{
				FloorType:     floorType,
				QualifiedName: src.QualifiedName,
				Score:         best.source.RelevanceScore,
				UsedSummary:   usedSummary,
			})
		}
	}

	return included, events
}

// freshnessLabel returns a human-readable freshness annotation.
// Thresholds match TierThresholds: HIGH >= 0.7, LOW < 0.4.
func freshnessLabel(freshness float64) string {
	switch {
	case freshness >= 0.7:
		return "fresh"
	case freshness >= 0.4:
		return "moderately stale"
	default:
		return "stale"
	}
}
