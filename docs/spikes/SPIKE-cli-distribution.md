# SPIKE: CLI Distribution Patterns for Ariadne

> Research spike for distributing the `ari` CLI binary across multiple platforms and package managers

**Time-boxed**: Research spike (findings, not production code)
**Date**: 2026-01-07
**Author**: Technology Scout (rnd)

---

## Executive Summary

This spike researches best practices for distributing the Ariadne CLI (`ari`) binary to end users. After analyzing goreleaser patterns, Homebrew tap setup, and alternative distribution methods, we recommend **goreleaser** as the release automation tool with **Homebrew tap** as the primary macOS/Linux distribution channel and **Scoop bucket** for Windows. GitHub Releases serves as the canonical source for direct binary downloads.

**Verdict**: Proceed with goreleaser + Homebrew tap setup. The tooling is mature, well-documented, and aligns with industry standards (gh CLI, terraform, etc.).

---

## 1. Context

### Current State

- **CLI binary**: `ari`
- **Go module**: `github.com/autom8y/knossos`
- **Target repository**: `github.com/autom8y/knossos` (post-rename)
- **Build system**: None formalized (manual `go build`)

### Success Criteria

- Published release with GitHub Releases artifacts
- Homebrew tap for macOS/Linux installation
- Clear installation documentation
- Automated release workflow via GitHub Actions

---

## 2. GoReleaser Configuration

### Overview

GoReleaser automates the entire release process: building binaries for multiple platforms, creating archives, generating changelogs, and publishing to package managers.

### Core Configuration

```yaml
# .goreleaser.yaml
version: 2

project_name: ari

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: ari
    main: ./cmd/ari
    binary: ari
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: ari-archive
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'
      - Merge pull request
      - Merge branch

release:
  github:
    owner: autom8y
    name: knossos
  draft: false
  prerelease: auto
  name_template: "v{{.Version}}"
```

### Build Matrix

| OS | Architecture | Format | Notes |
|----|--------------|--------|-------|
| linux | amd64, arm64 | tar.gz | Standard Linux distributions |
| darwin | amd64, arm64 | tar.gz | Intel and Apple Silicon Macs |
| windows | amd64, arm64 | zip | Windows 10/11 |

### Version Injection Pattern

The `main.go` already supports version injection via ldflags:

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

GoReleaser automatically populates these from git tags and build metadata.

### Testing Before Release

```bash
# Validate configuration
goreleaser check

# Build without releasing (snapshot)
goreleaser release --snapshot --clean

# Skip publishing for inspection
goreleaser release --skip=publish
```

### Sources

