# Knossos Concepts

This guide explains how Knossos works using plain language. Technical terms are introduced gradually with clear definitions.

## The Core Idea

Claude Code is powerful, but a single AI assistant doing requirements, design, implementation, testing, and review in one conversation leads to context loss, inconsistent quality, and no accountability. Knossos solves this by splitting work across **specialized agents** that each do one thing well.

## Agents

An **agent** is a Claude Code subagent with a specific role, tools, and behavioral instructions. Each agent is defined by a markdown file in `.claude/agents/`.

For example, in the `review` rite:

- **Scanner** — reads your codebase and identifies areas of concern. Has access to Bash, Glob, Grep, Read. Cannot edit files.
- **Assessor** — evaluates scanner findings and prioritizes by impact. Can verify findings by reading code but cannot modify it.
- **Reporter** — synthesizes everything into a structured review document.

Agents are constrained by design:
- Each agent only has the tools it needs (a scanner doesn't need Edit)
- Each agent knows what it decides, what it escalates, and what it must not do
- Agents produce typed artifacts (scan-findings, assessment, review-report) that flow to the next agent

## Orchestrator

Every rite has an **orchestrator** — a coordinating agent that routes work to specialists. The orchestrator:

- Receives your request and decides which specialist to invoke first
- Reviews specialist output and decides the next phase
- Enforces handoff criteria (did the scanner produce valid findings before routing to assessor?)
- Never executes work directly — it only coordinates

The orchestrator is named **Pythia** by convention (after the Oracle at Delphi). You don't interact with Pythia directly — the main Claude Code thread consults it behind the scenes.

## Rites

A **rite** is a complete workflow definition. It bundles:

| Component | What it defines | File |
|-----------|----------------|------|
| **Manifest** | Agents, dependencies, defaults | `manifest.yaml` |
| **Workflow** | Phases, sequence, routing rules | `workflow.yaml` |
| **Agent prompts** | Behavioral instructions per agent | `agents/*.md` |
| **Skills** | Reference knowledge agents can load | `mena/**/INDEX.lego.md` |

Think of a rite as a "playbook" — it defines WHO does WHAT in WHICH ORDER.

### Example: The Review Rite

```
scan ──> assess ──> report
 │         │          │
 │         │          └─ Reporter produces final document
 │         └─ Assessor prioritizes findings
 └─ Scanner reads codebase structure
```

3 phases, 3 specialists, 1 orchestrator coordinating.

## Phases

A **phase** is a stage in the workflow. Each phase:

- Is handled by one specialist agent
- Produces a typed artifact (document)
- Has criteria that must be met before advancing to the next phase

Phases run sequentially by default. Some rites support conditional phases that only run at higher complexity levels.

## Sessions

A **session** tracks a unit of work from start to finish. Sessions:

- Have a unique ID and creation timestamp
- Track which phase you're in
- Can be **parked** (saved for later) and **resumed**
- Produce a summary when **wrapped** (completed)

You don't need sessions for quick tasks. They're useful when work spans multiple conversations or when you want progress tracking.

## Skills (Reference Knowledge)

**Skills** are reference documents that agents can load on-demand. Unlike agent prompts (always in context for that agent), skills are loaded only when needed — saving context window space.

For example, the `review-ref` skill contains:
- Review methodology and scan heuristics
- Severity model (critical/high/medium/low definitions)
- Report format template

An agent with `skills: [review-ref]` can load this reference when it needs methodology guidance.

## Commands (Slash Commands)

**Commands** are actions you invoke by typing `/` in Claude Code. They're the user-facing interface to Knossos operations:

- `/go` — start or resume a session
- `/commit` — create a well-structured git commit
- `/consult` — get guidance on which workflow to use

Commands are action-oriented (do something and exit). Skills are knowledge-oriented (stay in context as reference).

## How It All Fits Together

```
You type: /go "Review this codebase"
            │
            ▼
    ┌─── Orchestrator (Pythia) ───┐
    │  "Route to scanner first"    │
    └──────────┬───────────────────┘
               ▼
    ┌─── Scanner ─────────────────┐
    │  Reads files, finds issues   │
    │  Produces: SCAN-findings.md  │
    └──────────┬───────────────────┘
               ▼
    ┌─── Orchestrator ────────────┐
    │  "Scan complete, route to    │
    │   assessor"                  │
    └──────────┬───────────────────┘
               ▼
    ┌─── Assessor ────────────────┐
    │  Evaluates, prioritizes      │
    │  Produces: ASSESS-report.md  │
    └──────────┬───────────────────┘
               ▼
    ┌─── Orchestrator ────────────┐
    │  "Assessment complete, route │
    │   to reporter"               │
    └──────────┬───────────────────┘
               ▼
    ┌─── Reporter ────────────────┐
    │  Synthesizes final document  │
    │  Produces: REVIEW-report.md  │
    └──────────────────────────────┘
```

## Glossary

| Term | Plain meaning | Technical detail |
|------|--------------|-----------------|
| Rite | Workflow definition | Bundle of manifest + workflow + agents + skills |
| Agent | Specialist AI with a specific role | Claude Code subagent with constrained tools |
| Orchestrator | Coordinator that routes work | Read-only agent that decides phase transitions |
| Phase | Stage in the workflow | Sequential step producing a typed artifact |
| Session | Tracked unit of work | State machine: active → parked → wrapped |
| Skill | Reference document loaded on-demand | Markdown file in `.claude/skills/` |
| Command | User-invoked action (slash command) | Markdown file in `.claude/commands/` |
| Manifest | Rite configuration | YAML declaring agents, phases, dependencies |
| `ari` | CLI tool (Ariadne) | Go binary managing sync, sessions, hooks |
| Sync | Update `.claude/` from rite sources | `ari sync` materializes agents/skills/commands |
