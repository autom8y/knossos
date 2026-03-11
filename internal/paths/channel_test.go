package paths_test

import (
	"testing"

	"github.com/autom8y/knossos/internal/paths"
)

func TestChannelByName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantName    string
		wantDir     string
		wantContext string
		wantErr     bool
	}{
		{"empty defaults to claude", "", "claude", ".claude", "CLAUDE.md", false},
		{"explicit claude", "claude", "claude", ".claude", "CLAUDE.md", false},
		{"gemini", "gemini", "gemini", ".gemini", "GEMINI.md", false},
		{"unknown", "foo", "", "", "", true},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable for parallel tests
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ch, err := paths.ChannelByName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ChannelByName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if ch.Name() != tt.wantName {
					t.Errorf("Name() = %v, want %v", ch.Name(), tt.wantName)
				}
				if ch.DirName() != tt.wantDir {
					t.Errorf("DirName() = %v, want %v", ch.DirName(), tt.wantDir)
				}
				if ch.ContextFile() != tt.wantContext {
					t.Errorf("ContextFile() = %v, want %v", ch.ContextFile(), tt.wantContext)
				}
			}
		})
	}
}

func TestResolver_ChannelDir(t *testing.T) {
	t.Parallel()
	r := paths.NewResolver("/fake/root")

	claudePath := r.ChannelDir(paths.ClaudeChannel{})
	if claudePath != "/fake/root/.claude" {
		t.Errorf("expected /fake/root/.claude, got %s", claudePath)
	}

	geminiPath := r.ChannelDir(paths.GeminiChannel{})
	if geminiPath != "/fake/root/.gemini" {
		t.Errorf("expected /fake/root/.gemini, got %s", geminiPath)
	}
}

func TestClaudeChannel_BackwardCompat(t *testing.T) {
	t.Parallel()
	r := paths.NewResolver("/fake/root")
	
	claudeDir := r.ClaudeDir()
	channelDir := r.ChannelDir(paths.ClaudeChannel{})
	
	if claudeDir != channelDir {
		t.Errorf("ClaudeDir() %s != ChannelDir(ClaudeChannel{}) %s", claudeDir, channelDir)
	}
}
