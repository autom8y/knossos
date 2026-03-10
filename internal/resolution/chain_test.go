package resolution

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// mkRiteDir creates a directory with a manifest.yaml marker file.
func mkRiteDir(t *testing.T, base, name string) {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte("name: "+name), 0644); err != nil {
		t.Fatal(err)
	}
}

// manifestValidator checks that manifest.yaml exists (disk or FS).
func manifestValidator(item ResolvedItem) bool {
	if item.Fsys != nil {
		_, err := fs.Stat(item.Fsys, filepath.Join(item.Path, "manifest.yaml"))
		return err == nil
	}
	_, err := os.Stat(filepath.Join(item.Path, "manifest.yaml"))
	return err == nil
}

// --- Resolve tests ---

func TestResolve_TopDownEarlyExit(t *testing.T) {
	projectDir := t.TempDir()
	platformDir := t.TempDir()
	mkRiteDir(t, projectDir, "security")
	mkRiteDir(t, platformDir, "security")

	chain := NewChain(
		Tier{Label: "project", Dir: projectDir},
		Tier{Label: "platform", Dir: platformDir},
	)

	item, err := chain.Resolve("security", manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if item.Source != "project" {
		t.Errorf("expected source 'project', got %q", item.Source)
	}
	if item.Name != "security" {
		t.Errorf("expected name 'security', got %q", item.Name)
	}
}

func TestResolve_FallsThrough(t *testing.T) {
	projectDir := t.TempDir() // empty — no rites
	platformDir := t.TempDir()
	mkRiteDir(t, platformDir, "security")

	chain := NewChain(
		Tier{Label: "project", Dir: projectDir},
		Tier{Label: "platform", Dir: platformDir},
	)

	item, err := chain.Resolve("security", manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if item.Source != "platform" {
		t.Errorf("expected source 'platform', got %q", item.Source)
	}
}

func TestResolve_NotFound(t *testing.T) {
	emptyDir := t.TempDir()
	chain := NewChain(
		Tier{Label: "project", Dir: emptyDir},
	)

	_, err := chain.Resolve("nonexistent", manifestValidator)
	if err == nil {
		t.Fatal("expected error for missing item")
	}
}

func TestResolve_FSFallback(t *testing.T) {
	emptyDisk := t.TempDir()
	embeddedFS := fstest.MapFS{
		"rites/security/manifest.yaml": &fstest.MapFile{Data: []byte("name: security")},
	}

	chain := NewChain(
		Tier{Label: "project", Dir: emptyDisk},
		Tier{Label: "embedded", Dir: "rites", FS: embeddedFS},
	)

	item, err := chain.Resolve("security", manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if item.Source != "embedded" {
		t.Errorf("expected source 'embedded', got %q", item.Source)
	}
	if item.Fsys == nil {
		t.Error("expected non-nil Fsys for embedded item")
	}
}

// --- ResolveAll tests ---

func TestResolveAll_HigherPriorityShadows(t *testing.T) {
	projectDir := t.TempDir()
	platformDir := t.TempDir()
	mkRiteDir(t, projectDir, "security")
	mkRiteDir(t, platformDir, "security")

	chain := NewChain(
		Tier{Label: "project", Dir: projectDir},
		Tier{Label: "platform", Dir: platformDir},
	)

	items, err := chain.ResolveAll(manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items["security"].Source != "project" {
		t.Errorf("expected 'project' to shadow 'platform', got %q", items["security"].Source)
	}
}

func TestResolveAll_MergesAcrossTiers(t *testing.T) {
	projectDir := t.TempDir()
	platformDir := t.TempDir()
	mkRiteDir(t, projectDir, "custom-rite")
	mkRiteDir(t, platformDir, "security")
	mkRiteDir(t, platformDir, "hygiene")

	chain := NewChain(
		Tier{Label: "project", Dir: projectDir},
		Tier{Label: "platform", Dir: platformDir},
	)

	items, err := chain.ResolveAll(manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items["custom-rite"].Source != "project" {
		t.Errorf("expected custom-rite from project, got %q", items["custom-rite"].Source)
	}
	if items["security"].Source != "platform" {
		t.Errorf("expected security from platform, got %q", items["security"].Source)
	}
}

func TestResolveAll_EmptyTiersSkipped(t *testing.T) {
	chain := NewChain(
		Tier{Label: "project", Dir: "/nonexistent/path/12345"},
		Tier{Label: "user", Dir: ""},
	)

	items, err := chain.ResolveAll(manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestResolveAll_FSEntries(t *testing.T) {
	embeddedFS := fstest.MapFS{
		"rites/security/manifest.yaml": &fstest.MapFile{Data: []byte("name: security")},
		"rites/hygiene/manifest.yaml":  &fstest.MapFile{Data: []byte("name: hygiene")},
	}

	chain := NewChain(
		Tier{Label: "embedded", Dir: "rites", FS: embeddedFS},
	)

	items, err := chain.ResolveAll(manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items["security"].Fsys == nil {
		t.Error("expected non-nil Fsys for embedded security")
	}
}

func TestEmptyChain(t *testing.T) {
	chain := NewChain()

	_, err := chain.Resolve("anything", manifestValidator)
	if err == nil {
		t.Error("expected error from empty chain Resolve")
	}

	items, err := chain.ResolveAll(manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items from empty chain, got %d", len(items))
	}
}

func TestValidateFilters(t *testing.T) {
	dir := t.TempDir()
	// Create two entries: one with manifest, one without
	mkRiteDir(t, dir, "valid-rite")
	if err := os.MkdirAll(filepath.Join(dir, "invalid-rite"), 0755); err != nil {
		t.Fatal(err)
	}

	chain := NewChain(Tier{Label: "test", Dir: dir})

	items, err := chain.ResolveAll(manifestValidator)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 valid item, got %d", len(items))
	}
	if _, ok := items["valid-rite"]; !ok {
		t.Error("expected 'valid-rite' to pass validation")
	}
}

// --- Builder tests ---

func TestRiteChain_TierOrder(t *testing.T) {
	chain := RiteChain("/project", "/user", "/org", "/platform", nil)
	tiers := chain.Tiers()
	if len(tiers) != 4 {
		t.Fatalf("expected 4 tiers (no embedded), got %d", len(tiers))
	}
	expected := []string{"project", "user", "org", "platform"}
	for i, want := range expected {
		if tiers[i].Label != want {
			t.Errorf("tier %d: expected label %q, got %q", i, want, tiers[i].Label)
		}
	}
}

func TestRiteChain_WithEmbedded(t *testing.T) {
	fs := fstest.MapFS{}
	chain := RiteChain("/project", "/user", "", "/platform", fs)
	tiers := chain.Tiers()
	// org skipped (empty), so: project, user, platform, embedded
	if len(tiers) != 4 {
		t.Fatalf("expected 4 tiers, got %d", len(tiers))
	}
	if tiers[3].Label != "embedded" {
		t.Errorf("expected last tier 'embedded', got %q", tiers[3].Label)
	}
}

func TestRiteChain_EmptyTiersSkipped(t *testing.T) {
	chain := RiteChain("", "/user", "", "", nil)
	tiers := chain.Tiers()
	if len(tiers) != 1 {
		t.Fatalf("expected 1 tier (user only), got %d", len(tiers))
	}
	if tiers[0].Label != "user" {
		t.Errorf("expected 'user', got %q", tiers[0].Label)
	}
}

func TestProcessionChain_TierOrder(t *testing.T) {
	chain := ProcessionChain("/proj/proc", "/user/proc", "/org/proc", "/plat/proc", nil)
	tiers := chain.Tiers()
	if len(tiers) != 4 {
		t.Fatalf("expected 4 tiers, got %d", len(tiers))
	}
	if tiers[0].Label != "project" || tiers[3].Label != "platform" {
		t.Errorf("unexpected tier order: %v", tiers)
	}
}

func TestContextChain_UserHighestPriority(t *testing.T) {
	chain := ContextChain("/user", "/project", "/org", "/platform")
	tiers := chain.Tiers()
	if len(tiers) != 4 {
		t.Fatalf("expected 4 tiers, got %d", len(tiers))
	}
	if tiers[0].Label != "user" {
		t.Errorf("expected user as highest priority, got %q", tiers[0].Label)
	}
	if tiers[1].Label != "project" {
		t.Errorf("expected project second, got %q", tiers[1].Label)
	}
}
