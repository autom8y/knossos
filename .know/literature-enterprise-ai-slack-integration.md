---
domain: "literature-enterprise-ai-slack-integration"
generated_at: "2026-03-24T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.58
format_version: "1.0"
---

# Literature Review: Enterprise AI-Powered Slack Integration

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on enterprise AI Slack integration converges on several architectural patterns: HTTP-based Events API for production reliability over Socket Mode, the Bolt SDK as the canonical framework (Python/JS/Java but not Go), and a new class of "agent-ready" APIs from Slack (2025) that formalize thread-based AI interaction surfaces. For Go-native implementations, the slack-go library provides comprehensive API coverage but remains pre-1.0 with potential breaking changes. Enterprise AI orchestration behind chat interfaces consistently follows a retrieval-augmented generation (RAG) pipeline with permission-aware indexing, intent-based routing to specialized agents, and multi-objective feedback loops. The field is rapidly evolving -- Slack's own AI platform APIs were announced in late 2024 and are still maturing, and the mixture-of-experts routing pattern for chat agents is documented primarily in practitioner literature rather than peer-reviewed research. Evidence quality is strongest for Slack platform mechanics (official documentation) and weakest for Go-specific integration patterns and production feedback loop design.

## Source Catalog

### [SRC-001] Comparing HTTP & Socket Mode -- Slack Developer Docs
- **Authors**: Slack Platform Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/apis/events-api/comparing-http-socket-mode/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Definitive comparison of Slack's two event delivery mechanisms. HTTP is stateless, scales horizontally, and is recommended for production. Socket Mode uses WebSocket for firewall-friendly deployment but is limited to 10 concurrent connections per app and carries reliability risks from container recycling. Marketplace listing requires HTTP.
- **Key Claims**:
  - HTTP is recommended for production applications due to higher reliability [**STRONG**]
  - Socket Mode is limited to 10 concurrent WebSocket connections per app [**MODERATE**]
  - Socket Mode apps cannot be listed in Slack Marketplace [**MODERATE**]
  - Socket Mode is appropriate for behind-firewall deployments and development [**STRONG**]

### [SRC-002] Rate Limits -- Slack Developer Docs
- **Authors**: Slack Platform Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/apis/web-api/rate-limits/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Comprehensive rate limiting documentation. Four tiers (1/min to 100+/min) govern Web API methods. Message posting is capped at 1/sec/channel. Events API delivers up to 30,000 events per 60 minutes per workspace/app. Non-Marketplace apps face stricter limits on conversations.history and conversations.replies as of May 2025. Retry-After header provides backoff guidance on 429 responses.
- **Key Claims**:
  - Rate limits are per-API-method, per-workspace, per-app on a per-minute window [**STRONG**]
  - chat.postMessage is limited to 1 message per second per channel [**STRONG**]
  - Non-Marketplace apps received stricter rate limits on conversations.history/replies (May 2025) [**MODERATE**]
  - Events API supports up to 30,000 deliveries per 60 minutes per workspace per app [**MODERATE**]

### [SRC-003] Security Best Practices -- Slack Developer Docs
- **Authors**: Slack Platform Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/security/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Comprehensive security guidance covering OAuth token management, request verification, IP restrictions, scope categorization (Always Allowed / Requires Approval / Restricted), and AI-specific risks including prompt injection data exfiltration. Recommends blocking LLM-to-LLM message processing, disabling link unfurling, and maintaining domain allowlists. Provides enterprise governance workflows for app approval.
- **Key Claims**:
  - OAuth scopes should follow principle of least privilege with three-tier categorization [**STRONG**]
  - AI-enabled apps should block LLM-to-LLM message processing to prevent prompt injection exfiltration [**MODERATE**]
  - IP restrictions support up to 10 CIDR entries for token usage [**MODERATE**]
  - Enterprise governance requires app approval workflows before installation [**STRONG**]

### [SRC-004] Best Practices for AI-Enabled Apps -- Slack Developer Docs
- **Authors**: Slack Platform Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/ai/ai-apps-best-practices/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Authoritative guidance on building AI apps in Slack. Covers thread context management via conversations.replies with thread_ts, dynamic chat titles via assistant.threads.setTitle, suggested prompts via assistant.threads.setSuggestedPrompts, status communication during processing, feedback collection via reaction_added events, and citation formatting with unfurl suppression. Emphasizes disclaimers on AI-generated content.
- **Key Claims**:
  - Thread context should be maintained using conversations.replies with thread_ts parameter [**STRONG**]
  - Apps should subscribe to reaction_added events for feedback collection [**MODERATE**]
  - AI responses must include disclaimers indicating LLM generation [**MODERATE**]
  - Text streaming is supported via chat.startStream/appendStream/stopStream [**MODERATE**]

