package perspective

import (
	"fmt"
	"strings"
)

// RunAudit evaluates audit checks against a resolved perspective document.
// It returns an AuditOverlay with findings filtered to MVP layers (L1, L3, L4, L5, L9).
func RunAudit(doc *PerspectiveDocument, ctx *ParseContext) *AuditOverlay {
	var findings []AuditFinding

	// Phase 1 checks (AUDIT-001 through AUDIT-006)
	findings = append(findings, checkMissingContract(doc)...)
	findings = append(findings, checkToolConflict(doc)...)
	findings = append(findings, checkMemoryNoSeed(doc)...)
	findings = append(findings, checkMCPNotWired(doc)...)
	findings = append(findings, checkModelInherit(doc, ctx)...)
	findings = append(findings, checkWriteGuardNoExtraPaths(doc)...)

	// Phase 2 checks (AUDIT-007 through AUDIT-010)
	findings = append(findings, checkSkillsWithoutSkillTool(doc)...)
	findings = append(findings, checkOrphanAgent(doc)...)
	findings = append(findings, checkMustProduceNotInArtifacts(doc)...)
	findings = append(findings, checkUpstreamDownstreamNotInRite(doc, ctx)...)

	// Phase 3 checks (AUDIT-011)
	findings = append(findings, checkZeroReachableSkills(doc)...)

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

// --- Phase 2 audit checks ---

// checkSkillsWithoutSkillTool checks if skills are preloaded but Skill tool is disallowed (L2+L4).
func checkSkillsWithoutSkillTool(doc *PerspectiveDocument) []AuditFinding {
	l2 := getLayerData[*PerceptionData](doc, "L2")
	if l2 == nil {
		return nil
	}

	hasPreloaded := len(l2.ExplicitSkills) > 0 || len(l2.PolicyInjectedSkills) > 0
	if hasPreloaded && !l2.SkillToolAvailable {
		var skillNames []string
		skillNames = append(skillNames, l2.ExplicitSkills...)
		skillNames = append(skillNames, l2.PolicyInjectedSkills...)
		return []AuditFinding{{
			ID:             "AUDIT-007",
			Severity:       SeverityCritical,
			Category:       CategoryInconsistency,
			LayersAffected: []string{"L2", "L4"},
			Title:          "Skills preloaded but Skill tool disallowed",
			Description:    "Agent has preloaded skills but Skill is in disallowedTools, making on-demand skill invocation impossible",
			Evidence:       strings.Join(skillNames, ", "),
			Recommendation: "Remove Skill from disallowedTools or remove explicit skills from agent frontmatter",
		}}
	}
	return nil
}

// checkOrphanAgent checks if the agent is not in any workflow phase (L6).
func checkOrphanAgent(doc *PerspectiveDocument) []AuditFinding {
	l6 := getLayerData[*PositionData](doc, "L6")
	if l6 == nil {
		return nil
	}

	if !l6.InWorkflow {
		return []AuditFinding{{
			ID:             "AUDIT-008",
			Severity:       SeverityWarning,
			Category:       CategoryGap,
			LayersAffected: []string{"L6"},
			Title:          "Agent not in any workflow phase",
			Description:    "Agent does not appear in any workflow.yaml phase. This may be intentional (e.g., orchestrator agents)",
			Recommendation: "Verify this agent is intentionally out-of-workflow, or add it to a workflow phase",
		}}
	}
	return nil
}

// checkMustProduceNotInArtifacts checks if contract.must_produce items are not in artifact types (L4+L7).
func checkMustProduceNotInArtifacts(doc *PerspectiveDocument) []AuditFinding {
	l4 := getLayerData[*ConstraintData](doc, "L4")
	l7 := getLayerData[*SurfaceData](doc, "L7")
	if l4 == nil || l7 == nil || l4.BehavioralContract == nil {
		return nil
	}

	artifactSet := make(map[string]bool, len(l7.ArtifactTypes))
	for _, a := range l7.ArtifactTypes {
		artifactSet[a] = true
	}

	var missing []string
	for _, mp := range l4.BehavioralContract.MustProduce {
		if !artifactSet[mp] {
			missing = append(missing, mp)
		}
	}

	if len(missing) > 0 {
		return []AuditFinding{{
			ID:             "AUDIT-009",
			Severity:       SeverityWarning,
			Category:       CategoryInconsistency,
			LayersAffected: []string{"L4", "L7"},
			Title:          "contract.must_produce not in workflow artifact types",
			Description:    "Behavioral contract requires producing artifacts not declared in workflow phase produces",
			Evidence:       strings.Join(missing, ", "),
			Recommendation: "Add the missing artifact types to the workflow phase produces field, or update the contract",
		}}
	}
	return nil
}

// checkUpstreamDownstreamNotInRite checks if upstream/downstream references agents not in the rite (L6).
func checkUpstreamDownstreamNotInRite(doc *PerspectiveDocument, ctx *ParseContext) []AuditFinding {
	l6 := getLayerData[*PositionData](doc, "L6")
	if l6 == nil {
		return nil
	}

	// Build set of valid agent names from rite manifest
	riteAgents := make(map[string]bool)
	if agents, ok := ctx.RiteManifest["agents"].([]any); ok {
		for _, a := range agents {
			if m, ok := a.(map[string]any); ok {
				if name, ok := m["name"].(string); ok {
					riteAgents[name] = true
				}
			}
		}
	}

	// Check upstream/downstream from raw frontmatter (knossos-only fields)
	var invalid []string
	for _, field := range []string{"upstream", "downstream"} {
		edges := stringSliceFromMap(ctx.AgentFrontmatterRaw, field)
		for _, agentRef := range edges {
			if !riteAgents[agentRef] {
				invalid = append(invalid, field+":"+agentRef)
			}
		}
	}

	if len(invalid) > 0 {
		return []AuditFinding{{
			ID:             "AUDIT-010",
			Severity:       SeverityCritical,
			Category:       CategoryInconsistency,
			LayersAffected: []string{"L6"},
			Title:          "upstream/downstream references agent not in rite",
			Description:    "Agent frontmatter references agents not found in the rite manifest agents list",
			Evidence:       strings.Join(invalid, ", "),
			Recommendation: "Fix the agent references or add the missing agents to the rite manifest",
		}}
	}
	return nil
}

// --- Phase 3 audit checks ---

// checkZeroReachableSkills checks if the agent has zero reachable skills despite Skill tool being available (L2).
func checkZeroReachableSkills(doc *PerspectiveDocument) []AuditFinding {
	l2 := getLayerData[*PerceptionData](doc, "L2")
	if l2 == nil {
		return nil
	}

	if l2.SkillToolAvailable && l2.TotalReachable == 0 {
		return []AuditFinding{{
			ID:             "AUDIT-011",
			Severity:       SeverityWarning,
			Category:       CategoryGap,
			LayersAffected: []string{"L2"},
			Title:          "Zero reachable skills with Skill tool available",
			Description:    "Agent has Skill tool available but cannot reach any skills (no explicit, injected, referenced, or on-demand skills)",
			Recommendation: "Add skills to agent frontmatter, configure skill policies, or remove Skill from tools if not needed",
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
