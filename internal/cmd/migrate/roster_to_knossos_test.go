package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test fixtures
const testUserManifestRoster = `{
  "manifest_version": "1.0",
  "last_sync": "2026-01-15T10:00:00Z",
  "agents": {
    "moirai.md": {
      "source": "roster",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "abc123"
    },
    "custom.md": {
      "source": "user",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "def456"
    }
  }
}`

const testUserManifestKnossos = `{
  "manifest_version": "1.0",
  "last_sync": "2026-01-15T10:00:00Z",
  "agents": {
    "moirai.md": {
      "source": "knossos",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "abc123"
    },
    "custom.md": {
      "source": "user",
      "installed_at": "2026-01-15T10:00:00Z",
      "checksum": "def456"
    }
  }
}`

const testUserManifestDiverged = `{
  "manifest_version": "1.0",
  "agents": {
    "agent1.md": {
      "source": "roster-diverged",
      "checksum": "xyz"
    }
  }
}`

const testUserManifestMixed = `{
  "manifest_version": "1.0",
  "agents": {
    "a1.md": {"source": "roster"},
    "a2.md": {"source": "knossos"},
    "a3.md": {"source": "user"}
  }
}`

const testCEMManifestRoster = `{
  "schema_version": 3,
  "roster": {
    "path": "/Users/test/Code/knossos",
    "commit": "abc123",
    "ref": "main",
    "last_sync": "2026-01-07T17:57:29Z"
  },
  "team": {
    "name": "10x-dev",
    "roster_path": "/Users/test/Code/roster/rites/10x-dev"
  },
  "managed_files": [
    {"path": ".claude/commands", "source": "roster"},
    {"path": ".claude/hooks", "source": "roster"}
  ]
}`

const testCEMManifestKnossos = `{
  "schema_version": 3,
  "knossos": {
    "path": "/Users/test/Code/knossos",
    "commit": "abc123",
    "ref": "main",
    "last_sync": "2026-01-07T17:57:29Z"
  },
  "team": {
    "name": "10x-dev",
    "knossos_path": "/Users/test/Code/roster/rites/10x-dev"
  },
  "managed_files": [
    {"path": ".claude/commands", "source": "knossos"},
    {"path": ".claude/hooks", "source": "knossos"}
  ]
}`

func TestRewriteUserManifest_RosterToKnossos(t *testing.T) {
	rewritten, count, err := rewriteUserManifestBytes([]byte(testUserManifestRoster))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 entry rewritten, got %d", count)
	}

	// Verify the source field changed
	var manifest map[string]interface{}
	if err := json.Unmarshal(rewritten, &manifest); err != nil {
		t.Fatalf("failed to unmarshal rewritten manifest: %v", err)
	}

	agents := manifest["agents"].(map[string]interface{})
	moirai := agents["moirai.md"].(map[string]interface{})
	if moirai["source"] != "knossos" {
		t.Errorf("expected source=knossos, got %v", moirai["source"])
	}

	// Verify user source unchanged
	custom := agents["custom.md"].(map[string]interface{})
	if custom["source"] != "user" {
		t.Errorf("expected source=user, got %v", custom["source"])
	}
}

func TestRewriteUserManifest_DivergedToKnossosDiverged(t *testing.T) {
	rewritten, count, err := rewriteUserManifestBytes([]byte(testUserManifestDiverged))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 entry rewritten, got %d", count)
	}

	var manifest map[string]interface{}
	json.Unmarshal(rewritten, &manifest)
	agents := manifest["agents"].(map[string]interface{})
	agent1 := agents["agent1.md"].(map[string]interface{})
	if agent1["source"] != "knossos-diverged" {
		t.Errorf("expected source=knossos-diverged, got %v", agent1["source"])
	}
}

func TestRewriteUserManifest_AlreadyMigrated(t *testing.T) {
	rewritten, count, err := rewriteUserManifestBytes([]byte(testUserManifestKnossos))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 entries rewritten (already migrated), got %d", count)
	}

	// Verify bytes unchanged
	if string(rewritten) != testUserManifestKnossos {
		t.Errorf("manifest changed when it should be idempotent")
	}
}

