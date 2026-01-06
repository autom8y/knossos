package team

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/inscription"
	"github.com/autom8y/ariadne/internal/paths"
)

// SwitchOptions configures rite switch behavior.
type SwitchOptions struct {
	TargetRite string
	RemoveAll  bool
	KeepAll    bool
	PromoteAll bool
	Update     bool
	DryRun     bool
	NoSync     bool // Skip inscription sync on rite switch
}

// HasOrphanStrategy returns true if an orphan handling flag is set.
func (o *SwitchOptions) HasOrphanStrategy() bool {
	return o.RemoveAll || o.KeepAll || o.PromoteAll
}

// OrphanStrategy returns the strategy name.
func (o *SwitchOptions) OrphanStrategy() string {
	if o.RemoveAll {
		return "remove-all"
	}
	if o.KeepAll {
		return "keep-all"
	}
	if o.PromoteAll {
		return "promote-all"
	}
	return ""
}

// SwitchResult represents the result of a rite switch.
type SwitchResult struct {
	Rite               string             `json:"team"` // Keep JSON tag for backward compatibility
	PreviousRite       string             `json:"previous_team"` // Keep JSON tag for backward compatibility
	SwitchedAt         time.Time          `json:"switched_at"`
	AgentsInstalled    []string           `json:"agents_installed"`
	OrphansHandled     *OrphanResult      `json:"orphans_handled,omitempty"`
	ClaudeMDUpdated    bool               `json:"claude_md_updated"`
	ManifestPath       string             `json:"manifest_path"`
	InscriptionSynced  bool               `json:"inscription_synced,omitempty"`
	InscriptionVersion string             `json:"inscription_version,omitempty"`
	SyncConflicts      []InscriptionConflict `json:"sync_conflicts,omitempty"`
}

