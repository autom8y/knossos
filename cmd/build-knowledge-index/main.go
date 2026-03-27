// Command build-knowledge-index generates a pre-baked knowledge-index.json for
// Docker image embedding. This eliminates cold-start LLM calls in the container
// by running summary generation at build time.
//
// Usage:
//
//	go run ./cmd/build-knowledge-index \
//	  --catalog deploy/registry/domains.yaml \
//	  --content deploy/content \
//	  --output deploy/knowledge-index.json
//
// Requires ANTHROPIC_API_KEY for Haiku summary generation.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/llm"
	registryorg "github.com/autom8y/knossos/internal/registry/org"
	"github.com/autom8y/knossos/internal/search/knowledge"
)

func main() {
	catalogPath := flag.String("catalog", "deploy/registry/domains.yaml", "Path to domains.yaml catalog")
	contentDir := flag.String("content", "deploy/content", "Path to content directory (deploy/content/)")
	outputPath := flag.String("output", "deploy/knowledge-index.json", "Output path for knowledge-index.json")
	timeout := flag.Duration("timeout", 10*time.Minute, "Build timeout")
	flag.Parse()

	// Configure structured logging.
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	if err := run(*catalogPath, *contentDir, *outputPath, *timeout); err != nil {
		slog.Error("build failed", "error", err)
		os.Exit(1)
	}
}

