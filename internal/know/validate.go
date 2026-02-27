// Package know provides shared parsing for .know/ codebase knowledge files.
package know

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/frontmatter"
)

// BrokenRef describes a single reference that failed validation.
type BrokenRef struct {
	Type    string // "file", "function", "commit"
	Ref     string // the raw reference string extracted
	Context string // surrounding text (truncated) for locating in source
	Error   string // why it failed: "file not found", "function not found in source", "git object not found"
}

// ValidationReport summarizes validation results for one domain.
type ValidationReport struct {
	Domain      string      // domain name from frontmatter
	TotalRefs   int         // total references extracted
	BrokenCount int         // len(Broken)
	Broken      []BrokenRef // details of each broken reference
}

// Regex patterns for reference extraction from .know/ markdown files.
// Scoped to internal/ and cmd/ trees to avoid false positives on prose, URLs, or config paths.
var (
	// reFilePath matches backtick-quoted file paths in recognized source trees.
	// Covers Go (internal/, cmd/), TypeScript/Python (src/, lib/, app/), and config files.
	// Example: `internal/know/know.go`, `src/lib/utils.ts`
	reFilePath = regexp.MustCompile("`((?:internal|cmd|src|lib|app)/[a-zA-Z0-9_/.-]+\\.[a-zA-Z0-9]+)`")

	// reBareFilePath matches unquoted file paths in recognized source trees.
	// Example: internal/know/know.go at word boundary.
	reBareFilePath = regexp.MustCompile(`(?:^|[\s(])((internal|cmd|src|lib|app)/[a-zA-Z0-9_/.-]+\.[a-zA-Z0-9]+)(?:[\s),:]|$)`)

	// reFuncRef matches backtick-quoted exported function names (PascalCase, 3+ chars).
	// Excludes single-letter or two-letter identifiers to avoid noise.
	// Example: `ValidateDomain()` or `BuildDomainStatus`
	reFuncRef = regexp.MustCompile("`([A-Z][a-zA-Z0-9_]{2,}(?:\\([^)]*\\))?)`")

	// reCommitHash matches backtick-quoted 7-40 hex character commit hashes.
	// Requires backtick quoting to distinguish from hex strings in prose.
	reCommitHash = regexp.MustCompile("`([0-9a-f]{7,40})`")
)

// refKey is used for deduplication: (type, ref) tuples.
type refKey struct {
	refType string
	ref     string
}

// extractedRef holds an extracted reference with its surrounding context.
type extractedRef struct {
	refType string
	ref     string
	context string
}

// ValidateDomain reads .know/{domain}.md, extracts references, and verifies each.
// rootDir is the project root (for file path resolution and git operations).
// Returns a report; returns error only for I/O failures (missing file, unreadable).
func ValidateDomain(rootDir, domain string) (*ValidationReport, error) {
	knowDir := filepath.Join(rootDir, ".know")
	path := filepath.Join(knowDir, domain+".md")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading .know/%s.md: %w", domain, err)
	}

	// Parse frontmatter to get domain name; fall back to filename stem.
	domainName := domain
	yamlBytes, body, err := frontmatter.Parse(data)
	if err == nil {
		var meta Meta
		if yaml.Unmarshal(yamlBytes, &meta) == nil && meta.Domain != "" {
			domainName = meta.Domain
		}
	} else {
		// No frontmatter: body is the full file content.
		body = data
	}

	refs := extractRefs(string(body))

	report := &ValidationReport{
		Domain:    domainName,
		TotalRefs: len(refs),
	}

	for _, r := range refs {
		var broken *BrokenRef
		switch r.refType {
		case "file":
			broken = verifyFileRef(rootDir, r)
		case "function":
			broken = verifyFuncRef(rootDir, r)
		case "commit":
			broken = verifyCommitRef(rootDir, r)
		}
		if broken != nil {
			report.Broken = append(report.Broken, *broken)
		}
	}

	report.BrokenCount = len(report.Broken)
	return report, nil
}

// ValidateAll globs .know/*.md and validates each domain.
// Returns one report per domain. Returns error only for I/O failures on the directory.
func ValidateAll(rootDir string) ([]ValidationReport, error) {
	knowDir := filepath.Join(rootDir, ".know")
	entries, err := os.ReadDir(knowDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read .know/ directory: %w", err)
	}

	var reports []ValidationReport
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		domain := strings.TrimSuffix(entry.Name(), ".md")
		report, err := ValidateDomain(rootDir, domain)
		if err != nil {
			// Log to stderr and continue; don't abort the whole run for one bad file.
			fmt.Fprintf(os.Stderr, "warn: cannot validate %s: %v\n", entry.Name(), err)
			continue
		}
		reports = append(reports, *report)
	}

	return reports, nil
}

