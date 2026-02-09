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
	ActiveRite    string `json:"active_rite,omitempty"`
	ExecutionMode string `json:"execution_mode,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	SchemaVersion string `json:"schema_version,omitempty"`
	GitBranch     string `json:"git_branch,omitempty"`
	GitChanges    int    `json:"git_changes,omitempty"`
	SailsColor    string `json:"sails_color,omitempty"`
	SailsBase     string `json:"sails_base,omitempty"`
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
	b.WriteString(fmt.Sprintf("Rite: %s\n", s.ActiveRite))
	b.WriteString(fmt.Sprintf("Mode: %s\n", s.ExecutionMode))
	if s.GitBranch != "" {
		b.WriteString(fmt.Sprintf("Branch: %s (%d changes)\n", s.GitBranch, s.GitChanges))
	}
	// Display sails info
	if s.SailsColor != "" {
		sailsInfo := fmt.Sprintf("Sails: %s", s.SailsColor)
		if s.SailsBase != "" && s.SailsBase != s.SailsColor {
			sailsInfo += fmt.Sprintf(" (base: %s)", s.SailsBase)
		}
		b.WriteString(sailsInfo + "\n")
	} else {
		b.WriteString("Sails: not generated\n")
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
	Rite          string `json:"rite"`
	CreatedAt     string `json:"created_at"`
	SchemaVersion string `json:"schema_version"`
}

// Text implements Textable for CreateOutput.
func (c CreateOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Created session: %s\n", c.SessionID))
	b.WriteString(fmt.Sprintf("Initiative: %s\n", c.Initiative))
	b.WriteString(fmt.Sprintf("Complexity: %s\n", c.Complexity))
	b.WriteString(fmt.Sprintf("Rite: %s\n", c.Rite))
	return b.String()
}

// SeedCreateOutput represents session creation result when using --seed mode.
// The session is created in an ephemeral worktree and seeded (copied) to the main repo.
type SeedCreateOutput struct {
	SessionID   string `json:"session_id"`
	Status      string `json:"status"`
	Seeded      bool   `json:"seeded"`
	SeededTo    string `json:"seeded_to"`
	ParkReason  string `json:"park_reason"`
	Initiative  string `json:"initiative"`
	Complexity  string `json:"complexity"`
	Rite        string `json:"rite"`
	CreatedAt   string `json:"created_at"`
	ParkedAt    string `json:"parked_at"`
	ProjectRoot string `json:"project_root,omitempty"`
}

// Text implements Textable for SeedCreateOutput.
func (s SeedCreateOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Seeded session: %s\n", s.SessionID))
	b.WriteString(fmt.Sprintf("Status: %s\n", s.Status))
	b.WriteString(fmt.Sprintf("Initiative: %s\n", s.Initiative))
	b.WriteString(fmt.Sprintf("Complexity: %s\n", s.Complexity))
	b.WriteString(fmt.Sprintf("Rite: %s\n", s.Rite))
	b.WriteString(fmt.Sprintf("Seeded to: %s\n", s.SeededTo))
	b.WriteString(fmt.Sprintf("Park reason: %s\n", s.ParkReason))
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
	Stale      bool   `json:"stale,omitempty"`
	StaleHint  string `json:"stale_hint,omitempty"`
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
		status := s.Status
		if s.Stale && s.StaleHint != "" {
			status = s.Status + "  [STALE - " + s.StaleHint + "]"
		}
		rows[i] = []string{
			prefix + s.SessionID,
			status,
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
	SessionID      string   `json:"session_id"`
	Status         string   `json:"status,omitempty"`
	PreviousStatus string   `json:"previous_status,omitempty"`
	FromPhase      string   `json:"from_phase,omitempty"`
	ToPhase        string   `json:"to_phase,omitempty"`
	TransitionedAt string   `json:"transitioned_at,omitempty"`
	ParkedAt       string   `json:"parked_at,omitempty"`
	ParkedReason   string   `json:"parked_reason,omitempty"`
	ResumedAt      string   `json:"resumed_at,omitempty"`
	ArchivedAt     string   `json:"archived_at,omitempty"`
	Archived       bool     `json:"archived,omitempty"`
	ArchivePath    string   `json:"archive_path,omitempty"`
	SailsColor     string   `json:"sails_color,omitempty"`
	SailsBase      string   `json:"sails_base,omitempty"`
	SailsReasons   []string `json:"sails_reasons,omitempty"`
	SailsPath      string   `json:"sails_path,omitempty"`
}

// Text implements Textable for TransitionOutput.
// Displays wrap summary with sails color and appropriate warnings.
func (t TransitionOutput) Text() string {
	if t.Status != "ARCHIVED" {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Session %s archived\n", t.SessionID))

	// Display sails color with appropriate formatting
	if t.SailsColor != "" {
		b.WriteString(fmt.Sprintf("Sails: %s", t.SailsColor))
		if t.SailsBase != "" && t.SailsBase != t.SailsColor {
			b.WriteString(fmt.Sprintf(" (base: %s)", t.SailsBase))
		}
		b.WriteString("\n")

		// Display warnings based on color
		switch t.SailsColor {
		case "BLACK":
			b.WriteString("\nWARNING: BLACK sails - known failures detected.\n")
			b.WriteString("Do NOT ship this session to production.\n")
			if len(t.SailsReasons) > 0 {
				b.WriteString("Reasons:\n")
				for _, reason := range t.SailsReasons {
					b.WriteString(fmt.Sprintf("  - %s\n", reason))
				}
			}
		case "GRAY":
			b.WriteString("\nINFO: GRAY sails - confidence unknown.\n")
			b.WriteString("Consider QA review before shipping.\n")
			b.WriteString("Use /qa to run adversarial testing and upgrade to WHITE.\n")
		case "WHITE":
			b.WriteString("\nShip with confidence.\n")
		}
	}

	if t.Archived && t.ArchivePath != "" {
		b.WriteString(fmt.Sprintf("Archived to: %s\n", t.ArchivePath))
	}

	return b.String()
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
	Timestamp string                 `json:"timestamp"`
	Event     string                 `json:"event"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	FromPhase string                 `json:"from_phase,omitempty"`
	ToPhase   string                 `json:"to_phase,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
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

// MigrateOutput represents migration result.
type MigrateOutput struct {
	Migrated      []MigrationResult `json:"migrated"`
	Skipped       []SkipResult      `json:"skipped"`
	Failed        []FailResult      `json:"failed"`
	TotalMigrated int               `json:"total_migrated"`
	TotalSkipped  int               `json:"total_skipped"`
	TotalFailed   int               `json:"total_failed"`
	DryRun        bool              `json:"dry_run"`
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

// RecoverOutput represents session recovery result.
type RecoverOutput struct {
	StaleLocks            []string `json:"stale_locks,omitempty"`
	RemovedLocks          []string `json:"removed_locks,omitempty"`
	ActiveSession         string   `json:"active_session,omitempty"`
	CCMapOrphans          []string `json:"cc_map_orphans,omitempty"`
	RemovedCCMapOrphans   []string `json:"removed_cc_map_orphans,omitempty"`
	CurrentSessionCleaned bool     `json:"current_session_cleaned,omitempty"`
	DryRun                bool     `json:"dry_run"`
	Summary               string   `json:"summary"`
}

// FrayOutput represents session fray (fork) result.
type FrayOutput struct {
	ParentID     string `json:"parent_id"`
	ChildID      string `json:"child_id"`
	ChildDir     string `json:"child_dir"`
	FrayPoint    string `json:"fray_point"`
	WorktreePath string `json:"worktree_path,omitempty"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// Text implements Textable for FrayOutput.
