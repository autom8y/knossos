package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

type TargetChannel interface {
	Name() string        // "claude" or "gemini"
	DirName() string     // ".claude" or ".gemini"
	ContextFile() string // "CLAUDE.md" or "GEMINI.md"
}

type ClaudeChannel struct{}
func (ClaudeChannel) Name() string        { return "claude" }
func (ClaudeChannel) DirName() string     { return ".claude" }
func (ClaudeChannel) ContextFile() string { return "CLAUDE.md" }

type GeminiChannel struct{}
func (GeminiChannel) Name() string        { return "gemini" }
func (GeminiChannel) DirName() string     { return ".gemini" }
func (GeminiChannel) ContextFile() string { return "GEMINI.md" }

// AllChannels returns all supported target channels in projection order.
func AllChannels() []TargetChannel {
	return []TargetChannel{ClaudeChannel{}, GeminiChannel{}}
}

func ChannelByName(name string) (TargetChannel, error) {
	switch name {
	case "claude", "":
		return ClaudeChannel{}, nil
	case "gemini":
		return GeminiChannel{}, nil
	default:
		return nil, fmt.Errorf("unknown channel: %q", name)
	}
}

func (r *Resolver) ChannelDir(ch TargetChannel) string {
	return filepath.Join(r.projectRoot, ch.DirName())
}

// UserChannelDir returns the user-level directory for a specific channel.
// For channel="" or "claude": ~/.claude
// For channel="gemini": ~/.gemini
//
// On unrecognized channel names, falls back to ~/.claude intentionally.
// All callers (ForChannel path helpers, user_scope, org_scope) pass validated
// channel strings or empty string (which ChannelByName normalizes to "claude").
// Returning an error here would require error handling in ~10 callers that
// construct paths -- the fallback is the pragmatic choice (HA-3-030 scope).
func UserChannelDir(channel string) string {
	homeDir, _ := os.UserHomeDir()
	ch, err := ChannelByName(channel)
	if err != nil {
		return filepath.Join(homeDir, ".claude")
	}
	return filepath.Join(homeDir, ch.DirName())
}
