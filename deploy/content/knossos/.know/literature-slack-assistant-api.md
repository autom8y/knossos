---
domain: "literature-slack-assistant-api"
generated_at: "2026-03-25T00:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.72
format_version: "1.0"
---

# Literature Review: Slack Assistant API

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The Slack Assistant API (branded "Agents & AI Apps") is Slack's dedicated framework for building AI-powered apps that interact with users via a split-view panel and threaded conversations, rather than traditional bot DMs. The literature shows strong consensus on the core configuration requirements: enabling the "Agents & AI Apps" feature toggle, subscribing to `assistant_thread_started`, `assistant_thread_context_changed`, and `message.im` events, and granting `assistant:write`, `chat:write`, and `im:history` scopes. The most common pitfall -- the "Sending messages to this app has been turned off" error -- is consistently attributed to missing App Home message tab configuration or absent `chat:write` scope, though the assistant API introduces its own distinct entry point that replaces the traditional Messages tab. Evidence quality is moderate to strong for official documentation but weaker for architectural best practices and troubleshooting, which rely heavily on community sources.

## Source Catalog

### [SRC-001] Developing apps with AI features | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/ai/developing-ai-apps/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: The canonical guide for building Slack AI apps. Defines the complete setup flow: enable "Agents & AI Apps" toggle, which auto-adds `assistant:write` scope; subscribe to `assistant_thread_started`, `assistant_thread_context_changed`, and `message.im` events; grant `chat:write` and `im:history` scopes. Documents the split-view architecture, streaming APIs (`chat.startStream`, `chat.appendStream`, `chat.stopStream`), suggested prompts, thread titles, and feedback mechanisms. Notes that slash commands are unsupported in the split view and workspace guests cannot access AI apps.
- **Key Claims**:
  - Enabling "Agents & AI Apps" feature toggle automatically adds `assistant:write` scope and provides a split-view entry point [**STRONG**]
  - Three events are required: `assistant_thread_started`, `assistant_thread_context_changed`, `message.im` [**STRONG**]
  - Required scopes are `assistant:write`, `chat:write`, `im:history` (plus `conversations:info` for channel context) [**STRONG**]
  - The Messages tab is replaced by Chat and History tabs when "Agents & AI Apps" is enabled [**MODERATE**]
  - Slash commands are not supported in the split view container [**MODERATE**]
  - Workspace guests cannot access apps with Agents & AI Apps enabled [**MODERATE**]

### [SRC-002] App manifest reference | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/reference/app-manifest/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: The complete manifest schema reference. Defines all top-level fields (`_metadata`, `display_information`, `settings`, `features`, `oauth_config`), the `features.bot_user` block (with `display_name` and `always_online`), `features.app_home` block (with `home_tab_enabled`, `messages_tab_enabled`, `messages_tab_read_only_enabled`), `settings.event_subscriptions` (with `request_url`, `bot_events`, `user_events`), and `oauth_config.scopes` (with `bot` and `user` arrays). Documents that event subscriptions require either a `request_url` or `socket_mode_enabled`. Maximum 100 bot_events and 255 scopes.
- **Key Claims**:
  - `display_information.name` is the only required field; all other sections are optional [**STRONG**]
  - Event subscriptions require either `request_url` or `socket_mode_enabled` to be set [**STRONG**]
  - `features.app_home.messages_tab_enabled` controls whether users can message the bot in App Home [**STRONG**]
  - `features.bot_user.display_name` is required if the bot_user block is included [**MODERATE**]
  - Manifest supports v1 and v2 schema versions with differences in function parameter handling [**MODERATE**]

### [SRC-003] Using AI in Slack Bolt.js Apps | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/tools/bolt-js/concepts/ai-apps/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Documents the `Assistant` class in Bolt.js for handling AI-specific events. Provides the three-handler pattern: `threadStarted`, `threadContextChanged`, `userMessage`. Documents available utilities per handler (`say()`, `setSuggestedPrompts()`, `saveThreadContext()`, `setTitle()`, `setStatus()`, `getThreadContext()`, `client`). Defines the `ThreadContextStore` interface with `get` and `save` methods. Notes that thread context is NOT included with individual user messages -- must be stored/retrieved via ThreadContextStore.
- **Key Claims**:
  - The Assistant class requires three handler callbacks: `threadStarted`, `threadContextChanged`, `userMessage` [**STRONG**]
  - Thread context is not included with `message.im` events; apps must use ThreadContextStore to persist and retrieve it [**STRONG**]
  - The `DefaultThreadContextStore` stores context via message metadata, eliminating need for custom storage in most cases [**MODERATE**]
  - `chatStream()` utility abstracts the three-method streaming sequence (`startStream`, `appendStream`, `stopStream`) [**MODERATE**]

