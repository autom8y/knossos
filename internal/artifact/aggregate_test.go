package artifact

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAggregator_AggregateSession_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	sessionID := "session-20260105-143022-abc12345"

	// Aggregate empty session (should succeed with no changes)
	if err := aggregator.AggregateSession(sessionID); err != nil {
		t.Fatalf("Failed to aggregate empty session: %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 0 {
		t.Errorf("Expected 0 artifacts, got %d", len(projectReg.Artifacts))
	}
}

func TestAggregator_AggregateSession_SingleArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	sessionID := "session-20260105-143022-abc12345"
	entry := Entry{
		ArtifactID:   "TDD-test",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-test.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}

	if err := registry.Register(sessionID, entry); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}

	if err := aggregator.AggregateSession(sessionID); err != nil {
		t.Fatalf("Failed to aggregate session: %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(projectReg.Artifacts))
	}
	if projectReg.Artifacts[0].ArtifactID != "TDD-test" {
		t.Errorf("Expected TDD-test, got %s", projectReg.Artifacts[0].ArtifactID)
	}
	if projectReg.SessionsIndexed != 1 {
		t.Errorf("Expected 1 session indexed, got %d", projectReg.SessionsIndexed)
	}
}

func TestAggregator_AggregateSession_MultipleSessions(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	// Register artifacts from session 1
	session1 := "session-20260105-143022-abc12345"
	entry1 := Entry{
		ArtifactID:   "PRD-feature-a",
		ArtifactType: TypePRD,
		Path:         "docs/requirements/PRD-feature-a.md",
		Phase:        PhaseRequirements,
		Specialist:   "product-owner",
		SessionID:    session1,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}
	if err := registry.Register(session1, entry1); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}
	if err := aggregator.AggregateSession(session1); err != nil {
		t.Fatalf("Failed to aggregate session 1: %v", err)
	}

	// Register artifacts from session 2
	session2 := "session-20260105-153022-def67890"
	entry2 := Entry{
		ArtifactID:   "TDD-feature-a",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-feature-a.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    session2,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}
	if err := registry.Register(session2, entry2); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}
	if err := aggregator.AggregateSession(session2); err != nil {
		t.Fatalf("Failed to aggregate session 2: %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 2 {
		t.Fatalf("Expected 2 artifacts, got %d", len(projectReg.Artifacts))
	}
	if projectReg.SessionsIndexed != 2 {
		t.Errorf("Expected 2 sessions indexed, got %d", projectReg.SessionsIndexed)
	}
}

func TestAggregator_AggregateSession_UpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	sessionID := "session-20260105-143022-abc12345"
	entry1 := Entry{
		ArtifactID:   "TDD-update",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-update-v1.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    false,
	}
	if err := registry.Register(sessionID, entry1); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}
	if err := aggregator.AggregateSession(sessionID); err != nil {
		t.Fatalf("Failed to aggregate session: %v", err)
	}

	// Update the same artifact
	entry2 := Entry{
		ArtifactID:   "TDD-update",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-update-v2.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}
	if err := registry.Register(sessionID, entry2); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}
	if err := aggregator.AggregateSession(sessionID); err != nil {
		t.Fatalf("Failed to aggregate session: %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact (not duplicated), got %d", len(projectReg.Artifacts))
	}
	if projectReg.Artifacts[0].Path != ".ledge/specs/TDD-update-v2.md" {
		t.Errorf("Expected graduated path .ledge/specs/TDD-update-v2.md, got %s", projectReg.Artifacts[0].Path)
	}
	if projectReg.Artifacts[0].Validated != true {
		t.Errorf("Expected validated=true, got %v", projectReg.Artifacts[0].Validated)
	}
}

