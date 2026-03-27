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
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/llm"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/reason"
	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/search/bm25"
	"github.com/autom8y/knossos/internal/search/knowledge"
	"github.com/autom8y/knossos/internal/triage"
)

// queryOptions holds flag values for the query subcommand.
type queryOptions struct {
	org        string
	diagnostic bool
	noTriage   bool
}

// newQueryCmd creates the serve query subcommand.
func newQueryCmd(ctx *cmdContext) *cobra.Command {
	opts := queryOptions{}

	cmd := &cobra.Command{
		Use:   "query [question]",
		Short: "Run a query through the Clew reasoning pipeline locally",
		Long: `Execute a query through the full Clew intelligence pipeline without Slack,
Docker, or ECS. Exercises the same code paths as the Slack handler: triage
orchestrator, pipeline.QueryWithTriage, trust scoring, and response generation.

This command is the primary iteration tool for Clew development. The feedback
loop drops from 15 minutes (Docker build + ECS deploy + Slack) to ~5 seconds.

Configuration is resolved via the same hierarchy as 'ari serve':
  - ANTHROPIC_API_KEY must be set (required for reasoning)
  - Org context from --org flag, KNOSSOS_ORG env var, or active org
  - Knowledge index loaded from pre-baked JSON or built synchronously

Examples:
  ari serve query "How is knossos structured?"
  ari serve query --diagnostic "What are the scar tissue patterns?"
  ari serve query --no-triage "Explain the sync pipeline"
  ari serve query --org autom8y "What testing conventions exist?"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuery(ctx, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.org, "org", "",
		"Organization name (env: KNOSSOS_ORG, default: active org)")
	cmd.Flags().BoolVar(&opts.diagnostic, "diagnostic", false,
		"Show detailed breakdown: triage stages, trust scores, token counts, per-stage latency")
	cmd.Flags().BoolVar(&opts.noTriage, "no-triage", false,
		"Skip triage and use the v1 direct query path (Pipeline.Query)")

	return cmd
}

// runQuery executes a single query through the reasoning pipeline.
func runQuery(ctx *cmdContext, opts queryOptions, question string) error {
	printer := ctx.GetPrinter(output.FormatText)
	totalStart := time.Now()

	// Track per-stage latency for diagnostics.
	var timings queryTimings

	// Step 1: Build the reasoning pipeline (reuses serve.go buildPipeline).
	stageStart := time.Now()
	pipelineResult := buildPipeline()
	timings.pipelineBuild = time.Since(stageStart)

	if pipelineResult.pipeline == nil {
		return common.PrintAndReturn(printer,
			errors.New(errors.CodeUsageError, "reasoning pipeline could not be initialized; ensure ANTHROPIC_API_KEY is set"))
	}

	// Step 2: Build LLM client for triage (shared with knowledge index).
	var llmClient *llm.AnthropicClient
	var triageOrch *triage.Orchestrator
	if !opts.noTriage {
		stageStart = time.Now()
		var llmErr error
		llmClient, llmErr = llm.NewAnthropicClient(llm.DefaultClientConfig())
		if llmErr != nil {
			slog.Warn("LLM client not available, triage disabled", "error", llmErr)
		}

		if llmClient != nil && pipelineResult.searchIndex != nil {
			// Build Clew-specific BM25 index with higher length normalization.
			var clewIdx *bm25.Index
			if pipelineResult.searchIndex.HasBM25() {
				clewIdx = buildClewBM25Index(pipelineResult.catalog)
			}
			triageSearchIdx := &triageSearchAdapter{
				searchIndex:     pipelineResult.searchIndex,
				clewBM25:        clewIdx,
				catalog:         pipelineResult.catalog,
				knowledgeIdxPtr: pipelineResult.knowledgeIdxPtr,
			}
			embeddingModel := &triage.StubEmbeddingModel{}
			triageOrch = triage.NewOrchestrator(llmClient, triageSearchIdx, embeddingModel)
		}
		timings.triageBuild = time.Since(stageStart)
	}

	// Step 3: Load or build knowledge index synchronously (CLI blocks until ready).
	stageStart = time.Now()
	knowledgeIdx := loadPrebakedKnowledgeIndex()
	if knowledgeIdx == nil && pipelineResult.catalog != nil && llmClient != nil {
		slog.Info("no pre-baked knowledge index, building synchronously")
		buildCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		knowledgeIdx = buildKnowledgeIndex(buildCtx, pipelineResult, llmClient)
	}
	timings.knowledgeIndex = time.Since(stageStart)

	// Print diagnostic header if requested.
	if opts.diagnostic {
		fmt.Fprintf(os.Stdout, "\n%s Query %s\n", diagSep, strings.Repeat("━", 39))
		fmt.Fprintf(os.Stdout, "%q\n", question)
	}

	// Step 4: Run triage (if enabled).
	var triageResult *triage.TriageResult
	var triageInput *reason.TriageResultInput
	if triageOrch != nil && !opts.noTriage {
		stageStart = time.Now()
		var triageErr error
		triageResult, triageErr = triageOrch.Assess(context.Background(), question, nil)
		timings.triage = time.Since(stageStart)

		if triageErr != nil {
			slog.Warn("triage assessment failed, falling back to v1 path", "error", triageErr)
		} else if triageResult != nil {
			triageInput = convertTriageResult(triageResult)
		}

		if opts.diagnostic {
			printTriageDiagnostic(triageResult, timings.triage)
		}
	}

	// Step 5: Execute the query through the pipeline.
	stageStart = time.Now()
	var resp *response.ReasoningResponse
	var queryErr error

	if triageInput != nil && len(triageInput.Candidates) > 0 {
		resp, queryErr = pipelineResult.pipeline.QueryWithTriage(context.Background(), triageInput)
	} else {
		resp, queryErr = pipelineResult.pipeline.Query(context.Background(), question)
	}
	timings.generation = time.Since(stageStart)

	if queryErr != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeUsageError, "pipeline query failed", queryErr))
	}

	timings.total = time.Since(totalStart)

	// Step 6: Format and print the output.
	if opts.diagnostic {
		printDiagnosticOutput(resp, pipelineResult, knowledgeIdx, timings)
	} else {
		printStandardOutput(resp)
	}

	return nil
}

// queryTimings tracks per-stage latency for diagnostic output.
type queryTimings struct {
	pipelineBuild  time.Duration
	triageBuild    time.Duration
	knowledgeIndex time.Duration
	triage         time.Duration
	generation     time.Duration
	total          time.Duration
}

// diagSep is the diagnostic section separator.
const diagSep = "━━━"

// printStandardOutput prints the response in the default (non-diagnostic) format.
func printStandardOutput(resp *response.ReasoningResponse) {
	// Print the answer.
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, resp.Answer)

	// Print citations if present.
	if len(resp.Citations) > 0 {
		fmt.Fprintf(os.Stdout, "\nSources:\n")
		for _, c := range resp.Citations {
			if c.Section != "" {
				fmt.Fprintf(os.Stdout, "  - %s > %s\n", c.QualifiedName, c.Section)
			} else {
				fmt.Fprintf(os.Stdout, "  - %s\n", c.QualifiedName)
			}
		}
	}

	// Print trust tier.
	fmt.Fprintf(os.Stdout, "\nTrust: %s (confidence: %.3f)\n", resp.Tier.String(), resp.Confidence.Overall)
}

// printDiagnosticOutput prints the full diagnostic breakdown.
func printDiagnosticOutput(
	resp *response.ReasoningResponse,
	pr pipelineComponents,
	knowledgeIdx *knowledge.KnowledgeIndex,
	timings queryTimings,
) {
	w := os.Stdout

	// Trust section.
	fmt.Fprintf(w, "\n%s Trust %s\n", diagSep, strings.Repeat("━", 39))
	fmt.Fprintf(w, "Tier: %s | Confidence: %.3f\n", resp.Tier.String(), resp.Confidence.Overall)
	fmt.Fprintf(w, "  Retrieval: %.2f | Freshness: %.2f | Coverage: %.2f\n",
		resp.Confidence.Retrieval, resp.Confidence.Freshness, resp.Confidence.Coverage)

	// Context assembly section.
	fmt.Fprintf(w, "\n%s Context Assembly %s\n", diagSep, strings.Repeat("━", 31))
	if len(resp.Citations) > 0 {
		fmt.Fprintf(w, "Sources cited: %d\n", len(resp.Citations))
		for i, c := range resp.Citations {
			if c.Section != "" {
				fmt.Fprintf(w, "  #%d %s > %s\n", i+1, c.QualifiedName, c.Section)
			} else {
				fmt.Fprintf(w, "  #%d %s\n", i+1, c.QualifiedName)
			}
		}
	} else {
		fmt.Fprintf(w, "Sources cited: 0\n")
	}

	// Provenance chain.
	if resp.Provenance != nil && len(resp.Provenance.Sources) > 0 {
		fmt.Fprintf(w, "Provenance chain: %d sources\n", len(resp.Provenance.Sources))
		for _, s := range resp.Provenance.Sources {
			fmt.Fprintf(w, "  - %s (freshness: %.2f)\n", s.QualifiedName, s.FreshnessAtQuery)
		}
	}

	// Response section.
	fmt.Fprintf(w, "\n%s Response %s\n", diagSep, strings.Repeat("━", 37))
	fmt.Fprintln(w, resp.Answer)

	// Gap admission (if LOW tier).
	if resp.Gap != nil {
		fmt.Fprintf(w, "\n%s Gap Admission %s\n", diagSep, strings.Repeat("━", 33))
		fmt.Fprintf(w, "Reason: %s\n", resp.Gap.Reason)
		if len(resp.Gap.Suggestions) > 0 {
			fmt.Fprintf(w, "Suggestions:\n")
			for _, s := range resp.Gap.Suggestions {
				fmt.Fprintf(w, "  - %s\n", s)
			}
		}
	}

	// Metrics section.
	fmt.Fprintf(w, "\n%s Metrics %s\n", diagSep, strings.Repeat("━", 38))
	fmt.Fprintf(w, "Total: %s | Pipeline build: %s | Triage: %s | Generation: %s\n",
		formatDuration(timings.total),
		formatDuration(timings.pipelineBuild),
		formatDuration(timings.triage),
		formatDuration(timings.generation),
	)

	if resp.TokensUsed.TotalTokens > 0 {
		fmt.Fprintf(w, "Tokens: %d in / %d out | Est. cost: $%.4f\n",
			resp.TokensUsed.PromptTokens,
			resp.TokensUsed.CompletionTokens,
			resp.TokensUsed.EstimatedCostUSD,
		)
	}

	// Index stats.
	fmt.Fprintf(w, "Search index: %s | Knowledge index: %s | Catalog: %s\n",
		boolReady(pr.searchIndex != nil),
		boolReady(knowledgeIdx != nil),
		boolReady(pr.catalog != nil),
	)

	if pr.catalog != nil {
		fmt.Fprintf(w, "Catalog domains: %d\n", pr.catalog.DomainCount())
	}
	if pr.searchIndex != nil {
		fmt.Fprintf(w, "BM25 available: %v\n", pr.searchIndex.HasBM25())
	}

	// Intent.
	fmt.Fprintf(w, "Intent: tier=%s answerable=%v domains=%v\n",
		resp.Intent.Tier, resp.Intent.Answerable, resp.Intent.Domains)

	if resp.Degraded {
		fmt.Fprintf(w, "DEGRADED: %s\n", resp.DegradedReason)
	}

	fmt.Fprintln(w)
}

// printTriageDiagnostic prints triage-specific diagnostic information.
func printTriageDiagnostic(result *triage.TriageResult, elapsed time.Duration) {
	w := os.Stdout
	fmt.Fprintf(w, "\n%s Triage %s\n", diagSep, strings.Repeat("━", 38))

	if result == nil {
		fmt.Fprintf(w, "Triage: skipped or returned no results (%s)\n", formatDuration(elapsed))
		return
	}

	fmt.Fprintf(w, "Refined query: %q\n", result.RefinedQuery)
	fmt.Fprintf(w, "Candidates: %d | Model calls: %d | Latency: %s\n",
		len(result.Candidates), result.ModelCallCount, formatDuration(elapsed))

	for i, c := range result.Candidates {
		fmt.Fprintf(w, "  #%d %-45s score=%.2f freshness=%.2f\n",
			i+1, c.QualifiedName, c.RelevanceScore, c.Freshness)
		if c.Rationale != "" {
			fmt.Fprintf(w, "     %s\n", c.Rationale)
		}
	}
}

// convertTriageResult converts a triage.TriageResult to a reason.TriageResultInput.
// This is the same conversion done by the Slack handler's triagePipelineQueryAdapter
// and triageOrchestratorAdapter.
func convertTriageResult(result *triage.TriageResult) *reason.TriageResultInput {
	if result == nil {
		return nil
	}

	candidates := make([]reason.TriageCandidateInput, len(result.Candidates))
	for i, c := range result.Candidates {
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

	return &reason.TriageResultInput{
		RefinedQuery:   result.RefinedQuery,
		Candidates:     candidates,
		ModelCallCount: result.ModelCallCount,
	}
}

// formatDuration formats a duration for human display.
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dus", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// boolReady returns a human-readable readiness indicator.
func boolReady(ready bool) string {
	if ready {
		return "ready"
	}
	return "not available"
}
