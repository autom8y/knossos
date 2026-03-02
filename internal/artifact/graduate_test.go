package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/paths"
)

func TestGraduateSession_EmptyRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)

	result, err := GraduateSession(resolver, "session-20260105-143022-abc12345")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result.Graduated) != 0 {
		t.Errorf("Expected 0 graduated, got %d", len(result.Graduated))
	}
	if len(result.Warnings) != 0 {
		t.Errorf("Expected 0 warnings, got %d", len(result.Warnings))
	}
}

func TestGraduateSession_CopiesArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)
	sessionID := "session-20260105-143022-abc12345"

	// Create source artifact
	srcRelPath := filepath.Join(".sos", "sessions", sessionID, "ADR-001.md")
	srcPath := filepath.Join(tmpDir, srcRelPath)
	if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPath, []byte("# ADR-001: Use Go\n\nDecision content."), 0644); err != nil {
		t.Fatal(err)
	}

	// Create session artifact registry
	registry := NewRegistry(tmpDir)
	entry := Entry{
		ArtifactID:   "ADR-001",
		ArtifactType: TypeADR,
		Path:         srcRelPath,
		Phase:        PhaseDesign,
		Specialist:   "context-architect",
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
		Validated:    true,
	}
	if err := registry.Register(sessionID, entry); err != nil {
		t.Fatal(err)
	}

	result, err := GraduateSession(resolver, sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Graduated) != 1 {
		t.Fatalf("Expected 1 graduated, got %d", len(result.Graduated))
	}

	ge := result.Graduated[0]
	if ge.ArtifactID != "ADR-001" {
		t.Errorf("Expected artifact ID ADR-001, got %s", ge.ArtifactID)
	}
	if ge.Category != "decisions" {
		t.Errorf("Expected category decisions, got %s", ge.Category)
	}
	if ge.GraduatedPath != filepath.Join(".ledge", "decisions", "ADR-001.md") {
		t.Errorf("Unexpected graduated path: %s", ge.GraduatedPath)
	}

	// Verify file exists at destination
	destPath := filepath.Join(tmpDir, ge.GraduatedPath)
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Graduated file not found: %v", err)
	}

	// Verify provenance frontmatter
	contentStr := string(content)
	if !strings.Contains(contentStr, "session_id: "+sessionID) {
		t.Error("Missing session_id in provenance frontmatter")
	}
	if !strings.Contains(contentStr, "graduated_at:") {
		t.Error("Missing graduated_at in provenance frontmatter")
	}
	if !strings.Contains(contentStr, "original_path: "+srcRelPath) {
		t.Error("Missing original_path in provenance frontmatter")
	}
	// Verify original content preserved
	if !strings.Contains(contentStr, "# ADR-001: Use Go") {
		t.Error("Original content not preserved in graduated file")
	}
}

func TestGraduateSession_MissingSourceWarning(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)
	sessionID := "session-20260105-143022-abc12345"

	// Register an artifact whose source file doesn't exist
	registry := NewRegistry(tmpDir)
	entry := Entry{
		ArtifactID:   "PRD-missing",
		ArtifactType: TypePRD,
		Path:         "docs/PRD-missing.md",
		Phase:        PhaseRequirements,
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
	}
	if err := registry.Register(sessionID, entry); err != nil {
		t.Fatal(err)
	}

	result, err := GraduateSession(resolver, sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Graduated) != 0 {
		t.Errorf("Expected 0 graduated (source missing), got %d", len(result.Graduated))
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(result.Warnings))
	}
	if !strings.Contains(result.Warnings[0], "cannot read source") {
		t.Errorf("Unexpected warning: %s", result.Warnings[0])
	}
}

