package source

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestSourceResolver_EmbeddedFallback(t *testing.T) {
	// Create in-memory fs.FS with a test rite
	fsys := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: test-rite\nversion: 1.0\n"),
		},
		"rites/test-rite/agents/agent-one.md": &fstest.MapFile{
			Data: []byte("# Agent One\n"),
		},
	}

	resolver := NewSourceResolverWithPaths("/nonexistent-project", "", "", "")
	resolver.WithEmbeddedFS(fsys)

	// Should find rite in embedded FS when filesystem sources don't exist
	resolved, err := resolver.ResolveRite("test-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}

	if resolved.Source.Type != SourceEmbedded {
		t.Errorf("Expected source type %q, got %q", SourceEmbedded, resolved.Source.Type)
	}
	if resolved.Name != "test-rite" {
		t.Errorf("Expected rite name %q, got %q", "test-rite", resolved.Name)
	}
	if resolved.RitePath != "rites/test-rite" {
		t.Errorf("Expected rite path %q, got %q", "rites/test-rite", resolved.RitePath)
	}
	if resolved.ManifestPath != "rites/test-rite/manifest.yaml" {
		t.Errorf("Expected manifest path %q, got %q", "rites/test-rite/manifest.yaml", resolved.ManifestPath)
	}
	if resolved.TemplatesDir != "knossos/templates" {
		t.Errorf("Expected templates dir %q, got %q", "knossos/templates", resolved.TemplatesDir)
	}
}

func TestSourceResolver_FilesystemOverridesEmbedded(t *testing.T) {
	// Create a temp directory with a project rite
	tmpDir := t.TempDir()
	createTestRite(t, tmpDir, "test-rite")

	// Create embedded FS with the same rite
	fsys := fstest.MapFS{
		"rites/test-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: test-rite\nversion: 2.0\n"),
		},
	}

	resolver := NewSourceResolverWithPaths(tmpDir, "", "", "")
	resolver.WithEmbeddedFS(fsys)

	resolved, err := resolver.ResolveRite("test-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}

	// Filesystem (project) should override embedded
	if resolved.Source.Type != SourceProject {
		t.Errorf("Expected source type %q (filesystem wins), got %q", SourceProject, resolved.Source.Type)
	}
}

func TestSourceResolver_EmbeddedNotFoundReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/other-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: other-rite\nversion: 1.0\n"),
		},
	}

	resolver := NewSourceResolverWithPaths("/nonexistent-project", "", "", "")
	resolver.WithEmbeddedFS(fsys)

	_, err := resolver.ResolveRite("missing-rite", "")
	if err == nil {
		t.Fatal("Expected error for missing rite, got nil")
	}
}

func TestSourceResolver_NoEmbeddedFS(t *testing.T) {
	resolver := NewSourceResolverWithPaths("/nonexistent-project", "", "", "")
	// No embedded FS set

	_, err := resolver.ResolveRite("any-rite", "")
	if err == nil {
		t.Fatal("Expected error when no sources available, got nil")
	}
}

func TestSourceResolver_ListIncludesEmbedded(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/embedded-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: embedded-rite\nversion: 1.0\n"),
		},
		"rites/another-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: another-rite\nversion: 1.0\n"),
		},
	}

	resolver := NewSourceResolverWithPaths("/nonexistent-project", "", "", "")
	resolver.WithEmbeddedFS(fsys)

	rites, err := resolver.ListAvailableRites()
	if err != nil {
		t.Fatalf("ListAvailableRites failed: %v", err)
	}

	if len(rites) != 2 {
		t.Fatalf("Expected 2 rites, got %d", len(rites))
	}

	// Both should be embedded source type
	for _, r := range rites {
		if r.Source.Type != SourceEmbedded {
			t.Errorf("Expected source type %q for rite %q, got %q", SourceEmbedded, r.Name, r.Source.Type)
		}
	}
}

func TestSourceResolver_ListShadowsEmbedded(t *testing.T) {
	// Create a temp dir with a project rite that shadows an embedded one
	tmpDir := t.TempDir()
	createTestRite(t, tmpDir, "shadowed-rite")

	fsys := fstest.MapFS{
		"rites/shadowed-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: shadowed-rite\nversion: 2.0\n"),
		},
		"rites/embedded-only/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: embedded-only\nversion: 1.0\n"),
		},
	}

	resolver := NewSourceResolverWithPaths(tmpDir, "", "", "")
	resolver.WithEmbeddedFS(fsys)

	rites, err := resolver.ListAvailableRites()
	if err != nil {
		t.Fatalf("ListAvailableRites failed: %v", err)
	}

	if len(rites) != 2 {
		t.Fatalf("Expected 2 rites (1 project + 1 embedded-only), got %d", len(rites))
	}

	// Find the shadowed rite and verify it comes from project, not embedded
	for _, r := range rites {
		if r.Name == "shadowed-rite" && r.Source.Type != SourceProject {
			t.Errorf("Expected project source for shadowed-rite, got %q", r.Source.Type)
		}
		if r.Name == "embedded-only" && r.Source.Type != SourceEmbedded {
			t.Errorf("Expected embedded source for embedded-only, got %q", r.Source.Type)
		}
	}
}

