# PRD: /consult Meta-Advisor Enhancement

## Overview

The `/consult` command should serve as the primary cognitive load absorber for the roster ecosystem. With multiple teams, skills, and agents (see `roster/rites/` for current inventory), users need a meta-advisor that handles routing complexity so they can describe intent in natural language and receive optimal team/skill/workflow paths without internalizing all concepts upfront.

## Background

The roster ecosystem has grown to significant complexity:

- **Teams**: Multiple rites in `roster/rites/` (dynamically discoverable)
- **Skills**: Distributed across managed and user locations (see `RITE_SKILL_MATRIX.md`)
- **Agents**: Specialized roles with specific invocation patterns (per team)

Current `/consult` implementation provides ecosystem guidance and team recommendations but has structural gaps:

1. **No explicit reference to `prompting` skill** - When `/consult` recommends a workflow, it should reference the `prompting` skill for exact invocation patterns, ensuring users get copy-paste commands rather than general guidance.

2. **Workflow knowledge not systematically leveraged** - The `10x-workflow` skill contains routing logic, quality gates, and phase transitions that `/consult` should draw from for accurate recommendations.

3. **Role as "cognitive load absorber" undocumented** - Neither CLAUDE.md nor the consult-ref skill explicitly position `/consult` as THE entry point for confused users.

4. **Team profiles fragmented** - Team information is scattered across `rites/*/TEAM.yaml`, skill descriptions, and ad-hoc knowledge rather than structured routing tables.

5. **No "confused? start here" positioning** - The ecosystem lacks a clear "I don't know where to go" entry point.

**Key Constraint**: `/consult` advises but does not execute. User receives routing guidance and then manually invokes the recommended team/workflow.

## User Stories

### Discovery and Navigation

- **US-1**: As a new roster user, I want to describe my goal in plain language and receive a specific team/command recommendation so that I can start working without reading all documentation.

- **US-2**: As a developer confused about which team to use, I want `/consult` to ask clarifying questions and narrow down options so that I get the right recommendation, not just a list of possibilities.

- **US-3**: As an experienced user exploring new capabilities, I want `/consult` to surface relevant skills I may not know about so that I discover functionality organically.

### Invocation Pattern Accuracy

- **US-4**: As a user receiving a `/consult` recommendation, I want exact copy-paste invocation patterns (from `prompting` skill) so that I can immediately execute without translation.

- **US-5**: As a user with a recommended workflow, I want to see the expected phase transitions and quality gates so that I understand what the full journey looks like before starting.

### Team Discovery

- **US-6**: As a user, I want to ask "what teams can help with X?" and receive a ranked list with rationale so that I understand my options.

- **US-7**: As a user switching between projects with different needs, I want `/consult` to recommend appropriate teams based on my described context (e.g., "production is down" vs. "planning next quarter") so that I use the right specialists.

- **US-8**: As a user, I want a dedicated `/teams` command (separate from `/consult`) to show all teams with their current status (active/dormant), core competencies, and quick-switch commands so that I have a complete reference. `/consult` is for advice; `/teams` is for inventory.

### Cognitive Load Reduction

- **US-9**: As an overwhelmed new user, I want to see a single starting point ("unsure? start with `/consult`") prominently in CLAUDE.md so that I know where to go when confused.

- **US-10**: As a user mid-workflow who gets stuck, I want to ask `/consult "I'm stuck on X"` and receive context-aware guidance that considers my active session so that I can unblock myself.

## Functional Requirements

### Must Have

#### Skill Integration

- **FR-1.1**: `/consult` MUST reference the `prompting` skill when providing invocation recommendations, extracting exact patterns rather than generating ad-hoc commands.

- **FR-1.2**: `/consult` MUST reference `10x-workflow` skill for workflow phase transitions, quality gates, and agent routing decisions.

- **FR-1.3**: Agent prompts invoking `/consult` MUST include explicit skill references (e.g., "reference `prompting` skill for invocation patterns") to ensure agents get proper instruction, not for user-facing verbosity.

#### Team Routing Knowledge

- **FR-2.1**: `/consult` MUST have structured access to all 11 team profiles including:
  - Team name and quick-switch command
  - Core competencies and use cases
  - Agent roster with primary artifacts
  - Complexity levels supported
  - Status (active/dormant)

- **FR-2.2**: `/consult` MUST support intent-to-team matching using natural language queries with confidence scoring.

- **FR-2.3**: `/consult` MUST offer alternatives when multiple teams could handle a request, with explicit tradeoff explanations.

- **FR-2.4**: Create a `team-discovery` skill that provides structured team routing data accessible by both `/consult` and other agents. This skill reads from `roster/rites/*/TEAM.yaml` and provides consistent team metadata.

#### Cognitive Load Absorber Positioning

- **FR-3.1**: CLAUDE.md Getting Help table MUST include `/consult` as routing option. Per PRD-claude-md-descriptive-architecture, positioning goes in table format, not prose.

- **FR-3.2**: The `consult-ref` skill MUST document its role as "cognitive load absorber" - the designated entry point for confused users.

- **FR-3.3**: `/consult` with no arguments MUST provide a minimal, scannable ecosystem overview (not exhaustive listing) with clear next-step suggestions.

#### Response Quality

- **FR-4.1**: `/consult` responses SHOULD follow the 4-part structure: Assessment, Recommendation, Command-Flow, Alternatives. (Guidance, not mandate—quality responses cover these areas but format flexibility allowed.)

- **FR-4.2**: Command-Flow sections MUST include phase-by-phase breakdown with expected artifacts at each step.

- **FR-4.3**: `/consult` MUST ask clarifying questions when intent is ambiguous rather than guessing.

### Should Have

#### Context Awareness

