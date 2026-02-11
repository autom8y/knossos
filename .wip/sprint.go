package session

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// SprintOutput represents sprint command result.
type SprintOutput struct {
	SprintID  string `json:"sprint_id"`
	SessionID string `json:"session_id"`
	Goal      string `json:"goal"`
	Status    string `json:"status"`
	Action    string `json:"action"`
}

// Text implements Textable for SprintOutput.
func (s SprintOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Sprint %s: %s\n", s.Action, s.SprintID))
	b.WriteString(fmt.Sprintf("Session: %s\n", s.SessionID))
	b.WriteString(fmt.Sprintf("Goal: %s\n", s.Goal))
	b.WriteString(fmt.Sprintf("Status: %s\n", s.Status))
	return b.String()
}

func newSprintCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sprint",
		Short: "Manage sprints within a session",
		Long: `Create, complete, and delete sprints within the current session.

Sprints are stored in named directories under the session's sprints/ folder.
Each sprint has a SPRINT_CONTEXT.md with YAML frontmatter tracking goals and tasks.

Examples:
  ari session sprint create "Implement auth" --task "Login API" --task "Session mgmt"
  ari session sprint mark-complete <sprint-id>
  ari session sprint delete <sprint-id>`,
	}

	cmd.AddCommand(newSprintCreateCmd(ctx))
	cmd.AddCommand(newSprintMarkCompleteCmd(ctx))
	cmd.AddCommand(newSprintDeleteCmd(ctx))

	return cmd
}

