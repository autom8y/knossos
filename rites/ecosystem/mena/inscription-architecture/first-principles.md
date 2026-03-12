# First Principles: Inscription Architecture

> "Session-specific knowledge should NOT be in the context file. The context file should just maintain alignment with the active rite to ensure standardized workflows are followed."

These six principles govern all inscription architecture decisions across the knossos/satellite ecosystem.

---

## Principle 1: The Context File is a Behavioral Contract

The context file tells the harness **what it can do** and **how to work**, not what it is currently doing.

### What the Context File Answers

1. **What can the harness do here?** (Capabilities: skills, agents, hooks)
2. **How should the harness work here?** (Workflow: routing patterns, handoff protocols)
3. **Who is the harness working as?** (Identity: active rite, available agents)

### What the Context File Is NOT

| Not This | Use This Instead |
|----------|------------------|
| Knowledge base | Skills for domain knowledge |
| Session log | SESSION_CONTEXT for work in progress |
| Task tracker | Todo tools, session files |
| Scratchpad | Conversation context for transient state |

### The Minimal Viable Context File

```markdown
# Context File

This project uses a {N}-agent workflow ({rite-name}).

## Available Agents
{list of agents with one-line descriptions}

## Skills
{list of skills the harness can invoke}

## Workflow
{routing guidance: when to invoke which agent}
```

Everything else is either infrastructure documentation (synced from knossos) or project extensions (satellite-owned).

---

## Principle 2: Stable Content Only

If content changes more than once per session, it does not belong in the context file.

### Stability Tiers

| Tier | Change Frequency | In Context File? | Examples |
|------|------------------|-----------------|----------|
| **STATIC** | Once (at init) | Rarely needed | Rite catalog source reference |
| **STABLE** | Weeks/months | **Yes** | Rite, skills, workflow |
| **DYNAMIC** | Days/weekly | **No** | Current initiative |
| **EPHEMERAL** | Minutes/hourly | **No** | Git state, current task |

### The Stability Boundary

```
Context file contains: STATIC + STABLE
Context file excludes: DYNAMIC + EPHEMERAL
```

### Applying the Classification

| Content | Stability | In Context File? | Alternative |
|---------|-----------|-----------------|-------------|
| Skill documentation | STABLE | Yes | - |
| Agent catalog | STABLE | Yes (regenerated) | - |
| Workflow patterns | STABLE | Yes (synced) | - |
| Active rite name | STABLE | Yes (regenerated) | ACTIVE_RITE file |
| Current initiative | DYNAMIC | No | SESSION_CONTEXT |
| Sprint goals | DYNAMIC | No | PRD, project docs |
| Git branch | EPHEMERAL | No | Hook output |
| Current task | EPHEMERAL | No | Conversation context |

---

## Principle 3: Separation by Source of Truth

Each concern has exactly one owner. Content placement follows ownership.

### The Five Concerns

| Concern | Owner | In Context File? | Sync Behavior |
|---------|-------|-----------------|---------------|
| **Ecosystem Infrastructure** | Knossos | Yes | SYNC |
| **Project Identity** | Satellite | Yes | PRESERVE |
| **Rite Configuration** | Rite | Yes | REGENERATE |
| **Session State** | Session | **No** | N/A |
| **Workflow Guidance** | Knossos | Yes | SYNC |

### The Critical Distinction

```
Context file = Ecosystem Infrastructure + Project Identity + Rite Configuration

Context file != Session State
```

Session state is injected by hooks at session start. It appears in the harness context but never writes back to the context file.

### Source of Truth Tests

| Test | Answer | Result |
|------|--------|--------|
| Knossos | SYNC |
| Satellite | PRESERVE or PROJECT |
| ACTIVE_RITE + agents/ | REGENERATE |
| Session files | NOT IN CONTEXT FILE |

---

## Principle 4: Injection for Transient State

Hooks inject ephemeral context at session start. It appears in the harness view but never writes to the context file.

### The Three Layers

```
+-----------------------------------------------------------------+
|                    SESSION CONTEXT                               |
|  (Injected by hooks at session start)                            |
|  - Current git state                                             |
|  - Active session info                                           |
|  - Worktree context                                              |
|  - Rite routing hints                                            |
+-----------------------------------------------------------------+
                            |
                            | supplements
                            v
+-----------------------------------------------------------------+
|                PROJECT CONTEXT FILE                              |
|  (.channel/CLAUDE.md)                                            |
|  - Rite configuration (from knossos)                              |
|  - Project identity (satellite-owned)                            |
|  - Ecosystem infrastructure (synced from knossos)                 |
+-----------------------------------------------------------------+
                            |
                            | supplements
                            v
+-----------------------------------------------------------------+
|                 GLOBAL CONTEXT FILE                              |
|  (~/.channel/CLAUDE.md)                                          |
|  - Personal preferences                                          |
|  - Global tool configurations                                    |
|  - User-wide skills                                              |
+-----------------------------------------------------------------+
```

