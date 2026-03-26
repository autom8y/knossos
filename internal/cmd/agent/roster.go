// Package agent implements ari agent commands.
// roster.go implements the `ari agent roster` subcommand.
package agent

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	agentpkg "github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
	"gopkg.in/yaml.v3"
)

// rosterEntry represents a single agent in the roster output.
type rosterEntry struct {
	Name        string `json:"name"`
	Section     string `json:"section"` // "standing", "summoned", or "available"
	Description string `json:"description,omitempty"`
	SummonedAt  string `json:"summoned_at,omitempty"` // only for summoned agents
}

// rosterOutput is the structured output for ari agent roster.
type rosterOutput struct {
	Standing  []rosterEntry `json:"standing"`
	Summoned  []rosterEntry `json:"summoned"`
	Available []rosterEntry `json:"available"`
}

// Text implements output.Textable.
func (r rosterOutput) Text() string {
	var b strings.Builder

	b.WriteString("== Standing Agents ==\n")
	if len(r.Standing) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, e := range r.Standing {
			if e.Description != "" {
				fmt.Fprintf(&b, "  %-20s  %s\n", e.Name, e.Description)
			} else {
				fmt.Fprintf(&b, "  %s\n", e.Name)
			}
		}
	}

	b.WriteString("\n== Summoned (Active) ==\n")
	if len(r.Summoned) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, e := range r.Summoned {
			if e.SummonedAt != "" {
				fmt.Fprintf(&b, "  %-20s  summoned %s\n", e.Name, e.SummonedAt)
			} else {
				fmt.Fprintf(&b, "  %s\n", e.Name)
			}
		}
	}

	b.WriteString("\n== Available to Summon ==\n")
	if len(r.Available) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, e := range r.Available {
			if e.Description != "" {
				fmt.Fprintf(&b, "  %-20s  %s\n", e.Name, e.Description)
			} else {
				fmt.Fprintf(&b, "  %s\n", e.Name)
			}
		}
	}

	return b.String()
}

type rosterOptions struct{}

func newRosterCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roster",
		Short: "Show the agent roster: standing, summoned, and available agents",
		Long: `Displays three sections of the agent roster:

  Standing:  Core platform agents (pythia, moirai, metis) — always active.
  Summoned:  Agents you have summoned with 'ari agent summon'.
  Available: Agents available to summon (tier: summonable in source).

Examples:
  ari agent roster               # Show full roster
  ari agent roster -o json       # JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRoster(ctx, rosterOptions{})
		},
	}

	return cmd
}

func runRoster(ctx *cmdContext, opts rosterOptions) error {
	printer := ctx.GetPrinter(output.FormatText)

	// Resolve user channel dir (for summoned list + standing descriptions)
	userChannelDir, pathErr := paths.UserChannelDir("claude")
	if pathErr != nil {
		userChannelDir = "" // degrade gracefully
	}

	// Section 1: Standing agents
	standing := buildStandingSection(userChannelDir)

	// Section 2: Summoned (active) — read from USER_PROVENANCE_MANIFEST.yaml
	summoned := buildSummonedSection(userChannelDir)

	// Build a set of already-summoned names to exclude from available
	summonedNames := make(map[string]bool)
	for _, e := range summoned {
		summonedNames[e.Name] = true
	}

	// Section 3: Available to summon
	available := buildAvailableSection(summonedNames)

	result := rosterOutput{
		Standing:  standing,
		Summoned:  summoned,
		Available: available,
	}

	return printer.Print(result)
}

// buildStandingSection returns the hard-coded standing agents.
// Attempts to read description from the user channel dir's agents folder
// (where they may have been materialized by ari sync).
func buildStandingSection(userChannelDir string) []rosterEntry {
	names := []string{"pythia", "moirai", "metis"}
	entries := make([]rosterEntry, 0, len(names))

	for _, name := range names {
		entry := rosterEntry{
			Name:    name,
			Section: "standing",
		}

		// Try to read description from materialized agent file
		if userChannelDir != "" {
			agentPath := filepath.Join(userChannelDir, "agents", name+".md")
			if data, err := os.ReadFile(agentPath); err == nil {
				if fm, err := agentpkg.ParseAgentFrontmatter(data); err == nil && fm.Description != "" {
					entry.Description = truncateDescription(fm.Description)
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

// buildSummonedSection reads USER_PROVENANCE_MANIFEST.yaml and returns entries
// with source matching "summon:*".
func buildSummonedSection(userChannelDir string) []rosterEntry {
	if userChannelDir == "" {
		return nil
	}

	manifestPath := provenance.UserManifestPath(userChannelDir)
	manifest, err := provenance.LoadOrBootstrap(manifestPath)
	if err != nil {
		return nil
	}

	var entries []rosterEntry
	for key, entry := range manifest.Entries {
		if !strings.HasPrefix(entry.SourcePath, "summon:") {
			continue
		}
		// Key is "agents/{name}.md" — extract name
		if !strings.HasPrefix(key, "agents/") || !strings.HasSuffix(key, ".md") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(key, "agents/"), ".md")

		summonedAt := ""
		if !entry.LastSynced.IsZero() {
			summonedAt = entry.LastSynced.Format(time.RFC3339)
			// Annotate entries that have not been refreshed in over 24 hours.
			// A stale summoned agent may indicate a missed dismiss from a previous session.
			if time.Since(entry.LastSynced) > 24*time.Hour {
				summonedAt += " [stale?]"
			}
		}

		entries = append(entries, rosterEntry{
			Name:       name,
			Section:    "summoned",
			SummonedAt: summonedAt,
		})
	}

	// Sort for stable output
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

	return entries
}

// buildAvailableSection scans agent source files (KNOSSOS_HOME/agents/ + embedded)
// for files with tier: summonable in frontmatter. Excludes already-summoned and
// standing agents.
func buildAvailableSection(summonedNames map[string]bool) []rosterEntry {
	seen := make(map[string]bool)
	var entries []rosterEntry

	addEntry := func(name string, data []byte) {
		if standingAgents[name] || summonedNames[name] || seen[name] {
			return
		}
		seen[name] = true

		// Check for tier: summonable
		tier := extractTierFromBytes(data)
		if tier != "summonable" {
			return
		}

		entry := rosterEntry{
			Name:    name,
			Section: "available",
		}

		// Try to read description
		if fm, err := agentpkg.ParseAgentFrontmatter(data); err == nil && fm.Description != "" {
			entry.Description = truncateDescription(fm.Description)
		}

		entries = append(entries, entry)
	}

	// Scan KNOSSOS_HOME/agents/
	knossosHome := os.Getenv("KNOSSOS_HOME")
	if knossosHome != "" {
		agentsDir := filepath.Join(knossosHome, "agents")
		if dirEntries, err := os.ReadDir(agentsDir); err == nil {
			for _, e := range dirEntries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".md")
				data, err := os.ReadFile(filepath.Join(agentsDir, e.Name()))
				if err != nil {
					continue
				}
				addEntry(name, data)
			}
		}
	}

	// Scan embedded agents FS
	embeddedAgents := common.EmbeddedAgents()
	if embeddedAgents != nil {
		if dirEntries, err := fs.ReadDir(embeddedAgents, "agents"); err == nil {
			for _, e := range dirEntries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".md")
				data, err := fs.ReadFile(embeddedAgents, "agents/"+e.Name())
				if err != nil {
					continue
				}
				addEntry(name, data)
			}
		}
	}

	// Sort for stable output
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

	return entries
}

// extractTierFromBytes extracts the tier field from agent frontmatter bytes.
// Returns empty string if frontmatter is invalid or tier is absent.
func extractTierFromBytes(content []byte) string {
	yamlBytes, _, err := frontmatter.Parse(content)
	if err != nil {
		return ""
	}
	var fmMap map[string]any
	if err := yaml.Unmarshal(yamlBytes, &fmMap); err != nil {
		return ""
	}
	if v, ok := fmMap["tier"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// truncateDescription truncates a description to a reasonable display length.
// Trims to first line if multi-line, then truncates to 60 chars.
func truncateDescription(desc string) string {
	// Take first line only for multi-line descriptions
	if idx := strings.Index(desc, "\n"); idx >= 0 {
		desc = desc[:idx]
	}
	desc = strings.TrimSpace(desc)
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}
	return desc
}
