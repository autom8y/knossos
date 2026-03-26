---
domain: feat/llm-client-abstraction
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/llm/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# LLM Client Abstraction Layer

## Purpose and Design Rationale

Shared transport for all Clew Haiku callsites (BC-01). Transport-only: API key management, HTTP transport. ZERO prompt engineering -- callers own their prompts. Minimal interface: `Complete(ctx, req) (string, error)`. Three parallel LLM client interfaces exist by design: llm.Client (triage/summarization), response.ClaudeClient (reasoning with structured output), knowledge.LLMClient (narrow, RR-007 import isolation).

## Conceptual Model

Single-method `Client` interface. `CompletionRequest`: SystemPrompt, UserMessage, MaxTokens, Model. `AnthropicClient`: wraps Go SDK, creates new client per-call, API key validated at construction. `MockClient`: shared test double imported by triage and other tests.

## Implementation Map

`internal/llm/client.go` (interface + AnthropicClient, 152 lines), `client_test.go` (MockClient + unit tests, 126 lines). Default model: claude-haiku-4-5, max tokens: 800. Consumers: triage.Orchestrator, conversation.LLMSummarizer, knowledge index builder. Wired as single shared instance in serve.go.

## Boundaries and Failure Modes

Missing API key at startup: hard error from constructor; callers disable triage/knowledge. API call failure: wrapped error, callers fail-open. Empty response: returns ("", nil), callers must handle. ToolUseBlock in response: JSON-marshaled gracefully. No retry/backoff (SDK-level only). No streaming. No temperature control.

## Knowledge Gaps

1. Per-call client construction may have performance implications
2. Model pinning (claude-haiku-4-5) requires code change for deprecation
3. search/knowledge bridge adapter type not confirmed
