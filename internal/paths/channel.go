package paths

import (
	"fmt"
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
