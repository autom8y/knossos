package procession

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/config"
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

// isolateHome overrides KNOSSOS_HOME to prevent picking up real templates.
func isolateHome(t *testing.T) {
	t.Helper()
	config.ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", t.TempDir())
	t.Cleanup(config.ResetKnossosHome)
}

func TestRenderToDir_MatchingRite(t *testing.T) {
	isolateHome(t)
	projectDir := t.TempDir()
	tmpDir := t.TempDir()
	writeTestTemplate(t, projectDir, "my-workflow", "security")

	count, err := RenderToDir(projectDir, tmpDir, stubRender, "security")
	if err != nil {
		t.Fatalf("RenderToDir: %v", err)
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
	isolateHome(t)
	projectDir := t.TempDir()
	tmpDir := t.TempDir()
	writeTestTemplate(t, projectDir, "my-workflow", "security")

	count, err := RenderToDir(projectDir, tmpDir, stubRender, "docs")
	if err != nil {
		t.Fatalf("RenderToDir: %v", err)
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
	isolateHome(t)
	projectDir := t.TempDir()
	tmpDir := t.TempDir()
	writeTestTemplate(t, projectDir, "my-workflow", "security")

	count, err := RenderToDir(projectDir, tmpDir, stubRender, "")
	if err != nil {
		t.Fatalf("RenderToDir: %v", err)
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
	isolateHome(t)
	projectDir := t.TempDir()
	tmpDir := t.TempDir()
	writeTestTemplate(t, projectDir, "sec-workflow", "security")
	writeTestTemplate(t, projectDir, "doc-workflow", "docs")

	count, err := RenderToDir(projectDir, tmpDir, stubRender, "security")
	if err != nil {
		t.Fatalf("RenderToDir: %v", err)
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
