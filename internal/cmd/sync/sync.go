// Package sync implements the ari sync commands.
package sync

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// cmdContext holds shared state for sync commands.
type cmdContext struct {
	common.BaseContext
}

// NewSyncCmd creates the unified sync command.
func NewSyncCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	var (
		scope             string
		riteName          string
		source            string
		resource          string
		dryRun            bool
		recover_          bool
		overwriteDiverged bool
		keepOrphans       bool
		soft              bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize rite and user resources",
		Long: `Sync generates and updates Claude Code configuration.

By default, syncs both rite scope (project .claude/) and user scope (~/.claude/).
Use --scope to limit to a specific scope.

Rite Scope:
  Generates .claude/ from the active rite (agents, mena, hooks, rules, CLAUDE.md).
  Requires a project context with ACTIVE_RITE (or --rite flag).
  Source resolution: project > user > knossos > embedded.

User Scope:
  Syncs user-level resources from $KNOSSOS_HOME to ~/.claude/.
  Works without a project context.
  Resources: agents, mena (commands + skills), hooks.

Examples:
  ari sync                              # Sync everything (default)
  ari sync --scope=rite                 # Rite only
  ari sync --scope=user                 # User only (works outside projects)
  ari sync --rite=ecosystem             # Specific rite
  ari sync --resource=agents            # Filter to just agents
  ari sync --dry-run                    # Preview
  ari sync --overwrite-diverged         # Overwrite locally modified files
  ari sync --recover                    # Adopt existing untracked files
  ari sync --keep-orphans               # Don't auto-remove orphaned files`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate scope
			syncScope := materialize.SyncScope(scope)
			if scope == "" {
				syncScope = materialize.ScopeAll
			}
			if !syncScope.IsValid() {
				return fmt.Errorf("invalid --scope value: %q (must be rite, user, or all)", scope)
			}

			// Validate resource
			syncResource := materialize.SyncResource(resource)
			if !syncResource.IsValid() {
				return fmt.Errorf("invalid --resource value: %q (must be agents, mena, or hooks)", resource)
			}

			// Build SyncOptions
			opts := materialize.SyncOptions{
				Scope:             syncScope,
				RiteName:          riteName,
				Source:            source,
				Resource:          syncResource,
				DryRun:            dryRun,
				Recover:           recover_,
				OverwriteDiverged: overwriteDiverged,
				KeepOrphans:       keepOrphans,
				Soft:              soft,
			}

			return runSync(ctx, opts, cmd)
		},
	}

	// Flags per D14
	cmd.Flags().StringVar(&scope, "scope", "", "Sync scope: rite, user, or all (default: all)")
	cmd.Flags().StringVar(&riteName, "rite", "", "Rite name (defaults to ACTIVE_RITE)")
	cmd.Flags().StringVar(&source, "source", "", "Rite source: path or 'knossos' alias")
	cmd.Flags().StringVar(&resource, "resource", "", "Filter to resource type: agents, mena, or hooks")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&recover_, "recover", false, "Adopt existing untracked files into manifest")
	cmd.Flags().BoolVar(&overwriteDiverged, "overwrite-diverged", false, "Overwrite locally modified files")
	cmd.Flags().BoolVar(&keepOrphans, "keep-orphans", false, "Preserve orphaned knossos files")
	cmd.Flags().BoolVar(&soft, "soft", false, "CC-safe mode: update only agents and CLAUDE.md (skip hooks/mena/rules)")

	// Does NOT require project (user scope works without project)
	common.SetNeedsProject(cmd, false, false)

	return cmd
}

func runSync(ctx *cmdContext, opts materialize.SyncOptions, cmd *cobra.Command) error {
	printer := ctx.GetPrinter(output.FormatText)

	// Resolve project directory for rite scope
	projectDir, _ := os.Getwd()
	projectDirExplicit := cmd.Root().PersistentFlags().Changed("project-dir")
	if projectDirExplicit && ctx.ProjectDir != nil && *ctx.ProjectDir != "" {
		projectDir = *ctx.ProjectDir
	}

	// Create resolver and materializer
	resolver := paths.NewResolver(projectDir)
	var m *materialize.Materializer
	if opts.Source != "" {
		m = materialize.NewMaterializerWithSource(resolver, opts.Source)
	} else {
		m = materialize.NewMaterializer(resolver)
	}

	// Wire embedded assets
	if embRites := common.EmbeddedRites(); embRites != nil {
		m.WithEmbeddedFS(embRites)
	}
	if embTemplates := common.EmbeddedTemplates(); embTemplates != nil {
		m.WithEmbeddedTemplates(embTemplates)
	}
	if embHooks := common.EmbeddedHooks(); embHooks != nil {
		m.WithEmbeddedHooks(embHooks)
	}

	// Execute unified sync
	result, err := m.Sync(opts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Format output
	out := formatSyncResult(result, opts)
	if opts.DryRun {
		out["dry_run"] = true
	}

	return printer.Print(out)
}

func formatSyncResult(result *materialize.SyncResult, opts materialize.SyncOptions) map[string]any {
	out := map[string]any{
		"status": "success",
	}

	if result.RiteResult != nil {
		rite := map[string]any{
			"status": result.RiteResult.Status,
		}
		if result.RiteResult.RiteName != "" {
			rite["rite"] = result.RiteResult.RiteName
		}
		if result.RiteResult.Source != "" {
			rite["source"] = result.RiteResult.Source
		}
		if result.RiteResult.SourcePath != "" {
			rite["source_path"] = result.RiteResult.SourcePath
		}
		if len(result.RiteResult.OrphansDetected) > 0 {
			rite["orphans_detected"] = result.RiteResult.OrphansDetected
			rite["orphan_action"] = result.RiteResult.OrphanAction
		}
		if result.RiteResult.LegacyBackupPath != "" {
			rite["legacy_backup"] = result.RiteResult.LegacyBackupPath
		}
		if result.RiteResult.SoftMode {
			rite["soft_mode"] = true
			rite["deferred_stages"] = result.RiteResult.DeferredStages
		}
		out["rite"] = rite
	}

	if result.UserResult != nil {
		user := map[string]any{
			"status": result.UserResult.Status,
			"totals": result.UserResult.Totals,
		}
		if len(result.UserResult.Errors) > 0 {
			user["errors"] = result.UserResult.Errors
		}
		if len(result.UserResult.Resources) > 0 {
			resources := map[string]any{}
			for k, v := range result.UserResult.Resources {
				resources[string(k)] = v
			}
			user["resources"] = resources
		}
		out["user"] = user
	}

	return out
}

// getPrinter creates an output printer from the context.
func (c *cmdContext) getPrinter() *output.Printer {
	return c.GetPrinter(output.FormatText)
}
