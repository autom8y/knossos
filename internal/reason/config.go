package reason

import (
	reasoncontext "github.com/autom8y/knossos/internal/reason/context"
	"github.com/autom8y/knossos/internal/reason/response"
)

// ReasoningConfig is the top-level configuration for the reasoning pipeline.
type ReasoningConfig struct {
	// Generator controls Claude API interaction.
	Generator response.GeneratorConfig

	// Assembler controls context window assembly.
	Assembler reasoncontext.AssemblerConfig

	// SearchLimit is the maximum number of search results to retrieve.
	// Default: 20.
	SearchLimit int

	// Org is the organization name for system prompt rendering.
	// Default: "autom8y".
	Org string
}

// DefaultReasoningConfig returns production defaults.
func DefaultReasoningConfig() ReasoningConfig {
	return ReasoningConfig{
		Generator:   response.DefaultGeneratorConfig(),
		Assembler:   reasoncontext.DefaultAssemblerConfig(),
		SearchLimit: 20,
		Org:         "autom8y",
	}
}
