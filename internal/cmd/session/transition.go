package session

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
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
		Long: `Transitions between workflow phases within an active session.

Valid phases: requirements, design, implementation, validation, complete

Phases must progress forward. Artifact validation is performed by default
and can be skipped with --force.

Examples:
  ari session transition design
  ari session transition implementation
  ari session transition complete --force`,
		Args: cobra.ExactArgs(1),
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
		printer.PrintError(err)
		return err
	}

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		return err
	}

	if sessionID == "" {
		err := errors.ErrSessionNotFound("")
		printer.PrintError(err)
		return err
	}

	// Acquire exclusive lock
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout, "ari-session-transition")
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer sessionLock.Release()

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := sess.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		printer.PrintError(err)
		return err
	}

	// Check session is active
	if sessCtx.Status != sess.StatusActive {
		err := errors.ErrLifecycleViolation(string(sessCtx.Status), "phase transition",
			"session must be ACTIVE to transition phases")
		printer.PrintError(err)
		return err
	}

	fromPhase := sessCtx.CurrentPhase

	// Validate phase progression
	if !sess.CanTransitionPhase(sess.Phase(fromPhase), sess.Phase(targetPhase)) {
		err := errors.NewWithDetails(errors.CodeLifecycleViolation,
			"invalid phase transition: phases must progress forward",
			map[string]interface{}{
				"from_phase": fromPhase,
				"to_phase":   targetPhase,
			})
		printer.PrintError(err)
		return err
	}

	// Validate artifacts if not forced
	artifactsValidated := true
	if !opts.force {
		missing := validateArtifacts(resolver.ProjectRoot(), sess.Phase(targetPhase))
		if len(missing) > 0 {
			err := errors.NewWithDetails(errors.CodeLifecycleViolation,
				"cannot transition: missing required artifacts",
				map[string]interface{}{
					"from_phase":       fromPhase,
					"to_phase":         targetPhase,
					"missing_artifacts": missing,
				})
			printer.PrintError(err)
			return err
		}
	} else {
		artifactsValidated = false
	}

	// Rotate SESSION_CONTEXT on phase transition to keep context compact
	sessionDir := resolver.SessionDir(sessionID)
	rotResult, rotErr := sess.RotateSessionContext(sessionDir, sess.DefaultMaxLines, sess.DefaultKeepLines)
	if rotErr != nil {
		printer.VerboseLog("warn", "failed to rotate SESSION_CONTEXT on transition", map[string]interface{}{"error": rotErr.Error()})
	} else if rotResult.Rotated {
		printer.VerboseLog("info", "rotated SESSION_CONTEXT on transition", map[string]interface{}{
			"archived_lines": rotResult.ArchivedLines,
			"kept_lines":     rotResult.KeptLines,
		})
	}

	// Update phase
	now := time.Now().UTC()
	sessCtx.CurrentPhase = targetPhase

	// Save context
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Emit event
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitPhaseTransition(sessionID, fromPhase, targetPhase, artifactsValidated); err != nil {
		printer.VerboseLog("warn", "failed to emit transition event", map[string]interface{}{"error": err.Error()})
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
func validateArtifacts(projectRoot string, targetPhase sess.Phase) []string {
	var missing []string
	docsDir := filepath.Join(projectRoot, "docs")

	switch targetPhase {
	case sess.PhaseDesign:
		// Requires PRD
		if !hasArtifacts(filepath.Join(docsDir, "requirements"), "PRD-*.md") {
			missing = append(missing, "PRD: No PRD found in docs/requirements/")
		}
	case sess.PhaseImplementation:
		// Requires TDD
		if !hasArtifacts(filepath.Join(docsDir, "design"), "TDD-*.md") {
			missing = append(missing, "TDD: No TDD found in docs/design/")
		}
	case sess.PhaseComplete:
		// Requires test plan
		if !hasArtifacts(filepath.Join(docsDir, "testing"), "TP-*.md") {
			missing = append(missing, "Test Plan: No test plan found in docs/testing/")
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
