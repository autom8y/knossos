package provenance

import (
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
