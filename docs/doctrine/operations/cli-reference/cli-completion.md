---
last_verified: 2026-02-26
---

# CLI Reference: completion

> Shell autocompletion scripts.

Generate autocompletion scripts for various shells.

**Family**: completion
**Commands**: 4
**Priority**: LOW

---

## Commands

### ari completion bash

Generate bash autocompletion script.

**Synopsis**:
```bash
ari completion bash
```

**Installation**:
```bash
# Add to ~/.bashrc
source <(ari completion bash)

# Or save to file
ari completion bash > /usr/local/etc/bash_completion.d/ari
```

---

### ari completion zsh

Generate zsh autocompletion script.

**Synopsis**:
```bash
ari completion zsh
```

**Installation**:
```bash
# Add to ~/.zshrc
source <(ari completion zsh)

# Or save to fpath
ari completion zsh > "${fpath[1]}/_ari"
```

---

### ari completion fish

Generate fish autocompletion script.

**Synopsis**:
```bash
ari completion fish
```

**Installation**:
```bash
ari completion fish > ~/.config/fish/completions/ari.fish
```

---

### ari completion powershell

Generate PowerShell autocompletion script.

**Synopsis**:
```bash
ari completion powershell
```

**Installation**:
```powershell
ari completion powershell | Out-String | Invoke-Expression
```

---

## See Also

- [CLI Reference Index](index.md) — All 32 command families
