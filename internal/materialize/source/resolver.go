package source

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// SourceResolver resolves rite sources with configurable fallback chain.
type SourceResolver struct {
	projectRoot     string
	projectRitesDir string
	userRitesDir    string
	knossosHome     string
	EmbeddedFS      fs.FS // Embedded rites filesystem (compiled-in fallback)

	// Cached resolutions (lazy-loaded)
	mu       sync.RWMutex
	resolved map[string]*ResolvedRite
}

// NewSourceResolver creates a new source resolver for the given project root.
func NewSourceResolver(projectRoot string) *SourceResolver {
	return &SourceResolver{
		projectRoot:     projectRoot,
		projectRitesDir: filepath.Join(projectRoot, ".knossos", "rites"),
		userRitesDir:    paths.UserRitesDir(),
		knossosHome:     config.KnossosHome(),
		resolved:        make(map[string]*ResolvedRite),
	}
}

// KnossosHome returns the resolved knossos home directory.
func (r *SourceResolver) KnossosHome() string {
	return r.knossosHome
}

// WithEmbeddedFS sets the embedded filesystem for fallback rite resolution.
// Returns the receiver for method chaining.
func (r *SourceResolver) WithEmbeddedFS(fsys fs.FS) *SourceResolver {
	r.EmbeddedFS = fsys
	return r
}

// ResolveRite finds the source for a rite with the configured fallback chain.
//
// Resolution order (highest to lowest priority):
//  1. ExplicitSource (if --source flag provided)
//  2. Project satellite rites (.knossos/rites/{rite}/)
//  3. User rites (~/.local/share/knossos/rites/{rite}/)
//  4. Knossos platform ($KNOSSOS_HOME/rites/{rite}/)
//  5. Embedded rites (compiled into binary)
//
// Returns error if rite is not found in any source.
func (r *SourceResolver) ResolveRite(riteName string, explicitSource string) (*ResolvedRite, error) {
	// Check cache first (unless explicit source overrides)
	if explicitSource == "" {
		r.mu.RLock()
		if cached, ok := r.resolved[riteName]; ok {
			r.mu.RUnlock()
			return cached, nil
		}
		r.mu.RUnlock()
	}

	var result *ResolvedRite
	var checkedPaths []string

	// 1. Explicit source (--source flag)
	if explicitSource != "" {
		source, err := r.parseExplicitSource(explicitSource)
		if err != nil {
			return nil, err
		}
		result, err = r.checkSource(riteName, source)
		if err != nil {
			return nil, errors.NewWithDetails(errors.CodeRiteNotFound,
				fmt.Sprintf("rite '%s' not found in explicit source: %s", riteName, explicitSource),
				map[string]any{
					"rite":   riteName,
					"source": explicitSource,
				})
		}
		return result, nil // Don't cache explicit source resolutions
	}

	// 2. Project rites
	if r.projectRitesDir != "" {
		source := RiteSource{
			Type:        SourceProject,
			Path:        r.projectRitesDir,
			Description: "project rites directory",
		}
		if res, err := r.checkSource(riteName, source); err == nil {
			result = res
		} else {
			checkedPaths = append(checkedPaths, source.Path)
		}
	}

	// 3. User rites
	if result == nil && r.userRitesDir != "" {
		source := RiteSource{
			Type:        SourceUser,
			Path:        r.userRitesDir,
			Description: "user-level rites",
		}
		if res, err := r.checkSource(riteName, source); err == nil {
			result = res
		} else {
			checkedPaths = append(checkedPaths, source.Path)
		}
	}

	// 4. Knossos platform
	if result == nil && r.knossosHome != "" {
		source := RiteSource{
			Type:        SourceKnossos,
			Path:        filepath.Join(r.knossosHome, "rites"),
			Description: fmt.Sprintf("Knossos platform at %s", r.knossosHome),
		}
		if res, err := r.checkSource(riteName, source); err == nil {
			result = res
		} else {
			checkedPaths = append(checkedPaths, source.Path)
		}
	}

	// 5. Embedded rites (compiled-in fallback)
	if result == nil && r.EmbeddedFS != nil {
		if res, err := r.checkEmbeddedSource(riteName); err == nil {
			result = res
		} else {
			checkedPaths = append(checkedPaths, "embedded://rites/"+riteName)
		}
	}

	if result == nil {
		return nil, errors.NewWithDetails(errors.CodeRiteNotFound,
			fmt.Sprintf("rite '%s' not found in any source", riteName),
			map[string]any{
				"rite":          riteName,
				"checked_paths": checkedPaths,
				"hint":          "Use --source to specify explicit path, or ensure KNOSSOS_HOME is set",
			})
	}

	// Cache result
	r.mu.Lock()
	r.resolved[riteName] = result
	r.mu.Unlock()

	return result, nil
}