### [SRC-004] assistant_thread_started event reference | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/reference/events/assistant_thread_started/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Complete event reference for `assistant_thread_started`. Fires when users open a new assistant thread (via DM or side-container within a channel). Payload includes `assistant_thread` object with `user_id`, `channel_id` (the DM channel), `thread_ts`, and `context` object containing `channel_id` (the channel the user is currently viewing), `team_id`, and `enterprise_id`. No scopes are required to receive this event. Developers should call `conversations.info` to verify channel access before using context data.
- **Key Claims**:
  - No scopes are required to receive the `assistant_thread_started` event [**STRONG**]
  - The event fires both via DM and via side-container within a channel [**STRONG**]
  - The `context.channel_id` represents the channel the user is currently viewing, not the DM channel [**STRONG**]
  - Apps should call `conversations.info` to verify access before using channel context [**MODERATE**]

### [SRC-005] assistant_thread_context_changed event reference | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/reference/events/assistant_thread_context_changed/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Complete event reference for `assistant_thread_context_changed`. Fires when a user navigates to a different channel while the assistant container is open. Payload structure mirrors `assistant_thread_started` with updated `context.channel_id`. No scopes required to receive. Used to track the active context of a user in Slack.
- **Key Claims**:
  - No scopes are required to receive the `assistant_thread_context_changed` event [**STRONG**]
  - The event fires when users switch channels while the assistant container is open [**STRONG**]

### [SRC-006] assistant:write scope reference | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/reference/scopes/assistant.write/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Documents the `assistant:write` scope, which enables an app to act as an AI Assistant. Unlocks three methods: `assistant.threads.setStatus`, `assistant.threads.setSuggestedPrompts`, and `assistant.threads.setTitle`. Works with bot tokens. Automatically added when the "Agents & AI Apps" feature toggle is enabled.
- **Key Claims**:
  - `assistant:write` scope unlocks `setStatus`, `setSuggestedPrompts`, and `setTitle` methods [**STRONG**]
  - The scope is automatically added when enabling the "Agents & AI Apps" feature toggle [**MODERATE**]
  - `assistant.threads.setStatus` now also accepts `chat:write` scope as an alternative [**MODERATE**]

### [SRC-007] App Home | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/surfaces/app-home/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Documents the App Home surface, specifically the Messages tab configuration. The Messages tab requires `chat:write` scope and optionally `im:history` for responding. Must subscribe to `message.im` event. Critical finding: enabling "Agents & AI Apps" feature replaces the Messages tab with Chat and History tabs, making the two configurations mutually exclusive.
- **Key Claims**:
  - The Messages tab requires `chat:write` scope at minimum and `im:history` for response capability [**STRONG**]
  - Enabling "Agents & AI Apps" replaces the Messages tab with Chat and History tabs [**STRONG**]
  - `message.im` event subscription is required to receive user messages in App Home [**STRONG**]

### [SRC-008] AI Assistant Tutorial (Bolt.js) | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/tools/bolt-js/tutorials/ai-assistant/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Step-by-step tutorial for building an AI assistant. Demonstrates the complete manifest-to-code flow. Key architectural insight: the assistant pattern uses specialized event handlers rather than generic message listeners. Documents the importance of setting thread titles from the first message, using `setStatus()` for loading indicators, and fetching channel/thread history for context. Notes that `message.im` does not provide thread context and that HTTP is recommended over Socket Mode for production.
- **Key Claims**:
  - Assistants use specialized event handlers (`threadStarted`, `threadContextChanged`, `userMessage`) rather than generic message listeners [**STRONG**]
  - HTTP mode is recommended over Socket Mode for production deployments [**MODERATE**]
  - `message.im` events do not include thread context; custom storage is needed [**MODERATE**]
  - Bot must handle `not_in_channel` errors and join channels before retrying API calls [**WEAK**]

