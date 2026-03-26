---
domain: feat/explain-command
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/explain/**/*.go"
  - "./internal/cmd/explain/concepts/*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Concept Documentation Browser (ari explain)

## Purpose and Design Rationale

`ari explain` solves the onboarding and vocabulary problem in a domain-heavy framework. Knossos introduces dense mythological vocabulary (rite, dromena, legomena, mena, inscription, sails, tribute, clew, naxos) that has no intuitive meaning to new users.

**Design decisions**: Embedded concept files via `//go:embed concepts/*.md`. Project-optional (`needsProject=false`). Markdown frontmatter format for concepts. Two output modes (list all vs single lookup). `cc_term` bridging field maps knossos terms to CC terms.

## Conceptual Model

### Lookup Chain

1. Normalize: `strings.ToLower(strings.TrimSpace(input))`
2. Exact match in registry
3. Alias match (e.g., `skills` → `legomena`, `commands` → `dromena`)
4. Levenshtein suggestion (threshold: dist <= 3 AND dist < len(input)/2)
5. Error with "Available concepts:" list

### 13 Embedded Concepts

`agent`, `dromena`, `inscription`, `knossos`, `know`, `ledge`, `legomena`, `mena`, `rite`, `sails`, `session`, `sos`, `tribute`

### Project Context Injection

When a project root is available, dynamic context is appended (active rite, session count, agent count, mena count). 10 context functions in `/Users/tomtenuta/Code/knossos/internal/cmd/explain/context.go`.

## Implementation Map

5 files in `/Users/tomtenuta/Code/knossos/internal/cmd/explain/`: `explain.go`, `concepts.go`, `models.go`, `context.go`, `explain_test.go` (53 test cases).

### Key Types

- `ConceptEntry` — parsed concept with name, summary, description, seeAlso, aliases, ccTerm
- `ConceptOutput` / `ConceptListOutput` — output types implementing `output.Textable` / `output.Tabular`

## Boundaries and Failure Modes

- `init()` panics on embedded file corruption (intentional — compile-time error)
- Context injection is silent-fail (returns empty string on any filesystem error)
- Levenshtein threshold: 2-character inputs never get suggestions
- `readSessionStatus` in context.go re-implements minimal frontmatter parsing (parallel to `internal/session`)

## Knowledge Gaps

1. No ADR for explain command design.
2. `context.go` is fully untested (18 functions, 0% coverage) — documented as DEBT-103.
3. Error format inconsistency on failed lookup (stdlib `fmt.Errorf` vs `*errors.Error`).
