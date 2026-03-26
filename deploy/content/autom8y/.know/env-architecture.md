---
domain: env-architecture
generated_at: "2026-03-15T12:00:00Z"
expires_after: "30d"
source_scope:
  - "/Users/tomtenuta/Code/a8/scripts/a8-devenv.sh"
  - ".a8/autom8y/org-secrets.conf"
  - ".a8/autom8y/env.defaults"
  - ".a8/autom8y/secretspec.toml"
generator: manual
confidence: 0.97
format_version: "1.0"
update_mode: "manual"
incremental_cycle: 0
max_incremental_cycles: 0
---

# Environment Architecture

## 1. Overview

This file documents the runtime environment and secret loading architecture for the autom8y
ecosystem. It covers config resolution, the 6-layer env loading model, the org-secrets SSM
overlay, all three cache mechanisms, discovery logic, and known risk areas.

**Four source files** are covered:

| File | Location | Role |
|------|----------|------|
| `a8-devenv.sh` | `/Users/tomtenuta/Code/a8/scripts/a8-devenv.sh` (1598 lines) | Core library — symlinked to `~/.config/direnv/lib/a8.sh` |
| `ecosystem.conf` | `.a8/autom8y/ecosystem.conf` | Identity and org config; sourced at load time |
| `org-secrets.conf` | `.a8/autom8y/org-secrets.conf` | SSM manifest (13 entries) |
| `env.defaults` | `.a8/autom8y/env.defaults` | Layer 1 non-secret ecosystem defaults |
| `secretspec.toml` | `.a8/autom8y/secretspec.toml` | Secret inventory for validation |

**Not covered**: Terraform bridge (`tf-bridge.conf`), worktree trust mechanics
(`_a8_direnv_trust_worktree`), the legacy `required-secrets.conf`, CodeArtifact token
usage details. See `.ledge/reviews/env-architecture-audit.md` for full forensic detail.

---

## 2. System Architecture

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  SOURCE FILES                                                                ║
║                                                                              ║
║  ~/.config/direnv/lib/a8.sh ──symlink──> /a8/scripts/a8-devenv.sh           ║
║  .a8/autom8y/ecosystem.conf       (identity; config resolution target)       ║
║  .a8/autom8y/org-secrets.conf     (SSM manifest: 13 entries)                 ║
║  .a8/autom8y/env.defaults         (Layer 1 committed non-secrets)            ║
║  .a8/autom8y/secrets.shared       (Layer 2 encrypted offline fallback)       ║
║  .env/defaults                    (Layer 3 project non-secrets)              ║
║  .env/secrets                     (Layer 4 project encrypted secrets)        ║
║  .env/${AUTOM8Y_ENV}              (Layer 5 per-env gitignored overrides)     ║
║  .envrc.local                     (Layer 6 per-developer gitignored)        ║
╚══════════════════════════════════════════════════════════════════════════════╝
                              |
                              v  direnv enters directory
╔══════════════════════════════════════════════════════════════════════════════╗
║  RUNTIME (direnv shell)                                                      ║
║                                                                              ║
║  Auto-execute: _a8_resolve_config -> _a8_generate_functions                  ║
║  .envrc: use_a8 -> _a8_use_handler                                           ║
║                                                                              ║
║  Override chain (later wins):                                                ║
║  Layer 1 < Layer 2 < Layer 3 < Layer 4 < Layer 5 < Layer 6                  ║
║                                               SSM overwrites Layer 2 values ║
║                                                                              ║
║  Caches (per-session, filesystem):                                           ║
║  ~/.cache/a8_codeartifact_token   (11h TTL)                                  ║
║  ~/.cache/a8_dotenvx_key          (24h TTL)                                  ║
║  ~/.cache/a8_secret_*             (24h TTL, 1 file per manifest entry)       ║
╚══════════════════════════════════════════════════════════════════════════════╝
                              |
                              v  AWS API calls (non-blocking on failure)
╔══════════════════════════════════════════════════════════════════════════════╗
║  AWS / EXTERNAL                                                              ║
║                                                                              ║
║  AWS SSM Parameter Store                                                     ║
║    /autom8y/platform/dotenvx/private-key   (dotenvx decrypt key)            ║
║    /autom8y/platform/<relative_path>       (13 org secrets)                 ║
║                                                                              ║
║  AWS CodeArtifact                                                            ║
║    Domain: ${A8_CODEARTIFACT_DOMAIN}       (UV index token)                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

