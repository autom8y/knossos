package source

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/resolution"
)

// SourceResolver resolves rite sources with configurable fallback chain.
type SourceResolver struct {
	projectRoot     string
	projectRitesDir string
	userRitesDir    string
	orgRitesDir     string // Org-level rites directory (between user and knossos)
	knossosHome     string
	activeOrg       string // Active org name (from config.ActiveOrg() at construction)
	dataDir         string // XDG data dir for org path computation (defaults to paths.DataDir())
	EmbeddedFS      fs.FS  // Embedded rites filesystem (compiled-in fallback)

	// Cached resolutions (lazy-loaded)
	mu       sync.RWMutex
	resolved map[string]*ResolvedRite
}

// NewSourceResolver creates a new source resolver for the given project root.
// Tier paths are read from global config. For explicit paths, use NewSourceResolverWithPaths.
func NewSourceResolver(projectRoot string) *SourceResolver {
	return &SourceResolver{
		projectRoot:     projectRoot,
		projectRitesDir: filepath.Join(projectRoot, ".knossos", "rites"),
		userRitesDir:    paths.UserRitesDir(),
		orgRitesDir:     paths.OrgRitesDir(config.ActiveOrg()),
		knossosHome:     config.KnossosHome(),
		activeOrg:       config.ActiveOrg(),
		dataDir:         paths.DataDir(),
		resolved:        make(map[string]*ResolvedRite),
	}
}

