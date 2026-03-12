// Package status implements the ari status command for unified project health dashboard.
package status

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/know"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// HealthDashboard is the unified output for ari status.
type HealthDashboard struct {
	Channel ChannelHealth  `json:"channel"`
	Knossos KnossosHealth `json:"knossos"`
	Know    KnowHealth    `json:"know"`
	Ledge   LedgeHealth   `json:"ledge"`
	SOS     SOSHealth     `json:"sos"`
	Healthy bool          `json:"healthy"`
	Errors  []string      `json:"errors,omitempty"`
}

// ChannelHealth reports channel directory state.
type ChannelHealth struct {
	Exists      bool   `json:"exists"`
	ActiveRite  string `json:"active_rite,omitempty"`
	AgentCount  int    `json:"agent_count"`
	LastSync    string `json:"last_sync,omitempty"`
	LastSyncAge string `json:"last_sync_age,omitempty"`
}

// KnossosHealth reports .knossos/ directory state.
type KnossosHealth struct {
	Exists             bool     `json:"exists"`
	SatelliteRiteCount int      `json:"satellite_rite_count"`
	SatelliteRites     []string `json:"satellite_rites,omitempty"`
}

// KnowHealth reports .know/ directory state.
type KnowHealth struct {
	Exists      bool                `json:"exists"`
	DomainCount int                 `json:"domain_count"`
	FreshCount  int                 `json:"fresh_count"`
	StaleCount  int                 `json:"stale_count"`
	Domains     []know.DomainStatus `json:"domains,omitempty"`
}

// LedgeHealth reports .ledge/ directory state.
type LedgeHealth struct {
	Exists        bool `json:"exists"`
	DecisionCount int  `json:"decision_count"`
	SpecCount     int  `json:"spec_count"`
	ReviewCount   int  `json:"review_count"`
	SpikeCount    int  `json:"spike_count"`
	TotalCount    int  `json:"total_count"`
}

// SOSHealth reports .sos/ directory state.
type SOSHealth struct {
	Exists         bool   `json:"exists"`
	ActiveCount    int    `json:"active_count"`
	ParkedCount    int    `json:"parked_count"`
	ArchivedCount  int    `json:"archived_count"`
	TotalCount     int    `json:"total_count"`
	CurrentSession string `json:"current_session,omitempty"`
}

// Text implements output.Textable for human-readable dashboard output.
func (h HealthDashboard) Text() string {
	var b strings.Builder

	b.WriteString("=== Project Health Dashboard ===\n")

	// Channel directory
	b.WriteString("\nchannel/\n")
	if !h.Channel.Exists {
		b.WriteString("  (not found)\n")
	} else {
		if h.Channel.ActiveRite != "" {
			fmt.Fprintf(&b, "  Active Rite:  %s\n", h.Channel.ActiveRite)
		} else {
			b.WriteString("  Active Rite:  (none)\n")
		}
		fmt.Fprintf(&b, "  Agents:       %d\n", h.Channel.AgentCount)
		if h.Channel.LastSync != "" {
			sync := h.Channel.LastSync
			if h.Channel.LastSyncAge != "" {
				sync += " (" + h.Channel.LastSyncAge + ")"
			}
			fmt.Fprintf(&b, "  Last Sync:    %s\n", sync)
		} else {
			b.WriteString("  Last Sync:    (no provenance manifest)\n")
		}
	}

	// .knossos/
	b.WriteString("\n.knossos/\n")
	if !h.Knossos.Exists {
		b.WriteString("  (not found)\n")
	} else {
		if h.Knossos.SatelliteRiteCount == 0 {
			b.WriteString("  Satellite Rites: 0\n")
		} else {
			fmt.Fprintf(&b, "  Satellite Rites: %d (%s)\n",
				h.Knossos.SatelliteRiteCount,
				strings.Join(h.Knossos.SatelliteRites, ", "))
		}
	}

	// .know/
	b.WriteString("\n.know/\n")
	switch {
	case !h.Know.Exists:
		b.WriteString("  (not found)\n")
	case h.Know.DomainCount == 0:
		b.WriteString("  Domains: 0\n")
	default:
		fmt.Fprintf(&b, "  Domains: %d (%d fresh, %d stale)\n",
			h.Know.DomainCount, h.Know.FreshCount, h.Know.StaleCount)
		for _, d := range h.Know.Domains {
			freshLabel := "fresh"
			expiresLabel := "expires"
			if !d.Fresh {
				freshLabel = "STALE"
				expiresLabel = "expired"
			}
			fmt.Fprintf(&b, "    %-20s %-8s %s %s\n",
				d.Domain, freshLabel, expiresLabel, d.Expires)
		}
	}

	// .ledge/
	b.WriteString("\n.ledge/\n")
	if !h.Ledge.Exists {
		b.WriteString("  (not found)\n")
	} else {
		fmt.Fprintf(&b, "  Decisions: %d  Specs: %d  Reviews: %d  Spikes: %d\n",
			h.Ledge.DecisionCount, h.Ledge.SpecCount,
			h.Ledge.ReviewCount, h.Ledge.SpikeCount)
	}

	// .sos/
	b.WriteString("\n.sos/\n")
	if !h.SOS.Exists {
		b.WriteString("  (not found)\n")
	} else {
		fmt.Fprintf(&b, "  Sessions: %d (%d active, %d parked, %d archived)\n",
			h.SOS.TotalCount, h.SOS.ActiveCount,
			h.SOS.ParkedCount, h.SOS.ArchivedCount)
		if h.SOS.CurrentSession != "" {
			fmt.Fprintf(&b, "  Current: %s\n", h.SOS.CurrentSession)
		}
	}

	// Errors
	if len(h.Errors) > 0 {
		b.WriteString("\nErrors:\n")
		for _, e := range h.Errors {
			fmt.Fprintf(&b, "  - %s\n", e)
		}
	}

	return b.String()
}

