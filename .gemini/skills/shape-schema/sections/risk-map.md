# Risk Map

Sprint-level risk assessment identifying what could go wrong, how likely it is, and what to do about it. Use for high-stakes initiatives or work in unfamiliar territory.

## Schema

```yaml
risk_map:
  - sprint: sprint-<N>
    risks:
      - description: "<what could go wrong>"
        probability: <low|medium|high>
        impact: <low|medium|high>
        mitigation: "<concrete preventive action>"
        contingency: "<what to do if the risk materializes>"
```

## Guidance

- **Risks are sprint-scoped.** A risk that spans the entire initiative belongs in every sprint it affects, with sprint-specific mitigation. Global risks with no sprint-level mitigation are not actionable.
- **Mitigation is preventive; contingency is reactive.** Mitigation reduces probability ("prototype with a mock before integrating the real API"). Contingency handles materialized risk ("if the API is unavailable, fall back to the cached schema and add a remediation sprint").
- **Probability and impact use simple scales.** Low/medium/high. No percentages, no elaborate matrices. The goal is triage, not precision.
- **Focus on risks agents can act on.** "The company might pivot" is not an actionable risk. "The upstream API schema might change during sprint-3" is -- mitigation: pin to a specific schema version.

## Anti-Patterns

- **Risk theater.** Listing 20 low-probability risks to appear thorough. Focus on the 3-5 risks that would actually derail the initiative.
- **Mitigation without contingency.** Prevention is not guaranteed. Every medium or high-impact risk needs a contingency plan.
