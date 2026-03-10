// Package procession implements the ari procession commands for cross-rite workflow management.
package procession

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/procession"
	"github.com/autom8y/knossos/internal/session"
)

// proceedOptions holds options for the proceed command.
type proceedOptions struct {
	artifacts string // comma-separated list of artifact paths
}

// newProceedCmd creates the procession proceed subcommand.
func newProceedCmd(ctx *cmdContext) *cobra.Command {
	var opts proceedOptions

	cmd := &cobra.Command{
		Use:   "proceed",
		Short: "Advance to the next station",
		Long: `Advance the procession to the next station, recording the current station
as completed. The current station is appended to completed_stations with a
timestamp and optional artifact paths.

If there is no next station, the procession is complete. The procession block
remains in the session context with an empty next_station field.

Examples:
  ari procession proceed
  ari procession proceed --artifacts=.sos/wip/sr/HANDOFF-audit-to-assess.md
  ari procession proceed --artifacts=path1.md,path2.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProceed(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.artifacts, "artifacts", "",
		"Comma-separated artifact paths produced during this station")

	return cmd
}

// proceedOutput represents the output of procession proceed.
type proceedOutput struct {
	CompletedStation string   `json:"completed_station"`
	CompletedRite    string   `json:"completed_rite"`
	CompletedAt      string   `json:"completed_at"`
	Artifacts        []string `json:"artifacts,omitempty"`
	NewCurrentStation string  `json:"new_current_station,omitempty"`
	NextStation      string   `json:"next_station,omitempty"`
	NextRite         string   `json:"next_rite,omitempty"`
	Complete         bool     `json:"complete"`
}

// Text implements output.Textable for proceedOutput.
func (o proceedOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Station completed: %s (%s)\n", o.CompletedStation, o.CompletedRite)
	fmt.Fprintf(&b, "Completed at:      %s\n", o.CompletedAt)
	for _, a := range o.Artifacts {
		fmt.Fprintf(&b, "  artifact: %s\n", a)
	}
	if o.Complete {
		b.WriteString("\nProcession complete. All stations finished.\n")
	} else {
		fmt.Fprintf(&b, "\nCurrent station: %s\n", o.NewCurrentStation)
		if o.NextRite != "" {
			fmt.Fprintf(&b, "Next station:    %s (%s)\n", o.NextStation, o.NextRite)
			fmt.Fprintf(&b, "\nTo switch rites: ari sync --rite %s\n", o.NextRite)
		} else {
			b.WriteString("Next station:    (final)\n")
		}
	}
	return b.String()
}

var _ output.Textable = proceedOutput{}

// runProceed executes the procession proceed command.
func runProceed(ctx *cmdContext, opts proceedOptions) error {
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

	// Load template to get rite for current station and compute next
	projectDir := resolver.ProjectRoot()
	templatePath := filepath.Join(projectDir, "processions", p.Type+".yaml")
	tmpl, err := procession.LoadTemplate(templatePath)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Get current station details
	currentStn := tmpl.GetStation(p.CurrentStation)
	if currentStn == nil {
		return common.PrintAndReturn(printer, errors.New(errors.CodeSchemaInvalid,
			fmt.Sprintf("current station %q not found in template %q", p.CurrentStation, p.Type)))
	}

	// Parse artifact paths from comma-separated flag
	var artifacts []string
	if opts.artifacts != "" {
		for _, a := range strings.Split(opts.artifacts, ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				artifacts = append(artifacts, trimmed)
			}
		}
	}

	// Append current station to completed_stations (append-only)
	now := time.Now().UTC().Format(time.RFC3339)
	completed := session.CompletedStation{
		Station:     p.CurrentStation,
		Rite:        currentStn.Rite,
		CompletedAt: now,
		Artifacts:   artifacts,
	}
	p.CompletedStations = append(p.CompletedStations, completed)

	// Compute new next station
	nextStationName := tmpl.NextStation(p.CurrentStation)

	// Check if procession is complete
	complete := nextStationName == ""

	if complete {
		// Final station: clear next fields; procession remains in context
		p.CurrentStation = ""
		p.NextStation = ""
		p.NextRite = ""
	} else {
		// Advance current to what was next
		p.CurrentStation = nextStationName

		// Compute the new next (one further ahead)
		newNext := tmpl.NextStation(p.CurrentStation)
		p.NextStation = newNext
		p.NextRite = ""
		if newNext != "" {
			if ns := tmpl.GetStation(newNext); ns != nil {
				p.NextRite = ns.Rite
			}
		}
	}

	// Save session context
	if err := sessCtx.Save(ctxPath); err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to save session context", err))
	}

	out := proceedOutput{
		CompletedStation:  completed.Station,
		CompletedRite:     completed.Rite,
		CompletedAt:       now,
		Artifacts:         artifacts,
		Complete:          complete,
	}
	if !complete {
		out.NewCurrentStation = p.CurrentStation
		out.NextStation = p.NextStation
		out.NextRite = p.NextRite
	}

	return printer.Print(out)
}
