// Package rite implements rite discovery, management, and switching for Ariadne.
// This file contains the context loader for YAML-based rite context injection.
package rite

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/paths"
)

// ContextFileName is the standard name for rite context YAML files.
const ContextFileName = "context.yaml"

// ContextLoader handles loading and caching of rite context files.
type ContextLoader struct {
	ritesDir string
	userDir  string

	mu    sync.RWMutex
	cache map[string]*RiteContext
}

// NewContextLoader creates a new context loader using the paths resolver.
func NewContextLoader(resolver *paths.Resolver) *ContextLoader {
	return &ContextLoader{
		ritesDir: resolver.RitesDir(),
		userDir:  paths.UserRitesDir(),
		cache:    make(map[string]*RiteContext),
	}
}

// NewContextLoaderWithPaths creates a context loader with explicit paths.
func NewContextLoaderWithPaths(ritesDir, userDir string) *ContextLoader {
	return &ContextLoader{
		ritesDir: ritesDir,
		userDir:  userDir,
		cache:    make(map[string]*RiteContext),
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

// loadFromFiles attempts to load context from YAML files or fallback to orchestrator.
func (cl *ContextLoader) loadFromFiles(riteName string) (*RiteContext, error) {
	// Try user teams directory first (higher priority)
	if cl.userDir != "" {
		contextPath := filepath.Join(cl.userDir, riteName, ContextFileName)
		if ctx, err := cl.loadFromYAML(contextPath); err == nil {
			return ctx, nil
		}
	}

	// Try project rites directory
	if cl.ritesDir != "" {
		contextPath := filepath.Join(cl.ritesDir, riteName, ContextFileName)
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
func (cl *ContextLoader) generateFromOrchestrator(riteName string) (*RiteContext, error) {
	// Try to find orchestrator.yaml
	var orchestratorPath string
	var found bool

	// Check user teams
	if cl.userDir != "" {
		path := filepath.Join(cl.userDir, riteName, "orchestrator.yaml")
		if _, err := os.Stat(path); err == nil {
			orchestratorPath = path
			found = true
		}
	}

	// Check project rites
	if !found && cl.ritesDir != "" {
		path := filepath.Join(cl.ritesDir, riteName, "orchestrator.yaml")
		if _, err := os.Stat(path); err == nil {
			orchestratorPath = path
			found = true
		}
	}

	if !found {
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
	ctx.Domain = orchestrator.Team.Domain

	// Add basic info rows
	if orchestrator.Team.Name != "" {
		ctx.AddRow("Team", orchestrator.Team.Name)
	}
	if orchestrator.Team.Domain != "" {
		ctx.AddRow("Domain", orchestrator.Team.Domain)
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
	Team struct {
		Name   string `yaml:"name"`
		Domain string `yaml:"domain"`
		Color  string `yaml:"color,omitempty"`
	} `yaml:"team"`
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
	// Check user rites first
	if cl.userDir != "" {
		path := filepath.Join(cl.userDir, riteName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check project rites
	if cl.ritesDir != "" {
		path := filepath.Join(cl.ritesDir, riteName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return where it would be (project rites)
	return filepath.Join(cl.ritesDir, riteName, ContextFileName)
}

// HasContextFile checks if a rite has a context.yaml file.
func (cl *ContextLoader) HasContextFile(riteName string) bool {
	// Check user rites first
	if cl.userDir != "" {
		path := filepath.Join(cl.userDir, riteName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Check project rites
	if cl.ritesDir != "" {
		path := filepath.Join(cl.ritesDir, riteName, ContextFileName)
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
