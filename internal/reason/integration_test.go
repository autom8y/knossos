package reason

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/autom8y/knossos/internal/trust"

	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/intent"
	"github.com/autom8y/knossos/internal/reason/response"
)

// ---- Test infrastructure for integration tests ----

// buildMultiRepoCatalog constructs a DomainCatalog with 3 repos and 7 domains
// for cross-repo integration testing. Uses controlled timestamps for deterministic
// freshness decay calculations.
func buildMultiRepoCatalog(now time.Time) *registryorg.DomainCatalog {
	return &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		SyncedAt:      now.Format(time.RFC3339),
		Repos: []registryorg.RepoEntry{
			{
				Name: "knossos",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "autom8y::knossos::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
						GeneratedAt:   now.Add(-12 * time.Hour).Format(time.RFC3339),
						SourceHash:    "abc1234",
					},
					{
						QualifiedName: "autom8y::knossos::scar-tissue",
						Domain:        "scar-tissue",
						Path:          ".know/scar-tissue.md",
						GeneratedAt:   now.Add(-8 * 24 * time.Hour).Format(time.RFC3339),
						SourceHash:    "def5678",
					},
					{
						QualifiedName: "autom8y::knossos::conventions",
						Domain:        "conventions",
						Path:          ".know/conventions.md",
						GeneratedAt:   now.Add(-2 * 24 * time.Hour).Format(time.RFC3339),
						SourceHash:    "ghi9012",
					},
				},
			},
			{
				Name: "autom8y-web",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "autom8y::autom8y-web::architecture",
						Domain:        "architecture",
						Path:          ".know/architecture.md",
						GeneratedAt:   now.Add(-3 * 24 * time.Hour).Format(time.RFC3339),
						SourceHash:    "jkl3456",
					},
					{
						QualifiedName: "autom8y::autom8y-web::feat/dashboard",
						Domain:        "feat/dashboard",
						Path:          ".know/feat/dashboard.md",
						GeneratedAt:   now.Add(-1 * 24 * time.Hour).Format(time.RFC3339),
						SourceHash:    "mno7890",
					},
				},
			},
			{
				Name: "platform-infra",
				Domains: []registryorg.DomainEntry{
					{
						QualifiedName: "autom8y::platform-infra::release",
						Domain:        "release",
						Path:          ".know/release/history.md",
						GeneratedAt:   now.Add(-5 * 24 * time.Hour).Format(time.RFC3339),
						SourceHash:    "pqr1234",
					},
					{
						QualifiedName: "autom8y::platform-infra::test-coverage",
						Domain:        "test-coverage",
						Path:          ".know/test-coverage.md",
						GeneratedAt:   now.Add(-6 * 24 * time.Hour).Format(time.RFC3339),
						SourceHash:    "stu5678",
					},
				},
			},
		},
	}
}

// buildIntegrationPipeline constructs a Pipeline with a multi-repo catalog
// and the provided mock client.
func buildIntegrationPipeline(mock *response.MockClaudeClient, catalog *registryorg.DomainCatalog) *Pipeline {
	classifier := intent.NewClassifier()
	assembler := reasoncontext.NewAssembler(&testTokenCounter{}, reasoncontext.DefaultAssemblerConfig())
	generator := response.NewGenerator(mock, response.DefaultGeneratorConfig())
	scorer := trust.NewScorer(trust.DefaultConfig())
	searchIndex := buildTestSearchIndex()

	config := DefaultReasoningConfig()
	config.SearchLimit = 10

	return NewPipeline(classifier, assembler, generator, scorer, searchIndex, catalog, config)
}

// ---- Group A: Cross-Repo Registry (C1, C2) ----

func TestIntegration_CrossRepoRegistry_RepoCount(t *testing.T) {
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)

	assert.GreaterOrEqual(t, catalog.RepoCount(), 3,
		"PT-09 C1: catalog must contain at least 3 repos")
}

