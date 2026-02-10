{{/* agent-routing section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START agent-routing -->
## Agent Routing

**Pythia** coordinates each rite's workflow — routing tasks to specialists, verifying phase gates, and managing handoffs. In orchestrated sessions, the main thread delegates to specialists via Task tool.

Every agent defines its authority via **Exousia** (jurisdiction contract):
- **You Decide**: Actions within the agent's autonomous authority
- **You Escalate**: Situations requiring Pythia or user input
- **You Do NOT Decide**: Boundaries the agent must never cross

Without a session, execute directly or use `/task`. Routing guidance: `/consult`.
<!-- KNOSSOS:END agent-routing -->
