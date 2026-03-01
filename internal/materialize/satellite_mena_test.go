package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// setupKnossosRiteSharedMena creates a minimal knossos-style rites/ structure
// at the given knossosHome with a shared rite containing one dromenon and one legomenon.
// Returns the path to the shared mena directory.
func setupKnossosRiteSharedMena(t *testing.T, knossosHome string) {
	t.Helper()
	sharedMena := filepath.Join(knossosHome, "rites", "shared", "mena")
	droDir := filepath.Join(sharedMena, "shared-cmd")
	legoDir := filepath.Join(sharedMena, "shared-skill")

	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("setupKnossosRiteSharedMena: mkdir dro: %v", err)
	}
	if err := os.MkdirAll(legoDir, 0755); err != nil {
		t.Fatalf("setupKnossosRiteSharedMena: mkdir lego: %v", err)
	}
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: shared-cmd\ndescription: shared command\n---\n# Shared Command\n"), 0644); err != nil {
		t.Fatalf("setupKnossosRiteSharedMena: write dro: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), []byte("---\nname: shared-skill\ndescription: shared skill\n---\n# Shared Skill\n"), 0644); err != nil {
		t.Fatalf("setupKnossosRiteSharedMena: write lego: %v", err)
	}
}

// setupKnossosRiteDepMena creates a dependency rite mena directory inside
// knossosHome/rites/{depName}/mena/ with one legomenon.
func setupKnossosRiteDepMena(t *testing.T, knossosHome, depName string) {
	t.Helper()
	depMena := filepath.Join(knossosHome, "rites", depName, "mena")
	legoDir := filepath.Join(depMena, "dep-skill")

	if err := os.MkdirAll(legoDir, 0755); err != nil {
		t.Fatalf("setupKnossosRiteDepMena: mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), []byte("---\nname: dep-skill\ndescription: dep skill\n---\n# Dep Skill\n"), 0644); err != nil {
		t.Fatalf("setupKnossosRiteDepMena: write: %v", err)
	}
}

// setupSatelliteRite creates a minimal satellite project with a local rite.
// Returns the satellite project root directory.
func setupSatelliteRite(t *testing.T, riteName string) string {
	t.Helper()
	satelliteRoot := t.TempDir()
	riteDir := filepath.Join(satelliteRoot, "rites", riteName)
	agentsDir := filepath.Join(riteDir, "agents")

	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("setupSatelliteRite: mkdir agents: %v", err)
	}
	manifest := "name: " + riteName + "\nversion: \"1.0.0\"\ndescription: Test satellite rite\nentry_agent: analyst\nagents:\n  - name: analyst\n    role: Data analyst\n"
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("setupSatelliteRite: write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "analyst.md"), []byte("# Analyst\nYou are a data analyst.\n"), 0644); err != nil {
		t.Fatalf("setupSatelliteRite: write agent: %v", err)
	}
	return satelliteRoot
}

// TestMaterializeMena_SatelliteLocalRite_SharedMenaResolvesFromKnossosHome verifies
// that shared rite mena (dromena and legomena) is materialized from $KNOSSOS_HOME/rites/shared/mena/
// when the active rite is a satellite-local rite (lives in the satellite's rites/ dir).
//
// This test covers the primary gap: satellite-local rites previously resolved shared mena
// from {satellite}/rites/shared/mena/ which does not exist, silently dropping all 12 shared entries.
func TestMaterializeMena_SatelliteLocalRite_SharedMenaResolvesFromKnossosHome(t *testing.T) {
	// Set up a fake knossos home with shared rite mena
	knossosHome := t.TempDir()
	setupKnossosRiteSharedMena(t, knossosHome)

	// Point KNOSSOS_HOME at our fake knossos directory
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", knossosHome)
	t.Cleanup(config.ResetKnossosHome)

	// Create a satellite project with a local rite (does NOT have rites/shared/)
	satelliteRoot := setupSatelliteRite(t, "data-analyst")
	claudeDir := filepath.Join(satelliteRoot, ".claude")

	// Create a resolved rite pointing to the satellite's local rite directory
	// (simulating how the source resolver resolves a satellite-local rite)
	ritePath := filepath.Join(satelliteRoot, "rites", "data-analyst")
	resolved := &ResolvedRite{
		Name: "data-analyst",
		Source: RiteSource{
			Type:        SourceProject,
			Path:        filepath.Join(satelliteRoot, "rites"),
			Description: "satellite project rites",
		},
		RitePath:     ritePath,
		ManifestPath: filepath.Join(ritePath, "manifest.yaml"),
	}

	manifest := &RiteManifest{
		Name:    "data-analyst",
		Version: "1.0.0",
	}

	// Set up materializer with the satellite project as root
	// The source resolver is built from satelliteRoot, so KnossosHome() will pick up
	// the env var we just set.
	resolver := paths.NewResolver(satelliteRoot)
	m := NewMaterializer(resolver)

	if err := m.materializeMena(manifest, claudeDir, resolved, provenance.NullCollector{}, false); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify: shared dromenon should appear in .claude/commands/
	sharedCmd := filepath.Join(claudeDir, "commands", "shared-cmd.md")
	if _, err := os.Stat(sharedCmd); os.IsNotExist(err) {
		t.Errorf("FAIL: shared dromenon not materialized; expected at %s", sharedCmd)
	}

	// Verify: shared legomenon should appear in .claude/skills/
	sharedSkill := filepath.Join(claudeDir, "skills", "shared-skill", "SKILL.md")
	if _, err := os.Stat(sharedSkill); os.IsNotExist(err) {
		t.Errorf("FAIL: shared legomenon not materialized; expected at %s", sharedSkill)
	}
}

