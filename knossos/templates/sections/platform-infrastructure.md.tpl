{{/* platform-infrastructure section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START platform-infrastructure -->
## Platform

**Entry**: `/go` — cold-start dispatcher. Detects session state, resumes parked work, or routes new tasks.

**Sessions**: Managed by Moirai agent via `/start`, `/park`, `/continue`, `/wrap`. Moirai loads Fate skills (Clotho/Lachesis/Atropos) for progressive context. Mutate `*_CONTEXT.md` only via `Task(moirai, "...")`.

**Hooks**: Auto-inject session context on start; autopark on stop. CLI reference: `ari --help`.
<!-- KNOSSOS:END platform-infrastructure -->