func TestAggregator_BuildIndexes(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	artifacts := []Entry{
		{
			ArtifactID:   "PRD-feature-a",
			ArtifactType: TypePRD,
			Phase:        PhaseRequirements,
			Specialist:   "product-owner",
			SessionID:    "session-1",
		},
		{
			ArtifactID:   "TDD-feature-a",
			ArtifactType: TypeTDD,
			Phase:        PhaseDesign,
			Specialist:   "context-architect",
			SessionID:    "session-1",
		},
		{
			ArtifactID:   "PRD-feature-b",
			ArtifactType: TypePRD,
			Phase:        PhaseRequirements,
			Specialist:   "product-owner",
			SessionID:    "session-2",
		},
	}

	indexes := aggregator.buildIndexes(artifacts)

	// Check phase index
	if len(indexes.ByPhase[PhaseRequirements]) != 2 {
		t.Errorf("Expected 2 requirements artifacts, got %d", len(indexes.ByPhase[PhaseRequirements]))
	}
	if len(indexes.ByPhase[PhaseDesign]) != 1 {
		t.Errorf("Expected 1 design artifact, got %d", len(indexes.ByPhase[PhaseDesign]))
	}

	// Check type index
	if len(indexes.ByType[TypePRD]) != 2 {
		t.Errorf("Expected 2 PRD artifacts, got %d", len(indexes.ByType[TypePRD]))
	}
	if len(indexes.ByType[TypeTDD]) != 1 {
		t.Errorf("Expected 1 TDD artifact, got %d", len(indexes.ByType[TypeTDD]))
	}

	// Check specialist index
	if len(indexes.BySpecialist["product-owner"]) != 2 {
		t.Errorf("Expected 2 product-owner artifacts, got %d", len(indexes.BySpecialist["product-owner"]))
	}
	if len(indexes.BySpecialist["context-architect"]) != 1 {
		t.Errorf("Expected 1 context-architect artifact, got %d", len(indexes.BySpecialist["context-architect"]))
	}

	// Check session index
	if len(indexes.BySession["session-1"]) != 2 {
		t.Errorf("Expected 2 session-1 artifacts, got %d", len(indexes.BySession["session-1"]))
	}
	if len(indexes.BySession["session-2"]) != 1 {
		t.Errorf("Expected 1 session-2 artifact, got %d", len(indexes.BySession["session-2"]))
	}
}

func TestAggregator_AggregateAll_NoSessions(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	// AggregateAll with no sessions directory
	if err := aggregator.AggregateAll(); err != nil {
		t.Fatalf("Expected no error for missing sessions dir, got %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 0 {
		t.Errorf("Expected 0 artifacts, got %d", len(projectReg.Artifacts))
	}
}

func TestAggregator_AggregateAll_MultipleSessions(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	// Create sessions directory
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions directory: %v", err)
	}

	// Register artifacts in session 1
	session1 := "session-20260105-143022-abc12345"
	entry1 := Entry{
		ArtifactID:   "PRD-test-1",
		ArtifactType: TypePRD,
		Path:         "docs/requirements/PRD-test-1.md",
		Phase:        PhaseRequirements,
		Specialist:   "product-owner",
		SessionID:    session1,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}
	if err := registry.Register(session1, entry1); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}

	// Register artifacts in session 2
	session2 := "session-20260105-153022-def67890"
	entry2 := Entry{
		ArtifactID:   "TDD-test-2",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-test-2.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    session2,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}
	if err := registry.Register(session2, entry2); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}

	// Aggregate all sessions
	if err := aggregator.AggregateAll(); err != nil {
		t.Fatalf("Failed to aggregate all: %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 2 {
		t.Fatalf("Expected 2 artifacts, got %d", len(projectReg.Artifacts))
	}
	if projectReg.SessionsIndexed != 2 {
		t.Errorf("Expected 2 sessions indexed, got %d", projectReg.SessionsIndexed)
	}

	// Verify indexes
	if len(projectReg.Indexes.ByPhase[PhaseRequirements]) != 1 {
		t.Errorf("Expected 1 requirements artifact in index")
	}
	if len(projectReg.Indexes.ByPhase[PhaseDesign]) != 1 {
		t.Errorf("Expected 1 design artifact in index")
	}
}

