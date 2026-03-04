package userscope

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/autom8y/knossos/internal/materialize/mena"
	"github.com/autom8y/knossos/internal/provenance"
)

// TestCollectMena_FlattensDromena verifies that CollectMena resolves dromena
// flat names from frontmatter, flattening nested paths like operations/spike -> spike.
func TestCollectMena_FlattensDromena(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested dro: mena/operations/spike/INDEX.dro.md with name: spike
	spikeDir := filepath.Join(tmpDir, "mena", "operations", "spike")
	os.MkdirAll(spikeDir, 0755)
	os.WriteFile(filepath.Join(spikeDir, "INDEX.dro.md"), []byte("---\nname: spike\n---\n# Spike\n"), 0644)
	os.WriteFile(filepath.Join(spikeDir, "examples.md"), []byte("# Examples\n"), 0644)

	sources := []mena.MenaSource{{Path: filepath.Join(tmpDir, "mena")}}
	opts := mena.MenaProjectionOptions{Filter: mena.ProjectAll}

	resolution, err := mena.CollectMena(sources, opts)
	if err != nil {
		t.Fatalf("CollectMena failed: %v", err)
	}

	// Should be collected as directory entry with flat name "spike"
	entry, ok := resolution.Entries["operations/spike"]
	if !ok {
		t.Fatalf("Expected entry for 'operations/spike', got entries: %v", keysOf(resolution.Entries))
	}
	if entry.FlatName != "spike" {
		t.Errorf("Expected FlatName 'spike', got %q", entry.FlatName)
	}
	if entry.MenaType != "dro" {
		t.Errorf("Expected MenaType 'dro', got %q", entry.MenaType)
	}
}

// TestCollectMena_CompanionHiding verifies that non-INDEX .md files in dro
// directories get user-invocable: false injected via syncUserMena.
func TestCollectMena_CompanionHiding(t *testing.T) {
	tmpDir := t.TempDir()

	// Create dro with companion
	droDir := filepath.Join(tmpDir, "mena", "my-cmd")
	os.MkdirAll(droDir, 0755)
	os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\n---\n# Command\n"), 0644)
	os.WriteFile(filepath.Join(droDir, "examples.md"), []byte("# Examples\n"), 0644)

	// Set up target dirs
	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	commandsDir := filepath.Join(userClaudeDir, "commands")
	skillsDir := filepath.Join(userClaudeDir, "skills")
	os.MkdirAll(commandsDir, 0755)
	os.MkdirAll(skillsDir, 0755)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	collisionChecker := &CollisionChecker{}
	s := &syncer{}

	result, err := s.syncUserMena(tmpDir, userClaudeDir, manifest, collisionChecker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserMena failed: %v", err)
	}

	// Companion file should have user-invocable: false injected
	companionPath := filepath.Join(commandsDir, "my-cmd", "examples.md")
	content, err := os.ReadFile(companionPath)
	if err != nil {
		t.Fatalf("Failed to read companion file: %v", err)
	}
	if !containsBytes(content, "user-invocable: false") {
		t.Errorf("Expected companion to have 'user-invocable: false', got:\n%s", string(content))
	}

	// Promoted INDEX (now at parent level as my-cmd.md) should NOT have user-invocable: false
	promotedPath := filepath.Join(commandsDir, "my-cmd.md")
	promotedContent, err := os.ReadFile(promotedPath)
	if err != nil {
		t.Fatalf("Failed to read promoted INDEX file at %s: %v", promotedPath, err)
	}
	if containsBytes(promotedContent, "user-invocable: false") {
		t.Errorf("Promoted INDEX should NOT have 'user-invocable: false'")
	}

	if result.Summary.Added != 2 {
		t.Errorf("Expected 2 added files, got %d", result.Summary.Added)
	}
}

