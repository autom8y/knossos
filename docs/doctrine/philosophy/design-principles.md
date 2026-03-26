---
last_verified: 2026-03-26
---

# Design Principles — Quick Reference

> Compact lookup. Full narrative: [knossos-doctrine.md Section XI](knossos-doctrine.md#xi-design-principles)

| # | Principle | Key Implementation | Anti-Pattern |
|---|-----------|-------------------|--------------|
| 1 | The Clew Is Sacred | `internal/hook/clewcontract/`, `internal/session/` | Unrecorded decisions or actions |
| 2 | Honest Signals Over Comfortable Lies | `internal/sails/` | Ignoring GRAY signals, forcing WHITE |
| 3 | Mutation Through the Fates | `agents/moirai.md`, `internal/session/` | Direct `SESSION_CONTEXT.md` edits |
| 4 | Rites Over Teams | `rites/`, `internal/rite/`, `internal/materialize/` | N/A — structural principle |
| 5 | Heroes Are Mortal | Cognitive budget, Agent tool delegation | Loading all rites simultaneously |
| 6 | The Labyrinth Grows | `rites/*/manifest.yaml`, extensible hook system | N/A — evolutionary principle |
| 7 | Return Is the Victory | `ari session wrap`, quality gates | Abandoning sessions instead of wrapping |
| 8 | The Inscription Prepares | `knossos/templates/`, `internal/inscription/` | Manual edits in Knossos sections |
| 9 | The Palace Observes Xenia | `internal/materialize/`, `selective_write_test.go`, provenance `OwnerUser` entries | Overwriting user-owned content during materialization (see SCAR-005) |
| 10 | The Evans Principle | knossos-doctrine.md itself; every documentation file | Maintaining doctrine that describes aspirational rather than implemented reality |

All paths relative to repository root.

---

**See Also:**
- [knossos-doctrine.md](knossos-doctrine.md) (Section XI: full narrative with principle relationships and evolution notes)
- [mythology-concordance.md](mythology-concordance.md) (mapping myth to implementation)
- [../reference/INDEX.md](../reference/INDEX.md) (navigation hub)