// parseExplicitSource parses the --source flag value.
//
// Supports:
//   - "knossos" or "knossos:" -> KNOSSOS_HOME
//   - "/absolute/path" -> explicit path
//   - "~/relative/path" -> expanded relative path
func (r *SourceResolver) parseExplicitSource(source string) (RiteSource, error) {
	// Handle "knossos" alias
	if source == "knossos" || source == "knossos:" {
		home := config.KnossosHome()
		if home == "" {
			return RiteSource{}, errors.New(errors.CodeGeneralError,
				"KNOSSOS_HOME not set. Set KNOSSOS_HOME environment variable or use explicit path.")
		}
		return RiteSource{
			Type:        SourceKnossos,
			Path:        filepath.Join(home, "rites"),
			Description: fmt.Sprintf("Knossos platform at %s", home),
		}, nil
	}

	// Expand ~ to home directory
	expandedPath := source
	if len(source) > 0 && source[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return RiteSource{}, errors.Wrap(errors.CodeGeneralError, "failed to expand ~", err)
		}
		expandedPath = filepath.Join(homeDir, source[1:])
	}

	// If path doesn't contain /rites, append it
	if filepath.Base(expandedPath) != "rites" {
		ritesPath := filepath.Join(expandedPath, "rites")
		if info, err := os.Stat(ritesPath); err == nil && info.IsDir() {
			expandedPath = ritesPath
		}
	}

	// Validate path exists
	if _, err := os.Stat(expandedPath); err != nil {
		return RiteSource{}, errors.NewWithDetails(errors.CodeFileNotFound,
			fmt.Sprintf("source path does not exist: %s", expandedPath),
			map[string]any{"source": source, "expanded": expandedPath})
	}

	return RiteSource{
		Type:        SourceExplicit,
		Path:        expandedPath,
		Description: fmt.Sprintf("explicit source: %s", source),
	}, nil
}

// checkSource checks if a rite exists in a source location.
func (r *SourceResolver) checkSource(riteName string, source RiteSource) (*ResolvedRite, error) {
	ritePath := filepath.Join(source.Path, riteName)
	manifestPath := filepath.Join(ritePath, "manifest.yaml")

	// Check manifest exists
	if _, err := os.Stat(manifestPath); err != nil {
		return nil, err
	}

	// Determine templates directory based on source type
	var templatesDir string
	switch source.Type {
	case SourceKnossos:
		// Knossos templates are in knossos/templates/ subdirectory
		templatesDir = filepath.Join(filepath.Dir(source.Path), "knossos", "templates")
	case SourceProject:
		templatesDir = filepath.Join(r.projectRoot, "templates")
		// For knossos platform self-hosting: templates live at knossos/templates/
		// rather than the standard templates/ location. Detect via sections/ marker.
		if _, err := os.Stat(filepath.Join(templatesDir, "sections")); err != nil {
			alt := filepath.Join(r.projectRoot, "knossos", "templates")
			if _, err := os.Stat(filepath.Join(alt, "sections")); err == nil {
				templatesDir = alt
			}
		}
	case SourceExplicit:
		// Check if templates dir exists alongside rites dir
		templatesDir = filepath.Join(filepath.Dir(source.Path), "templates")
		if _, err := os.Stat(templatesDir); err != nil {
			templatesDir = "" // No templates available
		}
	case SourceUser:
		templatesDir = "" // User rites don't have templates
	}

	return &ResolvedRite{
		Name:         riteName,
		Source:       source,
		RitePath:     ritePath,
		ManifestPath: manifestPath,
		TemplatesDir: templatesDir,
	}, nil
}

// ListAvailableRites returns all rites available from all configured sources.
// Higher-priority sources shadow lower-priority ones (project > user > knossos).
func (r *SourceResolver) ListAvailableRites() ([]ResolvedRite, error) {
	var result []ResolvedRite
	seen := make(map[string]bool)

	// Collect from all sources in priority order
	sources := []RiteSource{
		{Type: SourceProject, Path: r.projectRitesDir},
		{Type: SourceUser, Path: r.userRitesDir},
		{Type: SourceKnossos, Path: filepath.Join(r.knossosHome, "rites")},
	}

	for _, source := range sources {
		if source.Path == "" {
			continue
		}

		entries, err := os.ReadDir(source.Path)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			riteName := entry.Name()
			if seen[riteName] {
				continue // Already found in higher-priority source
			}

			if resolved, err := r.checkSource(riteName, source); err == nil {
				result = append(result, *resolved)
				seen[riteName] = true
			}
		}
	}

	// 5. Embedded rites (lowest priority)
	if r.EmbeddedFS != nil {
		entries, err := fs.ReadDir(r.EmbeddedFS, "rites")
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() || seen[entry.Name()] {
					continue
				}
				if resolved, err := r.checkEmbeddedSource(entry.Name()); err == nil {
					result = append(result, *resolved)
					seen[entry.Name()] = true
				}
			}
		}
	}

	return result, nil
}

// checkEmbeddedSource checks if a rite exists in the embedded filesystem.
func (r *SourceResolver) checkEmbeddedSource(riteName string) (*ResolvedRite, error) {
	manifestPath := "rites/" + riteName + "/manifest.yaml"
	if _, err := fs.Stat(r.EmbeddedFS, manifestPath); err != nil {
		return nil, err
	}

	return &ResolvedRite{
		Name: riteName,
		Source: RiteSource{
			Type:        SourceEmbedded,
			Path:        "embedded://rites/" + riteName,
			Description: "compiled-in rite definition",
		},
		RitePath:     "rites/" + riteName,
		ManifestPath: manifestPath,
		TemplatesDir: "knossos/templates",
	}, nil
}

// ClearCache clears the resolution cache.
func (r *SourceResolver) ClearCache() {
	r.mu.Lock()
	r.resolved = make(map[string]*ResolvedRite)
	r.mu.Unlock()
}
