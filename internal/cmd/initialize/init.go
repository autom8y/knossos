// Package initialize implements the "ari init" command for bootstrapping
// a Knossos project. The package is named "initialize" because "init" is
// a reserved keyword in Go.
package initialize

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/config"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
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
  .knossos/           Satellite project config (rite overrides)
  .sos/               Session state and lifecycle
  .ledge/             Work product artifacts
    decisions/        ADRs and design decisions
    specs/            PRDs and technical specs
    reviews/          Audit reports and code reviews
    spikes/           Exploration and research (git-ignored by default)
    shelf/            Tracked work products (survives .gitignore)

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
    decisions/        ADRs and design decisions
    specs/            PRDs and technical specs
    reviews/          Audit reports and code reviews
    spikes/           Exploration and research (git-ignored by default)
    shelf/            Tracked work products (survives .gitignore)

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
		return common.PrintAndReturn(printer, fmt.Errorf("failed to get current directory: %w", err))
	}

	// Only check explicit --project-dir when running through cobra (cmd is non-nil).
	if cmd != nil {
		projectDirExplicit := cmd.Root().PersistentFlags().Changed("project-dir")
		if projectDirExplicit && ctx.ProjectDir != nil && *ctx.ProjectDir != "" {
			projectDir = *ctx.ProjectDir
		}
	}

	knossosDir := filepath.Join(projectDir, ".knossos")
	manifestPath := filepath.Join(knossosDir, "KNOSSOS_MANIFEST.yaml")

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

	claudeDir := filepath.Join(projectDir, ".claude")
	if _, err := os.Stat(claudeDir); err == nil && !force {
		// .claude/ exists but no KNOSSOS_MANIFEST.yaml in .knossos/ -- not Knossos-managed.
		errMsg := errors.New(errors.CodeUsageError, ".claude/ exists but is not Knossos-managed; use --force to initialize")
		return common.PrintAndReturn(printer, errMsg)
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
		// This provides /commit, /sos start, /go, guidance skills, etc. to all users
		// regardless of whether KNOSSOS_HOME is set or the source tree exists.
		extractEmbeddedMenaToXDG(embMena)
	}
	if embProc := common.EmbeddedProcessions(); embProc != nil {
		mat.WithEmbeddedProcessions(embProc)
	}
	// Bootstrap config/hooks.yaml from embedded bytes if not already present.
	if hooksYAML := common.EmbeddedHooksYAML(); len(hooksYAML) > 0 {
		hooksPath := filepath.Join(projectDir, "config", "hooks.yaml")
		if _, err := os.Stat(hooksPath); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(hooksPath), 0755); err == nil {
				_ = os.WriteFile(hooksPath, hooksYAML, 0644)
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
			return common.PrintAndReturn(printer, err)
		}

		sourceType := "filesystem"
		if syncResult.RiteResult != nil && syncResult.RiteResult.Source != "" {
			sourceType = syncResult.RiteResult.Source
		}

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
		return common.PrintAndReturn(printer, err)
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

// xdgVersionSentinel is the filename written inside the XDG mena directory to
// record which binary version performed the last extraction. Re-extraction is
// triggered when the sentinel is absent or its content differs from the running
// binary version.
const xdgVersionSentinel = ".ari-version"

