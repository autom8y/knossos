# Playbook: Technology Evaluation

> Structured assessment of new technologies before adoption

## When to Use

- Evaluating a new library, framework, or tool
- Comparing technology alternatives
- Assessing build vs buy decisions
- Prototyping integration with third-party services
- Moonshot exploration of emerging technologies

## Prerequisites

- Clear problem statement (what are we trying to solve?)
- Success criteria defined
- Time-box agreed (exploration can be unbounded otherwise)

## Command Sequence

### Phase 1: Initialize

```bash
/rnd
```
**Expected output**: Team switched to rnd-pack, roster displayed
**Decision point**: If quick feasibility check, use `/spike` instead.

### Phase 2: Start Session

```bash
/start "Technology name/domain" --complexity=EVALUATION
```
**Expected output**: Session created, context established
**Decision point**: Adjust complexity level:
- SPIKE: Quick feasibility (hours)
- EVALUATION: Structured assessment (days)
- MOONSHOT: Long-term architecture exploration

### Phase 3: Technology Scouting

Technology Scout surveys the landscape.

**Expected output**: tech-assessment with options, tradeoffs, initial recommendation
**Decision point**: Narrow to 2-3 candidates for deeper evaluation.

### Phase 4: Integration Research

```bash
/handoff integration-researcher
```
**Expected output**: integration-map showing how technology fits existing stack
**Decision point**: Are there blocking integration concerns?

### Phase 5: Prototyping

```bash
/handoff prototype-builder
```
**Expected output**: Working prototype, documented learnings
**Decision point**: Does prototype validate or invalidate hypothesis?

### Phase 6: Synthesis

```bash
/handoff moonshot-architect
```
**Expected output**: Final recommendation with migration path if adopting
**Decision point**: Go/no-go for adoption.

### Phase 7: Finalize

```bash
/wrap
```
**Expected output**: Session summary, decision record

## Variations

- **SPIKE complexity**: Skip formal scouting, jump to prototype
- **Build vs Buy**: Include cost analysis in scouting
- **MOONSHOT**: Extended exploration, checkpoint regularly

## Success Criteria

- [ ] Technology landscape understood
- [ ] Integration feasibility confirmed
- [ ] Prototype validates core use case
- [ ] Clear recommendation with rationale
- [ ] Migration path documented (if adopting)

## Rollback

If evaluation reveals blockers:
```bash
/park
# Document learnings for future reference
/wrap                          # Close with "no-go" decision
```

## Decision Framework

When synthesizing recommendation:

| Factor | Weight | Questions |
|--------|--------|-----------|
| Fit | High | Does it solve our problem? |
| Integration | High | How hard to integrate? |
| Maturity | Medium | Production-ready? Community support? |
| Cost | Medium | License, infrastructure, maintenance |
| Learning curve | Low | Team ramp-up time |