func TestGraduateSession_CodeSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)
	sessionID := "session-20260105-143022-abc12345"

	// Create a code artifact — should be skipped
	srcPath := filepath.Join(tmpDir, "internal", "handler.go")
	if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPath, []byte("package internal"), 0644); err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry(tmpDir)
	entry := Entry{
		ArtifactID:   "CODE-handler",
		ArtifactType: TypeCode,
		Path:         "internal/handler.go",
		Phase:        PhaseImplementation,
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
	}
	if err := registry.Register(sessionID, entry); err != nil {
		t.Fatal(err)
	}

	result, err := GraduateSession(resolver, sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Graduated) != 0 {
		t.Errorf("Expected 0 graduated (code type skipped), got %d", len(result.Graduated))
	}
	if len(result.Warnings) != 0 {
		t.Errorf("Expected 0 warnings, got %d", len(result.Warnings))
	}

	// Verify no .ledge directory was created
	ledgeDir := filepath.Join(tmpDir, ".ledge")
	if _, err := os.Stat(ledgeDir); !os.IsNotExist(err) {
		t.Error("No .ledge directory should exist for code-only artifacts")
	}
}

func TestGraduateSession_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)
	sessionID := "session-20260105-143022-abc12345"

	// Create source artifact
	srcRelPath := filepath.Join(".sos", "sessions", sessionID, "SPIKE-caching.md")
	srcPath := filepath.Join(tmpDir, srcRelPath)
	if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPath, []byte("# Spike: Caching\n\nContent."), 0644); err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry(tmpDir)
	entry := Entry{
		ArtifactID:   "SPIKE-caching",
		ArtifactType: TypeSpike,
		Path:         srcRelPath,
		Phase:        PhaseDesign,
		SessionID:    sessionID,
		RegisteredAt: time.Now().UTC(),
	}
	if err := registry.Register(sessionID, entry); err != nil {
		t.Fatal(err)
	}

	// Graduate twice
	result1, err := GraduateSession(resolver, sessionID)
	if err != nil {
		t.Fatalf("First graduation failed: %v", err)
	}

	result2, err := GraduateSession(resolver, sessionID)
	if err != nil {
		t.Fatalf("Second graduation failed: %v", err)
	}

	if len(result1.Graduated) != len(result2.Graduated) {
		t.Errorf("Idempotency broken: first=%d, second=%d", len(result1.Graduated), len(result2.Graduated))
	}

	// Verify file still exists and is valid
	destPath := filepath.Join(tmpDir, result1.Graduated[0].GraduatedPath)
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Graduated file not found after second run: %v", err)
	}
	if !strings.Contains(string(content), "# Spike: Caching") {
		t.Error("Content corrupted after second graduation")
	}
}

func TestGraduateSession_MultipleTypes(t *testing.T) {
	tmpDir := t.TempDir()
	resolver := paths.NewResolver(tmpDir)
	sessionID := "session-20260105-143022-abc12345"

	// Create multiple artifact types
	artifacts := []struct {
		id       string
		typ      ArtifactType
		path     string
		content  string
		category string
	}{
		{"ADR-002", TypeADR, "session/ADR-002.md", "# ADR 002", "decisions"},
		{"PRD-feature", TypePRD, "session/PRD-feature.md", "# PRD Feature", "specs"},
		{"REVIEW-auth", TypeReview, "session/REVIEW-auth.md", "# Review Auth", "reviews"},
		{"CODE-impl", TypeCode, "internal/impl.go", "package internal", ""}, // skipped
	}

	registry := NewRegistry(tmpDir)
	for _, a := range artifacts {
		srcPath := filepath.Join(tmpDir, a.path)
		if err := os.MkdirAll(filepath.Dir(srcPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(srcPath, []byte(a.content), 0644); err != nil {
			t.Fatal(err)
		}
		if err := registry.Register(sessionID, Entry{
			ArtifactID:   a.id,
			ArtifactType: a.typ,
			Path:         a.path,
			SessionID:    sessionID,
			RegisteredAt: time.Now().UTC(),
		}); err != nil {
			t.Fatal(err)
		}
	}

	result, err := GraduateSession(resolver, sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// 3 graduated (code skipped)
	if len(result.Graduated) != 3 {
		t.Fatalf("Expected 3 graduated, got %d", len(result.Graduated))
	}

	// Verify categories
	categories := make(map[string]bool)
	for _, g := range result.Graduated {
		categories[g.Category] = true
	}
	for _, expected := range []string{"decisions", "specs", "reviews"} {
		if !categories[expected] {
			t.Errorf("Missing category: %s", expected)
		}
	}
}