// TestMaterializeMena_SatelliteLocalRite_DependencyMenaResolvesFromKnossosHome verifies
// that dependency rite mena resolves from $KNOSSOS_HOME/rites/{dep}/mena/ when the
// active rite declares dependencies on knossos-core rites.
func TestMaterializeMena_SatelliteLocalRite_DependencyMenaResolvesFromKnossosHome(t *testing.T) {
	knossosHome := t.TempDir()
	setupKnossosRiteSharedMena(t, knossosHome)
	setupKnossosRiteDepMena(t, knossosHome, "10x-dev")

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", knossosHome)
	t.Cleanup(config.ResetKnossosHome)

	satelliteRoot := setupSatelliteRite(t, "data-analyst")
	claudeDir := filepath.Join(satelliteRoot, ".claude")

	ritePath := filepath.Join(satelliteRoot, "rites", "data-analyst")
	resolved := &ResolvedRite{
		Name: "data-analyst",
		Source: RiteSource{
			Type:        SourceProject,
			Path:        filepath.Join(satelliteRoot, "rites"),
			Description: "satellite project rites",
		},
		RitePath:     ritePath,
		ManifestPath: filepath.Join(ritePath, "manifest.yaml"),
	}

	manifest := &RiteManifest{
		Name:         "data-analyst",
		Version:      "1.0.0",
		Dependencies: []string{"10x-dev"},
	}

	resolver := paths.NewResolver(satelliteRoot)
	m := NewMaterializer(resolver)

	if err := m.materializeMena(manifest, claudeDir, resolved, provenance.NullCollector{}, false); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Shared mena must appear
	sharedCmd := filepath.Join(claudeDir, "commands", "shared-cmd.md")
	if _, err := os.Stat(sharedCmd); os.IsNotExist(err) {
		t.Errorf("FAIL: shared dromenon not materialized; expected at %s", sharedCmd)
	}

	// Dependency mena must appear
	depSkill := filepath.Join(claudeDir, "skills", "dep-skill", "SKILL.md")
	if _, err := os.Stat(depSkill); os.IsNotExist(err) {
		t.Errorf("FAIL: dependency legomenon not materialized; expected at %s", depSkill)
	}
}

