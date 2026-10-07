<div align="center">
  <img src="./logo.svg" alt="Veil" width="256" />
  <h1>Veil</h1>
  <p>Self-hosted media library for your local NAS.</p>
</div>

---

Scrapes, enriches, and streams movies, series, and galleries via a plugin-based architecture. Go backend + SvelteKit frontend + TypeScript plugins running in Bun subprocesses.

## Stack

| Layer              | Technology                           |
| ------------------ | ------------------------------------ |
| Backend            | Go (`github.com/playingwithclouds/veil`) |
| API                | GraphQL (`graph-gophers/graphql-go`) |
| Frontend           | SvelteKit (`web/`)                   |
| Database           | SQLite, in process (`modernc.org/sqlite`, pure Go) |
| Blob storage       | Files under `DATA_DIR/blobs`         |
| Plugin runtime     | Bun (TypeScript, stdin/stdout JSON)  |
| Browser automation | Chromium via Puppeteer               |

## Quick start

```sh
docker compose up
```

| Service           | URL                           |
| ----------------- | ----------------------------- |
| Frontend          | <http://localhost:3000>         |
| Backend (GraphQL) | <http://localhost:8080/graphql> |

Everything the backend persists lives in `DATA_DIR` (default `./data`, `/data`
in the container): the SQLite database `veil.db` (migrated on startup from
`internal/db/migrations`) and `blobs/` (cached images, stream cache, plugin
icons, served at `/api/blob/`).

## Architecture

```
Frontend (SvelteKit)
    │  GraphQL + SSE
    ▼
Backend (Go)
    ├── GraphQL API      /graphql
    ├── SSE events       /api/events
    ├── Search           /api/search
    ├── Stream proxy     /api/stream/
    ├── Blobs            /api/blob/
    └── Plugin system
            │  stdin/stdout JSON (Bun subprocess)
            ▼
        Plugin (TypeScript npm package)
```

The backend wires together:

- **Job queue** — persistent queue in SQLite with retry and per-kind rate limits
- **Worker pool** — concurrent job execution with optional VPN gate
- **Pipeline orchestrator** — coordinates discover → scrape → enrich → download flow
- **Stream resolver** — resolves a source URL to a direct playable URL via resolver plugins, with an on-disk HLS/MP4 cache
- **Subscription updater** — polls subscribed series for new episodes every 6 hours
- **Plugin registry** — fsnotify hot-reload; auto-runs `bun install` on first load

## Plugin system

Plugins are npm packages (TypeScript) placed in subdirectories of `./plugins/`. Each directory needs a `package.json` with a `"main"` field.

### Protocol

Every invocation is a single JSON line on stdin, one JSON line response on stdout:

```jsonc
// stdin
{ "capability": "scrape", "args": { "url": "https://..." } }

// stdout — either
{ "result": { ... } }
// or
{ "error": "something went wrong" }
```

For `scrape`, the plugin may emit multiple NDJSON lines (one per media item: series first, then seasons, then episodes).

### Capabilities

| Capability          | Description                                                            |
| ------------------- | ---------------------------------------------------------------------- |
| `meta`              | Required. Returns plugin name, version, and capability list            |
| `discover`          | Bulk-list all available items; returns minimal stubs with `source_url` |
| `scrape`            | Fetch full detail for one URL                                          |
| `enrich`            | Fill missing fields from a secondary source                            |
| `resolve`           | Turn a source URL into a direct playable stream URL                    |
| `search`            | Search by query string                                                 |
| `browse:categories` | List content categories                                                |
| `browse:actors`     | List or search performers                                              |

All capabilities have NSFW equivalents prefixed with `nsfw:` (e.g. `nsfw:discover`).

### Plugin contract

1. Must handle `capability: "meta"` and return `{ name, version, capabilities[] }`.
2. Declare which domains you handle in `meta.domains` (for `resolve`); use `"*"` as wildcard.
3. Declare configurable fields in `meta.settings` — the backend injects values as env vars.
4. TypeScript types in `src/types.ts` must mirror `internal/plugins/types.go`.

### Writing a plugin

```ts
// index.ts (referenced by package.json "main")
import { readInput, writeResult, writeError } from "@playingwithclouds/veil-sdk";

const input = await readInput();

if (input.capability === "meta") {
  writeResult({
    name: "my-plugin",
    version: "0.1.0",
    capabilities: ["discover", "scrape"],
  });
} else if (input.capability === "discover") {
  // fetch items, return DiscoverResult
} else if (input.capability === "scrape") {
  // fetch full item, emit ScrapeResult lines (NDJSON)
} else {
  writeError("unsupported capability");
}
```

Drop the directory into `./plugins/` — the backend hot-reloads it automatically.

## Media types

```
MediaBase (title, description, poster, genres, tags, cast, images, downloads, content_rating)
├── Movie         (year, duration_minutes, trailer_urls)
├── Series        (year, status)  →  Season[]  →  Episode[]
├── Gallery       (year)
├── AdultMovie    (year, duration_minutes, acts)
│   └── Scene     (number, duration_minutes, acts)
├── AdultGallery  (year, acts)
└── AdultPerson   (performer profile)
```

`ContentRating`: `nsfw bool`, `rating` (G/PG/PG-13/R/NC-17/X/NR/FSK…), `warnings[]`.

## Ingest pipeline

```
discover  →  bulk-insert minimal stubs into DB
scrape    →  full series/movie + all seasons/episodes (no stream links yet)
enrich    →  fill missing fields (cast, poster, description) from a secondary source
resolve   →  turn a source URL into a direct playable URL (on demand at play time)
download  →  cache a resolved stream (HLS segments or MP4) to the blob store
```

