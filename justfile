# Ariadne build automation
# Use: just <recipe>
#
# NOTE: CGO_ENABLED=0 is required for all build/test commands to avoid
# macOS dyld issues (missing LC_UUID load command) that cause test binaries
# to abort. This is a known Go/macOS compatibility issue with CGO-linked
# test binaries. Linux CI runs without this limitation.
#
# See: .github/workflows/ariadne-tests.yml for CI configuration

# Default recipe - build the binary
default: build

# Build ari binary (CGO disabled for macOS compatibility)
build:
    CGO_ENABLED=0 go build -ldflags "-X github.com/autom8y/knossos/internal/cmd/root.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev) -X github.com/autom8y/knossos/internal/cmd/root.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none) -X github.com/autom8y/knossos/internal/cmd/root.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o ari ./cmd/ari/main.go

# Build with verbose output
build-verbose:
    CGO_ENABLED=0 go build -v -ldflags "-X github.com/autom8y/knossos/internal/cmd/root.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev) -X github.com/autom8y/knossos/internal/cmd/root.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none) -X github.com/autom8y/knossos/internal/cmd/root.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o ari ./cmd/ari/main.go

# Run all tests
test:
    CGO_ENABLED=0 go test ./...

# Run tests with verbose output
test-verbose:
    CGO_ENABLED=0 go test -v ./...

# Run specific package tests
test-sails:
    CGO_ENABLED=0 go test -v ./internal/sails/...

# Validate frontmatter on all INDEX.md command files
audit-frontmatter:
    @echo "Auditing command frontmatter..."
    @CGO_ENABLED=0 go test ./internal/materialize/... -run TestFrontmatterAllIndexFiles -v
    @echo "✓ Frontmatter audit passed."

# Lint the codebase
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -f ari
    go clean -testcache

# Install to GOPATH/bin
install:
    CGO_ENABLED=0 go install ./cmd/ari

# Show binary info
info:
    @file ari 2>/dev/null || echo "Binary not built yet"
    @ls -lh ari 2>/dev/null || true
