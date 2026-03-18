package materialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	procmena "github.com/autom8y/knossos/internal/materialize/procession"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/procession"
	"github.com/autom8y/knossos/internal/provenance"
)

// projectRoot returns the knossos project root by walking up from the test file location.
func projectRoot(t *testing.T) string {
	t.Helper()
	// Walk up from internal/materialize/ to project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	root := filepath.Join(wd, "..", "..")
	// Verify the archetypes directory exists
	if _, err := os.Stat(filepath.Join(root, "knossos", "archetypes", "orchestrator.md.tpl")); err != nil {
		t.Fatalf("cannot locate project root from %s: %v", wd, err)
	}
	return root
}

func tenxDevData() OrchestratorData {
	return OrchestratorData{
		RiteName:    "10x-dev",
		Description: "Routes development work through requirements, design, implementation, and validation phases. Use when: building features or systems requires full lifecycle coordination. Triggers: coordinate, orchestrate, development workflow, feature development, implementation planning.",
		Color:       "blue",
		Skills:      []string{"orchestrator-templates", "10x-workflow"},
		PhaseRouting: `| requirements-analyst | New feature or system requested, PRD needed |
| architect | Requirements complete, architecture design needed |
| principal-engineer | Design complete, implementation needed |
| qa-adversary | Implementation complete, validation needed |
`,
		HandoffCriteria: `| requirements | - Product requirements document complete<- User stories and acceptance criteria defined<- Success metrics established< |
| design | - Architecture document with rationale<- Test-driven design (TDD) approach defined<- Technical risks identified< |
| implementation | - Code passes linting and type checking<- All unit tests pass<- Code review approval obtained< |
| validation | - Test plan complete and executed<- All tests pass<- Deployment readiness verified< |
`,
		RiteAntiPatterns: `- **Skipping design phase for MODULE complexity (always design first)**
- **Implementing without acceptance criteria defined**
- **Validating against incomplete or ambiguous requirements**`,
		CrossRiteProtocol: "<!-- TODO: Define how cross-rite concerns are routed and resolved -->",
		EntryPointSection: `## Entry Point Selection

The default workflow starts with Requirements Analyst, but certain work types benefit from alternative entry points. Select the entry agent based on work type:

| Work Type | Entry Agent | Rationale |
|-----------|-------------|-----------|
| **New feature** | requirements-analyst | Scope must be defined before design or implementation |
| **Enhancement** | requirements-analyst | Existing features need updated requirements |
| **Technical refactoring** | architect | Design-first; no new requirements, but architecture decisions needed |
| **Performance optimization** | architect | Requires analysis of bottlenecks and design tradeoffs |
| **Bug fix** | principal-engineer | Problem is known; fix and verify |
| **Security fix** | principal-engineer | Immediate remediation; design review post-implementation if needed |
| **Hotfix** | principal-engineer | Time-critical; minimal ceremony |

### Selection Criteria

1. **Does this add user-facing capability?** -> requirements-analyst
2. **Does this change system structure without adding features?** -> architect
3. **Is this fixing known broken behavior?** -> principal-engineer
4. **Is this time-critical remediation?** -> principal-engineer

### Entry Point Implications

- **requirements-analyst entry**: Full PRD -> TDD -> Code -> QA flow
- **architect entry**: TDD -> Code -> QA flow (skip PRD when requirements are implicit in technical need)
- **principal-engineer entry**: Code -> QA flow (skip PRD and TDD when scope is self-evident)

When uncertain, default to requirements-analyst. It is cheaper to skip phases than to backtrack.`,
	}
}

func forgeData() OrchestratorData {
	return OrchestratorData{
		RiteName:    "forge",
		Description: "Routes agent rite creation through design, prompts, workflow, platform integration, catalog, and validation phases. Use when: building new agent rites or expanding the agent ecosystem. Triggers: coordinate, orchestrate, forge workflow, agent creation, rite buildout.",
		Color:       "cyan",
		Skills:      []string{"orchestrator-templates", "forge-ref"},
		PhaseRouting: `| agent-designer | New agent rite concept, design phase needed |
| prompt-architect | Design complete, agent prompts needed |
| workflow-engineer | Prompts ready, workflow configuration needed |
| platform-engineer | Workflow ready, knossos integration needed |
| agent-curator | Platform integration complete, catalog update needed |
| eval-specialist | Catalog complete, evaluation and validation needed |
`,
		HandoffCriteria: `| design | - Rite specification documented<- Agent roles defined<- Workflow phases mapped< |
| prompts | - Agent prompt files created<- System instructions finalized<- Tool access configured< |
| workflow | - Workflow configuration complete<- Phase transitions defined<- Complexity levels documented< |
| platform | - Agents registered in knossos<- Integration tests passing<- ari sync validated< |
| catalog | - Knowledge base updated<- Rite documentation added<- Integration guide written< |
| validation | - Evaluation report complete<- Rite readiness confirmed<- Production deployment approved< |
`,
		RiteAntiPatterns: `- **Creating agents without workflow context (agents must fit rite lifecycle)**
- **Skipping prompt validation (prompts must be tested before deployment)**
- **Agent proliferation (consolidate similar roles, avoid agent sprawl)**`,
		CrossRiteProtocol: `Notify ecosystem of knossos changes affecting sync/knossos. Coordinate with target rite on agent specifications.

When routing cross-rite concerns:
1. Identify the affected rite(s)
2. Include current session context in handoff
3. Notify user of cross-rite escalation
4. Track resolution in throughline`,
	}
}

