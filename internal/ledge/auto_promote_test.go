package ledge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/artifact"
)

func TestAutoPromoteSession_AllPromotable(t *testing.T) {
	dir, resolver := setupTestProject(t)

	// Create graduated artifacts in .ledge/
	entries := []artifact.GraduatedEntry{
		{ArtifactID: "a1", GraduatedPath: ".ledge/decisions/adr-001.md", Category: "decisions"},
		{ArtifactID: "a2", GraduatedPath: ".ledge/specs/spec-001.md", Category: "specs"},
		{ArtifactID: "a3", GraduatedPath: ".ledge/reviews/review-001.md", Category: "reviews"},
	}

	for _, e := range entries {
		p := filepath.Join(dir, e.GraduatedPath)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("---\nsession_id: s1\n---\n\n# Content\n"), 0644)
	}

	result, err := AutoPromoteSession(resolver, entries)
	if err != nil {
		t.Fatalf("AutoPromoteSession failed: %v", err)
	}

	if len(result.Promoted) != 3 {
		t.Errorf("expected 3 promoted, got %d", len(result.Promoted))
	}
	if result.Skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", result.Skipped)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %v", result.Warnings)
	}

	// Verify files moved to shelf
	for _, p := range result.Promoted {
		shelfPath := filepath.Join(dir, p.ShelfPath)
		if _, err := os.Stat(shelfPath); os.IsNotExist(err) {
			t.Errorf("shelf file not found: %s", shelfPath)
		}
	}
}

func TestAutoPromoteSession_WithSpikes(t *testing.T) {
	dir, resolver := setupTestProject(t)

	entries := []artifact.GraduatedEntry{
		{ArtifactID: "a1", GraduatedPath: ".ledge/decisions/adr.md", Category: "decisions"},
		{ArtifactID: "a2", GraduatedPath: ".ledge/specs/spec.md", Category: "specs"},
		{ArtifactID: "a3", GraduatedPath: ".ledge/reviews/rev.md", Category: "reviews"},
		{ArtifactID: "a4", GraduatedPath: ".ledge/spikes/spike.md", Category: "spikes"},
	}

	for _, e := range entries {
		p := filepath.Join(dir, e.GraduatedPath)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("content"), 0644)
	}

	result, err := AutoPromoteSession(resolver, entries)
	if err != nil {
		t.Fatalf("AutoPromoteSession failed: %v", err)
	}

	if len(result.Promoted) != 3 {
		t.Errorf("expected 3 promoted, got %d", len(result.Promoted))
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}
}

func TestAutoPromoteSession_SourceMissing(t *testing.T) {
	_, resolver := setupTestProject(t)

	entries := []artifact.GraduatedEntry{
		{ArtifactID: "a1", GraduatedPath: ".ledge/decisions/missing.md", Category: "decisions"},
	}

	result, err := AutoPromoteSession(resolver, entries)
	if err != nil {
		t.Fatalf("AutoPromoteSession failed: %v", err)
	}

	if len(result.Promoted) != 0 {
		t.Errorf("expected 0 promoted, got %d", len(result.Promoted))
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings for missing source (silent skip), got %v", result.Warnings)
	}
}

func TestAutoPromoteSession_DestExists(t *testing.T) {
	dir, resolver := setupTestProject(t)

	// Create source in .ledge/
	src := filepath.Join(dir, ".ledge", "reviews", "dup.md")
	os.WriteFile(src, []byte("source content"), 0644)

	// Pre-create destination in shelf
	dest := filepath.Join(dir, ".ledge", "shelf", "reviews", "dup.md")
	os.WriteFile(dest, []byte("already there"), 0644)

	entries := []artifact.GraduatedEntry{
		{ArtifactID: "a1", GraduatedPath: ".ledge/reviews/dup.md", Category: "reviews"},
	}

	result, err := AutoPromoteSession(resolver, entries)
	if err != nil {
		t.Fatalf("AutoPromoteSession failed: %v", err)
	}

	if len(result.Promoted) != 0 {
		t.Errorf("expected 0 promoted, got %d", len(result.Promoted))
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
	if !strings.Contains(result.Warnings[0], "already exists") {
		t.Errorf("warning should mention 'already exists', got: %s", result.Warnings[0])
	}
}

func TestAutoPromoteSession_EmptyGraduated(t *testing.T) {
	_, resolver := setupTestProject(t)

	result, err := AutoPromoteSession(resolver, nil)
	if err != nil {
		t.Fatalf("AutoPromoteSession failed: %v", err)
	}

	if len(result.Promoted) != 0 {
		t.Errorf("expected 0 promoted, got %d", len(result.Promoted))
	}
	if result.Skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", result.Skipped)
	}
}

