package procession

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/autom8y/knossos/internal/procession"
)

// stubRender is a minimal RenderFunc for testing.
func stubRender(projectRoot, templateName string, data any) ([]byte, error) {
	return []byte("# stub " + templateName), nil
}

// writeTestTemplate writes a procession template YAML to projectDir/processions/.
func writeTestTemplate(t *testing.T, projectDir, name, entryRite string) {
	t.Helper()
	dir := filepath.Join(projectDir, "processions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll processions/: %v", err)
	}
	content := `name: ` + name + `
description: "Test template"
stations:
  - name: start
    rite: ` + entryRite + `
    goal: "Start"
    produces: [artifact]
  - name: finish
    rite: other
    goal: "Finish"
    produces: [result]
artifact_dir: .sos/wip/` + name + `/
`
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write template %s: %v", name, err)
	}
}

// makeTestProcession creates a ResolvedProcession for rendering tests.
// This avoids filesystem setup and global state mutation.
func makeTestProcession(name, entryRite string) ResolvedProcession {
	return ResolvedProcession{
		Name:   name,
		Source: "project",
		Template: &procession.Template{
			Name:        name,
			Description: "Test template",
			Stations: []procession.Station{
				{Name: "start", Rite: entryRite, Goal: "Start", Produces: []string{"artifact"}},
				{Name: "finish", Rite: "other", Goal: "Finish", Produces: []string{"result"}},
			},
			ArtifactDir: ".sos/wip/" + name + "/",
		},
	}
}

// --- ResolveTemplate tests ---

func TestResolveTemplate_Found(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()
	writeTestTemplate(t, projectDir, "my-workflow", "security")

	procDir := filepath.Join(projectDir, "processions")
	rp, err := ResolveTemplateWithDirs("my-workflow", procDir, "", "", "", nil)
	if err != nil {
		t.Fatalf("ResolveTemplateWithDirs: %v", err)
	}
	if rp.Name != "my-workflow" {
		t.Errorf("Name = %q, want %q", rp.Name, "my-workflow")
	}
	if rp.Source != "project" {
		t.Errorf("Source = %q, want %q", rp.Source, "project")
	}
	if rp.Template == nil {
		t.Fatal("Template is nil")
	}
	if len(rp.Template.Stations) != 2 {
		t.Errorf("Stations = %d, want 2", len(rp.Template.Stations))
	}
}

func TestResolveTemplate_NotFound(t *testing.T) {
	t.Parallel()
	projectDir := t.TempDir()

	procDir := filepath.Join(projectDir, "processions")
	_, err := ResolveTemplateWithDirs("nonexistent", procDir, "", "", "", nil)
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention template name, got: %v", err)
	}
}

func TestResolveTemplate_ProjectShadowsPlatform(t *testing.T) {
	t.Parallel()
	// Write template at platform tier
	platformDir := t.TempDir()
	writeTestTemplate(t, platformDir, "my-workflow", "platform-rite")

	// Write template at project tier (higher priority)
	projectDir := t.TempDir()
	writeTestTemplate(t, projectDir, "my-workflow", "project-rite")

	projectProcDir := filepath.Join(projectDir, "processions")
	platformProcDir := filepath.Join(platformDir, "processions")
	rp, err := ResolveTemplateWithDirs("my-workflow", projectProcDir, "", "", platformProcDir, nil)
	if err != nil {
		t.Fatalf("ResolveTemplateWithDirs: %v", err)
	}
	if rp.Source != "project" {
		t.Errorf("Source = %q, want %q (project should shadow platform)", rp.Source, "project")
	}
	if rp.Template.Stations[0].Rite != "project-rite" {
		t.Errorf("Rite = %q, want %q", rp.Template.Stations[0].Rite, "project-rite")
	}
}

func TestResolveTemplate_PlatformFallback(t *testing.T) {
	t.Parallel()
	// Write template at platform tier only
	platformDir := t.TempDir()
	writeTestTemplate(t, platformDir, "platform-only", "security")

	// Empty project dir (no templates)
	projectDir := t.TempDir()

	projectProcDir := filepath.Join(projectDir, "processions")
	platformProcDir := filepath.Join(platformDir, "processions")
	rp, err := ResolveTemplateWithDirs("platform-only", projectProcDir, "", "", platformProcDir, nil)
	if err != nil {
		t.Fatalf("ResolveTemplateWithDirs: %v", err)
	}
	if rp.Source != "platform" {
		t.Errorf("Source = %q, want %q", rp.Source, "platform")
	}
}

