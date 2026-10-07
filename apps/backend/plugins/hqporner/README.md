# veil-plugin-hqporner

Veil plugin for [hqporner.com](https://hqporner.com) — an adult (NSFW) video site. Supports bulk catalog discovery, title search, and full metadata scraping including the embedded player source.

---

## Overview

hqporner.com is an adult streaming site with a large HD catalog. This plugin integrates it into Veil's plugin pipeline — crawling the catalog in bulk, resolving titles via search, and scraping full metadata on demand.

Every item is adult content, so each scraped `Movie` observation carries `adult: true`. There are no series/episodes on this source — everything is modelled as a single movie.

## Capabilities

| Capability | Description                                                                          |
| ---------- | ------------------------------------------------------------------------------------ |
| `discover` | Crawls the paginated HD listing (`/hdporn/<page>`), returning minimal video stubs    |
| `search`   | Queries hqporner's search endpoint (`/?q=<query>`) and returns matching videos       |
| `scrape`   | Fetches full metadata (title, cast, categories, runtime, embed source) for one video |
| `scene:list:page` | `scene:list` with a `url` lists an `/actress/<slug>` or `/category/<slug>` page (paginated `/<page>`); subscriptions use it to follow performers |

## URL patterns

- Listing page: `/hdporn/<page>` — video cards under `section.box.feature`
- Video card: `h3.meta-data-title > a[href="/hdporn/<id>-<slug>.html"]`, poster `img#cover_<id>`
- Search: `/?q=<query>` and `/?q=<query>&p=<page>`
- Video page: `/hdporn/<id>-<slug>.html`
  - Title: `h1.main-h1`
  - Runtime: `li.icon.fa-clock-o` in the header meta list
  - Cast: `a[href^="/actress/"]`
  - Categories: `a[href^="/category/"]`
  - Player source: `altplayer.php?i=<embedUrl>` marker (fallback: first non-ad `iframe`)

## Data mapping

- `external_id` — `hqporner-<numericId>`
- `adult` — always `true`
- `genres` / `tags` — category chips
- `credits` — actresses as `cast`
- `downloads` — single embed URL, resolved later by a resolver plugin

## Development

```sh
bun install
bun test
```

Live tests hit hqporner.com and need network access.
