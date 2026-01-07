---
artifact_id: ADR-0013
title: "Moirai Consolidation: 4 Agents to 1 Unified Agent with Fate Skills"
created_at: "2026-01-07T23:30:00Z"
author: architect
status: accepted
context: "The current 4-agent Moirai architecture (Moirai router + Clotho + Lachesis + Atropos) introduces routing overhead, cognitive burden, and maintenance complexity. Users must understand Fate taxonomy despite invoking Moirai generically, and the router adds latency before reaching the actual Fate agent."
decision: "Consolidate into a single unified Moirai agent that handles all session/sprint state operations directly, with Fates preserved as internal progressive disclosure skills loaded on-demand."
consequences:
  - type: positive
    description: "Single invocation pattern (Task(moirai, ...)) eliminates routing confusion"
  - type: positive
    description: "Reduced latency by eliminating router-to-Fate subprocess hop"
  - type: positive
    description: "Simplified maintenance with 1 agent file instead of 5"
  - type: positive
    description: "Progressive disclosure via skills preserves Fate domain knowledge without agent overhead"
  - type: negative
    description: "Breaking change: Task(clotho/lachesis/atropos, ...) calls will fail"
    mitigation: "Hard break accepted; remove old agent files and update all references simultaneously"
  - type: negative
    description: "Skill loading adds ~50-100 lines of context per operation"
    mitigation: "Skills are loaded lazily; only relevant Fate skill loaded per operation"
  - type: neutral
    description: "Users never see Fates directly; they use slash commands or Task(moirai, ...)"
  - type: neutral
    description: "CLI remains authoritative for session operations; Moirai wraps ari commands"
supersedes: ADR-0005
related_artifacts:
  - PRD-moirai-consolidation
  - TDD-moirai-unified-agent
  - TDD-fate-skills
tags:
  - architecture
  - session-management
  - moirai
  - consolidation
schema_version: "1.0"
---

# ADR-0013: Moirai Consolidation

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-01-07 |
| **Deciders** | Architecture Team |
| **Supersedes** | ADR-0005 (Moirai Centralized State Authority) |
| **Superseded by** | N/A |

## Context

The Moirai session management system evolved from a single concept to a 4-agent architecture following Greek mythology:

```
Current Architecture (Being Replaced):

User/Agent --> Task(moirai) --> Moirai Router --> Task(clotho|lachesis|atropos) --> File Mutation
                                    |
                              4 separate agents:
                              - user-agents/moirai.md (router)
                              - user-agents/clotho.md (creation)
                              - user-agents/lachesis.md (measurement)
                              - user-agents/atropos.md (termination)
                              - user-agents/moirai-shared.md (shared definitions)
```

### Problems Identified

1. **Cognitive Overhead**: Users must learn the Fate taxonomy (Clotho spins, Lachesis measures, Atropos cuts) to understand routing, but then invoke Moirai generically anyway.

2. **Routing Latency**: Extra subprocess hop through Moirai router before reaching actual Fate adds 50-100ms per operation.

3. **Broken Write Guard**: The hook blocks direct writes but suggests `Task(moirai, ...)` which then routes to a Fate, adding complexity. Main thread agents bypass Moirai entirely via `STATE_MATE_BYPASS` mechanism.

4. **Maintenance Complexity**: 5 separate files (4 agents + shared) require coordinated updates for any operation change.

5. **No Discovery**: Users have no way to progressively learn Moirai operations without reading all agent files.

### Forces

- **Simplicity**: Users should learn one invocation pattern, not four
- **Performance**: Eliminate unnecessary subprocess hops
- **Maintainability**: Single source of truth for session operations
- **Domain Knowledge**: Preserve Fate taxonomy for semantic clarity
- **CLI Authority**: `ari` CLI remains the authoritative source for state changes

## Decision

### Decision 1: Consolidate 4 Agents into 1 Unified Moirai Agent

Create a single `.claude/agents/moirai.md` agent that:

- Handles ALL session state operations directly (no routing to sub-agents)
- Parses both natural language and structured commands
- Returns structured JSON responses
- Logs all mutations to audit trail with reasoning

**Operations Handled:**

