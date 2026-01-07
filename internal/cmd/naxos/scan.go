package naxos

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/naxos"
)

type scanOptions struct {
	inactiveThreshold   time.Duration
	staleSailsThreshold time.Duration
	includeArchived     bool
}

func newScanCmd(ctx *cmdContext) *cobra.Command {
	var opts scanOptions

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for orphaned sessions",
		Long: `Scans the sessions directory for orphaned sessions that may need cleanup.

Detection criteria:
  1. Inactive sessions: ACTIVE status but no activity for >24h (default)
  2. Stale gray sails: PARKED with gray/unknown sails for >7d (default)
  3. Incomplete wraps: Sessions marked for wrap but never completed

The scan produces a report with suggested actions. No automatic cleanup
is performed - the user determines what action to take.

Examples:
  ari naxos scan                         # Default thresholds
  ari naxos scan --inactive-threshold 12h    # More aggressive inactive check
  ari naxos scan --stale-threshold 14d       # Longer stale sails threshold
  ari naxos scan --include-archived          # Also scan archived sessions
  ari naxos scan --json                      # Machine-readable output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(ctx, opts)
		},
	}

	cmd.Flags().DurationVar(&opts.inactiveThreshold, "inactive-threshold", 24*time.Hour,
		"How long a session can be inactive before flagging")
	cmd.Flags().DurationVar(&opts.staleSailsThreshold, "stale-threshold", 7*24*time.Hour,
		"How long gray sails can persist before flagging")
	cmd.Flags().BoolVar(&opts.includeArchived, "include-archived", false,
		"Include archived sessions in scan")

	return cmd
}

func runScan(ctx *cmdContext, opts scanOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Build scan configuration
	config := naxos.ScanConfig{
		InactiveThreshold:   opts.inactiveThreshold,
		StaleSailsThreshold: opts.staleSailsThreshold,
		IncludeArchived:     opts.includeArchived,
	}

	// Create scanner and run
	scanner := naxos.NewScanner(resolver, config)
	result, err := scanner.Scan()
	if err != nil {
		return err
	}

	// Convert to output format and print
	output := naxos.FromScanResult(result)
	return printer.Print(output)
}
