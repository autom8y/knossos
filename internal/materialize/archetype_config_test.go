package materialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadOrchestratorConfig_ReviewYAML(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	riteFS := os.DirFS(filepath.Join(root, "rites", "review"))

	config, err := loadOrchestratorConfig(riteFS)
	if err != nil {
		t.Fatalf("loadOrchestratorConfig() error = %v", err)
	}

	if config.Rite.Name != "review" {
		t.Errorf("Rite.Name = %q, want %q", config.Rite.Name, "review")
	}
	if config.Rite.Color != "cyan" {
		t.Errorf("Rite.Color = %q, want %q", config.Rite.Color, "cyan")
	}
	if config.Frontmatter.Description == "" {
		t.Error("Frontmatter.Description should not be empty")
	}
	if len(config.Routing) == 0 {
		t.Error("Routing should not be empty")
	}
	if _, ok := config.Routing["signal-sifter"]; !ok {
		t.Error("Routing should contain signal-sifter")
	}
	if len(config.HandoffCriteria) == 0 {
		t.Error("HandoffCriteria should not be empty")
	}
	if _, ok := config.HandoffCriteria["scan"]; !ok {
		t.Error("HandoffCriteria should contain scan phase")
	}
	if len(config.Antipatterns) == 0 {
		t.Error("Antipatterns should not be empty")
	}
	if config.CrossRiteProtocol == "" {
		t.Error("CrossRiteProtocol should not be empty")
	}
}

func TestLoadOrchestratorConfig_ClinicYAML(t *testing.T) {
	t.Parallel()
	root := projectRoot(t)
	riteFS := os.DirFS(filepath.Join(root, "rites", "clinic"))

	config, err := loadOrchestratorConfig(riteFS)
	if err != nil {
		t.Fatalf("loadOrchestratorConfig() error = %v", err)
	}

	if config.Rite.Name != "clinic" {
		t.Errorf("Rite.Name = %q, want %q", config.Rite.Name, "clinic")
	}
	if config.Rite.Domain != "debugging and investigation" {
		t.Errorf("Rite.Domain = %q, want %q", config.Rite.Domain, "debugging and investigation")
	}
	if len(config.Routing) != 4 {
		t.Errorf("Routing length = %d, want 4", len(config.Routing))
	}
	if len(config.HandoffCriteria) != 4 {
		t.Errorf("HandoffCriteria length = %d, want 4", len(config.HandoffCriteria))
	}
	if config.WorkflowPosition.Upstream == "" {
		t.Error("WorkflowPosition.Upstream should not be empty")
	}
}

func TestLoadOrchestratorConfig_MissingFile(t *testing.T) {
	t.Parallel()
	riteFS := fstest.MapFS{} // empty FS

	config, err := loadOrchestratorConfig(riteFS)
	if err == nil {
		t.Fatal("expected error for missing orchestrator.yaml")
	}
	if config != nil {
		t.Error("config should be nil on error")
	}
}

func TestLoadOrchestratorConfig_InvalidYAML(t *testing.T) {
	t.Parallel()
	riteFS := fstest.MapFS{
		"orchestrator.yaml": &fstest.MapFile{Data: []byte("{{invalid yaml")},
	}

	config, err := loadOrchestratorConfig(riteFS)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if config != nil {
		t.Error("config should be nil on error")
	}
}

