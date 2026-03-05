package perspective

import (
	"fmt"
	"strings"
)

// RunAudit evaluates audit checks against a resolved perspective document.
// It returns an AuditOverlay with findings filtered to MVP layers (L1, L3, L4, L5, L9).
func RunAudit(doc *PerspectiveDocument, ctx *ParseContext) *AuditOverlay {
	var findings []AuditFinding

	findings = append(findings, checkMissingContract(doc)...)
	findings = append(findings, checkToolConflict(doc)...)
	findings = append(findings, checkMemoryNoSeed(doc)...)
	findings = append(findings, checkMCPNotWired(doc)...)
	findings = append(findings, checkModelInherit(doc, ctx)...)
	findings = append(findings, checkWriteGuardNoExtraPaths(doc)...)

	summary := SeveritySummary{}
	for _, f := range findings {
		switch f.Severity {
		case SeverityCritical:
			summary.Critical++
		case SeverityWarning:
			summary.Warning++
		case SeverityInfo:
			summary.Info++
		}
	}

	return &AuditOverlay{
		Findings:        findings,
		SeveritySummary: summary,
	}
}

// checkMissingContract checks if the agent has no behavioral contract (L4).
func checkMissingContract(doc *PerspectiveDocument) []AuditFinding {
	l4 := getLayerData[*ConstraintData](doc, "L4")
	if l4 == nil {
		return nil
	}

	if l4.BehavioralContract == nil {
		return []AuditFinding{{
			ID:             "AUDIT-001",
			Severity:       SeverityWarning,
			Category:       CategoryGap,
			LayersAffected: []string{"L4"},
			Title:          "Missing behavioral contract",
			Description:    "No contract.must_not constraints defined in agent source frontmatter",
			Recommendation: "Add a contract: section with must_not rules to define behavioral boundaries",
		}}
	}
	return nil
}

// checkToolConflict checks if any tool appears in both tools and disallowedTools (L3+L4).
func checkToolConflict(doc *PerspectiveDocument) []AuditFinding {
	l3 := getLayerData[*CapabilityData](doc, "L3")
	l4 := getLayerData[*ConstraintData](doc, "L4")
	if l3 == nil || l4 == nil {
		return nil
	}

	disallowed := make(map[string]bool, len(l4.DisallowedTools))
	for _, t := range l4.DisallowedTools {
		disallowed[t] = true
	}

	var conflicts []string
	for _, t := range l3.Tools {
		if disallowed[t] {
			conflicts = append(conflicts, t)
		}
	}

	if len(conflicts) > 0 {
		return []AuditFinding{{
			ID:             "AUDIT-002",
			Severity:       SeverityCritical,
			Category:       CategoryInconsistency,
			LayersAffected: []string{"L3", "L4"},
			Title:          "Tool in both tools and disallowedTools",
			Description:    "The following tools appear in both the tools list and disallowedTools list",
			Evidence:       strings.Join(conflicts, ", "),
			Recommendation: "Remove conflicting tools from either tools or disallowedTools",
		}}
	}
	return nil
}

// checkMemoryNoSeed checks if memory is enabled but no seed file exists (L5).
func checkMemoryNoSeed(doc *PerspectiveDocument) []AuditFinding {
	l5 := getLayerData[*MemoryData](doc, "L5")
	if l5 == nil {
		return nil
	}

	if l5.Enabled && l5.SeedFile != nil && !l5.SeedFile.Exists {
		return []AuditFinding{{
			ID:             "AUDIT-003",
			Severity:       SeverityWarning,
			Category:       CategoryDegradation,
			LayersAffected: []string{"L5"},
			Title:          "Memory enabled but no seed file",
			Description:    fmt.Sprintf("Memory scope is %q but seed file does not exist at %s", l5.Scope, l5.SeedFile.Path),
			Recommendation: "Create a seed MEMORY.md at the expected path for first-invocation resilience",
		}}
	}
	return nil
}

// checkMCPNotWired checks if any MCP tool reference lacks a wired server (L3).
func checkMCPNotWired(doc *PerspectiveDocument) []AuditFinding {
	l3 := getLayerData[*CapabilityData](doc, "L3")
	if l3 == nil {
		return nil
	}

	var unwired []string
	for _, mcp := range l3.MCPTools {
		if !mcp.ServerWired {
			unwired = append(unwired, mcp.Reference)
		}
	}

	if len(unwired) > 0 {
		return []AuditFinding{{
			ID:             "AUDIT-004",
			Severity:       SeverityCritical,
			Category:       CategoryGap,
			LayersAffected: []string{"L3"},
			Title:          "MCP tool without wired server",
			Description:    "The following MCP tool references have no matching server in the rite manifest mcp_servers",
			Evidence:       strings.Join(unwired, ", "),
			Recommendation: "Add the missing MCP server to the rite manifest mcp_servers section",
		}}
	}
	return nil
}

// checkModelInherit checks if model is "inherit" without a manifest default (L1).
func checkModelInherit(doc *PerspectiveDocument, ctx *ParseContext) []AuditFinding {
	l1 := getLayerData[*IdentityData](doc, "L1")
	if l1 == nil {
		return nil
	}

	if l1.Model != "inherit" {
		return nil
	}

	defaults := getAgentDefaults(ctx.RiteManifest)
	if _, ok := defaults["model"]; !ok {
		return []AuditFinding{{
			ID:             "AUDIT-005",
			Severity:       SeverityWarning,
			Category:       CategoryGap,
			LayersAffected: []string{"L1"},
			Title:          "model: inherit without manifest default",
			Description:    "Agent uses model: inherit but the rite manifest agent_defaults has no model field",
			Recommendation: "Add a model field to agent_defaults in the rite manifest, or set model explicitly on the agent",
		}}
	}
	return nil
}

// checkWriteGuardNoExtraPaths checks if write-guard is enabled but has no agent-specific extra paths (L4).
func checkWriteGuardNoExtraPaths(doc *PerspectiveDocument) []AuditFinding {
	l4 := getLayerData[*ConstraintData](doc, "L4")
	if l4 == nil {
		return nil
	}

	if l4.WriteGuard != nil && l4.WriteGuard.Enabled && len(l4.WriteGuard.AgentExtraPaths) == 0 {
		return []AuditFinding{{
			ID:             "AUDIT-006",
			Severity:       SeverityInfo,
			Category:       CategoryGap,
			LayersAffected: []string{"L4"},
			Title:          "Write-guard enabled but no agent extra paths",
			Description:    "Agent relies only on shared and rite base paths for write-guard enforcement",
			Recommendation: "Consider adding agent-specific write-guard extra-paths if the agent needs to write to additional locations",
		}}
	}
	return nil
}

// getLayerData extracts typed layer data from a perspective document.
// Returns nil if the layer is missing or the data type doesn't match.
func getLayerData[T any](doc *PerspectiveDocument, layerKey string) T {
	var zero T
	env, ok := doc.Layers[layerKey]
	if !ok || env == nil {
		return zero
	}
	data, ok := env.Data.(T)
	if !ok {
		return zero
	}
	return data
}
