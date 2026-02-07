package usersync

import (
	"fmt"
	"strings"
)

// Text implements output.Textable for Result.
func (r Result) Text() string {
	var b strings.Builder

	if r.DryRun {
		b.WriteString("[DRY RUN] ")
	}
	b.WriteString(fmt.Sprintf("Syncing user %s...\n", r.Resource))
	b.WriteString(fmt.Sprintf("  Source: %s\n", r.Source))
	b.WriteString(fmt.Sprintf("  Target: %s\n", r.Target))
	b.WriteString("\n")

	// Report changes
	for _, name := range r.Changes.Added {
		b.WriteString(fmt.Sprintf("  Added: %s\n", name))
	}
	for _, name := range r.Changes.Updated {
		b.WriteString(fmt.Sprintf("  Updated: %s\n", name))
	}
	for _, entry := range r.Changes.Skipped {
		if strings.Contains(entry.Reason, "collision") {
			b.WriteString(fmt.Sprintf("  Collision: %s (%s)\n", entry.Name, entry.Reason))
		} else {
			b.WriteString(fmt.Sprintf("  Skipped: %s (%s)\n", entry.Name, entry.Reason))
		}
	}

	// Summary line
	b.WriteString(fmt.Sprintf("\nSummary: %d added, %d updated, %d skipped, %d unchanged",
		r.Summary.Added, r.Summary.Updated, r.Summary.Skipped, r.Summary.Unchanged))
	if r.Summary.Collisions > 0 {
		b.WriteString(fmt.Sprintf(", %d collisions", r.Summary.Collisions))
	}
	b.WriteString("\n")

	return b.String()
}

// AllResult contains results for all resource types.
type AllResult struct {
	SyncedAt  string            `json:"synced_at"`
	DryRun    bool              `json:"dry_run"`
	Resources map[string]Result `json:"resources"`
	Totals    Summary           `json:"totals"`
	Errors    []ResourceError   `json:"errors,omitempty"`
}

// ResourceError captures an error for a specific resource type.
type ResourceError struct {
	Resource ResourceType `json:"resource"`
	Err      string       `json:"error"`
}

// Error implements the error interface.
func (e ResourceError) Error() string {
	return e.Err
}

// Text implements output.Textable for AllResult.
func (r AllResult) Text() string {
	var b strings.Builder

	if r.DryRun {
		b.WriteString("[DRY RUN] ")
	}
	b.WriteString("Syncing all user resources...\n\n")

	// Resource summaries
	order := []ResourceType{ResourceAgents, ResourceMena, ResourceHooks}
	for _, rt := range order {
		if result, ok := r.Resources[string(rt)]; ok {
			b.WriteString(fmt.Sprintf("%s: %d added, %d updated, %d skipped, %d unchanged",
				capitalizeFirst(string(rt)),
				result.Summary.Added,
				result.Summary.Updated,
				result.Summary.Skipped,
				result.Summary.Unchanged))
			if result.Summary.Collisions > 0 {
				b.WriteString(fmt.Sprintf(", %d collisions", result.Summary.Collisions))
			}
			b.WriteString("\n")
		}
	}

	// Errors
	for _, e := range r.Errors {
		b.WriteString(fmt.Sprintf("%s: ERROR - %s\n", capitalizeFirst(string(e.Resource)), e.Err))
	}

	// Totals
	b.WriteString(fmt.Sprintf("\nTotals: %d added, %d updated, %d skipped, %d unchanged",
		r.Totals.Added, r.Totals.Updated, r.Totals.Skipped, r.Totals.Unchanged))
	if r.Totals.Collisions > 0 {
		b.WriteString(fmt.Sprintf(", %d collisions", r.Totals.Collisions))
	}
	b.WriteString("\n")

	return b.String()
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
