---
domain: workstation
generated_by: hygiene-sprint
last_updated: "2026-03-10"
---

# Workstation Module

Documentation for new developers on how the workstation module is structured and why tools live where they do.

## 1. Tool Ownership Table

| Tool Domain | Manager | Examples | Version Strategy |
|---|---|---|---|
| System dev tools | devbox (Nix) | jq, yq, shellcheck, shfmt, terraform, golangci-lint, yamllint, actionlint | Pinned per-repo in `devbox.json` |
| Language runtimes | mise | Go, Python, just, bash | Pinned in `.mise.toml` (project root) |
| UX/productivity tools | Homebrew | eza, bat, fd, ripgrep, fzf, starship, atuin, tmux, htop | Latest via `brew upgrade` |
| Python packages | uv / pyproject.toml | ruff, mypy, pytest (as dev deps) | Pinned in `pyproject.toml` |
| Dotfiles & identity | chezmoi | zshrc, gitconfig, gitignore_global, starship config, atuin config, mise config, LaunchAgent | Managed in `config/chezmoi/`, applied to `~/` |

Each manager owns its domain exclusively. Do not install system dev tools via Homebrew or language runtimes via devbox — the boundaries are intentional and prevent version drift.

## 2. Directory Ownership

```
config/workstation/          # Shell runtime files + Homebrew declarations
  Brewfile.platform          #   Tier 1: Required for a8 to function
  Brewfile.dev               #   Tier 2: Opinionated developer productivity
  Brewfile.personal.example  #   Tier 3: Template for user-specific (gitignored)
  scripts/cleanup.sh         #   LaunchAgent target (cron-like maintenance)
  zsh/                       #   Shell init, antidote plugins, helper functions
    a8-init.sh               #   Main shell init (sourced from ~/.zshrc)
    .zsh_plugins.txt         #   antidote plugin declarations
    plugins/                 #   Local platform plugins (a8, ari, just, git)

config/chezmoi/              # Dotfile source state (chezmoi manages)
  .chezmoi.toml.tmpl         #   Config template (promptStringOnce for git identity)
  .chezmoiroot               #   Points to home/ (chezmoi convention for nested repos)
  home/                      #   Maps to ~/  at chezmoi apply time
    dot_zshrc                #     → ~/.zshrc
    dot_gitconfig.tmpl       #     → ~/.gitconfig (Go template)
    dot_gitignore_global     #     → ~/.gitignore_global
    dot_config/              #     → ~/.config/
    private_Library/         #     → ~/Library/ (LaunchAgent)
    .chezmoiscripts/         #     One-time/on-change install scripts
```

Files in `config/workstation/` are sourced at runtime from the repo. Files in `config/chezmoi/home/` are deployed to `~/` by `chezmoi apply`. These are distinct responsibilities — do not conflate them.

## 3. Brewfile Tier System

Three tiers control what gets installed and when:

- **Platform** (`Brewfile.platform`): Tools required for a8 to function. If removing a tool would break `a8-devenv.sh`, `just dev-setup`, or CI, it belongs here. Installed first.
- **Dev** (`Brewfile.dev`): Opinionated productivity tools that any developer would benefit from. Not required for a8 to function. Installed second.
- **Personal** (`Brewfile.personal`, gitignored): User-specific tools. Copy `Brewfile.personal.example` to get started. Not committed to the repo.

When adding a new Homebrew dependency, pick the lowest tier that satisfies the requirement. A tool that only benefits your personal workflow does not belong in `Brewfile.platform`.

## 4. chezmoi `.chezmoiroot` Explained

The `config/chezmoi/.chezmoiroot` file contains `home`. This is a chezmoi convention for repositories where the chezmoi source state is not at the repository root. The `home/` subdirectory within `config/chezmoi/` is treated as if it were `~/.local/share/chezmoi/` — chezmoi maps its contents to `~/` at apply time.

The `dot_` prefix convention maps to `.` in the destination:

| Source file | Destination |
|---|---|
| `dot_zshrc` | `~/.zshrc` |
| `dot_gitconfig.tmpl` | `~/.gitconfig` |
| `dot_gitignore_global` | `~/.gitignore_global` |
| `dot_config/` | `~/.config/` |
| `private_Library/` | `~/Library/` |

The `private_` prefix marks files that chezmoi will create with mode `0600` (not world-readable). Use this for anything containing tokens or identity information.

## 5. Key Design Decision: What Lives Where

The module has two distinct file populations that serve different purposes:

- Files that are **deployed as dotfiles** (written to `~/`) live in `config/chezmoi/home/` and are managed by `chezmoi apply`. These are the entry points that a user's shell and tools load at startup.
- Files that are **sourced at runtime** from the repo (not deployed) live in `config/workstation/zsh/` and are referenced by `a8-init.sh` at shell startup.

This split exists because antidote plugins and the shell init file need to be sourced from a known repo path, not from `~/`. chezmoi manages the entry point (`~/.zshrc`) which then sources the repo-local runtime files. The result is that you can update `config/workstation/zsh/` files and they take effect on next shell open without running `chezmoi apply`.

**Decision rule**: If a file needs to exist at a `~/` path for a tool to find it, it belongs in `config/chezmoi/home/`. If a file is sourced or executed from within the repo by the shell init chain, it belongs in `config/workstation/zsh/`.

## 6. Version Pinning Strategy

`devbox.json` uses explicit `package@version` syntax (never `@latest`). The lock file (`devbox.lock`) is the ground truth for resolved Nix store paths.

**Current pinned versions:**

| Package | Version |
|---|---|
| jq | 1.8.1 |
| yq-go | 4.52.4 |
| shellcheck | 0.11.0 |
| shfmt | 3.12.0 |
| terraform | 1.14.6 |
| golangci-lint | 2.10.1 |
| yamllint | 1.37.1 |
| actionlint | 1.7.11 |

**Upgrade procedure**: Change the version in `devbox.json`, run `devbox install`, commit both `devbox.json` and `devbox.lock` together.

## 7. CI Integration Points

| Workflow | Job | What it does | Tool source |
|---|---|---|---|
| `go-ci.yml` | `shell-lint` | Runs shellcheck, shfmt, actionlint on scripts | `jetify-com/devbox-install-action@v0.12.0` (same versions as local dev) |
| `go-ci.yml` | `lint` | Go static analysis | `golangci/golangci-lint-action@v7` (separate from devbox, better caching) |
| `go-ci.yml` | `test` / `build` | Go test + build | `actions/setup-go@v5` (Go is a mise concern, not devbox) |
| `workstation-ci.yml` | `chezmoi-verify` | Validates chezmoi templates render | Installs chezmoi via official installer |

`workstation-ci.yml` triggers on push/PR to `config/chezmoi/**`. It runs `chezmoi doctor` and applies templates to a temp destination to catch rendering errors without side effects.

## 8. Test Coverage

Bats smoke tests live in `test/workstation/workstation-recipes.bats`. They exercise read-only just recipes only (non-destructive).

**What's covered:**

| Test | Assertion |
|---|---|
| `workstation-status exits 0` | Exit code 0 |
| `workstation-status prints section headers` | Output contains expected section headings (Dotfiles, Dev tools, Homebrew) |
| `workstation-diff exits 0 even when no diff` | Exit code 0 regardless of diff state |

Run locally: `bats test/workstation/`

Not yet in CI -- bats requires local tool installation.

## 9. Legacy Cleanup Completed

The `com.tomtenuta.maintenance` cleanup block was removed from the LaunchAgent script as part of the platform hardening initiative. No `tomtenuta` references remain anywhere under `config/`.