### [SRC-009] App Manifests for Slack Agents | Vercel Academy
- **Authors**: Vercel
- **Year**: 2025
- **Type**: blog post (educational)
- **URL/DOI**: https://vercel.com/academy/slack-agents/manifest-and-scopes
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Provides a complete, worked manifest example for a Slack agent app with all scopes and events. Documents the scope-to-feature mapping principle: every event requires matching scopes, and Slack validates consistency. Shows a comprehensive bot_events list including all four assistant-relevant events alongside channel and group events. Emphasizes the "start minimal, add as needed" approach to scopes.
- **Key Claims**:
  - Slack validates that event subscriptions have matching scopes and rejects inconsistent manifests [**MODERATE**]
  - A complete agent manifest requires: `channels:history`, `chat:write`, `commands`, `app_mentions:read`, `groups:history`, `im:history`, `mpim:history`, `assistant:write`, `reactions:write`, `reactions:read` scopes [**MODERATE**]
  - Every event requires matching scopes (e.g., `message.channels` needs `channels:history`) [**MODERATE**]

### [SRC-010] Fix: 'Sending Messages to This App Has Been Turned Off' | W3Tutorials
- **Authors**: W3Tutorials
- **Year**: 2025
- **Type**: blog post
- **URL/DOI**: https://www.w3tutorials.net/blog/can-t-send-direct-message-to-slack-bot-feature-turned-off/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Comprehensive troubleshooting guide for the "sending messages turned off" error. Identifies five root causes: missing `chat:write` scope (most common), no active DM conversation, app not reinstalled after permission updates, user blocked the app, and invalid/expired bot tokens. Documents the fix sequence: verify scopes, add missing scopes, reinstall app, create DM channel via `conversations.open`, and check user blocks. Notes that Slack does not auto-create DMs.
- **Key Claims**:
  - Missing `chat:write` scope is the single most common cause of the "sending messages turned off" error [**MODERATE**]
  - App must be reinstalled after adding new scopes for them to take effect [**MODERATE**]
  - Slack does not auto-create DM channels; bot or user must initiate first [**WEAK**]
  - Users can block apps via workspace settings, causing the error even with correct configuration [**WEAK**]
  - Token validity can be verified using the `auth.test` API method [**WEAK**]

### [SRC-011] How We Boosted AI Usage by Building a GPT Chatbot in Slack | Lingvano (Medium)
- **Authors**: Lingvano team
- **Year**: 2024
- **Type**: blog post
- **URL/DOI**: https://medium.com/lingvano/how-we-boosted-ai-usage-across-our-organization-by-building-a-gpt-chatbot-in-slack-589d44c2b4c5
- **Verified**: yes
- **Relevance**: 3
- **Summary**: First-person account of encountering and resolving the "sending messages turned off" error. The team discovered that navigating to Features > App Home and activating "Allow users to send Slash commands and messages from the messages tab" toggle resolved the issue. Provides practical evidence that this is a commonly encountered configuration gap. Also demonstrates that manifest-based setup can miss this toggle.
- **Key Claims**:
  - The "Allow users to send Slash commands and messages from the messages tab" toggle in App Home is required for traditional bot DMs [**MODERATE**]
  - Manifest-based app creation may not automatically set all App Home toggles correctly [**WEAK**]

### [SRC-012] Manage app agents and assistants | Slack Help Center
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://slack.com/help/articles/33077521383059-Manage-app-agents-and-assistants
- **Verified**: yes
- **Relevance**: 3
- **Summary**: Admin-facing guide for managing AI agent apps at the workspace or organization level. Workspace owners/admins can enable or disable the "AI agent experience" or "AI assistant experience" per app. When enabled, AI apps appear at the top of Slack. Requires paid plan. This is a server-side gate: even correctly configured apps will not function if the admin has disabled the AI experience.
- **Key Claims**:
  - Workspace admins can enable/disable the AI agent/assistant experience per app [**MODERATE**]
  - AI assistant apps require a paid Slack plan [**MODERATE**]
  - Admin-level disabling overrides app-level configuration [**WEAK**]

