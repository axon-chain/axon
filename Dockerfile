# syntax=docker/dockerfile:1.7

# ── Stage 1: Build ──────────────────────────────────────────
FROM golang:1.25.7-trixie AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    git make gcc libc-dev && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 make build

# ── Stage 2: Runtime ────────────────────────────────────────
FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl jq python3 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/build/axond /usr/local/bin/axond

EXPOSE 17656 17657 11317 19090 18545 18546

VOLUME /root/.axond

ENTRYPOINT ["axond"]
