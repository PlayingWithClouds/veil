# Veil

Self-hosted adult media library, single user. Go backend + SvelteKit frontend in a Bun monorepo, plugin-based scrapers. The backend is self-contained (in-process SQLite + blobs on disk) so it can run on a server or embedded in a phone app.

## Git

Work happens on `main`. No branches, no PRs. Deferred work goes into GitHub issues on `PlayingWithClouds/veil`, created with `gh` as the `PlayingWithClouds` account (check `gh auth status`; switch with `gh auth switch` if another account is active).

Commit the worktree first if it's dirty, then write your changes. Commit after every big change and push if a remote is configured. Short, to-the-point commit title; a body explaining the change when the title doesn't carry it; ask if you're unsure what to write. Always say the commit was made by you, not a human.

No trailers, ever: no `Co-Authored-By`.

## Writing

Commits carry only what matters. Say the thing, explain what a reader won't see for themselves, stop. No restating the diff, no summarising what you just said, no section that exists because the format seemed to want one.

## Layout

| Path | Purpose |
| ---- | ------- |
| `apps/backend/` | Go backend (module `github.com/playingwithclouds/veil`) |
| `apps/backend/plugins/` | Scraper plugins (TypeScript source; bundled by `apps/backend/scripts/bundle-plugins.ts`) |
| `apps/backend/internal/db/migrations/` | Versioned SQLite schema (`NNN_*.sql`), embedded, applied on startup, append-only |
| `apps/web/` | SvelteKit 2 + Svelte 5 (runes) + Tailwind v4 frontend |
| `apps/mobile/` | Native Kotlin + Jetpack Compose Android app with the embedded Go backend (see Mobile app) |
| `packages/veil-sdk/` | Shared TS types + genql GraphQL client + plugin SDK |
| `packages/config/` | Shared eslint/tsconfig |

### Key backend paths

| Path | Purpose |
| ---- | ------- |
| `cmd/server/main.go` | Entry point, wires everything |
| `internal/db/` | SQLite open/migrate + query helpers (`"table:id"` record ids, JSON/BOOLEAN column decoding, `$name` params) |
| `internal/cache/` | Expiring key-value store on the SQLite `kv` table |
| `internal/storage/` | On-disk blob store, served at `/api/blob/`; optional at-rest encryption of `stream-cache/` (`BLOB_ENCRYPTION_KEY`, see Mobile app) |
| `internal/blobcache/` | Cache size limit: LRU eviction of `stream-cache/` + `img-cache/` blobs |
| `internal/plugins/` | Registry (hot reload) + embedded JS runtime (goja, native HTML parser) + types; the Bun runner remains only for the parity test |
| `internal/ingest/` | Observation→canonical merge, relation linking, collection members, markers |
| `internal/pipeline/` | Orchestrator: scene visits (`visit.go`), job execution (download, speed-check, enrich) |
| `internal/jobs/` | DB-backed job queue + worker pool (VPN-gated) |
| `internal/discovery/` | Plugin keyword search + stub ingest, shared by `/api/search` and search subscriptions |
| `internal/updater/` | Scheduler that re-runs due search subscriptions |
| `internal/media/` | Repository: all GraphQL read/write queries |
| `internal/api/graphql/` | gqlgen resolvers; SDL in `schema/*.graphql` |
| `internal/stream/` | HLS/MP4 proxy + write-through blob cache |
| `internal/resolve/` | Stream resolution chain (cache → domain plugin → wildcard → direct) |
| `internal/alike/` | Cross-plugin alternate-source matching (pHash + performer/duration score) |
| `internal/recommend/` | Recommendation engine: taste profile, candidate sources, ranking, re-ranking; all tunables in `weights.go` |

After changing SDL: `go run github.com/99designs/gqlgen generate` in `apps/backend`, then `bun run codegen` in `packages/veil-sdk` (regenerates the genql client the frontend uses), then `bun run check` in `apps/web`.

## Entity model

Flat tags, no hierarchy.