| Operation | Former Fate | Description |
|-----------|-------------|-------------|
| `create_sprint` | Clotho | Create new sprint within session |
| `start_sprint` | Clotho | Activate pending sprint |
| `mark_complete` | Lachesis | Record task completion |
| `transition_phase` | Lachesis | Progress workflow phase |
| `update_field` | Lachesis | Update context field |
| `park_session` | Lachesis | Pause session with reason |
| `resume_session` | Lachesis | Resume from parked state |
| `handoff` | Lachesis | Record agent transition |
| `record_decision` | Lachesis | Log decision |
| `append_content` | Lachesis | Append to context body |
| `wrap_session` | Atropos | Archive session with sails |
| `generate_sails` | Atropos | Compute confidence signal |
| `delete_sprint` | Atropos | Remove or archive sprint |

### Decision 2: Fates Become Progressive Disclosure Skills

Transform Fate agents into Moirai-internal skills at `.claude/skills/moirai/`:

```
.claude/skills/moirai/
├── SKILL.md          # Entry point with routing table
├── clotho.md         # Creation operations (Spinner)
├── lachesis.md       # Measurement operations (Measurer)
└── atropos.md        # Termination operations (Cutter)
```

**Why skills, not inline documentation:**

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Inline everything in agent | Single file, no loading | Agent prompt bloat (~800 lines), no progressive disclosure | Rejected |
| Separate documentation files | Organized | Not automatically loadable | Rejected |
| Skills (progressive disclosure) | Lazy loading, ~50-100 lines per domain, preserves Fate semantics | Adds Read operation per invocation | **Selected** |

**Skill loading pattern:**

```
1. User invokes: /park "taking a break"
2. Slash command routes to Moirai: Task(moirai, "park_session...")
3. Moirai parses operation: "park_session"
4. Moirai reads SKILL.md → park_session maps to lachesis
5. Moirai reads lachesis.md → extracts park_session specification
6. Moirai executes: ari session park --reason "taking a break"
7. Moirai returns structured JSON response
```

### Decision 3: CLI Remains Authoritative

Moirai delegates to `ari` CLI for session state changes rather than reimplementing:

| Moirai Operation | CLI Command |
|------------------|-------------|
| `create_session` | `ari session create` |
| `park_session` | `ari session park` |
| `resume_session` | `ari session resume` |
| `wrap_session` | `ari session wrap` |
| `transition_phase` | `ari session transition` |
| `handoff` | `ari handoff execute` |
| `generate_sails` | `ari sails check` |

**Why CLI authority:**

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Agent-native implementation | Full control | Dual implementation, divergence risk | Rejected |
| Dual implementation | Flexibility | Maintenance nightmare, consistency issues | Rejected |
| CLI delegation | Single source of truth, CLI already tested | Adds Bash call | **Selected** |

### Decision 4: Hard Break Compatibility

Remove old Fate agent files rather than deprecation or shimming:

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Soft deprecation | Gradual migration | Prolongs confusion, maintenance burden | Rejected |
| Shim layer | Backward compatibility | Added complexity, masks real interface | Rejected |
| Hard break | Clean architecture, forces update | Breaking change | **Selected** |

**Files to remove:**

| File | Reason |
|------|--------|
| `user-agents/moirai.md` | Replaced by `.claude/agents/moirai.md` |
| `user-agents/clotho.md` | Replaced by skill |
| `user-agents/lachesis.md` | Replaced by skill |
| `user-agents/atropos.md` | Replaced by skill |
| `user-agents/moirai-shared.md` | Absorbed into skills |

**Migration impact:**

- `Task(clotho, ...)` will fail (agent removed)
- `Task(lachesis, ...)` will fail (agent removed)
- `Task(atropos, ...)` will fail (agent removed)
- `Task(moirai, ...)` continues to work (new unified agent)

## Consequences

### Positive

1. **Single Invocation Pattern**: Users learn `Task(moirai, "operation")` only; no Fate selection required
2. **Reduced Latency**: Eliminates router-to-Fate subprocess hop (50-100ms savings)
3. **Simplified Maintenance**: 1 agent + 4 skills vs. 5 agent files
4. **Progressive Disclosure**: Skills load only relevant domain knowledge (~50-200 lines vs. ~800 lines)
5. **Preserved Semantics**: Fate taxonomy remains in skills for domain understanding
6. **CLI Consistency**: Single source of truth for session operations prevents divergence
7. **Better Write Guard**: Clearer error messages reference unified Moirai, not multiple Fates

### Negative

1. **Breaking Change**: All existing `Task(clotho|lachesis|atropos, ...)` calls will fail immediately
   - *Mitigation*: Remove old files and update all references in atomic commit

