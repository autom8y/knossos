package mena

import (
	"path/filepath"
	"testing"
)

func TestBuildSourceChain_FullChain(t *testing.T) {
	opts := SourceChainOptions{
		RitePath:        "/rites/myrite",
		RitesBase:       "/rites",
		Dependencies:    []string{"other"},
		PlatformMenaDir: "/platform/mena",
	}
	sources := BuildSourceChain(opts)

	// Expected order: platform, shared, dep, rite-local
	if len(sources) != 4 {
		t.Fatalf("expected 4 sources, got %d: %+v", len(sources), sources)
	}
	if sources[0].Path != "/platform/mena" {
		t.Errorf("sources[0] should be platform, got %q", sources[0].Path)
	}
	if sources[1].Path != filepath.Join("/rites", "shared", "mena") {
		t.Errorf("sources[1] should be shared, got %q", sources[1].Path)
	}
	if sources[2].Path != filepath.Join("/rites", "other", "mena") {
		t.Errorf("sources[2] should be dep, got %q", sources[2].Path)
	}
	if sources[3].Path != filepath.Join("/rites/myrite", "mena") {
		t.Errorf("sources[3] should be rite-local, got %q", sources[3].Path)
	}
}

func TestBuildSourceChain_EmptyRitesBase(t *testing.T) {
	opts := SourceChainOptions{
		RitePath:        "/rites/myrite",
		RitesBase:       "",
		PlatformMenaDir: "/platform/mena",
	}
	sources := BuildSourceChain(opts)

	// Expected: platform + rite-local only (no shared, no deps)
	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d: %+v", len(sources), sources)
	}
	if sources[0].Path != "/platform/mena" {
		t.Errorf("sources[0] should be platform, got %q", sources[0].Path)
	}
	if sources[1].Path != filepath.Join("/rites/myrite", "mena") {
		t.Errorf("sources[1] should be rite-local, got %q", sources[1].Path)
	}
}

func TestBuildSourceChain_NoPlatform(t *testing.T) {
	opts := SourceChainOptions{
		RitePath:        "/rites/myrite",
		RitesBase:       "/rites",
		Dependencies:    []string{"other"},
		PlatformMenaDir: "",
	}
	sources := BuildSourceChain(opts)

	// Expected: shared + dep + rite-local (no platform)
	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d: %+v", len(sources), sources)
	}
	if sources[0].Path != filepath.Join("/rites", "shared", "mena") {
		t.Errorf("sources[0] should be shared, got %q", sources[0].Path)
	}
	if sources[1].Path != filepath.Join("/rites", "other", "mena") {
		t.Errorf("sources[1] should be dep, got %q", sources[1].Path)
	}
	if sources[2].Path != filepath.Join("/rites/myrite", "mena") {
		t.Errorf("sources[2] should be rite-local, got %q", sources[2].Path)
	}
}

func TestBuildSourceChain_SharedFilteredFromDeps(t *testing.T) {
	// "shared" in dependencies list should not result in a duplicate shared entry
	opts := SourceChainOptions{
		RitePath:        "/rites/myrite",
		RitesBase:       "/rites",
		Dependencies:    []string{"shared", "other"},
		PlatformMenaDir: "/platform/mena",
	}
	sources := BuildSourceChain(opts)

	// Expected: platform, shared (implicit), other (not "shared"), rite-local
	// "shared" from Dependencies is filtered; shared appears once at position 1
	if len(sources) != 4 {
		t.Fatalf("expected 4 sources, got %d: %+v", len(sources), sources)
	}

	// Verify "shared" appears exactly once
	sharedCount := 0
	for _, s := range sources {
		if s.Path == filepath.Join("/rites", "shared", "mena") {
			sharedCount++
		}
	}
	if sharedCount != 1 {
		t.Errorf("expected shared to appear exactly once, got %d times", sharedCount)
	}

	// Verify "other" appears
	otherFound := false
	for _, s := range sources {
		if s.Path == filepath.Join("/rites", "other", "mena") {
			otherFound = true
		}
	}
	if !otherFound {
		t.Error("expected 'other' dep to appear in sources")
	}
}

func TestBuildSourceChain_EmptyDependencies(t *testing.T) {
	opts := SourceChainOptions{
		RitePath:        "/rites/myrite",
		RitesBase:       "/rites",
		Dependencies:    nil,
		PlatformMenaDir: "/platform/mena",
	}
	sources := BuildSourceChain(opts)

	// Expected: platform + shared + rite-local (no dep entries)
	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d: %+v", len(sources), sources)
	}
	if sources[0].Path != "/platform/mena" {
		t.Errorf("sources[0] should be platform")
	}
	if sources[1].Path != filepath.Join("/rites", "shared", "mena") {
		t.Errorf("sources[1] should be shared")
	}
	if sources[2].Path != filepath.Join("/rites/myrite", "mena") {
		t.Errorf("sources[2] should be rite-local")
	}
}
