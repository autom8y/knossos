package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
)

// TestRoutingDroToCommands verifies that INDEX.dro.md files
// are routed to .claude/commands/ with extension stripped
func TestRoutingDroToCommands(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create mena directory with a dromena command
	menaDir := filepath.Join(tmpDir, "mena", "test-cmd")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatalf("Failed to create mena dir: %v", err)
	}

	indexContent := `---
name: test-cmd
description: A test command
---
# Test Command

This is a test command.
`
	if err := os.WriteFile(filepath.Join(menaDir, "INDEX.dro.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.dro.md: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeMena(manifest, claudeDir, nil); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify: file should be in .claude/commands/ with stripped name, NOT in .claude/skills/
	commandsPath := filepath.Join(claudeDir, "commands", "test-cmd", "INDEX.md")
	skillsPath := filepath.Join(claudeDir, "skills", "test-cmd", "INDEX.md")

	if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
		t.Errorf("Expected dromena to be in .claude/commands/test-cmd/INDEX.md (stripped), but it does not exist")
	}

	if _, err := os.Stat(skillsPath); err == nil {
		t.Errorf("Dromena should NOT be in .claude/skills/test-cmd/, but it exists")
	}

	// Verify: old un-stripped name should NOT exist
	oldPath := filepath.Join(claudeDir, "commands", "test-cmd", "INDEX.dro.md")
	if _, err := os.Stat(oldPath); err == nil {
		t.Errorf("Extension stripping failed: INDEX.dro.md should not exist in output, but it does")
	}
}

