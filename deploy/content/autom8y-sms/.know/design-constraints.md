---
domain: design-constraints
generated_at: "2026-03-25T12:09:30Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "9934462"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Design Constraints

## Tension Catalog Completeness

### TENSION-001: Dual Orchestrator Pattern -- Conversation vs. Reminder

**Location**: `src/autom8_sms/services/orchestrator.py` (ConversationOrchestrator) and `src/autom8_sms/services/reminder/orchestrator.py` (ReminderOrchestrator)

Two orchestrators exist with entirely separate construction, lifecycle, and dependency chains. They share the same `TwilioClient` and both interact with the data service, but there is no shared base class or interface. The naming `ConversationOrchestrator` lives in `services/orchestrator.py` while `ReminderOrchestrator` lives in `services/reminder/orchestrator.py` -- asymmetric nesting with no corresponding `services/conversation/` namespace.

**Naming mismatch**: The package is `autom8_sms` (underscore, "sms") while the pyproject.toml distribution name is `autom8y-sms` (hyphen, "y"). The installed package name (`autom8y-sms`) differs from the Python import namespace (`autom8_sms`).

### TENSION-002: Scheduling Client as Optional -- Two Incompatible Modes in One Class

**Location**: `src/autom8_sms/services/orchestrator.py`, line 144

`ConversationOrchestrator` carries `scheduling_client: DataSchedulingProtocol | None = None` as a dataclass field. This makes the class semantically split into two modes (scheduling-enabled, scheduling-disabled), but both modes are served by the same class. Seven `# type: ignore[union-attr]` comments are the load-bearing scar of this design: callers must treat the field as not-None after passing a guard, but the type system cannot prove it. ADR-007 documents this explicitly.

### TENSION-003: PILOT_OFFICE_PHONES Evaluated Twice in Separate Paths

**Location**: `src/autom8_sms/handlers/client_lead.py` line 67 and `src/autom8_sms/services/orchestrator.py` line 1101

The pilot gate is enforced at two independent points: once in the Lambda handler (via `_get_pilot_phones()` at module load time, stored as `PILOT_PHONES`) and once inside `ConversationOrchestrator._is_scheduling_enabled()` (reads `PILOT_OFFICE_PHONES` at call time via `os.environ.get`). The handler gate filters which conversations are processed at all; the orchestrator gate controls whether scheduling tools are injected into the prompt. These are logically related but separately maintained.

### TENSION-004: SmsDataClient Wraps Interop Client but Bypasses Its Protocol for get_business

**Location**: `src/autom8_sms/clients/data_service.py`, lines 267-297

`SmsDataClient` delegates most operations to `DataMessageClient` via the interop protocol. But `get_business()` uses a private accessor (`_get_http_client()`, lines 322-332) to obtain the underlying HTTP transport and calls `/api/v1/business/{phone}` directly -- bypassing the interop protocol entirely. The `_get_http_client()` method inspects attribute names (`http`, `_http`) through `hasattr` duck typing. This is structurally fragile: interop internals changing would silently break `get_business()`.

### TENSION-005: BusinessInfo Defined in conversation.py -- Name/Location Mismatch

**Location**: `src/autom8_sms/models/conversation.py`, lines 95-110

`BusinessInfo`, `LeadInfo`, `AddressInfo`, and `MessageRecord` are all defined in `models/conversation.py`. The file name implies it owns conversation-specific models, but it is actually the primary domain model module for all data entities. A parallel module `models/twilio.py` handles only the single `TwilioSmsResponse` model.

### TENSION-006: Duplicate SmsDataClient/DataServiceClient Alias

**Location**: `src/autom8_sms/clients/data_service.py`, line 352; `src/autom8_sms/clients/__init__.py`

`DataServiceClient = SmsDataClient` is a backward-compatible alias retained in the module. Both names are exported from `clients/__init__.py`. The alias persists without a documented deprecation timeline.

### TENSION-007: ConversationOrchestrator._get_conversation_id() Imports from Handler

**Location**: `src/autom8_sms/services/orchestrator.py`, lines 1163-1168

The orchestrator service layer reaches up into the handler layer to import `conversation_id_var` from `autom8_sms.handlers.client_lead`. This is a layering violation: services should not import from handlers. The code silently falls back to a generated UUID on `ImportError` or `LookupError`, meaning the layering violation is semi-optional.

### TENSION-008: console/_genai_attrs.py is a Shim with Explicit TODO for Removal

**Location**: `src/autom8_sms/console/_genai_attrs.py`, lines 1-8

