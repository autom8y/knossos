# Progressive Craft Layers L0-L4 (F8)

Build-order framework for the harden phase. Establishes sequencing from structural correctness through progressive polish. Source: Freiberg's progressive craft model [WEAK single-source -- reversible if layer ordering proves too sequential].

## The Principle

Build from the foundation outward. Performance and accessibility are Layer 0 -- non-negotiable before any interaction layer begins. Each subsequent layer adds craft, but only after the previous layer is solid. Skipping layers creates fragile implementations that look polished but break under constraint.

## The Layers

### L0: Performance and Structure (Non-Negotiable Base)

**What to build**:
- Progressive enhancement: content accessible without JavaScript
- Performance budget compliance: within rendering-manifest allocations
- CLS mitigation: layout space reserved for all dynamic regions
- Semantic HTML structure: landmarks, headings, roles
- Reduced-motion support: `animation-play-state: paused` on all animations by default; re-enable with `prefers-reduced-motion: no-preference`

**Exit criterion**: If disabled JavaScript, content is still present and readable. Performance budgets met. All animations respect `prefers-reduced-motion`. Static lint passes.

**Rationale**: Freiberg: "Is having a slow website with immaculate attention to visual craft desirable?" Performance and accessibility are foundational, not finishing touches.

### L1: Core Interaction

**What to build**:
- Primary user flow functions correctly (form submission, navigation, state changes)
- Keyboard navigation complete: all interactive elements reachable and operable
- Focus management: explicit focus direction on state changes, modals, dynamic content
- ARIA: roles, states, properties, live regions
- Integration tests pass for all user-meaningful behaviors

**Exit criterion**: A keyboard-only user can complete all primary tasks. Integration tests pass. ARIA audit passes.

**Rationale**: Core interaction correctness before any motion. Motion applied to a broken interaction magnifies the brokenness.

### L2: Motion Choreography

**What to build**:
- Spring physics for interactive elements (per motion-architecture-spec)
- Keyframe animations for ambient elements (per motion-architecture-spec)
- Stagger choreography for related elements (20-30ms per element)
- Timing ceiling enforcement: interactive animations at or below 200ms
- Interruptibility: interactive animations respond to mid-flight interruption

**Exit criterion**: All animations respect the motion-architecture-spec from the intent phase. Interruptibility verified. Timing ceiling verified.

**Rationale**: Motion choreography layer builds on a working interaction (L1) and a performant structure (L0). Motion applied before these layers risks covering up structural problems.

### L3: Novelty (Budgeted at 10%)

**What to build**:
- Novel motion elements within the 10% novelty budget
- Branded easing curves or custom spring configurations
- Unexpected but spatially honest spatial metaphors
- Personality expression within the interaction's frequency tier

**Exit criterion**: Novel elements are within the 10% budget allocation from the motion-architecture-spec. Novel motion does not compromise L0 (performance), L1 (interaction correctness), or L2 (choreography coherence).

**Rationale**: Novelty is the reward layer, not the foundation. It requires solid L0-L2 to build upon and a budget constraint to prevent it from consuming the entire motion vocabulary.

### L4: Polish

**What to build**:
- Reduced-motion alternative verification: inspect every animation and confirm `prefers-reduced-motion` handling is complete
- Cross-browser animation consistency: verify spring physics and custom curves render consistently
- Progressive enhancement edge cases: verify all functionality at JS-disabled and reduced-motion states
- Visual regression baseline: capture screenshots at final state for regression reference
- Component isolation fixtures: all states represented (default, loading, error, empty, edge-case)

**Exit criterion**: All L0-L3 exit criteria met. Reduced-motion alternatives verified. Cross-browser consistency confirmed. Component isolation fixtures exist.

**Rationale**: Polish layer verifies and documents, not builds. Anything requiring structural changes at L4 is a L0-L3 gap -- address at the appropriate layer.

## Layer Application in Practice

The layers are sequential dependencies, not sequential phases. In practice:
1. Build L0 first -- do not proceed until structural foundation is solid
2. Build L1 -- core interaction before any motion
3. Layer L2 onto a working L1 -- motion enhances; it does not compensate
4. Apply L3 within budget -- novelty is a resource to spend deliberately
5. Verify L4 -- polish is documentation and verification, not new construction