func TestConvertOrchestratorConfigToData_FullMapping(t *testing.T) {
	t.Parallel()
	config := &OrchestratorConfig{
		Skills:            []string{"orchestrator-templates", "test-ref"},
		Antipatterns:      []string{"Pattern A", "Pattern B"},
		CrossRiteProtocol: "Route to security for auth issues",
	}
	config.Rite.Name = "test"
	config.Rite.Color = "green"
	config.Frontmatter.Description = "Test description"
	config.Routing = map[string]string{
		"analyst": "Requirements needed",
		"builder": "Design complete",
	}
	config.HandoffCriteria = map[string][]string{
		"analyst": {"Requirements doc complete", "Acceptance criteria defined"},
		"builder": {"Code complete", "Tests pass"},
	}
	config.WorkflowPosition.Upstream = "User request"
	config.WorkflowPosition.Downstream = "10x-dev for implementation"

	data := convertOrchestratorConfigToData(config)

	// Direct fields
	if data["description"] != "Test description" {
		t.Errorf("description = %q, want %q", data["description"], "Test description")
	}
	if data["color"] != "green" {
		t.Errorf("color = %q, want %q", data["color"], "green")
	}
	if data["cross_rite_protocol"] != "Route to security for auth issues" {
		t.Errorf("cross_rite_protocol = %q", data["cross_rite_protocol"])
	}

	// Skills ([]any for yaml compat)
	skills, ok := data["skills"].([]any)
	if !ok || len(skills) != 2 {
		t.Fatalf("skills type/len wrong: %T %v", data["skills"], data["skills"])
	}
	if skills[0] != "orchestrator-templates" || skills[1] != "test-ref" {
		t.Errorf("skills = %v", skills)
	}

	// Phase routing (sorted by key)
	routing, ok := data["phase_routing"].(string)
	if !ok {
		t.Fatalf("phase_routing type wrong: %T", data["phase_routing"])
	}
	if !strings.Contains(routing, "| analyst | Requirements needed |") {
		t.Errorf("phase_routing missing analyst row:\n%s", routing)
	}
	if !strings.Contains(routing, "| builder | Design complete |") {
		t.Errorf("phase_routing missing builder row:\n%s", routing)
	}
	// Verify sorted order: analyst before builder
	analystIdx := strings.Index(routing, "analyst")
	builderIdx := strings.Index(routing, "builder")
	if analystIdx > builderIdx {
		t.Error("phase_routing should have analyst before builder (sorted)")
	}

	// Handoff criteria
	handoff, ok := data["handoff_criteria"].(string)
	if !ok {
		t.Fatalf("handoff_criteria type wrong: %T", data["handoff_criteria"])
	}
	if !strings.Contains(handoff, "| analyst | - Requirements doc complete<- Acceptance criteria defined< |") {
		t.Errorf("handoff_criteria format wrong:\n%s", handoff)
	}

	// Antipatterns
	anti, ok := data["rite_anti_patterns"].(string)
	if !ok {
		t.Fatalf("rite_anti_patterns type wrong: %T", data["rite_anti_patterns"])
	}
	if !strings.Contains(anti, "- **Pattern A**") {
		t.Errorf("rite_anti_patterns missing Pattern A:\n%s", anti)
	}

	// Position in workflow
	pos, ok := data["position_in_workflow"].(string)
	if !ok {
		t.Fatalf("position_in_workflow type wrong: %T", data["position_in_workflow"])
	}
	if !strings.Contains(pos, "**Upstream**: User request") {
		t.Errorf("position_in_workflow missing upstream:\n%s", pos)
	}
	if !strings.Contains(pos, "**Downstream**: 10x-dev for implementation") {
		t.Errorf("position_in_workflow missing downstream:\n%s", pos)
	}
}

func TestConvertOrchestratorConfigToData_RoutingTableSorted(t *testing.T) {
	t.Parallel()
	config := &OrchestratorConfig{}
	config.Routing = map[string]string{
		"zebra":    "Last alphabetically",
		"alpha":    "First alphabetically",
		"middle":   "In between",
	}

	data := convertOrchestratorConfigToData(config)
	routing := data["phase_routing"].(string)

	alphaIdx := strings.Index(routing, "alpha")
	middleIdx := strings.Index(routing, "middle")
	zebraIdx := strings.Index(routing, "zebra")

	if alphaIdx > middleIdx || middleIdx > zebraIdx {
		t.Errorf("routing not sorted alphabetically:\n%s", routing)
	}
}

func TestConvertOrchestratorConfigToData_HandoffCriteriaOrdering(t *testing.T) {
	t.Parallel()
	config := &OrchestratorConfig{}
	config.Routing = map[string]string{
		"beta":  "Second",
		"alpha": "First",
	}
	config.HandoffCriteria = map[string][]string{
		"alpha":   {"Item A1"},
		"beta":    {"Item B1", "Item B2"},
		"orphan":  {"Item O1"}, // not in routing — appended at end
	}

	data := convertOrchestratorConfigToData(config)
	handoff := data["handoff_criteria"].(string)

	alphaIdx := strings.Index(handoff, "| alpha |")
	betaIdx := strings.Index(handoff, "| beta |")
	orphanIdx := strings.Index(handoff, "| orphan |")

	if alphaIdx > betaIdx {
		t.Error("alpha should come before beta (sorted routing keys)")
	}
	if betaIdx > orphanIdx {
		t.Error("orphan should come after routing keys")
	}

	// Multi-item format
	if !strings.Contains(handoff, "| beta | - Item B1<- Item B2< |") {
		t.Errorf("multi-item format wrong:\n%s", handoff)
	}
}

