# Cross-Rite Handoff Protocol

How work transitions between rites. The outgoing rite produces a handoff artifact; the incoming rite's Potnia loads it as context. Without this protocol, rite transitions lose continuity.

## Schema

```yaml
cross_rite_handoffs:
  - from_rite: <outgoing-rite>
    to_rite: <incoming-rite>
    artifact:
      path: ".ledge/spikes/<initiative-slug>-<from-rite>-handoff.md"
      produces: "<what the outgoing rite writes into the artifact>"
      consumes: "<what the incoming rite expects to find>"
    verification:
      - "<condition the incoming Potnia checks before starting>"
```

## Guidance

- **Artifacts go to `.ledge/`.** Handoff artifacts are durable work products, not ephemeral scratch. Use `.ledge/spikes/` for exploration handoffs, `.ledge/decisions/` for ADR-backed handoffs.
- **Naming convention: `{initiative}-{from-rite}-handoff.md`.** This makes handoffs discoverable by initiative slug and traceable to the producing rite.
- **`produces` and `consumes` must align.** The outgoing rite writes specific sections or data; the incoming rite expects to find them. Misalignment means the incoming Potnia starts with incomplete context.
- **Verification is the incoming rite's entry check.** Before starting sprint work, the incoming Potnia verifies the handoff artifact exists and contains what `consumes` specifies. This is a lightweight pre-flight, not a full checkpoint.
- **One artifact per transition.** If a rite transition requires multiple documents, consolidate into one handoff artifact with sections. Multiple artifacts create discovery problems.

## Anti-Patterns

- **Handoff by conversation history.** Rite transitions reset context. If the incoming Potnia needs information from the outgoing rite, it must be in a file, not in conversation memory.
- **Omitting `consumes`.** Without explicit expectations, the incoming Potnia cannot verify the handoff is complete. This creates silent context gaps.
