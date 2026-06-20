# syntax=docker/dockerfile:1

# ----- Builder -------------------------------------------------------------
# Pinned to the same Go version as go.mod / CI (.github/workflows/ci.yml).
FROM golang:1.26 AS builder

WORKDIR /src

# Download modules first so the layer is cached when only source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Version metadata injected via ldflags (see cmd/version.go). Passed from
# fly.toml [build.args] or the CD workflow; defaults keep local builds working.
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# CGO disabled → a fully static binary that runs on distroless/static.
# The package main lives in ./cmd, so build that path with an explicit name.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}" \
    -o /api-gateway ./cmd

# ----- Runtime -------------------------------------------------------------
# distroless/static ships CA certificates (needed for HTTPS JWKS / upstreams)
# and runs as an unprivileged user. Nothing else is in the image.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /api-gateway /api-gateway
# Third entry in the config search order (internal/config/config.go).
COPY gateway.yaml /etc/gateway/gateway.yaml

EXPOSE 8080

ENTRYPOINT ["/api-gateway", "serve"]
