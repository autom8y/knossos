---
domain: feat/explain-command
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/explain/**/*.go"
  - "./internal/concept/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.96
format_version: "1.0"
---

# Concept Documentation Browser (ari explain)

## Purpose and Design Rationale

Zero-dependency, offline concept glossary for knossos domain vocabulary. Discoverability for newcomers to mythology terminology. Harness-agnostic bridging via harness_term field. Project-aware contextualization with live counts. Concept package extracted from cmd/explain to resolve TENSION-015 (internal/search needed concept data without crossing CLI layer boundary).

## Conceptual Model

**16 embedded concepts** with YAML frontmatter (summary, see_also, aliases, harness_term). **Lookup chain:** exact match -> alias match -> Levenshtein suggestion (distance <=3 AND < len/2). **Context injection:** 10 of 16 concepts have live inspection functions reading project state. **Two output modes:** list (table of name+summary) and lookup (full description + context). **Display name:** "{name} ({harness_term})" when harness_term present.

## Implementation Map

`internal/concept/concept.go` (registry via init(), LookupConcept, AllConcepts, parseConcept, levenshtein, suggestConcept). 16 embedded .md files in concepts/. `internal/cmd/explain/` (explain.go, concepts.go re-export shim, context.go with 10 contextFuncs, models.go output types). Search integration: CollectConcepts in internal/search/collectors.go feeds ari ask index.

## Boundaries and Failure Modes

panic-on-init for malformed concept files (broken build detection). Context functions hardcode ClaudeChannel (Gemini-only projects get wrong counts). Sessions directory unreadable: graceful "0 sessions" message. Unrecognized concept: error with sorted list + optional suggestion. No NLP/fuzzy search (Levenshtein only). Six concepts without context functions.

## Knowledge Gaps

1. potnia.md references "pythia" in see_also but no pythia concept exists
2. contextSOS implementation is shallow ("directory exists" only)
3. No pythia concept file despite codebase references