func TestIntegration_CrossRepoRegistry_DomainCount(t *testing.T) {
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)

	assert.GreaterOrEqual(t, catalog.DomainCount(), 7,
		"PT-09 C1: catalog must contain at least 7 domains")
}

func TestIntegration_CrossRepoRegistry_LookupCrossRepo(t *testing.T) {
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)

	// Verify cross-repo qualified name lookups (C2).
	crossRepoNames := []string{
		"autom8y::knossos::architecture",
		"autom8y::autom8y-web::architecture",
		"autom8y::platform-infra::release",
		"autom8y::autom8y-web::feat/dashboard",
		"autom8y::knossos::scar-tissue",
		"autom8y::knossos::conventions",
		"autom8y::platform-infra::test-coverage",
	}

	for _, qn := range crossRepoNames {
		t.Run(qn, func(t *testing.T) {
			entry, found := catalog.LookupDomain(qn)
			require.True(t, found, "PT-09 C2: LookupDomain must find %q", qn)
			assert.Equal(t, qn, entry.QualifiedName)
			assert.NotEmpty(t, entry.Domain, "domain must be populated")
			assert.NotEmpty(t, entry.Path, "path must be populated")
			assert.NotEmpty(t, entry.SourceHash, "source hash must be populated")
		})
	}
}

func TestIntegration_CrossRepoRegistry_LookupMissing(t *testing.T) {
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)

	_, found := catalog.LookupDomain("autom8y::nonexistent::missing")
	assert.False(t, found, "lookup for nonexistent domain must return false")
}

// ---- Group B: Provenance Chains with Real DomainEntry Data (C4) ----

func TestIntegration_ProvenanceChain_FromCatalogEntries(t *testing.T) {
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)
	decay := trust.DefaultConfig().Decay

	// Build ProvenanceLinkInputs from real catalog entries.
	inputs := []trust.ProvenanceLinkInput{
		linkInputFromCatalog(catalog, "autom8y::knossos::architecture"),
		linkInputFromCatalog(catalog, "autom8y::knossos::scar-tissue"),
		linkInputFromCatalog(catalog, "autom8y::platform-infra::release"),
	}

	chain := trust.NewProvenanceChain(inputs, &decay, now)

	// Verify chain structure (C4).
	require.Equal(t, 3, chain.Len(), "chain must have 3 sources")
	assert.False(t, chain.IsEmpty())

	for _, link := range chain.Sources {
		assert.NotEmpty(t, link.QualifiedName, "each link must have QualifiedName")
		assert.NotEmpty(t, link.SourceHash, "each link must have SourceHash")
		assert.NotEmpty(t, link.FilePath, "each link must have FilePath")
		assert.NotEmpty(t, link.Domain, "each link must have Domain")
	}
}

func TestIntegration_ProvenanceChain_FreshnessDecayCurves(t *testing.T) {
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)
	decay := trust.DefaultConfig().Decay

	inputs := []trust.ProvenanceLinkInput{
		linkInputFromCatalog(catalog, "autom8y::knossos::architecture"),
		linkInputFromCatalog(catalog, "autom8y::knossos::scar-tissue"),
		linkInputFromCatalog(catalog, "autom8y::platform-infra::release"),
	}

	chain := trust.NewProvenanceChain(inputs, &decay, now)

	// knossos::architecture -- 12h old, architecture half-life 14d.
	// Expected: exp(-ln(2)/14 * 0.5) ~ 0.976
	archLink := chain.Sources[0]
	expectedArchFreshness := math.Exp(-math.Ln2 / 14.0 * 0.5)
	assert.InDelta(t, expectedArchFreshness, archLink.FreshnessAtQuery, 0.02,
		"architecture freshness: 12h old, 14d half-life")

	// knossos::scar-tissue -- 8d old, scar-tissue half-life 10d.
	// Expected: exp(-ln(2)/10 * 8) ~ 0.574
	scarLink := chain.Sources[1]
	expectedScarFreshness := math.Exp(-math.Ln2 / 10.0 * 8.0)
	assert.InDelta(t, expectedScarFreshness, scarLink.FreshnessAtQuery, 0.02,
		"scar-tissue freshness: 8d old, 10d half-life")

	// platform-infra::release -- 5d old, release half-life 3d.
	// Expected: exp(-ln(2)/3 * 5) ~ 0.315
	releaseLink := chain.Sources[2]
	expectedReleaseFreshness := math.Exp(-math.Ln2 / 3.0 * 5.0)
	assert.InDelta(t, expectedReleaseFreshness, releaseLink.FreshnessAtQuery, 0.02,
		"release freshness: 5d old, 3d half-life")
}

