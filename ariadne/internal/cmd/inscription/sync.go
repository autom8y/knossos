package inscription

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/inscription"
	"github.com/autom8y/ariadne/internal/output"
)

type syncOptions struct {
	force    bool
	dryRun   bool
	noBackup bool
	rite     string
}

func newSyncCmd(ctx *cmdContext) *cobra.Command {
	var opts syncOptions

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize CLAUDE.md with templates",
		Long: `Synchronize CLAUDE.md with Knossos templates and project state.

This command:
  1. Loads the KNOSSOS_MANIFEST.yaml (or creates a default)
  2. Generates content for knossos-owned and regenerate regions
  3. Merges with existing CLAUDE.md, preserving satellite regions
  4. Creates a backup and writes the updated file
  5. Updates manifest with new hashes and version

Region ownership determines sync behavior:
  - knossos: Always overwritten from templates
  - satellite: Never overwritten (user-owned)
  - regenerate: Regenerated from project state

Examples:
  ari inscription sync              # Normal sync
  ari inscription sync --dry-run    # Preview changes
  ari inscription sync --force      # Force full regeneration
  ari inscription sync --rite foo   # Sync with specific rite`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(ctx, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force full regeneration regardless of hashes")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without writing")
	cmd.Flags().BoolVar(&opts.noBackup, "no-backup", false, "Skip backup creation")
	cmd.Flags().StringVar(&opts.rite, "rite", "", "Rite name to sync for")

	return cmd
}

func runSync(ctx *cmdContext, opts syncOptions) error {
	printer := ctx.getPrinter()
	pipeline := ctx.getPipeline()

	syncOpts := inscription.SyncOptions{
		Force:    opts.force,
		RiteName: opts.rite,
		DryRun:   opts.dryRun,
		NoBackup: opts.noBackup,
	}

	if opts.dryRun {
		return runSyncDryRun(ctx, pipeline, syncOpts, printer)
	}

	result, err := pipeline.Sync(syncOpts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	out := SyncOutput{
		Success:            result.Success,
		RegionsSynced:      result.RegionsSynced,
		BackupPath:         result.BackupPath,
		Duration:           result.Duration.String(),
		InscriptionVersion: result.InscriptionVersion,
	}

	if len(result.Conflicts) > 0 {
		out.Conflicts = make([]ConflictOutput, len(result.Conflicts))
		for i, c := range result.Conflicts {
			out.Conflicts[i] = ConflictOutput{
				Region:    c.Region,
				Type:      string(c.Type),
				Message:   c.Message,
				Preserved: c.Preserved,
			}
		}
	}

	return printer.Print(out)
}

func runSyncDryRun(ctx *cmdContext, pipeline *inscription.Pipeline, opts inscription.SyncOptions, printer *output.Printer) error {
	preview, err := pipeline.DryRun(opts)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	out := SyncPreviewOutput{
		DryRun:         true,
		WouldSync:      preview.WouldSync,
		WouldPreserve:  preview.WouldPreserve,
		CurrentVersion: preview.CurrentVersion,
		NewVersion:     preview.NewVersion,
	}

	if len(preview.Conflicts) > 0 {
		out.Conflicts = make([]ConflictOutput, len(preview.Conflicts))
		for i, c := range preview.Conflicts {
			out.Conflicts[i] = ConflictOutput{
				Region:    c.Region,
				Type:      string(c.Type),
				Message:   c.Message,
				Preserved: c.Preserved,
			}
		}
	}

	return printer.Print(out)
}

// SyncOutput represents sync result for output.
type SyncOutput struct {
	Success            bool             `json:"success"`
	RegionsSynced      []string         `json:"regions_synced"`
	Conflicts          []ConflictOutput `json:"conflicts,omitempty"`
	BackupPath         string           `json:"backup_path,omitempty"`
	Duration           string           `json:"duration"`
	InscriptionVersion string           `json:"inscription_version"`
}

// Text implements output.Textable for SyncOutput.
func (s SyncOutput) Text() string {
	var b strings.Builder

	if s.Success {
		b.WriteString(fmt.Sprintf("Synced CLAUDE.md (v%s)\n", s.InscriptionVersion))
		b.WriteString(fmt.Sprintf("Regions updated: %d\n", len(s.RegionsSynced)))

		if len(s.RegionsSynced) > 0 {
			b.WriteString("  - ")
			b.WriteString(strings.Join(s.RegionsSynced, "\n  - "))
			b.WriteString("\n")
		}

		if len(s.Conflicts) > 0 {
			b.WriteString(fmt.Sprintf("\nConflicts: %d\n", len(s.Conflicts)))
			for _, c := range s.Conflicts {
				icon := "!"
				if c.Preserved {
					icon = "~"
				}
				b.WriteString(fmt.Sprintf("  %s %s: %s\n", icon, c.Region, c.Message))
			}
		}

		if s.BackupPath != "" {
			b.WriteString(fmt.Sprintf("\nBackup: %s\n", s.BackupPath))
		}

		b.WriteString(fmt.Sprintf("Duration: %s\n", s.Duration))
	} else {
		b.WriteString("Sync failed\n")
	}

	return b.String()
}

// SyncPreviewOutput represents dry-run preview for output.
type SyncPreviewOutput struct {
	DryRun         bool             `json:"dry_run"`
	WouldSync      []string         `json:"would_sync"`
	WouldPreserve  []string         `json:"would_preserve"`
	Conflicts      []ConflictOutput `json:"conflicts,omitempty"`
	CurrentVersion string           `json:"current_version"`
	NewVersion     string           `json:"new_version"`
}

// Text implements output.Textable for SyncPreviewOutput.
func (s SyncPreviewOutput) Text() string {
	var b strings.Builder

	b.WriteString("=== DRY RUN (no changes made) ===\n\n")

	b.WriteString(fmt.Sprintf("Current version: %s -> New version: %s\n\n", s.CurrentVersion, s.NewVersion))

	if len(s.WouldSync) > 0 {
		b.WriteString(fmt.Sprintf("Would sync %d regions:\n", len(s.WouldSync)))
		for _, r := range s.WouldSync {
			b.WriteString(fmt.Sprintf("  + %s\n", r))
		}
	}

	if len(s.WouldPreserve) > 0 {
		b.WriteString(fmt.Sprintf("\nWould preserve %d regions:\n", len(s.WouldPreserve)))
		for _, r := range s.WouldPreserve {
			b.WriteString(fmt.Sprintf("  ~ %s\n", r))
		}
	}

	if len(s.Conflicts) > 0 {
		b.WriteString(fmt.Sprintf("\nConflicts detected: %d\n", len(s.Conflicts)))
		for _, c := range s.Conflicts {
			action := "overwrite"
			if c.Preserved {
				action = "preserve"
			}
			b.WriteString(fmt.Sprintf("  ! %s: %s (will %s)\n", c.Region, c.Message, action))
		}
	}

	return b.String()
}

// ConflictOutput represents a conflict for output.
type ConflictOutput struct {
	Region    string `json:"region"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Preserved bool   `json:"preserved"`
}

// BackupOutput represents a backup entry for output.
type BackupOutput struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
}
