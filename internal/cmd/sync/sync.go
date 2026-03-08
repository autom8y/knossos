// Package sync implements the ari sync commands.
package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
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
		orgName           string
		dryRun            bool
		recoverMode          bool
		overwriteDiverged bool
		keepOrphans       bool
		soft              bool
		budget            bool
		elCheapo          bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize rite and user resources",
		Long: `Sync generates and updates Claude Code configuration.

By default, syncs rite scope (project .claude/), org scope, and user scope (~/.claude/).
Use --scope to limit to a specific scope.

Rite Scope:
  Generates .claude/ from the active rite (agents, mena, hooks, rules, CLAUDE.md).
  Requires a project context with ACTIVE_RITE (or --rite flag).
  Source resolution: project > user > org > knossos > embedded.

Org Scope:
  Syncs org-level agents and mena to ~/.claude/.
  Requires an active org (KNOSSOS_ORG env var, ari org set, or --org flag).

User Scope:
  Syncs user-level resources from $KNOSSOS_HOME to ~/.claude/.
  Works without a project context.
  Resources: agents, mena (commands + skills), hooks.

Examples:
  ari sync                              # Sync everything (default)
  ari sync --scope=rite                 # Rite only
  ari sync --scope=user                 # User only (works outside projects)
  ari sync --scope=org                  # Org only
  ari sync --org=autom8y                # Sync with specific org
  ari sync --rite=ecosystem             # Specific rite
  ari sync --resource=agents            # Filter to just agents
  ari sync --dry-run                    # Preview
  ari sync --overwrite-diverged         # Overwrite locally modified files
  ari sync --recover                    # Adopt existing untracked files
  ari sync --keep-orphans               # Don't auto-remove orphaned files
  ari sync --budget                     # Show context token budget after sync`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate scope
			syncScope := materialize.SyncScope(scope)
			if scope == "" {
				syncScope = materialize.ScopeAll
			}
			if !syncScope.IsValid() {
				return errors.New(errors.CodeUsageError, fmt.Sprintf("invalid --scope value: %q (must be rite, org, user, or all)", scope))
			}

			// Validate resource
			syncResource := materialize.SyncResource(resource)
			if !syncResource.IsValid() {
				return errors.New(errors.CodeUsageError, fmt.Sprintf("invalid --resource value: %q (must be agents, mena, or hooks)", resource))
			}

			// El-cheapo only affects rite scope
			if elCheapo && (syncScope == materialize.ScopeUser || syncScope == materialize.ScopeOrg) {
				return errors.New(errors.CodeUsageError, "--el-cheapo only affects rite scope; use --scope=rite or default")
			}

			// Build SyncOptions
			opts := materialize.SyncOptions{
				Scope:             syncScope,
				RiteName:          riteName,
				Source:            source,
				Resource:          syncResource,
				OrgName:           orgName,
				DryRun:            dryRun,
				Recover:           recoverMode,
				OverwriteDiverged: overwriteDiverged,
				KeepOrphans:       keepOrphans,
				Soft:              soft,
				ElCheapo:          elCheapo,
			}

			return runSync(ctx, opts, budget, cmd)
		},
	}

	// Flags per D14
	cmd.Flags().StringVar(&scope, "scope", "", "Sync scope: rite, org, user, or all (default: all)")
	cmd.Flags().StringVar(&riteName, "rite", "", "Rite name (defaults to ACTIVE_RITE)")
	cmd.Flags().StringVar(&source, "source", "", "Rite source: path or 'knossos' alias")
	cmd.Flags().StringVar(&orgName, "org", "", "Organization name (defaults to active org)")
	cmd.Flags().StringVar(&resource, "resource", "", "Filter to resource type: agents, mena, or hooks")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&recoverMode, "recover", false, "Adopt existing untracked files into manifest")
	cmd.Flags().BoolVar(&overwriteDiverged, "overwrite-diverged", false, "Overwrite locally modified files")
	cmd.Flags().BoolVar(&keepOrphans, "keep-orphans", false, "Preserve orphaned knossos files")
	cmd.Flags().BoolVar(&soft, "soft", false, "CC-safe mode: update only agents and CLAUDE.md (skip hooks/mena/rules)")
	cmd.Flags().BoolVar(&budget, "budget", false, "Show context token budget after sync")
	cmd.Flags().BoolVar(&elCheapo, "el-cheapo", false, "Force all agents to haiku model (ephemeral, reverted on session exit)")

	// Does NOT require project (user scope works without project)
	common.SetNeedsProject(cmd, false, false)

	return cmd
}

