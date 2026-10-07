# veil-plugin-missav

Veil plugin for [missav.ws](https://missav.ws) — an adult JAV (NSFW) streaming site. Supports catalog discovery, full metadata scraping, and HLS stream resolution.

---

## Overview

missav.ws is an adult JAV streaming site. This plugin integrates it into Veil's plugin pipeline — crawling the newest listing in bulk, scraping full metadata on demand, and resolving the playable HLS master from the page's obfuscated player.

Every item is a single adult JAV title, modelled as one `Scene`.

## Capabilities

| Capability       | Description                                                                                                          |
| ---------------- | -------------------------------------------------------------------------------------------------------------------- |
| `scene:list`     | With a query, searches `/en/search/<query>` via the Cloudflare solver; without one, crawls the newest listing (`/en/new?page=<n>`). Returns minimal stubs |
| `scene:find`     | Fetches full metadata (title, cast, genres, code, release date) for one video URL                                     |
| `stream:resolve` | Unpacks the packed player script to the surrit.com HLS master playlist                                                |

### Cloudflare note

Search and filter pages (`/en/search/<query>`, `/en/genres/<genre>`,
`/en/actresses/<name>`) sit behind a Cloudflare JS challenge that a plain
server-side fetch cannot clear, so `scene:list` with a query goes through
`fetchHtmlSmart` (FlareSolverr when configured, plain fetch otherwise). The
newest listing and video detail pages are not challenged, so queryless
`scene:list` and `scene:find` use plain fetches.

## URL patterns

- Listing page: `/en/new?page=<n>` — video cards `div.thumbnail.group`
- Video card: `a.text-secondary[href="/en/<code>"]`, poster `img[data-src]` (fourhoi.com cover), hover clip `video[data-src]`, runtime overlay span with clock text (`H:MM:SS`)
- Video page: `/en/<code>`
  - Title: `meta[property="og:title"]` (English display title)
  - Poster: `meta[property="og:image"]`
  - Info panel: `div.text-secondary` rows keyed by a leading `<span>Label:</span>`
    - `Code:`, `Title:` (original), `Release date:`, `Actress:`, `Genre:`
  - Cast: `a[href*="/en/actresses/"]` (excluding `/actresses/ranking`)
  - Genres: `a[href*="/en/genres/"]`
  - Stream: a dean-edwards packed script encoding `https://surrit.com/<uuid>/playlist.m3u8`

## Data mapping

- `external_id` — `missav-<code>` (trailing path segment, e.g. `missav-ebon-006`)
- `tags` — genre chips plus the DVD-style code (e.g. `EBON-006`)
- `performers` — actresses in page order
- `duration_seconds` — parsed from the listing card's runtime overlay (list stubs)
- `downloads` — the video page URL, handed to the resolver

## Streams

The video page embeds the stream URL inside a dean-edwards packed script. The
resolver unpacks it and returns the surrit.com HLS master (`playlist.m3u8`). That
master is served from a Cloudflare-fronted CDN and expects a missav `Referer`
at playback, which is included in the resolve result headers.

## Development

```sh
bun install
bun test
```

Live tests hit missav.ws and need network access.
