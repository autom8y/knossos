---
domain: feat/observability-platform
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/observe/**/*.go"
  - "./deploy/terraform/logging.tf"
  - "./deploy/terraform/ecs.tf"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# CloudWatch EMF Metrics and OTEL Tracing

## Purpose and Design Rationale

Two complementary, independently degradable signals for ECS/Fargate deployment. CloudWatch EMF emits metrics as structured JSON log lines (auto-extracted by awslogs driver -- no sidecar). OTEL traces optional and fail-open (noop when endpoint empty). CostTracker accumulates token counts as first-class operational signal.

## Conceptual Model

**4 components:** Structured logging (slog JSON to stderr), CloudWatch EMF (MetricsRecorder interface, 34 metrics from Thermia consultation), OTEL tracing (OTLP HTTP exporter, OTELMiddleware, InstrumentedPipeline with GenAI semantic conventions), Cost tracking (process-lifetime accumulation).

## Implementation Map

`internal/observe/logging.go`, `metrics.go` (EMFRecorder + NopRecorder), `tracer.go` (InitTracer), `middleware.go` (OTELMiddleware), `pipeline.go` (InstrumentedPipeline), `cost.go` (CostTracker). Terraform: `/ecs/clew` log group (30d retention), containerInsights enabled.

## Boundaries and Failure Modes

OTEL disabled: noop provider (zero impact). OTEL init failure: warning, continue. Missing pipeline: InstrumentedPipeline not created. No CloudWatch alarms exist in Terraform. InstrumentedPipeline wraps Query only (not QueryStream). CostTracker not exposed externally.

## Knowledge Gaps

1. No CloudWatch metric alarms defined
2. InstrumentedPipeline streaming coverage gap
3. containerInsights interaction with EMF not documented
