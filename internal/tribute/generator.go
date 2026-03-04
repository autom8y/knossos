package tribute

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// Generator creates TRIBUTE.md for a session.
type Generator struct {
	// SessionPath is the path to the session directory.
	SessionPath string

	// ProjectRoot is the project root directory (optional, enables graduated artifacts).
	ProjectRoot string

	// Now is a function that returns the current time (for testing).
	Now func() time.Time
}

// NewGenerator creates a new Generator for the given session.
func NewGenerator(sessionPath string) *Generator {
	return &Generator{
		SessionPath: sessionPath,
		Now:         time.Now,
	}
}

// Generate creates TRIBUTE.md and returns the result.
func (g *Generator) Generate() (*GenerateResult, error) {
	if g.SessionPath == "" {
		return nil, errors.New(errors.CodeUsageError, "session path is required")
	}

	// Verify session directory exists
	info, err := os.Stat(g.SessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewWithDetails(errors.CodeSessionNotFound,
				"session directory not found",
				map[string]any{"path": g.SessionPath})
		}
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to access session directory", err)
	}
	if !info.IsDir() {
		return nil, errors.NewWithDetails(errors.CodeUsageError,
			"path is not a directory",
			map[string]any{"path": g.SessionPath})
	}

	extractor := NewExtractor(g.SessionPath)

	// Step 1: Load SESSION_CONTEXT.md (required)
	ctx, err := extractor.ExtractSessionContext()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to load SESSION_CONTEXT.md", err)
	}

	// Step 2: Extract events (graceful degradation if missing)
	events, err := extractor.ExtractEvents()
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to extract events", err)
	}

	// Step 3: Extract data from events
	artifacts := extractor.ExtractArtifacts(events)
	decisions := extractor.ExtractDecisions(events)
	phases := extractor.ExtractPhases(events)
	handoffs := extractor.ExtractHandoffs(events)
	metrics := extractor.ExtractMetrics(events)

	// Step 3b: Extract graduated artifacts from registry (graceful degradation)
	var graduatedArtifacts []GraduatedArtifact
	if g.ProjectRoot != "" && ctx.SessionID != "" {
		graduatedArtifacts = extractor.ExtractGraduatedArtifacts(ctx.SessionID, g.ProjectRoot)
	}

	// Step 4: Load WHITE_SAILS.yaml (graceful degradation if missing)
	sailsData, _ := extractor.ExtractWhiteSails() // Ignore error - optional

	// Step 5: Extract notes from body
	notes := extractor.ExtractNotes(ctx.Body)

	// Step 6: Calculate timing
	now := g.Now().UTC()
	startedAt := ctx.CreatedAt
	endedAt := now
	if ctx.ArchivedAt != nil {
		endedAt = *ctx.ArchivedAt
	}
	duration := endedAt.Sub(startedAt)

	// Build result
	result := &GenerateResult{
		FilePath:           filepath.Join(g.SessionPath, "TRIBUTE.md"),
		SessionID:          ctx.SessionID,
		Initiative:         ctx.Initiative,
		Complexity:         ctx.Complexity,
		Duration:           duration,
		Rite:               ctx.ActiveRite,
		FinalPhase:         ctx.CurrentPhase,
		StartedAt:          startedAt,
		EndedAt:            endedAt,
		Artifacts:          artifacts,
		Decisions:          decisions,
		Phases:             phases,
		Handoffs:           handoffs,
		GraduatedArtifacts: graduatedArtifacts,
		Commits:            []Commit{}, // Phase 2 - empty for now
		Metrics:            metrics,
		Notes:              notes,
		GeneratedAt:        now,
	}

	// Add sails data if available
	if sailsData != nil {
		result.SailsColor = sailsData.Color
		result.SailsBase = sailsData.ComputedBase
		result.SailsProofs = sailsData.Proofs
	}

	// Step 7: Render TRIBUTE.md
	renderer := NewRenderer()
	content, err := renderer.Render(result)
	if err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to render TRIBUTE.md", err)
	}

	// Step 8: Write TRIBUTE.md (idempotent - overwrites existing)
	if err := os.WriteFile(result.FilePath, content, 0644); err != nil {
		return nil, errors.Wrap(errors.CodeGeneralError, "failed to write TRIBUTE.md", err)
	}

	return result, nil
}

// GenerateFromProject creates a Generator for a specific session in a project.
func GenerateFromProject(projectRoot string, sessionID string) (*Generator, error) {
	if projectRoot == "" {
		return nil, errors.New(errors.CodeUsageError, "project root is required")
	}
	if sessionID == "" {
		return nil, errors.New(errors.CodeSessionNotFound, "no session ID provided")
	}

	resolver := paths.NewResolver(projectRoot)
	sessionDir := resolver.SessionDir(strings.TrimSpace(sessionID))

	// Verify session directory exists
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return nil, errors.New(errors.CodeSessionNotFound, "session directory not found: "+sessionID)
	}

	gen := NewGenerator(sessionDir)
	gen.ProjectRoot = projectRoot
	return gen, nil
}

// GenerateFromSessionID creates a Generator for a specific session ID.
func GenerateFromSessionID(projectRoot, sessionID string) (*Generator, error) {
	resolver := paths.NewResolver(projectRoot)

	// Check sessions directory first
	sessionPath := resolver.SessionDir(sessionID)
	if _, err := os.Stat(filepath.Join(sessionPath, "SESSION_CONTEXT.md")); err == nil {
		gen := NewGenerator(sessionPath)
		gen.ProjectRoot = projectRoot
		return gen, nil
	}

	// Check archive directory
	archivePath := filepath.Join(resolver.ArchiveDir(), sessionID)
	if _, err := os.Stat(filepath.Join(archivePath, "SESSION_CONTEXT.md")); err == nil {
		gen := NewGenerator(archivePath)
		gen.ProjectRoot = projectRoot
		return gen, nil
	}

	return nil, errors.ErrSessionNotFound(sessionID)
}
