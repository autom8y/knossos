package serve

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/search/knowledge"
)

func TestNewServeCmd(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""

	cmd := NewServeCmd(&output, &verbose, &projectDir)

	if cmd.Use != "serve" {
		t.Errorf("expected Use 'serve', got %q", cmd.Use)
	}

	// Verify flags exist with correct defaults.
	// Port and max-concurrent default to 0 (meaning "resolve from hierarchy").
	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("expected --port flag")
	}
	if portFlag.DefValue != "0" {
		t.Errorf("expected port default 0, got %q", portFlag.DefValue)
	}

	secretFlag := cmd.Flags().Lookup("slack-signing-secret")
	if secretFlag == nil {
		t.Fatal("expected --slack-signing-secret flag")
	}

	tokenFlag := cmd.Flags().Lookup("slack-bot-token")
	if tokenFlag == nil {
		t.Fatal("expected --slack-bot-token flag")
	}

	drainFlag := cmd.Flags().Lookup("drain-timeout")
	if drainFlag == nil {
		t.Fatal("expected --drain-timeout flag")
	}
	if drainFlag.DefValue != "0s" {
		t.Errorf("expected drain-timeout default 0s, got %q", drainFlag.DefValue)
	}

	maxConcurrentFlag := cmd.Flags().Lookup("max-concurrent")
	if maxConcurrentFlag == nil {
		t.Fatal("expected --max-concurrent flag")
	}
	if maxConcurrentFlag.DefValue != "0" {
		t.Errorf("expected max-concurrent default 0, got %q", maxConcurrentFlag.DefValue)
	}

	// Verify new flags exist.
	orgFlag := cmd.Flags().Lookup("org")
	if orgFlag == nil {
		t.Fatal("expected --org flag")
	}
	if orgFlag.DefValue != "" {
		t.Errorf("expected org default empty, got %q", orgFlag.DefValue)
	}

	envFileFlag := cmd.Flags().Lookup("env-file")
	if envFileFlag == nil {
		t.Fatal("expected --env-file flag")
	}
	if envFileFlag.DefValue != "" {
		t.Errorf("expected env-file default empty, got %q", envFileFlag.DefValue)
	}
}

func TestNewServeCmd_NeedsProject(t *testing.T) {
	output := "text"
	verbose := false
	projectDir := ""

	cmd := NewServeCmd(&output, &verbose, &projectDir)

	// ari serve should NOT require project context
	val, ok := cmd.Annotations["needsProject"]
	if !ok {
		t.Fatal("expected needsProject annotation to be set")
	}
	if val != "false" {
		t.Errorf("expected needsProject=false, got %q", val)
	}
}

func TestRunServe_MissingSigningSecret(t *testing.T) {
	// Ensure env vars are clean.
	t.Setenv("SLACK_SIGNING_SECRET", "")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("KNOSSOS_ORG", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	output := "text"
	verbose := false
	projectDir := ""

	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &output,
			Verbose:    &verbose,
			ProjectDir: &projectDir,
		},
	}

	opts := serveOptions{}
	err := runServe(ctx, opts)
	if err == nil {
		t.Fatal("expected error for missing signing secret")
	}
}

func TestRunServe_MissingBotToken(t *testing.T) {
	t.Setenv("SLACK_SIGNING_SECRET", "test-secret")
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("KNOSSOS_ORG", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MAX_CONCURRENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("DRAIN_TIMEOUT", "")

	output := "text"
	verbose := false
	projectDir := ""

	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     &output,
			Verbose:    &verbose,
			ProjectDir: &projectDir,
		},
	}

	opts := serveOptions{}
	err := runServe(ctx, opts)
	if err == nil {
		t.Fatal("expected error for missing bot token")
	}
}

// ---- WS-3: SummaryLookup wiring tests ----

// mockCatalog implements knowledge.DomainCatalog for testing.
type mockCatalog struct {
	domains []knowledge.CatalogDomainEntry
}

func (m *mockCatalog) ListDomains() []knowledge.CatalogDomainEntry {
	return m.domains
}

func (m *mockCatalog) LookupDomain(qualifiedName string) (knowledge.CatalogDomainEntry, bool) {
	for _, d := range m.domains {
		if d.QualifiedName == qualifiedName {
			return d, true
		}
	}
	return knowledge.CatalogDomainEntry{}, false
}

