// Package ledge implements ledge artifact management operations.
package ledge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// PromoteResult contains the outcome of a promote operation.
type PromoteResult struct {
	SourcePath string `json:"source_path"`
	ShelfPath  string `json:"shelf_path"`
	Category   string `json:"category"`
}

// validCategories are the ledge categories that can be promoted to shelf.
var validCategories = map[string]bool{
	"decisions": true,
	"specs":     true,
	"reviews":   true,
}

// Promote moves an artifact from .ledge/{category}/ to .ledge/shelf/{category}/,
// adding promotion frontmatter. Returns an error if the source doesn't exist,
// isn't in a valid ledge category, or the destination already exists.
func Promote(resolver *paths.Resolver, sourcePath string) (*PromoteResult, error) {
	projectRoot := resolver.ProjectRoot()

	// Resolve absolute path
	absSource := sourcePath
	if !filepath.IsAbs(absSource) {
		absSource = filepath.Join(projectRoot, absSource)
	}

	// Verify source exists
	if _, err := os.Stat(absSource); err != nil {
		return nil, fmt.Errorf("source not found: %s", sourcePath)
	}

	// Determine category from path
	ledgeDir := resolver.LedgeDir()
	relToLedge, err := filepath.Rel(ledgeDir, absSource)
	if err != nil || strings.HasPrefix(relToLedge, "..") {
		return nil, fmt.Errorf("source %s is not inside .ledge/", sourcePath)
	}

	parts := strings.SplitN(relToLedge, string(filepath.Separator), 2)
	if len(parts) < 2 {
		return nil, errors.New(errors.CodeUsageError, "source must be inside a category directory (e.g., .ledge/reviews/file.md)")
	}

	category := parts[0]
	filename := parts[1]

	if !validCategories[category] {
		return nil, fmt.Errorf("category %q is not promotable (valid: decisions, specs, reviews)", category)
	}

	// Build destination path
	shelfDir := filepath.Join(resolver.LedgeShelfDir(), category)
	destPath := filepath.Join(shelfDir, filename)

	// Check if already promoted
	if _, err := os.Stat(destPath); err == nil {
		return nil, fmt.Errorf("destination already exists: .ledge/shelf/%s/%s", category, filename)
	}

	// Read source content
	content, err := os.ReadFile(absSource)
	if err != nil {
		return nil, fmt.Errorf("cannot read source: %w", err)
	}

	// Ensure shelf category directory exists
	if err := paths.EnsureDir(shelfDir); err != nil {
		return nil, fmt.Errorf("cannot create shelf directory: %w", err)
	}

	// Add promotion frontmatter
	now := time.Now().UTC().Format(time.RFC3339)
	relSource, _ := filepath.Rel(projectRoot, absSource)
	promoted := stampPromotionFrontmatter(content, now, relSource)

	// Write to shelf
	if err := os.WriteFile(destPath, promoted, 0644); err != nil {
		return nil, fmt.Errorf("cannot write to shelf: %w", err)
	}

	// Remove source (move semantics)
	if err := os.Remove(absSource); err != nil {
		return nil, fmt.Errorf("promoted to shelf but cannot remove source: %w", err)
	}

	relDest, _ := filepath.Rel(projectRoot, destPath)

	return &PromoteResult{
		SourcePath: relSource,
		ShelfPath:  relDest,
		Category:   category,
	}, nil
}

// stampPromotionFrontmatter adds promoted_at and promoted_from fields to content.
// If the content already has YAML frontmatter, the fields are inserted before the
// closing delimiter. If not, a new frontmatter block is prepended.
func stampPromotionFrontmatter(content []byte, promotedAt, promotedFrom string) []byte {
	promoLines := fmt.Sprintf("promoted_at: %s\npromoted_from: %s\n", promotedAt, promotedFrom)

	s := string(content)

	// Check if content has existing frontmatter
	if strings.HasPrefix(s, "---\n") {
		// Find closing delimiter
		closeIdx := strings.Index(s[4:], "\n---\n")
		if closeIdx >= 0 {
			// Insert promotion fields before closing delimiter
			insertAt := 4 + closeIdx + 1 // right before "---\n"
			return []byte(s[:insertAt] + promoLines + s[insertAt:])
		}
	}

	// No existing frontmatter — prepend new block
	fm := fmt.Sprintf("---\n%s---\n\n", promoLines)
	return append([]byte(fm), content...)
}
