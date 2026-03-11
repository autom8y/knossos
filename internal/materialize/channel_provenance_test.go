package materialize

import (
	"testing"

	"github.com/autom8y/knossos/internal/provenance"
)

func TestMaterialize_ChannelInProvenance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		channel     string
		wantChannel string
	}{
		{
			name:        "gemini channel recorded in manifest entries",
			channel:     "gemini",
			wantChannel: "gemini",
		},
		{
			name:        "empty channel recorded as empty (claude default)",
			channel:     "",
			wantChannel: "",
		},
		{
			name:        "claude channel recorded as claude",
			channel:     "claude",
			wantChannel: "claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify that NewKnossosEntry correctly propagates channel
			// through the provenance collector (the mechanism used by all
			// rite-scope materialize stages).
			collector := provenance.NewCollector()

			entry := provenance.NewKnossosEntry(
				provenance.ScopeRite,
				"rites/10x-dev/agents/architect.md",
				"project",
				"sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				tt.channel,
			)
			collector.Record("agents/architect.md", entry)

			entries := collector.Entries()
			recorded, ok := entries["agents/architect.md"]
			if !ok {
				t.Fatal("expected entry for agents/architect.md")
			}
			if recorded.Channel != tt.wantChannel {
				t.Errorf("recorded.Channel = %q, want %q", recorded.Channel, tt.wantChannel)
			}
		})
	}
}

func TestMaterialize_ChannelIsolation(t *testing.T) {
	t.Parallel()

	// Verify that rite-scope entries carry channel while user-scope entries
	// remain channel-agnostic (empty), as required by the multi-channel design.
	collector := provenance.NewCollector()

	// Rite-scope entry with gemini channel
	riteEntry := provenance.NewKnossosEntry(
		provenance.ScopeRite,
		"rites/10x-dev/agents/architect.md",
		"project",
		"sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"gemini",
	)
	collector.Record("agents/architect.md", riteEntry)

	// User-scope entry with no channel (channel-agnostic)
	userEntry := provenance.NewKnossosEntry(
		provenance.ScopeUser,
		"agents/moirai.md",
		"embedded",
		"sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		"",
	)
	collector.Record("agents/moirai.md", userEntry)

	entries := collector.Entries()

	riteRecorded := entries["agents/architect.md"]
	if riteRecorded.Channel != "gemini" {
		t.Errorf("rite entry Channel = %q, want %q", riteRecorded.Channel, "gemini")
	}

	userRecorded := entries["agents/moirai.md"]
	if userRecorded.Channel != "" {
		t.Errorf("user entry Channel = %q, want empty (channel-agnostic)", userRecorded.Channel)
	}
}
