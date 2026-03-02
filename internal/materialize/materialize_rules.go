package materialize

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/autom8y/knossos/internal/checksum"
	"github.com/autom8y/knossos/internal/paths"
	"github.com/autom8y/knossos/internal/provenance"
)

// knownRuleTemplateNames returns the set of knossos-managed rule filenames from the
// filesystem templates dir. Used by materializeRules to identify stale files to remove.
func (m *Materializer) knownRuleTemplateNames(_ *ResolvedRite) map[string]bool {
	names := make(map[string]bool)
	fsRulesDir := filepath.Join(m.templatesDir, "rules")
	if entries, err := os.ReadDir(fsRulesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				names[entry.Name()] = true
			}
		}
	}
	return names
}

// materializeRules copies rule files from templates/rules to .claude/rules/
// Platform rules are overwritten from templates; user-created rules are preserved.
// On rite switch, stale knossos-managed rules are removed before writing new ones.
// Provenance is determined by template filename: any .md file whose name matches
// a template source file is knossos-managed; all others are user-created.
func (m *Materializer) materializeRules(claudeDir string, resolved *ResolvedRite, collector provenance.Collector) error {
	rulesDir := filepath.Join(claudeDir, "rules")
	if err := paths.EnsureDir(rulesDir); err != nil {
		return err
	}

	projectRoot := m.resolver.ProjectRoot()

	// Skip knossos-internal rules when templates come from outside the project.
	// Template rules (trigger on internal/**, rites/**, knossos/**) are development
	// guides for the knossos codebase — harmful noise on foreign projects.
	// Also clean up stale knossos-managed rules from any previous sync.
	if m.templatesDir != "" && !strings.HasPrefix(m.templatesDir, projectRoot) {
		knownTemplateNames := m.knownRuleTemplateNames(resolved)
		if existingRules, err := os.ReadDir(rulesDir); err == nil {
			for _, entry := range existingRules {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				if knownTemplateNames[entry.Name()] {
					os.Remove(filepath.Join(rulesDir, entry.Name()))
				}
			}
		}
		return nil
	}
	if resolved != nil && resolved.Source.Type == SourceEmbedded {
		knownTemplateNames := m.knownRuleTemplateNames(resolved)
		if existingRules, err := os.ReadDir(rulesDir); err == nil {
			for _, entry := range existingRules {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				if knownTemplateNames[entry.Name()] {
					os.Remove(filepath.Join(rulesDir, entry.Name()))
				}
			}
		}
		return nil
	}

	// Collect template rule names and content from filesystem templates dir.
	templateRules := make(map[string][]byte)

	{
		sourceRulesDir := filepath.Join(m.templatesDir, "rules")
		entries, err := os.ReadDir(sourceRulesDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil // No template rules = no-op
			}
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			content, err := os.ReadFile(filepath.Join(sourceRulesDir, entry.Name()))
			if err != nil {
				return err
			}
			templateRules[entry.Name()] = content
		}
	}

	// Build the complete set of known template names for stale rule detection.
	// templateRules already contains all filesystem template entries.
	allTemplateNames := make(map[string]bool)
	for name := range templateRules {
		allTemplateNames[name] = true
	}

	// Remove only STALE knossos-managed rules: files that match a known template name
	// but are NOT in the current rite's template set. Do NOT pre-delete rules that will
	// be rewritten — writeIfChanged() handles atomic overwrite. Pre-deletion causes
	// CC's file watcher to see DELETE events that crash active sessions.
	if existingRules, err := os.ReadDir(rulesDir); err == nil {
		for _, entry := range existingRules {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			if allTemplateNames[entry.Name()] && templateRules[entry.Name()] == nil {
				os.Remove(filepath.Join(rulesDir, entry.Name()))
			}
		}
	}

	// Write current template rules and record provenance
	for name, content := range templateRules {
		dstPath := filepath.Join(rulesDir, name)
		written, err := writeIfChanged(dstPath, content, 0644)
		if err != nil {
			return err
		}
		if written {
			sourcePath := filepath.Join(m.templatesDir, "rules", name)
			srcRelPath, _ := filepath.Rel(projectRoot, sourcePath)
			collector.Record("rules/"+name, provenance.NewKnossosEntry(
				provenance.ScopeRite,
				srcRelPath,
				"template",
				checksum.Bytes(content),
			))
		}
	}

	return nil
}
