package worktree

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	ariErrors "github.com/autom8y/knossos/internal/errors"
)

// ImportOutput represents the output of worktree import.
type ImportOutput struct {
	Success      bool   `json:"success"`
	WorktreeID   string `json:"worktree_id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Rite         string `json:"rite"`
	FromArchive  string `json:"from_archive"`
	OriginalID   string `json:"original_id"`
	CreatedAt    string `json:"created_at"`
	Instructions string `json:"instructions"`
}

// Text implements output.Textable for ImportOutput.
func (i ImportOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Imported worktree: %s\n", i.WorktreeID))
	b.WriteString(fmt.Sprintf("  Name: %s\n", i.Name))
	b.WriteString(fmt.Sprintf("  Path: %s\n", i.Path))
	if i.Rite != "" && i.Rite != "none" {
		b.WriteString(fmt.Sprintf("  Rite: %s\n", i.Rite))
	}
	b.WriteString(fmt.Sprintf("  From archive: %s\n", i.FromArchive))
	b.WriteString(fmt.Sprintf("  Original ID: %s\n", i.OriginalID))
	b.WriteString(fmt.Sprintf("\nTo start working: cd %s && claude\n", i.Path))
	return b.String()
}

func newImportCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <archive-path>",
		Short: "Import worktree from archive",
		Long: `Import a worktree from a previously exported tar.gz archive.

The imported worktree will have:
- A new unique ID (preserving original ID in metadata)
- Same name and rite as the exported worktree
- Files restored to the archived git ref
- Ecosystem setup (materialization, rite) applied

Examples:
  ari worktree import ./feature-auth.tar.gz
  ari worktree import ~/backups/worktree.tar.gz`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(ctx, args[0])
		},
	}

	return cmd
}

func runImport(ctx *cmdContext, archivePath string) error {
	printer := ctx.getPrinter()

	mgr, err := ctx.getManager()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Validate archive exists
	if !strings.HasSuffix(archivePath, ".tar.gz") && !strings.HasSuffix(archivePath, ".tgz") {
		err := ariErrors.New(ariErrors.CodeUsageError, "archive must be a .tar.gz or .tgz file")
		printer.PrintError(err)
		return err
	}

	wt, err := mgr.Import(archivePath)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	result := ImportOutput{
		Success:      true,
		WorktreeID:   wt.ID,
		Name:         wt.Name,
		Path:         wt.Path,
		Rite:         wt.Rite,
		FromArchive:  archivePath,
		OriginalID:   wt.FromRef, // FromRef contains the original git ref
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		Instructions: fmt.Sprintf("cd %s && claude", wt.Path),
	}

	return printer.Print(result)
}
