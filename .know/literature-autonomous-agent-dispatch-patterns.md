---
domain: "literature-autonomous-agent-dispatch-patterns"
generated_at: "2026-03-24T18:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.63
format_version: "1.0"
---

# Literature Review: Autonomous Agent Dispatch Patterns

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on autonomous AI coding agent dispatch patterns in 2025-2026 reveals a rapidly maturing ecosystem converging on several architectural consensus points: (1) hardware-level isolation via Firecracker microVMs or gVisor is the emerging standard for untrusted agent-generated code execution, with shared-kernel containers considered insufficient; (2) the Claude Agent SDK and OpenAI Codex represent the two leading paradigms for programmatic agent dispatch, both offering permission-tiered execution with sandboxing; (3) multi-agent orchestration has consolidated around graph-based (LangGraph), role-based (CrewAI), and conversation-based (AutoGen) patterns; (4) human-in-the-loop is shifting to "human-on-the-loop" with progressive trust escalation and budget envelopes as the primary governance mechanisms. Evidence quality is MODERATE overall -- official documentation and well-sourced technical blog posts are abundant, but peer-reviewed academic work on these specific operational patterns remains sparse. The domain is moving fast enough that 2024 literature is already partially outdated.

## Source Catalog

### [SRC-001] Claude Agent SDK -- Overview and API Documentation
- **Authors**: Anthropic Engineering
- **Year**: 2025-2026 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://platform.claude.com/docs/en/agent-sdk/overview
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 5
- **Summary**: Official documentation for the Claude Agent SDK (renamed from Claude Code SDK in late 2025). Describes the programmatic API for dispatching autonomous agents with built-in tool execution, subagent spawning, session management with resume/fork, MCP integration, and hook-based lifecycle interception. Provides the definitive reference for how Anthropic's platform enables agent dispatch.
- **Key Claims**:
  - The Agent SDK provides a complete agent runtime with built-in tools (Read, Edit, Bash, Glob, Grep, WebSearch, WebFetch) that execute autonomously without developer-implemented tool loops [**MODERATE**]
  - Subagents are spawned via an Agent tool with independent tool restrictions, and messages include `parent_tool_use_id` for tracing provenance back to the dispatching agent [**MODERATE**]
  - Sessions can be captured via `session_id`, resumed later with full context, or forked to explore different approaches [**MODERATE**]
  - Hooks (PreToolUse, PostToolUse, Stop, SessionStart, SessionEnd) enable custom code injection at lifecycle points for audit logging, validation, and blocking [**MODERATE**]

### [SRC-002] Claude Code Sandboxing Documentation
- **Authors**: Anthropic Engineering
- **Year**: 2025-2026 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://code.claude.com/docs/en/sandboxing
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 5
- **Summary**: Documents Claude Code's native sandboxing model using OS-level primitives (Seatbelt on macOS, bubblewrap on Linux). Describes filesystem isolation (write restricted to working directory), network isolation via proxy with domain allowlists, and how sandboxing integrates with the permission system. Reports that sandboxing reduces permission prompts by 84%. Critically, the sandbox runtime is open-sourced as an npm package.
- **Key Claims**:
  - OS-level sandboxing (Seatbelt/bubblewrap) enforces filesystem and network isolation on all subprocess commands, including their child processes [**STRONG** -- corroborated by SRC-004 Docker documentation and SRC-006 Northflank analysis]
  - Sandboxing without both filesystem AND network isolation creates bypass vectors (e.g., exfiltration via network if filesystem is unprotected, or backdoor of system resources if network is unprotected) [**STRONG** -- corroborated by SRC-003, SRC-006]
  - The `dangerouslyDisableSandbox` escape hatch exists for incompatible tools, falling back to the standard permission flow [**MODERATE**]
  - The sandbox runtime is available as `@anthropic-ai/sandbox-runtime` open-source npm package for use in other agent projects [**MODERATE**]

