package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
)

// ExportOutput represents the output of worktree export.
type ExportOutput struct {
	Success      bool   `json:"success"`
	WorktreeID   string `json:"worktree_id"`
	Name         string `json:"name"`
	ArchivePath  string `json:"archive_path"`
	ArchiveSize  int64  `json:"archive_size"`
	ExportedAt   string `json:"exported_at"`
	Instructions string `json:"instructions"`
}

// Text implements output.Textable for ExportOutput.
func (e ExportOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Exported worktree: %s (%s)\n", e.Name, e.WorktreeID))
	b.WriteString(fmt.Sprintf("  Archive: %s\n", e.ArchivePath))
	b.WriteString(fmt.Sprintf("  Size: %s\n", formatSize(e.ArchiveSize)))
	b.WriteString(fmt.Sprintf("\nTo import: ari worktree import %s\n", e.ArchivePath))
	return b.String()
}

func newExportCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <id-or-name> <target-path>",
		Short: "Export worktree to archive",
		Long: `Export a worktree to a tar.gz archive including all metadata.

The archive includes:
- All worktree files (excluding .git)
- Worktree metadata (rite, complexity, etc.)
- Current git ref for reproducibility

Examples:
  ari worktree export feature-auth ./feature-auth.tar.gz
  ari worktree export wt-20260104-143052-a1b2 ~/backups/worktree.tar.gz`,
		Args: common.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(ctx, args[0], args[1])
		},
	}

	return cmd
}

func runExport(ctx *cmdContext, idOrName, targetPath string) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Get worktree info for output
	status, err := mgr.Status(idOrName)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Ensure .tar.gz extension
	if !strings.HasSuffix(targetPath, ".tar.gz") && !strings.HasSuffix(targetPath, ".tgz") {
		targetPath += ".tar.gz"
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = targetPath
	}

	if err := mgr.Export(idOrName, absPath); err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Get archive size
	var archiveSize int64
	if info, err := os.Stat(absPath); err == nil {
		archiveSize = info.Size()
	}

	result := ExportOutput{
		Success:      true,
		WorktreeID:   status.ID,
		Name:         status.Name,
		ArchivePath:  absPath,
		ArchiveSize:  archiveSize,
		ExportedAt:   time.Now().UTC().Format(time.RFC3339),
		Instructions: fmt.Sprintf("ari worktree import %s", absPath),
	}

	return printer.Print(result)
}

// formatSize formats a byte size into human-readable form.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
