// Package lint implements the ari lint command for source validation.
package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/mena"
	"github.com/autom8y/knossos/internal/output"
)

// skillAtPattern matches @skill-name references in body content.
// Excludes email-style patterns (@user.com), @handles in examples,
// and regex patterns. Matches: @word-word at word boundaries.
var skillAtPattern = regexp.MustCompile(`(?m)(?:^|[\s(` + "`" + `])@([a-z][a-z0-9](?:[a-z0-9-]*[a-z0-9])?)(?:[#/\s` + "`" + `.,;:)\]]|$)`)

// Source path leak detection patterns.
var (
	// Read() calls with knossos source paths — functional breakage in satellites.
	readSourcePathPattern = regexp.MustCompile(`Read\("rites/`)
	// rites/*/mena/ references in links or documentation.
	riteMenaPathPattern = regexp.MustCompile(`rites/[^/]+/mena/`)
	// Source extensions (.lego.md or .dro.md) in body references.
	sourceExtPattern = regexp.MustCompile(`\.(lego|dro)\.md`)
	// Inline backtick content — stripped before pattern matching to avoid
	// false positives on paths used as literal examples in prose.
	inlineCodePattern = regexp.MustCompile("`[^`]+`")
)

// Severity levels for lint findings.
const (
	SevCritical = "CRIT"
	SevHigh     = "HIGH"
	SevMedium   = "MED"
	SevLow      = "LOW"
)

// workflowCommands lists commands that must be model-invocable (no disable-model-invocation).
// These are the pipeline commands that /one invokes autonomously during orchestrated sessions.
var workflowCommands = map[string]bool{
	"spike": true, "hotfix": true, "sprint": true, "task": true,
	"build": true, "architect": true, "qa": true,
	"pr": true, "code-review": true, "commit": true,
}

// expectedForkState defines the deliberate fork/inline classification for each dromena.
//
// Classification principle (context-isolation semantics):
//
//	true  = FORK: command produces artifacts, reports, or side effects. A summary
//	        returning to the main conversation is sufficient. Forked commands run
//	        as subagents and CANNOT use the Task tool (SCAR-018). Commands that
//	        dispatch subagents via Task must be INLINE.
//	false = INLINE: command guides the ongoing conversation, manages session state,
//	        dispatches subagents via Task, or needs context to persist.
var expectedForkState = map[string]bool{
	// === INLINE (false): persistent context needed ===
	// Session lifecycle — hook-injected context must persist
	"go": false, "start": false, "park": false, "continue": false,
	"wrap": false, "handoff": false, "fray": false, "sos": false,
	// Interactive orchestration — guide ongoing conversation
	"consult": false, "task": false, "sprint": false, "hotfix": false,
	"build": false, "architect": false, "qa": false, "commit": false,
	// Context management — shapes subsequent work
	"frame": false, "one": false,
	// Task-dispatching orchestrators — fork blocks Task tool (SCAR-018)
	"know": false, "dion": false, "land": false, "radar": false,
	"theoria": false, "spike": false, "release": false,
	"consolidate": false, "eval-agent": false, "validate-rite": false,
	"forge-rite": false, "new-rite": false,
	"minus-1": false, "zero": false,

	// === FORK (true): isolated execution, summary sufficient ===
	// Workflow pipeline commands — model-invocable, inline for /one autonomy
	"pr": false, "code-review": false,
	"research": true, "interview": true,
	// CLI wrappers / one-shot actions
	"rite": true, "sessions": true,
	"worktree": true, "sync": true, "sync-debug": true, "ecosystem": true,
	// Rite-switching (one-shot CLI wrappers)
	"10x": true, "debt": true, "docs": true, "forge": true,
	"hygiene": true, "intelligence": true, "rnd": true,
	"security": true, "sre": true, "strategy": true,
	"arch": true, "slop-chop": true, "clinic": true,
	"releaser": true, "review": true, "thermia": true,
}

// Finding is a single lint issue.
type Finding struct {
	File     string `json:"file"`
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}

// LintReport is the full lint output.
type LintReport struct {
	Agents   []Finding `json:"agents,omitempty"`
	Dromena  []Finding `json:"dromena,omitempty"`
	Legomena []Finding `json:"legomena,omitempty"`
	Summary  struct {
		Total    int `json:"total"`
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Low      int `json:"low"`
		Files    int `json:"files_checked"`
	} `json:"summary"`
}

