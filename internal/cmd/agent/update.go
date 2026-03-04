package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	agentpkg "github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
)

// agentUpdateOutput is the structured output for ari agent update.
type agentUpdateOutput struct {
	Entries []agentUpdateEntry `json:"entries"`
	Updated int                `json:"updated"`
	Skipped int                `json:"skipped"`
	Errors  int                `json:"errors"`
}

type agentUpdateEntry struct {
	Path     string `json:"path"`
	Status   string `json:"status"` // "updated", "skipped", "would_update", "error"
	Sections int    `json:"sections,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Text implements output.Textable.
func (u agentUpdateOutput) Text() string {
	var b strings.Builder

	for _, entry := range u.Entries {
		switch entry.Status {
		case "updated":
			b.WriteString(fmt.Sprintf("UPDATED  %s (%d sections regenerated)\n", entry.Path, entry.Sections))
		case "would_update":
			b.WriteString(fmt.Sprintf("WOULD UPDATE %s (%d sections would change)\n", entry.Path, entry.Sections))
		case "skipped":
			b.WriteString(fmt.Sprintf("SKIPPED  %s (%s)\n", entry.Path, entry.Reason))
		case "error":
			b.WriteString(fmt.Sprintf("ERROR    %s: %s\n", entry.Path, entry.Error))
		}
	}

	b.WriteString(fmt.Sprintf("\nSummary: %d updated, %d skipped, %d errors\n",
		u.Updated, u.Skipped, u.Errors))
	return b.String()
}

type updateOptions struct {
	rite   string
	all    bool
	dryRun bool
	paths  []string
}

func newUpdateCmd(ctx *cmdContext) *cobra.Command {
	var opts updateOptions

	cmd := &cobra.Command{
		Use:   "update [path...]",
		Short: "Update platform-owned sections in agent files",
		Long: `Regenerates platform-owned and derived sections while preserving author content.

This command:
- Reads existing agent files
- Looks up the archetype from the 'type' frontmatter field
- Regenerates platform-owned sections from templates
- Regenerates derived sections from frontmatter data
- Preserves author-owned sections exactly as-is
- Preserves unknown sections not in the archetype

Platform sections are defined in archetype templates and regenerated on update.
Author sections are never modified (or added with TODO markers if missing).
Derived sections are generated from frontmatter (tools, upstream/downstream).

Examples:
  ari agent update rites/ecosystem/agents/orchestrator.md
  ari agent update --rite ecosystem
  ari agent update --all
  ari agent update --dry-run --rite ecosystem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.paths = args
			return runUpdate(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.rite, "rite", "r", "", "Update all agents in a specific rite")
	cmd.Flags().BoolVar(&opts.all, "all", false, "Update all agents in all rites")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Show what would change without writing files")

	return cmd
}

func runUpdate(ctx *cmdContext, opts updateOptions) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	// Determine which agent files to update
	var agentPaths []string
	var err error

	switch {
	case opts.all:
		agentPaths, err = findAllAgentFiles(resolver)
	case opts.rite != "":
		agentPaths, err = findAgentFilesInRite(resolver, opts.rite)
	case len(opts.paths) > 0:
		agentPaths = opts.paths
	default:
		return errors.New(errors.CodeUsageError,
			"must specify agent paths, --rite, or --all")
	}

	if err != nil {
		printer.PrintError(err)
		return err
	}

	if len(agentPaths) == 0 {
		printer.PrintLine("No agent files found")
		return nil
	}

	// Process each agent file
	stats := &updateStats{}
	for _, path := range agentPaths {
		if err := updateAgentFile(resolver, path, opts.dryRun, stats); err != nil {
			relPath, _ := filepath.Rel(resolver.ProjectRoot(), path)
			stats.entries = append(stats.entries, agentUpdateEntry{
				Path:   relPath,
				Status: "error",
				Error:  err.Error(),
			})
			stats.errors++
		}
	}

	out := agentUpdateOutput{
		Entries: stats.entries,
		Updated: stats.updated,
		Skipped: stats.skipped,
		Errors:  stats.errors,
	}

	if err := printer.Print(out); err != nil {
		return err
	}

	if stats.errors > 0 {
		return errors.New(errors.CodeGeneralError, "some agent files failed to update")
	}

	return nil
}

type updateStats struct {
	updated int
	skipped int
	errors  int
	entries []agentUpdateEntry
}

