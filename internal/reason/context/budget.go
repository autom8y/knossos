package context

// BudgetManager tracks and enforces token budgets during context assembly.
type BudgetManager struct {
	limit     int
	consumed  int
	included  int
	skipped   int
}

// NewBudgetManager creates a BudgetManager with the given token limit.
func NewBudgetManager(limit int) *BudgetManager {
	return &BudgetManager{
		limit: limit,
	}
}

// CanFit returns true if the given token count fits within the remaining budget.
func (bm *BudgetManager) CanFit(tokens int) bool {
	return bm.consumed+tokens <= bm.limit
}

// Consume records token usage. Returns false if the tokens would exceed the budget.
func (bm *BudgetManager) Consume(tokens int) bool {
	if !bm.CanFit(tokens) {
		bm.skipped++
		return false
	}
	bm.consumed += tokens
	bm.included++
	return true
}

// Consumed returns the number of tokens consumed so far.
func (bm *BudgetManager) Consumed() int {
	return bm.consumed
}

// Skip records a skipped item without consuming tokens.
// Use when manually deciding to skip a candidate outside of Consume().
func (bm *BudgetManager) Skip() {
	bm.skipped++
}

// Remaining returns the number of tokens remaining in the budget.
func (bm *BudgetManager) Remaining() int {
	r := bm.limit - bm.consumed
	if r < 0 {
		return 0
	}
	return r
}

// Report returns a BudgetReport summarizing token allocation.
func (bm *BudgetManager) Report() BudgetReport {
	utilization := 0.0
	if bm.limit > 0 {
		utilization = float64(bm.consumed) / float64(bm.limit)
		if utilization > 1.0 {
			utilization = 1.0
		}
	}
	return BudgetReport{
		SourceMaterialTokens: bm.consumed,
		BudgetLimit:          bm.limit,
		SourcesIncluded:      bm.included,
		SourcesSkipped:       bm.skipped,
		BudgetUtilization:    utilization,
	}
}
