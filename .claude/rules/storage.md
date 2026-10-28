---
paths:
  - "apps/backend/internal/storage/**"
  - "apps/backend/internal/blobcache/**"
  - "apps/backend/internal/stream/**"
  - "apps/backend/internal/resolve/**"
  - "apps/backend/cmd/server/**"
---

# Storage, cache, streaming

## Rules

- Posters, `images/` and plugin icons are never evicted.
- `recordings/` and `live-buffer/` sit outside the cache prefixes; the evictor never touches them.
- Cache files may be deleted by hand, so lookups must tolerate missing blobs (check the blob, and every HLS segment, else go to origin).

## Layout

`DATA_DIR` (default `apps/backend/data`): `veil.db`, `blobs/`, `plugins/`. Blob store served at `/api/blob/`.

Env: `DNS_SERVERS` (comma-separated) or `DNS_SERVERS_FILE` (re-read per lookup) where Go can't find the system resolver (Android). `HOST` bind address (default all). `PLUGIN_INDEX` plugin index URL. `PUBLIC_URL` base of backend-built URLs.

## Streaming

`internal/stream`: HLS/MP4 proxy with write-through blob cache; live cam sessions bypass it. `internal/resolve`: cache → domain plugin → wildcard → direct.

## Cache eviction (`internal/blobcache`)

`stream-cache/` + `img-cache/` capped at `CACHE_MAX_BYTES` (default 300 MB, `blobcache.DefaultMaxBytes`). Over cap: LRU evicted to 90%, checked at start, every 15 min, shortly after writes. Last use = mtime, bumped when served. An HLS stream (manifests + segments) evicts as one, with its kv entries.

- Kept, not counted: downloads (blob URLs in download jobs); streams in playback or being cached (any file used in the last 5 min, or during a cache operation) until idle.
- Kept, counted: other blobs used in the last 5 min.
- At start: kv entries for missing blobs dropped, downloads with missing files marked failed, half-written `.put-*` removed.

## Encryption at rest (`internal/storage/encrypt.go`)

With `BLOB_ENCRYPTION_KEY` (base64, 32 bytes), `stream-cache/` (streams, downloads), `recordings/` and `live-buffer/` are encrypted with chunked AES-256-GCM, seekable so range requests work. `BLOB_ENCRYPTION_WRITES=off` stops new encryption but still reads encrypted blobs. Plain blobs stay readable.
