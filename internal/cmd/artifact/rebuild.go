package artifact

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// rebuildOutput is the structured output for ari artifact rebuild.
type rebuildOutput struct {
	SessionsScanned  int   `json:"sessions_scanned"`
	ArtifactsIndexed int   `json:"artifacts_indexed"`
	RebuildTimeMs    int64 `json:"rebuild_time_ms"`
	DryRun           bool  `json:"dry_run"`
}

// Text implements output.Textable.
func (r rebuildOutput) Text() string {
	var b strings.Builder
	b.WriteString("\nRebuilt project registry:\n")
	b.WriteString(fmt.Sprintf("  Sessions scanned: %d\n", r.SessionsScanned))
	b.WriteString(fmt.Sprintf("  Artifacts indexed: %d\n", r.ArtifactsIndexed))
	b.WriteString(fmt.Sprintf("  Time: %dms\n", r.RebuildTimeMs))
	return b.String()
}

func newRebuildCmd(ctx *cmdContext) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild the project registry from all session registries",
		Long: `Rebuild the project registry from all session registries.

This scans all session directories and aggregates their artifact registries
into the project-level registry. Use this for recovery or initial index build.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.getPrinter()
			aggregator := ctx.getAggregator()
			registry := ctx.getRegistry()

			start := time.Now()

			if dryRun {
				// Load current project registry to show stats
				projectReg, err := registry.LoadProjectRegistry()
				if err != nil {
					return common.PrintAndReturn(printer, err)
				}

				return printer.Print(rebuildOutput{
					SessionsScanned:  projectReg.SessionsIndexed,
					ArtifactsIndexed: projectReg.ArtifactCount,
					RebuildTimeMs:    0,
					DryRun:           true,
				})
			}

			// Perform rebuild
			if err := aggregator.AggregateAll(); err != nil {
				return common.PrintAndReturn(printer, err)
			}

			elapsed := time.Since(start)

			// Load result
			projectReg, err := registry.LoadProjectRegistry()
			if err != nil {
				return common.PrintAndReturn(printer, err)
			}

			return printer.Print(rebuildOutput{
				SessionsScanned:  projectReg.SessionsIndexed,
				ArtifactsIndexed: projectReg.ArtifactCount,
				RebuildTimeMs:    elapsed.Milliseconds(),
				DryRun:           false,
			})
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be rebuilt without writing")

	return cmd
}