**Key rule**: SSM runs after Layer 2. `_a8_load_org_secrets` is called by `_a8_use_handler`
AFTER `_a8_load_env` completes. SSM exports are unconditional — they always overwrite
the `secrets.shared` fallback values loaded in Layer 2.

---

## 3. 6-Layer Loading Model

All layers load inside `_a8_load_env` (lines 364–454), except Layer 6 which loads in
`_a8_use_handler` (line 1531). Later layers override earlier layers.

| Layer | Name | File Path | Scope | Committed? | Encrypted? | Absent Behavior |
|-------|------|-----------|-------|------------|------------|-----------------|
| 1 | Ecosystem non-secrets | `${_A8_MANIFEST_DIR}/env.defaults` | Org-wide | Yes | No (plaintext) | Silent skip |
| 2 | Ecosystem secrets (offline fallback) | `${_A8_MANIFEST_DIR}/secrets.shared` | Org-wide | Yes | Yes (dotenvx) | Silent skip; `log_warn` if key missing |
| 3 | Project non-secrets | `.env/defaults` | Per-repo | Yes | No (plaintext) | Silent skip |
| 4 | Project secrets | `.env/secrets` | Per-repo | Yes | Yes (dotenvx) | Silent skip; `log_warn` if key missing |
| 5 | Per-environment overrides | `.env/${AUTOM8Y_ENV}` | Per-repo per-env | No (gitignored) | No | Silent skip |
| 6 | Developer local overrides | `.envrc.local` | Per-repo per-dev | No (gitignored) | No | Silent skip (direnv built-in) |

**SSM Authority Note**: `_a8_load_org_secrets` runs after all 6 layers complete. It
exports org secrets directly via `export "${ev}=${val}"` with no skip-if-set guard.
This unconditional export means SSM values always overwrite Layer 2 values. If SSM is
unreachable and cache is absent, the Layer 2 `secrets.shared` values persist as fallback.

**Layer 1 known variables** (from `env.defaults`):
`LOG_LEVEL`, `DEBUG`, `CLOUDFLARE_ACCOUNT_ID`, `GRAFANA_URL`, `AMP_QUERY_ENDPOINT`,
`GRAFANA_LOKI_URL`, `GRAFANA_LOKI_INSTANCE_ID`, `GRAFANA_TEMPO_URL`,
`GRAFANA_TEMPO_OTLP_ENDPOINT`, `GRAFANA_TEMPO_INSTANCE_ID`, `VPC_ID`, `SUBNET_IDS`,
`SECURITY_GROUP_IDS`, `MYSQL_SECRET_ARN`, `AUTH_DB_SECRET_ARN`, `MONITORING_EMAIL`

**Layer 2 variables**: Same set as `org-secrets.conf` entries — this is the encrypted
offline copy of the 13 SSM-authoritative values.

**Layer loading mechanism**:
- Plaintext layers: `set -a; source "$file"; set +a`
- Encrypted layers: `_a8_source_encrypted "$file"` — auto-detects `DOTENV_PUBLIC_KEY`
  marker; dispatches to `dotenvx get --format eval` or plain source

**Legacy shim** (lines 436–441): `.env/shared` sources after Layer 5 if present.
Not a numbered layer. Comment marks it for removal post-migration. Overrides Layer 5 —
counterintuitive ordering.

---

## 4. Config Resolution

`_a8_resolve_config` (lines 110–233) discovers and sources `ecosystem.conf`. It is
idempotent: the session cache guard at line 112 (`[ "${_A8_CONFIG_RESOLVED:-}" = "1" ]`)
makes all subsequent calls within a direnv session no-ops.

### Precedence Levels (A → B → 2.5 → C)

| Priority | Level | Trigger | Path Pattern | Multi-match Behavior |
|----------|-------|---------|--------------|----------------------|
| Highest | A | `A8_ECOSYSTEM` env var set | `.a8/${A8_ECOSYSTEM}/ecosystem.conf` | N/A — exact path |
| 2 | B | `A8_ECOSYSTEM` unset; glob in CWD | `.a8/*/ecosystem.conf` | 0 matches: fall through; 2+ matches: `log_error` + return 1 |
| 3 | 2.5 | Levels A and B found nothing | Walk CWD → `/` looking for `.a8/*/ecosystem.conf` | First match wins |
| Lowest | C | All above failed | `~/.config/a8/*/ecosystem.conf` | 0 matches: `log_error` + return 1; 2+ matches: `log_error` + return 1 |

