---
summary: The Evans Principle — reconstructions that outlive their evidence become obstacles to understanding.
see_also: [know, session, sos]
aliases: [evans]
---
Named for Sir Arthur Evans, whose concrete reconstructions at Knossos — originally helpful interpretive aids — became permanent obstacles that now prevent scholars from seeing what the actual evidence says. Applied to context management: session state, cached knowledge, and reconstructed context must be bound to their evidence lifespan. When a session wraps, its context is archived, not left to rot. When .know/ files expire, they are regenerated from current evidence. Stale reconstructions — hardcoded assumptions, orphaned session artifacts, context that outlives the work it described — are the Evans trap. The writeguard hook enforces this principle: direct mutations to context files are blocked because uncontrolled edits create reconstructions that outlive their evidence.
