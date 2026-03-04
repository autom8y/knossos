package artifact

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

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
					printer.PrintError(err)
					return err
				}

				result := map[string]interface{}{
					"sessions_scanned":  projectReg.SessionsIndexed,
					"artifacts_indexed": projectReg.ArtifactCount,
					"rebuild_time_ms":   0,
					"dry_run":           true,
				}
				return printer.Print(result)
			}

			// Perform rebuild
			if err := aggregator.AggregateAll(); err != nil {
				printer.PrintError(err)
				return err
			}

			elapsed := time.Since(start)

			// Load result
			projectReg, err := registry.LoadProjectRegistry()
			if err != nil {
				printer.PrintError(err)
				return err
			}

			result := map[string]interface{}{
				"sessions_scanned":  projectReg.SessionsIndexed,
				"artifacts_indexed": projectReg.ArtifactCount,
				"rebuild_time_ms":   elapsed.Milliseconds(),
				"dry_run":           false,
			}
			if err := printer.Print(result); err != nil {
				return err
			}

			// Print summary in text mode
			if *ctx.Output == "text" {
				fmt.Printf("\nRebuilt project registry:\n")
				fmt.Printf("  Sessions scanned: %d\n", projectReg.SessionsIndexed)
				fmt.Printf("  Artifacts indexed: %d\n", projectReg.ArtifactCount)
				fmt.Printf("  Time: %dms\n", elapsed.Milliseconds())
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be rebuilt without writing")

	return cmd
}
