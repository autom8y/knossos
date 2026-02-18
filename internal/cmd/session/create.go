package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/hook/clewcontract"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

type createOptions struct {
	complexity string
	rite       string
	seed       bool
	seedPrefix string
	seedKeep   bool
}

func newCreateCmd(ctx *cmdContext) *cobra.Command {
	var opts createOptions

	cmd := &cobra.Command{
		Use:   "create <initiative>",
		Short: "Create a new session",
		Long: `Create a new session, transitioning from NONE to ACTIVE state.

The initiative argument is a short description of the work being done.
Complexity defaults to MODULE if not specified.

Seed mode (--seed) creates the session in an ephemeral worktree, immediately
parks it, and copies it to the main repo. This enables parallel session
preparation without hitting the single-session-per-terminal constraint.

Examples:
  ari session create "user-auth feature"
  ari session create "deploy pipeline" -c SYSTEM -r sre
  ari session create "hotfix login" -c PATCH
  ari session create "parallel work" --seed`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(ctx, args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.complexity, "complexity", "c", "MODULE",
		"Complexity level: PATCH, MODULE, SYSTEM, INITIATIVE, MIGRATION")
	cmd.Flags().StringVarP(&opts.rite, "rite", "r", "",
		"Rite (practice bundle) to activate (default: from ACTIVE_RITE)")
	cmd.Flags().BoolVar(&opts.seed, "seed", false,
		"Create session in ephemeral worktree, park it, and copy to main repo")
	cmd.Flags().StringVar(&opts.seedPrefix, "seed-prefix", "/tmp/knossos-seed-",
		"Custom prefix for ephemeral worktree path")
	cmd.Flags().BoolVar(&opts.seedKeep, "seed-keep", false,
		"Keep worktree after seeding (for debugging)")

	return cmd
}

func runCreate(ctx *cmdContext, initiative string, opts createOptions) error {
	// If seed mode is enabled, delegate to seeded creation flow
	if opts.seed {
		return runCreateSeeded(ctx, initiative, opts)
	}

	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	// Get rite from flag or ACTIVE_RITE file
	rite := opts.rite
	if rite == "" {
		rite = ctx.getActiveRite()
	}

	// Validate complexity
	if !isValidComplexity(opts.complexity) {
		err := errors.New(errors.CodeUsageError, "invalid complexity: must be PATCH, MODULE, SYSTEM, INITIATIVE, or MIGRATION")
		printer.PrintError(err)
		return err
	}

	// Ensure sessions directory exists
	if err := paths.EnsureDir(resolver.SessionsDir()); err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to create sessions directory", err)
		printer.PrintError(err)
		return err
	}

	// Acquire exclusive lock for creation (using a special "create" lock)
	createLock, err := lockMgr.Acquire("__create__", lock.Exclusive, lock.DefaultTimeout, "ari-session-create")
	if err != nil {
		printer.PrintError(err)
		return err
	}
	defer createLock.Release()

	// Check for existing session
	currentID, err := session.FindActiveSession(resolver.SessionsDir())
	if err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to read current session", err)
		printer.PrintError(err)
		return err
	}

	if currentID != "" {
		// Check if session directory exists
		if _, err := os.Stat(resolver.SessionDir(currentID)); err == nil {
			// Load session to check status
			ctxPath := resolver.SessionContextFile(currentID)
			existingCtx, err := session.LoadContext(ctxPath)
			if err == nil {
				err := errors.ErrSessionExists(currentID, string(existingCtx.Status))
				printer.PrintError(err)
				return err
			}
		}
	}

	// Validate FSM transition (defense-in-depth: NONE -> ACTIVE)
	fsm := session.NewFSM()
	if err := fsm.ValidateTransition(session.StatusNone, session.StatusActive); err != nil {
		printer.PrintError(err)
		return err
	}

	// Create new session context
	newCtx := session.NewContext(initiative, opts.complexity, rite)
	sessionDir := resolver.SessionDir(newCtx.SessionID)

	// Create session directory
	if err := paths.EnsureDir(sessionDir); err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to create session directory", err)
		printer.PrintError(err)
		return err
	}

	// Save context
	ctxPath := resolver.SessionContextFile(newCtx.SessionID)
	if err := newCtx.Save(ctxPath); err != nil {
		// Cleanup on failure
		os.RemoveAll(sessionDir)
		printer.PrintError(err)
		return err
	}

	// Emit lifecycle event
	writer := clewcontract.NewBufferedEventWriter(sessionDir, clewcontract.DefaultFlushInterval)
	defer writer.Close()
	writer.Write(clewcontract.NewSessionCreatedEvent(newCtx.SessionID, initiative, opts.complexity, rite))
	if err := writer.Flush(); err != nil {
		printer.VerboseLog("warn", "failed to write event", map[string]interface{}{"error": err.Error()})
	}

	// Output result
	result := output.CreateOutput{
		SessionID:     newCtx.SessionID,
		SessionDir:    sessionDir,
		Status:        string(newCtx.Status),
		Initiative:    initiative,
		Complexity:    opts.complexity,
		Rite:          rite,
		CreatedAt:     newCtx.CreatedAt.Format(time.RFC3339),
		SchemaVersion: newCtx.SchemaVersion,
	}

	return printer.Print(result)
}

