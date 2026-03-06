package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
	"github.com/autom8y/knossos/internal/perspective"
)

type embodyOptions struct {
	riteName string
	audit    bool
	simulate bool
}

func newEmbodyCmd(ctx *cmdContext) *cobra.Command {
	var opts embodyOptions

	cmd := &cobra.Command{
		Use:   "embody <agent-name>",
		Short: "Show an agent's full experiential context",
		Long: `Reconstructs an agent's full context as a first-person perspective view.

Resolves identity, perception, capability, constraint, memory, position,
surface, and provenance layers from source files (not materialized output)
to capture all metadata including knossos-only fields stripped during
materialization.

Examples:
  ari agent embody pythia                          # Default perspective
  ari agent embody principal-engineer --rite 10x-dev  # Specific rite
  ari agent embody qa-adversary --audit            # With audit overlay
  ari agent embody pythia -o json                  # JSON output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEmbody(ctx, opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.riteName, "rite", "r", "", "Rite to resolve agent from (default: active rite)")
	cmd.Flags().BoolVar(&opts.audit, "audit", false, "Enable audit overlay with consistency checks")
	cmd.Flags().BoolVar(&opts.simulate, "simulate", false, "Enable simulate mode (capability mapping — Phase 3)")

	return cmd
}

func runEmbody(ctx *cmdContext, opts embodyOptions, agentName string) error {
	printer := ctx.GetPrinter(output.FormatText)
	resolver := ctx.GetResolver()

	mode := "default"
	if opts.audit {
		mode = "audit"
	}
	if opts.simulate {
		mode = "simulate"
	}

	perspOpts := perspective.PerspectiveOptions{
		AgentName:   agentName,
		RiteName:    opts.riteName,
		Mode:        mode,
		ProjectRoot: resolver.ProjectRoot(),
	}

	start := time.Now()

	// Build parse context
	parseCtx, err := perspective.NewParseContext(perspOpts)
	if err != nil {
		err := errors.Wrap(errors.CodeGeneralError, "failed to build perspective context", err)
		printer.PrintError(err)
		return err
	}

	// Assemble perspective document
	doc := perspective.Assemble(parseCtx, perspOpts, start)

	// Run audit if requested
	if opts.audit {
		doc.AuditOverlay = perspective.RunAudit(doc, parseCtx)
	}

	// Simulate mode skeleton (Phase 3 — not yet implemented)
	if opts.simulate {
		doc.SimulateOverlay = &perspective.SimulateOverlay{
			Prompt: "(simulate mode not yet implemented — Phase 3)",
		}
	}

	// Output
	if err := printer.Print(embodyOutput{doc: doc}); err != nil {
		return err
	}

	return nil
}

// embodyOutput wraps PerspectiveDocument for output formatting.
type embodyOutput struct {
	doc *perspective.PerspectiveDocument
}

// Text implements output.Textable for human-readable text output.
func (o embodyOutput) Text() string {
	doc := o.doc
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\n=== Perspective: %s (%s) ===\n", doc.Agent, doc.Rite))

	// L1 Identity
	if env, ok := doc.Layers["L1"]; ok {
		b.WriteString("\n> Identity\n")
		if id, ok := env.Data.(*perspective.IdentityData); ok {
			b.WriteString(fmt.Sprintf("  Name: %s\n", id.Name))
			if id.Role != "" {
				b.WriteString(fmt.Sprintf("  Role: %s\n", id.Role))
			}
			if id.Type != "" {
				b.WriteString(fmt.Sprintf("  Type: %s\n", id.Type))
			}
			if id.Model != "" {
				b.WriteString(fmt.Sprintf("  Model: %s\n", id.Model))
			}
			b.WriteString(fmt.Sprintf("  System Prompt: %d lines\n", id.SystemPromptLines))
		}
		writeStatusLine(&b, env)
	}

	// L2 Perception
	if env, ok := doc.Layers["L2"]; ok {
		b.WriteString("\n> Perception\n")
		if perc, ok := env.Data.(*perspective.PerceptionData); ok {
			b.WriteString(fmt.Sprintf("  Explicit Skills: %d\n", len(perc.ExplicitSkills)))
			if len(perc.ExplicitSkills) > 0 {
				b.WriteString(fmt.Sprintf("    %s\n", truncateList(perc.ExplicitSkills, 5)))
			}
			b.WriteString(fmt.Sprintf("  Policy Injected: %d\n", len(perc.PolicyInjectedSkills)))
			b.WriteString(fmt.Sprintf("  Policy Referenced: %d\n", len(perc.PolicyReferencedSkills)))
			b.WriteString(fmt.Sprintf("  On-Demand: %d\n", len(perc.OnDemandSkills)))
			skillTool := "available"
			if !perc.SkillToolAvailable {
				skillTool = "DISALLOWED"
			}
			b.WriteString(fmt.Sprintf("  Skill Tool: %s\n", skillTool))
			b.WriteString(fmt.Sprintf("  Totals: %d preloaded, %d reachable\n", perc.TotalPreloaded, perc.TotalReachable))
		}
		writeStatusLine(&b, env)
	}

	// L3 Capability
	if env, ok := doc.Layers["L3"]; ok {
		b.WriteString("\n> Capability\n")
		if cap, ok := env.Data.(*perspective.CapabilityData); ok {
			if len(cap.Tools) > 0 {
				toolSummary := truncateList(cap.CCNativeTools, 5)
				b.WriteString(fmt.Sprintf("  Tools: %s (%d total)\n", toolSummary, len(cap.Tools)))
			} else {
				b.WriteString("  Tools: none\n")
			}
			if len(cap.MCPTools) > 0 {
				var names []string
				for _, m := range cap.MCPTools {
					wired := "wired"
					if !m.ServerWired {
						wired = "NOT wired"
					}
					names = append(names, fmt.Sprintf("%s (%s)", m.Reference, wired))
				}
				b.WriteString(fmt.Sprintf("  MCP: %s\n", strings.Join(names, ", ")))
			} else {
				b.WriteString("  MCP: none\n")
			}
			if cap.ToolsFromDefaults {
				b.WriteString("  Tools Source: manifest agent_defaults\n")
			}
			if len(cap.Hooks) > 0 {
				b.WriteString(fmt.Sprintf("  Hooks: %d declared\n", len(cap.Hooks)))
			}
		}
		writeStatusLine(&b, env)
	}

	// L4 Constraint
	if env, ok := doc.Layers["L4"]; ok {
		b.WriteString("\n> Constraint\n")
		if con, ok := env.Data.(*perspective.ConstraintData); ok {
			if len(con.DisallowedTools) > 0 {
				b.WriteString(fmt.Sprintf("  Disallowed: %s (%d tools)\n",
					truncateList(con.DisallowedTools, 5), len(con.DisallowedTools)))
			} else {
				b.WriteString("  Disallowed: none\n")
			}
			if con.WriteGuard != nil && con.WriteGuard.Enabled {
				b.WriteString(fmt.Sprintf("  Write Guard: enabled (%d paths)\n", len(con.WriteGuard.AllowPaths)))
			} else {
				b.WriteString("  Write Guard: disabled\n")
			}
			if con.BehavioralContract != nil {
				var parts []string
				if len(con.BehavioralContract.MustNot) > 0 {
					parts = append(parts, fmt.Sprintf("%d must_not", len(con.BehavioralContract.MustNot)))
				}
				if len(con.BehavioralContract.MustUse) > 0 {
					parts = append(parts, fmt.Sprintf("%d must_use", len(con.BehavioralContract.MustUse)))
				}
				if len(con.BehavioralContract.MustProduce) > 0 {
					parts = append(parts, fmt.Sprintf("%d must_produce", len(con.BehavioralContract.MustProduce)))
				}
				b.WriteString(fmt.Sprintf("  Contract: %s\n", strings.Join(parts, ", ")))
			} else {
				b.WriteString("  Contract: none\n")
			}
		}
		writeStatusLine(&b, env)
	}

	// L5 Memory
	if env, ok := doc.Layers["L5"]; ok {
		b.WriteString("\n> Memory\n")
		if mem, ok := env.Data.(*perspective.MemoryData); ok {
			if mem.Enabled {
				b.WriteString(fmt.Sprintf("  Scope: %s\n", mem.Scope))
			} else {
				b.WriteString("  Scope: disabled\n")
			}
			if mem.SeedFile != nil {
				if mem.SeedFile.Exists {
					lines := 0
					if mem.SeedFile.LineCount != nil {
						lines = *mem.SeedFile.LineCount
					}
					b.WriteString(fmt.Sprintf("  Seed: exists (%d lines)\n", lines))
				} else {
					b.WriteString("  Seed: not found\n")
				}
			}
			if mem.RuntimeMemory != nil {
				if mem.RuntimeMemory.PathResolvable {
					accessible := "accessible"
					if !mem.RuntimeMemory.ContentAccessible {
						accessible = "not found"
					}
					b.WriteString(fmt.Sprintf("  Runtime: %s\n", accessible))
				} else {
					b.WriteString("  Runtime: OPAQUE (CC path hash)\n")
				}
			}
		}
		writeStatusLine(&b, env)
	}

	// L6 Position
	if env, ok := doc.Layers["L6"]; ok {
		b.WriteString("\n> Position\n")
		if pos, ok := env.Data.(*perspective.PositionData); ok {
			if pos.InWorkflow {
				b.WriteString(fmt.Sprintf("  Phase: %s (%d/%d)\n", pos.WorkflowPhase, pos.PhaseIndex+1, pos.TotalPhases))
				if pos.PhasePredecessor != "" {
					b.WriteString(fmt.Sprintf("  Predecessor: %s\n", pos.PhasePredecessor))
				}
				if pos.PhaseSuccessor != "" {
					b.WriteString(fmt.Sprintf("  Successor: %s\n", pos.PhaseSuccessor))
				}
				if pos.PhaseCondition != "" {
					b.WriteString(fmt.Sprintf("  Condition: %s\n", pos.PhaseCondition))
				}
				if pos.PhaseProduces != "" {
					b.WriteString(fmt.Sprintf("  Produces: %s\n", pos.PhaseProduces))
				}
			} else {
				b.WriteString("  Phase: not in workflow\n")
			}
			if pos.IsEntryPoint {
				b.WriteString("  Entry Point: yes (workflow)\n")
			}
			if pos.IsEntryAgent {
				b.WriteString("  Entry Agent: yes (manifest)\n")
			}
			if len(pos.BackRoutes) > 0 {
				b.WriteString(fmt.Sprintf("  Back Routes: %d targeting this agent\n", len(pos.BackRoutes)))
			}
			if len(pos.ComplexityGates) > 0 {
				b.WriteString(fmt.Sprintf("  Complexity Gates: %s\n", strings.Join(pos.ComplexityGates, ", ")))
			}
			if len(pos.HandoffCriteria) > 0 {
				b.WriteString(fmt.Sprintf("  Handoff Criteria: %d items\n", len(pos.HandoffCriteria)))
			}
		}
		writeStatusLine(&b, env)
	}

	// L7 Surface
	if env, ok := doc.Layers["L7"]; ok {
		b.WriteString("\n> Surface\n")
		if surf, ok := env.Data.(*perspective.SurfaceData); ok {
			if len(surf.DromenaOwned) > 0 {
				b.WriteString(fmt.Sprintf("  Dromena: %s\n", strings.Join(surf.DromenaOwned, ", ")))
			} else {
				b.WriteString("  Dromena: none\n")
			}
			if len(surf.LegomenaAvailable) > 0 {
				b.WriteString(fmt.Sprintf("  Legomena: %s\n", truncateList(surf.LegomenaAvailable, 5)))
			} else {
				b.WriteString("  Legomena: none\n")
			}
			if len(surf.ArtifactTypes) > 0 {
				b.WriteString(fmt.Sprintf("  Artifacts: %s\n", strings.Join(surf.ArtifactTypes, ", ")))
			}
			if len(surf.Commands) > 0 {
				var cmdNames []string
				for _, c := range surf.Commands {
					cmdNames = append(cmdNames, c.Name)
				}
				b.WriteString(fmt.Sprintf("  Commands: %s\n", strings.Join(cmdNames, ", ")))
			}
			if len(surf.ContractMustProduce) > 0 {
				b.WriteString(fmt.Sprintf("  Must Produce: %s\n", strings.Join(surf.ContractMustProduce, ", ")))
			}
		}
		writeStatusLine(&b, env)
	}

	// L9 Provenance
	if env, ok := doc.Layers["L9"]; ok {
		b.WriteString("\n> Provenance\n")
		if prov, ok := env.Data.(*perspective.ProvenanceData); ok {
			if prov.Owner != "" {
				b.WriteString(fmt.Sprintf("  Owner: %s\n", prov.Owner))
				diverged := "no"
				if prov.Diverged {
					diverged = "YES"
				}
				b.WriteString(fmt.Sprintf("  Diverged: %s\n", diverged))
				if !prov.LastSynced.IsZero() {
					b.WriteString(fmt.Sprintf("  Last Sync: %s\n", prov.LastSynced.Format(time.RFC3339)))
				}
			} else {
				b.WriteString("  No provenance entry found\n")
			}
		}
		writeStatusLine(&b, env)
	}

	// Assembly metadata
	b.WriteString(fmt.Sprintf("\nAssembly: %d resolved, %d degraded, %d failed (%dms)\n",
		doc.AssemblyMetadata.LayersResolved,
		doc.AssemblyMetadata.LayersDegraded,
		doc.AssemblyMetadata.LayersFailed,
		doc.AssemblyMetadata.ResolutionTimeMs))

	// Audit overlay
	if doc.AuditOverlay != nil {
		b.WriteString("\n=== Audit ===\n")
		if len(doc.AuditOverlay.Findings) == 0 {
			b.WriteString("No findings.\n")
		}
		for _, f := range doc.AuditOverlay.Findings {
			prefix := "  "
			switch f.Severity {
			case perspective.SeverityCritical:
				prefix = "! "
			case perspective.SeverityWarning:
				prefix = "? "
			case perspective.SeverityInfo:
				prefix = "  "
			}
			b.WriteString(fmt.Sprintf("%s%s [%s] %s\n", prefix, f.Severity, f.ID, f.Title))
			b.WriteString(fmt.Sprintf("  Layers: %s | %s\n", strings.Join(f.LayersAffected, ", "), f.Description))
			if f.Evidence != "" {
				b.WriteString(fmt.Sprintf("  Evidence: %s\n", f.Evidence))
			}
		}
		s := doc.AuditOverlay.SeveritySummary
		b.WriteString(fmt.Sprintf("\nSummary: %d CRITICAL, %d WARNING, %d INFO\n",
			s.Critical, s.Warning, s.Info))
	}

	return b.String()
}

// MarshalJSON delegates to the underlying document for JSON output.
func (o embodyOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.doc)
}

// MarshalYAML delegates to the underlying document for YAML output.
func (o embodyOutput) MarshalYAML() (any, error) {
	return o.doc, nil
}

// writeStatusLine writes the layer status and any gaps to the builder.
func writeStatusLine(b *strings.Builder, env *perspective.LayerEnvelope) {
	if env.Status != perspective.StatusResolved {
		fmt.Fprintf(b, "  [%s]", env.Status)
		for _, g := range env.Gaps {
			fmt.Fprintf(b, " %s: %s", g.Field, g.Reason)
		}
		b.WriteString("\n")
	}
}

// truncateList returns the first n items joined by comma, with ellipsis if truncated.
func truncateList(items []string, n int) string {
	if len(items) <= n {
		return strings.Join(items, ", ")
	}
	return strings.Join(items[:n], ", ") + ", ..."
}
