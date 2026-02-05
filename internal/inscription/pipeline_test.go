package inscription

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewPipeline(t *testing.T) {
	pipeline := NewPipeline("/project")

	if pipeline.ClaudeMDPath != "/project/.claude/CLAUDE.md" {
		t.Errorf("NewPipeline() ClaudeMDPath = %q, want '/project/.claude/CLAUDE.md'", pipeline.ClaudeMDPath)
	}
	if pipeline.ManifestPath != "/project/.claude/KNOSSOS_MANIFEST.yaml" {
		t.Errorf("NewPipeline() ManifestPath = %q, want '/project/.claude/KNOSSOS_MANIFEST.yaml'", pipeline.ManifestPath)
	}
	if pipeline.TemplateDir != "/project/knossos/templates" {
		t.Errorf("NewPipeline() TemplateDir = %q, want '/project/knossos/templates'", pipeline.TemplateDir)
	}
	if pipeline.BackupDir != "/project/.claude/backups" {
		t.Errorf("NewPipeline() BackupDir = %q, want '/project/.claude/backups'", pipeline.BackupDir)
	}
	if pipeline.ProjectRoot != "/project" {
		t.Errorf("NewPipeline() ProjectRoot = %q, want '/project'", pipeline.ProjectRoot)
	}
}

func TestNewPipelineWithPaths(t *testing.T) {
	pipeline := NewPipelineWithPaths(
		"/custom/path/CLAUDE.md",
		"/custom/manifest.yaml",
		"/custom/templates",
		"/custom/backups",
	)

	if pipeline.ClaudeMDPath != "/custom/path/CLAUDE.md" {
		t.Errorf("NewPipelineWithPaths() ClaudeMDPath = %q", pipeline.ClaudeMDPath)
	}
	if pipeline.ManifestPath != "/custom/manifest.yaml" {
		t.Errorf("NewPipelineWithPaths() ManifestPath = %q", pipeline.ManifestPath)
	}
	if pipeline.TemplateDir != "/custom/templates" {
		t.Errorf("NewPipelineWithPaths() TemplateDir = %q", pipeline.TemplateDir)
	}
	if pipeline.BackupDir != "/custom/backups" {
		t.Errorf("NewPipelineWithPaths() BackupDir = %q", pipeline.BackupDir)
	}
}

func TestPipeline_Validate_NoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	pipeline := NewPipeline(tmpDir)

	result, err := pipeline.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should not fail, just warn about missing manifest
	if result == nil {
		t.Fatal("Validate() result should not be nil")
	}

	// Should have a warning about missing manifest
	hasManifestWarning := false
	for _, issue := range result.Issues {
		if strings.Contains(strings.ToLower(issue.Message), "manifest") {
			hasManifestWarning = true
			break
		}
	}
	if !hasManifestWarning {
		t.Errorf("Validate() should warn about missing manifest, got issues: %v", result.Issues)
	}
}

func TestPipeline_Validate_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Create a valid manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
  quick-start:
    owner: regenerate
    source: ACTIVE_RITE
`
	os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	result, err := pipeline.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.SchemaVersion != "1.0" {
		t.Errorf("Validate() SchemaVersion = %q, want '1.0'", result.SchemaVersion)
	}

	if result.RegionCount != 2 {
		t.Errorf("Validate() RegionCount = %d, want 2", result.RegionCount)
	}
}

func TestPipeline_Validate_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Create an invalid manifest (missing required fields)
	manifestContent := `regions:
  test:
    owner: invalid-owner
`
	os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	result, err := pipeline.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Valid {
		t.Error("Validate() should mark invalid manifest as invalid")
	}
}

func TestPipeline_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "5"
regions:
  execution-mode:
    owner: knossos
  project-custom:
    owner: satellite
section_order:
  - execution-mode
  - project-custom
`
	os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	preview, err := pipeline.DryRun(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("DryRun() error = %v", err)
	}

	if preview.CurrentVersion != "5" {
		t.Errorf("DryRun() CurrentVersion = %q, want '5'", preview.CurrentVersion)
	}

	if preview.NewVersion != "6" {
		t.Errorf("DryRun() NewVersion = %q, want '6'", preview.NewVersion)
	}

	// Should sync knossos regions
	if !contains(preview.WouldSync, "execution-mode") {
		t.Error("DryRun() should include 'execution-mode' in WouldSync")
	}

	// Should preserve satellite regions
	if !contains(preview.WouldPreserve, "project-custom") {
		t.Error("DryRun() should include 'project-custom' in WouldPreserve")
	}
}

