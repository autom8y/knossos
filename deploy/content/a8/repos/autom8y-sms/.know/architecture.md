---
domain: architecture
generated_at: "2026-03-25T12:09:30Z"
expires_after: "7d"
source_scope:
  - "./src/**/*.py"
  - "./app/**/*.py"
  - "./pyproject.toml"
generator: theoros
source_hash: "9934462"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Architecture

**Language**: Python 3.12
**Package**: `autom8_sms` (source at `src/autom8_sms/`)
**Build system**: Hatchling via `pyproject.toml`
**Entry point (CLI)**: `sms-console` -> `autom8_sms.console.__main__:main`
**Deployment target**: AWS Lambda (two functions), invoked by EventBridge on a schedule
**Package version**: 0.1.0 -- SMS conversation service for client leads, described as "Zapier replacement"

---

## Package Structure

The source tree under `src/autom8_sms/` has the following top-level packages (directories) and standalone modules:

### Top-level module: `src/autom8_sms/__init__.py`
- **Purpose**: Version marker and architecture docstring. Describes the high-level flow: `Twilio Webhook -> Lambda -> Claude API -> Twilio API` with `autom8_data` for message storage.
- **Exports**: `__version__ = "0.1.0"`

### `src/autom8_sms/config.py`
- **Purpose**: Configuration layer. Defines three classes:
  - `TwilioConfig` (pydantic `BaseSettings`, prefix `TWILIO_`): `account_sid`, `auth_token`
  - `SmsServiceConfig` (extends `Autom8yBaseSettings`, prefix `SMS_`): `autom8y_env`, `data_service_url`, `autom8y_auth_url`, `log_level`, `follow_up_delay_seconds`, `max_message_length`, `system_prompt`
  - `Config`: Composite container. `config.twilio` and `config.service` accessors.
- **Pattern**: Global lazy singleton via `get_config()` / `_config: Config | None`.

### `src/autom8_sms/handlers/` -- Lambda entry points
- `__init__.py`: Re-exports `lambda_handler` from `client_lead`.
- `client_lead.py`: Lambda handler for conversation processing. EventBridge `rate(1 minute)`. Polls pending conversations, applies pilot phone filter (`PILOT_OFFICE_PHONES` env var), per-conversation timeout guard (default 45s), and delegates to `ConversationOrchestrator.process_conversation()`. Contains `conversation_id_var: ContextVar[str | None]` for structured log correlation.
- `appointment_reminder.py`: Lambda handler for appointment reminders. EventBridge `rate(5 minutes)`. Delegates entirely to `ReminderOrchestrator.run()`. Emits a dead-man's-switch CloudWatch metric via `emit_success_timestamp(NAMESPACE)`.

### `src/autom8_sms/services/` -- Domain orchestration
- `__init__.py`: Empty.
- `orchestrator.py` (hub module): `ConversationOrchestrator` -- central business logic. Manages the full SMS response pipeline. Key sub-operations: context fetch -> prompt selection -> tool resolution -> `_generate_response()` tool_use dispatch loop -> send SMS -> log outbound. Contains `GenerationResult` and `ProcessingResult` dataclasses. Imports from `clients`, `models`, `prompts`, `tools`, and `autom8y_interop`.
- `reminder/__init__.py`: Empty.
- `reminder/orchestrator.py`: `ReminderOrchestrator` -- serial send loop for appointment reminders. Steps: calculate window -> query eligible appointments -> for each: format -> send SMS -> mark sent.
- `reminder/config.py`: `ReminderConfig` (`Autom8yBaseSettings`, prefix `REMINDER_`). Singleton via `get_reminder_config()` using `@lru_cache(maxsize=1)`.
- `reminder/templates.py`: `format_reminder_message()` -- formats appointment datetime in business-local timezone using `zoneinfo`.

