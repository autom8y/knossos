package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autom8y/knossos/internal/paths"
)

// TestRoutingInvokableTrue verifies that commands with invokable: true
// are routed to .claude/commands/
func TestRoutingInvokableTrue(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source", "commands")
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create source command with invokable: true
	cmdDir := filepath.Join(sourceDir, "test-cmd")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("Failed to create command dir: %v", err)
	}

	indexContent := `---
name: test-cmd
description: A test command
invokable: true
---
# Test Command

This is a test command.
`
	if err := os.WriteFile(filepath.Join(cmdDir, "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.md: %v", err)
	}

	// Create user commands directory at project level
	// getUserCommandsDir() will find this automatically
	userCmdsDir := filepath.Join(tmpDir, "user-commands", "test-cmd")
	if err := os.MkdirAll(userCmdsDir, 0755); err != nil {
		t.Fatalf("Failed to create user commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(userCmdsDir, "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write user INDEX.md: %v", err)
	}

	// Create materializer
	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	// Create a minimal manifest
	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeCommands(manifest, claudeDir); err != nil {
		t.Fatalf("materializeCommands failed: %v", err)
	}

	// Verify: file should be in .claude/commands/, NOT in .claude/skills/
	commandsPath := filepath.Join(claudeDir, "commands", "test-cmd", "INDEX.md")
	skillsPath := filepath.Join(claudeDir, "skills", "test-cmd", "INDEX.md")

	if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
		t.Errorf("Expected command to be in .claude/commands/test-cmd/INDEX.md, but it does not exist")
	}

	if _, err := os.Stat(skillsPath); err == nil {
		t.Errorf("Command should NOT be in .claude/skills/test-cmd/INDEX.md, but it exists")
	}
}

// TestRoutingInvokableFalse verifies that commands with invokable: false
// are routed to .claude/skills/
func TestRoutingInvokableFalse(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create source command with invokable: false
	userCmdsDir := filepath.Join(tmpDir, "user-commands", "test-ref")
	if err := os.MkdirAll(userCmdsDir, 0755); err != nil {
		t.Fatalf("Failed to create command dir: %v", err)
	}

	indexContent := `---
name: test-ref
description: A test reference
invokable: false
category: reference
---
# Test Reference

This is a test reference.
`
	if err := os.WriteFile(filepath.Join(userCmdsDir, "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.md: %v", err)
	}

	// Create materializer
	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	// Create a minimal manifest
	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeCommands(manifest, claudeDir); err != nil {
		t.Fatalf("materializeCommands failed: %v", err)
	}

	// Verify: file should be in .claude/skills/, NOT in .claude/commands/
	commandsPath := filepath.Join(claudeDir, "commands", "test-ref", "INDEX.md")
	skillsPath := filepath.Join(claudeDir, "skills", "test-ref", "INDEX.md")

	if _, err := os.Stat(skillsPath); os.IsNotExist(err) {
		t.Errorf("Expected command to be in .claude/skills/test-ref/INDEX.md, but it does not exist")
	}

	if _, err := os.Stat(commandsPath); err == nil {
		t.Errorf("Command should NOT be in .claude/commands/test-ref/INDEX.md, but it exists")
	}
}

// TestRoutingDefaultIsInvokable verifies that commands without invokable field
// default to invokable=true and route to .claude/commands/
func TestRoutingDefaultIsInvokable(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create source command without invokable field
	userCmdsDir := filepath.Join(tmpDir, "user-commands", "test-default")
	if err := os.MkdirAll(userCmdsDir, 0755); err != nil {
		t.Fatalf("Failed to create command dir: %v", err)
	}

	indexContent := `---
name: test-default
description: A test command with default invokable
---
# Test Default Command

This command has no invokable field and should default to invokable=true.
`
	if err := os.WriteFile(filepath.Join(userCmdsDir, "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.md: %v", err)
	}

	// Create materializer
	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	// Create a minimal manifest
	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeCommands(manifest, claudeDir); err != nil {
		t.Fatalf("materializeCommands failed: %v", err)
	}

	// Verify: file should be in .claude/commands/ (default behavior)
	commandsPath := filepath.Join(claudeDir, "commands", "test-default", "INDEX.md")
	skillsPath := filepath.Join(claudeDir, "skills", "test-default", "INDEX.md")

	if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
		t.Errorf("Expected command to be in .claude/commands/test-default/INDEX.md, but it does not exist")
	}

	if _, err := os.Stat(skillsPath); err == nil {
		t.Errorf("Command should NOT be in .claude/skills/test-default/INDEX.md, but it exists")
	}
}

