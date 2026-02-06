package manifest_test

import (
	"testing"

	"github.com/autom8y/knossos/internal/manifest"
)

func TestNewSchemaValidator(t *testing.T) {
	v, err := manifest.NewSchemaValidator()
	if err != nil {
		t.Fatalf("NewSchemaValidator() error = %v", err)
	}
	if v == nil {
		t.Fatal("NewSchemaValidator() returned nil")
	}
}

func TestValidate(t *testing.T) {
	v, err := manifest.NewSchemaValidator()
	if err != nil {
		t.Fatalf("NewSchemaValidator() error = %v", err)
	}

	tests := []struct {
		name       string
		content    map[string]interface{}
		schemaName string
		wantValid  bool
	}{
		{
			name:       "valid manifest - has version",
			content:    map[string]interface{}{"version": "1.0"},
			schemaName: manifest.SchemaManifest,
			wantValid:  true,
		},
		{
			name:       "invalid manifest - missing version",
			content:    map[string]interface{}{"project": map[string]interface{}{"name": "test"}},
			schemaName: manifest.SchemaManifest,
			wantValid:  false,
		},
		{
			name: "valid team manifest",
			content: map[string]interface{}{
				"version": "1.0",
				"name":    "test-rite",
				"workflow": map[string]interface{}{
					"type":        "sequential",
					"entry_point": "start",
				},
				"agents": []interface{}{},
			},
			schemaName: manifest.SchemaTeamManifest,
			wantValid:  true,
		},
		{
			name: "invalid team manifest - missing workflow",
			content: map[string]interface{}{
				"version": "1.0",
				"name":    "test-rite",
				"agents":  []interface{}{},
			},
			schemaName: manifest.SchemaTeamManifest,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manifest.Manifest{
				Path:    "/test/path",
				Content: tt.content,
			}

			result, err := v.Validate(m, tt.schemaName, false)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("Validate() Valid = %v, want %v, issues: %v", result.Valid, tt.wantValid, result.Issues)
			}
		})
	}
}

func TestDetectSchemaFromPath(t *testing.T) {
	tests := []struct {
		path       string
		wantSchema string
		wantErr    bool
	}{
		{
			path:       "/project/.claude/manifest.json",
			wantSchema: manifest.SchemaManifest,
		},
		{
			path:       "/project/teams/my-team/manifest.yaml",
			wantSchema: manifest.SchemaTeamManifest,
		},
		{
			path:       "/project/teams/my-team/manifest.yml",
			wantSchema: manifest.SchemaTeamManifest,
		},
		{
			path:    "/random/file.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			schema, err := manifest.DetectSchemaFromPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectSchemaFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && schema != tt.wantSchema {
				t.Errorf("DetectSchemaFromPath() = %q, want %q", schema, tt.wantSchema)
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.0", true},
		{"1.2", true},
		{"10.20", true},
		{"1.0.0", true},
		{"1", false},
		{"1.", false},
		{".1", false},
		{"a.b", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Test through validation
			v, _ := manifest.NewSchemaValidator()
			m := &manifest.Manifest{
				Path:    "/test",
				Content: map[string]interface{}{"version": tt.version},
			}
			result, _ := v.Validate(m, manifest.SchemaManifest, false)

			// Valid version should pass, invalid should fail
			gotValid := result.Valid
			if gotValid != tt.want {
				t.Errorf("version %q: got valid=%v, want valid=%v", tt.version, gotValid, tt.want)
			}
		})
	}
}
