package materialize

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestFrontmatterAllMenaFiles validates frontmatter in all INDEX.dro.md and INDEX.lego.md files.
// This test walks both mena/ and rites/*/mena/ directories to ensure
// all mena files have valid frontmatter according to the unified schema.
func TestFrontmatterAllMenaFiles(t *testing.T) {
	projectRoot := findProjectRoot(t)

	var failures []string
	var checked int

	// Walk mena/ directory (project-level)
	menaDir := filepath.Join(projectRoot, "mena")
	if err := filepath.Walk(menaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Logf("Warning: cannot access %s: %v", path, err)
			return nil
		}

		if info.IsDir() || !isMenaIndex(info.Name()) {
			return nil
		}

		checked++
		if err := validateMenaFile(t, path); err != nil {
			failures = append(failures, path+": "+err.Error())
		}

		return nil
	}); err != nil {
		t.Fatalf("Failed to walk mena/: %v", err)
	}

	// Walk rites/*/mena/ directories
	ritesDir := filepath.Join(projectRoot, "rites")
	riteDirs, err := os.ReadDir(ritesDir)
	if err != nil {
		t.Fatalf("Failed to read rites directory: %v", err)
	}

	for _, riteEntry := range riteDirs {
		if !riteEntry.IsDir() {
			continue
		}

		riteMenaDir := filepath.Join(ritesDir, riteEntry.Name(), "mena")
		if _, err := os.Stat(riteMenaDir); os.IsNotExist(err) {
			continue // No mena directory in this rite
		}

		if err := filepath.Walk(riteMenaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				t.Logf("Warning: cannot access %s: %v", path, err)
				return nil
			}

			if info.IsDir() || !isMenaIndex(info.Name()) {
				return nil
			}

			checked++
			if err := validateMenaFile(t, path); err != nil {
				failures = append(failures, path+": "+err.Error())
			}

			return nil
		}); err != nil {
			t.Fatalf("Failed to walk rite mena in %s: %v", riteEntry.Name(), err)
		}
	}

	// Report results
	if checked == 0 {
		t.Fatal("No INDEX.dro.md or INDEX.lego.md files found - this suggests the test is not finding the correct directories")
	}

	t.Logf("Validated frontmatter in %d mena index files", checked)

	if len(failures) > 0 {
		t.Errorf("Found %d files with invalid frontmatter:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// TestFrontmatterContextField validates that the context field parses correctly from YAML.
func TestFrontmatterContextField(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantCtx  string
		wantName string
	}{
		{
			name: "context fork parses",
			yaml: `name: test-cmd
description: A test command
context: fork`,
			wantCtx:  "fork",
			wantName: "test-cmd",
		},
		{
			name: "no context field",
			yaml: `name: test-cmd
description: A test command`,
			wantCtx:  "",
			wantName: "test-cmd",
		},
		{
			name: "full frontmatter with context",
			yaml: `name: code-review
description: Structured review
model: opus
disable-model-invocation: true
context: fork`,
			wantCtx:  "fork",
			wantName: "code-review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fm MenaFrontmatter
			if err := yaml.Unmarshal([]byte(tt.yaml), &fm); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if fm.Context != tt.wantCtx {
				t.Errorf("Context = %q, want %q", fm.Context, tt.wantCtx)
			}
			if fm.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", fm.Name, tt.wantName)
			}
		})
	}
}

// isMenaIndex returns true if the filename is a mena index file
// (INDEX.dro.md, INDEX.lego.md, or legacy INDEX.md).
func isMenaIndex(name string) bool {
	return name == "INDEX.dro.md" || name == "INDEX.lego.md" || name == "INDEX.md"
}

// validateMenaFile reads a mena index file and validates its frontmatter.
func validateMenaFile(t *testing.T, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Inline frontmatter parsing (ParseMenaFrontmatter was deleted per Sprint 1 Batch 5)
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return err
	}

	// Find closing delimiter
	var endIndex int
	if idx := bytes.Index(content[4:], []byte("\n---\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(content[4:], []byte("\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(content[4:], []byte("\r\n---\r\n")); idx != -1 {
		endIndex = idx
	} else if idx := bytes.Index(content[4:], []byte("\r\n---\n")); idx != -1 {
		endIndex = idx
	} else {
		return err
	}

	frontmatterBytes := content[4 : 4+endIndex]

	var fm MenaFrontmatter
	if err := yaml.Unmarshal(frontmatterBytes, &fm); err != nil {
		return err
	}

	if err := fm.Validate(); err != nil {
		return err
	}

	t.Logf("  %s", path)
	return nil
}

// findProjectRoot walks up from the current directory to find the project root.
// The project root is identified by the presence of go.mod.
func findProjectRoot(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
