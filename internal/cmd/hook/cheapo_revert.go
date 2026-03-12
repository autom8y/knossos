package hook

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/hook"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/paths"
)

// CheapoRevertOutput represents the output of the cheapo-revert hook.
type CheapoRevertOutput struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Text implements output.Textable.
func (c CheapoRevertOutput) Text() string {
	return c.Message
}

// newCheapoRevertCmd creates the cheapo-revert hook subcommand.
func newCheapoRevertCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cheapo-revert",
		Short: "Revert el-cheapo model override on session exit",
		Long: `Reverts the el-cheapo model override by running a normal rite-scope sync.

This hook is triggered on Stop events when el-cheapo mode was active.
It re-materializes the channel directory without the model override,
restoring original agent models and settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runCheapoRevert(ctx)
			})
		},
	}

	return cmd
}

// runCheapoRevert implements the cheapo-revert hook with proper timeout and event guard.
func runCheapoRevert(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	hookEnv := ctx.getHookEnv()

	// Event guard: only process Stop events (or empty for direct CLI/test invocation).
	if hookEnv.Event != "" && hookEnv.Event != hook.EventStop {
		return printer.Print(CheapoRevertOutput{
			Status:  "skipped",
			Message: "not a Stop event",
		})
	}

	// Resolve project directory
	projectDir, _ := os.Getwd()
	if ctx.ProjectDir != nil && *ctx.ProjectDir != "" {
		projectDir = *ctx.ProjectDir
	}

	// Check for el-cheapo marker
	markerPath := filepath.Join(projectDir, ".knossos", ".el-cheapo-active")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		return printer.Print(CheapoRevertOutput{
			Status:  "skipped",
			Message: "no el-cheapo marker found",
		})
	}

	// Run normal rite-scope sync (without el-cheapo) to revert
	resolver := paths.NewResolver(projectDir)
	m := NewWiredMaterializer(resolver)

	// Sync with rite scope only (no el-cheapo) — this reverts everything
	opts := materialize.SyncOptions{
		Scope: materialize.ScopeRite,
	}
	_, err := m.Sync(opts)
	if err != nil {
		// Graceful degradation — don't fail the hook
		return printer.Print(CheapoRevertOutput{
			Status:  "error",
			Message: fmt.Sprintf("revert sync failed: %v", err),
		})
	}

	return printer.Print(CheapoRevertOutput{
		Status:  "reverted",
		Message: "el-cheapo model override reverted",
	})
}
