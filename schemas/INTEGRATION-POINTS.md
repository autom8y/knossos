# Integration Points for Orchestrator Template System

## Overview

The orchestrator template system (orchestrator-base.md.tpl + orchestrator.yaml + generate-orchestrator.sh) integrates with three key infrastructure systems. This document specifies the integration contracts and assumptions.

## 1. swap-team.sh Integration

### Contract

`swap-team.sh` reads the frontmatter of orchestrator.md to extract agent metadata for the AGENT_MANIFEST. The generated orchestrator.md must produce valid YAML frontmatter that parses correctly.

### Frontmatter Format Requirement

```yaml
---
name: orchestrator
role: "..."
description: "..."
tools: Read, Skill
model: opus
color: <color>
---
```

### Critical Fields

| Field | Required | Type | Validation |
|-------|----------|------|-----------|
| `name` | YES | string | Must equal "orchestrator" (literal) |
| `role` | YES | string | Human-readable one-liner (no pipes or quotes) |
| `description` | YES | string | May be quoted single-line or YAML pipe multiline |
| `tools` | YES | string | Always "Read, Skill" (canonical) |
| `model` | YES | string | Always "opus" (canonical) |
| `color` | YES | string | Valid hex (#RRGGBB) or named color |

### Parsing Assumptions from swap-team.sh

The script extracts metadata using grep/sed patterns:
- Line 1167: `sed -n '/^---$/,/^---$/p' "$agent_file" | grep "^name:" | head -1 | sed 's/^name:[[:space:]]*//'`
- Line 1172-1182: Description extraction handles quoted single-line or YAML pipe multiline
- Line 1199: Role extraction with quote stripping

**Generator Responsibility**: Ensure frontmatter format is parseable by these sed patterns.

### Validation Rule

The generator MUST produce frontmatter that round-trips through swap-team.sh parsing without loss of information.

### Change Restrictions

- Do NOT add new required frontmatter fields without updating swap-team.sh parsing logic
- Do NOT change the name field (must always be "orchestrator")
- Do NOT change tools/model fields (they are canonical across all orchestrators)

---

## 2. workflow.yaml Dependency

### Contract

The orchestrator template system reads workflow.yaml to extract specialist names and complexity levels. These values populate {{SPECIALIST_ENUM}}, {{COMPLEXITY_ENUM}}, and routing table content.

### Data Sources from workflow.yaml

| Template Placeholder | Source | Format | Example |
|---|---|---|---|
| `{{SPECIALIST_ENUM}}` | `phases[].agent` | Pipe-separated quoted strings | `"technology-scout" \| "integration-researcher" \| ...` |
| `{{COMPLEXITY_ENUM}}` | `complexity_levels[].name` | Pipe-separated quoted strings | `"SPIKE" \| "EVALUATION" \| "MOONSHOT"` |
| Routing table specialists | `orchestrator.yaml routing{}.` keys | Must match phase agents | `technology-scout`, `integration-researcher` |

### Validation Requirements

When orchestrator.yaml is processed:

1. **Specialist Validation**: All keys in `orchestrator.yaml routing{}` MUST match agent names from `workflow.yaml phases[].agent`
   - Mismatch = broken routing table in generated orchestrator.md

2. **Complexity Validation**: If orchestrator.md refers to complexity levels, they MUST match `workflow.yaml complexity_levels[].name`
   - Example: If template hardcodes "SPIKE", verify that "SPIKE" exists in workflow.yaml

3. **Phase Ordering**: The `{{SPECIALIST_ENUM}}` should match the phase order in workflow.yaml for readability

### Integration Test Requirement

When generating orchestrator.md for a team:
1. Read both orchestrator.yaml and workflow.yaml
2. Cross-validate specialist names exist in workflow.yaml
3. Fail generation if mismatches found (clear error message)

### Data Consistency Note

**Pre-existing issue discovered during POC**: Original rnd-pack orchestrator.md references complexity "PROTOTYPE" but workflow.yaml defines "EVALUATION". This is data inconsistency, not a generator bug. Generator should validate and report such issues.

---

## 3. CEM Sync & File Validation

### Contract

Generated orchestrator.md files are placed in `.claude/agents/` alongside other agent definitions. They must pass all standard CEM validation checks (checksum, syntax, etc.).

### File Placement

```
/Users/tomtenuta/Code/roster/teams/{team-name}/agents/orchestrator.md
```

### CEM Validation Requirements

Generated files must:
1. Be valid Markdown (no syntax errors)
2. Pass checksum validation if CEM sync is enabled
3. Not break existing agent loading mechanisms
4. Preserve frontmatter format for agent discovery

### No Special CEM Changes Required

The CEM sync system treats generated orchestrator.md like any other agent file. No special handling needed:
- Generated files use standard agent naming (orchestrator.md)
- Frontmatter format is already supported by agent discovery
- No "generated" flag tracking needed in AGENT_MANIFEST (optional enhancement)

### Change Implications

If frontmatter format changes:
- All generated files must be regenerated
- CEM checksums will change (expected)
- swap-team.sh parsing logic must be updated in parallel

---

## 4. AGENT_MANIFEST.json (Optional)

### Current State

AGENT_MANIFEST.json tracks active agents and their origin (team vs. user-added).

### Optional Enhancement

Could add `generated: true` metadata to orchestrator agents to indicate they were auto-generated. Example:

```json
{
  "orchestrator.md": {
    "source": "team",
    "origin": "rnd-pack",
    "generated": true,
    "template_version": "1.0"
  }
}
```

### Benefits
- Prevents accidental manual edits of generated files
- Enables regeneration validation ("did you mean to edit a generated file?")
- Tracks template version for migration purposes

### Decision Point for Phase 2

This is an optional enhancement. Can be added later if needed for:
- Regeneration workflows
- Breaking change migrations
- Audit trails

---

## Integration Testing Matrix

The following teams should be tested in Phase 2 to verify integration:

| Team | Specialists | Phases | Complexity | Notes |
|---|---|---|---|---|
| rnd-pack | 4 | 4 | 3 levels | Linear flow, baseline case |
| security-pack | 4 | 4 | 3 levels | Different naming conventions |
| ecosystem-pack | 5 | 5 | 4 levels | Longest pipeline, critical path |
| doc-team-pack | 4 | 4 | 3 levels | Different domain (documentation) |
| strategy-pack | 4 | 4 | 3 levels | Different complexity enum |

### Integration Test Criteria

For each team, verify:
1. **frontmatter parses cleanly** via swap-team.sh grep/sed
2. **Specialists match workflow.yaml** (no mismatches)
3. **Complexity enums valid** (no undefined levels referenced)
4. **All sections present** (passes validate-orchestrator.sh)
5. **No placeholders remain** (all {{}} replaced)
6. **Markdown is valid** (no syntax errors)
7. **Routing table is sensible** (specialists match routing keys)

---

## Known Constraints & Future Compatibility

### Constraints

1. **Single template for all teams**: orchestrator-base.md.tpl is shared. Teams customize via orchestrator.yaml only.
   - Breaking change: If template structure changes, all team configs must still produce valid output
   - Mitigation: Version the template; maintain backward compatibility or plan migration

2. **ASCII diagram generation**: Current workflow diagram is hardcoded for 4-agent linear layout
   - Breaking change risk: Teams with >6 specialists need different generation logic
   - Mitigation: Phase 2 to implement dynamic diagram generation or diagram override in orchestrator.yaml

3. **Frontmatter format is frozen**: Cannot change field names or structure without updating swap-team.sh
   - Breaking change: Any frontmatter schema change requires coordinated update across both scripts

### Future Extensibility

Design allows for:
- Adding team-specific sections via orchestrator.yaml (extension_points)
- Adding skill references dynamically (already parameterized)
- Adding custom anti-patterns per team (already parameterized)
- Versioning the template schema (not yet implemented)

---

## Decision Log

### swap-team.sh Changes Required?

**Decision**: NONE

**Rationale**: POC verified that current frontmatter parsing in swap-team.sh (lines 1167-1199) works correctly with generated files. Generator produces output that round-trips cleanly.

**Validation**: See PROTO-orchestrator-template-generator.md Integration Points Verified section.

### Template Versioning Strategy

**Decision**: Deferred to Phase 2

**Rationale**: POC uses unversioned template. As templates evolve, need strategy for:
- Tracking which version each team was generated from
- Regenerating when template updates
- Handling backward compatibility

**Recommended Phase 2 approach**: Add `template_version: "1.0"` to orchestrator.yaml schema; update AGENT_MANIFEST to track.

### Complexity Enum Validation

**Decision**: Generator should validate, not enforce

**Rationale**: If orchestrator.yaml references undefined complexity levels, that's a configuration error. Generator should report it clearly and exit 1, not silently use defaults.

**Phase 2 requirement**: Generate-orchestrator.sh must validate all routing specialists exist in workflow.yaml before generating.

---

## Next Steps (Phase 2 Integration Engineer)

When implementing the generator integration:

1. **Verify frontmatter round-trip** through swap-team.sh parsing (use POC test)
2. **Add workflow.yaml validation** to generate-orchestrator.sh
3. **Document diagram generation strategy** (dynamic vs. override)
4. **Implement comprehensive error handling** with clear messages for:
   - Missing orchestrator.yaml
   - Missing workflow.yaml
   - Specialist name mismatches
   - Undefined complexity levels
5. **Create CI integration** for regeneration validation

---

## File References

| File | Role | Owner |
|---|---|---|
| `/Users/tomtenuta/Code/roster/templates/orchestrator-base.md.tpl` | Template with placeholders | Shared (frozen after Phase 1) |
| `/Users/tomtenuta/Code/roster/teams/{team}/orchestrator.yaml` | Team-specific config | Each team |
| `/Users/tomtenuta/Code/roster/templates/generate-orchestrator.sh` | Generator script | Integration Engineer (Phase 2) |
| `/Users/tomtenuta/Code/roster/templates/validate-orchestrator.sh` | Validation script | Integration Engineer (Phase 2) |
| `/Users/tomtenuta/Code/roster/swap-team.sh` | Team swapper (no changes) | Existing (no modifications) |
| `/Users/tomtenuta/Code/roster/teams/{team}/workflow.yaml` | Team workflow definition | Each team |
| `/Users/tomtenuta/Code/roster/teams/{team}/agents/orchestrator.md` | Generated output | Generated (do not edit) |
