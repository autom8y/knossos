package complaint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/output"
)

// Complaint represents a parsed complaint YAML artifact.
type Complaint struct {
	ID          string   `json:"id"       yaml:"id"`
	FiledBy     string   `json:"filed_by" yaml:"filed_by"`
	FiledAt     string   `json:"filed_at" yaml:"filed_at"`
	Title       string   `json:"title"    yaml:"title"`
	Severity    string   `json:"severity" yaml:"severity"`
	Description string   `json:"description" yaml:"description"`
	Tags        []string `json:"tags"     yaml:"tags"`
	Status      string   `json:"status"   yaml:"status"`

	// Deep-file optional fields
	SuggestedFix   string          `json:"suggested_fix,omitempty"   yaml:"suggested_fix,omitempty"`
	EffortEstimate string          `json:"effort_estimate,omitempty" yaml:"effort_estimate,omitempty"`
	Zone           string          `json:"zone,omitempty"            yaml:"zone,omitempty"`
	RelatedScars   []string        `json:"related_scars,omitempty"   yaml:"related_scars,omitempty"`
	Evidence       *ComplaintEvidence `json:"evidence,omitempty"     yaml:"evidence,omitempty"`
}

// ComplaintEvidence holds optional evidence block from deep-file complaints.
type ComplaintEvidence struct {
	SessionID string   `json:"session_id,omitempty" yaml:"session_id,omitempty"`
	EventRefs []string `json:"event_refs,omitempty" yaml:"event_refs,omitempty"`
	Context   string   `json:"context,omitempty"    yaml:"context,omitempty"`
}

// filedDate extracts the YYYY-MM-DD portion from filed_at (ISO-8601).
// Falls back to the raw string if it is too short to contain a date prefix.
func (c *Complaint) filedDate() string {
	if len(c.FiledAt) >= 10 {
		return c.FiledAt[:10]
	}
	return c.FiledAt
}

// truncateTitle truncates a title to at most maxLen characters, appending "..."
// when truncation occurs.
func truncateTitle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// listOptions holds flag values for the list subcommand.
type listOptions struct {
	status   string
	severity string
}

