// Package initialize implements the "ari init" command for bootstrapping
// a Knossos project. The package is named "initialize" because "init" is
// a reserved keyword in Go.
package initialize

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/config"

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
		return fmt.Sprintf(`Initialized project with rite '%s'

What was created:
  .claude/agents/     Agent prompts for the %s workflow
  .claude/skills/     Reference knowledge agents can load
  .claude/commands/   Slash commands (type / in Claude Code to see them)
  .claude/CLAUDE.md   Project instructions (always in context)
  .claude/settings.json  Hook configuration
  .knossos/           Satellite project config (rite overrides)
  .sos/               Session state and lifecycle
  .ledge/             Work product artifacts

Next steps:
  1. Open this project in Claude Code
  2. Type /go to start a session
  3. Describe what you want to do — the agents will coordinate`, o.Rite, o.Rite)
	}
	return `Initialized project (minimal scaffold)

What was created:
  .claude/CLAUDE.md   Project instructions (always in context)
  .knossos/           Satellite project config (rite overrides)
  .sos/               Session state and lifecycle
  .ledge/             Work product artifacts

Next steps:
  1. Open this project in Claude Code
  2. Run 'ari init --rite review' to add a workflow
  3. Available rites: review, slop-chop, 10x-dev, and more`
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
	if embAgents := common.EmbeddedAgents(); embAgents != nil {
		mat.WithEmbeddedAgents(embAgents)
	}
	if embMena := common.EmbeddedMena(); embMena != nil {
		mat.WithEmbeddedMena(embMena)
		// Extract embedded platform mena to XDG data dir if not already present.
		// This provides /commit, /start, /go, guidance skills, etc. to all users
		// regardless of whether KNOSSOS_HOME is set or the source tree exists.
		extractEmbeddedMenaToXDG(embMena)
	}
	// Bootstrap config/hooks.yaml from embedded bytes if not already present.
	if hooksYAML := common.EmbeddedHooksYAML(); len(hooksYAML) > 0 {
		hooksPath := filepath.Join(projectDir, "config", "hooks.yaml")
		if _, err := os.Stat(hooksPath); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(hooksPath), 0755); err == nil {
				os.WriteFile(hooksPath, hooksYAML, 0644)
			}
		}
	}

	// Materialize based on whether a rite was specified.
	// Scaffold project-level directories (.knossos/, .sos/, .ledge/).
	scaffoldProjectDirs(projectDir)

	if riteName != "" {
		printer.VerboseLog("info", fmt.Sprintf("Initializing with rite: %s", riteName), map[string]any{
			"project_dir": projectDir,
			"rite":        riteName,
			"source":      source,
		})

		syncResult, err := mat.Sync(materialize.SyncOptions{
			Scope:       materialize.ScopeAll,
			RiteName:    riteName,
			KeepOrphans: true,
		})
		if err != nil {
			printer.PrintError(err)
			return err
		}

		sourceType := "filesystem"
		if syncResult.RiteResult != nil && syncResult.RiteResult.Source != "" {
			sourceType = syncResult.RiteResult.Source
		}

		// Generate settings.json with required hooks if it doesn't exist.
		// This ensures agent-guard hooks fire on foreign projects.
		writeDefaultSettings(claudeDir)

		out := initOutput{
			Initialized: true,
			ProjectDir:  projectDir,
			Rite:        riteName,
			Source:      sourceType,
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

	// Generate settings.json with required hooks if it doesn't exist.
	writeDefaultSettings(claudeDir)

	out := initOutput{
		Initialized: true,
		ProjectDir:  projectDir,
		Source:      "minimal",
		Mode:        "minimal",
		Message:     "Initialized Knossos project (minimal scaffold)",
	}
	return printer.Print(out)
}

// extractEmbeddedMenaToXDG extracts platform mena from the embedded FS to the
// XDG data directory. This is the hybrid distribution model: binary embeds mena,
// first init extracts to XDG cache so subsequent syncs read from filesystem.
// Idempotent: skips extraction if XDG mena dir already exists.
func extractEmbeddedMenaToXDG(embMena fs.FS) {
	xdgMena := filepath.Join(config.XDGDataDir(), "mena")
	if _, err := os.Stat(xdgMena); err == nil {
		return // Already extracted
	}
	if err := os.MkdirAll(xdgMena, 0755); err != nil {
		return // Best-effort
	}
	fs.WalkDir(embMena, "mena", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		rel, relErr := filepath.Rel("mena", path)
		if relErr != nil || rel == "." {
			return nil
		}
		dest := filepath.Join(xdgMena, rel)
		if d.IsDir() {
			os.MkdirAll(dest, 0755)
			return nil
		}
		content, readErr := fs.ReadFile(embMena, path)
		if readErr != nil {
			return nil
		}
		os.MkdirAll(filepath.Dir(dest), 0755)
		os.WriteFile(dest, content, 0644)
		return nil
	})
}

// scaffoldProjectDirs creates the project-level directory structure:
//   - .knossos/ — satellite project config (rites, mena overrides)
//   - .sos/ — session state and lifecycle
//   - .ledge/ — work product artifacts
//
// Idempotent: directories are created only if they don't already exist.
func scaffoldProjectDirs(projectDir string) {
	dirs := []string{
		filepath.Join(projectDir, ".knossos"),
		filepath.Join(projectDir, ".sos"),
		filepath.Join(projectDir, ".ledge"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755) // Best-effort, non-fatal.
	}
}

// writeDefaultSettings writes settings.json with the agent-guard hook configuration
// if no settings.json exists yet. Non-fatal: hooks are optional infrastructure.
func writeDefaultSettings(claudeDir string) {
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		// Already exists (or stat error) -- don't overwrite user settings.
		return
	}
	settingsContent := []byte(`{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "ari hook agent-guard --output json"
          }
        ]
      }
    ]
  }
}
`)
	// Best-effort write -- failures are non-fatal since hooks are optional.
	os.WriteFile(settingsPath, settingsContent, 0644)
}