**Level 2.5 use case**: Satellite repos that sit under a workspace root containing the
monorepo's `.a8/` directory inherit org config without needing their own `.a8/` tree.

### `ecosystem.conf` Required Keys

| Key | Required? | Default if absent | Purpose |
|-----|-----------|-------------------|---------|
| `A8_ORG_NAME` | Hard required | None — `log_error` + return 1 | Org identity; drives all function names |
| `A8_AWS_ACCOUNT` | Required | None — `log_error` | AWS account number |
| `A8_AWS_REGION` | Required | None — `log_error` | AWS region |
| `A8_CODEARTIFACT_DOMAIN` | Required | None — `log_error` | CodeArtifact domain name |
| `A8_ENV_CANONICAL` | Optional | `local staging production test` | Valid env names |
| `A8_SSM_PREFIX` | Optional | `/${A8_ORG_NAME}` (e.g., `/autom8y`) | SSM path prefix |

### Derived Internal Variables (set by `_a8_resolve_config`)

| Variable | Derivation |
|----------|------------|
| `_A8_ORG_NAME` | `${A8_ORG_NAME}` |
| `_A8_AWS_ACCOUNT` | `${A8_AWS_ACCOUNT:-}` |
| `_A8_AWS_REGION` | `${A8_AWS_REGION:-}` |
| `_A8_CODEARTIFACT_DOMAIN` | `${A8_CODEARTIFACT_DOMAIN:-}` |
| `_A8_ENV_CANONICAL` | `${A8_ENV_CANONICAL:-local staging production test}` |
| `_A8_SSM_PREFIX` | `${A8_SSM_PREFIX:-/${A8_ORG_NAME}}` |
| `_A8_MANIFEST_DIR` | `ecosystem.conf` directory path (stripped of `/ecosystem.conf`) |
| `_A8_CONFIG_PATH` | Full path to `ecosystem.conf` |
| `_A8_ENV_VAR_NAME` | `$(printf '%s' "${_A8_ORG_NAME}" \| tr '[:lower:]' '[:upper:]')_ENV` → e.g., `AUTOM8Y_ENV` |
| `_A8_CONFIG_RESOLVED` | `1` (session cache flag) |

---

## 5. Execution Sequence

```
direnv enters directory
  |
  v
sources ~/.config/direnv/lib/a8.sh  (= a8-devenv.sh)
  |
  v
Auto-execute block (lines 1587-1598)
  |-- [1] _a8_resolve_config        (find + source ecosystem.conf; set _A8_* vars)
  |-- [2] _a8_generate_functions    (eval 10 public API functions into shell scope)
  |
  v
.envrc calls: use_a8 [--env=NAME]
  |
  v
_a8_use_handler (lines 1472-1532)
  |-- [3]  _a8_trust_mise           (mise trust + source ~/.mise/env)
  |-- [4]  _a8_get_dotenvx_key      (SSM fetch or cache -> export DOTENV_PRIVATE_KEY)
  |-- [5]  _a8_load_env             (Layers 1-5 + legacy shim)
  |          |-- Layer 1: source env.defaults
  |          |-- Layer 2: _a8_source_encrypted secrets.shared
  |          |-- Layer 3: source .env/defaults
  |          |-- Layer 4: _a8_source_encrypted .env/secrets
  |          |-- Layer 5: source .env/${resolved_env}
  |          |-- [legacy]: source .env/shared (if present)
  |-- [6]  _a8_load_org_secrets     (3-pass SSM batch fetch -> exports 13 secrets)
  |-- [7]  _a8_tf_env_bridge        (export TF_VAR_environment if TF files present)
  |-- [8]  _a8_get_artifact_token   (AWS CodeArtifact token -> cache -> UV index creds)
  |-- [9]  _a8_resolve_tool_tokens  (gh auth token -> GITHUB_TOKEN)
  |-- [10] _a8_activate_venv        (uv .venv activation)
  |-- [11] _a8_validate_secrets     (warn-only secretspec audit)
  |-- [12] source_env_if_exists .envrc.local   (Layer 6)
```

