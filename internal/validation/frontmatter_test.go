package validation

import (
	"testing"
)

func TestExtractFrontmatter_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		wantKeys []string
		wantErr  bool
	}{
		{
			name: "simple frontmatter",
			content: `---
title: Test Document
status: draft
---
# Content here`,
			wantKeys: []string{"title", "status"},
			wantErr:  false,
		},
		{
			name: "complex frontmatter",
			content: `---
artifact_id: PRD-test-feature
title: Test Feature
status: approved
created_at: 2025-12-29T20:00:00Z
author: test-user
success_criteria:
  - Criterion 1
  - Criterion 2
---
# PRD Content`,
			wantKeys: []string{"artifact_id", "title", "status", "created_at", "author", "success_criteria"},
			wantErr:  false,
		},
		{
			name: "frontmatter with nested objects",
			content: `---
components:
  - name: Component A
    responsibility: Does things
  - name: Component B
    responsibility: Does other things
coverage:
  functional: 95
  integration: 80
---
# Content`,
			wantKeys: []string{"components", "coverage"},
			wantErr:  false,
		},
		{
			name: "frontmatter with CRLF line endings",
			content: "---\r\ntitle: Test\r\n---\r\n# Content",
			wantKeys: []string{"title"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ExtractFrontmatter([]byte(tt.content))

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("ExtractFrontmatter() returned nil result")
				}

				for _, key := range tt.wantKeys {
					if _, ok := result.Data[key]; !ok {
						t.Errorf("ExtractFrontmatter() missing key %q", key)
					}
				}

				if result.StartLine != 1 {
					t.Errorf("ExtractFrontmatter() StartLine = %d, want 1", result.StartLine)
				}
			}
		})
	}
}

func TestExtractFrontmatter_Invalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		content     string
		errContains string
	}{
		{
			name:        "no opening delimiter",
			content:     "title: Test\n---\n# Content",
			errContains: "missing opening '---'",
		},
		{
			name:        "no closing delimiter",
			content:     "---\ntitle: Test\n# Content",
			errContains: "unclosed frontmatter",
		},
		{
			name:        "empty file",
			content:     "",
			errContains: "empty file",
		},
		{
			name:        "malformed YAML",
			content:     "---\ntitle: [unclosed\n---\n",
			errContains: "invalid YAML",
		},
		{
			name:        "opening delimiter not on first line",
			content:     "\n---\ntitle: Test\n---\n",
			errContains: "missing opening '---'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ExtractFrontmatter([]byte(tt.content))

			if err == nil {
				t.Error("ExtractFrontmatter() expected error, got nil")
				return
			}

			if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
				t.Errorf("ExtractFrontmatter() error = %q, want to contain %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestExtractFrontmatter_EmptyFrontmatter(t *testing.T) {
	t.Parallel()
	// Empty frontmatter between delimiters returns an error
	// because there's no meaningful YAML content
	content := "---\n---\n# Content"
	_, err := ExtractFrontmatter([]byte(content))

	if err == nil {
		t.Error("ExtractFrontmatter() expected error for empty frontmatter, got nil")
	}

	if !contains(err.Error(), "empty frontmatter") {
		t.Errorf("ExtractFrontmatter() error = %q, want to contain 'empty frontmatter'", err.Error())
	}
}

func TestHasFrontmatter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "with LF frontmatter",
			content: "---\ntitle: Test\n---\n",
			want:    true,
		},
		{
			name:    "with CRLF frontmatter",
			content: "---\r\ntitle: Test\r\n---\r\n",
			want:    true,
		},
		{
			name:    "no frontmatter - starts with heading",
			content: "# Title\n\nContent",
			want:    false,
		},
		{
			name:    "no frontmatter - starts with text",
			content: "Some content\n---\n",
			want:    false,
		},
		{
			name:    "too short",
			content: "---",
			want:    false,
		},
		{
			name:    "empty",
			content: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasFrontmatter([]byte(tt.content)); got != tt.want {
				t.Errorf("HasFrontmatter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildFrontmatter(t *testing.T) {
	t.Parallel()
	data := map[string]interface{}{
		"title":  "Test Document",
		"status": "draft",
	}

	result, err := BuildFrontmatter(data)
	if err != nil {
		t.Fatalf("BuildFrontmatter() error = %v", err)
	}

	// Should start with ---
	if result[:4] != "---\n" {
		t.Errorf("BuildFrontmatter() should start with '---\\n', got %q", result[:4])
	}

	// Should end with ---
	if result[len(result)-4:] != "---\n" {
		t.Errorf("BuildFrontmatter() should end with '---\\n', got %q", result[len(result)-4:])
	}

	// Should contain the data
	if !contains(result, "title:") {
		t.Error("BuildFrontmatter() missing 'title' field")
	}
	if !contains(result, "status:") {
		t.Error("BuildFrontmatter() missing 'status' field")
	}
}

func TestExtractFrontmatter_RawYAML(t *testing.T) {
	t.Parallel()
	content := `---
title: Test
author: Alice
---
# Content`

	result, err := ExtractFrontmatter([]byte(content))
	if err != nil {
		t.Fatalf("ExtractFrontmatter() error = %v", err)
	}

	if result.RawYAML == "" {
		t.Error("ExtractFrontmatter() RawYAML should not be empty")
	}

	if !contains(result.RawYAML, "title:") {
		t.Error("ExtractFrontmatter() RawYAML should contain 'title:'")
	}
	if !contains(result.RawYAML, "author:") {
		t.Error("ExtractFrontmatter() RawYAML should contain 'author:'")
	}
}

func TestExtractFrontmatter_LineNumbers(t *testing.T) {
	t.Parallel()
	content := `---
title: Test
author: Alice
status: draft
---
# Content`

	result, err := ExtractFrontmatter([]byte(content))
	if err != nil {
		t.Fatalf("ExtractFrontmatter() error = %v", err)
	}

	if result.StartLine != 1 {
		t.Errorf("ExtractFrontmatter() StartLine = %d, want 1", result.StartLine)
	}

	// EndLine should be line 5 (the closing ---)
	if result.EndLine != 5 {
		t.Errorf("ExtractFrontmatter() EndLine = %d, want 5", result.EndLine)
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
