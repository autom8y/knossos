---
name: interview
description: "Structured requirements interview producing a spec artifact. Use when: starting a new feature, rite, or initiative where requirements are unclear. Inversion of prompt engineering — the model interviews the user, not the other way around."
argument-hint: "<topic> [--output=<path>] [--depth=shallow|standard|deep]"
context: fork
---

# /interview — Structured Requirements Interview

Run a sustained, phased interview that probes for non-obvious concerns, surfaces tradeoffs, and produces a concise spec artifact. The model interviews YOU — inverting the traditional prompt engineering relationship.

## Core Principles

1. **Ask, don't assume.** When you encounter ambiguity, surface a structured question rather than guessing. Every assumption you skip over becomes a bug someone has to find later.

2. **Front-load decisions when they're cheap to change.** Every question is a fork in the road; every answer narrows the solution space. By the time execution begins, the decision tree has been navigated together — tradeoffs are explicit, not buried in code review.

3. **Questions must not be obvious.** Probe for non-obvious concerns: tradeoffs, edge cases, failure modes, UX implications, things the user hasn't thought of yet. Act like a senior architect or PM pushing the spec toward completeness, not a form collecting known answers.

4. **Structured multiple-choice over open-ended text.** Use AskUserQuestion with options that have short labels and descriptions. This lets the user convey intent accurately in seconds. Reduce cognitive load. Free-text escape hatches ("Other") exist but should not be the default path.

5. **Interview continuously until complete, then write the spec.** This is not one question — it's a sustained, iterative dialogue that drills progressively deeper. The output artifact is written only after sufficient coverage, not after a fixed number of questions.

6. **Separate planning from execution.** The spec produced becomes the source of truth for a fresh execution session. Context isolation means the executor gets a clean, unambiguous mandate without inheriting messy exploration context.

## Behavior

### 1. Parse Input and Establish Scope

Extract topic from arguments. Determine depth:
- **shallow**: 3-5 questions, focused on intent and constraints. Good for small features.
- **standard** (default): 8-15 questions across all phases. Good for modules and services.
- **deep**: 15-25+ questions, exhaustive coverage. Good for initiatives and architectural decisions.

If no topic is provided, open with a single question: "What are we building?"

### 2. Load Context

Before asking the first substantive question:
- Read relevant files in the working directory to understand the codebase
- Check for existing specs, PRDs, or design docs related to the topic
- Note the active rite and session (if any) — this shapes output format

Do NOT display what you found. Use it to ask BETTER questions.

### 3. Conduct Phased Interview

Progress through four phases. Each phase has a PURPOSE — move to the next phase when that purpose is satisfied, not after a fixed count.

#### Phase 1: UNDERSTAND (Intent & Constraints)
**Purpose**: Establish what the user wants and what limits exist.

Probe for:
- Core intent (what problem are we solving?)
- Success criteria (how do we know it works?)
- Hard constraints (time, tech stack, compatibility, compliance)
- What explicitly is NOT in scope

Question style: Broad, divergent. Open the solution space.

#### Phase 2: DESIGN (Architecture & Tradeoffs)
**Purpose**: Navigate key technical decisions with the user.

Probe for:
- Architectural approach (patterns, libraries, services)
- Tradeoffs between competing approaches (present them as options)
- Integration points and failure modes
- Data model and state management
- Edge cases the user hasn't considered

Question style: Focused, convergent. Close the solution space through explicit tradeoff decisions.

#### Phase 3: REVIEW (Confirm Alignment)
**Purpose**: Verify the interviewer's understanding matches the user's intent.

Present a brief synthesis of decisions made so far. Ask:
- "Does this capture your intent?"
- Surface any contradictions or gaps discovered during design phase
- Final opportunity for course correction before committing to the spec

Question style: Confirmatory. One or two questions maximum.

#### Phase 4: PLAN (Write the Spec)
**Purpose**: Produce the artifact.

Write the spec to the output path. The spec must:
- Include ONLY the recommended approach (not all alternatives explored)
- Be detailed enough to execute but concise enough to review in 2 minutes
- Reference specific files, paths, and entities from the codebase
- Include a phased implementation plan if the work spans multiple steps
- Note any open questions that were deferred (not silently dropped)

