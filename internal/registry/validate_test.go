package registry

import (
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
		EntryAgent: "pythia",
		Agents: []manifestAgent{
			{Name: "pythia"},
			{Name: "analyst"},
		},
		Legomena: []string{"conventions"},
		Dromena:  []string{"go"},
	}
	writeManifest(t, dir, m)

	// Create agent files.
	writeFile(t, filepath.Join(dir, "agents", "pythia.md"))
	writeFile(t, filepath.Join(dir, "agents", "analyst.md"))

	// Create legomena (directory-based INDEX pattern).
	writeFile(t, filepath.Join(dir, "mena", "conventions", "INDEX.lego.md"))

	// Create dromena (directory-based INDEX pattern).
	writeFile(t, filepath.Join(dir, "mena", "go", "INDEX.dro.md"))

	warnings, err := ValidateRiteReferences(dir)
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
			{Name: "pythia"},
			{Name: "ghost"}, // no file will be created
		},
	}
	writeManifest(t, dir, m)

	// Create only pythia, not ghost.
	writeFile(t, filepath.Join(dir, "agents", "pythia.md"))

	warnings, err := ValidateRiteReferences(dir)
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

	warnings, err := ValidateRiteReferences(dir)
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

	_, err := ValidateRiteReferences(dir)
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

	warnings, err := ValidateRiteReferences(dir)
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

	warnings, err := ValidateRiteReferences(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for flat dromena pattern, got %d: %+v", len(warnings), warnings)
	}
}
