package rite

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/inscription"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/paths"
	ritelib "github.com/autom8y/knossos/internal/rite"
	"gopkg.in/yaml.v3"
)

type statusOptions struct {
	riteName string
}

func newStatusCmd(ctx *cmdContext) *cobra.Command {
	var opts statusOptions

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show rite status",
		Long: `Shows detailed status of the current or specified rite (practice bundle).

Includes agent installation status, workflow phases, manifest validity,
inscription sync status, and any orphaned agents.

Examples:
  ari rite status
  ari rite status -r ecosystem
  ari rite status -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite to query status for (default: active)")

	return cmd
}

func runStatus(ctx *cmdContext, opts statusOptions) error {
	printer := ctx.getPrinter()
	discovery := ctx.getDiscovery()
	resolver := ctx.GetResolver()

	// Get rite name (from flag or active)
	riteName := opts.riteName
	if riteName == "" {
		riteName = discovery.ActiveRiteName()
		if riteName == "" {
			return errors.New(errors.CodeFileNotFound, "No active rite set")
		}
	}

	// Get rite info
	t, err := discovery.Get(riteName)
	if err != nil {
		return err
	}

	// Load workflow for phase info (non-fatal: rite status should work even if workflow is missing)
	workflowPath := filepath.Join(t.Path, "workflow.yaml")
	workflow, err := ritelib.LoadWorkflow(workflowPath)
	if err != nil {
		workflow = nil
	}

	// Load rite manifest for canonical agent list
	manifest, manifestErr := discovery.GetManifest(riteName)

	// Build agent status list from manifest (canonical) or workflow (fallback)
	var agents []output.AgentStatus
	if manifestErr == nil && len(manifest.Agents) > 0 {
		agents = buildAgentStatusFromManifest(manifest, resolver)
	} else if workflow != nil {
		agents = buildAgentStatusList(t, workflow, resolver)
	}

	// Check manifest validity (KNOSSOS_MANIFEST.yaml)
	manifestValid := false
	knossosManifestPath := resolver.KnossosManifestFile()
	if data, err := os.ReadFile(knossosManifestPath); err == nil {
		var manifest inscription.Manifest
		if err := yaml.Unmarshal(data, &manifest); err == nil {
			manifestValid = manifest.ActiveRite == riteName
		}
	}

	// Check inscription sync - simple check if rite name appears in inscription file
	inscriptionSynced := false
	if content, err := os.ReadFile(resolver.ContextFileForChannel(paths.ClaudeChannel{})); err == nil {
		inscriptionSynced = strings.Contains(string(content), riteName)
	}

	// Orphans are now handled by materialization, not status command
	var orphans []string

	result := output.RiteStatusOutput{
		Rite:           t.Name,
		IsActive:       t.Active,
		Path:           t.Path,
		Description:    t.Description,
		WorkflowType:   t.WorkflowType,
		Agents:         agents,
		Phases:         workflowPhaseNames(workflow),
		EntryPoint:     t.EntryPoint,
		Orphans:        orphans,
		ManifestValid:  manifestValid,
		InscriptionSynced: inscriptionSynced,
	}

	return printer.Print(result)
}

func buildAgentStatusFromManifest(manifest *ritelib.RiteManifest, resolver *paths.Resolver) []output.AgentStatus {
	agents := make([]output.AgentStatus, 0, len(manifest.Agents))

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

	for _, ref := range manifest.Agents {
		file := ref.File
		if file == "" {
			file = ref.Name + ".md"
		}
		agents = append(agents, output.AgentStatus{
			Name:      ref.Name,
			File:      file,
			Role:      ref.Role,
			Produces:  ref.Produces,
			Installed: installedAgents[file],
		})
	}

	return agents
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

// workflowPhaseNames returns phase names from a workflow, or nil if workflow is nil.
func workflowPhaseNames(w *ritelib.Workflow) []string {
	if w == nil {
		return nil
	}
	return w.PhaseNames()
}