### Layer Responsibilities

| Layer | Content Type | Modified By |
|-------|--------------|-------------|
| Global | Personal preferences, global tools | User manually |
| Project | Rite + Project + Infrastructure | ari sync, knossos |
| Session | Transient state, current work | Hooks (read-only to context file) |

### Hook Output Example

```markdown
## Project Context (auto-loaded)

| Property | Value |
|----------|-------|
| **Project** | /Users/dev/myproject |
| **Active Rite** | docs |
| **Git** | feature/add-auth (3 uncommitted) |
```

This content is:
- Generated fresh on each session start
- Never written to the context file
- Only exists in the harness conversation context
- Can change between sessions without file modification

---

## Principle 5: Single Purpose per Content

Each piece of content has one owner, one sync behavior, one location.

### Content Classification

| Content | Owner | Sync | Location |
|---------|-------|------|----------|
| Skills Architecture table | Knossos | SYNC | Context file |
| Project-specific conventions | Satellite | PRESERVE | Context file (## Project:*) |
| Quick Start agent table | Rite | REGENERATE | Context file |
| Current initiative | Session | N/A | SESSION_CONTEXT |
| Git state | Hooks | N/A | Hook output only |

### The Anti-Duplication Rule

If content exists in multiple places, one becomes stale. Pick the authoritative source and reference it:

- **Derived content**: Regenerate from source (rite sections from ACTIVE_RITE + agents/)
- **Transient content**: Inject via hooks, never persist
- **Project extensions**: Use `## Project:*` namespace

---

## Principle 6: The Decay Test

Content that decays (becomes stale) without active maintenance does not belong in the context file.

### The Test

> "If I don't update this for a month, is the context file incorrect?"

| Answer | Verdict |
|--------|---------|
| **No** (still accurate) | Belongs in the context file (STABLE) |
| **Yes** (becomes stale) | Does not belong (DYNAMIC/EPHEMERAL) |

### Applying the Decay Test

| Content | After 1 Month | Belongs? |
|---------|---------------|----------|
| Skills documentation | Still accurate | Yes |
| Workflow patterns | Still accurate | Yes |
| Agent catalog | Still accurate (unless rite swapped) | Yes |
| "Currently working on X" | Stale immediately | No |
| Git branch name | Stale within hours | No |
| "Last updated: DATE" | Stale immediately | No |

### The Ultimate Test

> "If this content disappeared, would the harness work less effectively in this project?"

- **YES**: The content describes capabilities, workflow, or identity that the harness needs
- **NO**: The content is informational, historical, or transient

The context file should contain only content that passes the YES test.

---

## Summary Table

| Principle | Rule | Test |
|-----------|------|------|
| 1. Behavioral Contract | What the harness can do and how | Does it define capabilities or workflow? |
| 2. Stable Content Only | Changes less than once per session | Will it be accurate in a month? |
| 3. Separation by Source | One owner per content | Where is the authoritative source? |
| 4. Injection for Transient | Hooks inject, context file stores | Does it change per session? |
| 5. Single Purpose | One owner, one location | Is it duplicated anywhere? |
| 6. Decay Test | No maintenance = still accurate | Does it rot without updates? |

---

## Decision Record

Key architectural decisions and their rationale:

| Decision | Rationale |
|----------|-----------|
| Session state excluded from context file | Changes too frequently, creates maintenance burden |
| Rite sections regenerated, not copied | Satellites have their own rites from ACTIVE_RITE |
| Hooks inject transient context | Separation of stable (file) vs ephemeral (context) |
| PRESERVE as default for unknown sections | Encourages experimentation, safer than deletion |
| `## Project:*` pattern for extensions | Clear namespace, prevents conflicts with knossos sections |

---

## What Goes Where (Quick Reference)

```
Context file (stable behavioral contract)
+-- Skills (what the harness can invoke)
+-- Agents (who is available)
+-- Workflow (how work flows)
+-- Project extensions (## Project:*)
+-- Infrastructure docs

SESSION_CONTEXT (transient session state)
+-- Current initiative
+-- Current phase
+-- Parked status
+-- Handoff context
+-- Session metadata

Hook Output (ephemeral context)
+-- Git state
+-- Worktree info
+-- Session suggestions
+-- Rite routing hints

~/.channel/CLAUDE.md (user preferences)
+-- Personal tool configs
+-- Global skills
+-- User-wide defaults
```

---

## Related Files

- [ownership-model.md](ownership-model.md) - Detailed section ownership
- [boundary-test.md](boundary-test.md) - Validation checklist
- [anti-patterns.md](anti-patterns.md) - What NOT to put in the context file
