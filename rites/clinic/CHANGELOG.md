# Clinic Rite Changelog

## [1.0.0] - 2026-02-26

### Added

- Initial release of the clinic rite
- Five agents: pythia (orchestrator), triage-nurse, pathologist, diagnostician, attending
- Four-phase sequential workflow: intake -> examination -> diagnosis -> treatment
- Three back-routes: evidence_gap (max 3), diagnosis_insufficient (max 2), scope_expansion (max 1, user-confirmed)
- Evidence architecture: `.claude/wip/ERRORS/{investigation-slug}/` with index.yaml as shared coordination artifact
- Single complexity level: INVESTIGATION (emergent depth, no pre-classification)
- Cross-rite handoff artifacts: handoff-10x-dev.md, handoff-sre.md, handoff-debt-triage.md
- SRE escalation inbound protocol (incident-commander -> triage-nurse)
- Session resume via index.yaml status field
- clinic-ref skill with full methodology documentation
- `/clinic` quick-switch command

### Design Decisions

- Single complexity level (INVESTIGATION) rather than tiered complexity: production bugs do not admit pre-classification; depth emerges from what agents find
- Back-routes as first-class workflow elements, not error conditions: real debugging loops require returning to evidence gathering
- Pathologist writes all evidence to disk immediately, never to context: token economics require this for 80-120k+ token investigations
- Diagnostician loads evidence selectively via index.yaml, not all at once: same token economics constraint
- Attending produces handoff recommendations; user decides whether to act: clinic never auto-invokes downstream rites
