// Package output provides format-aware output printing for Ariadne.
// It handles JSON, YAML, and text (table) output formats.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

// Format represents an output format.
type Format string

const (
	// FormatText is human-readable text output.
	FormatText Format = "text"
	// FormatJSON is machine-readable JSON output.
	FormatJSON Format = "json"
	// FormatYAML is YAML output.
	FormatYAML Format = "yaml"
)

// ParseFormat parses a format string.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "yaml":
		return FormatYAML
	default:
		return FormatText
	}
}

// Printer handles formatted output.
type Printer struct {
	format  Format
	out     io.Writer
	errOut  io.Writer
	verbose bool
}

// NewPrinter creates a new printer with the given format.
func NewPrinter(format Format, out, errOut io.Writer, verbose bool) *Printer {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	return &Printer{
		format:  format,
		out:     out,
		errOut:  errOut,
		verbose: verbose,
	}
}

// Print outputs data in the configured format.
func (p *Printer) Print(data interface{}) error {
	switch p.format {
	case FormatJSON:
		return p.printJSON(data)
	case FormatYAML:
		return p.printYAML(data)
	default:
		return p.printText(data)
	}
}

// PrintSuccess outputs a success message.
// For JSON format, wraps the data. For text, silent (per TDD).
func (p *Printer) PrintSuccess(data interface{}) error {
	switch p.format {
	case FormatJSON:
		return p.printJSON(data)
	case FormatYAML:
		return p.printYAML(data)
	default:
		// Silent success for mutations per TDD
		return nil
	}
}

// PrintError outputs an error.
func (p *Printer) PrintError(err error) error {
	if p.format == FormatJSON {
		// Check if error has JSON method
		if jsonErr, ok := err.(interface{ JSON() string }); ok {
			fmt.Fprintln(p.errOut, jsonErr.JSON())
			return nil
		}
		// Wrap in standard error format
		wrapper := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "GENERAL_ERROR",
				"message": err.Error(),
			},
		}
		enc := json.NewEncoder(p.errOut)
		enc.SetIndent("", "  ")
		return enc.Encode(wrapper)
	}
	fmt.Fprintf(p.errOut, "Error: %s\n", err.Error())
	return nil
}

// PrintText outputs raw text (bypasses format switching).
func (p *Printer) PrintText(text string) {
	fmt.Fprint(p.out, text)
}

// PrintLine outputs a text line (bypasses format switching).
func (p *Printer) PrintLine(text string) {
	fmt.Fprintln(p.out, text)
}