func slopChopData() OrchestratorData {
	return OrchestratorData{
		RiteName:    "slop-chop",
		Description: "Coordinates slop-chop AI code quality gate phases. Routes work through detection,\nanalysis, decay, remediation, and verdict phases. Use when: reviewing AI-assisted\ncode for hallucinations, logic errors, temporal debt, and other AI-specific pathologies.\nTriggers: coordinate, orchestrate, slop-chop workflow, AI code review, quality gate.",
		Color:       "red",
		Skills:      []string{"orchestrator-templates", "slop-chop-ref"},
		ContractMustNot: []string{
			"Execute analysis or detection work directly",
			"Use tools beyond Read",
			"Respond with prose instead of CONSULTATION_RESPONSE format",
		},
		ExousiaYouDecide: `
- Phase sequencing and complexity gating (which phases run)
- Which specialist handles the current phase
- When handoff criteria are met to advance
- Whether to pause pending clarification`,
		ExousiaYouEscalate: `
- Conflicting findings between specialists
- Scope changes mid-analysis (DIFF needs MODULE-level review)
- Configuration conflicts in ` + "`.slop-chop.yaml`" + ` overrides`,
		ExousiaYouDoNotDecide: `
- Detection methodology (hallucination-hunter)
- Individual finding severity (each specialist owns their domain)
- Pass/fail verdict (gate-keeper)
- Fix implementations (remedy-smith)
- Temporal staleness classification (cruft-cutter)`,
		PhaseRouting: `<!-- TODO: Define which specialist handles which phase and routing conditions -->
`,
		HandoffCriteria:   "",
		RiteAntiPatterns:  "",
		CrossRiteProtocol: "<!-- TODO: Define how cross-rite concerns are routed and resolved -->",
		CustomSections: `## Phase Routing and Complexity Gating

| Specialist | Route When | Complexity |
|------------|------------|------------|
| hallucination-hunter | Entry: code review needed | ALL |
| logic-surgeon | Detection complete | ALL |
| cruft-cutter | Analysis complete, temporal scan needed | MODULE+ |
| remedy-smith | Temporal scan complete, remediation needed | MODULE+ |
| gate-keeper | All analysis complete, verdict needed | ALL |

**DIFF** (3 phases): detection --> analysis --> verdict. Skip cruft-cutter and remedy-smith.
**MODULE / CODEBASE** (5 phases): detection --> analysis --> decay --> remediation --> verdict.

### Artifact Chain

Each specialist receives ALL prior artifacts. Include paths in every specialist prompt:
- logic-surgeon: [detection-report]
- cruft-cutter: [detection-report, analysis-report]
- remedy-smith: [detection-report, analysis-report, decay-report]
- gate-keeper: ALL prior artifacts (varies by complexity)

### Handoff Criteria

| Phase | Advance When |
|-------|-------------|
| detection | Import/registry verification complete for all in-scope files; severity ratings assigned |
| analysis | Logic + test quality assessed; bloat scan complete; unreviewed-output signals documented |
| decay | Temporal debt scan complete; comment artifacts classified; staleness scores assigned |
| remediation | Every finding has remedy or explicit waiver; auto-fixes validated; safe/unsafe justified |
| verdict | Verdict issued with evidence; CI output generated; cross-rite referrals documented |`,
	}
}

func TestRenderArchetype_TemplateParsesWithoutError(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	tplPath := filepath.Join(root, "knossos", "archetypes", "orchestrator.md.tpl")
	content, err := os.ReadFile(tplPath)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}

	// Verify the template parses successfully
	_, err = RenderArchetypeFromString(string(content), "orchestrator.md.tpl", tenxDevData())
	if err != nil {
		t.Fatalf("template failed to parse/render: %v", err)
	}
}

