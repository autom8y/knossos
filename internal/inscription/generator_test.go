package inscription

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"test": {Owner: OwnerKnossos},
		},
	}
	ctx := &RenderContext{ActiveRite: "test-rite"}

	gen := NewGenerator("/templates", manifest, ctx)

	if gen.TemplateDir != "/templates" {
		t.Errorf("NewGenerator() TemplateDir = %q, want '/templates'", gen.TemplateDir)
	}
	if gen.Manifest != manifest {
		t.Error("NewGenerator() Manifest not set correctly")
	}
	if gen.Context != ctx {
		t.Error("NewGenerator() Context not set correctly")
	}
}

func TestGenerator_GenerateSection_Knossos(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"execution-mode": {Owner: OwnerKnossos},
		},
	}
	ctx := &RenderContext{ActiveRite: "10x-dev"}
	gen := NewGenerator("", manifest, ctx)

	content, err := gen.GenerateSection("execution-mode")
	if err != nil {
		t.Fatalf("GenerateSection() error = %v", err)
	}

	if !strings.Contains(content, "## Execution Mode") {
		t.Error("GenerateSection() should contain '## Execution Mode'")
	}
	if !strings.Contains(content, "Native") {
		t.Error("GenerateSection() should contain execution mode table")
	}
}

func TestGenerator_GenerateSection_Regenerate(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"quick-start": {Owner: OwnerRegenerate, Source: "ACTIVE_RITE+agents"},
		},
	}
	ctx := &RenderContext{
		ActiveRite: "10x-dev",
		AgentCount: 3,
		Agents: []AgentInfo{
			{Name: "architect", Role: "Designs systems", Produces: "TDD"},
			{Name: "engineer", Role: "Builds code", Produces: "Code"},
		},
	}
	gen := NewGenerator("", manifest, ctx)

	content, err := gen.GenerateSection("quick-start")
	if err != nil {
		t.Fatalf("GenerateSection() error = %v", err)
	}

	if !strings.Contains(content, "## Quick Start") {
		t.Error("GenerateSection() should contain '## Quick Start'")
	}
	if !strings.Contains(content, "10x-dev") {
		t.Error("GenerateSection() should contain rite name")
	}
	if !strings.Contains(content, "architect") {
		t.Error("GenerateSection() should contain agent names")
	}
}

func TestGenerator_GenerateSection_Satellite(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"project-custom": {Owner: OwnerSatellite},
		},
	}
	gen := NewGenerator("", manifest, nil)

	content, err := gen.GenerateSection("project-custom")
	if err != nil {
		t.Fatalf("GenerateSection() error = %v", err)
	}

	// Satellite regions return empty (content comes from existing file)
	if content != "" {
		t.Errorf("GenerateSection() satellite should return empty, got %q", content)
	}
}

func TestGenerator_GenerateSection_NotFound(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{},
	}
	gen := NewGenerator("", manifest, nil)

	_, err := gen.GenerateSection("nonexistent")
	if err == nil {
		t.Error("GenerateSection() expected error for nonexistent region")
	}
}

func TestGenerator_RenderRegion(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"execution-mode": {Owner: OwnerKnossos},
		},
	}
	gen := NewGenerator("", manifest, nil)

	wrapped, err := gen.RenderRegion("execution-mode")
	if err != nil {
		t.Fatalf("RenderRegion() error = %v", err)
	}

	// Should be wrapped with markers
	if !strings.HasPrefix(wrapped, "<!-- KNOSSOS:START execution-mode -->") {
		t.Error("RenderRegion() should start with START marker")
	}
	if !strings.HasSuffix(wrapped, "<!-- KNOSSOS:END execution-mode -->") {
		t.Error("RenderRegion() should end with END marker")
	}
}

func TestGenerator_RenderRegion_Regenerate_Options(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"quick-start": {Owner: OwnerRegenerate, Source: "ACTIVE_RITE"},
		},
	}
	ctx := &RenderContext{ActiveRite: "test", AgentCount: 2}
	gen := NewGenerator("", manifest, ctx)

	wrapped, err := gen.RenderRegion("quick-start")
	if err != nil {
		t.Fatalf("RenderRegion() error = %v", err)
	}

	// Should include options in marker
	if !strings.Contains(wrapped, "regenerate=true") {
		t.Error("RenderRegion() regenerate marker should have regenerate=true option")
	}
	if !strings.Contains(wrapped, "source=ACTIVE_RITE") {
		t.Error("RenderRegion() regenerate marker should have source option")
	}
}