**Step 4 before Step 5**: `_a8_get_dotenvx_key` runs before `_a8_load_env` because
Layer 2 and Layer 4 need `DOTENV_PRIVATE_KEY` to decrypt. If step 4 fails (SSM
unreachable, cache expired), steps 5-Layer-2 and 5-Layer-4 emit `log_warn` and skip
the encrypted files.

**Step 6 after Step 5**: SSM runs after all 6 layers. SSM values overwrite Layer 2.

---

## 6. Org Secrets (SSM)

### SSM Path Convention

```
Full SSM path = ${_A8_SSM_PREFIX}/platform/${ssm_relative_path}

For autom8y:
  _A8_SSM_PREFIX = /autom8y   (leading slash is part of the value)
  Full path example: /autom8y/platform/grafana/auth-token
```

### Manifest (`org-secrets.conf`) — 13 Entries

| SSM Relative Path | Env Var | Category |
|-------------------|---------|----------|
| `grafana/auth-token` | `AUTOM8Y_GRAFANA_AUTH` | Grafana |
| `cloudflare/api-token` | `AUTOM8Y_CLOUDFLARE_API_TOKEN` | Cloudflare |
| `slack/alerts-bot-token` | `AUTOM8Y_OBSERVABILITY_SLACK_BOT_TOKEN` | Slack |
| `slack/automation-bot-token` | `AUTOM8Y_SLACK_BOT_TOKEN` | Slack |
| `grafana/loki-api-key` | `AUTOM8Y_GRAFANA_LOKI_API_KEY` | Grafana Loki |
| `grafana/tempo-api-key` | `AUTOM8Y_GRAFANA_TEMPO_API_KEY` | Grafana Tempo |
| `grafana/tempo-api-key-write` | `AUTOM8Y_GRAFANA_TEMPO_API_KEY_WRITE` | Grafana Tempo |
| `google/gcal-sa-key-json` | `GOOGLE_SA_KEY_JSON` | Google Calendar |
| `auth/service-api-key` | `SERVICE_API_KEY` | S2S Auth |
| `meta/app-id` | `AUTOM8Y_META_APP_ID` | Meta Ads |
| `meta/app-secret` | `AUTOM8Y_META_APP_SECRET` | Meta Ads |
| `meta/access-token` | `AUTOM8Y_META_ACCESS_TOKEN` | Meta Ads |
| `meta/account-id` | `AUTOM8Y_META_ACCOUNT_ID` | Meta Ads |

**Note**: The ads satellite uses `ADS_META_*` var names for the same SSM paths. The
manifest uses `AUTOM8Y_META_*`. The satellite renaming happens outside this manifest.

### 3-Pass Algorithm (`_a8_load_org_secrets`, lines 868–1000)

**Pass 1 — Read manifest, check caches** (lines 883–932):
Reads `org-secrets.conf` line by line (`IFS=: read -r ssm_path env_var`). For each
valid entry, stores in indexed vars (`_os_ssm_paths_N`, `_os_env_vars_N`). Checks
`~/.cache/a8_secret_${sanitized_path}` for 24h validity. Cache hit: store value in
`_os_values_N`, set `_os_need_fetch_N=0`. Cache miss/expired: `_os_need_fetch_N=1`.

**Pass 2 — Batch fetch uncached secrets** (lines 934–966):
- **Batch path** (jq available): Collects all `_os_need_fetch_N=1` entries into
  `batch_paths`; calls `_a8_fetch_org_secrets_batch $batch_paths` (unquoted — word
  splitting is intentional for multiple `--names` args). Batch function writes cache
  files but does NOT export env vars.
- **Sequential fallback** (no jq): Calls `_a8_fetch_org_secret "$p" "$v"` per entry.
  This function writes cache AND exports the env var directly.

**Pass 3 — Export env vars** (lines 968–997):
Re-reads freshly written (or previously cached) cache files to populate `val`. Exports
`export "${ev}=${val}"` only when `val` is non-empty (line 989 guard). Cleans up all
`_os_*` indexed variables. Logs debug count.

### Offline Behavior (SSM Unreachable)

1. `_a8_get_dotenvx_key` runs first. SSM unreachable: `DOTENV_PRIVATE_KEY` unset unless cached.
2. `_a8_load_env` runs. Layer 2 decryption fails with `log_warn` if key unset; OR succeeds from cache.
3. `_a8_load_org_secrets` runs. All SSM fetches fail silently (`2>/dev/null || true`).
4. Cache valid (< 24h): values served from cache; env vars exported normally.
5. Cache absent or expired AND SSM unreachable: no export. Layer 2 `secrets.shared` values persist.
6. No `log_warn` or `log_error` for SSM unreachability itself — completely silent degradation.

