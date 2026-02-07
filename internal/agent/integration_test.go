package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// projectRoot returns the project root by walking up from this test file
// until we find go.mod.
func projectRoot(t *testing.T) string {
	t.Helper()

	// Get the directory of this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file location")
	}

	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (no go.mod)")
		}
		dir = parent
	}
}

// TestAllRiteAgentSpecs validates that every agent in rites/*/agents/*.md
// parses successfully and passes WARN-mode validation.
func TestAllRiteAgentSpecs(t *testing.T) {
	root := projectRoot(t)
	pattern := filepath.Join(root, "rites", "*", "agents", "*.md")

	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to glob agent files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no agent files found; expected 50+ in rites/*/agents/")
	}

	t.Logf("Found %d rite agent files", len(files))

	av := newTestValidator(t)

	mode := ValidationModeWarn
	if os.Getenv("AGENT_STRICT") == "1" {
		mode = ValidationModeStrict
		t.Log("Running in STRICT mode (AGENT_STRICT=1)")
	}

	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		t.Run(rel, func(t *testing.T) {
			result, err := av.ValidateAgentFile(path, mode)
			if err != nil {
				t.Fatalf("validation error: %v", err)
			}

			if !result.Valid {
				for _, issue := range result.Issues {
					t.Errorf("  ISSUE: [%s] %s", issue.Field, issue.Message)
				}
			}

			for _, w := range result.Warnings {
				t.Logf("  WARN: %s", w)
			}

			if result.Frontmatter != nil {
				if result.Frontmatter.Name == "" {
					t.Error("parsed frontmatter has empty name")
				}
			}
		})
	}
}

// TestAllUserAgentSpecs validates agents/*.md parse correctly.
func TestAllUserAgentSpecs(t *testing.T) {
	root := projectRoot(t)
	pattern := filepath.Join(root, "agents", "*.md")

	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to glob agent files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("no agent files found")
	}

	t.Logf("Found %d agent files", len(files))

	knownBroken := map[string]string{}

	av := newTestValidator(t)

	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		base := filepath.Base(path)
		t.Run(rel, func(t *testing.T) {
			// agents always use WARN mode (they may predate enhanced schema)
			result, err := av.ValidateAgentFile(path, ValidationModeWarn)
			if err != nil {
				t.Fatalf("validation error: %v", err)
			}

			if !result.Valid {
				if reason, ok := knownBroken[base]; ok {
					t.Logf("  KNOWN ISSUE (skipping): %s", reason)
					return
				}
				for _, issue := range result.Issues {
					t.Errorf("  ISSUE: [%s] %s", issue.Field, issue.Message)
				}
			}

			for _, w := range result.Warnings {
				t.Logf("  WARN: %s", w)
			}
		})
	}
}
