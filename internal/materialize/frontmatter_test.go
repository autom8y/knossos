package materialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFrontmatterAllIndexFiles validates frontmatter in all INDEX.md command files.
// This test walks both user-commands/ and rites/*/commands/ directories to ensure
// all command files have valid frontmatter according to the unified schema.
func TestFrontmatterAllIndexFiles(t *testing.T) {
	projectRoot := findProjectRoot(t)

	var failures []string
	var checked int

	// Walk user-commands directory
	userCommandsDir := filepath.Join(projectRoot, "user-commands")
	if err := filepath.Walk(userCommandsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Don't fail on directory access errors, just skip
			t.Logf("Warning: cannot access %s: %v", path, err)
			return nil
		}

		if info.IsDir() || info.Name() != "INDEX.md" {
			return nil
		}

		checked++
		if err := validateIndexFile(t, path); err != nil {
			failures = append(failures, path+": "+err.Error())
		}

		return nil
	}); err != nil {
		t.Fatalf("Failed to walk user-commands: %v", err)
	}

	// Walk rites/*/commands directories
	ritesDir := filepath.Join(projectRoot, "rites")
	riteDirs, err := os.ReadDir(ritesDir)
	if err != nil {
		t.Fatalf("Failed to read rites directory: %v", err)
	}

	for _, riteEntry := range riteDirs {
		if !riteEntry.IsDir() {
			continue
		}

		commandsDir := filepath.Join(ritesDir, riteEntry.Name(), "commands")
		if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
			continue // No commands directory in this rite
		}

		if err := filepath.Walk(commandsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				t.Logf("Warning: cannot access %s: %v", path, err)
				return nil
			}

			if info.IsDir() || info.Name() != "INDEX.md" {
				return nil
			}

			checked++
			if err := validateIndexFile(t, path); err != nil {
				failures = append(failures, path+": "+err.Error())
			}

			return nil
		}); err != nil {
			t.Fatalf("Failed to walk rite commands in %s: %v", riteEntry.Name(), err)
		}
	}

	// Report results
	if checked == 0 {
		t.Fatal("No INDEX.md files found - this suggests the test is not finding the correct directories")
	}

	t.Logf("Validated frontmatter in %d INDEX.md files", checked)

	if len(failures) > 0 {
		t.Errorf("Found %d files with invalid frontmatter:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// validateIndexFile reads an INDEX.md file and validates its frontmatter.
func validateIndexFile(t *testing.T, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fm, err := ParseCommandFrontmatter(content)
	if err != nil {
		return err
	}

	if err := fm.Validate(); err != nil {
		return err
	}

	t.Logf("✓ %s", path)
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
