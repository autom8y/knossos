package perspective

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/provenance"
)

// newTestParseContext creates a ParseContext from a temporary test fixture directory.
// It creates minimal rite source files with the given agent frontmatter and body.
func newTestParseContext(t *testing.T, agentFrontmatter, agentBody string, opts ...testContextOpt) *ParseContext {
	t.Helper()
	dir := t.TempDir()

	// Create directory structure
	riteName := "test-rite"
	riteDir := filepath.Join(dir, "rites", riteName)
	agentsDir := filepath.Join(riteDir, "agents")
	knossosDir := filepath.Join(dir, ".knossos")
	claudeDir := filepath.Join(dir, ".claude")
	claudeAgentsDir := filepath.Join(claudeDir, "agents")

	for _, d := range []string{agentsDir, knossosDir, claudeDir, claudeAgentsDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Write ACTIVE_RITE
	if err := os.WriteFile(filepath.Join(knossosDir, "ACTIVE_RITE"), []byte(riteName), 0644); err != nil {
		t.Fatal(err)
	}

	// Write agent source file
	agentName := "test-agent"
	content := "---\n" + agentFrontmatter + "\n---\n" + agentBody
	if err := os.WriteFile(filepath.Join(agentsDir, agentName+".md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Write minimal rite manifest
	manifestContent := `name: test-rite
version: "1.0"
entry_agent: test-agent
agents:
  - name: test-agent
    role: test role
`
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write shared manifest with hook_defaults
	sharedDir := filepath.Join(dir, "rites", "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatal(err)
	}
	sharedManifest := `name: shared
hook_defaults:
  write_guard:
    allow_paths: [".ledge/", ".know/"]
    timeout: 3
`
	if err := os.WriteFile(filepath.Join(sharedDir, "manifest.yaml"), []byte(sharedManifest), 0644); err != nil {
		t.Fatal(err)
	}

	// Apply optional modifications
	cfg := &testContextConfig{dir: dir, riteName: riteName, agentName: agentName}
	for _, o := range opts {
		o(cfg)
	}

	ctx, err := NewParseContext(PerspectiveOptions{
		AgentName:   agentName,
		RiteName:    riteName,
		Mode:        "default",
		ProjectRoot: dir,
	})
	if err != nil {
		t.Fatalf("NewParseContext: %v", err)
	}

	return ctx
}

type testContextConfig struct {
	dir       string
	riteName  string
	agentName string
}

type testContextOpt func(*testContextConfig)

// withSeedFile creates a seed MEMORY.md file for the agent.
func withSeedFile(content string) testContextOpt {
	return func(cfg *testContextConfig) {
		seedDir := filepath.Join(cfg.dir, ".claude", "agent-memory", cfg.agentName)
		if err := os.MkdirAll(seedDir, 0755); err != nil {
			panic(err)
		}
		if err := os.WriteFile(filepath.Join(seedDir, "MEMORY.md"), []byte(content), 0644); err != nil {
			panic(err)
		}
	}
}

// withWorkflow writes a workflow.yaml file in the rite source dir.
func withWorkflow(content string) testContextOpt {
	return func(cfg *testContextConfig) {
		path := filepath.Join(cfg.dir, "rites", cfg.riteName, "workflow.yaml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			panic(err)
		}
	}
}

// withOrchestrator writes an orchestrator.yaml file in the rite source dir.
func withOrchestrator(content string) testContextOpt {
	return func(cfg *testContextConfig) {
		path := filepath.Join(cfg.dir, "rites", cfg.riteName, "orchestrator.yaml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			panic(err)
		}
	}
}

// withRiteManifest overwrites the rite manifest.yaml with custom content.
func withRiteManifest(content string) testContextOpt {
	return func(cfg *testContextConfig) {
		path := filepath.Join(cfg.dir, "rites", cfg.riteName, "manifest.yaml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			panic(err)
		}
	}
}

// withSkillsDirs creates skill directories under .claude/skills/.
func withSkillsDirs(names []string) testContextOpt {
	return func(cfg *testContextConfig) {
		for _, name := range names {
			dir := filepath.Join(cfg.dir, ".claude", "skills", name)
			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}
		}
	}
}

// --- Test: Identity Resolver ---

func TestResolveIdentity(t *testing.T) {
	tests := []struct {
		name           string
		frontmatter    string
		body           string
		wantName       string
		wantRole       string
		wantType       string
		wantModel      string
		wantLines      int
		wantStatus     LayerStatus
		wantExcerptLen int
	}{
		{
			name: "full identity",
			frontmatter: `name: potnia
description: Orchestrates development lifecycle
role: orchestrator
type: orchestrator
model: opus
color: blue
schema_version: "1.0"
maxTurns: 40`,
			body:           "# System Prompt\n\nYou are the orchestrator.\n",
			wantName:       "potnia",
			wantRole:       "orchestrator",
			wantType:       "orchestrator",
			wantModel:      "opus",
			wantLines:      3,
			wantStatus:     StatusResolved,
			wantExcerptLen: 43, // includes trailing newline from body
		},
		{
			name: "minimal identity",
			frontmatter: `name: minimal
description: A minimal agent`,
			body:       "Short body.",
			wantName:   "minimal",
			wantLines:  1,
			wantStatus: StatusResolved,
		},
		{
			name:        "missing name triggers partial",
			frontmatter: `description: no name agent`,
			body:        "body",
			wantStatus:  StatusPartial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestParseContext(t, tt.frontmatter, tt.body)
			env := resolveIdentity(ctx)

			if env.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", env.Status, tt.wantStatus)
			}

			id, ok := env.Data.(*IdentityData)
			if !ok {
				t.Fatal("Data is not *IdentityData")
			}

			if tt.wantName != "" && id.Name != tt.wantName {
				t.Errorf("name = %q, want %q", id.Name, tt.wantName)
			}
			if tt.wantRole != "" && id.Role != tt.wantRole {
				t.Errorf("role = %q, want %q", id.Role, tt.wantRole)
			}
			if tt.wantType != "" && id.Type != tt.wantType {
				t.Errorf("type = %q, want %q", id.Type, tt.wantType)
			}
			if tt.wantModel != "" && id.Model != tt.wantModel {
				t.Errorf("model = %q, want %q", id.Model, tt.wantModel)
			}
			if tt.wantLines > 0 && id.SystemPromptLines != tt.wantLines {
				t.Errorf("system_prompt_lines = %d, want %d", id.SystemPromptLines, tt.wantLines)
			}
			if tt.wantExcerptLen > 0 && len(id.SystemPromptExcerpt) != tt.wantExcerptLen {
				t.Errorf("system_prompt_excerpt len = %d, want %d", len(id.SystemPromptExcerpt), tt.wantExcerptLen)
			}
		})
	}
}

