package sync

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/paths"
)

func newMaterializeCmd(ctx *cmdContext) *cobra.Command {
	var force bool
	var riteName string
	var source string
	var dryRun bool
	var removeAll bool
	var keepAll bool
	var promoteAll bool
	var minimal bool

	cmd := &cobra.Command{
		Use:   "materialize",
		Short: "Generate .claude/ directory from templates and rite manifests",
		Long: `Materialize generates the complete .claude/ directory structure from:
  - templates/ (hooks, CLAUDE.md sections)
  - rites/{active}/ (agents, skills)
  - rites/shared/ (shared skills)

This is an idempotent operation - it can be run multiple times safely.

By default, it uses the current ACTIVE_RITE. Use --rite to specify a different rite.

Rite Source Resolution (in priority order):
  1. --source flag: explicit path or "knossos" alias for $KNOSSOS_HOME
  2. Project rites: ./rites/{rite}/
  3. User rites: ~/.local/share/knossos/rites/{rite}/
  4. Knossos platform: $KNOSSOS_HOME/rites/{rite}/

Cross-Cutting Mode:
  Use --minimal to generate base infrastructure without a rite (no agents/skills).
  This is suitable for projects using session tracking without orchestrated workflows.

Orphan Handling:
  When switching rites, agents from the previous rite become "orphans".
  By default, orphans are preserved (--keep-all behavior).
  Use --remove-all to delete orphans (with backup) or --promote-all to move them to user-level.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate mutually exclusive flags
			exclusiveCount := 0
			if removeAll {
				exclusiveCount++
			}
			if keepAll {
				exclusiveCount++
			}
			if promoteAll {
				exclusiveCount++
			}
			if exclusiveCount > 1 {
				return fmt.Errorf("--remove-all, --keep-all, and --promote-all are mutually exclusive")
			}

			// Build options
			opts := materialize.Options{
				Force:      force || removeAll || promoteAll,
				DryRun:     dryRun,
				RemoveAll:  removeAll,
				KeepAll:    keepAll || exclusiveCount == 0, // Default to keep
				PromoteAll: promoteAll,
				Minimal:    minimal,
			}

			// Check if --project-dir was explicitly set by user
			// (vs auto-discovered by PersistentPreRunE)
			projectDirExplicit := cmd.Root().PersistentFlags().Changed("project-dir")

			return runMaterialize(ctx, riteName, source, opts, projectDirExplicit)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force regeneration, overwriting local changes")
	cmd.Flags().StringVar(&riteName, "rite", "", "Rite to materialize (defaults to current ACTIVE_RITE)")
	cmd.Flags().StringVar(&source, "source", "", "Rite source: path or 'knossos' for $KNOSSOS_HOME")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&removeAll, "remove-all", false, "Remove all orphan agents (with backup)")
	cmd.Flags().BoolVar(&keepAll, "keep-all", false, "Preserve all orphan agents (default)")
	cmd.Flags().BoolVar(&promoteAll, "promote-all", false, "Move orphan agents to user-level (~/.claude/agents/)")
	cmd.Flags().BoolVar(&minimal, "minimal", false, "Generate base infrastructure only (no rite/agents/skills)")

	// Materialize doesn't require existing project - it can bootstrap new projects
	common.SetNeedsProject(cmd, false, false)

	return cmd
}

func runMaterialize(ctx *cmdContext, riteName string, source string, opts materialize.Options, projectDirExplicit bool) error {
	printer := ctx.getPrinter()

	// Always use current working directory for materialize (bootstrap mode)
	// This ensures we create .claude/ in cwd, not in some parent project
	projectDir, err := os.Getwd()
	if err != nil {
		printer.PrintError(fmt.Errorf("failed to get current directory: %w", err))
		return err
	}

	// Only allow explicit --project-dir flag to override cwd
	// (ignore auto-discovered project dir from PersistentPreRunE)
	if projectDirExplicit && ctx.ProjectDir != nil && *ctx.ProjectDir != "" {
		projectDir = *ctx.ProjectDir
	}

	// Create resolver with explicit project directory
	resolver := paths.NewResolver(projectDir)

	// Create materializer with source resolution
	var m *materialize.Materializer
	if source != "" {
		m = materialize.NewMaterializerWithSource(resolver, source)
	} else {
		m = materialize.NewMaterializer(resolver)
	}

	// Wire embedded assets if available
	if embRites := common.EmbeddedRites(); embRites != nil {
		m.WithEmbeddedFS(embRites)
	}
	if embTemplates := common.EmbeddedTemplates(); embTemplates != nil {
		m.WithEmbeddedTemplates(embTemplates)
	}
	if embHooks := common.EmbeddedHooks(); embHooks != nil {
		m.WithEmbeddedHooks(embHooks)
	}

	// Detect if running inside Claude Code — use staged materialization
	// to prevent CC's file watcher from seeing intermediate states.
	inSession := os.Getenv("CLAUDE_SESSION_ID") != ""
	if inSession {
		printer.VerboseLog("info", "Detected active Claude Code session — using staged materialization", nil)
	}

	// Handle minimal mode (cross-cutting, no rite required)
	if opts.Minimal {
		printer.VerboseLog("info", "Materializing minimal .claude/ (cross-cutting mode)", nil)
		var result *materialize.Result
		var err error
		if inSession {
			result, err = m.StagedMaterialize(func(sm *materialize.Materializer) (*materialize.Result, error) {
				return sm.MaterializeMinimal(opts)
			})
		} else {
			result, err = m.MaterializeMinimal(opts)
		}
		if err != nil {
			printer.PrintError(err)
			return err
		}

		output := map[string]any{
			"status":     "success",
			"message":    "Successfully materialized minimal .claude/ (cross-cutting mode)",
			"mode":       "minimal",
			"claude_dir": resolver.ClaudeDir(),
		}
		if result.LegacyBackupPath != "" {
			output["legacy_backup"] = result.LegacyBackupPath
			output["migration_hint"] = "Legacy CLAUDE.md backed up. Add custom content to the 'user-content' section."
		}
		if opts.DryRun {
			output["dry_run"] = true
			output["message"] = "[DRY RUN] Would materialize minimal .claude/"
		}
		return printer.Print(output)
	}

	// Determine which rite to materialize
	if riteName == "" {
		// Read current ACTIVE_RITE
		activeRitePath := resolver.ActiveRiteFile()
		data, err := os.ReadFile(activeRitePath)
		if err != nil {
			if os.IsNotExist(err) {
				hint := "no ACTIVE_RITE found, please specify --rite or use --minimal for cross-cutting mode"
				if source == "" {
					hint += " (use --source to specify rite location)"
				}
				printer.PrintError(fmt.Errorf(hint))
				return err
			}
			printer.PrintError(err)
			return err
		}
		riteName = string(data)
		// Trim whitespace/newlines
		riteName = string(bytes.TrimSpace([]byte(riteName)))
	}

	if riteName == "" {
		err := fmt.Errorf("rite name cannot be empty")
		printer.PrintError(err)
		return err
	}

	// Run materialization
	logDetails := map[string]any{"rite": riteName}
	if source != "" {
		logDetails["source"] = source
	}
	printer.VerboseLog("info", fmt.Sprintf("Materializing .claude/ for rite: %s", riteName), logDetails)

	var result *materialize.Result
	if inSession {
		result, err = m.StagedMaterialize(func(sm *materialize.Materializer) (*materialize.Result, error) {
			return sm.MaterializeWithOptions(riteName, opts)
		})
	} else {
		result, err = m.MaterializeWithOptions(riteName, opts)
	}
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	output := map[string]any{
		"status":     "success",
		"message":    fmt.Sprintf("Successfully materialized .claude/ for rite '%s'", riteName),
		"rite":       riteName,
		"claude_dir": resolver.ClaudeDir(),
	}

	// Add source info from resolution result
	if result.Source != "" {
		output["source_type"] = result.Source
		output["source_path"] = result.SourcePath
	}

	// Add orphan info if present
	if len(result.OrphansDetected) > 0 {
		output["orphans_detected"] = result.OrphansDetected
		output["orphan_action"] = result.OrphanAction
	}

	// Add hooks info
	if result.HooksSkipped {
		output["hooks_skipped"] = true
	}

	// Add legacy backup info if migration occurred
	if result.LegacyBackupPath != "" {
		output["legacy_backup"] = result.LegacyBackupPath
		output["migration_hint"] = "Legacy CLAUDE.md backed up. Add custom content to the 'user-content' section in the new file."
	}

	// Add dry-run indicator
	if opts.DryRun {
		output["dry_run"] = true
		output["message"] = fmt.Sprintf("[DRY RUN] Would materialize .claude/ for rite '%s'", riteName)
	}

	return printer.Print(output)
}
