package usersync

import (
	"os"
	"path/filepath"

	"github.com/autom8y/knossos/internal/provenance"
)

// cleanupOldManifests removes all legacy JSON manifest files and creates backups.
// Called after successful sync for all resource types.
func (s *Syncer) cleanupOldManifests() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	claudeDir := filepath.Join(homeDir, ".claude")
	oldManifests := []string{
		filepath.Join(claudeDir, "USER_AGENT_MANIFEST.json"),
		filepath.Join(claudeDir, "USER_MENA_MANIFEST.json"),
		filepath.Join(claudeDir, "USER_HOOKS_MANIFEST.json"),
		filepath.Join(claudeDir, "USER_COMMAND_MANIFEST.json"),
		filepath.Join(claudeDir, "USER_SKILL_MANIFEST.json"),
	}
	for _, path := range oldManifests {
		// Backup before removal for safety
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Already gone or unreadable
		}
		backupPath := path + ".v2-backup"
		os.WriteFile(backupPath, data, 0644) // Best effort
		os.Remove(path)
	}
}

// loadManifest loads the unified provenance manifest.
func (s *Syncer) loadManifest() (*provenance.ProvenanceManifest, error) {
	return provenance.LoadOrBootstrap(s.manifestPath)
}

// saveManifest saves the unified provenance manifest.
func (s *Syncer) saveManifest(manifest *provenance.ProvenanceManifest) error {
	return provenance.Save(s.manifestPath, manifest)
}
