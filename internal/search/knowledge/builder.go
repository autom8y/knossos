package knowledge

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/autom8y/knossos/internal/search/knowledge/embedding"
	"github.com/autom8y/knossos/internal/search/knowledge/graph"
	"github.com/autom8y/knossos/internal/search/knowledge/summary"
)

// embeddingDimensions is the vector dimensionality for Sprint 7 text-to-vector.
// Smaller than production (1536) because we use hash-based vectors, not real embeddings.
const embeddingDimensions = 256

// Build constructs a KnowledgeIndex from the catalog and content store.
//
// Build sequence:
//  1. Load persisted KnowledgeIndex JSON (if PersistedPath is set and file exists)
//  2. For each domain: check NeedsReindex via source_hash comparison
//  3. Match: load from persisted. Mismatch: regenerate (Haiku + embeddings + edges)
//  4. Persist updated JSON
//  5. Post-build validation: all indexed domains have ContentStore content
//
// D-L5: Eager rebuild on source_hash change.
// BC-05: BM25 index is wrapped, not duplicated.
func Build(ctx context.Context, cfg BuildConfig) (*KnowledgeIndex, error) {
	if cfg.Catalog == nil {
		return nil, fmt.Errorf("catalog is required for KnowledgeIndex build")
	}

	domains := cfg.Catalog.ListDomains()
	if len(domains) == 0 {
		slog.Info("knowledge index build: empty catalog")
		return newEmptyIndex(cfg.BM25Index), nil
	}

	buildStart := time.Now()

	// Step 1: Try to load persisted index.
	var persisted *KnowledgeIndex
	if cfg.PersistedPath != "" {
		var err error
		persisted, err = Load(cfg.PersistedPath)
		if err != nil {
			slog.Debug("no persisted knowledge index, building from scratch",
				"path", cfg.PersistedPath, "error", err)
		} else {
			slog.Info("loaded persisted knowledge index",
				"domains", len(persisted.catalog),
				"summaries", persisted.summaries.Count(),
				"embeddings", persisted.embeddings.Count(),
			)
		}
	}

	// Initialize sub-stores from persisted or empty.
	summaryStore := summary.NewStore()
	embeddingStore := embedding.NewStore()
	catalog := make(map[string]*DomainMetadata)

	if persisted != nil {
		summaryStore = persisted.summaries
		embeddingStore = persisted.embeddings
		catalog = persisted.catalog
	}

	// Step 2-3: For each domain, check and regenerate as needed.
	// Uses errgroup with concurrency limit for parallel LLM calls.
	var reindexedCount, skippedCount, failedCount atomic.Int32
	domainInfos := make([]graph.DomainInfo, 0, len(domains))

	// Mutex protects shared, non-thread-safe stores: summaryStore, embeddingStore, catalog.
	var mu sync.Mutex

	// Pre-collect graph infos and identify which domains need reindexing.
	// This read-phase touches the stores before goroutines start, so no mutex needed yet.
	type domainWork struct {
		entry          CatalogDomainEntry
		meta           *DomainMetadata
		needsSummary   bool
		needsEmbedding bool
	}
	var work []domainWork

	for _, d := range domains {
		qn := d.QualifiedName
		currentHash := d.SourceHash

		meta := &DomainMetadata{
			QualifiedName:  qn,
			DomainType:     d.Domain,
			Repo:           repoFromQualifiedName(qn),
			SourceHash:     currentHash,
			GeneratedAt:    d.GeneratedAt,
			FreshnessScore: 0, // BC-12: zero in Tier 1.
		}

		gi := graph.DomainInfo{
			QualifiedName: qn,
			DomainType:    d.Domain,
			Repo:          meta.Repo,
		}
		if existing, ok := catalog[qn]; ok {
			gi.SourceScope = existing.SourceScope
			meta.SourceScope = existing.SourceScope
		}
		domainInfos = append(domainInfos, gi)

		needsSummary := summaryStore.NeedsRegeneration(qn, currentHash)
		needsEmbedding := embeddingStore.NeedsRecompute(qn, currentHash)

		if !needsSummary && !needsEmbedding {
			catalog[qn] = meta
			meta.IndexedAt = time.Now()
			skippedCount.Add(1)
			continue
		}

		work = append(work, domainWork{
			entry:          d,
			meta:           meta,
			needsSummary:   needsSummary,
			needsEmbedding: needsEmbedding,
		})
	}

	// Process domains requiring reindexing in parallel.
	// Concurrency limit of 10 is conservative for Haiku API rate limits;
	// empirical tuning deferred to Sprint 4.
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for _, w := range work {
		w := w // capture loop variable
		g.Go(func() error {
			if gCtx.Err() != nil {
				return nil // Parent context cancelled; stop scheduling.
			}

			qn := w.entry.QualifiedName
			currentHash := w.entry.SourceHash

			// Load content for reindexing.
			if cfg.ContentStore == nil || !cfg.ContentStore.HasContent(qn) {
				slog.Warn("content not available for domain, skipping",
					"domain", qn)
				mu.Lock()
				if _, ok := catalog[qn]; ok {
					skippedCount.Add(1)
				} else {
					failedCount.Add(1)
				}
				mu.Unlock()
				return nil
			}

			content, err := cfg.ContentStore.LoadContent(qn)
			if err != nil {
				slog.Warn("failed to load content for domain",
					"domain", qn, "error", err)
				failedCount.Add(1)
				return nil
			}

			// Regenerate summary.
			if w.needsSummary && cfg.LLMClient != nil {
				sections := parseSections(content)
				llmAdapter := &llmClientAdapter{client: cfg.LLMClient}

				domainCtx, cancel := context.WithTimeout(gCtx, 30*time.Second)
				mu.Lock()
				_, genErr := summaryStore.Generate(domainCtx, qn, content, currentHash, sections, llmAdapter)
				mu.Unlock()
				cancel()

				if genErr != nil {
					slog.Warn("summary generation failed, using stale if available",
						"domain", qn, "error", genErr)
					// Non-fatal: domain still usable via BM25.
				}
			} else if w.needsSummary && cfg.LLMClient == nil {
				slog.Debug("LLM client not available, skipping summary generation",
					"domain", qn)
			}

			// Regenerate embedding from summary text.
			if w.needsEmbedding {
				mu.Lock()
				summaryText, hasSummary := summaryStore.GetSummary(qn)
				mu.Unlock()

				embedText := summaryText
				if !hasSummary {
					embedText = truncateForEmbedding(content, 2000)
				}
				if embedText != "" {
					vec := embedding.TextToVector(embedText, embeddingDimensions)
					if vec != nil {
						mu.Lock()
						embeddingStore.Add(qn, vec, currentHash)
						mu.Unlock()
					}
				}
			}

			mu.Lock()
			catalog[qn] = w.meta
			w.meta.IndexedAt = time.Now()
			mu.Unlock()
			reindexedCount.Add(1)
			return nil // Always nil -- individual domain failures are logged, not propagated.
		})
	}
	_ = g.Wait() // All goroutines return nil; errors are logged per-domain.

	reindexed := int(reindexedCount.Load())
	skipped := int(skippedCount.Load())
	failed := int(failedCount.Load())

	// Step 3b: Build entity graph from metadata (D-L6: deterministic, zero LLM cost).
	entityGraph := graph.Build(domainInfos)

	buildDuration := time.Since(buildStart)
	slog.Info("knowledge index build complete",
		"reindexed", reindexed,
		"skipped", skipped,
		"failed", failed,
		"total_domains", len(domains),
		"summaries", summaryStore.Count(),
		"embeddings", embeddingStore.Count(),
		"edges", entityGraph.EdgeCount(),
		"build_time", buildDuration,
	)

	idx := &KnowledgeIndex{
		bm25:       cfg.BM25Index,
		summaries:  summaryStore,
		embeddings: embeddingStore,
		graph:      entityGraph,
		catalog:    catalog,
	}

	// Step 4: Persist updated index.
	if cfg.PersistedPath != "" {
		if err := Save(cfg.PersistedPath, idx); err != nil {
			slog.Warn("failed to persist knowledge index",
				"path", cfg.PersistedPath, "error", err)
			// Non-fatal: index is usable in memory.
		}
	}

	// Step 5: Post-build validation.
	if cfg.ContentStore != nil {
		for qn := range catalog {
			if !cfg.ContentStore.HasContent(qn) {
				slog.Warn("indexed domain missing from content store",
					"domain", qn)
			}
		}
	}

	return idx, nil
}

