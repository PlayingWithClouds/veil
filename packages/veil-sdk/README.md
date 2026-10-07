# @playingwithclouds/veil-sdk

Shared TypeScript library for building Veil plugins. Provides all media types, the stdin/stdout JSON protocol handler (`runPlugin`), and HTTP utilities with browser-compatible headers.

---

## Overview

All Veil plugins communicate with the Go backend over a stdin/stdout JSON protocol. This SDK handles that protocol so plugin authors only need to implement their capability logic. It also exports all shared media types and a set of HTTP helpers with browser-compatible headers.

## Exports

| Export                                                     | Description                                                                                                     |
| ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `runPlugin(handlers)`                                      | Reads one JSON invocation from stdin, dispatches to the correct handler, writes the result to stdout, and exits |
| Types (`Movie`, `Series`, `Episode`, `ScrapeResult`, etc.) | All media types mirrored from `internal/plugins/types.go` in the Go backend                                     |
| `fetchHtml(url, headers?)`                                 | `fetch` wrapper with browser UA and Accept headers                                                              |
| `followRedirect(url)`                                      | Follows redirects and returns the final URL                                                                     |
| `resolveUrl(base, path)`                                   | Resolves a relative or protocol-relative path against a base URL                                                |
| `domainOf(url)`                                            | Extracts the hostname without `www.` prefix                                                                     |

## Data client

A fully-typed client over the backend GraphQL API, for both plugins and the frontend. Built on a [genql](https://genql.dev)-generated client (`src/generated`, regenerate with `bun run codegen`) wrapped in an ergonomic per-resource facade.

```ts
import { createVeilClient } from "@playingwithclouds/veil-sdk/client";

const veil = createVeilClient({ url: "http://localhost:8080/graphql" });

const movies = await veil.movie.list({ limit: 50 });      // search/limit/offset
const movie  = await veil.movie.get(id);                  // by id
const made   = await veil.movie.create({ tmdbId: 603, title: "The Matrix", originalTitle: "The Matrix" });
await veil.movie.update(id, { tagline: "..." });
await veil.movie.remove(id);

const seasons = await veil.season.list({ seriesId });     // child lists take their parent id
```

Every read/write returns **all scalar fields** by default. Pass a genql selection as the last argument to pick fields or pull in relations — return types narrow to match:

```ts
const slim = await veil.movie.get(id, { id: true, title: true });   // typed as { id, title } | null
```

Resources: `movie`, `series`, `season`, `episode`, `person`, `collection`, `genre`, `keyword`, `network`, `productionCompany`, `watchProvider`, `credit`, `image`, `video`. Each exposes `get`, `list`, `create`, `update`, `remove`.

For subscriptions, settings, and bespoke mutations not covered by the facade, use the raw genql client:

```ts
const { jobs } = await veil.client.query({ jobs: { id: true, status: true } });
```

Regenerate the typed client whenever the backend schema changes:

```sh
bun run codegen   # concatenates apps/backend SDL -> runs genql -> src/generated
```

## Protocol

Veil invokes a plugin by running `bun run <main>` and writing a single JSON object to stdin:

```json
{ "capability": "scrape", "args": { "url": "https://example.com/movie/123" } }
```

`runPlugin` dispatches to the registered handler and writes the result to stdout:

```json
{ "result": { ... } }
```

On error:

```json
{ "error": "message" }
```

**Scrape** is the exception — it emits one `ScrapeResult` per line (NDJSON), root item first, then children (seasons → episodes):

```
{"type":"series","series":{...}}
{"type":"season","season":{...}}
{"type":"episode","episode":{...}}
```

## Usage

```ts
import { runPlugin } from "@playingwithclouds/veil-sdk";

runPlugin({
  meta: {
    name: "my-plugin",
    version: "0.1.0",
    capabilities: ["scrape"],
  },
  scrape: async (url) => {
    // return ScrapeResult[]
  },
});
```

## Media type hierarchy

```
MediaBase
├── Movie          — year, duration_minutes, trailer_urls
├── Series         — year, status
│   └── Season     — series_external_id, number
│       └── Episode — season_number, number, duration_minutes, air_date
├── Gallery        — year
├── AdultMovie     — acts, series_name, series_number
├── Scene          — number, acts
├── AdultGallery   — acts
└── AdultPerson    — full performer profile
```

## Installation

This package is a local dependency — installed by referencing the path:

```json
{ "@playingwithclouds/veil-sdk": "file:../sdk" }
```

```sh
bun install
```

## License

MIT
