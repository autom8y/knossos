---
description: "Performance Budgets companion for ui-ref skill."
---

# Performance Budgets

> Core Web Vitals thresholds, JavaScript budgets, CLS prevention, and rendering strategy performance impact.

## Core Web Vitals: Minimum Acceptable Performance

Thresholds measured at P75 of real user experiences. These are floors, not goals.

| Metric | Good | Needs Improvement | Poor |
|--------|------|------------------|------|
| **LCP** (Largest Contentful Paint) | < 2.5s | 2.5s – 4s | > 4s |
| **INP** (Interaction to Next Paint) | < 200ms | 200ms – 500ms | > 500ms |
| **CLS** (Cumulative Layout Shift) | < 0.1 | 0.1 – 0.25 | > 0.25 |

As of January 2026, only 55.7% of origins pass all three CWV metrics. INP replaced FID in March 2024. Sites passing all three see 24% lower bounce rates and better organic rankings.

## JavaScript Budget

Performance budgets must target P75 global devices — mid-range Android phones 15–25% as fast as premium hardware.

| Target | Max Gzipped JS | Network | Device |
|--------|---------------|---------|--------|
| 3-second TTI (JS-heavy) | 365 KB | 7.2 Mbps down, 94ms RTT | P75 mobile |
| 5-second TTI (JS-heavy) | 650 KB | Same | Same |
| 3-second TTI (markup-based) | 75 KB | Same | Same |
| Critical-path HTML/CSS/fonts | ~150 KB | Same | Same |

**JavaScript cost**: gzipped JS expands 5–7× in memory, then must be parsed, compiled, and executed on the main thread. JavaScript is the most expensive resource per byte. 45% of global mobile connections still occur on 2G/3G.

**Agent rule**: when adding a dependency or generating substantial client-side code, consider the cumulative bundle impact. Flag when combined dependencies exceed budget thresholds. Framework cost counts against the budget before any application code.

## CLS Prevention

Cumulative Layout Shift occurs when visible elements change position after initial render. It is a rendering coordination problem — the server must communicate layout intent.

**Required patterns**:
- All images and videos: explicit `width` and `height` attributes, or CSS `aspect-ratio`
- Dynamic content regions (ads, embeds, lazy-loaded sections): reserve space with `min-height`
- Web fonts: `font-display: optional` for body text (no CLS); `font-display: swap` + size-adjusted fallbacks for display/heading fonts only
- Streaming SSR loading boundaries: placeholder dimensions matching final content
- Never insert DOM elements above existing visible content without user-initiated action
- Animations: use CSS `transform`, never layout-triggering properties (`top`, `left`, `width`, `height`, `margin`)

## LCP Optimization

- Avoid render-blocking resources above the fold
- Ensure hero images use proper loading strategies (preload critical images)
- Minimize critical-path CSS
- Implement streaming SSR to deliver page shell before all data resolves

## INP Optimization

- Avoid long main-thread tasks (>50ms) during interaction handlers
- Defer non-critical JavaScript
- Use passive event listeners where appropriate
- Avoid synchronous computation in interaction handlers

## Rendering Strategy Impact on Performance

| Strategy | LCP | INP | CLS | JS Payload | Use Case |
|----------|-----|-----|-----|-----------|----------|
| SSG | Excellent | Neutral | Low risk | Minimal | Marketing, docs, blogs |
| ISR | Excellent (cache hit) | Neutral | Low risk | Minimal | Product catalogs, CMS content |
| Streaming SSR | Good | Good | Medium risk (needs layout reservation) | Moderate | Data-heavy pages with multiple API sources |
| Full SSR | Good | Moderate (full hydration) | Low-Medium | Moderate-High | Authenticated pages, per-request data |
| Islands | Excellent | Excellent | Low risk | Minimal | Content sites with isolated interactive widgets |
| SPA (CSR) | Poor | Poor initially | High risk | Maximum | Internal tools, admin panels only |

**Default**: SSG > ISR > Streaming SSR > Full SSR > CSR. Apply the most static strategy that satisfies freshness requirements.

## Performance Testing as a Quality Gate

Performance belongs in the testing pyramid, not as a separate concern:
- **Bundle size assertion**: fail CI if total JS exceeds 365 KB gzipped
- **Lighthouse CI**: fail if LCP > 2.5s, INP > 200ms, CLS > 0.1
- **Output format**: Lighthouse JSON with metric-level pass/fail (agent-parseable)

Performance tests gate deployment, not every PR.
