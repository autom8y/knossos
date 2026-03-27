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

// BuildFromCatalog creates a BM25 index from a DomainCatalog using default
// ari ask BM25 parameters. Iterates all domains, loads content via the
// ContentLoader, tokenizes, splits into sections, and adds to the index.
// Fail-open per domain: content load failures are logged and skipped.
func BuildFromCatalog(catalog *registryorg.DomainCatalog, loader ContentLoader) (*Index, error) {
	return BuildFromCatalogWithScorer(catalog, loader, nil)
}

// BuildFromCatalogWithScorer creates a BM25 index from a DomainCatalog using
// a custom BM25 scorer. When scorer is nil, default ari ask parameters are used.
// This enables Clew knowledge retrieval to use different length normalization
// than ari ask (R-4 isolation).
func BuildFromCatalogWithScorer(catalog *registryorg.DomainCatalog, loader ContentLoader, scorer *BM25) (*Index, error) {
	if catalog == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	var idx *Index
	if scorer != nil {
		idx = NewIndexWithScorer(scorer)
	} else {
		idx = NewIndex()
	}
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

// domainTypeVocabulary maps domain types to specialist vocabulary that
// differentiates them in BM25 scoring. These terms are injected during
// indexing to give each domain type a distinct BM25 signature.
//
// This replaces the prior approach of repeating generic domain names 3x,
// which created homogeneous amplification — "architecture" appearing as
// a domain name across every repo made all architecture documents
// collectively dominant in any BM25 result set.
var domainTypeVocabulary = map[string][]string{
	"architecture":       {"structure", "layers", "packages", "components", "modules", "boundaries", "dependencies"},
	"conventions":        {"practices", "standards", "patterns", "idioms", "style", "guidelines", "naming"},
	"scar-tissue":        {"bugs", "failures", "regressions", "incidents", "lessons", "pitfalls", "defenses"},
	"design-constraints": {"constraints", "frozen", "decisions", "tradeoffs", "invariants", "tensions", "boundaries"},
	"test-coverage":      {"tests", "coverage", "gaps", "assertions", "fixtures", "verification", "validation"},
	"feat":               {"feature", "capability", "workflow", "behavior", "interface", "integration"},
	"release":            {"release", "deploy", "version", "changelog", "migration", "upgrade"},
	"literature":         {"research", "scholarship", "evidence", "citations", "methodology", "findings"},
	"radar":              {"opportunities", "signals", "health", "improvements", "recommendations"},
}

// buildBoostedContent prepends metadata fields to content for BM25 indexing.
// G-4: BM25 must index domain names, qualified names, repo names, and section
// headings alongside the summary/body text. Without this, queries referencing
// repo names (like "autom8y-sms") produce zero BM25 signal.
//
// Amplification strategy (contextual-equilibrium redesign):
//   - Repo name and qualified name are boosted 2x (identity signal)
//   - Domain-type-specific vocabulary is injected 2x (differentiating signal)
//   - Section headings are boosted 1x (structural signal)
//   - Generic domain name is included 1x only (not repeated)
//
// This replaces the prior 3x domain-name repetition which created homogeneous
// amplification across architecture documents from different repos.
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

	// Identity signal: repo name and qualified name boosted 2x.
	for i := 0; i < 2; i++ {
		b.WriteString(entry.QualifiedName)
		b.WriteString(" ")
		if repoName != "" {
			b.WriteString(repoName)
			b.WriteString(" ")
		}
	}

	// Domain name: included once only (not repeated to avoid homogeneous amplification).
	b.WriteString(entry.Domain)
	b.WriteString(" ")

	// Domain-type vocabulary: inject differentiating terms 2x.
	if vocab, ok := domainTypeVocabulary[entry.Domain]; ok {
		for i := 0; i < 2; i++ {
			for _, term := range vocab {
				b.WriteString(term)
				b.WriteString(" ")
			}
		}
	}

	// Section headings: structural signal 1x.
	for _, h := range headings {
		b.WriteString(h)
		b.WriteString(" ")
	}

	// Append full content body (no truncation).
	b.WriteString(content)

	return b.String()
}