### `src/autom8_sms/models/` -- Domain models (Pydantic)
- `conversation.py`: Core domain models:
  - `MessageRecord`: SMS message. Has `direction` field (inbound/outbound/outbound-claude/outbound-scheduled). Carries WS-3 instrumentation fields: `prompt_version`, `input_tokens`, `output_tokens`, `model_id`.
  - `LeadInfo`: Lead data (phone, office_phone, full_name, calendar_link, status).
  - `AddressInfo`: Business address with `format_full_address()` and IANA `timezone` field.
  - `BusinessInfo`: Business/chiropractor record. Contains `twil_phone` -- the Twilio sending number (critical: fetched per-lead from the business record, not from config).
  - `ConversationContext`: Aggregate. Contains `messages`, `lead`, `business`, `address`, `lead_status`, `next_appointment_datetime`, `calendar_link_override`. Properties: `office_phone`, `twil_phone`, `office_location`, `calendar_link`. Method: `to_claude_messages()`.
- `twilio.py`: `TwilioSmsResponse` -- Twilio API response model.

### `src/autom8_sms/clients/` -- External service adapters
- `data_service.py`: `SmsDataClient` -- wraps `autom8y_interop.DataMessageClient`. Provides `DataMessageProtocol` delegation (get_messages, create_message, get_lead, get_address, get_conversation_context, get_pending_conversations) plus SMS-specific extensions: `get_business()` (calls `/api/v1/business/{phone}`, returns `BusinessInfo` with `twil_phone`) and `get_recent_messages()`. Uses circuit breaker via `resolve_message_client()` with `SMS_DATA_` env prefix. Backward-compatible alias: `DataServiceClient = SmsDataClient`.
- `twilio.py`: `TwilioClient` -- direct HTTP to Twilio REST API (`https://api.twilio.com/2010-04-01`). Uses `autom8y_http.Autom8yHttpClient` (NOT the Twilio SDK, per ADR-AUTH-EMAIL-001). Circuit breaker enabled via `WP-CB-SMS`. Auth via HTTP Basic (`TWILIO_ACCOUNT_SID:TWILIO_AUTH_TOKEN`). One method: `send_sms(to, body, from_)`.
- `models.py`: `BusinessResponse` -- Pydantic envelope for `GET /api/v1/business/{phone}`. Field: `data: dict[str, Any] | None`. Extra fields ignored (ADR-WS1-001).

### `src/autom8_sms/tools/` -- AI tool definitions and dispatch
- `definitions.py`: Claude `tool_use` JSON schemas for the Anthropic API. Defines: `CHECK_AVAILABILITY_TOOL`, `BOOK_APPOINTMENT_TOOL`, `CANCEL_APPOINTMENT_TOOL`, `GET_APPOINTMENTS_TOOL`, `RESCHEDULE_APPOINTMENT_TOOL`. Convenience lists: `SCHEDULING_TOOLS`, `CANCEL_TOOLS`, `RESCHEDULE_TOOLS`.
- `models.py`: Typed models for dispatch results. `ToolDispatchResult` (frozen dataclass): `content: str`, `availability_response: AvailabilityResponse | None`. Also typed result shapes for use in tests: `FormattedAvailabilityResult`, `FormattedBookingResult`, `FormattedCancelResult`.
- `formatting.py`: Format interop response objects into JSON strings for Claude `tool_result` content. Functions: `format_availability_result()`, `format_booking_result()`, `format_cancel_result()`, `format_appointments_result()`, `format_reschedule_result()`.
- `validation.py`: Server-side input validation for AI-generated tool calls. SD-03 (booking_context allowlist), SD-04 (business hours -- checks slot match against last `AvailabilityResponse`), SD-05 (temporal bounds -- past/future clamp). Functions: `build_booking_context()`, `validate_temporal_bounds_availability()`, `validate_temporal_bounds_booking()`, `validate_business_hours()`. Constants: `MAX_TOOL_LOOP_ITERATIONS = 5`, `ALLOWED_TOOL_NAMES`.