// --- Test: Capability Resolver ---

func TestResolveCapability(t *testing.T) {
	tests := []struct {
		name              string
		frontmatter       string
		wantToolCount     int
		wantCCNativeCount int
		wantMCPCount      int
		wantUnknownCount  int
		wantFromDefaults  bool
	}{
		{
			name: "standard tools",
			frontmatter: `name: test
description: test
tools: Bash, Read, Write, Edit, Glob, Grep`,
			wantToolCount:     6,
			wantCCNativeCount: 6,
		},
		{
			name: "with MCP tools",
			frontmatter: `name: test
description: test
tools:
  - Bash
  - Read
  - mcp:github/create_issue`,
			wantToolCount:     3,
			wantCCNativeCount: 2,
			wantMCPCount:      1,
		},
		{
			name: "unknown tools",
			frontmatter: `name: test
description: test
tools:
  - Bash
  - CustomTool`,
			wantToolCount:     2,
			wantCCNativeCount: 1, // Bash is CC native
			wantUnknownCount:  1,
		},
		{
			name: "no tools",
			frontmatter: `name: test
description: test`,
			wantToolCount:    0,
			wantFromDefaults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestParseContext(t, tt.frontmatter, "body")
			env := resolveCapability(ctx)

			if env.Status != StatusResolved {
				t.Errorf("status = %s, want RESOLVED", env.Status)
			}

			cap, ok := env.Data.(*CapabilityData)
			if !ok {
				t.Fatal("Data is not *CapabilityData")
			}

			if len(cap.Tools) != tt.wantToolCount {
				t.Errorf("tools count = %d, want %d (tools: %v)", len(cap.Tools), tt.wantToolCount, cap.Tools)
			}
			if len(cap.CCNativeTools) != tt.wantCCNativeCount {
				t.Errorf("cc_native count = %d, want %d", len(cap.CCNativeTools), tt.wantCCNativeCount)
			}
			if len(cap.MCPTools) != tt.wantMCPCount {
				t.Errorf("mcp count = %d, want %d", len(cap.MCPTools), tt.wantMCPCount)
			}
			if len(cap.UnknownTools) != tt.wantUnknownCount {
				t.Errorf("unknown count = %d, want %d", len(cap.UnknownTools), tt.wantUnknownCount)
			}
			if cap.ToolsFromDefaults != tt.wantFromDefaults {
				t.Errorf("tools_from_defaults = %v, want %v", cap.ToolsFromDefaults, tt.wantFromDefaults)
			}
		})
	}
}

// --- Test: Constraint Resolver ---

func TestResolveConstraint(t *testing.T) {
	tests := []struct {
		name           string
		frontmatter    string
		wantDisallowed int
		wantWriteGuard bool
		wantContract   bool
		wantStatus     LayerStatus
	}{
		{
			name: "full constraints",
			frontmatter: `name: test
description: test
tools: Bash, Read
disallowedTools: Write, Edit
write-guard: true
contract:
  must_use: [Read]
  must_not: [never write CLAUDE.md]`,
			wantDisallowed: 2,
			wantWriteGuard: true,
			wantContract:   true,
			wantStatus:     StatusResolved,
		},
		{
			name: "no contract is partial",
			frontmatter: `name: test
description: test
tools: Bash
disallowedTools: Write`,
			wantDisallowed: 1,
			wantWriteGuard: false,
			wantContract:   false,
			wantStatus:     StatusPartial,
		},
		{
			name: "write-guard opt-out",
			frontmatter: `name: test
description: test
write-guard: false`,
			wantWriteGuard: false,
			wantStatus:     StatusPartial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestParseContext(t, tt.frontmatter, "body")
			env := resolveConstraint(ctx)

			if env.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", env.Status, tt.wantStatus)
			}

			con, ok := env.Data.(*ConstraintData)
			if !ok {
				t.Fatal("Data is not *ConstraintData")
			}

			if len(con.DisallowedTools) != tt.wantDisallowed {
				t.Errorf("disallowed count = %d, want %d", len(con.DisallowedTools), tt.wantDisallowed)
			}
			if (con.WriteGuard != nil && con.WriteGuard.Enabled) != tt.wantWriteGuard {
				t.Errorf("write_guard enabled = %v, want %v", con.WriteGuard != nil && con.WriteGuard.Enabled, tt.wantWriteGuard)
			}
			if (con.BehavioralContract != nil) != tt.wantContract {
				t.Errorf("has contract = %v, want %v", con.BehavioralContract != nil, tt.wantContract)
			}
		})
	}
}