// extractRefs scans markdown body text and extracts deduplicated references.
// Returns file, function, and commit references found in the content.
func extractRefs(body string) []extractedRef {
	seen := make(map[refKey]bool)
	var refs []extractedRef

	addRef := func(refType, ref, context string) {
		k := refKey{refType, ref}
		if seen[k] {
			return
		}
		seen[k] = true
		// Truncate context to 80 chars for readability in reports.
		if len(context) > 80 {
			context = context[:77] + "..."
		}
		refs = append(refs, extractedRef{refType: refType, ref: ref, context: strings.TrimSpace(context)})
	}

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		// Extract backtick-quoted file paths.
		for _, m := range reFilePath.FindAllStringSubmatch(line, -1) {
			addRef("file", m[1], line)
		}

		// Extract bare (unquoted) file paths.
		for _, m := range reBareFilePath.FindAllStringSubmatch(line, -1) {
			addRef("file", m[1], line)
		}

		// Extract exported function references.
		// We must not overlap with file path matches, so skip lines that have already
		// yielded a file path for each match position.
		for _, m := range reFuncRef.FindAllStringSubmatch(line, -1) {
			ref := m[1]
			// Strip trailing parens for a clean name to search.
			stripped := strings.TrimRight(ref, "()")
			// If the "function" reference looks like a path segment, skip it.
			if strings.Contains(stripped, "/") || strings.Contains(stripped, ".") {
				continue
			}
			addRef("function", ref, line)
		}

		// Extract commit hashes.
		for _, m := range reCommitHash.FindAllStringSubmatch(line, -1) {
			// Reject if it looks like a path or non-SHA pattern (all zeros is trivial).
			hash := m[1]
			if hash == strings.Repeat("0", len(hash)) {
				continue
			}
			addRef("commit", hash, line)
		}
	}

	return refs
}

// verifyFileRef checks that a file path reference exists on disk.
func verifyFileRef(rootDir string, r extractedRef) *BrokenRef {
	path := filepath.Join(rootDir, r.ref)
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return &BrokenRef{
			Type:    "file",
			Ref:     r.ref,
			Context: r.context,
			Error:   "file not found",
		}
	}
	// Other errors (permission, etc.) are not treated as broken.
	return nil
}

// verifyFuncRef checks that an exported function name exists somewhere in Go source files.
// Uses git grep for speed and .gitignore respect. Falls back to "skip" if git is unavailable.
func verifyFuncRef(rootDir string, r extractedRef) *BrokenRef {
	// Strip trailing parens to get the bare function name.
	funcName := strings.TrimSuffix(strings.TrimSuffix(r.ref, "()"), "(")
	// Also strip any remaining paren content like "(opts)"
	if idx := strings.Index(funcName, "("); idx >= 0 {
		funcName = funcName[:idx]
	}

	// git grep -l searches for the identifier as a func, const, type, or var declaration.
	// This covers exported functions, constants, types, and package-level variables.
	pattern := "(func|const|type|var).*" + funcName
	cmd := exec.Command("git", "grep", "-l", "-E", pattern, "--", "*.go")
	cmd.Dir = rootDir
	out, err := cmd.Output()
	if err != nil {
		// git grep returns exit code 1 when no matches found, and other codes for errors.
		// Exit code 1 with no output = not found. Other errors = git unavailable, skip.
		if isGitGrepNoMatch(err) {
			return &BrokenRef{
				Type:    "function",
				Ref:     r.ref,
				Context: r.context,
				Error:   "not found in source",
			}
		}
		// git unavailable or other error: skip (best-effort, don't report as broken).
		return nil
	}

	// Any output means at least one match was found.
	if strings.TrimSpace(string(out)) == "" {
		return &BrokenRef{
			Type:    "function",
			Ref:     r.ref,
			Context: r.context,
			Error:   "not found in source",
		}
	}

	return nil
}

// isGitGrepNoMatch returns true when the error is git grep's "no matches" exit code (1).
// Exit code 2+ indicates a real git error (unavailable, bad syntax, etc.).
func isGitGrepNoMatch(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode() == 1
	}
	return false
}

// verifyCommitRef checks that a commit hash is a known git object.
// Falls back to "skip" if git is unavailable (best-effort validation).
func verifyCommitRef(rootDir string, r extractedRef) *BrokenRef {
	cmd := exec.Command("git", "cat-file", "-t", r.ref)
	cmd.Dir = rootDir
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			// Exit 128 = git object not found.
			return &BrokenRef{
				Type:    "commit",
				Ref:     r.ref,
				Context: r.context,
				Error:   "git object not found",
			}
		}
		// git unavailable or other error: skip (best-effort).
		return nil
	}
	return nil
}
