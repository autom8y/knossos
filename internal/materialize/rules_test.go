package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/provenance"

	"github.com/autom8y/knossos/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterializeRules_WipesOldKnossosRules(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	rulesDir := filepath.Join(claudeDir, "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0755))

	// Simulate a rule from a previous rite's templates
	require.NoError(t, os.WriteFile(
		filepath.Join(rulesDir, "internal-session.md"),
		[]byte("old session rule"), 0644))

	// Create templates with a different set of rules
	templatesDir := filepath.Join(projectDir, "templates")
	sourceRulesDir := filepath.Join(templatesDir, "rules")
	require.NoError(t, os.MkdirAll(sourceRulesDir, 0755))
	// internal-session.md still exists in templates (will be updated)
	require.NoError(t, os.WriteFile(
		filepath.Join(sourceRulesDir, "internal-session.md"),
		[]byte("new session rule"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	err := m.materializeRules(claudeDir, nil, provenance.NullCollector{})
	require.NoError(t, err)

	// The old rule should be replaced with new content
	got, err := os.ReadFile(filepath.Join(rulesDir, "internal-session.md"))
	require.NoError(t, err)
	assert.Equal(t, "new session rule", string(got))
}

func TestMaterializeRules_PreservesUserRules(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	rulesDir := filepath.Join(claudeDir, "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0755))

	// User-created rule (not in templates)
	require.NoError(t, os.WriteFile(
		filepath.Join(rulesDir, "my-custom-rule.md"),
		[]byte("user rule content"), 0644))

	// Template rule
	require.NoError(t, os.WriteFile(
		filepath.Join(rulesDir, "internal-agent.md"),
		[]byte("old agent rule"), 0644))

	// Create templates
	templatesDir := filepath.Join(projectDir, "templates")
	sourceRulesDir := filepath.Join(templatesDir, "rules")
	require.NoError(t, os.MkdirAll(sourceRulesDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sourceRulesDir, "internal-agent.md"),
		[]byte("new agent rule"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	err := m.materializeRules(claudeDir, nil, provenance.NullCollector{})
	require.NoError(t, err)

	// User rule should survive
	got, err := os.ReadFile(filepath.Join(rulesDir, "my-custom-rule.md"))
	require.NoError(t, err)
	assert.Equal(t, "user rule content", string(got))

	// Template rule should be updated
	got, err = os.ReadFile(filepath.Join(rulesDir, "internal-agent.md"))
	require.NoError(t, err)
	assert.Equal(t, "new agent rule", string(got))
}

func TestMaterializeRules_RemovesStaleTemplateRule(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	rulesDir := filepath.Join(claudeDir, "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0755))

	// Old template rule that no longer has a template source
	require.NoError(t, os.WriteFile(
		filepath.Join(rulesDir, "mena.md"),
		[]byte("old mena rule"), 0644))

	// Templates directory has mena.md as a known template name
	templatesDir := filepath.Join(projectDir, "templates")
	sourceRulesDir := filepath.Join(templatesDir, "rules")
	require.NoError(t, os.MkdirAll(sourceRulesDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sourceRulesDir, "mena.md"),
		[]byte("updated mena rule"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	err := m.materializeRules(claudeDir, nil, provenance.NullCollector{})
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(rulesDir, "mena.md"))
	require.NoError(t, err)
	assert.Equal(t, "updated mena rule", string(got))
}

func TestMaterializeRules_Idempotent(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	require.NoError(t, os.MkdirAll(filepath.Join(claudeDir, "rules"), 0755))

	templatesDir := filepath.Join(projectDir, "templates")
	sourceRulesDir := filepath.Join(templatesDir, "rules")
	require.NoError(t, os.MkdirAll(sourceRulesDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sourceRulesDir, "internal-session.md"),
		[]byte("session rule"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	// Run twice
	require.NoError(t, m.materializeRules(claudeDir, nil, provenance.NullCollector{}))
	require.NoError(t, m.materializeRules(claudeDir, nil, provenance.NullCollector{}))

	// Same output
	got, err := os.ReadFile(filepath.Join(claudeDir, "rules", "internal-session.md"))
	require.NoError(t, err)
	assert.Equal(t, "session rule", string(got))

	// Only the template rule exists
	entries, err := os.ReadDir(filepath.Join(claudeDir, "rules"))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// TestMaterializeRules_EmbeddedSource verifies that embedded rites produce zero rules
// AND clean up stale knossos-managed rules from a previous filesystem-sourced sync.
// Template rules are knossos-internal development guides (internal/**, rites/**, etc.)
// and are harmful noise on foreign (non-knossos) projects.
func TestMaterializeRules_EmbeddedSource(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	rulesDir := filepath.Join(claudeDir, "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0755))

	// Pre-populate stale knossos-managed rules simulating a prior filesystem sync.
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "internal-agent.md"), []byte("stale rule"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "mena.md"), []byte("stale mena"), 0644))
	// User-created rule that must survive.
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "my-custom.md"), []byte("user rule"), 0644))

	// Set up filesystem templates dir so knownRuleTemplateNames() can resolve stale names.
	templatesDir := filepath.Join(projectDir, "knossos", "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "rules", "internal-agent.md"), []byte("source"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "rules", "mena.md"), []byte("source"), 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	resolved := &ResolvedRite{
		RitePath: "rites/test",
		Source:   RiteSource{Type: SourceEmbedded, Path: "embedded"},
	}

	err := m.materializeRules(claudeDir, resolved, provenance.NullCollector{})
	require.NoError(t, err)

	// Stale knossos-managed rules must be removed.
	_, err = os.Stat(filepath.Join(rulesDir, "internal-agent.md"))
	assert.True(t, os.IsNotExist(err), "stale internal-agent.md must be removed for embedded source")

	_, err = os.Stat(filepath.Join(rulesDir, "mena.md"))
	assert.True(t, os.IsNotExist(err), "stale mena.md must be removed for embedded source")

	// User-created rule must survive.
	got, err := os.ReadFile(filepath.Join(rulesDir, "my-custom.md"))
	require.NoError(t, err)
	assert.Equal(t, "user rule", string(got))
}
