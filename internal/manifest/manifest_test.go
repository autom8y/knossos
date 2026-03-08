package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/manifest"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		format    string
		wantErr   bool
		checkFunc func(*manifest.Manifest) bool
	}{
		{
			name:    "valid JSON manifest",
			content: `{"version": "1.0", "project": {"name": "test"}}`,
			format:  "json",
			wantErr: false,
			checkFunc: func(m *manifest.Manifest) bool {
				return m.Content["version"] == "1.0"
			},
		},
		{
			name: "valid YAML manifest",
			content: `version: "1.0"
project:
  name: test`,
			format:  "yaml",
			wantErr: false,
			checkFunc: func(m *manifest.Manifest) bool {
				return m.Content["version"] == "1.0"
			},
		},
		{
			name:    "invalid JSON",
			content: `{"version": }`,
			format:  "json",
			wantErr: true,
		},
		{
			name:    "invalid YAML",
			content: `version: "1.0"\n  bad: indent`,
			format:  "yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			ext := ".json"
			if tt.format == "yaml" {
				ext = ".yaml"
			}
			tmpFile := filepath.Join(t.TempDir(), "manifest"+ext)
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			m, err := manifest.Load(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil && !tt.checkFunc(m) {
				t.Error("Load() manifest content check failed")
			}
		})
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := manifest.Load("/nonexistent/path/manifest.json")
	if err == nil {
		t.Error("Load() expected error for nonexistent file")
	}
}

func TestManifestClone(t *testing.T) {
	original := &manifest.Manifest{
		Path:   "/test/path",
		Format: manifest.FormatJSON,
		Content: map[string]any{
			"version": "1.0",
			"project": map[string]any{
				"name": "test",
			},
		},
	}

	clone := original.Clone()

	// Check it's a deep copy
	if clone.Path != original.Path {
		t.Error("Clone() path mismatch")
	}

	// Modify clone and check original is unchanged
	clone.Content["version"] = "2.0"
	if original.Content["version"] != "1.0" {
		t.Error("Clone() did not create deep copy")
	}
}

func TestManifestToJSON(t *testing.T) {
	m := &manifest.Manifest{
		Content: map[string]any{
			"version": "1.0",
		},
	}

	jsonBytes, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// ToJSON uses json.Marshal (no indent), so expect compact JSON
	expected := `{"version":"1.0"}`
	if string(jsonBytes) != expected {
		t.Errorf("ToJSON() = %q, want %q", string(jsonBytes), expected)
	}
}

// --- DEBT-178: Rite Manifest Validation Tests ---

func TestValidateRiteManifest_Valid(t *testing.T) {
	m := &manifest.Manifest{
		Path: "/test/rites/test-rite/manifest.yaml",
		Content: map[string]any{
			"name":        "test-rite",
			"entry_agent": "potnia",
			"agents": []any{
				map[string]any{"name": "potnia", "role": "orchestrator"},
				map[string]any{"name": "builder", "role": "builds things"},
			},
		},
	}
	warnings := manifest.ValidateRiteManifest(m)
	if len(warnings) != 0 {
		t.Errorf("ValidateRiteManifest() returned %d warnings for valid manifest, want 0: %v", len(warnings), warnings)
	}
}

func TestValidateRiteManifest_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name         string
		content      map[string]any
		wantWarnings int
		wantPaths    []string
	}{
		{
			name: "missing name",
			content: map[string]any{
				"entry_agent": "potnia",
				"agents":      []any{map[string]any{"name": "potnia"}},
			},
			wantWarnings: 1,
			wantPaths:    []string{"$.name"},
		},
		{
			name: "missing entry_agent",
			content: map[string]any{
				"name":   "test-rite",
				"agents": []any{map[string]any{"name": "builder"}},
			},
			wantWarnings: 1,
			wantPaths:    []string{"$.entry_agent"},
		},
		{
			name: "empty name and entry_agent",
			content: map[string]any{
				"name":        "",
				"entry_agent": "",
				"agents":      []any{map[string]any{"name": "builder"}},
			},
			wantWarnings: 2,
			wantPaths:    []string{"$.name", "$.entry_agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manifest.Manifest{Path: "/test", Content: tt.content}
			warnings := manifest.ValidateRiteManifest(m)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("got %d warnings, want %d: %v", len(warnings), tt.wantWarnings, warnings)
			}
			for i, wantPath := range tt.wantPaths {
				if i < len(warnings) && warnings[i].Path != wantPath {
					t.Errorf("warning[%d].Path = %q, want %q", i, warnings[i].Path, wantPath)
				}
			}
			// All warnings must have severity "warning" (not "error") per TD-3
			for _, w := range warnings {
				if w.Severity != "warning" {
					t.Errorf("severity = %q, want %q (TD-3: warnings only)", w.Severity, "warning")
				}
			}
		})
	}
}

func TestValidateRiteManifest_EntryAgentNotInList(t *testing.T) {
	m := &manifest.Manifest{
		Path: "/test",
		Content: map[string]any{
			"name":        "test-rite",
			"entry_agent": "nonexistent-agent",
			"agents": []any{
				map[string]any{"name": "builder", "role": "builds things"},
			},
		},
	}
	warnings := manifest.ValidateRiteManifest(m)
	found := false
	for _, w := range warnings {
		if w.Path == "$.entry_agent" && w.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for entry_agent not found in agents list")
	}
}

func TestValidateRiteManifest_AgentMissingName(t *testing.T) {
	m := &manifest.Manifest{
		Path: "/test",
		Content: map[string]any{
			"name":        "test-rite",
			"entry_agent": "builder",
			"agents": []any{
				map[string]any{"name": "builder", "role": "builds things"},
				map[string]any{"role": "unnamed agent"},
			},
		},
	}
	warnings := manifest.ValidateRiteManifest(m)
	found := false
	for _, w := range warnings {
		if w.Path == "$.agents[1].name" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for agent[1] missing name field")
	}
}

func TestValidateRiteManifest_NilManifest(t *testing.T) {
	warnings := manifest.ValidateRiteManifest(nil)
	if warnings != nil {
		t.Errorf("expected nil for nil manifest, got %v", warnings)
	}
}

func TestLoadFormat(t *testing.T) {
	tests := []struct {
		name    string
		ext     string
		content string
		want    manifest.Format
	}{
		{
			name:    "JSON format",
			ext:     ".json",
			content: `{"version": "1.0"}`,
			want:    manifest.FormatJSON,
		},
		{
			name:    "YAML format",
			ext:     ".yaml",
			content: "version: \"1.0\"\n",
			want:    manifest.FormatYAML,
		},
		{
			name:    "YML format",
			ext:     ".yml",
			content: "version: \"1.0\"\n",
			want:    manifest.FormatYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "manifest"+tt.ext)
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			m, err := manifest.Load(tmpFile)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if m.Format != tt.want {
				t.Errorf("Load() format = %v, want %v", m.Format, tt.want)
			}
		})
	}
}