- **FR-S.1**: `/consult` SHOULD detect active session context and tailor recommendations accordingly (e.g., suggesting `/handoff` when mid-workflow).

- **FR-S.2**: `/consult` is stateless; session context is injected via hooks. Hooks provide active session state so `/consult` can tailor responses without maintaining internal memory.

- **FR-S.3**: `/consult --verbose` SHOULD provide expanded explanations of recommendations including skill source references.

#### Discovery Features

- **FR-S.4**: `/consult "what skills relate to X"` SHOULD surface relevant skills with brief descriptions and invocation patterns.

- **FR-S.5**: `/consult` SHOULD suggest related capabilities ("You might also be interested in...") after addressing primary query.

- **FR-S.6**: `/consult` SHOULD integrate with session history to provide "what worked last time you did X" recommendations. Sessions are already tracked and can be canonicalized for context availability.

### Could Have

- **FR-C.1**: `/consult --wizard` COULD provide interactive multi-step guidance for complex scenarios.

- **FR-C.2**: `/consult` COULD maintain a lightweight usage log to identify frequently asked patterns for ecosystem improvement.

## Non-Functional Requirements

- **NFR-1**: Response Latency - `/consult` responses MUST complete within 5 seconds for simple queries, 15 seconds for complex routing analysis.

- **NFR-2**: Accuracy - Team recommendations should match user intent without requiring clarification in most cases. (Qualitative measure for power user context; no formal metric.)

- **NFR-3**: Discoverability - New users MUST encounter `/consult` positioning within first 3 interactions with ecosystem.

- **NFR-4**: Maintainability - Team routing knowledge MUST be updatable without modifying `/consult` core logic (data-driven routing).

- **NFR-5**: Consistency - Invocation patterns provided by `/consult` MUST match current `prompting` skill content exactly.

## Edge Cases

| Case | Expected Behavior |
|------|------------------|
| Query matches multiple teams equally | Present top 3 with explicit tradeoff analysis, ask user to specify constraint |
| Query uses unknown terminology | Ask clarifying question, suggest possible interpretations |
| User requests dormant team capability | Explain dormant status, offer alternative from active rites, note future availability |
| Query is too vague ("help me code") | Narrow down with 2-3 clarifying questions before recommending |
| User is mid-session and asks for different team | Warn about active session, offer `/park` + team switch or session continuation |
| Query matches skill but not team | Recommend skill directly with invocation pattern, note that no team switch needed |
| No teams match the query | Acknowledge gap, suggest closest alternatives, offer to escalate to ecosystem maintainer |
| User asks about deprecated command | Explain deprecation, provide replacement command with updated pattern |
| Conflicting constraints in query | Surface conflict explicitly, ask user to prioritize |
| Query in non-English | Process query, respond in same language if possible, or acknowledge limitation |

## Success Criteria

- [ ] `/consult` response includes skill references for invocation patterns
- [ ] `/consult` response includes workflow skill reference for phase/gate decisions
- [ ] CLAUDE.md Getting Help table includes `/consult` routing
- [ ] `consult-ref` skill includes "cognitive load absorber" role documentation
- [ ] All teams discoverable via `/teams` command with accurate profile information
- [ ] `team-discovery` skill created and accessible by agents
- [ ] `/consult` with no arguments displays scannable overview with clear next steps
- [ ] Intent matching works for representative queries across common use cases
- [ ] User can complete "describe intent -> receive recommendation -> invoke successfully" flow

**Test Plan Note**: Specific test cases (e.g., "/consult 'I want to build a feature'" → 10x-dev-pack) should be enumerated during QA phase, not in PRD success criteria.

## Dependencies

| Dependency | Type | Owner | Status |
|------------|------|-------|--------|
| `prompting` skill content | Internal | roster | Complete - patterns exist |
| `10x-workflow` skill content | Internal | roster | Complete - routing logic exists |
| Team profiles in `rites/*/TEAM.yaml` | Internal | roster | Complete - 11 teams defined |
| `consult-ref` skill | Internal | roster | Exists - needs enhancement |
| CLAUDE.md structure | Internal | roster | Exists - needs positioning update |

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Skill content drift causes stale `/consult` recommendations | Medium | High | Add validation that `/consult` references align with current skill content |
| Cognitive load absorber positioning not noticed | Medium | Medium | Add positioning in multiple locations (CLAUDE.md, `/help`, session start) |
| Team profiles incomplete or inconsistent | Low | Medium | Validate all 11 team profiles as prerequisite |
| Performance degradation from skill lookups | Low | Low | Cache skill references, lazy-load detailed patterns |
| Recommendations become outdated as ecosystem evolves | Medium | High | Establish maintenance protocol for `/consult` knowledge base |

## Out of Scope

- **Active delegation** - `/consult` advises only; user must invoke recommended team/workflow manually
- **Team count reduction** - All 11 teams remain; dormant teams are future value and `/consult` handles discovery
- **Underlying complexity simplification** - Goal is navigability, not reduction; power users retain full access
- **Automated skill invocation** - `/consult` surfaces skill patterns but does not invoke Skill tool itself
- **Cross-session learning** - Recommendations based on current session context only, not historical patterns
- **Natural language workflow execution** - User describes intent, receives command, must execute command themselves

## Open Questions

*All questions resolved during stakeholder discussion:*

1. ~~Should `/consult` invoke skills directly or just reference them?~~ **Resolved**: Reference only; user invokes.
2. ~~How should dormant teams appear in recommendations?~~ **Resolved**: Listed with dormant status, alternatives from active rites offered.
3. ~~Where exactly in CLAUDE.md should cognitive load absorber positioning go?~~ **Resolved**: Quick Start or Getting Help section with prominent visibility.

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-consult-meta-advisor.md` | Created |
