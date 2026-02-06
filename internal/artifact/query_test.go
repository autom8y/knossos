package artifact

import (
	"testing"
	"time"
)

func setupTestData(t *testing.T, registry *Registry, aggregator *Aggregator) {
	// Session 1: PRD and TDD for feature-a
	session1 := "session-20260105-143022-abc12345"
	entries1 := []Entry{
		{
			ArtifactID:   "PRD-feature-a",
			ArtifactType: TypePRD,
			Path:         "docs/requirements/PRD-feature-a.md",
			Phase:        PhaseRequirements,
			Specialist:   "product-owner",
			SessionID:    session1,
			RegisteredAt: time.Now().UTC(),
			Validated:    true,
		},
		{
			ArtifactID:   "TDD-feature-a",
			ArtifactType: TypeTDD,
			Path:         "docs/design/TDD-feature-a.md",
			Phase:        PhaseDesign,
			Specialist:   "context-architect",
			SessionID:    session1,
			RegisteredAt: time.Now().UTC(),
			Validated:    true,
		},
	}
	for _, entry := range entries1 {
		if err := registry.Register(session1, entry); err != nil {
			t.Fatalf("Failed to register entry: %v", err)
		}
	}
	if err := aggregator.AggregateSession(session1); err != nil {
		t.Fatalf("Failed to aggregate session1: %v", err)
	}

	// Session 2: ADR and code for feature-b
	session2 := "session-20260105-153022-def67890"
	entries2 := []Entry{
		{
			ArtifactID:   "ADR-0001-choice",
			ArtifactType: TypeADR,
			Path:         "docs/decisions/ADR-0001-choice.md",
			Phase:        PhaseDesign,
			Specialist:   "context-architect",
			SessionID:    session2,
			RegisteredAt: time.Now().UTC(),
			Validated:    true,
		},
		{
			ArtifactID:   "code-implementation",
			ArtifactType: TypeCode,
			Path:         "internal/feature/feature.go",
			Phase:        PhaseImplementation,
			Specialist:   "integration-engineer",
			SessionID:    session2,
			RegisteredAt: time.Now().UTC(),
			Validated:    true,
		},
	}
	for _, entry := range entries2 {
		if err := registry.Register(session2, entry); err != nil {
			t.Fatalf("Failed to register entry: %v", err)
		}
	}
	if err := aggregator.AggregateSession(session2); err != nil {
		t.Fatalf("Failed to aggregate session2: %v", err)
	}
}

func TestQuerier_Query_NoFilter(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	result, err := querier.Query(QueryFilter{})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 4 {
		t.Errorf("Expected 4 artifacts, got %d", result.Count)
	}
	if len(result.Entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(result.Entries))
	}
}

func TestQuerier_QueryByPhase(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	// Query requirements phase
	result, err := querier.Query(QueryFilter{Phase: PhaseRequirements})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 requirements artifact, got %d", result.Count)
	}
	if result.Entries[0].ArtifactID != "PRD-feature-a" {
		t.Errorf("Expected PRD-feature-a, got %s", result.Entries[0].ArtifactID)
	}

	// Query design phase
	result, err = querier.Query(QueryFilter{Phase: PhaseDesign})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 design artifacts, got %d", result.Count)
	}

	// Query implementation phase
	result, err = querier.Query(QueryFilter{Phase: PhaseImplementation})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 implementation artifact, got %d", result.Count)
	}
}

func TestQuerier_QueryByType(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	// Query PRD type
	result, err := querier.Query(QueryFilter{Type: TypePRD})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 PRD, got %d", result.Count)
	}

	// Query TDD type
	result, err = querier.Query(QueryFilter{Type: TypeTDD})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 TDD, got %d", result.Count)
	}

	// Query ADR type
	result, err = querier.Query(QueryFilter{Type: TypeADR})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 ADR, got %d", result.Count)
	}

	// Query code type
	result, err = querier.Query(QueryFilter{Type: TypeCode})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 code artifact, got %d", result.Count)
	}
}