func TestSourceResolver_EmbeddedCaching(t *testing.T) {
	fsys := fstest.MapFS{
		"rites/cached-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: cached-rite\nversion: 1.0\n"),
		},
	}

	resolver := NewSourceResolverWithPaths("/nonexistent-project", "", "", "")
	resolver.WithEmbeddedFS(fsys)

	// First resolution
	resolved1, err := resolver.ResolveRite("cached-rite", "")
	if err != nil {
		t.Fatalf("First resolve failed: %v", err)
	}

	// Second resolution should use cache
	resolved2, err := resolver.ResolveRite("cached-rite", "")
	if err != nil {
		t.Fatalf("Second resolve failed: %v", err)
	}

	if resolved1 != resolved2 {
		t.Error("Expected cached result to return same pointer")
	}
}

// createTestRite creates a minimal rite directory structure for testing.
func createTestRite(t *testing.T, baseDir, riteName string) {
	t.Helper()
	riteDir := filepath.Join(baseDir, ".knossos", "rites", riteName)
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatalf("Failed to create rite dir: %v", err)
	}
	manifest := "name: " + riteName + "\nversion: 1.0\n"
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}
}

// createOrgRite creates a minimal rite in an org directory structure for testing.
func createOrgRite(t *testing.T, orgDir, riteName string) {
	t.Helper()
	riteDir := filepath.Join(orgDir, "rites", riteName)
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatalf("Failed to create org rite dir: %v", err)
	}
	manifest := "name: " + riteName + "\nversion: 1.0\n"
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed to write org manifest: %v", err)
	}
}

// createUserRite creates a minimal rite in a user-level directory for testing.
func createUserRite(t *testing.T, userRitesDir, riteName string) {
	t.Helper()
	riteDir := filepath.Join(userRitesDir, riteName)
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatalf("Failed to create user rite dir: %v", err)
	}
	manifest := "name: " + riteName + "\nversion: 1.0\n"
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed to write user manifest: %v", err)
	}
}

// --- Org Tier Tests (6-Tier Resolution) ---

func TestSourceResolver_OrgTierResolvesRite(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "org")
	createOrgRite(t, orgDir, "org-rite")

	resolver := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	resolved, err := resolver.ResolveRite("org-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}
	if resolved.Source.Type != SourceOrg {
		t.Errorf("Expected source type %q, got %q", SourceOrg, resolved.Source.Type)
	}
	if resolved.Name != "org-rite" {
		t.Errorf("Expected rite name %q, got %q", "org-rite", resolved.Name)
	}
}

func TestSourceResolver_UserShadowsOrg(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "org")
	userDir := filepath.Join(tmpDir, "user-rites")

	// Same rite name in both org and user
	createOrgRite(t, orgDir, "shared-rite")
	createUserRite(t, userDir, "shared-rite")

	resolver := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    userDir,
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	resolved, err := resolver.ResolveRite("shared-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}
	// User (tier 3) should shadow org (tier 4)
	if resolved.Source.Type != SourceUser {
		t.Errorf("Expected user to shadow org, got source type %q", resolved.Source.Type)
	}
}

