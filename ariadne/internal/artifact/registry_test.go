package artifact

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegistry_SessionRegistryPath(t *testing.T) {
	registry := NewRegistry("/test/project")
	path := registry.SessionRegistryPath("session-20260105-143022-abc12345")
	expected := "/test/project/.claude/sessions/session-20260105-143022-abc12345/artifacts.yaml"
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestRegistry_ProjectRegistryPath(t *testing.T) {
	registry := NewRegistry("/test/project")
	path := registry.ProjectRegistryPath()
	expected := "/test/project/.claude/artifacts/registry.yaml"
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestRegistry_LoadSessionRegistry_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	sessionID := "session-20260105-143022-abc12345"
	sessionReg, err := registry.LoadSessionRegistry(sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if sessionReg.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, sessionReg.SessionID)
	}
	if sessionReg.SchemaVersion != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", sessionReg.SchemaVersion)
	}
	if len(sessionReg.Artifacts) != 0 {
		t.Errorf("Expected empty artifacts, got %d", len(sessionReg.Artifacts))
	}
}

func TestRegistry_SaveAndLoadSessionRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	sessionID := "session-20260105-143022-abc12345"
	entry := Entry{
		ArtifactID:   "TDD-test-artifact",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-test.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}

	sessionReg := &SessionRegistry{
		SchemaVersion: "1.0",
		SessionID:     sessionID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Artifacts:     []Entry{entry},
	}

	if err := registry.SaveSessionRegistry(sessionReg); err != nil {
		t.Fatalf("Failed to save session registry: %v", err)
	}

	loaded, err := registry.LoadSessionRegistry(sessionID)
	if err != nil {
		t.Fatalf("Failed to load session registry: %v", err)
	}

	if loaded.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, loaded.SessionID)
	}
	if loaded.ArtifactCount != 1 {
		t.Errorf("Expected 1 artifact, got %d", loaded.ArtifactCount)
	}
	if loaded.Artifacts[0].ArtifactID != "TDD-test-artifact" {
		t.Errorf("Expected artifact ID TDD-test-artifact, got %s", loaded.Artifacts[0].ArtifactID)
	}
}

func TestRegistry_Register_NewEntry(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	sessionID := "session-20260105-143022-abc12345"
	entry := Entry{
		ArtifactID:   "TDD-new-artifact",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-new.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}

	if err := registry.Register(sessionID, entry); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}

	loaded, err := registry.LoadSessionRegistry(sessionID)
	if err != nil {
		t.Fatalf("Failed to load session registry: %v", err)
	}

	if len(loaded.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(loaded.Artifacts))
	}
	if loaded.Artifacts[0].ArtifactID != "TDD-new-artifact" {
		t.Errorf("Expected artifact ID TDD-new-artifact, got %s", loaded.Artifacts[0].ArtifactID)
	}
}

func TestRegistry_Register_UpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	sessionID := "session-20260105-143022-abc12345"
	entry1 := Entry{
		ArtifactID:   "TDD-update-test",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-update.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    false,
	}

	if err := registry.Register(sessionID, entry1); err != nil {
		t.Fatalf("Failed to register entry: %v", err)
	}

	// Update with validated=true
	entry2 := Entry{
		ArtifactID:   "TDD-update-test",
		ArtifactType: TypeTDD,
		Path:         "docs/design/TDD-update-v2.md",
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}

	if err := registry.Register(sessionID, entry2); err != nil {
		t.Fatalf("Failed to update entry: %v", err)
	}

	loaded, err := registry.LoadSessionRegistry(sessionID)
	if err != nil {
		t.Fatalf("Failed to load session registry: %v", err)
	}

	if len(loaded.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact (not duplicated), got %d", len(loaded.Artifacts))
	}
	if loaded.Artifacts[0].Validated != true {
		t.Errorf("Expected validated=true, got %v", loaded.Artifacts[0].Validated)
	}
	if loaded.Artifacts[0].Path != "docs/design/TDD-update-v2.md" {
		t.Errorf("Expected updated path, got %s", loaded.Artifacts[0].Path)
	}
}

