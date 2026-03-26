# Clew Alerting Configuration

9 alerts from the Thermia cache consultation, all with post-deploy suppression where applicable.

## Alert Summary

| # | Alert | Condition | Severity | Post-Deploy Suppression |
|---|-------|-----------|----------|------------------------|
| 1 | Miss rate high | >40% sustained 10m | WARNING | 15m |
| 2 | Miss rate critical | >80% sustained 5m | CRITICAL | 10m |
| 3 | Miss rate increasing | +5pp/min over 15m | WARNING | 20m |
| 4 | Active threads high | >200 sustained 10m | WARNING | None |
| 5 | Eviction goroutine stalled | 0 evictions + >50 threads + >2h uptime | WARNING | 2h |
| 6 | Build startup timeout | P95 >75s | WARNING | None |
| 7 | Build knowledge failures | >10 stale fallbacks | WARNING | None |
| 8 | Content domains missing | >0 | WARNING | None |
| 9 | Stage 4 cache miss nonzero | >0 (Tier 1 only) | INFO | None |

## Metrics Transport

Metrics are emitted as CloudWatch Embedded Metric Format (EMF) via structured slog JSON logging.
CloudWatch automatically extracts metrics from JSON log lines containing the `_aws` metadata block
when using the awslogs driver on ECS/Fargate.

## Post-Deploy Suppression

Alerts 1-3 and 5 include post-deploy suppression to prevent false alarms when the
ConversationManager starts with an empty cache. Suppression is keyed off the
`clew_startup_timestamp` gauge metric.

Implementation options:
1. CloudWatch Composite Alarms with a "recently deployed" alarm as suppression input
2. CloudWatch Math Expressions comparing current time vs startup_timestamp
3. PagerDuty/OpsGenie maintenance windows triggered by the deploy pipeline

## Files

- `cloudwatch-alarms.json` -- Alarm definitions (for `aws cloudwatch put-metric-alarm` or Terraform)

## Runbooks

See the Thermia consultation Part 3 for full runbook text:
`.ledge/spikes/THERMIA-clew-v2-cache-consultation.md`