func TestGenerator_SetSectionTemplate(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"custom": {Owner: OwnerKnossos},
		},
	}
	ctx := &RenderContext{ActiveRite: "test-rite"}
	gen := NewGenerator("", manifest, ctx)

	// Set custom template
	gen.SetSectionTemplate("custom", "## Custom Section\n\nRite: {{.ActiveRite}}")

	content, err := gen.GenerateSection("custom")
	if err != nil {
		t.Fatalf("GenerateSection() error = %v", err)
	}

	if !strings.Contains(content, "## Custom Section") {
		t.Error("GenerateSection() should use custom template")
	}
	if !strings.Contains(content, "Rite: test-rite") {
		t.Error("GenerateSection() should render template variables")
	}
}

func TestGenerator_TemplateFuncs(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"test": {Owner: OwnerKnossos},
		},
	}
	ctx := &RenderContext{
		ActiveRite: "test",
		Agents: []AgentInfo{
			{Name: "architect", Role: "Designs", Produces: "TDD"},
		},
	}
	gen := NewGenerator("", manifest, ctx)

	// Test lower function
	gen.SetSectionTemplate("test", `{{lower "HELLO"}}`)
	content, _ := gen.GenerateSection("test")
	if content != "hello" {
		t.Errorf("lower template func failed, got %q", content)
	}

	// Test upper function
	gen.SetSectionTemplate("test", `{{upper "hello"}}`)
	content, _ = gen.GenerateSection("test")
	if content != "HELLO" {
		t.Errorf("upper template func failed, got %q", content)
	}

	// Test agents function
	gen.SetSectionTemplate("test", `{{agents}}`)
	content, _ = gen.GenerateSection("test")
	if !strings.Contains(content, "architect") {
		t.Error("agents template func should include agent names")
	}

	// Test term function
	gen.SetSectionTemplate("test", `{{term "knossos"}}`)
	content, _ = gen.GenerateSection("test")
	if !strings.Contains(content, "labyrinth") {
		t.Error("term template func should return definition")
	}
}

func TestGenerator_LoadAgentTable(t *testing.T) {
	tests := []struct {
		name   string
		agents []AgentInfo
		want   string
	}{
		{
			name:   "empty agents",
			agents: nil,
			want:   "| Agent | Role | Produces |",
		},
		{
			name: "with agents",
			agents: []AgentInfo{
				{Name: "architect", Role: "Designs systems", Produces: "TDD"},
				{Name: "engineer", Role: "Builds code", Produces: "Code"},
			},
			want: "**architect**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator("", nil, &RenderContext{Agents: tt.agents})
			table := gen.loadAgentTable()

			if !strings.Contains(table, tt.want) {
				t.Errorf("loadAgentTable() = %q, want to contain %q", table, tt.want)
			}
		})
	}
}

func TestGenerator_LookupTerminology(t *testing.T) {
	tests := []struct {
		term string
		want string
	}{
		{"knossos", "labyrinth"},
		{"ariadne", "clew"},
		{"theseus", "amnesia"},
		{"unknown", "unknown"}, // Returns term if not found
	}

	gen := NewGenerator("", nil, nil)

	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			got := gen.lookupTerminology(tt.term)
			if !strings.Contains(strings.ToLower(got), strings.ToLower(tt.want)) {
				t.Errorf("lookupTerminology(%q) = %q, want to contain %q", tt.term, got, tt.want)
			}
		})
	}
}

func TestGenerator_LookupTerminology_CustomVars(t *testing.T) {
	ctx := &RenderContext{
		KnossosVars: map[string]string{
			"term_custom": "My custom definition",
		},
	}
	gen := NewGenerator("", nil, ctx)

	got := gen.lookupTerminology("custom")
	if got != "My custom definition" {
		t.Errorf("lookupTerminology('custom') = %q, want 'My custom definition'", got)
	}
}