- **scene** — atomic unit: one video. No series/season/episode; episodic content is a collection. Carries a studio link, performers + tags, preview assets, streams.
- **gallery** — image set; parallels scene for stills.
- **image** — one row per image. `content = true` rows (gallery pages, standalone images) are browsable/taggable/collectible; `content = false` rows are assets (posters, logos, profile shots).
- **performer** — bio + `aliases`, `tags`.
- **studio** — `aliases`, `tags`, `parent` link (network → studio hierarchy), `source_url` identity.
- **collection** — ordered members of any entity type (scene/gallery/image/performer/studio), no nesting. `user_created` = user playlist; otherwise scraper-created (site series, DVD; `created_by`/`source_url`).
- **tag** — flat; `aliases` merge synonyms at ingest.
- **scene_marker** — tag and/or label at `seconds` (optional `end_seconds` span). `personal` = user-created; otherwise global (e.g. plugin position tagger, `created_by`).

Tag search is inherited: `scenes(tagId)` matches the scene's own tags OR its studio's OR any credited performer's, ranked direct-first and surfaced as `tagMatch: "direct" | "inherited"`. Same for galleries/images (images also inherit via their owning gallery).

## Data rules

- Results from a fetch never expose the raw record ID structure — only ever `..._id: string`.
- Discovery is search-driven: search/browse results are ingested as stubs only (listing fields, no scrape). A scene's detail page is fetched once, when it is first visited (`ensureSceneStreams` → `Orchestrator.VisitScene`, tracked by `scene.detail_fetched_at`). That fetch also links the site's related videos (`scene_related`, source `site`); below 24 related, background fallback searches by performer, tag, then title fill the rest on the origin plugin. A gallery's page is fetched on open while it has no images (`ensureGalleryImages` → `Orchestrator.VisitGallery`).
- Subscriptions (`search_subscription`, `kind` = search | studio | performer | tag) re-run every `interval_hours`. Searches go through the same discovery path as live search. Studio/performer/tag subscriptions list the entity's site page via `scene:list:page` when a plugin supports it, else keyword-search the name on its origin plugins (keeping only scenes credited to it or not yet credited), and always add library scenes credited to it. Results collect in `search_subscription_item`; the first run's results count as seen, later arrivals as new.
- Ingest is observation-first: raw plugin payloads land in `observation`, merge into canonical records by identity (source_url → external_id → name), per-field history in `changes`.
- Migrations are append-only (`db.Open` applies them and tracks versions in `_migration`); never edit an applied migration, add a new numbered file.
- Lists of links (`tags`, `performers`, `aliases`) are JSON columns; match them with `json_each`. Bind lists as `$param` and use `IN (SELECT value FROM json_each($param))`.

## Recommendations

`internal/recommend` runs candidate generation → ranking → re-ranking per request, no ML or external calls.