// --- Test: Memory Resolver ---

func TestResolveMemory(t *testing.T) {
	tests := []struct {
		name           string
		frontmatter    string
		seedFile       bool
		wantScope      string
		wantEnabled    bool
		wantSeedExists bool
		wantStatus     LayerStatus
	}{
		{
			name: "project scope with seed",
			frontmatter: `name: test-agent
description: test
memory: project`,
			seedFile:       true,
			wantScope:      "project",
			wantEnabled:    true,
			wantSeedExists: true,
			wantStatus:     StatusPartial, // project scope has OPAQUE runtime path
		},
		{
			name: "boolean true normalizes to project",
			frontmatter: `name: test-agent
description: test
memory: true`,
			wantScope:      "project",
			wantEnabled:    true,
			wantSeedExists: false,
			wantStatus:     StatusPartial,
		},
		{
			name: "disabled memory",
			frontmatter: `name: test-agent
description: test`,
			wantScope:   "",
			wantEnabled: false,
			wantStatus:  StatusResolved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []testContextOpt
			if tt.seedFile {
				opts = append(opts, withSeedFile("# Memory\nLine 1\nLine 2\n"))
			}
			ctx := newTestParseContext(t, tt.frontmatter, "body", opts...)
			env := resolveMemory(ctx)

			if env.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", env.Status, tt.wantStatus)
			}

			mem, ok := env.Data.(*MemoryData)
			if !ok {
				t.Fatal("Data is not *MemoryData")
			}

			if mem.Scope != tt.wantScope {
				t.Errorf("scope = %q, want %q", mem.Scope, tt.wantScope)
			}
			if mem.Enabled != tt.wantEnabled {
				t.Errorf("enabled = %v, want %v", mem.Enabled, tt.wantEnabled)
			}
			if mem.SeedFile != nil && mem.SeedFile.Exists != tt.wantSeedExists {
				t.Errorf("seed_file.exists = %v, want %v", mem.SeedFile.Exists, tt.wantSeedExists)
			}
		})
	}
}

// --- Test: Provenance Resolver ---