func TestAggregator_AggregateAll_IgnoresNonSessionDirs(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	// Create sessions directory with mixed content
	sessionsDir := filepath.Join(tmpDir, ".sos", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("Failed to create sessions directory: %v", err)
	}

	// Create a non-session directory
	if err := os.MkdirAll(filepath.Join(sessionsDir, "not-a-session"), 0755); err != nil {
		t.Fatalf("Failed to create non-session directory: %v", err)
	}

	// Create a file (not a directory)
	if err := os.WriteFile(filepath.Join(sessionsDir, ".current-session"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Register artifact in valid session
	sessionID := "session-20260105-143022-abc12345"
	entry := Entry{
		ArtifactID:   "PRD-valid",
		ArtifactType: TypePRD,
		Path:         "docs/requirements/PRD-valid.md",
		Phase:        PhaseRequirements,
		Specialist:   "product-owner",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}
	if err := registry.Register(sessionID, entry); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}

	// Aggregate all - should only pick up valid session
	if err := aggregator.AggregateAll(); err != nil {
		t.Fatalf("Failed to aggregate all: %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact (ignoring non-session dirs), got %d", len(projectReg.Artifacts))
	}
}

func TestAggregator_AggregateSession_GraduatesPaths(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)
	aggregator := NewAggregator(registry)

	sessionID := "session-20260302-100000-grad1234"

	// Register artifacts of different types
	entries := []Entry{
		{
			ArtifactID: "ADR-auth", ArtifactType: TypeADR,
			Path: ".sos/sessions/" + sessionID + "/ADR-auth.md",
			Phase: PhaseDesign, Specialist: "architect", SessionID: sessionID,
			RegisteredAt: time.Now().UTC(), Validated: true,
		},
		{
			ArtifactID: "PRD-billing", ArtifactType: TypePRD,
			Path: "docs/requirements/PRD-billing.md",
			Phase: PhaseRequirements, Specialist: "product-owner", SessionID: sessionID,
			RegisteredAt: time.Now().UTC(), Validated: true,
		},
		{
			ArtifactID: "review-auth", ArtifactType: TypeReview,
			Path: ".sos/sessions/" + sessionID + "/review-auth.md",
			Phase: PhaseValidation, Specialist: "audit-lead", SessionID: sessionID,
			RegisteredAt: time.Now().UTC(), Validated: true,
		},
		{
			ArtifactID: "code-handler", ArtifactType: TypeCode,
			Path: "internal/auth/handler.go",
			Phase: PhaseImplementation, Specialist: "developer", SessionID: sessionID,
			RegisteredAt: time.Now().UTC(), Validated: true,
		},
		{
			ArtifactID: "spike-caching", ArtifactType: TypeSpike,
			Path: ".sos/sessions/" + sessionID + "/spike-caching.md",
			Phase: PhaseDesign, Specialist: "architect", SessionID: sessionID,
			RegisteredAt: time.Now().UTC(), Validated: true,
		},
	}

	for _, entry := range entries {
		if err := registry.Register(sessionID, entry); err != nil {
			t.Fatalf("Failed to register %s: %v", entry.ArtifactID, err)
		}
	}

	// Verify session registry has original paths (not graduated)
	sessionReg, err := registry.LoadSessionRegistry(sessionID)
	if err != nil {
		t.Fatalf("Failed to load session registry: %v", err)
	}
	for _, a := range sessionReg.Artifacts {
		if a.ArtifactType != TypeCode && a.Path != entries[0].Path && a.ArtifactType == TypeADR {
			// Session registry should have original path
			if a.Path != ".sos/sessions/"+sessionID+"/ADR-auth.md" {
				t.Errorf("Session registry should have original path, got %s", a.Path)
			}
		}
	}

	// Aggregate (graduation point)
	if err := aggregator.AggregateSession(sessionID); err != nil {
		t.Fatalf("Failed to aggregate session: %v", err)
	}

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if len(projectReg.Artifacts) != 5 {
		t.Fatalf("Expected 5 artifacts, got %d", len(projectReg.Artifacts))
	}

	// Verify graduated paths
	pathMap := make(map[string]string) // artifactID -> path
	for _, a := range projectReg.Artifacts {
		pathMap[a.ArtifactID] = a.Path
	}

	expectations := map[string]string{
		"ADR-auth":      ".ledge/decisions/ADR-auth.md",
		"PRD-billing":   ".ledge/specs/PRD-billing.md",
		"review-auth":   ".ledge/reviews/review-auth.md",
		"code-handler":  "internal/auth/handler.go", // unchanged
		"spike-caching": ".ledge/spikes/spike-caching.md",
	}

	for id, expected := range expectations {
		if got := pathMap[id]; got != expected {
			t.Errorf("Artifact %s: path = %q, want %q", id, got, expected)
		}
	}
}
