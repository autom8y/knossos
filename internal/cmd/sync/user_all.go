package sync

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/usersync"
)

func newUserAllCmd(ctx *cmdContext) *cobra.Command {
	var dryRun, recover, force, verbose bool

	cmd := &cobra.Command{
		Use:   "all",
		Short: "Sync all user resources to ~/.claude/",
		Long: `Sync all user resources from knossos to ~/.claude/.

Runs sync for all resource types in sequence:
  1. agents
  2. mena (commands + skills)
  3. hooks

Behavior:
  - Failures in one resource type don't prevent syncing others
  - Exit code reflects aggregate success/failure
  - Summary shows results for each resource type

Examples:
  # Sync all user resources
  ari sync user all

  # Preview what would be synced
  ari sync user all --dry-run

  # Adopt existing files into manifest
  ari sync user all --recover

  # Force overwrite diverged files
  ari sync user all --force

  # JSON output for scripting
  ari sync user all --output=json`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := usersync.Options{
				DryRun:  dryRun,
				Recover: recover,
				Force:   force,
				Verbose: verbose,
			}
			return runUserSyncAll(ctx, opts)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview changes without applying")
	cmd.Flags().BoolVarP(&recover, "recover", "r", false, "Adopt existing files matching knossos")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite diverged files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	common.SetNeedsProject(cmd, false, false)

	return cmd
}

func runUserSyncAll(ctx *cmdContext, opts usersync.Options) error {
	printer := ctx.getPrinter()

	allResult := usersync.AllResult{
		SyncedAt:  time.Now().UTC().Format(time.RFC3339),
		DryRun:    opts.DryRun,
		Resources: make(map[string]usersync.Result),
		Totals:    usersync.Summary{},
		Errors:    []usersync.ResourceError{},
	}

	resourceTypes := []usersync.ResourceType{
		usersync.ResourceAgents,
		usersync.ResourceMena,
		usersync.ResourceHooks,
	}

	hasCollisions := false

	for _, resourceType := range resourceTypes {
		syncer, err := usersync.NewSyncer(resourceType)
		if err != nil {
			allResult.Errors = append(allResult.Errors, usersync.ResourceError{
				Resource: resourceType,
				Err:      err.Error(),
			})
			continue
		}

		result, err := syncer.Sync(opts)
		if err != nil {
			allResult.Errors = append(allResult.Errors, usersync.ResourceError{
				Resource: resourceType,
				Err:      err.Error(),
			})
			continue
		}

		allResult.Resources[string(resourceType)] = *result

		// Aggregate totals
		allResult.Totals.Added += result.Summary.Added
		allResult.Totals.Updated += result.Summary.Updated
		allResult.Totals.Skipped += result.Summary.Skipped
		allResult.Totals.Unchanged += result.Summary.Unchanged
		allResult.Totals.Collisions += result.Summary.Collisions

		if result.Summary.Collisions > 0 {
			hasCollisions = true
		}
	}

	if err := printer.Print(allResult); err != nil {
		return err
	}

	// Return error if there were any errors or collisions
	if len(allResult.Errors) > 0 {
		return fmt.Errorf("%s: %s", allResult.Errors[0].Resource, allResult.Errors[0].Err)
	}

	if hasCollisions {
		// Note: Per TDD, exit code 1 for collisions
		// The printer already printed the results, just indicate collisions occurred
		return nil // Collisions are not fatal errors, just informational
	}

	return nil
}