// extractEmbeddedMenaToXDG extracts platform mena from the embedded FS to the
// XDG data directory. This is the hybrid distribution model: binary embeds mena,
// first init extracts to XDG cache so subsequent syncs read from filesystem.
//
// Version-aware: writes a .ari-version sentinel after extraction. On subsequent
// calls it reads the sentinel and skips extraction only if the recorded version
// matches the current binary version. If versions differ (or sentinel is absent
// but the directory already exists), the XDG mena directory is wiped and
// re-extracted from the embedded FS so the installed-user copy stays current
// across binary upgrades.
func extractEmbeddedMenaToXDG(embMena fs.FS) {
	currentVersion := common.BuildVersion()
	xdgMena := filepath.Join(config.XDGDataDir(), "mena")
	sentinelPath := filepath.Join(xdgMena, xdgVersionSentinel)

	if _, err := os.Stat(xdgMena); err == nil {
		// XDG mena dir exists -- check version sentinel.
		if data, readErr := os.ReadFile(sentinelPath); readErr == nil {
			if string(data) == currentVersion {
				return // Already extracted at this version.
			}
		}
		// Sentinel absent or version mismatch -- wipe and re-extract.
		if removeErr := os.RemoveAll(xdgMena); removeErr != nil {
			slog.Warn("extractEmbeddedMena: RemoveAll failed", "path", xdgMena, "error", removeErr)
			return // Best-effort
		}
	}

	if err := os.MkdirAll(xdgMena, 0755); err != nil {
		slog.Warn("extractEmbeddedMena: MkdirAll failed", "path", xdgMena, "error", err)
		return // Best-effort
	}

	_ = fs.WalkDir(embMena, "mena", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Warn("extractEmbeddedMena: WalkDir skip", "path", path, "error", err)
			return nil // skip errors
		}
		rel, relErr := filepath.Rel("mena", path)
		if relErr != nil || rel == "." {
			return nil
		}
		dest := filepath.Join(xdgMena, rel)
		if d.IsDir() {
			_ = os.MkdirAll(dest, 0755)
			return nil
		}
		content, readErr := fs.ReadFile(embMena, path)
		if readErr != nil {
			slog.Warn("extractEmbeddedMena: ReadFile failed", "path", path, "error", readErr)
			return nil
		}
		_ = os.MkdirAll(filepath.Dir(dest), 0755)
		if writeErr := os.WriteFile(dest, content, 0644); writeErr != nil {
			slog.Warn("extractEmbeddedMena: WriteFile failed", "path", dest, "error", writeErr)
		}
		return nil
	})

	// Write version sentinel so subsequent calls can detect stale extractions.
	if writeErr := os.WriteFile(sentinelPath, []byte(currentVersion), 0644); writeErr != nil {
		slog.Warn("extractEmbeddedMena: sentinel write failed", "path", sentinelPath, "error", writeErr)
	}
}

// scaffoldProjectDirs creates the project-level directory structure:
//   - .knossos/ — satellite project config (rites, mena overrides)
//   - .sos/ — session state and lifecycle
//   - .ledge/ — work product artifacts (with subdirectories)
//
// Idempotent: directories are created only if they don't already exist.
// Each .ledge subdirectory gets a .gitkeep so empty dirs survive git.
func scaffoldProjectDirs(projectDir string) {
	dirs := []string{
		filepath.Join(projectDir, ".knossos"),
		filepath.Join(projectDir, ".sos"),
	}
	for _, d := range dirs {
		_ = os.MkdirAll(d, 0755) // Best-effort, non-fatal.
	}

	ledgeDir := filepath.Join(projectDir, ".ledge")
	ledgeSubdirs := []string{"decisions", "specs", "reviews", "spikes"}
	for _, sub := range ledgeSubdirs {
		subDir := filepath.Join(ledgeDir, sub)
		_ = os.MkdirAll(subDir, 0755)
		// .gitkeep so empty dirs survive git.
		gitkeep := filepath.Join(subDir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			_ = os.WriteFile(gitkeep, []byte(""), 0644)
		}
	}

	// .ledge/shelf/ — tracked production-quality work products (mirrored categories).
	shelfDir := filepath.Join(ledgeDir, "shelf")
	shelfSubdirs := []string{"decisions", "specs", "reviews"}
	for _, sub := range shelfSubdirs {
		subDir := filepath.Join(shelfDir, sub)
		_ = os.MkdirAll(subDir, 0755)
		gitkeep := filepath.Join(subDir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			_ = os.WriteFile(gitkeep, []byte(""), 0644)
		}
	}
	// Root shelf .gitkeep (in case no subdirs have content yet).
	shelfGitkeep := filepath.Join(shelfDir, ".gitkeep")
	if _, err := os.Stat(shelfGitkeep); os.IsNotExist(err) {
		_ = os.WriteFile(shelfGitkeep, []byte(""), 0644)
	}

	// .sos/land/ — tracked cross-session synthesis (survives gitignore negation).
	landDir := filepath.Join(projectDir, ".sos", "land")
	_ = os.MkdirAll(landDir, 0755)
	landGitkeep := filepath.Join(landDir, ".gitkeep")
	if _, err := os.Stat(landGitkeep); os.IsNotExist(err) {
		_ = os.WriteFile(landGitkeep, []byte(""), 0644)
	}

	// Root .ledge/.gitignore: preserve decisions and specs, ignore session scratch.
	writeLedgeGitignore(ledgeDir)

	// .ledge/spikes/.gitignore: spike artifacts may be large, opt-in tracking.
	writeSpikesGitignore(filepath.Join(ledgeDir, "spikes"))

	// Root .gitignore: standardized Knossos ephemeral artifact patterns.
	writeProjectGitignore(projectDir)
}

