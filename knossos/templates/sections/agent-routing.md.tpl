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

### Throughline Resume Protocol

The main thread MAY track subagent IDs for throughline agents (Pythia, Moirai) and pass `resume: {agentId}` on subsequent Task calls. This gives the agent full history of its prior consultations within the workflow.

- Agent IDs are valid only within the current CC session
- Clear stored IDs on rite switch or session wrap
- If resume fails (invalid ID, session changed), fall back to fresh invocation
- Resume is opportunistic -- orchestrated workflows function correctly without it
<!-- KNOSSOS:END agent-routing -->
