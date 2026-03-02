---
name: clinic-ref-anti-patterns
description: "Clinic investigation anti-patterns by agent. Use when: reviewing agent behavior, catching premature diagnosis, preventing context hoarding. Triggers: clinic anti-pattern, premature diagnosis, context hoarding, vague criteria."
---

# Clinic: Anti-Patterns

| Agent | Anti-Pattern | Correct Behavior |
|-------|-------------|------------------|
| triage-nurse | Premature diagnosis in intake report | Document symptoms, not theories |
| triage-nurse | Vague evidence collection plan | "Check the logs" → specify system, data type, time range |
| pathologist | Context hoarding (keeping evidence in context) | Write to E{NNN}.txt immediately |
| pathologist | Analytical summaries in index.yaml | Factual description only |
| diagnostician | Premature convergence | Check that ALL symptoms map to identified cause(s) |
| diagnostician | Re-running commands | Evidence is on disk — read it; if missing, trigger back-route |
| diagnostician | Loading all evidence files | Read index first, load selectively |
| attending | Re-diagnosing instead of using diagnosis.md | If insufficient, trigger back-route |
| attending | Vague acceptance criteria | Specific, testable conditions |
| attending | Ignoring monitoring gaps | If observability was absent, produce handoff-sre.md |
