package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/output"
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
It re-materializes .claude/ without the model override, restoring
original agent models and settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := output.NewPrinter(output.FormatJSON, nil, nil, false)

			// Resolve project directory
			projectDir, _ := os.Getwd()
			if ctx.ProjectDir != nil && *ctx.ProjectDir != "" {
				projectDir = *ctx.ProjectDir
			}

			// Check for el-cheapo marker
			markerPath := filepath.Join(projectDir, ".claude", ".el-cheapo-active")
			if _, err := os.Stat(markerPath); os.IsNotExist(err) {
				out := CheapoRevertOutput{
					Status:  "skipped",
					Message: "no el-cheapo marker found",
				}
				return printer.Print(out)
			}

			// Run normal rite-scope sync (without el-cheapo) to revert
			resolver := paths.NewResolver(projectDir)
			m := materialize.NewMaterializer(resolver)

			// Wire embedded assets
			if embRites := common.EmbeddedRites(); embRites != nil {
				m.WithEmbeddedFS(embRites)
			}
			if embTemplates := common.EmbeddedTemplates(); embTemplates != nil {
				m.WithEmbeddedTemplates(embTemplates)
			}
			if embAgents := common.EmbeddedAgents(); embAgents != nil {
				m.WithEmbeddedAgents(embAgents)
			}
			if embMena := common.EmbeddedMena(); embMena != nil {
				m.WithEmbeddedMena(embMena)
			}

			// Sync with rite scope only (no el-cheapo) — this reverts everything
			opts := materialize.SyncOptions{
				Scope: materialize.ScopeRite,
			}
			_, err := m.Sync(opts)
			if err != nil {
				// Output error as JSON for CC hook consumption
				errOut := map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("revert sync failed: %v", err),
				}
				data, _ := json.Marshal(errOut)
				fmt.Println(string(data))
				return nil // Don't fail the hook — graceful degradation
			}

			out := CheapoRevertOutput{
				Status:  "reverted",
				Message: "el-cheapo model override reverted",
			}
			return printer.Print(out)
		},
	}

	return cmd
}