### `src/autom8_sms/prompts/` -- Claude system prompts
- `__init__.py`: Re-exports `SDR_SYSTEM_PROMPT`, `SDR_PROMPT_VERSION`, `SCHEDULED_LEAD_PROMPT`, `SCHEDULED_PROMPT_VERSION`, `RESCHEDULE_ALLOWED`, `RESCHEDULE_STAFF_ONLY`.
- `sdr_prompt.py`: `SDR_SYSTEM_PROMPT` (string with `{placeholders}`). `PROMPT_VERSION = "sdr-v2"`. `SCHEDULING_TOOL_INSTRUCTIONS` -- injected when scheduling tools are enabled.
- `scheduled_prompt.py`: `SCHEDULED_LEAD_PROMPT` for leads with existing appointments. `PROMPT_VERSION = "scheduled-v2"`. `CANCEL_INSTRUCTIONS`, `RESCHEDULE_ALLOWED` (>48h), `RESCHEDULE_STAFF_ONLY` (<48h).

### `src/autom8_sms/narrative/` -- Observability plugin
- `__init__.py`: Empty.
- `rules.py`: `get_rules()` -- returns `NarrativeRuleSet` with 22 span-rendering rules for the `autom8y_devx` narrative engine. Registered via entry point `autom8y_devx.narrative_rules`. Handles: `gen_ai.decision`, `gen_ai.response_validation`, `gen_ai.tool_call.*`, `gen_ai.response_formatting`, `sms.webhook.*`, `sms.message.send`. Priority 10 (overrides core at priority 100).

### `src/autom8_sms/console/` -- Developer REPL (excluded from coverage)
- `__main__.py`: CLI entry point. Args: `--business` (required), `--lead` (optional). Creates `InstrumentedOrchestrator` wrapping a stub-based `ConversationOrchestrator`, then runs `ConsoleSession.run()`.
- `_stubs.py`: Creates console orchestrator with stub clients.
- `_session.py`: Interactive REPL loop.
- `_renderer.py`: Terminal display.
- `_instrument.py`: `InstrumentedOrchestrator` -- wraps orchestrator to capture traces.
- `_fork.py`, `_replay.py`, `_otel.py`, `_genai_attrs.py`, `_models.py`: Console-specific supporting modules.

### `src/autom8_sms/debug.py`
- **Purpose**: Developer debug entry point. Runs `orchestrator._generate_response()` against real context without sending SMS. Excluded from coverage (`omit` in `pyproject.toml`).

---

## Layer Boundaries

The import graph follows a strict downward-only dependency direction:

```
handlers/           (Lambda entry points -- AWS boundary)
    |
services/orchestrator.py  (hub: imports all domain packages)
services/reminder/orchestrator.py  (imports clients, tools subset)
    |
clients/             (external adapter layer)
    |
models/              (pure domain types, no internal imports)
prompts/             (pure string constants, no internal imports)
tools/               (definitions, validation, formatting -- no circular deps)
    |
config.py            (pure env config, leaf)
```

**Leaf packages** (no internal autom8_sms imports): `config.py`, `models/`, `prompts/`, `tools/` (mostly -- `tools/validation.py` imports only `autom8y_log`/`autom8y_telemetry`).

**Hub packages** (import many siblings): `services/orchestrator.py` imports from `clients/`, `models/`, `prompts/`, `tools/` -- it is the dependency hub of the codebase. `handlers/client_lead.py` imports from `clients/` and `services/`.

**External SDK layer** (leaf from project perspective): `autom8y_interop`, `autom8y_ai`, `autom8y_http`, `autom8y_log`, `autom8y_telemetry`, `autom8y_config` -- these are private PyPI packages from the `autom8y` CodeArtifact index.

**TYPE_CHECKING guard pattern**: Used extensively to break potential circular imports. `services/orchestrator.py` guards `AIClientProtocol`, `ToolCall`, `AvailabilityResponse`, `DataSchedulingProtocol` behind `if TYPE_CHECKING:`. `clients/data_service.py` guards `DataMessageProtocol`.

**Boundary enforcement**: No circular imports observed. `clients/` does not import from `services/`. `models/` and `prompts/` have no internal imports (pure leaf).