// Text implements output.Textable.
func (r LintReport) Text() string {
	var b strings.Builder

	printSection := func(name string, findings []Finding) {
		if len(findings) == 0 {
			return
		}
		fmt.Fprintf(&b, "\n%s\n", name)
		b.WriteString(strings.Repeat("-", len(name)) + "\n")
		for _, f := range findings {
			fmt.Fprintf(&b, "  [%s] %s: %s (%s)\n", f.Severity, f.File, f.Message, f.Rule)
		}
	}

	b.WriteString("Lint Report: Knossos Source Validation\n")
	b.WriteString(strings.Repeat("=", 40) + "\n")

	printSection("Agents", r.Agents)
	printSection("Dromena", r.Dromena)
	printSection("Legomena", r.Legomena)

	fmt.Fprintf(&b, "\nSummary: %d issues (%d critical, %d high, %d medium, %d low) across %d files\n",
		r.Summary.Total, r.Summary.Critical, r.Summary.High, r.Summary.Medium, r.Summary.Low, r.Summary.Files)

	return b.String()
}

type cmdContext struct {
	common.BaseContext
}

// NewLintCmd creates the lint command.
func NewLintCmd(outputFlag *string, verboseFlag *bool, projectDir *string) *cobra.Command {
	ctx := &cmdContext{
		BaseContext: common.BaseContext{
			Output:     outputFlag,
			Verbose:    verboseFlag,
			ProjectDir: projectDir,
		},
	}

	var scope string

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate mena and agent sources before sync",
		Long: `Lint validates source files to catch errors before projection.

Checks agents, dromena (.dro.md), and legomena (.lego.md) for:
- Missing or malformed frontmatter
- Required fields (name, description, etc.)
- Agent archetype mismatches (maxTurns, type, color)
- Dromena context:fork allowlist mismatches
- Dromena context:fork + Task tool conflicts (SCAR-018)
- Workflow commands must be model-invocable
- Legomena missing Triggers keyword in description

Examples:
  ari lint                    # Lint all sources
  ari lint --scope=agents     # Agents only
  ari lint --scope=dromena    # Dromena only
  ari lint --scope=legomena   # Legomena only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(ctx, scope)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Limit to: agents, dromena, or legomena")

	common.SetNeedsProject(cmd, true, true)

	return cmd
}

func runLint(ctx *cmdContext, scope string) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()
	projectRoot := resolver.ProjectRoot()
	sources := buildAllMenaSources(projectRoot)

	report := &LintReport{}

	if scope == "" || scope == "agents" {
		lintAgents(projectRoot, report)
	}
	if scope == "" || scope == "dromena" {
		lintDromena(projectRoot, sources, report)
		lintMenaNamespace(projectRoot, sources, report)
	}
	if scope == "" || scope == "legomena" {
		lintLegomena(projectRoot, sources, report)
	}
	// Session artifact boundary check runs for all scopes (it checks shared mena dirs)
	if scope == "" {
		lintSessionArtifactsInSharedMena(projectRoot, report)
	}

	// Compute summary
	all := append(append(report.Agents, report.Dromena...), report.Legomena...)
	report.Summary.Total = len(all)
	for _, f := range all {
		switch f.Severity {
		case SevCritical:
			report.Summary.Critical++
		case SevHigh:
			report.Summary.High++
		case SevMedium:
			report.Summary.Medium++
		case SevLow:
			report.Summary.Low++
		}
	}

	return printer.Print(*report)
}

// --- Agent linting ---

type agentFrontmatter struct {
	Name             string                          `yaml:"name"`
	Description      string                          `yaml:"description"`
	Type             string                          `yaml:"type"`
	Tools            frontmatter.FlexibleStringSlice `yaml:"tools"`
	Model            string                          `yaml:"model"`
	Color            string                          `yaml:"color"`
	MaxTurns         int                             `yaml:"maxTurns"`
	MaxTurnsOverride bool                            `yaml:"maxTurns-override"`
	Domain           string                          `yaml:"domain"`
}

// approvedAgentNames is the naming provenance registry. All agent names must
// be registered here. Values indicate tier classification:
//   - "tier-1": Linear B attested (Bronze Age archaeological evidence)
//   - "tier-2": Classical Greek (Homer, Hesiod, Herodotus, etc.)
//   - "tier-3": Hellenistic/Scholarly (Panhellenic practice, Alexandrian scholarship)
//   - "functional": Non-mythological descriptive name
var approvedAgentNames = map[string]string{
	// Tier 1: Linear B attested
	"potnia": "tier-1", // KN Gg 702: da-pu₂-ri-to-jo po-ti-ni-ja

	// Tier 2: Classical Greek
	"dionysus": "tier-2", // Homer, Hesiod, Euripides
	"moirai":   "tier-2", // Hesiod Theogony, Plato Republic
	"pythia":   "tier-2", // Herodotus, Pausanias (Delphic oracle)

	// Tier 3: Hellenistic/Scholarly
	"theoros": "tier-3", // Panhellenic practice (Herodotus, Thucydides)

	// Functional: non-mythological descriptive names
	"analytics-engineer":       "functional",
	"agent-curator":            "functional",
	"agent-designer":           "functional",
	"architect":                "functional",
	"architect-enforcer":       "functional",
	"attending":                "functional",
	"audit-lead":               "functional",
	"business-model-analyst":   "functional",
	"capacity-engineer":        "functional",
	"cartographer":             "functional",
	"case-reporter":            "functional",
	"chaos-engineer":           "functional",
	"code-smeller":             "functional",
	"compatibility-tester":     "functional",
	"compliance-architect":     "functional",
	"competitive-analyst":      "functional",
	"context-architect":        "functional",
	"context-engineer":         "functional",
	"cruft-cutter":             "functional",
	"debt-collector":           "functional",
	"dependency-analyst":       "functional",
	"dependency-resolver":      "functional",
	"diagnostician":            "functional",
	"doc-auditor":              "functional",
	"doc-reviewer":             "functional",
	"documentation-engineer":   "functional",
	"domain-forensics":         "functional",
	"ecosystem-analyst":        "functional",
	"eval-specialist":          "functional",
	"experimentation-lead":     "functional",
	"gate-keeper":              "functional",
	"hallucination-hunter":     "functional",
	"heat-mapper":              "functional",
	"incident-commander":       "functional",
	"information-architect":    "functional",
	"insights-analyst":         "functional",
	"integration-engineer":     "functional",
	"integration-researcher":   "functional",
	"janitor":                  "functional",
	"logic-surgeon":            "functional",
	"market-researcher":        "functional",
	"moonshot-architect":       "functional",
	"observability-engineer":   "functional",
	"pathologist":              "functional",
	"pattern-profiler":         "functional",
	"penetration-tester":       "functional",
	"pipeline-monitor":         "functional",
	"platform-engineer":        "functional",
	"principal-engineer":       "functional",
	"prompt-architect":         "functional",
	"prototype-engineer":       "functional",
	"qa-adversary":             "functional",
	"release-executor":         "functional",
	"release-planner":          "functional",
	"remedy-smith":             "functional",
	"remediation-planner":      "functional",
	"requirements-analyst":     "functional",
	"risk-assessor":            "functional",
	"roadmap-strategist":       "functional",
	"security-reviewer":        "functional",
	"signal-sifter":            "functional",
	"sprint-planner":           "functional",
	"structure-evaluator":      "functional",
	"systems-thermodynamicist": "functional",
	"tech-transfer":            "functional",
	"tech-writer":              "functional",
	"technology-scout":         "functional",
	"thermal-monitor":          "functional",
	"threat-modeler":           "functional",
	"topology-cartographer":    "functional",
	"triage-nurse":             "functional",
	"user-researcher":          "functional",
	"workflow-engineer":        "functional",
}

var archetypeMaxTurns = map[string]int{
	"orchestrator": 40,
	"specialist":   150,
	"analyst":      150,
	"designer":     150,
	"engineer":     150,
	"reviewer":     100,
}

func lintAgents(projectRoot string, report *LintReport) {
	// Lint agents from rite sources (not .claude/ — we lint sources)
	agentDirs := findAgentDirs(projectRoot)

	for _, dir := range agentDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			report.Summary.Files++
			path := filepath.Join(dir, entry.Name())
			relPath := mustRel(projectRoot, path)
			lintAgentFile(path, relPath, report)
		}
	}
}

func findAgentDirs(projectRoot string) []string {
	var dirs []string
	// Check rites/*/agents/
	riteDir := filepath.Join(projectRoot, "rites")
	rites, _ := os.ReadDir(riteDir)
	for _, r := range rites {
		if r.IsDir() {
			d := filepath.Join(riteDir, r.Name(), "agents")
			if info, err := os.Stat(d); err == nil && info.IsDir() {
				dirs = append(dirs, d)
			}
		}
	}
	return dirs
}

func lintAgentFile(path, relPath string, report *LintReport) {
	data, err := os.ReadFile(path)
	if err != nil {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevCritical, Rule: "read-error",
			Message: fmt.Sprintf("cannot read file: %v", err),
		})
		return
	}

	yamlBytes, body, err := frontmatter.Parse(data)
	if err != nil {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevCritical, Rule: "frontmatter-missing",
			Message: "no valid frontmatter found",
		})
		return
	}

	var fm agentFrontmatter
	if err := yaml.Unmarshal(yamlBytes, &fm); err != nil {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevCritical, Rule: "frontmatter-parse",
			Message: fmt.Sprintf("invalid YAML: %v", err),
		})
		return
	}

	// Required fields
	if fm.Name == "" {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevCritical, Rule: "name-missing",
			Message: "frontmatter missing 'name' field",
		})
	}
	if fm.Description == "" {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevCritical, Rule: "description-missing",
			Message: "frontmatter missing 'description' field",
		})
	}
	if fm.Type == "" {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevHigh, Rule: "type-missing",
			Message: "frontmatter missing 'type' field",
		})
	}

	// Type validation
	if fm.Type != "" {
		if _, ok := archetypeMaxTurns[fm.Type]; !ok {
			report.Agents = append(report.Agents, Finding{
				File: relPath, Severity: SevHigh, Rule: "type-invalid",
				Message: fmt.Sprintf("type %q not a recognized archetype (orchestrator|specialist|analyst|designer|engineer|reviewer)", fm.Type),
			})
		}
	}

	// maxTurns archetype check — generous tolerance for orchestrators
	// which need 20-40 turns for multi-phase coordination.
	// Agents with maxTurns-override: true have intentional deviations.
	if fm.MaxTurnsOverride {
		// Intentional deviation acknowledged — skip check
	} else if fm.Type != "" && fm.MaxTurns > 0 {
		if expected, ok := archetypeMaxTurns[fm.Type]; ok {
			deviation := fm.MaxTurns - expected
			if deviation < 0 {
				deviation = -deviation
			}
			// Threshold: 50% of archetype default or 50, whichever is larger
			threshold := max(expected/2, 50)
			if deviation > threshold {
				report.Agents = append(report.Agents, Finding{
					File: relPath, Severity: SevMedium, Rule: "maxTurns-deviation",
					Message: fmt.Sprintf("maxTurns=%d deviates from %s archetype default of %d by %d", fm.MaxTurns, fm.Type, expected, deviation),
				})
			}
		}
	}

	// Description quality (non-orchestrators should be multi-line)
	if fm.Type != "orchestrator" && fm.Description != "" {
		lines := strings.Split(strings.TrimSpace(fm.Description), "\n")
		if len(lines) == 1 && len(fm.Description) < 200 {
			report.Agents = append(report.Agents, Finding{
				File: relPath, Severity: SevHigh, Rule: "description-single-line",
				Message: "non-orchestrator agent has single-line description (should include use cases and examples)",
			})
		}
	}

	// Valid CC color check
	validAgentColors := map[string]bool{
		"red": true, "blue": true, "green": true, "yellow": true,
		"purple": true, "orange": true, "pink": true, "cyan": true,
	}
	if fm.Color != "" && !validAgentColors[fm.Color] {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevHigh, Rule: "agent-invalid-color",
			Message: fmt.Sprintf("color=%q is not a valid CC color (valid: red, blue, green, yellow, purple, orange, pink, cyan)", fm.Color),
		})
	}

	// File size warning for embedded reference
	if len(data) > 15000 { // ~200 lines
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevMedium, Rule: "agent-oversized",
			Message: fmt.Sprintf("agent file is %d bytes — consider extracting reference content to skills", len(data)),
		})
	}

	// @skill-name anti-pattern check (body content only)
	checkSkillAtRefs(string(body), relPath, &report.Agents)

	// Naming provenance check — all agent names must be registered
	if fm.Name != "" {
		if _, ok := approvedAgentNames[fm.Name]; !ok {
			report.Agents = append(report.Agents, Finding{
				File: relPath, Severity: SevLow, Rule: "naming-provenance",
				Message: fmt.Sprintf("agent name %q is not in the approved naming registry — register in lint.go approvedAgentNames", fm.Name),
			})
		}
	}

	// Domain-jurisdiction check — orchestrator domain must match rite directory
	if fm.Domain != "" && fm.Type == "orchestrator" && strings.HasPrefix(relPath, "rites/") {
		parts := strings.SplitN(relPath, "/", 4) // rites/{rite}/agents/{file}
		if len(parts) >= 2 {
			riteName := parts[1]
			if fm.Domain != riteName {
				report.Agents = append(report.Agents, Finding{
					File: relPath, Severity: SevMedium, Rule: "domain-jurisdiction",
					Message: fmt.Sprintf("orchestrator domain %q does not match rite directory %q", fm.Domain, riteName),
				})
			}
		}
	}
}

// --- Dromena linting ---

func lintDromena(projectRoot string, sources []mena.MenaSource, report *LintReport) {
	mena.Walk(sources, ".dro.md", func(entry mena.WalkEntry) {
		report.Summary.Files++
		relPath := mustRel(projectRoot, entry.Path)
		lintDromenFile(entry.Path, relPath, entry.Data, report)
	})
}

func lintDromenFile(_, relPath string, data []byte, report *LintReport) {
	yamlBytes, body, err := frontmatter.Parse(data)
	if err != nil {
		report.Dromena = append(report.Dromena, Finding{
			File: relPath, Severity: SevCritical, Rule: "frontmatter-missing",
			Message: "no valid frontmatter found",
		})
		return
	}

	// Use flexible map parsing since argument-hint uses YAML-incompatible bracket syntax
	fm := parseFrontmatterLenient(yamlBytes)
	if fm == nil {
		report.Dromena = append(report.Dromena, Finding{
			File: relPath, Severity: SevCritical, Rule: "frontmatter-parse",
			Message: "frontmatter has no parseable fields",
		})
		return
	}

	if strVal(fm, "name") == "" {
		report.Dromena = append(report.Dromena, Finding{
			File: relPath, Severity: SevCritical, Rule: "name-missing",
			Message: "frontmatter missing 'name' field",
		})
	}
	if strVal(fm, "description") == "" {
		report.Dromena = append(report.Dromena, Finding{
			File: relPath, Severity: SevHigh, Rule: "description-missing",
			Message: "frontmatter missing 'description' field",
		})
	}

	// context:fork allowlist check — enforce deliberate fork/inline classification
	name := strVal(fm, "name")
	hasFork := strVal(fm, "context") == "fork"
	isRiteScoped := strings.HasPrefix(relPath, "rites/")
	if expected, known := expectedForkState[name]; known {
		if expected && !hasFork {
			report.Dromena = append(report.Dromena, Finding{
				File: relPath, Severity: SevMedium, Rule: "context-fork-expected",
				Message: fmt.Sprintf("dromena %q should have context: fork (self-contained command)", name),
			})
		} else if !expected && hasFork && !isRiteScoped {
			// Only flag for global-scope dromena. Rite-scoped overrides (e.g.,
			// 10x-dev/architect-ref with fork vs global /architect without)
			// are intentional — the rite version is a self-contained reference
			// command, not the global interactive version.
			report.Dromena = append(report.Dromena, Finding{
				File: relPath, Severity: SevMedium, Rule: "context-fork-unexpected",
				Message: fmt.Sprintf("dromena %q should NOT have context: fork (interactive/contextual command)", name),
			})
		}
	} else if name != "" {
		report.Dromena = append(report.Dromena, Finding{
			File: relPath, Severity: SevLow, Rule: "context-fork-unclassified",
			Message: fmt.Sprintf("dromena %q not in fork/inline allowlist — classify in lint.go expectedForkState", name),
		})
	}

	// SCAR-018: context:fork + Task tool conflict — forked commands cannot dispatch subagents
	if hasFork {
		allowedTools := strVal(fm, "allowed-tools")
		if strings.Contains(allowedTools, "Task") {
			report.Dromena = append(report.Dromena, Finding{
				File: relPath, Severity: SevHigh, Rule: "fork-task-conflict",
				Message: fmt.Sprintf("dromena %q has context: fork with Task in allowed-tools; forked commands run as subagents which cannot use Task (SCAR-018)", name),
			})
		}
	}

	// Workflow commands must be model-invocable for /one autonomous execution
	if workflowCommands[name] {
		if strVal(fm, "disable-model-invocation") == "true" {
			report.Dromena = append(report.Dromena, Finding{
				File: relPath, Severity: SevHigh, Rule: "workflow-not-model-invocable",
				Message: fmt.Sprintf("workflow dromena %q has disable-model-invocation: true; workflow commands must be model-invocable for /one pipeline autonomy", name),
			})
		}
	}

	// @skill-name anti-pattern check
	checkSkillAtRefs(string(body), relPath, &report.Dromena)

	// Source path leak check
	checkSourcePathLeaks(string(body), relPath, &report.Dromena)
}

// --- Dromena namespace collision detection ---

func lintMenaNamespace(projectRoot string, sources []mena.MenaSource, report *LintReport) {
	// Collect all dromena names to detect collisions.
	// Scope-aware: only flag collisions within the same scope (global or same rite).
	// Cross-scope shadowing (rite overrides global) is intentional.
	type nameSource struct {
		name    string
		relPath string
		scope   string // "global" or rite name
	}

	var entries []nameSource

	mena.Walk(sources, ".dro.md", func(entry mena.WalkEntry) {
		yamlBytes, _, err := frontmatter.Parse(entry.Data)
		if err != nil {
			return
		}
		fm := parseFrontmatterLenient(yamlBytes)
		if fm == nil {
			return
		}
		name := strVal(fm, "name")
		if name == "" {
			return // already flagged by name-missing rule
		}

		relPath := mustRel(projectRoot, entry.Path)

		// Determine scope from path
		scope := "global"
		if strings.HasPrefix(relPath, "rites/") {
			parts := strings.SplitN(relPath, "/", 3)
			if len(parts) >= 2 {
				scope = parts[1] // rite name
			}
		}

		entries = append(entries, nameSource{name: name, relPath: relPath, scope: scope})
	})

	// Build (scope, name) → files map — only flag within-scope collisions
	type scopeKey struct {
		scope string
		name  string
	}
	scopeFiles := make(map[scopeKey][]string)
	for _, e := range entries {
		key := scopeKey{scope: e.scope, name: e.name}
		scopeFiles[key] = append(scopeFiles[key], e.relPath)
	}

	// Flag collisions within same scope only
	for key, files := range scopeFiles {
		if len(files) < 2 {
			continue
		}
		for _, f := range files {
			report.Dromena = append(report.Dromena, Finding{
				File: f, Severity: SevHigh, Rule: "name-collision",
				Message: fmt.Sprintf("dromena name %q collides with %d other file(s) in %s scope: %s", key.name, len(files)-1, key.scope, strings.Join(files, ", ")),
			})
		}
	}
}

// --- Legomena linting ---

type legomenFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func lintLegomena(projectRoot string, sources []mena.MenaSource, report *LintReport) {
	mena.Walk(sources, ".lego.md", func(entry mena.WalkEntry) {
		report.Summary.Files++
		relPath := mustRel(projectRoot, entry.Path)
		lintLegomenFile(entry.Path, relPath, entry.Data, report)
	})
}

func lintLegomenFile(_, relPath string, data []byte, report *LintReport) {
	yamlBytes, body, err := frontmatter.Parse(data)
	if err != nil {
		report.Legomena = append(report.Legomena, Finding{
			File: relPath, Severity: SevCritical, Rule: "frontmatter-missing",
			Message: "no valid frontmatter found",
		})
		return
	}

	var fm legomenFrontmatter
	if err := yaml.Unmarshal(yamlBytes, &fm); err != nil {
		report.Legomena = append(report.Legomena, Finding{
			File: relPath, Severity: SevCritical, Rule: "frontmatter-parse",
			Message: fmt.Sprintf("invalid YAML: %v", err),
		})
		return
	}

	if fm.Name == "" {
		report.Legomena = append(report.Legomena, Finding{
			File: relPath, Severity: SevCritical, Rule: "name-missing",
			Message: "frontmatter missing 'name' field",
		})
	}
	if fm.Description == "" {
		report.Legomena = append(report.Legomena, Finding{
			File: relPath, Severity: SevHigh, Rule: "description-missing",
			Message: "frontmatter missing 'description' field",
		})
	}

	// Triggers keyword check
	if fm.Description != "" && !strings.Contains(strings.ToLower(fm.Description), "triggers:") {
		report.Legomena = append(report.Legomena, Finding{
			File: relPath, Severity: SevHigh, Rule: "triggers-missing",
			Message: "description lacks 'Triggers:' keyword for CC autonomous loading",
		})
	}

	// Monolithic file check
	if len(data) > 25000 { // ~500 lines
		report.Legomena = append(report.Legomena, Finding{
			File: relPath, Severity: SevMedium, Rule: "legomen-oversized",
			Message: fmt.Sprintf("legomen file is %d bytes — consider INDEX + companion pattern", len(data)),
		})
	}

	// @skill-name anti-pattern check
	checkSkillAtRefs(string(body), relPath, &report.Legomena)

	// Source path leak check
	checkSourcePathLeaks(string(body), relPath, &report.Legomena)
}

// --- Session artifact boundary check ---

// sessionArtifactPatterns detect session-specific content that should NOT
// appear in shared mena directories. Shared mena is permanent platform
// knowledge; session artifacts belong in .sos/wip/ (ephemeral).
var sessionArtifactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^session[_-]?id\s*:`),
	regexp.MustCompile(`(?i)^throughline\s*:`),
	regexp.MustCompile(`(?i)^session[_-]?ref\s*:`),
	regexp.MustCompile(`(?i)^sprint[_-]?session\s*:`),
}

