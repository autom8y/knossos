package land

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
)

func newSynthesizeCmd(ctx *cmdContext) *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "synthesize",
		Short: "Synthesize session knowledge into persistent form",
		Long: `Synthesize cross-session knowledge from .sos/archive/ into .sos/land/.

This is an infrastructure stub. The actual synthesis logic (Dionysus agent
delegation) will be implemented in a future initiative.

Examples:
  ari land synthesize
  ari land synthesize --domain=architecture`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSynthesize(ctx, domain)
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Knowledge domain to synthesize (e.g., architecture, conventions)")

	return cmd
}

type synthesizeOutput struct {
	Status    string `json:"status"`
	LandDir   string `json:"land_dir"`
	ArchiveOK bool   `json:"archive_exists"`
	Domain    string `json:"domain,omitempty"`
	Message   string `json:"message"`
}

// Text implements output.Textable.
func (o synthesizeOutput) Text() string {
	return o.Message
}

func runSynthesize(ctx *cmdContext, domain string) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	// Ensure .sos/land/ exists
	landDir := resolver.LandDir()
	if err := paths.EnsureDir(landDir); err != nil {
		return fmt.Errorf("cannot create land directory: %w", err)
	}

	// Check if archive has session data
	archiveDir := resolver.ArchiveDir()
	archiveExists := false
	if entries, err := os.ReadDir(archiveDir); err == nil && len(entries) > 0 {
		archiveExists = true
	}

	msg := fmt.Sprintf(`Land directory: %s
Archive data: %s

Synthesis is not yet implemented. Future plans:
  - Dionysus agent will transform session archives into persistent knowledge
  - Domain-scoped synthesis (architecture, conventions, scar-tissue)
  - Output to .sos/land/{domain}.md for git tracking
`, filepath.Join(".sos", "land"), archiveStatus(archiveExists))

	if domain != "" {
		msg += fmt.Sprintf("\nRequested domain: %s\n", domain)
	}

	return printer.Print(synthesizeOutput{
		Status:    "stub",
		LandDir:   landDir,
		ArchiveOK: archiveExists,
		Domain:    domain,
		Message:   msg,
	})
}

func archiveStatus(exists bool) string {
	if exists {
		return "available"
	}
	return "empty (no archived sessions yet)"
}
