package session

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/lock"
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
		Short: "Fork session into parallel strand",
		Long: `Fork the current (or specified) session into a new parallel strand.

The parent session is parked and a new child session is created that carries
forward the parent's initiative, complexity, rite, and phase. The child's
body includes the parent's context with a "Frayed from" header.

By default, a git worktree is created for filesystem isolation. Use
--no-worktree to skip worktree creation (child shares the same working tree).

Examples:
  ari session fray
  ari session fray --no-worktree
  ari session fray --from session-20260206-120000-abcdef01

Context:
  Use for parallel workstreams within the same initiative.
  Parent is auto-parked; child inherits phase, rite, and complexity.
  The child emits strand_resolved on its parent when wrapped.
  Worktree creation is best-effort -- failure does not block the fray.
  Use --from to fray a specific session instead of the current one.`,
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
			return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
		}
	}

	if parentID == "" {
		err := errors.ErrSessionNotFound("")
		return common.PrintAndReturn(printer, err)
	}

	projectDir := ""
	if ctx.ProjectDir != nil {
		projectDir = *ctx.ProjectDir
	}

	result, err := fraySession(projectDir, parentID, opts)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	return printer.Print(result)
}

// fraySession performs the core fray operation. Extracted for testability.
func fraySession(projectDir, parentID string, opts frayOptions) (*output.FrayOutput, error) {
	resolver := paths.NewResolver(projectDir)
	lockMgr := lock.NewManager(resolver.LocksDir())
	fsm := session.NewFSM()

	// Acquire exclusive lock on parent session
	parentLock, err := lockMgr.Acquire(parentID, lock.Exclusive, lock.DefaultTimeout, "ari-session-fray")
	if err != nil {
		return nil, err
	}
	defer func() { _ = parentLock.Release() }()
	emitLockEvent(resolver, parentID, "ari-session-fray")

	// Load parent context
	parentCtxPath := resolver.SessionContextFile(parentID)
	parentCtx, err := session.LoadContext(parentCtxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.ErrSessionNotFound(parentID)
		}
		return nil, err
	}

	// Validate parent can transition to PARKED (fray parks the parent)
	if err := fsm.ValidateTransition(parentCtx.Status, session.StatusParked); err != nil {
		return nil, err
	}

	// Create child session inheriting parent's attributes
	childCtx := session.NewContext(parentCtx.Initiative, parentCtx.Complexity, parentCtx.ActiveRite)
	childCtx.SchemaVersion = "2.3"
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
		_ = os.RemoveAll(childDir)
		return nil, err
	}

	// Park parent
	now := time.Now().UTC()
	parentCtx.Status = session.StatusParked
	parentCtx.ParkedAt = &now
	parentCtx.ParkedReason = fmt.Sprintf("Frayed to %s", childCtx.SessionID)
	parentCtx.ParkSource = "fray"
	parentCtx.Strands = append(parentCtx.Strands, session.Strand{
		SessionID: childCtx.SessionID,
		Status:    "ACTIVE",
	})

	if err := parentCtx.Save(parentCtxPath); err != nil {
		_ = os.RemoveAll(childDir)
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

// emitFrayEvents emits clew events for a fray operation.
// All event emissions are non-fatal — failures are silently ignored.
func emitFrayEvents(resolver *paths.Resolver, parentID, childID, frayPoint string) {
	// Emit clew session_frayed event on parent
	parentDir := resolver.SessionDir(parentID)
	frayWriter := clewcontract.NewBufferedEventWriter(parentDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = frayWriter.Close() }()
	frayWriter.Write(clewcontract.NewSessionFrayedEvent(parentID, childID, frayPoint))
	_ = frayWriter.Flush() // Best-effort flush for fray events

	// Emit session_created lifecycle event on child
	childDir := resolver.SessionDir(childID)
	childWriter := clewcontract.NewBufferedEventWriter(childDir, clewcontract.DefaultFlushInterval)
	defer func() { _ = childWriter.Close() }()
	childWriter.Write(clewcontract.NewSessionCreatedEvent(childID, "", "", ""))
	_ = childWriter.Flush()
}