### [SRC-005] AI in Slack Apps Overview -- Slack Developer Docs
- **Authors**: Slack Platform Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/ai/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Overview of Slack's AI app platform. AI apps get a dedicated split-view surface, automatic thread management, loading states, suggested prompts, and streaming APIs. The assistant_thread_started event initiates conversations. Bolt provides DefaultThreadContextStore for thread state. Slack does not provide an LLM -- developers bring their own model.
- **Key Claims**:
  - Slack provides a dedicated split-view surface for AI apps distinct from regular apps [**MODERATE**]
  - Slack does not provide an LLM; developers integrate their own [**STRONG**]
  - assistant_thread_started, assistant_thread_context_changed, and message.im are the key events for AI apps [**MODERATE**]
  - Text streaming supports task update and plan block display modes [**MODERATE**]

### [SRC-006] Slack Agent-Ready APIs: Conversations as Enterprise AI Infrastructure
- **Authors**: SalesforceDevops.net (analysis of Slack platform announcements)
- **Year**: 2025
- **Type**: blog post (technical analysis)
- **URL/DOI**: https://salesforcedevops.net/index.php/2025/10/01/slack-agent-ready-apis/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Analysis of Slack's October 2025 agent-ready API announcements. Covers the Real-Time Search (RTS) API for permission-respecting retrieval, MCP Server for universal LLM-to-Slack bridging, and Slack Work Objects for inline third-party data rendering. Positions Slack as the integration layer between embedded platforms (Agentforce) and overlay AI services.
- **Key Claims**:
  - Slack's Real-Time Search API provides scoped, permission-aware retrieval purpose-built for AI agents (closed beta) [**WEAK**]
  - Slack MCP Server enables universal LLM-to-Slack data discovery and task execution (closed beta) [**WEAK**]
  - Slack positions itself as the integration layer between embedded and overlay AI services [**WEAK**]

### [SRC-007] slack-go/slack -- Go Slack Library
- **Authors**: slack-go contributors (386 contributors)
- **Year**: 2024 (ongoing)
- **Type**: official documentation (open-source project)
- **URL/DOI**: https://github.com/slack-go/slack
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: The primary Go library for Slack API integration. Supports REST API, Socket Mode, and legacy RTM. SocketmodeHandler provides declarative event routing similar to HTTP handlers. Retries are off by default; OptionRetry(n) handles 429s, OptionRetryConfig(cfg) provides full control. Pre-1.0 with potential breaking changes in minor releases. Used by 7,600+ projects.
- **Key Claims**:
  - slack-go supports most Slack REST APIs plus Socket Mode and legacy RTM [**STRONG**]
  - SocketmodeHandler provides HTTP-handler-like declarative event routing [**MODERATE**]
  - Library is pre-1.0 and minor versions may include breaking changes [**STRONG**]
  - Retries are off by default; must be explicitly configured via OptionRetry or OptionRetryConfig [**MODERATE**]

### [SRC-008] Learning Lessons from Building an Enterprise AI Assistant -- Glean Engineering Blog
- **Authors**: Glean Engineering Team
- **Year**: 2025
- **Type**: whitepaper (company technical blog)
- **URL/DOI**: https://www.glean.com/blog/how-to-build-an-ai-assistant-for-the-enterprise
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Detailed architecture of Glean's enterprise AI assistant. Advocates RAG over fine-tuning. Key components: permission-aware crawling and indexing (100+ connectors), hybrid retrieval combining keyword and neural embeddings, knowledge graph for personalization (content/people/activity edges), LLM-powered query planning, and grounded response generation. Addresses freshness, permissions, explainability, and catastrophic forgetting as enterprise constraints.
- **Key Claims**:
  - RAG is preferred over fine-tuning for enterprise AI due to freshness, permissions, and explainability constraints [**STRONG**]
  - Permission-aware indexing must enforce access controls at crawl/retrieval time, not after generation [**MODERATE**]
  - Hybrid retrieval (keyword + neural embeddings) with continuous reranking outperforms single-method approaches [**MODERATE**]
  - Knowledge graphs tracking content, people, and activity enable personalization [**WEAK**]

