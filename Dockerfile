# ─────────────────────────────────────────────────────────────────────────────
# Neptune-Pamm-Server — multi-stage build
# ─────────────────────────────────────────────────────────────────────────────

# ─── Stage 1: build ──────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Build metadata (passed from `make docker`)
ARG VERSION=dev
ARG COMMIT=none

# Static build, no cgo
ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /src

# Install git (for VCS stamping) and certs
RUN apk add --no-cache git ca-certificates tzdata

# Cache module downloads first (layer reused unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /out/neptune-pamm-server ./cmd/main.go

# ─── Stage 2: runtime ────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Certs + timezone data from the builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# The static binary
COPY --from=builder /out/neptune-pamm-server /app/neptune-pamm-server

# Runs as the distroless `nonroot` user (uid 65532)
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/neptune-pamm-server"]