// TestMaterializeMena_SatelliteLocalRite_RiteMenaOverridesShared verifies that
// rite-local mena overrides same-named shared mena (rite > shared priority).
func TestMaterializeMena_SatelliteLocalRite_RiteMenaOverridesShared(t *testing.T) {
	knossosHome := t.TempDir()
	setupKnossosRiteSharedMena(t, knossosHome)

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", knossosHome)
	t.Cleanup(config.ResetKnossosHome)

	satelliteRoot := setupSatelliteRite(t, "data-analyst")
	claudeDir := filepath.Join(satelliteRoot, ".claude")

	// Create a rite-local mena entry with the same name as the shared one ("shared-cmd")
	// but different content — the rite version should win.
	riteLocalMenaDir := filepath.Join(satelliteRoot, "rites", "data-analyst", "mena", "shared-cmd")
	if err := os.MkdirAll(riteLocalMenaDir, 0755); err != nil {
		t.Fatalf("mkdir rite mena: %v", err)
	}
	if err := os.WriteFile(filepath.Join(riteLocalMenaDir, "INDEX.dro.md"), []byte("---\nname: shared-cmd\ndescription: rite override\n---\n# Rite Override\n"), 0644); err != nil {
		t.Fatalf("write rite mena: %v", err)
	}

	ritePath := filepath.Join(satelliteRoot, "rites", "data-analyst")
	resolved := &ResolvedRite{
		Name: "data-analyst",
		Source: RiteSource{
			Type:        SourceProject,
			Path:        filepath.Join(satelliteRoot, "rites"),
			Description: "satellite project rites",
		},
		RitePath:     ritePath,
		ManifestPath: filepath.Join(ritePath, "manifest.yaml"),
	}

	manifest := &RiteManifest{
		Name:    "data-analyst",
		Version: "1.0.0",
	}

	resolver := paths.NewResolver(satelliteRoot)
	m := NewMaterializer(resolver)

	if err := m.materializeMena(manifest, claudeDir, resolved, provenance.NullCollector{}, false); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify: rite-local version of shared-cmd must be projected
	cmdFile := filepath.Join(claudeDir, "commands", "shared-cmd.md")
	content, err := os.ReadFile(cmdFile)
	if err != nil {
		t.Fatalf("shared-cmd.md not found at %s: %v", cmdFile, err)
	}
	if string(content) != "---\nname: shared-cmd\ndescription: rite override\n---\n# Rite Override\n" {
		t.Errorf("FAIL: rite mena did not override shared mena; content = %q", string(content))
	}

	// Verify: shared legomenon should still appear (not overridden by rite)
	sharedSkill := filepath.Join(claudeDir, "skills", "shared-skill", "SKILL.md")
	if _, err := os.Stat(sharedSkill); os.IsNotExist(err) {
		t.Errorf("FAIL: shared legomenon not materialized; expected at %s", sharedSkill)
	}
}

// TestMaterializeMena_KnossosCoreSelf_NoRegression verifies that knossos syncing
// its own rites (the self-hosting case) continues to work correctly after the fix.
// In this case ritesBase == KNOSSOS_HOME/rites/ so the fix should be a no-op.
func TestMaterializeMena_KnossosCoreSelf_NoRegression(t *testing.T) {
	// Simulate the knossos self-hosting case: projectRoot == knossosHome
	// Both the rite and shared/ are in the same rites/ directory.
	knossosHome := t.TempDir()

	// Set up the knossos project with its own rites structure
	setupKnossosRiteSharedMena(t, knossosHome)

	// Create a knossos-core rite at knossosHome/rites/ecosystem/
	knossosCoreRite := filepath.Join(knossosHome, "rites", "ecosystem")
	if err := os.MkdirAll(knossosCoreRite, 0755); err != nil {
		t.Fatalf("mkdir core rite: %v", err)
	}
	coreManifestContent := "name: ecosystem\nversion: \"1.0.0\"\ndescription: Core rite\nentry_agent: pythia\nagents:\n  - name: pythia\n    role: Orchestrator\n"
	if err := os.WriteFile(filepath.Join(knossosCoreRite, "manifest.yaml"), []byte(coreManifestContent), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", knossosHome)
	t.Cleanup(config.ResetKnossosHome)

	claudeDir := filepath.Join(knossosHome, ".claude")

	// Resolved rite points to knossosHome/rites/ecosystem — filepath.Dir gives knossosHome/rites/
	// which IS the same as $KNOSSOS_HOME/rites, so sharedRitesBase should stay as ritesBase.
	ritePath := knossosCoreRite
	resolved := &ResolvedRite{
		Name: "ecosystem",
		Source: RiteSource{
			Type:        SourceProject,
			Path:        filepath.Join(knossosHome, "rites"),
			Description: "knossos project rites",
		},
		RitePath:     ritePath,
		ManifestPath: filepath.Join(ritePath, "manifest.yaml"),
	}

	manifest := &RiteManifest{
		Name:    "ecosystem",
		Version: "1.0.0",
	}

	resolver := paths.NewResolver(knossosHome)
	m := NewMaterializer(resolver)

	if err := m.materializeMena(manifest, claudeDir, resolved, provenance.NullCollector{}, false); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Shared mena must still appear for knossos-core rites
	sharedCmd := filepath.Join(claudeDir, "commands", "shared-cmd.md")
	if _, err := os.Stat(sharedCmd); os.IsNotExist(err) {
		t.Errorf("REGRESSION: shared dromenon not materialized for knossos-core rite; expected at %s", sharedCmd)
	}

	sharedSkill := filepath.Join(claudeDir, "skills", "shared-skill", "SKILL.md")
	if _, err := os.Stat(sharedSkill); os.IsNotExist(err) {
		t.Errorf("REGRESSION: shared legomenon not materialized for knossos-core rite; expected at %s", sharedSkill)
	}
}