func TestIntegration_ProvenanceChain_QualifiedNames(t *testing.T) {
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)
	decay := trust.DefaultConfig().Decay

	inputs := []trust.ProvenanceLinkInput{
		linkInputFromCatalog(catalog, "autom8y::knossos::architecture"),
		linkInputFromCatalog(catalog, "autom8y::autom8y-web::feat/dashboard"),
	}

	chain := trust.NewProvenanceChain(inputs, &decay, now)
	names := chain.QualifiedNames()

	assert.Equal(t, []string{
		"autom8y::knossos::architecture",
		"autom8y::autom8y-web::feat/dashboard",
	}, names)
}

// ---- Group C: Trust Scoring All Three Tiers (C3) ----

func TestIntegration_TrustScoring_HighTier(t *testing.T) {
	scorer := trust.NewScorer(trust.DefaultConfig())

	// HIGH: fresh architecture (0.97), high retrieval (0.9), full coverage (1.0).
	score := scorer.Score(trust.ScoreInput{
		Freshness:        0.97,
		RetrievalQuality: 0.9,
		DomainCoverage:   1.0,
	})

	assert.Equal(t, trust.TierHigh, score.Tier, "PT-09 C3: fresh+relevant+covered must be HIGH")
	assert.GreaterOrEqual(t, score.Overall, 0.7, "HIGH tier requires overall >= 0.7")
	assert.Nil(t, score.Gap, "HIGH tier must not have GapAdmission")
}

func TestIntegration_TrustScoring_MediumTier(t *testing.T) {
	scorer := trust.NewScorer(trust.DefaultConfig())

	// MEDIUM: moderate scar-tissue (0.57), moderate retrieval (0.6), full coverage (1.0).
	score := scorer.Score(trust.ScoreInput{
		Freshness:        0.57,
		RetrievalQuality: 0.6,
		DomainCoverage:   1.0,
	})

	assert.Equal(t, trust.TierMedium, score.Tier, "PT-09 C3: moderate inputs must be MEDIUM")
	assert.GreaterOrEqual(t, score.Overall, 0.4, "MEDIUM tier requires overall >= 0.4")
	assert.Less(t, score.Overall, 0.7, "MEDIUM tier requires overall < 0.7")
	assert.Nil(t, score.Gap, "MEDIUM tier must not have GapAdmission")
}

func TestIntegration_TrustScoring_LowTier(t *testing.T) {
	scorer := trust.NewScorer(trust.DefaultConfig())

	// LOW: stale/missing (0.1), low retrieval (0.1), poor coverage (0.3).
	staleDomains := []trust.StaleDomainInfo{
		{QualifiedName: "autom8y::knossos::scar-tissue", Domain: "scar-tissue", Repo: "knossos", Freshness: 0.1},
	}
	score := scorer.Score(trust.ScoreInput{
		Freshness:        0.1,
		RetrievalQuality: 0.1,
		DomainCoverage:   0.3,
		MissingDomains:   []string{"deployment"},
		StaleDomains:     staleDomains,
	})

	assert.Equal(t, trust.TierLow, score.Tier, "PT-09 C3: poor inputs must be LOW")
	assert.Less(t, score.Overall, 0.4, "LOW tier requires overall < 0.4")
	assert.NotNil(t, score.Gap, "LOW tier must have GapAdmission")
}