**`narrative/` is cross-cutting**: Imports `autom8_devx_types` (not `autom8_sms`). It is registered via Python entry points, not imported by the service itself.

---

## Entry Points and API Surface

### Lambda entry points (production)

1. **`lambda_handler`** in `src/autom8_sms/handlers/client_lead.py`
   - Trigger: EventBridge `rate(1 minute)`
   - Decorated with `@instrument_lambda` from `autom8y_telemetry.aws`
   - Calls `asyncio.run(process_pending_conversations())`
   - Returns `{"statusCode": 200|500, "body": {...}}`

2. **`lambda_handler`** in `src/autom8_sms/handlers/appointment_reminder.py`
   - Trigger: EventBridge `rate(5 minutes)`
   - Decorated with `@instrument_lambda`
   - Calls `asyncio.run(_run_reminders())`
   - Emits dead-man's-switch metric `Autom8y/SmsReminder`

### CLI entry point (developer tooling)

- **`sms-console`** -> `autom8_sms.console.__main__:main`
  - Flags: `--business` (required, E.164 phone), `--lead` (optional)
  - Starts interactive REPL against stub clients
  - Example: `uv run sms-console --business "+17203303721"`

- **`autom8_sms.debug`** (module, not script): Developer dry-run mode, prints Claude response without sending.

### Plugin entry point

- `autom8y_devx.narrative_rules` -> `sms = autom8_sms.narrative.rules:get_rules`
  - Consumed by the `autom8y_devx` observability console at runtime.

### Key exported interfaces

- `ConversationOrchestrator.process_conversation(lead_phone, new_message)` -- main business logic surface
- `ReminderOrchestrator.run()` -- reminder send loop
- `SmsDataClient.from_env()` -- data client factory (context manager)
- `TwilioClient.from_env()` -- Twilio client factory (context manager)
- `get_config() -> Config` -- global config singleton
- `get_rules() -> NarrativeRuleSet` -- observability plugin entry

---

## Key Abstractions

### 1. `ConversationContext` (`src/autom8_sms/models/conversation.py`)
The central data carrier in the service. Aggregates all data needed to process one conversation: message history, lead info, business info, address. Properties derive computed values: `twil_phone` (from business), `office_location` (formatted address), `calendar_link` (prefers BusinessOffer override over lead field). `to_claude_messages()` converts message history to Anthropic message format by mapping `direction` -> `role`.

### 2. `ConversationOrchestrator` (`src/autom8_sms/services/orchestrator.py`)
The hub class. Holds `claude_client: AIClientProtocol`, `twilio_client: TwilioClient`, `data_client: SmsDataClient`, `scheduling_client: DataSchedulingProtocol | None`. The `from_env()` factory conditionally creates `scheduling_client` only when `PILOT_OFFICE_PHONES` is set. The `_generate_response()` method runs the tool_use dispatch loop (up to `MAX_TOOL_LOOP_ITERATIONS = 5` iterations). Prompt selection is dynamic: SDR prompt for new leads, scheduled-lead prompt for leads with appointments.

### 3. `AIClientProtocol` (from `autom8y_ai`)
The interface `ConversationOrchestrator` depends on for Claude calls. Consumed via `AnthropicAdapter.from_env()` at runtime but typed as `AIClientProtocol` for testability (stub injection in console and tests).

### 4. `DataSchedulingProtocol` (from `autom8y_interop.data.protocols`)
Interface for scheduling operations. Methods: `check_availability`, `book_appointment`, `cancel_appointment`, `get_appointments`, `reschedule_appointment`, `mark_reminder_sent`, `get_reminder_eligible_appointments`. Only instantiated when scheduling is active (pilot gate).

### 5. `ToolDispatchResult` (`src/autom8_sms/tools/models.py`)
Frozen dataclass returned by every tool handler. `content: str` (JSON for Claude), `availability_response: AvailabilityResponse | None` (internal, stripped before message assembly). The availability response is carried through the dispatch loop for SD-04 business hours validation on subsequent booking calls.

