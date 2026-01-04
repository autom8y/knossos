package team

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// getTestDataPath returns the absolute path to testdata/teams.
func getTestDataPath(t *testing.T) string {
	t.Helper()
	testdataPath := filepath.Join("..", "..", "testdata", "teams")
	absPath, err := filepath.Abs(testdataPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}
	return absPath
}

func TestTeamContext_ToMarkdown(t *testing.T) {
	tests := []struct {
		name string
		ctx  *TeamContext
		want string
	}{
		{
			name: "empty context",
			ctx:  NewTeamContext("test-team"),
			want: "",
		},
		{
			name: "nil context",
			ctx:  nil,
			want: "",
		},
		{
			name: "single row",
			ctx: &TeamContext{
				SchemaVersion: "1.0",
				TeamName:      "test-team",
				ContextRows: []ContextRow{
					{Key: "Status", Value: "Active"},
				},
			},
			want: "| | |\n|---|---|\n| **Status** | Active |\n",
		},
		{
			name: "multiple rows",
			ctx: &TeamContext{
				SchemaVersion: "1.0",
				TeamName:      "test-team",
				ContextRows: []ContextRow{
					{Key: "Status", Value: "Active"},
					{Key: "Environment", Value: "Production"},
					{Key: "Version", Value: "1.0"},
				},
			},
			want: "| | |\n|---|---|\n| **Status** | Active |\n| **Environment** | Production |\n| **Version** | 1.0 |\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.ToMarkdown()
			if got != tt.want {
				t.Errorf("ToMarkdown() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestTeamContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ctx     *TeamContext
		wantErr bool
	}{
		{
			name: "valid context",
			ctx: &TeamContext{
				SchemaVersion: "1.0",
				TeamName:      "test-team",
			},
			wantErr: false,
		},
		{
			name: "missing team name",
			ctx: &TeamContext{
				SchemaVersion: "1.0",
			},
			wantErr: true,
		},
		{
			name: "missing schema version",
			ctx: &TeamContext{
				TeamName: "test-team",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ctx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTeamContext_AddRow(t *testing.T) {
	ctx := NewTeamContext("test-team")

	ctx.AddRow("Key1", "Value1")
	ctx.AddRow("Key2", "Value2")

	if len(ctx.ContextRows) != 2 {
		t.Errorf("len(ContextRows) = %d, want 2", len(ctx.ContextRows))
	}

	if ctx.GetRow("Key1") != "Value1" {
		t.Errorf("GetRow(Key1) = %q, want %q", ctx.GetRow("Key1"), "Value1")
	}

	if ctx.GetRow("NonExistent") != "" {
		t.Errorf("GetRow(NonExistent) = %q, want empty string", ctx.GetRow("NonExistent"))
	}
}

func TestTeamContext_HasRows(t *testing.T) {
	ctx := NewTeamContext("test-team")

	if ctx.HasRows() {
		t.Error("HasRows() = true on new context, want false")
	}

	ctx.AddRow("Key", "Value")

	if !ctx.HasRows() {
		t.Error("HasRows() = false after AddRow, want true")
	}
}

func TestContextLoader_Load_FromYAML(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	ctx, err := loader.Load("valid-team")
	if err != nil {
		t.Fatalf("Load(valid-team) error = %v", err)
	}

	if ctx.TeamName != "valid-team" {
		t.Errorf("TeamName = %q, want %q", ctx.TeamName, "valid-team")
	}

	if ctx.DisplayName != "Valid Test Team" {
		t.Errorf("DisplayName = %q, want %q", ctx.DisplayName, "Valid Test Team")
	}

	if ctx.Domain != "testing" {
		t.Errorf("Domain = %q, want %q", ctx.Domain, "testing")
	}

	// Check context rows
	if len(ctx.ContextRows) != 3 {
		t.Errorf("len(ContextRows) = %d, want 3", len(ctx.ContextRows))
	}

	if ctx.GetRow("Status") != "Testing" {
		t.Errorf("GetRow(Status) = %q, want %q", ctx.GetRow("Status"), "Testing")
	}

	// Check metadata
	if ctx.Metadata["purpose"] != "unit testing" {
		t.Errorf("Metadata[purpose] = %q, want %q", ctx.Metadata["purpose"], "unit testing")
	}
}

func TestContextLoader_Load_FallbackToOrchestrator(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	ctx, err := loader.Load("minimal-team")
	if err != nil {
		t.Fatalf("Load(minimal-team) error = %v", err)
	}

	if ctx.TeamName != "minimal-team" {
		t.Errorf("TeamName = %q, want %q", ctx.TeamName, "minimal-team")
	}

	if ctx.Domain != "testing" {
		t.Errorf("Domain = %q, want %q", ctx.Domain, "testing")
	}

	// Should have rows generated from orchestrator
	if !ctx.HasRows() {
		t.Error("HasRows() = false after fallback, want true")
	}
}

func TestContextLoader_Load_TeamNotFound(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	_, err := loader.Load("non-existent-team")
	if err == nil {
		t.Fatal("Load(non-existent-team) error = nil, want error")
	}
}

func TestContextLoader_Load_EmptyTeamName(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	_, err := loader.Load("")
	if err == nil {
		t.Fatal("Load('') error = nil, want error")
	}
}

func TestContextLoader_Load_MalformedYAML(t *testing.T) {
	// Create a temporary directory with malformed YAML
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, "bad-team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		t.Fatalf("failed to create team dir: %v", err)
	}

	// Write malformed YAML
	malformedYAML := `schema_version: "1.0"
team_name: bad-team
context_rows:
  - key: "unclosed string
    value: broken`

	if err := os.WriteFile(filepath.Join(teamDir, "context.yaml"), []byte(malformedYAML), 0644); err != nil {
		t.Fatalf("failed to write malformed YAML: %v", err)
	}

	loader := NewContextLoaderWithPaths(tmpDir, "")

	_, err := loader.Load("bad-team")
	if err == nil {
		t.Fatal("Load(bad-team) error = nil, want parse error")
	}
}

func TestContextLoader_Caching(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	// First load
	ctx1, err := loader.Load("valid-team")
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}

	// Should be cached now
	if !loader.IsCached("valid-team") {
		t.Error("IsCached(valid-team) = false after Load, want true")
	}

	// Second load should return same instance (from cache)
	ctx2, err := loader.Load("valid-team")
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}

	if ctx1 != ctx2 {
		t.Error("second Load() returned different instance, want same (cached)")
	}

	// Invalidate
	loader.Invalidate("valid-team")

	if loader.IsCached("valid-team") {
		t.Error("IsCached(valid-team) = true after Invalidate, want false")
	}

	// Load again
	ctx3, err := loader.Load("valid-team")
	if err != nil {
		t.Fatalf("third Load() error = %v", err)
	}

	if ctx1 == ctx3 {
		t.Error("third Load() returned same instance after Invalidate, want different")
	}
}

func TestContextLoader_InvalidateAll(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	// Load multiple teams
	_, _ = loader.Load("valid-team")
	_, _ = loader.Load("minimal-team")

	if !loader.IsCached("valid-team") || !loader.IsCached("minimal-team") {
		t.Error("teams not cached after Load")
	}

	// Invalidate all
	loader.InvalidateAll()

	if loader.IsCached("valid-team") {
		t.Error("valid-team still cached after InvalidateAll")
	}
	if loader.IsCached("minimal-team") {
		t.Error("minimal-team still cached after InvalidateAll")
	}
}

func TestContextLoader_HasContextFile(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	// valid-team has context.yaml
	if !loader.HasContextFile("valid-team") {
		t.Error("HasContextFile(valid-team) = false, want true")
	}

	// minimal-team only has orchestrator.yaml
	if loader.HasContextFile("minimal-team") {
		t.Error("HasContextFile(minimal-team) = true, want false")
	}

	// non-existent team
	if loader.HasContextFile("non-existent") {
		t.Error("HasContextFile(non-existent) = true, want false")
	}
}

func TestContextLoader_GetContextPath(t *testing.T) {
	teamsDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(teamsDir, "")

	path := loader.GetContextPath("valid-team")
	expectedSuffix := filepath.Join("valid-team", "context.yaml")

	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("GetContextPath(valid-team) = %q, want suffix %q", path, expectedSuffix)
	}
}

func TestContextLoader_UserDirPriority(t *testing.T) {
	// Create temp directories for project and user teams
	projectDir := t.TempDir()
	userDir := t.TempDir()

	teamName := "priority-team"

	// Create team in both directories
	projectTeamDir := filepath.Join(projectDir, teamName)
	userTeamDir := filepath.Join(userDir, teamName)

	if err := os.MkdirAll(projectTeamDir, 0755); err != nil {
		t.Fatalf("failed to create project team dir: %v", err)
	}
	if err := os.MkdirAll(userTeamDir, 0755); err != nil {
		t.Fatalf("failed to create user team dir: %v", err)
	}

	// Write project context
	projectContext := `schema_version: "1.0"
team_name: priority-team
context_rows:
  - key: Source
    value: Project`

	// Write user context (should take priority)
	userContext := `schema_version: "1.0"
team_name: priority-team
context_rows:
  - key: Source
    value: User`

	if err := os.WriteFile(filepath.Join(projectTeamDir, "context.yaml"), []byte(projectContext), 0644); err != nil {
		t.Fatalf("failed to write project context: %v", err)
	}
	if err := os.WriteFile(filepath.Join(userTeamDir, "context.yaml"), []byte(userContext), 0644); err != nil {
		t.Fatalf("failed to write user context: %v", err)
	}

	loader := NewContextLoaderWithPaths(projectDir, userDir)

	ctx, err := loader.Load(teamName)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// User directory should take priority
	if ctx.GetRow("Source") != "User" {
		t.Errorf("GetRow(Source) = %q, want %q (user dir should take priority)", ctx.GetRow("Source"), "User")
	}
}

func TestContextLoader_SaveContext(t *testing.T) {
	teamsDir := t.TempDir()

	// Create team directory
	teamDir := filepath.Join(teamsDir, "save-team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		t.Fatalf("failed to create team dir: %v", err)
	}

	loader := NewContextLoaderWithPaths(teamsDir, "")

	ctx := NewTeamContext("save-team")
	ctx.DisplayName = "Save Test Team"
	ctx.Domain = "testing"
	ctx.AddRow("Key", "Value")

	if err := loader.SaveContext(ctx); err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Verify file was created
	contextPath := filepath.Join(teamDir, "context.yaml")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Fatalf("context.yaml not created at %s", contextPath)
	}

	// Load it back
	loaded, err := loader.Load("save-team")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.DisplayName != "Save Test Team" {
		t.Errorf("loaded DisplayName = %q, want %q", loaded.DisplayName, "Save Test Team")
	}

	if loaded.GetRow("Key") != "Value" {
		t.Errorf("loaded GetRow(Key) = %q, want %q", loaded.GetRow("Key"), "Value")
	}
}

func TestContextLoader_SaveContext_InvalidContext(t *testing.T) {
	teamsDir := t.TempDir()
	loader := NewContextLoaderWithPaths(teamsDir, "")

	// Context without required fields
	ctx := &TeamContext{}

	if err := loader.SaveContext(ctx); err == nil {
		t.Error("SaveContext() with invalid context should return error")
	}
}

func TestContextLoader_SaveContext_TeamNotFound(t *testing.T) {
	teamsDir := t.TempDir()
	loader := NewContextLoaderWithPaths(teamsDir, "")

	ctx := NewTeamContext("non-existent-team")

	if err := loader.SaveContext(ctx); err == nil {
		t.Error("SaveContext() with non-existent team should return error")
	}
}
