# Strategy Pack Workflow

## Phase Flow

```
┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐      ┌───────────────────┐
│      Market       │─────▶│   Competitive     │─────▶│  Business Model   │─────▶│     Roadmap       │
│    Researcher     │      │     Analyst       │      │     Analyst       │      │    Strategist     │
└───────────────────┘      └───────────────────┘      └───────────────────┘      └───────────────────┘
  Market Analysis       Competitive Intel         Financial Model         Strategic Roadmap
```

## Phases

| Phase | Agent | Artifact | Entry Criteria |
|-------|-------|----------|----------------|
| market-research | market-researcher | Market Analysis | User request |
| competitive-analysis | competitive-analyst | Competitive Intel | Market analysis complete, complexity >= STRATEGIC |
| business-modeling | business-model-analyst | Financial Model | Competitive analysis complete (or market if TACTICAL) |
| strategic-planning | roadmap-strategist | Strategic Roadmap | Financial model complete |

## Complexity Levels

- **TACTICAL**: Single decision, existing market data
  - Phases: business-modeling, strategic-planning
- **STRATEGIC**: New market entry, major product bet
  - Phases: market-research, competitive-analysis, business-modeling, strategic-planning
- **TRANSFORMATION**: Business model change, company pivot
  - Phases: market-research, competitive-analysis, business-modeling, strategic-planning

## Phase Skipping

At TACTICAL complexity, market-research and competitive-analysis phases are skipped for quick decisions using existing data.