### [SRC-013] apps.manifest.validate method | Slack Developer Docs
- **Authors**: Slack (Salesforce)
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.slack.dev/reference/methods/apps.manifest.validate/
- **Verified**: yes
- **Relevance**: 3
- **Summary**: Documents the manifest validation API method. Returns structured error responses with `message` and `pointer` fields indicating the exact location of schema violations. Common validation errors include: missing request URL or socket mode for event subscriptions, missing request URL for interactivity. The `invalid_manifest` error type wraps all schema violations.
- **Key Claims**:
  - Manifest validation returns structured errors with JSON pointers to the failing field [**STRONG**]
  - "Event Subscription requires either Request URL or Socket Mode Enabled" is a common validation error [**MODERATE**]
  - The `apps.manifest.validate` method can pre-validate before deployment [**MODERATE**]

## Thematic Synthesis

### Theme 1: The Assistant API Is a Distinct Surface, Not Just a Bot Enhancement

**Consensus**: The Slack Assistant API (branded "Agents & AI Apps") introduces a fundamentally different interaction model from traditional bot DMs. When enabled, it replaces the App Home Messages tab with Chat and History tabs, provides a split-view panel within channels, and uses its own event lifecycle (`assistant_thread_started`, `assistant_thread_context_changed`) alongside `message.im`. This is not an incremental improvement to bot messaging but a new surface with its own entry point, thread model, and UI. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-004], [SRC-007], [SRC-008]

**Controversy**: Whether traditional bot DMs and assistant-mode DMs can coexist. [SRC-007] explicitly states that enabling Agents & AI Apps replaces the Messages tab, making the two mutually exclusive. [SRC-011] describes the traditional "messages tab" toggle as the fix for DM issues, but this only applies to non-assistant apps. Developers must choose one model or the other.
**Dissenting sources**: [SRC-011] describes traditional App Home message tab configuration as the solution, while [SRC-001] and [SRC-007] indicate that the assistant API replaces this entirely.

**Practical Implications**:
- Choose between traditional bot DM model and assistant API model at app design time; they cannot be combined
- If using the assistant API, do not configure `messages_tab_enabled` in the manifest -- it will be superseded
- The split-view panel means users interact with your app in-context within channels, not by navigating to a DM

**Evidence Strength**: STRONG

### Theme 2: The Configuration Stack Has Multiple Independent Gates That Must All Be Open

**Consensus**: A working Slack assistant app requires alignment across at least five independent configuration layers: (1) the "Agents & AI Apps" feature toggle in app settings, (2) correct OAuth scopes (`assistant:write`, `chat:write`, `im:history`), (3) event subscriptions (`assistant_thread_started`, `assistant_thread_context_changed`, `message.im`), (4) an event delivery mechanism (Request URL or Socket Mode), and (5) workspace admin enablement of the AI experience. Failure at any single layer produces silent failure or the "sending messages turned off" error. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-006], [SRC-009], [SRC-010], [SRC-012], [SRC-013]

**Practical Implications**:
- Use `apps.manifest.validate` to catch scope/event mismatches before deployment
- After adding scopes, always reinstall the app -- new scopes do not activate without reinstallation
- Check workspace admin settings if the app is correctly configured but users still cannot interact
- The `assistant:write` scope is auto-added by the feature toggle but `chat:write` and `im:history` must be explicitly added

**Evidence Strength**: STRONG

### Theme 3: Thread Context Management Is the Core Architectural Challenge

**Consensus**: The assistant thread model requires explicit context management because `message.im` events do not carry thread context. The `assistant_thread_started` event provides initial context (which channel the user is viewing), and `assistant_thread_context_changed` fires when they navigate, but this context must be stored and retrieved for every subsequent user message. The `ThreadContextStore` interface (with `get` and `save` methods) is the designated abstraction. [**STRONG**]
**Sources**: [SRC-001], [SRC-003], [SRC-004], [SRC-005], [SRC-008]

**Controversy**: Whether the default in-message-metadata context store is sufficient for production use. [SRC-003] notes that `DefaultThreadContextStore` works "in most cases" by storing context in message metadata, but [SRC-008] suggests custom storage for more sophisticated context persistence.
**Dissenting sources**: [SRC-003] argues default storage is sufficient for most cases, while [SRC-008] implies custom storage is needed for production-grade context management.

**Practical Implications**:
- Always call `saveThreadContext()` in both `threadStarted` and `threadContextChanged` handlers
- For production apps, evaluate whether the default metadata-based store meets latency and reliability requirements
- Call `conversations.info` to verify channel access before using context data, as the bot may not be a member of the user's current channel
- The `context.channel_id` in the event payload is the channel the user is viewing, not the DM channel where the thread lives