func TestGenerator_GenerateQuickStartContent(t *testing.T) {
	ctx := &RenderContext{
		ActiveRite: "10x-dev",
		AgentCount: 5,
		Agents: []AgentInfo{
			{Name: "architect", Role: "System design", Produces: "TDD"},
		},
	}
	gen := NewGenerator("", nil, ctx)

	content, err := gen.generateQuickStartContent()
	if err != nil {
		t.Fatalf("generateQuickStartContent() error = %v", err)
	}

	checks := []string{
		"## Quick Start",
		"5-agent workflow",
		"10x-dev",
		"architect",
		"prompting",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("generateQuickStartContent() missing %q", check)
		}
	}
}

func TestGenerator_GenerateAgentConfigsContent(t *testing.T) {
	ctx := &RenderContext{
		Agents: []AgentInfo{
			{Name: "architect", File: "architect.md", Role: "System design"},
			{Name: "engineer", File: "engineer.md", Role: "Builds code"},
		},
	}
	gen := NewGenerator("", nil, ctx)

	content, err := gen.generateAgentConfigsContent()
	if err != nil {
		t.Fatalf("generateAgentConfigsContent() error = %v", err)
	}

	if !strings.Contains(content, "## Agents") {
		t.Error("generateAgentConfigsContent() missing header")
	}
	if !strings.Contains(content, "architect.md") {
		t.Error("generateAgentConfigsContent() missing agent file")
	}
	if !strings.Contains(content, "engineer.md") {
		t.Error("generateAgentConfigsContent() missing second agent file")
	}
}

func TestGenerator_GetDefaultSectionContent(t *testing.T) {
	gen := NewGenerator("", nil, nil)

	// Test each known section
	sections := []string{
		"execution-mode",
		"agent-routing",
		"commands",
		"platform-infrastructure",
		"navigation",
		"slash-commands",
		"quick-start",
		"agent-configurations",
	}

	for _, section := range sections {
		t.Run(section, func(t *testing.T) {
			content, err := gen.getDefaultSectionContent(section)
			if err != nil {
				t.Errorf("getDefaultSectionContent(%q) error = %v", section, err)
			}
			if content == "" {
				t.Errorf("getDefaultSectionContent(%q) returned empty", section)
			}
		})
	}
}

func TestGenerator_GetDefaultSectionContent_Unknown(t *testing.T) {
	gen := NewGenerator("", nil, nil)

	_, err := gen.getDefaultSectionContent("unknown-section")
	if err == nil {
		t.Error("getDefaultSectionContent('unknown-section') expected error")
	}
}

func TestGenerator_GenerateAll(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"execution-mode": {Owner: OwnerKnossos},
			"quick-start":    {Owner: OwnerRegenerate, Source: "ACTIVE_RITE"},
			"project-custom": {Owner: OwnerSatellite},
		},
		SectionOrder: []string{"execution-mode", "quick-start", "project-custom"},
	}
	ctx := &RenderContext{ActiveRite: "test", AgentCount: 2}
	gen := NewGenerator("", manifest, ctx)

	result, err := gen.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll() error = %v", err)
	}

	// Should have knossos and regenerate content
	if _, ok := result["execution-mode"]; !ok {
		t.Error("GenerateAll() missing 'execution-mode'")
	}
	if _, ok := result["quick-start"]; !ok {
		t.Error("GenerateAll() missing 'quick-start'")
	}

	// Satellite content is now generated (merger preserves existing if present)
	if _, ok := result["project-custom"]; !ok {
		t.Error("GenerateAll() should include satellite 'project-custom' for initial template")
	}
}

func TestGenerator_TemplateFile(t *testing.T) {
	// Create temp directory with template
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, "sections")
	os.MkdirAll(sectionsDir, 0755)

	templateContent := "## Test Section\n\nRite: {{.ActiveRite}}"
	templatePath := filepath.Join(sectionsDir, "test-section.md.tpl")
	os.WriteFile(templatePath, []byte(templateContent), 0644)

	manifest := &Manifest{
		Regions: map[string]*Region{
			"test-section": {Owner: OwnerKnossos},
		},
	}
	ctx := &RenderContext{ActiveRite: "my-rite"}
	gen := NewGenerator(tmpDir, manifest, ctx)

	content, err := gen.GenerateSection("test-section")
	if err != nil {
		t.Fatalf("GenerateSection() error = %v", err)
	}

	if !strings.Contains(content, "## Test Section") {
		t.Error("GenerateSection() should use template file content")
	}
	if !strings.Contains(content, "Rite: my-rite") {
		t.Error("GenerateSection() should render template variables")
	}
}