// TestCollectMena_LegoPreservesPath verifies that legomena paths are NOT flattened.
func TestCollectMena_LegoPreservesPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested legomena
	legoDir := filepath.Join(tmpDir, "mena", "guidance", "standards")
	os.MkdirAll(legoDir, 0755)
	os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), []byte("---\nname: standards\n---\n# Standards\n"), 0644)
	os.WriteFile(filepath.Join(legoDir, "code-conventions.md"), []byte("# Code\n"), 0644)

	sources := []mena.MenaSource{{Path: filepath.Join(tmpDir, "mena")}}
	opts := mena.MenaProjectionOptions{Filter: mena.ProjectAll}

	resolution, err := mena.CollectMena(sources, opts)
	if err != nil {
		t.Fatalf("CollectMena failed: %v", err)
	}

	entry, ok := resolution.Entries["guidance/standards"]
	if !ok {
		t.Fatalf("Expected entry for 'guidance/standards'")
	}
	// Legomena should preserve full path (no flattening)
	if entry.FlatName != "guidance/standards" {
		t.Errorf("Expected lego FlatName 'guidance/standards', got %q", entry.FlatName)
	}
	if entry.MenaType != "lego" {
		t.Errorf("Expected MenaType 'lego', got %q", entry.MenaType)
	}
}

// TestCollectMena_StandaloneFlattening verifies standalone dro files get flattened.
func TestCollectMena_StandaloneFlattening(t *testing.T) {
	tmpDir := t.TempDir()

	// Create standalone dro in a grouping directory
	groupDir := filepath.Join(tmpDir, "mena", "operations")
	os.MkdirAll(groupDir, 0755)
	os.WriteFile(filepath.Join(groupDir, "architect.dro.md"), []byte("---\nname: architect\n---\n# Architect\n"), 0644)

	sources := []mena.MenaSource{{Path: filepath.Join(tmpDir, "mena")}}
	opts := mena.MenaProjectionOptions{Filter: mena.ProjectAll}

	resolution, err := mena.CollectMena(sources, opts)
	if err != nil {
		t.Fatalf("CollectMena failed: %v", err)
	}

	sf, ok := resolution.Standalones["operations/architect.dro.md"]
	if !ok {
		t.Fatalf("Expected standalone for 'operations/architect.dro.md', got: %v", keysOfStandalone(resolution.Standalones))
	}
	// Should be flattened to just "architect.md"
	if sf.FlatName != "architect.md" {
		t.Errorf("Expected FlatName 'architect.md', got %q", sf.FlatName)
	}
	if sf.MenaType != "dro" {
		t.Errorf("Expected MenaType 'dro', got %q", sf.MenaType)
	}
}

// TestSyncUserMena_CollisionSkipsRiteContent verifies that files matching rite
// manifest entries are skipped during user-scope sync.
func TestSyncUserMena_CollisionSkipsRiteContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create dro source
	droDir := filepath.Join(tmpDir, "mena", "my-cmd")
	os.MkdirAll(droDir, 0755)
	os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\n---\n# Cmd\n"), 0644)

	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userClaudeDir, "commands"), 0755)
	os.MkdirAll(filepath.Join(userClaudeDir, "skills"), 0755)

	// Create collision checker that reports collision for commands/my-cmd.md
	// (promoted dromena INDEX.md form — rite scope also promotes)
	checker := &CollisionChecker{
		manifestLoaded: true,
		riteEntries:    map[string]bool{"commands/my-cmd.md": true},
	}

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	s := &syncer{}

	result, err := s.syncUserMena(tmpDir, userClaudeDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserMena failed: %v", err)
	}

	// Should be skipped due to collision
	if result.Summary.Skipped == 0 {
		t.Errorf("Expected skipped entries due to collision, got none")
	}
	if result.Summary.Added != 0 {
		t.Errorf("Expected 0 added (collision), got %d", result.Summary.Added)
	}
}