func TestResolveProvenance(t *testing.T) {
	t.Run("no entry yields partial", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: test`, "body")
		env := resolveProvenance(ctx)

		if env.Status != StatusPartial {
			t.Errorf("status = %s, want PARTIAL", env.Status)
		}

		prov, ok := env.Data.(*ProvenanceData)
		if !ok {
			t.Fatal("Data is not *ProvenanceData")
		}
		if prov.Owner != "" {
			t.Errorf("owner = %q, want empty", prov.Owner)
		}
	})

	t.Run("matching checksum not diverged", func(t *testing.T) {
		dir := t.TempDir()
		riteName := "test-rite"
		riteDir := filepath.Join(dir, "rites", riteName)
		agentsDir := filepath.Join(riteDir, "agents")
		knossosDir := filepath.Join(dir, ".knossos")
		claudeDir := filepath.Join(dir, ".claude")
		claudeAgentsDir := filepath.Join(claudeDir, "agents")

		for _, d := range []string{agentsDir, knossosDir, claudeDir, claudeAgentsDir} {
			if err := os.MkdirAll(d, 0755); err != nil {
				t.Fatal(err)
			}
		}

		_ = os.WriteFile(filepath.Join(knossosDir, "ACTIVE_RITE"), []byte(riteName), 0644)

		agentContent := "---\nname: test-agent\ndescription: test\n---\nbody"
		_ = os.WriteFile(filepath.Join(agentsDir, "test-agent.md"), []byte(agentContent), 0644)

		materializedContent := "---\nname: test-agent\ndescription: test\n---\nbody"
		_ = os.WriteFile(filepath.Join(claudeAgentsDir, "test-agent.md"), []byte(materializedContent), 0644)

		_ = os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte("name: test-rite\nversion: \"1.0\"\nagents: []\n"), 0644)

		sharedDir := filepath.Join(dir, "rites", "shared")
		_ = os.MkdirAll(sharedDir, 0755)
		_ = os.WriteFile(filepath.Join(sharedDir, "manifest.yaml"), []byte("name: shared\n"), 0644)

		// Compute checksum of materialized file
		cs, _ := checksum.File(filepath.Join(claudeAgentsDir, "test-agent.md"))

		// Write provenance manifest with matching checksum
		manifest := &provenance.ProvenanceManifest{
			SchemaVersion: "2.0",
			LastSync:      time.Now().UTC(),
			ActiveRite:    riteName,
			Entries: map[string]*provenance.ProvenanceEntry{
				"agents/test-agent.md": {
					Owner:      provenance.OwnerKnossos,
					Scope:      provenance.ScopeRite,
					SourcePath: "rites/test-rite/agents/test-agent.md",
					SourceType: "project",
					Checksum:   cs,
					LastSynced: time.Now().UTC(),
				},
			},
		}
		_ = provenance.Save(provenance.ManifestPath(knossosDir), manifest)

		ctx, err := NewParseContext(PerspectiveOptions{
			AgentName:   "test-agent",
			RiteName:    riteName,
			Mode:        "default",
			ProjectRoot: dir,
		})
		if err != nil {
			t.Fatal(err)
		}

		env := resolveProvenance(ctx)
		if env.Status != StatusResolved {
			t.Errorf("status = %s, want RESOLVED", env.Status)
		}

		prov, _ := env.Data.(*ProvenanceData)
		if prov.Owner != "knossos" {
			t.Errorf("owner = %q, want knossos", prov.Owner)
		}
		if prov.Diverged {
			t.Error("diverged = true, want false")
		}
	})
}

// --- Test: Audit Checks ---

func TestAuditChecks(t *testing.T) {
	t.Run("missing contract warning", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: test
tools: Bash`, "body")
		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "audit"}, time.Now())
		overlay := RunAudit(doc, ctx)

		found := false
		for _, f := range overlay.Findings {
			if f.ID == "AUDIT-001" {
				found = true
				if f.Severity != SeverityWarning {
					t.Errorf("AUDIT-001 severity = %s, want WARNING", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected AUDIT-001 finding for missing contract")
		}
	})

	t.Run("tool conflict critical", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: test
tools:
  - Bash
  - Write
disallowedTools:
  - Write`, "body")
		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "audit"}, time.Now())
		overlay := RunAudit(doc, ctx)

		found := false
		for _, f := range overlay.Findings {
			if f.ID == "AUDIT-002" {
				found = true
				if f.Severity != SeverityCritical {
					t.Errorf("AUDIT-002 severity = %s, want CRITICAL", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected AUDIT-002 finding for tool conflict")
		}
	})

	t.Run("memory enabled no seed", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: test
memory: project`, "body")
		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "audit"}, time.Now())
		overlay := RunAudit(doc, ctx)

		found := false
		for _, f := range overlay.Findings {
			if f.ID == "AUDIT-003" {
				found = true
				if f.Severity != SeverityWarning {
					t.Errorf("AUDIT-003 severity = %s, want WARNING", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected AUDIT-003 finding for memory without seed")
		}
	})

	t.Run("write-guard no extra paths info", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: test
write-guard: true`, "body")
		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "audit"}, time.Now())
		overlay := RunAudit(doc, ctx)

		found := false
		for _, f := range overlay.Findings {
			if f.ID == "AUDIT-006" {
				found = true
				if f.Severity != SeverityInfo {
					t.Errorf("AUDIT-006 severity = %s, want INFO", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected AUDIT-006 finding for write-guard without agent extra paths")
		}
	})
}

// --- Test: Assemble ---

func TestAssemble(t *testing.T) {
	ctx := newTestParseContext(t, `name: test-agent
description: A test agent
tools: Bash, Read
model: opus`, "# System Prompt\nYou do things.\n")

	start := time.Now()
	doc := Assemble(ctx, PerspectiveOptions{
		AgentName: "test-agent",
		Mode:      "default",
	}, start)

	if doc.Version != "1.0" {
		t.Errorf("version = %q, want 1.0", doc.Version)
	}
	if doc.Agent != "test-agent" {
		t.Errorf("agent = %q, want test-agent", doc.Agent)
	}
	if doc.Rite != "test-rite" {
		t.Errorf("rite = %q, want test-rite", doc.Rite)
	}

	// Should have all 9 layers (L1-L9)
	expectedLayers := []string{"L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8", "L9"}
	for _, key := range expectedLayers {
		if _, ok := doc.Layers[key]; !ok {
			t.Errorf("missing layer %s", key)
		}
	}

	// Assembly metadata should have counts
	total := doc.AssemblyMetadata.LayersResolved + doc.AssemblyMetadata.LayersDegraded + doc.AssemblyMetadata.LayersFailed
	if total != 9 {
		t.Errorf("total layers = %d, want 9", total)
	}
}

// --- Test: Position Resolver ---

func TestResolvePosition(t *testing.T) {
	workflowContent := `
name: test-rite
phases:
  - name: analysis
    agent: analyst
    produces: gap-analysis
    next: implementation
  - name: implementation
    agent: test-agent
    produces: code
    next: validation
    condition: "complexity >= MODULE"
  - name: validation
    agent: validator
    produces: report
    next: null
entry_point:
  agent: analyst
complexity_levels:
  - name: PATCH
    phases: [analysis, validation]
  - name: MODULE
    phases: [analysis, implementation, validation]
back_routes:
  - source_phase: validation
    trigger: "fail(implementation)"
    target_phase: implementation
    target_agent: test-agent
    requires_user_confirmation: false
    condition: "Implementation bug"
commands:
  - name: build
    file: build.md
    description: "Build something"
    primary_agent: test-agent
`

	orchestratorContent := `
handoff_criteria:
  implementation:
    - "All tests pass"
    - "Code reviewed"
`

	t.Run("agent in workflow", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent
tools: Bash, Read`, "body",
			withWorkflow(workflowContent),
			withOrchestrator(orchestratorContent))

		env := resolvePosition(ctx)
		if env.Status != StatusResolved {
			t.Errorf("status = %s, want RESOLVED", env.Status)
		}
		pos, ok := env.Data.(*PositionData)
		if !ok {
			t.Fatal("data is not *PositionData")
		}
		if !pos.InWorkflow {
			t.Error("InWorkflow = false, want true")
		}
		if pos.PhaseIndex != 1 {
			t.Errorf("PhaseIndex = %d, want 1", pos.PhaseIndex)
		}
		if pos.TotalPhases != 3 {
			t.Errorf("TotalPhases = %d, want 3", pos.TotalPhases)
		}
		if pos.PhasePredecessor != "analyst" {
			t.Errorf("PhasePredecessor = %q, want analyst", pos.PhasePredecessor)
		}
		if pos.PhaseSuccessor != "validator" {
			t.Errorf("PhaseSuccessor = %q, want validator", pos.PhaseSuccessor)
		}
		if pos.PhaseCondition != "complexity >= MODULE" {
			t.Errorf("PhaseCondition = %q, want 'complexity >= MODULE'", pos.PhaseCondition)
		}
		if pos.PhaseProduces != "code" {
			t.Errorf("PhaseProduces = %q, want code", pos.PhaseProduces)
		}
		if len(pos.BackRoutes) != 1 {
			t.Fatalf("BackRoutes count = %d, want 1", len(pos.BackRoutes))
		}
		if pos.BackRoutes[0].SourcePhase != "validation" {
			t.Errorf("BackRoute source = %q, want validation", pos.BackRoutes[0].SourcePhase)
		}
		if len(pos.HandoffCriteria) != 2 {
			t.Errorf("HandoffCriteria count = %d, want 2", len(pos.HandoffCriteria))
		}
		// MODULE includes implementation phase
		found := false
		for _, g := range pos.ComplexityGates {
			if g == "MODULE" {
				found = true
			}
		}
		if !found {
			t.Errorf("ComplexityGates %v does not contain MODULE", pos.ComplexityGates)
		}
	})

	t.Run("agent not in workflow", func(t *testing.T) {
		wf := `
name: test-rite
phases:
  - name: analysis
    agent: other-agent
    produces: report
    next: null
`
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent`, "body", withWorkflow(wf))

		env := resolvePosition(ctx)
		if env.Status != StatusPartial {
			t.Errorf("status = %s, want PARTIAL", env.Status)
		}
		pos := env.Data.(*PositionData)
		if pos.InWorkflow {
			t.Error("InWorkflow = true, want false")
		}
		if len(env.Gaps) == 0 {
			t.Error("expected gap for agent not in workflow")
		}
	})

	t.Run("entry point agent", func(t *testing.T) {
		wf := `
name: test-rite
entry_point:
  agent: test-agent
phases:
  - name: analysis
    agent: test-agent
    produces: report
    next: null
`
		manifest := `name: test-rite
version: "1.0"
entry_agent: test-agent
agents:
  - name: test-agent
    role: test
`
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent`, "body",
			withWorkflow(wf), withRiteManifest(manifest))

		env := resolvePosition(ctx)
		pos := env.Data.(*PositionData)
		if !pos.IsEntryPoint {
			t.Error("IsEntryPoint = false, want true")
		}
		if !pos.IsEntryAgent {
			t.Error("IsEntryAgent = false, want true")
		}
	})
}

