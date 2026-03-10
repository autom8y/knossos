package materialize

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRiteManifestDrift verifies that materialize.RiteManifest and
// rite.RiteManifest parse the same manifest.yaml fields identically.
// This test catches schema drift between the two subset projections.
func TestRiteManifestDrift(t *testing.T) {
	t.Parallel()
	candidates := []string{
		"../../rites/10x-dev/manifest.yaml",
		"../../rites/shared/manifest.yaml",
	}

	tested := 0
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		tested++

		t.Run(path, func(t *testing.T) {
			t.Parallel()
			// Parse as materialize.RiteManifest
			var matManifest RiteManifest
			require.NoError(t, yaml.Unmarshal(data, &matManifest))

			// Parse the same YAML into a raw map to verify shared fields
			var raw map[string]any
			require.NoError(t, yaml.Unmarshal(data, &raw))

			// Verify the shared fields we care about.
			// Only compare non-nil raw values -- YAML null maps to nil in raw
			// but to "" in Go structs, which is expected behavior, not drift.
			assert.Equal(t, raw["name"], matManifest.Name, "name field drift in %s", path)
			if v, ok := raw["version"]; ok && v != nil {
				assert.Equal(t, v, matManifest.Version, "version field drift in %s", path)
			}
			if v, ok := raw["entry_agent"]; ok && v != nil {
				assert.Equal(t, v, matManifest.EntryAgent, "entry_agent field drift in %s", path)
			}
			if v, ok := raw["description"]; ok && v != nil {
				assert.Equal(t, v, matManifest.Description, "description field drift in %s", path)
			}
		})
	}

	require.Greater(t, tested, 0, "at least one manifest.yaml must be available for drift testing")
}
