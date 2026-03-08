package registry

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// writeManifest serializes m as YAML into ritePath/manifest.yaml.
func writeManifest(t *testing.T, ritePath string, m riteManifest) {
	t.Helper()
	data, err := yaml.Marshal(m)
	if err != nil {
		t.Fatalf("writeManifest: marshal failed: %v", err)
	}
	manifestPath := filepath.Join(ritePath, "manifest.yaml")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("writeManifest: write failed: %v", err)
	}
}

// writeFile creates a file at path, creating parent directories as needed.
func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("writeFile: mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("# placeholder\n"), 0644); err != nil {
		t.Fatalf("writeFile: write failed: %v", err)
	}
}

// TestValidateRiteReferences_ValidRite verifies that a fully populated rite
// produces zero warnings.
func TestValidateRiteReferences_ValidRite(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:       "test-rite",
		EntryAgent: "potnia",
		Agents: []manifestAgent{
			{Name: "potnia"},
			{Name: "analyst"},
		},
		Legomena: []string{"conventions"},
		Dromena:  []string{"go"},
	}
	writeManifest(t, dir, m)

	// Create agent files.
	writeFile(t, filepath.Join(dir, "agents", "potnia.md"))
	writeFile(t, filepath.Join(dir, "agents", "analyst.md"))

	// Create legomena (directory-based INDEX pattern).
	writeFile(t, filepath.Join(dir, "mena", "conventions", "INDEX.lego.md"))

	// Create dromena (directory-based INDEX pattern).
	writeFile(t, filepath.Join(dir, "mena", "go", "INDEX.dro.md"))

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_MissingAgentFile verifies that a declared agent
// with no corresponding .md file produces exactly 1 warning.
func TestValidateRiteReferences_MissingAgentFile(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name: "test-rite",
		Agents: []manifestAgent{
			{Name: "potnia"},
			{Name: "ghost"}, // no file will be created
		},
	}
	writeManifest(t, dir, m)

	// Create only potnia, not ghost.
	writeFile(t, filepath.Join(dir, "agents", "potnia.md"))

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d: %+v", len(warnings), warnings)
	}
	if len(warnings) > 0 && warnings[0].RefName != "ghost" {
		t.Errorf("expected warning for 'ghost', got RefName=%q", warnings[0].RefName)
	}
}

// TestValidateRiteReferences_MissingLegomena verifies that a declared legomena
// with no INDEX.lego.md file produces a warning.
func TestValidateRiteReferences_MissingLegomena(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:     "test-rite",
		Legomena: []string{"missing-skill"},
	}
	writeManifest(t, dir, m)

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d: %+v", len(warnings), warnings)
	}
	if len(warnings) > 0 && warnings[0].RefName != "missing-skill" {
		t.Errorf("expected warning for 'missing-skill', got RefName=%q", warnings[0].RefName)
	}
}

// TestValidateRiteReferences_NoManifest verifies that a missing manifest.yaml
// returns an error rather than a warning.
func TestValidateRiteReferences_NoManifest(t *testing.T) {
	dir := t.TempDir()
	// Intentionally do NOT write manifest.yaml.

	_, err := ValidateRiteReferences(dir, "", "")
	if err == nil {
		t.Error("expected error for missing manifest, got nil")
	}
}

// TestValidateRiteReferences_EntryAgentNotInList verifies that an entry_agent
// value not present in the agents list produces a warning.
func TestValidateRiteReferences_EntryAgentNotInList(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:       "test-rite",
		EntryAgent: "orchestrator", // not in agents list
		Agents: []manifestAgent{
			{Name: "analyst"},
		},
	}
	writeManifest(t, dir, m)

	// Create the declared agent file to avoid an extra warning.
	writeFile(t, filepath.Join(dir, "agents", "analyst.md"))

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundEntryWarning := false
	for _, w := range warnings {
		if w.RefName == "orchestrator" {
			foundEntryWarning = true
		}
	}
	if !foundEntryWarning {
		t.Errorf("expected warning for entry_agent 'orchestrator' not in agents list, got: %+v", warnings)
	}
}