func TestRewriteUserManifest_MixedSources(t *testing.T) {
	rewritten, count, err := rewriteUserManifestBytes([]byte(testUserManifestMixed))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 entry rewritten (only roster), got %d", count)
	}

	var manifest map[string]interface{}
	json.Unmarshal(rewritten, &manifest)
	agents := manifest["agents"].(map[string]interface{})

	a1 := agents["a1.md"].(map[string]interface{})
	if a1["source"] != "knossos" {
		t.Errorf("a1 source should be knossos, got %v", a1["source"])
	}

	a2 := agents["a2.md"].(map[string]interface{})
	if a2["source"] != "knossos" {
		t.Errorf("a2 source should remain knossos, got %v", a2["source"])
	}

	a3 := agents["a3.md"].(map[string]interface{})
	if a3["source"] != "user" {
		t.Errorf("a3 source should remain user, got %v", a3["source"])
	}
}

func TestRewriteUserManifest_EmptyManifest(t *testing.T) {
	emptyManifest := `{}`
	rewritten, count, err := rewriteUserManifestBytes([]byte(emptyManifest))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 entries rewritten (empty manifest), got %d", count)
	}

	if string(rewritten) != emptyManifest {
		t.Errorf("empty manifest changed")
	}
}

func TestRewriteUserManifest_InvalidJSON(t *testing.T) {
	_, _, err := rewriteUserManifestBytes([]byte("not json"))
	if err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}

func TestRewriteUserManifest_PreservesOtherFields(t *testing.T) {
	rewritten, _, err := rewriteUserManifestBytes([]byte(testUserManifestRoster))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var manifest map[string]interface{}
	json.Unmarshal(rewritten, &manifest)

	// Verify manifest_version preserved
	if manifest["manifest_version"] != "1.0" {
		t.Errorf("manifest_version not preserved")
	}

	// Verify last_sync preserved
	if manifest["last_sync"] != "2026-01-15T10:00:00Z" {
		t.Errorf("last_sync not preserved")
	}

	// Verify entry fields preserved
	agents := manifest["agents"].(map[string]interface{})
	moirai := agents["moirai.md"].(map[string]interface{})
	if moirai["checksum"] != "abc123" {
		t.Errorf("checksum not preserved")
	}
	if moirai["installed_at"] != "2026-01-15T10:00:00Z" {
		t.Errorf("installed_at not preserved")
	}
}

func TestRewriteCEM_RosterKeyRename(t *testing.T) {
	rewritten, count, err := rewriteCEMManifestBytes([]byte(testCEMManifestRoster))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 4 {
		t.Errorf("expected 4 changes (1 roster key + 1 team.roster_path + 2 managed_files), got %d", count)
	}

	var manifest map[string]interface{}
	json.Unmarshal(rewritten, &manifest)

	// Verify "knossos" key exists
	if _, ok := manifest["knossos"]; !ok {
		t.Errorf("knossos key not found")
	}

	// Verify "roster" key removed
	if _, ok := manifest["roster"]; ok {
		t.Errorf("roster key should be removed")
	}

	// Verify knossos key contents
	knossos := manifest["knossos"].(map[string]interface{})
	if knossos["path"] != "/Users/test/Code/knossos" {
		t.Errorf("knossos.path not preserved")
	}
}

func TestRewriteCEM_TeamRosterPath(t *testing.T) {
	rewritten, _, err := rewriteCEMManifestBytes([]byte(testCEMManifestRoster))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var manifest map[string]interface{}
	json.Unmarshal(rewritten, &manifest)

	team := manifest["team"].(map[string]interface{})

	// Verify knossos_path exists
	if _, ok := team["knossos_path"]; !ok {
		t.Errorf("knossos_path not found")
	}

	// Verify roster_path removed
	if _, ok := team["roster_path"]; ok {
		t.Errorf("roster_path should be removed")
	}

	// Verify value preserved
	if team["knossos_path"] != "/Users/test/Code/roster/rites/10x-dev" {
		t.Errorf("knossos_path value not preserved")
	}
}

func TestRewriteCEM_ManagedFilesSource(t *testing.T) {
	rewritten, _, err := rewriteCEMManifestBytes([]byte(testCEMManifestRoster))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var manifest map[string]interface{}
	json.Unmarshal(rewritten, &manifest)

	managedFiles := manifest["managed_files"].([]interface{})
	for _, fileVal := range managedFiles {
		fileMap := fileVal.(map[string]interface{})
		if fileMap["source"] != "knossos" {
			t.Errorf("managed_files source should be knossos, got %v", fileMap["source"])
		}
	}
}

func TestRewriteCEM_AlreadyMigrated(t *testing.T) {
	rewritten, count, err := rewriteCEMManifestBytes([]byte(testCEMManifestKnossos))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 changes (already migrated), got %d", count)
	}

	if string(rewritten) != testCEMManifestKnossos {
		t.Errorf("manifest changed when it should be idempotent")
	}
}