func TestSourceResolver_OrgShadowsKnossos(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "org")
	knossosDir := filepath.Join(tmpDir, "knossos")

	createOrgRite(t, orgDir, "shared-rite")

	// Create knossos rite
	riteDir := filepath.Join(knossosDir, "rites", "shared-rite")
	if err := os.MkdirAll(riteDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(riteDir, "manifest.yaml"), []byte("name: shared-rite\nversion: 1.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     knossosDir,
		resolved:        make(map[string]*ResolvedRite),
	}

	resolved, err := resolver.ResolveRite("shared-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}
	// Org (tier 4) should shadow knossos (tier 5)
	if resolved.Source.Type != SourceOrg {
		t.Errorf("Expected org to shadow knossos, got source type %q", resolved.Source.Type)
	}
}

func TestSourceResolver_ProjectShadowsAll(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "org")
	userDir := filepath.Join(tmpDir, "user-rites")

	// Same rite in project, user, and org
	createTestRite(t, tmpDir, "shared-rite")
	createUserRite(t, userDir, "shared-rite")
	createOrgRite(t, orgDir, "shared-rite")

	resolver := &SourceResolver{
		projectRoot:     tmpDir,
		projectRitesDir: filepath.Join(tmpDir, ".knossos", "rites"),
		userRitesDir:    userDir,
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	resolved, err := resolver.ResolveRite("shared-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}
	// Project (tier 2) should shadow all
	if resolved.Source.Type != SourceProject {
		t.Errorf("Expected project to shadow all, got source type %q", resolved.Source.Type)
	}
}

func TestSourceResolver_ListIncludesOrg(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "org")

	// Create org-only rite
	createOrgRite(t, orgDir, "org-only-rite")

	// Create embedded rite
	fsys := fstest.MapFS{
		"rites/embedded-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: embedded-rite\nversion: 1.0\n"),
		},
	}

	resolver := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		EmbeddedFS:      fsys,
		resolved:        make(map[string]*ResolvedRite),
	}

	rites, err := resolver.ListAvailableRites()
	if err != nil {
		t.Fatalf("ListAvailableRites failed: %v", err)
	}

	if len(rites) != 2 {
		t.Fatalf("Expected 2 rites (org + embedded), got %d", len(rites))
	}

	found := map[string]SourceType{}
	for _, r := range rites {
		found[r.Name] = r.Source.Type
	}
	if found["org-only-rite"] != SourceOrg {
		t.Errorf("Expected org-only-rite to have source type %q, got %q", SourceOrg, found["org-only-rite"])
	}
	if found["embedded-rite"] != SourceEmbedded {
		t.Errorf("Expected embedded-rite to have source type %q, got %q", SourceEmbedded, found["embedded-rite"])
	}
}

func TestSourceResolver_ListOrgShadowedByUser(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "org")
	userDir := filepath.Join(tmpDir, "user-rites")

	// Same rite in org and user
	createOrgRite(t, orgDir, "overlap-rite")
	createUserRite(t, userDir, "overlap-rite")

	// Org-only rite
	createOrgRite(t, orgDir, "org-exclusive")

	resolver := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    userDir,
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	rites, err := resolver.ListAvailableRites()
	if err != nil {
		t.Fatalf("ListAvailableRites failed: %v", err)
	}

	if len(rites) != 2 {
		t.Fatalf("Expected 2 rites (overlap from user + org-exclusive from org), got %d", len(rites))
	}

	found := map[string]SourceType{}
	for _, r := range rites {
		found[r.Name] = r.Source.Type
	}
	// overlap-rite should come from user (higher priority)
	if found["overlap-rite"] != SourceUser {
		t.Errorf("Expected overlap-rite from user, got %q", found["overlap-rite"])
	}
	if found["org-exclusive"] != SourceOrg {
		t.Errorf("Expected org-exclusive from org, got %q", found["org-exclusive"])
	}
}

