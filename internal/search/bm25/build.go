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

		// Document-level indexing.
		freqs, totalTerms := BuildTermFreqs(content)
		if totalTerms == 0 {
			continue
		}

		doc := &IndexedUnit{
			QualifiedName: d.QualifiedName,
			Domain:        d.Domain,
			Title:         d.Domain,
			RawText:       truncate(content, 200),
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
				RawText:       truncate(sec.Body, 150),
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

// truncate returns the first n characters of a string, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