func TestIntegration_TrustScoring_ZeroInput_ArithmeticFallback(t *testing.T) {
	scorer := trust.NewScorer(trust.DefaultConfig())
	cfg := trust.DefaultConfig()
	wf := cfg.Weights.Freshness
	wr := cfg.Weights.Retrieval
	wc := cfg.Weights.Coverage
	wSum := wf + wr + wc

	// WS-2.4: A single zero input no longer collapses to 0.0. Instead, the arithmetic
	// mean fallback produces a score proportional to the non-zero components. This prevents
	// false LOW-tier refusals when one signal happens to be zero but others are strong.
	testCases := []struct {
		name        string
		freshness   float64
		retrieval   float64
		coverage    float64
		wantOverall float64
	}{
		{"zero_freshness", 0.0, 0.9, 1.0,
			(wf*0.0 + wr*0.9 + wc*1.0) / wSum},
		{"zero_retrieval", 0.9, 0.0, 1.0,
			(wf*0.9 + wr*0.0 + wc*1.0) / wSum},
		{"zero_coverage", 0.9, 0.9, 0.0,
			(wf*0.9 + wr*0.9 + wc*0.0) / wSum},
		{"all_zero", 0.0, 0.0, 0.0, 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := scorer.Score(trust.ScoreInput{
				Freshness:        tc.freshness,
				RetrievalQuality: tc.retrieval,
				DomainCoverage:   tc.coverage,
			})
			assert.InDelta(t, tc.wantOverall, score.Overall, 0.001,
				"arithmetic mean fallback should produce expected overall")
		})
	}
}

// ---- Group D: GapAdmission References Real Data (C5) ----

func TestIntegration_GapAdmission_MissingDomains(t *testing.T) {
	scorer := trust.NewScorer(trust.DefaultConfig())

	score := scorer.Score(trust.ScoreInput{
		Freshness:        0.1,
		RetrievalQuality: 0.1,
		DomainCoverage:   0.3,
		MissingDomains:   []string{"deployment", "monitoring"},
		StaleDomains:     nil,
	})

	require.NotNil(t, score.Gap, "PT-09 C5: LOW tier must have GapAdmission")
	assert.Equal(t, []string{"deployment", "monitoring"}, score.Gap.MissingDomains,
		"Gap must reference queried domain names")
	assert.NotEmpty(t, score.Gap.Reason, "Gap must have a reason")
	assert.NotEmpty(t, score.Gap.Suggestions, "Gap must have suggestions")
	assert.Contains(t, score.Gap.Reason, "deployment",
		"Reason must reference missing domains")
}

func TestIntegration_GapAdmission_StaleDomains(t *testing.T) {
	scorer := trust.NewScorer(trust.DefaultConfig())

	staleDomains := []trust.StaleDomainInfo{
		{
			QualifiedName: "autom8y::knossos::scar-tissue",
			Domain:        "scar-tissue",
			Repo:          "knossos",
			Freshness:     0.1,
		},
		{
			QualifiedName: "autom8y::platform-infra::release",
			Domain:        "release",
			Repo:          "platform-infra",
			Freshness:     0.05,
		},
	}

	score := scorer.Score(trust.ScoreInput{
		Freshness:        0.1,
		RetrievalQuality: 0.1,
		DomainCoverage:   0.3,
		MissingDomains:   nil,
		StaleDomains:     staleDomains,
	})

	require.NotNil(t, score.Gap, "PT-09 C5: LOW tier must have GapAdmission")
	require.Len(t, score.Gap.StaleDomains, 2, "Gap must have 2 stale domains")
	assert.Equal(t, "autom8y::knossos::scar-tissue", score.Gap.StaleDomains[0].QualifiedName)
	assert.Equal(t, "autom8y::platform-infra::release", score.Gap.StaleDomains[1].QualifiedName)
	assert.NotEmpty(t, score.Gap.Reason, "Gap must have a reason")
	assert.Contains(t, score.Gap.Reason, "autom8y::knossos::scar-tissue",
		"Reason must reference stale qualified names")
	assert.NotEmpty(t, score.Gap.Suggestions, "Gap must have suggestions")
}