func TestSourceResolver_CacheOrgAware(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir1 := filepath.Join(tmpDir, "org1")
	orgDir2 := filepath.Join(tmpDir, "org2")

	createOrgRite(t, orgDir1, "my-rite")
	createOrgRite(t, orgDir2, "my-rite")

	// Resolve with org1
	resolver1 := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     filepath.Join(orgDir1, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	res1, err := resolver1.ResolveRite("my-rite", "")
	if err != nil {
		t.Fatalf("Resolve with org1 failed: %v", err)
	}

	// Simulate org switch by creating resolver with different orgRitesDir
	resolver2 := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     filepath.Join(orgDir2, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        resolver1.resolved, // Share cache to prove keys differ
	}

	res2, err := resolver2.ResolveRite("my-rite", "")
	if err != nil {
		t.Fatalf("Resolve with org2 failed: %v", err)
	}

	// Different org dirs should produce different cache entries (not stale)
	if res1.Source.Path == res2.Source.Path {
		t.Error("Expected different source paths for different orgs, got same")
	}
}

func TestSourceResolver_EmptyOrgSkipped(t *testing.T) {
	// When no org is configured, orgRitesDir is empty — should be skipped
	fsys := fstest.MapFS{
		"rites/fallback-rite/manifest.yaml": &fstest.MapFile{
			Data: []byte("name: fallback-rite\nversion: 1.0\n"),
		},
	}

	resolver := NewSourceResolverWithPaths("/nonexistent-project", "", "", "")
	resolver.WithEmbeddedFS(fsys)

	resolved, err := resolver.ResolveRite("fallback-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}
	// Should fall through to embedded since no org is configured
	if resolved.Source.Type != SourceEmbedded {
		t.Errorf("Expected embedded fallback when no org configured, got %q", resolved.Source.Type)
	}
}

func TestSourceResolver_OrgNoTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "org")
	createOrgRite(t, orgDir, "org-rite")

	resolver := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	resolved, err := resolver.ResolveRite("org-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}
	// Org rites should not carry templates
	if resolved.TemplatesDir != "" {
		t.Errorf("Expected empty templates dir for org rite, got %q", resolved.TemplatesDir)
	}
}

func TestSourceResolver_ExplicitOrgAlias(t *testing.T) {
	tmpDir := t.TempDir()
	orgDir := filepath.Join(tmpDir, "orgs", "test-org")
	createOrgRite(t, orgDir, "aliased-rite")

	// Set up XDG paths to point to our temp dir
	t.Setenv("XDG_DATA_HOME", tmpDir)

	resolver := &SourceResolver{
		projectRoot:     "/nonexistent-project",
		projectRitesDir: "/nonexistent-project/.knossos/rites",
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     filepath.Join(orgDir, "rites"),
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	// "org:test-org" should resolve to the named org's rites dir
	source, err := resolver.parseExplicitSource("org:test-org")
	if err != nil {
		t.Fatalf("parseExplicitSource(org:test-org) failed: %v", err)
	}
	if source.Type != SourceOrg {
		t.Errorf("Expected source type %q, got %q", SourceOrg, source.Type)
	}
}

func TestSourceResolver_ExplicitOrgNoOrg(t *testing.T) {
	t.Setenv("KNOSSOS_ORG", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // Prevent ActiveOrg() from reading active-org file

	resolver := &SourceResolver{
		resolved: make(map[string]*ResolvedRite),
	}

	// "org" without active org should error
	_, err := resolver.parseExplicitSource("org")
	if err == nil {
		t.Fatal("Expected error when using 'org' alias with no active org")
	}
}

// --- SCAR Regression Tests ---

// TestSCAR023_TemplatePathResolution_SelfHosting is a regression test for SCAR-023.
//
// Background: SourceProject template resolution looked only at $PROJECT/templates/.
// In the knossos self-hosting case (where knossos is the project), templates live at
// $PROJECT/knossos/templates/. The fix added a fallback: when templates/sections/
// does not exist, check knossos/templates/sections/ as an alternative.
//
// This test creates a project structure that mimics the knossos self-hosting layout
// (knossos/templates/sections/ exists but templates/sections/ does not) and verifies
// that the resolver correctly falls back to the knossos/templates/ path.
func TestSCAR023_TemplatePathResolution_SelfHosting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the knossos self-hosting template layout:
	// knossos/templates/sections/ exists, but templates/sections/ does NOT.
	knossosTemplatesDir := filepath.Join(tmpDir, "knossos", "templates")
	if err := os.MkdirAll(filepath.Join(knossosTemplatesDir, "sections"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a project-level rite (satellite local)
	createTestRite(t, tmpDir, "self-hosted-rite")

	resolver := &SourceResolver{
		projectRoot:     tmpDir,
		projectRitesDir: filepath.Join(tmpDir, ".knossos", "rites"),
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     "",
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	resolved, err := resolver.ResolveRite("self-hosted-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}

	if resolved.Source.Type != SourceProject {
		t.Errorf("Expected source type %q, got %q", SourceProject, resolved.Source.Type)
	}

	// The critical assertion: templates dir must resolve to knossos/templates/
	// because templates/sections/ does not exist (self-hosting fallback).
	if resolved.TemplatesDir != knossosTemplatesDir {
		t.Errorf("SCAR-023 regression: TemplatesDir = %q, want %q. "+
			"When templates/sections/ does not exist but knossos/templates/sections/ does, "+
			"the resolver must fall back to knossos/templates/. "+
			"See commit bff1293.", resolved.TemplatesDir, knossosTemplatesDir)
	}
}

// TestSCAR023_TemplatePathResolution_StandardProject is the complementary test:
// when templates/sections/ exists at the standard location, the resolver must
// use it directly without fallback.
func TestSCAR023_TemplatePathResolution_StandardProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both template locations (standard takes priority)
	standardTemplatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(filepath.Join(standardTemplatesDir, "sections"), 0755); err != nil {
		t.Fatal(err)
	}
	// Also create knossos/templates/ to verify it is NOT used when standard exists
	if err := os.MkdirAll(filepath.Join(tmpDir, "knossos", "templates", "sections"), 0755); err != nil {
		t.Fatal(err)
	}

	createTestRite(t, tmpDir, "standard-rite")

	resolver := &SourceResolver{
		projectRoot:     tmpDir,
		projectRitesDir: filepath.Join(tmpDir, ".knossos", "rites"),
		userRitesDir:    "/nonexistent-user-rites",
		orgRitesDir:     "",
		knossosHome:     "/nonexistent-knossos-home",
		resolved:        make(map[string]*ResolvedRite),
	}

	resolved, err := resolver.ResolveRite("standard-rite", "")
	if err != nil {
		t.Fatalf("ResolveRite failed: %v", err)
	}

	// Standard templates/ must be used when it exists (no fallback needed)
	if resolved.TemplatesDir != standardTemplatesDir {
		t.Errorf("TemplatesDir = %q, want %q (standard path should take priority)", resolved.TemplatesDir, standardTemplatesDir)
	}
}
