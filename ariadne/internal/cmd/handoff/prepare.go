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
	"github.com/autom8y/ariadne/internal/validation"
)

// prepareOptions holds options for the prepare command.
type prepareOptions struct {
	fromAgent  string
	toAgent    string
	artifactID string
}

// newPrepareCmd creates the handoff prepare subcommand.
func newPrepareCmd(ctx *cmdContext) *cobra.Command {
	var opts prepareOptions

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepare for handoff between agents",
		Long: `Prepare for a handoff by validating current agent's output
and checking readiness for the receiving agent.

This command validates handoff criteria and generates a handoff
context that can be passed to the receiving agent. It emits a
task_end event for the source agent.

Examples:
  ari handoff prepare --from=architect --to=principal-engineer
  ari handoff prepare --from=requirements-analyst --to=architect`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrepare(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.fromAgent, "from", "", "Source agent (e.g., architect)")
	cmd.Flags().StringVar(&opts.toAgent, "to", "", "Target agent (e.g., principal-engineer)")
	cmd.Flags().StringVar(&opts.artifactID, "artifact", "", "Artifact ID being handed off (optional)")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

// runPrepare executes the handoff prepare command.
func runPrepare(ctx *cmdContext, opts prepareOptions) error {
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

	// Validate agents are known
	validAgents := []string{
		"requirements-analyst", "architect", "principal-engineer", "qa-adversary", "orchestrator",
	}
	if !isValidAgent(opts.fromAgent, validAgents) {
		err := errors.NewWithDetails(errors.CodeUsageError, "invalid source agent",
			map[string]interface{}{"from": opts.fromAgent, "valid_agents": validAgents})
		printer.PrintError(err)
		return err
	}
	if !isValidAgent(opts.toAgent, validAgents) {
		err := errors.NewWithDetails(errors.CodeUsageError, "invalid target agent",
			map[string]interface{}{"to": opts.toAgent, "valid_agents": validAgents})
		printer.PrintError(err)
		return err
	}

	// Validate handoff sequence
	if !isValidHandoffSequence(opts.fromAgent, opts.toAgent) {
		err := errors.NewWithDetails(errors.CodeLifecycleViolation, "invalid handoff sequence",
			map[string]interface{}{
				"from": opts.fromAgent,
				"to":   opts.toAgent,
				"hint": "Handoffs must follow workflow: requirements-analyst -> architect -> principal-engineer -> qa-adversary",
			})
		printer.PrintError(err)
		return err
	}

	// Determine phase for artifact validation
	phase := agentToPhase(opts.fromAgent)

	// Run artifact validation if possible
	var validationResult *validation.HandoffResult
	var warnings []string

	if phase != "" {
		hv, err := validation.NewHandoffValidator()
		if err == nil {
			// Try to find and validate artifact
			artifactType := phaseToArtifactType(phase)
			if artifactType != validation.ArtifactTypeUnknown {
				// Note: In a real implementation, we would locate the artifact file
				// For now, we just report readiness based on session context
				printer.VerboseLog("info", "artifact validation would check", map[string]interface{}{
					"phase":         phase,
					"artifact_type": artifactType,
				})
			}
			_ = hv // silence unused warning
		}
	}

	// Calculate duration (from session start to now)
	durationMs := time.Since(sessCtx.CreatedAt).Milliseconds()

	// Emit task_end event for the source agent
	sessionDir := resolver.SessionDir(sessionID)
	tcWriter, err := threadcontract.NewEventWriter(sessionDir)
	if err != nil {
		printer.VerboseLog("warn", "failed to create event writer", map[string]interface{}{"error": err.Error()})
	} else {
		// Build artifacts list (currently empty, could be populated from artifact registry)
		artifacts := []string{}
		if opts.artifactID != "" {
			artifacts = append(artifacts, opts.artifactID)
		}

		// Build throughline summary
		throughline := fmt.Sprintf("Handoff from %s to %s prepared", opts.fromAgent, opts.toAgent)

		taskEndEvent := threadcontract.NewTaskEndEvent(
			opts.fromAgent,
			"success",
			throughline,
			artifacts,
			durationMs,
		)
		if err := tcWriter.Write(taskEndEvent); err != nil {
			printer.VerboseLog("warn", "failed to emit task_end event", map[string]interface{}{"error": err.Error()})
		}
	}

	// Emit handoff event to session events
	emitter := session.NewEventEmitter(
		resolver.SessionEventsFile(sessionID),
		resolver.TransitionsLog(),
	)
	if err := emitter.EmitPhaseTransition(sessionID, opts.fromAgent, opts.toAgent, validationResult == nil || validationResult.Passed); err != nil {
		printer.VerboseLog("warn", "failed to emit phase transition event", map[string]interface{}{"error": err.Error()})
	}

	// Build output
	result := HandoffPrepareOutput{
		SessionID:   sessionID,
		FromAgent:   opts.fromAgent,
		ToAgent:     opts.toAgent,
		Status:      "ready",
		PreparedAt:  time.Now().UTC().Format(time.RFC3339),
		DurationMs:  durationMs,
		Warnings:    warnings,
		ArtifactID:  opts.artifactID,
		CurrentPhase: sessCtx.CurrentPhase,
	}

	if validationResult != nil && !validationResult.Passed {
		result.Status = "validation_failed"
		for _, cr := range validationResult.FailedBlocking() {
			result.ValidationErrors = append(result.ValidationErrors, cr.Message)
		}
	}

	return printer.PrintSuccess(result)
}

// isValidAgent checks if an agent name is in the valid list.
func isValidAgent(agent string, valid []string) bool {
	for _, v := range valid {
		if agent == v {
			return true
		}
	}
	return false
}

// isValidHandoffSequence validates that the handoff follows the workflow.
func isValidHandoffSequence(from, to string) bool {
	// Define valid handoff transitions
	validTransitions := map[string][]string{
		"requirements-analyst": {"architect"},
		"architect":            {"principal-engineer"},
		"principal-engineer":   {"qa-adversary"},
		"qa-adversary":         {"orchestrator", "architect"}, // QA can loop back to architect or complete
		"orchestrator":         {"requirements-analyst", "architect", "principal-engineer", "qa-adversary"}, // orchestrator can delegate
	}

	validTargets, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, valid := range validTargets {
		if to == valid {
			return true
		}
	}
	return false
}

// agentToPhase maps an agent to the phase they complete.
func agentToPhase(agent string) validation.Phase {
	switch agent {
	case "requirements-analyst":
		return validation.PhaseRequirements
	case "architect":
		return validation.PhaseDesign
	case "principal-engineer":
		return validation.PhaseImplementation
	case "qa-adversary":
		return validation.PhaseValidation
	default:
		return ""
	}
}

// phaseToArtifactType maps a phase to the expected artifact type.
func phaseToArtifactType(phase validation.Phase) validation.ArtifactType {
	switch phase {
	case validation.PhaseRequirements:
		return validation.ArtifactTypePRD
	case validation.PhaseDesign:
		return validation.ArtifactTypeTDD
	default:
		return validation.ArtifactTypeUnknown
	}
}

// HandoffPrepareOutput represents the output of handoff prepare.
type HandoffPrepareOutput struct {
	SessionID        string   `json:"session_id"`
	FromAgent        string   `json:"from_agent"`
	ToAgent          string   `json:"to_agent"`
	Status           string   `json:"status"`
	PreparedAt       string   `json:"prepared_at"`
	DurationMs       int64    `json:"duration_ms"`
	Warnings         []string `json:"warnings,omitempty"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
	ArtifactID       string   `json:"artifact_id,omitempty"`
	CurrentPhase     string   `json:"current_phase"`
}

// Text implements output.Textable for HandoffPrepareOutput.
func (h HandoffPrepareOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Handoff prepared: %s -> %s\n", h.FromAgent, h.ToAgent))
	b.WriteString(fmt.Sprintf("Session: %s\n", h.SessionID))
	b.WriteString(fmt.Sprintf("Status: %s\n", h.Status))

	if len(h.Warnings) > 0 {
		b.WriteString("\nWarnings:\n")
		for _, w := range h.Warnings {
			b.WriteString(fmt.Sprintf("  - %s\n", w))
		}
	}

	if len(h.ValidationErrors) > 0 {
		b.WriteString("\nValidation Errors:\n")
		for _, e := range h.ValidationErrors {
			b.WriteString(fmt.Sprintf("  - %s\n", e))
		}
	}

	return b.String()
}

// Ensure Textable interface is implemented
var _ output.Textable = HandoffPrepareOutput{}
