// Package procession implements the ari procession commands for cross-rite workflow management.
package procession

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// newStatusCmd creates the procession status subcommand.
func newStatusCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current procession state",
		Long: `Show the current procession state for the active session,
including the current station, completed stations, next station, and artifact directory.

Examples:
  ari procession status
  ari procession status -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx)
		},
	}
	return cmd
}

// statusOutput represents the output of procession status.
type statusOutput struct {
	ProcessionID       string                    `json:"procession_id"`
	Type               string                    `json:"type"`
	CurrentStation     string                    `json:"current_station"`
	CompletedStations  []session.CompletedStation `json:"completed_stations,omitempty"`
	NextStation        string                    `json:"next_station,omitempty"`
	NextRite           string                    `json:"next_rite,omitempty"`
	ArtifactDir        string                    `json:"artifact_dir"`
}

// Text implements output.Textable for statusOutput.
func (o statusOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Procession: %s\n", o.ProcessionID)
	fmt.Fprintf(&b, "Type:       %s\n", o.Type)
	fmt.Fprintf(&b, "Current:    %s\n", o.CurrentStation)

	if len(o.CompletedStations) > 0 {
		b.WriteString("Completed:\n")
		for _, cs := range o.CompletedStations {
			fmt.Fprintf(&b, "  - %s (%s) at %s\n", cs.Station, cs.Rite, cs.CompletedAt)
			for _, a := range cs.Artifacts {
				fmt.Fprintf(&b, "    artifact: %s\n", a)
			}
		}
	} else {
		b.WriteString("Completed:  (none)\n")
	}

	if o.NextStation != "" {
		fmt.Fprintf(&b, "Next:       %s (%s)\n", o.NextStation, o.NextRite)
	} else {
		b.WriteString("Next:       (final station)\n")
	}
	fmt.Fprintf(&b, "Artifacts:  %s\n", o.ArtifactDir)
	return b.String()
}

var _ output.Textable = statusOutput{}

// runStatus executes the procession status command.
func runStatus(ctx *cmdContext) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Get session ID
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
	}
	if sessionID == "" {
		return common.PrintAndReturn(printer, errors.New(errors.CodeSessionNotFound, "No active session. Use 'ari session create' first."))
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			return common.PrintAndReturn(printer, errors.ErrSessionNotFound(sessionID))
		}
		return common.PrintAndReturn(printer, err)
	}

	// Check procession is active
	if sessCtx.Procession == nil {
		return common.PrintAndReturn(printer, errors.New(errors.CodeUsageError,
			"No active procession. Use 'ari procession create --template=<name>' to start one."))
	}

	p := sessCtx.Procession
	return printer.Print(statusOutput{
		ProcessionID:      p.ID,
		Type:              p.Type,
		CurrentStation:    p.CurrentStation,
		CompletedStations: p.CompletedStations,
		NextStation:       p.NextStation,
		NextRite:          p.NextRite,
		ArtifactDir:       p.ArtifactDir,
	})
}