// --- Test: Surface Resolver ---

func TestResolveSurface(t *testing.T) {
	t.Run("full surface", func(t *testing.T) {
		manifest := `name: test-rite
version: "1.0"
entry_agent: test-agent
agents:
  - name: test-agent
    role: test
dromena:
  - sync-debug
legomena:
  - ecosystem-ref
  - doc-ecosystem
`
		wf := `
name: test-rite
phases:
  - name: implementation
    agent: test-agent
    produces: code
    next: null
commands:
  - name: build
    file: build.md
    description: "Build something"
    primary_agent: test-agent
  - name: deploy
    file: deploy.md
    primary_agent: other-agent
`
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent
contract:
  must_produce: [code, docs]`, "body",
			withRiteManifest(manifest), withWorkflow(wf))

		env := resolveSurface(ctx)
		if env.Status != StatusResolved {
			t.Errorf("status = %s, want RESOLVED", env.Status)
		}
		surf := env.Data.(*SurfaceData)
		if len(surf.DromenaOwned) != 1 || surf.DromenaOwned[0] != "sync-debug" {
			t.Errorf("DromenaOwned = %v, want [sync-debug]", surf.DromenaOwned)
		}
		if len(surf.LegomenaAvailable) != 2 {
			t.Errorf("LegomenaAvailable count = %d, want 2", len(surf.LegomenaAvailable))
		}
		if len(surf.ArtifactTypes) != 1 || surf.ArtifactTypes[0] != "code" {
			t.Errorf("ArtifactTypes = %v, want [code]", surf.ArtifactTypes)
		}
		if len(surf.Commands) != 1 || surf.Commands[0].Name != "build" {
			t.Errorf("Commands = %v, want [build]", surf.Commands)
		}
		if len(surf.ContractMustProduce) != 2 {
			t.Errorf("ContractMustProduce count = %d, want 2", len(surf.ContractMustProduce))
		}
	})

	t.Run("minimal surface", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent`, "body")

		env := resolveSurface(ctx)
		if env.Status != StatusResolved {
			t.Errorf("status = %s, want RESOLVED", env.Status)
		}
		surf := env.Data.(*SurfaceData)
		if len(surf.DromenaOwned) != 0 {
			t.Errorf("DromenaOwned = %v, want empty", surf.DromenaOwned)
		}
	})
}

