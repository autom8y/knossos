// Package procession implements the ari procession commands for cross-rite workflow management.
package procession

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/errors"
	procmena "github.com/autom8y/knossos/internal/materialize/procession"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/session"
)

// createOptions holds options for the create command.
type createOptions struct {
	templateName string
}

// newCreateCmd creates the procession create subcommand.
func newCreateCmd(ctx *cmdContext) *cobra.Command {
	var opts createOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Start a new procession from a template",
		Long: `Start a new cross-rite coordinated workflow from a named template.

Resolves the template through the 5-tier resolution chain (project > user >
org > platform > embedded), creates the artifact directory, and stores the
procession state in the active session context. The procession ID is generated
as {template-name}-{YYYY-MM-DD}.

Examples:
  ari procession create --template=security-remediation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.templateName, "template", "", "Procession template name (required)")
	_ = cmd.MarkFlagRequired("template")

	return cmd
}

// createOutput represents the output of procession create.
type createOutput struct {
	ProcessionID   string `json:"procession_id"`
	Type           string `json:"type"`
	CurrentStation string `json:"current_station"`
	CurrentRite    string `json:"current_rite"`
	NextStation    string `json:"next_station,omitempty"`
	NextRite       string `json:"next_rite,omitempty"`
	ArtifactDir    string `json:"artifact_dir"`
}

// Text implements output.Textable for createOutput.
func (o createOutput) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Procession created: %s\n", o.ProcessionID)
	fmt.Fprintf(&b, "Type:               %s\n", o.Type)
	fmt.Fprintf(&b, "Current station:    %s (%s)\n", o.CurrentStation, o.CurrentRite)
	if o.NextStation != "" {
		fmt.Fprintf(&b, "Next station:       %s (%s)\n", o.NextStation, o.NextRite)
	} else {
		b.WriteString("Next station:       (final)\n")
	}
	fmt.Fprintf(&b, "Artifact dir:       %s\n", o.ArtifactDir)
	return b.String()
}

var _ output.Textable = createOutput{}

// runCreate executes the procession create command.
func runCreate(ctx *cmdContext, opts createOptions) error {
	printer := ctx.getPrinter()
	resolver := ctx.GetResolver()

	// Get session ID
	sessionID, err := ctx.GetSessionID()
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to get session ID", err))
	}
	if sessionID == "" {
		return common.PrintAndReturn(printer, errors.New(errors.CodeSessionNotFound, "No active session. Use 'ari session create' first."))
	}

	// Load session context
	ctxPath := resolver.SessionContextFile(sessionID)
	sessCtx, err := session.LoadContext(ctxPath)
	if err != nil {
		if errors.IsNotFound(err) {
			return common.PrintAndReturn(printer, errors.ErrSessionNotFound(sessionID))
		}
		return common.PrintAndReturn(printer, err)
	}

	// Reject if a procession is already active
	if sessCtx.Procession != nil {
		return common.PrintAndReturn(printer, errors.New(errors.CodeLifecycleViolation,
			fmt.Sprintf("session already has an active procession %q; use 'ari procession abandon' to terminate it first",
				sessCtx.Procession.ID)))
	}

	// Resolve template through the 5-tier resolution chain
	projectDir := resolver.ProjectRoot()
	rp, err := procmena.ResolveTemplate(opts.templateName, projectDir, common.EmbeddedProcessions())
	if err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeFileNotFound, "template resolution failed", err))
	}
	tmpl := rp.Template

	// Generate procession ID: {template-name}-{YYYY-MM-DD}
	today := time.Now().UTC().Format("2006-01-02")
	processionID := fmt.Sprintf("%s-%s", tmpl.Name, today)

	// Create artifact directory relative to project root
	artifactDir := filepath.Join(projectDir, tmpl.ArtifactDir)
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError,
			fmt.Sprintf("failed to create artifact directory: %s", artifactDir), err))
	}

	// First station
	firstStation := tmpl.Stations[0]

	// Compute next station and rite from template
	nextStationName := tmpl.NextStation(firstStation.Name)
	nextRite := ""
	if nextStationName != "" {
		if ns := tmpl.GetStation(nextStationName); ns != nil {
			nextRite = ns.Rite
		}
	}

	// Set procession on session context
	sessCtx.Procession = &session.Procession{
		ID:             processionID,
		Type:           tmpl.Name,
		CurrentStation: firstStation.Name,
		NextStation:    nextStationName,
		NextRite:       nextRite,
		ArtifactDir:    tmpl.ArtifactDir,
	}

	// Save session context
	if err := sessCtx.Save(ctxPath); err != nil {
		return common.PrintAndReturn(printer, errors.Wrap(errors.CodeGeneralError, "failed to save session context", err))
	}

	return printer.Print(createOutput{
		ProcessionID:   processionID,
		Type:           tmpl.Name,
		CurrentStation: firstStation.Name,
		CurrentRite:    firstStation.Rite,
		NextStation:    nextStationName,
		NextRite:       nextRite,
		ArtifactDir:    tmpl.ArtifactDir,
	})
}