func newSprintCreateCmd(ctx *cmdContext) *cobra.Command {
	var tasks []string

	cmd := &cobra.Command{
		Use:   "create <goal>",
		Short: "Create a new sprint",
		Long: `Create a new sprint within the current session.

The sprint starts in ACTIVE status. Tasks can be provided with --task flags.

Examples:
  ari session sprint create "Implement auth"
  ari session sprint create "Sprint 1" --task "Login API" --task "Session mgmt"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSprintCreate(ctx, args[0], tasks)
		},
	}

	cmd.Flags().StringArrayVar(&tasks, "task", nil, "Task description (repeatable)")

	return cmd
}

func runSprintCreate(ctx *cmdContext, goal string, tasks []string) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Get current session
	sessionID, err := ctx.GetSessionID()
	if err != nil || sessionID == "" {
		err := errors.New(errors.CodeSessionNotFound, "no active session found")
		printer.PrintError(err)
		return err
	}

	// Load session context to verify it's ACTIVE
	sessCtx, err := session.LoadContext(resolver.SessionContextFile(sessionID))
	if err != nil {
		printer.PrintError(err)
		return err
	}
	if sessCtx.Status != session.StatusActive {
		err := errors.New(errors.CodeLifecycleViolation,
			fmt.Sprintf("session is %s, must be ACTIVE to create sprint", sessCtx.Status))
		printer.PrintError(err)
		return err
	}

	// Check for existing active sprint
	existingID, err := findSprintByStatus(resolver, sessionID, session.SprintStatusActive)
	if err == nil && existingID != "" {
		err := errors.New(errors.CodeLifecycleViolation,
			fmt.Sprintf("active sprint already exists: %s", existingID))
		printer.PrintError(err)
		return err
	}

	// Create sprint
	sprintCtx := session.NewSprintContext(sessionID, goal, tasks)

	// Create sprint directory
	sprintDir := resolver.SprintDir(sessionID, sprintCtx.SprintID)
	if err := paths.EnsureDir(sprintDir); err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to create sprint directory", err)
		printer.PrintError(err)
		return err
	}

	// Save SPRINT_CONTEXT.md
	ctxPath := resolver.SprintContextFile(sessionID, sprintCtx.SprintID)
	if err := sprintCtx.Save(ctxPath); err != nil {
		os.RemoveAll(sprintDir)
		printer.PrintError(err)
		return err
	}

	// Emit event (non-fatal)
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitSprintCreated(sessionID, sprintCtx.SprintID, goal); err != nil {
		printer.VerboseLog("warn", "failed to emit sprint created event", map[string]interface{}{"error": err.Error()})
	}

	result := SprintOutput{
		SprintID:  sprintCtx.SprintID,
		SessionID: sessionID,
		Goal:      goal,
		Status:    string(sprintCtx.Status),
		Action:    "created",
	}

	return printer.Print(result)
}

func newSprintMarkCompleteCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-complete [sprint-id]",
		Short: "Mark a sprint as completed",
		Long: `Mark the specified or current active sprint as completed.

If no sprint ID is provided, finds and completes the current ACTIVE sprint.

Examples:
  ari session sprint mark-complete
  ari session sprint mark-complete sprint-20260208-120000-abcdef01`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sprintID := ""
			if len(args) > 0 {
				sprintID = args[0]
			}
			return runSprintMarkComplete(ctx, sprintID)
		},
	}

	return cmd
}

func runSprintMarkComplete(ctx *cmdContext, sprintID string) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	sessionID, err := ctx.GetSessionID()
	if err != nil || sessionID == "" {
		err := errors.New(errors.CodeSessionNotFound, "no active session found")
		printer.PrintError(err)
		return err
	}

	// Resolve sprint ID
	if sprintID == "" {
		found, err := findSprintByStatus(resolver, sessionID, session.SprintStatusActive)
		if err != nil || found == "" {
			err := errors.New(errors.CodeFileNotFound, "no active sprint found")
			printer.PrintError(err)
			return err
		}
		sprintID = found
	}

	// Load sprint context
	ctxPath := resolver.SprintContextFile(sessionID, sprintID)
	sprintCtx, err := session.LoadSprintContext(ctxPath)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	if sprintCtx.Status != session.SprintStatusActive {
		err := errors.New(errors.CodeLifecycleViolation,
			fmt.Sprintf("sprint is %s, must be ACTIVE to mark complete", sprintCtx.Status))
		printer.PrintError(err)
		return err
	}

	// Transition to COMPLETED
	now := time.Now().UTC()
	sprintCtx.Status = session.SprintStatusCompleted
	sprintCtx.CompletedAt = &now

	if err := sprintCtx.Save(ctxPath); err != nil {
		printer.PrintError(err)
		return err
	}

	// Emit event (non-fatal)
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitSprintCompleted(sessionID, sprintID); err != nil {
		printer.VerboseLog("warn", "failed to emit sprint completed event", map[string]interface{}{"error": err.Error()})
	}

	result := SprintOutput{
		SprintID:  sprintID,
		SessionID: sessionID,
		Goal:      sprintCtx.Goal,
		Status:    string(sprintCtx.Status),
		Action:    "completed",
	}

	return printer.Print(result)
}

func newSprintDeleteCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <sprint-id>",
		Short: "Delete a sprint",
		Long: `Delete a sprint directory and its SPRINT_CONTEXT.md.

The sprint must not be ACTIVE. Complete it first with mark-complete.

Examples:
  ari session sprint delete sprint-20260208-120000-abcdef01`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSprintDelete(ctx, args[0])
		},
	}

	return cmd
}

func runSprintDelete(ctx *cmdContext, sprintID string) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	sessionID, err := ctx.GetSessionID()
	if err != nil || sessionID == "" {
		err := errors.New(errors.CodeSessionNotFound, "no active session found")
		printer.PrintError(err)
		return err
	}

	// Load sprint context to validate
	ctxPath := resolver.SprintContextFile(sessionID, sprintID)
	sprintCtx, err := session.LoadSprintContext(ctxPath)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	if sprintCtx.Status == session.SprintStatusActive {
		err := errors.New(errors.CodeLifecycleViolation,
			"cannot delete ACTIVE sprint; use mark-complete first")
		printer.PrintError(err)
		return err
	}

	// Delete sprint directory
	sprintDir := resolver.SprintDir(sessionID, sprintID)
	if err := os.RemoveAll(sprintDir); err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to delete sprint directory", err)
		printer.PrintError(err)
		return err
	}

	// Emit event (non-fatal)
	emitter := ctx.getEventEmitter(sessionID)
	if err := emitter.EmitSprintDeleted(sessionID, sprintID); err != nil {
		printer.VerboseLog("warn", "failed to emit sprint deleted event", map[string]interface{}{"error": err.Error()})
	}

	result := SprintOutput{
		SprintID:  sprintID,
		SessionID: sessionID,
		Goal:      sprintCtx.Goal,
		Status:    "DELETED",
		Action:    "deleted",
	}

	return printer.Print(result)
}

// findSprintByStatus scans the sprints directory for a sprint with the given status.
func findSprintByStatus(resolver *paths.Resolver, sessionID string, status session.SprintStatus) (string, error) {
	sprintsDir := resolver.SprintsDir(sessionID)
	entries, err := os.ReadDir(sprintsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ctxPath := resolver.SprintContextFile(sessionID, entry.Name())
		sprintCtx, err := session.LoadSprintContext(ctxPath)
		if err != nil {
			continue
		}
		if sprintCtx.Status == status {
			return entry.Name(), nil
		}
	}

	return "", nil
}