2. **Skill Loading Overhead**: Each operation requires Read of SKILL.md + domain skill
   - *Mitigation*: Skills are compact (~50-200 lines); total context < 250 lines per operation

3. **Loss of Direct Fate Invocation**: Power users cannot invoke Fates directly
   - *Mitigation*: Not a real use case; users always went through Moirai or slash commands

### Neutral

1. **User Invisibility**: Users never see Fates; they use slash commands (`/park`, `/wrap`) or `Task(moirai, ...)`
2. **Audit Trail**: Logs preserve Fate domain attribution for traceability
3. **Schema Unchanged**: Session/sprint context schemas remain unchanged

## Alternatives Considered

### Alternative 1: Keep 4-Agent Pattern

Retain current architecture with improvements to routing and documentation.

**Rejected because:**
- Does not address cognitive overhead (users still need to understand Fate taxonomy)
- Routing latency remains
- Maintenance complexity unchanged
- Write guard still confusing

### Alternative 2: Merge Only Fates (Keep Router)

Merge Clotho/Lachesis/Atropos into single agent but keep Moirai as router.

**Rejected because:**
- Adds unnecessary abstraction layer
- Router becomes passthrough with no value
- Still requires understanding Moirai vs. "merged Fate"

### Alternative 3: Pure Skill Approach (No Agent)

Eliminate Moirai agent entirely; slash commands invoke skills directly.

**Rejected because:**
- Skills cannot execute tools (Read/Write/Bash)
- Would require main thread to execute operations
- Loses centralized audit logging and validation
- Write guard cannot invoke agents

## Implementation

### Phase 1: Create Unified Agent
1. Write `.claude/agents/moirai.md` with all operations
2. Implement operation parser (structured + natural language)
3. Add Fate domain routing table

### Phase 2: Create Fate Skills
1. Write `.claude/skills/moirai/SKILL.md` (routing table)
2. Write `.claude/skills/moirai/clotho.md` (creation)
3. Write `.claude/skills/moirai/lachesis.md` (measurement)
4. Write `.claude/skills/moirai/atropos.md` (termination)

### Phase 3: CLI Integration
1. Map operations to `ari` commands
2. Implement CLI output parsing
3. Add error code translation

### Phase 4: Update Slash Commands
1. Update `/park` to invoke unified Moirai
2. Update `/wrap` to invoke unified Moirai
3. Update `/handoff` to invoke unified Moirai

### Phase 5: Remove Old Files
1. Delete `user-agents/moirai.md`
2. Delete `user-agents/clotho.md`
3. Delete `user-agents/lachesis.md`
4. Delete `user-agents/atropos.md`
5. Delete `user-agents/moirai-shared.md`

### Phase 6: Update Documentation
1. Update `.claude/CLAUDE.md` State Management section
2. Update any references to old Fate agents
3. Create `moirai-ref.md` discovery skill

## Validation

### Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Routing Errors | 0 | No FATE_MISMATCH in audit log |
| Invocation Pattern | 1 | Users use only `Task(moirai, ...)` |
| Write Guard Catches | 100% | All direct writes blocked |
| Latency | <500ms | End-to-end operation timing |
| Audit Coverage | 100% | All mutations have reasoning |

### Test Cases

| Test | Description | Pass Criteria |
|------|-------------|---------------|
| park_session | Park active session | Session state PARKED, CLI invoked |
| wrap_session | Wrap with WHITE sails | Session ARCHIVED, sails generated |
| create_sprint | Create new sprint | Sprint file created, pending status |
| mark_complete | Complete task | Task status completed, artifact logged |
| handoff | Transfer to agent | Agent transition recorded |

## Related Decisions

- **ADR-0005**: Moirai Centralized State Authority (superseded by this ADR)
- **ADR-0001**: Session State Machine Redesign (defines FSM that unified Moirai enforces)
- **ADR-0009**: Knossos Roster Identity (establishes Moirai/Fates naming)

## References

- PRD: `docs/requirements/PRD-moirai-consolidation.md`
- TDD (Unified Agent): `docs/design/TDD-moirai-unified-agent.md`
- TDD (Fate Skills): `docs/design/TDD-fate-skills.md`
- Current Moirai Router: `user-agents/moirai.md`
- Current Fates: `user-agents/clotho.md`, `user-agents/lachesis.md`, `user-agents/atropos.md`