### 4. Output the Artifact

Default output path: `.wip/INTERVIEW-{topic-slug}.md`

Override with `--output=<path>`.

#### Spec Artifact Format

```markdown
# {Topic}

## Intent
{One paragraph: what we're building and why}

## Decisions
| Decision | Choice | Rationale |
|----------|--------|-----------|
| {key decision} | {what was chosen} | {why, including rejected alternatives} |

## Scope
**In scope**: {bulleted list}
**Out of scope**: {bulleted list}

## Design
{Technical approach — architecture, data model, integration points.
Reference specific files and paths. Keep it scannable.}

## Implementation Plan
{Ordered steps. Each step references files to create or modify.
Phased if the work is large enough to warrant it.}

## Open Questions
{Anything deferred. Empty section if fully resolved.}
```

### 5. Suggest Next Steps

After writing the spec:
```
Spec written to: {path}

Next steps:
  /build "{topic}"     — Execute from this spec
  /architect "{topic}" — Expand into full TDD + ADRs first
  /start "{topic}"     — Start a tracked session
```

## AskUserQuestion Patterns

Every question MUST use AskUserQuestion with structured options. Follow these patterns:

### Binary Decision
```
question: "Should the API support batch operations?"
options:
  - label: "Yes, batch support"
    description: "Adds complexity but reduces round-trips for bulk operations"
  - label: "No, single-item only"
    description: "Simpler API surface, clients loop themselves"
```

### Architecture Fork
```
question: "How should we handle state persistence?"
options:
  - label: "PostgreSQL"
    description: "Relational, ACID, good for structured data with joins"
  - label: "Redis"
    description: "Fast KV store, good for sessions/cache, less durable"
  - label: "SQLite"
    description: "Embedded, zero-config, good for single-node deployments"
```

### Non-Obvious Probe
```
question: "What happens when a user's session expires mid-operation?"
options:
  - label: "Fail and retry"
    description: "Operation fails cleanly, user re-authenticates and retries"
  - label: "Queue and resume"
    description: "Operation is saved, completes after re-auth without user re-triggering"
  - label: "Silent refresh"
    description: "Auto-refresh token in background, user never sees interruption"
```

## Anti-Patterns

- **Asking what you can look up.** If the codebase already answers a question, don't ask it. Read the code first.
- **Leading questions.** "Should we use the standard approach of X?" is not a real question. Present genuine alternatives.
- **Asking permission to ask questions.** Don't say "Can I ask you some questions about this?" Just start asking.
- **One giant question dump.** Ask 1-2 questions at a time. Let answers inform the next question. This is a dialogue, not a form.
- **Obvious questions.** "What language should we use?" when the repo is 100% Go. Read the room.
- **Parroting back answers.** Don't repeat what the user said before the next question. Acknowledge briefly and move forward.
- **Questions about preferences that don't affect the design.** Naming conventions, comment style, etc. — follow codebase conventions, don't ask.

## Depth Calibration

| Depth | Questions | Phases | Best For |
|-------|-----------|--------|----------|
| shallow | 3-5 | UNDERSTAND + PLAN | Bug fixes, small features, clear requirements |
| standard | 8-15 | All four phases | New features, modules, API design |
| deep | 15-25+ | All four phases, multiple rounds | Initiatives, architectural decisions, rite design |

## Integration with Rites

The interview spec artifact is designed to feed directly into:
- **10x-dev**: Replaces or supplements the PRD that Requirements Analyst produces
- **forge**: Provides the rite concept brief for Agent Designer
- **arch**: Feeds into the architecture analysis scope
- **Any rite**: The spec is a universal "here's what we decided" artifact

## Example

```bash
# Standard interview for a new feature
/interview "Add webhook support for event notifications"

# Deep interview for an architectural decision
/interview "Migrate from REST to gRPC" --depth=deep

# Shallow interview, custom output
/interview "Fix auth token refresh" --depth=shallow --output=docs/specs/auth-refresh.md
```
