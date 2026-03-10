package mena

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestFrontmatterAllMenaFiles validates frontmatter in all INDEX.dro.md and INDEX.lego.md files.
func TestFrontmatterAllMenaFiles(t *testing.T) {
	t.Parallel()
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
			continue
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

	if checked == 0 {
		t.Fatal("No INDEX.dro.md or INDEX.lego.md files found")
	}

	t.Logf("Validated frontmatter in %d mena index files", checked)

	if len(failures) > 0 {
		t.Errorf("Found %d files with invalid frontmatter:\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

// TestFrontmatterContextField validates that the context field parses correctly from YAML.
func TestFrontmatterContextField(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		yaml     string
		wantCtx  string
		wantName string
	}{
		{
			name:     "context fork parses",
			yaml:     "name: test-cmd\ndescription: A test command\ncontext: fork",
			wantCtx:  "fork",
			wantName: "test-cmd",
		},
		{
			name:     "no context field",
			yaml:     "name: test-cmd\ndescription: A test command",
			wantCtx:  "",
			wantName: "test-cmd",
		},
		{
			name:     "full frontmatter with context",
			yaml:     "name: code-review\ndescription: Structured review\nmodel: opus\ndisable-model-invocation: true\ncontext: fork",
			wantCtx:  "fork",
			wantName: "code-review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

func isMenaIndex(name string) bool {
	return name == "INDEX.dro.md" || name == "INDEX.lego.md" || name == "INDEX.md"
}

func validateMenaFile(t *testing.T, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return err
	}

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