func TestGenerator_TemplateError(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"test": {Owner: OwnerKnossos},
		},
	}
	gen := NewGenerator("", manifest, nil)

	// Invalid template syntax
	gen.SetSectionTemplate("test", "{{.Invalid}")

	_, err := gen.GenerateSection("test")
	if err == nil {
		t.Error("GenerateSection() expected error for invalid template")
	}
}

func TestGenerator_RegenerateFromSource_UnknownSource(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"test": {Owner: OwnerRegenerate, Source: "unknown/source"},
		},
	}
	gen := NewGenerator("", manifest, nil)

	_, err := gen.GenerateSection("test")
	if err == nil {
		t.Error("GenerateSection() expected error for unknown source")
	}
}

func TestGenerator_NilContext(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"quick-start": {Owner: OwnerRegenerate, Source: "ACTIVE_RITE"},
		},
	}
	gen := NewGenerator("", manifest, nil)

	content, err := gen.generateQuickStartContent()
	if err != nil {
		t.Fatalf("generateQuickStartContent() error = %v", err)
	}

	// Should return default content without context
	if !strings.Contains(content, "## Quick Start") {
		t.Error("generateQuickStartContent() should return default content")
	}
}

func TestGenerator_BuildMarkerOptions(t *testing.T) {
	manifest := &Manifest{
		Regions: map[string]*Region{
			"knossos-region":    {Owner: OwnerKnossos},
			"regenerate-region": {Owner: OwnerRegenerate, Source: "agents/*.md"},
			"satellite-region":  {Owner: OwnerSatellite},
		},
	}
	gen := NewGenerator("", manifest, nil)

	// Knossos regions have no special options
	opts := gen.buildMarkerOptions("knossos-region")
	if len(opts) != 0 {
		t.Errorf("buildMarkerOptions() knossos region should have no options, got %v", opts)
	}

	// Regenerate regions have options
	opts = gen.buildMarkerOptions("regenerate-region")
	if opts["regenerate"] != "true" {
		t.Error("buildMarkerOptions() regenerate region missing regenerate=true")
	}
	if opts["source"] != "agents/*.md" {
		t.Errorf("buildMarkerOptions() regenerate region source = %q, want 'agents/*.md'", opts["source"])
	}
}

func TestAgentInfo_Fields(t *testing.T) {
	agent := AgentInfo{
		Name:     "architect",
		File:     "architect.md",
		Role:     "System design",
		Produces: "TDD",
	}

	if agent.Name != "architect" {
		t.Error("AgentInfo Name field not set")
	}
	if agent.File != "architect.md" {
		t.Error("AgentInfo File field not set")
	}
	if agent.Role != "System design" {
		t.Error("AgentInfo Role field not set")
	}
	if agent.Produces != "TDD" {
		t.Error("AgentInfo Produces field not set")
	}
}

func TestRenderContext_Fields(t *testing.T) {
	ctx := &RenderContext{
		ActiveRite:  "test-rite",
		AgentCount:  5,
		Agents:      []AgentInfo{{Name: "test"}},
		KnossosVars: map[string]string{"key": "value"},
		ProjectRoot: "/project",
	}

	if ctx.ActiveRite != "test-rite" {
		t.Error("RenderContext ActiveRite not set")
	}
	if ctx.AgentCount != 5 {
		t.Error("RenderContext AgentCount not set")
	}
	if len(ctx.Agents) != 1 {
		t.Error("RenderContext Agents not set")
	}
	if ctx.KnossosVars["key"] != "value" {
		t.Error("RenderContext KnossosVars not set")
	}
	if ctx.ProjectRoot != "/project" {
		t.Error("RenderContext ProjectRoot not set")
	}
}
