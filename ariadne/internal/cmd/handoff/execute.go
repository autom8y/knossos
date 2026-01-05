// Package handoff implements the ari handoff commands for agent handoff management.
package handoff

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/hook/threadcontract"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/session"
)

// executeOptions holds options for the execute command.
type executeOptions struct {
	artifactID string
	toAgent    string
	dryRun     bool
}

// newExecuteCmd creates the handoff execute subcommand.
func newExecuteCmd(ctx *cmdContext) *cobra.Command {
	var opts executeOptions

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a prepared handoff",
		Long: `Execute a handoff that has been prepared, recording the
transition in the session audit log and emitting a task_start
event for the receiving agent.

The handoff is recorded to events.jsonl for tracking purposes.
This command delegates actual state mutations to state-mate
when running within a Claude Code session.

Examples:
  ari handoff execute --artifact=TDD-user-auth --to=principal-engineer
  ari handoff execute --artifact=PRD-user-auth --to=architect`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExecute(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.artifactID, "artifact", "", "Artifact ID being handed off")
	cmd.Flags().StringVar(&opts.toAgent, "to", "", "Target agent receiving the handoff")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview handoff without executing")
	_ = cmd.MarkFlagRequired("artifact")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

// runExecute executes the handoff execute command.
func runExecute(ctx *cmdContext, opts executeOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()

	// Get session ID
	sessionID, err := ctx.getSessionID()
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.New(errors.CodeSessionNotFound, "No active session. Use 'ari session create' first.")
		printer.PrintError(err)
		return err
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		printer.PrintError(err)
		return err
	}

	// Validate session is ACTIVE
	if sessCtx.Status != session.StatusActive {
		err := errors.ErrLifecycleViolation(string(sessCtx.Status), "handoff",
			"session must be ACTIVE for handoff")
		printer.PrintError(err)
		return err
	}

	// Validate target agent
	validAgents := []string{
		"requirements-analyst", "architect", "principal-engineer", "qa-adversary", "orchestrator",
	}
	if !isValidAgent(opts.toAgent, validAgents) {
		err := errors.NewWithDetails(errors.CodeUsageError, "invalid target agent",
			map[string]interface{}{"to": opts.toAgent, "valid_agents": validAgents})
		printer.PrintError(err)
		return err
	}

	// Determine target phase
	targetPhase := agentToTargetPhase(opts.toAgent)

	// Build result (for dry-run or actual execution)
	now := time.Now().UTC()
	result := HandoffExecuteOutput{
		SessionID:   sessionID,
		ArtifactID:  opts.artifactID,
		ToAgent:     opts.toAgent,
		TargetPhase: targetPhase,
		ExecutedAt:  now.Format(time.RFC3339),
		DryRun:      opts.dryRun,
	}

	if opts.dryRun {
		result.Status = "would_execute"
		result.Message = fmt.Sprintf("Would hand off artifact %s to %s", opts.artifactID, opts.toAgent)
		return printer.Print(result)
	}

	// Emit task_start event for the receiving agent
	sessionDir := resolver.SessionDir(sessionID)
	tcWriter, err := threadcontract.NewEventWriter(sessionDir)
	if err != nil {
		printer.VerboseLog("warn", "failed to create event writer", map[string]interface{}{"error": err.Error()})
	} else {
		taskDescription := fmt.Sprintf("Handoff received: implement %s", opts.artifactID)
		taskStartEvent := threadcontract.NewTaskStartEvent(
			opts.toAgent,
			taskDescription,
			sessionID,
		)
		if err := tcWriter.Write(taskStartEvent); err != nil {
			printer.VerboseLog("warn", "failed to emit task_start event", map[string]interface{}{"error": err.Error()})
		}
	}

	// Record handoff event in session events
	emitter := session.NewEventEmitter(
		resolver.SessionEventsFile(sessionID),
		resolver.TransitionsLog(),
	)

	// Emit handoff executed event
	handoffEvent := session.Event{
		Timestamp: now.Format(time.RFC3339),
		Event:     "HANDOFF_EXECUTED",
		To:        opts.toAgent,
		Metadata: map[string]interface{}{
			"artifact_id":  opts.artifactID,
			"target_phase": targetPhase,
			"from_phase":   sessCtx.CurrentPhase,
		},
	}
	if err := emitter.Emit(handoffEvent); err != nil {
		printer.VerboseLog("warn", "failed to emit handoff event", map[string]interface{}{"error": err.Error()})
	}
	if err := emitter.EmitToAudit(sessionID, handoffEvent); err != nil {
		printer.VerboseLog("warn", "failed to emit audit event", map[string]interface{}{"error": err.Error()})
	}

	result.Status = "executed"
	result.Message = fmt.Sprintf("Handoff to %s executed successfully", opts.toAgent)

	return printer.PrintSuccess(result)
}

// agentToTargetPhase maps a receiving agent to the phase they work in.
func agentToTargetPhase(agent string) string {
	switch agent {
	case "requirements-analyst":
		return "requirements"
	case "architect":
		return "design"
	case "principal-engineer":
		return "implementation"
	case "qa-adversary":
		return "validation"
	case "orchestrator":
		return "orchestration"
	default:
		return "unknown"
	}
}

// HandoffExecuteOutput represents the output of handoff execute.
type HandoffExecuteOutput struct {
	SessionID   string `json:"session_id"`
	ArtifactID  string `json:"artifact_id"`
	ToAgent     string `json:"to_agent"`
	TargetPhase string `json:"target_phase"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	ExecutedAt  string `json:"executed_at"`
	DryRun      bool   `json:"dry_run,omitempty"`
}

// Text implements output.Textable for HandoffExecuteOutput.
func (h HandoffExecuteOutput) Text() string {
	var b strings.Builder
	if h.DryRun {
		b.WriteString("[DRY RUN] ")
	}
	b.WriteString(fmt.Sprintf("Handoff executed: %s -> %s\n", h.ArtifactID, h.ToAgent))
	b.WriteString(fmt.Sprintf("Session: %s\n", h.SessionID))
	b.WriteString(fmt.Sprintf("Target Phase: %s\n", h.TargetPhase))
	b.WriteString(fmt.Sprintf("Status: %s\n", h.Status))
	return b.String()
}

// Ensure Textable interface is implemented
var _ output.Textable = HandoffExecuteOutput{}
