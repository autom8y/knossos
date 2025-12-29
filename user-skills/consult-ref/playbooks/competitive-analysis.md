# Playbook: Competitive Analysis

> Strategic intelligence for product and market decisions

## When to Use

- Entering a new market
- Planning product roadmap
- Responding to competitor moves
- Positioning and pricing decisions
- Investment or acquisition due diligence

## Prerequisites

- Clear market or product scope
- Access to public competitor information
- Stakeholder alignment on strategic questions

## Command Sequence

### Phase 1: Initialize

```bash
/strategy
```
**Expected output**: Team switched to strategy-pack, roster displayed
**Decision point**: If just market sizing, may be simpler scope.

### Phase 2: Start Session

```bash
/start "Analysis scope" --complexity=STRATEGIC
```
**Expected output**: Session created, context established
**Decision point**: Adjust complexity level:
- TACTICAL: Single competitor or feature comparison
- STRATEGIC: Market segment analysis
- TRANSFORMATION: Industry-level strategic planning

### Phase 3: Market Research

Market Researcher gathers competitive intelligence.

**Expected output**: competitive-analysis with competitor profiles, positioning maps
**Decision point**: Are there gaps in competitive data to address?

### Phase 4: Financial Analysis

```bash
/handoff financial-modeler
```
**Expected output**: Unit economics comparison, pricing analysis, market sizing
**Decision point**: Skip if not financially focused.

### Phase 5: Strategic Synthesis

```bash
/handoff strategic-planner
```
**Expected output**: Strategic roadmap with prioritized initiatives
**Decision point**: Validate with leadership before acting.

### Phase 6: Finalize

```bash
/wrap
```
**Expected output**: Session summary, strategic recommendations

## Variations

- **TACTICAL complexity**: Skip financial modeling, focus on feature comparison
- **Pricing focus**: Emphasize financial modeler phase
- **Roadmap planning**: Extend strategic synthesis with OKRs

## Success Criteria

- [ ] Key competitors identified and profiled
- [ ] Competitive positioning understood
- [ ] Strategic opportunities identified
- [ ] Actionable recommendations provided
- [ ] Stakeholder alignment on next steps

## Rollback

If strategic direction changes:
```bash
/park
# Reassess with new information
/continue
/handoff strategic-planner    # Revise recommendations
```

## Analysis Frameworks

Use appropriate frameworks for synthesis:

| Framework | When to Use |
|-----------|-------------|
| Porter's Five Forces | Industry structure |
| SWOT | Organization positioning |
| BCG Matrix | Portfolio decisions |
| TAM/SAM/SOM | Market sizing |
| Jobs-to-be-Done | Customer needs |