func TestRewriteCEM_NoRosterKey(t *testing.T) {
	noRoster := `{"schema_version": 3}`
	rewritten, count, err := rewriteCEMManifestBytes([]byte(noRoster))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 changes (no roster key), got %d", count)
	}

	if string(rewritten) != noRoster {
		t.Errorf("manifest changed when no roster key present")
	}
}

func TestScanEnvVars_DetectsRosterHome(t *testing.T) {
	// Set test env var
	os.Setenv("ROSTER_HOME", "/test/path")
	defer os.Unsetenv("ROSTER_HOME")

	mappings := scanRosterEnvVars()

	found := false
	for _, m := range mappings {
		if m.Old == "ROSTER_HOME" && m.New == "KNOSSOS_HOME" && m.Value == "/test/path" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("ROSTER_HOME not detected or mapped correctly")
	}
}

func TestScanEnvVars_NoRosterVars(t *testing.T) {
	// Ensure no ROSTER_* vars set (may fail in dev env, but should pass in clean CI)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "ROSTER_") {
			t.Skip("ROSTER_* variables present in environment, skipping clean env test")
		}
	}

	mappings := scanRosterEnvVars()

	if len(mappings) != 0 {
		t.Errorf("expected no mappings in clean env, got %d", len(mappings))
	}
}

func TestScanEnvVars_MultipleVars(t *testing.T) {
	os.Setenv("ROSTER_HOME", "/home")
	os.Setenv("ROSTER_VERBOSE", "true")
	defer func() {
		os.Unsetenv("ROSTER_HOME")
		os.Unsetenv("ROSTER_VERBOSE")
	}()

	mappings := scanRosterEnvVars()

	if len(mappings) < 2 {
		t.Errorf("expected at least 2 mappings, got %d", len(mappings))
	}
}

func TestScanEnvVars_PrefixVars(t *testing.T) {
	os.Setenv("ROSTER_PREF_FOO", "bar")
	defer os.Unsetenv("ROSTER_PREF_FOO")

	mappings := scanRosterEnvVars()

	found := false
	for _, m := range mappings {
		if m.Old == "ROSTER_PREF_FOO" && m.New == "KNOSSOS_PREF_FOO" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("ROSTER_PREF_FOO not detected with correct mapping")
	}
}

func TestGenerateScript_WithVars(t *testing.T) {
	envVars := []EnvVarMapping{
		{Old: "ROSTER_HOME", New: "KNOSSOS_HOME", Value: "/test"},
		{Old: "ROSTER_VERBOSE", New: "KNOSSOS_VERBOSE", Value: "true"},
	}

	script := generateMigrationScript(envVars)

	// Verify shebang
	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Errorf("script should start with shebang")
	}

	// Verify sed commands present
	if !strings.Contains(script, "ROSTER_HOME") {
		t.Errorf("script should contain ROSTER_HOME replacement")
	}
	if !strings.Contains(script, "KNOSSOS_HOME") {
		t.Errorf("script should contain KNOSSOS_HOME")
	}

	// Verify platform detection
	if !strings.Contains(script, "uname -s") {
		t.Errorf("script should detect platform for sed compatibility")
	}
}

func TestGenerateScript_NoVars(t *testing.T) {
	script := generateMigrationScript([]EnvVarMapping{})

	if !strings.Contains(script, "No ROSTER_* environment variables detected") {
		t.Errorf("script should indicate no variables detected")
	}
}

func TestGenerateScript_HasShebang(t *testing.T) {
	script := generateMigrationScript([]EnvVarMapping{
		{Old: "ROSTER_HOME", New: "KNOSSOS_HOME", Value: "/test"},
	})

	if !strings.HasPrefix(script, "#!/bin/bash\n") {
		t.Errorf("script must start with #!/bin/bash")
	}
}