**Evidence Strength**: STRONG

### Theme 4: The "Sending Messages Turned Off" Error Has Multiple Root Causes Depending on App Type

**Consensus**: The "Sending messages to this app has been turned off" error is the most commonly reported Slack bot configuration issue. For traditional (non-assistant) bots, the primary cause is the missing "Allow users to send Slash commands and messages from the messages tab" toggle in App Home settings, followed by missing `chat:write` scope. For assistant-mode apps, the error landscape is different because the Messages tab is replaced entirely. [**MODERATE**]
**Sources**: [SRC-010], [SRC-011], [SRC-007], [SRC-001]

**Controversy**: The relative importance of causes. [SRC-010] identifies missing `chat:write` scope as "the single most common cause," while [SRC-011] identifies the App Home message tab toggle as the fix they needed. The discrepancy likely reflects different app configurations (assistant vs. traditional bot).
**Dissenting sources**: [SRC-010] emphasizes `chat:write` scope as primary fix, while [SRC-011] emphasizes the App Home toggle.

**Practical Implications**:
- For traditional bots: ensure App Home "messages tab" is enabled and "Allow users to send" is toggled on
- For assistant apps: ensure "Agents & AI Apps" feature toggle is enabled (this replaces the messages tab)
- For both: verify `chat:write` scope is present and app has been reinstalled after scope changes
- Check admin-level settings ([SRC-012]) if per-app configuration is correct but the error persists
- Use `auth.test` API to verify token validity as a debugging step

**Evidence Strength**: MODERATE

### Theme 5: Manifest-Driven Configuration Enables Reproducibility but Has Validation Gaps

**Consensus**: App manifests (YAML or JSON) are the recommended way to configure Slack apps, enabling version-controlled, git-tracked, environment-reproducible configuration. Slack provides `apps.manifest.validate` for pre-deployment checking. However, manifests do not capture all configuration state -- the "Agents & AI Apps" feature toggle appears to be a dashboard-only setting, and scope/event mismatch errors are common. [**MODERATE**]
**Sources**: [SRC-002], [SRC-009], [SRC-013], [SRC-011]

**Controversy**: Whether manifests fully represent app configuration. [SRC-009] and [SRC-002] present manifests as the single source of truth, while [SRC-011]'s experience suggests manifest creation did not set all required App Home toggles, and [SRC-001] notes the "Agents & AI Apps" toggle is enabled in "app settings" (not the manifest).
**Dissenting sources**: [SRC-009] treats manifests as complete configuration, while [SRC-011] found manifest creation left the App Home message tab disabled.

**Practical Implications**:
- Use manifests for reproducible configuration but verify dashboard-only settings manually
- Run `apps.manifest.validate` before every deployment to catch schema errors early
- After manifest updates, always reinstall the app to activate new scopes
- Keep the manifest in version control and treat it as infrastructure-as-code
- Be aware that the "Agents & AI Apps" toggle may need to be set outside the manifest

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Enabling "Agents & AI Apps" feature toggle provides split-view entry point and auto-adds `assistant:write` scope -- Sources: [SRC-001], [SRC-006]
- Three event subscriptions are required for assistant apps: `assistant_thread_started`, `assistant_thread_context_changed`, `message.im` -- Sources: [SRC-001], [SRC-003], [SRC-008]
- Required scopes are `assistant:write`, `chat:write`, and `im:history` -- Sources: [SRC-001], [SRC-003], [SRC-006]
- Event subscriptions require either Request URL or Socket Mode enabled -- Sources: [SRC-002], [SRC-013]
- `messages_tab_enabled` in the manifest controls traditional bot DM capability in App Home -- Sources: [SRC-002], [SRC-007]
- Enabling Agents & AI Apps replaces the Messages tab with Chat and History tabs -- Sources: [SRC-001], [SRC-007]
- No scopes are required to receive `assistant_thread_started` and `assistant_thread_context_changed` events -- Sources: [SRC-004], [SRC-005]
- Thread context is not included in `message.im` events; must be stored/retrieved via ThreadContextStore -- Sources: [SRC-003], [SRC-008]
- `context.channel_id` in assistant events represents the user's currently viewed channel, not the DM channel -- Sources: [SRC-004], [SRC-005]
- The `Assistant` class requires three handler callbacks: `threadStarted`, `threadContextChanged`, `userMessage` -- Sources: [SRC-001], [SRC-003]
- Manifest validation returns structured errors with JSON pointers to failing fields -- Sources: [SRC-013]
- `assistant:write` scope unlocks `setStatus`, `setSuggestedPrompts`, and `setTitle` methods -- Sources: [SRC-006]