// --- Test: Perception Resolver ---

func TestResolvePerception(t *testing.T) {
	t.Run("explicit skills only", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent
skills:
  - ecosystem-ref
  - doc-ecosystem`, "body",
			withSkillsDirs([]string{"ecosystem-ref", "doc-ecosystem", "forge-ref", "conventions"}))

		capData := &CapabilityData{Tools: []string{"Read", "Bash", "Skill"}}
		conData := &ConstraintData{DisallowedTools: []string{}}

		env := resolvePerception(ctx, capData, conData)
		if env.Status != StatusResolved {
			t.Errorf("status = %s, want RESOLVED", env.Status)
		}
		perc := env.Data.(*PerceptionData)
		if len(perc.ExplicitSkills) != 2 {
			t.Errorf("ExplicitSkills count = %d, want 2", len(perc.ExplicitSkills))
		}
		if !perc.SkillToolAvailable {
			t.Error("SkillToolAvailable = false, want true")
		}
		if perc.TotalPreloaded != 2 {
			t.Errorf("TotalPreloaded = %d, want 2", perc.TotalPreloaded)
		}
		// On-demand should include forge-ref and conventions (not already preloaded)
		if len(perc.OnDemandSkills) != 2 {
			t.Errorf("OnDemandSkills count = %d, want 2: %v", len(perc.OnDemandSkills), perc.OnDemandSkills)
		}
	})

	t.Run("skill policy inject", func(t *testing.T) {
		manifest := `name: test-rite
version: "1.0"
entry_agent: test-agent
agents:
  - name: test-agent
    role: test
skill_policies:
  - skill: clinic-ref
    mode: inject
    requires_tools: [Read]
`
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent`, "body", withRiteManifest(manifest))

		capData := &CapabilityData{Tools: []string{"Read", "Bash"}}
		conData := &ConstraintData{DisallowedTools: []string{}}

		env := resolvePerception(ctx, capData, conData)
		perc := env.Data.(*PerceptionData)
		if len(perc.PolicyInjectedSkills) != 1 || perc.PolicyInjectedSkills[0] != "clinic-ref" {
			t.Errorf("PolicyInjectedSkills = %v, want [clinic-ref]", perc.PolicyInjectedSkills)
		}
		if len(perc.EffectivePolicies) != 1 {
			t.Fatalf("EffectivePolicies count = %d, want 1", len(perc.EffectivePolicies))
		}
		if !perc.EffectivePolicies[0].Applied {
			t.Errorf("policy Applied = false, want true")
		}
	})

	t.Run("skill tool disallowed", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent
skills:
  - ecosystem-ref`, "body",
			withSkillsDirs([]string{"ecosystem-ref", "forge-ref"}))

		capData := &CapabilityData{Tools: []string{"Read"}}
		conData := &ConstraintData{DisallowedTools: []string{"Skill"}}

		env := resolvePerception(ctx, capData, conData)
		perc := env.Data.(*PerceptionData)
		if perc.SkillToolAvailable {
			t.Error("SkillToolAvailable = true, want false")
		}
		if len(perc.OnDemandSkills) != 0 {
			t.Errorf("OnDemandSkills = %v, want empty (Skill disallowed)", perc.OnDemandSkills)
		}
	})

	t.Run("policy requires_tools not met", func(t *testing.T) {
		manifest := `name: test-rite
version: "1.0"
entry_agent: test-agent
agents:
  - name: test-agent
    role: test
skill_policies:
  - skill: clinic-ref
    mode: inject
    requires_tools: [Bash]
`
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent`, "body", withRiteManifest(manifest))

		capData := &CapabilityData{Tools: []string{"Read"}} // no Bash
		conData := &ConstraintData{DisallowedTools: []string{}}

		env := resolvePerception(ctx, capData, conData)
		perc := env.Data.(*PerceptionData)
		if len(perc.PolicyInjectedSkills) != 0 {
			t.Errorf("PolicyInjectedSkills = %v, want empty (requires_tools not met)", perc.PolicyInjectedSkills)
		}
		if len(perc.EffectivePolicies) != 1 || perc.EffectivePolicies[0].Applied {
			t.Error("policy should not be applied when requires_tools not met")
		}
	})
}