func runSync(ctx *cmdContext, opts materialize.SyncOptions, showBudget bool, cmd *cobra.Command) error {
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
	if embAgents := common.EmbeddedAgents(); embAgents != nil {
		m.WithEmbeddedAgents(embAgents)
	}
	if embMena := common.EmbeddedMena(); embMena != nil {
		m.WithEmbeddedMena(embMena)
	}
	// Execute unified sync
	result, err := m.Sync(opts)
	if err != nil {
		return err
	}

	// Format output
	out := formatSyncResult(result, opts)
	if opts.DryRun {
		out.DryRun = true
	}

	// Budget report (appended to sync output)
	if showBudget {
		claudeDir := filepath.Join(projectDir, ".claude")
		budgetData := map[string]any{}
		if err := formatBudgetReport(claudeDir, budgetData); err != nil {
			printer.VerboseLog("warn", "budget calculation failed", map[string]any{"error": err.Error()})
		} else if b, ok := budgetData["budget"]; ok {
			out.Budget = b
		}
	}

	if err := printer.Print(out); err != nil {
		return err
	}

	// Print human-readable budget text after structured output
	if showBudget {
		claudeDir := filepath.Join(projectDir, ".claude")
		text, err := budgetText(claudeDir)
		if err == nil {
			printer.PrintText(text)
		}
	}

	return nil
}

func formatSyncResult(result *materialize.SyncResult, opts materialize.SyncOptions) *output.SyncResultOutput {
	out := &output.SyncResultOutput{
		Status: "success",
	}

	if result.RiteResult != nil {
		out.Rite = &output.SyncRiteResult{
			Status:          result.RiteResult.Status,
			Error:           result.RiteResult.Error,
			RiteName:        result.RiteResult.RiteName,
			Source:          result.RiteResult.Source,
			SourcePath:      result.RiteResult.SourcePath,
			OrphansDetected: result.RiteResult.OrphansDetected,
			OrphanAction:    result.RiteResult.OrphanAction,
			LegacyBackup:    result.RiteResult.LegacyBackupPath,
			SoftMode:        result.RiteResult.SoftMode,
			DeferredStages:  result.RiteResult.DeferredStages,
			ElCheapoMode:    result.RiteResult.ElCheapoMode,
			RiteSwitched:    result.RiteResult.RiteSwitched,
			PreviousRite:    result.RiteResult.PreviousRite,
		}
	}

	if result.OrgResult != nil {
		out.Org = &output.SyncOrgResult{
			Status:  result.OrgResult.Status,
			Error:   result.OrgResult.Error,
			OrgName: result.OrgResult.OrgName,
			Source:  result.OrgResult.Source,
			Agents:  result.OrgResult.Agents,
			Mena:    result.OrgResult.Mena,
		}
	}

	if result.UserResult != nil {
		userOut := &output.SyncUserResult{
			Status: result.UserResult.Status,
			Totals: result.UserResult.Totals,
		}
		if len(result.UserResult.Errors) > 0 {
			userOut.Errors = result.UserResult.Errors
		}
		if len(result.UserResult.Resources) > 0 {
			resources := map[string]any{}
			for k, v := range result.UserResult.Resources {
				resources[string(k)] = v
			}
			userOut.Resources = resources
		}
		out.User = userOut
	}

	return out
}