- [GoReleaser Quick Start](https://goreleaser.com/quick-start/)
- [GoReleaser Builds Configuration](https://goreleaser.com/customization/builds/go/)
- [GoReleaser Changelog](https://goreleaser.com/customization/changelog/)

---

## 3. Homebrew Tap Setup

### Repository Structure

A Homebrew tap is a Git repository following a specific naming convention:

```
homebrew-tap/                # github.com/autom8y/homebrew-tap
├── Formula/
│   └── ari.rb              # Formula for ari CLI
├── Casks/                  # Optional: for GUI apps
└── README.md
```

**Naming Requirement**: Repository must be named `homebrew-<something>` (e.g., `homebrew-tap`).

### Formula Generation via GoReleaser

Add to `.goreleaser.yaml`:

```yaml
brews:
  - name: ari
    ids:
      - ari-archive
    repository:
      owner: autom8y
      name: homebrew-tap
      branch: main
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/autom8y/knossos"
    description: "Ariadne CLI - The thread through the Knossos labyrinth"
    license: "MIT"
    install: |
      bin.install "ari"
    test: |
      system "#{bin}/ari", "--version"
    dependencies:
      - name: git
        type: optional
```

### Manual Formula Example

If not using GoReleaser automation:

```ruby
# Formula/ari.rb
class Ari < Formula
  desc "Ariadne CLI - The thread through the Knossos labyrinth"
  homepage "https://github.com/autom8y/knossos"
  version "1.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/autom8y/knossos/releases/download/v1.0.0/ari_1.0.0_darwin_arm64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
    on_intel do
      url "https://github.com/autom8y/knossos/releases/download/v1.0.0/ari_1.0.0_darwin_amd64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/autom8y/knossos/releases/download/v1.0.0/ari_1.0.0_linux_arm64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
    on_intel do
      url "https://github.com/autom8y/knossos/releases/download/v1.0.0/ari_1.0.0_linux_amd64.tar.gz"
      sha256 "CHECKSUM_HERE"
    end
  end

  def install
    bin.install "ari"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/ari --version")
  end
end
```

### Installation Experience

```bash
# Add tap
brew tap autom8y/tap

# Install
brew install ari

# Update
brew upgrade ari
```

### Cross-Repository Authentication

GoReleaser's default `GITHUB_TOKEN` cannot push to another repository. Options:

1. **Personal Access Token (PAT)**: Create a PAT with `repo` scope, store as `HOMEBREW_TAP_TOKEN` secret
2. **GitHub App**: More secure for organizations, generates installation tokens
3. **Deploy Key**: SSH key with write access to tap repository

**Recommendation**: Use a PAT for simplicity; upgrade to GitHub App if security requirements increase.

### Sources

- [Homebrew Taps Documentation](https://docs.brew.sh/Taps)
- [GoReleaser Homebrew Configuration](https://goreleaser.com/customization/homebrew/)
- [How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)

---

## 4. Other Distribution Methods

### 4.1 Scoop (Windows)

Scoop is the Windows equivalent of Homebrew. GoReleaser can generate Scoop manifests:

```yaml
# Add to .goreleaser.yaml
scoops:
  - name: ari
    ids:
      - ari-archive
    repository:
      owner: autom8y
      name: scoop-bucket
      branch: main
      token: "{{ .Env.SCOOP_BUCKET_TOKEN }}"
    homepage: "https://github.com/autom8y/knossos"
    description: "Ariadne CLI - The thread through the Knossos labyrinth"
    license: MIT
```

**Repository Structure**:
```
scoop-bucket/               # github.com/autom8y/scoop-bucket
├── ari.json               # Generated manifest
└── README.md
```

**Installation**:
```powershell
scoop bucket add autom8y https://github.com/autom8y/scoop-bucket
scoop install ari
```

### 4.2 Linux Packages (deb/rpm/apk)

GoReleaser integrates nFPM for native Linux packages:

```yaml
# Add to .goreleaser.yaml
nfpms:
  - id: ari-packages
    package_name: ari
    vendor: Autom8y
    homepage: https://github.com/autom8y/knossos
    maintainer: Autom8y <support@autom8y.com>
    description: "Ariadne CLI - The thread through the Knossos labyrinth"
    license: MIT
    formats:
      - deb
      - rpm
      - apk
    bindir: /usr/bin
    contents:
      - src: ./LICENSE
        dst: /usr/share/doc/ari/LICENSE
```

**Distribution Options**:
- Upload to GitHub Releases (manual download)
- Host APT/YUM repository (more infrastructure)
- Submit to official repositories (long process)

**Recommendation**: Start with GitHub Releases artifacts. Consider APT/YUM repos if Linux adoption justifies the infrastructure cost.

### 4.3 Nix (NUR)

GoReleaser can generate Nix derivations for the Nix User Repository:

```yaml
# Add to .goreleaser.yaml
nix:
  - name: ari
    repository:
      owner: autom8y
      name: nur-packages
      branch: main
    homepage: "https://github.com/autom8y/knossos"
    description: "Ariadne CLI - The thread through the Knossos labyrinth"
    license: mit
    install: |
      mkdir -p $out/bin
      cp ari $out/bin/ari
```

**Note**: Nix support is less mature. Consider only if Nix users specifically request it.

### 4.4 go install

Users can install directly from source:

```bash
go install github.com/autom8y/knossos/cmd/ari@latest
```

**Caveats**:
- Requires Go toolchain installed
- Version may not match releases (depends on user's Go version)
- ldflags version injection not applied (shows "dev")
- Not recommended as primary installation method

### Distribution Method Comparison

| Method | Platform | Effort | User Experience | Recommendation |
|--------|----------|--------|-----------------|----------------|
| **GitHub Releases** | All | Low | Manual download | Always (baseline) |
| **Homebrew Tap** | macOS/Linux | Low | Excellent | Yes (primary) |
| **Scoop Bucket** | Windows | Low | Excellent | Yes (Windows users) |
| **deb/rpm** | Linux | Medium | Good | Phase 2 |
| **Nix NUR** | NixOS | Medium | Good | On request |
| **go install** | All | None | Fair | Documented, not promoted |

### Sources

- [GoReleaser Scoop Configuration](https://goreleaser.com/customization/scoop/)
- [nFPM Documentation](https://nfpm.goreleaser.com/)
- [GoReleaser Nix Configuration](https://goreleaser.com/customization/nix/)

---

## 5. Installation UX Patterns

### Pattern Comparison

| Pattern | Security | Simplicity | Recommendation |
|---------|----------|------------|----------------|
| **Package Manager** | High (signed packages) | High | Primary method |
| **curl \| sh** | Medium (HTTPS only) | High | Acceptable with HTTPS |
| **Binary Download** | Medium (checksums) | Medium | Always available |
| **go install** | High (source-based) | Medium | Developer-focused |

### Recommended Installation Script

For users without Homebrew, provide a simple install script:

```bash
#!/bin/sh
# install.sh - Install ari CLI
set -e

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
VERSION=$(curl -s https://api.github.com/repos/autom8y/knossos/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')

# Download and install
TARBALL="ari_${VERSION}_${OS}_${ARCH}.tar.gz"
curl -sLO "https://github.com/autom8y/knossos/releases/download/v${VERSION}/${TARBALL}"
tar xzf "$TARBALL"
sudo mv ari /usr/local/bin/
rm "$TARBALL"

echo "ari v${VERSION} installed to /usr/local/bin/ari"
```

**Usage**:
```bash
curl -sSL https://raw.githubusercontent.com/autom8y/knossos/main/install.sh | sh
```

### Installation Documentation Structure

```markdown
## Installation

### macOS / Linux (Homebrew)

```bash
brew tap autom8y/tap
brew install ari
```

### Windows (Scoop)

```powershell
scoop bucket add autom8y https://github.com/autom8y/scoop-bucket
scoop install ari
```

### Manual Download

Download the appropriate binary from [GitHub Releases](https://github.com/autom8y/knossos/releases).

### From Source

```bash
go install github.com/autom8y/knossos/cmd/ari@latest
```
```

### Sources

- [golangci-lint Installation Patterns](https://golangci-lint.run/docs/welcome/install/local/)
- [CLI UX Patterns](http://lucasfcosta.com/2022/06/01/ux-patterns-cli-tools.html)
- [Curl to Shell Security Analysis](https://www.arp242.net/curl-to-sh.html)

---

## 6. GitHub Actions Workflow

### Release Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for changelog generation

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
          SCOOP_BUCKET_TOKEN: ${{ secrets.SCOOP_BUCKET_TOKEN }}
```

### Required Secrets

| Secret | Purpose | Scope |
|--------|---------|-------|
| `GITHUB_TOKEN` | GitHub Releases (automatic) | Current repo |
| `HOMEBREW_TAP_TOKEN` | Push to homebrew-tap repo | `repo` scope PAT |
| `SCOOP_BUCKET_TOKEN` | Push to scoop-bucket repo | `repo` scope PAT |

### Release Process

1. **Tag the release**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions triggers** on tag push

3. **GoReleaser executes**:
   - Builds binaries for all platforms
   - Creates archives with checksums
   - Generates changelog from commits
   - Creates GitHub Release with assets
   - Updates Homebrew tap formula
   - Updates Scoop bucket manifest

4. **Users can install** via their preferred method

### Pre-release Testing Workflow

```yaml
# .github/workflows/build.yml
name: Build

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build Snapshot
        uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: build --snapshot --clean
```

### Sources

- [GoReleaser GitHub Actions](https://goreleaser.com/ci/actions/)
- [goreleaser-action Repository](https://github.com/goreleaser/goreleaser-action)

---

## 7. Version Management and Changelog

### Semantic Versioning

GoReleaser enforces semantic versioning. Tags must follow the pattern:

```
v1.0.0        # Standard release
v1.0.0-beta.1 # Pre-release
v1.0.0-rc.1   # Release candidate
```

### Changelog Generation Options

| Method | Pros | Cons |
|--------|------|------|
| **git** (default) | Simple, no dependencies | Basic formatting |
| **github** | Includes author usernames | Requires API access |
| **github-native** | Uses GitHub's native generation | Less customization |
| **semantic-release** | Conventional commits | Additional tooling |

**Recommendation**: Start with `git` method. Consider `github-native` for richer changelogs.

### Conventional Commits Integration

If using Conventional Commits, GoReleaser can categorize changes:

```yaml
changelog:
  sort: asc
  use: github
  groups:
    - title: Features
      regexp: '^feat(\(.+\))?: .+'
      order: 0
    - title: Bug Fixes
      regexp: '^fix(\(.+\))?: .+'
      order: 1
    - title: Documentation
      regexp: '^docs(\(.+\))?: .+'
      order: 2
    - title: Other
      order: 999
  filters:
    exclude:
      - '^chore:'
      - Merge pull request
```

### Sources

- [GoReleaser Semantic Versioning](https://goreleaser.com/limitations/semver/)
- [GoReleaser Changelog](https://goreleaser.com/customization/changelog/)
- [Semantic Release Cookbook](https://goreleaser.com/cookbooks/semantic-release/)

---

## 8. Cross-Platform Build Considerations

### CGO Considerations

The recommended configuration disables CGO:

```yaml
builds:
  - env:
      - CGO_ENABLED=0
```

**Benefits**:
- Static binaries with no system dependencies
- Cross-compilation works without toolchains
- Simpler CI/CD pipeline

**Trade-offs**:
- Cannot use CGO-dependent libraries
- Some crypto operations may be slower (Go's pure Go implementations)

Current `ari` dependencies do not require CGO.

### Platform-Specific Considerations

| Platform | Consideration | Mitigation |
|----------|---------------|------------|
| **darwin/arm64** | Apple Silicon | Native support via cross-compilation |
| **windows** | Path separators | Use `filepath.Join` (already done) |
| **linux/arm64** | ARM servers (AWS Graviton) | Include in build matrix |

### Testing Cross-Platform Builds

```bash
# Test all platforms locally
goreleaser build --snapshot --clean

# Verify artifacts
ls dist/
```

---

## 9. Recommendations Summary

### Phase 1: Initial Release (Recommended)

1. **Configure GoReleaser** with builds for darwin/linux/windows on amd64/arm64
2. **Create `homebrew-tap` repository** at `github.com/autom8y/homebrew-tap`
3. **Set up GitHub Actions** release workflow
4. **Generate PAT** for cross-repo publishing
5. **Tag v1.0.0** and verify release

### Phase 2: Expand Distribution

1. **Add Scoop bucket** for Windows users
2. **Generate deb/rpm packages** via nFPM
3. **Create install.sh script** for curl-pipe-sh installation
4. **Document all installation methods**

### Phase 3: Advanced (On Demand)

1. **Nix NUR** if requested
2. **APT/YUM repository** if Linux adoption warrants
3. **Code signing** for enhanced security

### Repository Structure (Post-Implementation)

```
autom8y/
├── knossos/                  # Main repository
│   ├── ariadne/              # CLI source
│   │   ├── cmd/ari/
│   │   └── go.mod
│   ├── .goreleaser.yaml      # Release configuration
│   ├── .github/workflows/
│   │   └── release.yml       # Release automation
│   └── install.sh            # Manual install script
├── homebrew-tap/             # Homebrew formulas
│   └── Formula/ari.rb
└── scoop-bucket/             # Scoop manifests
    └── ari.json
```

---

## 10. Risk Assessment

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Cross-repo token exposure | Low | High | Use GitHub Secrets, minimal scope PAT |
| Broken formula after release | Medium | Medium | Test with snapshot builds first |
| Windows path issues | Low | Medium | Use filepath.Join consistently |
| Architecture mismatch | Low | Low | Explicit build matrix |

### Operational Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Forgotten tap repository | Medium | Low | Document in release checklist |
| Version mismatch between repos | Low | Medium | GoReleaser handles atomically |
| Abandoned alternative channels | Medium | Low | Start with only Homebrew/Scoop |

---

## Appendix: Complete GoReleaser Configuration

```yaml
# .goreleaser.yaml
version: 2

project_name: ari

before:
  hooks:
    - go mod tidy

builds:
  - id: ari
    dir: ariadne
    main: ./cmd/ari
    binary: ari
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: ari-archive
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'
      - Merge pull request
      - Merge branch

release:
  github:
    owner: autom8y
    name: knossos
  draft: false
  prerelease: auto
  name_template: "v{{.Version}}"

brews:
  - name: ari
    ids:
      - ari-archive
    repository:
      owner: autom8y
      name: homebrew-tap
      branch: main
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/autom8y/knossos"
    description: "Ariadne CLI - The thread through the Knossos labyrinth"
    license: "MIT"
    install: |
      bin.install "ari"
    test: |
      system "#{bin}/ari", "--version"

scoops:
  - name: ari
    ids:
      - ari-archive
    repository:
      owner: autom8y
      name: scoop-bucket
      branch: main
      token: "{{ .Env.SCOOP_BUCKET_TOKEN }}"
    homepage: "https://github.com/autom8y/knossos"
    description: "Ariadne CLI - The thread through the Knossos labyrinth"
    license: MIT

nfpms:
  - id: ari-packages
    package_name: ari
    vendor: Autom8y
    homepage: https://github.com/autom8y/knossos
    maintainer: Autom8y <support@autom8y.com>
    description: "Ariadne CLI - The thread through the Knossos labyrinth"
    license: MIT
    formats:
      - deb
      - rpm
    bindir: /usr/bin
```

---

## References

### Official Documentation

- [GoReleaser Documentation](https://goreleaser.com/)
- [Homebrew Documentation](https://docs.brew.sh/)
- [Scoop Documentation](https://scoop.sh/)
- [nFPM Documentation](https://nfpm.goreleaser.com/)

### Reference Implementations

- [GoReleaser's Own Config](https://github.com/goreleaser/goreleaser/blob/main/.goreleaser.yaml)
- [GitHub CLI Release Config](https://github.com/cli/cli/blob/trunk/.goreleaser.yml)
- [golangci-lint Installation](https://golangci-lint.run/docs/welcome/install/local/)

### Repositories to Create

- `github.com/autom8y/homebrew-tap` - Homebrew formulas
- `github.com/autom8y/scoop-bucket` - Scoop manifests (optional, Phase 2)