func TestRenderArchetype_10xDev(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	result, err := RenderArchetype(root, "orchestrator.md.tpl", tenxDevData())
	if err != nil {
		t.Fatalf("RenderArchetype() error = %v", err)
	}

	output := string(result)

	// Verify frontmatter structure
	sections := []struct {
		name    string
		content string
	}{
		{"frontmatter start", "---\nname: potnia"},
		{"description", "Routes development work through requirements"},
		{"color", "color: blue"},
		{"skills orchestrator-templates", "- orchestrator-templates"},
		{"skills 10x-workflow", "- 10x-workflow"},
		{"disallowedTools", "disallowedTools:"},
		{"contract", "contract:"},
		{"frontmatter end", "---\n\n# Potnia"},

		// Body sections in order
		{"opening paragraph", "consultative throughline** for 10x-dev work"},
		{"consultation role", "## Consultation Role (CRITICAL)"},
		{"what you do", "### What You DO"},
		{"what you do not do", "### What You DO NOT DO"},
		{"litmus test", "### The Litmus Test"},
		{"tool access", "## Tool Access"},
		{"consultation protocol", "## Consultation Protocol"},
		{"input", "### Input: CONSULTATION_REQUEST"},
		{"output", "### Output: CONSULTATION_RESPONSE"},
		{"position in workflow", "## Position in Workflow"},
		{"exousia", "## Exousia"},
		{"you decide", "### You Decide"},
		{"you escalate", "### You Escalate"},
		{"you do not decide", "### You Do NOT Decide"},
		{"phase routing", "## Phase Routing"},
		{"phase routing content", "| requirements-analyst | New feature or system requested"},
		{"behavioral constraints", "## Behavioral Constraints"},
		{"handling failures", "## Handling Failures"},
		{"acid test", "## The Acid Test"},
		{"cross-rite protocol", "## Cross-Rite Protocol"},
		{"skills reference", "## Skills Reference"},
		{"anti-patterns", "## Anti-Patterns"},
		{"core responsibilities", "## Core Responsibilities"},
		{"entry point section", "## Entry Point Selection"},
		{"entry point table", "| **New feature** | requirements-analyst |"},
		{"behavioral constraints DO NOT", "## Behavioral Constraints (DO NOT)"},
		{"handoff criteria", "## Handoff Criteria"},
		{"handoff content", "| requirements | - Product requirements document complete"},
		{"anti-patterns to avoid", "## Anti-Patterns to Avoid"},
		{"rite-specific anti-patterns", "### Rite-Specific Anti-Patterns"},
		{"rite-specific content", "Skipping design phase for MODULE complexity"},
	}

	for _, tc := range sections {
		if !strings.Contains(output, tc.content) {
			t.Errorf("missing section %q: expected content %q not found in output", tc.name, tc.content)
		}
	}

	// Verify section ordering (each section must appear after the previous)
	orderedMarkers := []string{
		"# Potnia",
		"## Consultation Role (CRITICAL)",
		"## Tool Access",
		"## Consultation Protocol",
		"## Position in Workflow",
		"## Exousia",
		"## Phase Routing",
		"## Behavioral Constraints\n",
		"## Handling Failures",
		"## The Acid Test",
		"## Cross-Rite Protocol",
		"## Skills Reference",
		"## Anti-Patterns\n",
		"## Core Responsibilities",
		"## Entry Point Selection",
		"## Behavioral Constraints (DO NOT)",
		"## Handoff Criteria",
		"## Anti-Patterns to Avoid",
		"### Rite-Specific Anti-Patterns",
	}

	lastIdx := -1
	for _, marker := range orderedMarkers {
		idx := strings.Index(output, marker)
		if idx == -1 {
			t.Errorf("section marker %q not found in output", marker)
			continue
		}
		if idx <= lastIdx {
			t.Errorf("section %q (at %d) appears before or at previous section (at %d) — wrong order", marker, idx, lastIdx)
		}
		lastIdx = idx
	}
}

func TestRenderArchetype_Forge(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	result, err := RenderArchetype(root, "orchestrator.md.tpl", forgeData())
	if err != nil {
		t.Fatalf("RenderArchetype() error = %v", err)
	}

	output := string(result)

	checks := []struct {
		name    string
		content string
	}{
		{"rite name", "consultative throughline** for forge work"},
		{"color", "color: cyan"},
		{"skills forge-ref", "- forge-ref"},
		{"phase routing agent-designer", "| agent-designer | New agent rite concept"},
		{"phase routing eval-specialist", "| eval-specialist | Catalog complete"},
		{"handoff criteria design", "| design | - Rite specification documented"},
		{"handoff criteria validation", "| validation | - Evaluation report complete"},
		{"cross-rite protocol", "Notify ecosystem of knossos changes"},
		{"rite-specific anti-patterns", "Creating agents without workflow context"},
		{"no entry point section", "## Behavioral Constraints (DO NOT)"},
	}

	for _, tc := range checks {
		if !strings.Contains(output, tc.content) {
			t.Errorf("forge: missing %q: expected %q not found", tc.name, tc.content)
		}
	}

	// Forge should NOT have an Entry Point Selection section
	if strings.Contains(output, "## Entry Point Selection") {
		t.Error("forge should not contain Entry Point Selection section")
	}
}

func TestRenderArchetype_SlopChop(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	result, err := RenderArchetype(root, "orchestrator.md.tpl", slopChopData())
	if err != nil {
		t.Fatalf("RenderArchetype() error = %v", err)
	}

	output := string(result)

	checks := []struct {
		name    string
		content string
	}{
		{"rite name", "consultative throughline** for slop-chop"},
		{"color", "color: red"},
		{"custom contract", "Execute analysis or detection work directly"},
		{"custom exousia decide", "Phase sequencing and complexity gating"},
		{"custom exousia escalate", "Conflicting findings between specialists"},
		{"custom exousia do not decide", "Detection methodology (hallucination-hunter)"},
		{"custom sections", "## Phase Routing and Complexity Gating"},
		{"artifact chain", "### Artifact Chain"},
		{"complexity gating", "**DIFF** (3 phases): detection --> analysis --> verdict"},
	}

	for _, tc := range checks {
		if !strings.Contains(output, tc.content) {
			t.Errorf("slop-chop: missing %q: expected %q not found", tc.name, tc.content)
		}
	}
}

func TestRenderArchetype_DefaultContractMustNot(t *testing.T) {
	t.Parallel()
	// When ContractMustNot is nil/empty, the template should use defaults
	root := projectRoot(t)
	data := tenxDevData()
	data.ContractMustNot = nil

	result, err := RenderArchetype(root, "orchestrator.md.tpl", data)
	if err != nil {
		t.Fatalf("RenderArchetype() error = %v", err)
	}

	output := string(result)
	for _, expected := range defaultContractMustNot() {
		if !strings.Contains(output, expected) {
			t.Errorf("default contract.must_not entry %q not found in output", expected)
		}
	}
}

