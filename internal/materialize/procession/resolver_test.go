package procession

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// isolateEnv prevents the resolver from finding platform-level processions
// via KNOSSOS_HOME or XDG_DATA_HOME during testing.
func isolateEnv(t *testing.T) {
	t.Helper()
	for _, env := range []string{"KNOSSOS_HOME", "XDG_DATA_HOME"} {
		orig := os.Getenv(env)
		t.Setenv(env, t.TempDir())
		_ = orig
	}
}

const validTemplate = `name: test-procession
description: "A test procession"
stations:
  - name: first
    rite: security
    goal: "Do the first thing"
    produces: [report]
  - name: second
    rite: hygiene
    goal: "Do the second thing"
    produces: [fixes]
artifact_dir: .sos/wip/test-procession/
`

const anotherTemplate = `name: another-procession
description: "Another test procession"
stations:
  - name: alpha
    rite: debt-triage
    goal: "Alpha step"
    produces: [inventory]
  - name: beta
    rite: hygiene
    goal: "Beta step"
    produces: [cleanup]
artifact_dir: .sos/wip/another-procession/
`

func TestResolveProcessions_EmptyDir(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "processions"), 0o755)

	results, err := ResolveProcessions(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestResolveProcessions_ProjectLevel(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	procDir := filepath.Join(dir, "processions")
	os.MkdirAll(procDir, 0o755)
	os.WriteFile(filepath.Join(procDir, "test-procession.yaml"), []byte(validTemplate), 0o644)

	results, err := ResolveProcessions(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "test-procession" {
		t.Errorf("expected name test-procession, got %s", results[0].Name)
	}
	if results[0].Source != "project" {
		t.Errorf("expected source project, got %s", results[0].Source)
	}
}

func TestResolveProcessions_EmbeddedFallback(t *testing.T) {
	isolateEnv(t)
	embFS := fstest.MapFS{
		"processions/test-procession.yaml": &fstest.MapFile{Data: []byte(validTemplate)},
	}

	results, err := ResolveProcessions("", embFS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Source != "embedded" {
		t.Errorf("expected source embedded, got %s", results[0].Source)
	}
}

func TestResolveProcessions_ProjectShadowsEmbedded(t *testing.T) {
	isolateEnv(t)
	// Embedded has test-procession
	embFS := fstest.MapFS{
		"processions/test-procession.yaml": &fstest.MapFile{Data: []byte(validTemplate)},
	}

	// Project also has test-procession (should win)
	dir := t.TempDir()
	procDir := filepath.Join(dir, "processions")
	os.MkdirAll(procDir, 0o755)
	os.WriteFile(filepath.Join(procDir, "test-procession.yaml"), []byte(validTemplate), 0o644)

	results, err := ResolveProcessions(dir, embFS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Source != "project" {
		t.Errorf("expected project to shadow embedded, got source %s", results[0].Source)
	}
}

func TestResolveProcessions_MultipleTemplates(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	procDir := filepath.Join(dir, "processions")
	os.MkdirAll(procDir, 0o755)
	os.WriteFile(filepath.Join(procDir, "test-procession.yaml"), []byte(validTemplate), 0o644)
	os.WriteFile(filepath.Join(procDir, "another-procession.yaml"), []byte(anotherTemplate), 0o644)

	results, err := ResolveProcessions(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	if !names["test-procession"] || !names["another-procession"] {
		t.Errorf("expected both templates, got %v", names)
	}
}

func TestResolveProcessions_InvalidTemplateSkipped(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	procDir := filepath.Join(dir, "processions")
	os.MkdirAll(procDir, 0o755)
	// Valid template
	os.WriteFile(filepath.Join(procDir, "good.yaml"), []byte(validTemplate), 0o644)
	// Invalid template (missing required fields)
	os.WriteFile(filepath.Join(procDir, "bad.yaml"), []byte("name: bad\n"), 0o644)

	results, err := ResolveProcessions(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (invalid skipped), got %d", len(results))
	}
	if results[0].Name != "test-procession" {
		t.Errorf("expected test-procession, got %s", results[0].Name)
	}
}

func TestResolveProcessions_GitkeepIgnored(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	procDir := filepath.Join(dir, "processions")
	os.MkdirAll(procDir, 0o755)
	os.WriteFile(filepath.Join(procDir, ".gitkeep"), []byte(""), 0o644)

	results, err := ResolveProcessions(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
