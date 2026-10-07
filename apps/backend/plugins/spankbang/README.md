# spankbang plugin

Scraper for [spankbang.com](https://spankbang.com). Capabilities: `scene:find` (full scrape of one video URL), `scene:list` (keyword search with `query`, newest-uploads catalog without), `stream:resolve`.

## Cloudflare

The site is behind Cloudflare's managed challenge and answers plain requests with `403`. Every fetch goes through `fetchHtmlSmart`, which retries via **FlareSolverr** (`FLARESOLVERR_URL`). Without a solver configured the plugin cannot fetch.

## URL patterns

- Newest listing: `/new_videos/<page>/`
- Search: `/s/<query>/<page>/` (spaces as `+`)
- Video page: `/<code>/video/<slug>` — `<code>` (e.g. `7gu17`) is the stable id.

## Results scoping

Listing and search pages open with an 8-item **promoted block** (a `[data-testid="video-list"]` that is identical across every page), then the real results follow — under `[data-testid="search-result"]` on search pages, a second `[data-testid="video-list"]` on browse pages. The parser picks the container with the **most** cards, so the promoted block is never mistaken for results.

## Video model

- title: `h1` (falls back to `og:title`)
- poster: `og:image`
- duration: `og:video:duration` (seconds)
- date: first `<time datetime>`
- tags: `[data-testid="video-tags"]` `/s/<tag>/` links, falling back to `meta[name=keywords]`
- performers: none — spankbang has no reliable per-video performer credits; performer names, when tagged, appear among the tags.

## Streams (all qualities)

The video page embeds `var stream_data = { '<quality>': ['<mp4>'], … }` covering every progressive quality (240p–4k) plus an adaptive HLS master (`master.m3u8`). `scene:find` emits **all** of them as downloads (each directly streamable and downloadable); `stream:resolve` returns the single best source (highest mp4, else HLS) for immediate playback. The mp4/HLS URLs are signed but end in `.mp4`/`.m3u8`, so the backend treats them as direct media — no re-resolution needed.