// TestRoutingLegoToSkills verifies that INDEX.lego.md files
// are routed to .claude/skills/ with extension stripped
func TestRoutingLegoToSkills(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create mena directory with a legomena reference
	menaDir := filepath.Join(tmpDir, "mena", "test-ref")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatalf("Failed to create mena dir: %v", err)
	}

	indexContent := `---
name: test-ref
description: A test reference
---
# Test Reference

This is a test reference.
`
	if err := os.WriteFile(filepath.Join(menaDir, "INDEX.lego.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.lego.md: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeMena(manifest, claudeDir, nil); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify: file should be in .claude/skills/ with stripped name, NOT in .claude/commands/
	skillsPath := filepath.Join(claudeDir, "skills", "test-ref", "INDEX.md")
	commandsPath := filepath.Join(claudeDir, "commands", "test-ref", "INDEX.md")

	if _, err := os.Stat(skillsPath); os.IsNotExist(err) {
		t.Errorf("Expected legomena to be in .claude/skills/test-ref/INDEX.md (stripped), but it does not exist")
	}

	if _, err := os.Stat(commandsPath); err == nil {
		t.Errorf("Legomena should NOT be in .claude/commands/test-ref/, but it exists")
	}

	// Verify: old un-stripped name should NOT exist
	oldPath := filepath.Join(claudeDir, "skills", "test-ref", "INDEX.lego.md")
	if _, err := os.Stat(oldPath); err == nil {
		t.Errorf("Extension stripping failed: INDEX.lego.md should not exist in output, but it does")
	}
}

// TestRoutingDefaultIsDro verifies that commands with plain INDEX.md
// default to dromena routing (.claude/commands/) for backward compatibility
func TestRoutingDefaultIsDro(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create mena directory with a plain INDEX.md (no .dro/.lego extension)
	menaDir := filepath.Join(tmpDir, "mena", "test-default")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatalf("Failed to create mena dir: %v", err)
	}

	indexContent := `---
name: test-default
description: A test command with default routing
---
# Test Default Command

This command has a plain INDEX.md and should default to dromena routing.
`
	if err := os.WriteFile(filepath.Join(menaDir, "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.md: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeMena(manifest, claudeDir, nil); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify: plain INDEX.md defaults to .claude/commands/ (no stripping needed)
	commandsPath := filepath.Join(claudeDir, "commands", "test-default", "INDEX.md")
	skillsPath := filepath.Join(claudeDir, "skills", "test-default", "INDEX.md")

	if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
		t.Errorf("Expected default to be in .claude/commands/test-default/INDEX.md, but it does not exist")
	}

	if _, err := os.Stat(skillsPath); err == nil {
		t.Errorf("Default should NOT be in .claude/skills/test-default/, but it exists")
	}
}

// TestRoutingSupportingFilesFollowIndex verifies that supporting files
// are routed to the same destination as the INDEX file based on extension,
// and that INDEX extension is stripped in output
func TestRoutingSupportingFilesFollowIndex(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create mena directory with legomena INDEX and supporting files
	menaDir := filepath.Join(tmpDir, "mena", "test-with-files")
	if err := os.MkdirAll(menaDir, 0755); err != nil {
		t.Fatalf("Failed to create mena dir: %v", err)
	}

	indexContent := `---
name: test-with-files
description: A test reference with supporting files
---
# Test Reference

This is a test reference with supporting files.
`
	if err := os.WriteFile(filepath.Join(menaDir, "INDEX.lego.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.lego.md: %v", err)
	}

	// Create supporting files (plain .md, no extension convention)
	if err := os.WriteFile(filepath.Join(menaDir, "behavior.md"), []byte("# Behavior\n\nTest behavior."), 0644); err != nil {
		t.Fatalf("Failed to write behavior.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(menaDir, "examples.md"), []byte("# Examples\n\nTest examples."), 0644); err != nil {
		t.Fatalf("Failed to write examples.md: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeMena(manifest, claudeDir, nil); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify: ALL files should be in .claude/skills/ (following INDEX.lego.md)
	// INDEX.lego.md is stripped to INDEX.md in output
	skillsIndexPath := filepath.Join(claudeDir, "skills", "test-with-files", "INDEX.md")
	skillsBehaviorPath := filepath.Join(claudeDir, "skills", "test-with-files", "behavior.md")
	skillsExamplesPath := filepath.Join(claudeDir, "skills", "test-with-files", "examples.md")

	if _, err := os.Stat(skillsIndexPath); os.IsNotExist(err) {
		t.Errorf("Expected INDEX.md (stripped from INDEX.lego.md) to be in .claude/skills/test-with-files/, but it does not exist")
	}

	if _, err := os.Stat(skillsBehaviorPath); os.IsNotExist(err) {
		t.Errorf("Expected behavior.md to be in .claude/skills/test-with-files/, but it does not exist")
	}

	if _, err := os.Stat(skillsExamplesPath); os.IsNotExist(err) {
		t.Errorf("Expected examples.md to be in .claude/skills/test-with-files/, but it does not exist")
	}

	// Verify: files should NOT be in .claude/commands/
	commandsIndexPath := filepath.Join(claudeDir, "commands", "test-with-files", "INDEX.md")
	if _, err := os.Stat(commandsIndexPath); err == nil {
		t.Errorf("Files should NOT be in .claude/commands/test-with-files/, but INDEX.md exists")
	}

	// Verify: old un-stripped name should NOT exist
	oldPath := filepath.Join(claudeDir, "skills", "test-with-files", "INDEX.lego.md")
	if _, err := os.Stat(oldPath); err == nil {
		t.Errorf("Extension stripping failed: INDEX.lego.md should not exist in output, but it does")
	}
}

// TestRoutingMixedMena verifies that multiple mena with different
// extensions are routed to their correct destinations with extensions stripped
func TestRoutingMixedMena(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	menaBaseDir := filepath.Join(tmpDir, "mena")

	// Create dromena command
	droDir := filepath.Join(menaBaseDir, "dro-cmd")
	if err := os.MkdirAll(droDir, 0755); err != nil {
		t.Fatalf("Failed to create dro dir: %v", err)
	}
	droContent := `---
name: dro-cmd
description: A dromena command
---
# Dromena Command
`
	if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), []byte(droContent), 0644); err != nil {
		t.Fatalf("Failed to write dro INDEX: %v", err)
	}

	// Create legomena reference
	legoDir := filepath.Join(menaBaseDir, "lego-ref")
	if err := os.MkdirAll(legoDir, 0755); err != nil {
		t.Fatalf("Failed to create lego dir: %v", err)
	}
	legoContent := `---
name: lego-ref
description: A legomena reference
---
# Legomena Reference
`
	if err := os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), []byte(legoContent), 0644); err != nil {
		t.Fatalf("Failed to write lego INDEX: %v", err)
	}

	// Create default (plain INDEX.md) command
	defaultDir := filepath.Join(menaBaseDir, "default-cmd")
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		t.Fatalf("Failed to create default dir: %v", err)
	}
	defaultContent := `---
name: default-cmd
description: A default command
---
# Default Command
`
	if err := os.WriteFile(filepath.Join(defaultDir, "INDEX.md"), []byte(defaultContent), 0644); err != nil {
		t.Fatalf("Failed to write default INDEX: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeMena(manifest, claudeDir, nil); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify dromena is in .claude/commands/ (stripped to INDEX.md)
	droPath := filepath.Join(claudeDir, "commands", "dro-cmd", "INDEX.md")
	if _, err := os.Stat(droPath); os.IsNotExist(err) {
		t.Errorf("Expected dro-cmd to be in .claude/commands/ as INDEX.md (stripped), but it does not exist")
	}

	// Verify legomena is in .claude/skills/ (stripped to INDEX.md)
	legoPath := filepath.Join(claudeDir, "skills", "lego-ref", "INDEX.md")
	if _, err := os.Stat(legoPath); os.IsNotExist(err) {
		t.Errorf("Expected lego-ref to be in .claude/skills/ as INDEX.md (stripped), but it does not exist")
	}

	// Verify default is in .claude/commands/
	defaultPath := filepath.Join(claudeDir, "commands", "default-cmd", "INDEX.md")
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		t.Errorf("Expected default-cmd to be in .claude/commands/, but it does not exist")
	}

	// Verify no cross-contamination (check stripped names)
	droInSkills := filepath.Join(claudeDir, "skills", "dro-cmd", "INDEX.md")
	if _, err := os.Stat(droInSkills); err == nil {
		t.Errorf("dro-cmd should NOT be in .claude/skills/, but it exists")
	}

	legoInCommands := filepath.Join(claudeDir, "commands", "lego-ref", "INDEX.md")
	if _, err := os.Stat(legoInCommands); err == nil {
		t.Errorf("lego-ref should NOT be in .claude/commands/, but it exists")
	}
}