The file is a local shim for `autom8y_telemetry.genai` which does not exist in the installed version. The TODO marker acknowledges the shim is load-bearing until the upstream SDK evolves. No issue tracker reference or timeline.

### TENSION-009: get_conversation_context Overrides Interop at the Parse Layer

**Location**: `src/autom8_sms/clients/data_service.py`, lines 214-244

`SmsDataClient.get_conversation_context()` calls the interop SDK then immediately `model_dump()`s the result and re-parses every sub-object into SMS-specific typed models. This double-parse happens on every conversation. The motivation is domain isolation (SMS needs `twil_phone`, which is not in the interop model).

### TENSION-010: Config Split Across Three Classes with Different env-prefix Conventions

**Location**: `src/autom8_sms/config.py` and `src/autom8_sms/services/reminder/config.py`

`TwilioConfig` uses `TWILIO_` prefix. `SmsServiceConfig` uses `SMS_` prefix but reads `AUTOM8Y_ENV` and `AUTOM8Y_DATA_URL` bypassing the prefix via `AliasChoices`. `ReminderConfig` uses `REMINDER_` prefix but also reads six env vars that bypass the `REMINDER_` prefix.

---

## Trade-off Documentation

### TENSION-001 Trade-off: Dual Orchestrators

- **Current state**: Two orchestrators with independent construction, no shared interface.
- **Ideal state**: A shared `BaseOrchestrator` or separate service concerns.
- **Why current persists**: The reminder workflow (EventBridge-triggered, no Claude, batch loop) and conversation workflow (webhook-triggered, Claude + tool dispatch loop) are structurally different enough that a shared base would be thin.

### TENSION-002 Trade-off: Optional Scheduling Client

- **Current state**: `scheduling_client: DataSchedulingProtocol | None = None` with 7 `type: ignore` suppressions.
- **Ideal state**: Two separate orchestrator subclasses or a Strategy pattern injection.
- **Why current persists**: ADR-007 explicitly chose the optional-field approach to avoid refactoring the orchestrator constructor API while scheduling was in pilot mode.

### TENSION-003 Trade-off: Double Pilot Gate

- **Current state**: Handler filters conversations before dispatch; orchestrator re-checks the same env var for tool injection.
- **Ideal state**: A single gating layer or explicit passing of the gate result.
- **Why current persists**: The two checks serve distinct purposes (which conversations to process vs. which tool set to use). Defense-in-depth by design.

### TENSION-004 Trade-off: Protocol Bypass for get_business

- **Current state**: `_get_http_client()` uses duck-typing attribute inspection to reach the underlying HTTP transport.
- **Ideal state**: `DataMessageProtocol` exposes a `get_business()` method.
- **Why current persists**: The `autom8y-interop` library's protocol covers messages/leads/addresses/context but not business records. Adding to interop requires cross-repo coordination.

### TENSION-007 Trade-off: Handler-to-Service Import

- **Current state**: `ConversationOrchestrator._get_conversation_id()` imports from the handler layer.
- **Ideal state**: Correlation ID passed as a parameter into `process_conversation()`.
- **Why current persists**: The correlation ID is a `ContextVar` set by the handler's per-invocation logic. Threading it as a parameter would require call-stack changes.

### ADR-Referenced Trade-offs

- **ADR-007**: Optional scheduling client with `# type: ignore` suppressions instead of class hierarchy.
- **ADR-008**: Claude tool_use messages use raw dicts (`list[Any]`) rather than typed models to preserve Anthropic protocol fidelity.
- **ADR-006**: Timezone read from `context.address.timezone`; raises `ValueError` when absent rather than defaulting to UTC.
- **ADR-002**: `SmsDataClient` circuit breaker thresholds (`cb_failure_threshold=3, cb_recovery_timeout=30s`) are tighter than SDK defaults for faster Lambda fail-over.
- **ADR-WS1-001**: `extra="ignore"` on response models survives upstream field additions at the cost of silently discarding unknown fields.
- **ADR-WS-ATOMIC-003**: Reschedule resolves `appointment_id` server-side rather than accepting from Claude's tool input.

---

## Abstraction Gap Mapping

### GAP-001: No Shared Scheduling Client Lifecycle Interface

**Type**: Missing abstraction

Both `ConversationOrchestrator.from_env()` and `ReminderOrchestrator.from_env()` call `resolve_scheduling_client()` with overlapping but not identical parameters. No shared factory or lifecycle manager exists.