### MODERATE Evidence
- Missing `chat:write` scope is the most common cause of the "sending messages turned off" error -- Sources: [SRC-010]
- App must be reinstalled after adding new scopes for them to take effect -- Sources: [SRC-010]
- The "Allow users to send Slash commands and messages from the messages tab" toggle is required for traditional bot DMs -- Sources: [SRC-011]
- Workspace admins can enable/disable the AI agent/assistant experience per app, overriding app config -- Sources: [SRC-012]
- AI assistant apps require a paid Slack plan -- Sources: [SRC-001], [SRC-012]
- Slack validates that event subscriptions have matching scopes and rejects inconsistent manifests -- Sources: [SRC-009]
- `assistant.threads.setStatus` accepts either `assistant:write` or `chat:write` scope -- Sources: [SRC-006]
- HTTP mode is recommended over Socket Mode for production deployments -- Sources: [SRC-008]
- Slash commands are not supported in the split view container -- Sources: [SRC-001]
- Workspace guests cannot access apps with Agents & AI Apps enabled -- Sources: [SRC-001]
- The DefaultThreadContextStore stores context via message metadata -- Sources: [SRC-003]

### WEAK Evidence
- Slack does not auto-create DM channels; bot or user must initiate first -- Sources: [SRC-010]
- Users can block apps via workspace settings, causing the DM error even with correct configuration -- Sources: [SRC-010]
- Manifest-based app creation may not automatically set all App Home toggles correctly -- Sources: [SRC-011]
- Admin-level disabling overrides app-level configuration -- Sources: [SRC-012]
- Bot must handle `not_in_channel` errors gracefully and join channels before retrying -- Sources: [SRC-008]

### UNVERIFIED
- The "Agents & AI Apps" feature toggle has no manifest-level equivalent and must be set via the dashboard -- Basis: inference from documentation gaps across [SRC-001], [SRC-002], [SRC-009]; no source explicitly confirms or denies manifest-level control
- OAuth authorization URLs must independently list all required scopes even when the manifest specifies them -- Basis: single GitHub issue [SRC-related: bolt-js #2437], not corroborated by official documentation

## Knowledge Gaps

- **Manifest representation of "Agents & AI Apps" toggle**: No source definitively documents whether this feature toggle has a manifest-level field or must always be set via the Slack dashboard. The manifest reference ([SRC-002]) does not list it, and the AI docs ([SRC-001]) describe it as a dashboard action, but explicit confirmation of its absence from the manifest schema is lacking.

- **Error taxonomy for assistant-mode apps**: The "sending messages turned off" troubleshooting literature ([SRC-010], [SRC-011]) predates or does not account for the assistant API model. There is no authoritative guide specifically for debugging assistant-mode DM failures as distinct from traditional bot DM failures.

- **Rate limits and quotas for assistant API methods**: While [SRC-001] mentions the `chat.update` 3-second rate limit, there is no comprehensive documentation of rate limits for `assistant.threads.setStatus`, `setSuggestedPrompts`, `setTitle`, or the streaming APIs.

- **Multi-workspace / Enterprise Grid behavior**: The assistant API documentation focuses on single-workspace setup. Behavior in Enterprise Grid (org-wide app deployment, cross-workspace context) is only touched on at the admin level ([SRC-012]) with no developer-facing detail.

- **Socket Mode reliability for assistant apps**: [SRC-008] recommends HTTP for production but provides no data on Socket Mode limitations specific to the assistant API (event ordering, reconnection behavior, context loss).

## Domain Calibration

This review covers a domain with official documentation that is comprehensive but relatively new (2024-2025 vintage). The high proportion of STRONG and MODERATE claims reflects the density of official Slack documentation, but much of the troubleshooting and architectural guidance comes from community sources. The assistant API is evolving rapidly -- the transition from `api.slack.com` to `docs.slack.dev` domains was ongoing during research, and some features (streaming, feedback blocks) appear to be recent additions. Treat findings as current but verify against the latest documentation before implementing.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research slack-assistant-api` on 2026-03-25.
