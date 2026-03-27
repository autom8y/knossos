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

// InstrumentedPipeline wraps a QueryRunner with OTEL tracing, cost tracking,
// and Thermia metrics recording.
// It satisfies the slack.QueryRunner interface, allowing it to be substituted
// for the raw pipeline without modifying frozen packages.
type InstrumentedPipeline struct {
	inner   slack.QueryRunner
	tracer  trace.Tracer
	cost    *CostTracker
	metrics MetricsRecorder
}

// NewInstrumentedPipeline creates an instrumented wrapper around the inner pipeline.
// metrics may be nil (metrics recording will be skipped).
func NewInstrumentedPipeline(inner slack.QueryRunner, cost *CostTracker, metrics MetricsRecorder) *InstrumentedPipeline {
	if metrics == nil {
		metrics = NopRecorder{}
	}
	return &InstrumentedPipeline{
		inner:   inner,
		tracer:  Tracer("clew.pipeline"),
		cost:    cost,
		metrics: metrics,
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
	elapsed := time.Since(start)
	durationMs := elapsed.Milliseconds()

	span.SetAttributes(attribute.Int64("clew.duration_ms", durationMs))

	// Record end-to-end query latency metric (cm_path unknown at this layer, use "unknown").
	p.metrics.RecordQueryLatency("unknown", elapsed)

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

		// CE mechanism metrics from assembly diagnostics.
		if resp.CEDiagnostics != nil {
			diag := resp.CEDiagnostics
			for i := 0; i < diag.SectionCandidatesPacked; i++ {
				p.metrics.IncrSectionCandidate()
			}
			for _, e := range diag.DiversityFloorEvents {
				p.metrics.IncrDiversityFloorEnforced(e.FloorType)
			}
			for _, h := range diag.TypeCeilingHits {
				if h.Skipped {
					p.metrics.IncrTypeCeilingHit(h.DomainType)
				}
			}
			totalTokens := 0
			for _, v := range diag.TypeTokenBreakdown {
				totalTokens += v
			}
			if totalTokens > 0 {
				for typ, tokens := range diag.TypeTokenBreakdown {
					fraction := float64(tokens) / float64(totalTokens)
					p.metrics.RecordAssemblerTypeFraction(typ, fraction)
				}
			}
		}
	}

	return resp, nil
}
