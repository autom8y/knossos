# Makefile -- E2E distribution validation targets for ari CLI
#
# Complements justfile (which handles build/test/lint).
# This Makefile adds e2e targets that require Docker.
#
# Usage:
#   make e2e-linux   -- build Docker image and run Linux E2E harness locally
#   make e2e-local   -- run e2e-validate.sh directly on host (macOS developers)

.PHONY: e2e-linux e2e-local

## e2e-linux: Build Docker image and run E2E validation with Linuxbrew.
##            Requires Docker. Validates linux_amd64/arm64 distribution path.
##            VERSION is auto-detected from latest GitHub release if gh CLI is available.
e2e-linux:
	docker build -f Dockerfile.e2e -t ari-e2e .
	@echo ""
	@echo "Running E2E validation inside Docker (Linuxbrew)..."
	@echo ""
	docker run --rm ari-e2e $(if $(VERSION),--version $(VERSION),)

## e2e-local: Run e2e-validate.sh directly on the host.
##            Intended for macOS developers with Homebrew installed.
##            Skips Docker; validates real macOS Homebrew path.
##            Set VERSION=v0.x.y to pin a specific version.
e2e-local:
	@echo "Running E2E validation on host..."
	@echo ""
	./scripts/e2e-validate.sh $(if $(VERSION),--version $(VERSION),)

## help: Show available targets.
help:
	@grep -E '^## ' Makefile | sed 's/## //'
