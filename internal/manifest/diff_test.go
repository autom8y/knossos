package manifest_test

import (
	"testing"

	"github.com/autom8y/knossos/internal/manifest"
)

func TestDiff(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		a         map[string]interface{}
		b         map[string]interface{}
		wantDiff  bool
		additions int
		mods      int
		removals  int
	}{
		{
			name:     "identical manifests",
			a:        map[string]interface{}{"version": "1.0"},
			b:        map[string]interface{}{"version": "1.0"},
			wantDiff: false,
		},
		{
			name:      "added field",
			a:         map[string]interface{}{"version": "1.0"},
			b:         map[string]interface{}{"version": "1.0", "name": "test"},
			wantDiff:  true,
			additions: 1,
		},
		{
			name:     "modified field",
			a:        map[string]interface{}{"version": "1.0"},
			b:        map[string]interface{}{"version": "2.0"},
			wantDiff: true,
			mods:     1,
		},
		{
			name:     "removed field",
			a:        map[string]interface{}{"version": "1.0", "name": "test"},
			b:        map[string]interface{}{"version": "1.0"},
			wantDiff: true,
			removals: 1,
		},
		{
			name: "nested modification",
			a: map[string]interface{}{
				"project": map[string]interface{}{
					"name": "old",
				},
			},
			b: map[string]interface{}{
				"project": map[string]interface{}{
					"name": "new",
				},
			},
			wantDiff: true,
			mods:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mA := &manifest.Manifest{Content: tt.a}
			mB := &manifest.Manifest{Content: tt.b}

			result, err := manifest.Diff(mA, mB, manifest.ManifestDiffOptions{})
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}

			if result.HasChanges != tt.wantDiff {
				t.Errorf("Diff() HasChanges = %v, want %v", result.HasChanges, tt.wantDiff)
			}
			if result.Additions != tt.additions {
				t.Errorf("Diff() Additions = %d, want %d", result.Additions, tt.additions)
			}
			if result.Modifications != tt.mods {
				t.Errorf("Diff() Modifications = %d, want %d", result.Modifications, tt.mods)
			}
			if result.Deletions != tt.removals {
				t.Errorf("Diff() Deletions = %d, want %d", result.Deletions, tt.removals)
			}
		})
	}
}

func TestDiffFormatUnified(t *testing.T) {
	t.Parallel()
	mA := &manifest.Manifest{
		Content: map[string]interface{}{
			"version": "1.0",
			"old":     "value",
		},
	}
	mB := &manifest.Manifest{
		Content: map[string]interface{}{
			"version": "2.0",
			"new":     "value",
		},
	}

	result, err := manifest.Diff(mA, mB, manifest.ManifestDiffOptions{})
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	output := result.FormatUnified()

	// Should contain version modification
	if output == "" {
		t.Error("FormatUnified() returned empty string for manifests with differences")
	}
}
