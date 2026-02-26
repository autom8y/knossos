package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	sess "github.com/autom8y/knossos/internal/session"
)

// newContextCmd creates the "ari session context" subgroup.
func newContextCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Query and project session context",
		Long: `Commands for querying and projecting session context for agents.

Context:
  Subcommands provide role-adaptive context projections.
  The SubagentStart hook calls 'snapshot' automatically on agent spawn.`,
	}

	cmd.AddCommand(newSnapshotCmd(ctx))

	return cmd
}

// snapshotOptions holds the flag values for `ari session context snapshot`.
type snapshotOptions struct {
	role      string
	agentName string
}

// newSnapshotCmd creates the "ari session context snapshot" command.
func newSnapshotCmd(ctx *cmdContext) *cobra.Command {
	var opts snapshotOptions

	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Generate role-adaptive context snapshot",
		Long: `Generate a session context projection sized for the target agent role.

Orchestrators get full context (last 10 timeline entries, all decisions, blockers).
Specialists get a scoped view (last 5 + up to 3 agent-scoped entries, blockers).
Background agents get minimal context (phase, complexity, status only).

The SubagentStart hook calls this automatically when spawning subagents.
Use this command directly to preview what a spawned agent would receive,
or to refresh stale context mid-conversation.

Examples:
  ari session context snapshot --role=orchestrator
  ari session context snapshot --role=specialist --agent=context-architect
  ari session context snapshot --role=background
  ari session context snapshot -o json

Context:
  SubagentStart hook calls this automatically -- agents rarely need direct calls.
  Call directly to refresh stale context mid-conversation or preview agent views.
  Read-only: no locks, no writes, no events emitted.
  Use -o json for structured consumption; default output is markdown.
  Prefer this over 'ari session timeline' for injecting context into agents.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshot(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.role, "role", "orchestrator",
		"Agent role: orchestrator, specialist, background")
	cmd.Flags().StringVar(&opts.agentName, "agent", "",
		"Agent name (for specialist scoping)")

	return cmd
}

// runSnapshot implements the snapshot command.
// It is read-only: no locks, no writes, no events emitted.
func runSnapshot(ctx *cmdContext, opts snapshotOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Resolve session ID — required.
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err)
	}
	if sessionID == "" {
		return errors.ErrSessionNotFound("")
	}

	// Validate role.
	role := sess.SnapshotRole(opts.role)
	switch role {
	case sess.RoleOrchestrator, sess.RoleSpecialist, sess.RoleBackground:
		// valid
	default:
		return errors.New(errors.CodeValidationFailed,
			fmt.Sprintf("invalid role %q: must be orchestrator, specialist, or background", opts.role))
	}

	// Load session context (frontmatter + body).
	ctxPath := resolver.SessionContextFile(sessionID)
	if _, statErr := os.Stat(ctxPath); os.IsNotExist(statErr) {
		return errors.ErrSessionNotFound(sessionID)
	}

	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		return err
	}

	// Determine events path.
	eventsPath := resolver.SessionEventsFile(sessionID)

	// Generate snapshot (pure read — no locks needed).
	config := sess.SnapshotConfig{
		Role:      role,
		AgentName: opts.agentName,
	}

	snap, err := sess.GenerateSnapshot(sessCtx, eventsPath, config)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to generate snapshot", err)
	}

	// Output the result.
	// For JSON output: write raw JSON from RenderJSON() directly.
	// For text output: wrap the markdown in SnapshotOutput.
	outputFormat := ""
	if ctx.Output != nil {
		outputFormat = *ctx.Output
	}

	if outputFormat == "json" {
		raw, marshalErr := snap.RenderJSON()
		if marshalErr != nil {
			return errors.Wrap(errors.CodeGeneralError, "failed to marshal snapshot JSON", marshalErr)
		}
		// Write raw JSON + newline directly to avoid double-encoding.
		printer.PrintText(string(raw))
		printer.PrintText("\n")
		return nil
	}

	// Text output: render markdown and wrap in SnapshotOutput.
	return printer.Print(output.SnapshotOutput{
		Markdown: snap.RenderMarkdown(),
	})
}