**Locations**: `src/autom8_sms/services/orchestrator.py` lines 151-165; `src/autom8_sms/services/reminder/orchestrator.py` lines 55-76.

### GAP-002: Tool Dispatch Is a Large if/elif Chain Without Registry

**Type**: Missing abstraction (dispatcher registry)

`_dispatch_tool_call()` in `src/autom8_sms/services/orchestrator.py` routes via `if/elif` over 5 tool names. Adding a new tool requires editing three locations: `tools/definitions.py` (schema), `tools/validation.py` (allowlist), and the dispatch chain.

### GAP-003: Formatted Result Types Duplicated Between Typing and Runtime

**Type**: Dual representation

`src/autom8_sms/tools/models.py` defines typed Pydantic models describing JSON shapes produced by formatting functions, but the formatters produce raw `dict` objects serialized to JSON -- they do not instantiate the Pydantic models. The typed models exist for documentation and test validation only.

### GAP-004: Prompt Version Management Is Manual

**Type**: Missing automation

`PROMPT_VERSION` in both prompt modules must be manually bumped when prompt text changes. No automated check enforces this.

### GAP-005: console/_genai_attrs.py Is a Zombie Abstraction

**Type**: Zombie abstraction (shim)

Exists solely because `autom8y_telemetry.genai` does not yet exist upstream. Has explicit `TODO` but no removal timeline.

### GAP-006: SmsServiceConfig.system_prompt Is Unused

**Type**: Zombie field

`src/autom8_sms/config.py` defines `system_prompt` as a configurable field, but the orchestrator never reads it -- prompts come from the `prompts/` package.

---

## Load-Bearing Code Identification

### LB-001: _dispatch_tool_call -- Central Tool Routing

**Location**: `src/autom8_sms/services/orchestrator.py`, lines 478-560

This single method routes all five scheduling tools. Changing its signature or name requires coordinated updates to 5 handler methods, test fixtures, `tools/models.py`, and `console/_instrument.py` (which monkey-patches it at line 201).

### LB-002: SmsDataClient.get_conversation_context -- Context Assembly

**Location**: `src/autom8_sms/clients/data_service.py`, lines 195-247

Every conversation processing call flows through this. Assembles `ConversationContext` from interop's composite endpoint. All prompt selection, tool enablement, and scheduling logic depends on this object being correctly populated.

### LB-003: TwilioClient._build_http_client -- Circuit Breaker Config

**Location**: `src/autom8_sms/clients/twilio.py`, lines 52-71

Circuit breaker configured inline with hardcoded defaults. Referenced by ADR-002.

### LB-004: ALLOWED_TOOL_NAMES Allowlist

**Location**: `src/autom8_sms/tools/validation.py`, lines 38-46

Must stay synchronized with `tools/definitions.py` and the `_dispatch_tool_call` chain. No test asserts the allowlist equals the set of defined tools.

### LB-005: MessageRecord nullable fields for backward compat

**Location**: `src/autom8_sms/models/conversation.py`, lines 37-43

`prompt_version`, `input_tokens`, `output_tokens`, `model_id` are all `| None` with WS-3 backward compat comment. Making them non-nullable would require a data migration.

### LB-006: conversation_id_var ContextVar

**Location**: `src/autom8_sms/handlers/client_lead.py` line 30

Referenced by `ConversationOrchestrator._get_conversation_id()` via a deferred import. If moved or renamed, the fallback UUID generation activates silently, and log correlation breaks.

### LB-007: Console _dispatch_tool_call Monkey-patch

**Location**: `src/autom8_sms/console/_instrument.py` lines 201, 222

Depends on: (1) method being an instance method, (2) Python attribute assignment allowing override, (3) the signature not changing. All three are currently true; any becoming false breaks the console silently.

---

## Evolution Constraint Documentation

### Changeability Ratings

| Area | Rating | Notes |
|------|--------|-------|
| `prompts/sdr_prompt.py` -- prompt text | safe | String constant; bump `PROMPT_VERSION` manually |
| `prompts/scheduled_prompt.py` -- prompt text | safe | Same pattern |
| `tools/definitions.py` -- tool schemas | coordinated | Adding/removing tools requires allowlist and dispatch chain changes |
| `models/conversation.py` -- MessageRecord nullable fields | migration | Removing nullability requires data migration |
| `models/conversation.py` -- ConversationContext structure | coordinated | Every prompt builder and context assembler depends on its shape |
| `clients/data_service.py` -- `get_business()` | coordinated | Bypasses interop protocol; breaks silently if interop internals change |
| `services/orchestrator.py` -- `_dispatch_tool_call` | frozen | Monkey-patched by `_instrument.py`; rename breaks instrumentation |
| `config.py` -- `SmsServiceConfig.system_prompt` | safe (zombie) | Not used; can be removed |
| `console/_genai_attrs.py` | migration | Remove when `autom8y-telemetry` ships `genai` submodule |
| Lambda handler entry points (`handlers/`) | frozen | AWS EventBridge points to these by name |
| `pyproject.toml` entry point `sms-console` | frozen | Published CLI entry point |
| `DataServiceClient` alias | frozen | Backward-compat obligation until all callers migrated |