// TestSyncUserMena_PreservesUserOwned verifies that owner=user entries are never overwritten.
func TestSyncUserMena_PreservesUserOwned(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source
	droDir := filepath.Join(tmpDir, "mena", "my-cmd")
	os.MkdirAll(droDir, 0755)
	os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\n---\n# New Content\n"), 0644)

	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userClaudeDir, "commands"), 0755)
	os.MkdirAll(filepath.Join(userClaudeDir, "skills"), 0755)
	// Promoted dro INDEX.md lives at commands/my-cmd.md (parent level)
	promotedFile := filepath.Join(userClaudeDir, "commands", "my-cmd.md")
	os.WriteFile(promotedFile, []byte("# User Modified\n"), 0644)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"commands/my-cmd.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
		},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	_, err := s.syncUserMena(tmpDir, userClaudeDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserMena failed: %v", err)
	}

	// User content should be preserved
	content, _ := os.ReadFile(promotedFile)
	if string(content) != "# User Modified\n" {
		t.Errorf("Expected user content preserved, got %q", string(content))
	}
}

// TestSyncUserMena_Idempotent verifies that running sync twice produces zero changes.
func TestSyncUserMena_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source
	droDir := filepath.Join(tmpDir, "mena", "my-cmd")
	os.MkdirAll(droDir, 0755)
	os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\n---\n# Cmd\n"), 0644)

	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	os.MkdirAll(filepath.Join(userClaudeDir, "commands"), 0755)
	os.MkdirAll(filepath.Join(userClaudeDir, "skills"), 0755)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := &CollisionChecker{}
	s := &syncer{}

	// First sync
	result1, err := s.syncUserMena(tmpDir, userClaudeDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("First sync failed: %v", err)
	}
	if result1.Summary.Added == 0 {
		t.Errorf("First sync should add files")
	}

	// Second sync - should produce zero changes
	result2, err := s.syncUserMena(tmpDir, userClaudeDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("Second sync failed: %v", err)
	}
	if result2.Summary.Added != 0 {
		t.Errorf("Second sync should add 0 files, got %d", result2.Summary.Added)
	}
	if result2.Summary.Updated != 0 {
		t.Errorf("Second sync should update 0 files, got %d", result2.Summary.Updated)
	}
}

// TestWipeKnossosOwnedMenaEntries verifies that knossos-produced commands/
// and skills/ entries are removed (even if marked owner: user by old pipeline),
// while genuinely user-created entries are preserved.
func TestWipeKnossosOwnedMenaEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mena source: "my-cmd" dromena directory
	menaDir := filepath.Join(tmpDir, "knossos", "mena", "my-cmd")
	os.MkdirAll(menaDir, 0755)
	os.WriteFile(filepath.Join(menaDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\n---\n# Cmd\n"), 0644)

	// Create target files: knossos-produced command + genuinely user-created command + agent
	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	knossosFile := filepath.Join(userClaudeDir, "commands", "my-cmd", "INDEX.md")
	userFile := filepath.Join(userClaudeDir, "commands", "user-custom", "INDEX.md")
	agentFile := filepath.Join(userClaudeDir, "agents", "my-agent.md")
	os.MkdirAll(filepath.Dir(knossosFile), 0755)
	os.MkdirAll(filepath.Dir(userFile), 0755)
	os.MkdirAll(filepath.Dir(agentFile), 0755)
	os.WriteFile(knossosFile, []byte("# Old\n"), 0644)
	os.WriteFile(userFile, []byte("# User\n"), 0644)
	os.WriteFile(agentFile, []byte("# Agent\n"), 0644)

	// All marked owner: user (simulating old pipeline behavior)
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"commands/my-cmd/INDEX.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
			"commands/user-custom/INDEX.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
			"agents/my-agent.md": {
				Owner: provenance.OwnerKnossos,
				Scope: provenance.ScopeUser,
			},
		},
	}

	knossosHome := filepath.Join(tmpDir, "knossos")
	wipeKnossosOwnedMenaEntries(knossosHome, userClaudeDir, manifest, false)

	// Knossos-produced commands entry should be removed (even though owner: user)
	if _, exists := manifest.Entries["commands/my-cmd/INDEX.md"]; exists {
		t.Error("Expected knossos-produced commands entry to be removed")
	}
	if _, err := os.Stat(knossosFile); !os.IsNotExist(err) {
		t.Error("Expected knossos-produced commands file to be removed")
	}

	// Genuinely user-created entry should be preserved
	if _, exists := manifest.Entries["commands/user-custom/INDEX.md"]; !exists {
		t.Error("Expected user-created entry to be preserved")
	}
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		t.Error("Expected user-created file to be preserved")
	}

	// Agents entry should be preserved (not commands/ or skills/)
	if _, exists := manifest.Entries["agents/my-agent.md"]; !exists {
		t.Error("Expected agents entry to be preserved (not mena)")
	}
}