func TestQuerier_QueryBySpecialist(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	// Query product-owner
	result, err := querier.Query(QueryFilter{Specialist: "product-owner"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 product-owner artifact, got %d", result.Count)
	}

	// Query context-architect
	result, err = querier.Query(QueryFilter{Specialist: "context-architect"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 context-architect artifacts, got %d", result.Count)
	}

	// Query integration-engineer
	result, err = querier.Query(QueryFilter{Specialist: "integration-engineer"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 integration-engineer artifact, got %d", result.Count)
	}
}

func TestQuerier_QueryBySession(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	// Query session 1
	result, err := querier.Query(QueryFilter{SessionID: "session-20260105-143022-abc12345"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 artifacts from session 1, got %d", result.Count)
	}

	// Query session 2
	result, err = querier.Query(QueryFilter{SessionID: "session-20260105-153022-def67890"})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 artifacts from session 2, got %d", result.Count)
	}
}

func TestQuerier_Query_MultipleFilters(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	// Query: phase=design AND specialist=context-architect
	result, err := querier.Query(QueryFilter{
		Phase:      PhaseDesign,
		Specialist: "context-architect",
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 design artifacts from context-architect, got %d", result.Count)
	}

	// Query: phase=design AND type=tdd
	result, err = querier.Query(QueryFilter{
		Phase: PhaseDesign,
		Type:  TypeTDD,
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 design TDD, got %d", result.Count)
	}
	if result.Entries[0].ArtifactID != "TDD-feature-a" {
		t.Errorf("Expected TDD-feature-a, got %s", result.Entries[0].ArtifactID)
	}

	// Query: session AND type (should return 0 - no match)
	result, err = querier.Query(QueryFilter{
		SessionID: "session-20260105-143022-abc12345",
		Type:      TypeADR,
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("Expected 0 artifacts (no ADR in session 1), got %d", result.Count)
	}
}

func TestQuerier_ListPhases(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	counts, err := querier.ListPhases()
	if err != nil {
		t.Fatalf("ListPhases failed: %v", err)
	}

	if counts[PhaseRequirements] != 1 {
		t.Errorf("Expected 1 requirements artifact, got %d", counts[PhaseRequirements])
	}
	if counts[PhaseDesign] != 2 {
		t.Errorf("Expected 2 design artifacts, got %d", counts[PhaseDesign])
	}
	if counts[PhaseImplementation] != 1 {
		t.Errorf("Expected 1 implementation artifact, got %d", counts[PhaseImplementation])
	}
	if counts[PhaseValidation] != 0 {
		t.Errorf("Expected 0 validation artifacts, got %d", counts[PhaseValidation])
	}
}

func TestQuerier_ListTypes(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	counts, err := querier.ListTypes()
	if err != nil {
		t.Fatalf("ListTypes failed: %v", err)
	}

	if counts[TypePRD] != 1 {
		t.Errorf("Expected 1 PRD, got %d", counts[TypePRD])
	}
	if counts[TypeTDD] != 1 {
		t.Errorf("Expected 1 TDD, got %d", counts[TypeTDD])
	}
	if counts[TypeADR] != 1 {
		t.Errorf("Expected 1 ADR, got %d", counts[TypeADR])
	}
	if counts[TypeCode] != 1 {
		t.Errorf("Expected 1 code artifact, got %d", counts[TypeCode])
	}
}

func TestQuerier_ListSpecialists(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	counts, err := querier.ListSpecialists()
	if err != nil {
		t.Fatalf("ListSpecialists failed: %v", err)
	}

	if counts["product-owner"] != 1 {
		t.Errorf("Expected 1 product-owner artifact, got %d", counts["product-owner"])
	}
	if counts["context-architect"] != 2 {
		t.Errorf("Expected 2 context-architect artifacts, got %d", counts["context-architect"])
	}
	if counts["integration-engineer"] != 1 {
		t.Errorf("Expected 1 integration-engineer artifact, got %d", counts["integration-engineer"])
	}
}

func TestQuerier_ListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)
	querier := NewQuerier(registry)

	setupTestData(t, registry, aggregator)

	counts, err := querier.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if counts["session-20260105-143022-abc12345"] != 2 {
		t.Errorf("Expected 2 artifacts in session 1, got %d", counts["session-20260105-143022-abc12345"])
	}
	if counts["session-20260105-153022-def67890"] != 2 {
		t.Errorf("Expected 2 artifacts in session 2, got %d", counts["session-20260105-153022-def67890"])
	}
}

func TestQuerier_EmptyRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	querier := NewQuerier(registry)

	// Query empty registry
	result, err := querier.Query(QueryFilter{})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("Expected 0 artifacts, got %d", result.Count)
	}

	// List phases in empty registry
	counts, err := querier.ListPhases()
	if err != nil {
		t.Fatalf("ListPhases failed: %v", err)
	}

	if len(counts) != 0 {
		t.Errorf("Expected 0 phase counts, got %d", len(counts))
	}
}
