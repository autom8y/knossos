package session

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/hook/clewcontract"
	"github.com/autom8y/ariadne/internal/lock"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
	"github.com/autom8y/ariadne/internal/sails"
	"github.com/autom8y/ariadne/internal/session"
)

type wrapOptions struct {
	noArchive bool
	force     bool
}

func newWrapCmd(ctx *cmdContext) *cobra.Command {
	var opts wrapOptions

	cmd := &cobra.Command{
		Use:   "wrap",
		Short: "Complete and archive a session",
		Long:  `Completes a session, transitioning to ARCHIVED state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWrap(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.noArchive, "no-archive", false, "Don't move to archive directory")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Force wrap even with BLACK sails")

	return cmd
}

func runWrap(ctx *cmdContext, opts wrapOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.getResolver()
	lockMgr := ctx.getLockManager()
	fsm := session.NewFSM()

	sessionID, err := ctx.getSessionID()
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
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout)
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer sessionLock.Release()

	// Load session context
	sessionDir := resolver.SessionDir(sessionID)
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.ErrSessionNotFound(sessionID)
		}
		printer.PrintError(err)
		return err
	}

	// Validate transition
	if err := fsm.ValidateTransition(sessCtx.Status, session.StatusArchived); err != nil {
		printer.PrintError(err)
		return err
	}

	// Generate White Sails confidence signal before archiving
	var sailsResult *sails.GenerateResult
	sailsGen := sails.NewGenerator(sessionDir)
	sailsResult, sailsErr := sailsGen.Generate()
	if sailsErr != nil {
		// Don't block wrap on sails generation failure - warn and continue
		printer.VerboseLog("warn", "failed to generate sails", map[string]interface{}{"error": sailsErr.Error()})
	} else {
		// Quality gate: Block wrap if sails are BLACK (unless --force)
		if sailsResult.Color == sails.ColorBlack {
			if !opts.force {
				err := errors.NewWithDetails(errors.CodeQualityGateFailed,
					"cannot wrap session with BLACK sails: explicit blockers present",
					map[string]interface{}{
						"color":   string(sailsResult.Color),
						"reasons": sailsResult.Reasons,
					})
				printer.PrintError(err)
				return err
			}
			// If --force, emit warning but continue
			printer.VerboseLog("warn", "wrapping session with BLACK sails (--force used)", map[string]interface{}{
				"color":   string(sailsResult.Color),
				"reasons": sailsResult.Reasons,
			})
		}

		// Emit SAILS_GENERATED event to Clew Contract
		writer, writerErr := clewcontract.NewEventWriter(sessionDir)
		if writerErr != nil {
			printer.VerboseLog("warn", "failed to create event writer for sails", map[string]interface{}{"error": writerErr.Error()})
		} else {
			// Build evidence paths from collected proofs
			var evidencePaths *clewcontract.EvidencePaths
			if sailsResult.Proofs != nil {
				evidencePaths = &clewcontract.EvidencePaths{}
				if proof, ok := sailsResult.Proofs["tests"]; ok && proof.EvidencePath != "" {
					evidencePaths.Tests = proof.EvidencePath
				}
				if proof, ok := sailsResult.Proofs["build"]; ok && proof.EvidencePath != "" {
					evidencePaths.Build = proof.EvidencePath
				}
				if proof, ok := sailsResult.Proofs["lint"]; ok && proof.EvidencePath != "" {
					evidencePaths.Lint = proof.EvidencePath
				}
				if proof, ok := sailsResult.Proofs["adversarial"]; ok && proof.EvidencePath != "" {
					evidencePaths.Adversarial = proof.EvidencePath
				}
				if proof, ok := sailsResult.Proofs["integration"]; ok && proof.EvidencePath != "" {
					evidencePaths.Integration = proof.EvidencePath
				}
			}

			sailsEvent := clewcontract.NewSailsGeneratedEvent(sessionID, clewcontract.SailsGeneratedData{
				Color:         string(sailsResult.Color),
				ComputedBase:  string(sailsResult.ComputedBase),
				Reasons:       sailsResult.Reasons,
				FilePath:      sailsResult.FilePath,
				EvidencePaths: evidencePaths,
			})
			if writeErr := writer.Write(sailsEvent); writeErr != nil {
				printer.VerboseLog("warn", "failed to emit sails event", map[string]interface{}{"error": writeErr.Error()})
			}
		}
	}

	// Update context
	now := time.Now().UTC()
	previousStatus := sessCtx.Status
	sessCtx.Status = session.StatusArchived
	sessCtx.ArchivedAt = &now

	// Save context
	// Note: This direct save bypasses the Moirai write guard by design.
	// The write guard (user-hooks/session-guards/session-write-guard.sh) only
	// protects against Claude Code's Write/Edit tools via PreToolUse hooks.
	// Native ariadne commands like `ari session wrap` are the authorized mutation
	// path and bypass the guard through direct Go file I/O (os.WriteFile).
	// See: user-hooks/session-guards/session-write-guard.sh lines 32-35
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Clear current session
	if err := ctx.clearCurrentSessionID(); err != nil {
		printer.VerboseLog("warn", "failed to clear current session", map[string]interface{}{"error": err.Error()})
	}

	// Emit event
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitArchived(sessionID, string(previousStatus)); err != nil {
		printer.VerboseLog("warn", "failed to emit archive event", map[string]interface{}{"error": err.Error()})
	}

	// Emit Clew Contract session_end event
	tcWriter, err := clewcontract.NewEventWriter(sessionDir)
	if err == nil {
		durationMs := time.Since(sessCtx.CreatedAt).Milliseconds()

		// Collect cognitive budget metadata if available
		budget := collectCognitiveBudget(sessionDir)

		var sessionEndEvent clewcontract.Event
		if budget != nil {
			sessionEndEvent = clewcontract.NewSessionEndEventWithBudget(sessionID, "completed", durationMs, budget)
		} else {
			sessionEndEvent = clewcontract.NewSessionEndEvent(sessionID, "completed", durationMs)
		}

		if err := tcWriter.Write(sessionEndEvent); err != nil {
			printer.VerboseLog("warn", "failed to emit session_end event", map[string]interface{}{"error": err.Error()})
		}
	}

	// Move to archive if requested
	var archivePath string
	archived := false
	if !opts.noArchive {
		archiveDir := resolver.ArchiveDir()
		if err := paths.EnsureDir(archiveDir); err != nil {
			printer.VerboseLog("warn", "failed to create archive directory", map[string]interface{}{"error": err.Error()})
		} else {
			archivePath = archiveDir + "/" + sessionID
			// Only move if target doesn't exist
			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				if err := os.Rename(sessionDir, archivePath); err != nil {
					printer.VerboseLog("warn", "failed to move to archive", map[string]interface{}{"error": err.Error()})
				} else {
					archived = true
				}
			}
		}
	}

	// Output result
	result := output.TransitionOutput{
		SessionID:      sessionID,
		Status:         string(session.StatusArchived),
		PreviousStatus: string(previousStatus),
		ArchivedAt:     now.Format(time.RFC3339),
		Archived:       archived,
		ArchivePath:    archivePath,
	}

	// Add sails information to output if generation succeeded
	if sailsResult != nil {
		result.SailsColor = string(sailsResult.Color)
		result.SailsBase = string(sailsResult.ComputedBase)
		result.SailsReasons = sailsResult.Reasons
		result.SailsPath = sailsResult.FilePath
	}

	return printer.PrintSuccess(result)
}

// collectCognitiveBudget attempts to collect cognitive budget metadata from the session.
// Returns nil if CLEW_RECORD.ndjson doesn't exist or cannot be read.
// Falls back to THREAD_RECORD.ndjson for legacy sessions.
// Future: Integrate with ARIADNE_MSG_WARN/ARIADNE_MSG_PARK thresholds.
func collectCognitiveBudget(sessionDir string) map[string]interface{} {
	// Try new path first, fall back to legacy path
	clewRecordPath := sessionDir + "/CLEW_RECORD.ndjson"
	threadRecordPath := sessionDir + "/THREAD_RECORD.ndjson"

	recordPath := clewRecordPath
	if _, err := os.Stat(clewRecordPath); os.IsNotExist(err) {
		// Fall back to legacy path for backwards compatibility
		if _, err := os.Stat(threadRecordPath); os.IsNotExist(err) {
			return nil
		}
		recordPath = threadRecordPath
	}

	// Read and count tool events
	file, err := os.Open(recordPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	toolCounts := make(map[string]int)
	totalEvents := 0

	// Simple line-by-line count (NDJSON format)
	// Future: Parse JSON to get more detailed metrics
	scanner := os.NewFile(file.Fd(), recordPath)
	buffer := make([]byte, 4096)
	for {
		n, err := scanner.Read(buffer)
		if err != nil {
			break
		}
		for i := 0; i < n; i++ {
			if buffer[i] == '\n' {
				totalEvents++
			}
		}
	}

	// Return basic budget data
	// Future enhancements:
	// - Parse individual events to categorize by tool type
	// - Track message count vs thresholds (ARIADNE_MSG_WARN)
	// - Include park suggestions if threshold exceeded
	return map[string]interface{}{
		"total_tool_calls": totalEvents,
		"tool_counts":      toolCounts, // Placeholder for future detailed counts
	}
}