type cmdContext struct {
	common.BaseContext
}

// NewStatusCmd creates the ari status command.
func NewStatusCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show unified project health dashboard",
		Long: `Display a unified health overview of all Knossos directory trees:
channel directory, .knossos/, .know/, .ledge/, and .sos/.

Reports active rite, agent count, sync recency, knowledge freshness,
artifact counts, and session state in a single view.

This is a read-only command — it does not modify any state.

Examples:
  ari status              # Human-readable dashboard
  ari status -o json      # Machine-readable JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := ctx.GetPrinter(output.FormatText)
			resolver := ctx.GetResolver()
			projectDir := resolver.ProjectRoot()

			dashboard := collect(resolver, projectDir)
			if err := printer.Print(dashboard); err != nil {
				return err
			}

			if !dashboard.Healthy {
				// Return a non-nil error to trigger exit code 1.
				// The dashboard has already been printed, so use SilenceErrors.
				return errors.New(errors.CodeValidationFailed, fmt.Sprintf("project unhealthy: %s", strings.Join(dashboard.Errors, "; ")))
			}
			return nil
		},
	}

	common.SetNeedsProject(cmd, true, true)
	return cmd
}

// collect gathers health data from all directory trees.
func collect(resolver *paths.Resolver, projectDir string) HealthDashboard {
	var errs []string

	channel := collectChannel(resolver)
	if !channel.Exists {
		errs = append(errs, "channel directory not found")
	}

	knossos := collectKnossos(resolver)
	knowHealth := collectKnow(projectDir)
	ledge := collectLedge(resolver)
	sos := collectSOS(resolver)

	return HealthDashboard{
		Channel: channel,
		Knossos: knossos,
		Know:    knowHealth,
		Ledge:   ledge,
		SOS:     sos,
		Healthy: len(errs) == 0,
		Errors:  errs,
	}
}

// collectChannel gathers channel directory health data.
func collectChannel(resolver *paths.Resolver) ChannelHealth {
	channelDir := resolver.ClaudeDir()
	if _, err := os.Stat(channelDir); os.IsNotExist(err) {
		return ChannelHealth{Exists: false}
	}

	health := ChannelHealth{Exists: true}

	// Active rite
	health.ActiveRite = resolver.ReadActiveRite()

	// Agent count
	agentsDir := resolver.AgentsDir()
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				health.AgentCount++
			}
		}
	}

	// Last sync from provenance manifest (now in .knossos/)
	manifestPath := provenance.ManifestPath(resolver.KnossosDir())
	if manifest, err := provenance.Load(manifestPath); err == nil {
		if !manifest.LastSync.IsZero() {
			health.LastSync = manifest.LastSync.UTC().Format(time.RFC3339)
			health.LastSyncAge = formatAge(manifest.LastSync)
		}
	}

	return health
}

// collectKnossos gathers .knossos/ health data.
func collectKnossos(resolver *paths.Resolver) KnossosHealth {
	knossosDir := resolver.KnossosDir()
	if _, err := os.Stat(knossosDir); os.IsNotExist(err) {
		return KnossosHealth{Exists: false}
	}

	health := KnossosHealth{Exists: true}

	ritesDir := resolver.RitesDir()
	entries, err := os.ReadDir(ritesDir)
	if err != nil {
		return health
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Check for manifest.yaml to confirm it's a valid rite
		manifestPath := filepath.Join(ritesDir, entry.Name(), "manifest.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			health.SatelliteRites = append(health.SatelliteRites, entry.Name())
			health.SatelliteRiteCount++
		}
	}

	return health
}

// collectKnow gathers .know/ health data.
func collectKnow(projectDir string) KnowHealth {
	knowDir := filepath.Join(projectDir, ".know")
	if _, err := os.Stat(knowDir); os.IsNotExist(err) {
		return KnowHealth{Exists: false}
	}

	health := KnowHealth{Exists: true}

	domains, err := know.ReadMeta(projectDir, projectDir)
	if err != nil {
		return health
	}

	health.Domains = domains
	health.DomainCount = len(domains)
	for _, d := range domains {
		if d.Fresh {
			health.FreshCount++
		} else {
			health.StaleCount++
		}
	}

	return health
}

// collectLedge gathers .ledge/ health data.
func collectLedge(resolver *paths.Resolver) LedgeHealth {
	ledgeDir := resolver.LedgeDir()
	if _, err := os.Stat(ledgeDir); os.IsNotExist(err) {
		return LedgeHealth{Exists: false}
	}

	health := LedgeHealth{Exists: true}

	health.DecisionCount = countMDFiles(resolver.LedgeDecisionsDir())
	health.SpecCount = countMDFiles(resolver.LedgeSpecsDir())
	health.ReviewCount = countMDFiles(resolver.LedgeReviewsDir())
	health.SpikeCount = countMDFiles(resolver.LedgeSpikesDir())
	health.TotalCount = health.DecisionCount + health.SpecCount + health.ReviewCount + health.SpikeCount

	return health
}

// collectSOS gathers .sos/ health data.
func collectSOS(resolver *paths.Resolver) SOSHealth {
	sosDir := resolver.SOSDir()
	if _, err := os.Stat(sosDir); os.IsNotExist(err) {
		return SOSHealth{Exists: false}
	}

	health := SOSHealth{Exists: true}

	// Read current session
	currentFile := resolver.CurrentSessionFile()
	if data, err := os.ReadFile(currentFile); err == nil {
		health.CurrentSession = strings.TrimSpace(string(data))
	}

	// Scan sessions directory
	sessionsDir := resolver.SessionsDir()
	if entries, err := os.ReadDir(sessionsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if len(name) < 32 || name[:8] != "session-" {
				continue
			}
			contextPath := filepath.Join(sessionsDir, name, "SESSION_CONTEXT.md")
			status := readSessionStatus(contextPath)
			switch status {
			case "ACTIVE":
				health.ActiveCount++
			case "PARKED":
				health.ParkedCount++
			}
		}
	}

	// Scan archive directory
	archiveDir := resolver.ArchiveDir()
	if entries, err := os.ReadDir(archiveDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && len(entry.Name()) >= 32 && entry.Name()[:8] == "session-" {
				health.ArchivedCount++
			}
		}
	}

	health.TotalCount = health.ActiveCount + health.ParkedCount + health.ArchivedCount
	return health
}

// readSessionStatus reads the status field from SESSION_CONTEXT.md frontmatter.
// Returns empty string on any error.
func readSessionStatus(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	// First line must be "---"
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return ""
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			break
		}
		if value, ok := strings.CutPrefix(line, "status:"); ok {
			return strings.Trim(strings.TrimSpace(value), "\"'")
		}
	}

	return ""
}

// countMDFiles counts .md files in a directory.
func countMDFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			count++
		}
	}
	return count
}

// formatAge returns a human-readable age string like "2h ago" or "3d ago".
func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
