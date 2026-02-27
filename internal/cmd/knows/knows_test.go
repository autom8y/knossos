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
			Domain:      "architecture",
			Generated:   "2026-02-26T21:17:58Z",
			Expires:     "2026-03-05",
			Fresh:       true,
			TimeExpired: false,
			CodeChanged: false,
			SourceHash:  "a9149e7",
			Confidence:  0.88,
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

func TestKnowsOutput_Text_StaleDomain_Expired(t *testing.T) {
	domains := []know.DomainStatus{
		{
			Domain:      "architecture",
			Generated:   "2026-01-01T00:00:00Z",
			Expires:     "2026-01-08",
			Fresh:       false,
			TimeExpired: true,
			CodeChanged: false,
			SourceHash:  "old1234",
			Confidence:  0.70,
		},
	}
	out := KnowsOutput{Domains: domains, AllFresh: false}
	text := out.Text()

	if !strings.Contains(text, "stale (expired)") {
		t.Errorf("Text() should contain 'stale (expired)' for time-expired domain, got: %q", text)
	}
}

func TestKnowsOutput_Text_StaleDomain_CodeChanged(t *testing.T) {
	domains := []know.DomainStatus{
		{
			Domain:      "architecture",
			Generated:   "2026-02-26T00:00:00Z",
			Expires:     "2026-03-05",
			Fresh:       false,
			TimeExpired: false,
			CodeChanged: true,
			SourceHash:  "old1234",
			CurrentHash: "newsha9",
			Confidence:  0.85,
		},
	}
	out := KnowsOutput{Domains: domains, AllFresh: false}
	text := out.Text()

	if !strings.Contains(text, "stale (code changed)") {
		t.Errorf("Text() should contain 'stale (code changed)', got: %q", text)
	}
}

func TestKnowsOutput_Text_StaleDomain_BothReasons(t *testing.T) {
	domains := []know.DomainStatus{
		{
			Domain:      "architecture",
			Generated:   "2026-01-01T00:00:00Z",
			Expires:     "2026-01-08",
			Fresh:       false,
			TimeExpired: true,
			CodeChanged: true,
			SourceHash:  "old1234",
			CurrentHash: "newsha9",
			Confidence:  0.70,
		},
	}
	out := KnowsOutput{Domains: domains, AllFresh: false}
	text := out.Text()

	if !strings.Contains(text, "stale (expired + code changed)") {
		t.Errorf("Text() should contain 'stale (expired + code changed)', got: %q", text)
	}
}

func TestStalenessLabel_Fresh(t *testing.T) {
	d := know.DomainStatus{Fresh: true}
	if got := stalenessLabel(d); got != "fresh" {
		t.Errorf("stalenessLabel(fresh) = %q, want %q", got, "fresh")
	}
}

func TestStalenessLabel_ExpiredOnly(t *testing.T) {
	d := know.DomainStatus{Fresh: false, TimeExpired: true, CodeChanged: false}
	if got := stalenessLabel(d); got != "stale (expired)" {
		t.Errorf("stalenessLabel(expired only) = %q, want %q", got, "stale (expired)")
	}
}

func TestStalenessLabel_CodeChangedOnly(t *testing.T) {
	d := know.DomainStatus{Fresh: false, TimeExpired: false, CodeChanged: true}
	if got := stalenessLabel(d); got != "stale (code changed)" {
		t.Errorf("stalenessLabel(code changed only) = %q, want %q", got, "stale (code changed)")
	}
}

func TestStalenessLabel_BothReasons(t *testing.T) {
	d := know.DomainStatus{Fresh: false, TimeExpired: true, CodeChanged: true}
	if got := stalenessLabel(d); got != "stale (expired + code changed)" {
		t.Errorf("stalenessLabel(both) = %q, want %q", got, "stale (expired + code changed)")
	}
}

