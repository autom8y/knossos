// Package output provides format-aware output printing for Ariadne.
// It handles JSON, YAML, and text (table) output formats.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
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

// ValidateFormat checks if a format string is valid.
func ValidateFormat(s string) error {
	switch strings.ToLower(s) {
	case "text", "json", "yaml", "":
		return nil
	default:
		return fmt.Errorf("invalid output format %q (must be text, json, or yaml)", s)
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
func (p *Printer) Print(data any) error {
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
func (p *Printer) PrintSuccess(data any) error {
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

// PrintError outputs an error to stderr. The output error (if any) is
// silently discarded since callers are already in an error-handling path
// and cannot meaningfully react to a stderr write failure.
func (p *Printer) PrintError(err error) {
	if p.format == FormatJSON {
		// Check if error has JSON method
		if jsonErr, ok := err.(interface{ JSON() string }); ok {
			fmt.Fprintln(p.errOut, jsonErr.JSON())
			return
		}
		// Wrap in standard error format
		wrapper := map[string]any{
			"error": map[string]any{
				"code":    "GENERAL_ERROR",
				"message": err.Error(),
			},
		}
		enc := json.NewEncoder(p.errOut)
		enc.SetIndent("", "  ")
		_ = enc.Encode(wrapper)
		return
	}
	fmt.Fprintf(p.errOut, "Error: %s\n", err.Error())
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
func (p *Printer) VerboseLog(level, msg string, fields map[string]any) {
	if !p.verbose {
		return
	}
	entry := map[string]any{
		"level": level,
		"msg":   msg,
		"ts":    time.Now().UTC().Format(time.RFC3339),
	}
	maps.Copy(entry, fields)
	data, _ := json.Marshal(entry)
	fmt.Fprintln(p.errOut, string(data))
}

func (p *Printer) printJSON(data any) error {
	enc := json.NewEncoder(p.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (p *Printer) printYAML(data any) error {
	enc := yaml.NewEncoder(p.out)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(data)
}

func (p *Printer) printText(data any) error {
	// Handle Tabular interface for table output
	if t, ok := data.(Tabular); ok {
		return p.printTable(t)
	}

	// Handle Textable interface for custom text output
	if t, ok := data.(Textable); ok {
		fmt.Fprintln(p.out, t.Text())
		return nil
	}

	// Fallback: warn that a type is missing Text()/Tabular interface, then dump anyway
	fmt.Fprintf(p.errOut, "warning: output type %T has no Text() or Tabular() interface\n", data)
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
	defer func() { _ = w.Flush() }()

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

// Note: SessionListOutput intentionally does NOT implement Textable.
// The Tabular interface (Headers/Rows) is the correct rendering path.
// If no sessions exist, Rows() returns an empty table which the printer handles.

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
	GraduatedCount int      `json:"graduated_count,omitempty"`
	PromotedCount  int      `json:"promoted_count,omitempty"`
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
	Timestamp string         `json:"timestamp"`
	Event     string         `json:"event"`
	From      string         `json:"from,omitempty"`
	To        string         `json:"to,omitempty"`
	FromPhase string         `json:"from_phase,omitempty"`
	ToPhase   string         `json:"to_phase,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
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

// ClaimOutput represents session claim (CC binding) result.
type ClaimOutput struct {
	SessionID   string `json:"session_id"`
	CCSessionID string `json:"cc_session_id"`
	Status      string `json:"status"`
}

// Text implements Textable for ClaimOutput.
func (c ClaimOutput) Text() string {
	return fmt.Sprintf("Claimed session %s (status: %s)", c.SessionID, c.Status)
}

// --- Sync Output Structures ---

// SyncResultOutput represents the result of a sync operation.
type SyncResultOutput struct {
	Status string          `json:"status"`
	DryRun bool            `json:"dry_run,omitempty"`
	Rite   *SyncRiteResult `json:"rite,omitempty"`
	Org    *SyncOrgResult  `json:"org,omitempty"`
	User   *SyncUserResult `json:"user,omitempty"`
	Budget any             `json:"budget,omitempty"`
}

// SyncRiteResult represents rite scope sync result.
type SyncRiteResult struct {
	Status          string   `json:"status"`
	Error           string   `json:"error,omitempty"`
	RiteName        string   `json:"rite,omitempty"`
	Source          string   `json:"source,omitempty"`
	SourcePath      string   `json:"source_path,omitempty"`
	OrphansDetected []string `json:"orphans_detected,omitempty"`
	OrphanAction    string   `json:"orphan_action,omitempty"`
	LegacyBackup    string   `json:"legacy_backup,omitempty"`
	SoftMode        bool     `json:"soft_mode,omitempty"`
	DeferredStages  []string `json:"deferred_stages,omitempty"`
	ElCheapoMode    bool     `json:"el_cheapo_mode,omitempty"`
	RiteSwitched    bool     `json:"rite_switched,omitempty"`
	PreviousRite    string   `json:"previous_rite,omitempty"`
}

// SyncOrgResult represents org scope sync result.
type SyncOrgResult struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	OrgName string `json:"org,omitempty"`
	Source  string `json:"source,omitempty"`
	Agents  int    `json:"agents,omitempty"`
	Mena    int    `json:"mena,omitempty"`
}

// SyncUserResult represents user scope sync result.
type SyncUserResult struct {
	Status    string         `json:"status"`
	Totals    any            `json:"totals"`
	Resources map[string]any `json:"resources,omitempty"`
	Errors    any            `json:"errors,omitempty"`
}

// Text implements Textable for SyncResultOutput.
func (s SyncResultOutput) Text() string {
	var b strings.Builder
	if s.DryRun {
		b.WriteString("[DRY RUN] ")
	}
	b.WriteString(fmt.Sprintf("Sync: %s\n", s.Status))

	if s.Rite != nil {
		b.WriteString(fmt.Sprintf("  Rite: %s", s.Rite.Status))
		if s.Rite.RiteName != "" {
			b.WriteString(fmt.Sprintf(" (%s)", s.Rite.RiteName))
		}
		b.WriteString("\n")
		if s.Rite.Error != "" {
			if s.Rite.Status == "skipped" {
				b.WriteString(fmt.Sprintf("    Reason: %s\n", s.Rite.Error))
			} else {
				b.WriteString(fmt.Sprintf("  Error: %s\n", s.Rite.Error))
			}
		}
		if len(s.Rite.OrphansDetected) > 0 {
			if s.Rite.RiteSwitched && s.Rite.OrphanAction == "removed" {
				b.WriteString(fmt.Sprintf("  Agents: %d replaced (rite switch: %s -> %s)\n",
					len(s.Rite.OrphansDetected), s.Rite.PreviousRite, s.Rite.RiteName))
			} else {
				b.WriteString(fmt.Sprintf("  Orphans: %d detected (%s)\n", len(s.Rite.OrphansDetected), s.Rite.OrphanAction))
			}
		}
		if s.Rite.SoftMode {
			b.WriteString(fmt.Sprintf("  Soft mode: deferred %s\n", strings.Join(s.Rite.DeferredStages, ", ")))
		}
		if s.Rite.ElCheapoMode {
			b.WriteString("  El-cheapo mode: all agents using haiku\n")
		}
	}

	if s.Org != nil {
		b.WriteString(fmt.Sprintf("  Org: %s", s.Org.Status))
		if s.Org.OrgName != "" {
			b.WriteString(fmt.Sprintf(" (%s)", s.Org.OrgName))
		}
		if s.Org.Agents > 0 || s.Org.Mena > 0 {
			b.WriteString(fmt.Sprintf(" [agents:%d, mena:%d]", s.Org.Agents, s.Org.Mena))
		}
		b.WriteString("\n")
		if s.Org.Error != "" {
			if s.Org.Status == "skipped" {
				b.WriteString(fmt.Sprintf("    Reason: %s\n", s.Org.Error))
			} else {
				b.WriteString(fmt.Sprintf("  Error: %s\n", s.Org.Error))
			}
		}
	}

	if s.User != nil {
		b.WriteString(fmt.Sprintf("  User: %s\n", s.User.Status))
	}

	return b.String()
}

// TimelineEntryOutput represents a single timeline entry in JSON output.
type TimelineEntryOutput struct {
	Time     string `json:"time"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
}

// TimelineOutput represents the timeline for a session.
type TimelineOutput struct {
	SessionID string                `json:"session_id"`
	Entries   []TimelineEntryOutput `json:"entries"`
	Total     int                   `json:"total"`
	Filtered  int                   `json:"filtered"`
}

// Text implements Textable for TimelineOutput.
func (t TimelineOutput) Text() string {
	var b strings.Builder

	if t.SessionID != "" {
		b.WriteString(fmt.Sprintf("Timeline for %s:\n", t.SessionID))
	}

	if len(t.Entries) == 0 {
		b.WriteString("(no entries)\n")
		return b.String()
	}

	for _, e := range t.Entries {
		b.WriteString(fmt.Sprintf("- %s | %-8s | %s\n", e.Time, e.Category, e.Summary))
	}

	return b.String()
}

// LogOutput represents the result of an `ari session log` command.
type LogOutput struct {
	SessionID string `json:"session_id"`
	Type      string `json:"type"`
	Entry     string `json:"entry"`
}

// Text implements Textable for LogOutput.
func (l LogOutput) Text() string {
	return fmt.Sprintf("logged: %s\n", l.Entry)
}

// FieldOutput represents a single field get or set result.
type FieldOutput struct {
	Key           string `json:"key"`
	Value         string `json:"value"`
	PreviousValue string `json:"previous_value,omitempty"`
}

// Text implements Textable for FieldOutput.
func (f FieldOutput) Text() string {
	if f.PreviousValue != "" {
		return fmt.Sprintf("%s: %s (was: %s)\n", f.Key, f.Value, f.PreviousValue)
	}
	return fmt.Sprintf("%s: %s\n", f.Key, f.Value)
}

// FieldAllOutput represents a full snapshot of all session frontmatter fields.
type FieldAllOutput struct {
	SessionID     string `json:"session_id"`
	Status        string `json:"status"`
	Initiative    string `json:"initiative"`
	Complexity    string `json:"complexity"`
	CurrentPhase  string `json:"current_phase"`
	ActiveRite    string `json:"active_rite"`
	SchemaVersion string `json:"schema_version"`
	CreatedAt     string `json:"created_at"`
}

// Text implements Textable for FieldAllOutput.
func (f FieldAllOutput) Text() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("session_id: %s\n", f.SessionID))
	b.WriteString(fmt.Sprintf("status: %s\n", f.Status))
	b.WriteString(fmt.Sprintf("initiative: %s\n", f.Initiative))
	b.WriteString(fmt.Sprintf("complexity: %s\n", f.Complexity))
	b.WriteString(fmt.Sprintf("current_phase: %s\n", f.CurrentPhase))
	b.WriteString(fmt.Sprintf("active_rite: %s\n", f.ActiveRite))
	b.WriteString(fmt.Sprintf("schema_version: %s\n", f.SchemaVersion))
	b.WriteString(fmt.Sprintf("created_at: %s\n", f.CreatedAt))
	return b.String()
}

// SnapshotOutput is a wrapper around a rendered session context snapshot.
// The Text() method returns the pre-rendered markdown for text output.
// For JSON output, the command bypasses this struct and writes raw JSON directly.
type SnapshotOutput struct {
	// Markdown is the pre-rendered markdown for text output.
	Markdown string
}

// Text implements Textable for SnapshotOutput.
func (s SnapshotOutput) Text() string {
	return s.Markdown
}

// QueryStrand represents a child session strand in query output.
// Mirrors hook.StrandOutput but lives in the output package for coupling independence.
type QueryStrand struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	FrameRef  string `json:"frame_ref,omitempty"`
	LandedAt  string `json:"landed_at,omitempty"`
}

// QueryOutput represents the output of the `ari session query` command.
// The text format mirrors the hook context YAML frontmatter so agents receive
// the same field layout regardless of whether they read from hook injection or
// an on-demand pull.
type QueryOutput struct {
	SessionID     string       `json:"session_id,omitempty"`
	Status        string       `json:"status,omitempty"`
	Initiative    string       `json:"initiative,omitempty"`
	Complexity    string       `json:"complexity,omitempty"`
	ActiveRite    string       `json:"active_rite,omitempty"`
	ExecutionMode string       `json:"execution_mode,omitempty"`
	CurrentPhase  string       `json:"current_phase,omitempty"`
	FrayedFrom    string       `json:"frayed_from,omitempty"`
	FrameRef      string       `json:"frame_ref,omitempty"`
	ParkSource    string       `json:"park_source,omitempty"`
	ClaimedBy     string       `json:"claimed_by,omitempty"`
	Strands       []QueryStrand `json:"strands,omitempty"`
	HasSession    bool         `json:"has_session"`
}

// Text implements Textable for QueryOutput.
// Produces YAML frontmatter matching the hook context format so agents can
// parse query output the same way they parse hook-injected context.
func (q QueryOutput) Text() string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("# Session Context (ari session query)\n")

	if !q.HasSession {
		b.WriteString("has_session: false\n")
		b.WriteString("---\n")
		return b.String()
	}

	// Required fields
	b.WriteString(fmt.Sprintf("session_id: %s\n", q.SessionID))
	b.WriteString(fmt.Sprintf("status: %s\n", q.Status))
	b.WriteString(fmt.Sprintf("initiative: %q\n", q.Initiative))
	b.WriteString(fmt.Sprintf("active_rite: %s\n", q.ActiveRite))
	b.WriteString(fmt.Sprintf("execution_mode: %s\n", q.ExecutionMode))

	// Optional scalar fields (omitempty)
	if q.CurrentPhase != "" {
		b.WriteString(fmt.Sprintf("current_phase: %s\n", q.CurrentPhase))
	}
	if q.Complexity != "" {
		b.WriteString(fmt.Sprintf("complexity: %s\n", q.Complexity))
	}
	if q.FrayedFrom != "" {
		b.WriteString(fmt.Sprintf("frayed_from: %s\n", q.FrayedFrom))
	}
	if q.FrameRef != "" {
		b.WriteString(fmt.Sprintf("frame_ref: %s\n", q.FrameRef))
	}
	if q.ParkSource != "" {
		b.WriteString(fmt.Sprintf("park_source: %s\n", q.ParkSource))
	}
	if q.ClaimedBy != "" {
		b.WriteString(fmt.Sprintf("claimed_by: %s\n", q.ClaimedBy))
	}

	// Strands rendered as YAML list (omitempty)
	if len(q.Strands) > 0 {
		b.WriteString("strands:\n")
		for _, s := range q.Strands {
			b.WriteString(fmt.Sprintf("  - session_id: %s\n", s.SessionID))
			b.WriteString(fmt.Sprintf("    status: %s\n", s.Status))
			if s.FrameRef != "" {
				b.WriteString(fmt.Sprintf("    frame_ref: %s\n", s.FrameRef))
			}
			if s.LandedAt != "" {
				b.WriteString(fmt.Sprintf("    landed_at: %q\n", s.LandedAt))
			}
		}
	}

	b.WriteString("---\n")
	return b.String()
}
