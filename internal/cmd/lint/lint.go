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
	"github.com/autom8y/knossos/internal/output"
)

// skillAtPattern matches @skill-name references in body content.
// Excludes email-style patterns (@user.com), @handles in examples,
// and regex patterns. Matches: @word-word at word boundaries.
var skillAtPattern = regexp.MustCompile(`(?m)(?:^|[\s(` + "`" + `])@([a-z][a-z0-9](?:[a-z0-9-]*[a-z0-9])?)(?:[#/\s` + "`" + `.,;:)\]]|$)`)

// Severity levels for lint findings.
const (
	SevCritical = "CRIT"
	SevHigh     = "HIGH"
	SevMedium   = "MED"
	SevLow      = "LOW"
)

// expectedForkState defines the deliberate fork/inline classification for each dromena.
// true = should have context: fork (self-contained, artifact output, no conversation context needed)
// false = should NOT have context: fork (interactive, contextual, or orchestrating)
var expectedForkState = map[string]bool{
	// Inline: interactive, contextual, or orchestrating commands
	"go":        false,
	"start":     false,
	"commit":    false,
	"consult":   false,
	"qa":        false,
	"one":       false,
	"task":      false,
	"build":     false,
	"architect": false,
	"hotfix":    false,
	"sprint":    false,
	// Inline: session lifecycle (need hook-injected context)
	"park":     false,
	"continue": false,
	"wrap":     false,
	"handoff":  false,
	"fray":     false,
	// Fork: self-contained CLI wrappers and one-shot actions
	"pr":          true,
	"code-review": true,
	"spike":       true,
	"minus-1":     true,
	"zero":        true,
	"rite":        true,
	"sessions":    true,
	"worktree":    true,
	"sync":        true,
	"theoria":     true,
	"ecosystem":   true,
	// Fork: rite-switching commands (one-shot CLI wrappers)
	"10x":          true,
	"debt":         true,
	"docs":         true,
	"forge":        true,
	"hygiene":      true,
	"intelligence": true,
	"rnd":          true,
	"security":     true,
	"sre":          true,
	"strategy":     true,
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
		b.WriteString(fmt.Sprintf("\n%s\n", name))
		b.WriteString(strings.Repeat("-", len(name)) + "\n")
		for _, f := range findings {
			b.WriteString(fmt.Sprintf("  [%s] %s: %s (%s)\n", f.Severity, f.File, f.Message, f.Rule))
		}
	}

	b.WriteString("Lint Report: Knossos Source Validation\n")
	b.WriteString(strings.Repeat("=", 40) + "\n")

	printSection("Agents", r.Agents)
	printSection("Dromena", r.Dromena)
	printSection("Legomena", r.Legomena)

	b.WriteString(fmt.Sprintf("\nSummary: %d issues (%d critical, %d high, %d medium, %d low) across %d files\n",
		r.Summary.Total, r.Summary.Critical, r.Summary.High, r.Summary.Medium, r.Summary.Low, r.Summary.Files))

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
- Legomena missing Triggers keyword in description

