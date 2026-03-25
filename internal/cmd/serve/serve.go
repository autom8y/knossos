// Package serve implements the ari serve command for the HTTP webhook server.
package serve

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/envload"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/observe"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/reason"
	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/intent"
	"github.com/autom8y/knossos/internal/reason/response"
	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/autom8y/knossos/internal/search"
	"github.com/autom8y/knossos/internal/serve"
	"github.com/autom8y/knossos/internal/serve/health"
	"github.com/autom8y/knossos/internal/serve/webhook"
	internalslack "github.com/autom8y/knossos/internal/slack"
	"github.com/autom8y/knossos/internal/tokenizer"
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

	// Create cost tracker and instrumented pipeline wrapper.
	costTracker := observe.NewCostTracker()
	var queryRunner internalslack.QueryRunner
	if pipelineResult.pipeline != nil {
		queryRunner = observe.NewInstrumentedPipeline(pipelineResult.pipeline, costTracker)
	}

	// Create Slack client and handler.
	slackClient := internalslack.NewSlackClient(cfg.SlackBotToken)
	slackCfg := internalslack.DefaultSlackConfig()
	slackCfg.BotToken = cfg.SlackBotToken
	slackHandler, _ := internalslack.NewSlackHandler(queryRunner, slackClient, slackCfg)

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

	slog.Info("server configured",
		"port", cfg.Port,
		"drain_timeout", cfg.DrainTimeout,
		"max_concurrent", cfg.MaxConcurrent,
		"pipeline_ready", pipelineResult.pipeline != nil,
		"otel_endpoint", cfg.OTELEndpoint,
	)

	// Start blocks until shutdown signal.
	if err := srv.Start(context.Background()); err != nil {
		return common.PrintAndReturn(printer,
			errors.Wrap(errors.CodeServerStartFailed, "server failed to start", err))
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
