package inscription

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newBackupsCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backups",
		Short: "List available CLAUDE.md backups",
		Long: `List all available CLAUDE.md backups.

Backups are created automatically before each sync operation.
By default, the 5 most recent backups are retained.

Backup files are stored in .knossos/backups/ with timestamp naming:
  CLAUDE.md.YYYY-MM-DDTHH-MM-SSZ

Examples:
  ari inscription backups           # List all backups
  ari inscription backups --json    # JSON output for scripting`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackups(ctx)
		},
	}

	return cmd
}

func runBackups(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	pipeline := ctx.getPipeline()

	backups, err := pipeline.ListBackups()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	out := BackupsListOutput{
		Backups: make([]BackupOutput, len(backups)),
		Total:   len(backups),
	}

	for i, b := range backups {
		out.Backups[i] = BackupOutput{
			Path:      b.Path,
			Name:      b.Name,
			Timestamp: b.Timestamp,
			Size:      b.Size,
		}
	}

	return printer.Print(out)
}

// BackupsListOutput represents backup list for output.
type BackupsListOutput struct {
	Backups []BackupOutput `json:"backups"`
	Total   int            `json:"total"`
}

// Headers implements output.Tabular for BackupsListOutput.
func (b BackupsListOutput) Headers() []string {
	return []string{"TIMESTAMP", "NAME", "SIZE"}
}

// Rows implements output.Tabular for BackupsListOutput.
func (b BackupsListOutput) Rows() [][]string {
	rows := make([][]string, len(b.Backups))
	for i, backup := range b.Backups {
		rows[i] = []string{
			backup.Timestamp.Format(time.RFC3339),
			backup.Name,
			formatSize(backup.Size),
		}
	}
	return rows
}

// Text implements output.Textable for BackupsListOutput.
func (b BackupsListOutput) Text() string {
	if len(b.Backups) == 0 {
		return "No backups available\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available backups (%d):\n\n", b.Total))

	for i, backup := range b.Backups {
		marker := "  "
		if i == 0 {
			marker = "* " // Most recent
		}
		sb.WriteString(fmt.Sprintf("%s%s  %s\n", marker, backup.Timestamp.Format("2006-01-02 15:04:05"), backup.Name))
	}

	sb.WriteString("\n* = most recent (used by 'ari inscription rollback')\n")

	return sb.String()
}

// formatSize formats a byte size for display.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
