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
// Valid channels are derived from AllChannels(), not hardcoded (HA-6-027).
// If name is empty, it returns ClaudeChannel for backward compatibility (ADR-0031).
func ChannelByName(name string) (TargetChannel, error) {
	if name == "" {
		return ClaudeChannel{}, nil
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
func UserChannelDir(channel string) (string, error) {
	homeDir, _ := os.UserHomeDir()
	ch, err := ChannelByName(channel)
	if err != nil {
		return "", fmt.Errorf("invalid channel %q, cannot resolve user channel dir: %w", channel, err)
	}
	return filepath.Join(homeDir, ch.DirName()), nil
}
