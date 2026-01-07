// Package hook implements the ari hook commands.
package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/hook"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
	"github.com/autom8y/ariadne/internal/session"
)

// ContextOutput represents the output of the context hook.
type ContextOutput struct {
	SessionID     string `json:"session_id,omitempty"`
	Status        string `json:"status,omitempty"`
	Initiative    string `json:"initiative,omitempty"`
	Rite          string `json:"rite,omitempty"`
	CurrentPhase  string `json:"current_phase,omitempty"`
	ExecutionMode string `json:"execution_mode,omitempty"`
	HasSession    bool   `json:"has_session"`
}

// Text implements output.Textable for markdown output.
func (c ContextOutput) Text() string {
	if !c.HasSession {
		return "No active session"
	}

	var b strings.Builder
	b.WriteString("## Session Context\n")
	b.WriteString("| Field | Value |\n")
	b.WriteString("|-------|-------|\n")
	b.WriteString(fmt.Sprintf("| Session | %s |\n", c.SessionID))
	b.WriteString(fmt.Sprintf("| Status | %s |\n", c.Status))
	b.WriteString(fmt.Sprintf("| Initiative | %s |\n", c.Initiative))
	b.WriteString(fmt.Sprintf("| Rite | %s |\n", c.Rite))
	b.WriteString(fmt.Sprintf("| Mode | %s |\n", c.ExecutionMode))
	return b.String()
}

// newContextCmd creates the context hook subcommand.
func newContextCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Inject session context on SessionStart",
		Long: `Reads session context and outputs it for Claude Code injection.

This hook is triggered on SessionStart events. It reads:
- SESSION_CONTEXT.md if a session exists
- ACTIVE_RITE file for team context

Output is formatted as a markdown table suitable for Claude context.

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runContext(ctx)
			})
		},
	}

	return cmd
}

func runContext(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	// Early exit if hooks disabled
	if ctx.shouldEarlyExit() {
		return outputNoSession(printer)
	}

	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a SessionStart event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventSessionStart {
		printer.VerboseLog("debug", "skipping context hook for non-SessionStart event",
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputNoSession(printer)
	}

	// Get resolver for path lookups
	resolver := ctx.getResolver()
	if resolver.ProjectRoot() == "" {
		// Try to discover project from environment
		if hookEnv.ProjectDir != "" {
			resolver = newResolverFromPath(hookEnv.ProjectDir)
		} else {
			return outputNoSession(printer)
		}
	}

	// Get current session ID
	sessionID, err := ctx.getCurrentSessionID()
	if err != nil {
		printer.VerboseLog("warn", "failed to read current session", map[string]interface{}{"error": err.Error()})
		return outputNoSession(printer)
	}

	if sessionID == "" {
		return outputNoSession(printer)
	}

	// Trim any whitespace/newlines from session ID
	sessionID = strings.TrimSpace(sessionID)

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.VerboseLog("warn", "failed to load session context",
			map[string]interface{}{"session_id": sessionID, "error": err.Error()})
		return outputNoSession(printer)
	}

	// Read active rite with backward compatibility
	activeTeam := readActiveRite(resolver)
	if activeTeam == "" {
		activeTeam = sessCtx.ActiveRite
	}

	// Determine execution mode
	mode := determineExecutionMode(sessCtx, activeTeam)

	// Build output
	result := ContextOutput{
		SessionID:     sessCtx.SessionID,
		Status:        string(sessCtx.Status),
		Initiative:    sessCtx.Initiative,
		Rite:          activeTeam,
		CurrentPhase:  sessCtx.CurrentPhase,
		ExecutionMode: mode,
		HasSession:    true,
	}

	return printer.Print(result)
}

// outputNoSession outputs the no-session response.
func outputNoSession(printer *output.Printer) error {
	result := ContextOutput{HasSession: false}
	return printer.Print(result)
}

// readActiveRite reads the ACTIVE_RITE file.
func readActiveRite(resolver *paths.Resolver) string {
	ritePath := resolver.ActiveRiteFile()
	if data, err := os.ReadFile(ritePath); err == nil {
		return strings.TrimSpace(string(data))
	}

	return ""
}

// determineExecutionMode determines the execution mode based on session and team.
func determineExecutionMode(sessCtx *session.Context, activeTeam string) string {
	// No session = native mode
	if sessCtx == nil {
		return "native"
	}

	// Session with team = orchestrated mode
	if activeTeam != "" && activeTeam != "none" {
		return "orchestrated"
	}

	// Session without team = cross-cutting mode
	return "cross-cutting"
}

// newResolverFromPath creates a resolver from a project path.
func newResolverFromPath(projectDir string) *paths.Resolver {
	return paths.NewResolver(projectDir)
}
