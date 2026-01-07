# ADR-0015: Content Organization

| Field | Value |
|-------|-------|
| **Status** | ACCEPTED |
| **Date** | 2026-01-07 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |
| **Related** | SPIKE-content-organization.md, SPIKE-knossos-consolidation-architecture.md, TDD-ariadne-manifest.md |

---

## Context

The Knossos platform organizes content (skills, agents, rites) across multiple locations that have evolved organically during development. The current structure includes:

1. **Runtime materialized content** (`.claude/skills/`, `.claude/agents/`) - populated by `swap-rite.sh`
2. **Source content** (`rites/*/skills/`, `rites/*/agents/`) - canonical definitions
3. **Shared content** (`rites/shared/skills/`) - cross-rite primitives
4. **User overrides** (`user-skills/`, `user-agents/`) - user-level defaults

This distributed structure works but lacks:
- Formal manifest schema defining rite contents
- Validation of skill references in agents
- Clear documentation of the content hierarchy
- Explicit shared skill inclusion rules

The upcoming consolidation (see SPIKE-knossos-consolidation-architecture.md) requires formalizing this structure to enable Go-based sync and full `.claude/` generation.

---

## Decision

We adopt the following content organization architecture:

### 1. Rite Structure

Each rite follows this directory structure:

```
rites/{rite-name}/
+-- manifest.yaml       # REQUIRED: Rite metadata and configuration
+-- agents/             # REQUIRED: Agent definitions
+-- skills/             # OPTIONAL: Rite-specific skills
+-- README.md           # REQUIRED: Human documentation
+-- workflow.yaml       # OPTIONAL: Orchestration workflow
```

### 2. Manifest Schema

The `manifest.yaml` file is the authoritative source for rite configuration:

```yaml
version: "1.1"
type: rite                    # rite | shared
name: {rite-name}             # Must match directory name
description: {description}

agents:
  - name: {agent-name}
    file: agents/{agent-name}.md
    role: {role description}
    produces: {artifact type}

skills:
  include:                    # Rite-specific skills to include
    - {skill-name}
  shared:                     # Shared skills to import from rites/shared/
    - cross-rite-handoff
  exclude: []                 # Explicit exclusions (rare)

workflow:
  type: sequential            # sequential | parallel | hybrid
  entry_point: {agent-name}

complexity:
  levels:
    - name: {LEVEL_NAME}
      description: {when to use}
      phases: [phase1, phase2]
```

### 3. Shared Skills

The `rites/shared/` directory uses a specialized manifest:

```yaml
version: "1.0"
type: shared
name: shared
description: Cross-rite primitives

skills:
  - name: cross-rite-handoff
    path: skills/cross-rite-handoff/
    description: HANDOFF artifact schema
```

Shared skills have no agents, no workflow, and are automatically available to all rites.

### 4. User Overrides

User-level content remains in dedicated directories:

```
user-skills/      # Synced to ~/.claude/skills/
user-agents/      # Synced to ~/.claude/agents/
```

These are NOT rites but provide user-level defaults that persist across projects.

### 5. Materialization Model

The `.claude/` directory is fully generated:

```
.claude/                      # GITIGNORED - entirely generated
+-- skills/                   # From: rite skills + shared skills + user skills
+-- agents/                   # From: rite agents + user agents
+-- hooks/                    # From: templates/hooks/
+-- ACTIVE_RITE               # Current rite name
+-- CLAUDE.md                 # Generated entry point
```

Generation order and precedence:
1. Shared skills (lowest precedence)
2. Rite-specific skills
3. User skills (highest precedence)

### 6. Skill Reference Syntax

Agents reference skills using `@` syntax:

```markdown
Produce artifact using `@skill-name#anchor`.
```

References are validated at sync time. Missing references produce warnings; in strict mode, they produce errors.

---

## Consequences

### Positive

1. **Clear ownership**: Every skill and agent belongs to exactly one rite (or shared/user)
2. **Explicit configuration**: Manifests declare what a rite contains
3. **Validated references**: Broken skill references caught at sync time
4. **Clean generation**: `.claude/` becomes fully reproducible from source
5. **Consistent structure**: All rites follow the same pattern

### Negative

1. **Migration effort**: 12 rites need manifest.yaml files created
2. **New validation**: Skill reference validation may surface existing issues
3. **Learning curve**: Contributors must understand manifest schema
4. **Tooling dependency**: Requires `ari` tooling for full benefit

### Neutral

1. **No content movement**: Source files stay in current locations
2. **Backward compatible**: Can be adopted incrementally
3. **Existing TDD**: Builds on TDD-ariadne-manifest.md foundations

---

## Alternatives Considered

### Alternative A: Flat Skills Directory

All skills in a single `skills/` directory, agents reference by path.

```
skills/
+-- rnd/doc-rnd/
+-- 10x-dev/doc-artifacts/
+-- shared/cross-rite-handoff/
```

**Rejected because**: Loses rite isolation; harder to understand skill ownership; doesn't support rite-specific overrides.

### Alternative B: Skills as Git Submodules

Each skill as its own repository, included via submodules.

```
rites/rnd/skills/doc-rnd/  -> git submodule
```

**Rejected because**: Over-engineered for current scale; adds git complexity; makes atomic changes across skills difficult.

### Alternative C: Central Registry

Skills and agents registered in central JSON/YAML, referenced by ID.

```yaml
# registry.yaml
skills:
  doc-rnd:
    location: rites/rnd/skills/doc-rnd/
    rites: [rnd]
```

**Rejected because**: Duplicates information; registry diverges from actual files; harder to maintain consistency.

---

## Implementation Notes

### Phase 1: Manifest Creation (Non-Breaking)

1. Create `manifest.yaml` for each rite based on current directory contents
2. Validate manifests with `ari manifest validate`
3. No behavior change to `swap-rite.sh`

### Phase 2: Manifest Integration

1. Update `swap-rite.sh` to read manifest for agent/skill lists
2. Add skill reference validation (warnings only)
3. Test thoroughly before enabling

### Phase 3: Full Generation

1. Gitignore `.claude/` directory
2. Migrate to `ari sync` for generation
3. Remove `swap-rite.sh` sync logic

---

## References

- SPIKE-content-organization.md - Detailed analysis and findings
- SPIKE-knossos-consolidation-architecture.md - Parent consolidation spike
- TDD-ariadne-manifest.md - Existing manifest schema design
- rites/shared/README.md - Shared skills documentation
