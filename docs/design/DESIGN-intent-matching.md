# Design: Intent-Matching Algorithm for Team Discovery

> Context Design artifact for Dynamic Team Discovery Integration
> Session: session-20260104-022657-e336a09d | Task: 003

## Overview

This document specifies the algorithm for matching user queries to team capabilities with confidence scoring. The design enables `/consult` to dynamically recommend teams based on parsed intent rather than hardcoded team lists.

### Problem Statement

Currently, `/consult` uses hardcoded team lists in `consult.md` and `consult-ref/SKILL.md`. This approach:
- Requires manual updates when teams are added/removed
- Cannot intelligently match user intent to team capabilities
- Lacks confidence scoring for ambiguous queries
- Does not leverage the rich metadata in `orchestrator.yaml` and `README.md`

### Design Goals

1. **Generative**: Auto-discover teams from filesystem (`teams/*/orchestrator.yaml`)
2. **Intelligent**: Match natural language intent to team capabilities
3. **Transparent**: Provide confidence scores with rationale
4. **Extensible**: Support new signal types without algorithm changes
5. **Performant**: Cache capability index for session duration

---

## Intent Extraction

### Signal Types

The algorithm extracts four categories of signals from user queries:

| Signal Type | Description | Weight | Examples |
|-------------|-------------|--------|----------|
| **Trigger signals** | Direct matches to orchestrator.yaml `frontmatter.description` triggers | 0.40 | "build a feature", "security review", "explore technology" |
| **Domain signals** | Matches to `team.domain` field | 0.25 | "documentation", "code quality", "analytics" |
| **Problem signals** | Inferred from problem domain keywords | 0.20 | "slow API" -> sre/10x, "vulnerability" -> security |
| **Exclusion signals** | Matches to README.md "Not for" sections | -0.50 | "one-off script" excludes 10x-dev-pack |

### Extraction Pipeline

```
user_query
    |
    v
+-------------------+
|  1. Tokenize      |  Split on whitespace, punctuation
+-------------------+
    |
    v
+-------------------+
|  2. Normalize     |  Lowercase, stem, remove stopwords
+-------------------+
    |
    v
+-------------------+
|  3. Extract       |  Identify signal-bearing tokens
|     Signals       |  Map to signal types
+-------------------+
    |
    v
+-------------------+
|  4. Weight        |  Apply signal-type weights
|     Signals       |  Handle negation/exclusion
+-------------------+
    |
    v
extracted_signals: SignalSet
```

### Signal Extraction Rules

#### Trigger Signal Extraction

Parse orchestrator.yaml `frontmatter.description` for trigger phrases:

```yaml
# Example from 10x-dev-pack/orchestrator.yaml
frontmatter:
  description: |
    Routes development work through requirements, design, implementation, and validation phases.
    Use when: building features or systems requires full lifecycle coordination.
    Triggers: coordinate, orchestrate, development workflow, feature development, implementation planning.
```

Extraction pattern:
1. Parse text after "Triggers:" as comma-separated list
2. Also extract phrases after "Use when:"
3. Store as lowercase, trimmed strings

#### Domain Signal Extraction

Direct match against `team.domain` field:

```yaml
# Each team has a domain
team:
  name: security-pack
  domain: security assessment
```

Domain mapping:
- Exact match: full weight (0.25)
- Partial match (word overlap): half weight (0.125)
- Synonym match: three-quarter weight (0.1875)

#### Problem Signal Extraction

Map common problem patterns to team domains:

| Problem Pattern | Primary Team | Secondary Team |
|-----------------|--------------|----------------|
| `slow`, `performance`, `latency` | sre-pack | 10x-dev-pack |
| `bug`, `error`, `broken`, `fix` | 10x-dev-pack | - |
| `vulnerability`, `CVE`, `auth`, `crypto` | security-pack | - |
| `documentation`, `docs`, `readme` | doc-team-pack | - |
| `refactor`, `cleanup`, `smells` | hygiene-pack | - |
| `debt`, `legacy`, `cruft` | debt-triage-pack | hygiene-pack |
| `analytics`, `metrics`, `tracking` | intelligence-pack | - |
| `research`, `explore`, `prototype` | rnd-pack | - |
| `market`, `strategy`, `business` | strategy-pack | - |
| `reliability`, `incident`, `SLO` | sre-pack | - |
| `sync`, `satellite`, `CEM`, `roster` | ecosystem-pack | - |
| `agent`, `team creation`, `forge` | forge-pack | - |

