// Package procession implements the ari procession commands for cross-rite workflow management.
package procession

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	procmena "github.com/autom8y/knossos/internal/materialize/procession"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// recedeOptions holds options for the recede command.
type recedeOptions struct {
	to string
}

// newRecedeCmd creates the procession recede subcommand.
func newRecedeCmd(ctx *cmdContext) *cobra.Command {
	var opts recedeOptions

	cmd := &cobra.Command{
		Use:   "recede",
		Short: "Move back to an earlier station",
		Long: `Move the procession back to a named earlier station. The completed_stations
log is NOT modified — it is append-only. This means recede is a position change,
not a state rollback. Use this when a station needs to be re-executed (e.g.,
when validation fails and remediation must be retried).

The --to station must:
  - Exist in the template
  - Appear before the current station in the template order

Examples:
  ari procession recede --to=remediate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecede(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.to, "to", "", "Station name to recede to (required)")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

// recedeOutput represents the output of procession recede.
type recedeOutput struct {
	NewCurrentStation string `json:"new_current_station"`
	NewCurrentRite    string `json:"new_current_rite"`
	NextStation       string `json:"next_station,omitempty"`
	NextRite          string `json:"next_rite,omitempty"`
}

// Text implements output.Textable for recedeOutput.
func (o recedeOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Receded to station: %s (%s)\n", o.NewCurrentStation, o.NewCurrentRite)
	if o.NextStation != "" {
		fmt.Fprintf(&b, "Next station:       %s (%s)\n", o.NextStation, o.NextRite)
	} else {
		b.WriteString("Next station:       (final)\n")
	}
	return b.String()
}

var _ output.Textable = recedeOutput{}

// runRecede executes the procession recede command.
func runRecede(ctx *cmdContext, opts recedeOptions) error {
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

	// Resolve template through the 5-tier resolution chain
	rp, err := procmena.ResolveTemplate(p.Type, resolver.ProjectRoot(), common.EmbeddedProcessions())
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeFileNotFound, "template resolution failed", err))
	}
	tmpl := rp.Template

	// Validate --to station exists in template
	targetStn := tmpl.GetStation(opts.to)
	if targetStn == nil {
		return common.PrintAndReturn(printer, errors.New(errors.CodeUsageError,
			fmt.Sprintf("station %q does not exist in procession template %q", opts.to, p.Type)))
	}

	// Validate --to station comes before current station in template order
	stationNames := tmpl.StationNames()
	toIdx := -1
	currentIdx := -1
	for i, name := range stationNames {
		if name == opts.to {
			toIdx = i
		}
		if name == p.CurrentStation {
			currentIdx = i
		}
	}

	// Handle the case where current station is "" (procession just completed)
	// In that case, currentIdx == -1 and any valid station is a valid recede target.
	if currentIdx != -1 && toIdx >= currentIdx {
		return common.PrintAndReturn(printer, errors.New(errors.CodeUsageError,
			fmt.Sprintf("recede --to=%q must be before current station %q in template order",
				opts.to, p.CurrentStation)))
	}

	// Set current_station to the target (position change only; completed_stations unchanged)
	p.CurrentStation = opts.to

	// Recompute next station and rite
	nextStationName := tmpl.NextStation(p.CurrentStation)
	p.NextStation = nextStationName
	p.NextRite = ""
	if nextStationName != "" {
		if ns := tmpl.GetStation(nextStationName); ns != nil {
			p.NextRite = ns.Rite
		}
	}

	// Save session context
	if err := sessCtx.Save(ctxPath); err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to save session context", err))
	}

	return printer.Print(recedeOutput{
		NewCurrentStation: p.CurrentStation,
		NewCurrentRite:    targetStn.Rite,
		NextStation:       p.NextStation,
		NextRite:          p.NextRite,
	})
}
