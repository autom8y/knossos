// Package agent implements ari agent commands.
// summon.go implements the `ari agent summon {name}` subcommand.
package agent

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	agentpkg "github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// standingAgents are hard-coded system agents that cannot be summoned or dismissed.
// These are core platform infrastructure.
var standingAgents = map[string]bool{
	"pythia": true,
	"moirai": true,
	"metis":  true,
}

// knossosOnlySummonFields are knossos-internal frontmatter fields stripped during summon
// projection. This mirrors the list in internal/materialize/agent_transform.go but is
// maintained independently to avoid coupling the summon command to the materialize package.
var knossosOnlySummonFields = []string{
	"type",
	"role",
	"upstream",
	"downstream",
	"produces",
	"contract",
	"schema_version",
	"write-guard",
	"aliases",
	"skill_policy_exclude",
	"skill_policy_override",
	"tier", // summon-specific: CC does not need tier metadata
}

type summonOptions struct {
	name string // populated from args[0]
}

func newSummonCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summon <name>",
		Short: "Summon an agent to your user-level Claude config",
		Long: `Summons a named agent to your user-level Claude configuration (~/.claude/agents/).

Summonable agents are those published with tier: summonable in their source
frontmatter. Standing agents (pythia, moirai, metis) cannot be summoned.

Examples:
  ari agent summon theoros       # Summon the theoros agent
  ari agent summon naxos         # Summon the naxos agent
  ari agent roster               # See available summonables first`,
		Args: common.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSummon(ctx, args, summonOptions{name: args[0]})
		},
	}

	return cmd
}

func runSummon(ctx *cmdContext, args []string, opts summonOptions) error {
	printer := ctx.GetPrinter(output.FormatText)

	name := opts.name

	// Guard: check standing agent deny-list
	if standingAgents[name] {
		err := errors.NewWithDetails(errors.CodeValidationFailed,
			fmt.Sprintf("%q is a standing agent and cannot be summoned", name),
			map[string]any{"agent": name, "standing_agents": []string{"pythia", "moirai", "metis"}})
		return common.PrintAndReturn(printer, err)
	}

	// Resolve agent source content
	content, err := resolveAgentSource(name)
	if err != nil {
		return common.PrintAndReturn(printer, err)
	}

	// Parse frontmatter to validate it's a valid agent
	fm, parseErr := agentpkg.ParseAgentFrontmatter(content)
	if parseErr != nil {
		err := errors.Wrap(errors.CodeParseError,
			fmt.Sprintf("agent %q has invalid frontmatter", name), parseErr)
		return common.PrintAndReturn(printer, err)
	}
	// Validate required fields
	if fm.Name == "" && fm.Description == "" {
		err := errors.New(errors.CodeValidationFailed,
			fmt.Sprintf("agent %q does not appear to be a valid agent (missing name/description)", name))
		return common.PrintAndReturn(printer, err)
	}

	// Check tier field: if present, must be "summonable"
	tier := extractTierField(content)
	if tier != "" && tier != "summonable" {
		slog.Warn("agent tier is not 'summonable'", "agent", name, "tier", tier)
		printer.PrintLine(fmt.Sprintf("warning: agent %q has tier %q (expected 'summonable') — proceeding anyway", name, tier))
	}

	// Collision check: get project knossos dir from resolver, check against rite provenance
	resolver := ctx.GetResolver()
	knossosDir := resolver.KnossosDir()
	checker := newCollisionCheckerForSummon(knossosDir)
	manifestKey := "agents/" + name + ".md"
	if checker.IsEffective() {
		if collides, reason := checker.CheckCollision(manifestKey); collides {
			err := errors.NewWithDetails(errors.CodeValidationFailed,
				fmt.Sprintf("agent %q collides with a rite-owned agent %s", name, reason),
				map[string]any{"agent": name, "key": manifestKey})
			return common.PrintAndReturn(printer, err)
		}
	}

	// Transform: strip knossos fields, inject name
	transformed, transformErr := transformForSummon(content, name)
	if transformErr != nil {
		err := errors.Wrap(errors.CodeGeneralError,
			fmt.Sprintf("failed to transform agent %q for summon", name), transformErr)
		return common.PrintAndReturn(printer, err)
	}

	// Resolve user channel dir
	userChannelDir, pathErr := paths.UserChannelDir("claude")
	if pathErr != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to resolve user channel directory", pathErr)
		return common.PrintAndReturn(printer, err)
	}

	// Ensure agents directory exists
	agentsDir := filepath.Join(userChannelDir, "agents")
	if mkErr := os.MkdirAll(agentsDir, 0755); mkErr != nil {
		err := errors.Wrap(errors.CodePermissionDenied,
			fmt.Sprintf("failed to create agents directory: %s", agentsDir), mkErr)
		return common.PrintAndReturn(printer, err)
	}

	// Write to ~/.claude/agents/{name}.md
	targetPath := filepath.Join(agentsDir, name+".md")
	if writeErr := os.WriteFile(targetPath, transformed, 0644); writeErr != nil {
		err := errors.Wrap(errors.CodePermissionDenied,
			fmt.Sprintf("failed to write agent file: %s", targetPath), writeErr)
		return common.PrintAndReturn(printer, err)
	}

	// Update USER_PROVENANCE_MANIFEST.yaml
	manifestPath := provenance.UserManifestPath(userChannelDir)
	manifest, loadErr := provenance.LoadOrBootstrap(manifestPath)
	if loadErr != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to load user provenance manifest", loadErr)
		return common.PrintAndReturn(printer, err)
	}

	cs := checksum.Bytes(transformed)
	manifest.Entries[manifestKey] = provenance.NewKnossosEntry(
		provenance.ScopeUser,
		"summon:"+name,
		"summon",
		cs,
		"", // channel "" normalizes to claude per NewKnossosEntry convention
	)
	manifest.LastSync = time.Now().UTC()

	if saveErr := provenance.Save(manifestPath, manifest); saveErr != nil {
		// Non-fatal: agent was written; provenance is tracking metadata only
		slog.Warn("failed to save user provenance manifest", "error", saveErr)
		printer.PrintLine(fmt.Sprintf("warning: agent written but provenance update failed: %s", saveErr))
	}

	printer.PrintLine(fmt.Sprintf("%s summoned. Restart CC to activate.", name))
	return nil
}

