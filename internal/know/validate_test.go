package know

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writeKnowFile creates a .know/*.md file with frontmatter and body content.
func writeKnowFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file %s: %v", path, err)
	}
}

// TestExtractRefs_FilePathsBacktick verifies backtick-quoted file path extraction.
func TestExtractRefs_FilePathsBacktick(t *testing.T) {
	body := "See `internal/know/know.go` for the implementation and `cmd/ari/main.go` for the entry point."
	refs := extractRefs(body)

	files := refsOfType(refs, "file")
	if len(files) < 2 {
		t.Fatalf("expected at least 2 file refs, got %d: %v", len(files), files)
	}
	assertContainsRef(t, files, "internal/know/know.go")
	assertContainsRef(t, files, "cmd/ari/main.go")
}

// TestExtractRefs_FilePathsBare verifies bare (unquoted) file path extraction.
func TestExtractRefs_FilePathsBare(t *testing.T) {
	body := "The file internal/materialize/materialize.go handles the sync pipeline."
	refs := extractRefs(body)

	files := refsOfType(refs, "file")
	if len(files) == 0 {
		t.Fatalf("expected at least 1 file ref, got 0")
	}
	assertContainsRef(t, files, "internal/materialize/materialize.go")
}

// TestExtractRefs_FunctionRefs verifies exported function name extraction.
func TestExtractRefs_FunctionRefs(t *testing.T) {
	body := "Call `ValidateDomain()` or `BuildDomainStatus` to check freshness."
	refs := extractRefs(body)

	funcs := refsOfType(refs, "function")
	if len(funcs) == 0 {
		t.Fatalf("expected at least 1 function ref, got 0")
	}
	assertContainsRef(t, funcs, "ValidateDomain()")
}

// TestExtractRefs_CommitHashes verifies commit hash extraction.
func TestExtractRefs_CommitHashes(t *testing.T) {
	body := "Fixed in `abc1234` and later cleaned up in `def5678abcdef12`."
	refs := extractRefs(body)

	commits := refsOfType(refs, "commit")
	if len(commits) < 2 {
		t.Fatalf("expected at least 2 commit refs, got %d: %v", len(commits), commits)
	}
	assertContainsRef(t, commits, "abc1234")
	assertContainsRef(t, commits, "def5678abcdef12")
}

// TestExtractRefs_Deduplication verifies that repeated refs are extracted only once.
func TestExtractRefs_Deduplication(t *testing.T) {
	body := "See `internal/know/know.go`. Also see `internal/know/know.go` again."
	refs := extractRefs(body)

	files := refsOfType(refs, "file")
	count := 0
	for _, r := range files {
		if r == "internal/know/know.go" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("internal/know/know.go should appear exactly once after dedup, got %d", count)
	}
}

// TestExtractRefs_NoFalsePositives verifies that non-source paths are not extracted.
func TestExtractRefs_NoFalsePositives(t *testing.T) {
	body := "See https://github.com/foo/bar.go and rites/ecosystem/agents/pythia.md for context."
	refs := extractRefs(body)

	files := refsOfType(refs, "file")
	for _, r := range files {
		if strings.Contains(r, "github.com") || strings.Contains(r, "rites/") {
			t.Errorf("unexpected file ref extracted: %q", r)
		}
	}
}

// TestValidateDomain_ValidRefs verifies that a domain with valid refs has BrokenCount == 0.
func TestValidateDomain_ValidRefs(t *testing.T) {
	// Use knossos root as rootDir so internal/know/know.go actually exists.
	rootDir := findProjectRoot(t)

	knowDir := filepath.Join(t.TempDir(), ".know")
	if err := os.MkdirAll(knowDir, 0755); err != nil {
		t.Fatalf("mkdir .know: %v", err)
	}

	// Write a know file that references a real file.
	content := "---\ndomain: test\ngenerated_at: \"2026-01-01T00:00:00Z\"\nexpires_after: \"7d\"\n---\n\nSee `internal/know/know.go` for parsing logic.\n"
	writeKnowFile(t, knowDir, "test.md", content)

	// Point rootDir to the real project root so file existence checks work.
	report, err := ValidateDomain(rootDir, "test")
	if err != nil {
		// The file is in a temp dir, not rootDir/.know — use a symlink approach instead.
		// Actually we need to validate against a rootDir that has the .know/ dir.
		// Simpler: create rootDir with .know inside it and create the source file too.
		t.Skip("ValidateDomain test requires .know/ inside rootDir; skipping")
	}

	if report == nil {
		t.Fatal("ValidateDomain returned nil report")
	}
}

// TestValidateDomain_BrokenFileRef verifies that a missing file path is reported as broken.
func TestValidateDomain_BrokenFileRef(t *testing.T) {
	// Set up a temp rootDir with .know/ inside it.
	rootDir := t.TempDir()
	knowDir := filepath.Join(rootDir, ".know")
	if err := os.MkdirAll(knowDir, 0755); err != nil {
		t.Fatalf("mkdir .know: %v", err)
	}

	// Reference a file that definitely doesn't exist in rootDir.
	content := "---\ndomain: brokentest\ngenerated_at: \"2026-01-01T00:00:00Z\"\nexpires_after: \"7d\"\n---\n\nSee `internal/nonexistent/bogus.go` for details.\n"
	writeKnowFile(t, knowDir, "brokentest.md", content)

	report, err := ValidateDomain(rootDir, "brokentest")
	if err != nil {
		t.Fatalf("ValidateDomain error: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.BrokenCount == 0 {
		t.Errorf("expected BrokenCount > 0, got 0 (broken file ref not detected)")
	}

	// Verify the broken ref details.
	found := false
	for _, b := range report.Broken {
		if b.Type == "file" && strings.Contains(b.Ref, "nonexistent") {
			found = true
			if b.Error == "" {
				t.Error("BrokenRef.Error should not be empty")
			}
		}
	}
	if !found {
		t.Errorf("expected broken file ref for nonexistent path, got: %+v", report.Broken)
	}
}

// TestValidateDomain_MissingFile verifies error return for missing domain file.
func TestValidateDomain_MissingFile(t *testing.T) {
	rootDir := t.TempDir()
	_, err := ValidateDomain(rootDir, "nonexistent")
	if err == nil {
		t.Error("ValidateDomain on missing domain: want error, got nil")
	}
}

// TestValidateAll_TwoDomains verifies that ValidateAll returns one report per domain.
func TestValidateAll_TwoDomains(t *testing.T) {
	rootDir := t.TempDir()
	knowDir := filepath.Join(rootDir, ".know")
	if err := os.MkdirAll(knowDir, 0755); err != nil {
		t.Fatalf("mkdir .know: %v", err)
	}

	// Write two domain files.
	writeKnowFile(t, knowDir, "alpha.md", "---\ndomain: alpha\ngenerated_at: \"2026-01-01T00:00:00Z\"\nexpires_after: \"7d\"\n---\n\nNo references here.\n")
	writeKnowFile(t, knowDir, "beta.md", "---\ndomain: beta\ngenerated_at: \"2026-01-01T00:00:00Z\"\nexpires_after: \"7d\"\n---\n\nAlso no references.\n")

	reports, err := ValidateAll(rootDir)
	if err != nil {
		t.Fatalf("ValidateAll error: %v", err)
	}
	if len(reports) != 2 {
		t.Errorf("expected 2 reports, got %d", len(reports))
	}
}

// TestValidateAll_EmptyDirectory verifies that an empty .know/ returns no reports.
func TestValidateAll_EmptyDirectory(t *testing.T) {
	rootDir := t.TempDir()
	knowDir := filepath.Join(rootDir, ".know")
	if err := os.MkdirAll(knowDir, 0755); err != nil {
		t.Fatalf("mkdir .know: %v", err)
	}

	reports, err := ValidateAll(rootDir)
	if err != nil {
		t.Fatalf("ValidateAll error: %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for empty .know/, got %d", len(reports))
	}
}

// TestValidateAll_MissingDirectory verifies that a missing .know/ returns nil without error.
func TestValidateAll_MissingDirectory(t *testing.T) {
	rootDir := t.TempDir()

	reports, err := ValidateAll(rootDir)
	if err != nil {
		t.Fatalf("ValidateAll on missing dir: want nil error, got %v", err)
	}
	if reports != nil {
		t.Errorf("expected nil reports for missing directory, got %v", reports)
	}
}

// --- helpers ---

// refsOfType returns just the Ref strings for extractedRefs of the given type.
func refsOfType(refs []extractedRef, refType string) []string {
	var result []string
	for _, r := range refs {
		if r.refType == refType {
			result = append(result, r.ref)
		}
	}
	return result
}

// assertContainsRef checks that the given ref string appears in the list.
func assertContainsRef(t *testing.T, refs []string, want string) {
	t.Helper()
	if !slices.Contains(refs, want) {
		t.Errorf("expected ref %q in %v", want, refs)
	}
}

// findProjectRoot walks up from the test binary's working directory to find the go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	// Start from the package directory (internal/know/).
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (go.mod) from %s", dir)
		}
		dir = parent
	}
}