func (f FrayOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Frayed session: %s -> %s\n", f.ParentID, f.ChildID))
	b.WriteString(fmt.Sprintf("Fray point: %s\n", f.FrayPoint))
	b.WriteString(fmt.Sprintf("Status: %s\n", f.Status))
	if f.WorktreePath != "" {
		b.WriteString(fmt.Sprintf("Worktree: %s\n", f.WorktreePath))
	}
	return b.String()
}

// RosterMigrateOutput represents the result of roster-to-knossos migration.
type RosterMigrateOutput struct {
	DryRun           bool               `json:"dry_run"`
	ManifestsFound   int                `json:"manifests_found"`
	ManifestsChanged int                `json:"manifests_changed"`
	ManifestsSkipped int                `json:"manifests_skipped"`
	EntriesRewritten int                `json:"entries_rewritten"`
	UserManifests    []ManifestMigResult `json:"user_manifests,omitempty"`
	CEMManifest      *ManifestMigResult `json:"cem_manifest,omitempty"`
	EnvVarsDetected  []EnvVarDetected   `json:"env_vars_detected,omitempty"`
	BackupsCreated   []string           `json:"backups_created,omitempty"`
	ScriptGenerated  bool               `json:"script_generated,omitempty"`
	ScriptPath       string             `json:"script_path,omitempty"`
	Errors           []string           `json:"errors,omitempty"`
}

