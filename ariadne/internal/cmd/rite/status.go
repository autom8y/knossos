package rite

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/output"
	"github.com/autom8y/ariadne/internal/paths"
	ritelib "github.com/autom8y/ariadne/internal/rite"
)

type statusOptions struct {
	riteName string
}

func newStatusCmd(ctx *cmdContext) *cobra.Command {
	var opts statusOptions

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show rite status",
		Long:  `Shows detailed status of the current or specified rite (practice bundle).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite to query status for (default: active)")
	cmd.Flags().StringVarP(&opts.riteName, "team", "t", "", "Deprecated: use --rite instead")

	cmd.Flags().MarkDeprecated("team", "use --rite instead")

	return cmd
}

func runStatus(ctx *cmdContext, opts statusOptions) error {
	printer := ctx.getPrinter()
	discovery := ctx.getDiscovery()
	resolver := ctx.getResolver()

	// Get rite name (from flag or active)
	riteName := opts.riteName
	if riteName == "" {
		riteName = discovery.ActiveRiteName()
		if riteName == "" {
			err := errors.New(errors.CodeFileNotFound, "No active team set")
			printer.PrintError(err)
			return err
		}
	}

	// Get rite info
	t, err := discovery.Get(riteName)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	// Load workflow for phase info
	workflowPath := filepath.Join(t.Path, "workflow.yaml")
	workflow, err := ritelib.LoadWorkflow(workflowPath)
	if err != nil {
		wrappedErr := errors.Wrap(errors.CodeGeneralError, "failed to load workflow", err)
		printer.PrintError(wrappedErr)
		return wrappedErr
	}

	// Build agent status list
	agents := buildAgentStatusList(t, workflow, resolver)

	// Check manifest validity
	manifest, manifestErr := ritelib.LoadAgentManifest(resolver.AgentManifestFile())
	manifestValid := manifestErr == nil && manifest.ActiveRite == riteName

	// Check CLAUDE.md sync
	updater := ritelib.NewClaudeMDUpdater(resolver.ClaudeMDFile())
	claudeMDSynced := updater.IsSynced(riteName)

	// Get orphans if any
	var orphans []string
	if manifestErr == nil {
		orphans = manifest.DetectOrphans(riteName)
	}

	result := output.RiteStatusOutput{
		Rite:           t.Name,
		IsActive:       t.Active,
		Path:           t.Path,
		Description:    t.Description,
		WorkflowType:   t.WorkflowType,
		Agents:         agents,
		Phases:         workflow.PhaseNames(),
		EntryPoint:     t.EntryPoint,
		Orphans:        orphans,
		ManifestValid:  manifestValid,
		ClaudeMDSynced: claudeMDSynced,
	}

	return printer.Print(result)
}

func buildAgentStatusList(t *ritelib.Rite, workflow *ritelib.Workflow, resolver *paths.Resolver) []output.AgentStatus {
	infos := workflow.GetAgentInfo()
	agents := make([]output.AgentStatus, 0, len(infos))

	// Get list of installed agents
	installedAgents := make(map[string]bool)
	agentsDir := resolver.AgentsDir()
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				installedAgents[entry.Name()] = true
			}
		}
	}

	for _, info := range infos {
		agent := output.AgentStatus{
			Name:      info.Name,
			File:      info.File,
			Role:      info.Role,
			Produces:  info.Produces,
			Installed: installedAgents[info.File],
		}
		agents = append(agents, agent)
	}

	return agents
}
