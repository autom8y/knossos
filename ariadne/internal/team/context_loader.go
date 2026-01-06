// Package team implements team pack discovery, management, and switching for Ariadne.
// This file contains the context loader for YAML-based team context injection.
package team

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/paths"
)

// ContextFileName is the standard name for team context YAML files.
const ContextFileName = "context.yaml"

// ContextLoader handles loading and caching of team context files.
type ContextLoader struct {
	teamsDir string
	userDir  string

	mu    sync.RWMutex
	cache map[string]*TeamContext
}

// NewContextLoader creates a new context loader using the paths resolver.
func NewContextLoader(resolver *paths.Resolver) *ContextLoader {
	return &ContextLoader{
		teamsDir: resolver.TeamsDir(),
		userDir:  paths.UserTeamsDir(),
		cache:    make(map[string]*TeamContext),
	}
}

// NewContextLoaderWithPaths creates a context loader with explicit paths.
func NewContextLoaderWithPaths(teamsDir, userDir string) *ContextLoader {
	return &ContextLoader{
		teamsDir: teamsDir,
		userDir:  userDir,
		cache:    make(map[string]*TeamContext),
	}
}

// Load returns the team context for the given team name.
// It first checks the cache, then tries to load from:
// 1. User rites directory ($XDG_DATA_HOME/ariadne/rites/{team}/context.yaml)
// 2. Project rites directory (rites/{team}/context.yaml)
// 3. Fallback: generates context from orchestrator.yaml
func (cl *ContextLoader) Load(teamName string) (*TeamContext, error) {
	if teamName == "" {
		return nil, errors.New(errors.CodeUsageError, "team name is required")
	}

	// Check cache first
	cl.mu.RLock()
	if ctx, ok := cl.cache[teamName]; ok {
		cl.mu.RUnlock()
		return ctx, nil
	}
	cl.mu.RUnlock()

	// Try to load from files
	ctx, err := cl.loadFromFiles(teamName)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cl.mu.Lock()
	cl.cache[teamName] = ctx
	cl.mu.Unlock()

	return ctx, nil
}

// loadFromFiles attempts to load context from YAML files or fallback to orchestrator.
func (cl *ContextLoader) loadFromFiles(teamName string) (*TeamContext, error) {
	// Try user teams directory first (higher priority)
	if cl.userDir != "" {
		contextPath := filepath.Join(cl.userDir, teamName, ContextFileName)
		if ctx, err := cl.loadFromYAML(contextPath); err == nil {
			return ctx, nil
		}
	}

	// Try project teams directory
	if cl.teamsDir != "" {
		contextPath := filepath.Join(cl.teamsDir, teamName, ContextFileName)
		if ctx, err := cl.loadFromYAML(contextPath); err == nil {
			return ctx, nil
		}
	}

	// Fallback: generate from orchestrator.yaml
	return cl.generateFromOrchestrator(teamName)
}

// loadFromYAML loads a TeamContext from a YAML file.
func (cl *ContextLoader) loadFromYAML(path string) (*TeamContext, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ctx TeamContext
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, errors.ErrParseError(path, "YAML", err)
	}

	// Validate the loaded context
	if err := ctx.Validate(); err != nil {
		return nil, errors.Wrap(errors.CodeSchemaInvalid, "invalid team context", err)
	}

	return &ctx, nil
}

// generateFromOrchestrator creates a TeamContext from an orchestrator.yaml file.
// This provides backward compatibility when no context.yaml exists.
func (cl *ContextLoader) generateFromOrchestrator(teamName string) (*TeamContext, error) {
	// Try to find orchestrator.yaml
	var orchestratorPath string
	var found bool

	// Check user teams
	if cl.userDir != "" {
		path := filepath.Join(cl.userDir, teamName, "orchestrator.yaml")
		if _, err := os.Stat(path); err == nil {
			orchestratorPath = path
			found = true
		}
	}

	// Check project teams
	if !found && cl.teamsDir != "" {
		path := filepath.Join(cl.teamsDir, teamName, "orchestrator.yaml")
		if _, err := os.Stat(path); err == nil {
			orchestratorPath = path
			found = true
		}
	}

	if !found {
		return nil, errors.ErrTeamNotFound(teamName)
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
	ctx := NewTeamContext(teamName)
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

// Invalidate removes a team from the cache, forcing a reload on next access.
func (cl *ContextLoader) Invalidate(teamName string) {
	cl.mu.Lock()
	delete(cl.cache, teamName)
	cl.mu.Unlock()
}

// InvalidateAll clears the entire cache.
func (cl *ContextLoader) InvalidateAll() {
	cl.mu.Lock()
	cl.cache = make(map[string]*TeamContext)
	cl.mu.Unlock()
}

// IsCached returns true if the team context is in the cache.
func (cl *ContextLoader) IsCached(teamName string) bool {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	_, ok := cl.cache[teamName]
	return ok
}

// GetContextPath returns the path where context.yaml would be for a team.
// Returns the first path that exists, or the project path if none exists.
func (cl *ContextLoader) GetContextPath(teamName string) string {
	// Check user teams first
	if cl.userDir != "" {
		path := filepath.Join(cl.userDir, teamName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check project teams
	if cl.teamsDir != "" {
		path := filepath.Join(cl.teamsDir, teamName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return where it would be (project teams)
	return filepath.Join(cl.teamsDir, teamName, ContextFileName)
}

// HasContextFile checks if a team has a context.yaml file.
func (cl *ContextLoader) HasContextFile(teamName string) bool {
	// Check user teams first
	if cl.userDir != "" {
		path := filepath.Join(cl.userDir, teamName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Check project teams
	if cl.teamsDir != "" {
		path := filepath.Join(cl.teamsDir, teamName, ContextFileName)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// SaveContext writes a TeamContext to the team's context.yaml file.
func (cl *ContextLoader) SaveContext(ctx *TeamContext) error {
	if err := ctx.Validate(); err != nil {
		return err
	}

	// Default to project teams directory
	teamDir := filepath.Join(cl.teamsDir, ctx.TeamName)
	if _, err := os.Stat(teamDir); os.IsNotExist(err) {
		return errors.ErrTeamNotFound(ctx.TeamName)
	}

	contextPath := filepath.Join(teamDir, ContextFileName)

	data, err := yaml.Marshal(ctx)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal context", err)
	}

	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to write context file", err)
	}

	// Invalidate cache
	cl.Invalidate(ctx.TeamName)

	return nil
}