// resolveAgentSource finds the agent source file by name.
// Search order:
//  1. $KNOSSOS_HOME/agents/{name}.md
//  2. Embedded agents FS
//
// Returns error with list of available agents if not found.
func resolveAgentSource(name string) ([]byte, error) {
	// 1. KNOSSOS_HOME filesystem
	knossosHome := os.Getenv("KNOSSOS_HOME")
	if knossosHome != "" {
		p := filepath.Join(knossosHome, "agents", name+".md")
		if data, err := os.ReadFile(p); err == nil {
			return data, nil
		}
	}

	// 2. Embedded agents FS
	embeddedAgents := common.EmbeddedAgents()
	if embeddedAgents != nil {
		data, err := fs.ReadFile(embeddedAgents, "agents/"+name+".md")
		if err == nil {
			return data, nil
		}
	}

	// Not found: build list of available summonables for the error message
	available := listAvailableAgentNames(knossosHome, embeddedAgents)
	return nil, errors.NewWithDetails(errors.CodeFileNotFound,
		fmt.Sprintf("agent %q not found in agent sources", name),
		map[string]any{
			"agent":     name,
			"available": available,
		})
}

// listAvailableAgentNames scans agent sources and returns names of available agents.
// Filters out standing agents.
func listAvailableAgentNames(knossosHome string, embeddedAgents fs.FS) []string {
	seen := make(map[string]bool)
	var names []string

	addFromDir := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			agentName := strings.TrimSuffix(e.Name(), ".md")
			if standingAgents[agentName] || seen[agentName] {
				continue
			}
			seen[agentName] = true
			names = append(names, agentName)
		}
	}

	if knossosHome != "" {
		addFromDir(filepath.Join(knossosHome, "agents"))
	}

	if embeddedAgents != nil {
		entries, err := fs.ReadDir(embeddedAgents, "agents")
		if err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				agentName := strings.TrimSuffix(e.Name(), ".md")
				if standingAgents[agentName] || seen[agentName] {
					continue
				}
				seen[agentName] = true
				names = append(names, agentName)
			}
		}
	}

	return names
}