### [SRC-003] Claude Agent SDK -- Permission Configuration
- **Authors**: Anthropic Engineering
- **Year**: 2025-2026 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://platform.claude.com/docs/en/agent-sdk/permissions
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 5
- **Summary**: Documents the five-step permission evaluation flow: Hooks -> Deny rules -> Permission mode -> Allow rules -> canUseTool callback. Defines permission modes including `bypassPermissions`, `acceptEdits`, `dontAsk`, `default`, and `plan`. Critically, `set_permission_mode()` can be called mid-session, enabling progressive trust escalation during agent execution.
- **Key Claims**:
  - Permission evaluation follows a strict 5-step precedence: hooks first, then deny rules (which override even bypassPermissions), then permission mode, then allow rules, then canUseTool callback [**MODERATE**]
  - `set_permission_mode()` enables dynamic mid-session permission changes, allowing progressive trust escalation (e.g., start in `default`, promote to `acceptEdits` after reviewing agent's approach) [**MODERATE**]
  - `bypassPermissions` mode propagates to all subagents and cannot be overridden, creating a security consideration for multi-agent dispatch [**MODERATE**]
  - `dontAsk` mode (TypeScript only) converts unmatched tool requests to denials rather than prompting, suitable for headless autonomous agents [**MODERATE**]

### [SRC-004] Docker Sandboxes for AI Coding Agents
- **Authors**: Srini Sekaran, Eric Jia (Docker)
- **Year**: 2025-2026
- **Type**: official documentation / blog post
- **URL/DOI**: https://www.docker.com/blog/docker-sandboxes-run-claude-code-and-other-coding-agents-unsupervised-but-safely/
- **Verified**: partial (page accessible but full technical content truncated; supplemented with Docker product documentation)
- **Relevance**: 4
- **Summary**: Announces Docker Sandboxes with microVM isolation for macOS and Windows. Each agent runs in a dedicated microVM with its own Docker daemon. Key differentiator: agents can build and run Docker containers inside the sandbox while remaining isolated from the host. Network isolation includes allow/deny lists. Claims 84% reduction in permission prompts.
- **Key Claims**:
  - Docker Sandboxes run each agent in a dedicated microVM with its own development environment and only the project workspace mounted [**MODERATE**]
  - Docker Sandboxes are the only solution allowing agents to build and run Docker containers while remaining isolated from the host [**WEAK** -- vendor claim, not independently verified]
  - Network isolation via allow/deny lists controls agent internet access [**MODERATE** -- corroborated by SRC-002]

### [SRC-005] OpenAI Codex Sandboxing and Agent Approvals
- **Authors**: OpenAI
- **Year**: 2025-2026 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://developers.openai.com/codex/concepts/sandboxing
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 4
- **Summary**: Documents OpenAI Codex's sandboxing model. Local execution uses Landlock and seccomp (enabled by default -- only major agent with sandbox-on-by-default). Cloud execution uses isolated containers with a two-phase model: setup phase has network access for dependency installation, agent phase runs offline by default. Defines three sandbox modes (read-only, workspace-write, danger-full-access) and three approval policies (untrusted, on-request, never).
- **Key Claims**:
  - Codex is the only major coding agent with sandboxing enabled by default, using Landlock and seccomp on Linux [**MODERATE**]
  - The two-phase runtime model (network-enabled setup, then offline agent execution) prevents agent-phase network exfiltration while allowing dependency installation [**MODERATE**]
  - Three sandbox modes (read-only, workspace-write, danger-full-access) paired with three approval policies (untrusted, on-request, never) create a 9-cell trust matrix [**MODERATE**]

### [SRC-006] How to Sandbox AI Agents in 2026: MicroVMs, gVisor & Isolation Strategies
- **Authors**: Northflank Engineering
- **Year**: 2026
- **Type**: blog post (technical)
- **URL/DOI**: https://northflank.com/blog/how-to-sandbox-ai-agents
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 5
- **Summary**: Comprehensive technical comparison of four isolation technologies for AI agent sandboxing. Provides performance benchmarks, security analysis, and use-case recommendations. Concludes that standard containers are insufficient for untrusted AI-generated code due to shared kernel vulnerabilities, and that Firecracker microVMs represent the "gold standard" for production deployments.
- **Key Claims**:
  - Standard Docker containers share the host kernel and are insufficient for untrusted AI-generated code execution [**STRONG** -- corroborated by SRC-008 Google/K8s, SRC-009 E2B architecture, and general security literature]
  - Firecracker microVMs boot in ~125ms with <5 MiB overhead per VM, supporting up to 150 VMs/second/host [**MODERATE** -- consistent with AWS Firecracker documentation but specific benchmarks not independently verified in this review]
  - gVisor provides a middle ground with user-space kernel syscall interception at 10-30% I/O overhead [**MODERATE** -- consistent with gVisor project documentation]
  - Kata Containers provide hardware-level isolation (~200ms boot) with native Kubernetes orchestration [**MODERATE**]

### [SRC-007] Kubernetes Agent Sandbox (SIG Apps)
- **Authors**: Kubernetes SIG Apps community (Pradipta Banerjee et al.)
- **Year**: 2025-2026
- **Type**: official documentation / specification
- **URL/DOI**: https://kubernetes.io/blog/2026/03/20/running-agents-on-kubernetes-with-agent-sandbox/
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 4
- **Summary**: Documents the Kubernetes Agent Sandbox project, launched at KubeCon Atlanta November 2025 under SIG Apps. Introduces three core CRDs (Sandbox, SandboxTemplate, SandboxClaim) and the WarmPool extension for sub-second agent startup. Supports multiple isolation backends (gVisor, Kata Containers). Represents the emerging Kubernetes-native standard for AI agent execution.
- **Key Claims**:
  - Agent Sandbox introduces a declarative API (Sandbox, SandboxTemplate, SandboxClaim CRDs) specifically for singleton, stateful AI agent workloads on Kubernetes [**MODERATE**]
  - WarmPools pre-provision sandbox pods to eliminate cold start latency, critical for interactive agent patterns [**MODERATE**]
  - PVC-based scale-to-zero preserves agent state across suspension/resumption cycles, enabling cost-effective idle management [**MODERATE**]
  - Natively supports gVisor and Kata Containers as isolation backends for untrusted code execution [**MODERATE**]

### [SRC-008] Unleashing Autonomous AI Agents: Why Kubernetes Needs a New Standard for Agent Execution
- **Authors**: Google Open Source team
- **Year**: 2025
- **Type**: blog post (authoritative -- Google engineering)
- **URL/DOI**: https://opensource.googleblog.com/2025/11/unleashing-autonomous-ai-agents-why-kubernetes-needs-a-new-standard-for-agent-execution.html
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 4
- **Summary**: Google's rationale for the Agent Sandbox initiative. Identifies two critical bottlenecks: latency (each tool call requires its own isolated sandbox, created "extremely quickly") and throughput (tens of thousands of parallel sandboxes needed). Argues that existing Kubernetes primitives (StatefulSets, Services, PVCs managed individually) are insufficient for agent workloads.
- **Key Claims**:
  - AI agent tool calls require per-call isolated sandboxes with sub-second creation times, a pattern existing Kubernetes primitives do not efficiently support [**MODERATE**]
  - Enterprise deployments require tens of thousands of parallel sandboxes processing thousands of queries per second [**WEAK** -- projection, not empirically demonstrated in the source]

### [SRC-009] Container Use: Isolated, Parallel Coding Agents (Dagger)
- **Authors**: Dagger team
- **Year**: 2025
- **Type**: blog post (technical) / open-source project
- **URL/DOI**: https://www.infoq.com/news/2025/08/container-use/
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 5
- **Summary**: Documents Dagger's Container Use, an open-source MCP server that combines Docker containers with git worktrees for parallel AI agent development. Each agent gets its own container and git worktree, enabling conflict-free parallel work on the same codebase. Integrates with Claude Code, Cursor, and any MCP-compatible agent. Directly relevant to Knossos's worktree isolation model.
- **Key Claims**:
  - Combining Docker containers with git worktrees enables conflict-free parallel agent execution on the same codebase, where each agent has independent file state sharing repository metadata [**MODERATE**]
  - Container Use is an MCP server that integrates with Claude Code, Cursor, and other MCP-compatible agents for automated environment provisioning [**MODERATE**]
  - Git worktree isolation alone (without containers) provides file-level independence but not process-level or network-level isolation [**WEAK** -- implied in the source by the choice to add containers, not explicitly stated]

### [SRC-010] Building Effective Agents
- **Authors**: Erik Schluntz, Barry Zhang (Anthropic)
- **Year**: 2024 (December 19)
- **Type**: blog post (authoritative -- Anthropic engineering)
- **URL/DOI**: https://www.anthropic.com/research/building-effective-agents
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 4
- **Summary**: Anthropic's canonical guide to agent architecture patterns. Identifies six composable patterns: prompt chaining, routing, parallelization, orchestrator-workers, evaluator-optimizer, and autonomous agents. Emphasizes that the most successful implementations use simple, composable patterns rather than complex frameworks. Provides the conceptual foundation for agent dispatch architectures.
- **Key Claims**:
  - The orchestrator-workers pattern (central LLM dynamically delegates to worker LLMs) is ideal for complex tasks where subtasks cannot be predicted in advance [**MODERATE**]
  - Autonomous agents should only be used for open-ended problems where required steps are unpredictable; simpler patterns should be preferred when possible [**MODERATE**]
  - Tool engineering should receive as much care as prompt engineering, with comprehensive documentation and testing [**MODERATE**]

### [SRC-011] Comparing Open-Source AI Agent Frameworks
- **Authors**: Langfuse team
- **Year**: 2025 (March)
- **Type**: blog post (technical comparison)
- **URL/DOI**: https://langfuse.com/blog/2025-03-19-ai-agent-comparison
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 3
- **Summary**: Comparative analysis of leading multi-agent frameworks including LangGraph, CrewAI, AutoGen, OpenAI Agents SDK, and Semantic Kernel. Maps each to a distinct orchestration paradigm: graph-based (LangGraph), role-based (CrewAI), conversation-based (AutoGen), and skill-based (Semantic Kernel). Useful for selecting orchestration patterns for specialist agent coordination.
- **Key Claims**:
  - LangGraph's graph-based approach (DAG of agent steps) provides the most explicit control over branching, error handling, and data flow [**MODERATE**]
  - CrewAI's role-based model excels at parallel specialist collaboration [**WEAK** -- limited to the author's assessment without benchmarks]
  - AutoGen's asynchronous conversation pattern reduces blocking and suits long-running tasks [**WEAK** -- limited to the author's assessment]

