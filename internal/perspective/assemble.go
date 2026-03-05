package perspective

import "time"

// Assemble builds a PerspectiveDocument by running all MVP layer resolvers
// against the shared parse context.
func Assemble(ctx *ParseContext, opts PerspectiveOptions, start time.Time) *PerspectiveDocument {
	doc := &PerspectiveDocument{
		Version:     "1.0",
		GeneratedAt: time.Now().UTC(),
		Agent:       opts.AgentName,
		Rite:        ctx.RiteName,
		SourcePath:  ctx.AgentSourcePath,
		Mode:        opts.Mode,
		Layers:      make(map[string]*LayerEnvelope),
	}

	// Resolve MVP layers: L1, L3, L4, L5, L9
	doc.Layers["L1"] = resolveIdentity(ctx)
	doc.Layers["L3"] = resolveCapability(ctx)
	doc.Layers["L4"] = resolveConstraint(ctx)
	doc.Layers["L5"] = resolveMemory(ctx)
	doc.Layers["L9"] = resolveProvenance(ctx)

	// Compute assembly metadata
	var resolved, degraded, failed int
	for _, env := range doc.Layers {
		switch env.Status {
		case StatusResolved:
			resolved++
		case StatusPartial, StatusOpaque:
			degraded++
		case StatusFailed:
			failed++
		}
	}

	doc.AssemblyMetadata = AssemblyMetadata{
		ResolutionTimeMs: int(time.Since(start).Milliseconds()),
		LayersResolved:   resolved,
		LayersDegraded:   degraded,
		LayersFailed:     failed,
	}

	return doc
}