// --- RenderToDir tests ---

func TestRenderToDir_MatchingRite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	processions := []ResolvedProcession{
		makeTestProcession("my-workflow", "security"),
	}
	count, err := RenderToDirWithProcessions("", tmpDir, stubRender, "security", processions)
	if err != nil {
		t.Fatalf("RenderToDirWithProcessions: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Both dromena and legomena should exist
	droPath := filepath.Join(tmpDir, "my-workflow", "INDEX.dro.md")
	if _, err := os.Stat(droPath); os.IsNotExist(err) {
		t.Errorf("dromena not created at %s", droPath)
	}

	legoPath := filepath.Join(tmpDir, "my-workflow-ref", "INDEX.lego.md")
	if _, err := os.Stat(legoPath); os.IsNotExist(err) {
		t.Errorf("legomena not created at %s", legoPath)
	}
}

func TestRenderToDir_NonMatchingRite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	processions := []ResolvedProcession{
		makeTestProcession("my-workflow", "security"),
	}
	count, err := RenderToDirWithProcessions("", tmpDir, stubRender, "docs", processions)
	if err != nil {
		t.Fatalf("RenderToDirWithProcessions: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (legomena still counts)", count)
	}

	// Dromena should NOT exist
	droPath := filepath.Join(tmpDir, "my-workflow", "INDEX.dro.md")
	if _, err := os.Stat(droPath); !os.IsNotExist(err) {
		t.Errorf("dromena should NOT be created for non-matching rite, but found at %s", droPath)
	}

	// Legomena should still exist (skills are universal)
	legoPath := filepath.Join(tmpDir, "my-workflow-ref", "INDEX.lego.md")
	if _, err := os.Stat(legoPath); os.IsNotExist(err) {
		t.Errorf("legomena should be created regardless of rite, not found at %s", legoPath)
	}
}

func TestRenderToDir_EmptyRite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	processions := []ResolvedProcession{
		makeTestProcession("my-workflow", "security"),
	}
	count, err := RenderToDirWithProcessions("", tmpDir, stubRender, "", processions)
	if err != nil {
		t.Fatalf("RenderToDirWithProcessions: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Dromena should NOT exist (empty rite = minimal mode, legomena only)
	droPath := filepath.Join(tmpDir, "my-workflow", "INDEX.dro.md")
	if _, err := os.Stat(droPath); !os.IsNotExist(err) {
		t.Errorf("dromena should NOT be created in minimal mode (empty rite), but found at %s", droPath)
	}

	// Legomena should exist
	legoPath := filepath.Join(tmpDir, "my-workflow-ref", "INDEX.lego.md")
	if _, err := os.Stat(legoPath); os.IsNotExist(err) {
		t.Errorf("legomena should be created in minimal mode, not found at %s", legoPath)
	}
}

func TestRenderToDir_MultipleTemplates(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	processions := []ResolvedProcession{
		makeTestProcession("sec-workflow", "security"),
		makeTestProcession("doc-workflow", "docs"),
	}
	count, err := RenderToDirWithProcessions("", tmpDir, stubRender, "security", processions)
	if err != nil {
		t.Fatalf("RenderToDirWithProcessions: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2 (both templates render legomena)", count)
	}

	// sec-workflow dromena SHOULD exist (matching rite)
	secDro := filepath.Join(tmpDir, "sec-workflow", "INDEX.dro.md")
	if _, err := os.Stat(secDro); os.IsNotExist(err) {
		t.Errorf("sec-workflow dromena should exist (matching rite)")
	}

	// doc-workflow dromena should NOT exist (non-matching rite)
	docDro := filepath.Join(tmpDir, "doc-workflow", "INDEX.dro.md")
	if _, err := os.Stat(docDro); !os.IsNotExist(err) {
		t.Errorf("doc-workflow dromena should NOT exist (non-matching rite)")
	}

	// Both legomena should exist
	secLego := filepath.Join(tmpDir, "sec-workflow-ref", "INDEX.lego.md")
	if _, err := os.Stat(secLego); os.IsNotExist(err) {
		t.Errorf("sec-workflow legomena should exist")
	}

	docLego := filepath.Join(tmpDir, "doc-workflow-ref", "INDEX.lego.md")
	if _, err := os.Stat(docLego); os.IsNotExist(err) {
		t.Errorf("doc-workflow legomena should exist")
	}
}