// runCreateSeeded creates a session in an ephemeral worktree, parks it immediately,
// and copies it to the main repo's sessions directory. This enables parallel session
// preparation without hitting the single-session-per-terminal constraint.
func runCreateSeeded(ctx *cmdContext, initiative string, opts createOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Get rite from flag or ACTIVE_RITE file
	rite := opts.rite
	if rite == "" {
		rite = ctx.getActiveRite()
	}

	// Validate complexity
	if !isValidComplexity(opts.complexity) {
		err := errors.New(errors.CodeUsageError, "invalid complexity: must be PATCH, MODULE, SYSTEM, INITIATIVE, or MIGRATION")
		printer.PrintError(err)
		return err
	}

	// Generate unique worktree path
	worktreePath := fmt.Sprintf("%s%d-%d", opts.seedPrefix, time.Now().Unix(), os.Getpid())

	// Get the main repo's project root for later use
	mainProjectRoot := resolver.ProjectRoot()

	// Create ephemeral worktree
	printer.VerboseLog("info", "creating ephemeral worktree", map[string]interface{}{"path": worktreePath})
	if err := createWorktree(worktreePath); err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to create worktree", err)
		printer.PrintError(err)
		return err
	}

	// Ensure cleanup of worktree unless --seed-keep is set
	cleanupWorktree := func() {
		if opts.seedKeep {
			printer.VerboseLog("info", "keeping worktree for debugging", map[string]interface{}{"path": worktreePath})
			return
		}
		printer.VerboseLog("info", "removing ephemeral worktree", map[string]interface{}{"path": worktreePath})
		removeWorktree(worktreePath)
	}

	// Create a resolver for the worktree
	worktreeResolver := paths.NewResolver(worktreePath)

	// Ensure sessions directory exists in worktree
	if err := paths.EnsureDir(worktreeResolver.SessionsDir()); err != nil {
		cleanupWorktree()
		err := errors.Wrap(errors.CodeGeneralError, "failed to create sessions directory in worktree", err)
		printer.PrintError(err)
		return err
	}

	// Ensure locks directory exists in worktree
	if err := paths.EnsureDir(worktreeResolver.LocksDir()); err != nil {
		cleanupWorktree()
		err := errors.Wrap(errors.CodeGeneralError, "failed to create locks directory in worktree", err)
		printer.PrintError(err)
		return err
	}

	// Create new session context (starts as ACTIVE)
	newCtx := session.NewContext(initiative, opts.complexity, rite)
	worktreeSessionDir := worktreeResolver.SessionDir(newCtx.SessionID)

	// Create session directory in worktree
	if err := paths.EnsureDir(worktreeSessionDir); err != nil {
		cleanupWorktree()
		err := errors.Wrap(errors.CodeGeneralError, "failed to create session directory in worktree", err)
		printer.PrintError(err)
		return err
	}

	// Immediately transition to PARKED status (seeded sessions start parked)
	now := time.Now().UTC()
	newCtx.Status = session.StatusParked
	newCtx.ParkedAt = &now
	newCtx.ParkedReason = "Seeded for parallel execution"

	// Save context in worktree
	worktreeCtxPath := worktreeResolver.SessionContextFile(newCtx.SessionID)
	if err := newCtx.Save(worktreeCtxPath); err != nil {
		os.RemoveAll(worktreeSessionDir)
		cleanupWorktree()
		printer.PrintError(err)
		return err
	}

	// Emit lifecycle events in worktree before copying
	writer := clewcontract.NewBufferedEventWriter(worktreeSessionDir, clewcontract.DefaultFlushInterval)
	defer writer.Close()
	writer.Write(clewcontract.NewSessionCreatedEvent(newCtx.SessionID, initiative, opts.complexity, rite))
	writer.Write(clewcontract.NewSessionParkedEvent(newCtx.SessionID, "seeded creation"))
	if err := writer.Flush(); err != nil {
		printer.VerboseLog("warn", "failed to write seeded events", map[string]interface{}{"error": err.Error()})
	}

	// Copy session from worktree to main repo
	mainSessionsDir := resolver.SessionsDir()
	if err := paths.EnsureDir(mainSessionsDir); err != nil {
		cleanupWorktree()
		err := errors.Wrap(errors.CodeGeneralError, "failed to ensure main sessions directory", err)
		printer.PrintError(err)
		return err
	}

	mainSessionDir := resolver.SessionDir(newCtx.SessionID)
	printer.VerboseLog("info", "copying session to main repo", map[string]interface{}{
		"from": worktreeSessionDir,
		"to":   mainSessionDir,
	})

	if err := copyDir(worktreeSessionDir, mainSessionDir); err != nil {
		// On copy failure, keep worktree for debugging unless explicitly asked to remove
		if !opts.seedKeep {
			printer.VerboseLog("warn", "copy failed, keeping worktree for debugging", map[string]interface{}{
				"path":  worktreePath,
				"error": err.Error(),
			})
		}
		err := errors.Wrap(errors.CodeGeneralError, "failed to copy session to main repo", err)
		printer.PrintError(err)
		return err
	}

	// Cleanup worktree (respects --seed-keep)
	cleanupWorktree()

	// Output result with seeding information
	result := output.SeedCreateOutput{
		SessionID:   newCtx.SessionID,
		Status:      string(newCtx.Status),
		Seeded:      true,
		SeededTo:    mainSessionDir,
		ParkReason:  newCtx.ParkedReason,
		Initiative:  initiative,
		Complexity:  opts.complexity,
		Rite:        rite,
		CreatedAt:   newCtx.CreatedAt.Format(time.RFC3339),
		ParkedAt:    newCtx.ParkedAt.Format(time.RFC3339),
		ProjectRoot: mainProjectRoot,
	}

	return printer.Print(result)
}

// createWorktree creates an ephemeral git worktree at the given path.
func createWorktree(path string) error {
	cmd := exec.Command("git", "worktree", "add", path, "--detach", "HEAD")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// removeWorktree removes a git worktree.
func removeWorktree(path string) error {
	// First try normal remove
	cmd := exec.Command("git", "worktree", "remove", path)
	if err := cmd.Run(); err != nil {
		// If normal remove fails, try force remove
		cmd = exec.Command("git", "worktree", "remove", path, "--force")
		return cmd.Run()
	}
	return nil
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, content, srcInfo.Mode())
}

func isValidComplexity(c string) bool {
	switch c {
	case "PATCH", "MODULE", "SYSTEM", "INITIATIVE", "MIGRATION":
		return true
	default:
		return false
	}
}