func TestIntegration_GapAdmission_MixedMissingAndStale(t *testing.T) {
	scorer := trust.NewScorer(trust.DefaultConfig())

	score := scorer.Score(trust.ScoreInput{
		Freshness:        0.1,
		RetrievalQuality: 0.1,
		DomainCoverage:   0.2,
		MissingDomains:   []string{"deployment"},
		StaleDomains: []trust.StaleDomainInfo{
			{QualifiedName: "autom8y::knossos::scar-tissue", Domain: "scar-tissue", Repo: "knossos", Freshness: 0.1},
		},
	})

	require.NotNil(t, score.Gap)
	assert.Contains(t, score.Gap.Reason, "deployment", "Reason must reference missing domain")
	assert.Contains(t, score.Gap.Reason, "autom8y::knossos::scar-tissue", "Reason must reference stale domain")
	assert.GreaterOrEqual(t, len(score.Gap.Suggestions), 2,
		"Suggestions must cover both missing and stale domains")
}

// ---- Group E: Full Pipeline Integration ----

func TestIntegration_FullPipeline_LowTier_EmptySearch(t *testing.T) {
	// Empty search index -> no results -> LOW tier -> Claude NOT called.
	mock := &response.MockClaudeClient{}
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)
	p := buildIntegrationPipeline(mock, catalog)

	resp, err := p.Query(context.Background(), "How does the deployment pipeline work?")

	require.NoError(t, err)
	require.NotNil(t, resp, "pipeline must always return non-nil response")
	assert.Equal(t, trust.TierLow, resp.Tier, "empty search -> LOW tier")
	assert.NotNil(t, resp.Gap, "LOW tier must have GapAdmission")
	assert.Equal(t, 0, mock.CallCount, "Claude must NOT be called for LOW tier (D-9)")
	assert.NotEmpty(t, resp.Answer, "LOW tier must have an answer")
}

func TestIntegration_FullPipeline_BackwardCompat_NilCatalog(t *testing.T) {
	// Nil catalog -> graceful LOW response, no panic.
	mock := &response.MockClaudeClient{}
	p := buildIntegrationPipeline(mock, nil)

	resp, err := p.Query(context.Background(), "What is the architecture?")

	require.NoError(t, err)
	require.NotNil(t, resp, "nil catalog must not cause nil response")
	assert.Equal(t, trust.TierLow, resp.Tier, "nil catalog -> LOW tier")
	assert.Equal(t, 0, mock.CallCount, "Claude must NOT be called with nil catalog")
}

func TestIntegration_FullPipeline_IntentClassification(t *testing.T) {
	// Verify all three intent tiers flow through the pipeline correctly.
	testCases := []struct {
		name         string
		query        string
		expectedTier string
		answerable   bool
	}{
		{"observe_intent", "What is the architecture?", "OBSERVE", true},
		{"record_intent", "Update the scar tissue documentation", "RECORD", false},
		{"act_intent", "Deploy the new release to staging", "ACT", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &response.MockClaudeClient{}
			now := time.Now().UTC()
			catalog := buildMultiRepoCatalog(now)
			p := buildIntegrationPipeline(mock, catalog)

			resp, err := p.Query(context.Background(), tc.query)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.expectedTier, resp.Intent.Tier)
			assert.Equal(t, tc.answerable, resp.Intent.Answerable)
		})
	}
}

func TestIntegration_FullPipeline_NonAnswerable_NeverCallsClaude(t *testing.T) {
	mock := &response.MockClaudeClient{}
	now := time.Now().UTC()
	catalog := buildMultiRepoCatalog(now)
	p := buildIntegrationPipeline(mock, catalog)

	// Record intent.
	resp, err := p.Query(context.Background(), "Write new conventions documentation")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, mock.CallCount, "Record intent must not call Claude")

	// Act intent.
	resp2, err2 := p.Query(context.Background(), "Run the migration script")
	require.NoError(t, err2)
	require.NotNil(t, resp2)
	assert.Equal(t, 0, mock.CallCount, "Act intent must not call Claude")
}

