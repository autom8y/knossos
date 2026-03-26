# Multi-stage build for Clew server (ari serve).
# Produces a minimal distroless image with the statically-linked binary.
#
# Registry data strategy: The domains.yaml catalog is pre-baked into the image
# from deploy/registry/domains.yaml. This is a build-time snapshot of the org's
# knowledge registry (128 domains across 8 repos). To refresh, run
# `ari registry sync --org autom8y` locally, copy the updated catalog to
# deploy/registry/domains.yaml, and rebuild. Future iterations will add
# webhook-driven incremental updates via HandlePushEvent.
#
# Content strategy: .know/ files are pre-baked from deploy/content/{repo}/.know/.
# This provides the BM25 index with actual domain content for search and synthesis.
# To refresh, run `deploy/scripts/collect-content.sh` which clones repos and
# copies their .know/ directories into deploy/content/. The ContentStore reads
# from /data/content/ at runtime.
#
# XDG path layout (distroless nonroot, HOME=/home/nonroot):
#   /home/nonroot/.local/share/knossos/registry/autom8y/domains.yaml
#   /home/nonroot/.local/share/knossos/orgs/autom8y/org.yaml
#   /data/content/{repo}/.know/*.md

# ---------- Builder ----------
FROM golang:1.25 AS builder

WORKDIR /build

# Cache module download.
COPY go.mod go.sum ./
RUN go mod download

# Copy source.
COPY . .

# Build static binary with stripped symbols.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /ari ./cmd/ari

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /ari /ari

# Pre-bake registry data into the image at the XDG data path.
# distroless nonroot HOME=/home/nonroot; Linux XDG fallback: ~/.local/share/
# The org.yaml is minimal (name only) — sufficient for DefaultOrgContext() resolution.
# The domains.yaml is the full catalog snapshot from deploy/registry/.
COPY --chown=nonroot:nonroot \
    deploy/registry/domains.yaml \
    /home/nonroot/.local/share/knossos/registry/autom8y/domains.yaml
COPY --chown=nonroot:nonroot \
    deploy/registry/org.yaml \
    /home/nonroot/.local/share/knossos/orgs/autom8y/org.yaml

# Pre-bake .know/ content for BM25 indexing and context assembly (C-1 fix).
# The ContentStore reads from /data/content/{repo}/.know/ at runtime.
# Directory structure: deploy/content/{repoName}/.know/*.md
COPY --chown=nonroot:nonroot \
    deploy/content/ \
    /data/content/

EXPOSE 8080

# KNOSSOS_ORG tells the serve command which org context to use for catalog loading.
# This env var is read by config.ActiveOrg() -> config.DefaultOrgContext().
ENV KNOSSOS_ORG=autom8y

USER nonroot:nonroot

ENTRYPOINT ["/ari", "serve"]
