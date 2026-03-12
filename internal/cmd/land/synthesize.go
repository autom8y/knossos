package land

import (
	"fmt"
	"github.com/autom8y/knossos/internal/cmd/common"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/session"
)

// validDomains lists the domains that Dionysus can synthesize.
var validDomains = []string{
	"initiative-history",
	"scar-tissue",
	"workflow-patterns",
	"all",
}

func newSynthesizeCmd(ctx *cmdContext) *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "synthesize",
		Short: "Enumerate session archives for Dionysus knowledge synthesis",
		Long: `Prepare infrastructure for cross-session knowledge synthesis via the Dionysus agent.

This command validates prerequisites, enumerates archived sessions in .sos/archive/,
and prints an inventory of available data. The inventory is used to construct a Task
tool invocation for the Dionysus agent, which performs the actual synthesis.

The Dionysus agent is invoked via the Task tool, not directly from the CLI.
Run this command first to confirm archives exist, then delegate to Dionysus:

  Task("dionysus", { domain: "scar-tissue", sessions: [...] })

Supported domains:
  initiative-history   Cross-session initiative outcomes and blockers
  scar-tissue          Recurring bugs, root causes, and defensive patterns
  workflow-patterns    Phase transitions, complexity calibration, rite usage
  all                  Synthesize all domains (default)

Examples:
  ari land synthesize
  ari land synthesize --domain=scar-tissue
  ari land synthesize -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSynthesize(ctx, domain)
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Knowledge domain to synthesize (initiative-history, scar-tissue, workflow-patterns, all)")

	return cmd
}

// sessionSummary is metadata extracted from a single archived session.
type sessionSummary struct {
	SessionID   string `json:"session_id"`
	Initiative  string `json:"initiative"`
	Complexity  string `json:"complexity"`
	ActiveRite  string `json:"active_rite"`
	CreatedAt   string `json:"created_at"`
	ArchivedAt  string `json:"archived_at,omitempty"`
	HasEvents   bool   `json:"has_events"`
	EventsBytes int64  `json:"events_bytes"`
	HasSails    bool   `json:"has_sails"`
}

// landFileSummary describes an existing .sos/land/{domain}.md file.
type landFileSummary struct {
	Domain              string `json:"domain"`
	Path                string `json:"path"`
	GeneratedAt         string `json:"generated_at,omitempty"`
	SessionsSynthesized int    `json:"sessions_synthesized,omitempty"`
}

// synthesizeOutput holds the full inventory for JSON output.
type synthesizeOutput struct {
	Status   string            `json:"status"`
	Domain   string            `json:"domain"`
	Sessions []sessionSummary  `json:"sessions"`
	LandDir  string            `json:"land_dir"`
	Existing []landFileSummary `json:"existing_land_files,omitempty"`
	Message  string            `json:"message"`
}

// Headers implements output.Tabular for synthesizeOutput.
func (o synthesizeOutput) Headers() []string {
	return []string{"SESSION ID", "INITIATIVE", "COMPLEXITY", "RITE", "ARCHIVED", "EVENTS"}
}

// Rows implements output.Tabular for synthesizeOutput.
func (o synthesizeOutput) Rows() [][]string {
	rows := make([][]string, len(o.Sessions))
	for i, s := range o.Sessions {
		archivedAt := s.ArchivedAt
		if archivedAt == "" {
			archivedAt = s.CreatedAt
		}
		// Trim to date portion for readability.
		if len(archivedAt) >= 10 {
			archivedAt = archivedAt[:10]
		}
		eventsLabel := "no"
		if s.HasEvents {
			eventsLabel = fmt.Sprintf("yes (%d B)", s.EventsBytes)
		}
		rows[i] = []string{
			s.SessionID,
			s.Initiative,
			s.Complexity,
			s.ActiveRite,
			archivedAt,
			eventsLabel,
		}
	}
	return rows
}

// Text implements output.Textable for synthesizeOutput.
// Provides a summary footer; the table is rendered via Tabular.
func (o synthesizeOutput) Text() string {
	return o.Message
}

func runSynthesize(ctx *cmdContext, domain string) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	// Validate domain flag.
	if domain != "" {
		if !isValidDomain(domain) {
			err := errors.New(errors.CodeUsageError,
				fmt.Sprintf("invalid domain %q: must be one of %s", domain, strings.Join(validDomains, ", ")))
			return common.PrintAndReturn(printer, err)
		}
	}
	effectiveDomain := domain
	if effectiveDomain == "" {
		effectiveDomain = "all"
	}

	// Ensure .sos/land/ exists.
	landDir := resolver.LandDir()
	if err := paths.EnsureDir(landDir); err != nil {
		return fmt.Errorf("cannot create land directory: %w", err)
	}

	// Enumerate sessions in .sos/archive/.
	archiveDir := resolver.ArchiveDir()
	archiveEntries, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			msg := fmt.Sprintf("No archive directory found at %s.\nRun `ari session wrap` to archive a completed session first.", archiveDir)
			format := output.ParseFormat(*ctx.Output)
			if format != output.FormatText {
				return printer.Print(synthesizeOutput{
					Status:  "no-archive",
					Domain:  effectiveDomain,
					LandDir: landDir,
					Message: msg,
				})
			}
			printer.PrintLine(msg)
			return nil
		}
		return fmt.Errorf("reading archive directory: %w", err)
	}

	var sessions []sessionSummary
	for _, entry := range archiveEntries {
		if !entry.IsDir() {
			continue
		}
		sessionDir := filepath.Join(archiveDir, entry.Name())
		summary, warn := loadSessionSummary(sessionDir)
		if warn != "" {
			fmt.Fprintf(os.Stderr, "warn: %s\n", warn)
			continue
		}
		sessions = append(sessions, summary)
	}

	if len(sessions) == 0 {
		msg := fmt.Sprintf("No archived sessions found in %s.\nRun `ari session wrap` to archive a completed session first.", archiveDir)
		format := output.ParseFormat(*ctx.Output)
		if format != output.FormatText {
			return printer.Print(synthesizeOutput{
				Status:  "empty",
				Domain:  effectiveDomain,
				LandDir: landDir,
				Message: msg,
			})
		}
		printer.PrintLine(msg)
		return nil
	}

	// Check for existing land files.
	existing := checkExistingLandFiles(landDir, effectiveDomain)

	// Build summary message for text mode.
	var msgBuilder strings.Builder
	fmt.Fprintf(&msgBuilder, "%d session(s) available for synthesis\n", len(sessions))
	fmt.Fprintf(&msgBuilder, "Domain:   %s\n", effectiveDomain)
	fmt.Fprintf(&msgBuilder, "Archive:  %s\n", archiveDir)
	fmt.Fprintf(&msgBuilder, "Land dir: %s\n", landDir)
	if len(existing) > 0 {
		msgBuilder.WriteString("\nExisting land files:\n")
		for _, lf := range existing {
			fmt.Fprintf(&msgBuilder, "  %s.md", lf.Domain)
			if lf.GeneratedAt != "" {
				fmt.Fprintf(&msgBuilder, " (generated: %s", lf.GeneratedAt)
				if lf.SessionsSynthesized > 0 {
					fmt.Fprintf(&msgBuilder, ", sessions: %d", lf.SessionsSynthesized)
				}
				msgBuilder.WriteString(")")
			}
			msgBuilder.WriteString("\n")
		}
	}
	msgBuilder.WriteString("\nTo synthesize, invoke Dionysus via Task tool with this inventory.")

	out := synthesizeOutput{
		Status:   "ready",
		Domain:   effectiveDomain,
		Sessions: sessions,
		LandDir:  landDir,
		Existing: existing,
		Message:  msgBuilder.String(),
	}

	// For text output, print the table then the summary footer.
	format := output.ParseFormat(*ctx.Output)
	if format == output.FormatText {
		if err := printer.Print(out); err != nil {
			return err
		}
		printer.PrintLine("")
		printer.PrintLine(out.Message)
		return nil
	}

	return printer.Print(out)
}

// loadSessionSummary parses the session directory and returns a summary.
// Returns a non-empty warning string (and zero summary) if the session should be skipped.
func loadSessionSummary(sessionDir string) (sessionSummary, string) {
	sessionID := filepath.Base(sessionDir)

	contextPath := filepath.Join(sessionDir, "SESSION_CONTEXT.md")
	ctx, err := session.LoadContext(contextPath)
	if err != nil {
		return sessionSummary{}, fmt.Sprintf("skipping %s: SESSION_CONTEXT.md missing or invalid: %v", sessionID, err)
	}

	summary := sessionSummary{
		SessionID:  ctx.SessionID,
		Initiative: ctx.Initiative,
		Complexity: ctx.Complexity,
		ActiveRite: ctx.ActiveRite,
	}
	if !ctx.CreatedAt.IsZero() {
		summary.CreatedAt = ctx.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if ctx.ArchivedAt != nil && !ctx.ArchivedAt.IsZero() {
		summary.ArchivedAt = ctx.ArchivedAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	// Check for events.jsonl.
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	if info, err := os.Stat(eventsPath); err == nil {
		summary.HasEvents = true
		summary.EventsBytes = info.Size()
	}

	// Check for WHITE_SAILS.yaml.
	sailsPath := filepath.Join(sessionDir, "WHITE_SAILS.yaml")
	if _, err := os.Stat(sailsPath); err == nil {
		summary.HasSails = true
	}

	return summary, ""
}

// checkExistingLandFiles returns summaries for existing .sos/land/{domain}.md files.
// When effectiveDomain is "all", checks all valid non-"all" domains.
func checkExistingLandFiles(landDir, effectiveDomain string) []landFileSummary {
	domainsToCheck := validDomains[:len(validDomains)-1] // exclude "all"
	if effectiveDomain != "all" {
		domainsToCheck = []string{effectiveDomain}
	}

	var existing []landFileSummary
	for _, d := range domainsToCheck {
		path := filepath.Join(landDir, d+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lf := landFileSummary{
			Domain: d,
			Path:   path,
		}
		// Parse simple frontmatter to extract generated_at and sessions_synthesized.
		meta := parseLandFileMeta(data)
		lf.GeneratedAt = meta.generatedAt
		lf.SessionsSynthesized = meta.sessionsSynthesized
		existing = append(existing, lf)
	}
	return existing
}

// landMeta holds the subset of land file frontmatter we care about.
type landMeta struct {
	generatedAt         string
	sessionsSynthesized int
}

// landFrontmatter is the YAML struct for parsing land file frontmatter.
type landFrontmatter struct {
	GeneratedAt         string `yaml:"generated_at"`
	SessionsSynthesized int    `yaml:"sessions_synthesized"`
}

// parseLandFileMeta extracts generated_at and sessions_synthesized from a land file.
// Returns zero values if frontmatter is absent or unparseable (graceful degradation).
func parseLandFileMeta(data []byte) landMeta {
	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return landMeta{}
	}
	endIdx := strings.Index(content[4:], "\n---")
	if endIdx == -1 {
		return landMeta{}
	}
	yamlContent := content[4 : endIdx+4]

	var fm landFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return landMeta{}
	}
	return landMeta{
		generatedAt:         fm.GeneratedAt,
		sessionsSynthesized: fm.SessionsSynthesized,
	}
}

// isValidDomain reports whether domain is in the validDomains list.
func isValidDomain(domain string) bool {
	return slices.Contains(validDomains, domain)
}
