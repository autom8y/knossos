package materialize

import (
	"github.com/autom8y/knossos/internal/provenance"
)

// CollisionChecker detects collisions between user and rite scope resources.
// Reads rite PROVENANCE_MANIFEST.yaml for O(1) lookup.
type CollisionChecker struct {
	riteEntries    map[string]bool
	manifestLoaded bool
}

// NewCollisionChecker creates a checker from rite manifest.
// claudeDir is the project .claude/. Empty string = no collision checking.
func NewCollisionChecker(claudeDir string) *CollisionChecker {
	c := &CollisionChecker{}
	if claudeDir != "" {
		c.loadRiteManifest(claudeDir)
	}
	return c
}

func (c *CollisionChecker) loadRiteManifest(claudeDir string) {
	c.manifestLoaded = true
	c.riteEntries = make(map[string]bool)
	manifestPath := provenance.ManifestPath(claudeDir)
	manifest, err := provenance.Load(manifestPath)
	if err != nil {
		return
	}
	for key, entry := range manifest.Entries {
		if entry.Scope == provenance.ScopeRite && entry.Owner == provenance.OwnerKnossos {
			c.riteEntries[key] = true
		}
	}
}

// CheckCollision checks if a manifest key collides with a rite entry.
func (c *CollisionChecker) CheckCollision(manifestKey string) (bool, string) {
	if !c.manifestLoaded || len(c.riteEntries) == 0 {
		return false, ""
	}
	if c.riteEntries[manifestKey] {
		return true, "(from manifest)"
	}
	return false, ""
}
