package search

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/cmd/explain"
	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/rite"
)

// CollectCommands traverses the Cobra command tree recursively and returns
// a SearchEntry for each non-hidden, non-root command.
func CollectCommands(root *cobra.Command) []SearchEntry {
	var entries []SearchEntry
	collectCommandsRecursive(root, root, &entries)
	return entries
}

// collectCommandsRecursive recurses through the command tree.
func collectCommandsRecursive(root, cmd *cobra.Command, entries *[]SearchEntry) {
	for _, sub := range cmd.Commands() {
		if sub.Hidden {
			continue
		}
		// Skip the generated help command.
		if sub.Name() == "help" {
			continue
		}

		// Use CommandPath() which gives e.g. "ari session create".
		// Strip the root name prefix to get just "session create".
		fullPath := sub.CommandPath()
		// Remove leading "ari " if present (root is "ari").
		name := strings.TrimPrefix(fullPath, root.Name()+" ")
		if name == fullPath {
			// Root command itself — skip.
			name = strings.TrimPrefix(fullPath, root.Name())
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		*entries = append(*entries, SearchEntry{
			Name:        name,
			Domain:      DomainCommand,
			Summary:     sub.Short,
			Description: sub.Long,
			Action:      "ari " + name + " --help",
		})

		// Recurse into subcommands.
		collectCommandsRecursive(root, sub, entries)
	}
}

// CollectConcepts returns entries from the explain concept registry.
func CollectConcepts() []SearchEntry {
	concepts := explain.AllConcepts()
	entries := make([]SearchEntry, 0, len(concepts))
	for _, c := range concepts {
		entries = append(entries, SearchEntry{
			Name:        c.Name,
			Domain:      DomainConcept,
			Summary:     c.Summary,
			Description: c.Description,
			Aliases:     c.Aliases,
			Action:      "ari explain " + c.Name,
		})
	}
	return entries
}

// CollectRites returns entries from the rite discovery system.
// Only project-scoped rites are collected; user/org rites are excluded so
// that results reflect the active project context rather than the host system.
// Returns an empty slice if resolver is nil or has no project root.
func CollectRites(resolver *paths.Resolver) []SearchEntry {
	if resolver == nil || resolver.ProjectRoot() == "" {
		return nil
	}

	// Use project-only discovery to avoid picking up unrelated user rites from
	// the developer's home directory.
	activeRite := resolver.ReadActiveRite()
	discovery := rite.NewDiscoveryWithPaths(resolver.RitesDir(), "", activeRite)
	rites, err := discovery.List()
	if err != nil {
		return nil
	}

	entries := make([]SearchEntry, 0, len(rites))
	for _, r := range rites {
		e := SearchEntry{
			Name:        r.Name,
			Domain:      DomainRite,
			Summary:     r.Description,
			Description: r.Description,
			Action:      "/" + r.Name,
			Boosted:     r.Active,
		}
		entries = append(entries, e)
	}
	return entries
}

// CollectAgents returns entries from .claude/agents/ directory.
// Returns an empty slice if resolver is nil, has no project root, or the
// directory is missing.
func CollectAgents(resolver *paths.Resolver) []SearchEntry {
	if resolver == nil || resolver.ProjectRoot() == "" {
		return nil
	}

	agentsDir := resolver.AgentsDir()
	dirEntries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil
	}

	var entries []SearchEntry
	for _, de := range dirEntries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(agentsDir, de.Name()))
		if err != nil {
			continue
		}

		fm, err := agent.ParseAgentFrontmatter(data)
		if err != nil {
			continue
		}
		if fm.Name == "" {
			continue
		}

		// Use first line of Description as summary; Role as action label.
		summary := firstLine(fm.Description)
		action := fm.Role
		if action == "" {
			action = summary
		}

		entries = append(entries, SearchEntry{
			Name:        fm.Name,
			Domain:      DomainAgent,
			Summary:     summary,
			Description: fm.Description,
			Action:      action,
			Keywords:    extractKeywords(fm.Description),
		})
	}
	return entries
}

// firstLine returns the first non-empty line of s, trimmed.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return s
}

// dronemaMeta holds the frontmatter fields we care about in dromena files.
type dronemaMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// CollectDromena returns entries from .claude/commands/ directory.
// Walks recursively for .md files and parses frontmatter.
// Returns an empty slice if resolver is nil, has no project root, or the
// directory is missing.
func CollectDromena(resolver *paths.Resolver) []SearchEntry {
	if resolver == nil || resolver.ProjectRoot() == "" {
		return nil
	}

	commandsDir := filepath.Join(resolver.ClaudeDir(), "commands")

	var entries []SearchEntry
	err := filepath.WalkDir(commandsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Directory missing or unreadable — skip gracefully.
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		yamlBytes, _, parseErr := frontmatter.Parse(data)
		if parseErr != nil {
			// No frontmatter — skip.
			return nil
		}

		var meta dronemaMeta
		if err := yaml.Unmarshal(yamlBytes, &meta); err != nil || meta.Name == "" {
			return nil
		}

		kw := extractKeywords(meta.Description)
		entries = append(entries, SearchEntry{
			Name:        meta.Name,
			Domain:      DomainDromena,
			Summary:     firstLine(meta.Description),
			Description: meta.Description,
			Action:      "/" + meta.Name,
			Keywords:    kw,
		})
		return nil
	})
	if err != nil {
		return nil
	}

	return entries
}

// orchestratorFile mirrors the relevant sections of orchestrator.yaml.
type orchestratorFile struct {
	Rite struct {
		Name   string `yaml:"name"`
		Domain string `yaml:"domain"`
	} `yaml:"rite"`
	Frontmatter struct {
		Description string `yaml:"description"`
	} `yaml:"frontmatter"`
	Routing map[string]string `yaml:"routing"`
}

// CollectRouting returns entries from orchestrator.yaml routing sections
// found under the knossos rites directory.
// Returns an empty slice if resolver is nil, has no project root, or no
// orchestrator files are readable.
func CollectRouting(resolver *paths.Resolver) []SearchEntry {
	if resolver == nil || resolver.ProjectRoot() == "" {
		return nil
	}

	ritesDir := resolver.RitesDir()

	dirEntries, err := os.ReadDir(ritesDir)
	if err != nil {
		return nil
	}

	var entries []SearchEntry
	for _, de := range dirEntries {
		if !de.IsDir() {
			continue
		}

		orchPath := filepath.Join(ritesDir, de.Name(), "orchestrator.yaml")
		data, err := os.ReadFile(orchPath)
		if err != nil {
			continue
		}

		var orch orchestratorFile
		if err := yaml.Unmarshal(data, &orch); err != nil {
			continue
		}

		kw := extractKeywords(orch.Frontmatter.Description)

		for specialist, trigger := range orch.Routing {
			entries = append(entries, SearchEntry{
				Name:        specialist,
				Domain:      DomainRouting,
				Summary:     trigger,
				Description: orch.Frontmatter.Description,
				Action:      trigger,
				Keywords:    kw,
			})
		}
	}

	return entries
}