// lintSessionArtifactsInMena walks platform mena/ and rites/shared/mena/,
// flagging any files with session-specific frontmatter. This enforces the
// SCAR-027 boundary: mena is permanent platform knowledge, not a dumping
// ground for session-scoped artifacts.
func lintSessionArtifactsInSharedMena(projectRoot string, report *LintReport) {
	menaDirs := []string{
		filepath.Join(projectRoot, "mena"),
		filepath.Join(projectRoot, "rites", "shared", "mena"),
	}
	for _, menaDir := range menaDirs {
		if _, err := os.Stat(menaDir); err != nil {
			continue
		}

		_ = filepath.WalkDir(menaDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				// Skip examples/ directories — they legitimately contain session_id
				// in template/example artifacts (not real session data).
				if d.Name() == "examples" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			report.Summary.Files++
			relPath := mustRel(projectRoot, path)

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			yamlBytes, _, parseErr := frontmatter.Parse(data)
			if parseErr != nil {
				return nil // No frontmatter -- not a session artifact
			}

			// Check frontmatter lines for session-specific fields
			fmLines := strings.Split(string(yamlBytes), "\n")
			for _, line := range fmLines {
				trimmed := strings.TrimSpace(line)
				for _, pat := range sessionArtifactPatterns {
					if pat.MatchString(trimmed) {
						report.Legomena = append(report.Legomena, Finding{
							File:     relPath,
							Severity: SevHigh,
							Rule:     "session-artifact-in-shared-mena",
							Message:  fmt.Sprintf("mena file contains session-specific field %q — session artifacts belong in .sos/wip/, not mena/", trimmed),
						})
						return nil // One finding per file is sufficient
					}
				}
			}

			return nil
		})
	}
}