Examples:
  ari lint                    # Lint all sources
  ari lint --scope=agents     # Agents only
  ari lint --scope=dromena    # Dromena only
  ari lint --scope=legomena   # Legomena only`,
		SilenceUsage: true,
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

	report := &LintReport{}

	if scope == "" || scope == "agents" {
		lintAgents(projectRoot, report)
	}
	if scope == "" || scope == "dromena" {
		lintDromena(projectRoot, report)
		lintMenaNamespace(projectRoot, report)
	}
	if scope == "" || scope == "legomena" {
		lintLegomena(projectRoot, report)
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
	Name        string                       `yaml:"name"`
	Description string                       `yaml:"description"`
	Type        string                       `yaml:"type"`
	Tools       frontmatter.FlexibleStringSlice `yaml:"tools"`
	Model       string                       `yaml:"model"`
	Color       string                       `yaml:"color"`
	MaxTurns    int                          `yaml:"maxTurns"`
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
	// which need 20-40 turns for multi-phase coordination
	if fm.Type != "" && fm.MaxTurns > 0 {
		if expected, ok := archetypeMaxTurns[fm.Type]; ok {
			deviation := fm.MaxTurns - expected
			if deviation < 0 {
				deviation = -deviation
			}
			// Threshold: 50% of archetype default or 50, whichever is larger
			threshold := expected / 2
			if threshold < 50 {
				threshold = 50
			}
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

	// Orchestrator color check
	if fm.Type == "orchestrator" && fm.Color != "" && fm.Color != "purple" {
		report.Agents = append(report.Agents, Finding{
			File: relPath, Severity: SevMedium, Rule: "orchestrator-color",
			Message: fmt.Sprintf("orchestrator has color=%q (convention: purple)", fm.Color),
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
}

// --- Dromena linting ---

func lintDromena(projectRoot string, report *LintReport) {
	walkMena(projectRoot, ".dro.md", func(path, relPath string, data []byte) {
		report.Summary.Files++
		lintDromenFile(path, relPath, data, report)
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
	if expected, known := expectedForkState[name]; known {
		if expected && !hasFork {
			report.Dromena = append(report.Dromena, Finding{
				File: relPath, Severity: SevMedium, Rule: "context-fork-expected",
				Message: fmt.Sprintf("dromena %q should have context: fork (self-contained command)", name),
			})
		} else if !expected && hasFork {
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

	// @skill-name anti-pattern check
	checkSkillAtRefs(string(body), relPath, &report.Dromena)
}

// --- Dromena namespace collision detection ---

func lintMenaNamespace(projectRoot string, report *LintReport) {
	// Collect all dromena names to detect collisions.
	// Scope-aware: only flag collisions within the same scope (global or same rite).
	// Cross-scope shadowing (rite overrides global) is intentional.
	type nameSource struct {
		name    string
		relPath string
		scope   string // "global" or rite name
	}

	var entries []nameSource

	walkMena(projectRoot, ".dro.md", func(path, relPath string, data []byte) {
		yamlBytes, _, err := frontmatter.Parse(data)
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

func lintLegomena(projectRoot string, report *LintReport) {
	walkMena(projectRoot, ".lego.md", func(path, relPath string, data []byte) {
		report.Summary.Files++
		lintLegomenFile(path, relPath, data, report)
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
}

// --- @skill-name check ---

// skillAtExclusions are known false positives: team handles, documentation examples.
var skillAtExclusions = map[string]bool{
	"api-team":      true,
	"product-lead":  true,
	"platform-team": true,
	"skill-name":    true, // appears in anti-pattern documentation
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

// --- Helpers ---

func walkMena(projectRoot, suffix string, fn func(path, relPath string, data []byte)) {
	// Walk mena/ directory tree
	menaDir := filepath.Join(projectRoot, "mena")
	filepath.Walk(menaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, suffix) {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		fn(path, mustRel(projectRoot, path), data)
		return nil
	})

	// Also walk rites/*/mena/
	riteDir := filepath.Join(projectRoot, "rites")
	rites, _ := os.ReadDir(riteDir)
	for _, r := range rites {
		if !r.IsDir() {
			continue
		}
		riteMena := filepath.Join(riteDir, r.Name(), "mena")
		filepath.Walk(riteMena, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, suffix) {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			fn(path, mustRel(projectRoot, path), data)
			return nil
		})
	}
}

// parseFrontmatterLenient tries strict YAML first, then falls back to line-by-line
// key extraction. This handles argument-hint values with brackets like
// "[--scope=rite|user|all]" which are invalid YAML.
func parseFrontmatterLenient(yamlBytes []byte) map[string]interface{} {
	var fm map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &fm); err == nil {
		return fm
	}

	// Fallback: line-by-line extraction of simple key: value pairs
	fm = make(map[string]interface{})
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

func strVal(m map[string]interface{}, key string) string {
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
