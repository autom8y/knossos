package materialize

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/autom8y/knossos/internal/materialize/hooks"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSCAR002_StagedMaterializeAbsent is a regression test for DEBT-023 / SCAR-002.
//
// Background: StagedMaterialize renamed .claude/ -> .claude.bak/ during materialization.
// CC's file watcher lost track of its config directory when the rename happened, causing
// hard freezes. The fix was to remove the function entirely (commit 95cf0bc).
//
// This test fails if StagedMaterialize is re-added to the Materializer type, asserting
// the permanent absence of the dangerous rename pattern. Using reflect to check method
// presence rather than a compile-time guard, so that a re-introduction is caught at
// test time with a clear failure message rather than silently compiling.
func TestSCAR002_StagedMaterializeAbsent(t *testing.T) {
	materializerType := reflect.TypeOf((*Materializer)(nil))

	_, exists := materializerType.MethodByName("StagedMaterialize")
	assert.False(t, exists,
		"SCAR-002 regression: StagedMaterialize must not exist on Materializer. "+
			"This method renames .claude/ -> .claude.bak/ which causes CC file watcher freeze. "+
			"Use per-file atomic writes (writeIfChanged) instead. See commit 95cf0bc.")
}

// TestSCAR002_MaterializeWithOptions_NoClaudeRename is a behavioral regression for DEBT-023 / SCAR-002.
//
// Verifies that a full MaterializeWithOptions run does NOT rename or move the .claude/
// directory. After materialization, .claude/ must exist at the original path, not as
// .claude.bak/ or any other renamed form.
//
// This test fails if directory-rename logic is re-introduced into the materialization
// pipeline, regardless of whether it is inside StagedMaterialize or any other path.
func TestSCAR002_MaterializeWithOptions_NoClaudeRename(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	// Setup minimum required directories and files.
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))

	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	// Run materialization.
	result, err := m.MaterializeWithOptions("test-rite", Options{Force: true})
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)

	// .claude/ must exist at original path — not renamed or moved.
	info, err := os.Stat(claudeDir)
	require.NoError(t, err, "SCAR-002 regression: .claude/ must exist after materialization")
	assert.True(t, info.IsDir(), ".claude/ must be a directory")

	// Ensure no .claude.bak/ was created (the renamed form from the old StagedMaterialize).
	backupDir := claudeDir + ".bak"
	_, bakErr := os.Stat(backupDir)
	assert.True(t, os.IsNotExist(bakErr),
		"SCAR-002 regression: .claude.bak/ must not exist after materialization; "+
			"directory rename pattern has been re-introduced")
}

// TestSCAR004_CorruptProvenanceManifest_PropagatesError is a regression test for
// DEBT-024 / SCAR-004.
//
// Background: LoadOrBootstrap and DetectDivergence errors were silently discarded via
// blank identifier in MaterializeMinimal and MaterializeWithOptions. Filesystem permission
// errors and corrupted manifest conditions were completely masked (commits e5655a9, e7277cf).
//
// The fix: all load sites now emit log.Printf WARN on non-file-not-found errors;
// parse/validation errors propagate and abort the pipeline.
//
// This test writes a provenance manifest with corrupt content (invalid YAML that cannot
// be parsed) and asserts that MaterializeWithOptions returns a non-nil error rather than
// silently bootstrapping. The test fails if error handling is reverted to silent discard.
func TestSCAR004_CorruptProvenanceManifest_PropagatesError(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	// Setup minimum required directories and files.
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))

	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Pre-create .claude/ with a corrupt provenance manifest.
	// The content is valid UTF-8 but is invalid YAML (unclosed mapping key),
	// so yaml.Unmarshal will return a parse error. This simulates a truncated
	// or otherwise corrupted manifest file on disk.
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	manifestPath := filepath.Join(claudeDir, provenance.ManifestFileName)
	corruptYAML := []byte("schema_version: \"1.0\"\nlast_sync: !!invalid-tag \"not-a-timestamp\"\n")
	require.NoError(t, os.WriteFile(manifestPath, corruptYAML, 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	// MaterializeWithOptions must return an error when the provenance manifest is corrupt.
	// Silently bootstrapping would mask data corruption and is the anti-pattern this scar guards.
	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true})
	assert.Error(t, err,
		"SCAR-004 regression: MaterializeWithOptions must propagate a non-nil error "+
			"when PROVENANCE_MANIFEST.yaml contains invalid/corrupt data. "+
			"Silent bootstrap (returning nil) defeats the provenance integrity guarantee. "+
			"See commits e5655a9, e7277cf.")
}