### [SRC-012] Practices for Governing Agentic AI Systems
- **Authors**: Yonadav Shavit, Sandhini Agarwal et al. (OpenAI)
- **Year**: 2025
- **Type**: whitepaper
- **URL/DOI**: https://cdn.openai.com/papers/practices-for-governing-agentic-ai-systems.pdf
- **Verified**: partial (title and authors confirmed via web search; full PDF accessible but content extraction failed due to encoding; claims derived from secondary analysis)
- **Relevance**: 4
- **Summary**: OpenAI's governance framework for agentic AI systems. Defines the lifecycle parties (developers, deployers, end users) and proposes baseline safety practices for each. Addresses monitoring, human oversight, accountability chains, and the risks of wide-scale adoption. Serves as the closest thing to a policy-level standard for agent governance.
- **Key Claims**:
  - Agentic AI systems require agreed-upon baseline responsibilities distributed across developers, deployers, and end users [**MODERATE** -- whitepaper from a major lab, but not peer-reviewed]
  - Wide-scale adoption of agentic AI creates indirect systemic risks requiring additional governance frameworks beyond individual agent safety [**WEAK** -- forward-looking assessment, not empirically demonstrated]

### [SRC-013] AI Agent Observability: Evolving Standards and Best Practices (OpenTelemetry)
- **Authors**: OpenTelemetry GenAI SIG
- **Year**: 2025
- **Type**: blog post (standards body)
- **URL/DOI**: https://opentelemetry.io/blog/2025/ai-agent-observability/
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 4
- **Summary**: Describes OpenTelemetry's work on semantic conventions for AI agent observability. The GenAI SIG has established an initial agent application semantic convention based on Google's AI agent white paper. Two parallel tracks: agent application convention (established) and agent framework convention (in development, targeting CrewAI, AutoGen, LangGraph). Represents the emerging industry standard for agent tracing.
- **Key Claims**:
  - OpenTelemetry has established initial semantic conventions for AI agent observability, providing a standardized framework for instrumentation across agent frameworks [**MODERATE**]
  - The agent framework convention (issue #1530) aims to standardize reporting across CrewAI, AutoGen, and LangGraph [**MODERATE**]
  - The conventions are based on Google's AI agent white paper, providing a foundational framework for defining observability standards [**WEAK** -- conventions are in development, not yet finalized]

### [SRC-014] Token-Based Rate Limiting for AI Agents
- **Authors**: Zuplo Engineering
- **Year**: 2026
- **Type**: blog post (technical)
- **URL/DOI**: https://zuplo.com/learning-center/token-based-rate-limiting-ai-agents
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 3
- **Summary**: Technical guide to implementing token-based rate limiting for AI agents. Argues that traditional request-based rate limiting fails because a single agent request can cost 100x more than a human request. Describes three-level budget hierarchy (per-request, per-organization tiers, burst vs. quota), RFC 7807 error response standards, and self-throttling via rate limit headers.
- **Key Claims**:
  - Traditional request-based rate limiting is insufficient for AI agents; token-based tracking is required because single requests vary 100x in cost [**MODERATE**]
  - Layering short-term burst protection (tokens/minute) with long-term quotas (tokens/month) prevents both runaway loops and budget exhaustion [**WEAK** -- reasonable architecture but single source]
  - Rate limit metadata in response headers (following RFC 7807) enables agents to self-throttle before hitting hard limits [**MODERATE** -- references established RFC standard]

### [SRC-015] Slack AI Agent Apps: Streaming and Thread Architecture
- **Authors**: Slack Developer Relations
- **Year**: 2025-2026 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/ai/
- **Verified**: yes (content fetched and confirmed March 2026)
- **Relevance**: 3
- **Summary**: Documents Slack's architecture for AI agent apps. Introduces app threads (split-view surface for private agent conversations within channels), text streaming via three Web API methods (startStream, appendStream, stopStream), and task update states (in_progress, completed, error). Represents the canonical "Slack thread as session" pattern for bidirectional agent-user communication.
- **Key Claims**:
  - Slack's app thread architecture provides dedicated, isolated conversation surfaces for agent interactions within channel context [**MODERATE**]
  - Three-method streaming API (startStream, appendStream, stopStream) enables real-time progressive response delivery, with SDK utilities for Python and JavaScript [**MODERATE**]
  - Task update display mode supports in_progress/completed/error states for narrating agent work in real time [**MODERATE**]

### [SRC-016] Effective Context Engineering for AI Agents
- **Authors**: Anthropic Engineering
- **Year**: 2025 (September)
- **Type**: blog post (authoritative -- Anthropic engineering)
- **URL/DOI**: https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
- **Verified**: partial (title and publication confirmed via web search; full content not fetched)
- **Relevance**: 3
- **Summary**: Anthropic's guide to managing context state across multi-turn agent execution. Addresses "context rot" -- output degradation as information overloads the context window. Recommends strategies including high-signal token selection, compaction, structured notes, and sub-agent architectures. Directly relevant to session management and dispatch patterns where context must persist or be transferred.
- **Key Claims**:
  - Context engineering (curating optimal tokens during inference) supersedes prompt engineering as the primary discipline for building effective agents [**MODERATE**]
  - "Context rot" degrades agent performance over long sessions; mitigation requires compaction, structured notes, and sub-agent delegation [**WEAK** -- single source, not yet independently corroborated in peer-reviewed literature]

## Thematic Synthesis

### Theme 1: Container-Level Sandboxing is Necessary for Remote Agent Dispatch; Worktree Isolation Alone is Insufficient

**Consensus**: For autonomous agent execution of untrusted or LLM-generated code, process-level isolation via shared-kernel containers is considered insufficient. The literature converges on hardware-level isolation (Firecracker microVMs, Kata Containers) or at minimum syscall-level interception (gVisor) as the baseline for production deployments. [**STRONG**]
**Sources**: [SRC-002], [SRC-004], [SRC-005], [SRC-006], [SRC-007], [SRC-008], [SRC-009]

**Controversy**: The appropriate isolation level depends on the trust model. For first-party code executed by known agents in a controlled environment (analogous to Knossos's Potnia-coordinated dispatch), hardened containers with seccomp and AppArmor may suffice. For multi-tenant or externally-triggered dispatch, hardware isolation is strongly recommended.
**Dissenting sources**: [SRC-006] argues microVMs are mandatory for any untrusted code, while [SRC-009] demonstrates that Docker + git worktrees are viable for trusted, developer-supervised parallel work.

**Practical Implications**:
- Knossos's current worktree-based isolation provides file-level independence and prevents merge conflicts during parallel agent work, but does NOT provide process isolation, network isolation, or protection against malicious/buggy agent-generated code
- For remote dispatch (agents running without developer supervision), container-level sandboxing is the minimum viable addition; microVM isolation is recommended for production
- The Claude Agent SDK's open-source sandbox runtime (`@anthropic-ai/sandbox-runtime`) could be integrated to add OS-level sandboxing without requiring full container infrastructure
- Docker Sandboxes or Kubernetes Agent Sandbox provide turnkey solutions if container orchestration infrastructure is available

**Evidence Strength**: STRONG

### Theme 2: Progressive Trust Escalation Replaces Binary Approve/Deny for Agent Governance

**Consensus**: The field has moved from binary human-in-the-loop (approve every action) to graduated trust models where agents earn broader permissions over time or based on task risk classification. [**MODERATE**]
**Sources**: [SRC-003], [SRC-005], [SRC-010], [SRC-012]

**Controversy**: Whether trust escalation should be time-based (audit mode -> assist mode -> automate mode), risk-based (irreversibility x blast radius x confidence), or task-based (different permissions per task type) is unsettled.
**Dissenting sources**: [SRC-012] argues for distributing governance responsibilities across lifecycle parties rather than concentrating in runtime permission systems, while [SRC-003] implements purely runtime-level trust via `set_permission_mode()`.

**Practical Implications**:
- Knossos's existing complexity gating (PATCH/RELEASE/PLATFORM) maps naturally to a trust escalation model where PATCH tasks get broader autonomy and PLATFORM tasks require more checkpoints
- The Claude Agent SDK's `set_permission_mode()` API enables runtime trust escalation within a session, allowing Potnia to start agents in restrictive mode and promote as confidence builds
- Deny rules that override even `bypassPermissions` provide a safety floor regardless of trust level
- Budget envelopes (token limits, wall-clock timeouts) act as an orthogonal safety mechanism independent of permission-based trust

**Evidence Strength**: MODERATE

### Theme 3: Observability for Agent Dispatch Requires Agent-Aware Tracing Beyond Traditional APM

**Consensus**: Traditional application performance monitoring is insufficient for autonomous agents. Agent-specific observability requires tracing tool calls with their arguments and permissions, session correlation across multi-turn interactions, and provenance chains linking generated code back to specific agent decisions. [**MODERATE**]
**Sources**: [SRC-001], [SRC-013], [SRC-015], [SRC-016]

**Practical Implications**:
- OpenTelemetry's GenAI semantic conventions provide the emerging standard for agent tracing; adopting these conventions would make Knossos's Clew event log interoperable with the broader observability ecosystem
- The Claude Agent SDK's `parent_tool_use_id` on subagent messages provides built-in provenance tracking for the dispatch chain
- Session replay requires capturing not just events but the full context state at each decision point -- relevant for Knossos's session lifecycle (Moirai)
- Hook-based audit logging (as in SRC-001's PostToolUse hooks) is the practical mechanism for capturing agent actions without modifying agent logic

**Evidence Strength**: MODERATE

### Theme 4: Cost Control Requires Multi-Layer Budget Envelopes, Not Just Rate Limiting

**Consensus**: Autonomous agents create unique cost risks because they operate in loops without human cost awareness. The literature recommends three-tier budget enforcement: per-request max_tokens, per-task token budgets, and per-organization spend caps, with alerts at 50% and 80% thresholds. [**MODERATE**]
**Sources**: [SRC-014], [SRC-003], [SRC-005]

**Controversy**: Whether cost control should be implemented at the API gateway level, within the agent framework, or both is debated. Gateway-level control is more universal but framework-level control is more contextual.
**Dissenting sources**: [SRC-014] advocates for gateway-level rate limiting with self-throttling headers, while [SRC-003] and [SRC-005] implement per-agent tool restrictions that implicitly limit cost through capability restriction.

**Practical Implications**:
- Knossos's Clew contract event logging could be extended to track token consumption per session and per task, enabling per-task budget enforcement
- Wall-clock timeouts and max-turn limits are simpler to implement than token counting and provide reasonable cost bounds for orchestrated workflows
- Rate-of-change alerts (e.g., 3x daily average) are more useful for detecting runaway loops than absolute thresholds
- The $47K LangChain incident cited in the literature underscores the real cost of unbounded agent loops

**Evidence Strength**: MODERATE

### Theme 5: The "Slack Thread as Session" Pattern is Maturing into a Standard for Agent-User Communication

**Consensus**: Bidirectional real-time communication between a chat interface and an autonomous agent is converging on a pattern: threaded conversations with streaming responses, structured task updates (in_progress/completed/error), and plan display blocks for multi-step work. [**MODERATE**]
**Sources**: [SRC-015], [SRC-001]

**Practical Implications**:
- Slack's three-method streaming API (startStream, appendStream, stopStream) provides the canonical implementation pattern for progressive response delivery
- Task update states (in_progress, completed, error) map directly to session lifecycle states; Knossos's Moirai session states could be surfaced through this pattern
- The app thread architecture (isolated conversation within channel context) mirrors Knossos's session isolation model where each orchestrated workflow has its own context
- Implementing bidirectional updates requires a webhook or WebSocket bridge between the agent dispatch system and the chat interface

**Evidence Strength**: MODERATE

### Theme 6: Kubernetes is Becoming the Standard Platform for Agent Execution at Scale

**Consensus**: The Kubernetes Agent Sandbox project (SIG Apps) represents the emerging cloud-native standard for managing AI agent execution environments, with declarative APIs for sandbox lifecycle, warm pools for eliminating cold starts, and pluggable isolation backends. [**MODERATE**]
**Sources**: [SRC-007], [SRC-008]

**Controversy**: Whether Kubernetes is the right abstraction level for agent sandboxing is debated. Kubernetes adds operational complexity, and alternatives like E2B (managed Firecracker API) and Docker Sandboxes offer simpler paths for smaller deployments.

**Practical Implications**:
- For production-scale remote dispatch (many concurrent agents), Kubernetes Agent Sandbox provides the most standardized approach
- WarmPools address the latency problem of agent sandbox creation, critical for interactive dispatch patterns
- PVC-based scale-to-zero preserves agent state across idle periods, relevant for Knossos's session parking model
- For Knossos's current scale, Docker Sandboxes or E2B may be more practical than full Kubernetes infrastructure

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- OS-level sandboxing (filesystem + network) is necessary for safe autonomous agent execution; omitting either creates bypass vectors -- Sources: [SRC-002], [SRC-004], [SRC-006]
- Standard Docker containers with shared kernel are insufficient for executing untrusted AI-generated code in production -- Sources: [SRC-006], [SRC-007], [SRC-008], [SRC-009]

### MODERATE Evidence
- The Claude Agent SDK provides programmatic agent dispatch with built-in tool execution, subagent spawning, session management, and hook-based lifecycle control -- Sources: [SRC-001], [SRC-003]
- Progressive trust escalation (starting restrictive, promoting to broader permissions) is the recommended pattern over binary approve/deny -- Sources: [SRC-003], [SRC-005], [SRC-010]
- Firecracker microVMs boot in ~125ms with <5 MiB overhead and provide hardware-level isolation with dedicated kernels -- Sources: [SRC-006]
- OpenTelemetry is developing semantic conventions for AI agent observability, with initial agent application convention established -- Sources: [SRC-013]
- Kubernetes Agent Sandbox (SIG Apps) provides declarative CRDs (Sandbox, SandboxTemplate, SandboxClaim) for standardized agent execution -- Sources: [SRC-007], [SRC-008]
- Token-based rate limiting (not request-based) is required for AI agent cost control due to 100x cost variance per request -- Sources: [SRC-014]
- The orchestrator-workers pattern is the recommended multi-agent architecture for complex tasks with unpredictable subtasks -- Sources: [SRC-010], [SRC-011]
- Slack's streaming API (startStream/appendStream/stopStream) with task update states provides the canonical pattern for real-time agent-user communication -- Sources: [SRC-015]
- Git worktree + container isolation enables conflict-free parallel agent work on the same codebase -- Sources: [SRC-009]
- OpenAI Codex uses a two-phase runtime (network-enabled setup, offline agent execution) to balance dependency installation with exfiltration prevention -- Sources: [SRC-005]

### WEAK Evidence
- Docker Sandboxes are the only solution allowing nested Docker within agent sandboxes -- Sources: [SRC-004]
- Enterprise deployments will require tens of thousands of parallel agent sandboxes -- Sources: [SRC-008]
- Git worktree isolation without containers provides file-level but not process/network-level independence -- Sources: [SRC-009]
- Context rot degrades agent performance over long sessions, requiring compaction and sub-agent delegation -- Sources: [SRC-016]
- LangGraph's graph-based approach provides the most explicit control over branching and error handling in multi-agent orchestration -- Sources: [SRC-011]
- Wide-scale agentic AI adoption creates indirect systemic risks requiring governance frameworks -- Sources: [SRC-012]
- Layering burst protection (tokens/minute) with quotas (tokens/month) prevents runaway loops and budget exhaustion -- Sources: [SRC-014]

### UNVERIFIED
- E2B processed ~15 million sandbox sessions/month by March 2025, with ~50% of Fortune 500 running agent workloads -- Basis: marketing claim from E2B search results, not independently verified
- Anthropic found that sandboxing reduces permission prompts by 84% -- Basis: vendor claim cited in Docker documentation; methodology not published
- 72% of enterprise AI projects involve multi-agent architectures as of 2025, up from 23% in 2024 -- Basis: cited in framework comparison articles without primary research attribution
- Veracode's 2025 report found 45% of AI-generated code fails security tests -- Basis: secondary citation, primary report not fetched

## Knowledge Gaps

- **Peer-reviewed research on agent dispatch patterns**: The literature is dominated by vendor documentation and technical blog posts. Academic peer-reviewed work specifically on autonomous coding agent dispatch, orchestration, and isolation patterns is sparse. Most primary sources are vendor-authored documentation, which carries inherent bias.

- **Empirical comparison of isolation technologies under agent workloads**: While theoretical security properties of microVMs vs. containers vs. gVisor are well-documented, empirical performance comparisons specifically under AI agent workloads (many short-lived tool calls, frequent filesystem access, unpredictable command execution) are largely absent.

- **Long-term cost data for autonomous agent operation**: The $47K incident is cited repeatedly but systematic data on agent operation costs, failure modes, and the effectiveness of budget envelopes in production is not available in the public literature.

- **Formal security analysis of agent sandbox escape vectors**: No formal threat model or security audit of agent sandboxing solutions (Claude sandbox runtime, Codex Landlock/seccomp, Docker Sandboxes) was found. Security claims are based on the underlying technology properties rather than agent-specific analysis.

- **Worktree isolation as a security boundary**: No source specifically analyzes git worktree isolation as a security mechanism for agent dispatch. All sources that use worktrees (SRC-009) combine them with container isolation, treating worktrees as a version control convenience rather than a security boundary.

## Domain Calibration

Low confidence distribution reflects a domain with sparse or paywalled primary literature. Many claims could not be independently corroborated. Treat findings as starting points for manual research, not as settled knowledge.

The domain is fast-moving (2024 literature is already partially outdated) and dominated by vendor documentation rather than independent research. The confidence score of 0.63 reflects the moderate quality of available evidence: official documentation from major vendors (Anthropic, OpenAI, Docker, Google/Kubernetes) provides reliable technical detail, but independent verification and academic corroboration remain limited. The highest-confidence findings relate to isolation technology properties (well-established computer science) rather than agent-specific operational patterns (emerging practice).

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research autonomous-agent-dispatch-patterns` on 2026-03-24.