// writeLedgeGitignore writes the root .ledge/.gitignore that ignores session-scratch
// patterns while preserving decisions/ and specs/ content. Idempotent.
func writeLedgeGitignore(ledgeDir string) {
	gitignorePath := filepath.Join(ledgeDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); !os.IsNotExist(err) {
		return
	}
	content := []byte(`# Session scratch — not committed
*.scratch
*.tmp
*.wip

# Subdirectories managed individually
# decisions/ and specs/ are fully tracked
# reviews/ is tracked
# spikes/ has its own .gitignore (opt-in)
`)
	_ = os.WriteFile(gitignorePath, content, 0644)
}

// knossosGitignoreBlock is the canonical gitignore block for Knossos-managed projects.
// Delimited by marker comments for idempotent detection and update.
const knossosGitignoreBlock = `# Knossos
.knossos/
.claude/CLAUDE.md
**/.sos/*
!**/.sos/land/
!**/.sos/land/**
**/.ledge/*
!**/.ledge/shelf/
!**/.ledge/shelf/**
# End Knossos
`

// writeProjectGitignore writes or updates the Knossos block in the project root
// .gitignore. Uses marker comments (# Knossos / # End Knossos) for idempotent
// block detection: creates the file if absent, appends the block if no markers
// are found, or replaces the existing marker-delimited region.
func writeProjectGitignore(projectDir string) {
	gitignorePath := filepath.Join(projectDir, ".gitignore")

	existing, err := os.ReadFile(gitignorePath)
	if err != nil {
		// No existing file — write the block as the entire file.
		_ = os.WriteFile(gitignorePath, []byte(knossosGitignoreBlock), 0644)
		return
	}

	content := string(existing)

	startIdx := strings.Index(content, "# Knossos\n")
	endMarker := "# End Knossos\n"
	endIdx := strings.Index(content, endMarker)

	if startIdx >= 0 && endIdx >= 0 && endIdx > startIdx {
		// Replace existing block (including markers).
		before := content[:startIdx]
		after := content[endIdx+len(endMarker):]
		newContent := before + knossosGitignoreBlock + after
		_ = os.WriteFile(gitignorePath, []byte(newContent), 0644)
		return
	}

	// No existing block — append with blank line separator.
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if len(content) > 0 {
		content += "\n"
	}
	content += knossosGitignoreBlock
	_ = os.WriteFile(gitignorePath, []byte(content), 0644)
}

// writeSpikesGitignore writes .ledge/spikes/.gitignore with an opt-in policy.
// Spike artifacts may be large; add specific files to git as needed. Idempotent.
func writeSpikesGitignore(spikesDir string) {
	gitignorePath := filepath.Join(spikesDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); !os.IsNotExist(err) {
		return
	}
	content := []byte(`# Spike artifacts may be large; add specific files to git as needed
*
!.gitignore
!.gitkeep
`)
	_ = os.WriteFile(gitignorePath, content, 0644)
}