func run(catalogPath, contentDir, outputPath string, timeout time.Duration) error {
	start := time.Now()

	// Step 1: Load domain catalog.
	catalog, err := registryorg.LoadCatalog(catalogPath)
	if err != nil {
		return fmt.Errorf("load catalog: %w", err)
	}
	slog.Info("catalog loaded",
		"path", catalogPath,
		"domains", catalog.DomainCount(),
		"repos", catalog.RepoCount(),
	)

	// Step 2: Create content store adapter.
	absContentDir, err := filepath.Abs(contentDir)
	if err != nil {
		return fmt.Errorf("resolve content dir: %w", err)
	}
	if info, err := os.Stat(absContentDir); err != nil || !info.IsDir() {
		return fmt.Errorf("content directory not found or not a directory: %s", absContentDir)
	}
	contentStore := &contentAdapter{
		contentDir: absContentDir,
		catalog:    catalog,
	}

	// Step 3: Create LLM client.
	llmClient, err := llm.NewAnthropicClient(llm.DefaultClientConfig())
	if err != nil {
		return fmt.Errorf("create LLM client (is ANTHROPIC_API_KEY set?): %w", err)
	}
	kiLLMClient := &llmAdapter{client: llmClient}

	// Step 4: Create catalog adapter.
	catalogAdapter := &catalogAdapterType{catalog: catalog}

	// Step 5: Resolve absolute output path.
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	// Step 6: Build knowledge index.
	cfg := knowledge.BuildConfig{
		Catalog:       catalogAdapter,
		ContentStore:  contentStore,
		LLMClient:     kiLLMClient,
		PersistedPath: absOutputPath,
		BM25Index:     nil, // BM25 is not needed for pre-bake; it's wired at runtime.
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	slog.Info("building knowledge index...",
		"output", absOutputPath,
		"timeout", timeout,
	)

	idx, err := knowledge.Build(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build knowledge index: %w", err)
	}

	elapsed := time.Since(start)
	slog.Info("knowledge index built successfully",
		"output", absOutputPath,
		"domains", idx.DomainCount(),
		"summaries", idx.SummaryCount(),
		"embeddings", idx.EmbeddingCount(),
		"edges", idx.EdgeCount(),
		"elapsed", elapsed.Round(time.Millisecond),
	)

	// Step 7: Verify the output file was written by Build's persistence step.
	if _, err := os.Stat(absOutputPath); err != nil {
		return fmt.Errorf("output file not found after build (persistence may have failed): %w", err)
	}

	info, _ := os.Stat(absOutputPath)
	slog.Info("output file verified",
		"path", absOutputPath,
		"size_bytes", info.Size(),
	)

	return nil
}

// ---- Adapters ----
// These mirror the adapters in internal/cmd/serve/serve.go but are self-contained
// to avoid pulling in the full serve command dependency chain.

// catalogAdapterType adapts *registryorg.DomainCatalog to knowledge.DomainCatalog.
type catalogAdapterType struct {
	catalog *registryorg.DomainCatalog
}

func (a *catalogAdapterType) ListDomains() []knowledge.CatalogDomainEntry {
	domains := a.catalog.ListDomains()
	out := make([]knowledge.CatalogDomainEntry, len(domains))
	for i, d := range domains {
		out[i] = knowledge.CatalogDomainEntry{
			QualifiedName: d.QualifiedName,
			Domain:        d.Domain,
			Path:          d.Path,
			GeneratedAt:   d.GeneratedAt,
			ExpiresAfter:  d.ExpiresAfter,
			SourceHash:    d.SourceHash,
			Confidence:    d.Confidence,
		}
	}
	return out
}

func (a *catalogAdapterType) LookupDomain(qualifiedName string) (knowledge.CatalogDomainEntry, bool) {
	d, ok := a.catalog.LookupDomain(qualifiedName)
	if !ok {
		return knowledge.CatalogDomainEntry{}, false
	}
	return knowledge.CatalogDomainEntry{
		QualifiedName: d.QualifiedName,
		Domain:        d.Domain,
		Path:          d.Path,
		GeneratedAt:   d.GeneratedAt,
		ExpiresAfter:  d.ExpiresAfter,
		SourceHash:    d.SourceHash,
		Confidence:    d.Confidence,
	}, true
}

func (a *catalogAdapterType) DomainCount() int {
	return a.catalog.DomainCount()
}

// contentAdapter adapts a pre-baked content directory to knowledge.ContentStore.
// Uses the catalog's Path field for reliable scoped domain resolution.
type contentAdapter struct {
	contentDir string
	catalog    *registryorg.DomainCatalog
}

func (a *contentAdapter) LoadContent(qualifiedName string) (string, error) {
	// Strategy 1: Use catalog Path field for precise resolution.
	// The Path field contains the repo-relative path (e.g., ".know/architecture.md"
	// or "services/ads/.know/architecture.md").
	if d, ok := a.catalog.LookupDomain(qualifiedName); ok && d.Path != "" {
		repoName := repoFromQualifiedName(qualifiedName)
		fullPath := filepath.Join(a.contentDir, repoName, d.Path)
		data, err := os.ReadFile(fullPath)
		if err == nil {
			return stripFrontmatter(string(data)), nil
		}
	}

	// Strategy 2: Fallback to derived path (same as serve.go adapter).
	parts := strings.SplitN(qualifiedName, "::", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid qualified name: %s", qualifiedName)
	}
	repoName := know.RepoFromQualifiedName(qualifiedName)
	domainName := parts[2]

	candidates := []string{
		filepath.Join(a.contentDir, repoName, ".know", domainName+".md"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return stripFrontmatter(string(data)), nil
	}

	return "", fmt.Errorf("content not found for %s", qualifiedName)
}

func (a *contentAdapter) HasContent(qualifiedName string) bool {
	content, err := a.LoadContent(qualifiedName)
	return err == nil && content != ""
}

// repoFromQualifiedName extracts the repo name from "org::repo[/scope]::domain".
// Duplicated from know.RepoFromQualifiedName to keep the adapter self-documenting.
func repoFromQualifiedName(qn string) string {
	parts := strings.SplitN(qn, "::", 3)
	if len(parts) < 2 {
		return ""
	}
	repo, _, _ := strings.Cut(parts[1], "/")
	return repo
}

// stripFrontmatter removes YAML frontmatter delimited by ---.
func stripFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---") {
		return text
	}
	rest := text[3:]
	_, after, found := strings.Cut(rest, "---")
	if !found {
		return text
	}
	return strings.TrimSpace(after)
}

// llmAdapter adapts *llm.AnthropicClient to knowledge.LLMClient.
type llmAdapter struct {
	client *llm.AnthropicClient
}

func (a *llmAdapter) Complete(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error) {
	return a.client.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserMessage:  userMessage,
		MaxTokens:    maxTokens,
	})
}
