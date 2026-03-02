## Context Design: Resolve Entry-Point Semantic Conflict (SD-4)

### Problem Statement

Two fields across `manifest.yaml` and `workflow.yaml` both contain "entry point" semantics but carry different values in most rites. There are 17 rites total (excluding `shared`). This creates confusion about which agent is the "real" entry point, and the codebase resolves the ambiguity through a priority rule in `internal/rite/discovery.go:254-258` that most authors are unaware of.

### Findings

| Field | File | Semantic | Consumer | Value in all rites |
|-------|------|----------|----------|--------------------|
| `entry_agent` | `manifest.yaml` | CC routing entry: which agent Task tool invokes first | `discovery.go`, `validate.go`, `status.go`, `info.go` | Always `pythia` |
| `entry_point.agent` | `workflow.yaml` | Domain entry: which specialist starts the domain workflow | `workflow.go`, `validate.go` (warning only) | First-phase specialist |

**Priority rule** (discovery.go:254-258): `entry_agent` always wins. `entry_point.agent` is a dead-letter fallback that only activates when `entry_agent` is empty (never in practice).

**Actual conflict scope**: Of 17 rites (excluding `shared`), 15 follow the standard mismatch pattern: `entry_agent: pythia` + `entry_point.agent: <first-phase-specialist>`. The remaining 2 rites (`thermia` and `releaser`) deliberately set `entry_point.agent: pythia` with separate fields for first specialist (`first_specialist` and `first_phase_agent` respectively) -- these are intentional designs documented with comments, not authoring errors. The original report identified 9 affected rites; the actual count is 15.

### Options Evaluated

#### Option A: Align both to `pythia`
- Edit 15 `workflow.yaml` files to set `entry_point.agent: pythia`
- **Rejected**: Destroys useful domain information. The workflow `entry_point` carries artifact metadata (type, path_template) that semantically belongs to the first specialist, not the orchestrator. Making pythia the entry_point would require moving or duplicating artifact templates, and would misrepresent the domain workflow structure.

#### Option B: Align both to the specialist
- Edit 15 `manifest.yaml` files to change `entry_agent` from `pythia` to the specialist
- **Rejected**: Factually wrong. CC always invokes pythia first via Task tool. Changing `entry_agent` would break the `internal/registry/validate.go` validation (which checks entry_agent is in agents list and has a file), and would misrepresent how CC actually routes work. The `Rite.EntryPoint` consumed by `status.go` and `info.go` would show the wrong agent.

#### Option C: Document the distinction (SELECTED)
- Add YAML comments to both files explaining the semantic difference
- Leave thermia as-is (its `entry_point.agent: pythia` is intentional and already documented with extensive comments and a separate `first_specialist` field)
- **Rationale**: These fields genuinely serve different purposes consumed by different systems. The manifest `entry_agent` is consumed by rite discovery/status/validation (CC-facing). The workflow `entry_point` is consumed by workflow loading and displayed in workflow descriptions (domain-facing). Documenting the distinction is cheaper than removing a field and cheaper than adding migration complexity. The "conflict" is not a bug -- it is an undocumented intentional design.

#### Option D: Remove `entry_point.agent` from workflow.yaml
- **Rejected**: The workflow `entry_point` carries `artifact` metadata (type, path_template) that has no home in manifest.yaml. Removing it would lose data or require a schema migration to manifest.yaml. Additionally, `entry_point` serves as human-readable documentation of "where the domain work starts" independent of orchestrator routing.

### Selected Approach: Option C -- Document the Distinction

### Components Affected

**manifest.yaml** (all 17 rites):
- Add YAML comment above `entry_agent` explaining it is the CC routing agent (invoked via Task tool)

**workflow.yaml** (15 rites -- all except thermia and releaser):
- Add YAML comment above `entry_point` explaining it is the domain workflow entry (first specialist in phase chain)

**thermia/workflow.yaml and releaser/workflow.yaml** (no changes):
- Both already have documentation explaining their deliberate choice to set `entry_point.agent: pythia` with separate fields for the first specialist (`first_specialist` and `first_phase_agent` respectively). These are intentional single-entry consultation models, not authoring errors.

### Backward Compatibility: COMPATIBLE

No schema changes. No Go code changes. No behavioral changes. All edits are YAML comments (ignored by parsers). No data corrections.

### Schema Definitions

No schema changes required. Both fields retain their existing types:

```yaml
# manifest.yaml
entry_agent: string  # Required. Agent invoked via Task tool. Always "pythia" for orchestrated rites.

# workflow.yaml
entry_point:
  agent: string      # Required. First specialist in the domain workflow phase chain.
  artifact:          # Optional. Artifact metadata for the entry specialist's output.
    type: string
    path_template: string
```

### File-Level Changes

#### manifest.yaml (17 rites)

Add comment block before `entry_agent` line in each manifest:

```yaml
# entry_agent: The agent CC invokes first via Task tool (orchestrator routing).
# This is distinct from workflow.yaml entry_point.agent (domain workflow entry).
entry_agent: pythia
```

Files to edit:
- `rites/ecosystem/manifest.yaml`
- `rites/docs/manifest.yaml`
- `rites/sre/manifest.yaml`
- `rites/security/manifest.yaml`
- `rites/intelligence/manifest.yaml`
- `rites/rnd/manifest.yaml`
- `rites/strategy/manifest.yaml`
- `rites/debt-triage/manifest.yaml`
- `rites/slop-chop/manifest.yaml`
- `rites/forge/manifest.yaml`
- `rites/10x-dev/manifest.yaml`
- `rites/thermia/manifest.yaml`
- `rites/clinic/manifest.yaml`
- `rites/review/manifest.yaml`
- `rites/hygiene/manifest.yaml`
- `rites/arch/manifest.yaml`
- `rites/releaser/manifest.yaml`

#### workflow.yaml (15 rites -- all except thermia and releaser which are already documented)

Add comment block before `entry_point` line:

```yaml
# entry_point: The first specialist in the domain workflow (not the CC routing agent).
# The CC routing agent is manifest.yaml:entry_agent (always pythia for orchestrated rites).
entry_point:
  agent: <existing-specialist>
```

#### thermia/workflow.yaml and releaser/workflow.yaml (no changes)

Both already have comments documenting their intentional choice to use `entry_point.agent: pythia` with separate fields for the first specialist. No additional comments needed.

### Integration Tests

No new integration tests required. This is a documentation-only change (YAML comments).

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| Standard (ecosystem) | Existing manifest drift test | Passes unchanged (comments ignored by YAML parser) |
| Complex (10x-dev) | Existing rite validation | Passes unchanged (comments ignored) |
| All rites | `CGO_ENABLED=0 go test ./internal/rite/...` | All existing tests pass |
| All rites | `CGO_ENABLED=0 go test ./internal/registry/...` | All existing tests pass |

### Migration Path

None required. All changes are backward compatible:
- YAML comments are ignored by all parsers
- No data changes to any rite

### Decisions Log

| Decision | Rationale |
|----------|-----------|
| Document rather than remove | Both fields serve distinct consumers (CC routing vs. domain workflow) and carry different metadata |
| Leave thermia and releaser workflow.yaml unchanged | Both deliberately set `entry_point.agent: pythia` with separate first-specialist fields; intentional design already documented with inline comments |
| Add comments to all 17 manifest.yaml and 15 workflow.yaml (skip thermia and releaser) | The semantic distinction applies universally; selective documentation would leave the confusion partially unresolved; thermia and releaser already self-document |
| No Go code changes | The priority rule in discovery.go is correct behavior, not a bug; it properly prefers the CC-facing entry_agent |
