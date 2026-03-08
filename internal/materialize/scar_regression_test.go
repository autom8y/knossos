package materialize

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
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
	materializerType := reflect.TypeFor[*Materializer]()

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
	knossosDir := filepath.Join(projectDir, ".knossos")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	// Setup minimum required directories and files.
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))

	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Pre-create .knossos/ with a corrupt provenance manifest.
	// The content is valid UTF-8 but is invalid YAML (unclosed mapping key),
	// so yaml.Unmarshal will return a parse error. This simulates a truncated
	// or otherwise corrupted manifest file on disk.
	require.NoError(t, os.MkdirAll(knossosDir, 0755))
	manifestPath := filepath.Join(knossosDir, provenance.ManifestFileName)
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
	cfg, parseErr := parseHooksYAMLForSCAR008(t, data)
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

// parseHooksYAMLForSCAR008 parses hooks.yaml content using the hooks package
// (same parser as the production code).
func parseHooksYAMLForSCAR008(t *testing.T, data []byte) ([]hooks.HookEntry, error) {
	t.Helper()
	tmpDir := t.TempDir()

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

// TestSCAR021_CrossRiteAgents_ProjectScopeExclusion is a labeled regression test for SCAR-021.
//
// Background: Global agents (pythia, moirai, context-engineer, theoros) were
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
	for _, name := range []string{"moirai", "pythia", "context-engineer", "theoros"} {
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
	for _, name := range []string{"moirai", "pythia", "context-engineer", "theoros"} {
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
	knossosDir := filepath.Join(projectDir, ".knossos")
	ritesDir := filepath.Join(projectDir, ".knossos", "rites")
	templatesDir := filepath.Join(projectDir, "templates")

	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "sections"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "rules"), 0755))

	setupRite(t, ritesDir, "test-rite",
		"name: test-workflow\nphases:\n  - build\n",
		[]Agent{{Name: "builder", Role: "builds things"}})

	// Pre-create .knossos/ with a provenance manifest that parses as valid YAML
	// but fails schema validation: schema_version has an invalid format ("bad-format"
	// does not match the required ^[0-9]+\.[0-9]+$ pattern) and last_sync is zero.
	require.NoError(t, os.MkdirAll(knossosDir, 0755))
	manifestPath := filepath.Join(knossosDir, provenance.ManifestFileName)
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

// --- Sprint 3.2: Individual SCAR Regression Tests ---

// repoRootFromThisFile resolves the repository root from this test file's location.
// This file lives at internal/materialize/scar_regression_test.go, so the repo root
// is two directories up.
func repoRootFromThisFile(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestSCAR015_ShellScripts_StderrLogging is a regression test for SCAR-015.
//
// Background: Shell log functions in sync scripts wrote to stdout instead of stderr.
// When called inside data-returning functions, log messages were embedded in manifest
// JSON keys, causing silent data corruption. Fix: all log functions redirect to stderr
// with >&2 (commit c8b551a).
//
// This test scans shell scripts that are part of the sync/materialization pipeline
// for log function definitions that write to stdout instead of stderr. Test/validation
// scripts (e2e-validate.sh, etc.) are excluded because their echo output to stdout
// is intentional user-facing output, not log contamination.
func TestSCAR015_ShellScripts_StderrLogging(t *testing.T) {
	repoRoot := repoRootFromThisFile(t)

	// Only check shell scripts in sync-related paths (the original SCAR domain).
	// e2e-validate.sh and other test scripts intentionally write to stdout.
	syncScriptDirs := []string{
		"rites",
		"scripts/sync",
		"scripts/hooks",
	}

	// Pattern: function definition for log/warn/info/error that could pollute stdout
	logFuncDef := regexp.MustCompile(`(?i)^\s*(?:log|warn|info|error|debug)\s*\(\)\s*\{`)
	stderrRedirect := regexp.MustCompile(`>&2`)

	var violations []string

	for _, dir := range syncScriptDirs {
		dirPath := filepath.Join(repoRoot, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		_ = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".sh") {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()

			relPath, _ := filepath.Rel(repoRoot, path)
			scanner := bufio.NewScanner(file)
			inLogFunc := false
			braceDepth := 0
			lineNum := 0

			for scanner.Scan() {
				lineNum++
				line := scanner.Text()

				// Detect log function definition
				if logFuncDef.MatchString(line) {
					inLogFunc = true
					braceDepth = 1
					continue
				}

				if inLogFunc {
					braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
					if braceDepth <= 0 {
						inLogFunc = false
						continue
					}

					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "#") {
						continue
					}
					// Inside a log function, echo/printf must redirect to stderr
					if (strings.HasPrefix(trimmed, "echo ") || strings.HasPrefix(trimmed, "printf ")) &&
						!stderrRedirect.MatchString(line) {
						violations = append(violations, fmt.Sprintf("%s:%d: %s", relPath, lineNum, trimmed))
					}
				}
			}
			return nil
		})
	}

	assert.Empty(t, violations,
		"SCAR-015 regression: shell log functions in sync scripts must redirect to stderr (>&2). "+
			"Log output mixed with data-returning function stdout corrupts manifest JSON. "+
			"See commit c8b551a.")
}