// updateAgentFile updates a single agent file.
func updateAgentFile(resolver interface{ ProjectRoot() string }, path string, dryRun bool, stats *updateStats) error {
	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(errors.CodeFileNotFound, "failed to read agent file", err)
	}

	// Parse the agent file
	parsed, err := agentpkg.ParseAgentSections(content)
	if err != nil {
		return errors.Wrap(errors.CodeParseError, "failed to parse agent file", err)
	}

	// Get archetype from type field
	archetypeName := parsed.Frontmatter.Type
	if archetypeName == "" {
		relPath, _ := filepath.Rel(resolver.ProjectRoot(), path)
		stats.entries = append(stats.entries, agentUpdateEntry{Path: relPath, Status: "skipped", Reason: "no type field"})
		stats.skipped++
		return nil
	}

	// Map non-standard types to specialist
	if archetypeName != "orchestrator" && archetypeName != "specialist" && archetypeName != "reviewer" {
		archetypeName = "specialist"
	}

	archetype, err := agentpkg.GetArchetype(archetypeName)
	if err != nil {
		relPath, _ := filepath.Rel(resolver.ProjectRoot(), path)
		stats.entries = append(stats.entries, agentUpdateEntry{Path: relPath, Status: "skipped", Reason: fmt.Sprintf("unknown archetype %q", archetypeName)})
		stats.skipped++
		return nil
	}

	// Check if there are any platform or derived sections to update
	if !hasSectionsToUpdate(archetype) {
		relPath, _ := filepath.Rel(resolver.ProjectRoot(), path)
		stats.entries = append(stats.entries, agentUpdateEntry{Path: relPath, Status: "skipped", Reason: "no platform sections to update"})
		stats.skipped++
		return nil
	}

	// Regenerate platform sections
	updated, err := agentpkg.RegeneratePlatformSections(parsed, archetype)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to regenerate sections", err)
	}

	// Reassemble the file
	newContent := agentpkg.AssembleAgentFile(updated)

	// Count how many sections changed
	changedCount := countChangedSections(parsed, updated)

	if changedCount == 0 {
		relPath, _ := filepath.Rel(resolver.ProjectRoot(), path)
		stats.entries = append(stats.entries, agentUpdateEntry{Path: relPath, Status: "skipped", Reason: "no changes"})
		stats.skipped++
		return nil
	}

	relPath, _ := filepath.Rel(resolver.ProjectRoot(), path)

	if dryRun {
		stats.entries = append(stats.entries, agentUpdateEntry{Path: relPath, Status: "would_update", Sections: changedCount})
		stats.updated++
		return nil
	}

	// Write back
	if err := os.WriteFile(path, newContent, 0644); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to write agent file", err)
	}

	stats.entries = append(stats.entries, agentUpdateEntry{Path: relPath, Status: "updated", Sections: changedCount})
	stats.updated++
	return nil
}

// hasSectionsToUpdate returns true if the archetype has platform or derived sections.
func hasSectionsToUpdate(archetype *agentpkg.Archetype) bool {
	for _, section := range archetype.Sections {
		if section.Ownership == agentpkg.OwnerPlatform || section.Ownership == agentpkg.OwnerDerived {
			return true
		}
	}
	return false
}

// countChangedSections counts how many sections have different content.
func countChangedSections(original, updated *agentpkg.ParsedAgent) int {
	changed := 0

	// Build map of original section content by name
	originalContent := make(map[string]string)
	for _, section := range original.Sections {
		if section.Name != "" {
			originalContent[section.Name] = strings.TrimSpace(section.Content)
		}
	}

	// Compare updated sections
	for _, section := range updated.Sections {
		if section.Name == "" {
			continue
		}
		if original, found := originalContent[section.Name]; found {
			if original != strings.TrimSpace(section.Content) {
				changed++
			}
		} else {
			// New section
			changed++
		}
	}

	return changed
}

// findAllAgentFiles finds all agent files in the project.
func findAllAgentFiles(resolver interface{ ProjectRoot() string }) ([]string, error) {
	var paths []string

	// Find in rites/*/agents/
	ritesDir := filepath.Join(resolver.ProjectRoot(), "rites")
	entries, err := os.ReadDir(ritesDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to read rites directory", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentsDir := filepath.Join(ritesDir, entry.Name(), "agents")
		ritePaths, err := findAgentFilesInDir(agentsDir)
		if err != nil {
			continue // Skip rites without agents directory
		}
		paths = append(paths, ritePaths...)
	}

	// Find in agents/
	userAgentsDir := filepath.Join(resolver.ProjectRoot(), "agents")
	userPaths, err := findAgentFilesInDir(userAgentsDir)
	if err == nil {
		paths = append(paths, userPaths...)
	}

	return paths, nil
}

// findAgentFilesInRite finds all agent files in a specific rite.
func findAgentFilesInRite(resolver interface{ ProjectRoot() string }, riteName string) ([]string, error) {
	agentsDir := filepath.Join(resolver.ProjectRoot(), "rites", riteName, "agents")
	return findAgentFilesInDir(agentsDir)
}

// findAgentFilesInDir finds all .md files in a directory.
func findAgentFilesInDir(dir string) ([]string, error) {
	var paths []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound,
			fmt.Sprintf("failed to read directory: %s", dir), err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	return paths, nil
}
