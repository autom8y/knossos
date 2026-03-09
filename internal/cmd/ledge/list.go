package ledge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
)

func newListCmd(ctx *cmdContext) *cobra.Command {
	var shelf bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ledge artifacts",
		Long: `List work product artifacts in the ledge.

By default lists .ledge/{category}/ contents (promotable artifacts).
Use --shelf to list .ledge/shelf/{category}/ contents (promoted artifacts).

Examples:
  ari ledge list           # List promotable artifacts
  ari ledge list --shelf   # List promoted artifacts on shelf`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, shelf)
		},
	}

	cmd.Flags().BoolVar(&shelf, "shelf", false, "List shelf (promoted) artifacts instead")

	return cmd
}

type listEntry struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Path     string `json:"path"`
}

type listOutput struct {
	Entries []listEntry `json:"entries"`
	Shelf   bool        `json:"shelf"`
}

// Text implements output.Textable.
func (o listOutput) Text() string {
	if len(o.Entries) == 0 {
		if o.Shelf {
			return "No artifacts on shelf.\n"
		}
		return "No artifacts in ledge.\n"
	}

	var b strings.Builder
	if o.Shelf {
		b.WriteString("Shelf (.ledge/shelf/):\n")
	} else {
		b.WriteString("Ledge (.ledge/):\n")
	}

	currentCat := ""
	for _, e := range o.Entries {
		if e.Category != currentCat {
			fmt.Fprintf(&b, "\n  %s/\n", e.Category)
			currentCat = e.Category
		}
		fmt.Fprintf(&b, "    %s\n", e.Name)
	}

	return b.String()
}

func runList(ctx *cmdContext, shelf bool) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	baseDir := resolver.LedgeDir()
	if shelf {
		baseDir = resolver.LedgeShelfDir()
	}

	categories := []string{"decisions", "specs", "reviews"}
	if !shelf {
		categories = append(categories, "spikes")
	}

	var entries []listEntry
	for _, cat := range categories {
		catDir := filepath.Join(baseDir, cat)
		files, err := os.ReadDir(catDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || f.Name() == ".gitkeep" || f.Name() == ".gitignore" {
				continue
			}
			relPath := filepath.Join(catDir, f.Name())
			if rel, err := filepath.Rel(resolver.ProjectRoot(), relPath); err == nil {
				relPath = rel
			}
			entries = append(entries, listEntry{
				Category: cat,
				Name:     f.Name(),
				Path:     relPath,
			})
		}
	}

	return printer.Print(listOutput{Entries: entries, Shelf: shelf})
}
