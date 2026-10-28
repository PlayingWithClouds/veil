---
paths:
  - "apps/backend/internal/ingest/**"
  - "apps/backend/internal/alike/**"
  - "apps/backend/internal/names/**"
  - "apps/backend/internal/db/**"
  - "apps/backend/internal/cache/**"
  - "apps/backend/internal/media/**"
  - "apps/backend/internal/pipeline/**"
  - "apps/backend/internal/jobs/**"
  - "apps/backend/internal/updater/**"
  - "apps/backend/internal/api/graphql/**"
---

# Data model and ingest

## Entities

Tags are flat. Collections don't nest.

- scene: one video, the atomic unit. No series/season/episode; episodic content is a collection. Studio link, performers, tags, preview assets, streams.
- gallery: image set, the stills counterpart of scene.
- image: one row per image. `content = true` (gallery pages, standalone) is browsable/taggable/collectible; `content = false` is an asset (poster, logo, profile shot).
- performer: bio, `aliases`, `tags`.
- studio: `aliases`, `tags`, `parent` (network → studio), identity `source_url`.
- collection: ordered members of any type (scene/gallery/image/performer/studio). `user_created` = playlist; otherwise scraper-made (site series, DVD; `created_by`/`source_url`).
- tag: `aliases` merge synonyms at ingest.
- scene_marker: tag and/or label at `seconds`, optional `end_seconds`. `personal` = user-made; otherwise global (`created_by`, e.g. plugin position tagger).

Tag search inherits: `scenes(tagId)` matches the scene's own tags, its studio's, or any credited performer's, direct first, exposed as `tagMatch: "direct" | "inherited"`. Same for galleries and images (images also via their owning gallery).

## Storage helpers

- `internal/db`: `"table:id"` record ids, JSON/BOOLEAN column decoding, `$name` params. `db.Open` applies migrations and tracks versions in `_migration`.
- `internal/cache`: expiring kv on the `kv` table.
- `internal/jobs`: DB-backed job queue + worker pool, VPN-gated. `internal/pipeline`: scene visits (`visit.go`), jobs (download, speed-check, enrich).

## Ingest

- Observation-first: raw plugin payloads land in `observation`, merge into canonical records by identity (source_url → external_id → name), per-field history in `changes`.
- Search/browse results are stubs only (listing fields, no scrape). A scene's detail page is fetched once, on first visit (`ensureSceneStreams` → `Orchestrator.VisitScene`, `scene.detail_fetched_at`). That fetch links the site's related videos (`scene_related`, source `site`); under 24, background searches by performer, tag, then title on the origin plugin fill the rest.
- A gallery page is fetched on open while it has no images (`ensureGalleryImages` → `Orchestrator.VisitGallery`).

## Duplicates and merges

- Performers, studios, tags match by spelling-insensitive `name_key` (`internal/names`); `RepairNamedRecords` merges older duplicates at startup.
- Scenes (`alike.SameSceneStrict`), galleries (`alike.SameGallery`) and scraper collections (same name + a shared member) merge via `MergeDuplicateMedia` at startup and, for new scenes/galleries, at ingest.
- A merged-away record's `source_url`/`external_id` stay in `entity_identity` and resolve to the survivor.
- `alike.SameScene` (loose, hide-only): runtimes within 5%, then identical credited performers or near-identical title. No runtime never matches. `alike` also does pHash + performer/duration alternate-source matching.
- Stage names: `Orchestrator.ProfilePerformers` (start + every 6 h, 100 per run) searches `freeones` for performers without a profile and ingests the page naming them. Its aliases are absorbed; any performer whose name is one of them (full names, 2+ words only) merges in (`mergeAliasTwins`, also in startup repair).
- Manual: `mergeRecords(survivorId, mergedId)`; web ⋮ → "Merge into…" on performer/studio pages, a button on tag pages.

## Subscriptions

`search_subscription` (`kind` search | studio | performer | tag) re-runs every `interval_hours` (`internal/updater`). Searches use the live-search discovery path. Entity kinds list the entity's site page via `scene:list:page` when a plugin supports it, else keyword-search the name on its origin plugins (keeping scenes credited to it or uncredited), and always add library scenes credited to it. Results go to `search_subscription_item`; the first run counts as seen, later arrivals as new.