### 6. `ValidationResult` (`src/autom8_sms/tools/validation.py`)
Frozen dataclass from each validation function. `valid: bool`, `error_message: str | None`, `clamped_end_date: date | None`. SD-05 may clamp (not reject) the end date on availability queries. SD-04 rejects booking if no prior availability check was performed (DEFECT-01 fix).

### 7. Prompt selection pattern (`services/orchestrator.py`)
`_get_prompt_version(context)` inspects `context.lead_status` and `context.next_appointment_datetime` to select prompt. `lead_status == "scheduled"` -> `SCHEDULED_LEAD_PROMPT`. Within 48 hours -> `RESCHEDULE_STAFF_ONLY`. Otherwise -> `RESCHEDULE_ALLOWED`. Default -> `SDR_SYSTEM_PROMPT`.

### 8. Tool set selection pattern (`services/orchestrator.py`)
`_resolve_tools(context)` selects tool lists based on lead status and pilot gate. Lead status `"scheduled"` with appointment < 48h -> `CANCEL_TOOLS`. Scheduled with appointment > 48h -> `RESCHEDULE_TOOLS`. New/unscheduled -> `SCHEDULING_TOOLS`. No scheduling client -> empty tools list.

### 9. `BusinessInfo.twil_phone` field
This is the critical field enabling SMS delivery. The Twilio "from" number is not in the global config -- it is fetched per-request from the business/chiropractor record via `SmsDataClient.get_business()` or the composite context endpoint. `ConversationContext.twil_phone` is a property that delegates to `business.twil_phone`.

### 10. `GenerationResult` dataclass
Returned by `_generate_response()`. Carries `response_text`, `prompt_version`, `input_tokens`, `output_tokens`, `model_id`. The `prompt_version` and token counts are persisted to the messages table for WS-3 SMS optimization instrumentation.

---

## Data Flow

### Primary flow: Inbound SMS -> Claude response

```
EventBridge (rate(1 min))
    -> lambda_handler (handlers/client_lead.py, @instrument_lambda)
    -> asyncio.run(process_pending_conversations())
    -> SmsDataClient.get_pending_conversations()          # polls /pending endpoint
    -> [pilot phone filter applied from PILOT_OFFICE_PHONES]
    -> ConversationOrchestrator.process_conversation(lead_phone, new_message)
        -> SmsDataClient.get_conversation_context(lead_phone)
            -> autom8y_interop.DataMessageClient.get_conversation_context()
                -> GET /api/v1/context/{lead_phone} (composite endpoint)
                    -> parsed into ConversationContext (messages, lead, business, address)
        -> _build_system_prompt(context)         # injects business context into prompt template
        -> _get_prompt_version(context)          # selects SDR or scheduled prompt
        -> _resolve_tools(context)               # selects SCHEDULING/CANCEL/RESCHEDULE/empty tools
        -> _generate_response(context, new_message)
            -> [tool_use dispatch loop, up to 5 iterations]
                -> claude_client.send_message_async(messages, system, max_tokens=300, tools)
                -> if stop_reason == "tool_use":
                    -> _dispatch_tool_call(tool_call, context, ...)
                        -> SD validation (SD-03, SD-04, SD-05)
                        -> DataSchedulingProtocol.check/book/cancel/reschedule/get()
                        -> format_*_result() -> JSON string
                    -> append tool_use + tool_result to messages
                    -> continue loop
                -> if stop_reason == "end_turn":
                    -> message_length truncation check (gen_ai.response_validation span)
                    -> return GenerationResult
        -> TwilioClient.send_sms(to=lead_phone, body=response, from_=context.twil_phone)
            -> POST https://api.twilio.com/2010-04-01/Accounts/{sid}/Messages.json
        -> SmsDataClient.create_message(outbound_message with instrumentation fields)
            -> POST /api/v1/messages (idempotent: 409 treated as success)
    -> ProcessingResult
    -> return {"statusCode": 200, "body": summary}
```

### Reminder flow

