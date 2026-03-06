package ledge

import (
	"fmt"
	"strings"

	"github.com/autom8y/knossos/internal/artifact"
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/paths"
)

// PromotedEntry represents a single artifact promoted to shelf.
type PromotedEntry struct {
	SourcePath string `json:"source_path"` // .ledge/{category}/{file}
	ShelfPath  string `json:"shelf_path"`  // .ledge/shelf/{category}/{file}
	Category   string `json:"category"`    // decisions, specs, reviews
}

// AutoPromoteResult contains the outcome of auto-promoting session artifacts.
type AutoPromoteResult struct {
	Promoted []PromotedEntry `json:"promoted"`
	Skipped  int             `json:"skipped"`            // Non-promotable categories
	Warnings []string        `json:"warnings,omitempty"` // Per-artifact errors that were handled
}

// AutoPromoteSession promotes graduated artifacts to .ledge/shelf/{category}/.
// Artifacts with non-promotable categories (e.g., spikes) are silently skipped.
// Per-artifact failures are captured as warnings, not errors.
// The caller is responsible for gating on sails color before calling this.
func AutoPromoteSession(resolver *paths.Resolver, graduated []artifact.GraduatedEntry) (*AutoPromoteResult, error) {
	if resolver == nil {
		return nil, errors.New(errors.CodeUsageError, "resolver is required")
	}

	result := &AutoPromoteResult{}

	for _, entry := range graduated {
		if !validCategories[entry.Category] {
			result.Skipped++
			continue
		}

		promoResult, err := Promote(resolver, entry.GraduatedPath)
		if err != nil {
			// Source not found: silently skip (idempotency — already promoted or missing)
			if strings.Contains(err.Error(), "source not found") {
				continue
			}
			// All other errors: warn and continue
			result.Warnings = append(result.Warnings, fmt.Sprintf("promote %s: %v", entry.GraduatedPath, err))
			continue
		}

		result.Promoted = append(result.Promoted, PromotedEntry{
			SourcePath: promoResult.SourcePath,
			ShelfPath:  promoResult.ShelfPath,
			Category:   promoResult.Category,
		})
	}

	return result, nil
}
