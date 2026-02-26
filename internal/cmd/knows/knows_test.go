package knows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/know"
)

// writeFrontmatter creates a .know/*.md file with the given YAML frontmatter.
func writeFrontmatter(t *testing.T, dir, filename, fm string) {
	t.Helper()
	content := fmt.Sprintf("---\n%s---\n\n# Body content\n", fm)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func TestKnowsOutput_Text_Empty(t *testing.T) {
	out := KnowsOutput{Domains: nil, AllFresh: true}
	text := out.Text()
	if !strings.Contains(text, "No codebase knowledge") {
		t.Errorf("empty Text() should mention no knowledge, got: %q", text)
	}
}

func TestKnowsOutput_Text_WithDomains(t *testing.T) {
	domains := []know.DomainStatus{
		{
			Domain:     "architecture",
			Generated:  "2026-02-26T21:17:58Z",
			Expires:    "2026-03-05",
			Fresh:      true,
			SourceHash: "a9149e7",
			Confidence: 0.88,
		},
	}
	out := KnowsOutput{Domains: domains, AllFresh: true}
	text := out.Text()

	if !strings.Contains(text, "architecture") {
		t.Errorf("Text() should contain domain name, got: %q", text)
	}
	if !strings.Contains(text, "fresh") {
		t.Errorf("Text() should contain 'fresh' status, got: %q", text)
	}
	if !strings.Contains(text, "a9149e7") {
		t.Errorf("Text() should contain source hash, got: %q", text)
	}
}

func TestKnowsOutput_Text_StaleDomain(t *testing.T) {
	domains := []know.DomainStatus{
		{
			Domain:     "architecture",
			Generated:  "2026-01-01T00:00:00Z",
			Expires:    "2026-01-08",
			Fresh:      false,
			SourceHash: "old1234",
			Confidence: 0.70,
		},
	}
	out := KnowsOutput{Domains: domains, AllFresh: false}
	text := out.Text()

	if !strings.Contains(text, "STALE") {
		t.Errorf("Text() should contain 'STALE' for expired domain, got: %q", text)
	}
}

func TestReadSingleDomain_Missing(t *testing.T) {
	dir := t.TempDir()
	err := readSingleDomain(dir, "nonexistent")
	if err == nil {
		t.Error("readSingleDomain with missing file: want error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %q", err.Error())
	}
}

func TestReadSingleDomain_Exists(t *testing.T) {
	dir := t.TempDir()
	content := "---\ndomain: architecture\n---\n\n# Architecture\nSome content here.\n"
	if err := os.WriteFile(filepath.Join(dir, "architecture.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := readSingleDomain(dir, "architecture")

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	captured := string(buf[:n])

	if err != nil {
		t.Errorf("readSingleDomain: unexpected error: %v", err)
	}
	if !strings.Contains(captured, "Architecture") {
		t.Errorf("readSingleDomain output should contain file content, got: %q", captured)
	}
}

func TestFreshDomainDetection(t *testing.T) {
	dir := t.TempDir()
	// Generated 1 day ago, expires in 7 days = fresh
	generatedAt := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	writeFrontmatter(t, dir, "architecture.md", fmt.Sprintf(`domain: architecture
generated_at: "%s"
expires_after: "7d"
source_hash: "abc1234"
confidence: 0.88
format_version: "1.0"
`, generatedAt))

	domains, err := know.ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("want 1 domain, got %d", len(domains))
	}
	if !domains[0].Fresh {
		t.Error("domain generated 1d ago with 7d expiry should be fresh")
	}
}

func TestStaleDomainDetection(t *testing.T) {
	dir := t.TempDir()
	// Generated 10 days ago, expires in 7 days = stale
	generatedAt := time.Now().UTC().Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	writeFrontmatter(t, dir, "architecture.md", fmt.Sprintf(`domain: architecture
generated_at: "%s"
expires_after: "7d"
source_hash: "abc1234"
confidence: 0.88
format_version: "1.0"
`, generatedAt))

	domains, err := know.ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("want 1 domain, got %d", len(domains))
	}
	if domains[0].Fresh {
		t.Error("domain generated 10d ago with 7d expiry should be stale")
	}
}

func TestMissingKnowDirectory(t *testing.T) {
	dir := t.TempDir()
	knowDir := filepath.Join(dir, ".know")
	// Don't create the directory

	domains, err := know.ReadMeta(knowDir)
	if err != nil {
		t.Errorf("ReadMeta on missing directory: want nil error, got %v", err)
	}
	if domains != nil {
		t.Errorf("ReadMeta on missing directory: want nil slice, got %v", domains)
	}
}

func TestMalformedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	// Write a file with no frontmatter
	if err := os.WriteFile(filepath.Join(dir, "broken.md"), []byte("# No frontmatter\n"), 0644); err != nil {
		t.Fatalf("write broken file: %v", err)
	}

	domains, err := know.ReadMeta(dir)
	if err != nil {
		t.Errorf("ReadMeta with broken file: want nil error, got %v", err)
	}
	// Broken file should be skipped silently
	if len(domains) != 0 {
		t.Errorf("want 0 domains (broken skipped), got %d", len(domains))
	}
}
