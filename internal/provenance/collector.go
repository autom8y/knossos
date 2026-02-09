package provenance

import "sync"

// Collector accumulates provenance entries during a materialization run.
// Pipeline stages call Record() after each successful file write.
// At the end of materialization, the orchestrating function calls Entries()
// to retrieve all recorded entries for manifest construction.
type Collector interface {
	// Record adds or updates a provenance entry for the given relative path.
	// relativePath is relative to .claude/ (e.g., "agents/orchestrator.md").
	// Duplicate paths overwrite previous entries (last-writer-wins, matching
	// the materialize pipeline's priority semantics).
	Record(relativePath string, entry *ProvenanceEntry)

	// Entries returns all recorded entries. The returned map must not be
	// modified by the caller.
	Entries() map[string]*ProvenanceEntry
}

// defaultCollector is the in-memory implementation of Collector.
// Thread-safe via mutex since pipeline stages might eventually parallelize.
type defaultCollector struct {
	mu      sync.Mutex
	entries map[string]*ProvenanceEntry
}

// NewCollector creates a new in-memory Collector.
func NewCollector() Collector {
	return &defaultCollector{
		entries: make(map[string]*ProvenanceEntry),
	}
}

// Record adds or overwrites a provenance entry for the given path.
func (c *defaultCollector) Record(relativePath string, entry *ProvenanceEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[relativePath] = entry
}

// Entries returns the accumulated entries map.
func (c *defaultCollector) Entries() map[string]*ProvenanceEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.entries
}

// NullCollector is a no-op Collector for dry-run and minimal modes
// where provenance tracking is not needed.
type NullCollector struct{}

// Record is a no-op.
func (NullCollector) Record(string, *ProvenanceEntry) {}

// Entries returns nil.
func (NullCollector) Entries() map[string]*ProvenanceEntry {
	return nil
}