// VerboseLog writes a JSON line to stderr for debugging.
func (p *Printer) VerboseLog(level, msg string, fields map[string]interface{}) {
	if !p.verbose {
		return
	}
	entry := map[string]interface{}{
		"level": level,
		"msg":   msg,
		"ts":    time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range fields {
		entry[k] = v
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(p.errOut, string(data))
}

func (p *Printer) printJSON(data interface{}) error {
	enc := json.NewEncoder(p.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (p *Printer) printYAML(data interface{}) error {
	enc := yaml.NewEncoder(p.out)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}

func (p *Printer) printText(data interface{}) error {
	// Handle Tabular interface for table output
	if t, ok := data.(Tabular); ok {
		return p.printTable(t)
	}

	// Handle Textable interface for custom text output
	if t, ok := data.(Textable); ok {
		fmt.Fprintln(p.out, t.Text())
		return nil
	}

	// Fallback to fmt
	fmt.Fprintln(p.out, data)
	return nil
}

// Tabular interface for types that can be rendered as tables.
type Tabular interface {
	Headers() []string
	Rows() [][]string
}

// Textable interface for types with custom text representation.
type Textable interface {
	Text() string
}

func (p *Printer) printTable(t Tabular) error {
	w := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Headers
	headers := t.Headers()
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)

	// Rows
	for _, row := range t.Rows() {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, cell)
		}
		fmt.Fprintln(w)
	}

	return nil
}

// --- Common Output Structures ---

// StatusOutput represents session status for JSON output.
type StatusOutput struct {
	SessionID     string `json:"session_id,omitempty"`
	SessionDir    string `json:"session_dir,omitempty"`
	HasSession    bool   `json:"has_session"`
	Status        string `json:"status"`
	Initiative    string `json:"initiative,omitempty"`
	Complexity    string `json:"complexity,omitempty"`
	CurrentPhase  string `json:"current_phase,omitempty"`
	ActiveTeam    string `json:"active_team,omitempty"`
	ExecutionMode string `json:"execution_mode,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	SchemaVersion string `json:"schema_version,omitempty"`
	GitBranch     string `json:"git_branch,omitempty"`
	GitChanges    int    `json:"git_changes,omitempty"`
}

// Text implements Textable for StatusOutput.
func (s StatusOutput) Text() string {
	if !s.HasSession {
		return "No active session"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Session: %s\n", s.SessionID))
	b.WriteString(fmt.Sprintf("Status: %s\n", s.Status))
	b.WriteString(fmt.Sprintf("Initiative: %s\n", s.Initiative))
	b.WriteString(fmt.Sprintf("Phase: %s\n", s.CurrentPhase))
	b.WriteString(fmt.Sprintf("Team: %s\n", s.ActiveTeam))
	b.WriteString(fmt.Sprintf("Mode: %s\n", s.ExecutionMode))
	if s.GitBranch != "" {
		b.WriteString(fmt.Sprintf("Branch: %s (%d changes)\n", s.GitBranch, s.GitChanges))
	}
	return b.String()
}

// CreateOutput represents session creation result.
type CreateOutput struct {
	SessionID     string `json:"session_id"`
	SessionDir    string `json:"session_dir"`
	Status        string `json:"status"`
	Initiative    string `json:"initiative"`
	Complexity    string `json:"complexity"`
	Team          string `json:"team"`
	CreatedAt     string `json:"created_at"`
	SchemaVersion string `json:"schema_version"`
}

// Text implements Textable for CreateOutput.
func (c CreateOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Created session: %s\n", c.SessionID))
	b.WriteString(fmt.Sprintf("Initiative: %s\n", c.Initiative))
	b.WriteString(fmt.Sprintf("Complexity: %s\n", c.Complexity))
	b.WriteString(fmt.Sprintf("Team: %s\n", c.Team))
	return b.String()
}

// SessionListOutput represents session list for JSON output.
type SessionListOutput struct {
	Sessions       []SessionSummary `json:"sessions"`
	Total          int              `json:"total"`
	CurrentSession string           `json:"current_session,omitempty"`
}

// SessionSummary is a brief session entry for listing.
type SessionSummary struct {
	SessionID  string `json:"session_id"`
	Status     string `json:"status"`
	Initiative string `json:"initiative"`
	Complexity string `json:"complexity"`
	CreatedAt  string `json:"created_at"`
	ParkedAt   string `json:"parked_at,omitempty"`
	Current    bool   `json:"current"`
}

// Headers implements Tabular for SessionListOutput.
func (l SessionListOutput) Headers() []string {
	return []string{"SESSION ID", "STATUS", "INITIATIVE", "CREATED"}
}

// Rows implements Tabular for SessionListOutput.
func (l SessionListOutput) Rows() [][]string {
	rows := make([][]string, len(l.Sessions))
	for i, s := range l.Sessions {
		prefix := "  "
		if s.Current {
			prefix = "* "
		}
		// Extract date from created_at
		date := s.CreatedAt
		if len(date) >= 10 {
			date = date[:10]
		}
		rows[i] = []string{
			prefix + s.SessionID,
			s.Status,
			s.Initiative,
			date,
		}
	}
	return rows
}

// Text implements Textable for SessionListOutput.
func (l SessionListOutput) Text() string {
	if len(l.Sessions) == 0 {
		return "No sessions found"
	}
	var b strings.Builder
	// Let tabular handle headers/rows
	return b.String()
}

// TransitionOutput represents a state transition result.
type TransitionOutput struct {
	SessionID     string `json:"session_id"`
	Status        string `json:"status,omitempty"`
	PreviousStatus string `json:"previous_status,omitempty"`
	FromPhase     string `json:"from_phase,omitempty"`
	ToPhase       string `json:"to_phase,omitempty"`
	TransitionedAt string `json:"transitioned_at,omitempty"`
	ParkedAt      string `json:"parked_at,omitempty"`
	ParkedReason  string `json:"parked_reason,omitempty"`
	ResumedAt     string `json:"resumed_at,omitempty"`
	ArchivedAt    string `json:"archived_at,omitempty"`
	Archived      bool   `json:"archived,omitempty"`
	ArchivePath   string `json:"archive_path,omitempty"`
}

// AuditOutput represents audit log output.
type AuditOutput struct {
	SessionID      string       `json:"session_id"`
	Events         []AuditEvent `json:"events"`
	Total          int          `json:"total"`
	FiltersApplied AuditFilters `json:"filters_applied"`
}

// AuditEvent represents a single audit event.
type AuditEvent struct {
	Timestamp   string                 `json:"timestamp"`
	Event       string                 `json:"event"`
	From        string                 `json:"from,omitempty"`
	To          string                 `json:"to,omitempty"`
	FromPhase   string                 `json:"from_phase,omitempty"`
	ToPhase     string                 `json:"to_phase,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuditFilters shows which filters were applied.
type AuditFilters struct {
	Limit     int    `json:"limit"`
	EventType string `json:"event_type"`
	Since     string `json:"since"`
}

// Headers implements Tabular for AuditOutput.
func (a AuditOutput) Headers() []string {
	return []string{"TIMESTAMP", "EVENT", "TRANSITION", "DETAILS"}
}

// Rows implements Tabular for AuditOutput.
func (a AuditOutput) Rows() [][]string {
	rows := make([][]string, len(a.Events))
	for i, e := range a.Events {
		transition := ""
		if e.From != "" || e.To != "" {
			transition = fmt.Sprintf("%s -> %s", e.From, e.To)
		} else if e.FromPhase != "" || e.ToPhase != "" {
			transition = fmt.Sprintf("%s -> %s", e.FromPhase, e.ToPhase)
		}

		details := ""
		if e.Metadata != nil {
			parts := make([]string, 0, len(e.Metadata))
			for k, v := range e.Metadata {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
			details = strings.Join(parts, ", ")
		}

		rows[i] = []string{e.Timestamp, e.Event, transition, details}
	}
	return rows
}

// LockOutput represents lock operation result.
type LockOutput struct {
	SessionID  string `json:"session_id"`
	Locked     bool   `json:"locked"`
	Unlocked   bool   `json:"unlocked,omitempty"`
	LockPath   string `json:"lock_path,omitempty"`
	HolderPID  int    `json:"holder_pid,omitempty"`
	AcquiredAt string `json:"acquired_at,omitempty"`
	WasStale   bool   `json:"was_stale,omitempty"`
}

// MigrateOutput represents migration result.
type MigrateOutput struct {
	Migrated       []MigrationResult `json:"migrated"`
	Skipped        []SkipResult      `json:"skipped"`
	Failed         []FailResult      `json:"failed"`
	TotalMigrated  int               `json:"total_migrated"`
	TotalSkipped   int               `json:"total_skipped"`
	TotalFailed    int               `json:"total_failed"`
	DryRun         bool              `json:"dry_run"`
}

// MigrationResult describes a successful migration.
type MigrationResult struct {
	SessionID      string   `json:"session_id"`
	FromVersion    string   `json:"from_version"`
	ToVersion      string   `json:"to_version"`
	StatusDerived  string   `json:"status_derived,omitempty"`
	FieldsMigrated []string `json:"fields_migrated,omitempty"`
	BackupPath     string   `json:"backup_path,omitempty"`
}

// SkipResult describes a skipped migration.
type SkipResult struct {
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"`
}

// FailResult describes a failed migration.
type FailResult struct {
	SessionID string `json:"session_id"`
	Error     string `json:"error"`
}
