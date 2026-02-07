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

// TestMenaScope_ValidScope verifies all valid and invalid scope values.
func TestMenaScope_ValidScope(t *testing.T) {
	tests := []struct {
		scope MenaScope
		want  bool
	}{
		{MenaScopeBoth, true},
		{MenaScopeUser, true},
		{MenaScopeProject, true},
		{MenaScope("global"), false},
		{MenaScope("both"), false},   // explicit "both" is not valid; omit instead
		{MenaScope("User"), false},   // case-sensitive
		{MenaScope("PROJECT"), false}, // case-sensitive
	}

	for _, tt := range tests {
		name := string(tt.scope)
		if name == "" {
			name = "<empty>"
		}
		t.Run(name, func(t *testing.T) {
			got := tt.scope.ValidScope()
			if got != tt.want {
				t.Errorf("MenaScope(%q).ValidScope() = %v, want %v", string(tt.scope), got, tt.want)
			}
		})
	}
}

// TestMenaScope_String verifies the String() method for all scope values.
func TestMenaScope_String(t *testing.T) {
	tests := []struct {
		scope MenaScope
		want  string
	}{
		{MenaScopeBoth, "both"},
		{MenaScopeUser, "user"},
		{MenaScopeProject, "project"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.scope.String()
			if got != tt.want {
				t.Errorf("MenaScope(%q).String() = %q, want %q", string(tt.scope), got, tt.want)
			}
		})
	}
}

// TestMenaFrontmatter_Validate_Scope verifies scope validation in Validate().
func TestMenaFrontmatter_Validate_Scope(t *testing.T) {
	baseFM := func(scope MenaScope) *MenaFrontmatter {
		return &MenaFrontmatter{
			Name:        "test",
			Description: "test description",
			Scope:       scope,
		}
	}

	// Valid cases
	validScopes := []MenaScope{MenaScopeBoth, MenaScopeUser, MenaScopeProject}
	for _, s := range validScopes {
		fm := baseFM(s)
		if err := fm.Validate(); err != nil {
			t.Errorf("Validate() should pass for scope %q, got error: %v", string(s), err)
		}
	}

	// Invalid cases
	invalidScopes := []MenaScope{
		MenaScope("global"),
		MenaScope("both"), // explicit "both" is invalid
		MenaScope("User"), // case-sensitive
	}
	for _, s := range invalidScopes {
		fm := baseFM(s)
		err := fm.Validate()
		if err == nil {
			t.Errorf("Validate() should fail for scope %q, but passed", string(s))
			continue
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "invalid scope") {
			t.Errorf("Error for scope %q should contain 'invalid scope', got: %s", string(s), errMsg)
		}
		if !strings.Contains(errMsg, string(s)) {
			t.Errorf("Error for scope %q should contain the invalid value, got: %s", string(s), errMsg)
		}
	}
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
