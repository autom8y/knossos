package hook

import (
	"github.com/autom8y/knossos/internal/cmd/common"
	"github.com/autom8y/knossos/internal/materialize"
	"github.com/autom8y/knossos/internal/paths"
)

// NewWiredMaterializer creates a Materializer with all 4 embedded FS sources wired.
// This is the standard pattern for hook commands that need to run sync operations.
func NewWiredMaterializer(resolver *paths.Resolver) *materialize.Materializer {
	m := materialize.NewMaterializer(resolver)

	if embRites := common.EmbeddedRites(); embRites != nil {
		m.WithEmbeddedFS(embRites)
	}
	if embTemplates := common.EmbeddedTemplates(); embTemplates != nil {
		m.WithEmbeddedTemplates(embTemplates)
	}
	if embAgents := common.EmbeddedAgents(); embAgents != nil {
		m.WithEmbeddedAgents(embAgents)
	}
	if embMena := common.EmbeddedMena(); embMena != nil {
		m.WithEmbeddedMena(embMena)
	}
	if embProc := common.EmbeddedProcessions(); embProc != nil {
		m.WithEmbeddedProcessions(embProc)
	}

	return m
}
