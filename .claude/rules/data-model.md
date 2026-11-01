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
  - "apps/backend/internal/api/graphql/**"
---

# Data model and ingest

## Schema is the source

`schema/*.schema` (OpenSchema, entry `schema/veil.schema`) defines every stored model. The long-term scope is in GitHub issues labelled `scope`; only part of it is built.

- `bun run schema` (root) generates `apps/backend/internal/schema/schema.go` and `packages/veil-sdk/src/schema/schema.ts`. Never edit those by hand.
- Database changes: edit the schema, then `openschema migrate schema/veil.schema -t migration-sqlite --numbered apps/backend/internal/db/migrations --name <name>` writes the next migration and updates `openschema.snapshot.json`. Commit both. Hand-written data migrations still go in a numbered file.
- `openschema` is not on npm: build it from the OpenSchema repo (`bun build src/cli/index.ts --compile --outfile ~/.local/bin/openschema`).
- Only decorators that change generated output: `@abstract`, `@value`, `@index`, `@table`, `@invisible`. Links are the `ref<T>` type. Behaviour goes in comments.
- Names are exact: tables `"Video"`, columns `"createdAt"`. Always `db.Quote` them.

## Storage helpers

- `internal/db`: uuid ids (`db.NewID`), `Insert`/`Update`/`Trash`/`Delete`, `Get[T]`, `QueryAs[T]`, `$name` params. Lists/maps/value objects are JSON text, booleans 0/1; `DecodeRow` decodes by the target struct's field types. Foreign keys are on. `db.Open` refuses databases from the pre-schema migration history.
- Records are trashed (`deletedAt`), not deleted; every read filters `"deletedAt" IS NULL`. Site-owned rows (streams, gallery images) are replaced for good.
- `internal/cache` (`CacheEntry`), `internal/settings` (`Setting`, plugin `""` = global), `internal/jobs` (`Job`; no worker runs jobs yet).

## Ingest (`internal/library`)

- Identity: one `Source` row per site a record was found on (`plugin` + `externalId`, unique; `externalId` falls back to the URL). Ingesting the same site record again updates the record its source points at.
- Performers, studios and tags without a site identity match by `nameKey` (`internal/names.Key`); a site identity, when given, is added as a source.
- `IngestFound` (a page from `<entity>:find`): root first (video/gallery), then standalone performers/studios/tags. Marks the root's source `fetchedAt`, replaces its streams (`Stream`, per source) and gallery images.
- `IngestListed` (a listing card): stub only; doesn't touch credits. Unfetched sources are read on first open (`pipeline.Visit` via `fetchVideo`/`fetchGallery`).
- Not back yet: cross-site merging of the same video (`alike.SameSceneStrict` is still there, unused), per-field provenance (`FieldSource`), revisions.

## API (`internal/api/graphql`)

Hand-written SDL in `schema/*.graphql`; library types are bound to the generated `internal/schema` structs in `gqlgen.yml`, and reference fields (`studio`, `performers`, `tags`, …) resolve to records. After SDL changes: `go run github.com/99designs/gqlgen generate`, then `bun run codegen` in `packages/veil-sdk`. gqlgen moves helper functions out of `*.resolvers.go` into a comment; keep helpers in `helpers.go`.
