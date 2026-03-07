package ledge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	ledgepkg "github.com/autom8y/knossos/internal/ledge"
	"github.com/autom8y/knossos/internal/output"
)

func newPromoteCmd(ctx *cmdContext) *cobra.Command {
	return &cobra.Command{
		Use:   "promote <path>",
		Short: "Promote a ledge artifact to shelf",
		Long: `Move an artifact from .ledge/{category}/ to .ledge/shelf/{category}/.

Adds promotion frontmatter (promoted_at, promoted_from) and removes the source
file. Promotable categories: decisions, specs, reviews.

Examples:
  ari ledge promote .ledge/reviews/GAP-auth-refactor.md
  ari ledge promote .ledge/decisions/ADR-0030.md`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPromote(ctx, args[0])
		},
	}
}

func runPromote(ctx *cmdContext, sourcePath string) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	result, err := ledgepkg.Promote(resolver, sourcePath)
	if err != nil {
		return err
	}

	return printer.Print(promoteOutput{*result})
}

type promoteOutput struct {
	ledgepkg.PromoteResult
}

// Text implements output.Textable.
func (o promoteOutput) Text() string {
	return fmt.Sprintf("Promoted %s → %s\n", o.SourcePath, o.ShelfPath)
}