### External Dependency Constraints

- `autom8y-ai[anthropic]>=1.3.0`: `AIClientProtocol` is the only interface to Claude. Any API changes require orchestrator changes.
- `autom8y-interop>=1.1.0`: `DataMessageClient`, `resolve_scheduling_client`, `resolve_message_client` are the integration boundary. Protocol shape determines available data.
- `autom8y-telemetry[otlp]>=0.6.0`: `trace_decision`, `trace_validation`, `record_side_effect` called throughout orchestrator.
- `requires-python = ">=3.12,<3.13"`: Pinned to exactly Python 3.12.

---

## Risk Zone Mapping

### RISK-001: Silent calendar_link_override Fallback

**Location**: `src/autom8_sms/clients/data_service.py`, line 236

If `lead_raw` is None or the upstream response omits `calendar_link`, this returns `None` silently. No warning is logged. A change in the interop API removing this key would produce a silent regression.

**Cross-reference**: TENSION-004

### RISK-002: BusinessInfo(**parsed.data) -- Unvalidated Dict Spread

**Location**: `src/autom8_sms/clients/data_service.py`, line 297

`parsed.data` is typed as `dict[str, Any]`. Extra keys not in the model are silently ignored (ADR-WS1-001). If the API response adds a field that conflicts with a Python reserved word, this could fail at runtime.

### RISK-003: PILOT_PHONES Module-Level Evaluation

**Location**: `src/autom8_sms/handlers/client_lead.py`, line 82

Evaluated once at Lambda cold start. If `PILOT_OFFICE_PHONES` env var is updated between cold starts, the running container uses the stale pilot set. The orchestrator reads the env var at call time, creating an inconsistency window.

**Cross-reference**: TENSION-003

### RISK-004: _get_http_client() Duck-Type Attribute Inspection

**Location**: `src/autom8_sms/clients/data_service.py`, lines 322-332

If neither `http` nor `_http` exists, the method returns `self._msg` itself. A subsequent `http.get_async(...)` call would raise `AttributeError` at runtime.

**Cross-reference**: TENSION-004

### RISK-005: Tool Loop Exhaustion Fallback Hides Bugs

**Location**: `src/autom8_sms/services/orchestrator.py`, lines 462-476

When `MAX_TOOL_LOOP_ITERATIONS` (5) is reached, the orchestrator returns `_build_fallback_response()` -- a graceful user-facing message. Tool loop exhaustion is not marked as `ProcessingResult.success = False` and is invisible to alerting systems.

### RISK-006: Timezone Resolution Raises ValueError -- Cascades Through Dispatch

**Location**: `src/autom8_sms/services/orchestrator.py`, lines 1108-1130

`_get_business_timezone()` raises `ValueError` when `address.timezone` is not configured (ADR-006). Propagates into `_dispatch_tool_call()`'s `except Exception` handler, which returns a fallback. If timezone is unconfigured for a business, scheduling tools are silently unavailable with no operator alerting.

---

## Knowledge Gaps

- ADR documents referenced inline (ADR-002, ADR-006, ADR-007, ADR-008, ADR-ENV-NAMING-CONVENTION, ADR-WS1-001, ADR-WS-ATOMIC-002, ADR-WS-ATOMIC-003, ADR-WS-LOOKUP-001) were not found in the repository -- they may live in an external decisions registry. The inline references are the only record of their content available in this codebase.
- The `autom8y-sms-test` package is an external test fixture dependency. Whether it exports `DataServiceClient` by name (driving the backward-compat alias) could not be confirmed.
- The `config.service.system_prompt` field appears unused but could not be confirmed by grep across all code paths including potential dynamic callers.
- Console subsystem (`console/`) is excluded from coverage and its risk zone behavior under async cancellation (RISK-006) is speculative.
- Lambda infrastructure coupling (timeout budgets, memory limits, concurrency) is referenced in comments but the actual Terraform/CDK/SAM configuration is not in this repository.