func TestRenderArchetype_MissingTemplate(t *testing.T) {
	t.Parallel()
	// Use a template name that doesn't exist anywhere (not just a bad projectRoot).
	_, err := RenderArchetype("/nonexistent", "nonexistent-archetype.md.tpl", tenxDevData())
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRenderArchetypeFromString_InvalidTemplate(t *testing.T) {
	t.Parallel()
	_, err := RenderArchetypeFromString("{{.Broken", "bad.tpl", tenxDevData())
	if err == nil {
		t.Fatal("expected error for invalid template syntax, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRenderArchetype_FrontmatterDelimiters(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	result, err := RenderArchetype(root, "orchestrator.md.tpl", tenxDevData())
	if err != nil {
		t.Fatalf("RenderArchetype() error = %v", err)
	}

	output := string(result)

	// Must start with ---
	if !strings.HasPrefix(output, "---\n") {
		t.Error("output must start with frontmatter delimiter ---")
	}

	// Must have exactly 2 --- delimiters (opening and closing frontmatter)
	count := strings.Count(output, "\n---\n")
	// The opening --- is at position 0 so it's "---\n" not "\n---\n"
	// Count closing delimiter
	if !strings.Contains(output, "\n---\n\n# Potnia") {
		t.Error("frontmatter closing delimiter must be followed by # Potnia heading")
	}

	// Verify no triple --- appears in the body (only in frontmatter)
	body := output[strings.Index(output[4:], "\n---\n")+4+5:]
	if strings.Contains(body, "\n---\n") {
		t.Error("unexpected --- delimiter found in body content")
	}
	_ = count // used for context, assertion is above
}

func TestRenderArchetype_SkillsYAMLList(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	result, err := RenderArchetype(root, "orchestrator.md.tpl", tenxDevData())
	if err != nil {
		t.Fatalf("RenderArchetype() error = %v", err)
	}

	output := string(result)

	// Skills must render as a YAML list with proper indentation
	if !strings.Contains(output, "skills:\n  - orchestrator-templates\n  - 10x-workflow\n") {
		// Extract the skills section for debugging
		start := strings.Index(output, "skills:")
		end := min(start+100, len(output))
		t.Errorf("skills not rendered as expected YAML list.\nGot:\n%s", output[start:end])
	}
}

// --- Integration tests: archetype wiring in materializeAgents ---

// setupArchetypeRite creates a minimal rite directory with an archetype agent and
// a normal source agent. Returns (projectDir, channelDir).
func setupArchetypeRite(t *testing.T) (string, string) {
	t.Helper()
	root := projectRoot(t) // real project root for archetype templates

	projectDir := t.TempDir()
	channelDir := filepath.Join(projectDir, ".claude")

	ritesDir := filepath.Join(projectDir, ".knossos", "rites", "test-arch")
	if err := os.MkdirAll(filepath.Join(ritesDir, "agents"), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Symlink knossos/archetypes into the temp project so RenderArchetype finds templates
	knossosDir := filepath.Join(projectDir, "knossos")
	if err := os.MkdirAll(knossosDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Symlink(
		filepath.Join(root, "knossos", "archetypes"),
		filepath.Join(knossosDir, "archetypes"),
	); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	// Write a source agent file for the non-archetype agent
	agentContent := "---\nname: engineer\ndescription: Implements code\ntools: Bash, Read\n---\n\n# Engineer\n\nBody.\n"
	if err := os.WriteFile(filepath.Join(ritesDir, "agents", "engineer.md"), []byte(agentContent), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Also write a source file for potnia (should be SKIPPED because archetype takes precedence)
	potniaSource := "# This should not appear — archetype rendering takes precedence\n"
	if err := os.WriteFile(filepath.Join(ritesDir, "agents", "potnia.md"), []byte(potniaSource), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	return projectDir, channelDir
}

func TestMaterializeAgents_ArchetypeRendersFromTemplate(t *testing.T) {
	t.Parallel()
	projectDir, channelDir := setupArchetypeRite(t)

	manifest := &RiteManifest{
		Name:        "test-arch",
		Description: "Test rite for archetype wiring",
		EntryAgent:  "potnia",
		Agents: []Agent{
			{Name: "potnia", Role: "Coordinates workflow", Archetype: "orchestrator"},
			{Name: "engineer", Role: "Implements code"},
		},
		ArchetypeData: map[string]map[string]any{
			"orchestrator": {
				"description":         "Coordinates test phases",
				"color":               "green",
				"skills":              []any{"orchestrator-templates"},
				"phase_routing":       "| engineer | Implementation needed |\n",
				"handoff_criteria":    "| impl | - Code complete |\n",
				"rite_anti_patterns":  "- **Test anti-pattern**",
				"cross_rite_protocol": "<!-- TODO -->",
			},
		},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		Name:         "test-arch",
		RitePath:     filepath.Join(projectDir, ".knossos", "rites", "test-arch"),
		ManifestPath: filepath.Join(projectDir, ".knossos", "rites", "test-arch", "manifest.yaml"),
		Source:       RiteSource{Type: SourceProject, Path: filepath.Join(projectDir, ".knossos", "rites", "test-arch")},
	}

	err := m.materializeAgents(manifest, resolved.RitePath, channelDir, resolved, provenance.NullCollector{}, nil, nil, nil, "", "", nil)
	if err != nil {
		t.Fatalf("materializeAgents() error = %v", err)
	}

	// Verify potnia was rendered from archetype template
	potniaPath := filepath.Join(channelDir, "agents", "potnia.md")
	potniaContent, err := os.ReadFile(potniaPath)
	if err != nil {
		t.Fatalf("expected potnia agent at %s: %v", potniaPath, err)
	}

	output := string(potniaContent)

	// Must contain archetype-rendered content, NOT the source file content
	if strings.Contains(output, "This should not appear") {
		t.Error("potnia should be rendered from archetype, not copied from source file")
	}

	// Must contain template-rendered content
	checks := []struct {
		name    string
		content string
	}{
		{"rite name in body", "consultative throughline** for test-arch"},
		{"color", "green"},
		{"phase routing", "| engineer | Implementation needed |"},
		{"handoff criteria", "| impl | - Code complete |"},
		{"anti-patterns", "Test anti-pattern"},
		{"heading", "# Potnia"},
	}
	for _, tc := range checks {
		if !strings.Contains(output, tc.content) {
			t.Errorf("archetype potnia missing %q: expected %q", tc.name, tc.content)
		}
	}
}

func TestMaterializeAgents_NonArchetypeAgentCopiedFromSource(t *testing.T) {
	t.Parallel()
	projectDir, channelDir := setupArchetypeRite(t)

	manifest := &RiteManifest{
		Name:        "test-arch",
		Description: "Test rite",
		EntryAgent:  "potnia",
		Agents: []Agent{
			{Name: "potnia", Role: "Coordinates", Archetype: "orchestrator"},
			{Name: "engineer", Role: "Implements code"},
		},
		ArchetypeData: map[string]map[string]any{
			"orchestrator": {
				"description":         "Coordinates test phases",
				"color":               "green",
				"skills":              []any{"orchestrator-templates"},
				"phase_routing":       "| engineer | Impl needed |\n",
				"handoff_criteria":    "| impl | - Done |\n",
				"rite_anti_patterns":  "- None",
				"cross_rite_protocol": "<!-- TODO -->",
			},
		},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		Name:         "test-arch",
		RitePath:     filepath.Join(projectDir, ".knossos", "rites", "test-arch"),
		ManifestPath: filepath.Join(projectDir, ".knossos", "rites", "test-arch", "manifest.yaml"),
		Source:       RiteSource{Type: SourceProject, Path: filepath.Join(projectDir, ".knossos", "rites", "test-arch")},
	}

	err := m.materializeAgents(manifest, resolved.RitePath, channelDir, resolved, provenance.NullCollector{}, nil, nil, nil, "", "", nil)
	if err != nil {
		t.Fatalf("materializeAgents() error = %v", err)
	}

	// Verify engineer was copied from source (not archetype)
	engPath := filepath.Join(channelDir, "agents", "engineer.md")
	engContent, err := os.ReadFile(engPath)
	if err != nil {
		t.Fatalf("expected engineer agent at %s: %v", engPath, err)
	}

	output := string(engContent)

	// Must contain source file content
	if !strings.Contains(output, "# Engineer") {
		t.Errorf("engineer should contain source heading:\n%s", output)
	}
	if !strings.Contains(output, "Body.") {
		t.Errorf("engineer should contain source body:\n%s", output)
	}
}

func TestMaterializeAgents_ArchetypeGoesThruTransformPipeline(t *testing.T) {
	t.Parallel()
	projectDir, channelDir := setupArchetypeRite(t)

	manifest := &RiteManifest{
		Name:        "test-arch",
		Description: "Test rite",
		EntryAgent:  "potnia",
		Agents: []Agent{
			{Name: "potnia", Role: "Coordinates", Archetype: "orchestrator"},
		},
		ArchetypeData: map[string]map[string]any{
			"orchestrator": {
				"description":         "Coordinates phases",
				"color":               "purple",
				"skills":              []any{"orchestrator-templates"},
				"phase_routing":       "| eng | Impl |\n",
				"handoff_criteria":    "| impl | - Done |\n",
				"rite_anti_patterns":  "- None",
				"cross_rite_protocol": "<!-- TODO -->",
			},
		},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		Name:         "test-arch",
		RitePath:     filepath.Join(projectDir, ".knossos", "rites", "test-arch"),
		ManifestPath: filepath.Join(projectDir, ".knossos", "rites", "test-arch", "manifest.yaml"),
		Source:       RiteSource{Type: SourceProject, Path: filepath.Join(projectDir, ".knossos", "rites", "test-arch")},
	}

	err := m.materializeAgents(manifest, resolved.RitePath, channelDir, resolved, provenance.NullCollector{}, nil, nil, nil, "", "", nil)
	if err != nil {
		t.Fatalf("materializeAgents() error = %v", err)
	}

	potniaPath := filepath.Join(channelDir, "agents", "potnia.md")
	potniaContent, err := os.ReadFile(potniaPath)
	if err != nil {
		t.Fatalf("expected potnia at %s: %v", potniaPath, err)
	}

	output := string(potniaContent)

	// The archetype template outputs type: orchestrator in frontmatter.
	// transformAgentContent strips knossos-only fields including "type".
	if strings.Contains(output, "\ntype:") {
		t.Error("transform pipeline should strip 'type' from archetype output")
	}

	// Name should be injected by transform pipeline
	if !strings.Contains(output, "name: potnia") {
		t.Errorf("transform pipeline should inject name:\n%s", output)
	}
}

func TestMaterializeAgents_NoArchetypeNoChange(t *testing.T) {
	t.Parallel()
	// When no agents have archetype set, behavior is identical to before.
	projectDir := t.TempDir()
	channelDir := filepath.Join(projectDir, ".claude")

	ritesDir := filepath.Join(projectDir, ".knossos", "rites", "plain")
	if err := os.MkdirAll(filepath.Join(ritesDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	agentContent := "---\nname: worker\ndescription: Works\n---\n\n# Worker\n"
	if err := os.WriteFile(filepath.Join(ritesDir, "agents", "worker.md"), []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	manifest := &RiteManifest{
		Name:       "plain",
		EntryAgent: "worker",
		Agents:     []Agent{{Name: "worker", Role: "Works"}},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		Name:         "plain",
		RitePath:     filepath.Join(projectDir, ".knossos", "rites", "plain"),
		ManifestPath: filepath.Join(projectDir, ".knossos", "rites", "plain", "manifest.yaml"),
		Source:       RiteSource{Type: SourceProject, Path: filepath.Join(projectDir, ".knossos", "rites", "plain")},
	}

	err := m.materializeAgents(manifest, resolved.RitePath, channelDir, resolved, provenance.NullCollector{}, nil, nil, nil, "", "", nil)
	if err != nil {
		t.Fatalf("materializeAgents() error = %v", err)
	}

	// Verify worker was copied from source
	workerPath := filepath.Join(channelDir, "agents", "worker.md")
	data, err := os.ReadFile(workerPath)
	if err != nil {
		t.Fatalf("expected worker at %s: %v", workerPath, err)
	}

	if !strings.Contains(string(data), "# Worker") {
		t.Error("worker content should come from source file")
	}
}

func TestMaterializeAgents_UnknownArchetypeErrors(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	channelDir := filepath.Join(projectDir, ".claude")

	ritesDir := filepath.Join(projectDir, ".knossos", "rites", "bad")
	if err := os.MkdirAll(filepath.Join(ritesDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	manifest := &RiteManifest{
		Name:       "bad",
		EntryAgent: "test",
		Agents:     []Agent{{Name: "test", Role: "Tests", Archetype: "nonexistent"}},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		Name:         "bad",
		RitePath:     filepath.Join(projectDir, ".knossos", "rites", "bad"),
		ManifestPath: filepath.Join(projectDir, ".knossos", "rites", "bad", "manifest.yaml"),
		Source:       RiteSource{Type: SourceProject, Path: filepath.Join(projectDir, ".knossos", "rites", "bad")},
	}

	err := m.materializeAgents(manifest, resolved.RitePath, channelDir, resolved, provenance.NullCollector{}, nil, nil, nil, "", "", nil)
	if err == nil {
		t.Fatal("expected error for unknown archetype, got nil")
	}
	if !strings.Contains(err.Error(), "unknown archetype") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMaterializeAgents_ArchetypeProvenanceRecorded(t *testing.T) {
	t.Parallel()
	projectDir, channelDir := setupArchetypeRite(t)

	manifest := &RiteManifest{
		Name:        "test-arch",
		Description: "Test rite",
		EntryAgent:  "potnia",
		Agents: []Agent{
			{Name: "potnia", Role: "Coordinates", Archetype: "orchestrator"},
			{Name: "engineer", Role: "Implements"},
		},
		ArchetypeData: map[string]map[string]any{
			"orchestrator": {
				"description":         "Coordinates",
				"color":               "blue",
				"skills":              []any{"orchestrator-templates"},
				"phase_routing":       "| eng | Impl |\n",
				"handoff_criteria":    "| impl | - Done |\n",
				"rite_anti_patterns":  "- None",
				"cross_rite_protocol": "<!-- TODO -->",
			},
		},
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		Name:         "test-arch",
		RitePath:     filepath.Join(projectDir, ".knossos", "rites", "test-arch"),
		ManifestPath: filepath.Join(projectDir, ".knossos", "rites", "test-arch", "manifest.yaml"),
		Source:       RiteSource{Type: SourceProject, Path: filepath.Join(projectDir, ".knossos", "rites", "test-arch")},
	}

	// Use a real collector to capture provenance records
	collector := provenance.NewCollector()

	err := m.materializeAgents(manifest, resolved.RitePath, channelDir, resolved, collector, nil, nil, nil, "", "", nil)
	if err != nil {
		t.Fatalf("materializeAgents() error = %v", err)
	}

	// Check that archetype agent has provenance with "archetype" source type
	entries := collector.Entries()
	potniaEntry, ok := entries["agents/potnia.md"]
	if !ok {
		t.Fatal("missing provenance entry for agents/potnia.md")
	}
	if potniaEntry.SourceType != "archetype" {
		t.Errorf("potnia provenance source_type = %q, want %q", potniaEntry.SourceType, "archetype")
	}
	if potniaEntry.SourcePath != "knossos/archetypes/orchestrator.md.tpl" {
		t.Errorf("potnia provenance source_path = %q, want %q", potniaEntry.SourcePath, "knossos/archetypes/orchestrator.md.tpl")
	}

	// Check that source agent has provenance with "project" source type
	engEntry, ok := entries["agents/engineer.md"]
	if !ok {
		t.Fatal("missing provenance entry for agents/engineer.md")
	}
	if engEntry.SourceType != "project" {
		t.Errorf("engineer provenance source_type = %q, want %q", engEntry.SourceType, "project")
	}
}

func TestBuildOrchestratorData_ExtractsAllFields(t *testing.T) {
	t.Parallel()
	manifest := &RiteManifest{
		Name:        "test-rite",
		Description: "Test description",
		ArchetypeData: map[string]map[string]any{
			"orchestrator": {
				"description":                  "Custom orchestrator desc",
				"color":                        "red",
				"skills":                       []any{"skill-a", "skill-b"},
				"contract_must_not":            []any{"Don't do X", "Don't do Y"},
				"phase_routing":                "| agent | route |\n",
				"handoff_criteria":             "| phase | criteria |\n",
				"rite_anti_patterns":           "- Pattern A",
				"cross_rite_protocol":          "Protocol text",
				"entry_point_section":          "## Entry\nContent",
				"custom_sections":              "## Custom\nContent",
				"exousia_you_decide":           "- Decide this",
				"exousia_you_escalate":         "- Escalate this",
				"exousia_you_do_not_decide":    "- Not this",
				"tool_access_section":          "Custom tools",
				"consultation_protocol_input":  "Custom input",
				"consultation_protocol_output": "Custom output",
				"position_in_workflow":         "Custom position",
				"core_responsibilities":        "- Extra responsibility",
				"skills_reference":             "Custom refs",
				"behavioral_constraints_do":    "Custom constraints",
			},
		},
	}

	agent := Agent{Name: "potnia", Role: "Coordinates", Archetype: "orchestrator"}
	data := buildOrchestratorData(agent, manifest)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"RiteName", data.RiteName, "test-rite"},
		{"Description", data.Description, "Custom orchestrator desc"},
		{"Color", data.Color, "red"},
		{"PhaseRouting", data.PhaseRouting, "| agent | route |\n"},
		{"HandoffCriteria", data.HandoffCriteria, "| phase | criteria |\n"},
		{"RiteAntiPatterns", data.RiteAntiPatterns, "- Pattern A"},
		{"CrossRiteProtocol", data.CrossRiteProtocol, "Protocol text"},
		{"EntryPointSection", data.EntryPointSection, "## Entry\nContent"},
		{"CustomSections", data.CustomSections, "## Custom\nContent"},
		{"ExousiaYouDecide", data.ExousiaYouDecide, "- Decide this"},
		{"ExousiaYouEscalate", data.ExousiaYouEscalate, "- Escalate this"},
		{"ExousiaYouDoNotDecide", data.ExousiaYouDoNotDecide, "- Not this"},
		{"ToolAccessSection", data.ToolAccessSection, "Custom tools"},
		{"ConsultationProtocolInput", data.ConsultationProtocolInput, "Custom input"},
		{"ConsultationProtocolOutput", data.ConsultationProtocolOutput, "Custom output"},
		{"PositionInWorkflow", data.PositionInWorkflow, "Custom position"},
		{"CoreResponsibilities", data.CoreResponsibilities, "- Extra responsibility"},
		{"SkillsReference", data.SkillsReference, "Custom refs"},
		{"BehavioralConstraintsDO", data.BehavioralConstraintsDO, "Custom constraints"},
	}

	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("buildOrchestratorData().%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}

	// Verify slice fields
	if len(data.Skills) != 2 || data.Skills[0] != "skill-a" || data.Skills[1] != "skill-b" {
		t.Errorf("Skills = %v, want [skill-a, skill-b]", data.Skills)
	}
	if len(data.ContractMustNot) != 2 || data.ContractMustNot[0] != "Don't do X" {
		t.Errorf("ContractMustNot = %v, want [Don't do X, Don't do Y]", data.ContractMustNot)
	}
}

func TestBuildOrchestratorData_MissingArchetypeData(t *testing.T) {
	t.Parallel()
	manifest := &RiteManifest{
		Name:        "empty-rite",
		Description: "No archetype data",
	}
	agent := Agent{Name: "potnia", Archetype: "orchestrator"}
	data := buildOrchestratorData(agent, manifest)

	if data.RiteName != "empty-rite" {
		t.Errorf("RiteName = %q, want %q", data.RiteName, "empty-rite")
	}
	if data.Description != "" {
		t.Errorf("Description should be empty when no archetype_data: %q", data.Description)
	}
	if data.Color != "" {
		t.Errorf("Color should be empty when no archetype_data: %q", data.Color)
	}
}

func TestMaterializeAgents_ArchetypeRendersFromConfigFile(t *testing.T) {
	t.Parallel()
	projectDir, channelDir := setupArchetypeRite(t)

	// Write an orchestrator.yaml to the test rite directory (simulating a real rite).
	// No ArchetypeData in the manifest — config file is the sole data source.
	ritesDir := filepath.Join(projectDir, ".knossos", "rites", "test-arch")
	orchestratorYAML := `
rite:
  name: test-arch
  domain: testing
  color: orange
frontmatter:
  description: "Coordinates config-driven test phases"
routing:
  engineer: "Implementation needed"
handoff_criteria:
  engineer:
    - "Code complete"
    - "Tests pass"
antipatterns:
  - "Skipping tests"
cross_rite_protocol: "Route to 10x-dev for implementation"
workflow_position:
  upstream: "User request"
  downstream: "10x-dev"
`
	if err := os.WriteFile(filepath.Join(ritesDir, "orchestrator.yaml"), []byte(orchestratorYAML), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	manifest := &RiteManifest{
		Name:        "test-arch",
		Description: "Test rite for config-driven archetype",
		EntryAgent:  "potnia",
		Agents: []Agent{
			{Name: "potnia", Role: "Coordinates workflow", Archetype: "orchestrator"},
			{Name: "engineer", Role: "Implements code"},
		},
		// ArchetypeData intentionally nil — real-world scenario
	}

	resolver := paths.NewResolver(projectDir)
	m := NewMaterializer(resolver)

	resolved := &ResolvedRite{
		Name:         "test-arch",
		RitePath:     ritesDir,
		ManifestPath: filepath.Join(ritesDir, "manifest.yaml"),
		Source:       RiteSource{Type: SourceProject, Path: ritesDir},
	}

	err := m.materializeAgents(manifest, resolved.RitePath, channelDir, resolved, provenance.NullCollector{}, nil, nil, nil, "", "", nil)
	if err != nil {
		t.Fatalf("materializeAgents() error = %v", err)
	}

	// Verify potnia was rendered from archetype template WITH config file data
	potniaPath := filepath.Join(channelDir, "agents", "potnia.md")
	potniaContent, err := os.ReadFile(potniaPath)
	if err != nil {
		t.Fatalf("expected potnia at %s: %v", potniaPath, err)
	}

	output := string(potniaContent)

	// Must NOT contain the source file content (archetype still wins)
	if strings.Contains(output, "This should not appear") {
		t.Error("potnia should be rendered from archetype, not copied from source file")
	}

	// Must contain config-file-derived content
	checks := []struct {
		name    string
		content string
	}{
		{"rite name in body", "consultative throughline** for test-arch"},
		{"color from config", "orange"},
		{"phase routing from config", "| engineer | Implementation needed |"},
		{"handoff criteria from config", "| engineer | - Code complete<- Tests pass< |"},
		{"antipattern from config", "Skipping tests"},
		{"cross-rite protocol from config", "Route to 10x-dev for implementation"},
		{"position from config", "**Upstream**: User request"},
		{"heading", "# Potnia"},
	}
	for _, tc := range checks {
		if !strings.Contains(output, tc.content) {
			t.Errorf("config-driven potnia missing %q: expected %q", tc.name, tc.content)
		}
	}

	// Transform pipeline should still strip knossos-only fields
	if strings.Contains(output, "\ntype:") {
		t.Error("transform pipeline should strip 'type' from archetype output")
	}
	if !strings.Contains(output, "name: potnia") {
		t.Error("transform pipeline should inject name")
	}
}

func TestRenderArchetypeAgent_UnknownArchetype(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	agent := Agent{Name: "test", Archetype: "unknown-type"}
	manifest := &RiteManifest{Name: "test"}

	_, err := renderArchetypeAgent(root, agent, manifest, RenderArchetype)
	if err == nil {
		t.Fatal("expected error for unknown archetype")
	}
	if !strings.Contains(err.Error(), "unknown archetype: unknown-type") {
		t.Errorf("unexpected error: %v", err)
	}
}

// securityRemediationTemplate returns a procession template matching
// the real processions/security-remediation.yaml for rendering tests.
func securityRemediationTemplate() *procession.Template {
	return &procession.Template{
		Name:        "security-remediation",
		Description: "Security findings lifecycle: audit, assess, plan, remediate, validate",
		ArtifactDir: ".sos/wip/security-remediation/",
		Stations: []procession.Station{
			{Name: "audit", Rite: "security", Goal: "Discover and catalog security findings", Produces: []string{"SECURITY-AUDIT.md"}},
			{Name: "assess", Rite: "debt-triage", Goal: "Score and prioritize findings", Produces: []string{"RISK-MATRIX.md"}},
			{Name: "plan", Rite: "debt-triage", Goal: "Create remediation roadmap", Produces: []string{"SPRINT-PACKAGE.md"}},
			{Name: "remediate", Rite: "10x-dev", Goal: "Implement fixes", Produces: []string{"CHANGELOG.md"}, AltRite: "hygiene"},
			{Name: "validate", Rite: "security", Goal: "Verify fixes", Produces: []string{"VALIDATION-REPORT.md"}},
		},
	}
}

func TestRenderArchetype_ProcessionWorkflow(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	data := procmena.BuildWorkflowData(securityRemediationTemplate())

	result, err := RenderArchetype(root, "procession-workflow.md.tpl", data)
	if err != nil {
		t.Fatalf("RenderArchetype(procession-workflow) error = %v", err)
	}

	output := string(result)

	// Verify frontmatter
	checks := []string{
		"name: security-remediation",
		"allowed-tools:",
		"model: opus",
		// Body sections
		"# /security-remediation -- Cross-Rite Procession Workflow",
		"## Station Map",
		"| 1 | audit | security |",
		"| 5 | validate | security |",
		"## Pre-computed Context",
		"## Behavior",
		"### Phase 0 -- Pre-Flight",
		"Case A: No session active",
		"Case B: Active procession of type `security-remediation`",
		"Case C: Active procession of DIFFERENT type",
		"Case D: Session active but no procession",
		"### Execution",
		`Skill("security-remediation-ref")`,
		`Skill("procession-ref")`,
		"## Anti-Patterns",
	}

	for _, want := range checks {
		if !strings.Contains(output, want) {
			t.Errorf("procession-workflow output missing: %q", want)
		}
	}
}

func TestRenderArchetype_ProcessionRef(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	data := procmena.BuildWorkflowData(securityRemediationTemplate())

	result, err := RenderArchetype(root, "procession-ref.md.tpl", data)
	if err != nil {
		t.Fatalf("RenderArchetype(procession-ref) error = %v", err)
	}

	output := string(result)

	checks := []string{
		"name: security-remediation-ref",
		"# security-remediation Procession Reference",
		"## Station Map",
		"| 1 | audit | security |",
		"## Station Goals",
		"**audit** (security)",
		"**remediate** (10x-dev, alt: hygiene)",
		"## Workflow",
		"Total stations**: 5",
		"## CLI Commands",
		"ari procession proceed",
	}

	for _, want := range checks {
		if !strings.Contains(output, want) {
			t.Errorf("procession-ref output missing: %q", want)
		}
	}
}
