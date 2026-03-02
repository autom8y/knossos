---
summary: A named workflow configuration that defines active agents, skills, hooks, and settings.
see_also: [session, agent, knossos]
aliases: [rites]
---
A named workflow configuration (like 10x-dev, architect, ecosystem) that defines which agents, skills, hooks, and settings are active for a project. Rites are stored in `.knossos/rites/` and activated via `ari sync --rite <name>`. The active rite determines the full shape of the `.claude/` directory.