func TestAutoPromoteSession_NilResolver(t *testing.T) {
	entries := []artifact.GraduatedEntry{
		{ArtifactID: "a1", GraduatedPath: ".ledge/decisions/adr.md", Category: "decisions"},
	}

	_, err := AutoPromoteSession(nil, entries)
	if err == nil {
		t.Fatal("expected error for nil resolver")
	}
}

func TestAutoPromoteSession_MixedCategories(t *testing.T) {
	dir, resolver := setupTestProject(t)

	entries := []artifact.GraduatedEntry{
		{ArtifactID: "a1", GraduatedPath: ".ledge/decisions/adr.md", Category: "decisions"},
		{ArtifactID: "a2", GraduatedPath: ".ledge/specs/spec.md", Category: "specs"},
		{ArtifactID: "a3", GraduatedPath: ".ledge/reviews/rev.md", Category: "reviews"},
		{ArtifactID: "a4", GraduatedPath: ".ledge/spikes/spike.md", Category: "spikes"},
		{ArtifactID: "a5", GraduatedPath: ".ledge/code/main.go", Category: "code"},
	}

	// Create source files for promotable categories
	for _, e := range entries {
		p := filepath.Join(dir, e.GraduatedPath)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("content"), 0644)
	}

	result, err := AutoPromoteSession(resolver, entries)
	if err != nil {
		t.Fatalf("AutoPromoteSession failed: %v", err)
	}

	if len(result.Promoted) != 3 {
		t.Errorf("expected 3 promoted, got %d", len(result.Promoted))
	}
	if result.Skipped != 2 {
		t.Errorf("expected 2 skipped (spikes + code), got %d", result.Skipped)
	}
}

func TestAutoPromoteSession_FrontmatterPreservation(t *testing.T) {
	dir, resolver := setupTestProject(t)

	graduated := "---\nsession_id: session-123\ngraduated_at: 2026-01-01T00:00:00Z\noriginal_path: .sos/sessions/s/adr.md\n---\n\n# ADR-001\nDecision content\n"
	src := filepath.Join(dir, ".ledge", "decisions", "adr-001.md")
	os.WriteFile(src, []byte(graduated), 0644)

	entries := []artifact.GraduatedEntry{
		{ArtifactID: "a1", GraduatedPath: ".ledge/decisions/adr-001.md", Category: "decisions"},
	}

	result, err := AutoPromoteSession(resolver, entries)
	if err != nil {
		t.Fatalf("AutoPromoteSession failed: %v", err)
	}

	if len(result.Promoted) != 1 {
		t.Fatalf("expected 1 promoted, got %d", len(result.Promoted))
	}

	// Read shelf file and verify both graduation and promotion frontmatter
	shelfPath := filepath.Join(dir, result.Promoted[0].ShelfPath)
	content, err := os.ReadFile(shelfPath)
	if err != nil {
		t.Fatalf("cannot read shelf file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "session_id: session-123") {
		t.Error("should preserve session_id from graduation")
	}
	if !strings.Contains(s, "graduated_at:") {
		t.Error("should preserve graduated_at from graduation")
	}
	if !strings.Contains(s, "promoted_at:") {
		t.Error("should add promoted_at")
	}
	if !strings.Contains(s, "promoted_from:") {
		t.Error("should add promoted_from")
	}
	if !strings.Contains(s, "# ADR-001") {
		t.Error("should preserve body content")
	}
}
