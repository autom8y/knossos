package usersync

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/config"
)

// CollisionChecker provides methods for detecting rite resource collisions.
type CollisionChecker struct {
	knossosHome  string
	ritesDir     string
	resourceType ResourceType
	nested       bool
}

// NewCollisionChecker creates a new collision checker for the given resource type.
func NewCollisionChecker(resourceType ResourceType, nested bool) *CollisionChecker {
	knossosHome := config.KnossosHome()
	return &CollisionChecker{
		knossosHome:  knossosHome,
		ritesDir:     filepath.Join(knossosHome, "rites"),
		resourceType: resourceType,
		nested:       nested,
	}
}

// CheckCollision checks if a resource name exists in any rite.
// Returns (hasCollision, riteName).
func (c *CollisionChecker) CheckCollision(name string) (bool, string) {
	if c.knossosHome == "" {
		return false, ""
	}

	if _, err := os.Stat(c.ritesDir); os.IsNotExist(err) {
		return false, ""
	}

	// Get the resource subdirectory name within rites (agents, skills, mena, hooks)
	subDir := c.resourceType.RiteSubDir()

	// For flat resources (agents), use just filename
	// For nested resources (skills, commands, hooks), use full relative path
	searchName := name
	if !c.nested {
		searchName = filepath.Base(name)
	}

	// Search each rite
	entries, err := os.ReadDir(c.ritesDir)
	if err != nil {
		return false, ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		riteName := entry.Name()
		resourcePath := filepath.Join(c.ritesDir, riteName, subDir, searchName)

		if _, err := os.Stat(resourcePath); err == nil {
			return true, riteName
		}
	}

	return false, ""
}

// GetRiteForResource finds which rite(s) contain a resource.
// Returns comma-separated list of rite names.
func GetRiteForResource(resourceType ResourceType, name string, nested bool) string {
	knossosHome := config.KnossosHome()
	if knossosHome == "" {
		return ""
	}

	ritesDir := filepath.Join(knossosHome, "rites")
	entries, err := os.ReadDir(ritesDir)
	if err != nil {
		return ""
	}

	subDir := resourceType.RiteSubDir()
	var matches []string

	// For flat resources, use just filename
	searchName := name
	if !nested {
		searchName = filepath.Base(name)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		riteName := entry.Name()
		resourcePath := filepath.Join(ritesDir, riteName, subDir, searchName)

		if _, err := os.Stat(resourcePath); err == nil {
			matches = append(matches, riteName)
		}
	}

	return strings.Join(matches, ", ")
}

// ListRiteResources lists all resources of a type from all rites.
// Returns a map of resource name to rite name(s).
func ListRiteResources(resourceType ResourceType, nested bool) map[string][]string {
	result := make(map[string][]string)

	knossosHome := config.KnossosHome()
	if knossosHome == "" {
		return result
	}

	ritesDir := filepath.Join(knossosHome, "rites")
	riteEntries, err := os.ReadDir(ritesDir)
	if err != nil {
		return result
	}

	subDir := resourceType.RiteSubDir()

	for _, riteEntry := range riteEntries {
		if !riteEntry.IsDir() {
			continue
		}

		riteName := riteEntry.Name()
		resourceDir := filepath.Join(ritesDir, riteName, subDir)

		if _, err := os.Stat(resourceDir); os.IsNotExist(err) {
			continue
		}

		// Walk the resource directory
		err := filepath.WalkDir(resourceDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}

			relPath, err := filepath.Rel(resourceDir, path)
			if err != nil {
				return nil
			}

			// For flat resources, use just the filename
			key := relPath
			if !nested {
				key = filepath.Base(relPath)
			}

			result[key] = append(result[key], riteName)
			return nil
		})
		if err != nil {
			continue
		}
	}

	return result
}
