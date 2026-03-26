package observe

import (
	"context"
	"fmt"
	"testing"

	"github.com/autom8y/knossos/internal/reason/response"
	"github.com/autom8y/knossos/internal/trust"
)

// mockQueryRunner is a test double for slack.QueryRunner.
type mockQueryRunner struct {
	resp *response.ReasoningResponse
	err  error
}

func (m *mockQueryRunner) Query(_ context.Context, _ string) (*response.ReasoningResponse, error) {
	return m.resp, m.err
}

func TestInstrumentedPipeline_DelegatesToInner(t *testing.T) {
	// Install noop tracer so spans are created but not exported.
	_, _ = InitTracer("test", "")

	expected := &response.ReasoningResponse{
		Answer: "test answer",
		Tier:   trust.TierHigh,
		Confidence: trust.ConfidenceScore{
			Overall: 0.85,
		},
		TokensUsed: response.TokenReport{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
			EstimatedCostUSD: 0.003,
		},
	}

	inner := &mockQueryRunner{resp: expected}
	cost := NewCostTracker()
	pipeline := NewInstrumentedPipeline(inner, cost, nil)

	resp, err := pipeline.Query(context.Background(), "test question")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if resp != expected {
		t.Error("Query() did not return inner response")
	}
}

func TestInstrumentedPipeline_RecordsCost(t *testing.T) {
	_, _ = InitTracer("test", "")

	tokens := response.TokenReport{
		PromptTokens:     200,
		CompletionTokens: 100,
		TotalTokens:      300,
		EstimatedCostUSD: 0.006,
	}

	inner := &mockQueryRunner{
		resp: &response.ReasoningResponse{
			Answer:     "answer",
			Tier:       trust.TierMedium,
			TokensUsed: tokens,
		},
	}
	cost := NewCostTracker()
	pipeline := NewInstrumentedPipeline(inner, cost, nil)

	_, err := pipeline.Query(context.Background(), "question")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	report := cost.Report()
	if report.QueryCount != 1 {
		t.Errorf("QueryCount = %d, want 1", report.QueryCount)
	}
	if report.TotalPromptTokens != 200 {
		t.Errorf("TotalPromptTokens = %d, want 200", report.TotalPromptTokens)
	}
	if report.TotalCompletionTokens != 100 {
		t.Errorf("TotalCompletionTokens = %d, want 100", report.TotalCompletionTokens)
	}
}

func TestInstrumentedPipeline_PropagatesError(t *testing.T) {
	_, _ = InitTracer("test", "")

	expectedErr := fmt.Errorf("pipeline failure")
	inner := &mockQueryRunner{err: expectedErr}
	cost := NewCostTracker()
	pipeline := NewInstrumentedPipeline(inner, cost, nil)

	_, err := pipeline.Query(context.Background(), "question")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("error = %q, want %q", err.Error(), expectedErr.Error())
	}

	// No cost should be recorded on error.
	report := cost.Report()
	if report.QueryCount != 0 {
		t.Errorf("QueryCount = %d, want 0 (error path)", report.QueryCount)
	}
}

func TestInstrumentedPipeline_NilCostTracker(t *testing.T) {
	_, _ = InitTracer("test", "")

	inner := &mockQueryRunner{
		resp: &response.ReasoningResponse{
			Answer: "answer",
			Tier:   trust.TierHigh,
			TokensUsed: response.TokenReport{
				PromptTokens: 50,
			},
		},
	}

	// nil CostTracker should not panic.
	pipeline := NewInstrumentedPipeline(inner, nil, nil)
	_, err := pipeline.Query(context.Background(), "question")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
}

func TestInstrumentedPipeline_TruncatesLongQuestion(t *testing.T) {
	_, _ = InitTracer("test", "")

	inner := &mockQueryRunner{
		resp: &response.ReasoningResponse{
			Answer: "answer",
			Tier:   trust.TierHigh,
		},
	}

	cost := NewCostTracker()
	pipeline := NewInstrumentedPipeline(inner, cost, nil)

	// Question longer than 256 chars should not cause issues.
	longQuestion := make([]byte, 1000)
	for i := range longQuestion {
		longQuestion[i] = 'a'
	}

	_, err := pipeline.Query(context.Background(), string(longQuestion))
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
}