// --- @skill-name check ---

// skillAtExclusions are known false positives: team handles, documentation examples.
var skillAtExclusions = map[string]bool{
	"api-team":      true,
	"product-lead":  true,
	"platform-team": true,
	"skill-name":    true, // appears in anti-pattern documentation
	"deprecated":    true, // JSDoc @deprecated
	"pytest":        true, // Python @pytest.mark
	"acme":          true, // npm @acme/sdk example
}

// checkSkillAtRefs scans body content for @skill-name references and appends findings.
func checkSkillAtRefs(body, relPath string, findings *[]Finding) {
	matches := skillAtPattern.FindAllStringSubmatch(body, -1)
	count := 0
	for _, m := range matches {
		if len(m) > 1 && skillAtExclusions[m[1]] {
			continue
		}
		count++
	}
	if count == 0 {
		return
	}
	*findings = append(*findings, Finding{
		File: relPath, Severity: SevHigh, Rule: "skill-at-syntax",
		Message: fmt.Sprintf("body contains %d @skill-name reference(s) — use plain skill name instead (see lexicon anti-patterns)", count),
	})
}

// --- Source path leak check ---

// checkSourcePathLeaks scans body content for knossos source paths that leak
// into materialized artifacts, causing failures or confusion in satellites.
func checkSourcePathLeaks(body, relPath string, findings *[]Finding) {
	lines := strings.Split(body, "\n")
	inCodeBlock := false

	var readLeaks, refLeaks, extLeaks int

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip lines that document the materialization pipeline itself
		lower := strings.ToLower(line)
		if strings.Contains(lower, "materializ") || strings.Contains(lower, "projected from") || strings.Contains(lower, "sync pipeline") {
			continue
		}

		// Pattern 1: Read() calls with rites/ source paths (HIGH)
		if readSourcePathPattern.MatchString(line) {
			readLeaks++
		}

		// Pattern 2: rites/*/mena/ in links or documentation (MEDIUM)
		// Only check outside code blocks and inline backticks to avoid false positives
		stripped := inlineCodePattern.ReplaceAllString(line, "")
		if !inCodeBlock && riteMenaPathPattern.MatchString(stripped) && !readSourcePathPattern.MatchString(stripped) {
			refLeaks++
		}

		// Pattern 3: .lego.md or .dro.md extensions in references (LOW)
		// Only in markdown link syntax or backtick paths, not in prose about the extension
		if !inCodeBlock && sourceExtPattern.MatchString(stripped) {
			// Exclude lines that discuss extensions conceptually (naming conventions,
			// audit criteria documenting what extensions mean, classification instructions).
			// Also exclude pinakes domain criteria files — these document source conventions
			// and legitimately reference .dro.md/.lego.md as domain vocabulary.
			if !strings.Contains(lower, "extension") && !strings.Contains(lower, "suffix") && !strings.Contains(lower, "infix") &&
				!strings.Contains(lower, "convention") && !strings.Contains(lower, " use .dro") && !strings.Contains(lower, " use .lego") &&
				!strings.Contains(lower, "classify by extension") && !strings.Contains(relPath, "pinakes/domains/mena-structure") {
				extLeaks++
			}
		}
	}

	if readLeaks > 0 {
		*findings = append(*findings, Finding{
			File: relPath, Severity: SevHigh, Rule: "source-path-read",
			Message: fmt.Sprintf("body contains %d Read() call(s) with rites/ source paths — use materialized channel paths (e.g. {channel}/skills/ or {channel}/commands/)", readLeaks),
		})
	}
	if refLeaks > 0 {
		*findings = append(*findings, Finding{
			File: relPath, Severity: SevMedium, Rule: "source-path-ref",
			Message: fmt.Sprintf("body contains %d reference(s) to rites/*/mena/ source paths — use materialized paths or skill names", refLeaks),
		})
	}
	if extLeaks > 0 {
		*findings = append(*findings, Finding{
			File: relPath, Severity: SevLow, Rule: "source-path-ext",
			Message: fmt.Sprintf("body contains %d reference(s) with .lego.md or .dro.md extensions — materialized files use .md", extLeaks),
		})
	}
}

