// Package rite implements rite discovery, management, and switching for Ariadne.
// This file contains the context loader for YAML-based rite context injection.
package rite

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/config"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// ContextFileName is the standard name for rite context YAML files.
const ContextFileName = "context.yaml"

// ContextLoader handles loading and caching of rite context files.
// Resolution order (highest priority first): user > project > org > platform.
type ContextLoader struct {
	ritesDir    string // project rites directory
	userDir     string // user rites directory (highest priority)
	orgDir      string // org rites directory
	platformDir string // platform rites directory (lowest priority)

	mu    sync.RWMutex
	cache map[string]*RiteContext
}

// NewContextLoader creates a new context loader using the paths resolver.
func NewContextLoader(resolver *paths.Resolver) *ContextLoader {
	return &ContextLoader{
		ritesDir:    resolver.RitesDir(),
		userDir:     paths.UserRitesDir(),
		orgDir:      paths.OrgRitesDir(config.ActiveOrg()),
		platformDir: PlatformRitesDir(),
		cache:       make(map[string]*RiteContext),
	}
}

// NewContextLoaderWithPaths creates a context loader with explicit paths.
// Resolution order (highest priority first): user > project > org > platform.
// Empty directory strings are silently skipped during resolution.
func NewContextLoaderWithPaths(ritesDir, userDir, orgDir, platformDir string) *ContextLoader {
	return &ContextLoader{
		ritesDir:    ritesDir,
		userDir:     userDir,
		orgDir:      orgDir,
		platformDir: platformDir,
		cache:       make(map[string]*RiteContext),
	}
}

// Load returns the rite context for the given rite name.
// It first checks the cache, then tries to load from:
// 1. User rites directory ($XDG_DATA_HOME/ariadne/rites/{rite}/context.yaml)
// 2. Project rites directory (rites/{rite}/context.yaml)
// 3. Fallback: generates context from orchestrator.yaml
func (cl *ContextLoader) Load(riteName string) (*RiteContext, error) {
	if riteName == "" {
		return nil, errors.New(errors.CodeUsageError, "rite name is required")
	}

	// Check cache first
	cl.mu.RLock()
	if ctx, ok := cl.cache[riteName]; ok {
		cl.mu.RUnlock()
		return ctx, nil
	}
	cl.mu.RUnlock()

	// Try to load from files
	ctx, err := cl.loadFromFiles(riteName)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cl.mu.Lock()
	cl.cache[riteName] = ctx
	cl.mu.Unlock()

	return ctx, nil
}

// contextDirs returns the search directories in priority order (user > project > org > platform).
// Empty directories are skipped.
func (cl *ContextLoader) contextDirs() []string {
	var dirs []string
	for _, d := range []string{cl.userDir, cl.ritesDir, cl.orgDir, cl.platformDir} {
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

// loadFromFiles attempts to load context from YAML files or fallback to orchestrator.
// Resolution order: user > project > org > platform.
func (cl *ContextLoader) loadFromFiles(riteName string) (*RiteContext, error) {
	for _, dir := range cl.contextDirs() {
		contextPath := filepath.Join(dir, riteName, ContextFileName)
		if ctx, err := cl.loadFromYAML(contextPath); err == nil {
			return ctx, nil
		}
	}

	// Fallback: generate from orchestrator.yaml
	return cl.generateFromOrchestrator(riteName)
}

// loadFromYAML loads a RiteContext from a YAML file.
func (cl *ContextLoader) loadFromYAML(path string) (*RiteContext, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ctx RiteContext
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, errors.ErrParseError(path, "YAML", err)
	}

	// Validate the loaded context
	if err := ctx.Validate(); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid rite context", err)
	}

	return &ctx, nil
}