func TestPipeline_Sync_CreatesBackup(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	backupsDir := filepath.Join(claudeDir, "backups")
	os.MkdirAll(claudeDir, 0755)

	// Create existing CLAUDE.md
	existingContent := "# CLAUDE.md\n\nOriginal content"
	os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte(existingContent), 0644)

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	result, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() should succeed")
	}

	// Check backup was created
	entries, _ := os.ReadDir(backupsDir)
	if len(entries) == 0 {
		t.Error("Sync() should create backup")
	}
}

func TestPipeline_Sync_NoBackupOption(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	backupsDir := filepath.Join(claudeDir, "backups")
	os.MkdirAll(claudeDir, 0755)

	// Create existing CLAUDE.md
	existingContent := "# CLAUDE.md\n\nOriginal content"
	os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte(existingContent), 0644)

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	result, err := pipeline.Sync(InscriptionSyncOptions{NoBackup: true})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.Success {
		t.Error("Sync() should succeed")
	}

	if result.BackupPath != "" {
		t.Error("Sync() with NoBackup should not have backup path")
	}

	// Check no backup was created
	entries, _ := os.ReadDir(backupsDir)
	if len(entries) > 0 {
		t.Error("Sync() with NoBackup should not create backup")
	}
}

func TestPipeline_Sync_DryRunDoesNotWrite(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	result, err := pipeline.Sync(InscriptionSyncOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !result.DryRun {
		t.Error("Sync() with DryRun should set DryRun flag in result")
	}

	// Check CLAUDE.md was NOT created
	claudeMDPath := filepath.Join(claudeDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMDPath); err == nil {
		t.Error("Sync() with DryRun should not create CLAUDE.md")
	}
}

func TestPipeline_Sync_UpdatesManifestVersion(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Create manifest with version 5
	manifestContent := `schema_version: "1.0"
inscription_version: "5"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	manifestPath := filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml")
	os.WriteFile(manifestPath, []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	result, err := pipeline.Sync(InscriptionSyncOptions{})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if result.InscriptionVersion != "6" {
		t.Errorf("Sync() InscriptionVersion = %q, want '6'", result.InscriptionVersion)
	}
}

func TestPipeline_ListBackups(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	backupsDir := filepath.Join(claudeDir, "backups")
	os.MkdirAll(backupsDir, 0755)

	// Create some backup files
	os.WriteFile(filepath.Join(backupsDir, "CLAUDE.md.2026-01-06T10-00-00Z"), []byte("backup1"), 0644)
	os.WriteFile(filepath.Join(backupsDir, "CLAUDE.md.2026-01-05T10-00-00Z"), []byte("backup2"), 0644)

	pipeline := NewPipeline(tmpDir)

	backups, err := pipeline.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("ListBackups() len = %d, want 2", len(backups))
	}
}

func TestPipeline_GetDiff(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Create existing CLAUDE.md with old content
	existingContent := `# CLAUDE.md

<!-- KNOSSOS:START execution-mode -->
## Execution Mode

Old content here
<!-- KNOSSOS:END execution-mode -->
`
	os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte(existingContent), 0644)

	// Create manifest
	manifestContent := `schema_version: "1.0"
inscription_version: "1"
regions:
  execution-mode:
    owner: knossos
section_order:
  - execution-mode
`
	os.WriteFile(filepath.Join(claudeDir, "KNOSSOS_MANIFEST.yaml"), []byte(manifestContent), 0644)

	pipeline := NewPipeline(tmpDir)

	diff, err := pipeline.GetDiff("execution-mode")
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}

	// Should show diff between old and new
	if !strings.Contains(diff, "execution-mode") {
		t.Error("GetDiff() should contain region name")
	}
}

func TestSyncOptions_Defaults(t *testing.T) {
	opts := InscriptionSyncOptions{}

	if opts.Force {
		t.Error("SyncOptions.Force should default to false")
	}
	if opts.DryRun {
		t.Error("SyncOptions.DryRun should default to false")
	}
	if opts.NoBackup {
		t.Error("SyncOptions.NoBackup should default to false")
	}
	if opts.Verbose {
		t.Error("SyncOptions.Verbose should default to false")
	}
	if opts.RiteName != "" {
		t.Error("SyncOptions.RiteName should default to empty")
	}
}

func TestSyncResult_Fields(t *testing.T) {
	result := &SyncResult{
		Success:            true,
		RegionsSynced:      []string{"region1", "region2"},
		Conflicts:          []Conflict{{Region: "conflict1"}},
		BackupPath:         "/path/to/backup",
		InscriptionVersion: "5",
		DryRun:             false,
	}

	if !result.Success {
		t.Error("SyncResult.Success not set")
	}
	if len(result.RegionsSynced) != 2 {
		t.Error("SyncResult.RegionsSynced not set")
	}
	if len(result.Conflicts) != 1 {
		t.Error("SyncResult.Conflicts not set")
	}
	if result.BackupPath != "/path/to/backup" {
		t.Error("SyncResult.BackupPath not set")
	}
	if result.InscriptionVersion != "5" {
		t.Error("SyncResult.InscriptionVersion not set")
	}
}

func TestValidationResult_Fields(t *testing.T) {
	result := &ValidationResult{
		Valid:         true,
		RegionCount:   5,
		SchemaVersion: "1.0",
		Issues: []ValidationIssue{
			{Severity: "warning", Message: "test warning"},
		},
	}

	if !result.Valid {
		t.Error("ValidationResult.Valid not set")
	}
	if result.RegionCount != 5 {
		t.Error("ValidationResult.RegionCount not set")
	}
	if result.SchemaVersion != "1.0" {
		t.Error("ValidationResult.SchemaVersion not set")
	}
	if len(result.Issues) != 1 {
		t.Error("ValidationResult.Issues not set")
	}
}

func TestValidationIssue_Fields(t *testing.T) {
	issue := ValidationIssue{
		Severity: "error",
		Region:   "test-region",
		Message:  "test error message",
	}

	if issue.Severity != "error" {
		t.Error("ValidationIssue.Severity not set")
	}
	if issue.Region != "test-region" {
		t.Error("ValidationIssue.Region not set")
	}
	if issue.Message != "test error message" {
		t.Error("ValidationIssue.Message not set")
	}
}

func TestSyncPreview_Fields(t *testing.T) {
	preview := &SyncPreview{
		WouldSync:      []string{"region1"},
		WouldPreserve:  []string{"region2"},
		Conflicts:      []Conflict{{Region: "conflict1"}},
		CurrentVersion: "1",
		NewVersion:     "2",
	}

	if len(preview.WouldSync) != 1 {
		t.Error("SyncPreview.WouldSync not set")
	}
	if len(preview.WouldPreserve) != 1 {
		t.Error("SyncPreview.WouldPreserve not set")
	}
	if len(preview.Conflicts) != 1 {
		t.Error("SyncPreview.Conflicts not set")
	}
	if preview.CurrentVersion != "1" {
		t.Error("SyncPreview.CurrentVersion not set")
	}
	if preview.NewVersion != "2" {
		t.Error("SyncPreview.NewVersion not set")
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"1", 1},
		{"5", 5},
		{"10", 10},
		{"123", 123},
		{"abc", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := parseVersion(tt.input)
			if got != tt.want {
				t.Errorf("parseVersion(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"short", 5, "short"},
		{"short", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !contains(slice, "banana") {
		t.Error("contains() should find 'banana'")
	}
	if contains(slice, "grape") {
		t.Error("contains() should not find 'grape'")
	}
	if contains(nil, "test") {
		t.Error("contains() should return false for nil slice")
	}
}

func TestSimpleDiff(t *testing.T) {
	// New region
	diff := simpleDiff("test", "", "new content")
	if !strings.Contains(diff, "+ (new region)") {
		t.Error("simpleDiff() should indicate new region")
	}

	// Removed region
	diff = simpleDiff("test", "old content", "")
	if !strings.Contains(diff, "- (region removed)") {
		t.Error("simpleDiff() should indicate removed region")
	}

	// Changed region
	diff = simpleDiff("test", "old", "new")
	if !strings.Contains(diff, "@@ region changed @@") {
		t.Error("simpleDiff() should indicate changed region")
	}
}

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantNil  bool
		wantRole string
		wantDesc string
	}{
		{
			name: "valid frontmatter with description",
			content: `---
name: architect
role: "Evaluates tradeoffs"
description: "System design authority who evaluates technical tradeoffs"
tools: Bash, Glob
---
# Architect

Main content here`,
			wantNil:  false,
			wantRole: "Evaluates tradeoffs",
			wantDesc: "System design authority who evaluates technical tradeoffs",
		},
		{
			name: "frontmatter without role",
			content: `---
name: qa-adversary
description: "Tests everything adversarially"
---
# QA`,
			wantNil:  false,
			wantRole: "",
			wantDesc: "Tests everything adversarially",
		},
		{
			name:     "no frontmatter",
			content:  "# Just a markdown file\n\nNo YAML here",
			wantNil:  true,
			wantRole: "",
			wantDesc: "",
		},
		{
			name:     "unclosed frontmatter",
			content:  "---\nname: test\nNo closing delimiter",
			wantNil:  true,
			wantRole: "",
			wantDesc: "",
		},
		{
			name:     "empty content",
			content:  "",
			wantNil:  true,
			wantRole: "",
			wantDesc: "",
		},
		{
			name:     "too short",
			content:  "---",
			wantNil:  true,
			wantRole: "",
			wantDesc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := extractFrontmatter(tt.content)

			if tt.wantNil {
				if fm != nil {
					t.Errorf("extractFrontmatter() = %v, want nil", fm)
				}
				return
			}

			if fm == nil {
				t.Fatal("extractFrontmatter() returned nil, want non-nil")
			}

			if fm.Role != tt.wantRole {
				t.Errorf("extractFrontmatter().Role = %q, want %q", fm.Role, tt.wantRole)
			}
			if fm.Description != tt.wantDesc {
				t.Errorf("extractFrontmatter().Description = %q, want %q", fm.Description, tt.wantDesc)
			}
		})
	}
}

