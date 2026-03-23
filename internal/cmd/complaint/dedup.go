package complaint

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/fileutil"
)

// dedupOptions holds flag values for the dedup subcommand.
type dedupOptions struct {
	dryRun bool
}

func newDedupCmd(ctx *cmdContext) *cobra.Command {
	var opts dedupOptions

	cmd := &cobra.Command{
		Use:   "dedup",
		Short: "Deduplicate complaint corpus",
		Long: `Collapse duplicate complaints by grouping on title-prefix pattern.

Keeps one representative per group and removes the rest. Use --dry-run
to preview what would be collapsed without making changes.

Examples:
  ari complaint dedup --dry-run
  ari complaint dedup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDedup(ctx, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false,
		"Preview dedup without making changes")

	return cmd
}

// dedupGroup represents a group of duplicate complaints.
type dedupGroup struct {
	Key            string   `json:"key"`
	Representative string   `json:"representative"`
	Members        []string `json:"members"`
	Count          int      `json:"count"`
}

// dedupOutput is the structured response for a dedup operation.
type dedupOutput struct {
	Groups      []dedupGroup `json:"groups"`
	TotalBefore int          `json:"total_before"`
	TotalAfter  int          `json:"total_after"`
	Removed     int          `json:"removed"`
	DryRun      bool         `json:"dry_run"`
}

// Text implements output.Textable.
func (o dedupOutput) Text() string {
	var b strings.Builder
	mode := "Dedup complete"
	if o.DryRun {
		mode = "Dedup dry run"
	}
	fmt.Fprintf(&b, "%s: %d complaints → %d groups (%d would be removed)\n\n",
		mode, o.TotalBefore, o.TotalAfter, o.Removed)

	for _, g := range o.Groups {
		if g.Count > 1 {
			fmt.Fprintf(&b, "  [%d] %s (keep: %s)\n", g.Count, g.Key, g.Representative)
		}
	}
	return b.String()
}

func runDedup(ctx *cmdContext, opts dedupOptions) error {
	printer := ctx.getPrinter()

	complaintsDir := resolveComplaintsDir(*ctx.ProjectDir)
	complaints, err := loadComplaints(complaintsDir)
	if err != nil || len(complaints) == 0 {
		return printer.Print(dedupOutput{DryRun: opts.dryRun})
	}

	// Group complaints by dedup key: filed_by + title-prefix pattern.
	groups := groupComplaints(complaints)

	totalBefore := len(complaints)
	totalAfter := len(groups)
	removed := totalBefore - totalAfter

	result := dedupOutput{
		TotalBefore: totalBefore,
		TotalAfter:  totalAfter,
		Removed:     removed,
		DryRun:      opts.dryRun,
	}

	for key, members := range groups {
		// Sort by filed_at ascending — keep the earliest as representative.
		sort.Slice(members, func(i, j int) bool {
			return members[i].FiledAt < members[j].FiledAt
		})

		memberIDs := make([]string, len(members))
		for i, m := range members {
			memberIDs[i] = m.ID
		}

		result.Groups = append(result.Groups, dedupGroup{
			Key:            key,
			Representative: members[0].ID,
			Members:        memberIDs,
			Count:          len(members),
		})
	}

	// Sort groups by count descending for output.
	sort.Slice(result.Groups, func(i, j int) bool {
		return result.Groups[i].Count > result.Groups[j].Count
	})

	if !opts.dryRun {
		// Remove duplicates, keeping the representative.
		for _, g := range result.Groups {
			if g.Count <= 1 {
				continue
			}
			for _, id := range g.Members[1:] {
				// Rewrite duplicate as merged: update status to resolved and add note.
				path, err := findComplaintFile(complaintsDir, id)
				if err != nil {
					continue
				}
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				var c Complaint
				if err := yaml.Unmarshal(data, &c); err != nil {
					continue
				}
				c.Status = "resolved"
				c.Description = fmt.Sprintf("Deduplicated into %s.\n\nOriginal: %s", g.Representative, c.Description)
				updated, err := yaml.Marshal(&c)
				if err != nil {
					continue
				}
				_ = fileutil.AtomicWriteFile(path, updated, 0644)
			}
		}
	}

	return printer.Print(result)
}

// groupComplaints groups complaints by a dedup key derived from filed_by and title prefix.
// Drift-detector complaints group by tag pattern (tool-fallback, retry-spiral, command-exploration).
// Agent-filed complaints group by title similarity (first 40 chars).
func groupComplaints(complaints []Complaint) map[string][]Complaint {
	groups := make(map[string][]Complaint)
	for _, c := range complaints {
		key := complaintDedupKey(c)
		groups[key] = append(groups[key], c)
	}
	return groups
}

// complaintDedupKey computes a grouping key for a complaint.
func complaintDedupKey(c Complaint) string {
	if c.FiledBy == "drift-detector" {
		// Group by drift pattern tag.
		for _, tag := range c.Tags {
			switch tag {
			case "tool-fallback", "retry-spiral", "command-exploration":
				return c.FiledBy + ":" + tag
			}
		}
		return c.FiledBy + ":unknown"
	}
	// Agent-filed: group by title prefix (first 40 chars).
	title := c.Title
	if len(title) > 40 {
		title = title[:40]
	}
	return c.FiledBy + ":" + title
}

// MarshalJSON implements custom JSON for dedup output.
func (o dedupOutput) MarshalJSON() ([]byte, error) {
	groups := o.Groups
	if groups == nil {
		groups = []dedupGroup{}
	}
	return json.Marshal(struct {
		Groups      []dedupGroup `json:"groups"`
		TotalBefore int          `json:"total_before"`
		TotalAfter  int          `json:"total_after"`
		Removed     int          `json:"removed"`
		DryRun      bool         `json:"dry_run"`
	}{Groups: groups, TotalBefore: o.TotalBefore, TotalAfter: o.TotalAfter, Removed: o.Removed, DryRun: o.DryRun})
}
