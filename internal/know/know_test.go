package know

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"7d", 7 * 24 * time.Hour, false},
		{"14d", 14 * 24 * time.Hour, false},
		{"30d", 30 * 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"0d", 0, false},
		{"2h", 2 * time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"90s", 90 * time.Second, false},
		{"1h30m", 90 * time.Minute, false},
		{"", 0, true},
		{"xd", 0, true},
		{"-1d", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDuration(%q) = %v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDuration(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestReadMeta_MissingDirectory(t *testing.T) {
	results, err := ReadMeta("/nonexistent/path/.know")
	if err != nil {
		t.Errorf("ReadMeta missing dir: want nil error, got %v", err)
	}
	if results != nil {
		t.Errorf("ReadMeta missing dir: want nil slice, got %v", results)
	}
}

func TestReadMeta_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	results, err := ReadMeta(dir)
	if err != nil {
		t.Errorf("ReadMeta empty dir: want nil error, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("ReadMeta empty dir: want 0 results, got %d", len(results))
	}
}

func TestReadMeta_FreshDomain(t *testing.T) {
	dir := t.TempDir()
	// generated 1 day ago, expires in 7 days = 6 days remaining = fresh
	generatedAt := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	writeFrontmatter(t, dir, "architecture.md", fmt.Sprintf(`domain: architecture
generated_at: "%s"
expires_after: "7d"
source_hash: "abc1234"
confidence: 0.88
format_version: "1.0"
`, generatedAt))

	results, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}

	got := results[0]
	if got.Domain != "architecture" {
		t.Errorf("domain = %q, want %q", got.Domain, "architecture")
	}
	if !got.Fresh {
		t.Errorf("Fresh = false, want true (generated 1d ago, expires in 7d)")
	}
	if got.SourceHash != "abc1234" {
		t.Errorf("SourceHash = %q, want %q", got.SourceHash, "abc1234")
	}
	if got.Confidence != 0.88 {
		t.Errorf("Confidence = %f, want 0.88", got.Confidence)
	}
}

func TestReadMeta_StaleDomain(t *testing.T) {
	dir := t.TempDir()
	// generated 10 days ago, expires in 7 days = stale
	generatedAt := time.Now().UTC().Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	writeFrontmatter(t, dir, "stale.md", fmt.Sprintf(`domain: stale
generated_at: "%s"
expires_after: "7d"
source_hash: "old123"
confidence: 0.70
format_version: "1.0"
`, generatedAt))

	results, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}

	got := results[0]
	if got.Fresh {
		t.Errorf("Fresh = true, want false (generated 10d ago, expires in 7d)")
	}
}

func TestReadMeta_MalformedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	// File with no frontmatter delimiter - should be skipped gracefully
	path := filepath.Join(dir, "broken.md")
	if err := os.WriteFile(path, []byte("# No frontmatter here\n"), 0644); err != nil {
		t.Fatalf("write broken file: %v", err)
	}

	results, err := ReadMeta(dir)
	if err != nil {
		t.Errorf("ReadMeta with malformed file: want nil error, got %v", err)
	}
	// Broken file is skipped
	if len(results) != 0 {
		t.Errorf("want 0 results (malformed skipped), got %d", len(results))
	}
}

func TestReadMeta_MultipleDomains(t *testing.T) {
	dir := t.TempDir()
	recentAt := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	oldAt := time.Now().UTC().Add(-30 * 24 * time.Hour).Format(time.RFC3339)

	writeFrontmatter(t, dir, "architecture.md", fmt.Sprintf(`domain: architecture
generated_at: "%s"
expires_after: "7d"
source_hash: "fresh1"
confidence: 0.90
format_version: "1.0"
`, recentAt))

	writeFrontmatter(t, dir, "conventions.md", fmt.Sprintf(`domain: conventions
generated_at: "%s"
expires_after: "14d"
source_hash: "stale2"
confidence: 0.75
format_version: "1.0"
`, oldAt))

	results, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}

	// Find by domain name
	byDomain := make(map[string]DomainStatus)
	for _, r := range results {
		byDomain[r.Domain] = r
	}

	arch, ok := byDomain["architecture"]
	if !ok {
		t.Error("missing architecture domain")
	} else if !arch.Fresh {
		t.Error("architecture should be fresh")
	}

	conv, ok := byDomain["conventions"]
	if !ok {
		t.Error("missing conventions domain")
	} else if conv.Fresh {
		t.Error("conventions should be stale (30d old, 14d expiry)")
	}
}

func TestReadMeta_IgnoresNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	// Write a non-.md file that should be ignored
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	recentAt := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	writeFrontmatter(t, dir, "architecture.md", fmt.Sprintf(`domain: architecture
generated_at: "%s"
expires_after: "7d"
source_hash: "abc"
confidence: 0.80
format_version: "1.0"
`, recentAt))

	results, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result (txt ignored), got %d", len(results))
	}
}
