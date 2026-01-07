package artifact

import (
	"os"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
)

// Aggregator handles session-to-project registry synchronization.
type Aggregator struct {
	registry *Registry
}

// NewAggregator creates a new Aggregator.
func NewAggregator(registry *Registry) *Aggregator {
	return &Aggregator{registry: registry}
}

// AggregateSession merges a session's artifacts into the project registry.
// This is the primary integration point called on mark_complete.
func (a *Aggregator) AggregateSession(sessionID string) error {
	// Load session registry
	sessionReg, err := a.registry.LoadSessionRegistry(sessionID)
	if err != nil {
		return err
	}

	if len(sessionReg.Artifacts) == 0 {
		return nil // Nothing to aggregate
	}

	// Load project registry
	projectReg, err := a.registry.LoadProjectRegistry()
	if err != nil {
		return err
	}

	// Build map of existing artifacts for deduplication
	existingMap := make(map[string]int)
	for i, entry := range projectReg.Artifacts {
		existingMap[entry.ArtifactID] = i
	}

	// Track sessions
	sessionsSet := make(map[string]bool)
	for _, entry := range projectReg.Artifacts {
		sessionsSet[entry.SessionID] = true
	}

	// Merge session artifacts
	for _, entry := range sessionReg.Artifacts {
		if idx, exists := existingMap[entry.ArtifactID]; exists {
			// Update existing entry
			projectReg.Artifacts[idx] = entry
		} else {
			// Add new entry
			projectReg.Artifacts = append(projectReg.Artifacts, entry)
		}
		sessionsSet[entry.SessionID] = true
	}

	// Rebuild indexes
	projectReg.Indexes = a.buildIndexes(projectReg.Artifacts)
	projectReg.SessionsIndexed = len(sessionsSet)

	return a.registry.SaveProjectRegistry(projectReg)
}

// AggregateAll rebuilds the project registry from all session registries.
// Use for recovery or initial index build.
func (a *Aggregator) AggregateAll() error {
	sessionsDir := a.registry.paths.SessionsDir()

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No sessions yet
		}
		return errors.Wrap(errors.CodeGeneralError, "failed to read sessions directory", err)
	}

	// Start with fresh project registry
	projectReg := &ProjectRegistry{
		SchemaVersion:   "1.0",
		ProjectRoot:     a.registry.projectRoot,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		Artifacts:       []Entry{},
		Indexes: ProjectIndexes{
			ByPhase:      make(map[Phase][]string),
			ByType:       make(map[ArtifactType][]string),
			BySpecialist: make(map[string][]string),
			BySession:    make(map[string][]string),
		},
	}

	sessionsSet := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "session-") {
			continue
		}

		sessionID := entry.Name()
		sessionReg, err := a.registry.LoadSessionRegistry(sessionID)
		if err != nil {
			// Log warning but continue
			continue
		}

		for _, artifact := range sessionReg.Artifacts {
			projectReg.Artifacts = append(projectReg.Artifacts, artifact)
			sessionsSet[sessionID] = true
		}
	}

	projectReg.Indexes = a.buildIndexes(projectReg.Artifacts)
	projectReg.SessionsIndexed = len(sessionsSet)

	return a.registry.SaveProjectRegistry(projectReg)
}

// buildIndexes creates the query indexes from artifact list.
func (a *Aggregator) buildIndexes(artifacts []Entry) ProjectIndexes {
	indexes := ProjectIndexes{
		ByPhase:      make(map[Phase][]string),
		ByType:       make(map[ArtifactType][]string),
		BySpecialist: make(map[string][]string),
		BySession:    make(map[string][]string),
	}

	for _, entry := range artifacts {
		id := entry.ArtifactID

		indexes.ByPhase[entry.Phase] = append(indexes.ByPhase[entry.Phase], id)
		indexes.ByType[entry.ArtifactType] = append(indexes.ByType[entry.ArtifactType], id)
		indexes.BySpecialist[entry.Specialist] = append(indexes.BySpecialist[entry.Specialist], id)
		indexes.BySession[entry.SessionID] = append(indexes.BySession[entry.SessionID], id)
	}

	return indexes
}