// ---- Group F: Backward Compatibility (C7) ----

func TestIntegration_BackwardCompat_NilCatalog_NoResponse(t *testing.T) {
	mock := &response.MockClaudeClient{}
	p := buildIntegrationPipeline(mock, nil)

	resp, err := p.Query(context.Background(), "How does the sync pipeline work?")

	require.NoError(t, err)
	require.NotNil(t, resp, "PT-09 C7: nil catalog must return non-nil response")
	// WS-2.4: With arithmetic mean fallback, nil catalog (freshness=0, retrieval>0, coverage=1.0)
	// no longer collapses to 0.0. The score may land in MEDIUM if BM25 returns any hits.
	// The important invariant is that the response is non-nil with no provenance sources.
	if resp.Provenance != nil {
		assert.True(t, resp.Provenance.IsEmpty(),
			"PT-09 C7: nil catalog -> empty provenance chain (no sources)")
	}
}

func TestIntegration_BackwardCompat_EmptyCatalog(t *testing.T) {
	emptyCatalog := &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		SyncedAt:      time.Now().Format(time.RFC3339),
		Repos:         nil,
	}

	mock := &response.MockClaudeClient{}
	p := buildIntegrationPipeline(mock, emptyCatalog)

	resp, err := p.Query(context.Background(), "What is the architecture?")

	require.NoError(t, err)
	require.NotNil(t, resp, "PT-09 C7: empty catalog must return non-nil response")
	assert.Equal(t, trust.TierLow, resp.Tier, "PT-09 C7: empty catalog -> LOW tier")
	assert.Equal(t, 0, mock.CallCount, "Claude must NOT be called with empty catalog")
}

func TestIntegration_BackwardCompat_EmptyCatalog_NoRepos(t *testing.T) {
	emptyCatalog := &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		SyncedAt:      time.Now().Format(time.RFC3339),
		Repos:         []registryorg.RepoEntry{},
	}

	mock := &response.MockClaudeClient{}
	p := buildIntegrationPipeline(mock, emptyCatalog)

	resp, err := p.Query(context.Background(), "How does billing work?")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, trust.TierLow, resp.Tier)
	assert.Equal(t, 0, mock.CallCount)
}

func TestIntegration_BackwardCompat_CatalogWithEmptyRepos(t *testing.T) {
	// Catalog with repos that have no domains.
	spareCatalog := &registryorg.DomainCatalog{
		SchemaVersion: "1.0",
		Org:           "autom8y",
		SyncedAt:      time.Now().Format(time.RFC3339),
		Repos: []registryorg.RepoEntry{
			{Name: "empty-repo-1", Domains: nil},
			{Name: "empty-repo-2", Domains: []registryorg.DomainEntry{}},
		},
	}

	mock := &response.MockClaudeClient{}
	p := buildIntegrationPipeline(mock, spareCatalog)

	resp, err := p.Query(context.Background(), "What is the test coverage?")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, trust.TierLow, resp.Tier)
}

// ---- Helpers ----

// linkInputFromCatalog looks up a domain entry by qualified name and converts it
// to a ProvenanceLinkInput. Panics if the domain is not found (test-only).
func linkInputFromCatalog(catalog *registryorg.DomainCatalog, qn string) trust.ProvenanceLinkInput {
	entry, found := catalog.LookupDomain(qn)
	if !found {
		panic("test setup error: domain not found in catalog: " + qn)
	}

	// Extract repo name from qualified name: "org::repo::domain" -> "repo".
	repo := ""
	for _, r := range catalog.Repos {
		for _, d := range r.Domains {
			if d.QualifiedName == qn {
				repo = r.Name
				break
			}
		}
		if repo != "" {
			break
		}
	}

	return trust.ProvenanceLinkInput{
		QualifiedName: entry.QualifiedName,
		GeneratedAt:   entry.GeneratedAt,
		SourceHash:    entry.SourceHash,
		FilePath:      entry.Path,
		Domain:        entry.Domain,
		Repo:          repo,
	}
}
