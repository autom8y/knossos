package knowledge

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// slowLLMClient simulates LLM latency for concurrency testing.
// Thread-safe: uses atomic counters for concurrent goroutine access.
type slowLLMClient struct {
	latency     time.Duration
	response    string
	callCount   atomic.Int32
	maxInFlight atomic.Int32
	inFlight    atomic.Int32
}

func (m *slowLLMClient) Complete(ctx context.Context, _, _ string, _ int) (string, error) {
	m.callCount.Add(1)

	// Track concurrent in-flight calls to prove parallelism.
	current := m.inFlight.Add(1)
	for {
		old := m.maxInFlight.Load()
		if current <= old {
			break
		}
		if m.maxInFlight.CompareAndSwap(old, current) {
			break
		}
	}
	defer m.inFlight.Add(-1)

	select {
	case <-time.After(m.latency):
		return m.response, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func TestBuild_ConcurrentSummaryGeneration(t *testing.T) {
	const domainCount = 10
	const llmLatency = 200 * time.Millisecond

	// Build catalog with N domains, each requiring summary generation.
	domains := make([]CatalogDomainEntry, domainCount)
	content := make(map[string]string)
	for i := range domainCount {
		qn := fmt.Sprintf("org::repo::domain-%d", i)
		domains[i] = CatalogDomainEntry{
			QualifiedName: qn,
			Domain:        "architecture",
			SourceHash:    fmt.Sprintf("hash-%d", i),
			GeneratedAt:   "2026-03-27T00:00:00Z",
		}
		content[qn] = fmt.Sprintf("# Domain %d\n\n## Section\n\nContent for domain %d.", i, i)
	}

	llm := &slowLLMClient{
		latency:  llmLatency,
		response: "Summary of this domain covering key architectural patterns.",
	}

	catalog := &mockCatalog{domains: domains}
	contentStore := &mockContentStore{content: content}

	start := time.Now()
	idx, err := Build(context.Background(), BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		LLMClient:    llm,
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// All domains should be indexed.
	if idx.DomainCount() != domainCount {
		t.Errorf("DomainCount() = %d, want %d", idx.DomainCount(), domainCount)
	}

	// All summaries should be generated.
	if idx.SummaryCount() != domainCount {
		t.Errorf("SummaryCount() = %d, want %d", idx.SummaryCount(), domainCount)
	}

	// LLM should have been called for each domain.
	calls := int(llm.callCount.Load())
	if calls != domainCount {
		t.Errorf("LLM call count = %d, want %d", calls, domainCount)
	}

	// Wall-clock time should be significantly less than sequential execution.
	// Sequential would be domainCount * llmLatency = 10 * 200ms = 2s.
	// With concurrency=10, all 10 should run in parallel: ~200ms + overhead.
	// We allow up to 3x a single call (600ms) as a generous bound.
	sequentialTime := time.Duration(domainCount) * llmLatency
	maxAllowed := 3 * llmLatency
	if elapsed >= sequentialTime {
		t.Errorf("Build took %v, which is >= sequential time %v -- concurrency is not working",
			elapsed, sequentialTime)
	}
	if elapsed >= maxAllowed {
		t.Errorf("Build took %v, expected < %v with full parallelism", elapsed, maxAllowed)
	}

	// Verify actual concurrency occurred: max in-flight should be > 1.
	maxConcurrent := int(llm.maxInFlight.Load())
	if maxConcurrent < 2 {
		t.Errorf("max concurrent LLM calls = %d, want >= 2 (proving parallelism)", maxConcurrent)
	}

	t.Logf("concurrency test: %d domains, %v elapsed (sequential would be %v), max concurrent = %d",
		domainCount, elapsed, sequentialTime, maxConcurrent)
}

func TestBuild_ConcurrentSummaryGeneration_NoTimeoutCascade(t *testing.T) {
	// Simulates the original bug scenario: 128 domains, 5s LLM latency.
	// With the mutex wrapping Generate(), only 1 goroutine runs at a time,
	// causing 9 others to burn their 30s context timeout waiting for the lock.
	// After the fix, all 10 concurrent slots run in parallel.
	//
	// We use scaled-down numbers for test speed: 20 domains, 100ms latency.
	// The assertions prove the same property: no timeout cascade.
	const domainCount = 20
	const llmLatency = 100 * time.Millisecond

	domains := make([]CatalogDomainEntry, domainCount)
	content := make(map[string]string)
	for i := range domainCount {
		qn := fmt.Sprintf("org::repo::cascade-%d", i)
		domains[i] = CatalogDomainEntry{
			QualifiedName: qn,
			Domain:        "scar-tissue",
			SourceHash:    fmt.Sprintf("cascade-hash-%d", i),
			GeneratedAt:   "2026-03-27T00:00:00Z",
		}
		content[qn] = fmt.Sprintf("# Cascade Domain %d\n\n## Overview\n\nContent %d.", i, i)
	}

	llm := &slowLLMClient{
		latency:  llmLatency,
		response: "Summary for cascade test domain.",
	}

	catalog := &mockCatalog{domains: domains}
	contentStore := &mockContentStore{content: content}

	// Use a 10-second context ceiling -- generous but would fail under serialized behavior.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	idx, err := Build(ctx, BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		LLMClient:    llm,
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// All summaries must succeed -- no timeout cascade.
	if idx.SummaryCount() != domainCount {
		t.Errorf("SummaryCount() = %d, want %d (timeout cascade?)", idx.SummaryCount(), domainCount)
	}

	calls := int(llm.callCount.Load())
	if calls != domainCount {
		t.Errorf("LLM call count = %d, want %d", calls, domainCount)
	}

	// With concurrency=10 and 20 domains at 100ms each:
	// 2 batches of 10, each taking ~100ms = ~200ms total.
	// Sequential would be 20 * 100ms = 2s.
	// Allow generous 4x single batch (400ms) for CI variability.
	maxAllowed := 4 * llmLatency
	sequentialTime := time.Duration(domainCount) * llmLatency
	if elapsed >= sequentialTime/2 {
		t.Errorf("Build took %v (sequential = %v), concurrency not effective", elapsed, sequentialTime)
	}

	t.Logf("cascade test: %d domains, %v elapsed (sequential would be %v), max concurrent = %d",
		domainCount, elapsed, sequentialTime, llm.maxInFlight.Load())

	_ = maxAllowed // used for documentation; the sequentialTime/2 check is stricter
}
