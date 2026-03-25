# Multi-stage build for Clew server (ari serve).
# Produces a minimal distroless image with the statically-linked binary.

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

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/ari", "serve"]
