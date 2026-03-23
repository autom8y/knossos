package complaint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/fileutil"
)

// validStatuses mirrors complaint_status_enum from complaint.schema.json.
var validStatuses = map[string]bool{
	"filed":    true,
	"triaged":  true,
	"accepted": true,
	"rejected": true,
	"resolved": true,
}

// updateOptions holds flag values for the update subcommand.
type updateOptions struct {
	status string
	id     string
}

func newUpdateCmd(ctx *cmdContext) *cobra.Command {
	var opts updateOptions

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update complaint status",
		Long: `Update the status of a complaint in .sos/wip/complaints/.

Validates the new status against the complaint schema enum:
  filed, triaged, accepted, rejected, resolved

Examples:
  ari complaint update --id=COMPLAINT-20260311-143022-drift-detect --status=triaged
  ari complaint update --id=COMPLAINT-20260311-091500-pythia --status=resolved`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.status, "status", "",
		"New status (filed, triaged, accepted, rejected, resolved)")
	cmd.Flags().StringVar(&opts.id, "id", "",
		"Complaint ID to update")

	_ = cmd.MarkFlagRequired("status")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// updateOutput is the structured response for a successful update.
type updateOutput struct {
	ID        string `json:"id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	Path      string `json:"path"`
}

// Text implements output.Textable.
func (o updateOutput) Text() string {
	return fmt.Sprintf("Updated %s: %s → %s\n", o.ID, o.OldStatus, o.NewStatus)
}

func runUpdate(ctx *cmdContext, opts updateOptions) error {
	printer := ctx.getPrinter()

	// Validate status against schema enum.
	if !validStatuses[opts.status] {
		valid := make([]string, 0, len(validStatuses))
		for k := range validStatuses {
			valid = append(valid, k)
		}
		return fmt.Errorf("invalid status %q: must be one of %s (per complaint.schema.json complaint_status_enum)",
			opts.status, strings.Join(valid, ", "))
	}

	// Resolve complaints directory.
	complaintsDir := resolveComplaintsDir(*ctx.ProjectDir)

	// Find the complaint file by ID.
	filePath, err := findComplaintFile(complaintsDir, opts.id)
	if err != nil {
		return err
	}

	// Read and parse the complaint.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading complaint %s: %w", opts.id, err)
	}

	var complaint Complaint
	if err := yaml.Unmarshal(data, &complaint); err != nil {
		return fmt.Errorf("parsing complaint %s: %w", opts.id, err)
	}

	oldStatus := complaint.Status
	if oldStatus == opts.status {
		return printer.Print(updateOutput{
			ID:        opts.id,
			OldStatus: oldStatus,
			NewStatus: opts.status,
			Path:      filePath,
		})
	}

	// Update status and write back atomically.
	complaint.Status = opts.status
	updated, err := yaml.Marshal(&complaint)
	if err != nil {
		return fmt.Errorf("marshaling complaint %s: %w", opts.id, err)
	}

	if err := fileutil.AtomicWriteFile(filePath, updated, 0644); err != nil {
		return fmt.Errorf("writing complaint %s: %w", opts.id, err)
	}

	return printer.Print(updateOutput{
		ID:        opts.id,
		OldStatus: oldStatus,
		NewStatus: opts.status,
		Path:      filePath,
	})
}

// findComplaintFile locates a complaint YAML file by ID in the complaints directory.
// Checks both .yaml and .yml extensions.
func findComplaintFile(dir, id string) (string, error) {
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(dir, id+ext)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("complaint %q not found in %s", id, dir)
}