---

## 7. Cache Mechanisms

### Cache 1: CodeArtifact Token

| Property | Value |
|----------|-------|
| File | `~/.cache/a8_codeartifact_token` |
| TTL | 11 hours (39600 seconds) — 1-hour buffer before AWS 12h expiry |
| Permissions | `chmod 600` |
| Set by | `_a8_get_artifact_token` (lines 472–511) |
| Format | `CACHED_TOKEN='<value>'` + `TOKEN_EXPIRY=<unix_timestamp>` |
| Scope guard | None — single shared file for all orgs (ADR-081) |
| Post-read unset | No — `CACHED_TOKEN` and `TOKEN_EXPIRY` may persist in shell scope |

### Cache 2: dotenvx Private Key

| Property | Value |
|----------|-------|
| File | `~/.cache/a8_dotenvx_key` |
| TTL | 24 hours (86400 seconds) |
| Permissions | `chmod 600` |
| Set by | `_a8_get_dotenvx_key` (lines 578–619) |
| Format | `CACHED_DOTENVX_KEY='<value>'` + `KEY_EXPIRY=<unix_timestamp>` |
| Scope guard | Skips entirely if `DOTENV_PRIVATE_KEY` already set (CI injection path) |
| Post-read unset | No — `CACHED_DOTENVX_KEY` and `KEY_EXPIRY` may persist in shell scope |

### Cache 3: Org Secrets (Per-Entry)

| Property | Value |
|----------|-------|
| File pattern | `~/.cache/a8_secret_<sanitized_relative_path>` |
| TTL | 24 hours (86400 seconds) |
| Permissions | `chmod 600` |
| Set by | `_a8_fetch_org_secret` (lines 694–745) or `_a8_fetch_org_secrets_batch` (lines 769–844) |
| Format | `CACHED_VALUE='<value>'` + `KEY_EXPIRY=<unix_timestamp>` |
| Scope guard | None — SSM always overwrites |
| Post-read unset | Yes — `unset CACHED_VALUE KEY_EXPIRY` called after read (lines 719, 928, 985) |
| Count | 1 file per `org-secrets.conf` entry → 13 files for autom8y |

### Sanitized Path Examples

```
grafana/auth-token       -> ~/.cache/a8_secret_grafana_auth-token
cloudflare/api-token     -> ~/.cache/a8_secret_cloudflare_api-token
slack/alerts-bot-token   -> ~/.cache/a8_secret_slack_alerts-bot-token
google/gcal-sa-key-json  -> ~/.cache/a8_secret_google_gcal-sa-key-json
meta/app-id              -> ~/.cache/a8_secret_meta_app-id
```

Rule: `ssm_relative_path` with `/` replaced by `_`. Prefix stripped at
`${_A8_SSM_PREFIX}/platform/` before sanitizing.

### All Cache Files: Shared Format

```bash
CACHED_<FIELDNAME>='<single-quote-escaped_value>'
<KEY|TOKEN>_EXPIRY=<unix_epoch_seconds>
```

Single-quote escape: `"${value//\'/\'\\\'\'}"` — replaces `'` with `'\''`.
`mkdir -p "$HOME/.cache"` called before every write.

---

## 8. Secret Validation

### Two Modes

| Mode | Function | Trigger | Behavior |
|------|----------|---------|----------|
| Automatic warn-only | `_a8_validate_secrets` (lines 1327–1420) | Every `use_a8` call for non-local envs | `log_warn` only; never returns 1; never modifies env |
| Manual verbose | `_a8_check_secrets` (lines 1436–1453) | `just check-secrets` | Full secretspec output to stdout; returns 1 if secretspec.toml or binary missing |

Profile selector: `${A8_SECRETSPEC_PROFILE:-human}` (undocumented knob).

### `_a8_validate_secrets` Three-Tier Fallback

| Tier | Condition | Mechanism |
|------|-----------|-----------|
| 1 | `secretspec` binary in PATH + `secretspec.toml` exists | `secretspec check --profile $profile` |
| 2 | No `secretspec` binary but `secretspec.toml` exists | Manual TOML parser using `sed` (BSD-compatible `[[:space:]]`); reads only `[profiles.human]` section |
| 3 | Neither `secretspec.toml` nor binary | Legacy `required-secrets.conf` fallback (if file present) |

