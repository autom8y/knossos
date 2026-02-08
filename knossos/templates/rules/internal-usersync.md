When modifying files in internal/usersync/:
- 3 resource types: agents (flat), mena (nested, dual-target), hooks (nested)
- 3 source states: knossos (unchanged), knossos-diverged (locally modified), user (user-created)
- User-created files are NEVER overwritten; diverged files require --force
- Mena routes to dual targets: .dro.md -> commands/, .lego.md -> skills/ (via DetectMenaType/RouteMenaFile)
- Collision detection: user resources skip if same name exists in any rite
- Checksums are SHA256 with "sha256:" prefix; used for change detection and divergence tracking
- Manifest tracks per-file source, checksum, install timestamp, mena type, and target
- Scope filtering: mena entries with scope=project are skipped in usersync pipeline (project-only via materialize)
