---
name: doc-intelligence
description: "User research, experimentation, and data insights templates. Use when: synthesizing user research, designing A/B tests, reporting experiment results. Triggers: research findings, user interview, experiment design, A/B test, insights report, user research."
---

# Intelligence Templates

> Templates for user research, experimentation, and data-driven decision-making.

## Purpose

Provides structured templates for intelligence workflows: user research synthesis with evidence-backed findings, A/B test pre-registration with statistical rigor, and experiment results analysis with ship/no-ship recommendations.

## Template Catalog

| Template | Purpose | Agent |
|----------|---------|-------|
| [Research Findings](templates/research-findings.md) | User research synthesis with participant quotes and themes | research-analyst |
| [Experiment Design](templates/experiment-design.md) | A/B test pre-registration with sample size and stopping rules | experiment-designer |
| [Insights Report](templates/insights-report.md) | Data-driven decision recommendation with segment analysis | data-analyst |

## When to Use Each Template

| Scenario | Template |
|----------|----------|
| Synthesizing user interviews | Research Findings |
| Planning an A/B test | Experiment Design |
| Analyzing experiment results | Insights Report |
| Deciding whether to ship | Insights Report |
| Understanding user behavior | Research Findings |
| Pre-registering experiment metrics | Experiment Design |

## Quality Gates Summary

| Template | Gate Criteria |
|----------|---------------|
| **Research Findings** | Findings backed by quotes, confidence levels assigned, themes with frequency |
| **Experiment Design** | Hypotheses in null/alt form, sample size calculated, guardrails defined |
| **Insights Report** | Statistical significance reported, alternatives ruled out, ship/no-ship paths documented |

## Progressive Disclosure

- [research-findings.md](templates/research-findings.md) - Research synthesis (RESEARCH-{slug})
- [experiment-design.md](templates/experiment-design.md) - A/B test specification (EXPERIMENT-{slug})
- [insights-report.md](templates/insights-report.md) - Decision recommendation (INSIGHT-{slug})