Dynamic var check in Tiers 2 and 3: `eval "val=\"\${${var_name}:-}\""`.
Tier 2 sed filter: `[A-Z_][A-Z0-9_]*` — provides injection mitigation for var names.

---

## 9. Discovery Mechanisms

### Repo Discovery (`_a8_discover_repos`, lines 1234–1259)

| Tier | Condition | Source |
|------|-----------|--------|
| 1 (explicit override) | `~/.config/a8/repos` file exists | One absolute path per line; `#` comments skipped |
| 2 (convention scan) | File absent | Sibling dirs of monorepo root matching `${_A8_ORG_NAME}-*/` that contain `.envrc` |

Tier 2 base: `$(git rev-parse --show-toplevel)` → parent dir → `autom8y-*/` glob.
Called only by `_a8_switch_env` for cross-repo status display.

### Service Discovery (`_a8_discover_services`, lines 1273–1286)

- Scan: `${monorepo_root}/services/*/`
- Filter: Contains `docker-compose.override.yml`
- Exclude: Directories named `_template`
- Output: Basenames only

### SDK Discovery (`_a8_discover_sdks`, lines 1300–1310)

- Scan: `${monorepo_root}/sdks/python/*/`
- Filter: Contains `pyproject.toml`
- Output: Basenames only

All three discovery functions require a git repo (`git rev-parse --show-toplevel`).
All return 0 on any failure.

---

## 10. Error Handling Reference

### Hard Failures (return 1)

| Condition | Function | Message |
|-----------|----------|---------|
| `ecosystem.conf` not found (all 4 levels) | `_a8_resolve_config` | Full setup guidance |
| Multiple `ecosystem.conf` matches | `_a8_resolve_config` | `log_error` + "set A8_ECOSYSTEM" guidance |
| `A8_ECOSYSTEM` set but path doesn't exist | `_a8_resolve_config` | `log_error` with create guidance |
| `A8_ORG_NAME` empty after sourcing | `_a8_resolve_config` | `"ecosystem.conf sourced but A8_ORG_NAME is empty"` |
| Required key missing in `ecosystem.conf` | `_a8_validate_config` | `"ecosystem.conf missing required keys: ${missing}"` |
| `_a8_generate_functions` with empty `_A8_ORG_NAME` | `_a8_generate_functions` | `"_a8_generate_functions called but _A8_ORG_NAME is empty"` |
| Non-canonical env name at switch time | `_a8_validate_env` | Two `log_error` messages |
| `secretspec.toml` missing (`check_secrets`) | `_a8_check_secrets` | `log_error` + return 1 |
| `secretspec` binary missing (`check_secrets`) | `_a8_check_secrets` | `log_error` + return 1 |

### Warnings (log_warn, continue)

| Condition | Function |
|-----------|----------|
| Non-canonical env name at load time | `_a8_load_env` — continues with that name (transition safety) |
| dotenvx key missing | `_a8_source_encrypted` |
| dotenvx binary missing | `_a8_source_encrypted` — includes install instructions |
| dotenvx decryption failure | `_a8_source_encrypted` |
| Malformed `org-secrets.conf` line | `_a8_load_org_secrets` |
| `uv sync` failure | `_a8_activate_venv` |
| `uv` binary missing | `_a8_activate_venv` |
| Unknown flag to `use_a8` | `_a8_use_handler` — continues (transition safety) |
| Config resolution failure at source time | Auto-execute block (emits 2x `log_warn`) |

### Silent Skips (return 0)

| Condition | Pattern |
|-----------|---------|
| SSM unreachable (any call) | `2>/dev/null \|\| true` |
| Any layer file absent | `if [ -f "$file" ]` guard |
| `_a8_fetch_org_secret` missing args | Returns 0 after `log_warn` |

### Asymmetric Env Name Behavior

Load-time non-canonical env: `log_warn` only, execution continues. This is intentional
for migration safety — new env names can be deployed before canonical list is updated.
Switch-time non-canonical env: `log_error` + return 1. Deliberate user action requires
a valid name.

---

## 11. Scar Tissue and Known Constraints

### IMPLICIT-DEVENV-01 (line 132)

