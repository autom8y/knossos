package observe

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/slack"
)

// Compile-time check: InstrumentedPipeline implements slack.QueryRunner.
var _ slack.QueryRunner = (*InstrumentedPipeline)(nil)

// InstrumentedPipeline wraps a QueryRunner with OTEL tracing and cost tracking.
// It satisfies the slack.QueryRunner interface, allowing it to be substituted
// for the raw pipeline without modifying frozen packages.
type InstrumentedPipeline struct {
	inner  slack.QueryRunner
	tracer trace.Tracer
	cost   *CostTracker
}

// NewInstrumentedPipeline creates an instrumented wrapper around the inner pipeline.
func NewInstrumentedPipeline(inner slack.QueryRunner, cost *CostTracker) *InstrumentedPipeline {
	return &InstrumentedPipeline{
		inner:  inner,
		tracer: Tracer("clew.pipeline"),
		cost:   cost,
	}
}

// Query delegates to the inner pipeline, recording a child span with
// OTEL GenAI conventions, confidence metrics, and cost data.
func (p *InstrumentedPipeline) Query(ctx context.Context, question string) (*response.ReasoningResponse, error) {
	ctx, span := p.tracer.Start(ctx, "clew.pipeline.query",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	// Truncate question for span attribute (avoid unbounded attribute values).
	truncated := question
	if len(truncated) > 256 {
		truncated = truncated[:256] + "..."
	}
	span.SetAttributes(attribute.String("clew.question", truncated))

	start := time.Now()
	resp, err := p.inner.Query(ctx, question)
	durationMs := time.Since(start).Milliseconds()

	span.SetAttributes(attribute.Int64("clew.duration_ms", durationMs))

	if err != nil {
		span.RecordError(err)
		return resp, err
	}

	if resp != nil {
		// Confidence attributes.
		span.SetAttributes(
			attribute.String("clew.confidence.tier", resp.Tier.String()),
			attribute.Float64("clew.confidence.overall", resp.Confidence.Overall),
		)

		// OTEL GenAI semantic conventions.
		span.SetAttributes(
			attribute.Int("gen_ai.usage.input_tokens", resp.TokensUsed.PromptTokens),
			attribute.Int("gen_ai.usage.output_tokens", resp.TokensUsed.CompletionTokens),
			attribute.Float64("gen_ai.usage.cost_usd", resp.TokensUsed.EstimatedCostUSD),
		)

		// Track cost.
		if p.cost != nil {
			p.cost.Record(resp.TokensUsed)
		}
	}

	return resp, nil
}