// TestRoutingSupportingFilesFollowIndex verifies that supporting files
// are routed to the same destination as INDEX.md based on invokable flag
func TestRoutingSupportingFilesFollowIndex(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	// Create source command with invokable: false and supporting files
	userCmdsDir := filepath.Join(tmpDir, "user-commands", "test-with-files")
	if err := os.MkdirAll(userCmdsDir, 0755); err != nil {
		t.Fatalf("Failed to create command dir: %v", err)
	}

	indexContent := `---
name: test-with-files
description: A test reference with supporting files
invokable: false
category: reference
---
# Test Reference

This is a test reference with supporting files.
`
	if err := os.WriteFile(filepath.Join(userCmdsDir, "INDEX.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write INDEX.md: %v", err)
	}

	// Create supporting files
	if err := os.WriteFile(filepath.Join(userCmdsDir, "behavior.md"), []byte("# Behavior\n\nTest behavior."), 0644); err != nil {
		t.Fatalf("Failed to write behavior.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(userCmdsDir, "examples.md"), []byte("# Examples\n\nTest examples."), 0644); err != nil {
		t.Fatalf("Failed to write examples.md: %v", err)
	}

	// Create materializer
	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	// Create a minimal manifest
	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeCommands(manifest, claudeDir); err != nil {
		t.Fatalf("materializeCommands failed: %v", err)
	}

	// Verify: ALL files should be in .claude/skills/ (following INDEX.md)
	skillsIndexPath := filepath.Join(claudeDir, "skills", "test-with-files", "INDEX.md")
	skillsBehaviorPath := filepath.Join(claudeDir, "skills", "test-with-files", "behavior.md")
	skillsExamplesPath := filepath.Join(claudeDir, "skills", "test-with-files", "examples.md")

	if _, err := os.Stat(skillsIndexPath); os.IsNotExist(err) {
		t.Errorf("Expected INDEX.md to be in .claude/skills/test-with-files/, but it does not exist")
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
}

// TestRoutingMixedCommands verifies that multiple commands with different
// invokable settings are routed to their correct destinations
func TestRoutingMixedCommands(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	userCmdsDir := filepath.Join(tmpDir, "user-commands")

	// Create invokable command
	invokableDir := filepath.Join(userCmdsDir, "invokable-cmd")
	if err := os.MkdirAll(invokableDir, 0755); err != nil {
		t.Fatalf("Failed to create invokable dir: %v", err)
	}
	invokableContent := `---
name: invokable-cmd
description: An invokable command
invokable: true
---
# Invokable Command
`
	if err := os.WriteFile(filepath.Join(invokableDir, "INDEX.md"), []byte(invokableContent), 0644); err != nil {
		t.Fatalf("Failed to write invokable INDEX.md: %v", err)
	}

	// Create non-invokable reference
	refDir := filepath.Join(userCmdsDir, "reference-cmd")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		t.Fatalf("Failed to create reference dir: %v", err)
	}
	refContent := `---
name: reference-cmd
description: A reference command
invokable: false
category: reference
---
# Reference Command
`
	if err := os.WriteFile(filepath.Join(refDir, "INDEX.md"), []byte(refContent), 0644); err != nil {
		t.Fatalf("Failed to write reference INDEX.md: %v", err)
	}

	// Create default (no invokable field) command
	defaultDir := filepath.Join(userCmdsDir, "default-cmd")
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
		t.Fatalf("Failed to write default INDEX.md: %v", err)
	}

	// Create materializer
	resolver := paths.NewResolver(tmpDir)
	m := NewMaterializer(resolver)

	// Create a minimal manifest
	manifest := &RiteManifest{
		Name:    "test",
		Version: "1.0.0",
	}

	if err := m.materializeCommands(manifest, claudeDir); err != nil {
		t.Fatalf("materializeCommands failed: %v", err)
	}

	// Verify invokable command is in .claude/commands/
	invokablePath := filepath.Join(claudeDir, "commands", "invokable-cmd", "INDEX.md")
	if _, err := os.Stat(invokablePath); os.IsNotExist(err) {
		t.Errorf("Expected invokable-cmd to be in .claude/commands/, but it does not exist")
	}

	// Verify reference is in .claude/skills/
	refPath := filepath.Join(claudeDir, "skills", "reference-cmd", "INDEX.md")
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		t.Errorf("Expected reference-cmd to be in .claude/skills/, but it does not exist")
	}

	// Verify default command is in .claude/commands/
	defaultPath := filepath.Join(claudeDir, "commands", "default-cmd", "INDEX.md")
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		t.Errorf("Expected default-cmd to be in .claude/commands/, but it does not exist")
	}

	// Verify cross-contamination doesn't occur
	invokableInSkills := filepath.Join(claudeDir, "skills", "invokable-cmd", "INDEX.md")
	if _, err := os.Stat(invokableInSkills); err == nil {
		t.Errorf("invokable-cmd should NOT be in .claude/skills/, but it exists")
	}

	refInCommands := filepath.Join(claudeDir, "commands", "reference-cmd", "INDEX.md")
	if _, err := os.Stat(refInCommands); err == nil {
		t.Errorf("reference-cmd should NOT be in .claude/commands/, but it exists")
	}
}