// TestSCAR016_ShellScripts_NoUnprotectedArithmeticIncrement is a regression test for SCAR-016.
//
// Background: Bash arithmetic expressions ((var++)) return exit code 1 when the value
// before increment is zero. With set -euo pipefail enabled, this causes immediate
// silent script exit. Fix: all patterns use ((var++)) || true (commit 1641792).
//
// This test scans all .sh files for ((var++)) patterns that lack the || true guard.
// It fails if an unprotected arithmetic increment is found in any shell script.
func TestSCAR016_ShellScripts_NoUnprotectedArithmeticIncrement(t *testing.T) {
	repoRoot := repoRootFromThisFile(t)

	// Pattern: ((identifier++)) or ((identifier--)) without || true
	arithmeticIncrement := regexp.MustCompile(`\(\([a-zA-Z_][a-zA-Z0-9_]*(\+\+|--)\)\)`)
	protectedPattern := regexp.MustCompile(`\(\([a-zA-Z_][a-zA-Z0-9_]*(\+\+|--)\)\)\s*\|\|\s*true`)

	var violations []string

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".sh") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(repoRoot, path)
		lines := strings.Split(string(data), "\n")

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if arithmeticIncrement.MatchString(line) && !protectedPattern.MatchString(line) {
				violations = append(violations, relPath+":"+fmt.Sprintf("%d", i+1)+": "+trimmed)
			}
		}
		return nil
	})
	require.NoError(t, err, "failed to walk repo for shell scripts")

	assert.Empty(t, violations,
		"SCAR-016 regression: bash arithmetic increments ((var++)) must use || true guard. "+
			"With set -euo pipefail, ((0++)) returns exit code 1 causing silent script death. "+
			"See commit 1641792.")
}

