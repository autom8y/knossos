// Package hook implements the ari hook commands.
package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"github.com/autom8y/knossos/internal/session"
)

// AutoparkOutput represents the output of the autopark hook.
type AutoparkOutput struct {
	SessionID       string   `json:"session_id,omitempty"`
	Status          string   `json:"status,omitempty"`
	PreviousStatus  string   `json:"previous_status,omitempty"`
	AutoParkedAt    string   `json:"auto_parked_at,omitempty"`
	WasParked       bool     `json:"was_parked"`
	Message         string   `json:"message,omitempty"`
	DismissedAgents []string `json:"dismissed_agents,omitempty"`
}

// Text implements output.Textable for text output.
func (a AutoparkOutput) Text() string {
	var parts []string
	if a.WasParked {
		parts = append(parts, "Session auto-parked: "+a.SessionID)
	} else if a.Message != "" {
		parts = append(parts, a.Message)
	}
	if len(a.DismissedAgents) > 0 {
		parts = append(parts, "Dismissed zombie agents: "+strings.Join(a.DismissedAgents, ", "))
	}
	return strings.Join(parts, "\n")
}

// newAutoparkCmd creates the autopark hook subcommand.
func newAutoparkCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autopark",
		Short: "Auto-park session on Stop event",
		Long: `Automatically transitions active sessions to PARKED on Stop event.

This hook is triggered on Stop events. It:
- Checks for an active session
- Transitions from ACTIVE to PARKED if applicable
- Records auto_parked_at timestamp
- Gracefully no-ops if no active session

Performance: <100ms target execution time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runAutopark(cmd, ctx)
			})
		},
	}

	return cmd
}

func runAutopark(cmd *cobra.Command, ctx *cmdContext) error {
	printer := ctx.getPrinter()
	return runAutoparkCore(cmd, ctx, printer)
}

// runAutoparkCore contains the actual logic with injected printer for testing.
func runAutoparkCore(cmd *cobra.Command, ctx *cmdContext, printer *output.Printer) error {
	// Get hook environment
	hookEnv := ctx.getHookEnv(cmd)

	// Authentication Check: Verify signature of raw payload
	if !hook.Verify(hookEnv.RawPayload, hookEnv.Signature) {
		return printer.Print(hook.OutputDenyAuth())
	}

	// Verify this is a Stop event (or allow for testing without event)
	if hookEnv.Event != "" && hookEnv.Event != hook.EventStop {
		printer.VerboseLog("debug", "skipping autopark hook for non-Stop event",
			map[string]any{"event": string(hookEnv.Event)})
		return outputNoPark(printer, "not a stop event", nil)
	}

	// Always dismiss zombie summoned agents on every Stop event, regardless of session state.
	// This is a safety net for dromenon closures that fail to dismiss summoned agents.
	dismissed := dismissZombieSummonedAgents(printer)

	// Resolve session context
	resolver, sessionID, err := ctx.resolveSession(hookEnv)
	if err != nil {
		printer.VerboseLog("warn", "failed to read current session", map[string]any{"error": err.Error()})
		return outputNoPark(printer, "no active session", dismissed)
	}

	if resolver.ProjectRoot() == "" || sessionID == "" {
		return outputNoPark(printer, "no active session", dismissed)
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		printer.VerboseLog("warn", "failed to load session context",
			map[string]any{"session_id": sessionID, "error": err.Error()})
		return outputNoPark(printer, "could not load session", dismissed)
	}

	// Only park if currently ACTIVE
	if sessCtx.Status != session.StatusActive {
		return outputNoPark(printer, "session not active (status: "+string(sessCtx.Status)+")", dismissed)
	}

	// Validate transition using FSM
	fsm := session.NewFSM()
	if !fsm.CanTransition(sessCtx.Status, session.StatusParked) {
		return outputNoPark(printer, "invalid transition from "+string(sessCtx.Status), dismissed)
	}

	// Record previous status
	previousStatus := sessCtx.Status

	// Update session state
	now := time.Now().UTC()
	sessCtx.Status = session.StatusParked
	sessCtx.ParkedAt = &now
	sessCtx.ParkedReason = "auto-parked on Stop"
	sessCtx.ParkSource = "auto"

	// Save session context
	if err := sessCtx.Save(ctxPath); err != nil {
		printer.VerboseLog("error", "failed to save session context",
			map[string]any{"session_id": sessionID, "error": err.Error()})
		return outputNoPark(printer, "failed to save session", dismissed)
	}

	// Log git status for audit purposes
	gitStatus := getGitStatusQuick()
	printer.VerboseLog("info", "session auto-parked", map[string]any{
		"session_id": sessionID,
		"git_status": gitStatus,
	})

	// Build output
	result := AutoparkOutput{
		SessionID:       sessionID,
		Status:          string(session.StatusParked),
		PreviousStatus:  string(previousStatus),
		AutoParkedAt:    now.Format(time.RFC3339),
		WasParked:       true,
		Message:         "Session auto-parked",
		DismissedAgents: dismissed,
	}

	return printer.Print(result)
}

// outputNoPark outputs a no-op response when parking didn't occur.
// dismissed carries any zombie agents removed during this Stop event.
func outputNoPark(printer *output.Printer, reason string, dismissed []string) error {
	result := AutoparkOutput{
		WasParked:       false,
		Message:         reason,
		DismissedAgents: dismissed,
	}
	return printer.Print(result)
}

// dismissZombieSummonedAgents scans USER_PROVENANCE_MANIFEST.yaml for entries whose
// SourcePath begins with "summon:" and removes their agent files and manifest entries.
// This is a safety net for dromenon closures that fail to call 'ari agent dismiss'.
// Returns the names of dismissed agents (may be nil if none found or on error).
func dismissZombieSummonedAgents(printer *output.Printer) []string {
	userChannelDir, err := paths.UserChannelDir("claude")
	if err != nil {
		printer.VerboseLog("warn", "zombie dismiss: failed to resolve user channel dir",
			map[string]any{"error": err.Error()})
		return nil
	}

	manifestPath := provenance.UserManifestPath(userChannelDir)
	manifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		printer.VerboseLog("warn", "zombie dismiss: failed to load user provenance",
			map[string]any{"error": err.Error()})
		return nil
	}

	// Collect keys for summon:* entries (agents installed via 'ari agent summon').
	// Sprint 1 stores these with SourcePath = "summon:{name}" (see dismiss.go L98).
	var toRemove []string
	for key, entry := range manifest.Entries {
		if strings.HasPrefix(entry.SourcePath, "summon:") {
			toRemove = append(toRemove, key)
		}
	}

	if len(toRemove) == 0 {
		return nil
	}

	// Remove agent files and manifest entries.
	var dismissed []string
	for _, key := range toRemove {
		agentPath := filepath.Join(userChannelDir, key)
		if err := os.Remove(agentPath); err != nil && !os.IsNotExist(err) {
			printer.VerboseLog("warn", "zombie dismiss: failed to remove agent file",
				map[string]any{"path": agentPath, "error": err.Error()})
			// Continue: remove the manifest entry even if the file is gone or unremovable.
		}
		delete(manifest.Entries, key)
		// Extract agent name from key: "agents/{name}.md" → "{name}"
		name := strings.TrimSuffix(filepath.Base(key), ".md")
		dismissed = append(dismissed, name)
	}

	// Persist the updated manifest.
	manifest.LastSync = time.Now().UTC()
	if err := provenance.Save(manifestPath, manifest); err != nil {
		printer.VerboseLog("warn", "zombie dismiss: failed to save manifest",
			map[string]any{"error": err.Error()})
	}

	return dismissed
}

// getGitStatusQuick returns a quick git status for logging.
func getGitStatusQuick() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "status", "--short")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	if strings.TrimSpace(string(out)) == "" {
		return "clean"
	}
	return "uncommitted"
}
