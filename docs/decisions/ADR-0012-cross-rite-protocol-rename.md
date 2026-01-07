# ADR-0012: Rename cross_team_protocol to cross_rite_protocol

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-01-07 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

The Knossos terminology migration from "team" to "rite" has been ongoing, with most terminology successfully migrated. However, the `cross_team_protocol` field in the orchestrator.yaml schema remains, creating inconsistency with the rest of the platform which now uses "rite" terminology exclusively.

### Current State

**Orchestrator Schema**: `schemas/orchestrator.yaml.schema.json`

The schema currently defines:
```json
"cross_team_protocol": {
  "type": "string",
  "description": "OPTIONAL: Cross-team handoff protocol..."
}
```

This field is referenced in:
- orchestrator.yaml.schema.json (lines 181-188, 255-258, 317, 320, 374, 423)
- All rite orchestrator.yaml files that define cross-rite coordination protocols

### Problem

The `cross_team_protocol` field name is the last remaining "team" terminology in active schemas and rite definitions. This creates:

1. **Semantic Confusion**: The description says "Cross-team" but the platform now uses "rite" terminology everywhere else
2. **Inconsistency**: All other schema fields use `rite` (e.g., `active_rite`, `rite_name`, etc.)
3. **Migration Incompleteness**: The terminology migration appears unfinished
4. **Documentation Drift**: New documentation uses "cross-rite" but the schema says "cross-team"

## Decision

Rename `cross_team_protocol` to `cross_rite_protocol` in the orchestrator.yaml schema.

This is a **BREAKING CHANGE** requiring:
1. Schema field rename in `schemas/orchestrator.yaml.schema.json`
2. Updates to all rite orchestrator.yaml files that use this field
3. Atomic commit including both schema and rite updates to maintain schema validity

## Consequences

### Positive

- **Terminology Consistency**: All schema fields now use "rite" terminology
- **Semantic Clarity**: Field name matches its actual function (cross-rite coordination)
- **Migration Completion**: Completes the Knossos terminology migration
- **Future Maintainability**: New rites won't encounter confusing legacy terminology

### Negative

- **Breaking Change**: Any external tools parsing orchestrator.yaml must update
- **Manual Migration**: All rites using this field must be updated simultaneously
- **Documentation Updates**: Examples and references must be updated

### Neutral

- **No Functionality Change**: This is purely a terminology update
- **Schema Version**: No schema version bump needed (semantic change only)
- **Backward Compatibility**: NOT maintained per zero-compatibility migration policy

## Implementation

### Phase 1: Schema Update

Update `schemas/orchestrator.yaml.schema.json`:
- Line 181-188: Rename field definition
- Lines 255-258, 317, 320, 374, 423: Update all references

### Phase 2: Rite Updates

Update all rites that use `cross_team_protocol`:
```bash
find rites/ -name "orchestrator.yaml" -exec grep -l "cross_team_protocol" {} \;
```

Replace with `cross_rite_protocol` in each file.

### Phase 3: Validation

Run orchestrator validation workflow:
```bash
.github/workflows/validate-orchestrators.yml
```

Ensure all rites pass schema validation.

### Rollback Strategy

Single atomic commit includes both schema and rite changes. Rollback via:
```bash
git revert <commit-hash>
```

## Related

- **SMELL-REPORT-knossos-terminology-migration-2026-01-07.md**: SM-015
- **REFACTOR-PLAN-knossos-terminology-2026-01-07.md**: RF-012
- **ADR-0009-knossos-roster-identity.md**: Original rite naming rationale

## Notes

This ADR completes the Knossos terminology migration by addressing the final schema-level "team" reference. All future work will use consistent "rite" terminology throughout the platform.
