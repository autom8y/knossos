// Package hook implements the ari hook commands.
package hook

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/hook"
	"github.com/autom8y/ariadne/internal/output"
)

// CommandCategory represents the category of a routed command.
type CommandCategory string

// Command categories for routing.
const (
	CategorySession      CommandCategory = "session"
	CategoryOrchestrator CommandCategory = "orchestrator"
	CategoryInitiative   CommandCategory = "initiative"
	CategoryGit          CommandCategory = "git"
	CategoryClew         CommandCategory = "clew"
)

// RouteOutput represents the output of the route hook.
type RouteOutput struct {
	Routed   bool            `json:"routed"`
	Command  string          `json:"command,omitempty"`
	Args     string          `json:"args,omitempty"`
	Category CommandCategory `json:"category,omitempty"`
}

// Text implements output.Textable for text output.
func (r RouteOutput) Text() string {
	if !r.Routed {
		return "not routed"
	}
	var b strings.Builder
	b.WriteString("Routed: ")
	b.WriteString(r.Command)
	if r.Args != "" {
		b.WriteString(" ")
		b.WriteString(r.Args)
	}
	b.WriteString(" (")
	b.WriteString(string(r.Category))
	b.WriteString(")")
	return b.String()
}

// commandMapping defines the routing rules for slash commands.
var commandMapping = map[string]CommandCategory{
	"/start":   CategorySession,
	"/park":    CategorySession,
	"/resume":  CategorySession,
	"/wrap":    CategorySession,
	"/consult": CategoryOrchestrator,
	"/task":    CategoryInitiative,
	"/sprint":  CategoryInitiative,
	"/commit":  CategoryGit,
	"/pr":      CategoryGit,
	"/stamp":   CategoryClew,
}

// newRouteCmd creates the route hook subcommand.
func newRouteCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Route slash commands on UserPromptSubmit",
		Long: `Detects and routes slash commands from user prompts.

This hook is triggered on UserPromptSubmit events. It:
- Parses CLAUDE_USER_MESSAGE for slash commands
- Returns routing information if a known command is detected
- Returns routed=false for regular messages

Supported commands:
  Session: /start, /park, /resume, /wrap
  Orchestrator: /consult
  Initiative: /task, /sprint
  Git: /commit, /pr
  Thread: /stamp

Note: This hook only performs detection/routing, not execution.

Performance: <5ms target execution time (simple string prefix matching).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runRoute(ctx)
			})
		},
	}

	return cmd
}

func runRoute(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a UserPromptSubmit event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventUserPromptSubmit {
		printer.VerboseLog("debug", "skipping route hook for non-UserPromptSubmit event",
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputNotRouted(printer)
	}

	// Get user message from environment
	userMessage := hookEnv.UserMessage
	if userMessage == "" {
		return outputNotRouted(printer)
	}

	// Parse the message for slash commands
	command, args, category := parseSlashCommand(userMessage)
	if command == "" {
		return outputNotRouted(printer)
	}

	// Build output
	result := RouteOutput{
		Routed:   true,
		Command:  command,
		Args:     args,
		Category: category,
	}

	return printer.Print(result)
}

// parseSlashCommand parses a user message and extracts slash command info.
// Returns empty strings if no valid slash command is found.
func parseSlashCommand(message string) (command string, args string, category CommandCategory) {
	// Trim leading whitespace
	message = strings.TrimSpace(message)

	// Must start with /
	if !strings.HasPrefix(message, "/") {
		return "", "", ""
	}

	// Extract the command (first word)
	parts := strings.SplitN(message, " ", 2)
	cmd := strings.ToLower(parts[0])

	// Look up the command category
	cat, ok := commandMapping[cmd]
	if !ok {
		return "", "", ""
	}

	// Extract args if present
	cmdArgs := ""
	if len(parts) > 1 {
		cmdArgs = strings.TrimSpace(parts[1])
	}

	return cmd, cmdArgs, cat
}

// outputNotRouted outputs the not-routed response.
func outputNotRouted(printer *output.Printer) error {
	result := RouteOutput{Routed: false}
	return printer.Print(result)
}

// runRouteWithPrinter is a helper that uses an injected printer for testing.
func runRouteWithPrinter(ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv()

	// Verify this is a UserPromptSubmit event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventUserPromptSubmit {
		printer.VerboseLog("debug", "skipping route hook for non-UserPromptSubmit event",
			map[string]interface{}{"event": string(hookEnv.Event)})
		return outputNotRoutedWithPrinter(printer)
	}

	// Get user message from environment
	userMessage := hookEnv.UserMessage
	if userMessage == "" {
		return outputNotRoutedWithPrinter(printer)
	}

	// Parse the message for slash commands
	command, args, category := parseSlashCommand(userMessage)
	if command == "" {
		return outputNotRoutedWithPrinter(printer)
	}

	// Build output
	result := RouteOutput{
		Routed:   true,
		Command:  command,
		Args:     args,
		Category: category,
	}

	return printer.Print(result)
}

// outputNotRoutedWithPrinter outputs the not-routed response with an injected printer.
func outputNotRoutedWithPrinter(printer *output.Printer) error {
	result := RouteOutput{Routed: false}
	return printer.Print(result)
}