#### Exclusion Signal Extraction

Parse README.md "Not for" sections:

```markdown
# From 10x-dev-pack/README.md
**Not for**: Documentation work, infrastructure automation, one-off scripts without testing requirements
```

Extraction:
1. Find line containing "Not for" or "**Not for**"
2. Parse comma-separated exclusion phrases
3. Apply negative weight when query matches exclusion

---

## Capability Index

### Data Structure

The capability index is the normalized representation of all team capabilities, built at runtime from filesystem sources.

```yaml
capability_index:
  10x-dev-pack:
    display_name: "10x Dev Pack"
    domain: "software development"
    command: "/10x"
    triggers:
      - "build a feature"
      - "coordinate development"
      - "implementation planning"
      - "PRD"
      - "technical design"
    artifacts:
      - "PRD"
      - "TDD"
      - "ADR"
      - "test plan"
    not_for:
      - "documentation work"
      - "infrastructure automation"
      - "one-off scripts"
    complexity_levels:
      - "SCRIPT"
      - "MODULE"
      - "SERVICE"
      - "PLATFORM"
    agents:
      - "requirements-analyst"
      - "architect"
      - "principal-engineer"
      - "qa-adversary"
    related_teams:
      - "doc-team-pack"
      - "rnd-pack"

  ecosystem-pack:
    display_name: "Ecosystem Pack"
    domain: "ecosystem infrastructure"
    command: "/ecosystem"
    triggers:
      - "satellite sync"
      - "CEM"
      - "roster"
      - "hook registration"
      - "skill registration"
    artifacts:
      - "gap analysis"
      - "context design"
      - "migration runbook"
      - "compatibility report"
    not_for:
      - "application code"
      - "team-specific workflows"
    complexity_levels:
      - "PATCH"
      - "MODULE"
      - "SYSTEM"
      - "MIGRATION"
    agents:
      - "ecosystem-analyst"
      - "context-architect"
      - "integration-engineer"
      - "documentation-engineer"
      - "compatibility-tester"
    related_teams:
      - "hygiene-pack"

  # ... additional teams follow same structure
```

### Schema Definition

```yaml
# capability_index.schema.yaml
type: object
additionalProperties:
  type: object
  required:
    - display_name
    - domain
    - command
    - triggers
  properties:
    display_name:
      type: string
      description: Human-readable team name
    domain:
      type: string
      description: Primary domain from orchestrator.yaml team.domain
    command:
      type: string
      pattern: "^/[a-z0-9-]+$"
      description: Quick-switch command
    triggers:
      type: array
      items:
        type: string
      minItems: 1
      description: Phrases that trigger this team recommendation
    artifacts:
      type: array
      items:
        type: string
      description: Artifact types this team produces
    not_for:
      type: array
      items:
        type: string
      description: Exclusion phrases from README.md
    complexity_levels:
      type: array
      items:
        type: string
      description: Supported complexity levels
    agents:
      type: array
      items:
        type: string
      description: Agent names in this team
    related_teams:
      type: array
      items:
        type: string
      description: Teams commonly handed off to/from
```

### Population Strategy

The capability index is populated from three sources in priority order:

1. **orchestrator.yaml** (required, primary source)
   - `team.name`, `team.domain`, `team.color`
   - `frontmatter.description` -> extract triggers
   - `routing` keys -> agent names
   - `handoff_criteria` keys -> artifacts
   - `skills` -> related capabilities

2. **README.md** (optional, enrichment)
   - "Triggers:" section -> additional trigger phrases
   - "Not for:" section -> exclusion phrases
   - "Complexity Levels:" section -> complexity values
   - "Related Teams:" section -> handoff targets