func TestStalenessLabel_Default(t *testing.T) {
	// Stale but neither TimeExpired nor CodeChanged: unparseable timestamp case
	d := know.DomainStatus{Fresh: false, TimeExpired: false, CodeChanged: false}
	if got := stalenessLabel(d); got != "stale" {
		t.Errorf("stalenessLabel(default) = %q, want %q", got, "stale")
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
	// Note: Fresh may be false if current git HEAD differs from "abc1234".
	// We only assert TimeExpired is false since we cannot control git hash in tests.
	if domains[0].TimeExpired {
		t.Error("domain generated 1d ago with 7d expiry should not be time-expired")
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
	if !domains[0].TimeExpired {
		t.Error("domain generated 10d ago with 7d expiry should have TimeExpired=true")
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

// TestCheckMode_CodeStale verifies that --check logic treats code-stale domains as stale.
// We test this through the stalenessLabel function and domain status fields rather than
// the full cobra command (which requires project resolution).
func TestCheckMode_CodeStale(t *testing.T) {
	// Simulate a domain that is time-fresh but code-changed
	d := know.DomainStatus{
		Domain:      "architecture",
		Fresh:       false,
		TimeExpired: false,
		CodeChanged: true,
		SourceHash:  "oldsha1",
		CurrentHash: "newsha9",
		Expires:     "2026-03-05",
	}

	if d.Fresh {
		t.Error("code-stale domain should have Fresh=false")
	}
	label := stalenessLabel(d)
	if label != "stale (code changed)" {
		t.Errorf("stalenessLabel = %q, want %q", label, "stale (code changed)")
	}
}

// TestValidateOutput_Text_AllValid verifies text output when all domains are valid.
func TestValidateOutput_Text_AllValid(t *testing.T) {
	out := ValidateOutput{
		Reports: []know.ValidationReport{
			{Domain: "architecture", TotalRefs: 23, BrokenCount: 0},
			{Domain: "conventions", TotalRefs: 18, BrokenCount: 0},
		},
		AllValid:   true,
		TotalRefs:  41,
		BrokenRefs: 0,
	}
	text := out.Text()

	if !strings.Contains(text, "architecture") {
		t.Errorf("Text() should contain 'architecture', got: %q", text)
	}
	if !strings.Contains(text, "conventions") {
		t.Errorf("Text() should contain 'conventions', got: %q", text)
	}
	if !strings.Contains(text, "valid") {
		t.Errorf("Text() should contain 'valid' status label, got: %q", text)
	}
	if strings.Contains(text, "BROKEN") {
		t.Errorf("Text() should not contain 'BROKEN' when all valid, got: %q", text)
	}
	if strings.Contains(text, "Broken References:") {
		t.Errorf("Text() should not show broken ref details when all valid, got: %q", text)
	}
}

// TestValidateOutput_Text_WithBroken verifies text output when some domains have broken refs.
func TestValidateOutput_Text_WithBroken(t *testing.T) {
	out := ValidateOutput{
		Reports: []know.ValidationReport{
			{
				Domain:      "scar-tissue",
				TotalRefs:   41,
				BrokenCount: 2,
				Broken: []know.BrokenRef{
					{
						Type:    "file",
						Ref:     "internal/materialize/throughline_cleanup_test.go",
						Context: "See `internal/materialize/throughline_cleanup_test.go` for the fix.",
						Error:   "file not found",
					},
					{
						Type:    "commit",
						Ref:     "80b176e",
						Context: "Fixed in `80b176e`.",
						Error:   "git object not found",
					},
				},
			},
			{
				Domain:      "architecture",
				TotalRefs:   23,
				BrokenCount: 0,
			},
		},
		AllValid:   false,
		TotalRefs:  64,
		BrokenRefs: 2,
	}
	text := out.Text()

	if !strings.Contains(text, "BROKEN") {
		t.Errorf("Text() should contain 'BROKEN' label for broken domain, got: %q", text)
	}
	if !strings.Contains(text, "Broken References:") {
		t.Errorf("Text() should contain 'Broken References:' section, got: %q", text)
	}
	if !strings.Contains(text, "scar-tissue") {
		t.Errorf("Text() should contain 'scar-tissue' domain name, got: %q", text)
	}
	if !strings.Contains(text, "[file]") {
		t.Errorf("Text() should contain '[file]' type label, got: %q", text)
	}
	if !strings.Contains(text, "[commit]") {
		t.Errorf("Text() should contain '[commit]' type label, got: %q", text)
	}
	if !strings.Contains(text, "80b176e") {
		t.Errorf("Text() should contain the broken commit hash, got: %q", text)
	}
	if !strings.Contains(text, "file not found") {
		t.Errorf("Text() should contain error description, got: %q", text)
	}

	// The valid domain should still appear as valid.
	if !strings.Contains(text, "architecture") {
		t.Errorf("Text() should still contain 'architecture' domain, got: %q", text)
	}
}
