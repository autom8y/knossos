package provenance

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewKnossosEntry_WithChannel(t *testing.T) {
	t.Parallel()
	entry := NewKnossosEntry(ScopeRite, "rites/eco/agents/foo.md", "project", "sha256:abc123", "gemini")

	if entry.Channel != "gemini" {
		t.Errorf("Channel = %q, want %q", entry.Channel, "gemini")
	}
	if entry.Owner != OwnerKnossos {
		t.Errorf("Owner = %q, want %q", entry.Owner, OwnerKnossos)
	}
}

func TestNewKnossosEntry_ClaudeDefault(t *testing.T) {
	t.Parallel()
	entry := NewKnossosEntry(ScopeRite, "rites/eco/agents/foo.md", "project", "sha256:abc123", "")

	if entry.Channel != "" {
		t.Errorf("Channel = %q, want empty (claude default)", entry.Channel)
	}
}

func TestStructuralEquality_Channel(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	manifestA := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"agents/foo.md": {
				Owner:      OwnerKnossos,
				Scope:      ScopeRite,
				SourcePath: "rites/eco/agents/foo.md",
				SourceType: "project",
				Channel:    "gemini",
				Checksum:   "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				LastSynced: now,
			},
		},
	}

	manifestB := &ProvenanceManifest{
		SchemaVersion: CurrentSchemaVersion,
		LastSync:      now,
		Entries: map[string]*ProvenanceEntry{
			"agents/foo.md": {
				Owner:      OwnerKnossos,
				Scope:      ScopeRite,
				SourcePath: "rites/eco/agents/foo.md",
				SourceType: "project",
				Channel:    "", // different channel
				Checksum:   "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				LastSynced: now,
			},
		},
	}

	if structurallyEqual(manifestA, manifestB) {
		t.Error("manifests with different channels should NOT be structurally equal")
	}

	// Same channel should be equal
	manifestB.Entries["agents/foo.md"].Channel = "gemini"
	if !structurallyEqual(manifestA, manifestB) {
		t.Error("manifests with same channels should be structurally equal")
	}
}

// TestManifestPathForChannel verifies channel-keyed manifest path resolution.
func TestManifestPathForChannel(t *testing.T) {
	t.Parallel()

	knossosDir := "/project/.knossos"

	tests := []struct {
		name    string
		channel string
		want    string
	}{
		{
			name:    "empty channel returns default manifest",
			channel: "",
			want:    filepath.Join(knossosDir, ManifestFileName),
		},
		{
			name:    "claude channel returns default manifest",
			channel: "claude",
			want:    filepath.Join(knossosDir, ManifestFileName),
		},
		{
			name:    "gemini channel returns gemini manifest",
			channel: "gemini",
			want:    filepath.Join(knossosDir, GeminiManifestFileName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ManifestPathForChannel(knossosDir, tt.channel)
			if got != tt.want {
				t.Errorf("ManifestPathForChannel(%q, %q) = %q, want %q", knossosDir, tt.channel, got, tt.want)
			}
		})
	}
}

// TestNewKnossosEntry_W001_ClaudeNormalization verifies that passing "claude" as
// channel is normalized to "" (W-001 fix), matching the newTypedEvent() convention.
func TestNewKnossosEntry_W001_ClaudeNormalization(t *testing.T) {
	t.Parallel()
	entry := NewKnossosEntry(ScopeRite, "rites/eco/agents/foo.md", "project",
		"sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "claude")

	if entry.Channel != "" {
		t.Errorf("Channel = %q, want empty (W-001: 'claude' should normalize to '')", entry.Channel)
	}
	if entry.Owner != OwnerKnossos {
		t.Errorf("Owner = %q, want %q", entry.Owner, OwnerKnossos)
	}
	if entry.Scope != ScopeRite {
		t.Errorf("Scope = %q, want %q", entry.Scope, ScopeRite)
	}
}

// TestGeminiManifestFileName verifies the constant value.
func TestGeminiManifestFileName(t *testing.T) {
	t.Parallel()
	if GeminiManifestFileName != "PROVENANCE_MANIFEST_GEMINI.yaml" {
		t.Errorf("GeminiManifestFileName = %q, want %q", GeminiManifestFileName, "PROVENANCE_MANIFEST_GEMINI.yaml")
	}
}

// TestManifestPathForChannel_BackwardCompat verifies that ManifestPath and
// ManifestPathForChannel produce identical paths for the claude channel.
func TestManifestPathForChannel_BackwardCompat(t *testing.T) {
	t.Parallel()
	knossosDir := "/some/project/.knossos"

	legacy := ManifestPath(knossosDir)
	forEmpty := ManifestPathForChannel(knossosDir, "")
	forClaude := ManifestPathForChannel(knossosDir, "claude")

	if legacy != forEmpty {
		t.Errorf("ManifestPath() = %q, ManifestPathForChannel('') = %q; should be identical", legacy, forEmpty)
	}
	if legacy != forClaude {
		t.Errorf("ManifestPath() = %q, ManifestPathForChannel('claude') = %q; should be identical", legacy, forClaude)
	}
}