// newEmptyIndex creates a KnowledgeIndex with no domains but a wrapped BM25 index.
func newEmptyIndex(bm25 BM25Searcher) *KnowledgeIndex {
	return &KnowledgeIndex{
		bm25:       bm25,
		summaries:  summary.NewStore(),
		embeddings: embedding.NewStore(),
		graph:      graph.New(),
		catalog:    make(map[string]*DomainMetadata),
	}
}

// llmClientAdapter wraps the knowledge.LLMClient interface to satisfy
// the summary.LLMClient interface.
type llmClientAdapter struct {
	client LLMClient
}

func (a *llmClientAdapter) Complete(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error) {
	return a.client.Complete(ctx, systemPrompt, userMessage, maxTokens)
}

// parseSections extracts section headings and bodies from markdown content.
// Uses ## as the section delimiter (matching bm25.SplitSections behavior).
// Returns a map of slug -> section body.
func parseSections(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	var currentHeading string
	var currentBody strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			// Save previous section.
			if currentHeading != "" {
				slug := slugify(currentHeading)
				if slug != "" {
					sections[slug] = currentBody.String()
				}
			}
			currentHeading = strings.TrimPrefix(line, "## ")
			currentBody.Reset()
		} else if currentHeading != "" {
			currentBody.WriteString(line)
			currentBody.WriteString("\n")
		}
	}

	// Save last section.
	if currentHeading != "" {
		slug := slugify(currentHeading)
		if slug != "" {
			sections[slug] = currentBody.String()
		}
	}

	return sections
}

// slugify converts a heading to a URL-friendly slug.
// Matches the behavior of bm25.Slugify without importing the parent package.
func slugify(heading string) string {
	s := strings.ToLower(strings.TrimSpace(heading))
	s = strings.ReplaceAll(s, " ", "-")

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if len(result) > 60 {
		result = result[:60]
	}
	return result
}

// repoFromQualifiedName extracts the repo from "org::repo::domain".
func repoFromQualifiedName(qn string) string {
	parts := strings.SplitN(qn, "::", 3)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// truncateForEmbedding truncates text to maxChars while trying to break at a word boundary.
func truncateForEmbedding(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	// Find last space before limit.
	idx := strings.LastIndex(text[:maxChars], " ")
	if idx > 0 {
		return text[:idx]
	}
	return text[:maxChars]
}