// NewSourceResolverWithPaths creates a source resolver with explicit tier paths.
// Empty strings are accepted and cause the corresponding tier to be skipped during
// resolution. This enables test injection without global state mutation.
func NewSourceResolverWithPaths(projectRoot, userRitesDir, orgRitesDir, knossosHome string) *SourceResolver {
	return &SourceResolver{
		projectRoot:     projectRoot,
		projectRitesDir: filepath.Join(projectRoot, ".knossos", "rites"),
		userRitesDir:    userRitesDir,
		orgRitesDir:     orgRitesDir,
		knossosHome:     knossosHome,
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
//  4. Org rites ($XDG_DATA_HOME/knossos/orgs/{org}/rites/{rite}/)
//  5. Knossos platform ($KNOSSOS_HOME/rites/{rite}/)
//  6. Embedded rites (compiled into binary)
//
// Returns error if rite is not found in any source.
func (r *SourceResolver) ResolveRite(riteName string, explicitSource string) (*ResolvedRite, error) {
	// Check cache first (unless explicit source overrides).
	// Cache key includes orgRitesDir so org switches invalidate stale entries.
	cacheKey := riteName + "\x00" + r.orgRitesDir
	if explicitSource == "" {
		r.mu.RLock()
		if cached, ok := r.resolved[cacheKey]; ok {
			r.mu.RUnlock()
			return cached, nil
		}
		r.mu.RUnlock()
	}

	// 1. Explicit source (--source flag) — SourceResolver-owned, not in Chain
	if explicitSource != "" {
		source, err := r.parseExplicitSource(explicitSource)
		if err != nil {
			return nil, err
		}
		result, err := r.checkSource(riteName, source)
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

	// 2-6. Resolution chain: project > user > org > platform > embedded
	chain := r.riteChain()
	item, err := chain.Resolve(riteName, manifestValidator)
	if err != nil {
		// Build checked paths from chain tiers for diagnostic detail
		checkedPaths := make([]string, 0, len(chain.Tiers()))
		for _, tier := range chain.Tiers() {
			if tier.FS != nil {
				checkedPaths = append(checkedPaths, "embedded://"+tier.Dir+"/"+riteName)
			} else {
				checkedPaths = append(checkedPaths, filepath.Join(tier.Dir, riteName))
			}
		}
		return nil, errors.NewWithDetails(errors.CodeRiteNotFound,
			fmt.Sprintf("rite '%s' not found in any source", riteName),
			map[string]any{
				"rite":          riteName,
				"checked_paths": checkedPaths,
				"hint":          "Use --source to specify explicit path, or ensure KNOSSOS_HOME is set",
			})
	}

	result := r.toResolvedRite(riteName, item)

	// Cache result
	r.mu.Lock()
	r.resolved[cacheKey] = result
	r.mu.Unlock()

	return result, nil
}

// parseExplicitSource parses the --source flag value.
//
// Supports:
//   - "knossos" or "knossos:" -> KNOSSOS_HOME
//   - "org" -> active org's rites directory
//   - "org:{name}" -> named org's rites directory
//   - "/absolute/path" -> explicit path
//   - "~/relative/path" -> expanded relative path
func (r *SourceResolver) parseExplicitSource(source string) (RiteSource, error) {
	// Handle "knossos" alias
	if source == "knossos" || source == "knossos:" {
		home := r.knossosHome
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

	// Handle "org" and "org:{name}" aliases
	if source == "org" || strings.HasPrefix(source, "org:") {
		orgName := r.activeOrg
		if strings.HasPrefix(source, "org:") {
			orgName = source[4:]
		}
		if orgName == "" {
			return RiteSource{}, errors.New(errors.CodeGeneralError,
				"No active org. Set KNOSSOS_ORG environment variable, run ari org set, or specify org:name.")
		}
		orgDir := filepath.Join(r.dataDir, "orgs", orgName, "rites")
		return RiteSource{
			Type:        SourceOrg,
			Path:        orgDir,
			Description: fmt.Sprintf("org '%s' rites", orgName),
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
// Used for explicit source resolution (--source flag) which is outside the Chain.
func (r *SourceResolver) checkSource(riteName string, source RiteSource) (*ResolvedRite, error) {
	ritePath := filepath.Join(source.Path, riteName)
	manifestPath := filepath.Join(ritePath, "manifest.yaml")

	// Check manifest exists
	if _, err := os.Stat(manifestPath); err != nil {
		return nil, err
	}

	return &ResolvedRite{
		Name:         riteName,
		Source:       source,
		RitePath:     ritePath,
		ManifestPath: manifestPath,
		TemplatesDir: r.resolveTemplatesDir(source),
	}, nil
}

// resolveTemplatesDir determines the templates directory for a given source.
func (r *SourceResolver) resolveTemplatesDir(source RiteSource) string {
	switch source.Type {
	case SourceKnossos:
		// Knossos templates are in knossos/templates/ subdirectory
		return filepath.Join(filepath.Dir(source.Path), "knossos", "templates")
	case SourceProject:
		templatesDir := filepath.Join(r.projectRoot, "templates")
		// For knossos platform self-hosting: templates live at knossos/templates/
		// rather than the standard templates/ location. Detect via sections/ marker.
		if _, err := os.Stat(filepath.Join(templatesDir, "sections")); err != nil {
			alt := filepath.Join(r.projectRoot, "knossos", "templates")
			if _, err := os.Stat(filepath.Join(alt, "sections")); err == nil {
				return alt
			}
		}
		return templatesDir
	case SourceExplicit:
		// Check if templates dir exists alongside rites dir
		templatesDir := filepath.Join(filepath.Dir(source.Path), "templates")
		if _, err := os.Stat(templatesDir); err != nil {
			return "" // No templates available
		}
		return templatesDir
	case SourceEmbedded:
		return "knossos/templates"
	case SourceUser, SourceOrg:
		return "" // User/org rites don't carry templates
	default:
		return ""
	}
}

// ListAvailableRites returns all rites available from all configured sources.
// Higher-priority sources shadow lower-priority ones (project > user > org > platform > embedded).
// Delegates to resolution.RiteChain for multi-tier enumeration with shadowing.
func (r *SourceResolver) ListAvailableRites() ([]ResolvedRite, error) {
	items, err := r.riteChain().ResolveAll(manifestValidator)
	if err != nil {
		return nil, err
	}

	result := make([]ResolvedRite, 0, len(items))
	for _, item := range items {
		resolved := r.toResolvedRite(item.Name, &item)
		result = append(result, *resolved)
	}

	return result, nil
}

// riteChain builds the resolution chain from the current resolver paths.
// Called on demand (cheap: just allocates a small struct with a tier slice).
func (r *SourceResolver) riteChain() *resolution.Chain {
	platformRitesDir := ""
	if r.knossosHome != "" {
		platformRitesDir = filepath.Join(r.knossosHome, "rites")
	}
	return resolution.RiteChain(
		r.projectRitesDir,
		r.userRitesDir,
		r.orgRitesDir,
		platformRitesDir,
		r.EmbeddedFS,
	)
}

// manifestValidator checks that a resolution item is a valid rite (has manifest.yaml).
func manifestValidator(item resolution.ResolvedItem) bool {
	if item.Fsys != nil {
		_, err := fs.Stat(item.Fsys, filepath.Join(item.Path, "manifest.yaml"))
		return err == nil
	}
	_, err := os.Stat(filepath.Join(item.Path, "manifest.yaml"))
	return err == nil
}

// toResolvedRite converts a resolution.ResolvedItem into a ResolvedRite,
// enriching it with source type mapping and templates directory resolution.
func (r *SourceResolver) toResolvedRite(riteName string, item *resolution.ResolvedItem) *ResolvedRite {
	// Handle embedded FS tier
	if item.Fsys != nil {
		return &ResolvedRite{
			Name: riteName,
			Source: RiteSource{
				Type:        SourceEmbedded,
				Path:        "embedded://rites/" + riteName,
				Description: "compiled-in rite definition",
			},
			RitePath:     item.Path,
			ManifestPath: item.Path + "/manifest.yaml",
			TemplatesDir: "knossos/templates",
		}
	}

	// Map tier label to source
	source := r.tierToSource(item.Source)
	return &ResolvedRite{
		Name:         riteName,
		Source:       source,
		RitePath:     item.Path,
		ManifestPath: filepath.Join(item.Path, "manifest.yaml"),
		TemplatesDir: r.resolveTemplatesDir(source),
	}
}

// tierToSource maps a resolution chain tier label to a RiteSource.
func (r *SourceResolver) tierToSource(tierLabel string) RiteSource {
	switch tierLabel {
	case "project":
		return RiteSource{Type: SourceProject, Path: r.projectRitesDir, Description: "project rites directory"}
	case "user":
		return RiteSource{Type: SourceUser, Path: r.userRitesDir, Description: "user-level rites"}
	case "org":
		return RiteSource{Type: SourceOrg, Path: r.orgRitesDir, Description: "org-level rites"}
	case "platform":
		return RiteSource{Type: SourceKnossos, Path: filepath.Join(r.knossosHome, "rites"),
			Description: fmt.Sprintf("Knossos platform at %s", r.knossosHome)}
	default:
		return RiteSource{Type: SourceExplicit, Path: "", Description: "unknown source"}
	}
}

// ClearCache clears the resolution cache.
func (r *SourceResolver) ClearCache() {
	r.mu.Lock()
	r.resolved = make(map[string]*ResolvedRite)
	r.mu.Unlock()
}