// TestSCAR018_KnowDromenon_NoContextFork is a regression test for SCAR-018.
//
// Background: The /know dromenon had context: fork set. Forked slash commands run as
// subagents which cannot use the Task tool. The command silently fell back to in-context
// observation instead of dispatching theoros subagents via the Argus Pattern.
// Fix: removed context: fork from /know (commit 4d92db4).
//
// This test verifies the /know dromenon does not have context: fork, guarding the
// specific fix. Additionally, it checks all dromena in the shared rites directory
// for the anti-pattern (Task + context: fork). Other directories may have intentional
// uses of context: fork with Task (e.g., spike with disable-model-invocation: true).
func TestSCAR018_KnowDromenon_NoContextFork(t *testing.T) {
	repoRoot := repoRootFromThisFile(t)

	// Primary assertion: the specific /know dromenon that triggered SCAR-018.
	// /know was relocated from rites/shared/mena/ to platform mena/ (mena placement remediation).
	knowPath := filepath.Join(repoRoot, "mena", "know", "INDEX.dro.md")
	data, err := os.ReadFile(knowPath)
	require.NoError(t, err, "/know dromenon must exist at mena/know/INDEX.dro.md")

	content := string(data)
	parts := strings.SplitN(content, "---", 3)
	require.True(t, len(parts) >= 3, "/know dromenon must have frontmatter")
	frontmatter := parts[1]

	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(trimmed, "context:"); ok {
			value := strings.TrimSpace(after)
			assert.NotEqual(t, "fork", value,
				"SCAR-018 regression: /know dromenon must NOT have context: fork. "+
					"Forked slash commands cannot use Task tool, which /know requires for "+
					"theoros dispatch via the Argus Pattern. See commit 4d92db4.")
		}
	}

	// Secondary assertion: check all platform and shared dromena for the anti-pattern.
	// Platform dromena are the highest-risk location because they apply to all scopes.
	var violations []string
	menaDirs := []string{
		filepath.Join(repoRoot, "mena"),
		filepath.Join(repoRoot, "rites", "shared", "mena"),
	}
	for _, menaDir := range menaDirs {
		_ = filepath.WalkDir(menaDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || d.Name() != "INDEX.dro.md" {
			return nil
		}
		fileData, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		fileContent := string(fileData)
		fileParts := strings.SplitN(fileContent, "---", 3)
		if len(fileParts) < 3 {
			return nil
		}
		fm := fileParts[1]

		hasTask := false
		hasContextFork := false
		hasDisableModelInvocation := false
		for _, fmLine := range strings.Split(fm, "\n") {
			trimmedLine := strings.TrimSpace(fmLine)
			if strings.HasPrefix(trimmedLine, "allowed-tools:") && strings.Contains(trimmedLine, "Task") {
				hasTask = true
			}
			if after, ok := strings.CutPrefix(trimmedLine, "context:"); ok {
				val := strings.TrimSpace(after)
				if val == "fork" {
					hasContextFork = true
				}
			}
			if strings.HasPrefix(trimmedLine, "disable-model-invocation:") && strings.Contains(trimmedLine, "true") {
				hasDisableModelInvocation = true
			}
		}
		// Dromena with disable-model-invocation: true intentionally fork
		// without needing Task for agent dispatch (e.g., /spike).
		if hasTask && hasContextFork && !hasDisableModelInvocation {
			relPath, _ := filepath.Rel(repoRoot, path)
			violations = append(violations, relPath)
		}
		return nil
	})
	}

	assert.Empty(t, violations,
		"SCAR-018 regression: platform/shared dromena with Task in allowed-tools must not have context: fork")
}

