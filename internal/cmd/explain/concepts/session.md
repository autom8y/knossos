---
summary: A tracked unit of work with lifecycle state (ACTIVE, PARKED, ARCHIVED).
see_also: [sos, sails, tribute]
aliases: [sessions]
---
A tracked unit of work with lifecycle state (ACTIVE, PARKED, ARCHIVED). Sessions track initiative, complexity, phase, and audit events. Stored in `.sos/sessions/`, each session is a directory containing SESSION_CONTEXT.md frontmatter and an events.jsonl audit log.
