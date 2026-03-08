# Thermia Rite Changelog

## [1.0.0] - 2026-02-27

### Added

- Initial release of the thermia rite
- Five agents: potnia (orchestrator), heat-mapper, systems-thermodynamicist, capacity-engineer, thermal-monitor
- Four-phase sequential workflow: assessment -> architecture -> specification -> validation
- Three complexity levels: QUICK (2 phases), STANDARD (4 phases), DEEP (4 phases at extended depth)
- Two back-routes: assessment_gap (architecture -> assessment, max 1), design_inconsistency (validation -> specification, max 1)
- 6-gate decision framework in heat-mapper for structured caching necessity evaluation
- Theoretical grounding: CAP theorem, Facebook TAO/Memcache NSDI papers, Lamport consistency hierarchy (thermodynamicist); working set theory, ARC, TinyLFU, XFetch, LIRS (capacity-engineer)
- Artifact chain: thermal-assessment -> cache-architecture -> capacity-specification -> observability-plan at `.sos/wip/thermia/`
- Cross-rite outbound routing from thermal-monitor: 10x-dev (primary), clinic, sre, arch, hygiene, debt-triage (secondary)
- `/thermia` quick-switch command (thermia-switch dromena)
- Consultative protocol: exhaust alternatives first, probe assumptions, challenge shallow answers, derive every number
- `thermia-ref` skill stub (post-launch scope; agents degrade gracefully without it)

### Design Decisions

- Thermal management theme: heat-mapper (hotspot assessment), systems-thermodynamicist (cooling architecture), capacity-engineer (system sizing and tuning), thermal-monitor (instrumentation and observability)
- "Exhaust alternatives first" mandate enforced at every level: 6-gate framework, heat-mapper's CRITICAL section, frontmatter examples, contract.must_not constraints
- Three complexity levels rather than emergent depth: cache architecture problems can be pre-classified (quick triage vs. full design vs. post-mortem) unlike production debugging
- Capacity-engineer requires derived numbers, not gut-feel: "2GB is not a capacity plan" is a first-class anti-pattern
- Thermal-monitor uses miss rate, not hit rate, as the primary health signal: hit rate is a vanity metric; miss rate is actionable
- Cross-rite notes absent from mid-workflow agents (heat-mapper, thermodynamicist, capacity-engineer) by design: outbound routing is Potnia and thermal-monitor responsibility
