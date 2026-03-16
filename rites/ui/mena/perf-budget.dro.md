## Context

Check performance budget compliance for a route, page, or component. Evaluates against Core Web Vitals thresholds (LCP < 2.5s, INP < 200ms, CLS < 0.1) and JavaScript budget (365 KB gzipped for 3-second TTI on P75 global devices).

## Your Task

You are performing a performance budget compliance check.

1. **Identify the target**: Ask the user which route, page, or component to evaluate if not specified. Accept current rendering strategy, estimated bundle size, or code to review.

2. **Rendering strategy assessment**:
   - Identify the current rendering strategy (SSG, ISR, Streaming SSR, Full SSR, CSR, Islands)
   - Flag if SSR is applied to content that could be static (missing cache efficiency)
   - Flag if CSR is applied to SEO-critical pages
   - Flag if a single strategy is applied uniformly to all routes

3. **JavaScript budget check**:
   - Target: 365 KB gzipped for 3-second TTI on P75 global devices
   - Account for framework cost first — it comes out of the total budget
   - Flag any individual dependency > 30 KB (candidate for lazy loading)
   - Check for non-above-the-fold components loaded eagerly
   - Identify code splitting opportunities (route-level, component-level, interaction-based)

4. **CLS risk assessment**:
   - Flag any images/videos without explicit `width`/`height` or `aspect-ratio`
   - Flag dynamic content regions without reserved space (`min-height`)
   - Flag web fonts without `font-display` strategy
   - Flag streaming SSR boundaries without placeholder dimensions
   - Flag any animations using layout-triggering properties (`top`, `left`, `width`, `height`, `margin`)

5. **LCP risk factors**:
   - Render-blocking resources above the fold
   - Hero images without proper loading priority
   - Client-side data fetching for primary page content (should be server-rendered)

6. **INP risk factors**:
   - Long synchronous tasks in interaction handlers
   - Missing passive event listeners
   - Synchronous computation in response to user input

7. **Report**:
   - Current strategy vs recommended strategy
   - Budget status: estimated JS size vs 365 KB threshold
   - CLS risks with specific file/element locations
   - Prioritized recommendations (highest impact first)
   - CI gate recommendations: bundle size assertion, Lighthouse CI thresholds
