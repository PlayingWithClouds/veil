---
paths:
  - "apps/backend/internal/searchrank/**"
  - "apps/backend/internal/discovery/**"
  - "apps/backend/internal/searchlab/**"
  - "apps/backend/cmd/search-lab/**"
  - "apps/backend/searchlab/**"
  - "apps/web/src/lib/liveSearch*"
  - "apps/web/src/lib/searchOrder*"
  - "apps/web/src/lib/searchLog*"
---

# Search

## Rules

- SSE streams in arrival order; clients must sort by `score`.
- All searchrank weights live in `searchrank/weights.go`.
- Typo credit must keep exact matches ahead.
- A remembered click only boosts a scene the search actually returns.
- The `content` filter never applies to library matches.
- A new search-lab idea is a new `Theory` in `internal/searchlab/theories.go`.

## `/api/search` (SSE)

Library matches first, then each plugin's results as they arrive.

- Library: FTS5 `scene_search` (migration 030, trigger-maintained) over title + performer/studio/tag names, bm25 order, title-substring fill (`ingest.SearchLibrary`). Performers, studios, tags, galleries rank exact → prefix → word-prefix → substring (`media.searchOrder`).
- Every result carries `score`: `3.0·text + 1.5·fusedRank·pluginShare + 0.4·(quality−0.5) + personalWeight·taste + 2.0·entity + 1.2·clickMemory`. fusedRank = reciprocal-rank fusion of the result's place in its site list (library bm25 order counts as a list). Quality = site rating, log view count. Taste from `recommend.Engine.Profile`.
- Cross-site copies (`alike.SameScene`, or same record) collapse: a later copy is dropped unless it scores better, then it's sent with `replaces` (old copy's source URL).
- A `site` event (`plugin, count, failed`) follows each plugin's keyword search.
- `content` param (`videos`, `cam-shows`, `jav`, `galleries`) keeps plugins whose `meta.content` carries it (`discovery.CarryingContent`; undeclared: scene lister = videos, gallery lister = galleries), passed on as `ListArgs.content` (kvs: `SiteConfig.content`).

## Scoring parts

- Entities: `searchrank.EntityIndex` (performer/studio/tag names, then aliases, by `names.Key`, cached 2 min, longest match first; one-word performer only as the whole query; near-miss of a 6+ letter name = `fuzzy`). The `matched` event lists them (`kind, id, name, alias?, fuzzy?`, `rewritten`). Credited scenes get entity credit (uncredited stub with a title naming it: 0.7; listed on the entity's page counts as credited). Up to 2 exactly spelled performers/studios (a tag only as the whole query) also list their site page via `scene:list:page` (`updater.Scheduler.PagePlugins`, `discovery.ListPage`) as an extra rank list.
- Typos: a word within 1 (4–7 letters) or 2 (8+) edits of a field word, same first letter, earns 0.6 of whole-word credit. Under 5 exact library matches, `scene_search_fuzzy` (FTS5 trigram, migration 033) adds candidates at text ≥ 0.4. Queries using an alias or near-miss also run `rewritten` on library and every plugin; its text counts 0.92.
- Taste weight: `searchrank.Specificity` (word count, share covered by entity/tags, library match count) sets it from 0.6 (broad, "anal") to 0.075 (title-like).
- Clicks: web search (Home, `/search`) logs the top 24 once the list is stable 2.5 s, plus clicks, via `recordImpressions` with surface `search` and `query` (`src/lib/searchLog.ts`) → `search_event` (90 days). `searchrank.History` builds per-(query, scene) click memory (30-day half-life, similar queries by word Jaccard) and per-plugin rank-share factors in [0.5, 1.5] from smoothed click rates of the query's words.

## Web client

`LiveSearch` (`src/lib/liveSearch.svelte.ts`; Home search, `HomeSearch`, TV `/search`) holds the first second, sorts it, sorts later arrivals in below the first 12 (`src/lib/searchOrder.ts`); Relevance / Newest / Most viewed / Rating / Duration switch. Library matches are their own list ("In your library"); a site copy replacing a library one is dropped. Site chips show spinner/count/warning. `ContentKindPicker` ("What are you looking for?", `?content=`) shows when sites carry more than one kind. A cam room's "Past shows" searches the username in `cam-shows`. Recognised performers/studios as channel cards (`SearchEntityStrip`), videos in a grid, galleries as a portrait shelf after two rows.

## Search lab (`apps/backend`)

`searchlab/queries.json` lists queries (easy → hard) and plugins. `go run ./cmd/search-lab fetch` (`FLARESOLVERR_URL` for missav) runs them through `data/plugins` bundles into `searchlab/snapshot.json`, keeping grades. `label` serves a grading page on 127.0.0.1:8099 (0 wrong, 1 loose, 2 fits, 3 exact; keys 0–3) → `searchlab/labels.tsv` (hand-editable). `eval` scores every theory (nDCG@10/24, P@10) and compares each site list's labelled precision with its list fit.
