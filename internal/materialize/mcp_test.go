package materialize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRiteManifest_MCPServerNames(t *testing.T) {
	manifest := &RiteManifest{
		MCPServers: []MCPServer{
			{Name: "github"},
			{Name: "terraform"},
			{Name: "go-semantic"},
		},
	}

	names := manifest.MCPServerNames()
	assert.Equal(t, []string{"github", "terraform", "go-semantic"}, names)
}

func TestRiteManifest_MCPServerNames_Empty(t *testing.T) {
	manifest := &RiteManifest{}

	names := manifest.MCPServerNames()
	assert.NotNil(t, names)
	assert.Empty(t, names)
}