// ManifestMigResult records migration outcome for a single manifest.
type ManifestMigResult struct {
	Path             string `json:"path"`
	EntriesRewritten int    `json:"entries_rewritten"`
	Skipped          bool   `json:"skipped"`
	SkipReason       string `json:"skip_reason,omitempty"`
	BackupPath       string `json:"backup_path,omitempty"`
}

// EnvVarDetected records a detected ROSTER_* environment variable.
type EnvVarDetected struct {
	Current string `json:"current"`
	Replace string `json:"replace"`
	Value   string `json:"value"`
}

// Text implements Textable for RosterMigrateOutput.
func (r RosterMigrateOutput) Text() string {
	var b strings.Builder

	if r.DryRun {
		b.WriteString("Roster-to-Knossos Migration (dry-run)\n\n")
	} else {
		b.WriteString("Roster-to-Knossos Migration\n\n")
	}

	// User manifests section
	if len(r.UserManifests) > 0 {
		b.WriteString("User Manifests:\n")
		for _, m := range r.UserManifests {
			if m.Skipped {
				b.WriteString(fmt.Sprintf("  %s  skipped (%s)\n", m.Path, m.SkipReason))
			} else {
				backupInfo := ""
				if m.BackupPath != "" && !r.DryRun {
					backupInfo = fmt.Sprintf(" (backup: %s)", m.BackupPath)
				}
				b.WriteString(fmt.Sprintf("  %s  %d entries rewritten%s\n", m.Path, m.EntriesRewritten, backupInfo))
			}
		}
		b.WriteString("\n")
	}

	// CEM manifest section
	if r.CEMManifest != nil {
		b.WriteString("CEM Manifest:\n")
		if r.CEMManifest.Skipped {
			b.WriteString(fmt.Sprintf("  %s  skipped (%s)\n", r.CEMManifest.Path, r.CEMManifest.SkipReason))
		} else {
			backupInfo := ""
			if r.CEMManifest.BackupPath != "" && !r.DryRun {
				backupInfo = fmt.Sprintf(" (backup: %s)", r.CEMManifest.BackupPath)
			}
			b.WriteString(fmt.Sprintf("  %s  %d fields rewritten%s\n", r.CEMManifest.Path, r.CEMManifest.EntriesRewritten, backupInfo))
		}
		b.WriteString("\n")
	}

	// Summary
	b.WriteString(fmt.Sprintf("Summary: %d manifests changed, %d skipped, %d entries rewritten\n",
		r.ManifestsChanged, r.ManifestsSkipped, r.EntriesRewritten))

	// Environment variables
	if len(r.EnvVarsDetected) > 0 {
		b.WriteString("\nEnvironment variables to update:\n")
		for _, ev := range r.EnvVarsDetected {
			b.WriteString(fmt.Sprintf("  %s -> %s (current: %s)\n", ev.Current, ev.Replace, ev.Value))
		}
	}

	// Errors
	if len(r.Errors) > 0 {
		b.WriteString("\nErrors:\n")
		for _, e := range r.Errors {
			b.WriteString(fmt.Sprintf("  - %s\n", e))
		}
	}

	// Dry-run prompt
	if r.DryRun {
		b.WriteString("\nUse --apply to execute this migration.\n")
	} else {
		b.WriteString("\nMigration complete.\n")
	}

	return b.String()
}
