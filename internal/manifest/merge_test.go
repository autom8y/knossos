package manifest_test

import (
	"testing"

	"github.com/autom8y/knossos/internal/manifest"
)

func TestMerge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		base         map[string]interface{}
		ours         map[string]interface{}
		theirs       map[string]interface{}
		strategy     manifest.MergeStrategy
		wantConflict bool
		checkResult  func(map[string]interface{}) bool
	}{
		{
			name:         "no changes",
			base:         map[string]interface{}{"version": "1.0"},
			ours:         map[string]interface{}{"version": "1.0"},
			theirs:       map[string]interface{}{"version": "1.0"},
			strategy:     manifest.StrategySmart,
			wantConflict: false,
			checkResult: func(m map[string]interface{}) bool {
				return m["version"] == "1.0"
			},
		},
		{
			name:         "ours only change",
			base:         map[string]interface{}{"version": "1.0"},
			ours:         map[string]interface{}{"version": "1.0", "name": "ours"},
			theirs:       map[string]interface{}{"version": "1.0"},
			strategy:     manifest.StrategySmart,
			wantConflict: false,
			checkResult: func(m map[string]interface{}) bool {
				return m["name"] == "ours"
			},
		},
		{
			name:         "theirs only change",
			base:         map[string]interface{}{"version": "1.0"},
			ours:         map[string]interface{}{"version": "1.0"},
			theirs:       map[string]interface{}{"version": "1.0", "name": "theirs"},
			strategy:     manifest.StrategySmart,
			wantConflict: false,
			checkResult: func(m map[string]interface{}) bool {
				return m["name"] == "theirs"
			},
		},
		{
			name:         "both change same field - conflict",
			base:         map[string]interface{}{"version": "1.0"},
			ours:         map[string]interface{}{"version": "2.0"},
			theirs:       map[string]interface{}{"version": "3.0"},
			strategy:     manifest.StrategySmart,
			wantConflict: true,
		},
		{
			name:         "both change same field - ours strategy",
			base:         map[string]interface{}{"version": "1.0"},
			ours:         map[string]interface{}{"version": "2.0"},
			theirs:       map[string]interface{}{"version": "3.0"},
			strategy:     manifest.StrategyOurs,
			wantConflict: false,
			checkResult: func(m map[string]interface{}) bool {
				return m["version"] == "2.0"
			},
		},
		{
			name:         "both change same field - theirs strategy",
			base:         map[string]interface{}{"version": "1.0"},
			ours:         map[string]interface{}{"version": "2.0"},
			theirs:       map[string]interface{}{"version": "3.0"},
			strategy:     manifest.StrategyTheirs,
			wantConflict: false,
			checkResult: func(m map[string]interface{}) bool {
				return m["version"] == "3.0"
			},
		},
		{
			name: "both add different fields - no conflict",
			base: map[string]interface{}{"version": "1.0"},
			ours: map[string]interface{}{"version": "1.0", "ours_field": "a"},
			theirs: map[string]interface{}{"version": "1.0", "theirs_field": "b"},
			strategy:     manifest.StrategySmart,
			wantConflict: false,
			checkResult: func(m map[string]interface{}) bool {
				return m["ours_field"] == "a" && m["theirs_field"] == "b"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := &manifest.Manifest{Content: tt.base}
			ours := &manifest.Manifest{Content: tt.ours}
			theirs := &manifest.Manifest{Content: tt.theirs}

			result, err := manifest.Merge(base, ours, theirs, manifest.ManifestMergeOptions{Strategy: tt.strategy})
			if err != nil {
				t.Fatalf("Merge() error = %v", err)
			}

			if result.HasConflicts != tt.wantConflict {
				t.Errorf("Merge() HasConflicts = %v, want %v", result.HasConflicts, tt.wantConflict)
			}

			if !tt.wantConflict && tt.checkResult != nil {
				if !tt.checkResult(result.Merged) {
					t.Error("Merge() result content check failed")
				}
			}
		})
	}
}

func TestMergeConflictMarkers(t *testing.T) {
	t.Parallel()
	base := &manifest.Manifest{Content: map[string]interface{}{"version": "1.0"}}
	ours := &manifest.Manifest{Content: map[string]interface{}{"version": "2.0"}}
	theirs := &manifest.Manifest{Content: map[string]interface{}{"version": "3.0"}}

	result, err := manifest.Merge(base, ours, theirs, manifest.ManifestMergeOptions{Strategy: manifest.StrategySmart})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if !result.HasConflicts {
		t.Fatal("expected conflicts")
	}

	// MergedMarkers is generated when conflicts exist
	if result.MergedMarkers == "" {
		t.Error("MergedMarkers is empty for merge with conflicts")
	}
}
