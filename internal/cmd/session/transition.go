package session

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	sess "github.com/autom8y/knossos/internal/session"
)

type transitionOptions struct {
	force bool
}

func newTransitionCmd(ctx *cmdContext) *cobra.Command {
	var opts transitionOptions

	cmd := &cobra.Command{
		Use:   "transition <phase>",
		Short: "Transition between workflow phases",
		Long: `Transition between workflow phases within an active session.

Valid phases: requirements, design, implementation, validation, complete

Phases must progress forward. Artifact validation is performed by default
and can be skipped with --force. Rotates SESSION_CONTEXT.md to keep
context compact across phase boundaries.

Examples:
  ari session transition design
  ari session transition implementation
  ari session transition complete --force

Context:
  Use this instead of 'ari session field-set current_phase' -- field-set
  rejects phase mutations and redirects here.
  Validates forward-only progression and checks for required artifacts.
  Use --force to skip artifact validation during rapid prototyping.
  Emits phase.transitioned event. Orchestrators call this at phase gates.`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTransition(ctx, args[0], opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip artifact validation")

	return cmd
}

func runTransition(ctx *cmdContext, targetPhase string, opts transitionOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	// Validate phase
	if !sess.IsValidPhase(targetPhase) {
		err := errors.New(errors.CodeUsageError, "invalid phase: must be requirements, design, implementation, validation, or complete")
		return common.PrintAndReturn(printer, err)
	}

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		return common.PrintAndReturn(printer, err)
	}

	// Acquire exclusive lock
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout, "ari-session-transition")
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}
	defer func() { _ = sessionLock.Release() }()
	emitLockEvent(resolver, sessionID, "ari-session-transition")

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		return common.PrintAndReturn(printer, err)
	}

	// Check session is active
	if sessCtx.Status != sess.StatusActive {
		err := errors.ErrLifecycleViolation(string(sessCtx.Status), "phase transition",
			"session must be ACTIVE to transition phases")
		return common.PrintAndReturn(printer, err)
	}

	fromPhase := sessCtx.CurrentPhase

	// Validate phase progression
	if !sess.CanTransitionPhase(sess.Phase(fromPhase), sess.Phase(targetPhase)) {
		err := errors.NewWithDetails(errors.CodeLifecycleViolation,
			"invalid phase transition: phases must progress forward",
			map[string]any{
				"from_phase": fromPhase,
				"to_phase":   targetPhase,
			})
		return common.PrintAndReturn(printer, err)
	}

	// Validate artifacts if not forced
	if !opts.force {
		missing := validateArtifacts(resolver.ProjectRoot(), sess.Phase(targetPhase))
		if len(missing) > 0 {
			err := errors.NewWithDetails(errors.CodeLifecycleViolation,
				"cannot transition: missing required artifacts",
				map[string]any{
					"from_phase":       fromPhase,
					"to_phase":         targetPhase,
					"missing_artifacts": missing,
				})
			return common.PrintAndReturn(printer, err)
		}
	}

	// Rotate SESSION_CONTEXT on phase transition to keep context compact
	sessionDir := resolver.SessionDir(sessionID)
	rotResult, rotErr := sess.RotateSessionContext(sessionDir, sess.DefaultMaxLines, sess.DefaultKeepLines)
	if rotErr != nil {
		printer.VerboseLog("warn", "failed to rotate SESSION_CONTEXT on transition", map[string]any{"error": rotErr.Error()})
	} else if rotResult.Rotated {
		printer.VerboseLog("info", "rotated SESSION_CONTEXT on transition", map[string]any{
			"archived_lines": rotResult.ArchivedLines,
			"kept_lines":     rotResult.KeptLines,
		})
	}

	// Update phase
	now := time.Now().UTC()
	sessCtx.CurrentPhase = targetPhase

	// Save context
	if err := sessCtx.Save(ctxPath); err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Emit lifecycle event
	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = writer.Close() }()
	writer.Write(clewcontract.NewPhaseTransitionedEvent(sessionID, fromPhase, targetPhase))
	if err := writer.Flush(); err != nil {
		printer.VerboseLog("warn", "failed to write event", map[string]any{"error": err.Error()})
	}

	// Output result
	result := output.TransitionOutput{
		SessionID:     sessionID,
		FromPhase:     fromPhase,
		ToPhase:       targetPhase,
		TransitionedAt: now.Format(time.RFC3339),
	}

	return printer.PrintSuccess(result)
}

// validateArtifacts checks for required artifacts for the target phase.
// All artifacts live in .ledge/specs/ (PRD, TDD, TP) or .ledge/decisions/ (ADR).
func validateArtifacts(projectRoot string, targetPhase sess.Phase) []string {
	var missing []string
	resolver := paths.NewResolver(projectRoot)

	switch targetPhase {
	case sess.PhaseDesign:
		if !hasArtifacts(resolver.LedgeSpecsDir(), "PRD-*.md") {
			missing = append(missing, "PRD: No PRD found in .ledge/specs/")
		}
	case sess.PhaseImplementation:
		if !hasArtifacts(resolver.LedgeSpecsDir(), "TDD-*.md") {
			missing = append(missing, "TDD: No TDD found in .ledge/specs/")
		}
	case sess.PhaseComplete:
		if !hasArtifacts(resolver.LedgeSpecsDir(), "TP-*.md") {
			missing = append(missing, "Test Plan: No test plan found in .ledge/specs/")
		}
	}

	return missing
}

// hasArtifacts checks if any files matching the pattern exist in the directory.
func hasArtifacts(dir, pattern string) bool {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return false
	}
	for _, match := range matches {
		if info, err := os.Stat(match); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}
