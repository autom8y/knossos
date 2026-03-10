package procession

import (
	"io/fs"
	"os"
	"path/filepath"
)

// RenderFunc renders an archetype template with data and returns the content.
// This is a function type to break the import cycle between materialize and
// materialize/procession.
type RenderFunc func(projectRoot, templateName string, data any) ([]byte, error)

// RenderToDir resolves all procession templates and renders their dromena
// (commands) and legomena (skills) into tmpDir.
//
// The currentRite parameter controls dromena projection:
//   - Non-empty: only render dromena when template's entry rite matches currentRite
//   - Empty: render legomena only (minimal/cross-cutting mode with no active rite)
//
// Legomena (skills) are always rendered regardless of currentRite — skills are
// universal reference material loaded on-demand via Skill().
//
// For each template, it creates (when applicable):
//   - {tmpDir}/{name}/INDEX.dro.md  (procession-workflow archetype, rite-filtered)
//   - {tmpDir}/{name}-ref/INDEX.lego.md (procession-ref archetype, always)
//
// Returns the number of templates where any artifact was rendered.
func RenderToDir(projectRoot string, tmpDir string, render RenderFunc, currentRite string, embeddedFS fs.FS) (int, error) {
	resolved, err := ResolveProcessions(projectRoot, embeddedFS)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, rp := range resolved {
		data := BuildWorkflowData(rp.Template)
		rendered := false

		// Render dromena (command) only when rite matches
		entryRite := ""
		if len(rp.Template.Stations) > 0 {
			entryRite = rp.Template.Stations[0].Rite
		}
		if currentRite != "" && entryRite == currentRite {
			droDir := filepath.Join(tmpDir, rp.Name)
			if err := os.MkdirAll(droDir, 0o755); err != nil {
				return count, err
			}

			droContent, err := render(projectRoot, "procession-workflow.md.tpl", data)
			if err != nil {
				return count, err
			}
			if err := os.WriteFile(filepath.Join(droDir, "INDEX.dro.md"), droContent, 0o644); err != nil {
				return count, err
			}
			rendered = true
		}

		// Render legomena (skill) — always, regardless of rite
		legoDir := filepath.Join(tmpDir, rp.Name+"-ref")
		if err := os.MkdirAll(legoDir, 0o755); err != nil {
			return count, err
		}

		legoContent, err := render(projectRoot, "procession-ref.md.tpl", data)
		if err != nil {
			return count, err
		}
		if err := os.WriteFile(filepath.Join(legoDir, "INDEX.lego.md"), legoContent, 0o644); err != nil {
			return count, err
		}
		rendered = true

		if rendered {
			count++
		}
	}

	return count, nil
}
