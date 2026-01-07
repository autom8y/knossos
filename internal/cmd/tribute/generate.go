package tribute

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/tribute"
)

// GenerateOutput represents the output of tribute generation.
type GenerateOutput struct {
	SessionID   string  `json:"session_id"`
	FilePath    string  `json:"file_path"`
	Initiative  string  `json:"initiative"`
	Complexity  string  `json:"complexity"`
	Duration    string  `json:"duration"`
	Rite        string  `json:"rite"`
	FinalPhase  string  `json:"final_phase"`
	SailsColor  string  `json:"sails_color,omitempty"`
	Artifacts   int     `json:"artifacts_count"`
	Decisions   int     `json:"decisions_count"`
	Phases      int     `json:"phases_count"`
	Handoffs    int     `json:"handoffs_count"`
	GeneratedAt string  `json:"generated_at"`
}

// Text implements output.Textable for GenerateOutput.
func (g GenerateOutput) Text() string {
	return fmt.Sprintf("Generated TRIBUTE.md for session %s\n"+
		"Path: %s\n"+
		"Initiative: %s\n"+
		"Duration: %s\n"+
		"Sails: %s\n",
		g.SessionID, g.FilePath, g.Initiative, g.Duration, g.SailsColor)
}

// newGenerateCmd creates the generate subcommand.
func newGenerateCmd(ctx *cmdContext) *cobra.Command {
	var sessionDir string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate TRIBUTE.md for a session",
		Long: `Generates a summary document (TRIBUTE.md) for a completed or active session.

The tribute summarizes:
- Session metadata (initiative, complexity, duration)
- Artifacts produced
- Decisions made
- Phase progression
- Agent handoffs
- White Sails attestation
- Session metrics

Examples:
  # Generate tribute for current session
  ari tribute generate

  # Generate tribute for a specific session
  ari tribute generate --session-id session-20260106-123456-abcd1234

  # Generate tribute for a session directory
  ari tribute generate --session-dir /path/to/session`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(ctx, sessionDir)
		},
	}

	cmd.Flags().StringVar(&sessionDir, "session-dir", "", "Path to session directory (overrides session lookup)")

	return cmd
}

// runGenerate executes tribute generation.
func runGenerate(ctx *cmdContext, sessionDir string) error {
	printer := ctx.getPrinter()

	var generator *tribute.Generator
	var err error

	if sessionDir != "" {
		// Use explicit session directory
		generator = tribute.NewGenerator(sessionDir)
	} else if ctx.SessionID != nil && *ctx.SessionID != "" {
		// Use session ID to find session
		if ctx.ProjectDir == nil || *ctx.ProjectDir == "" {
			return errors.New(errors.CodeProjectNotFound, "project directory required when using session ID")
		}
		generator, err = tribute.GenerateFromSessionID(*ctx.ProjectDir, *ctx.SessionID)
		if err != nil {
			printer.PrintError(err)
			return err
		}
	} else {
		// Use current session
		if ctx.ProjectDir == nil || *ctx.ProjectDir == "" {
			return errors.New(errors.CodeProjectNotFound, "project directory required")
		}
		generator, err = tribute.GenerateFromProject(*ctx.ProjectDir)
		if err != nil {
			printer.PrintError(err)
			return err
		}
	}

	printer.VerboseLog("info", "generating tribute", map[string]interface{}{
		"session_path": generator.SessionPath,
	})

	result, err := generator.Generate()
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Build output
	out := GenerateOutput{
		SessionID:   result.SessionID,
		FilePath:    result.FilePath,
		Initiative:  result.Initiative,
		Complexity:  result.Complexity,
		Duration:    formatDuration(result.Duration),
		Rite:        result.Rite,
		FinalPhase:  result.FinalPhase,
		SailsColor:  result.SailsColor,
		Artifacts:   len(result.Artifacts),
		Decisions:   len(result.Decisions),
		Phases:      len(result.Phases),
		Handoffs:    len(result.Handoffs),
		GeneratedAt: result.GeneratedAt.Format("2006-01-02T15:04:05Z"),
	}

	printer.VerboseLog("info", "tribute generated", map[string]interface{}{
		"file_path":  result.FilePath,
		"session_id": result.SessionID,
	})

	return printer.Print(out)
}

// formatDuration formats duration as human-readable string.
func formatDuration(d interface{}) string {
	switch v := d.(type) {
	case interface{ Hours() float64 }:
		hours := v.Hours()
		if hours < 1 {
			return fmt.Sprintf("%.0fm", hours*60)
		}
		return fmt.Sprintf("%.1fh", hours)
	default:
		return "unknown"
	}
}