// --- Helpers ---

// buildAllMenaSources constructs MenaSource entries for platform mena and all
// rite mena directories. Unlike BuildSourceChain (which builds a
// priority-ordered chain for the active rite), this function discovers ALL
// rites for source validation. Walk handles nonexistent directories
// gracefully, so sources pointing to rites without mena/ are harmless.
func buildAllMenaSources(projectRoot string) []mena.MenaSource {
	var sources []mena.MenaSource

	// Platform mena
	sources = append(sources, mena.MenaSource{
		Path: filepath.Join(projectRoot, "mena"),
	})

	// All rites (including shared)
	riteDir := filepath.Join(projectRoot, "rites")
	rites, _ := os.ReadDir(riteDir)
	for _, r := range rites {
		if r.IsDir() {
			sources = append(sources, mena.MenaSource{
				Path: filepath.Join(riteDir, r.Name(), "mena"),
			})
		}
	}

	return sources
}

// parseFrontmatterLenient tries strict YAML first, then falls back to line-by-line
// key extraction. This handles argument-hint values with brackets like
// "[--scope=rite|user|all]" which are invalid YAML.
func parseFrontmatterLenient(yamlBytes []byte) map[string]any {
	var fm map[string]any
	if err := yaml.Unmarshal(yamlBytes, &fm); err == nil {
		return fm
	}

	// Fallback: line-by-line extraction of simple key: value pairs
	fm = make(map[string]any)
	for _, line := range strings.Split(string(yamlBytes), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Skip continuation lines (indented or list items)
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if val != "" {
			fm[key] = val
		}
	}

	if len(fm) == 0 {
		return nil
	}
	return fm
}

func mustRel(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return filepath.Base(path)
	}
	return rel
}

func strVal(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}