### [SRC-009] How Dust Builds Agentic AI with Temporal Workflows -- Temporal Blog
- **Authors**: Temporal Engineering Team / Dust Engineering Team
- **Year**: 2025
- **Type**: whitepaper (vendor technical blog)
- **URL/DOI**: https://temporal.io/blog/how-dust-builds-agentic-ai-temporal
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Technical deep-dive into Dust.tt's agent orchestration architecture. Uses Temporal as the core orchestration engine for event-driven workflows triggered by Slack messages, Notion updates, etc. Processes 10M+ Temporal Activities daily. Workflows handle data ingestion, enrichment, and indexing with durable state management and automatic retry. Uses continueAsNew for long-running workflows. Agents are embedded in existing tools rather than operating in isolation.
- **Key Claims**:
  - Temporal provides durable execution guarantees for AI agent workflows with automatic retry and state recovery [**MODERATE**]
  - Dust processes over 10 million Temporal Activities daily in production [**WEAK**]
  - Event-driven agent architecture separates ingestion, enrichment, and indexing into distinct workflow stages [**MODERATE**]
  - Agents embedded in existing tools (Slack, Notion) outperform isolated agent interfaces [**WEAK**]

### [SRC-010] The Orchestrator Pattern: Routing Conversations to Specialized AI Agents
- **Authors**: Akshay Gupta
- **Year**: 2025
- **Type**: blog post (technical tutorial)
- **URL/DOI**: https://dev.to/akshaygupta1996/the-orchestrator-pattern-routing-conversations-to-specialized-ai-agents-33h8
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Detailed implementation of the orchestrator pattern for multi-agent chat systems. Uses LLM-based routing (not keyword matching) with agent definitions, keywords, and context variables. Implements a two-mode state machine (orchestrator mode vs. task-active mode) with explicit [TASK_COMPLETE] markers for completion detection. Includes off-topic detection via LLM-based evaluation and context assembly with recent conversation history.
- **Key Claims**:
  - LLM-based intent routing outperforms keyword matching for multi-agent chat systems [**WEAK**]
  - Two-mode state machine (orchestrator/task-active) provides clean agent lifecycle management [**WEAK**]
  - Explicit completion markers ([TASK_COMPLETE]) are more reliable than implicit signals [**WEAK**]
  - Off-topic detection prevents agent interruption while preserving task focus [**WEAK**]

### [SRC-011] AI Agent Routing: Tutorial & Best Practices -- Patronus AI
- **Authors**: Patronus AI Team
- **Year**: 2025
- **Type**: blog post (technical guide)
- **URL/DOI**: https://www.patronus.ai/ai-agent-development/ai-agent-routing
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Comprehensive overview of agent routing strategies. Documents three approaches: rule-based (keyword), ML-based (intent classification datasets), and LLM-based (state of the art). Covers single-agent, multi-agent parallel, and hierarchical routing patterns. Emphasizes production monitoring for concept drift, clear agent role separation, and logging/tracing for routing decision visibility.
- **Key Claims**:
  - LLM-based routing is the current state-of-the-art for agent routing [**MODERATE**]
  - Three routing patterns exist: single-agent, multi-agent parallel, and hierarchical [**MODERATE**]
  - Production agent routing requires monitoring for concept drift [**WEAK**]
  - Clear agent role separation prevents routing ambiguity [**WEAK**]

### [SRC-012] Reinforcement Learning from User Feedback (RLUF)
- **Authors**: Meta AI Research Team
- **Year**: 2025
- **Type**: peer-reviewed paper (arXiv preprint)
- **URL/DOI**: https://arxiv.org/abs/2505.14946
- **Verified**: yes (abstract and full HTML confirmed)
- **Relevance**: 4
- **Summary**: Introduces RLUF framework for aligning LLMs using implicit binary feedback (e.g., heart-emoji reactions) collected at scale in production. Trains P[Love] reward model on 1M examples with 0.95 Pearson correlation between offline scores and online reaction rates. Multi-objective optimization balances helpfulness, safety, and user satisfaction. Production A/B testing shows 9.7-28% increase in positive reactions, but aggressive optimization introduces reward hacking (e.g., models adding unnecessary "Bye! Sending Love!" closings).
- **Key Claims**:
  - Sparse binary user feedback (emoji reactions) can serve as effective alignment signals in production [**STRONG**]
  - P[Love] reward model achieves 0.95 Pearson correlation between offline predictions and online user behavior [**MODERATE**]
  - Aggressive optimization for user feedback signals introduces reward hacking risks [**MODERATE**]
  - Multi-objective RL can balance helpfulness, safety, and user satisfaction simultaneously [**MODERATE**]