// TestValidateRiteReferences_DromeneFlatPattern verifies that the flat
// mena/{name}.dro.md pattern is also accepted (not just the INDEX variant).
func TestValidateRiteReferences_DromeneFlatPattern(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:    "test-rite",
		Dromena: []string{"park"},
	}
	writeManifest(t, dir, m)

	// Use flat file pattern instead of directory INDEX.
	writeFile(t, filepath.Join(dir, "mena", "park.dro.md"))

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for flat dromena pattern, got %d: %+v", len(warnings), warnings)
	}
}

// --- Multi-source mena resolution tests ---

// makeRitesDir creates a rites/ base with a rite subdirectory and returns
// (ritePath, ritesBase).
func makeRitesDir(t *testing.T, riteName string) (string, string) {
	t.Helper()
	base := t.TempDir()
	ritesBase := filepath.Join(base, "rites")
	ritePath := filepath.Join(ritesBase, riteName)
	if err := os.MkdirAll(ritePath, 0755); err != nil {
		t.Fatalf("makeRitesDir: %v", err)
	}
	return ritePath, ritesBase
}

// TestValidateRiteReferences_LegomenaResolvedFromShared verifies that a
// legomena declared in the rite manifest resolves from rites/shared/mena/.
func TestValidateRiteReferences_LegomenaResolvedFromShared(t *testing.T) {
	ritePath, ritesBase := makeRitesDir(t, "myrite")

	m := riteManifest{
		Name:         "myrite",
		Legomena:     []string{"shared-skill"},
		Dependencies: []string{"shared"},
	}
	writeManifest(t, ritePath, m)

	// Skill exists in shared, not rite-local.
	writeFile(t, filepath.Join(ritesBase, "shared", "mena", "shared-skill", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(ritePath, ritesBase, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings (resolved from shared), got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_LegomenaResolvedFromDependency verifies that a
// legomena resolves from a dependency rite's mena directory.
func TestValidateRiteReferences_LegomenaResolvedFromDependency(t *testing.T) {
	ritePath, ritesBase := makeRitesDir(t, "consumer")

	m := riteManifest{
		Name:         "consumer",
		Legomena:     []string{"provider-ref"},
		Dependencies: []string{"shared", "provider"},
	}
	writeManifest(t, ritePath, m)

	// Skill exists in provider dependency, not rite-local or shared.
	writeFile(t, filepath.Join(ritesBase, "provider", "mena", "provider-ref", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(ritePath, ritesBase, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings (resolved from dependency), got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_DromenaResolvedFromShared verifies that dromena
// (both INDEX and flat patterns) resolve from shared mena.
func TestValidateRiteReferences_DromenaResolvedFromShared(t *testing.T) {
	ritePath, ritesBase := makeRitesDir(t, "myrite")

	m := riteManifest{
		Name:         "myrite",
		Dromena:      []string{"index-cmd", "flat-cmd"},
		Dependencies: []string{"shared"},
	}
	writeManifest(t, ritePath, m)

	// INDEX pattern in shared.
	writeFile(t, filepath.Join(ritesBase, "shared", "mena", "index-cmd", "INDEX.dro.md"))
	// Flat pattern in shared.
	writeFile(t, filepath.Join(ritesBase, "shared", "mena", "flat-cmd.dro.md"))

	warnings, err := ValidateRiteReferences(ritePath, ritesBase, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings (dromena in shared), got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_MissingInAllSources verifies that a legomena
// missing from rite-local, shared, and all dependencies still produces a warning.
func TestValidateRiteReferences_MissingInAllSources(t *testing.T) {
	ritePath, ritesBase := makeRitesDir(t, "myrite")

	m := riteManifest{
		Name:         "myrite",
		Legomena:     []string{"nonexistent"},
		Dependencies: []string{"shared"},
	}
	writeManifest(t, ritePath, m)

	// Create shared mena dir but NOT the skill.
	if err := os.MkdirAll(filepath.Join(ritesBase, "shared", "mena"), 0755); err != nil {
		t.Fatal(err)
	}

	warnings, err := ValidateRiteReferences(ritePath, ritesBase, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d: %+v", len(warnings), warnings)
	}
	if len(warnings) > 0 && warnings[0].RefName != "nonexistent" {
		t.Errorf("expected warning for 'nonexistent', got RefName=%q", warnings[0].RefName)
	}
}

// TestValidateRiteReferences_EmptyRitesBaseFallback verifies that an empty
// ritesBase degrades to rite-local-only checking.
func TestValidateRiteReferences_EmptyRitesBaseFallback(t *testing.T) {
	ritePath, ritesBase := makeRitesDir(t, "myrite")

	m := riteManifest{
		Name:         "myrite",
		Legomena:     []string{"shared-only"},
		Dependencies: []string{"shared"},
	}
	writeManifest(t, ritePath, m)

	// Skill exists in shared but we pass empty ritesBase.
	writeFile(t, filepath.Join(ritesBase, "shared", "mena", "shared-only", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(ritePath, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning (empty ritesBase = rite-local only), got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_NonexistentDependencyDir verifies graceful
// handling when a declared dependency directory does not exist on disk.
func TestValidateRiteReferences_NonexistentDependencyDir(t *testing.T) {
	ritePath, ritesBase := makeRitesDir(t, "myrite")

	m := riteManifest{
		Name:         "myrite",
		Legomena:     []string{"local-skill"},
		Dependencies: []string{"shared", "phantom"},
	}
	writeManifest(t, ritePath, m)

	// Skill exists rite-locally. "phantom" dependency dir does NOT exist.
	writeFile(t, filepath.Join(ritePath, "mena", "local-skill", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(ritePath, ritesBase, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings (resolved rite-locally despite phantom dep), got %d: %+v", len(warnings), warnings)
	}
}

// --- Agent skill reference validation tests ---

// writeAgentWithSkills creates an agent .md file with frontmatter containing skills.
func writeAgentWithSkills(t *testing.T, agentDir string, name string, skills []string) {
	t.Helper()
	var buf bytes.Buffer
	buf.WriteString("---\nname: " + name + "\ndescription: test agent\n")
	if len(skills) > 0 {
		buf.WriteString("skills:\n")
		for _, s := range skills {
			buf.WriteString("  - " + s + "\n")
		}
	}
	buf.WriteString("---\n")
	path := filepath.Join(agentDir, name+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("writeAgentWithSkills: mkdir: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("writeAgentWithSkills: write: %v", err)
	}
}

// TestValidateRiteReferences_AgentSkillsValid verifies that agent skills
// that resolve to existing legomena produce no warnings.
func TestValidateRiteReferences_AgentSkillsValid(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:   "test-rite",
		Agents: []manifestAgent{{Name: "analyst"}},
	}
	writeManifest(t, dir, m)
	writeAgentWithSkills(t, filepath.Join(dir, "agents"), "analyst", []string{"conventions"})
	writeFile(t, filepath.Join(dir, "mena", "conventions", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_AgentSkillsMissing verifies that an agent
// declaring a non-existent skill produces a warning.
func TestValidateRiteReferences_AgentSkillsMissing(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:   "test-rite",
		Agents: []manifestAgent{{Name: "analyst"}},
	}
	writeManifest(t, dir, m)
	writeAgentWithSkills(t, filepath.Join(dir, "agents"), "analyst", []string{"ghost-skill"})

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, w := range warnings {
		if w.RefName == "ghost-skill" && w.File == filepath.Join("agents", "analyst.md") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for ghost-skill in agents/analyst.md, got: %+v", warnings)
	}
}

// TestValidateRiteReferences_AgentDefaultsSkillsMissing verifies that
// agent_defaults.skills with a non-existent skill produces a warning.
func TestValidateRiteReferences_AgentDefaultsSkillsMissing(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:          "test-rite",
		AgentDefaults: manifestAgentDefaults{Skills: []string{"ghost-default"}},
	}
	writeManifest(t, dir, m)

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, w := range warnings {
		if w.RefName == "ghost-default" && w.File == "manifest.yaml" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for ghost-default in manifest.yaml, got: %+v", warnings)
	}
}

// TestValidateRiteReferences_AgentDefaultsSkillsValid verifies that
// agent_defaults.skills with existing legomena produce no skill warnings.
func TestValidateRiteReferences_AgentDefaultsSkillsValid(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:          "test-rite",
		AgentDefaults: manifestAgentDefaults{Skills: []string{"valid-default"}},
	}
	writeManifest(t, dir, m)
	writeFile(t, filepath.Join(dir, "mena", "valid-default", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_SkillPoliciesMissing verifies that a
// skill_policies entry referencing a non-existent skill produces a warning.
func TestValidateRiteReferences_SkillPoliciesMissing(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:          "test-rite",
		SkillPolicies: []manifestSkillPolicy{{Skill: "ghost-policy"}},
	}
	writeManifest(t, dir, m)

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, w := range warnings {
		if w.RefName == "ghost-policy" && w.File == "manifest.yaml" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for ghost-policy in manifest.yaml, got: %+v", warnings)
	}
}

// TestValidateRiteReferences_SkillPoliciesValid verifies that a
// skill_policies entry referencing an existing legomena produces no warning.
func TestValidateRiteReferences_SkillPoliciesValid(t *testing.T) {
	dir := t.TempDir()

	m := riteManifest{
		Name:          "test-rite",
		SkillPolicies: []manifestSkillPolicy{{Skill: "valid-policy"}},
	}
	writeManifest(t, dir, m)
	writeFile(t, filepath.Join(dir, "mena", "valid-policy", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_NestedSkillPath verifies that agent skills
// with nested paths (e.g., "guidance/file-verification") resolve correctly
// via the platform mena source.
func TestValidateRiteReferences_NestedSkillPath(t *testing.T) {
	dir := t.TempDir()
	platformMena := t.TempDir()

	m := riteManifest{
		Name:   "test-rite",
		Agents: []manifestAgent{{Name: "architect"}},
	}
	writeManifest(t, dir, m)
	writeAgentWithSkills(t, filepath.Join(dir, "agents"), "architect", []string{"guidance/file-verification"})

	// Nested skill exists in platform mena.
	writeFile(t, filepath.Join(platformMena, "guidance", "file-verification", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(dir, "", platformMena)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for nested skill path, got %d: %+v", len(warnings), warnings)
	}
}

// TestValidateRiteReferences_AgentSkillResolvedFromShared verifies that
// agent skills resolve from the shared mena source.
func TestValidateRiteReferences_AgentSkillResolvedFromShared(t *testing.T) {
	ritePath, ritesBase := makeRitesDir(t, "myrite")

	m := riteManifest{
		Name:         "myrite",
		Agents:       []manifestAgent{{Name: "analyst"}},
		Dependencies: []string{"shared"},
	}
	writeManifest(t, ritePath, m)
	writeAgentWithSkills(t, filepath.Join(ritePath, "agents"), "analyst", []string{"shared-skill"})

	// Skill exists in shared mena, not rite-local.
	writeFile(t, filepath.Join(ritesBase, "shared", "mena", "shared-skill", "INDEX.lego.md"))

	warnings, err := ValidateRiteReferences(ritePath, ritesBase, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings (skill in shared), got %d: %+v", len(warnings), warnings)
	}
}
