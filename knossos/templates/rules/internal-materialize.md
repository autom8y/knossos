---
paths:
  - "internal/materialize/**"
---

When modifying files in internal/materialize/:
- Idempotency invariant: running materialize twice must produce identical output
- 4-tier resolution order: rite > dependency > shared > user
- User content is NEVER destroyed (satellite regions, user-agents, user-hooks)
- writeIfChanged() prevents unnecessary file watcher triggers
- Dual pipeline: materialize (project, destructive) vs usersync (user, additive)
- MCP merge is union: add/update rite servers, preserve existing satellite servers
- Mena projection strips .dro.md/.lego.md extensions and routes to commands/ or skills/
- Provenance collector is threaded through all pipeline stages; record after each writeIfChanged
- Volatile infrastructure files (KNOSSOS_MANIFEST.yaml, sync/state.json, ACTIVE_RITE) are NOT tracked in provenance
- Orphan detection checks provenance manifest first, falls back to legacy rite-manifest-membership