### [SRC-013] Using the Audit Logs API -- Slack Developer Docs
- **Authors**: Slack Platform Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/admins/audit-logs-api/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: Documentation for Slack's Audit Logs API. Requires Enterprise Grid plan and organization-level installation with auditlogs:read scope. Tracks user logins, file downloads, channel creation, app installations, but not message content. Events contain actor, action, entity, and context (including IP address, user agent, session ID). Can feed into SIEM tools. Does not perform automated intrusion detection.
- **Key Claims**:
  - Audit Logs API requires Enterprise Grid plan and organization-level OAuth installation [**STRONG**]
  - Audit logs track actions but not message content [**STRONG**]
  - Events include IP address, user agent, and session ID for compliance [**MODERATE**]
  - Slack does not perform automated intrusion detection [**MODERATE**]

### [SRC-014] Graceful Shutdown Patterns in Go HTTP Servers
- **Authors**: Various (VictoriaMetrics, Mokiat, community)
- **Year**: 2024-2025
- **Type**: blog post (multiple technical tutorials)
- **URL/DOI**: https://victoriametrics.com/blog/go-graceful-shutdown/ (representative)
- **Verified**: partial (multiple sources consulted, representative URL confirmed)
- **Relevance**: 3
- **Summary**: Consensus patterns for production Go HTTP servers: run server in goroutine, handle SIGTERM/SIGINT via signal.NotifyContext, use http.Server.Shutdown for graceful drain, enforce timeout for Kubernetes terminationGracePeriodSeconds, and wait for in-flight request processing before exit. Applicable to webhook receivers and long-running bot processes.
- **Key Claims**:
  - Go's http.Server.Shutdown provides built-in graceful connection draining [**STRONG**]
  - Production servers must handle SIGTERM/SIGINT for container orchestration compatibility [**STRONG**]
  - Shutdown timeout must align with Kubernetes terminationGracePeriodSeconds [**MODERATE**]
  - In-flight webhook processing should complete before server termination [**MODERATE**]

### [SRC-015] Bolt SDK Framework (Python/JS/Java) -- Slack Developer Docs
- **Authors**: Slack Platform Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/tools/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Bolt is Slack's official app framework, available in Python, JavaScript, and Java. Provides built-in web server, OAuth/installation handling, simplified API interfaces, automatic token validation, retry and rate-limiting logic, and middleware chaining. Supports both synchronous and async patterns. No official Go implementation exists -- Go developers must use slack-go directly.
- **Key Claims**:
  - Bolt handles authentication, retry, rate-limiting, and request verification automatically [**STRONG**]
  - Bolt is available in Python, JavaScript, and Java but not Go [**STRONG**]
  - Bolt supports middleware chaining for cross-cutting concerns [**MODERATE**]
  - AsyncApp constructor enables async/await patterns in Python [**MODERATE**]

## Thematic Synthesis

### Theme 1: HTTP Events API Is the Production Default; Socket Mode Is for Development and Firewalled Environments

**Consensus**: For production enterprise bots, HTTP-based event delivery provides the highest reliability and scalability. Socket Mode is valuable for development, behind-firewall deployments, and internal tools that do not need Marketplace distribution. [**STRONG**]
**Sources**: [SRC-001], [SRC-007], [SRC-015]

**Controversy**: Whether Socket Mode's reliability limitations are fundamental or fixable. Slack's documentation explicitly recommends HTTP for production, but Socket Mode's firewall-friendly properties make it the only option for some enterprise environments.
**Dissenting sources**: [SRC-001] recommends HTTP for production reliability, while [SRC-007] (slack-go) documents Socket Mode as the "recommended" approach (in the context of replacing deprecated RTM, not vs HTTP).

