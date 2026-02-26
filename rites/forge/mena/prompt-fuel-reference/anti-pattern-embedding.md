# Anti-Pattern Embedding

## What AP-NN Entries Are

The HANDOFF's `## Prompt Anti-Pattern Catalog` contains AP-NN entries -- concrete wrong behaviors derived from real scars and golden path violations. Each entry specifies:

- **Source**: The SCAR-NNN or GOLD-NNN it was derived from
- **Anti-Pattern**: The specific wrong behavior
- **Rule**: The correct behavior

## Embedding Format

Translate each applicable AP-NN into a behavioral constraint using this template:

```markdown
- **DO NOT** {anti-pattern behavior}. **INSTEAD**: {correct behavior}. [Source: {SCAR/GOLD-NNN}]
```

**Example**:
```markdown
- **DO NOT** use os.RemoveAll on directories that may contain user content. **INSTEAD**: Build a managed-set of expected filenames, iterate directory, remove only managed files. [Source: SCAR-003]
- **DO NOT** read hook data from environment variables. **INSTEAD**: Use ParseStdin() for operational data (tool name, session ID). Only CLAUDE_PROJECT_DIR etc. are env vars. [Source: SCAR-005]
```

## Placement

Position the `## Anti-Patterns` section after `## Domain Knowledge` and before `## Protocol` or `## How You Work`.

```
## Domain Knowledge        <-- CRITICAL tier items
## Anti-Patterns            <-- AP-NN constraints (you are here)
## How You Work / Protocol  <-- agent methodology
```

## Grouping Rules

- Group related anti-patterns under a single `## Anti-Patterns` heading
- Order by severity: data-loss risks first, then correctness, then style
- If an agent has more than 7 anti-patterns, split into sub-groups with `###` headings (e.g., `### File Safety`, `### Hook Contracts`)
- Merge AP entries that share the same root cause into one constraint

## Source Traceability

Every embedded anti-pattern MUST reference its source:
- `[Source: SCAR-NNN]` for scar-derived anti-patterns
- `[Source: GOLD-NNN]` for golden path violations
- `[Source: SCAR-NNN + GOLD-NNN]` when multiple sources contribute

Never embed an anti-pattern without a source reference. Unsourced constraints are unverifiable and become stale without an audit trail.

## Filtering by Agent

Not every AP-NN applies to every agent. Filter by relevance:

- **Does this agent perform the action described in the anti-pattern?** If not, skip.
- **Could this agent accidentally trigger the anti-pattern?** If yes, include.
- **Is this anti-pattern covered by an EX-NN override already in Exousia?** If yes, include in Anti-Patterns section as reinforcement -- redundancy between Exousia and Anti-Patterns is intentional for high-severity items.
