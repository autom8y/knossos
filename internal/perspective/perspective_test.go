package perspective

import (
	"os"
	"path/filepath"
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
			frontmatter: `name: pythia
description: Orchestrates development lifecycle
role: orchestrator
type: orchestrator
model: opus
color: blue
schema_version: "1.0"
maxTurns: 40`,
			body:           "# System Prompt\n\nYou are the orchestrator.\n",
			wantName:       "pythia",
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

	// Should have all 5 MVP layers
	expectedLayers := []string{"L1", "L3", "L4", "L5", "L9"}
	for _, key := range expectedLayers {
		if _, ok := doc.Layers[key]; !ok {
			t.Errorf("missing layer %s", key)
		}
	}

	// Assembly metadata should have counts
	total := doc.AssemblyMetadata.LayersResolved + doc.AssemblyMetadata.LayersDegraded + doc.AssemblyMetadata.LayersFailed
	if total != 5 {
		t.Errorf("total layers = %d, want 5", total)
	}
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
