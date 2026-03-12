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
	ContextFilePath(projectRoot string) string // full path to context file
	SkillsDir(projectRoot string) string       // full path to skills directory
}

type ClaudeChannel struct{}
func (ClaudeChannel) Name() string        { return "claude" }
func (ClaudeChannel) DirName() string     { return ".claude" }
func (ClaudeChannel) ContextFile() string { return "CLAUDE.md" }
func (ClaudeChannel) ContextFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".claude", "CLAUDE.md")
}
func (ClaudeChannel) SkillsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".claude", "skills")
}

type GeminiChannel struct{}
func (GeminiChannel) Name() string        { return "gemini" }
func (GeminiChannel) DirName() string     { return ".gemini" }
func (GeminiChannel) ContextFile() string { return "GEMINI.md" }
func (GeminiChannel) ContextFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".gemini", "GEMINI.md")
}
func (GeminiChannel) SkillsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".gemini", "skills")
}

// AllChannels returns all supported target channels in projection order.
func AllChannels() []TargetChannel {
	return []TargetChannel{ClaudeChannel{}, GeminiChannel{}}
}

// ChannelByName returns the TargetChannel for the given name.
// Empty string defaults to "claude" for backward compatibility (HA-6-001).
// Valid channels are derived from AllChannels(), not hardcoded (HA-6-027).
func ChannelByName(name string) (TargetChannel, error) {
	if name == "" {
		return ClaudeChannel{}, nil // intentional default (HA-3-030)
	}
	for _, ch := range AllChannels() {
		if ch.Name() == name {
			return ch, nil
		}
	}
	return nil, fmt.Errorf("unknown channel: %q", name)
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
