package inscription

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

func newRollbackCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback [timestamp]",
		Short: "Restore CLAUDE.md from a backup",
		Long: `Restore CLAUDE.md from a previous backup.

If no timestamp is provided, restores from the most recent backup.
Use 'ari inscription backups' to list available backups.

The timestamp format matches the backup filename:
  YYYY-MM-DDTHH-MM-SSZ (e.g., 2026-01-06T10-30-00Z)

You can provide a partial timestamp to match:
  - Full: 2026-01-06T10-30-00Z
  - Date only: 2026-01-06
  - Partial: 2026-01-06T10

Examples:
  ari inscription rollback                    # Restore most recent backup
  ari inscription rollback 2026-01-06T10-30   # Restore specific backup`,
		Args: common.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			timestamp := ""
			if len(args) > 0 {
				timestamp = args[0]
			}
			return runRollback(ctx, timestamp)
		},
	}

	return cmd
}

func runRollback(ctx *cmdContext, timestamp string) error {
	printer := ctx.getPrinter()
	pipeline := ctx.getPipeline()

	// Get backup info before rollback
	backups, err := pipeline.ListBackups()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	if len(backups) == 0 {
		printer.PrintLine("No backups available")
		return nil
	}

	// Find the backup we'll restore
	var restoringFrom string
	if timestamp == "" {
		restoringFrom = backups[0].Name
	} else {
		for _, b := range backups {
			if contains(b.Name, timestamp) {
				restoringFrom = b.Name
				break
			}
		}
	}

	// Perform rollback
	err = pipeline.Rollback(timestamp)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	out := RollbackOutput{
		Success:      true,
		RestoredFrom: restoringFrom,
	}

	return printer.Print(out)
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RollbackOutput represents rollback result for output.
type RollbackOutput struct {
	Success      bool   `json:"success"`
	RestoredFrom string `json:"restored_from"`
}

// Text implements output.Textable for RollbackOutput.
func (r RollbackOutput) Text() string {
	if r.Success {
		return fmt.Sprintf("Restored CLAUDE.md from backup: %s\n", r.RestoredFrom)
	}
	return "Rollback failed\n"
}