- **Signals** (per scene, time-decayed with a 30-day half-life): o-event > like > completed watch > save (watchlist, user collection) ≈ A/B pick > partial watch > click (first visit or logged click); negative: dislike, early abandon (<10 s or <5% of `watch_history.max_progress_seconds`), A/B reject, shown-but-not-clicked (`recommendation_impression`). Blocked entities pin to −1.
- **Profile**: computed on demand (no cache table to invalidate); signalled scenes spread their weight onto performers, tags (÷√count), studio and site (first observing plugin), each dimension normalized to [−1, 1].
- **Candidate sources**, each item tagged with one: `scene_related` of the strongest signalled scenes (site-related highest prior), new search-subscription finds, newest scenes of top performers/studios/tags, exploration (random scenes of tags adjacent to the top tags, plus random), unvisited stubs from recent searches, newest scenes. Cold start = the last two.
- **Ranking**: affinity match + source prior + freshness − watched/fatigue/disliked-neighbour penalties. Blocked scenes, disliked scenes and inactive-plugin stubs are excluded. **Re-ranking**: per-24 page caps per performer/studio/site (plugin), no source or site three in a row, every 7th slot exploration.
- The ranked feed is memoized 30 min for `offset > 0` so pages stay stable; `offset 0` re-ranks once the ranking is 10 min old, or on `refresh: true` (Home's refresh button). The web logs impressions in batches via `recordImpressions` (`src/lib/recommendations.ts`).

## Plugin system

Plugins are TypeScript packages under `apps/backend/plugins/`, all importing `@playingwithclouds/veil-sdk` (`runPlugin` dispatcher + shared types). The backend never runs their source: `scripts/bundle-plugins.ts` bundles each into one CommonJS file in `DATA_DIR/plugins/<name>/plugin.js` (the installed-plugins folder, same on server and phone), and the backend runs it in an embedded JS engine (goja, `internal/plugins/embedded*.go`), a fresh runtime per invocation, hot-reloaded via fsnotify. `bun dev` runs the bundler in watch mode, so saving a plugin file reloads it within a second.

**Distribution via GitHub, never npm:** plugins are assets of this repository's rolling `plugins` release: per version an npm-style tarball `<folder>-<version>.tgz` (`package/package.json` + `package/plugin.js`, keyword `veil-plugin`), plus `index.json` listing every package's latest version with tarball URL and sha512 integrity. Packages are named `@playingwithclouds/veil-plugin-<name>`; the folder name is the package name minus scope and `veil-plugin-`. `internal/pluginstore` reads the index (`DefaultIndex`, override with `PLUGIN_INDEX`), installs packages (integrity checked), installs the default set on a fresh backend, updates to newer versions at start and every 6 h, and seeds from `PLUGIN_SEED_DIR` (the phone app's bundled set) without undoing updates or uninstalls (`.store.json` in the plugin folder). Dev bundles carry `localBuild: true` and are never replaced. To ship a fix: bump `version` in the plugin's `package.json` (meta reads it from there), then `bun scripts/publish-plugins.ts [plugin …]` in `apps/backend` (needs `gh` logged in as PlayingWithClouds; `--dry-run` to preview; skips versions already in the index; uploads `index.json` last). Every backend picks it up within 6 h or on "Check for updates" in Plugins; the phone checks on every launch.

**Runtime rules** (what a plugin may use): web APIs (`fetch`, `URL`, `URLSearchParams`, `AbortSignal.timeout`, `JSON`, promises), the SDK, and pure-JS npm packages (the bundler inlines them). No Node/Bun built-ins, files, processes or native addons. `node-html-parser` is special: bundles keep it external and the backend answers `parse()` natively in Go (x/net/html + cascadia), because parsing in JS under goja is ~60× slower. Only `querySelector(All)`, `getAttribute`/`hasAttribute`, `closest`, `text`/`rawText`, `innerHTML`/`outerHTML`/`toString`, `tagName`, `id`, `classList` and `parentNode` exist there (`embedded_html.go`); add to it before relying on anything else. `bun test` in a plugin still runs the real node-html-parser, so check changed parsing against the embedded runtime: `bun scripts/bundle-plugins.ts --out /tmp/bundles && VEIL_EMBEDDED_BUNDLES=/tmp/bundles go test ./internal/plugins -run EmbeddedBundlesLive -v` in `apps/backend` hits the live sites and compares every plugin's list/find output with Bun's (xhamster's tag order varies per request).

Plugins may send the site's own `rating` (0..10) and `view_count` on `Scene` and on listing `DiscoveredItem`s; cards show them. Only eporner does so far.

**Type mirrors — keep in sync in one commit:** `apps/backend/internal/plugins/types.go` ⇄ `packages/veil-sdk/src/types.ts`.

**Capabilities:** `meta`, `discover`, `scrape`, `search`, `enrich`, `resolve` (embed/page URL → playable mp4/HLS + headers; `meta.domains` routes, `"*"` = wildcard), `browse:categories`, `browse:actors`, `scene:list:page` (`scene:list` with a `url` arg lists a channel/performer/category page; eporner, hqporner).

**Protocol** (stdin/stdout JSON): `{ "capability", "args" }` in, `{ "result" } | { "error" }` out — except `scrape`, which streams NDJSON: one `ScrapeResult` per line, root entity (scene/gallery/image/collection) first, then related standalone entities (performer/studio/tag).

- `Scene.studio` is a structured `StudioRef { name, external_id?, source_url? }`.
- Plugins may emit `Collection` with ordered `member_urls`; unresolved members are auto-queued for scraping.
- Plugins may emit `Scene.markers` (tag/label + seconds/end_seconds) — ingested as global markers.
- `scene:find` should fill `Scene.related` with the page's related-videos list as `DiscoveredItem`s, using the same external_id scheme as the plugin's listings (reuse the listing card parser).

**FlareSolverr:** set `meta.requires_solver: true` when the site answers plain requests with a Cloudflare challenge, and fetch through `fetchHtmlSmart`. Without `FLARESOLVERR_URL` (e.g. embedded on a phone) the registry treats such plugins as unavailable: they drop out of search/browse, their stubs are hidden, visits skip them. Currently: missav, spankbang.

**Current plugins:** aylo (pornhub, redtube, youporn, tube8), eporner (API + enrich), hqporner, kvs (37 Kernel Video Sharing sites from one config list), missav, xhamster, spankbang, tnaflix, txxx-network (scenes); pornpics (galleries).

## Frontend

- GraphQL only, via `gqlClient` (`apps/web/src/lib/veil.ts`) with inline genql selections in `src/lib/*.ts` data modules; subscriptions over graphql-ws (`streamsChanged`).
- Internal links go through `src/lib/routes.ts` (`sceneUrl`, `performerUrl`, `tagUrl`, ...).
- Design tokens + component classes live in `src/routes/layout.css`, TV/D-pad focus in `src/lib/tv.css` + `FocusGrid`/`FocusList`. Reuse `SceneCard`, `TagChips`, `MemberCard`, `CollectionDialog`, `Lightbox` before writing new components. For anything new, follow the tokens and patterns already in `layout.css`.
- Stream playback URLs are proxied through the backend (`/api/stream/`, CORS *); images through `/api/img` (`cacheUrl` helper).

## Running

```sh
bun dev   # process-compose: plugin bundler (watch) + backend + web native, FlareSolverr via docker compose
```

Backend state lives in `DATA_DIR` (default `apps/backend/data`): `veil.db` + `blobs/` + `plugins/` (bundles). Wipe = stop the backend and delete that directory. `DNS_SERVERS` (comma-separated) or `DNS_SERVERS_FILE` (re-read per lookup) override the resolver where Go can't find the system one (Android). `HOST` sets the bind address (default all interfaces), `PLUGIN_INDEX` the plugin index URL.

Cache blobs (`stream-cache/`, `img-cache/`) are capped at `CACHE_MAX_BYTES` (default 300 MB, `blobcache.DefaultMaxBytes`): once over, the least recently used are evicted down to 90%, checked at start, every 15 min and shortly after writes. Last use is the file's mtime, bumped when served. An HLS stream (manifests + segments) is evicted as one, with its kv entries. Kept and not counted: downloads (blob URLs in download jobs), and streams in playback or being cached (any of their files used in the last 5 min, or during a running cache operation) until they go idle. Other blobs used in the last 5 min are counted but kept. Posters, `images/` and plugin icons are never evicted. Cache files may be deleted by hand: lookups check the blob (and every segment of an HLS stream) exists and otherwise go back to the origin; at start, kv entries for missing blobs are dropped, downloads whose file is gone are marked failed, and half-written `.put-*` files are removed.

### Mobile app

`apps/mobile` is a native Kotlin + Jetpack Compose (Material 3) Android app plus the Go backend running on the phone; no WebView, no Capacitor. Gradle root is `apps/mobile` (Kotlin DSL, version catalog in `gradle/libs.versions.toml`, one `:app` module, package `com.playingwithclouds.veil`). In `apps/mobile`: `bun run build` (backend for android/arm64 into `jniLibs` + release plugin bundles into `assets/plugins`), `bun run apk` (build + `./gradlew assembleDebug`, APK at `app/build/outputs/apk/debug/app-debug.apk`), `bun run android` (apk + adb install + launch), `bun run test` (JVM unit tests). `local.properties` (gitignored) must point `sdk.dir` at an Android SDK with platform 37.

- The backend ships as `jniLibs/arm64-v8a/libveil.so` (the only place Android lets an app execute files from; pure Go, `CGO_ENABLED=0`). `backend/EmbeddedBackend.kt` starts it from `VeilApplication` on `127.0.0.1:47831` with `DATA_DIR` in app storage, extracts the APK's plugin bundles (`assets/plugins`) as `PLUGIN_SEED_DIR` once per app install/update (`PluginSeeder`), keeps `DNS_SERVERS_FILE` current on network changes, restarts it when it crashes, and logs to logcat tag `VeilBackend`. Environment and DNS-file rules live in `BackendEnvironment` (unit tested).
- The app uses that on-device backend unless another server is saved in Settings (`ServerSettings`, `ServerForm`; saving restarts the app so no screen keeps data of the old server). `VeilApi` is the one place that builds the Apollo client (HTTP + graphql-ws) for the address in use and waits for `/health` at launch (`BootGate` in `ui/VeilApp.kt`, with the server form as fallback when the backend never answers).
- GraphQL: Apollo Kotlin generates typed models from `apps/backend/internal/api/graphql/schema/*.graphql`; the app's operations are `app/src/main/graphql/com/playingwithclouds/veil/*.graphql`. After changing SDL, re-check those operations (`./gradlew generateVeilApolloSources`). `data/*Repository.kt` wrap operations and map them to UI models; `/api/search` server-sent events are read by `api/LiveSearch.kt`.
- Backend-built blob/stream URLs carry `PUBLIC_URL` (default localhost). Pass them through `ServerSettings.backendUrl()` (or `ServerSettings.imageUrl()` / `RemoteImage` for images, which proxies remote images through `/api/img`) so loopback URLs are rebased onto the address the app actually uses (`BackendUrls`, unit tested).
- UI: `ui/VeilShell.kt` (tabs Home, Following, Galleries, Library side by side in one swipeable `HorizontalPager` on the `tabs` route, with the floating tab bar `ui/FloatingNavBar.kt` (solid, no glass, hidden while scrolling down; search lives in each tab's top bar via `SearchButton`); Home (`ui/home`) has a slim always-visible top bar (wordmark, search, menu), a chip row that hides on scroll (`HomeFilterRow`: site picker sheet, All, saved presets, runtime windows, top tags; `HomeFilter` decides ranked feed vs. tag listing, presets are backend `saved_filter` rows incl. sites), edge-to-edge `SceneCard`s with a meta line (studio · performer · site rating · site views · video age), a ⋮ for quick actions, swipe right for "not interested" (`SwipeToDismissCard`, dislike verdict) and a shelf after every 5 cards (scene rows from `recommendedRows`, performers/studios from the feed's credits); Continue watching lives in the Library hub; Galleries is the gallery plugin's category index; Library is the hub for downloads, watchlist, collections, history, performers, studios, tags; the gear menu holds Random, Plugins, Settings; page animations in `ui/PageTransitions.kt`; system back via navigation-compose), one package per screen under `ui/`, `ViewModel` + `StateFlow` + `LoadState`, endless lists through `ui/paging/PagedList`. Playback is Media3 ExoPlayer (`ui/scene/ScenePlayer.kt`): HLS and progressive through the backend's `/api/stream/` proxy, resume and progress saving, fullscreen with landscape lock and system back. Home logs recommendation impressions through `data/ImpressionLogger`.
- Privacy: every option sits in Settings → Privacy (`ui/settings/PrivacySection.kt`), stored by `privacy/PrivacyPreferences` (StateFlows over SharedPreferences). `MainActivity` is a `FragmentActivity` (biometric prompt) and applies the options.
  - Recents/screenshots: `FLAG_SECURE` while "Hide in recents and screenshots" (default on) or the app lock is on. Screenshots, `adb screencap` and recordings come out black, so turn it off to take screenshots.
  - App lock (`AppLock`, `LockPolicy`, `PinHasher`, `ui/privacy/LockScreen`): 4 to 8 digit PIN stored as salted PBKDF2, optional biometric unlock (`BiometricUnlock`), lock delay (immediately, 1, 5 min), wrong PINs throttled from the 5th try. The lock screen is drawn over the whole app in `VeilApp`; a cold start begins locked. Locking and the panic action also pause playback through `PrivacyEvents.pauseRequests`, which `rememberScenePlayer` collects.
  - Panic (`PanicController`, `PanicDetectors`): face-down (accelerometer, only while the app is in front) or a two-finger double tap (read in `MainActivity.dispatchTouchEvent`, never consumed) pauses playback, locks if a lock is set and sends the app to the background (home screen).
  - App name and icon (`Disguise`): one `activity-alias` per entry in the manifest (`.alias.DefaultAlias` is the real one, `Files`/`Gallery`/`Notes`/`Calculator` share the placeholder icon `@mipmap/ic_disguise`); `Disguises.apply` enables one and disables the rest. Labels live in `strings.xml`. The app itself keeps its name until the owner picks one.
  - Neutral media and notifications (`NeutralMedia`): every `MediaItem` carries the metadata "Playing" + app name and no artwork, so a media session shows nothing from the library; the app has no session or notification producer yet. `NeutralNotifications` holds the allowed wording and builds `VISIBILITY_SECRET` notifications for when one is added.
  - Incognito: in-memory only (ends with the process; the "You're incognito" label under the status bar ends it too). While on, `saveProgress`, `recordSearch` and the `ImpressionLogger` send nothing. Likes, ratings, lists and downloads still count, and a scene's first visit still stamps `detail_fetched_at`, which the engine reads as a click.
  - Encrypted downloads: the backend encrypts `stream-cache/` blobs (streams and downloads) at rest with chunked AES-256-GCM (`internal/storage/encrypt.go`, seekable, so range requests work) when `BLOB_ENCRYPTION_KEY` (base64, 32 bytes) is set; `BLOB_ENCRYPTION_WRITES=off` stops encrypting new blobs but still reads encrypted ones, and plain blobs stay readable. The app (`privacy/BlobKeyProvider`) generates a random data key, stores it sealed by an Android Keystore AES key, passes it in `BackendEnvironment`, and replaces it if it can no longer be opened (old encrypted blobs are then unreadable). The switch "Encrypt downloads" (default on) applies after a restart.
- Design system: tokens in `ui/theme` (`VeilColors`, `VeilSpacing` incl. the 16 dp `gutter`, `VeilShapes`, typography in `Theme.kt`, `VeilType` for the dense card styles), components in `ui/design` (buttons, pills incl. the flat `FeedChip` of feed filter rows, headers, surfaces, glass, `bleed`/`gutterPadding` layout helpers). Screens use only those: no raw colours, shapes, `sp`, or literal dp in paddings/gaps, enforced by `DesignTokenUsageTest`. Accent marks selection and progress; main actions stay white; things floating over content are blurring `glass` (not the tab bar, which is solid), every other button, pill and switch is `glassControl` (same look, no blur). Settings → Design kit (`ui/designkit`) shows every token and component on a device.

## Code style

Readability and maintainability over cleverness. Match the surrounding code when it conflicts with the rules below.

- Give every function a doc comment saying what it does, so it reads clearly at the call site (`// Name …` in Go, `/** … */` in TS). The signature carries the types — don't restate them in `@param`/`@returns` tags.
- Comments are technical, concise, and for the reader who arrives later. Add them where skimming the code isn't enough — not to restate it. Don't document a feature you just removed.
  - Good: `// Normalize external IDs before database lookup.`
  - Bad: `// Now we loop through the items and do the thing.`
- Descriptive names, never shortened. `camelCase` for variables, functions, parameters and object fields unless the language or the existing code says otherwise.
  - Good: `recordId`, `customerAccount`, `paymentMethod`
  - Bad: `rid`, `acct`, `pm`
- Explicit control flow over shorthand. Avoid `??`, ternaries and compact conditionals unless they clearly save a lot of repetition.
  - Good: `if (timeout === undefined) { timeout = DEFAULT_TIMEOUT }`
  - Bad: `const timeout = options.timeout ?? DEFAULT_TIMEOUT`
- Svelte: use its directives for dynamic classes (`class:rotate-0={open}`) over string-building (`class={open ? "rotate-0" : "rotate-90"}`).
- Don't declare a variable that's used once; use the expression inline.
- More than 2 levels of nesting is too much — use guard clauses, early returns, or extracted helpers.
- Short, focused functions, one responsibility each. Prefer a named helper over long inline logic with a comment above it.
- Don't change behaviour, public APIs, data shapes, validation or side effects unless asked.

## Tests

When your changes are written, consider whether they need a test, and add one next to the code it covers:

- Go: `*_test.go` beside the package, run `go test ./...` in `apps/backend`.
- Plugins: `tests/*.test.ts` in the plugin, run `bun test` there.
- Browser flows: Playwright, run `bunx playwright test`.

Failures mean investigate, not move on.
