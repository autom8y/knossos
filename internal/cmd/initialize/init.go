// Package initialize implements the "ari init" command for bootstrapping
// a Knossos project. The package is named "initialize" because "init" is
// a reserved keyword in Go.
package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

// cmdContext holds shared state for the init command.
type cmdContext struct {
	common.BaseContext
}

// initOutput represents the JSON output structure for ari init.
type initOutput struct {
	Initialized bool   `json:"initialized"`
	ProjectDir  string `json:"project_dir"`
	Rite        string `json:"rite,omitempty"`
	Source      string `json:"source,omitempty"`
	Mode        string `json:"mode"`
	Message     string `json:"message"`
}

// Text implements output.Textable for human-readable output.
func (o initOutput) Text() string {
	if !o.Initialized {
		return o.Message
	}
	if o.Rite != "" {
		return fmt.Sprintf("Initialized Knossos project with rite '%s' (source: %s)", o.Rite, o.Source)
	}
	return "Initialized Knossos project (minimal scaffold)"
}

// NewInitCmd creates the "ari init" command.
func NewInitCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	var riteName string
	var source string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Knossos project",
		Long: `Scaffolds .claude/ directory with CLAUDE.md, settings.local.json,
and KNOSSOS_MANIFEST.yaml. Optionally activates a rite.

Works without KNOSSOS_HOME set -- uses embedded rite definitions.

Examples:
  ari init                    # Minimal scaffold
  ari init --rite 10x-dev     # Scaffold with 10x-dev rite
  ari init --force            # Re-initialize existing project`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(ctx, riteName, source, force, cmd)
		},
	}

	cmd.Flags().StringVar(&riteName, "rite", "", "Rite to activate after scaffolding")
	cmd.Flags().StringVar(&source, "source", "", "Explicit rite source path")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing .claude/ directory")

	// This command does NOT require an existing project -- it creates the project context.
	common.SetNeedsProject(cmd, false, false)

	return cmd
}

func runInit(ctx *cmdContext, riteName, source string, force bool, cmd *cobra.Command) error {
	printer := ctx.GetPrinter(output.FormatText)

	// Determine project directory: use cwd unless --project-dir was explicitly set.
	projectDir, err := os.Getwd()
	if err != nil {
		printer.PrintError(fmt.Errorf("failed to get current directory: %w", err))
		return err
	}

	// Only check explicit --project-dir when running through cobra (cmd is non-nil).
	if cmd != nil {
		projectDirExplicit := cmd.Root().PersistentFlags().Changed("project-dir")
		if projectDirExplicit && ctx.ProjectDir != nil && *ctx.ProjectDir != "" {
			projectDir = *ctx.ProjectDir
		}
	}

	claudeDir := filepath.Join(projectDir, ".claude")
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")

	// Check existing state for idempotency.
	if _, err := os.Stat(manifestPath); err == nil && !force {
		// Already initialized -- exit 0 (not an error).
		result := initOutput{
			Initialized: false,
			ProjectDir:  projectDir,
			Mode:        "already_initialized",
			Message:     "Already initialized (use --force to reinitialize)",
		}
		return printer.Print(result)
	}

	if _, err := os.Stat(claudeDir); err == nil && !force {
		// .claude/ exists but no KNOSSOS_MANIFEST.yaml -- not Knossos-managed.
		errMsg := fmt.Errorf(".claude/ exists but is not Knossos-managed. Use --force to initialize.")
		printer.PrintError(errMsg)
		return errMsg
	}

	// Create resolver targeting the project directory.
	resolver := paths.NewResolver(projectDir)

	// Create materializer with source resolution.
	var mat *materialize.Materializer
	if source != "" {
		mat = materialize.NewMaterializerWithSource(resolver, source)
	} else {
		mat = materialize.NewMaterializer(resolver)
	}

	// Wire embedded assets if available.
	if embRites := common.EmbeddedRites(); embRites != nil {
		mat.WithEmbeddedFS(embRites)
	}
	if embTemplates := common.EmbeddedTemplates(); embTemplates != nil {
		mat.WithEmbeddedTemplates(embTemplates)
	}
	if embHooks := common.EmbeddedHooksYAML(); embHooks != nil {
		mat.WithEmbeddedHooks(embHooks)
	}

	// Materialize based on whether a rite was specified.
	if riteName != "" {
		printer.VerboseLog("info", fmt.Sprintf("Initializing with rite: %s", riteName), map[string]any{
			"project_dir": projectDir,
			"rite":        riteName,
			"source":      source,
		})

		result, err := mat.MaterializeWithOptions(riteName, materialize.Options{
			Force:   force,
			KeepAll: true,
		})
		if err != nil {
			printer.PrintError(err)
			return err
		}

		sourceType := result.Source
		if sourceType == "" {
			sourceType = "filesystem"
		}

		out := initOutput{
			Initialized: true,
			ProjectDir:  projectDir,
			Rite:        riteName,
			Source:       sourceType,
			Mode:        "rite",
			Message:     fmt.Sprintf("Initialized with rite '%s' (source: %s)", riteName, sourceType),
		}
		return printer.Print(out)
	}

	// No rite specified -- minimal scaffold.
	printer.VerboseLog("info", "Initializing minimal scaffold", map[string]any{
		"project_dir": projectDir,
	})

	_, err = mat.MaterializeMinimal(materialize.Options{})
	if err != nil {
		printer.PrintError(err)
		return err
	}

	out := initOutput{
		Initialized: true,
		ProjectDir:  projectDir,
		Source:      "minimal",
		Mode:        "minimal",
		Message:     "Initialized Knossos project (minimal scaffold)",
	}
	return printer.Print(out)
}