func newListCmd(ctx *cmdContext) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List filed complaints",
		Long: `List complaints from .sos/wip/complaints/ with optional filtering.

Complaint files are YAML artifacts filed by agents when they encounter
framework friction (CLI gaps, missing skills, broken hooks, routing failures).

Examples:
  ari complaint list
  ari complaint list --status=filed
  ari complaint list --severity=high
  ari complaint list --severity=critical --status=filed
  ari complaint list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.status, "status", "",
		"Filter by status (filed, triaged, accepted, rejected, resolved)")
	cmd.Flags().StringVar(&opts.severity, "severity", "",
		"Filter by severity (low, medium, high, critical)")

	return cmd
}

func runList(ctx *cmdContext, opts listOptions) error {
	printer := ctx.getPrinter()

	// Resolve complaints directory relative to project root.
	// When no project dir is available, fall back to cwd.
	complaintsDir := resolveComplaintsDir(*ctx.ProjectDir)

	complaints, err := loadComplaints(complaintsDir)
	if err != nil {
		// Non-fatal: missing directory is a valid empty state.
		complaints = nil
	}

	// Apply filters.
	complaints = filterComplaints(complaints, opts)

	// Sort by filed_at descending (most recent first).
	sort.Slice(complaints, func(i, j int) bool {
		return complaints[i].FiledAt > complaints[j].FiledAt
	})

	return printComplaints(printer, complaints)
}

// resolveComplaintsDir returns the path to .sos/wip/complaints/ for the given
// project directory. Falls back to cwd when projectDir is empty.
func resolveComplaintsDir(projectDir string) string {
	if projectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ""
		}
		projectDir = cwd
	}
	return filepath.Join(projectDir, ".sos", "wip", "complaints")
}

// loadComplaints reads all YAML complaint files from the given directory.
// It silently skips unparseable files and returns those it can read.
func loadComplaints(dir string) ([]Complaint, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var complaints []Complaint
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		if !strings.HasPrefix(name, "COMPLAINT-") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}

		var c Complaint
		if err := yaml.Unmarshal(data, &c); err != nil {
			continue
		}

		// Backfill ID from filename if the YAML body omits it.
		if c.ID == "" {
			c.ID = strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml")
		}

		complaints = append(complaints, c)
	}

	return complaints, nil
}

// filterComplaints applies status and severity filters to the slice.
func filterComplaints(complaints []Complaint, opts listOptions) []Complaint {
	if opts.status == "" && opts.severity == "" {
		return complaints
	}

	filtered := complaints[:0:0]
	for _, c := range complaints {
		if opts.status != "" && c.Status != opts.status {
			continue
		}
		if opts.severity != "" && c.Severity != opts.severity {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

// printComplaints dispatches to the appropriate format renderer.
// The output type implements both Tabular (for non-empty text tables) and
// Textable (for the empty-state message and the non-empty manual table).
// Since Printer checks Tabular first, we use a dedicated empty-state type
// when there are no complaints to avoid printing orphaned header columns.
func printComplaints(printer *output.Printer, complaints []Complaint) error {
	if len(complaints) == 0 {
		return printer.Print(emptyComplaintOutput{})
	}
	result := complaintListOutput{Complaints: complaints, Total: len(complaints)}
	return printer.Print(result)
}

// emptyComplaintOutput is a text-only output type for the zero-complaint case.
// It does NOT implement Tabular so the Printer routes through Textable instead,
// producing the human-friendly "No complaints found." message.
// For JSON consumers it still produces a structured empty array via MarshalJSON.
type emptyComplaintOutput struct{}

// Text implements output.Textable.
func (e emptyComplaintOutput) Text() string {
	return "No complaints found.\n"
}

// MarshalJSON produces the canonical empty-list response for JSON consumers.
func (e emptyComplaintOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Complaints []Complaint `json:"complaints"`
		Total      int         `json:"total"`
	}{Complaints: []Complaint{}, Total: 0})
}

// complaintListOutput is the structured output type for complaint list.
type complaintListOutput struct {
	Complaints []Complaint `json:"complaints"`
	Total      int         `json:"total"`
}

// Headers implements output.Tabular for text table rendering.
func (o complaintListOutput) Headers() []string {
	return []string{"ID", "SEVERITY", "TITLE", "STATUS", "FILED"}
}

// Rows implements output.Tabular for text table rendering.
func (o complaintListOutput) Rows() [][]string {
	if len(o.Complaints) == 0 {
		return [][]string{}
	}
	rows := make([][]string, len(o.Complaints))
	for i, c := range o.Complaints {
		rows[i] = []string{
			c.ID,
			c.Severity,
			truncateTitle(c.Title, 40),
			c.Status,
			c.filedDate(),
		}
	}
	return rows
}

// MarshalJSON implements custom JSON marshaling so that an empty complaint
// list renders as [] rather than null.
func (o complaintListOutput) MarshalJSON() ([]byte, error) {
	type alias complaintListOutput
	complaints := o.Complaints
	if complaints == nil {
		complaints = []Complaint{}
	}
	return json.Marshal(struct {
		Complaints []Complaint `json:"complaints"`
		Total      int         `json:"total"`
	}{Complaints: complaints, Total: o.Total})
}

// Text implements output.Textable for the empty-state message.
// The Tabular interface handles non-empty cases; this only fires
// when len(Complaints) == 0 AND the Tabular path produces no rows —
// but since Printer checks Tabular first, we need a fallback for
// the "no complaints" message. We achieve this by making
// complaintListOutput implement Textable as a fallback for empty state.
func (o complaintListOutput) Text() string {
	if len(o.Complaints) == 0 {
		return "No complaints found.\n"
	}
	// Non-empty: build table manually to keep column alignment.
	var b strings.Builder
	headers := o.Headers()
	rows := o.Rows()

	// Compute column widths.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Header row.
	for i, h := range headers {
		if i > 0 {
			b.WriteString("  ")
		}
		fmt.Fprintf(&b, "%-*s", widths[i], h)
	}
	b.WriteString("\n")

	// Data rows.
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				b.WriteString("  ")
			}
			fmt.Fprintf(&b, "%-*s", widths[i], cell)
		}
		b.WriteString("\n")
	}

	return b.String()
}
