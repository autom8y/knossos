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

func TestAllChannels(t *testing.T) {
	t.Parallel()

	channels := paths.AllChannels()
	if len(channels) != 2 {
		t.Fatalf("AllChannels() returned %d channels, want 2", len(channels))
	}

	if channels[0].Name() != "claude" {
		t.Errorf("AllChannels()[0].Name() = %q, want %q", channels[0].Name(), "claude")
	}
	if channels[1].Name() != "gemini" {
		t.Errorf("AllChannels()[1].Name() = %q, want %q", channels[1].Name(), "gemini")
	}
}

func TestContextFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		channel  paths.TargetChannel
		root     string
		wantPath string
	}{
		{"claude", paths.ClaudeChannel{}, "/project", "/project/.claude/CLAUDE.md"},
		{"gemini", paths.GeminiChannel{}, "/project", "/project/.gemini/GEMINI.md"},
		{"claude nested root", paths.ClaudeChannel{}, "/home/user/code/app", "/home/user/code/app/.claude/CLAUDE.md"},
		{"gemini nested root", paths.GeminiChannel{}, "/home/user/code/app", "/home/user/code/app/.gemini/GEMINI.md"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.channel.ContextFilePath(tt.root)
			if got != tt.wantPath {
				t.Errorf("ContextFilePath(%q) = %q, want %q", tt.root, got, tt.wantPath)
			}
		})
	}
}

func TestSkillsDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		channel paths.TargetChannel
		root    string
		wantDir string
	}{
		{"claude", paths.ClaudeChannel{}, "/project", "/project/.claude/skills"},
		{"gemini", paths.GeminiChannel{}, "/project", "/project/.gemini/skills"},
		{"claude nested root", paths.ClaudeChannel{}, "/home/user/code/app", "/home/user/code/app/.claude/skills"},
		{"gemini nested root", paths.GeminiChannel{}, "/home/user/code/app", "/home/user/code/app/.gemini/skills"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.channel.SkillsDir(tt.root)
			if got != tt.wantDir {
				t.Errorf("SkillsDir(%q) = %q, want %q", tt.root, got, tt.wantDir)
			}
		})
	}
}

// TestContextFilePath_ConsistentWithDirNameContextFile verifies that for CC and Gemini
// channels, ContextFilePath produces the same result as manually joining DirName + ContextFile.
// This will NOT hold for future channels like Codex (AGENTS.md at repo root).
func TestContextFilePath_ConsistentWithDirNameContextFile(t *testing.T) {
	t.Parallel()

	root := "/fake/root"
	for _, ch := range paths.AllChannels() {
		expected := root + "/" + ch.DirName() + "/" + ch.ContextFile()
		got := ch.ContextFilePath(root)
		if got != expected {
			t.Errorf("%s: ContextFilePath(%q) = %q, want %q (DirName/ContextFile)", ch.Name(), root, got, expected)
		}
	}
}