```
EventBridge (rate(5 min))
    -> lambda_handler (handlers/appointment_reminder.py)
    -> asyncio.run(_run_reminders())
    -> ReminderOrchestrator.run()
        -> calculate window: now -> now + window_hours (default 24h)
        -> DataSchedulingProtocol.get_reminder_eligible_appointments(window_start, window_end)
        -> for each eligible appointment:
            -> format_reminder_message(business_name, start_datetime, timezone, employee_name)
            -> [dry_run check]
            -> TwilioClient.send_sms(to=lead_phone, body=message, from_=office_phone)
            -> DataSchedulingProtocol.mark_reminder_sent(appointment_id)
                [Note: EC-7 narrow duplicate window if mark fails after send]
        -> emit_success_timestamp("Autom8y/SmsReminder")   # dead-man's-switch CloudWatch metric
    -> return summary dict
```

### Configuration data flow

```
Environment variables
    -> Pydantic BaseSettings validation at Lambda cold start
    -> TwilioConfig:  TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN
    -> SmsServiceConfig: AUTOM8Y_ENV, AUTOM8Y_DATA_URL, AUTOM8Y_AUTH_URL, SMS_SERVICE_LOG_LEVEL, etc.
    -> ReminderConfig: AUTOM8Y_ENV, AUTOM8Y_DATA_URL, AUTOM8Y_AUTH_URL, SERVICE_API_KEY,
                      TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, REMINDER_WINDOW_HOURS, REMINDER_DRY_RUN
    -> get_config() -> Config (lazy singleton, module-level _config)
    -> get_reminder_config() -> ReminderConfig (lru_cache singleton)
```

### Observability data flow

```
OpenTelemetry spans emitted during process_conversation():
    gen_ai.content.prompt        -> record_prompt_content (PII-gated)
    gen_ai.chat                  -> model, tokens, finish_reason
    gen_ai.response_validation   -> message_length check / SD-04 / SD-05
    gen_ai.content.completion    -> record_completion_content (PII-gated)
    gen_ai.decision              -> prompt_selection, tool_resolution, scheduling_instructions
    gen_ai.tool_call.*           -> per-tool spans via @trace_tool_call decorator
    gen_ai.response_formatting   -> per format_*_result() call
    sms.message.send             -> Twilio delivery, with record_side_effect()

narrative/rules.py -> NarrativeRuleSet (22 rules) -> consumed by autom8y_devx console
    renders human-readable descriptions of spans for developer observability UI
```

---

## Knowledge Gaps

1. **`console/` internals not fully read**: `_stubs.py`, `_session.py`, `_renderer.py`, `_instrument.py`, `_fork.py`, `_replay.py`, `_otel.py`, `_genai_attrs.py`, `_models.py` were not fully read. Console is excluded from coverage and is developer tooling only.

2. **`services/orchestrator.py` lines 650-end not read**: The `_handle_book_appointment`, `_handle_cancel_appointment`, `_handle_get_appointments`, `_handle_reschedule_appointment`, `_build_system_prompt`, `_build_fallback_response`, `_get_business_timezone`, `_get_prompt_version`, `_resolve_tools`, `_get_conversation_id`, and `close()` methods were not read due to file length. Their existence and signatures were inferred from call sites in read sections.

3. **`autom8y_interop` protocol shapes**: The exact signatures and field names of `DataMessageProtocol`, `DataSchedulingProtocol`, `AvailabilityResponse`, `BookingResponse`, `CancelResponse`, `RescheduleResponse`, `AppointmentListResponse` are inferred from usage -- not confirmed from source (private package).

4. **Dockerfile not examined**: Deployment configuration (Lambda packaging, layer setup, ADOT sidecar config) not documented. Referenced in tests (`test_dockerfile.py`) but not read.

5. **`autom8y_interop.data.lifecycle.resolve_scheduling_client`**: The `ReminderOrchestrator` uses this factory directly; the conversation orchestrator uses `autom8y_interop.data.resolve_scheduling_client`. Both differ from the `SmsDataClient`'s `resolve_message_client`. The distinction between these factories is not fully documented here.
