// Package root provides the root command for the ari CLI.
package root

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/autom8y/ariadne/internal/cmd/handoff"
	"github.com/autom8y/ariadne/internal/cmd/hook"
	"github.com/autom8y/ariadne/internal/cmd/manifest"
	"github.com/autom8y/ariadne/internal/cmd/session"
	"github.com/autom8y/ariadne/internal/cmd/sync"
	"github.com/autom8y/ariadne/internal/cmd/team"
	"github.com/autom8y/ariadne/internal/cmd/validate"
	"github.com/autom8y/ariadne/internal/cmd/worktree"
	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// GlobalOptions holds global flag values.
type GlobalOptions struct {
	Output     string
	Verbose    bool
	Config     string
	ProjectDir string
	SessionID  string
}

var globalOpts GlobalOptions

// SetVersion sets the version information (called from main).
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "ari",
	Short: "Ariadne - Claude Code workflow harness",
	Long: `Ariadne (ari) manages sessions, teams, manifests, and sync for Claude Code agentic workflows.

The thread that makes the maze survivable.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip project discovery for version command
		if cmd.Name() == "version" {
			return nil
		}

		// Initialize config
		if err := initConfig(); err != nil {
			return err
		}

		// Discover project root if not specified
		if globalOpts.ProjectDir == "" {
			projectRoot, err := paths.FindProjectRoot("")
			if err != nil {
				// Only fail if this is a command that needs a project
				if needsProject(cmd) {
					printer := output.NewPrinter(output.ParseFormat(globalOpts.Output), os.Stdout, os.Stderr, globalOpts.Verbose)
					printer.PrintError(errors.ErrProjectNotFound())
					return err
				}
			} else {
				globalOpts.ProjectDir = projectRoot
			}
		}

		return nil
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&globalOpts.Output, "output", "o", "text",
		"Output format: text, json, yaml")
	rootCmd.PersistentFlags().BoolVarP(&globalOpts.Verbose, "verbose", "v", false,
		"Enable verbose output (JSON lines to stderr)")
	rootCmd.PersistentFlags().StringVar(&globalOpts.Config, "config", "",
		"Config file (default: $XDG_CONFIG_HOME/ariadne/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&globalOpts.ProjectDir, "project-dir", "p", "",
		"Project root directory (overrides discovery)")
	rootCmd.PersistentFlags().StringVarP(&globalOpts.SessionID, "session-id", "s", "",
		"Session ID (overrides current)")

	// Add subcommands
	rootCmd.AddCommand(session.NewSessionCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir, &globalOpts.SessionID))
	rootCmd.AddCommand(team.NewTeamCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir))
	rootCmd.AddCommand(manifest.NewManifestCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir))
	rootCmd.AddCommand(sync.NewSyncCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir))
	rootCmd.AddCommand(validate.NewValidateCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir))
	rootCmd.AddCommand(handoff.NewHandoffCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir, &globalOpts.SessionID))
	rootCmd.AddCommand(worktree.NewWorktreeCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir))
	rootCmd.AddCommand(hook.NewHookCmd(&globalOpts.Output, &globalOpts.Verbose, &globalOpts.ProjectDir, &globalOpts.SessionID))
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		format := output.ParseFormat(globalOpts.Output)
		printer := output.NewPrinter(format, os.Stdout, os.Stderr, globalOpts.Verbose)

		if format == output.FormatJSON {
			printer.Print(map[string]string{
				"version": version,
				"commit":  commit,
				"date":    date,
				"go":      runtime.Version(),
				"os":      runtime.GOOS,
				"arch":    runtime.GOARCH,
			})
		} else {
			fmt.Printf("ari %s (%s, %s)\n", version, commit, date)
			fmt.Printf("%s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		}
	},
}

func initConfig() error {
	if globalOpts.Config != "" {
		viper.SetConfigFile(globalOpts.Config)
	} else {
		viper.AddConfigPath(paths.ConfigDir())
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	// Read config file if it exists (not an error if missing)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Apply config defaults
	if globalOpts.Output == "text" {
		if viper.IsSet("default_output") {
			globalOpts.Output = viper.GetString("default_output")
		}
	}

	return nil
}

// needsProject returns true if the command requires a project context.
func needsProject(cmd *cobra.Command) bool {
	// Version command doesn't need project
	if cmd.Name() == "version" {
		return false
	}
	// All session commands need project
	if cmd.Parent() != nil && cmd.Parent().Name() == "session" {
		return true
	}
	// All team commands need project
	if cmd.Parent() != nil && cmd.Parent().Name() == "team" {
		return true
	}
	// All manifest commands need project
	if cmd.Parent() != nil && cmd.Parent().Name() == "manifest" {
		return true
	}
	// All sync commands need project
	if cmd.Parent() != nil && cmd.Parent().Name() == "sync" {
		return true
	}
	// All validate commands need project
	if cmd.Parent() != nil && cmd.Parent().Name() == "validate" {
		return true
	}
	// All handoff commands need project
	if cmd.Parent() != nil && cmd.Parent().Name() == "handoff" {
		return true
	}
	// All worktree commands need project
	if cmd.Parent() != nil && cmd.Parent().Name() == "worktree" {
		return true
	}
	// Hook commands do NOT require project (they handle missing project gracefully)
	if cmd.Parent() != nil && cmd.Parent().Name() == "hook" {
		return false
	}
	if cmd.Name() == "hook" {
		return false
	}
	return true
}

// GetGlobalOptions returns the global options (for use by subcommands).
func GetGlobalOptions() *GlobalOptions {
	return &globalOpts
}