func TestRegistry_LoadProjectRegistry_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	projectReg, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if projectReg.ProjectRoot != tmpDir {
		t.Errorf("Expected project root %s, got %s", tmpDir, projectReg.ProjectRoot)
	}
	if projectReg.SchemaVersion != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", projectReg.SchemaVersion)
	}
	if len(projectReg.Artifacts) != 0 {
		t.Errorf("Expected empty artifacts, got %d", len(projectReg.Artifacts))
	}
	if projectReg.Indexes.ByPhase == nil {
		t.Error("Expected initialized ByPhase index")
	}
}

func TestRegistry_SaveAndLoadProjectRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	entry := Entry{
		ArtifactID:   "PRD-test-project",
		ArtifactType: TypePRD,
		Path:         "docs/requirements/PRD-test.md",
		Phase:        PhaseRequirements,
		Specialist:   "product-owner",
		SessionID:    "session-20260105-143022-abc12345",
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}

	projectReg := &ProjectRegistry{
		SchemaVersion:   "1.0",
		ProjectRoot:     tmpDir,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		SessionsIndexed: 1,
		Artifacts:       []Entry{entry},
		Indexes: ProjectIndexes{
			ByPhase:      map[Phase][]string{PhaseRequirements: {"PRD-test-project"}},
			ByType:       map[ArtifactType][]string{TypePRD: {"PRD-test-project"}},
			BySpecialist: map[string][]string{"product-owner": {"PRD-test-project"}},
			BySession:    map[string][]string{"session-20260105-143022-abc12345": {"PRD-test-project"}},
		},
	}

	if err := registry.SaveProjectRegistry(projectReg); err != nil {
		t.Fatalf("Failed to save project registry: %v", err)
	}

	loaded, err := registry.LoadProjectRegistry()
	if err != nil {
		t.Fatalf("Failed to load project registry: %v", err)
	}

	if loaded.ProjectRoot != tmpDir {
		t.Errorf("Expected project root %s, got %s", tmpDir, loaded.ProjectRoot)
	}
	if loaded.ArtifactCount != 1 {
		t.Errorf("Expected 1 artifact, got %d", loaded.ArtifactCount)
	}
	if len(loaded.Indexes.ByPhase[PhaseRequirements]) != 1 {
		t.Errorf("Expected 1 artifact in requirements phase")
	}
}

func TestRegistry_SaveSessionRegistry_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	sessionID := "session-20260105-143022-abc12345"
	sessionReg := &SessionRegistry{
		SchemaVersion: "1.0",
		SessionID:     sessionID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Artifacts:     []Entry{},
	}

	if err := registry.SaveSessionRegistry(sessionReg); err != nil {
		t.Fatalf("Failed to save session registry: %v", err)
	}

	// Verify directory was created
	sessionDir := filepath.Join(tmpDir, ".claude", "sessions", sessionID)
	info, err := os.Stat(sessionDir)
	if err != nil {
		t.Fatalf("Session directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected session directory to be a directory")
	}
}

func TestRegistry_SaveProjectRegistry_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	projectReg := &ProjectRegistry{
		SchemaVersion:   "1.0",
		ProjectRoot:     tmpDir,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		SessionsIndexed: 0,
		Artifacts:       []Entry{},
		Indexes: ProjectIndexes{
			ByPhase:      make(map[Phase][]string),
			ByType:       make(map[ArtifactType][]string),
			BySpecialist: make(map[string][]string),
			BySession:    make(map[string][]string),
		},
	}

	if err := registry.SaveProjectRegistry(projectReg); err != nil {
		t.Fatalf("Failed to save project registry: %v", err)
	}

	// Verify directory was created
	artifactsDir := filepath.Join(tmpDir, ".claude", "artifacts")
	info, err := os.Stat(artifactsDir)
	if err != nil {
		t.Fatalf("Artifacts directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected artifacts directory to be a directory")
	}
}