// TestWipeKnossosOwnedMenaEntries_OldStylePaths verifies that old non-flattened
// paths from the old pipeline are correctly identified and wiped.
func TestWipeKnossosOwnedMenaEntries_OldStylePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mena source with grouping: operations/spike
	spikeDir := filepath.Join(tmpDir, "knossos", "mena", "operations", "spike")
	os.MkdirAll(spikeDir, 0755)
	os.WriteFile(filepath.Join(spikeDir, "INDEX.dro.md"), []byte("---\nname: spike\n---\n# Spike\n"), 0644)

	// Create standalone: operations/architect.dro.md
	opsDir := filepath.Join(tmpDir, "knossos", "mena", "operations")
	os.WriteFile(filepath.Join(opsDir, "architect.dro.md"), []byte("---\nname: architect\n---\n# Architect\n"), 0644)

	// Create target files at OLD non-flattened paths (what old pipeline produced)
	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	oldSpikeFile := filepath.Join(userClaudeDir, "commands", "operations", "spike", "INDEX.md")
	oldArchFile := filepath.Join(userClaudeDir, "commands", "operations", "architect.md")
	os.MkdirAll(filepath.Dir(oldSpikeFile), 0755)
	os.WriteFile(oldSpikeFile, []byte("# Old Spike\n"), 0644)
	os.WriteFile(oldArchFile, []byte("# Old Architect\n"), 0644)

	// Manifest with old-style paths, all marked owner: user
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"commands/operations/spike/INDEX.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
			"commands/operations/architect.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
		},
	}

	knossosHome := filepath.Join(tmpDir, "knossos")
	wipeKnossosOwnedMenaEntries(knossosHome, userClaudeDir, manifest, false)

	// Both old-style entries should be wiped
	if _, exists := manifest.Entries["commands/operations/spike/INDEX.md"]; exists {
		t.Error("Expected old-style spike entry to be removed")
	}
	if _, exists := manifest.Entries["commands/operations/architect.md"]; exists {
		t.Error("Expected old-style architect entry to be removed")
	}

	// Files should be deleted
	if _, err := os.Stat(oldSpikeFile); !os.IsNotExist(err) {
		t.Error("Expected old-style spike file to be removed")
	}
	if _, err := os.Stat(oldArchFile); !os.IsNotExist(err) {
		t.Error("Expected old-style architect file to be removed")
	}
}

// TestWipeKnossosOwnedMenaEntries_UntrackedOrphans verifies that files on disk
// matching knossos patterns are removed even when not in the manifest.
func TestWipeKnossosOwnedMenaEntries_UntrackedOrphans(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mena source: operations/spike directory entry with name: spike
	spikeDir := filepath.Join(tmpDir, "knossos", "mena", "operations", "spike")
	os.MkdirAll(spikeDir, 0755)
	os.WriteFile(filepath.Join(spikeDir, "INDEX.dro.md"), []byte("---\nname: spike\n---\n# Spike\n"), 0644)

	// Create untracked file on disk at old-style path (NOT in manifest)
	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	orphanFile := filepath.Join(userClaudeDir, "commands", "operations", "spike.md")
	os.MkdirAll(filepath.Dir(orphanFile), 0755)
	os.WriteFile(orphanFile, []byte("# Old Spike\n"), 0644)

	// Also create a genuinely user-created file (NOT matching any knossos pattern)
	userFile := filepath.Join(userClaudeDir, "commands", "my-custom-cmd.md")
	os.WriteFile(userFile, []byte("# My Custom\n"), 0644)

	// Empty manifest — no entries tracked
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}

	knossosHome := filepath.Join(tmpDir, "knossos")
	wipeKnossosOwnedMenaEntries(knossosHome, userClaudeDir, manifest, false)

	// Untracked orphan should be removed
	if _, err := os.Stat(orphanFile); !os.IsNotExist(err) {
		t.Error("Expected untracked orphan file to be removed")
	}

	// Genuinely user-created file should survive
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		t.Error("Expected user-created file to survive wipe")
	}
}