// Filesystem integration tests
func TestMigrateUserManifests_DryRun(t *testing.T) {
	// Create temp dir with test manifests
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	manifestPath := filepath.Join(claudeDir, "USER_AGENT_MANIFEST.json")
	os.WriteFile(manifestPath, []byte(testUserManifestRoster), 0644)

	// Temporarily override user home
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	results, err := migrateUserManifests(true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Skipped {
		t.Errorf("manifest should not be skipped")
	}
	if result.EntriesRewritten != 1 {
		t.Errorf("expected 1 entry rewritten, got %d", result.EntriesRewritten)
	}

	// Verify file unchanged (dry-run)
	data, _ := os.ReadFile(manifestPath)
	if string(data) != testUserManifestRoster {
		t.Errorf("file should not be modified in dry-run")
	}
}

func TestMigrateUserManifests_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	manifestPath := filepath.Join(claudeDir, "USER_AGENT_MANIFEST.json")
	os.WriteFile(manifestPath, []byte(testUserManifestRoster), 0644)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	results, err := migrateUserManifests(false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := results[0]
	if result.Skipped {
		t.Errorf("manifest should not be skipped")
	}

	// Verify file changed
	data, _ := os.ReadFile(manifestPath)
	var manifest map[string]interface{}
	json.Unmarshal(data, &manifest)
	agents := manifest["agents"].(map[string]interface{})
	moirai := agents["moirai.md"].(map[string]interface{})
	if moirai["source"] != "knossos" {
		t.Errorf("source should be knossos after apply")
	}

	// Verify backup created
	backupPath := manifestPath + ".roster-backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("backup should be created")
	}
}

func TestMigrateUserManifests_ApplyNoBackup(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	manifestPath := filepath.Join(claudeDir, "USER_AGENT_MANIFEST.json")
	os.WriteFile(manifestPath, []byte(testUserManifestRoster), 0644)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	results, err := migrateUserManifests(false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file changed
	data, _ := os.ReadFile(manifestPath)
	var manifest map[string]interface{}
	json.Unmarshal(data, &manifest)
	agents := manifest["agents"].(map[string]interface{})
	moirai := agents["moirai.md"].(map[string]interface{})
	if moirai["source"] != "knossos" {
		t.Errorf("source should be knossos after apply")
	}

	// Verify no backup
	backupPath := manifestPath + ".roster-backup"
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Errorf("backup should not be created with no-backup flag")
	}

	// Verify BackupPath empty
	if results[0].BackupPath != "" {
		t.Errorf("BackupPath should be empty when backup=false")
	}
}

func TestMigrateUserManifests_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	manifestPath := filepath.Join(claudeDir, "USER_AGENT_MANIFEST.json")
	os.WriteFile(manifestPath, []byte(testUserManifestRoster), 0644)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// First migration
	results1, _ := migrateUserManifests(false, true)
	if results1[0].Skipped {
		t.Errorf("first migration should not be skipped")
	}

	// Second migration
	results2, _ := migrateUserManifests(false, true)
	if !results2[0].Skipped {
		t.Errorf("second migration should be skipped (idempotent)")
	}
	if results2[0].SkipReason != "already migrated" {
		t.Errorf("skip reason should be 'already migrated', got %s", results2[0].SkipReason)
	}
}

func TestMigrateUserManifests_MissingDir(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	results, err := migrateUserManifests(false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("should return empty results for missing .claude dir")
	}
}

func TestMigrateCEMManifest_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	cemDir := filepath.Join(tmpDir, ".claude", ".cem")
	os.MkdirAll(cemDir, 0755)

	cemPath := filepath.Join(cemDir, "manifest.json")
	os.WriteFile(cemPath, []byte(testCEMManifestRoster), 0644)

	result, err := migrateCEMManifest(tmpDir, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected result, got nil")
	}

	if result.Skipped {
		t.Errorf("manifest should not be skipped")
	}

	// Verify file changed
	data, _ := os.ReadFile(cemPath)
	var manifest map[string]interface{}
	json.Unmarshal(data, &manifest)

	if _, ok := manifest["knossos"]; !ok {
		t.Errorf("knossos key should exist")
	}
	if _, ok := manifest["roster"]; ok {
		t.Errorf("roster key should be removed")
	}
}

func TestMigrateCEMManifest_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := migrateCEMManifest(tmpDir, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("should return nil for missing CEM manifest")
	}
}

func TestBackupNotOverwritten(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	manifestPath := filepath.Join(claudeDir, "USER_AGENT_MANIFEST.json")
	os.WriteFile(manifestPath, []byte(testUserManifestRoster), 0644)

	// Create existing backup
	backupPath := manifestPath + ".roster-backup"
	originalBackup := []byte("original backup content")
	os.WriteFile(backupPath, originalBackup, 0644)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Run migration with backup=true
	results, _ := migrateUserManifests(false, true)

	// Verify backup unchanged
	backupData, _ := os.ReadFile(backupPath)
	if string(backupData) != string(originalBackup) {
		t.Errorf("existing backup should not be overwritten")
	}

	// Verify result indicates backup exists
	if !strings.Contains(results[0].BackupPath, "exists") {
		t.Errorf("BackupPath should indicate backup already exists")
	}
}