// InscriptionConflict represents a conflict from inscription sync.
type InscriptionConflict struct {
	Region    string `json:"region"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Preserved bool   `json:"preserved"`
}

// OrphanResult tracks orphan handling.
type OrphanResult struct {
	Strategy string   `json:"strategy"`
	Agents   []string `json:"agents"`
}

// DryRunResult represents a dry-run result.
type DryRunResult struct {
	DryRun                 bool     `json:"dry_run"`
	WouldSwitchTo          string   `json:"would_switch_to"`
	CurrentTeam            string   `json:"current_team"`
	WouldInstall           []string `json:"would_install"`
	OrphansDetected        []string `json:"orphans_detected"`
	OrphanStrategyRequired bool     `json:"orphan_strategy_required"`
	SuggestedFlags         []string `json:"suggested_flags,omitempty"`
}

// Switcher handles team switching operations.
type Switcher struct {
	resolver  *paths.Resolver
	discovery *Discovery
}

// NewSwitcher creates a new team switcher.
func NewSwitcher(resolver *paths.Resolver) *Switcher {
	return &Switcher{
		resolver:  resolver,
		discovery: NewDiscovery(resolver),
	}
}

// Switch performs a rite switch.
func (s *Switcher) Switch(opts SwitchOptions) (*SwitchResult, error) {
	// 1. Validate target rite exists
	rite, err := s.discovery.Get(opts.TargetRite)
	if err != nil {
		return nil, err
	}

	// 2. Load current manifest
	manifest, err := LoadManifest(s.resolver.AgentManifestFile())
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to load manifest", err)
	}

	previousRite := manifest.ActiveRite

	// 3. Detect orphans
	orphans := manifest.DetectOrphans(opts.TargetRite)
	if len(orphans) > 0 && !opts.HasOrphanStrategy() {
		return nil, errors.ErrOrphanConflict(orphans, previousRite, opts.TargetRite)
	}

	// 4. Dry run check
	if opts.DryRun {
		return s.dryRunResult(rite, manifest, orphans, opts)
	}

	// 5. Check if already on target rite and not forcing update
	if previousRite == opts.TargetRite && !opts.Update {
		// Return success with no changes
		return &SwitchResult{
			Rite:            opts.TargetRite,
			PreviousRite:    previousRite,
			SwitchedAt:      time.Now().UTC(),
			AgentsInstalled: []string{},
			ClaudeMDUpdated: false,
			ManifestPath:    s.resolver.AgentManifestFile(),
		}, nil
	}

	// 6. Create backup before making changes
	backup, err := s.createBackup()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to create backup", err)
	}

	// 7. Execute switch with rollback on failure
	result, err := s.executeSwitch(rite, manifest, orphans, opts)
	if err != nil {
		s.restoreBackup(backup)
		return nil, errors.Wrap(errors.CodeSwitchAborted, "switch failed, restored backup", err)
	}

	// 8. Clean up backup on success
	s.cleanupBackup(backup)

	return result, nil
}

// dryRunResult returns what would happen without making changes.
func (s *Switcher) dryRunResult(rite *Rite, manifest *Manifest, orphans []string, opts SwitchOptions) (*SwitchResult, error) {
	// This is a bit of a hack - return DryRunResult through a wrapper
	// The actual implementation will populate the proper structure
	agents := make([]string, len(rite.Agents))
	for i, a := range rite.Agents {
		agents[i] = a + ".md"
	}

	result := &SwitchResult{
		Rite:            opts.TargetRite,
		PreviousRite:    manifest.ActiveRite,
		SwitchedAt:      time.Now().UTC(),
		AgentsInstalled: agents,
		ClaudeMDUpdated: false,
		ManifestPath:    s.resolver.AgentManifestFile(),
	}

	if len(orphans) > 0 {
		result.OrphansHandled = &OrphanResult{
			Strategy: opts.OrphanStrategy(),
			Agents:   orphans,
		}
	}

	return result, nil
}

// executeSwitch performs the actual switch operation.
func (s *Switcher) executeSwitch(rite *Rite, manifest *Manifest, orphans []string, opts SwitchOptions) (*SwitchResult, error) {
	// 1. Handle orphans
	if len(orphans) > 0 {
		if err := s.handleOrphans(manifest, orphans, opts); err != nil {
			return nil, err
		}
	}

	// 2. Copy agents from rite to .claude/agents/
	agentsDir := s.resolver.AgentsDir()
	if err := paths.EnsureDir(agentsDir); err != nil {
		return nil, err
	}

	var installedAgents []string
	riteAgentsDir := s.resolver.TeamAgentsDir(opts.TargetRite)

	entries, err := os.ReadDir(riteAgentsDir)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read rite agents", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !filepath.HasPrefix(entry.Name(), "") {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		srcPath := filepath.Join(riteAgentsDir, entry.Name())
		dstPath := filepath.Join(agentsDir, entry.Name())

		if err := copyFile(srcPath, dstPath); err != nil {
			return nil, errors.Wrap(errors.CodeGeneralError, "failed to copy agent", err)
		}

		// Compute checksum and add to manifest
		checksum, _ := ComputeChecksum(dstPath)
		manifest.AddAgent(entry.Name(), "team", opts.TargetRite, checksum)
		installedAgents = append(installedAgents, entry.Name())
	}

	// 3. Update ACTIVE_RITE file
	if err := os.WriteFile(s.resolver.ActiveRiteFile(), []byte(opts.TargetRite), 0644); err != nil {
		return nil, errors.Wrap(errors.CodePermissionDenied, "failed to write ACTIVE_RITE", err)
	}

	// 4. Copy workflow.yaml to ACTIVE_WORKFLOW.yaml
	workflowSrc := s.resolver.TeamWorkflowFile(opts.TargetRite)
	workflowDst := s.resolver.ActiveWorkflowFile()
	if err := copyFile(workflowSrc, workflowDst); err != nil {
		// Non-critical, just log
	}

	// 5. Update manifest
	manifest.SetActiveTeam(opts.TargetRite)
	manifest.ClearOrphans()
	if err := manifest.Save(s.resolver.AgentManifestFile()); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to save manifest", err)
	}

	// 6. Build result structure
	result := &SwitchResult{
		Rite:            opts.TargetRite,
		PreviousRite:    manifest.ActiveRite,
		SwitchedAt:      time.Now().UTC(),
		AgentsInstalled: installedAgents,
		ClaudeMDUpdated: false,
		ManifestPath:    s.resolver.AgentManifestFile(),
	}

	if len(orphans) > 0 {
		result.OrphansHandled = &OrphanResult{
			Strategy: opts.OrphanStrategy(),
			Agents:   orphans,
		}
	}

	// 7. Trigger inscription sync (unless --no-sync)
	if !opts.NoSync {
		pipeline := inscription.NewPipeline(s.resolver.ProjectRoot())
		syncResult, err := pipeline.Sync(inscription.SyncOptions{
			RiteName: opts.TargetRite,
		})
		if err != nil {
			// Log warning but don't fail the switch
			// The switch itself succeeded, inscription sync is non-critical
			result.InscriptionSynced = false
		} else {
			result.InscriptionSynced = true
			result.InscriptionVersion = syncResult.InscriptionVersion
			result.ClaudeMDUpdated = true

			// Convert conflicts to output format
			if len(syncResult.Conflicts) > 0 {
				result.SyncConflicts = make([]InscriptionConflict, len(syncResult.Conflicts))
				for i, c := range syncResult.Conflicts {
					result.SyncConflicts[i] = InscriptionConflict{
						Region:    c.Region,
						Type:      string(c.Type),
						Message:   c.Message,
						Preserved: c.Preserved,
					}
				}
			}
		}
	} else {
		// Fallback to legacy ClaudeMDUpdater if inscription sync is skipped
		updater := NewClaudeMDUpdater(s.resolver.ClaudeMDFile())
		if err := updater.UpdateForTeam(rite); err == nil {
			result.ClaudeMDUpdated = true
		}
	}

	return result, nil
}

// handleOrphans processes orphaned agents according to the strategy.
func (s *Switcher) handleOrphans(manifest *Manifest, orphans []string, opts SwitchOptions) error {
	agentsDir := s.resolver.AgentsDir()

	for _, orphan := range orphans {
		if opts.RemoveAll {
			// Delete the agent file
			agentPath := filepath.Join(agentsDir, orphan)
			if err := os.Remove(agentPath); err != nil && !os.IsNotExist(err) {
				return err
			}
			manifest.RemoveAgent(orphan)
		} else if opts.KeepAll {
			// Mark as orphaned but keep
			manifest.MarkOrphaned(orphan)
		} else if opts.PromoteAll {
			// Promote to project-level
			manifest.PromoteToProject(orphan)
		}
	}

	return nil
}

// Backup structure for rollback
type backup struct {
	activeTeamPath   string
	activeTeamData   []byte
	manifestPath     string
	manifestData     []byte
	agentsDir        string
	agentBackups     map[string][]byte
}

// createBackup saves current state for potential rollback.
func (s *Switcher) createBackup() (*backup, error) {
	b := &backup{
		activeTeamPath: s.resolver.ActiveRiteFile(),
		manifestPath:   s.resolver.AgentManifestFile(),
		agentsDir:      s.resolver.AgentsDir(),
		agentBackups:   make(map[string][]byte),
	}

	// Backup ACTIVE_RITE
	if data, err := os.ReadFile(b.activeTeamPath); err == nil {
		b.activeTeamData = data
	}

	// Backup manifest
	if data, err := os.ReadFile(b.manifestPath); err == nil {
		b.manifestData = data
	}

	// Backup agent files
	if entries, err := os.ReadDir(b.agentsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			path := filepath.Join(b.agentsDir, entry.Name())
			if data, err := os.ReadFile(path); err == nil {
				b.agentBackups[entry.Name()] = data
			}
		}
	}

	return b, nil
}

// restoreBackup restores state from backup.
func (s *Switcher) restoreBackup(b *backup) {
	if b == nil {
		return
	}

	// Restore ACTIVE_RITE
	if b.activeTeamData != nil {
		os.WriteFile(b.activeTeamPath, b.activeTeamData, 0644)
	}

	// Restore manifest
	if b.manifestData != nil {
		os.WriteFile(b.manifestPath, b.manifestData, 0644)
	}

	// Restore agent files
	for name, data := range b.agentBackups {
		path := filepath.Join(b.agentsDir, name)
		os.WriteFile(path, data, 0644)
	}
}

// cleanupBackup removes backup data.
func (s *Switcher) cleanupBackup(b *backup) {
	// No files to clean up - we stored data in memory
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