3. **agents/*.md** (optional, agent roster)
   - Filename without extension -> agent name
   - Frontmatter `description` -> agent capabilities

### Build Algorithm

```
function build_capability_index():
    index = {}
    teams_dir = "$ROSTER_HOME/teams/"

    for team_dir in glob(teams_dir + "*-pack"):
        team_name = basename(team_dir)

        # Required: orchestrator.yaml
        orch_path = team_dir + "/orchestrator.yaml"
        if not exists(orch_path):
            log_warning("Skipping " + team_name + ": no orchestrator.yaml")
            continue

        orch = parse_yaml(orch_path)

        # Build team entry
        entry = {
            display_name: titlecase(team_name.replace("-pack", " Pack")),
            domain: orch.team.domain,
            command: "/" + team_name.replace("-pack", ""),
            triggers: extract_triggers(orch.frontmatter.description),
            artifacts: list(orch.handoff_criteria.keys()),
            complexity_levels: [],
            agents: list(orch.routing.keys()),
            related_teams: [],
            not_for: []
        }

        # Optional: README.md enrichment
        readme_path = team_dir + "/README.md"
        if exists(readme_path):
            readme = read_file(readme_path)
            entry.not_for = extract_not_for(readme)
            entry.complexity_levels = extract_complexity(readme)
            entry.related_teams = extract_related_teams(readme)
            entry.triggers += extract_readme_triggers(readme)

        # Optional: agent roster
        agent_files = glob(team_dir + "/agents/*.md")
        for agent_file in agent_files:
            agent_name = basename(agent_file).replace(".md", "")
            if agent_name not in entry.agents:
                entry.agents.append(agent_name)

        index[team_name] = entry

    return index
```

### Caching Strategy

The capability index is expensive to build (filesystem I/O, YAML parsing, regex extraction). Cache strategy:

1. **Build on first invocation**: When `/consult` or `team-discovery` is first called in a session
2. **Store in memory**: Keep index in session context (not persisted to disk)
3. **Invalidate on team change**: If `ACTIVE_RITE` changes, rebuild index
4. **TTL fallback**: Rebuild after 30 minutes of inactivity (stale data protection)

```
function get_or_build_index():
    cache_key = "capability_index"
    cache_ttl = 1800  # 30 minutes

    if cache_exists(cache_key) and cache_age(cache_key) < cache_ttl:
        return cache_get(cache_key)

    index = build_capability_index()
    cache_set(cache_key, index, ttl=cache_ttl)
    return index
```

---

## Scoring Algorithm

### Overview

The scoring algorithm computes a confidence score for each team based on extracted signals. Higher scores indicate stronger matches.

### Scoring Formula

```
confidence(team, signals) = clamp(0, 1, base_score + bonuses - penalties)

where:
  base_score = trigger_score * 0.40 + domain_score * 0.25 + problem_score * 0.20
  bonuses = complexity_bonus * 0.10 + recency_bonus * 0.05
  penalties = exclusion_penalty * 0.50
```

### Factor Calculations

#### Trigger Score (weight: 0.40)

```
function trigger_score(signals, team):
    team_triggers = lowercase(team.triggers)
    query_tokens = signals.normalized_tokens

    exact_matches = 0
    partial_matches = 0

    for trigger in team_triggers:
        trigger_tokens = tokenize(trigger)
        if all(t in query_tokens for t in trigger_tokens):
            exact_matches += 1
        elif any(t in query_tokens for t in trigger_tokens):
            partial_matches += 1

    if len(team_triggers) == 0:
        return 0

    # Exact matches worth 1.0, partial worth 0.3
    raw_score = (exact_matches * 1.0 + partial_matches * 0.3) / len(team_triggers)
    return min(1.0, raw_score)
```

#### Domain Score (weight: 0.25)

```
function domain_score(signals, team):
    query = signals.original_query.lower()
    domain = team.domain.lower()
    domain_words = domain.split()

    # Exact domain mention
    if domain in query:
        return 1.0

    # Word overlap
    overlap = sum(1 for word in domain_words if word in query)
    if overlap > 0:
        return 0.5 * (overlap / len(domain_words))

    # Synonym check (predefined synonym map)
    synonyms = get_domain_synonyms(domain)
    for synonym in synonyms:
        if synonym in query:
            return 0.75

    return 0.0
```

Domain synonym map:
```yaml
domain_synonyms:
  "software development": ["coding", "programming", "implementation", "building"]
  "documentation lifecycle": ["docs", "writing", "technical writing"]
  "code quality": ["refactoring", "cleanup", "code review"]
  "technical debt management": ["debt", "legacy", "cruft"]
  "site reliability engineering": ["ops", "operations", "infrastructure", "devops"]
  "security assessment": ["security", "vulnerabilities", "compliance", "audit"]
  "product analytics": ["analytics", "metrics", "data", "insights"]
  "technology exploration": ["research", "R&D", "exploration", "prototyping"]
  "business strategy": ["strategy", "market", "business"]
  "ecosystem infrastructure": ["CEM", "roster", "sync", "satellite"]
  "agent team creation": ["forge", "new team", "agent creation"]
```

#### Problem Score (weight: 0.20)

```
function problem_score(signals, team):
    # Problem patterns from signal extraction
    problem_map = get_problem_team_mapping()

    matched_problems = []
    for pattern, (primary, secondary) in problem_map.items():
        if pattern_matches(signals.original_query, pattern):
            matched_problems.append((primary, secondary))

    if not matched_problems:
        return 0.0

    # Score based on team match position
    primary_matches = sum(1 for p, s in matched_problems if p == team.name)
    secondary_matches = sum(1 for p, s in matched_problems if s == team.name)

    # Primary match: full credit, secondary: half credit
    return min(1.0, (primary_matches * 1.0 + secondary_matches * 0.5) / len(matched_problems))
```

#### Complexity Bonus (weight: 0.10)

When session context includes complexity level, boost teams that support it:

```
function complexity_bonus(session_context, team):
    if not session_context or not session_context.complexity:
        return 0.0

    session_complexity = session_context.complexity

    if session_complexity in team.complexity_levels:
        return 1.0

    # Partial credit for adjacent complexity levels
    complexity_order = ["SCRIPT", "MODULE", "SERVICE", "PLATFORM"]  # 10x example
    if session_complexity in complexity_order and team.complexity_levels:
        session_idx = complexity_order.index(session_complexity)
        for tc in team.complexity_levels:
            if tc in complexity_order:
                tc_idx = complexity_order.index(tc)
                if abs(session_idx - tc_idx) == 1:
                    return 0.5

    return 0.0
```

#### Recency Bonus (weight: 0.05)

Slight preference for recently used teams (session continuity):

```
function recency_bonus(session_context, team):
    if not session_context or not session_context.team_history:
        return 0.0

    history = session_context.team_history  # List of recently used team names

    if team.name == history[0]:  # Most recent
        return 1.0
    elif team.name in history[:3]:  # Top 3
        return 0.5
    elif team.name in history:
        return 0.2

    return 0.0
```

#### Exclusion Penalty (weight: 0.50)

Strong negative signal when query matches "Not for" phrases:

```
function exclusion_penalty(signals, team):
    if not team.not_for:
        return 0.0

    query = signals.original_query.lower()

    for exclusion in team.not_for:
        exclusion_lower = exclusion.lower()
        exclusion_tokens = tokenize(exclusion_lower)

        # Exact phrase match
        if exclusion_lower in query:
            return 1.0

        # Token overlap (2+ tokens)
        overlap = sum(1 for t in exclusion_tokens if t in query)
        if overlap >= 2:
            return 0.7

    return 0.0
```

### Confidence Thresholds

| Confidence Range | Interpretation | Response Behavior |
|------------------|----------------|-------------------|
| >= 0.80 | High confidence | Single recommendation with strong rationale |
| 0.50 - 0.79 | Medium confidence | Top 2-3 recommendations with explanations |
| 0.30 - 0.49 | Low confidence | Top 3 recommendations, suggest clarification |
| < 0.30 | No match | Ask clarifying questions, offer guided exploration |

### Response Templates by Threshold

#### High Confidence (>= 0.80)

```
=== Assessment ===
Goal: {extracted_goal}
Domain: {matched_domain}
Confidence: HIGH ({score:.0%})

=== Recommendation ===
Team: {team_name}
Command: {team_command}

Rationale: {trigger_matches_explanation}

=== Command-Flow ===
1. {first_command}
2. {second_command}
...
```

#### Medium Confidence (0.50 - 0.79)

```
=== Assessment ===
Goal: {extracted_goal}
Confidence: MEDIUM - Multiple teams may fit

=== Recommendations ===

1. {team_1_name} ({score_1:.0%})
   Best for: {team_1_domain}
   Why: {team_1_rationale}

2. {team_2_name} ({score_2:.0%})
   Best for: {team_2_domain}
   Why: {team_2_rationale}

[3. optional third recommendation]

=== Next Step ===
Choose based on your primary goal:
- {differentiator_1} → Use {team_1}
- {differentiator_2} → Use {team_2}
```

#### Low Confidence (< 0.50)

```
=== Assessment ===
Goal: Could not confidently determine intent
Query: "{original_query}"

I need more context to recommend the right team. Please clarify:

1. What is the primary outcome you're trying to achieve?
2. Is this new work or improving existing code?
3. Any specific constraints (security, performance, timeline)?

Or try one of these refined queries:
- /consult "build new feature for {domain}"
- /consult "fix bug in {component}"
- /consult "improve performance of {system}"
```

---

## Integration

### With team-discovery Skill

Update `user-skills/guidance/team-discovery/SKILL.md` to expose intent matching:

```markdown
## Intent Matching

Given a user query, the skill:
1. Builds or retrieves cached capability index
2. Extracts signals from query
3. Scores all teams against signals
4. Returns ranked list with confidence scores

### Usage

```bash
# Invoke via skill reference in /consult
# team-discovery returns structured team recommendations
```

### Output Schema

```yaml
intent_match_result:
  query: string
  recommendations:
    - team: string
      confidence: number (0.0 - 1.0)
      rationale: string
      command: string
  clarification_needed: boolean
  suggested_questions: string[] (if clarification_needed)
```
```

### With consult.md Command

Update `user-commands/navigation/consult.md` to use intent matching:

```markdown
## Query Processing (Mode 2)

When query provided:
1. Invoke `team-discovery` skill for intent matching
2. Receive ranked recommendations with confidence scores
3. Apply session context adjustments (complexity, recency)
4. Format output per confidence threshold template
5. Include invocation patterns from `prompting` skill
```

### Integration Points Diagram

```
User Query
    |
    v
/consult command
    |
    v
team-discovery skill
    |
    +-- Build capability index (if not cached)
    |       |
    |       +-- Read teams/*/orchestrator.yaml
    |       +-- Read teams/*/README.md
    |       +-- Read teams/*/agents/*.md
    |
    +-- Extract signals from query
    |
    +-- Score teams against signals
    |
    +-- Apply session context adjustments
    |
    v
Ranked recommendations
    |
    v
/consult formats response
    |
    +-- Reference prompting skill for invocation patterns
    +-- Reference 10x-workflow for phase context
    |
    v
User sees recommendation with confidence
```

---

## Pseudocode

### Main Entry Point

```
function match_intent(query: string, session_context: SessionContext?):
    # Step 1: Build or retrieve capability index
    capability_index = get_or_build_index()

    # Step 2: Extract signals from query
    signals = extract_signals(query)

    # Step 3: Score each team
    scores = {}
    for team_name, team in capability_index.items():
        score = calculate_team_score(signals, team, session_context)
        scores[team_name] = {
            score: score,
            team: team,
            rationale: build_rationale(signals, team, score)
        }

    # Step 4: Sort by score descending
    ranked = sorted(scores.items(), key=lambda x: x[1].score, reverse=True)

    # Step 5: Determine response based on top score
    top_score = ranked[0][1].score if ranked else 0

    if top_score >= 0.80:
        return high_confidence_response(ranked[0])
    elif top_score >= 0.50:
        return medium_confidence_response(ranked[:3])
    elif top_score >= 0.30:
        return low_confidence_response(ranked[:3], query)
    else:
        return no_match_response(query)


function calculate_team_score(signals, team, session_context):
    # Base score components
    t_score = trigger_score(signals, team) * 0.40
    d_score = domain_score(signals, team) * 0.25
    p_score = problem_score(signals, team) * 0.20

    base_score = t_score + d_score + p_score

    # Bonus components
    c_bonus = complexity_bonus(session_context, team) * 0.10
    r_bonus = recency_bonus(session_context, team) * 0.05

    bonuses = c_bonus + r_bonus

    # Penalty components
    e_penalty = exclusion_penalty(signals, team) * 0.50

    # Final score clamped to [0, 1]
    final_score = max(0, min(1, base_score + bonuses - e_penalty))

    return final_score


function build_rationale(signals, team, score):
    parts = []

    # Explain trigger matches
    matched_triggers = find_matching_triggers(signals, team)
    if matched_triggers:
        parts.append(f"Matches triggers: {', '.join(matched_triggers)}")

    # Explain domain match
    if domain_score(signals, team) > 0:
        parts.append(f"Domain '{team.domain}' aligns with query")

    # Explain exclusion if applicable
    if exclusion_penalty(signals, team) > 0:
        parts.append(f"Note: Query may match 'Not for' criteria")

    return "; ".join(parts) if parts else "General capability match"
```

---

## Edge Cases

### 1. No Matches Above Threshold

**Scenario**: Query is too vague or doesn't match any team capabilities.

**Example**: `/consult "help me"` or `/consult "do something"`

**Handling**:
- Return `no_match_response` with:
  - Acknowledgment that intent is unclear
  - List of clarifying questions
  - Examples of well-formed queries
  - Offer to display team roster (`/consult --team`)

### 2. Multiple Teams with Equal Confidence

**Scenario**: Two or more teams have identical scores.

**Example**: `/consult "improve system"` might match sre-pack, hygiene-pack, and 10x-dev-pack equally.

**Handling**:
- Break ties using secondary criteria in order:
  1. Recency bonus (most recently used team wins)
  2. Agent count (team with more specialists wins, implies broader capability)
  3. Alphabetical order (deterministic fallback)
- Present as "Multiple teams may fit equally well"
- Provide clear differentiators for user decision

### 3. Conflicting Signals

**Scenario**: Query matches both triggers AND exclusion phrases for the same team.

**Example**: `/consult "write documentation for my one-off script"`
- "write documentation" -> triggers doc-team-pack
- "one-off script" -> excluded by 10x-dev-pack

**Handling**:
- Exclusion penalty is applied AFTER base score calculation
- If net score remains positive, team is still recommended with caveat
- Rationale includes: "Note: Query partially matches 'Not for' criteria - verify scope"
- Confidence classification may drop one level

### 4. Empty or Vague Queries

**Scenario**: User provides minimal input.

**Examples**:
- `/consult ""` (empty)
- `/consult "?"` (punctuation only)
- `/consult "the"` (stopword only)

**Handling**:
- Detect empty/trivial query in extraction phase
- Skip scoring entirely
- Return general help (equivalent to `/consult` with no arguments)
- Display team roster and common starting points

### 5. Multi-Team Queries

**Scenario**: Query clearly requires multiple teams.

**Example**: `/consult "build a feature with security review and documentation"`

**Handling**:
- Identify multiple high-confidence matches
- Present as workflow recommendation:
  ```
  This goal spans multiple teams:
  1. 10x-dev-pack (build feature) -> then handoff to
  2. security-pack (security review) -> then handoff to
  3. doc-team-pack (documentation)

  Recommended workflow: /sprint with phased handoffs
  ```

### 6. Session Context Override

**Scenario**: Session context strongly suggests a team different from query match.

**Example**: Active session is `ecosystem-pack` but query is `/consult "write some tests"`

**Handling**:
- Session context provides bonus, not override
- If query strongly matches different team, recommend it
- Include note: "Note: You're currently in ecosystem-pack session. This recommendation would change your active team."

---

## Performance Considerations

### Index Caching Strategy

| Aspect | Approach |
|--------|----------|
| Cache location | In-memory (session-scoped) |
| Cache key | `capability_index` |
| TTL | 30 minutes |
| Invalidation | On team switch, manual refresh |
| Persistence | None (rebuild each session) |

### Lazy Loading

To minimize startup time, defer expensive operations:

1. **README.md parsing**: Only load when needed for NOT-for or enrichment
2. **Agent file enumeration**: Only on first full index build
3. **Synonym expansion**: Pre-compute and embed in code

### Query Normalization Overhead

Signal extraction adds ~5-10ms overhead. Acceptable for interactive use.

Optimization opportunities:
- Pre-tokenize trigger phrases (done at index build time)
- Use compiled regex patterns for problem signal matching
- Cache normalized query tokens within single request

### Memory Footprint

Estimated capability index size for 11 teams:
- ~500 bytes per team (strings, arrays)
- Total: ~5.5 KB
- Negligible for in-memory caching

---

## Future Enhancements

### Phase 2: ML-Based Similarity Scoring

Replace keyword matching with embedding-based semantic similarity:
- Embed team descriptions and user queries using Claude embeddings
- Compute cosine similarity for ranking
- Maintains interpretability via nearest-trigger explanation

### Phase 3: User Feedback Loop

Learn from user team selections:
- Track when user chooses different team than top recommendation
- Adjust trigger weights based on correction patterns
- Periodic retraining of problem->team mappings

### Phase 4: Cross-Session Learning

Persist intent patterns across sessions:
- Build user-specific intent profiles
- "You usually use 10x-dev-pack for 'feature' queries"
- Opt-in personalization with privacy controls

---

## Backward Compatibility

### Classification: COMPATIBLE

This design is additive:
- New algorithm supplements existing hardcoded behavior
- `/consult --team` continues to work (reads from index)
- No breaking changes to command syntax
- Graceful degradation if index build fails (fall back to hardcoded list)

### Migration Path

1. **Phase 1**: Implement capability index builder, validate against hardcoded lists
2. **Phase 2**: Add intent extraction, score calculation in parallel with existing logic
3. **Phase 3**: Switch consult.md to use team-discovery for recommendations
4. **Phase 4**: Remove hardcoded team lists from consult-ref/SKILL.md

### Rollback Strategy

If issues discovered:
- Set `CONSULT_USE_DYNAMIC_DISCOVERY=false` to disable
- Falls back to existing hardcoded behavior
- No data migration required

---

## Integration Tests

### Test Matrix

| Test Case | Input | Expected Outcome |
|-----------|-------|------------------|
| High confidence single match | `/consult "build a new feature"` | 10x-dev-pack with confidence >= 0.80 |
| High confidence security | `/consult "security review for auth"` | security-pack with confidence >= 0.80 |
| Medium confidence ambiguous | `/consult "improve the system"` | 2-3 teams with scores 0.50-0.79 |
| Low confidence vague | `/consult "help"` | Clarification questions returned |
| Exclusion penalty applied | `/consult "one-off script"` | 10x-dev-pack demoted, hygiene-pack or rnd-pack up |
| Multi-team detection | `/consult "feature with docs and security"` | Sprint workflow recommended |
| Empty query | `/consult ""` | General help displayed |
| Domain exact match | `/consult "documentation"` | doc-team-pack highest |
| Problem pattern match | `/consult "API is slow"` | sre-pack primary, 10x secondary |
| Session context bonus | Query in ecosystem-pack session | ecosystem-pack gets recency bonus |
| New team discovery | Add new `test-pack` team | Appears in recommendations |

### Satellite Diversity Coverage

| Satellite Type | Test Focus |
|----------------|------------|
| baseline | Index builds correctly from minimal orchestrator.yaml |
| minimal | Graceful handling when README.md missing |
| standard | Full enrichment from all sources |
| complex | Multi-agent teams with rich trigger sets |

---

## Files Changed

### New Files

| Path | Purpose |
|------|---------|
| `docs/design/DESIGN-intent-matching.md` | This design document |

### Modified Files (Implementation Phase)

| Path | Change |
|------|--------|
| `user-skills/guidance/team-discovery/SKILL.md` | Add intent matching section, output schema |
| `user-commands/navigation/consult.md` | Use team-discovery for query processing |
| `user-skills/guidance/consult-ref/SKILL.md` | Reference team-discovery, remove hardcoded lists |

---

## Appendix: Complete Team Inventory

Current teams and their key characteristics (as of 2026-01-04):

| Team | Domain | Key Triggers | Exclusions |
|------|--------|--------------|------------|
| 10x-dev-pack | software development | build feature, PRD, TDD, implementation | documentation, infrastructure, one-off scripts |
| debt-triage-pack | technical debt management | debt, legacy, sprint planning | implementation |
| doc-team-pack | documentation lifecycle | docs, writing, readme | code review, performance |
| ecosystem-pack | ecosystem infrastructure | CEM, roster, sync, satellite | application code, team workflows |
| forge-pack | agent team creation | new team, agent, forge | production features |
| hygiene-pack | code quality | refactor, cleanup, smells | new features |
| intelligence-pack | product analytics | analytics, metrics, A/B test | strategy, market research |
| rnd-pack | technology exploration | research, prototype, explore | production code |
| security-pack | security assessment | security, vulnerability, compliance | general code review |
| sre-pack | site reliability engineering | reliability, incident, SLO | feature development |
| strategy-pack | business strategy | market, strategy, business | tactical, engineering |

---

## Sign-Off Checklist

- [x] Solution architecture documented with rationale
- [x] Schema definitions complete with validation rules
- [x] Backward compatibility classified: COMPATIBLE
- [x] Migration path specified (phased rollout)
- [x] Settings merge algorithm: N/A (no settings changes)
- [x] Integration test matrix with expected outcomes
- [x] File changes specified at file/function level
- [x] No unresolved design decisions
