// Package artifact implements the Federated Artifact Registry for Ariadne.
// It provides queryable indexes of workflow artifacts at session and project levels.
package artifact

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// ArtifactType represents the type of workflow artifact.
type ArtifactType string

const (
	TypePRD      ArtifactType = "prd"
	TypeTDD      ArtifactType = "tdd"
	TypeADR      ArtifactType = "adr"
	TypeTestPlan ArtifactType = "test-plan"
	TypeCode     ArtifactType = "code"
	TypeRunbook  ArtifactType = "runbook"
)

// Phase represents the workflow phase that produced the artifact.
type Phase string

const (
	PhaseRequirements   Phase = "requirements"
	PhaseDesign         Phase = "design"
	PhaseImplementation Phase = "implementation"
	PhaseValidation     Phase = "validation"
)

// Entry represents a single artifact in the registry.
type Entry struct {
	ArtifactID       string            `yaml:"artifact_id" json:"artifact_id"`
	ArtifactType     ArtifactType      `yaml:"artifact_type" json:"artifact_type"`
	Path             string            `yaml:"path" json:"path"`
	Phase            Phase             `yaml:"phase" json:"phase"`
	Specialist       string            `yaml:"specialist" json:"specialist"`
	SessionID        string            `yaml:"session_id" json:"session_id"`
	TaskID           string            `yaml:"task_id,omitempty" json:"task_id,omitempty"`
	RegisteredAt     time.Time         `yaml:"registered_at" json:"registered_at"`
	Validated        bool              `yaml:"validated" json:"validated"`
	ValidationIssues []string          `yaml:"validation_issues,omitempty" json:"validation_issues,omitempty"`
	Metadata         map[string]any    `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// SessionRegistry represents artifacts registered within a single session.
type SessionRegistry struct {
	SchemaVersion string    `yaml:"schema_version" json:"schema_version"`
	SessionID     string    `yaml:"session_id" json:"session_id"`
	CreatedAt     time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt     time.Time `yaml:"updated_at" json:"updated_at"`
	ArtifactCount int       `yaml:"artifact_count" json:"artifact_count"`
	Artifacts     []Entry   `yaml:"artifacts" json:"artifacts"`
}

// ProjectRegistry represents the aggregated project-level artifact index.
type ProjectRegistry struct {
	SchemaVersion    string            `yaml:"schema_version" json:"schema_version"`
	ProjectRoot      string            `yaml:"project_root" json:"project_root"`
	CreatedAt        time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt        time.Time         `yaml:"updated_at" json:"updated_at"`
	ArtifactCount    int               `yaml:"artifact_count" json:"artifact_count"`
	SessionsIndexed  int               `yaml:"sessions_indexed" json:"sessions_indexed"`
	Artifacts        []Entry           `yaml:"artifacts" json:"artifacts"`
	Indexes          ProjectIndexes    `yaml:"indexes" json:"indexes"`
}

// ProjectIndexes contains pre-computed indexes for fast querying.
type ProjectIndexes struct {
	ByPhase      map[Phase][]string          `yaml:"by_phase" json:"by_phase"`
	ByType       map[ArtifactType][]string   `yaml:"by_type" json:"by_type"`
	BySpecialist map[string][]string         `yaml:"by_specialist" json:"by_specialist"`
	BySession    map[string][]string         `yaml:"by_session" json:"by_session"`
}

// Registry provides CRUD operations for artifact registries.
type Registry struct {
	projectRoot string
	paths       *paths.Resolver
}

// NewRegistry creates a new Registry for the given project root.
func NewRegistry(projectRoot string) *Registry {
	return &Registry{
		projectRoot: projectRoot,
		paths:       paths.NewResolver(projectRoot),
	}
}

// SessionRegistryPath returns the path to a session's artifacts.yaml.
func (r *Registry) SessionRegistryPath(sessionID string) string {
	return filepath.Join(r.paths.SessionDir(sessionID), "artifacts.yaml")
}

// ProjectRegistryPath returns the path to the project registry.
func (r *Registry) ProjectRegistryPath() string {
	return filepath.Join(r.paths.ClaudeDir(), "artifacts", "registry.yaml")
}

// LoadSessionRegistry loads the artifact registry for a session.
// Returns empty registry (not error) if file doesn't exist.
func (r *Registry) LoadSessionRegistry(sessionID string) (*SessionRegistry, error) {
	path := r.SessionRegistryPath(sessionID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty registry
			return &SessionRegistry{
				SchemaVersion: "1.0",
				SessionID:     sessionID,
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
				ArtifactCount: 0,
				Artifacts:     []Entry{},
			}, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read session registry", err)
	}

	var registry SessionRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse session registry", err)
	}

	return &registry, nil
}

// SaveSessionRegistry writes the session registry to disk.
func (r *Registry) SaveSessionRegistry(registry *SessionRegistry) error {
	path := r.SessionRegistryPath(registry.SessionID)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to create registry directory", err)
	}

	// Update metadata
	registry.UpdatedAt = time.Now().UTC()
	registry.ArtifactCount = len(registry.Artifacts)

	data, err := yaml.Marshal(registry)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal session registry", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Register adds an artifact to the session registry.
func (r *Registry) Register(sessionID string, entry Entry) error {
	registry, err := r.LoadSessionRegistry(sessionID)
	if err != nil {
		return err
	}

	// Check for duplicate
	for i, existing := range registry.Artifacts {
		if existing.ArtifactID == entry.ArtifactID {
			// Update existing entry
			registry.Artifacts[i] = entry
			return r.SaveSessionRegistry(registry)
		}
	}

	// Add new entry
	registry.Artifacts = append(registry.Artifacts, entry)
	return r.SaveSessionRegistry(registry)
}

// LoadProjectRegistry loads the project-level artifact index.
func (r *Registry) LoadProjectRegistry() (*ProjectRegistry, error) {
	path := r.ProjectRegistryPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectRegistry{
				SchemaVersion:   "1.0",
				ProjectRoot:     r.projectRoot,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
				ArtifactCount:   0,
				SessionsIndexed: 0,
				Artifacts:       []Entry{},
				Indexes: ProjectIndexes{
					ByPhase:      make(map[Phase][]string),
					ByType:       make(map[ArtifactType][]string),
					BySpecialist: make(map[string][]string),
					BySession:    make(map[string][]string),
				},
			}, nil
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to read project registry", err)
	}

	var registry ProjectRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse project registry", err)
	}

	return &registry, nil
}

// SaveProjectRegistry writes the project registry to disk.
func (r *Registry) SaveProjectRegistry(registry *ProjectRegistry) error {
	path := r.ProjectRegistryPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to create registry directory", err)
	}

	registry.UpdatedAt = time.Now().UTC()
	registry.ArtifactCount = len(registry.Artifacts)

	data, err := yaml.Marshal(registry)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal project registry", err)
	}

	return os.WriteFile(path, data, 0644)
}
