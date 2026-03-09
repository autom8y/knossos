package naxos

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/naxos"
)

type triageOptions struct {
	inactiveThreshold   time.Duration
	staleSailsThreshold time.Duration
	includeArchived     bool
	writeArtifact       bool
}

func newTriageCmd(ctx *cmdContext) *cobra.Command {
	var opts triageOptions

	cmd := &cobra.Command{
		Use:   "triage",
		Short: "Scan and classify orphaned sessions by severity",
		Long: `Scans sessions and classifies each orphaned session by severity (CRITICAL, HIGH, MEDIUM, LOW).

Severity rules:
  CRITICAL  INACTIVE >30 days, or INCOMPLETE_WRAP (any age)
  HIGH      INACTIVE >7 days, or STALE_SAILS >14 days
  MEDIUM    INACTIVE >24h, or STALE_SAILS >7 days
  LOW       Anything else

By default, triage writes a NAXOS_TRIAGE.md artifact to the sessions directory.
This artifact can be read by hooks and other tooling for fast summaries.

Examples:
  ari naxos triage                            # Default thresholds, writes artifact
  ari naxos triage --no-artifact              # Report only, no artifact written
  ari naxos triage --inactive-threshold 12h   # More aggressive inactive check
  ari naxos triage --stale-threshold 14d      # Longer stale sails threshold
  ari naxos triage --json                     # Machine-readable output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTriage(ctx, opts)
		},
	}

	cmd.Flags().DurationVar(&opts.inactiveThreshold, "inactive-threshold", 24*time.Hour,
		"How long a session can be inactive before flagging")
	cmd.Flags().DurationVar(&opts.staleSailsThreshold, "stale-threshold", 7*24*time.Hour,
		"How long gray sails can persist before flagging")
	cmd.Flags().BoolVar(&opts.includeArchived, "include-archived", false,
		"Include archived sessions in scan")
	cmd.Flags().BoolVar(&opts.writeArtifact, "no-artifact", false,
		"Skip writing NAXOS_TRIAGE.md artifact")

	return cmd
}

func runTriage(ctx *cmdContext, opts triageOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Build scan configuration.
	config := naxos.ScanConfig{
		InactiveThreshold:   opts.inactiveThreshold,
		StaleSailsThreshold: opts.staleSailsThreshold,
		IncludeArchived:     opts.includeArchived,
	}

	// Scan for orphaned sessions.
	scanner := naxos.NewScanner(resolver, config)
	scanResult, err := scanner.Scan()
	if err != nil {
		return err
	}

	// Triage: classify by severity and sort.
	triageResult := naxos.Triage(scanResult)

	// Write artifact unless suppressed.
	if !opts.writeArtifact {
		if err := naxos.WriteTriageArtifact(resolver.SessionsDir(), triageResult); err != nil {
			// Non-fatal: report the error but continue to print the result.
			printer.PrintLine("Warning: could not write triage artifact: " + err.Error())
		}
	}

	// Print the triage output.
	output := naxos.FromTriageResult(triageResult)
	return printer.Print(output)
}