**Bash 3.2 glob nullglob**: Non-matching globs expand to the literal string, not empty.
All glob loops use `[ -f "$f" ]` guard to filter out literal-string results.

```bash
for f in .a8/*/ecosystem.conf; do
    [ -f "$f" ] && matches+=("$f")  # SCAR TISSUE -- IMPLICIT-DEVENV-01
done
```

Appears at: line 132 (CWD scan), line 151 (global fallback scan),
line 1257 (sibling repo scan), line 1564 (worktree `.envrc` scan), line 1575 (worktree `.a8/` scan).

### IMPLICIT-DEVENV-02 (line 230)

**No associative arrays in bash 3.2**: Session cache stored as scalar `_A8_CONFIG_RESOLVED=1`.
Checked at `_a8_resolve_config` entry. Makes all subsequent calls within a direnv session no-ops.

### IMPLICIT-DEVENV-03 (line 269)

**`$@` escaping in heredoc**: Inside `eval "$(cat <<GENERATE_EOF ...)"`, `$@` must be
written as `\$@` to defer expansion to function call time. Literal `$@` would expand to
empty at heredoc expansion time.

### IMPLICIT-DEVENV-04 (line 270)

**Heredoc delimiter indentation**: `<<GENERATE_EOF` and its closing `GENERATE_EOF` must
be at column 0. Leading whitespace on the delimiter breaks bash 3.2 heredoc parsing.

### IMPLICIT-DEVENV-05 (line 223)

**No `${var^^}` in bash 3.2**: Uppercasing uses `tr '[:lower:]' '[:upper:]'`. Result
cached in `_A8_ENV_VAR_NAME` at config resolution time — single `tr` call per session.

### Additional Patterns (No Formal Marker)

| Pattern | Locations | Reason |
|---------|-----------|--------|
| Indexed vars as array substitute | `_a8_load_org_secrets`, `_a8_fetch_org_secrets_batch` | No `declare -A` in bash 3.2 |
| `eval "val=\"\${${key}:-}\""` for dynamic var read | Lines 46, 370, 911–914, 940–963, 1099, 1118, 1144, 1343, 1383–1386, 1409–1411, 1507–1508 | No `${!varname}` in bash 3.2 |
| `compgen -G` for glob existence check | `_a8_tf_env_bridge` lines 1148–1150 | Bash-specific but available since 3.2 |
| Unquoted `$chunk_paths` in `aws ssm get-parameters --names` | `_a8_fetch_org_secrets_batch` line 803 | Intentional word splitting for multiple path args |

### Known Undocumented Constraints

1. **Single `DOTENV_PRIVATE_KEY`**: All encrypted files (`secrets.shared`, `.env/secrets`)
   share one key. dotenvx per-file keys (e.g., `DOTENV_PRIVATE_KEY_SECRETS_SHARED`) are
   not used. Changing key strategy requires coordinating all encrypted files.

2. **`A8_SECRETSPEC_PROFILE` knob**: Overrides the secretspec profile used for validation
   (default: `human`). Not documented in `secretspec.toml` or `env.defaults`.

3. **direnv built-in dependency**: `watch_file`, `PATH_add`, `source_env_if_exists` are
   direnv built-ins assumed in scope. The script cannot be sourced outside direnv.

4. **`_A8_SSM_PREFIX` leading slash**: Default value is `/${A8_ORG_NAME}` — the leading
   slash is part of the value, not a path separator at construction time.

5. **`CACHED_TOKEN` / `CACHED_DOTENVX_KEY` scope leak**: Cache 1 and Cache 2 do NOT
   call `unset` after reading. These vars persist in shell scope. Cache 3 does unset.

---

## 12. Risk Areas