// extractTierField extracts the tier field value from agent frontmatter.
// Returns empty string if frontmatter is invalid or tier is absent.
func extractTierField(content []byte) string {
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

// transformForSummon performs a lightweight transform of agent source content
// for placement in the user-level channel directory.
//
// This is a parallel, simpler transform versus the full materialize pipeline:
//   - No write-guard resolution
//   - No skill policy application
//   - No MCP server injection
//   - No model override
//   - No channel body substitution
//
// Steps:
//  1. Parse frontmatter
//  2. Strip knossos-only fields (including tier)
//  3. Inject name
//  4. Reconstruct content
func transformForSummon(content []byte, agentName string) ([]byte, error) {
	yamlBytes, body, err := frontmatter.Parse(content)
	if err != nil {
		// No valid frontmatter — pass through unchanged
		return content, nil
	}

	var fmMap map[string]any
	if err := yaml.Unmarshal(yamlBytes, &fmMap); err != nil {
		// Invalid YAML — pass through unchanged
		return content, nil
	}

	// Strip knossos-only fields
	for _, field := range knossosOnlySummonFields {
		delete(fmMap, field)
	}

	// Inject name from the summon target name (matches file basename convention)
	fmMap["name"] = agentName

	// Reconstruct: ---\n + yaml + ---\n + body
	yamlOut, err := yaml.Marshal(fmMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transformed frontmatter: %w", err)
	}

	result := []byte("---\n")
	result = append(result, yamlOut...)
	result = append(result, []byte("---\n")...)
	result = append(result, body...)
	return result, nil
}

// newCollisionCheckerForSummon creates a CollisionChecker for summon validation.
// Uses the userscope package's checker but imported via the public type.
// If the knossos dir doesn't exist (e.g., in tests), returns a no-op checker.
func newCollisionCheckerForSummon(knossosDir string) collisionChecker {
	if _, err := os.Stat(knossosDir); os.IsNotExist(err) {
		return &noopCollisionChecker{}
	}
	return newRiteCollisionChecker(knossosDir)
}

// collisionChecker is an interface for rite-scope collision detection.
// Enables test injection without importing the userscope package directly.
type collisionChecker interface {
	IsEffective() bool
	CheckCollision(key string) (bool, string)
}

// noopCollisionChecker is a no-op implementation that never detects collisions.
// Used when the knossos directory is absent (e.g., project not initialized).
type noopCollisionChecker struct{}

func (n *noopCollisionChecker) IsEffective() bool                     { return false }
func (n *noopCollisionChecker) CheckCollision(string) (bool, string)  { return false, "" }

// riteCollisionChecker wraps provenance manifest to detect rite scope collisions.
// This reimplements the core logic from userscope.CollisionChecker inline to avoid
// a package import cycle: agent cmd -> userscope -> provenance (already imported here).
type riteCollisionChecker struct {
	riteEntries    map[string]bool
	manifestLoaded bool
}

func newRiteCollisionChecker(knossosDir string) *riteCollisionChecker {
	c := &riteCollisionChecker{}
	manifestPath := provenance.ManifestPathForChannel(knossosDir, "")
	manifest, err := provenance.Load(manifestPath)
	if err != nil {
		return c
	}
	c.manifestLoaded = true
	c.riteEntries = make(map[string]bool)
	for key, entry := range manifest.Entries {
		if entry.Scope == provenance.ScopeRite && entry.Owner == provenance.OwnerKnossos {
			c.riteEntries[key] = true
		}
	}
	return c
}

func (c *riteCollisionChecker) IsEffective() bool { return c.manifestLoaded }

func (c *riteCollisionChecker) CheckCollision(manifestKey string) (bool, string) {
	if !c.manifestLoaded || len(c.riteEntries) == 0 {
		return false, ""
	}
	if c.riteEntries[manifestKey] {
		return true, "(from manifest)"
	}
	for riteKey := range c.riteEntries {
		if strings.HasSuffix(riteKey, "/") && strings.HasPrefix(manifestKey, riteKey) {
			return true, "(inside rite directory)"
		}
	}
	return false, ""
}