// --- Test: Phase 2 Audit Checks ---

func TestPhase2AuditChecks(t *testing.T) {
	t.Run("AUDIT-007 skills without skill tool", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L2": {Data: &PerceptionData{
					ExplicitSkills:     []string{"ecosystem-ref"},
					PolicyInjectedSkills: []string{},
					SkillToolAvailable: false,
				}},
				"L4": {Data: &ConstraintData{DisallowedTools: []string{"Skill"}}},
			},
		}
		findings := checkSkillsWithoutSkillTool(doc)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].ID != "AUDIT-007" {
			t.Errorf("ID = %s, want AUDIT-007", findings[0].ID)
		}
		if findings[0].Severity != SeverityCritical {
			t.Errorf("Severity = %s, want CRITICAL", findings[0].Severity)
		}
	})

	t.Run("AUDIT-008 orphan agent", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L6": {Data: &PositionData{InWorkflow: false}},
			},
		}
		findings := checkOrphanAgent(doc)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].ID != "AUDIT-008" {
			t.Errorf("ID = %s, want AUDIT-008", findings[0].ID)
		}
	})

	t.Run("AUDIT-009 must_produce not in artifacts", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L4": {Data: &ConstraintData{
					DisallowedTools: []string{},
					BehavioralContract: &BehavioralContractData{
						MustProduce: []string{"code", "docs"},
					},
				}},
				"L7": {Data: &SurfaceData{ArtifactTypes: []string{"code"}}},
			},
		}
		findings := checkMustProduceNotInArtifacts(doc)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].ID != "AUDIT-009" {
			t.Errorf("ID = %s, want AUDIT-009", findings[0].ID)
		}
		if findings[0].Evidence != "docs" {
			t.Errorf("Evidence = %q, want docs", findings[0].Evidence)
		}
	})

	t.Run("AUDIT-008 no finding for in-workflow agent", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L6": {Data: &PositionData{InWorkflow: true}},
			},
		}
		findings := checkOrphanAgent(doc)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for in-workflow agent, got %d", len(findings))
		}
	})
}

// --- Test: Helper functions ---