| # | Risk | Severity | Locations | Testing Target |
|---|------|----------|-----------|----------------|
| R1 | `eval`-based dynamic var access — injection if input contains metacharacters | MEDIUM | Lines 46, 370, 911–914, 940–963, 1099, 1118, 1144, 1343, 1507–1508 | Validate `org-secrets.conf` env var names match `[A-Z_][A-Z0-9_]*` |
| R2 | Cache file sourcing — symlink or malicious write enables code execution | MEDIUM | Lines 485, 590, 714, 922, 981–983 | Verify cache files are regular files before source; consider format validation |
| R3 | `DOTENV_PUBLIC_KEY` filter in `_a8_source_encrypted` uses deny-list, not allow-list | LOW-MEDIUM | Lines 649–651 | Validate dotenvx output before eval; consider allow-list `grep -E '^[A-Z_][A-Z0-9_]*='` |
| R4 | Batch SSM fetch: `$chunk_paths` unquoted — spaces in SSM path would corrupt args | LOW | Lines 801–833 | Validate SSM path format in manifest parser; confirm no spaces in `org-secrets.conf` |
| R5 | Legacy `.env/shared` overrides Layer 5 (counterintuitive; blocks env-specific values) | LOW | Lines 436–441 | Audit all ecosystem repos for `.env/shared` presence; plan migration |
| R6 | `use_{org}()` backward compat entry points — silently break if `_A8_ORG_NAME` changes | LOW | Lines 257–287, 1592–1594 | Audit `.envrc` files; count `use_autom8y` vs `use_a8`; migrate to `use_a8` |
| R7 | Sequential vs batch SSM path inconsistency in Pass 2/3 of `_a8_load_org_secrets` | LOW | Lines 934–997 | Test with jq absent; verify env vars match batch-mode output |
| R8 | `_A8_ENV_VAR_NAME` used before assignment if `_a8_load_env` called without `_a8_resolve_config` | LOW | Lines 365, 1488–1490 | Add guard in `_a8_load_env`; test direct call path |

---

## 13. Knowledge Gaps

This document does not cover:

- `tf-bridge.conf` schema and `_a8_tf_bridge_all` runtime behavior — see audit report
- `_a8_direnv_trust_worktree` — called post-worktree-creation, not in normal `.envrc` flow
- `required-secrets.conf` legacy format — superseded by `secretspec.toml`; no instance observed in current manifest dir
- CodeArtifact token usage downstream of export (UV index credential mechanics)
- Cross-satellite variable renaming convention (`AUTOM8Y_META_*` vs `ADS_META_*`)
- Multi-ecosystem workspace setup (when `A8_ECOSYSTEM` override is needed)

For full forensic detail on all functions, line ranges, and undocumented behaviors, see
`.ledge/reviews/env-architecture-audit.md`.

---

## 14. Health Check Tools

Three CLI tools validate ecosystem env/secret health without requiring manual
code reading. All live under the monorepo `scripts/` directory.

### `a8-doctor.sh` -- Unified Health Check

Single command that composes all validators into a pass/fail report.

```bash
bash scripts/a8-doctor.sh              # human + JSON output
bash scripts/a8-doctor.sh --json-only  # machine-parseable JSON only
bash scripts/a8-doctor.sh --strict     # warnings become failures in secretspec lint
bash scripts/a8-doctor.sh --check config_resolution  # run one check only
```

**Checks**: config_resolution, credential_drift, secretspec_lint, dotenvx, env_audit.
**Exit codes**: 0 = all pass/skip, 1 = any fail, 2 = script error.
**Offline**: credential drift degrades to SKIP; other checks still run.

### `a8-credential-drift-check.sh` -- Credential Drift Detection

Compares SSM Parameter Store values against decrypted `secrets.shared` for every
entry in `org-secrets.conf`. Reports match/drift/ssm_only/shared_only status.
Never prints secret values -- only character lengths.

```bash
bash scripts/a8-credential-drift-check.sh             # full report
bash scripts/a8-credential-drift-check.sh --json-only  # CI mode
```

**Exit codes**: 0 = match, 1 = drift detected, 2 = error, 3 = all skipped.

### `lint_secretspec_xr.py` -- Cross-Repo Secretspec Lint

Validates workspace-root `secretspec.toml` consistency against satellite repos.
Rules: XR-E001 (required flag), XR-E002 (default value), XR-E003 (shared secret
coverage), XR-W001 (description drift), XR-W002 (profile mismatch), XR-W003
(cross-service naming). Requires Python 3.11+ (tomllib).

```bash
python3 scripts/lint_secretspec_xr.py          # advisory (always exit 0)
python3 scripts/lint_secretspec_xr.py --strict  # exit 1 on errors
python3 scripts/lint_secretspec_xr.py --json    # structured output for CI
```

### CI Integration

Run `a8-doctor.sh --json-only` as a merge gate. In CI without AWS credentials,
credential drift auto-skips (exit 0). For strict enforcement, add `--strict`.
Parse JSON output with `jq '.summary.fail'` to gate on failure count.