// TestSCAR008_BudgetHook_MustNotBeAsync is a regression test for SCAR-008.
//
// Background: The budget hook was accidentally configured with async: true. It fires on
// every PostToolUse event and completes in <5ms. Running it async created an
// "Async hook completed" notification for every single tool call, flooding logs.
// Fix: async: true was removed from the budget hook entry. Sub-100ms hooks must be sync.
//
// This test reads the authoritative hooks.yaml from the repository and asserts that the
// "ari hook budget" entry does NOT have async: true. It fails immediately if async is
// re-introduced, catching the regression at test time rather than in production.
func TestSCAR008_BudgetHook_MustNotBeAsync(t *testing.T) {
	// Resolve the repository root from this file's location.
	// This test file lives at internal/materialize/scar_regression_test.go;
	// the repo root is 2 directories up.
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	hooksYAMLPath := filepath.Join(repoRoot, "config", "hooks.yaml")
	data, err := os.ReadFile(hooksYAMLPath)
	require.NoError(t, err, "config/hooks.yaml must be readable from repo root")

	// Parse via the hooks package (same parser the pipeline uses).
	cfg, parseErr := parseHooksYAMLForSCAR008(data)
	require.NoError(t, parseErr, "config/hooks.yaml must be valid YAML")

	// Scan all entries: budget hook must not be async.
	for _, entry := range cfg {
		if entry.Command == "ari hook budget --output json" {
			assert.False(t, entry.Async,
				"SCAR-008 regression: 'ari hook budget' must NOT have async: true. "+
					"Running the budget hook async creates an 'Async hook completed' "+
					"notification on every tool call. Sub-100ms hooks must be synchronous. "+
					"See commit 85d66d5.")
		}
	}
}

// parseHooksYAMLForSCAR008 parses hooks.yaml content without importing the hooks package
// to avoid a direct dependency on the test needing special test setup.
// It uses the hooks package directly (same package as the production code).
func parseHooksYAMLForSCAR008(data []byte) ([]hooks.HookEntry, error) {
	tmpDir := t_tempDirForSCAR008()
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(configDir, "hooks.yaml"), data, 0644); err != nil {
		return nil, err
	}

	cfg := hooks.LoadHooksConfig(tmpDir)
	if cfg == nil {
		return nil, nil
	}
	return cfg.Hooks, nil
}

// t_tempDirForSCAR008 creates a temporary directory without t.TempDir() since
// this helper is called outside a test function context.
func t_tempDirForSCAR008() string {
	dir, _ := os.MkdirTemp("", "scar008-*")
	return dir
}

// TestSCAR021_CrossRiteAgents_ProjectScopeExclusion is a labeled regression test for SCAR-021.
//
// Background: Global agents (consultant, moirai, context-engineer, theoros) were
// materialized into project .claude/agents/ alongside rite agents, causing shadowing
// issues and orphan accumulation across rite switches. Fix: materializeCrossRiteAgents()
// was removed from the rite-scope pipeline. Cross-rite agents now exclusively use
// user-scope (commit 7ef0213).
//
// This test verifies that cross-rite agents from a top-level agents/ directory are NOT
// materialized into project-level .claude/agents/ during rite-scope sync.
func TestSCAR021_CrossRiteAgents_ProjectScopeExclusion(t *testing.T) {
	projectDir := t.TempDir()
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	claudeDir := filepath.Join(projectDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")

	// Setup a rite with one specialist agent.
	setupRite(t, ritesDir, "test-rite", "", []Agent{{Name: "principal-engineer", Role: "implements code"}})

	// Simulate cross-rite agents in the top-level agents/ directory.
	crossRiteDir := filepath.Join(projectDir, "agents")
	require.NoError(t, os.MkdirAll(crossRiteDir, 0755))
	for _, name := range []string{"moirai", "consultant", "context-engineer", "theoros"} {
		require.NoError(t, os.WriteFile(
			filepath.Join(crossRiteDir, name+".md"),
			[]byte("# "+name+"\n"),
			0644,
		))
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true, KeepAll: true})
	require.NoError(t, err)

	// Rite agent must be present.
	assert.FileExists(t, filepath.Join(agentsDir, "principal-engineer.md"),
		"SCAR-021 regression: rite agent must be materialized to project level")

	// Cross-rite agents must NOT be present at project level.
	for _, name := range []string{"moirai", "consultant", "context-engineer", "theoros"} {
		assert.NoFileExists(t, filepath.Join(agentsDir, name+".md"),
			"SCAR-021 regression: cross-rite agent %q must NOT be materialized to project .claude/agents/. "+
				"Cross-rite agents belong in user scope (~/.claude/agents/). "+
				"See commit 7ef0213.", name)
	}
}

// TestSCAR004_InvalidSchemaProvenanceManifest_PropagatesError is a second variant of the
// SCAR-004 regression test, covering the validation-failure path (valid YAML that parses
// successfully but fails schema validation in validateManifest).
//
// Both the parse-error and validation-error paths must propagate errors — this test
// covers the second path to ensure neither can be silently swallowed independently.
func TestSCAR004_InvalidSchemaProvenanceManifest_PropagatesError(t *testing.T) {
	projectDir := t.TempDir()
	claudeDir := filepath.Join(projectDir, ".claude")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))

	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Pre-create .claude/ with a provenance manifest that parses as valid YAML
	// but fails schema validation: schema_version has an invalid format ("bad-format"
	// does not match the required ^[0-9]+\.[0-9]+$ pattern) and last_sync is zero.
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	manifestPath := filepath.Join(claudeDir, provenance.ManifestFileName)
	invalidSchemaYAML := []byte("schema_version: \"bad-format\"\nentries: {}\n")
	require.NoError(t, os.WriteFile(manifestPath, invalidSchemaYAML, 0644))

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)
	m.templatesDir = templatesDir

	_, err := m.MaterializeWithOptions("test-rite", Options{Force: true})
	assert.Error(t, err,
		"SCAR-004 regression: MaterializeWithOptions must propagate a non-nil error "+
			"when PROVENANCE_MANIFEST.yaml passes YAML parsing but fails schema validation. "+
			"A schema_version of 'bad-format' must not be silently bootstrapped.")
}
