# veil-plugin-xhamster

Veil plugin for [xhamster.com](https://xhamster.com) — an adult video site. Supports bulk catalog listing, title search, and full metadata scraping.

---

## Overview

xhamster.com is an adult streaming site with a large catalog. This plugin integrates it into Veil's plugin pipeline — crawling the newest listing in bulk, resolving titles via search, and scraping full metadata on demand.

There are no series/episodes on this source — everything is modelled as a single `Scene`.

## Capabilities

| Capability       | Description                                                                                                                                            |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `scene:list`     | Reference stubs — searches `/search/<query>` when a query is set, otherwise crawls the paginated newest listing (`/newest/<page>`)                      |
| `scene:find`     | Fetches one full scene by URL (title, cast, categories, tags, runtime), emitted as NDJSON                                                               |
| `stream:resolve` | Extracts a plain HLS/mp4 stream from a video page when one is exposed                                                                                   |

## URL patterns

- Listing page: `/newest/<page>` — video cards `a.video-thumb__image-container`
- Video card: `a.video-thumb__image-container[href="/videos/<slug>-<id>"]`, title in `aria-label`, poster in child `img[src]`, preview clip in `data-previewvideo`, runtime badge in `[data-role="video-duration"]` (`MM:SS` / `H:MM:SS`)
- Search: `/search/<query>` and `/search/<query>?page=<page>`
- Video page: `/videos/<slug>-<id>`
  - Title: `meta[property="og:title"]`
  - Poster: `meta[property="og:image"]`
  - Runtime: `xplayerSettings.duration` (seconds) in the page's inline player config
  - Cast: `a[href*="/pornstars/"]` (index links under `/pornstars/all/` skipped)
  - Categories: `a[href*="/categories/"]`
  - Tags: `a[href*="/tags/"]`

## Data mapping

- `external_id` — `xhamster-<idToken>` (trailing hyphen segment of the video slug)
- `tags` — category chips plus tag chips
- `performers` — pornstar links on the video page
- `downloads` — the video page URL, handed to the resolver

## Streams

xhamster encrypts the per-quality `sources` URLs on the video page for in-browser
decryption. The resolver extracts a plain HLS master or mp4 when one is exposed
and throws otherwise — encrypted-only videos are not directly resolvable without
the site's player.

## Development

```sh
bun install
bun test
```

Live tests hit xhamster.com and need network access.