// TestWipeKnossosOwnedMenaEntries_DryRun verifies no changes in dry-run mode.
func TestWipeKnossosOwnedMenaEntries_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mena source
	menaDir := filepath.Join(tmpDir, "knossos", "mena", "my-cmd")
	os.MkdirAll(menaDir, 0755)
	os.WriteFile(filepath.Join(menaDir, "INDEX.dro.md"), []byte("---\nname: my-cmd\n---\n# Cmd\n"), 0644)

	// Create target file
	userClaudeDir := filepath.Join(tmpDir, "user-claude")
	knossosFile := filepath.Join(userClaudeDir, "commands", "my-cmd", "INDEX.md")
	os.MkdirAll(filepath.Dir(knossosFile), 0755)
	os.WriteFile(knossosFile, []byte("# Old\n"), 0644)

	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{
			"commands/my-cmd/INDEX.md": {
				Owner: provenance.OwnerUser,
				Scope: provenance.ScopeUser,
			},
		},
	}

	knossosHome := filepath.Join(tmpDir, "knossos")
	wipeKnossosOwnedMenaEntries(knossosHome, userClaudeDir, manifest, true)

	// Manifest entry should still exist
	if _, exists := manifest.Entries["commands/my-cmd/INDEX.md"]; !exists {
		t.Error("Expected manifest entry to be preserved in dry-run")
	}

	// File should still exist
	if _, err := os.Stat(knossosFile); os.IsNotExist(err) {
		t.Error("Expected file to be preserved in dry-run")
	}
}