Steps run as background jobs. `autoScrapeAfterDiscover` and `autoEnrichAfterScrape` settings trigger each stage automatically.

## Configuration (Settings)

Managed via the GraphQL `updateSettings` mutation or the UI. Key options:

| Setting                   | Default | Description                         |
| ------------------------- | ------- | ----------------------------------- |
| `maxConcurrentJobs`       | 4       | Worker pool size                    |
| `maxJobRetries`           | 3       | Retry limit per job                 |
| `autoScrapeAfterDiscover` | false   | Chain scrape after discover         |
| `autoEnrichAfterScrape`   | false   | Chain enrich after scrape           |
| `requireVpn`              | false   | Block workers when VPN disconnected |
| `downloadSpeedLimitKBps`  | 0       | 0 = unlimited                       |
| `tmdbApiKey`              | —       | Used by the TMDB enrichment plugin  |

Per-kind rate limits (`kindLimits`) control `maxConcurrent`, `retryInitialMs`, `retryMultiplier`, and `retryMaxMs` independently per job type.

## API

### GraphQL

`POST /graphql` — full schema at `GET /api/schema`.

Selected operations:

```graphql
# Trigger discovery across all enabled discover-capable plugins
mutation {
  discover(query: "Breaking Bad")
}

# Scrape a specific URL
mutation {
  scrape(pluginName: "s-to", url: "https://s.to/serie/breaking-bad") {
    success
    error
    mediaId
  }
}

# Stream an episode (resolves to direct URL on demand)
query {
  stream(url: "https://...") {
    url
    mimeType
    quality
    headers {
      name
      value
    }
  }
}

# Subscribe to a series for automatic refresh checks
mutation {
  subscribe(mediaId: "series:abc123") {
    id
  }
}
```

### SSE

`GET /api/events` — server-sent events for real-time updates.

Events: `mediaAdded`, `episodeAdded`, `jobUpdated`.

### Search

`GET /api/search?q=<query>` — full-text search across local DB and live plugin search.

## Development

### Prerequisites

- Docker + Docker Compose
- Go 1.25+ (for running outside Docker)
- Bun (for frontend and plugin development)

### Running with live reload

```sh
docker compose up
```

Backend uses [air](https://github.com/air-verse/air) for live reload (config in `.air.toml`). Frontend runs the Vite dev server.

### Building

```sh
make build          # go build + bun run build
```

### Regenerating GraphQL types

After editing `internal/api/graphql/schema.graphql`:

```sh
make generate
# or pull types from a running server:
GQL_SCHEMA_URL=http://localhost:8080/api/schema make generate
```

## Environment variables

| Variable           | Default                 | Description                                        |
| ------------------ | ----------------------- | -------------------------------------------------- |
| `PORT`             | `8080`                  | Backend listen port                                |
| `SURREAL_URL`      | —                       | SurrealDB WebSocket URL                            |
| `SURREAL_USER`     | —                       | SurrealDB username                                 |
| `SURREAL_PASS`     | —                       | SurrealDB password                                 |
| `SURREAL_NS`       | `shutterly`             | SurrealDB namespace                                |
| `SURREAL_DB`       | `media`                 | SurrealDB database                                 |
| `REDIS_URL`        | —                       | Redis connection URL                               |
| `MINIO_ENDPOINT`   | —                       | MinIO host:port                                    |
| `MINIO_ACCESS_KEY` | —                       | MinIO access key                                   |
| `MINIO_SECRET_KEY` | —                       | MinIO secret key                                   |
| `MINIO_BUCKET`     | `media`                 | MinIO bucket name                                  |
| `PLUGIN_DIR`       | `./plugins`             | Plugin directory                                   |
| `PUBLIC_URL`       | `http://localhost:8080` | Public base URL (used for HLS proxy URL rewriting) |

## Bundled plugins

| Plugin       | Capabilities         | Notes                                       |
| ------------ | -------------------- | ------------------------------------------- |
| `s-to`       | `discover`, `scrape` | German series site                          |
| `filmpalast` | varies               | Movie scraper                               |
| `kinoger`    | varies               | Movie scraper                               |
| `kinoz`      | varies               | Movie scraper                               |
| `tmdb`       | `enrich`             | TMDB metadata enrichment (requires API key) |
| `omdb`       | `enrich`             | OMDb metadata enrichment                    |
| `wikidata`   | `enrich`             | Wikidata enrichment                         |

Plugins are git submodules under `./plugins/`.

## Project layout

```
cmd/server/main.go              Entry point — wires all services
internal/
  api/
    graphql/                    GraphQL schema + resolvers
    search/                     Search handler
    sse/                        Server-sent events
  cache/                        Expiring key-value store (SQLite kv table)
  db/                           SQLite open/migrate + query helpers; migrations/
  ingest/                       Scoring + ingestion logic
  jobs/                         Queue, worker pool, job types
  media/                        Media repository (DB queries)
  pipeline/                     Orchestrator (discover/scrape/enrich/download)
  plugins/                      Plugin loader, registry, runner, types
  resolve/                      Stream URL resolver
  settings/                     App settings service
  storage/                      On-disk blob store + /api/blob/ handler
  stream/                       HLS proxy + on-disk segment cache
  subscriptions/                SSE hub
  updater/                      Periodic series refresh scheduler
  vpn/                          VPN connectivity check
plugins/                        Plugin npm packages (git submodules)
web/                            SvelteKit frontend
```
