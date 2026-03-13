package perspective

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/agent"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/frontmatter"
	"github.com/autom8y/knossos/internal/provenance"
	"gopkg.in/yaml.v3"
)

// PerspectiveOptions configures the perspective assembly pipeline.
type PerspectiveOptions struct {
	AgentName   string
	RiteName    string // empty = read from ACTIVE_RITE
	Mode        string // "default" or "audit"
	ProjectRoot string
}

// ParseContext holds all parsed source data shared across layer resolvers.
// It is populated once by NewParseContext and passed to each resolver function.
type ParseContext struct {
	// Agent data
	AgentFrontmatter    *agent.AgentFrontmatter
	AgentFrontmatterRaw map[string]any
	AgentBody           []byte

	// Manifest data (generic YAML for fields not in typed structs)
	SharedManifest map[string]any
	RiteManifest   map[string]any

	// Provenance
	Provenance *provenance.ProvenanceManifest

	// Workflow and orchestrator data (generic YAML for phase/routing fields)
	Workflow    map[string]any // parsed workflow.yaml
	Orchestrator map[string]any // parsed orchestrator.yaml

	// Materialized skills directories (dir names under channel skills/)
	MaterializedSkillsDirs []string

	// Resolved paths
	RiteSourcePath  string // absolute path to rite source directory
	AgentSourcePath string // absolute path to agent source file
	ProjectRoot     string
	KnossosDir      string
	ChannelDir      string
	RiteName        string
}

// NewParseContext builds a ParseContext by resolving the rite, reading the agent
// source file, and parsing all required manifests.
func NewParseContext(opts PerspectiveOptions) (*ParseContext, error) {
	ctx := &ParseContext{
		ProjectRoot: opts.ProjectRoot,
		KnossosDir:  filepath.Join(opts.ProjectRoot, ".knossos"),
		ChannelDir:  filepath.Join(opts.ProjectRoot, ".claude"), // HA-FS: actual CC channel directory path (SCAR-002)
	}

	// 1. Resolve rite name
	riteName := opts.RiteName
	if riteName == "" {
		data, err := os.ReadFile(filepath.Join(ctx.KnossosDir, "ACTIVE_RITE"))
		if err != nil {
			return nil, errors.Wrap(errors.CodeFileNotFound, "failed to read ACTIVE_RITE", err)
		}
		riteName = strings.TrimSpace(string(data))
	}
	ctx.RiteName = riteName

	// 2. Resolve rite source path using simple resolution
	// Check project rites first, then knossos platform rites
	riteSourcePath := resolveRiteSourcePath(opts.ProjectRoot, riteName)
	if riteSourcePath == "" {
		return nil, errors.NewWithDetails(errors.CodeFileNotFound,
			"rite source directory not found",
			map[string]any{"rite": riteName})
	}
	ctx.RiteSourcePath = riteSourcePath

	// 3. Read agent source file
	agentPath := filepath.Join(riteSourcePath, "agents", opts.AgentName+".md")
	agentContent, err := os.ReadFile(agentPath)
	if err != nil {
		return nil, errors.Wrap(errors.CodeFileNotFound,
			"agent source file not found: "+agentPath, err)
	}
	ctx.AgentSourcePath = agentPath

	// 4. Parse frontmatter via typed struct
	fm, err := agent.ParseAgentFrontmatter(agentContent)
	if err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse agent frontmatter", err)
	}
	ctx.AgentFrontmatter = fm

	// 5. Parse frontmatter into raw map for knossos-only fields
	yamlBytes, body, err := frontmatter.Parse(agentContent)
	if err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to parse agent frontmatter YAML", err)
	}
	ctx.AgentBody = body

	var rawMap map[string]any
	if err := yaml.Unmarshal(yamlBytes, &rawMap); err != nil {
		return nil, errors.Wrap(errors.CodeParseError, "failed to unmarshal agent frontmatter to map", err)
	}
	ctx.AgentFrontmatterRaw = rawMap

	// 6. Parse shared manifest
	sharedManifestPath := filepath.Join(opts.ProjectRoot, "rites", "shared", "manifest.yaml")
	ctx.SharedManifest = loadManifestAsMap(sharedManifestPath)

	// 7. Parse rite manifest
	riteManifestPath := filepath.Join(riteSourcePath, "manifest.yaml")
	ctx.RiteManifest = loadManifestAsMap(riteManifestPath)

	// 8. Parse workflow
	workflowPath := filepath.Join(riteSourcePath, "workflow.yaml")
	ctx.Workflow = loadManifestAsMap(workflowPath)

	// 9. Parse orchestrator
	orchestratorPath := filepath.Join(riteSourcePath, "orchestrator.yaml")
	ctx.Orchestrator = loadManifestAsMap(orchestratorPath)

	// 10. List materialized skills directories
	ctx.MaterializedSkillsDirs = listSkillsDirs(filepath.Join(ctx.ChannelDir, "skills"))

	// 11. Load provenance manifest
	provPath := provenance.ManifestPath(ctx.KnossosDir)
	prov, err := provenance.LoadOrBootstrap(provPath)
	if err != nil {
		// Non-fatal: use empty manifest
		prov = &provenance.ProvenanceManifest{
			Entries: make(map[string]*provenance.ProvenanceEntry),
		}
	}
	ctx.Provenance = prov

	return ctx, nil
}

// resolveRiteSourcePath returns the absolute path to a rite's source directory.
// It checks project rites/ first, then falls back to the embedded/knossos rites.
func resolveRiteSourcePath(projectRoot, riteName string) string {
	// 1. Check project-level rites directory (rites/<name>/)
	projectPath := filepath.Join(projectRoot, "rites", riteName)
	if hasManifest(projectPath) {
		return projectPath
	}

	// 2. Check satellite rites (.knossos/rites/<name>/)
	satellitePath := filepath.Join(projectRoot, ".knossos", "rites", riteName)
	if hasManifest(satellitePath) {
		return satellitePath
	}

	return ""
}

// hasManifest returns true if the directory contains a manifest.yaml file.
func hasManifest(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "manifest.yaml"))
	return err == nil
}

// listSkillsDirs returns the names of subdirectories under a skills directory.
// Returns nil if the directory cannot be read.
func listSkillsDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

// loadManifestAsMap reads a YAML manifest file into a generic map.
// Returns an empty map if the file cannot be read or parsed.
func loadManifestAsMap(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return map[string]any{}
	}
	return m
}