// generateFromOrchestrator creates a RiteContext from an orchestrator.yaml file.
// This provides backward compatibility when no context.yaml exists.
// Searches all tiers in priority order: user > project > org > platform.
func (cl *ContextLoader) generateFromOrchestrator(riteName string) (*RiteContext, error) {
	// Try to find orchestrator.yaml across all tiers
	var orchestratorPath string
	for _, dir := range cl.contextDirs() {
		path := filepath.Join(dir, riteName, "orchestrator.yaml")
		if _, err := os.Stat(path); err == nil {
			orchestratorPath = path
			break
		}
	}

	if orchestratorPath == "" {
		return nil, errors.ErrRiteNotFound(riteName)
	}

	// Load orchestrator.yaml
	data, err := os.ReadFile(orchestratorPath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to read orchestrator.yaml", err)
	}

	var orchestrator OrchestratorConfig
	if err := yaml.Unmarshal(data, &orchestrator); err != nil {
		return nil, errors.ErrParseError(orchestratorPath, "YAML", err)
	}

	// Generate context from orchestrator
	ctx := NewRiteContext(riteName)
	ctx.Description = orchestrator.Frontmatter.Description
	ctx.Domain = orchestrator.Rite.Domain

	// Add basic info rows
	if orchestrator.Rite.Name != "" {
		ctx.AddRow("Rite", orchestrator.Rite.Name)
	}
	if orchestrator.Rite.Domain != "" {
		ctx.AddRow("Domain", orchestrator.Rite.Domain)
	}
	if orchestrator.Frontmatter.Role != "" {
		ctx.AddRow("Role", orchestrator.Frontmatter.Role)
	}

	// Add routing information
	for agent, desc := range orchestrator.Routing {
		ctx.AddRow(agent, desc)
	}

	return ctx, nil
}

// OrchestratorConfig represents the structure of orchestrator.yaml files.
// This is used for fallback context generation.
type OrchestratorConfig struct {
	Rite struct {
		Name   string `yaml:"name"`
		Domain string `yaml:"domain"`
		Color  string `yaml:"color,omitempty"`
	} `yaml:"rite"`
	Frontmatter struct {
		Role        string `yaml:"role"`
		Description string `yaml:"description"`
	} `yaml:"frontmatter"`
	Routing          map[string]string `yaml:"routing"`
	WorkflowPosition struct {
		Upstream   string `yaml:"upstream"`
		Downstream string `yaml:"downstream"`
	} `yaml:"workflow_position"`
	HandoffCriteria map[string][]string `yaml:"handoff_criteria,omitempty"`
	Skills          []string            `yaml:"skills,omitempty"`
	Antipatterns    []string            `yaml:"antipatterns,omitempty"`
}

// Invalidate removes a rite from the cache, forcing a reload on next access.
func (cl *ContextLoader) Invalidate(riteName string) {
	cl.mu.Lock()
	delete(cl.cache, riteName)
	cl.mu.Unlock()
}

// InvalidateAll clears the entire cache.
func (cl *ContextLoader) InvalidateAll() {
	cl.mu.Lock()
	cl.cache = make(map[string]*RiteContext)
	cl.mu.Unlock()
}

// IsCached returns true if the rite context is in the cache.
func (cl *ContextLoader) IsCached(riteName string) bool {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	_, ok := cl.cache[riteName]
	return ok
}

// GetContextPath returns the path where context.yaml would be for a rite.
// Returns the first path that exists, or the project path if none exists.
func (cl *ContextLoader) GetContextPath(riteName string) string {
	for _, dir := range cl.contextDirs() {
		path := filepath.Join(dir, riteName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return where it would be (project rites)
	return filepath.Join(cl.ritesDir, riteName, ContextFileName)
}

// HasContextFile checks if a rite has a context.yaml file.
func (cl *ContextLoader) HasContextFile(riteName string) bool {
	for _, dir := range cl.contextDirs() {
		path := filepath.Join(dir, riteName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// SaveContext writes a RiteContext to the rite's context.yaml file.
func (cl *ContextLoader) SaveContext(ctx *RiteContext) error {
	if err := ctx.Validate(); err != nil {
		return err
	}

	// Default to project rites directory
	riteDir := filepath.Join(cl.ritesDir, ctx.RiteName)
	if _, err := os.Stat(riteDir); os.IsNotExist(err) {
		return errors.ErrRiteNotFound(ctx.RiteName)
	}

	contextPath := filepath.Join(riteDir, ContextFileName)

	data, err := yaml.Marshal(ctx)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal context", err)
	}

	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to write context file", err)
	}

	// Invalidate cache
	cl.Invalidate(ctx.RiteName)

	return nil
}
