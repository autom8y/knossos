package ledge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
)

func setupTestProject(t *testing.T) (string, *paths.Resolver) {
	t.Helper()
	dir := t.TempDir()

	// Create .claude/ so FindProjectRoot works
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

	// Create ledge categories
	for _, cat := range []string{"decisions", "specs", "reviews", "spikes"} {
		os.MkdirAll(filepath.Join(dir, ".ledge", cat), 0755)
	}

	// Create shelf categories
	for _, cat := range []string{"decisions", "specs", "reviews"} {
		os.MkdirAll(filepath.Join(dir, ".ledge", "shelf", cat), 0755)
	}

	return dir, paths.NewResolver(dir)
}

func TestPromote_BasicMove(t *testing.T) {
	dir, resolver := setupTestProject(t)

	// Create a test artifact in .ledge/reviews/
	src := filepath.Join(dir, ".ledge", "reviews", "test-review.md")
	os.WriteFile(src, []byte("# Test Review\nSome content\n"), 0644)

	result, err := Promote(resolver, ".ledge/reviews/test-review.md")
	if err != nil {
		t.Fatalf("Promote failed: %v", err)
	}

	if result.Category != "reviews" {
		t.Errorf("expected category=reviews, got %s", result.Category)
	}
	if result.ShelfPath != ".ledge/shelf/reviews/test-review.md" {
		t.Errorf("unexpected shelf path: %s", result.ShelfPath)
	}

	// Source should be removed
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source file should have been removed after promotion")
	}

	// Destination should exist with content
	dest := filepath.Join(dir, ".ledge", "shelf", "reviews", "test-review.md")
	content, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("cannot read promoted file: %v", err)
	}
	if !strings.Contains(string(content), "# Test Review") {
		t.Error("promoted file should contain original content")
	}
	if !strings.Contains(string(content), "promoted_at:") {
		t.Error("promoted file should contain promoted_at frontmatter")
	}
	if !strings.Contains(string(content), "promoted_from: .ledge/reviews/test-review.md") {
		t.Error("promoted file should contain promoted_from frontmatter")
	}
}

func TestPromote_WithExistingFrontmatter(t *testing.T) {
	dir, resolver := setupTestProject(t)

	// Create artifact with graduation frontmatter
	src := filepath.Join(dir, ".ledge", "decisions", "adr-001.md")
	graduated := "---\nsession_id: session-123\ngraduated_at: 2026-01-01T00:00:00Z\noriginal_path: .sos/sessions/s/adr.md\n---\n\n# ADR-001\nDecision content\n"
	os.WriteFile(src, []byte(graduated), 0644)

	result, err := Promote(resolver, ".ledge/decisions/adr-001.md")
	if err != nil {
		t.Fatalf("Promote failed: %v", err)
	}

	if result.Category != "decisions" {
		t.Errorf("expected category=decisions, got %s", result.Category)
	}

	dest := filepath.Join(dir, ".ledge", "shelf", "decisions", "adr-001.md")
	content, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("cannot read promoted file: %v", err)
	}

	s := string(content)
	// Should preserve graduation frontmatter AND add promotion fields
	if !strings.Contains(s, "session_id: session-123") {
		t.Error("should preserve session_id from graduation")
	}
	if !strings.Contains(s, "graduated_at:") {
		t.Error("should preserve graduated_at from graduation")
	}
	if !strings.Contains(s, "promoted_at:") {
		t.Error("should add promoted_at")
	}
	if !strings.Contains(s, "promoted_from: .ledge/decisions/adr-001.md") {
		t.Error("should add promoted_from")
	}
	if !strings.Contains(s, "# ADR-001") {
		t.Error("should preserve body content")
	}
}

func TestPromote_NonExistentSource(t *testing.T) {
	_, resolver := setupTestProject(t)

	_, err := Promote(resolver, ".ledge/reviews/nonexistent.md")
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
	if !strings.Contains(err.Error(), "source not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPromote_InvalidCategory(t *testing.T) {
	dir, resolver := setupTestProject(t)

	// Spikes are not promotable
	src := filepath.Join(dir, ".ledge", "spikes", "test.md")
	os.WriteFile(src, []byte("spike content"), 0644)

	_, err := Promote(resolver, ".ledge/spikes/test.md")
	if err == nil {
		t.Fatal("expected error for non-promotable category")
	}
	if !strings.Contains(err.Error(), "not promotable") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPromote_DestinationExists(t *testing.T) {
	dir, resolver := setupTestProject(t)

	src := filepath.Join(dir, ".ledge", "reviews", "dup.md")
	os.WriteFile(src, []byte("source"), 0644)

	dest := filepath.Join(dir, ".ledge", "shelf", "reviews", "dup.md")
	os.WriteFile(dest, []byte("already promoted"), 0644)

	_, err := Promote(resolver, ".ledge/reviews/dup.md")
	if err == nil {
		t.Fatal("expected error when destination already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPromote_OutsideLedge(t *testing.T) {
	_, resolver := setupTestProject(t)

	_, err := Promote(resolver, "/tmp/random-file.md")
	if err == nil {
		t.Fatal("expected error for file outside .ledge/")
	}
}

func TestPromote_AbsolutePath(t *testing.T) {
	dir, resolver := setupTestProject(t)

	src := filepath.Join(dir, ".ledge", "specs", "spec.md")
	os.WriteFile(src, []byte("spec content"), 0644)

	result, err := Promote(resolver, src)
	if err != nil {
		t.Fatalf("Promote with absolute path failed: %v", err)
	}
	if result.Category != "specs" {
		t.Errorf("expected category=specs, got %s", result.Category)
	}
}

func TestStampPromotionFrontmatter_NoExisting(t *testing.T) {
	content := []byte("# Just Markdown\nNo frontmatter here\n")
	result := stampPromotionFrontmatter(content, "2026-03-05T00:00:00Z", ".ledge/reviews/test.md")

	s := string(result)
	if !strings.HasPrefix(s, "---\n") {
		t.Error("should start with frontmatter delimiter")
	}
	if !strings.Contains(s, "promoted_at: 2026-03-05T00:00:00Z") {
		t.Error("should contain promoted_at")
	}
	if !strings.Contains(s, "# Just Markdown") {
		t.Error("should preserve original content")
	}
}

func TestStampPromotionFrontmatter_WithExisting(t *testing.T) {
	content := []byte("---\nsession_id: s1\ngraduated_at: 2026-01-01T00:00:00Z\n---\n\nBody\n")
	result := stampPromotionFrontmatter(content, "2026-03-05T00:00:00Z", ".ledge/reviews/test.md")

	s := string(result)
	// Should have exactly one frontmatter block
	count := strings.Count(s, "---\n")
	if count != 2 {
		t.Errorf("expected 2 frontmatter delimiters, got %d in:\n%s", count, s)
	}
	if !strings.Contains(s, "session_id: s1") {
		t.Error("should preserve existing frontmatter fields")
	}
	if !strings.Contains(s, "promoted_at: 2026-03-05T00:00:00Z") {
		t.Error("should add promoted_at")
	}
}
