package mena

import "testing"

// TestTransformMenaFilePath verifies INDEX.md promotion and SKILL.md rename
// logic for all combinations of mena type, filename, and directory depth.
func TestTransformMenaFilePath(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		dir          string
		isDro        bool
		wantNewBase  string
		wantPromoted bool
	}{
		{
			name:         "dro INDEX.md at root is promoted",
			base:         "INDEX.md",
			dir:          ".",
			isDro:        true,
			wantNewBase:  "INDEX.md",
			wantPromoted: true,
		},
		{
			name:         "lego INDEX.md at root renamed to SKILL.md",
			base:         "INDEX.md",
			dir:          ".",
			isDro:        false,
			wantNewBase:  "SKILL.md",
			wantPromoted: false,
		},
		{
			name:         "dro INDEX.md in subdir unchanged",
			base:         "INDEX.md",
			dir:          "schemas",
			isDro:        true,
			wantNewBase:  "INDEX.md",
			wantPromoted: false,
		},
		{
			name:         "lego INDEX.md in subdir unchanged",
			base:         "INDEX.md",
			dir:          "schemas",
			isDro:        false,
			wantNewBase:  "INDEX.md",
			wantPromoted: false,
		},
		{
			name:         "dro helper.md at root unchanged",
			base:         "helper.md",
			dir:          ".",
			isDro:        true,
			wantNewBase:  "helper.md",
			wantPromoted: false,
		},
		{
			name:         "lego helper.md at root unchanged",
			base:         "helper.md",
			dir:          ".",
			isDro:        false,
			wantNewBase:  "helper.md",
			wantPromoted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotPromoted := TransformMenaFilePath(tt.base, tt.dir, tt.isDro)
			if gotBase != tt.wantNewBase {
				t.Errorf("TransformMenaFilePath(%q, %q, %v) newBase = %q, want %q",
					tt.base, tt.dir, tt.isDro, gotBase, tt.wantNewBase)
			}
			if gotPromoted != tt.wantPromoted {
				t.Errorf("TransformMenaFilePath(%q, %q, %v) promoted = %v, want %v",
					tt.base, tt.dir, tt.isDro, gotPromoted, tt.wantPromoted)
			}
		})
	}
}