func TestConvertOrchestratorConfigToData_EmptyFields(t *testing.T) {
	t.Parallel()
	config := &OrchestratorConfig{}

	data := convertOrchestratorConfigToData(config)

	if len(data) != 0 {
		t.Errorf("expected empty data for empty config, got %d keys: %v", len(data), data)
	}
}

func TestEnrichArchetypeData_ConfigOnly(t *testing.T) {
	t.Parallel()
	manifest := &RiteManifest{
		Name: "test-rite",
		Agents: []Agent{
			{Name: "potnia", Archetype: "orchestrator"},
		},
		// No ArchetypeData — this is the real-world scenario
	}

	riteFS := fstest.MapFS{
		"orchestrator.yaml": &fstest.MapFile{Data: []byte(`
rite:
  name: test-rite
  color: purple
frontmatter:
  description: "Test orchestrator"
routing:
  worker: "Work needed"
handoff_criteria:
  worker:
    - "Work complete"
antipatterns:
  - "Doing work yourself"
cross_rite_protocol: "Route to 10x-dev"
`)},
	}

	enrichArchetypeData(manifest, riteFS)

	if manifest.ArchetypeData == nil {
		t.Fatal("ArchetypeData should be populated from config file")
	}
	data, ok := manifest.ArchetypeData["orchestrator"]
	if !ok {
		t.Fatal("ArchetypeData[orchestrator] should exist")
	}
	if data["color"] != "purple" {
		t.Errorf("color = %q, want %q", data["color"], "purple")
	}
	if data["description"] != "Test orchestrator" {
		t.Errorf("description = %q, want %q", data["description"], "Test orchestrator")
	}
	if !strings.Contains(data["phase_routing"].(string), "| worker |") {
		t.Error("phase_routing should contain worker")
	}
}

func TestEnrichArchetypeData_ManifestOverridesConfig(t *testing.T) {
	t.Parallel()
	manifest := &RiteManifest{
		Name: "test-rite",
		Agents: []Agent{
			{Name: "potnia", Archetype: "orchestrator"},
		},
		ArchetypeData: map[string]map[string]any{
			"orchestrator": {
				"color": "red", // Manifest value should win
			},
		},
	}

	riteFS := fstest.MapFS{
		"orchestrator.yaml": &fstest.MapFile{Data: []byte(`
rite:
  color: blue
frontmatter:
  description: "From config"
`)},
	}

	enrichArchetypeData(manifest, riteFS)

	data := manifest.ArchetypeData["orchestrator"]
	if data["color"] != "red" {
		t.Errorf("color = %q, want %q (manifest should override config)", data["color"], "red")
	}
	// Description comes from config since manifest didn't set it
	if data["description"] != "From config" {
		t.Errorf("description = %q, want %q (config should fill gaps)", data["description"], "From config")
	}
}

func TestEnrichArchetypeData_NoArchetypeAgent(t *testing.T) {
	t.Parallel()
	manifest := &RiteManifest{
		Name: "plain-rite",
		Agents: []Agent{
			{Name: "worker", Role: "Works"}, // No archetype
		},
	}

	riteFS := fstest.MapFS{
		"orchestrator.yaml": &fstest.MapFile{Data: []byte("rite:\n  name: plain\n")},
	}

	enrichArchetypeData(manifest, riteFS)

	if manifest.ArchetypeData != nil {
		t.Error("ArchetypeData should remain nil when no archetype agents exist")
	}
}

func TestEnrichArchetypeData_NilFS(t *testing.T) {
	t.Parallel()
	manifest := &RiteManifest{
		Name: "test",
		Agents: []Agent{
			{Name: "potnia", Archetype: "orchestrator"},
		},
	}

	enrichArchetypeData(manifest, nil)

	if manifest.ArchetypeData != nil {
		t.Error("ArchetypeData should remain nil with nil FS")
	}
}
