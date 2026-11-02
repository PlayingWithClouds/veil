---
paths:
  - "schema/**"
  - "apps/backend/internal/schema/**"
  - "apps/backend/internal/db/**"
  - "apps/backend/internal/library/**"
  - "apps/backend/internal/pipeline/**"
  - "apps/backend/internal/alike/**"
  - "apps/backend/internal/names/**"
  - "apps/backend/internal/cache/**"
  - "apps/backend/internal/jobs/**"
  - "apps/backend/internal/settings/**"
  - "apps/backend/internal/provenance/**"
  - "apps/backend/internal/user/**"
  - "apps/backend/internal/devices/**"
  - "apps/backend/internal/api/graphql/**"
---

# Data model and ingest

## Schema is the source

`schema/*.schema` (OpenSchema, entry `schema/veil.schema`) defines every stored model. The long-term scope is in GitHub issues labelled `scope`; only part of it is built.

- `bun run schema` (root) generates `apps/backend/internal/schema/schema.go`, `ordinals.go` (field ordinals per table from `schema/openschema.lock`, via `scripts/schema-ordinals.ts`) and `packages/veil-sdk/src/schema/schema.ts`. Never edit those by hand.
- Database changes: edit the schema, then `openschema migrate schema/veil.schema -t migration-sqlite --numbered apps/backend/internal/db/migrations --name <name>` writes the next migration and updates `openschema.snapshot.json`. Commit both. Hand-written data migrations still go in a numbered file.
- `openschema` is not on npm: build it from the OpenSchema repo (`bun build src/cli/index.ts --compile --outfile ~/.local/bin/openschema`).
- Only decorators that change generated output: `@abstract`, `@value`, `@index`, `@table`, `@invisible`. Links are the `ref<T>` type. Behaviour goes in comments.
- Names are exact: tables `"Video"`, columns `"createdAt"`. Always `db.Quote` them.

## Storage helpers

- `internal/db`: uuid ids (`db.NewID`), `Insert`/`Update`/`Trash`/`Delete`, `Get[T]`, `QueryAs[T]`, `$name` params. Lists/maps/value objects are JSON text, booleans 0/1; `DecodeRow` decodes by the target struct's field types. Foreign keys are on. `db.Open` refuses databases from the pre-schema migration history. `Tx` inside a transaction joins it.
- Records are trashed (`deletedAt`), not deleted; every read filters `"deletedAt" IS NULL`. Site-owned rows (streams, gallery images) are replaced for good.
- `internal/cache` (`CacheEntry`), `internal/settings` (`Setting`, plugin `""` = global).

## Writes: provenance and revisions (`internal/provenance`)

Every update of a video, gallery, performer, studio or tag goes through `provenance.Recorder.Update` with an `Author` (`User()`, `Plugin(name, sourceID)`, `System()`); never `db.Update` those tables directly.

- `FieldSource` per descriptive field of videos, galleries, performers, studios: the first source to fill a field owns it; a plugin only writes fields it owns or nobody owns. User writes lock the field (`locked`); `SetLocked(…, false)` releases it to the next source. New records `Claim` the fields their first source filled.
- `Revision` per changed field (before/after JSON, author), skipped when the value didn't change; also for `deletedAt` (trash/restore). Creation has none. `Revert` writes `before` back as a user edit (locked). Fields are named by ordinal (`schema.FieldOrdinals`, `provenance.Column`).
- `UserState` and `SavedFilter` have no revisions: `Revision.subjectKind` is a `RecordKind`, and those tables aren't one.

## Trash (`internal/library/trash.go`)

`Trash`/`Restore` (videos, galleries, performers, studios, tags) are revisions on `deletedAt`. A trashed record keeps its sources: re-ingesting updates it but it stays trashed, and credits by name land on the trashed performer/tag (`findByNameKey` includes trashed rows). `PurgeTrash` runs every 6 h for records older than `trashRetentionDays` (default 30): it deletes sources, streams, images, user state, provenance, revisions and clears references (studio columns, JSON id lists); events stay.

## User layer (`internal/user`)

- `UserState` per record (rating 0..10, favourite, notes, blocked, progress, viewCount, lastViewedAt), created on first write. Blocked records, and media crediting a blocked performer/studio/tag, are left out of every list.
- `Event` log: `Record` stamps the current device; `view` bumps viewCount/lastViewedAt, `progress` (`data.position`) sets the resume point, `finish` clears it. o-count, recent searches, history are queries over it/UserState.
- `SavedFilter` CRUD; a saved filter plays through `videos(filter:)`.

## Filter (`internal/library/filter.go`)

`schema.Filter` is the one filter for media lists (`MediaQuery`). Media is queried as `"media"` left-joined with its `UserState` as `"state"`. Performers, studios, categories (via `matchTags`) and sources match any; `includeTagIds` all, `excludeTagIds` none; duration and VR narrow videos only. `kinds` isn't applied by the per-kind lists.

## Jobs (`internal/jobs`)

`Queue` (`Job` rows, dedupe keys, retries) and `Runner`: `Register(kind, Limits, Task)` per kind (concurrency and backoff from settings `kindLimits`), `Start`, `Cancel`. Succeeded jobs stay listed; finished ones (except downloads) are purged after 7 days. Registered now: `scrape` (subject → `pipeline.Visit`, queued by `queueFetch`).

## Devices (`internal/devices`)

This install's `Device` row (`current`) is created on first start (`DEVICE_NAME` else host name, `DEVICE_PLATFORM` default `server`); events and jobs carry its id. `StorageLocation`s per device: a `library` and a `cache` location at `DATA_DIR/blobs` are ensured at start.

## Ingest (`internal/library`)

- Identity: one `Source` row per site a record was found on (`plugin` + `externalId`, unique; `externalId` falls back to the URL). Ingesting the same site record again updates the record its source points at.
- Performers, studios and tags without a site identity match by `nameKey` (`internal/names.Key`); a site identity, when given, is added as a source.
- `IngestFound` (a page from `<entity>:find`): root first (video/gallery), then standalone performers/studios/tags. Marks the root's source `fetchedAt`, replaces its streams (`Stream`, per source) and gallery images.
- `IngestListed` (a listing card): stub only; doesn't touch credits. Unfetched sources are read on first open (`pipeline.Visit` via `fetchVideo`/`fetchGallery`).
- Not back yet: cross-site merging of the same video (`alike.SameSceneStrict` is still there, unused).

## API (`internal/api/graphql`)

Hand-written SDL in `schema/*.graphql` (`library`, `user`, `system`, `plugins`, `app`); library types are bound to the generated `internal/schema` structs in `gqlgen.yml`, schema enums too (GraphQL values spelled as in the schema: `video`, `addedAt`), f32 fields through `scalars.Float32`, and reference fields (`studio`, `performers`, `tags`, …) resolve to records. After SDL changes: `go run github.com/99designs/gqlgen generate`, then `bun run codegen` in `packages/veil-sdk`. gqlgen moves helper functions out of `*.resolvers.go` into a comment; keep helpers in `helpers.go`.
