package observe

import (
	"math"
	"sync"
	"testing"

	"github.com/autom8y/knossos/internal/reason/response"
)

func TestCostTracker_AccumulatesCorrectly(t *testing.T) {
	ct := NewCostTracker()

	ct.Record(response.TokenReport{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		EstimatedCostUSD: 0.003,
	})

	ct.Record(response.TokenReport{
		PromptTokens:     200,
		CompletionTokens: 100,
		TotalTokens:      300,
		EstimatedCostUSD: 0.006,
	})

	report := ct.Report()

	if report.QueryCount != 2 {
		t.Errorf("QueryCount = %d, want 2", report.QueryCount)
	}
	if report.TotalPromptTokens != 300 {
		t.Errorf("TotalPromptTokens = %d, want 300", report.TotalPromptTokens)
	}
	if report.TotalCompletionTokens != 150 {
		t.Errorf("TotalCompletionTokens = %d, want 150", report.TotalCompletionTokens)
	}
	if report.TotalTokens != 450 {
		t.Errorf("TotalTokens = %d, want 450", report.TotalTokens)
	}
	if math.Abs(report.TotalCostUSD-0.009) > 1e-9 {
		t.Errorf("TotalCostUSD = %f, want 0.009", report.TotalCostUSD)
	}
}

func TestCostTracker_ZeroInitial(t *testing.T) {
	ct := NewCostTracker()
	report := ct.Report()

	if report.QueryCount != 0 {
		t.Errorf("initial QueryCount = %d, want 0", report.QueryCount)
	}
	if report.TotalTokens != 0 {
		t.Errorf("initial TotalTokens = %d, want 0", report.TotalTokens)
	}
	if report.TotalCostUSD != 0 {
		t.Errorf("initial TotalCostUSD = %f, want 0", report.TotalCostUSD)
	}
}

func TestCostTracker_ThreadSafe(t *testing.T) {
	ct := NewCostTracker()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ct.Record(response.TokenReport{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
				EstimatedCostUSD: 0.001,
			})
		}()
	}

	wg.Wait()

	report := ct.Report()
	if report.QueryCount != goroutines {
		t.Errorf("QueryCount = %d, want %d", report.QueryCount, goroutines)
	}
	if report.TotalPromptTokens != goroutines*10 {
		t.Errorf("TotalPromptTokens = %d, want %d", report.TotalPromptTokens, goroutines*10)
	}
	if report.TotalCompletionTokens != goroutines*5 {
		t.Errorf("TotalCompletionTokens = %d, want %d", report.TotalCompletionTokens, goroutines*5)
	}
}