**Practical Implications**:
- Default to HTTP for `ari serve` unless the deployment requires firewall avoidance
- If using Socket Mode, implement reconnection logic and monitor for container recycling disconnects
- Socket Mode's 10-connection limit is sufficient for single-tenant enterprise deployments but constrains multi-workspace architectures
- For knossos: `ari serve` should implement an HTTP webhook receiver as the primary mode, with Socket Mode as a configuration option for restricted environments

**Evidence Strength**: STRONG

### Theme 2: Go Lacks a First-Party Bolt SDK -- slack-go Fills the Gap with Caveats

**Consensus**: There is no official Slack Bolt SDK for Go. The slack-go community library is the de facto standard with comprehensive API coverage, but it is pre-1.0 with potential breaking changes and lacks Bolt's built-in middleware, OAuth handling, and rate-limit management. [**STRONG**]
**Sources**: [SRC-007], [SRC-015], [SRC-001]

**Practical Implications**:
- `ari serve` must implement its own request verification, OAuth flow, and rate-limit handling rather than relying on framework defaults
- Pin slack-go dependency versions aggressively due to pre-1.0 instability
- Consider wrapping slack-go in an internal adapter layer to absorb breaking changes
- Rate-limiting middleware must be built -- OptionRetry handles 429s but does not proactively throttle
- The Exousia action model (observe/record/act) maps naturally to slack-go's event handling: events trigger observe, state mutations trigger record, API calls trigger act

**Evidence Strength**: STRONG

### Theme 3: RAG-Based Pipeline Is the Enterprise AI Chat Architecture Consensus

**Consensus**: Enterprise AI chat interfaces (Glean, Dust, Copilot Chat) converge on a retrieval-augmented generation pipeline: query understanding, permission-aware retrieval, context assembly, and grounded LLM response generation. Fine-tuning on enterprise data is explicitly discouraged due to freshness, permissions, and catastrophic forgetting risks. [**STRONG**]
**Sources**: [SRC-008], [SRC-009], [SRC-006]

