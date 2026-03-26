---
description: "State Patterns companion for ui-architecture skill."
---

# State Patterns

> State classification, SWR caching, optimistic UI rules, signals reactivity, state machines.

## The Four State Categories

All application state falls into exactly four categories with fundamentally different management requirements. Mixing them is the root cause of most state management complexity.

| Category | Ownership | Characteristics | Where It Lives |
|----------|-----------|----------------|----------------|
| **Server State** | Remote source | Asynchronous, shared across clients, requires caching | Data-fetching layer (SWR/TanStack Query) |
| **Client State** | Local UI | Ephemeral — modal visibility, in-progress form input, selected tabs | Component-local or lightweight shared store |
| **URL State** | Browser URL | Shareable, bookmarkable — filters, pagination, sort order, view mode | Query parameters |
| **Derived State** | Computed from above | Any value computable from other state | Selectors/computed properties — never stored |

**Before writing any state management code**: classify every piece of state into one of these four categories.

**Detection rule**: if state stores both fetched API data and UI toggles in the same store, flag as God Store.

## State Placement Decision Tree

1. Is this data fetched from an API, shared across clients? → **Server cache** (SWR/TanStack Query)
2. Should this state be shareable via URL (filters, pagination, view mode)? → **URL query parameters**
3. Can this value be computed from existing state? → **Derived/computed** (selector, computed property)
4. Is this state used by only one component? → **Component-local state**
5. Is this state shared across unrelated components? → **Shared client state** (lightweight store)
6. Does this state involve complex ordered transitions with impossible state combinations? → **State machine**

## Stale-While-Revalidate: Default Cache Strategy for Server State

Serve cached (potentially stale) data immediately while revalidating in the background. This is the correct default for data-fetching code.

**Rules**:
- Always cache responses keyed by request parameters
- Serve from cache on subsequent requests, trigger background revalidation on cache hit
- Configure `staleTime` based on data volatility: seconds for feeds/chat, minutes for catalogs, hours for configuration

**Exceptions** (require cache-and-network or event-driven invalidation):
- Financial data, inventory counts, account balances, access permissions
- Any domain where stale reads cause real-world harm

**Cache invalidation strategy**:
| Data Criticality | Strategy |
|-----------------|---------|
| Non-critical, high-read (catalogs, blog posts) | TTL with SWR, staleness measured in minutes |
| Critical (account balances, inventory, permissions) | Event-driven via WebSocket/SSE push |
| Mixed criticality | Hybrid: event-driven for core, TTL for supporting data |
| Always pair event-driven with a TTL fallback | Handle missed events |

## Optimistic UI: Reversibility Assessment

Optimistic UI is appropriate only when all three conditions hold:
1. The operation succeeds >97% of the time
2. The action is reversible or the cost of a false positive is low
3. Server response arrives within the 2-second cognitive flow window

**Default pessimistic** (confirm-then-display): financial transactions, inventory decrements, medical records, access control, audit-logged records.

**Default optimistic with rollback**: social interactions (likes, comments, follows), preference toggles, drag-and-drop reordering.

**Rule**: every optimistic update must store previous state (or inverse operation) and restore it on failure. Pattern: snapshot before mutation → apply optimistic change → on API failure, restore snapshot and surface error.

**Heuristic**: "If this operation fails silently after showing success, what is the worst-case consequence?" — if the answer is non-trivial, use pessimistic.

## Minimal State: Derive Everything Else

The store contains only canonical, non-derivable data.

**Rules**:
- Check that no stored value is computable from other stored values (if `totalPrice` can be derived from `items`, it must be a selector)
- Normalize relational data: `{ entities: { [type]: { [id]: entity } }, ids: [id] }`
- Use memoized selectors (createSelector pattern) for derived values
- Flag any `setState` call that writes a value derivable from existing state

## URL as State Manager

Any state that should survive a page refresh, be shareable via link, or be navigable via browser history belongs in the URL.

**Belongs in URL**: search queries, filters, pagination, sort order, selected tabs, view modes, date ranges.

**Rules**:
- Omit default values from the URL to keep it clean
- Use `pushState` for distinct navigation actions, `replaceState` for refinements
- Never put sensitive data (tokens, passwords, PII) in URLs
- URL length limit: ~2000 characters practical
- Temporary UI states (modal open, dropdown expanded) do not belong in the URL

## Signals: Convergent Reactivity Primitive

Signals — reactive containers that hold a value, track dependents, and notify them on change — are the cross-framework consensus for fine-grained reactivity.

Three primitives:
1. **Signal/State** — writable reactive value
2. **Computed/Derived** — read-only value lazily derived with automatic memoization
3. **Effect/Watcher** — side-effect that runs when dependent signals change

**Agent rules for signals**:
- Computed values must be lazy (only evaluated when accessed) and glitch-free (no intermediate inconsistent states)
- Effects must not create circular dependencies
- Signal updates should be synchronous and batched where possible to prevent cascading re-renders

The TC39 Signals Proposal (Stage 1, 2024) aims to standardize this, with input from Angular, Vue, Solid, Preact, Svelte, MobX, and others.

## State Machines: When to Use

Use a state machine when:
- There are 3+ interrelated boolean flags where some combinations are invalid ("impossible states")
- The order of operations matters (multi-step wizards, authentication flows, payment processes)
- The system needs to be formally verifiable

**Detection rule**: N boolean flags create 2^N possible states. If only M are valid and M << 2^N, a state machine is warranted. Example: `isLoading`, `isError`, `isSuccess`, `isRetrying` → refactor to `idle | loading | success | error | retrying`.

**Do not use state machines for**: simple toggles, preference settings, basic CRUD forms. A state machine with fewer than 3 states and no guards is overengineering.
