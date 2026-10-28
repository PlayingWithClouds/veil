# Veil

Self-hosted adult media library, one user. Bun monorepo: Go backend, SvelteKit web, native Android app, TypeScript scraper plugins. Backend is self-contained (in-process SQLite, blobs on disk), so it runs on a server or embedded in the phone app.

Subsystem notes are in `.claude/rules/` and load with matching files. Read the relevant one before changing how a subsystem behaves.

## Layout

- `apps/backend/`: Go, module `github.com/playingwithclouds/veil`, entry `cmd/server/main.go`. Under `internal/`: `db/migrations/` schema, `api/graphql/` resolvers (SDL in `schema/*.graphql`), `media/` all GraphQL queries, `ingest/` `alike/` `names/` merging, `pipeline/` `jobs/` visits and job queue, `searchrank/` `discovery/` search, `recommend/` `updater/` recommendations and feeder, `plugins/` `pluginstore/` plugin runtime and install, `cams/` `livebuffer/` live cams, `stream/` `resolve/` `storage/` `blobcache/` streaming and blobs.
- `apps/backend/plugins/`: scraper plugins, bundled by `scripts/bundle-plugins.ts`.
- `apps/web/`: SvelteKit 2, Svelte 5 runes, Tailwind v4.
- `apps/mobile/`: Kotlin + Compose, embeds the Go backend.
- `packages/veil-sdk/`: shared TS types, genql client, plugin SDK. `packages/config/`: eslint/tsconfig.

## Commands

- `bun dev` (root): bundler watch + backend + web + FlareSolverr. State in `DATA_DIR` (default `apps/backend/data`); wipe = stop backend, delete it.
- After SDL changes, in order: `go run github.com/99designs/gqlgen generate` (`apps/backend`), `bun run codegen` (`packages/veil-sdk`), `bun run check` (`apps/web`), `./gradlew generateVeilApolloSources` (`apps/mobile`).
- Tests: `go test ./...` (`apps/backend`), `bun test` (plugin dir), `bunx playwright test`, `bun run test` (`apps/mobile`).
- Android: `bun run apk` / `bun run android` in `apps/mobile`.

## Git

- Work on `main`. No branches, no PRs.
- Deferred work → GitHub issue on `PlayingWithClouds/veil` via `gh` as `PlayingWithClouds` (`gh auth status`, `gh auth switch`).
- Dirty worktree at start: commit it first, so your changes stay separate.
- Commit after each substantial change; push if a remote exists.
- Title: short, what changed. Body: `AI Generated`, plus anything the diff doesn't show. Don't restate the diff or summarise.
- No trailers, no `Co-Authored-By`.

## Invariants

- Migrations are append-only: never edit an applied one, add a new numbered file.
- API results never expose raw record ids, only `..._id: string`.
- JSON list columns (`tags`, `performers`, `aliases`): match with `json_each`, bind lists as `$param` → `IN (SELECT value FROM json_each($param))`.
- Merges are irreversible, so merge matchers stay strict (`alike.SameSceneStrict`, `SameGallery`). The loose `alike.SameScene` only hides, never merges.
- `internal/plugins/types.go` ⇄ `packages/veil-sdk/src/types.ts`: change both in one commit.
- Plugins run in goja, not Node: web APIs, SDK, pure-JS packages only. `node-html-parser` is a native Go subset (`embedded_html.go`); passing `bun test` doesn't prove parsing works.
- Plugins ship as GitHub release assets, never npm. Fix = bump `version`, `bun scripts/publish-plugins.ts <plugin>` in `apps/backend`.
- Face identification names anybody on public face search websites. These websites are themselves responsible for the data they publish.
- UI uses the existing design system. Web: tokens/classes in `apps/web/src/routes/layout.css`; reuse `SceneCard`, `TagChips`, `MemberCard`, `CollectionDialog`, `Lightbox`. Android: only `ui/theme` tokens and `ui/design` components (`DesignTokenUsageTest`).

## Scope

Do what was asked, at the intended scope. Don't change behaviour, public APIs, data shapes, validation or side effects unless asked. If the request looks wrong or there's a clearly better way, say so in a sentence, then do it as asked. Make routine calls yourself; ask only when readings lead to materially different work.

## Working on your own

When I say you can do something yourself, I'm not watching. Don't end a turn with a summary that announces the next step, an offer to continue, or a list of non-blocking decisions; put status notes and recommendations alongside the next action and keep going. Stop only when nothing can move without me. Risky or irreversible actions still need my confirmation.

## Code style

Readability over cleverness. Surrounding code wins over these rules.

- Every function gets a doc comment saying what it does (`// Name …` Go, `/** … */` TS). No `@param`/`@returns`; the signature has the types.
- Other comments say what skimming won't: reason, constraint, gotcha. Never narrate code or mention removed features. Good: `// Normalize external IDs before database lookup.` Bad: `// Now we loop through the items.`
- Full names: `recordId`, not `rid`. `camelCase` unless the language or code says otherwise.
- Explicit control flow: `if (timeout === undefined) { timeout = DEFAULT_TIMEOUT }` over `??`, ternaries, compact conditionals, unless they save real repetition.
- Svelte: `class:rotate-0={open}`, not string-built classes.
- Inline single-use variables.
- Max two nesting levels: guard clauses, early returns, helpers.
- Short single-purpose functions; a named helper beats commented inline logic.

## Tests

After writing a change, decide if it needs a test; put it beside the code (`*_test.go`, plugin `tests/*.test.ts`, Playwright for browser flows). A failure means investigate, not work around.

## Tools

Edit with Read/Write/Edit, not `sed`, `awk` or Python.
