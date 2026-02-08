package session

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

type frayOptions struct {
	noWorktree bool
	fromID     string
}

func newFrayCmd(ctx *cmdContext) *cobra.Command {
	var opts frayOptions

	cmd := &cobra.Command{
		Use:   "fray",
		Short: "Fork session into a parallel strand",
		Long: `Forks the current (or specified) session into a new parallel strand.

The parent session is parked and a new child session is created that carries
forward the parent's initiative, complexity, rite, and phase. The child's
body includes the parent's context with a "Frayed from" header.

By default, a git worktree is created for filesystem isolation. Use
--no-worktree to skip worktree creation (child shares the same working tree).

Examples:
  ari session fray
  ari session fray --no-worktree
  ari session fray --from session-20260206-120000-abcdef01`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFray(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.noWorktree, "no-worktree", false, "Skip git worktree creation")
	cmd.Flags().StringVar(&opts.fromID, "from", "", "Source session ID (default: current session)")

	return cmd
}

func runFray(ctx *cmdContext, opts frayOptions) error {
	printer := ctx.getPrinter()

	// Resolve parent session ID
	parentID := opts.fromID
	if parentID == "" {
		var err error
		parentID, err = ctx.GetSessionID()
		if err != nil {
			printer.PrintError(errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
			return err
		}
	}

	if parentID == "" {
		err := errors.ErrSessionNotFound("")
		printer.PrintError(err)
		return err
	}

	projectDir := ""
	if ctx.ProjectDir != nil {
		projectDir = *ctx.ProjectDir
	}

	result, err := fraySession(projectDir, parentID, opts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	return printer.Print(result)
}

// fraySession performs the core fray operation. Extracted for testability.
func fraySession(projectDir, parentID string, opts frayOptions) (*output.FrayOutput, error) {
	resolver := paths.NewResolver(projectDir)

	// Load parent context
	parentCtxPath := resolver.SessionContextFile(parentID)
	parentCtx, err := session.LoadContext(parentCtxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.ErrSessionNotFound(parentID)
		}
		return nil, err
	}

	// Validate parent is ACTIVE
	if parentCtx.Status != session.StatusActive {
		return nil, errors.ErrLifecycleViolation(string(parentCtx.Status), "PARKED",
			"can only fray an ACTIVE session")
	}

	// Create child session inheriting parent's attributes
	childCtx := session.NewContext(parentCtx.Initiative, parentCtx.Complexity, parentCtx.ActiveRite)
	childCtx.SchemaVersion = "2.2"
	childCtx.CurrentPhase = parentCtx.CurrentPhase
	childCtx.FrayedFrom = parentID
	childCtx.FrayPoint = parentCtx.CurrentPhase
	childCtx.Rite = parentCtx.Rite

	// Copy parent body with fray header prepended
	frayHeader := fmt.Sprintf("\n## Frayed from %s\n\nContinuing from phase: %s\n",
		parentID, parentCtx.CurrentPhase)
	if parentCtx.Body != "" {
		childCtx.Body = frayHeader + parentCtx.Body
	} else {
		childCtx.Body = frayHeader
	}

	// Create child session directory
	childDir := resolver.SessionDir(childCtx.SessionID)
	if err := paths.EnsureDir(childDir); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create child session directory", err)
	}

	// Save child context
	childCtxPath := resolver.SessionContextFile(childCtx.SessionID)
	if err := childCtx.Save(childCtxPath); err != nil {
		os.RemoveAll(childDir)
		return nil, err
	}

	// Park parent
	now := time.Now().UTC()
	parentCtx.Status = session.StatusParked
	parentCtx.ParkedAt = &now
	parentCtx.ParkedReason = fmt.Sprintf("Frayed to %s", childCtx.SessionID)
	parentCtx.Strands = append(parentCtx.Strands, childCtx.SessionID)

	if err := parentCtx.Save(parentCtxPath); err != nil {
		os.RemoveAll(childDir)
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to update parent context", err)
	}

	// Create worktree if requested (default behavior)
	var worktreePath string
	if !opts.noWorktree {
		worktreePath = fmt.Sprintf("/tmp/knossos-fray-%d-%d", time.Now().Unix(), os.Getpid())
		if err := createWorktree(worktreePath); err != nil {
			// Non-fatal: worktree creation failure shouldn't block the fray
			// The session is already created, parent is parked — just skip the worktree
			worktreePath = ""
		}
	}

	// Update current session cache to point to child
	currentSessionFile := resolver.CurrentSessionFile()
	os.WriteFile(currentSessionFile, []byte(childCtx.SessionID), 0644)

	// Emit events (non-fatal)
	emitFrayEvents(resolver, parentID, childCtx.SessionID, parentCtx.CurrentPhase)

	// Build result
	result := &output.FrayOutput{
		ParentID:     parentID,
		ChildID:      childCtx.SessionID,
		ChildDir:     childDir,
		FrayPoint:    childCtx.FrayPoint,
		WorktreePath: worktreePath,
		Status:       string(childCtx.Status),
		CreatedAt:    childCtx.CreatedAt.Format(time.RFC3339),
	}

	return result, nil
}

// emitFrayEvents emits session and clew events for a fray operation.
// All event emissions are non-fatal — failures are silently ignored.
func emitFrayEvents(resolver *paths.Resolver, parentID, childID, frayPoint string) {
	// Emit SESSION_FRAYED on parent's event log
	parentEventsPath := resolver.SessionEventsFile(parentID)
	parentAuditPath := resolver.TransitionsLog()
	parentEmitter := session.NewEventEmitter(parentEventsPath, parentAuditPath)
	parentEmitter.EmitFrayed(parentID, childID, frayPoint)

	// Emit SESSION_CREATED on child's event log
	childEventsPath := resolver.SessionEventsFile(childID)
	childEmitter := session.NewEventEmitter(childEventsPath, parentAuditPath)
	childEmitter.EmitCreated(childID, "", "", "")

	// Emit clew session_frayed event on parent
	parentDir := resolver.SessionDir(parentID)
	frayWriter := clewcontract.NewBufferedEventWriter(parentDir, clewcontract.DefaultFlushInterval)
	defer frayWriter.Close()
	frayEvent := clewcontract.NewSessionFrayedEvent(parentID, childID, frayPoint)
	frayWriter.Write(frayEvent)
	frayWriter.Flush() // Best-effort flush for fray events
}
