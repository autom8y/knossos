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
	if !got.TimeExpired {
		t.Errorf("TimeExpired = false, want true")
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

	_, ok := byDomain["architecture"]
	if !ok {
		t.Error("missing architecture domain")
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

// TestBuildDomainStatus_MatchingHash verifies that a domain with matching source_hash is fresh.
func TestBuildDomainStatus_MatchingHash(t *testing.T) {
	now := time.Now().UTC()
	generatedAt := now.Add(-1 * time.Hour).Format(time.RFC3339)
	meta := Meta{
		Domain:       "architecture",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "abc1234",
	}
	// Same hash as current HEAD
	status := buildDomainStatus(meta, now, "abc1234")
	if !status.Fresh {
		t.Errorf("Fresh = false, want true: matching hash and within expiry")
	}
	if status.CodeChanged {
		t.Errorf("CodeChanged = true, want false: hashes match")
	}
	if status.TimeExpired {
		t.Errorf("TimeExpired = true, want false: generated 1h ago with 7d expiry")
	}
}

// TestBuildDomainStatus_DifferingHash verifies that a differing source_hash marks domain stale.
func TestBuildDomainStatus_DifferingHash(t *testing.T) {
	now := time.Now().UTC()
	generatedAt := now.Add(-1 * time.Hour).Format(time.RFC3339)
	meta := Meta{
		Domain:       "architecture",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "oldsha1",
	}
	// Different hash: code has changed since knowledge was generated
	status := buildDomainStatus(meta, now, "newsha9")
	if status.Fresh {
		t.Errorf("Fresh = true, want false: source_hash differs from HEAD")
	}
	if !status.CodeChanged {
		t.Errorf("CodeChanged = false, want true: hashes differ")
	}
	if status.TimeExpired {
		t.Errorf("TimeExpired = true, want false: generated 1h ago with 7d expiry")
	}
	if status.CurrentHash != "newsha9" {
		t.Errorf("CurrentHash = %q, want %q", status.CurrentHash, "newsha9")
	}
}

// TestBuildDomainStatus_ExpiredAndCodeChanged verifies combined staleness reasons.
func TestBuildDomainStatus_ExpiredAndCodeChanged(t *testing.T) {
	now := time.Now().UTC()
	generatedAt := now.Add(-10 * 24 * time.Hour).Format(time.RFC3339) // 10 days ago
	meta := Meta{
		Domain:       "conventions",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "oldsha1",
	}
	status := buildDomainStatus(meta, now, "newsha9")
	if status.Fresh {
		t.Errorf("Fresh = true, want false: both expired and code changed")
	}
	if !status.TimeExpired {
		t.Errorf("TimeExpired = false, want true: generated 10d ago with 7d expiry")
	}
	if !status.CodeChanged {
		t.Errorf("CodeChanged = false, want true: hashes differ")
	}
}

// TestBuildDomainStatus_EmptyCurrentHash verifies that missing git hash skips code check.
func TestBuildDomainStatus_EmptyCurrentHash(t *testing.T) {
	now := time.Now().UTC()
	generatedAt := now.Add(-1 * time.Hour).Format(time.RFC3339)
	meta := Meta{
		Domain:       "architecture",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "abc1234",
	}
	// Empty currentHash: git unavailable -- code check skipped
	status := buildDomainStatus(meta, now, "")
	if !status.Fresh {
		t.Errorf("Fresh = false, want true: git unavailable, should not penalize")
	}
	if status.CodeChanged {
		t.Errorf("CodeChanged = true, want false: cannot check when currentHash is empty")
	}
}

// TestBuildDomainStatus_EmptySourceHash verifies that missing stored hash skips code check.
func TestBuildDomainStatus_EmptySourceHash(t *testing.T) {
	now := time.Now().UTC()
	generatedAt := now.Add(-1 * time.Hour).Format(time.RFC3339)
	meta := Meta{
		Domain:       "architecture",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "", // not set in frontmatter
	}
	status := buildDomainStatus(meta, now, "abc1234")
	if !status.Fresh {
		t.Errorf("Fresh = false, want true: no stored hash, cannot determine code change")
	}
	if status.CodeChanged {
		t.Errorf("CodeChanged = true, want false: no stored hash to compare")
	}
}

// TestBuildDomainStatus_ScopedStaleness_OutOfScope verifies that when source_scope is set
// and only out-of-scope files changed, the domain is treated as fresh (not stale).
func TestBuildDomainStatus_ScopedStaleness_OutOfScope(t *testing.T) {
	// Override gitDiffNameOnly to return only out-of-scope files.
	orig := gitDiffNameOnly
	defer func() { gitDiffNameOnly = orig }()
	gitDiffNameOnly = func(fromHash, toHash string) ([]string, error) {
		// Only docs changed -- not in internal/ or cmd/ scope.
		return []string{"docs/README.md", "rites/shared/mena/foo.lego.md"}, nil
	}

	now := time.Now().UTC()
	generatedAt := now.Add(-1 * time.Hour).Format(time.RFC3339)
	meta := Meta{
		Domain:       "architecture",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "oldsha1",
		SourceScope:  []string{"internal/**/*.go", "cmd/**/*.go"},
	}
	status := buildDomainStatus(meta, now, "newsha9")
	if status.CodeChanged {
		t.Errorf("CodeChanged = true, want false: only out-of-scope files changed")
	}
	if !status.Fresh {
		t.Errorf("Fresh = false, want true: out-of-scope changes should not cause staleness")
	}
}

// TestBuildDomainStatus_ScopedStaleness_InScope verifies that when source_scope is set
// and in-scope files changed, the domain is treated as stale.
func TestBuildDomainStatus_ScopedStaleness_InScope(t *testing.T) {
	// Override gitDiffNameOnly to return an in-scope file.
	orig := gitDiffNameOnly
	defer func() { gitDiffNameOnly = orig }()
	gitDiffNameOnly = func(fromHash, toHash string) ([]string, error) {
		// A Go source file in internal/ changed.
		return []string{"internal/know/know.go", "docs/README.md"}, nil
	}

	now := time.Now().UTC()
	generatedAt := now.Add(-1 * time.Hour).Format(time.RFC3339)
	meta := Meta{
		Domain:       "architecture",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "oldsha1",
		SourceScope:  []string{"internal/**/*.go", "cmd/**/*.go"},
	}
	status := buildDomainStatus(meta, now, "newsha9")
	if !status.CodeChanged {
		t.Errorf("CodeChanged = false, want true: in-scope file changed")
	}
	if status.Fresh {
		t.Errorf("Fresh = true, want false: in-scope change marks domain stale")
	}
}

// TestBuildDomainStatus_ScopedStaleness_EmptyScope verifies fallback to hash comparison
// when SourceScope is empty.
func TestBuildDomainStatus_ScopedStaleness_EmptyScope(t *testing.T) {
	// Ensure gitDiffNameOnly is never called when SourceScope is empty.
	orig := gitDiffNameOnly
	defer func() { gitDiffNameOnly = orig }()
	gitDiffNameOnly = func(fromHash, toHash string) ([]string, error) {
		t.Error("gitDiffNameOnly called unexpectedly with empty SourceScope")
		return nil, nil
	}

	now := time.Now().UTC()
	generatedAt := now.Add(-1 * time.Hour).Format(time.RFC3339)
	meta := Meta{
		Domain:       "architecture",
		GeneratedAt:  generatedAt,
		ExpiresAfter: "7d",
		SourceHash:   "oldsha1",
		SourceScope:  nil, // empty: fall back to hash comparison
	}
	status := buildDomainStatus(meta, now, "newsha9")
	if !status.CodeChanged {
		t.Errorf("CodeChanged = false, want true: hashes differ and no scope defined")
	}
}

// TestMatchScope verifies glob pattern matching for various source_scope values.
func TestMatchScope(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// Double-star glob patterns
		{"internal/**/*.go", "internal/know/know.go", true},
		{"internal/**/*.go", "internal/materialize/mena/types.go", true},
		{"internal/**/*.go", "internal/cmd/session/session.go", true},
		{"internal/**/*.go", "cmd/ari/main.go", false},
		{"internal/**/*.go", "rites/shared/mena/foo.lego.md", false},
		{"cmd/**/*.go", "cmd/ari/main.go", true},
		{"cmd/**/*.go", "internal/know/know.go", false},
		// Exact match (no glob)
		{"go.mod", "go.mod", true},
		{"go.mod", "go.sum", false},
		// Leading "./" stripped
		{"./internal/**/*.go", "internal/know/know.go", true},
		{"./cmd/**/*.go", "cmd/ari/main.go", true},
		// Non-Go files should not match *.go patterns
		{"internal/**/*.go", "internal/know/README.md", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s~%s", tt.pattern, tt.path), func(t *testing.T) {
			got := matchScope(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("matchScope(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

// TestScopedStaleness_GitUnavailable verifies that scopedStaleness returns false
// (not stale) when git is unavailable.
func TestScopedStaleness_GitUnavailable(t *testing.T) {
	orig := gitDiffNameOnly
	defer func() { gitDiffNameOnly = orig }()
	gitDiffNameOnly = func(fromHash, toHash string) ([]string, error) {
		return nil, fmt.Errorf("git: command not found")
	}

	result := scopedStaleness("abc1234", "def5678", []string{"internal/**/*.go"})
	if result {
		t.Error("scopedStaleness with unavailable git: want false (not stale), got true")
	}
}

// TestScopedStaleness_EmptyScope verifies that scopedStaleness returns false when scope is empty.
func TestScopedStaleness_EmptyScope(t *testing.T) {
	result := scopedStaleness("abc1234", "def5678", nil)
	if result {
		t.Error("scopedStaleness with empty scope: want false, got true")
	}
}

// TestScopedStaleness_EmptyHashes verifies that scopedStaleness returns false for empty hashes.
func TestScopedStaleness_EmptyHashes(t *testing.T) {
	result := scopedStaleness("", "def5678", []string{"internal/**/*.go"})
	if result {
		t.Error("scopedStaleness with empty fromHash: want false, got true")
	}
	result = scopedStaleness("abc1234", "", []string{"internal/**/*.go"})
	if result {
		t.Error("scopedStaleness with empty toHash: want false, got true")
	}
}