// TestRoutingNestedGrouping verifies that grouping directories (no INDEX file)
// recurse into leaf directories, routing each leaf by its own INDEX type.
// This tests the fix for the legomena routing defect where nested grouping dirs
// (e.g., guidance/standards with INDEX.lego.md) were incorrectly routed to commands/.
// Also verifies extension stripping in projected output.
func TestRoutingNestedGrouping(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	menaBaseDir := filepath.Join(tmpDir, "mena")

	// Create grouping dir "guidance" (no INDEX file -- just a container)
	// with two leaf dirs: one legomena, one dromena
	legoLeaf := filepath.Join(menaBaseDir, "guidance", "standards")
	droLeaf := filepath.Join(menaBaseDir, "guidance", "file-verification")
	if err := os.MkdirAll(legoLeaf, 0755); err != nil {
		t.Fatalf("Failed to create lego leaf: %v", err)
	}
	if err := os.MkdirAll(droLeaf, 0755); err != nil {
		t.Fatalf("Failed to create dro leaf: %v", err)
	}

	// guidance/standards/ -- legomena (INDEX.lego.md)
	legoIndex := `---
name: standards
description: Code conventions and tech stack
---
# Standards
`
	if err := os.WriteFile(filepath.Join(legoLeaf, "INDEX.lego.md"), []byte(legoIndex), 0644); err != nil {
		t.Fatalf("Failed to write lego INDEX: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legoLeaf, "code-conventions.md"), []byte("# Code Conventions\n"), 0644); err != nil {
		t.Fatalf("Failed to write supporting file: %v", err)
	}

	// guidance/file-verification/ -- dromena (INDEX.dro.md)
	droIndex := `---
name: file-verification
description: File verification command
---
# File Verification
`
	if err := os.WriteFile(filepath.Join(droLeaf, "INDEX.dro.md"), []byte(droIndex), 0644); err != nil {
		t.Fatalf("Failed to write dro INDEX: %v", err)
	}

	// Also create a flat top-level entry to ensure it still works
	flatDir := filepath.Join(menaBaseDir, "flat-cmd")
	if err := os.MkdirAll(flatDir, 0755); err != nil {
		t.Fatalf("Failed to create flat dir: %v", err)
	}
	flatIndex := `---
name: flat-cmd
description: A flat command
---
# Flat Command
`
	if err := os.WriteFile(filepath.Join(flatDir, "INDEX.dro.md"), []byte(flatIndex), 0644); err != nil {
		t.Fatalf("Failed to write flat INDEX: %v", err)
	}

	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeMena(manifest, claudeDir, nil); err != nil {
		t.Fatalf("materializeMena failed: %v", err)
	}

	// Verify: guidance/standards -> skills/guidance/standards/ (legomena, stripped)
	skillsStandards := filepath.Join(claudeDir, "skills", "guidance", "standards", "INDEX.md")
	if _, err := os.Stat(skillsStandards); os.IsNotExist(err) {
		t.Errorf("Expected guidance/standards to be in skills/ as INDEX.md (stripped), but %s does not exist", skillsStandards)
	}
	skillsConventions := filepath.Join(claudeDir, "skills", "guidance", "standards", "code-conventions.md")
	if _, err := os.Stat(skillsConventions); os.IsNotExist(err) {
		t.Errorf("Expected supporting file to follow INDEX to skills/, but %s does not exist", skillsConventions)
	}

	// Verify: guidance/standards NOT in commands/
	commandsStandards := filepath.Join(claudeDir, "commands", "guidance", "standards", "INDEX.md")
	if _, err := os.Stat(commandsStandards); err == nil {
		t.Errorf("Legomena guidance/standards should NOT be in commands/, but %s exists", commandsStandards)
	}

	// Verify: guidance/file-verification -> commands/guidance/file-verification/ (dromena, stripped)
	commandsFV := filepath.Join(claudeDir, "commands", "guidance", "file-verification", "INDEX.md")
	if _, err := os.Stat(commandsFV); os.IsNotExist(err) {
		t.Errorf("Expected guidance/file-verification to be in commands/ as INDEX.md (stripped), but %s does not exist", commandsFV)
	}

	// Verify: guidance/file-verification NOT in skills/
	skillsFV := filepath.Join(claudeDir, "skills", "guidance", "file-verification", "INDEX.md")
	if _, err := os.Stat(skillsFV); err == nil {
		t.Errorf("Dromena guidance/file-verification should NOT be in skills/, but %s exists", skillsFV)
	}

	// Verify: flat-cmd -> commands/ (still works, stripped)
	flatPath := filepath.Join(claudeDir, "commands", "flat-cmd", "INDEX.md")
	if _, err := os.Stat(flatPath); os.IsNotExist(err) {
		t.Errorf("Expected flat-cmd to still be in commands/ as INDEX.md (stripped), but %s does not exist", flatPath)
	}
}

// TestDetectMenaType verifies the extension-based type detection
func TestDetectMenaType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"INDEX.dro.md", "dro"},
		{"INDEX.lego.md", "lego"},
		{"INDEX.md", "dro"},           // default
		{"commit.dro.md", "dro"},      // standalone dromena
		{"standards.lego.md", "lego"}, // standalone legomena
		{"behavior.md", "dro"},        // plain .md defaults to dro
		{"README.md", "dro"},          // unrelated file defaults to dro
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectMenaType(tt.filename)
			if got != tt.expected {
				t.Errorf("DetectMenaType(%q) = %q, want %q", tt.filename, got, tt.expected)
			}
		})
	}
}
