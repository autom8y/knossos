package org

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/paths"
)

// validOrgName matches kebab-case identifiers: lowercase letters, digits, hyphens.
var validOrgName = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)

func newInitCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <org-name>",
		Short: "Bootstrap an organization directory",
		Long: `Create the org directory structure at the XDG data path.

Creates:
  $XDG_DATA_HOME/knossos/orgs/<org-name>/
    org.yaml        # Org metadata
    rites/           # Org-level rites
    agents/          # Org-level agents
    mena/            # Org-level mena (commands + skills)

Org names must be kebab-case (lowercase letters, digits, hyphens).

Examples:
  ari org init autom8y
  ari org init my-team`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := args[0]
			return runInit(ctx, orgName)
		},
	}
	return cmd
}

func runInit(ctx *cmdContext, orgName string) error {
	printer := ctx.getPrinter()

	// Validate org name
	if len(orgName) < 2 {
		return fmt.Errorf("org name must be at least 2 characters: %q", orgName)
	}
	if !validOrgName.MatchString(orgName) {
		return fmt.Errorf("invalid org name %q: must be kebab-case (lowercase letters, digits, hyphens)", orgName)
	}
	// Prevent path traversal
	if filepath.Base(orgName) != orgName {
		return fmt.Errorf("invalid org name %q: must not contain path separators", orgName)
	}

	orgDir := paths.OrgDataDir(orgName)

	// Check if already exists
	if _, err := os.Stat(orgDir); err == nil {
		return fmt.Errorf("org %q already exists at %s", orgName, orgDir)
	}

	// Create directory structure
	dirs := []string{
		orgDir,
		filepath.Join(orgDir, "rites"),
		filepath.Join(orgDir, "agents"),
		filepath.Join(orgDir, "mena"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write org.yaml
	orgYAML := fmt.Sprintf("name: %s\n", orgName)
	orgYAMLPath := filepath.Join(orgDir, "org.yaml")
	if err := os.WriteFile(orgYAMLPath, []byte(orgYAML), 0644); err != nil {
		return fmt.Errorf("failed to write org.yaml: %w", err)
	}

	printer.Print(map[string]interface{}{
		"status":  "created",
		"org":     orgName,
		"path":    orgDir,
		"message": fmt.Sprintf("Org %q initialized at %s", orgName, orgDir),
	})

	return nil
}
