package materialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		HandoffCriteria: "",
		RiteAntiPatterns: "",
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
		{"frontmatter start", "---\nname: pythia"},
		{"description", "Routes development work through requirements"},
		{"color", "color: blue"},
		{"skills orchestrator-templates", "- orchestrator-templates"},
		{"skills 10x-workflow", "- 10x-workflow"},
		{"disallowedTools", "disallowedTools:"},
		{"contract", "contract:"},
		{"frontmatter end", "---\n\n# Pythia"},

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
		"# Pythia",
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
	_, err := RenderArchetype("/nonexistent", "orchestrator.md.tpl", tenxDevData())
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read archetype template") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRenderArchetypeFromString_InvalidTemplate(t *testing.T) {
	_, err := RenderArchetypeFromString("{{.Broken", "bad.tpl", tenxDevData())
	if err == nil {
		t.Fatal("expected error for invalid template syntax, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRenderArchetype_FrontmatterDelimiters(t *testing.T) {
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
	if !strings.Contains(output, "\n---\n\n# Pythia") {
		t.Error("frontmatter closing delimiter must be followed by # Pythia heading")
	}

	// Verify no triple --- appears in the body (only in frontmatter)
	body := output[strings.Index(output[4:], "\n---\n")+4+5:]
	if strings.Contains(body, "\n---\n") {
		t.Error("unexpected --- delimiter found in body content")
	}
	_ = count // used for context, assertion is above
}

func TestRenderArchetype_SkillsYAMLList(t *testing.T) {
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
		end := start + 100
		if end > len(output) {
			end = len(output)
		}
		t.Errorf("skills not rendered as expected YAML list.\nGot:\n%s", output[start:end])
	}
}