// TestSCAR020_SessionDromena_ExplicitSessionIDPassing is a regression test for SCAR-020.
//
// Background: All session lifecycle dromena failed to pass session ID to CLI subprocess
// commands. The SessionStart hook injects session ID into LLM context, but bash
// subprocesses cannot access LLM context. Fix: session dromena updated to
// explicitly instruct the LLM to extract session ID and pass it via -s flag
// (commit 6f35325).
//
// This test reads the session dromena (/fray, /sos) and verifies each contains
// explicit instructions for session ID extraction and passing.
// Updated post-3cf6c15: legacy dromena (continue, park, wrap) removed; sos added.
func TestSCAR020_SessionDromena_ExplicitSessionIDPassing(t *testing.T) {
	repoRoot := repoRootFromThisFile(t)

	// Session dromena that must explicitly instruct session ID passing.
	sessionDromena := []string{
		"mena/session/fray/INDEX.dro.md",
		"mena/session/sos/INDEX.dro.md",
	}

	// Patterns indicating explicit session ID extraction instruction.
	// The dromena must tell the LLM to extract the session ID from context
	// and pass it to CLI subprocesses (typically via -s flag or session_id parameter).
	sessionIDPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)session.?id`),
		regexp.MustCompile(`(?i)extract.*session`),
		regexp.MustCompile(`(?i)-s\s+.*session`),
		regexp.MustCompile(`(?i)session_id=`),
	}

	for _, dromenon := range sessionDromena {
		t.Run(filepath.Base(filepath.Dir(dromenon)), func(t *testing.T) {
			path := filepath.Join(repoRoot, dromenon)
			data, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				t.Skipf("dromenon %s not found (may have been restructured)", dromenon)
				return
			}
			require.NoError(t, err)

			content := string(data)
			found := false
			for _, pat := range sessionIDPatterns {
				if pat.MatchString(content) {
					found = true
					break
				}
			}

			assert.True(t, found,
				"SCAR-020 regression: session dromenon %s must contain explicit session ID "+
					"extraction/passing instructions. Bash subprocesses cannot access LLM context; "+
					"the dromenon must instruct the LLM to extract the session ID from the hook-injected "+
					"Session Context table and pass it to CLI commands. See commit 6f35325.", dromenon)
		})
	}
}

// TestSCAR027_SharedMena_NoSessionArtifacts is a regression test for SCAR-027.
//
// Background: An ARCH-REVIEW session artifact was accidentally added to
// rites/shared/mena/ as a legomenon. Legomena in shared mena are permanent
// platform knowledge, not session artifacts. Fix: file was reverted and
// a lint rule (session-artifact-in-shared-mena) was added.
//
// This test checks that INDEX files (entry points) and non-template files in
// platform mena/ and rites/shared/mena/ do not contain session-specific
// frontmatter markers. Template/schema/example files are excluded because
// they legitimately reference session_id as a field definition.
func TestSCAR027_SharedMena_NoSessionArtifacts(t *testing.T) {
	repoRoot := repoRootFromThisFile(t)

	// Session artifact frontmatter markers.
	sessionMarkers := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^session[_-]?id\s*:`),
		regexp.MustCompile(`(?i)^audit[_-]?id\s*:`),
		regexp.MustCompile(`(?i)^sprint[_-]?id\s*:`),
		regexp.MustCompile(`(?i)^throughline\s*:`),
		regexp.MustCompile(`(?i)^session[_-]?ref\s*:`),
		regexp.MustCompile(`(?i)^sprint[_-]?session\s*:`),
	}

	// Directories that contain template schemas or examples are allowed to
	// reference session_id as a field definition (not actual session data).
	excludeDirs := map[string]bool{
		"examples":         true,
		"schemas":          true,
		"shared-templates": true,
	}

	var violations []string

	// Walk both platform mena and shared rite mena.
	menaDirs := []string{
		filepath.Join(repoRoot, "mena"),
		filepath.Join(repoRoot, "rites", "shared", "mena"),
	}
	for _, menaDir := range menaDirs {
		if _, statErr := os.Stat(menaDir); os.IsNotExist(statErr) {
			continue
		}
		err := filepath.WalkDir(menaDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if excludeDirs[d.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}

			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}

			relPath, _ := filepath.Rel(repoRoot, path)
			content := string(data)

			// Only check frontmatter (between --- delimiters)
			parts := strings.SplitN(content, "---", 3)
			if len(parts) < 3 {
				return nil
			}
			frontmatter := parts[1]

			for _, line := range strings.Split(frontmatter, "\n") {
				trimmed := strings.TrimSpace(line)
				for _, pat := range sessionMarkers {
					if pat.MatchString(trimmed) {
						violations = append(violations, relPath+": "+trimmed)
					}
				}
			}
			return nil
		})
		require.NoError(t, err, "failed to walk mena directory: "+menaDir)
	}

	assert.Empty(t, violations,
		"SCAR-027 regression: mena entry files must not contain session artifacts. "+
			"Mena is permanent platform knowledge. Session artifacts belong in .sos/wip/. "+
			"See MEMORY.md Golden Rule and commit f0971e4.")
}
