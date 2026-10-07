# ---- dev stage (air live reload) ----
# Use the Debian-based image — bun is a glibc binary and won't run on musl/Alpine.
FROM golang:1.25 AS dev
RUN apt-get update && apt-get install -y \
    curl unzip \
    && rm -rf /var/lib/apt/lists/*
# Install bun (needed to run plugin subprocesses at runtime)
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH="/root/.bun/bin:$PATH"
RUN go install github.com/air-verse/air@latest
WORKDIR /app
# Pre-download modules so the first air build isn't cold.
# The actual source is bind-mounted at runtime; this just warms the cache.
# Register the SDK globally so plugins can resolve link:@playingwithclouds/veil-sdk via bun install.
# /app/veil-sdk is bind-mounted from ../veil-sdk at runtime.
CMD ["sh", "-c", "go mod download && (cd /app/veil-sdk && bun link) && exec air"]

# ---- prod builder ----
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /shutterly ./cmd/server

# ---- prod runtime ----
# Plugins run in the backend's embedded JS runtime, so no Bun is needed here.
FROM alpine:3
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /shutterly .
RUN mkdir -p /data/plugins
# SQLite database, blobs (cached images, stream cache) and installed plugin
# bundles (/data/plugins, from scripts/bundle-plugins.ts).
ENV DATA_DIR=/data
VOLUME /data
EXPOSE 8080
CMD ["/app/shutterly"]