func TestLoadAgents_WithYAMLFrontmatter(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create agent with YAML frontmatter
	architectContent := `---
name: architect
role: "Evaluates tradeoffs"
description: "System design authority who evaluates technical tradeoffs and produces TDDs."
tools: Bash, Glob
---
# Architect

Main content here.`

	if err := os.WriteFile(filepath.Join(agentsDir, "architect.md"), []byte(architectContent), 0644); err != nil {
		t.Fatalf("Failed to write architect.md: %v", err)
	}

	// Create agent without frontmatter (legacy format)
	legacyContent := `# Legacy Agent

This is a legacy agent without YAML frontmatter.
Produces: artifacts`

	if err := os.WriteFile(filepath.Join(agentsDir, "legacy.md"), []byte(legacyContent), 0644); err != nil {
		t.Fatalf("Failed to write legacy.md: %v", err)
	}

	// Load agents
	pipeline := NewPipeline(tmpDir)
	agents, err := pipeline.loadAgents(agentsDir)
	if err != nil {
		t.Fatalf("loadAgents() error = %v", err)
	}

	if len(agents) != 2 {
		t.Fatalf("loadAgents() returned %d agents, want 2", len(agents))
	}

	// Find architect agent
	var architectAgent *AgentInfo
	var legacyAgent *AgentInfo
	for i := range agents {
		if agents[i].Name == "architect" {
			architectAgent = &agents[i]
		}
		if agents[i].Name == "legacy" {
			legacyAgent = &agents[i]
		}
	}

	// Verify architect agent parsed YAML frontmatter
	if architectAgent == nil {
		t.Fatal("architect agent not found")
	}
	// Should use description (more verbose) over role
	if architectAgent.Role != "System design authority who evaluates technical tradeoffs and produces TDDs." {
		t.Errorf("architect.Role = %q, want description from frontmatter", architectAgent.Role)
	}
	// Role should NOT be "---" (the YAML delimiter)
	if architectAgent.Role == "---" {
		t.Error("architect.Role should not be '---' (YAML delimiter)")
	}

	// Verify legacy agent used fallback parsing
	if legacyAgent == nil {
		t.Fatal("legacy agent not found")
	}
	// Role should not be "---"
	if legacyAgent.Role == "---" {
		t.Error("legacy.Role should not be '---'")
	}
}
