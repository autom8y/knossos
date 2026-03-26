---
description: "Rendering Strategies companion for ui-architecture skill."
---

# Rendering Strategies

> SSG/ISR/Streaming SSR/Full SSR/Islands/Resumability decision framework, hydration spectrum, per-route configuration.

## The Rendering Hierarchy

Default to the most static strategy that satisfies freshness requirements. Each step down trades cache efficiency for data freshness.

**Hierarchy** (most static to most dynamic):
SSG → ISR (time-based) → ISR (on-demand) → Streaming SSR → Full SSR → CSR

## Per-Route Strategy Selection

Rendering strategy is a per-route configuration decision, not a global application setting. A single application should use SSG for marketing pages, ISR for product listings, SSR for authenticated dashboards, and CSR for admin panels.

### Decision Matrix

| Question | If Yes | If No |
|----------|--------|-------|
| Does content change less than once per day? | **SSG** | Continue |
| Does content change periodically (hourly/daily)? | **ISR** (time-based, interval = change frequency) | Continue |
| Does content change on specific events (publish, webhook)? | **ISR** (on-demand, triggered by webhook) | Continue |
| Does content require per-request data with multiple sources of varying latency? | **Streaming SSR** | Continue |
| Does content require authentication or real-time personalization? | **SSR** with caching headers | Continue |
| Pure client interaction, no SEO requirement? | **CSR** | Re-evaluate — most routes fit above |

**Escalation**: If a route does not clearly fit one strategy, default to ISR with 60–300 second revalidation interval.

## Strategy Reference

| Strategy | LCP | INP | CLS | JS Payload | Ideal Use Case |
|----------|-----|-----|-----|-----------|----------------|
| **SSG** | Excellent | Neutral | Low risk | Minimal | Marketing, docs, blogs, landing pages |
| **ISR** | Excellent (cache hit) | Neutral | Low risk | Minimal | Product catalogs, CMS-driven content |
| **Streaming SSR** | Good | Good | Medium risk (requires layout reservation) | Moderate | Data-heavy pages, multiple API sources |
| **Full SSR** | Good | Moderate | Low-Medium | Moderate-High | Authenticated pages, per-request personalization |
| **Islands** | Excellent | Excellent | Low risk | Minimal | Content sites with isolated interactive widgets |
| **Resumability** | Good | Excellent | Low risk | Minimal at load | Interactive apps, mobile-first, TTI-critical |
| **SPA (CSR)** | Poor | Poor initially | High risk | Maximum | Internal tools, admin panels only |

## Islands Architecture

Treats a page as a static HTML document containing independently-hydrating interactive regions ("islands"). Each island hydrates independently, fails independently, and ships only its own JavaScript.

**When to use**: content-heavy pages with localized interactivity — blog post with comment form, product page with add-to-cart button.

**Rules**:
- Each island independently loadable (no dependency on parent component tree)
- Each island declares its hydration trigger: `on load`, `on visible`, `on idle`, `on interaction`
- Default state of any page region is static HTML; interactivity must be explicitly opted into
- Flag any architecture where static content requires JavaScript to render

## The Hydration Spectrum

Choose the minimum hydration that satisfies interactivity requirements.

| Interactivity Density | Strategy | Rationale |
|----------------------|----------|-----------|
| 0% (pure content) | No hydration | Zero JavaScript needed |
| 1–20% (content with isolated interactions) | Partial hydration / Islands | Only interactive regions need JS |
| 20–50% (mixed content and interaction) | Progressive hydration | Spread hydration over time, prioritize visible |
| 50–80% (interaction-heavy with some static) | Selective hydration | Framework prioritizes automatically |
| 80–100% (application-like) | Full hydration or Resumability | Full tree needs activation |

**Hydration trigger preference** (most restrictive first): `on interaction` > `on visible` > `on idle` > `on load`. Flag any page that hydrates below-the-fold components on initial load.

## Streaming SSR

Delivers HTML progressively using HTTP chunked transfer encoding. The page shell (header, navigation, layout) ships immediately while data-dependent regions arrive as their data resolves.

**Rules**:
- Page shell must be the first chunk sent, before any data fetches resolve
- Each data-dependent region wrapped in a loading boundary
- Loading boundaries must reserve layout space to prevent CLS
- Most critical content (above-the-fold) resolves in earliest stream chunk
- Flag any SSR implementation that waits for all data fetches before sending any HTML

## Zero JavaScript Default

A page should deliver readable, navigable, semantically meaningful content with zero JavaScript. JavaScript enhances that baseline.

**Progressive enhancement hierarchy**:
1. Semantic HTML with server-rendered content
2. CSS for layout, style, and basic interactions (`<details>/<summary>`, `:hover`, transitions)
3. JavaScript for stateful interactivity that HTML/CSS cannot provide

**Platform capabilities reducing JavaScript requirements** (Baseline 2025):
- `<dialog>` for modals (focus trapping built in)
- `<details>/<summary>` for accordions
- Popover API for tooltips and dropdowns
- View Transitions API for animated page transitions
- Container queries for responsive components
- `:has()` selector for parent-based styling

Before generating JavaScript for a UI behavior, check if a platform primitive exists.

## ISR Revalidation: Time-Based vs On-Demand

| Type | Use When | Tradeoff |
|------|---------|---------|
| Time-based | Content updates at predictable frequency | Simple; may serve stale content for up to one full interval |
| On-demand | Content updates on specific events (publish webhook) | Fresher; requires integration with content pipeline |

For high-traffic routes, time-based with short intervals (30–120 seconds) often provides sufficient freshness with simpler architecture.

## Edge Rendering Decision

| Question | If Yes | If No |
|----------|--------|-------|
| Response fully generated from cached data, KV store, or request headers? | **Edge** | Continue |
| Requires database reads with strong consistency? | **Origin** | Continue |
| Requires database writes? | **Origin** | Continue |
| Requires > 128MB memory or > 30s execution? | **Origin** | Continue |
| Traffic globally distributed with latency sensitivity? | **Edge** with SWR | Origin if regionally concentrated |

**Never edge render with origin database dependencies** — cross-region round trips negate latency benefits.