// Helper: get keys of resolved entries map
func keysOf(m map[string]mena.MenaResolvedEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper: get keys of resolved standalones map
func keysOfStandalone(m map[string]mena.MenaResolvedStandalone) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestSyncUserMena_IncludesSharedRiteMena verifies that syncUserMena includes
// rites/shared/mena/ as a source so cross-rite features (/know, /radar, etc.)
// are available in user scope (~/.claude/commands/).
// Regression test for: shared mena missing from user-scope sync.
func TestSyncUserMena_IncludesSharedRiteMena(t *testing.T) {
	tmpDir := t.TempDir()

	// Create platform mena (KNOSSOS_HOME/mena/)
	platformDir := filepath.Join(tmpDir, "mena", "nav-cmd")
	os.MkdirAll(platformDir, 0755)
	os.WriteFile(filepath.Join(platformDir, "INDEX.dro.md"),
		[]byte("---\nname: nav-cmd\ndescription: test\n---\n# Nav\n"), 0644)

	// Create shared rite mena (KNOSSOS_HOME/rites/shared/mena/)
	sharedKnowDir := filepath.Join(tmpDir, "rites", "shared", "mena", "know")
	os.MkdirAll(sharedKnowDir, 0755)
	os.WriteFile(filepath.Join(sharedKnowDir, "INDEX.dro.md"),
		[]byte("---\nname: know\ndescription: Generate knowledge\n---\n# Know\n"), 0644)

	sharedSkillDir := filepath.Join(tmpDir, "rites", "shared", "mena", "shared-ref")
	os.MkdirAll(sharedSkillDir, 0755)
	os.WriteFile(filepath.Join(sharedSkillDir, "INDEX.lego.md"),
		[]byte("---\nname: shared-ref\ndescription: Shared reference\n---\n# Ref\n"), 0644)

	userClaudeDir := filepath.Join(t.TempDir(), ".claude")
	manifest := &provenance.ProvenanceManifest{
		Entries: map[string]*provenance.ProvenanceEntry{},
	}
	checker := NewCollisionChecker(t.TempDir()) // no rite manifest → not effective → no collisions

	s := &syncer{}
	result, err := s.syncUserMena(tmpDir, userClaudeDir, manifest, checker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserMena failed: %v", err)
	}

	// Platform dro mena should be promoted to parent level (commands/nav-cmd.md)
	navCmd := filepath.Join(userClaudeDir, "commands", "nav-cmd.md")
	if _, err := os.Stat(navCmd); os.IsNotExist(err) {
		t.Errorf("Expected promoted platform dromenon at %s", navCmd)
	}

	// Shared rite dro mena should ALSO be promoted
	knowCmd := filepath.Join(userClaudeDir, "commands", "know.md")
	if _, err := os.Stat(knowCmd); os.IsNotExist(err) {
		t.Errorf("Expected promoted shared dromenon /know at %s", knowCmd)
	}

	// Shared lego mena should have INDEX.md renamed to SKILL.md
	sharedSkill := filepath.Join(userClaudeDir, "skills", "shared-ref", "SKILL.md")
	if _, err := os.Stat(sharedSkill); os.IsNotExist(err) {
		t.Errorf("Expected shared legomenon SKILL.md at %s", sharedSkill)
	}

	// Verify added count includes shared mena entries
	if result.Summary.Added < 3 {
		t.Errorf("Expected at least 3 added entries (platform + shared dro + shared lego), got %d", result.Summary.Added)
	}
}

// TestSyncUserMena_EmbeddedIncludesSharedRiteMena verifies that the embedded
// fallback path includes rites/shared/mena/ entries alongside platform mena.
func TestSyncUserMena_EmbeddedIncludesSharedRiteMena(t *testing.T) {
	userClaudeDir := t.TempDir()

	// Build an embedded FS with platform mena + shared rite mena
	embeddedMena := fstest.MapFS{
		"mena/operations/commit/INDEX.dro.md": &fstest.MapFile{
			Data: []byte("---\nname: commit\ndescription: Commit\n---\n# commit\n"),
		},
	}
	embeddedRites := fstest.MapFS{
		"rites/shared/mena/smell-detection/INDEX.lego.md": &fstest.MapFile{
			Data: []byte("---\nname: smell-detection\ndescription: Smell detection\n---\n# smell-detection\n"),
		},
	}

	s := &syncer{
		embeddedMena:  embeddedMena,
		embeddedRites: embeddedRites,
	}
	manifest, _ := provenance.LoadOrBootstrap(filepath.Join(t.TempDir(), "manifest.yaml"))
	collisionChecker := NewCollisionChecker(t.TempDir())

	result, err := s.syncUserMenaFromEmbedded(userClaudeDir, manifest, collisionChecker, SyncOptions{})
	if err != nil {
		t.Fatalf("syncUserMenaFromEmbedded failed: %v", err)
	}

	// Platform dro should be promoted
	commitCmd := filepath.Join(userClaudeDir, "commands", "commit.md")
	if _, statErr := os.Stat(commitCmd); os.IsNotExist(statErr) {
		t.Errorf("Expected platform dromenon /commit at %s", commitCmd)
	}

	// Shared lego should be projected with SKILL.md rename
	smellSkill := filepath.Join(userClaudeDir, "skills", "smell-detection", "SKILL.md")
	if _, statErr := os.Stat(smellSkill); os.IsNotExist(statErr) {
		t.Errorf("Expected shared legomenon smell-detection at %s", smellSkill)
	}

	// Both entries should be added
	if result.Summary.Added < 2 {
		t.Errorf("Expected at least 2 added entries (platform dro + shared lego), got %d", result.Summary.Added)
	}
}

// Helper: check if byte slice contains a string
func containsBytes(data []byte, substr string) bool {
	return len(data) > 0 && len(substr) > 0 && bytesContains(data, []byte(substr))
}

func bytesContains(b, sub []byte) bool {
	for i := 0; i <= len(b)-len(sub); i++ {
		if string(b[i:i+len(sub)]) == string(sub) {
			return true
		}
	}
	return false
}
