---
paths:
  - "internal/usersync/**"
---

When modifying files in internal/usersync/:
- 3 resource types: agents (flat), mena (nested, dual-target), hooks (nested)
- Uses provenance.ProvenanceEntry types (unified schema, ADR-0026 Phase 4a)
- Owner model: knossos/user/untracked. Divergence is COMPUTED (checksum comparison), not stored
- User-created files are NEVER overwritten; diverged files require --force
- Mena routes to dual targets: .dro.md -> commands/, .lego.md -> skills/ (via DetectMenaType/RouteMenaFile)
- Single manifest: USER_PROVENANCE_MANIFEST.yaml (YAML, shared across all resource types)
- Manifest keys namespaced: agents/{name}, commands/{dir}/, skills/{dir}/, hooks/{path}
- Collision detection: manifest-based (reads rite PROVENANCE_MANIFEST.yaml), fallback to directory scan
- Orphan detection: auto-removes knossos-owned entries whose source is deleted; never removes user-owned
- Scope filtering: mena entries with scope=project are skipped in usersync pipeline (project-only via materialize)
