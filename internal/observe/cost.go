package observe

import (
	"sync"

	"github.com/autom8y/knossos/internal/reason/response"
)

// CostReport is a point-in-time snapshot of accumulated token usage and cost.
type CostReport struct {
	TotalPromptTokens     int64
	TotalCompletionTokens int64
	TotalTokens           int64
	TotalCostUSD          float64
	QueryCount            int64
}

// CostTracker accumulates token usage and estimated cost across pipeline invocations.
// Safe for concurrent use.
type CostTracker struct {
	mu                    sync.Mutex
	totalPromptTokens     int64
	totalCompletionTokens int64
	totalTokens           int64
	totalCostUSD          float64
	queryCount            int64
}

// NewCostTracker creates a new zero-valued CostTracker.
func NewCostTracker() *CostTracker {
	return &CostTracker{}
}

// Record adds a TokenReport to the running totals.
func (c *CostTracker) Record(report response.TokenReport) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalPromptTokens += int64(report.PromptTokens)
	c.totalCompletionTokens += int64(report.CompletionTokens)
	c.totalTokens += int64(report.TotalTokens)
	c.totalCostUSD += report.EstimatedCostUSD
	c.queryCount++
}

// Report returns a snapshot of the accumulated cost data.
func (c *CostTracker) Report() CostReport {
	c.mu.Lock()
	defer c.mu.Unlock()
	return CostReport{
		TotalPromptTokens:     c.totalPromptTokens,
		TotalCompletionTokens: c.totalCompletionTokens,
		TotalTokens:           c.totalTokens,
		TotalCostUSD:          c.totalCostUSD,
		QueryCount:            c.queryCount,
	}
}
