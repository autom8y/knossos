package bm25

import (
	"fmt"
	"os"
	"strings"

	registryorg "github.com/autom8y/knossos/internal/registry/org"
)

// ContentLoader loads the text content of a domain entry.
// Abstracted as an interface for testability — production uses LocalContentLoader,
// future implementations may use GitHub API for cross-repo content.
type ContentLoader interface {
	// LoadContent returns the markdown body (frontmatter stripped) for a domain entry.
	// Returns an error if the content cannot be loaded.
	LoadContent(entry registryorg.DomainEntry) (string, error)
}

// LocalContentLoader reads .know/ files from the local filesystem.
// It resolves file paths relative to a set of known repository root directories.
type LocalContentLoader struct {
	// RepoPaths maps repository names to their local filesystem root paths.
	// e.g., {"knossos": "/Users/tom/Code/knossos", "auth": "/Users/tom/Code/auth"}
	RepoPaths map[string]string
}

// LoadContent reads a domain file from the local filesystem, strips YAML frontmatter,
// and returns the body text.
func (l *LocalContentLoader) LoadContent(entry registryorg.DomainEntry) (string, error) {
	// Parse repo name from qualified name.
	parts := strings.SplitN(entry.QualifiedName, "::", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid qualified name: %s", entry.QualifiedName)
	}
	repoName := parts[1]

	repoRoot, ok := l.RepoPaths[repoName]
	if !ok {
		return "", fmt.Errorf("no local path for repo %s", repoName)
	}

	filePath := repoRoot + "/" + entry.Path
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", filePath, err)
	}

	return stripFrontmatter(string(data)), nil
}

// stripFrontmatter removes YAML frontmatter delimited by ---.
func stripFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---") {
		return text
	}
	rest := text[3:]
	idx := strings.Index(rest, "---")
	if idx < 0 {
		return text
	}
	return strings.TrimSpace(rest[idx+3:])
}

// BuildFromCatalog creates a BM25 index from a DomainCatalog.
// Iterates all domains, loads content via the ContentLoader, tokenizes,
// splits into sections, and adds to the index.
// Fail-open per domain: content load failures are logged and skipped.
func BuildFromCatalog(catalog *registryorg.DomainCatalog, loader ContentLoader) (*Index, error) {
	if catalog == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	idx := NewIndex()
	domains := catalog.ListDomains()

	for _, d := range domains {
		content, err := loader.LoadContent(d)
		if err != nil {
			// Fail-open: skip domains whose content cannot be loaded.
			continue
		}

		if strings.TrimSpace(content) == "" {
			continue
		}

		// G-4: Build boosted content that includes domain names, qualified names,
		// repo names, and section headings alongside the full content body.
		// This ensures queries like "autom8y-sms" that reference repo/domain
		// names directly produce BM25 signal even when the name does not appear
		// in the prose body.
		boostedContent := buildBoostedContent(d, content)

		// Document-level indexing with field-boosted content.
		freqs, totalTerms := BuildTermFreqs(boostedContent)
		if totalTerms == 0 {
			continue
		}

		doc := &IndexedUnit{
			QualifiedName: d.QualifiedName,
			Domain:        d.Domain,
			Title:         d.Domain,
			RawText:       content, // RawText keeps original content for display.
			TermFreqs:     freqs,
			TotalTerms:    totalTerms,
			GeneratedAt:   d.GeneratedAt,
		}
		idx.AddDocument(doc)

		// Section-level indexing.
		sections := SplitSections(content)
		for _, sec := range sections {
			if sec.Heading == "" {
				continue
			}
			secFreqs, secTotal := BuildTermFreqs(sec.Body)
			if secTotal < 3 {
				continue // Skip very short sections.
			}

			slug := Slugify(sec.Heading)
			secUnit := &IndexedUnit{
				QualifiedName: SectionQualifiedName(d.QualifiedName, slug),
				Domain:        d.Domain,
				Title:         sec.Heading,
				RawText:       sec.Body,
				TermFreqs:     secFreqs,
				TotalTerms:    secTotal,
				GeneratedAt:   d.GeneratedAt,
			}
			idx.AddSection(secUnit)
		}
	}

	idx.Finalize()
	return idx, nil
}

// buildBoostedContent prepends metadata fields to content for BM25 indexing.
// G-4: BM25 must index domain names, qualified names, repo names, and section
// headings alongside the summary/body text. Without this, queries referencing
// repo names (like "autom8y-sms") produce zero BM25 signal.
//
// The metadata is repeated 3 times to boost its weight relative to body text.
// This is a simple field-boosting strategy that avoids multi-field BM25 complexity.
func buildBoostedContent(entry registryorg.DomainEntry, content string) string {
	var b strings.Builder

	// Extract repo name from qualified name.
	parts := strings.SplitN(entry.QualifiedName, "::", 3)
	repoName := ""
	if len(parts) >= 2 {
		repoName = parts[1]
	}

	// Collect section headings from content.
	sections := SplitSections(content)
	var headings []string
	for _, sec := range sections {
		if sec.Heading != "" {
			headings = append(headings, sec.Heading)
		}
	}

	// Repeat metadata fields 3x for BM25 field boosting.
	for i := 0; i < 3; i++ {
		b.WriteString(entry.Domain)
		b.WriteString(" ")
		b.WriteString(entry.QualifiedName)
		b.WriteString(" ")
		if repoName != "" {
			b.WriteString(repoName)
			b.WriteString(" ")
		}
		for _, h := range headings {
			b.WriteString(h)
			b.WriteString(" ")
		}
	}

	// Append full content body (no truncation -- removed 200-char limit).
	b.WriteString(content)

	return b.String()
}