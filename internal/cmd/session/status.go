package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/lock"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

func newStatusCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show session state",
		Long: `Returns current session state with comprehensive metadata.

Includes session ID, status, initiative, complexity, current phase,
active rite, execution mode, git branch, and White Sails color.
Returns has_session=false if no active session exists.

Examples:
  ari session status
  ari session status -o json
  ari session status -s session-20260105-143000-abc12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx)
		},
	}

	return cmd
}

func runStatus(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()
	lockMgr := ctx.GetLockManager()

	sessionID, err := ctx.GetSessionID()
	if err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err)
		printer.PrintError(err)
		return err
	}

	// If no session, return has_session: false
	if sessionID == "" {
		result := output.StatusOutput{
			HasSession: false,
			Status:     "NONE",
		}
		return printer.Print(result)
	}

	// Check if session directory exists
	sessionDir := resolver.SessionDir(sessionID)
	ctxPath := resolver.SessionContextFile(sessionID)

	if _, err := os.Stat(ctxPath); os.IsNotExist(err) {
		// Session ID set but no context file
		result := output.StatusOutput{
			SessionID:  sessionID,
			HasSession: false,
			Status:     "NONE",
		}
		return printer.Print(result)
	}

	// Acquire shared lock for consistent read
	sessionLock, err := lockMgr.Acquire(sessionID, lock.Shared, lock.DefaultTimeout, "ari-session-status")
	if err != nil {
		// Non-fatal - continue without lock
		printer.VerboseLog("warn", "failed to acquire lock", map[string]interface{}{"error": err.Error()})
	} else {
		defer sessionLock.Release()
	}

	// Load session context
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Determine execution mode
	executionMode := deriveExecutionMode(sessCtx, ctx.getActiveRite())

	// Get git info
	gitBranch, gitChanges := getGitInfo()

	// Load WHITE_SAILS.yaml if exists
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	var sailsColor, sailsBase string
	if data, err := os.ReadFile(sailsPath); err == nil {
		var sailsData struct {
			Color        string `yaml:"color"`
			ComputedBase string `yaml:"computed_base"`
		}
		if yaml.Unmarshal(data, &sailsData) == nil {
			sailsColor = sailsData.Color
			sailsBase = sailsData.ComputedBase
		}
	}

	// Build result
	result := output.StatusOutput{
		SessionID:     sessionID,
		SessionDir:    sessionDir,
		HasSession:    true,
		Status:        string(sessCtx.Status),
		Initiative:    sessCtx.Initiative,
		Complexity:    sessCtx.Complexity,
		CurrentPhase:  sessCtx.CurrentPhase,
		ActiveRite:    sessCtx.ActiveRite,
		ExecutionMode: executionMode,
		CreatedAt:     sessCtx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		SchemaVersion: sessCtx.SchemaVersion,
		GitBranch:     gitBranch,
		GitChanges:    gitChanges,
		SailsColor:    sailsColor,
		SailsBase:     sailsBase,
	}

	return printer.Print(result)
}

// deriveExecutionMode determines the execution mode based on session and rite state.
func deriveExecutionMode(ctx *session.Context, activeRite string) string {
	// No session = native
	if ctx == nil {
		return "native"
	}

	// Parked sessions are cross-cutting
	if ctx.Status == session.StatusParked {
		return "cross-cutting"
	}

	// Archived sessions are native (not active)
	if ctx.Status == session.StatusArchived {
		return "native"
	}

	// Check rite configuration
	if activeRite == "" || activeRite == "none" {
		return "cross-cutting"
	}

	// Active session with rite = orchestrated
	return "orchestrated"
}

// getGitInfo returns the current git branch and number of changes.
func getGitInfo() (string, int) {
	// Get branch
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return "not a git repo", 0
	}
	branch := strings.TrimSpace(string(branchOut))

	// Get change count
	statusCmd := exec.Command("git", "status", "--short")
	statusOut, err := statusCmd.Output()
	if err != nil {
		return branch, 0
	}

	changes := 0
	for _, line := range strings.Split(string(statusOut), "\n") {
		if strings.TrimSpace(line) != "" {
			changes++
		}
	}

	return branch, changes
}