func TestDedupStrings(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"a", "b", "a", "c"}, []string{"a", "b", "c"}},
		{[]string{"x"}, []string{"x"}},
		{[]string{}, []string{}},
		{nil, []string{}},
	}

	for _, tt := range tests {
		result := dedupStrings(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("dedupStrings(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// --- Test: Horizon Resolver ---

func TestResolveHorizon(t *testing.T) {
	t.Run("tools not available computed", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent
tools: Read, Bash`, "body",
			withSkillsDirs([]string{"ecosystem-ref", "forge-ref", "conventions"}))

		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "default"}, time.Now())

		hor := getLayerData[*HorizonData](doc, "L8")
		if hor == nil {
			t.Fatal("L8 data is nil")
		}
		if hor.ToolsNotAvailable == nil {
			t.Fatal("ToolsNotAvailable is nil")
		}
		// Agent has Read and Bash; all other CC tools should be not available
		// knownCCTools has 14 tools total, agent has 2
		if len(hor.ToolsNotAvailable) != 12 {
			t.Errorf("ToolsNotAvailable count = %d, want 12 (14 total - 2 agent tools)", len(hor.ToolsNotAvailable))
		}
		// Verify sorted (deterministic output)
		for i := 1; i < len(hor.ToolsNotAvailable); i++ {
			if hor.ToolsNotAvailable[i] < hor.ToolsNotAvailable[i-1] {
				t.Errorf("ToolsNotAvailable not sorted: %v", hor.ToolsNotAvailable)
				break
			}
		}
	})

	t.Run("phases not in computed", func(t *testing.T) {
		wf := `
name: test-rite
phases:
  - name: analysis
    agent: analyst
    next: implementation
  - name: implementation
    agent: test-agent
    next: validation
  - name: validation
    agent: validator
    next: null
`
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent
tools: Read`, "body", withWorkflow(wf))

		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "default"}, time.Now())

		hor := getLayerData[*HorizonData](doc, "L8")
		if hor == nil {
			t.Fatal("L8 data is nil")
		}
		// Agent is in "implementation", should NOT be in "analysis" and "validation"
		if len(hor.PhasesNotIn) != 2 {
			t.Errorf("PhasesNotIn count = %d, want 2: %v", len(hor.PhasesNotIn), hor.PhasesNotIn)
		}
	})

	t.Run("memory blind spots", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent
memory: project`, "body")

		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "default"}, time.Now())

		hor := getLayerData[*HorizonData](doc, "L8")
		if hor == nil {
			t.Fatal("L8 data is nil")
		}
		// project scope → blind spots for user and local
		if len(hor.MemoryBlindSpots) != 2 {
			t.Errorf("MemoryBlindSpots count = %d, want 2: %v", len(hor.MemoryBlindSpots), hor.MemoryBlindSpots)
		}
	})

	t.Run("memory disabled all blind", func(t *testing.T) {
		ctx := newTestParseContext(t, `name: test-agent
description: Test agent`, "body")

		doc := Assemble(ctx, PerspectiveOptions{AgentName: "test-agent", Mode: "default"}, time.Now())

		hor := getLayerData[*HorizonData](doc, "L8")
		if hor == nil {
			t.Fatal("L8 data is nil")
		}
		if len(hor.MemoryBlindSpots) != 1 {
			t.Errorf("MemoryBlindSpots count = %d, want 1 (all blind): %v", len(hor.MemoryBlindSpots), hor.MemoryBlindSpots)
		}
	})
}

// --- Test: Simulate ---

func TestRunSimulate(t *testing.T) {
	t.Run("tool match via keyword", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L2": {Data: &PerceptionData{
					ExplicitSkills: []string{},
					OnDemandSkills: []string{},
				}},
				"L3": {Data: &CapabilityData{Tools: []string{"Read", "Bash", "Grep"}}},
				"L4": {Data: &ConstraintData{DisallowedTools: []string{}}},
				"L6": {Data: &PositionData{}},
			},
		}
		sim := RunSimulate(doc, "read a file and search for patterns")

		if len(sim.ToolMatches) == 0 {
			t.Fatal("expected tool matches")
		}
		if len(sim.CanAttempt) == 0 {
			t.Error("expected can_attempt entries for matched tools")
		}
		// "read" should match Read tool, "search" should match Grep/Glob/WebSearch, "file" should match Read/Write/Edit/Glob
		foundRead := false
		for _, m := range sim.ToolMatches {
			if m.Name == "Read" {
				foundRead = true
			}
		}
		if !foundRead {
			t.Errorf("expected Read in tool matches, got: %v", sim.ToolMatches)
		}
	})

	t.Run("constraint hit", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L2": {Data: &PerceptionData{}},
				"L3": {Data: &CapabilityData{Tools: []string{"Read"}}},
				"L4": {Data: &ConstraintData{
					DisallowedTools: []string{},
					BehavioralContract: &BehavioralContractData{
						MustNot: []string{"never edit CLAUDE.md directly"},
					},
				}},
				"L6": {Data: &PositionData{}},
			},
		}
		sim := RunSimulate(doc, "edit the CLAUDE.md file")

		if len(sim.ConstraintHits) == 0 {
			t.Error("expected constraint hit for 'edit' matching must_not rule")
		}
	})

	t.Run("disallowed tool in cannot_attempt", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L2": {Data: &PerceptionData{}},
				"L3": {Data: &CapabilityData{Tools: []string{"Read"}}},
				"L4": {Data: &ConstraintData{DisallowedTools: []string{"Write"}}},
				"L6": {Data: &PositionData{}},
			},
		}
		sim := RunSimulate(doc, "write a file")

		foundDisallowed := false
		for _, c := range sim.CannotAttempt {
			if c == "Write (disallowed)" {
				foundDisallowed = true
			}
		}
		if !foundDisallowed {
			t.Errorf("expected 'Write (disallowed)' in CannotAttempt, got: %v", sim.CannotAttempt)
		}
	})
}

// --- Test: Phase 3 Audit Checks ---

func TestPhase3AuditChecks(t *testing.T) {
	t.Run("AUDIT-011 zero reachable skills", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L2": {Data: &PerceptionData{
					SkillToolAvailable: true,
					TotalReachable:     0,
				}},
			},
		}
		findings := checkZeroReachableSkills(doc)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].ID != "AUDIT-011" {
			t.Errorf("ID = %s, want AUDIT-011", findings[0].ID)
		}
		if findings[0].Severity != SeverityWarning {
			t.Errorf("Severity = %s, want WARNING", findings[0].Severity)
		}
	})

	t.Run("AUDIT-011 no finding when skills reachable", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L2": {Data: &PerceptionData{
					SkillToolAvailable: true,
					TotalReachable:     5,
				}},
			},
		}
		findings := checkZeroReachableSkills(doc)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("AUDIT-011 no finding when skill tool not available", func(t *testing.T) {
		doc := &PerspectiveDocument{
			Layers: map[string]*LayerEnvelope{
				"L2": {Data: &PerceptionData{
					SkillToolAvailable: false,
					TotalReachable:     0,
				}},
			},
		}
		findings := checkZeroReachableSkills(doc)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings when Skill tool unavailable, got %d", len(findings))
		}
	})
}

// --- Test: Tokenizer ---

func TestTokenize(t *testing.T) {
	tokens := tokenize("Read a file and search for patterns!")
	if len(tokens) == 0 {
		t.Fatal("expected tokens")
	}
	// Single-char tokens should be filtered
	for _, tok := range tokens {
		if len(tok) <= 1 {
			t.Errorf("single-char token not filtered: %q", tok)
		}
	}
	// Should be lowercase
	for _, tok := range tokens {
		lower := strings.ToLower(tok)
		if tok != lower {
			t.Errorf("token not lowercase: %q", tok)
		}
	}
}

func TestStringFromMap(t *testing.T) {
	m := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": nil,
	}

	if got := stringFromMap(m, "key1"); got != "value1" {
		t.Errorf("got %q, want value1", got)
	}
	if got := stringFromMap(m, "key2"); got != "" {
		t.Errorf("got %q, want empty (non-string)", got)
	}
	if got := stringFromMap(m, "missing"); got != "" {
		t.Errorf("got %q, want empty (missing)", got)
	}
}
