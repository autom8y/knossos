package perspective

import "time"

// Assemble builds a PerspectiveDocument by running all layer resolvers
// against the shared parse context in topological dependency order.
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

	// Step 1: Independent layers (no cross-layer dependencies)
	doc.Layers["L1"] = resolveIdentity(ctx)
	doc.Layers["L3"] = resolveCapability(ctx)
	doc.Layers["L4"] = resolveConstraint(ctx)
	doc.Layers["L5"] = resolveMemory(ctx)
	doc.Layers["L6"] = resolvePosition(ctx)
	doc.Layers["L7"] = resolveSurface(ctx)
	doc.Layers["L9"] = resolveProvenance(ctx)

	// Step 2: L2 depends on L3 (tools) and L4 (disallowedTools)
	capData := getLayerData[*CapabilityData](doc, "L3")
	conData := getLayerData[*ConstraintData](doc, "L4")
	if capData != nil && conData != nil {
		doc.Layers["L2"] = resolvePerception(ctx, capData, conData)
	} else {
		doc.Layers["L2"] = &LayerEnvelope{
			Status:           StatusFailed,
			ResolutionMethod: "L2 depends on L3 and L4 which failed to resolve",
			Data:             &PerceptionData{},
		}
	}

	// Step 3: L8 depends on all other layers (inverse computation)
	doc.Layers["L8"] = resolveHorizon(ctx, doc)

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
