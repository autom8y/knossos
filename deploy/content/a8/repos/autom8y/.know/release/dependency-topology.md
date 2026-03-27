---
domain: release/dependency-topology
generated_at: "2026-03-07T12:00:00Z"
expires_after: "30d"
source_scope:
  - "./.know/release/"
generator: dependency-resolver
source_hash: "c9d08d0"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

## Cross-SDK Dependency DAG

```
autom8y-config (1.2.0) --> autom8y-log (0.5.6) --> autom8y-telemetry (0.5.2)
      |                          |                         |
      |                          |-->  autom8y-ai (1.1.0)  |
      |                          |                         |
      |                          +--> autom8y-meta (0.2.1) |
      |                                                    |
      +--> autom8y-core (1.1.1) --> autom8y-auth (1.1.1) [optional: telemetry, log]
      |         |
      |         +--> autom8y-interop (1.0.0)
      |
      +--> autom8y-http (0.5.0) --> autom8y-stripe (1.3.1)
      |         |
      |         +--> autom8y-ai (1.1.0)
      |         +--> autom8y-meta (0.2.1)
      |         +--> autom8y-auth (1.1.1)
      |
      +--> autom8y-slack (0.2.0)
      +--> autom8y-cache (0.4.0) [optional extra]
      +--> autom8y-events (0.1.0) [new, zero consumers]
```

## Validated Publish Order (4-phase)

| Phase | Label | Packages | Parallel? |
|-------|-------|----------|-----------|
| 1 | foundation | autom8y-config | No |
| 2 | log-layer | autom8y-log | No |
| 3 | consumers-of-log-and-config | autom8y-telemetry, autom8y-ai, autom8y-meta | Yes |
| 4 | observability-consumers | autom8y-auth | No |

## Blast Radius

| Package | Severity | Direct SDK Consumers | Satellite Consumers |
|---------|----------|---------------------|---------------------|
| autom8y-config | CRITICAL | log, slack, telemetry, meta, cache, events | ads, asana, data, sms (all 4) |
| autom8y-log | HIGH | http, ai, telemetry, meta | ads, asana, data, sms (all 4) |
| autom8y-http | HIGH | ai, auth, meta, stripe, interop | ads, asana, data, sms (all 4) |
| autom8y-telemetry | HIGH | (auth optional) | ads, asana, data (optional), sms |
| autom8y-core | HIGH | auth, interop | asana, data, sms |
| autom8y-auth | MEDIUM | -- | ads, asana (optional), data (optional) |
| autom8y-cache | MEDIUM | -- | asana, data |
| autom8y-interop | MEDIUM | -- | ads, sms |
| autom8y-meta | LOW | -- | ads |
| autom8y-ai | LOW | -- | sms |
| autom8y-stripe | LOW | -- | (pull-payments only, no satellite) |
| autom8y-events | NONE | -- | (zero consumers) |

## Satellite Consumer Map (verified 2026-03-07)

| Satellite | Production SDK Dependencies | Optional SDK Dependencies |
|-----------|---------------------------|--------------------------|
| autom8y-ads | auth, config, http, interop, log, meta, telemetry (7) | -- |
| autom8y-asana | cache, config, core, http, log, telemetry (6) | auth |
| autom8y-data | cache, config, core, http, log (5) | auth, telemetry |
| autom8y-sms | ai, config, core, http, interop, log, telemetry (7) | -- |

## Satellite Inter-Dependencies

None. All 4 satellites are leaf nodes consuming only monorepo SDKs. Parallel execution is always safe.

## Version Drift Notes

| Consumer | SDK | Pin Floor | Published | Delta | Action |
|----------|-----|-----------|-----------|-------|--------|
| autom8y-sms | autom8y-ai | >=0.1.0 | 1.1.0 | major | Recommend tighten to >=1.1.0 |
| ads/asana/data | autom8y-auth | >=1.1.0 | 1.1.1 | patch | Optional tighten |
| ads | autom8y-meta | >=0.2.0 | 0.2.1 | patch | Optional tighten |
