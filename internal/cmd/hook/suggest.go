// Package hook implements the ari hook commands.
package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
	"github.com/autom8y/knossos/internal/suggest"
)

// SuggestPhaseCacheFile is the filename for caching the last-seen phase in a session directory.
// Used by the suggest hook to detect phase transitions between invocations.
const SuggestPhaseCacheFile = ".suggest-phase-cache"

// SuggestOutput represents the output of the suggest hook.
type SuggestOutput struct {
	Detected    bool                   `json:"detected"`
	Transition  string                 `json:"transition,omitempty"` // e.g., "design -> implementation"
	Suggestions []suggest.Suggestion   `json:"suggestions,omitempty"`
	Message     string                 `json:"message,omitempty"`
}

// Text implements output.Textable for text output.
func (s SuggestOutput) Text() string {
	if s.Message != "" {
		return s.Message
	}
	if !s.Detected {
		return "no phase transition detected"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Phase transition: %s\n", s.Transition))
	for _, sg := range s.Suggestions {
		b.WriteString(fmt.Sprintf("- %s\n", sg.Text))
	}
	return b.String()
}

// newSuggestCmd creates the suggest hook subcommand.
func newSuggestCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suggest",
		Short: "Detect phase transitions and surface contextual suggestions",
		Long: `Detects phase transitions in the current session and generates
proactive suggestions when a phase change is found.

This hook is triggered on PostToolUse events (async). It:
- Reads SESSION_CONTEXT.md for the current phase
- Compares against a cached previous phase (.suggest-phase-cache)
- If the phase changed: generates phase-transition suggestions
- If no change: outputs empty (fast path)

The cache file is stored in the session directory and is automatically
cleaned up when the session is removed.

Performance: <60ms total (async, no latency impact on tool execution).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runSuggest(ctx)
			})
		},
	}

	return cmd
}

func runSuggest(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runSuggestCore(ctx, printer)
}

// runSuggestCore contains the suggest hook logic with injected printer for testing.
func runSuggestCore(ctx *cmdContext, printer *output.Printer) error {
	hookEnv := ctx.getHookEnv()

	// Verify this is a PostToolUse event (or empty for testing)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventPostToolUse {
		return printer.Print(SuggestOutput{Message: "not a PostToolUse event"})
	}

	// Resolve session
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil || sessionID == "" || resolver.ProjectRoot() == "" {
		return printer.Print(SuggestOutput{Message: "no active session"})
	}

	// Load session context for current phase
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.VerboseLog("warn", "suggest: failed to load session context",
			map[string]any{"error": err.Error()})
		return printer.Print(SuggestOutput{Message: "session context unreadable"})
	}

	currentPhase := sessCtx.CurrentPhase
	if currentPhase == "" {
		return printer.Print(SuggestOutput{Message: "no phase set"})
	}

	// Read previous phase from cache
	sessionDir := resolver.SessionDir(sessionID)
	cachePath := filepath.Join(sessionDir, SuggestPhaseCacheFile)
	previousPhase := readPhaseCache(cachePath)

	// Write current phase to cache for next invocation (best-effort)
	writePhaseCache(cachePath, currentPhase)

	// If no previous phase (first invocation) or same phase: no transition
	if previousPhase == "" || previousPhase == currentPhase {
		return printer.Print(SuggestOutput{Detected: false})
	}

	// Phase transition detected
	transitionInput := &suggest.PhaseTransitionInput{
		PreviousPhase: previousPhase,
		CurrentPhase:  currentPhase,
		Rite:          sessCtx.ActiveRite,
		Complexity:    sessCtx.Complexity,
	}

	suggestions := suggest.PhaseTransitionSuggestions(transitionInput)

	return printer.Print(SuggestOutput{
		Detected:    true,
		Transition:  fmt.Sprintf("%s -> %s", previousPhase, currentPhase),
		Suggestions: suggestions,
	})
}

// readPhaseCache reads the cached phase from the session directory.
// Returns empty string if the file does not exist or cannot be read.
func readPhaseCache(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// writePhaseCache writes the current phase to the cache file.
// Best-effort: errors are silently ignored.
func writePhaseCache(path, phase string) {
	_ = os.WriteFile(path, []byte(phase), 0644)
}
