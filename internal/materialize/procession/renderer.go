package procession

import (
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
// For each template, it creates:
//   - {tmpDir}/{name}/INDEX.dro.md  (procession-workflow archetype)
//   - {tmpDir}/{name}-ref/INDEX.lego.md (procession-ref archetype)
//
// Returns the number of templates rendered.
func RenderToDir(projectRoot string, tmpDir string, render RenderFunc) (int, error) {
	resolved, err := ResolveProcessions(projectRoot, nil)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, rp := range resolved {
		data := BuildWorkflowData(rp.Template)

		// Render dromena (command)
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

		// Render legomena (skill)
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

		count++
	}

	return count, nil
}
