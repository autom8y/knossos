package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/autom8y/knossos/internal/paths"
)

// GraduatedEntry represents a single artifact that was graduated to .ledge/.
type GraduatedEntry struct {
	ArtifactID    string       `json:"artifact_id"`
	OriginalPath  string       `json:"original_path"`
	GraduatedPath string       `json:"graduated_path"`
	ArtifactType  ArtifactType `json:"artifact_type"`
	Category      string       `json:"category"`
}

// GraduationResult contains the outcome of graduating session artifacts.
type GraduationResult struct {
	Graduated []GraduatedEntry `json:"graduated"`
	Warnings  []string         `json:"warnings,omitempty"`
}

// GraduateSession copies registered artifacts from session-local paths to
// permanent .ledge/{category}/ locations with YAML provenance frontmatter.
// Returns an empty result (not error) if the session has no artifact registry.
func GraduateSession(resolver *paths.Resolver, sessionID string) (*GraduationResult, error) {
	registry := NewRegistry(resolver.ProjectRoot())

	sessionReg, err := registry.LoadSessionRegistry(sessionID)
	if err != nil {
		return nil, err
	}

	result := &GraduationResult{}

	if len(sessionReg.Artifacts) == 0 {
		return result, nil
	}

	now := time.Now().UTC()

	for _, entry := range sessionReg.Artifacts {
		category := LedgeCategoryForType(entry.ArtifactType)
		if category == "" {
			continue
		}

		graduatedPath := registry.GraduatedPath(entry)
		srcPath := filepath.Join(resolver.ProjectRoot(), entry.Path)
		destPath := filepath.Join(resolver.ProjectRoot(), graduatedPath)

		content, readErr := os.ReadFile(srcPath)
		if readErr != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("cannot read source %s: %v", entry.Path, readErr))
			continue
		}

		if mkdirErr := paths.EnsureDir(filepath.Dir(destPath)); mkdirErr != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("cannot create directory for %s: %v", graduatedPath, mkdirErr))
			continue
		}

		provenance := fmt.Sprintf("---\nsession_id: %s\ngraduated_at: %s\noriginal_path: %s\n---\n\n",
			sessionID,
			now.Format(time.RFC3339),
			entry.Path,
		)
		graduated := append([]byte(provenance), content...)

		if writeErr := os.WriteFile(destPath, graduated, 0644); writeErr != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("cannot write graduated artifact %s: %v", graduatedPath, writeErr))
			continue
		}

		result.Graduated = append(result.Graduated, GraduatedEntry{
			ArtifactID:    entry.ArtifactID,
			OriginalPath:  entry.Path,
			GraduatedPath: graduatedPath,
			ArtifactType:  entry.ArtifactType,
			Category:      category,
		})
	}

	return result, nil
}