func (m *mockCatalog) DomainCount() int {
	return len(m.domains)
}

// mockContentStore implements knowledge.ContentStore for testing.
type mockContentStore struct {
	content map[string]string
}

func (m *mockContentStore) LoadContent(qualifiedName string) (string, error) {
	c, ok := m.content[qualifiedName]
	if !ok {
		return "", nil
	}
	return c, nil
}

func (m *mockContentStore) HasContent(qualifiedName string) bool {
	_, ok := m.content[qualifiedName]
	return ok
}

// mockLLMClient implements knowledge.LLMClient for testing.
type mockLLMClient struct {
	response string
}

func (m *mockLLMClient) Complete(_ context.Context, _, _ string, _ int) (string, error) {
	return m.response, nil
}

func TestSummaryLookup_ReturnsEmptyBeforeIndexBuild(t *testing.T) {
	// WS-3: Before the knowledge index is built, SummaryLookup should
	// return ("", false) -- fail-open pattern.
	var kiPtr atomic.Pointer[knowledge.KnowledgeIndex]

	lookup := func(qualifiedName string) (string, bool) {
		if ki := kiPtr.Load(); ki != nil {
			return ki.GetSummary(qualifiedName)
		}
		return "", false
	}

	summary, ok := lookup("org::repo::architecture")
	if ok {
		t.Error("expected ok=false before index build")
	}
	if summary != "" {
		t.Errorf("expected empty summary before index build, got %q", summary)
	}
}

func TestSummaryLookup_ReturnsSummaryAfterIndexBuild(t *testing.T) {
	// WS-3: After the knowledge index is stored in the atomic pointer,
	// SummaryLookup should return real summaries.
	var kiPtr atomic.Pointer[knowledge.KnowledgeIndex]

	lookup := func(qualifiedName string) (string, bool) {
		if ki := kiPtr.Load(); ki != nil {
			return ki.GetSummary(qualifiedName)
		}
		return "", false
	}

	// Build a KnowledgeIndex with a summary via the Build function.
	catalog := &mockCatalog{
		domains: []knowledge.CatalogDomainEntry{
			{QualifiedName: "org::repo::arch", Domain: "architecture", SourceHash: "h1"},
		},
	}
	contentStore := &mockContentStore{
		content: map[string]string{
			"org::repo::arch": "Architecture content for the project.",
		},
	}
	llmClient := &mockLLMClient{response: "A summary of the architecture domain."}

	idx, err := knowledge.Build(context.Background(), knowledge.BuildConfig{
		Catalog:      catalog,
		ContentStore: contentStore,
		LLMClient:    llmClient,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Before storing: should still return empty.
	summary, ok := lookup("org::repo::arch")
	if ok {
		t.Error("expected ok=false before storing index")
	}
	if summary != "" {
		t.Errorf("expected empty summary before storing index, got %q", summary)
	}

	// Store the index into the atomic pointer (simulates background build completion).
	kiPtr.Store(idx)

	// After storing: should return real summary.
	summary, ok = lookup("org::repo::arch")
	if !ok {
		t.Error("expected ok=true after storing index")
	}
	if summary == "" {
		t.Error("expected non-empty summary after storing index")
	}

	// Missing domain should still return false.
	summary, ok = lookup("org::repo::nonexistent")
	if ok {
		t.Error("expected ok=false for nonexistent domain")
	}
	if summary != "" {
		t.Errorf("expected empty summary for nonexistent domain, got %q", summary)
	}
}

func TestBuildPipeline_SummaryLookupWired(t *testing.T) {
	// WS-3: Verify that buildPipeline() returns a knowledgeIdxPtr that is
	// non-nil and that the SummaryLookup closure is connected.
	// buildPipeline() may return an empty result if ANTHROPIC_API_KEY is
	// not set, but the atomic pointer should still be allocated.
	result := buildPipeline()

	if result.knowledgeIdxPtr == nil {
		t.Fatal("expected knowledgeIdxPtr to be non-nil after buildPipeline()")
	}

	// The pointer should initially be nil (no index loaded yet).
	if result.knowledgeIdxPtr.Load() != nil {
		t.Error("expected knowledgeIdxPtr to be nil initially")
	}
}
