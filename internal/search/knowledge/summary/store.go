// Package summary provides Haiku-powered domain and section summarization
// for the KnowledgeIndex.
//
// Summaries are generated at index build time and cached by source_hash.
// When a domain's source_hash is unchanged, the existing summary is reused.
// Cost: ~$0.013/domain (Haiku pricing at ~15K input tokens).
//
// RR-007: This package MUST NOT import internal/search/ or any parent package.
package summary

import (
	"context"
	"fmt"
	"strings"
)

// LLMClient abstracts LLM completion calls for summary generation.
type LLMClient interface {
	Complete(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error)
}

// DomainSummary holds the generated summary for a single domain.
type DomainSummary struct {
	// QualifiedName is the canonical domain address.
	QualifiedName string `json:"qualified_name"`

	// DomainSummary is a 2-3 sentence summary of the entire domain.
	DomainSummary string `json:"domain_summary"`

	// SectionSummaries maps section slug to a 1-sentence summary.
	SectionSummaries map[string]string `json:"section_summaries,omitempty"`

	// SourceHash is the source_hash at the time of generation.
	// Used for cache invalidation.
	SourceHash string `json:"source_hash"`
}

// Store manages domain summaries with source_hash-based caching.
type Store struct {
	summaries map[string]*DomainSummary // qualifiedName -> summary
}

// NewStore creates an empty summary store.
func NewStore() *Store {
	return &Store{
		summaries: make(map[string]*DomainSummary),
	}
}

// NewStoreFromMap creates a store pre-populated with existing summaries.
// Used when loading persisted KnowledgeIndex JSON.
func NewStoreFromMap(summaries map[string]*DomainSummary) *Store {
	if summaries == nil {
		summaries = make(map[string]*DomainSummary)
	}
	return &Store{summaries: summaries}
}

// GetSummary returns the domain summary text for the given qualified name.
// Returns empty string and false if no summary is available.
func (s *Store) GetSummary(qualifiedName string) (string, bool) {
	ds, ok := s.summaries[qualifiedName]
	if !ok || ds == nil {
		return "", false
	}
	return ds.DomainSummary, true
}

// GetDomainSummary returns the full DomainSummary struct.
func (s *Store) GetDomainSummary(qualifiedName string) (*DomainSummary, bool) {
	ds, ok := s.summaries[qualifiedName]
	return ds, ok
}

// NeedsRegeneration returns true if the summary for the given domain is missing
// or has a different source_hash than the current one.
func (s *Store) NeedsRegeneration(qualifiedName, sourceHash string) bool {
	ds, ok := s.summaries[qualifiedName]
	if !ok || ds == nil {
		return true
	}
	return ds.SourceHash != sourceHash
}

// Set stores a pre-built summary.
func (s *Store) Set(summary *DomainSummary) {
	if summary == nil {
		return
	}
	s.summaries[summary.QualifiedName] = summary
}

// GenerateSummary performs the LLM call and returns a *DomainSummary without
// writing it to the store. This allows callers to perform the I/O-heavy LLM
// call outside a mutex and then store the result with Set() under lock.
//
// Returns nil and an error if the LLM call fails. The caller should handle
// degradation (e.g., use stale summary from persisted index).
func (s *Store) GenerateSummary(ctx context.Context, qualifiedName, content, sourceHash string, sections map[string]string, client LLMClient) (*DomainSummary, error) {
	if client == nil {
		return nil, fmt.Errorf("LLM client is nil, cannot generate summary for %s", qualifiedName)
	}

	systemPrompt := `You are a technical documentation summarizer. Your task is to create concise summaries of software knowledge documents.

Rules:
- Domain summary: exactly 2-3 sentences capturing the key purpose and content of this knowledge domain.
- Section summaries: exactly 1 sentence per section, capturing the key information.
- Be specific and technical. Avoid vague language.
- Reference specific technologies, patterns, and decisions mentioned in the content.
- Output format: Start with the domain summary on its own lines, then a blank line, then one line per section in the format "SECTION: slug | summary sentence"`

	// Build user message with content and section headings.
	var userMsg strings.Builder
	userMsg.WriteString("Summarize this knowledge domain:\n\n")
	userMsg.WriteString("Domain: ")
	userMsg.WriteString(qualifiedName)
	userMsg.WriteString("\n\n")

	// Truncate content to ~12K chars to stay within Haiku input limits.
	truncatedContent := content
	if len(truncatedContent) > 12000 {
		truncatedContent = truncatedContent[:12000] + "\n\n[Content truncated]"
	}
	userMsg.WriteString(truncatedContent)

	if len(sections) > 0 {
		userMsg.WriteString("\n\nSections to summarize:\n")
		for slug := range sections {
			userMsg.WriteString("- ")
			userMsg.WriteString(slug)
			userMsg.WriteString("\n")
		}
	}

	resp, err := client.Complete(ctx, systemPrompt, userMsg.String(), 800)
	if err != nil {
		return nil, fmt.Errorf("generate summary for %s: %w", qualifiedName, err)
	}

	// Parse the response into domain summary and section summaries.
	ds := &DomainSummary{
		QualifiedName:    qualifiedName,
		SourceHash:       sourceHash,
		SectionSummaries: make(map[string]string),
	}

	lines := strings.Split(strings.TrimSpace(resp), "\n")
	var domainLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "SECTION:"); ok {
			// Parse "SECTION: slug | summary"
			rest = strings.TrimSpace(rest)
			parts := strings.SplitN(rest, "|", 2)
			if len(parts) == 2 {
				slug := strings.TrimSpace(parts[0])
				summary := strings.TrimSpace(parts[1])
				if slug != "" && summary != "" {
					ds.SectionSummaries[slug] = summary
				}
			}
		} else {
			domainLines = append(domainLines, line)
		}
	}

	ds.DomainSummary = strings.Join(domainLines, " ")

	// Validate: domain summary should be non-empty.
	if ds.DomainSummary == "" {
		// Use the raw response as the summary if parsing failed.
		ds.DomainSummary = strings.TrimSpace(resp)
	}

	return ds, nil
}

// Generate creates a domain summary using the LLM client and stores it.
// Convenience method that calls GenerateSummary() + Set(). For concurrent
// callers that need to hold a mutex only during the store write, use
// GenerateSummary() + Set() separately instead.
//
// The content should be the full .know/ file body (frontmatter stripped).
// sections is a map of slug -> section body text for per-section summarization.
func (s *Store) Generate(ctx context.Context, qualifiedName, content, sourceHash string, sections map[string]string, client LLMClient) (*DomainSummary, error) {
	ds, err := s.GenerateSummary(ctx, qualifiedName, content, sourceHash, sections, client)
	if err != nil {
		return nil, err
	}
	s.Set(ds)
	return ds, nil
}

// All returns all stored summaries. Used for persistence.
func (s *Store) All() map[string]*DomainSummary {
	return s.summaries
}

// Count returns the number of stored summaries.
func (s *Store) Count() int {
	return len(s.summaries)
}
