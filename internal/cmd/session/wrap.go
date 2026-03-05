package session

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/artifact"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/ledge"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/naxos"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/sails"
	"github.com/autom8y/knossos/internal/session"
)

type wrapOptions struct {
	noArchive   bool
	noGraduate  bool
	force       bool
	autoPromote bool
}

func newWrapCmd(ctx *cmdContext) *cobra.Command {
	var opts wrapOptions

	cmd := &cobra.Command{
		Use:   "wrap",
		Short: "Complete and archive a session",
		Long: `Complete a session, transitioning to ARCHIVED state.

Before archiving, generate a White Sails confidence signal. If sails
are BLACK (explicit blockers present), the wrap is blocked unless --force
is used. The session directory is moved to the archive unless --no-archive
is specified.

After a successful wrap, scan for stale parked sessions and report
them to stderr with a suggestion to wrap them as well.

Examples:
  ari session wrap
  ari session wrap --no-archive
  ari session wrap --force          # Wrap even with BLACK sails

Context:
  Lifecycle command -- invoke via Moirai, not specialists directly.
  BLACK sails block wrap by default (quality gate). Use --force to override.
  Cleans up all lock artifacts, CC map entries, and Moirai locks.
  Emits session.archived and session.end events with cognitive budget.
  Use 'ari session gc' to batch-archive stale parked sessions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWrap(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.noArchive, "no-archive", false, "Don't move to archive directory")
	cmd.Flags().BoolVar(&opts.noGraduate, "no-graduate", false, "Skip artifact graduation to .ledge/")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Force wrap even with BLACK sails")
	cmd.Flags().BoolVar(&opts.autoPromote, "auto-promote", false, "Promote WHITE-sails artifacts to shelf after graduation")

	return cmd
}

func runWrap(ctx *cmdContext, opts wrapOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()
	fsm := session.NewFSM()

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

	// Check if session is already archived before acquiring lock
	archiveSessionPath := resolver.ArchiveDir() + "/" + sessionID
	if _, statErr := os.Stat(archiveSessionPath); statErr == nil {
		err := errors.NewWithDetails(errors.CodeLifecycleViolation,
			fmt.Sprintf("Session %s is already archived", sessionID),
			map[string]any{
				"session_id":   sessionID,
				"archive_path": archiveSessionPath,
				"hint":         "Session was previously wrapped and cannot be wrapped again.",
			})
		printer.PrintError(err)
		return err
	}

	// Acquire exclusive lock
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Exclusive, lock.DefaultTimeout, "ari-session-wrap")
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer func() { _ = sessionLock.Release() }()
	emitLockEvent(resolver, sessionID, "ari-session-wrap")

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
		printer.VerboseLog("warn", "failed to generate sails", map[string]any{"error": sailsErr.Error()})
	} else {
		// Quality gate: Block wrap if sails are BLACK (unless --force)
		if sailsResult.Color == sails.ColorBlack {
			if !opts.force {
				err := errors.NewWithDetails(errors.CodeQualityGateFailed,
					"cannot wrap session with BLACK sails: explicit blockers present",
					map[string]any{
						"color":   string(sailsResult.Color),
						"reasons": sailsResult.Reasons,
					})
				printer.PrintError(err)
				return err
			}
			// If --force, emit warning but continue
			printer.VerboseLog("warn", "wrapping session with BLACK sails (--force used)", map[string]any{
				"color":   string(sailsResult.Color),
				"reasons": sailsResult.Reasons,
			})
		}

		// Emit SAILS_GENERATED event to Clew Contract
		sailsWriter := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
		defer func() { _ = sailsWriter.Close() }()

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
		sailsWriter.Write(sailsEvent)
		if flushErr := sailsWriter.Flush(); flushErr != nil {
			printer.VerboseLog("warn", "failed to emit sails event", map[string]any{"error": flushErr.Error()})
		}
	}

	// Update context
	now := time.Now().UTC()
	previousStatus := sessCtx.Status
	sessCtx.Status = session.StatusArchived
	sessCtx.ArchivedAt = &now

	// Save context
	// Note: This direct save bypasses the Moirai write guard by design.
	// The write guard (hooks/session-guards/session-write-guard.sh) only
	// protects against Claude Code's Write/Edit tools via PreToolUse hooks.
	// Native ariadne commands like `ari session wrap` are the authorized mutation
	// path and bypass the guard through direct Go file I/O (os.WriteFile).
	// See: hooks/session-guards/session-write-guard.sh lines 32-35
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Emit Clew Contract events
	endWriter := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = endWriter.Close() }()
	{
		// Lifecycle event
		endWriter.Write(clewcontract.NewSessionArchivedEvent(sessionID, string(previousStatus)))

		// Session end event with budget
		durationMs := time.Since(sessCtx.CreatedAt).Milliseconds()
		budget := collectCognitiveBudget(sessionDir)

		var sessionEndEvent clewcontract.Event
		if budget != nil {
			sessionEndEvent = clewcontract.NewSessionEndEventWithBudget(sessionID, "completed", durationMs, budget)
		} else {
			sessionEndEvent = clewcontract.NewSessionEndEvent(sessionID, "completed", durationMs)
		}

		endWriter.Write(sessionEndEvent)
		if flushErr := endWriter.Flush(); flushErr != nil {
			printer.VerboseLog("warn", "failed to write events", map[string]any{"error": flushErr.Error()})
		}
	}

	// Graduate artifacts to .ledge/ (non-fatal)
	var graduatedCount int
	var gradResult *artifact.GraduationResult
	if !opts.noGraduate {
		var gradErr error
		gradResult, gradErr = artifact.GraduateSession(resolver, sessionID)
		if gradErr != nil {
			printer.VerboseLog("warn", "artifact graduation failed", map[string]any{"error": gradErr.Error()})
		} else {
			graduatedCount = len(gradResult.Graduated)
			for _, w := range gradResult.Warnings {
				printer.VerboseLog("warn", "artifact graduation warning", map[string]any{"warning": w})
			}
		}
	}

	// Auto-promote graduated artifacts to shelf (non-fatal, opt-in)
	var promotedCount int
	if opts.autoPromote && gradResult != nil && len(gradResult.Graduated) > 0 {
		if sailsResult != nil && sailsResult.Color == sails.ColorWhite {
			promoResult, promoErr := ledge.AutoPromoteSession(resolver, gradResult.Graduated)
			if promoErr != nil {
				printer.VerboseLog("warn", "auto-promotion failed", map[string]any{"error": promoErr.Error()})
			} else {
				promotedCount = len(promoResult.Promoted)
				for _, w := range promoResult.Warnings {
					printer.VerboseLog("warn", "auto-promotion warning", map[string]any{"warning": w})
				}
				// Emit per-artifact events
				for _, p := range promoResult.Promoted {
					endWriter.WriteTyped(
						clewcontract.NewTypedArtifactPromotedEvent(
							sessionID, p.SourcePath, p.ShelfPath, p.Category,
						),
					)
				}
			}
		}
	}

	// If this was a frayed session, emit strand_resolved on parent
	if sessCtx.FrayedFrom != "" {
		parentDir := resolver.SessionDir(sessCtx.FrayedFrom)
		strandWriter := clewcontract.NewBufferedEventWriter(parentDir, clewcontract.DefaultFlushInterval)
		defer func() { _ = strandWriter.Close() }()
		event := clewcontract.NewStrandResolvedEvent(sessCtx.FrayedFrom, sessionID, "wrapped")
		strandWriter.Write(event)
		if flushErr := strandWriter.Flush(); flushErr != nil {
			printer.VerboseLog("warn", "failed to emit strand_resolved event", map[string]any{"error": flushErr.Error()})
		}
	}

	// Clean up all lock artifacts for this session before archive move.
	// All cleanup is best-effort: failures are logged as warnings but do not
	// block the wrap, since the session context is already ARCHIVED.

	// 1. Advisory session lock (.locks/{id}.lock)
	_ = lockMgr.ForceRelease(sessionID)

	// 2. Moirai lock (.moirai-lock in session dir)
	// Must be removed before the archive move so it doesn't persist in the archive.
	moiraiLockPath := sessionDir + "/.moirai-lock"
	if removeErr := os.Remove(moiraiLockPath); removeErr != nil && !os.IsNotExist(removeErr) {
		printer.VerboseLog("warn", "failed to remove moirai lock before archive", map[string]any{"error": removeErr.Error()})
	}

	// 3. CC map entries pointing to this session
	if clearErr := session.ClearCCMapForSession(resolver, sessionID); clearErr != nil {
		printer.VerboseLog("warn", "failed to clear CC map entries", map[string]any{"error": clearErr.Error()})
	}

	// Move to archive if requested
	var archivePath string
	archived := false
	if !opts.noArchive {
		archiveDir := resolver.ArchiveDir()
		if err := paths.EnsureDir(archiveDir); err != nil {
			printer.VerboseLog("warn", "failed to create archive directory", map[string]any{"error": err.Error()})
		} else {
			archivePath = archiveDir + "/" + sessionID
			// Only move if target doesn't exist
			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				if err := os.Rename(sessionDir, archivePath); err != nil {
					printer.VerboseLog("warn", "failed to move to archive", map[string]any{"error": err.Error()})
				} else {
					archived = true
					// Defensive: verify source directory is gone after rename
					// (os.Rename should be atomic, but guard against edge cases)
					if _, statErr := os.Stat(sessionDir); statErr == nil {
						if removeErr := os.RemoveAll(sessionDir); removeErr != nil {
							printer.VerboseLog("warn", "failed to remove ghost session directory after archive", map[string]any{"error": removeErr.Error()})
						}
					}
				}
			} else if err == nil {
				// Archive target already exists (from a previous interrupted wrap).
				// The session context was already updated to ARCHIVED above,
				// so the live directory is a stale ghost — remove it.
				archived = true
				if removeErr := os.RemoveAll(sessionDir); removeErr != nil {
					printer.VerboseLog("warn", "failed to remove ghost session directory", map[string]any{"error": removeErr.Error()})
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

	result.GraduatedCount = graduatedCount
	result.PromotedCount = promotedCount

	// Add sails information to output if generation succeeded
	if sailsResult != nil {
		result.SailsColor = string(sailsResult.Color)
		result.SailsBase = string(sailsResult.ComputedBase)
		result.SailsReasons = sailsResult.Reasons
		result.SailsPath = sailsResult.FilePath
	}

	if err := printer.PrintSuccess(result); err != nil {
		return err
	}

	// Scan for stale parked sessions after successful wrap
	staleThreshold := staleSessionThreshold()
	staleSessions := naxos.ScanStaleSessions(resolver.SessionsDir(), staleThreshold, sessionID)
	if len(staleSessions) > 0 {
		fmt.Fprintf(os.Stderr, "\n---\n")
		fmt.Fprintf(os.Stderr, "Hint: %d stale parked session(s) found:\n", len(staleSessions))
		for _, s := range staleSessions {
			fmt.Fprintf(os.Stderr, "  %s  PARKED  %s ago\n", s.ID, naxos.FormatDuration(s.Age))
		}
		fmt.Fprintf(os.Stderr, "Consider: ari session wrap <session-id>\n")
	}

	return nil
}

// collectCognitiveBudget attempts to collect cognitive budget metadata from the session.
// Returns nil if CLEW_RECORD.ndjson doesn't exist or cannot be read.
// Falls back to THREAD_RECORD.ndjson for legacy sessions.
// Future: Integrate with ARI_MSG_WARN/ARI_MSG_PARK thresholds.
func collectCognitiveBudget(sessionDir string) map[string]any {
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
	defer func() { _ = file.Close() }()

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
		for i := range n {
			if buffer[i] == '\n' {
				totalEvents++
			}
		}
	}

	// Return basic budget data
	// Future enhancements:
	// - Parse individual events to categorize by tool type
	// - Track message count vs thresholds (ARI_MSG_WARN)
	// - Include park suggestions if threshold exceeded
	return map[string]any{
		"total_tool_calls": totalEvents,
		"tool_counts":      toolCounts, // Placeholder for future detailed counts
	}
}
