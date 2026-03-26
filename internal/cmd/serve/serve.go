// Package serve implements the ari serve command for the HTTP webhook server.
package serve

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/envload"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/llm"
	"github.com/autom8y/knossos/internal/observe"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/reason"
	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/intent"
	"github.com/autom8y/knossos/internal/reason/response"
	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/search/knowledge"
	"github.com/autom8y/knossos/internal/serve"
	"github.com/autom8y/knossos/internal/serve/health"
	"github.com/autom8y/knossos/internal/serve/webhook"
	internalslack "github.com/autom8y/knossos/internal/slack"
	"github.com/autom8y/knossos/internal/slack/conversation"
	"github.com/autom8y/knossos/internal/slack/streaming"
	"github.com/autom8y/knossos/internal/tokenizer"
	"github.com/autom8y/knossos/internal/triage"
	"github.com/autom8y/knossos/internal/trust"
)

// serveOptions holds flag values for the serve command.
type serveOptions struct {
	org                string
	envFile            string
	port               int
	slackSigningSecret string
	slackBotToken      string
	drainTimeout       time.Duration
	maxConcurrent      int
}

// cmdContext holds shared state for serve commands.
type cmdContext struct {
	common.BaseContext
}

// NewServeCmd creates the serve command.
func NewServeCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	opts := serveOptions{}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP webhook server",
		Long: `Start an HTTP server for Slack webhook event processing.

The server provides:
  - Slack webhook signature verification (HMAC-SHA256)
  - Reasoning pipeline (knowledge retrieval + trust scoring + Claude generation)
  - Health endpoints (/health for liveness, /ready for readiness)
  - Graceful shutdown on SIGTERM/SIGINT
  - Request ID propagation and structured logging
  - OpenTelemetry tracing (optional, via OTEL_EXPORTER_OTLP_ENDPOINT)
  - Concurrency limiting for pipeline queries

Configuration is resolved via a four-tier hierarchy (highest wins):
  1. CLI flags (--port, --slack-signing-secret, etc.)
  2. Process environment variables (SLACK_SIGNING_SECRET, PORT, etc.)
  3. Org env file ($XDG_DATA_HOME/knossos/orgs/{org}/serve.env)
  4. Hardcoded defaults (port=8080, log_level=INFO, max_concurrent=10)

Required secrets (must be set via any tier):
  SLACK_SIGNING_SECRET  Slack app signing secret
  SLACK_BOT_TOKEN       Slack bot OAuth token
  ANTHROPIC_API_KEY     Claude API key (required for reasoning)

Optional configuration:
  PORT                          Server port (default: 8080)
  LOG_LEVEL                     Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
  MAX_CONCURRENT                Max concurrent pipeline queries (default: 10)
  OTEL_EXPORTER_OTLP_ENDPOINT  OTLP collector endpoint (empty = noop tracing)

Examples:
  ari serve --org autom8y
  ari serve --org autom8y --port 3000
  ari serve --env-file /path/to/custom.env
  SLACK_SIGNING_SECRET=xxx SLACK_BOT_TOKEN=xoxb-xxx ari serve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.org, "org", "",
		"Organization name (env: KNOSSOS_ORG, default: active org)")
	cmd.Flags().StringVar(&opts.envFile, "env-file", "",
		"Path to env file (default: $XDG_DATA_HOME/knossos/orgs/{org}/serve.env)")
	cmd.Flags().IntVar(&opts.port, "port", 0,
		"Server port (env: PORT, default: 8080)")
	cmd.Flags().StringVar(&opts.slackSigningSecret, "slack-signing-secret", "",
		"Slack signing secret (env: SLACK_SIGNING_SECRET)")
	cmd.Flags().StringVar(&opts.slackBotToken, "slack-bot-token", "",
		"Slack bot token (env: SLACK_BOT_TOKEN)")
	cmd.Flags().DurationVar(&opts.drainTimeout, "drain-timeout", 0,
		"Graceful shutdown drain timeout (default: 30s)")
	cmd.Flags().IntVar(&opts.maxConcurrent, "max-concurrent", 0,
		"Max concurrent pipeline queries (env: MAX_CONCURRENT, default: 10)")

	// ari serve does NOT require project context
	common.SetNeedsProject(cmd, false, true)

	return cmd
}

// runServe starts the HTTP server with the configured options.
func runServe(ctx *cmdContext, opts serveOptions) error {
	printer := ctx.GetPrinter(output.FormatText)

	// Resolve org context (nil is valid -- means pure env-var mode).
	orgCtx := resolveOrgContext(opts.org)

	// Load configuration via the four-tier hierarchy.
	cfg, err := envload.Load(orgCtx, envload.Overrides{
		SlackSigningSecret: opts.slackSigningSecret,
		SlackBotToken:      opts.slackBotToken,
		Port:               opts.port,
		MaxConcurrent:      opts.maxConcurrent,
		DrainTimeout:       opts.drainTimeout,
		EnvFile:            opts.envFile,
	})
	if err != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeUsageError, "failed to load configuration", err))
	}

	// Validate required secrets.
	if cfg.SlackSigningSecret == "" {
		return common.PrintAndReturn(printer,
			errors.New(errors.CodeUsageError, "SLACK_SIGNING_SECRET is required (set via env var, org env file, or --slack-signing-secret)"))
	}
	if cfg.SlackBotToken == "" {
		return common.PrintAndReturn(printer,
			errors.New(errors.CodeUsageError, "SLACK_BOT_TOKEN is required (set via env var, org env file, or --slack-bot-token)"))
	}

	// Configure structured logging (JSON to stderr with trace context).
	observe.ConfigureStructuredLogging(cfg.LogLevel)

	// Initialise OTEL tracer (noop when endpoint is empty).
	shutdownTracer, err := observe.InitTracer("clew", cfg.OTELEndpoint)
	if err != nil {
		slog.Warn("OTEL tracer initialization failed, continuing without tracing", "error", err)
	} else {
		defer func() {
			if shutdownErr := shutdownTracer(context.Background()); shutdownErr != nil {
				slog.Warn("OTEL tracer shutdown error", "error", shutdownErr)
			}
		}()
	}

	// Build server config.
	serverCfg := serve.DefaultConfig()
	serverCfg.Port = cfg.Port
	serverCfg.DrainTimeout = cfg.DrainTimeout

	// Create health checker and server.
	checker := health.NewChecker()
	srv := serve.New(serverCfg, serve.WithHealthChecker(checker))

	// Build the reasoning pipeline (fail-open: logs warnings for missing dependencies).
	// Returns intermediate components for health check registration.
	pipelineResult := buildPipeline()

	// Create metrics recorder (CloudWatch EMF via structured slog).
	metricsRecorder := observe.NewEMFRecorder()

	// Create cost tracker and instrumented pipeline wrapper.
	costTracker := observe.NewCostTracker()
	var queryRunner internalslack.QueryRunner
	if pipelineResult.pipeline != nil {
		queryRunner = observe.NewInstrumentedPipeline(pipelineResult.pipeline, costTracker, metricsRecorder)
	}

	// Build LLM client (shared by triage, knowledge index, conversation manager).
	// BC-01: llm.Client in internal/llm/, shared by all callsites.
	llmClient, llmErr := llm.NewAnthropicClient(llm.DefaultClientConfig())
	if llmErr != nil {
		slog.Warn("LLM client not available, triage and knowledge index features disabled", "error", llmErr)
	}

	// Build KnowledgeIndex.
	// BC-05: Wraps existing BM25 index (ONE index, not duplicated).
	// BC-10: Restart-required for cache coherence. No hot-reload.
	// BC-11: Load from pre-baked JSON if available.
	var knowledgeIdx *knowledge.KnowledgeIndex
	if pipelineResult.catalog != nil {
		// H-3: 90-second timeout aligns with ECS health check start period.
		buildCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		knowledgeIdx = buildKnowledgeIndex(buildCtx, pipelineResult, llmClient)
	}

	// Build triage orchestrator (Sprint 5).
	// BC-02: triage.Orchestrator is a new top-level package.
	var triageAdapter internalslack.TriageRunner
	if llmClient != nil && pipelineResult.searchIndex != nil {
		// Create search index adapter for triage.
		triageSearchIdx := &triageSearchAdapter{searchIndex: pipelineResult.searchIndex, catalog: pipelineResult.catalog}
		// StubEmbeddingModel triggers BM25 fallback (BC-06).
		embeddingModel := &triage.StubEmbeddingModel{}
		triageOrch := triage.NewOrchestrator(llmClient, triageSearchIdx, embeddingModel)
		triageAdapter = &triageOrchestratorAdapter{orch: triageOrch}
		slog.Info("triage orchestrator initialized (Sprint 5: BM25 fallback mode)")
	}

	// Create Slack client.
	slackClient := internalslack.NewSlackClient(cfg.SlackBotToken)
	slackCfg := internalslack.DefaultSlackConfig()
	slackCfg.BotToken = cfg.SlackBotToken

	// Construct ConversationManager with LLM summarizer and Slack thread fetcher.
	var convMgr *conversation.Manager
	convConfig := conversation.DefaultConfig()
	var summarizer conversation.Summarizer
	if llmClient != nil {
		summarizer = conversation.NewLLMSummarizer(llmClient)
	}
	// SlackThreadFetcher is nil for now (conversations.replies integration is
	// wired via the SlackClient in a future PR). ConversationManager degrades
	// gracefully without it (DORMANT threads return nil).
	convMgr = conversation.NewManager(convConfig, summarizer, nil, metricsRecorder)
	slog.Info("conversation manager initialized",
		"ttl", convConfig.TTL,
		"max_recent_messages", convConfig.MaxRecentMessages,
		"cleanup_interval", convConfig.CleanupInterval,
	)

	// Construct StreamSender for progressive response rendering.
	streamSender := streaming.NewSender(cfg.SlackBotToken, "")

	// Build TriagePipeline adapter (WARNING-03 fix: passes triage candidates
	// through to Pipeline.QueryWithTriage for BC-07 weighted-mean freshness).
	var triagePipelineAdapter internalslack.TriageQueryRunner
	if pipelineResult.pipeline != nil {
		triagePipelineAdapter = &triagePipelineQueryAdapter{pipeline: pipelineResult.pipeline}
	}

	// Create handler with full dependencies.
	slackHandler, ctxStore, stopDedup := internalslack.NewSlackHandlerWithDeps(internalslack.HandlerDeps{
		Pipeline:        queryRunner,
		Client:          slackClient,
		Config:          slackCfg,
		TriageRunner:    triageAdapter,
		TriagePipeline:  triagePipelineAdapter,
		ConversationMgr: convMgr,
		StreamSender:    streamSender,
		Metrics:         metricsRecorder,
	})

	// Register webhook verification middleware wrapping the Slack handler.
	verifier := webhook.NewVerifier(cfg.SlackSigningSecret)
	srv.RegisterHandler("POST", "/slack/events", verifier.Handler(slackHandler))

	// Wire OTEL tracing middleware.
	srv.Use(observe.OTELMiddleware())

	// Wire concurrency limit middleware.
	srv.Use(serve.ConcurrencyLimit(cfg.MaxConcurrent))

	// Register health checks.
	checker.Register("slack", func(hctx context.Context) error {
		return slackClient.HealthCheck(hctx)
	})
	checker.Register("reasoning", func(_ context.Context) error {
		if pipelineResult.pipeline == nil {
			return fmt.Errorf("reasoning pipeline not initialized")
		}
		return nil
	})

	// Additional health checks for ECS deployment.
	checker.Register("catalog", func(_ context.Context) error {
		if pipelineResult.catalog == nil || pipelineResult.catalog.DomainCount() == 0 {
			return fmt.Errorf("domain catalog not loaded or empty")
		}
		return nil
	})
	checker.Register("search_index", func(_ context.Context) error {
		if pipelineResult.searchIndex == nil {
			return fmt.Errorf("search index not built")
		}
		return nil
	})
	checker.Register("claude_api", func(_ context.Context) error {
		if cfg.AnthropicAPIKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY not configured")
		}
		return nil
	})
	// Knowledge index health check.
	checker.Register("knowledge_index", func(_ context.Context) error {
		if knowledgeIdx == nil {
			return fmt.Errorf("knowledge index not built")
		}
		if knowledgeIdx.DomainCount() == 0 {
			return fmt.Errorf("knowledge index has no domains")
		}
		return nil
	})

	// BC-11: Validate cross-cache coherence before health check goes green.
	// This is fail-open: mismatches are logged as warnings, not errors.
	validateStartupCoherence(pipelineResult.catalog, pipelineResult.searchIndex, knowledgeIdx)

	slog.Info("server configured",
		"port", cfg.Port,
		"drain_timeout", cfg.DrainTimeout,
		"max_concurrent", cfg.MaxConcurrent,
		"pipeline_ready", pipelineResult.pipeline != nil,
		"otel_endpoint", cfg.OTELEndpoint,
	)

	// Start blocks until shutdown signal.
	if err := srv.Start(context.Background()); err != nil {
		stopDedup()
		ctxStore.Stop()
		if convMgr != nil {
			convMgr.Stop()
		}
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeServerStartFailed, "server failed to start", err))
	}

	// Stop background goroutines on graceful shutdown.
	stopDedup()
	ctxStore.Stop()
	if convMgr != nil {
		convMgr.Stop()
	}

	return nil
}

// resolveOrgContext resolves the org from the --org flag, KNOSSOS_ORG env var, or active-org file.
// Returns nil if no org is configured -- this is valid (pure env-var mode).
func resolveOrgContext(orgFlag string) config.OrgContext {
	orgName := orgFlag
	if orgName == "" {
		orgName = config.ActiveOrg()
	}
	if orgName == "" {
		slog.Debug("no org configured, skipping org env file layer")
		return nil
	}

	orgCtx, err := config.NewOrgContext(orgName)
	if err != nil {
		slog.Warn("failed to create org context, skipping org env file layer",
			"org", orgName, "error", err)
		return nil
	}

	slog.Info("org context resolved", "org", orgName)
	return orgCtx
}

// pipelineComponents holds intermediate pipeline components for health check registration.
type pipelineComponents struct {
	pipeline    *reason.Pipeline
	catalog     *registryorg.DomainCatalog
	searchIndex *search.SearchIndex
}

// buildPipeline constructs the full reasoning pipeline from environment configuration.
// Returns a pipelineComponents struct; pipeline is nil if critical dependencies
// (Claude API key) are unavailable. Non-critical dependencies degrade gracefully.
func buildPipeline() pipelineComponents {
	result := pipelineComponents{}

	// Step 1: Intent classifier (no external dependencies).
	classifier := intent.NewClassifier()

	// Step 2: Token counter (used by context assembler).
	counter, err := tokenizer.New()
	if err != nil {
		slog.Warn("tokenizer initialization failed, pipeline disabled", "error", err)
		return result
	}

	// Step 3: Context assembler.
	reasoningCfg := reason.DefaultReasoningConfig()
	assembler := reasoncontext.NewAssembler(counter, reasoningCfg.Assembler)

	// Step 4: Claude client (requires ANTHROPIC_API_KEY).
	claudeClient, err := response.NewAnthropicClient()
	if err != nil {
		slog.Warn("claude client initialization failed, pipeline disabled", "error", err)
		return result
	}

	// Step 5: Response generator.
	generator := response.NewGenerator(claudeClient, reasoningCfg.Generator)

	// Step 6: Trust scorer.
	trustCfg := trust.DefaultConfig()
	scorer := trust.NewScorer(trustCfg)

	// Step 7: Search index (uses a minimal cobra root; knowledge domains are the key channel).
	minimalRoot := &cobra.Command{Use: "ari"}
	result.searchIndex = search.Build(minimalRoot, nil)

	// Step 8: Load DomainCatalog from registry for provenance chains (TD-01 fix).
	// Fail-open: nil catalog means provenance chains will be empty but pipeline still works.
	if orgCtx, err := config.DefaultOrgContext(); err == nil {
		catalogPath := registryorg.CatalogPath(orgCtx)
		if loaded, err := registryorg.LoadCatalog(catalogPath); err == nil {
			result.catalog = loaded
			slog.Info("domain catalog loaded", "org", orgCtx.Name(), "domains", loaded.DomainCount())
		} else {
			slog.Warn("domain catalog not found, provenance chains will be empty", "error", err)
		}
	} else {
		slog.Debug("no org context configured, provenance chains will be empty")
	}

	result.pipeline = reason.NewPipeline(
		classifier,
		assembler,
		generator,
		scorer,
		result.searchIndex,
		result.catalog,
		reasoningCfg,
	)

	return result
}

// ---- Triage adapters (bridges triage package types to handler types) ----

// triageSearchAdapter adapts *search.SearchIndex to triage.SearchIndex interface.
type triageSearchAdapter struct {
	searchIndex *search.SearchIndex
	catalog     *registryorg.DomainCatalog
}

func (a *triageSearchAdapter) SearchByBM25(query string, k int) []triage.BM25Result {
	if a.searchIndex == nil || !a.searchIndex.HasBM25() {
		return nil
	}

	results := a.searchIndex.Search(query, search.SearchOptions{
		Limit:   k,
		Domains: []search.Domain{search.DomainKnowledge},
	})

	var out []triage.BM25Result
	for _, r := range results {
		if r.Domain == search.DomainKnowledge {
			out = append(out, triage.BM25Result{
				QualifiedName: r.Name,
				Score:         float64(r.Score),
				Domain:        string(r.Domain),
				RawText:       r.Description,
			})
		}
	}
	return out
}

func (a *triageSearchAdapter) GetMetadata(qualifiedName string) (*triage.DomainMetadata, bool) {
	if a.catalog == nil {
		return nil, false
	}

	entry, ok := a.catalog.LookupDomain(qualifiedName)
	if !ok {
		return nil, false
	}

	return &triage.DomainMetadata{
		QualifiedName:  entry.QualifiedName,
		DomainType:     entry.Domain,
		Repo:           repoFromQualifiedName(entry.QualifiedName),
		FreshnessScore: 0, // Tier 1: freshness computed at query time, not stored.
		GeneratedAt:    entry.GeneratedAt,
	}, true
}

func (a *triageSearchAdapter) ListAllDomains() []triage.DomainMetadata {
	if a.catalog == nil {
		return nil
	}

	domains := a.catalog.ListDomains()
	var out []triage.DomainMetadata
	for _, d := range domains {
		out = append(out, triage.DomainMetadata{
			QualifiedName:  d.QualifiedName,
			DomainType:     d.Domain,
			Repo:           repoFromQualifiedName(d.QualifiedName),
			FreshnessScore: 0, // Tier 1: freshness computed at query time.
			GeneratedAt:    d.GeneratedAt,
		})
	}
	return out
}

// triageOrchestratorAdapter adapts *triage.Orchestrator to internalslack.TriageRunner.
type triageOrchestratorAdapter struct {
	orch *triage.Orchestrator
}

func (a *triageOrchestratorAdapter) Assess(ctx context.Context, query string, threadHistory []internalslack.TriageThreadMessage) (*internalslack.TriageResultData, error) {
	// Convert handler thread messages to triage thread messages.
	var triageHistory []triage.ThreadMessage
	for _, m := range threadHistory {
		triageHistory = append(triageHistory, triage.ThreadMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}

	result, err := a.orch.Assess(ctx, query, triageHistory)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	// Convert triage result to handler type.
	data := &internalslack.TriageResultData{
		RefinedQuery:   result.RefinedQuery,
		ModelCallCount: result.ModelCallCount,
	}
	for _, c := range result.Candidates {
		data.Candidates = append(data.Candidates, internalslack.TriageCandidateData{
			QualifiedName:       c.QualifiedName,
			RelevanceScore:      c.RelevanceScore,
			EmbeddingSimilarity: c.EmbeddingSimilarity,
			Freshness:           c.Freshness,
			Rationale:           c.Rationale,
			DomainType:          c.DomainType,
			RelatedDomains:      c.RelatedDomains,
		})
	}
	return data, nil
}

// repoFromQualifiedName extracts the repo from "org::repo::domain".
func repoFromQualifiedName(qn string) string {
	parts := strings.SplitN(qn, "::", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// ---- Sprint 7: KnowledgeIndex build ----

// buildKnowledgeIndex constructs the KnowledgeIndex during server startup.
// BC-05: Wraps the existing BM25 index (ONE index, not duplicated).
// BC-10: Restart-required for cache coherence. No hot-reload.
// BC-11: Loads from pre-baked JSON if available.
// D-L5: Eager rebuild on source_hash change.
func buildKnowledgeIndex(ctx context.Context, pr pipelineComponents, llmClient *llm.AnthropicClient) *knowledge.KnowledgeIndex {
	// Build catalog adapter.
	if pr.catalog == nil {
		slog.Warn("no domain catalog, knowledge index disabled")
		return nil
	}

	catalogAdapter := &knowledgeCatalogAdapter{catalog: pr.catalog}
	contentAdapter := resolveKnowledgeContentStore()

	// Build BM25 adapter (wraps the existing BM25 index from search.SearchIndex).
	var bm25Adapter knowledge.BM25Searcher
	if pr.searchIndex != nil && pr.searchIndex.HasBM25() {
		bm25Adapter = &knowledgeBM25Adapter{searchIndex: pr.searchIndex}
	}

	// Build LLM adapter.
	var kiLLMClient knowledge.LLMClient
	if llmClient != nil {
		kiLLMClient = &knowledgeLLMAdapter{client: llmClient}
	}

	// Resolve persisted path: prefer env var, then default container path.
	persistedPath := knowledge.DefaultPersistedPath
	if envPath := os.Getenv("CLEW_KNOWLEDGE_INDEX_PATH"); envPath != "" {
		persistedPath = envPath
	}

	cfg := knowledge.BuildConfig{
		Catalog:       catalogAdapter,
		ContentStore:  contentAdapter,
		LLMClient:     kiLLMClient,
		PersistedPath: persistedPath,
		BM25Index:     bm25Adapter,
	}

	idx, err := knowledge.Build(ctx, cfg)
	if err != nil {
		slog.Warn("knowledge index build failed", "error", err)
		return nil
	}

	slog.Info("knowledge index ready",
		"domains", idx.DomainCount(),
		"summaries", idx.SummaryCount(),
		"embeddings", idx.EmbeddingCount(),
		"edges", idx.EdgeCount(),
	)

	return idx
}

// validateStartupCoherence checks cross-cache consistency after all indexes are built.
// BC-11: All caches should be derived from the same build generation.
// This is fail-open: mismatches are logged as warnings, not errors.
// See HANDOFF Section 2.5: 3 consistency invariants.
func validateStartupCoherence(catalog *registryorg.DomainCatalog, searchIndex *search.SearchIndex, knowledgeIdx *knowledge.KnowledgeIndex) {
	if catalog == nil {
		slog.Warn("startup coherence: catalog is nil, skipping validation")
		return
	}

	domains := catalog.ListDomains()
	totalDomains := len(domains)
	if totalDomains == 0 {
		slog.Warn("startup coherence: catalog has 0 domains")
		return
	}

	// Resolve content store for validation (same resolution as buildKnowledgeIndex).
	contentStore := resolveKnowledgeContentStore()

	// Invariant 1: Every cataloged domain should have content available.
	var contentMissing []string
	if contentStore != nil {
		for _, d := range domains {
			if !contentStore.HasContent(d.QualifiedName) {
				contentMissing = append(contentMissing, d.QualifiedName)
			}
		}
	}

	// Invariant 2: KnowledgeIndex should have metadata for every cataloged domain.
	var knowledgeMissing []string
	if knowledgeIdx != nil {
		for _, d := range domains {
			if _, ok := knowledgeIdx.GetMetadata(d.QualifiedName); !ok {
				knowledgeMissing = append(knowledgeMissing, d.QualifiedName)
			}
		}
	}

	// Invariant 3: BM25 index should be available if content exists.
	bm25Available := searchIndex != nil && searchIndex.HasBM25()
	if contentStore != nil && !bm25Available {
		slog.Warn("startup coherence: content available but BM25 index not built")
	}

	// Log results.
	if len(contentMissing) > 0 {
		slog.Warn("startup coherence: catalog-content mismatch",
			"missing_count", len(contentMissing),
			"total_domains", totalDomains,
			"missing_domains", contentMissing,
		)
	}
	if len(knowledgeMissing) > 0 {
		slog.Warn("startup coherence: catalog-knowledge mismatch",
			"missing_count", len(knowledgeMissing),
			"total_domains", totalDomains,
			"missing_domains", knowledgeMissing,
		)
	}

	coherent := len(contentMissing) == 0 && len(knowledgeMissing) == 0
	knowledgeDomains := 0
	if knowledgeIdx != nil {
		knowledgeDomains = knowledgeIdx.DomainCount()
	}
	slog.Info("startup coherence validation complete",
		"coherent", coherent,
		"total_domains", totalDomains,
		"content_missing", len(contentMissing),
		"knowledge_missing", len(knowledgeMissing),
		"bm25_available", bm25Available,
		"knowledge_domains", knowledgeDomains,
	)
}

// resolveKnowledgeContentStore returns a content store adapter for the knowledge index.
// Uses the same resolution strategy as the BM25 content loader.
func resolveKnowledgeContentStore() knowledge.ContentStore {
	// Check env var override first.
	if envDir := os.Getenv("CLEW_CONTENT_DIR"); envDir != "" {
		if info, err := os.Stat(envDir); err == nil && info.IsDir() {
			return &knowledgeContentAdapter{contentDir: envDir}
		}
	}

	// Check for pre-baked content directory.
	defaultDir := "/data/content"
	if info, err := os.Stat(defaultDir); err == nil && info.IsDir() {
		return &knowledgeContentAdapter{contentDir: defaultDir}
	}

	// No content source -- return nil (Build handles gracefully).
	return nil
}

// ---- Sprint 7: Knowledge index adapter types ----

// knowledgeCatalogAdapter adapts *registryorg.DomainCatalog to knowledge.DomainCatalog.
type knowledgeCatalogAdapter struct {
	catalog *registryorg.DomainCatalog
}

func (a *knowledgeCatalogAdapter) ListDomains() []knowledge.CatalogDomainEntry {
	domains := a.catalog.ListDomains()
	out := make([]knowledge.CatalogDomainEntry, len(domains))
	for i, d := range domains {
		out[i] = knowledge.CatalogDomainEntry{
			QualifiedName: d.QualifiedName,
			Domain:        d.Domain,
			Path:          d.Path,
			GeneratedAt:   d.GeneratedAt,
			ExpiresAfter:  d.ExpiresAfter,
			SourceHash:    d.SourceHash,
			Confidence:    d.Confidence,
		}
	}
	return out
}

func (a *knowledgeCatalogAdapter) LookupDomain(qualifiedName string) (knowledge.CatalogDomainEntry, bool) {
	d, ok := a.catalog.LookupDomain(qualifiedName)
	if !ok {
		return knowledge.CatalogDomainEntry{}, false
	}
	return knowledge.CatalogDomainEntry{
		QualifiedName: d.QualifiedName,
		Domain:        d.Domain,
		Path:          d.Path,
		GeneratedAt:   d.GeneratedAt,
		ExpiresAfter:  d.ExpiresAfter,
		SourceHash:    d.SourceHash,
		Confidence:    d.Confidence,
	}, true
}

func (a *knowledgeCatalogAdapter) DomainCount() int {
	return a.catalog.DomainCount()
}

// knowledgeContentAdapter adapts a pre-baked content directory to knowledge.ContentStore.
// RR-007: knowledge/ sub-packages do not import content/ directly.
type knowledgeContentAdapter struct {
	contentDir string
}

func (a *knowledgeContentAdapter) LoadContent(qualifiedName string) (string, error) {
	// Resolve path: look up the domain in the catalog to get the file path.
	// Since we don't have direct access to the catalog here, we use the
	// qualifiedName to derive the repo and search for the content.
	parts := strings.SplitN(qualifiedName, "::", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid qualified name: %s", qualifiedName)
	}
	repoName := parts[1]
	domainName := parts[2]

	// Try common .know/ path patterns.
	candidates := []string{
		fmt.Sprintf("%s/%s/.know/%s.md", a.contentDir, repoName, domainName),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return stripFrontmatter(string(data)), nil
	}

	return "", fmt.Errorf("content not found for %s", qualifiedName)
}

func (a *knowledgeContentAdapter) HasContent(qualifiedName string) bool {
	content, err := a.LoadContent(qualifiedName)
	return err == nil && content != ""
}

// stripFrontmatter removes YAML frontmatter delimited by ---.
func stripFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---") {
		return text
	}
	rest := text[3:]
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return text
	}
	return strings.TrimSpace(rest[idx+3:])
}

// knowledgeBM25Adapter adapts *search.SearchIndex to knowledge.BM25Searcher.
// BC-05: ONE BM25 index -- the knowledge package wraps the existing one.
type knowledgeBM25Adapter struct {
	searchIndex *search.SearchIndex
}

func (a *knowledgeBM25Adapter) SearchDocuments(query string, k int) []knowledge.BM25SearchHit {
	results := a.searchIndex.Search(query, search.SearchOptions{
		Limit:   k,
		Domains: []search.Domain{search.DomainKnowledge},
	})

	var hits []knowledge.BM25SearchHit
	for _, r := range results {
		if r.Domain == search.DomainKnowledge {
			hits = append(hits, knowledge.BM25SearchHit{
				QualifiedName: r.Name,
				Score:         float64(r.Score),
				Domain:        string(r.Domain),
				RawText:       r.Description,
				MatchType:     "document",
			})
		}
	}
	return hits
}

func (a *knowledgeBM25Adapter) SearchSections(query string, k int) []knowledge.BM25SearchHit {
	// The existing search index merges doc and section results via RRF.
	// For the knowledge index's BM25 wrapper, we return the same results
	// and let the coordinator handle dedup.
	return nil
}

// knowledgeLLMAdapter adapts *llm.AnthropicClient to knowledge.LLMClient.
type knowledgeLLMAdapter struct {
	client *llm.AnthropicClient
}

func (a *knowledgeLLMAdapter) Complete(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error) {
	return a.client.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserMessage:  userMessage,
		MaxTokens:    maxTokens,
	})
}

// ---- WARNING-03 fix: TriagePipeline adapter ----

// triagePipelineQueryAdapter adapts *reason.Pipeline to internalslack.TriageQueryRunner.
// This bridge converts handler-local TriageResultInputData to reason.TriageResultInput
// so that triage candidates (RelevanceScores, Freshness) reach the assembler for
// BC-07 weighted-mean freshness computation.
type triagePipelineQueryAdapter struct {
	pipeline *reason.Pipeline
}

func (a *triagePipelineQueryAdapter) QueryWithTriage(ctx context.Context, triageInput *internalslack.TriageResultInputData) (*response.ReasoningResponse, error) {
	if triageInput == nil {
		return a.pipeline.Query(ctx, "")
	}

	// Convert handler-local types to reason/ types.
	candidates := make([]reason.TriageCandidateInput, len(triageInput.Candidates))
	for i, c := range triageInput.Candidates {
		candidates[i] = reason.TriageCandidateInput{
			QualifiedName:       c.QualifiedName,
			RelevanceScore:      c.RelevanceScore,
			EmbeddingSimilarity: c.EmbeddingSimilarity,
			Freshness:           c.Freshness,
			Rationale:           c.Rationale,
			DomainType:          c.DomainType,
			RelatedDomains:      c.RelatedDomains,
		}
	}

	return a.pipeline.QueryWithTriage(ctx, &reason.TriageResultInput{
		RefinedQuery:   triageInput.RefinedQuery,
		Candidates:     candidates,
		ModelCallCount: triageInput.ModelCallCount,
	})
}
