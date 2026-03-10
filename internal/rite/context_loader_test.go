package rite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// getTestDataPath returns the absolute path to testdata/rites.
func getTestDataPath(t *testing.T) string {
	t.Helper()
	testdataPath := filepath.Join("..", "..", "testdata", "rites")
	absPath, err := filepath.Abs(testdataPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}
	return absPath
}

func TestRiteContext_ToMarkdown(t *testing.T) {
	tests := []struct {
		name string
		ctx  *RiteContext
		want string
	}{
		{
			name: "empty context",
			ctx:  NewRiteContext("test-rite"),
			want: "",
		},
		{
			name: "nil context",
			ctx:  nil,
			want: "",
		},
		{
			name: "single row",
			ctx: &RiteContext{
				SchemaVersion: "1.0",
				RiteName:      "test-rite",
				ContextRows: []ContextRow{
					{Key: "Status", Value: "Active"},
				},
			},
			want: "| | |\n|---|---|\n| **Status** | Active |\n",
		},
		{
			name: "multiple rows",
			ctx: &RiteContext{
				SchemaVersion: "1.0",
				RiteName:      "test-rite",
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

func TestRiteContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ctx     *RiteContext
		wantErr bool
	}{
		{
			name: "valid context",
			ctx: &RiteContext{
				SchemaVersion: "1.0",
				RiteName:      "test-rite",
			},
			wantErr: false,
		},
		{
			name: "missing rite name",
			ctx: &RiteContext{
				SchemaVersion: "1.0",
			},
			wantErr: true,
		},
		{
			name: "missing schema version",
			ctx: &RiteContext{
				RiteName: "test-rite",
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

func TestRiteContext_AddRow(t *testing.T) {
	ctx := NewRiteContext("test-rite")

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

func TestRiteContext_HasRows(t *testing.T) {
	ctx := NewRiteContext("test-rite")

	if ctx.HasRows() {
		t.Error("HasRows() = true on new context, want false")
	}

	ctx.AddRow("Key", "Value")

	if !ctx.HasRows() {
		t.Error("HasRows() = false after AddRow, want true")
	}
}

func TestContextLoader_Load_FromYAML(t *testing.T) {
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	ctx, err := loader.Load("valid-rite")
	if err != nil {
		t.Fatalf("Load(valid-rite) error = %v", err)
	}

	if ctx.RiteName != "valid-rite" {
		t.Errorf("RiteName = %q, want %q", ctx.RiteName, "valid-rite")
	}

	if ctx.DisplayName != "Valid Test Rite" {
		t.Errorf("DisplayName = %q, want %q", ctx.DisplayName, "Valid Test Rite")
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
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	ctx, err := loader.Load("minimal-rite")
	if err != nil {
		t.Fatalf("Load(minimal-rite) error = %v", err)
	}

	if ctx.RiteName != "minimal-rite" {
		t.Errorf("RiteName = %q, want %q", ctx.RiteName, "minimal-rite")
	}

	if ctx.Domain != "testing" {
		t.Errorf("Domain = %q, want %q", ctx.Domain, "testing")
	}

	// Should have rows generated from orchestrator
	if !ctx.HasRows() {
		t.Error("HasRows() = false after fallback, want true")
	}
}

func TestContextLoader_Load_RiteNotFound(t *testing.T) {
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	_, err := loader.Load("non-existent-rite")
	if err == nil {
		t.Fatal("Load(non-existent-rite) error = nil, want error")
	}
}

func TestContextLoader_Load_EmptyRiteName(t *testing.T) {
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	_, err := loader.Load("")
	if err == nil {
		t.Fatal("Load('') error = nil, want error")
	}
}

func TestContextLoader_Load_MalformedYAML(t *testing.T) {
	// Create a temporary directory with malformed YAML
	tmpDir := t.TempDir()
	riteDir := filepath.Join(tmpDir, "bad-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatalf("failed to create rite dir: %v", err)
	}

	// Write malformed YAML
	malformedYAML := `schema_version: "1.0"
rite_name: bad-rite
context_rows:
  - key: "unclosed string
    value: broken`

	if err := os.WriteFile(filepath.Join(riteDir, "context.yaml"), []byte(malformedYAML), 0644); err != nil {
		t.Fatalf("failed to write malformed YAML: %v", err)
	}

	loader := NewContextLoaderWithPaths(tmpDir, "", "", "")

	_, err := loader.Load("bad-rite")
	if err == nil {
		t.Fatal("Load(bad-rite) error = nil, want parse error")
	}
}

func TestContextLoader_Caching(t *testing.T) {
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	// First load
	ctx1, err := loader.Load("valid-rite")
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}

	// Should be cached now
	if !loader.IsCached("valid-rite") {
		t.Error("IsCached(valid-rite) = false after Load, want true")
	}

	// Second load should return same instance (from cache)
	ctx2, err := loader.Load("valid-rite")
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}

	if ctx1 != ctx2 {
		t.Error("second Load() returned different instance, want same (cached)")
	}

	// Invalidate
	loader.Invalidate("valid-rite")

	if loader.IsCached("valid-rite") {
		t.Error("IsCached(valid-rite) = true after Invalidate, want false")
	}

	// Load again
	ctx3, err := loader.Load("valid-rite")
	if err != nil {
		t.Fatalf("third Load() error = %v", err)
	}

	if ctx1 == ctx3 {
		t.Error("third Load() returned same instance after Invalidate, want different")
	}
}

func TestContextLoader_InvalidateAll(t *testing.T) {
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	// Load multiple rites
	_, _ = loader.Load("valid-rite")
	_, _ = loader.Load("minimal-rite")

	if !loader.IsCached("valid-rite") || !loader.IsCached("minimal-rite") {
		t.Error("rites not cached after Load")
	}

	// Invalidate all
	loader.InvalidateAll()

	if loader.IsCached("valid-rite") {
		t.Error("valid-rite still cached after InvalidateAll")
	}
	if loader.IsCached("minimal-rite") {
		t.Error("minimal-rite still cached after InvalidateAll")
	}
}

func TestContextLoader_HasContextFile(t *testing.T) {
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	// valid-rite has context.yaml
	if !loader.HasContextFile("valid-rite") {
		t.Error("HasContextFile(valid-rite) = false, want true")
	}

	// minimal-rite only has orchestrator.yaml
	if loader.HasContextFile("minimal-rite") {
		t.Error("HasContextFile(minimal-rite) = true, want false")
	}

	// non-existent rite
	if loader.HasContextFile("non-existent") {
		t.Error("HasContextFile(non-existent) = true, want false")
	}
}

func TestContextLoader_GetContextPath(t *testing.T) {
	ritesDir := getTestDataPath(t)
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	path := loader.GetContextPath("valid-rite")
	expectedSuffix := filepath.Join("valid-rite", "context.yaml")

	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("GetContextPath(valid-rite) = %q, want suffix %q", path, expectedSuffix)
	}
}

func TestContextLoader_UserDirPriority(t *testing.T) {
	// Create temp directories for project and user rites
	projectDir := t.TempDir()
	userDir := t.TempDir()

	riteName := "priority-rite"

	// Create rite in both directories
	projectRiteDir := filepath.Join(projectDir, riteName)
	userRiteDir := filepath.Join(userDir, riteName)

	if err := os.MkdirAll(projectRiteDir, 0755); err != nil {
		t.Fatalf("failed to create project rite dir: %v", err)
	}
	if err := os.MkdirAll(userRiteDir, 0755); err != nil {
		t.Fatalf("failed to create user rite dir: %v", err)
	}

	// Write project context
	projectContext := `schema_version: "1.0"
rite_name: priority-rite
context_rows:
  - key: Source
    value: Project`

	// Write user context (should take priority)
	userContext := `schema_version: "1.0"
rite_name: priority-rite
context_rows:
  - key: Source
    value: User`

	if err := os.WriteFile(filepath.Join(projectRiteDir, "context.yaml"), []byte(projectContext), 0644); err != nil {
		t.Fatalf("failed to write project context: %v", err)
	}
	if err := os.WriteFile(filepath.Join(userRiteDir, "context.yaml"), []byte(userContext), 0644); err != nil {
		t.Fatalf("failed to write user context: %v", err)
	}

	loader := NewContextLoaderWithPaths(projectDir, userDir, "", "")

	ctx, err := loader.Load(riteName)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// User directory should take priority
	if ctx.GetRow("Source") != "User" {
		t.Errorf("GetRow(Source) = %q, want %q (user dir should take priority)", ctx.GetRow("Source"), "User")
	}
}

func TestContextLoader_SaveContext(t *testing.T) {
	ritesDir := t.TempDir()

	// Create rite directory
	riteDir := filepath.Join(ritesDir, "save-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatalf("failed to create rite dir: %v", err)
	}

	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	ctx := NewRiteContext("save-rite")
	ctx.DisplayName = "Save Test Rite"
	ctx.Domain = "testing"
	ctx.AddRow("Key", "Value")

	if err := loader.SaveContext(ctx); err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Verify file was created
	contextPath := filepath.Join(riteDir, "context.yaml")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Fatalf("context.yaml not created at %s", contextPath)
	}

	// Load it back
	loaded, err := loader.Load("save-rite")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.DisplayName != "Save Test Rite" {
		t.Errorf("loaded DisplayName = %q, want %q", loaded.DisplayName, "Save Test Rite")
	}

	if loaded.GetRow("Key") != "Value" {
		t.Errorf("loaded GetRow(Key) = %q, want %q", loaded.GetRow("Key"), "Value")
	}
}

func TestContextLoader_SaveContext_InvalidContext(t *testing.T) {
	ritesDir := t.TempDir()
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	// Context without required fields
	ctx := &RiteContext{}

	if err := loader.SaveContext(ctx); err == nil {
		t.Error("SaveContext() with invalid context should return error")
	}
}

func TestContextLoader_SaveContext_RiteNotFound(t *testing.T) {
	ritesDir := t.TempDir()
	loader := NewContextLoaderWithPaths(ritesDir, "", "", "")

	ctx := NewRiteContext("non-existent-rite")

	if err := loader.SaveContext(ctx); err == nil {
		t.Error("SaveContext() with non-existent rite should return error")
	}
}