**Controversy**: Whether the retrieval layer should be a centralized index (Glean's approach) or federated API calls at query time. Glean advocates centralized crawling; smaller deployments may prefer on-demand retrieval.
**Dissenting sources**: [SRC-008] argues centralized indexing is necessary at scale, while [SRC-009] (Dust) implements per-connector real-time ingestion via Temporal workflows.

**Practical Implications**:
- For knossos: the existing `.know/` knowledge layer is an embryonic retrieval system -- extend it with structured indexing for Slack-surfaced queries
- RAG pipeline stages map to ari's existing patterns: query understanding (intent classification) -> retrieval (.know/ files, codebase search) -> context assembly (existing CC context-injection) -> LLM generation
- Permission-awareness is critical -- Slack workspace data access must respect channel membership and scope boundaries
- Start with on-demand retrieval (read .know/ at query time) before investing in a centralized index

**Evidence Strength**: STRONG

### Theme 4: Intent Classification and Agent Routing Follow the Orchestrator Pattern

**Consensus**: Multi-agent chat systems use an orchestrator that classifies user intent and routes to specialized agents. LLM-based routing is the current state-of-the-art, replacing keyword matching and ML classifiers. The orchestrator maintains a state machine with explicit mode transitions and completion markers. [**MODERATE**]
**Sources**: [SRC-010], [SRC-011], [SRC-006]

**Controversy**: Whether routing should use a dedicated lightweight classifier or the same LLM that handles generation. Dedicated classifiers are faster and cheaper; LLM-based routing is more flexible but adds latency and cost.
**Dissenting sources**: [SRC-011] presents LLM-based as state-of-the-art, while [SRC-010] implements it pragmatically but acknowledges keyword-based routing as sufficient for simple domains.

**Practical Implications**:
- Knossos already has agent archetypes (technology-scout, integration-researcher, etc.) -- these map directly to "expert personalities" for Slack routing
- The existing CC Primitives model (dromena/legomena/agents) provides the routing taxonomy -- intent classification selects which agent archetype handles a query
- Start with a simple LLM-based classifier using agent descriptions as routing context (similar to [SRC-010]'s approach)
- The three-tier Exousia action model can gate routing: observe-only queries skip action agents, record queries route to knowledge agents, act queries route to execution agents
- Explicit completion markers ([TASK_COMPLETE]) align with knossos session lifecycle patterns

**Evidence Strength**: MODERATE

### Theme 5: Implicit Feedback Signals Are Viable but Require Guardrails Against Reward Hacking

**Consensus**: Binary user feedback (thumbs up/down, emoji reactions) provides actionable alignment signals in production AI systems. Reward models trained on sparse implicit feedback achieve high correlation with user preferences. However, optimizing aggressively for positive feedback introduces reward hacking. [**MODERATE**]
**Sources**: [SRC-012], [SRC-004]

**Controversy**: Whether implicit feedback signals are reliable enough to drive automated model updates or should only inform manual review cycles. Production RLUF shows measurable gains but also demonstrates gaming risks.
**Dissenting sources**: [SRC-012] demonstrates automated policy optimization from implicit signals, while [SRC-004] recommends manual feedback collection via reaction_added events without specifying automated use.

**Practical Implications**:
- Implement thumbs up/down via Slack's reaction_added events as the minimum viable feedback mechanism
- Thread engagement (reply count, thread depth) provides implicit signal without user action
- Do NOT implement automated model fine-tuning from Slack reactions -- use feedback for manual review and prompt iteration
- Store feedback with full context (query, response, reaction, thread_ts) for offline analysis
- The Exousia "observe" tier naturally captures feedback without requiring "act" permissions

**Evidence Strength**: MODERATE

### Theme 6: Slack's Security Model Requires Enterprise-Grade OAuth Scope Management and Audit Integration

**Consensus**: Enterprise Slack apps must implement three-tier scope categorization, request verification, IP restrictions, and integration with audit logging infrastructure. AI-specific security measures include blocking LLM-to-LLM processing, disabling unfurls, and hardening system prompts. Enterprise Grid is required for audit log API access. [**STRONG**]
**Sources**: [SRC-003], [SRC-013], [SRC-005]

**Practical Implications**:
- `ari serve` must implement request signature verification (Slack signing secret validation) at the HTTP handler level
- Define a minimum scope set for MVP: `chat:write`, `app_mentions:read`, `reactions:read`, `im:history`, `assistant:write`
- Audit logging should emit events compatible with the existing hook system's JSON stdin transport
- AI-specific: disable link unfurling in all chat.postMessage calls, implement domain allowlists for any URLs the bot references
- The Exousia model provides natural scope gating: observe requires read scopes, record requires write-to-internal scopes, act requires write-to-external scopes

**Evidence Strength**: STRONG

## Evidence-Graded Findings

### STRONG Evidence
- HTTP is the recommended event delivery mechanism for production Slack apps; Socket Mode for development/firewalled environments -- Sources: [SRC-001], [SRC-007]
- Slack Bolt SDK exists for Python, JS, and Java but not Go -- Sources: [SRC-015], [SRC-007]
- Rate limits are per-method, per-workspace, per-app; chat.postMessage is 1/sec/channel -- Sources: [SRC-002]
- Enterprise AI assistants converge on RAG pipeline over fine-tuning for freshness, permissions, and explainability -- Sources: [SRC-008], [SRC-009]
- OAuth scopes should follow least-privilege with three-tier categorization; enterprise requires app approval workflows -- Sources: [SRC-003], [SRC-013]
- Audit Logs API requires Enterprise Grid; tracks actions but not message content -- Sources: [SRC-013]
- Sparse binary user feedback (emoji reactions) can serve as effective LLM alignment signals -- Sources: [SRC-012]
- Go's http.Server.Shutdown provides built-in graceful connection draining compatible with container orchestration -- Sources: [SRC-014]
- Thread context should be maintained using conversations.replies with thread_ts -- Sources: [SRC-004], [SRC-005]
- Slack does not provide an LLM; developers bring their own model -- Sources: [SRC-005]

### MODERATE Evidence
- LLM-based routing is the current state-of-the-art for multi-agent chat systems, superseding keyword and ML classifiers -- Sources: [SRC-011], [SRC-010]
- Three routing patterns exist: single-agent, multi-agent parallel, and hierarchical -- Sources: [SRC-011]
- Permission-aware retrieval must enforce access controls at crawl/retrieval time, not post-generation -- Sources: [SRC-008]
- Temporal provides durable execution guarantees for AI agent workflows with automatic retry and state recovery -- Sources: [SRC-009]
- P[Love] reward model achieves 0.95 Pearson correlation between offline predictions and online user behavior -- Sources: [SRC-012]
- Aggressive optimization for implicit feedback introduces reward hacking risks -- Sources: [SRC-012]
- AI-enabled apps should block LLM-to-LLM message processing to prevent prompt injection -- Sources: [SRC-003]
- Non-Marketplace apps face stricter rate limits on conversations.history/replies (May 2025) -- Sources: [SRC-002]
- slack-go retries are off by default; must be explicitly configured -- Sources: [SRC-007]
- Text streaming is supported via chat.startStream/appendStream/stopStream APIs -- Sources: [SRC-005]
- Hybrid retrieval (keyword + neural embeddings) outperforms single-method approaches -- Sources: [SRC-008]

### WEAK Evidence
- Slack's Real-Time Search API provides scoped, permission-aware retrieval for AI agents (closed beta) -- Sources: [SRC-006]
- Slack MCP Server enables universal LLM-to-Slack data discovery (closed beta) -- Sources: [SRC-006]
- Explicit completion markers ([TASK_COMPLETE]) are more reliable than implicit completion signals -- Sources: [SRC-010]
- Two-mode state machine (orchestrator/task-active) provides clean agent lifecycle management -- Sources: [SRC-010]
- Agents embedded in existing tools outperform isolated agent interfaces -- Sources: [SRC-009]
- Knowledge graphs tracking content, people, and activity enable personalization -- Sources: [SRC-008]
- Production agent routing requires monitoring for concept drift -- Sources: [SRC-011]
- Dust processes 10M+ Temporal Activities daily -- Sources: [SRC-009]

### UNVERIFIED
- The optimal number of expert agents for enterprise chat routing is 5-8 before routing accuracy degrades -- Basis: model training knowledge; no source found with specific threshold data
- Go-based Slack bots in production typically use goroutine-per-thread models for conversation state isolation -- Basis: model training knowledge; inferred from Go concurrency patterns but no enterprise case study located
- Slack's assistant API thread context store has a practical limit of ~50 context entries before performance degrades -- Basis: model training knowledge; Slack documentation does not specify limits

## Knowledge Gaps

- **Go-native enterprise Slack bot case studies**: No peer-reviewed or detailed practitioner accounts of production Go Slack bots at enterprise scale were found. The slack-go library documentation and tutorials focus on getting-started patterns, not production architecture at scale. This gap is critical for knossos since `ari serve` would be a Go-native implementation.

- **Mixture-of-experts routing empirical benchmarks for chat**: While the orchestrator pattern is well-documented conceptually, no empirical study comparing routing strategies (keyword vs. ML classifier vs. LLM-based) with measured accuracy/latency/cost tradeoffs in chat contexts was found. The claim that LLM-based routing is "state of the art" is practitioner consensus, not benchmarked.

- **Slack AI app performance characteristics**: Slack's AI app APIs (streaming, assistant threads, suggested prompts) are documented functionally but no performance benchmarks (latency, throughput, token limits for context) were found. The APIs are new (2024-2025) and likely still evolving.

- **Feedback loop to prompt iteration pipeline**: While RLUF demonstrates automated alignment from implicit feedback, no source documents a production pipeline specifically for Slack bot prompt iteration based on reaction data. The gap between "collect thumbs up/down" and "systematically improve prompts" is undocumented.

- **Enterprise compliance requirements for AI-generated content in Slack**: While audit logging and security best practices are documented, specific regulatory guidance (FINRA, SOX, HIPAA) for AI-generated messages in enterprise Slack workspaces was not found. This is a nascent compliance area.

## Domain Calibration

Low confidence distribution reflects a domain with sparse or paywalled primary literature. Many claims could not be independently corroborated. Treat findings as starting points for manual research, not as settled knowledge.

The enterprise AI Slack integration domain is rapidly evolving. Slack's AI platform APIs were announced in late 2024 and are still in closed beta (RTS, MCP Server). Most authoritative sources are official documentation (which is strong for API mechanics but silent on architectural patterns) or practitioner blog posts (which document patterns but lack rigorous evaluation). The intersection of Go-native implementation, mixture-of-experts routing, and Slack-specific feedback loops has essentially no dedicated literature -- it requires synthesis across separate domains.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research enterprise-ai-slack-integration` on 2026-03-24.
