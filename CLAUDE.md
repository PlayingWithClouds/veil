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
| `internal/stream/` | HLS/MP4 proxy + write-through blob cache; live sessions (cams) bypass the cache |
| `internal/resolve/` | Stream resolution chain (cache → domain plugin → wildcard → direct) |
| `internal/alike/` | Cross-plugin alternate-source matching (pHash + performer/duration score) |
| `internal/cams/` | Live cam rooms: listing through the cam plugins (60 s page cache), `cam_model` upserts, follow, online probe, live stream URL, recorder (`recorder.go`, `recording_scene.go`) and auto-record watcher (`autorecord.go`) |
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
- Same thing, several records: performers, studios and tags match by spelling-insensitive `name_key` (`internal/names`), and `RepairNamedRecords` merges the ones that predate it at startup. Scenes (`alike.SameSceneStrict`), galleries (`alike.SameGallery`) and scraper collections (same name + a shared member) are merged by `MergeDuplicateMedia` at startup and, for a newly created scene or gallery, at ingest. A merged-away record's `source_url`/`external_id` stay in `entity_identity` and resolve to the survivor. Merges are irreversible, so the matchers stay strict; the feed's `dropDuplicates` uses the looser `SameScene` and only hides.
- Stage names: `Orchestrator.ProfilePerformers` (start + every 6 h, 100 per run) searches the `freeones` plugin for performers without a profile and ingests the page that names them; its aliases are absorbed and any other performer whose name is one of them (full names only, two words or more) is merged in (`mergeAliasTwins`, also run by the startup repair). `mergeRecords(survivorId, mergedId)` merges two records by hand (web: ⋮ → "Merge into…" on performer/studio pages, a button on tag pages).
- Migrations are append-only (`db.Open` applies them and tracks versions in `_migration`); never edit an applied migration, add a new numbered file.
- Lists of links (`tags`, `performers`, `aliases`) are JSON columns; match them with `json_each`. Bind lists as `$param` and use `IN (SELECT value FROM json_each($param))`.

## Recommendations

`internal/recommend` runs candidate generation → ranking → re-ranking per request, no ML or external calls.

- **Signals** (per scene, time-decayed with a 30-day half-life): o-event > like > completed watch > save (watchlist, user collection) ≈ A/B pick > partial watch > click (first visit or logged click); negative: dislike, early abandon (<10 s or <5% of `watch_history.max_progress_seconds`), A/B reject, shown-but-not-clicked (`recommendation_impression`). Blocked entities pin to −1.
- **Profile**: computed on demand (no cache table to invalidate); signalled scenes spread their weight onto performers, tags (÷√count), studio and site (first observing plugin), each dimension normalized to [−1, 1].
- **Candidate sources**, each item tagged with one: `scene_related` of the strongest signalled scenes (site-related highest prior), new search-subscription finds, newest scenes of top performers/studios/tags, the feeder's finds (`discover:taste`, `discover:catalog`), exploration (random scenes of tags adjacent to the top tags, plus random), unvisited stubs from recent searches, newest scenes. Cold start = catalog finds, searches, newest.
- **Feeder** (`updater.Feeder`): the engine only ranks what is stored, so the feeder keeps it supplied without a search. Every 12 h (and when a ranking has fewer than 100 unwatched, unshown candidates, at most hourly) it lists one catalog page of every scene plugin (one page deeper per run, three deep) and the site pages, else name searches, of the top 3 performers, 2 studios and 3 tags, ingests them as stubs and records them in `feed_discovery` (kept 30 days, counted 14). Taste finds are scaled by the entity's current affinity. A run that stored anything clears the memoized feed.
- **Ranking**: affinity match + source prior + freshness − watched/fatigue/disliked-neighbour penalties. Blocked scenes, disliked scenes and inactive-plugin stubs are excluded. **Re-ranking**: per-24 page caps per performer/studio/site (plugin), no source or site three in a row, every 7th slot exploration.
- **Cross-site duplicates**: after scoring, `dropDuplicates` keeps the best-ranked copy of a video several sites carry (`alike.SameScene`: runtimes within 5 %, then identical credited performers or near-identical title; scenes without a runtime never match). `recommendations(excludeTagIds)` drops scenes carrying a tag (own, studio's or performer's) from the shared ranking, `scenes(includeTagIds, excludeTagIds)` does the same for listings, and `tasteProfile` exposes the profile's strongest liked and disliked tags, performers, studios and sites.
- The ranked feed is memoized 30 min for `offset > 0` so pages stay stable; `offset 0` re-ranks once the ranking is 10 min old, or on `refresh: true` (Home's refresh button). The web logs impressions in batches via `recordImpressions` (`src/lib/recommendations.ts`).

## Plugin system

Plugins are TypeScript packages under `apps/backend/plugins/`, all importing `@playingwithclouds/veil-sdk` (`runPlugin` dispatcher + shared types). The backend never runs their source: `scripts/bundle-plugins.ts` bundles each into one CommonJS file in `DATA_DIR/plugins/<name>/plugin.js` (the installed-plugins folder, same on server and phone), and the backend runs it in an embedded JS engine (goja, `internal/plugins/embedded*.go`), a fresh runtime per invocation, hot-reloaded via fsnotify. `bun dev` runs the bundler in watch mode, so saving a plugin file reloads it within a second.

**Distribution via GitHub, never npm:** plugins are assets of this repository's rolling `plugins` release: per version an npm-style tarball `<folder>-<version>.tgz` (`package/package.json` + `package/plugin.js`, keyword `veil-plugin`), plus `index.json` listing every package's latest version with tarball URL and sha512 integrity. Packages are named `@playingwithclouds/veil-plugin-<name>`; the folder name is the package name minus scope and `veil-plugin-`. `internal/pluginstore` reads the index (`DefaultIndex`, override with `PLUGIN_INDEX`), installs packages (integrity checked), installs the default set on a fresh backend, updates to newer versions at start and every 6 h, and seeds from `PLUGIN_SEED_DIR` (the phone app's bundled set) without undoing updates or uninstalls (`.store.json` in the plugin folder). Dev bundles carry `localBuild: true` and are never replaced. To ship a fix: bump `version` in the plugin's `package.json` (meta reads it from there), then `bun scripts/publish-plugins.ts [plugin …]` in `apps/backend` (needs `gh` logged in as PlayingWithClouds; `--dry-run` to preview; skips versions already in the index; uploads `index.json` last). Every backend picks it up within 6 h or on "Check for updates" in Plugins; the phone checks on every launch.

**Runtime rules** (what a plugin may use): web APIs (`fetch`, `URL`, `URLSearchParams`, `AbortSignal.timeout`, `JSON`, promises), the SDK, and pure-JS npm packages (the bundler inlines them). No Node/Bun built-ins, files, processes or native addons. `node-html-parser` is special: bundles keep it external and the backend answers `parse()` natively in Go (x/net/html + cascadia), because parsing in JS under goja is ~60× slower. Only `querySelector(All)`, `getAttribute`/`hasAttribute`, `closest`, `text`/`rawText`, `innerHTML`/`outerHTML`/`toString`, `tagName`, `id`, `classList` and `parentNode` exist there (`embedded_html.go`); add to it before relying on anything else. `bun test` in a plugin still runs the real node-html-parser, so check changed parsing against the embedded runtime: `bun scripts/bundle-plugins.ts --out /tmp/bundles && VEIL_EMBEDDED_BUNDLES=/tmp/bundles go test ./internal/plugins -run EmbeddedBundlesLive -v` in `apps/backend` hits the live sites and compares every plugin's list/find output with Bun's (xhamster's tag order varies per request).

Plugins may send the site's own `rating` (0..10) and `view_count` on `Scene` and on listing `DiscoveredItem`s; cards show them. Only eporner does so far.

**Type mirrors — keep in sync in one commit:** `apps/backend/internal/plugins/types.go` ⇄ `packages/veil-sdk/src/types.ts`.

**Capabilities:** `meta`, `discover`, `scrape`, `search`, `enrich`, `resolve` (embed/page URL → playable mp4/HLS + headers; `meta.domains` routes, `"*"` = wildcard), `browse:categories`, `browse:actors`, `scene:list:page` (`scene:list` with a `url` arg lists a channel/performer/category page; eporner, hqporner), `cam:list` (rooms online now: args query/tag/gender/limit/offset, returns `CamRoom`s + total) and `cam:stream` (room username → live HLS master as a `ResolveResult`), `face:identify` (see Face identification).

**Protocol** (stdin/stdout JSON): `{ "capability", "args" }` in, `{ "result" } | { "error" }` out — except `scrape`, which streams NDJSON: one `ScrapeResult` per line, root entity (scene/gallery/image/collection) first, then related standalone entities (performer/studio/tag). A list may write `{ "items": [...] }` lines before its result so live search shows them while it keeps crawling (SDK `ItemSink` for page crawls, `emitItems` for site fan-outs; spankbang, kvs, aylo do); the result still carries every item and `Runner.ListStream` forwards each item once.

- `Scene.studio` is a structured `StudioRef { name, external_id?, source_url? }`.
- Plugins may emit `Collection` with ordered `member_urls`; unresolved members are auto-queued for scraping.
- Plugins may emit `Scene.markers` (tag/label + seconds/end_seconds) — ingested as global markers.
- `scene:find` should fill `Scene.related` with the page's related-videos list as `DiscoveredItem`s, using the same external_id scheme as the plugin's listings (reuse the listing card parser).

**FlareSolverr:** set `meta.requires_solver: true` when the site answers plain requests with a Cloudflare challenge, and fetch through `fetchHtmlSmart`. Without `FLARESOLVERR_URL` (e.g. embedded on a phone) the registry treats such plugins as unavailable: they drop out of search/browse, their stubs are hidden, visits skip them. Currently: missav, spankbang.

**Current plugins:** aylo (pornhub, redtube, youporn, tube8), eporner (API + enrich), hqporner, kvs (37 Kernel Video Sharing sites from one config list), missav, xhamster, spankbang, tnaflix, txxx-network (scenes); pornpics (galleries); pornstarbyface, facedigger (`face:identify` only); chaturbate (live cams: `/api/ts/roomlist/room-list/` and `get_edge_hls_url_ajax`, both need the `X-Requested-With: XMLHttpRequest` header). Stripchat is not supported: its room list is open but its HLS playlists are gated by Mouflon and serve only an ad loop without the player's key.

**Face identification ("Who is this?"):** `identifyFace(mediaId, imageId?)` (`internal/face`) sends a scene's thumbnail/poster, a gallery image (loaded through the image cache) or a supplied frame (`frameBase64`: on the web the scene page pauses the player and grabs the frame on screen, `VideoPlayer.captureFrame`, which needs the `crossOrigin = 'anonymous'` video from `playback.ts`; "Pick another moment" lets the user scrub and search again) to every active `face:identify` plugin (args `image_base64` + `mime_type`, result `candidates[{ name, similarity 0..1, image_url?, profile_urls }]`), concurrently with a timeout, merges by `name_key`, returns the stored picture's `pictureUrl` for the dialog to show and marks performers already in the library; `creditPerformerByName` credits the pick through `ingest.CreditPerformer` (same name matching as ingest). **Performer-only rule:** names (candidates) may only come from services whose database holds professional performers or cam models (stage personas). Never general face search (PimEyes/FaceCheck style) or databases that include private people (amateur/leaked material must not be unmasked), never bypass a captcha/bot check, never use a site whose terms forbid automation or the user's region. Tube-site videos with a similar face may be offered beside the candidates as `videos` (`FaceVideoMatch`: the video's own title, similarity, page link, thumbnail): they are links only, shown under "Similar videos", and never become a candidate, a performer name or a credit. The `facedigger` plugin (job API: create, poll, fetch; no key or captcha) reads two of FaceDigger's categories: `cams`, whose titles are cam models' names, becomes candidates; `porn` (videos, titles possibly of private people) becomes `videos`. Its `onlyfans` category lists creator accounts and is never read. It refuses to run until the owner turns on the plugin setting `FACEDIGGER_ACCEPT_AGREEMENT`, since FaceDigger's user agreement (not an EU resident, no stalking/harassment, ...) is the owner's to accept. The embedded runtime provides `setTimeout`, `Blob`/`FormData` (bytes and string parts only) and `Uint8Array` fetch bodies for uploads; `base64ToBytes` in the SDK decodes the picture.

## Cams

A cam model is a live room, not a scene: it comes and goes. Cam plugins (`cam:list`, `cam:stream`; see Plugin system) never produce scenes, so none of the scene discovery, ingest or recommendation code sees them.

- `cam_model` (id `cam_model:<id>`, identity `source_url`) stores every room a listing has shown: name, thumbnail, tags, gender, subject, viewers, `online`, `last_seen_online_at`, and the `followed` flag. Rows are upserted whenever a listing page is fetched, so followed rooms stay known while offline.
- `camRooms(sources, tag, query, gender, limit, offset)` lists live to the plugins on request (`internal/cams`): each plugin is paged in chunks of 100, chunks are cached in memory for 60 s so paging stays stable, plugins keep their own popularity order and are merged by viewers. Rooms in a private or password show are skipped. `camSources` feeds the site filter.
- Online status of one room (`camModel`) or the Following row (`followedCamModels`, online first) is a `cam:stream` probe, reused for 45 s. `followCamModel` / `unfollowCamModel` flip the flag.
- `camStream(id)` resolves the room and returns a proxy URL from `Proxy.LiveManifestURL`: a live session, so playlists are fetched fresh, LL-HLS tags (parts, preload hints, blocking reload) are stripped, and segments are relayed without writing the blob store or kv cache. The site's token URL expires within about a minute, so the player asks for a stream right before loading it and again on retry.
- Recording: `Recorder` (`internal/cams/recorder.go`) pulls a room's HLS stream in the backend, so a capture outlives the page. `startCamRecording(id, quality)` picks the rendition from the master playlist (`quality` = `source`, the best, or a target height like `480`: the tallest at or below it, else the shortest above; `chooseVariant`), and re-resolves the stream and re-picks on every reconnect, since token URLs expire within about a minute and renditions change. The video playlist is followed with its separate audio rendition when the master has one (Chaturbate does), and each segment is stored once (deduplicated by URL path, discontinuity marked after a gap). A capture ends on `stopCamRecording`, when the broadcast ends, after `offlineGrace` (90 s) without a reachable stream, or after the `camRecordingMaxMinutes` setting (default 240 min). Footage under 5 s is dropped.
- Where recordings live: `recordings/<id>/` in the blob store (`index.m3u8`, `seg-NNNNN.ts|m4s`, `init-N.mp4`, and with separate audio `video.m3u8`, `audio.m3u8`, `aseg-*`, `ainit-*`, `poster.jpg`). The prefix is outside `stream-cache/`/`img-cache/`, so the cache evictor never touches it, and it is encrypted at rest with `BLOB_ENCRYPTION_KEY` like `stream-cache/`. `index.m3u8` is rewritten every 10 segments, so `Recorder.Recover` (at start) files a capture a crash interrupted from the playlist it had flushed. `cam_recording` rows (status `recording|finished|failed`, stop reason, quality, duration, size, scene) keep the history.
- A finished recording is a normal scene: source URL `veil-recording:<id>` (`ingest.RecordingSourcePrefix`, exempt from duplicate merging because two sessions of one model share performer and runtime), title "<name> — live <date time>", the model's performer (found or created by `name_key` through ingest), the cam site as studio, tags "Live cam" plus the room's own, the room thumbnail as poster, `detail_fetched_at` set, and one stream (format `hls`) pointing at `index.m3u8`. `stream.CacheService.IsRecording` keeps `Resolve`/`ManifestURL` from copying it into the stream cache; it plays straight from the blob store.
- `camRecordings(camModelId, activeOnly)` returns the running recordings (live duration, size, quality) then finished ones; `setCamAutoRecord(id, enabled, quality)` stores `cam_model.auto_record` and `record_quality` (turning it on also follows; unfollowing turns it off). `AutoRecorder` polls followed auto-record models every 2 minutes through the cached online probe, staggered, and starts a recording at the model's quality when she is live and none runs for her; one room never has two captures.
- Web: Cams in the sidebar and phone drawer, `/cams` (live grid with `ScrollSentinel`, search, gender and site filters, tag filter via `?tag=`, Following row, a "Recording" strip and REC/Auto badges on cards) and `/cams/[id]` (the scene player `VideoPlayer` in `live` mode: LIVE badge and "Go live", no seek bar, speed, markers or progress saving; quality menu from the hls levels, last pick in `localStorage` `veil.cam.quality`; unmuted start with a "Tap to unmute" prompt when the browser refuses; a stream error asks for a fresh URL up to 3 times before the offline state), Record button with a quality menu, follow, auto-record toggle and quality, and the finished recordings as `SceneCard`s. `ActiveRecordings` lists running recordings (avatar, name, pulsing dot, time, quality, Stop) in the sidebar and drawer, polled every 10 s by `stores/camRecordings.ts`. The Android app does not show cams yet.

## Frontend

- GraphQL only, via `gqlClient` (`apps/web/src/lib/veil.ts`) with inline genql selections in `src/lib/*.ts` data modules; subscriptions over graphql-ws (`streamsChanged`).
- Internal links go through `src/lib/routes.ts` (`sceneUrl`, `performerUrl`, `tagUrl`, ...).
- Design tokens + component classes live in `src/routes/layout.css`, TV/D-pad focus in `src/lib/tv.css` + `FocusGrid`/`FocusList`. Reuse `SceneCard`, `TagChips`, `MemberCard`, `CollectionDialog`, `Lightbox` before writing new components. For anything new, follow the tokens and patterns already in `layout.css`.
- Split view (web only): `/split` (`routes/split`, `splitUrl()`) plays up to 4 cams and/or scenes at once, 1 pane full, 2 side by side, 3-4 in a 2x2 grid, with a drag-to-reorder tab bar and a fullscreen button for the grid. Panes live in `lib/stores/split.ts` (localStorage `veil:split`: `{kind, mediaId, title}` per pane + focused index; playback restarts on reload). `SplitAddButton` (hover button on `SceneCard`, `ExploreSceneCard`, `CamCard`, sidebar followed cams, `ActiveRecordings`) calls `openInSplit`, which adds a pane and navigates to `/split`; with 4 open it replaces the focused pane and says so. Each `SplitPane` resolves its own stream (cam: `fetchCamStream`, re-resolved on error; scene: best stream) into a private-element `VideoPlayer` (not `shared`, so the mini player is untouched). Audio focus: only the focused pane is unmuted (`VideoPlayer` `silenced`, which also isolates it from the session's live-audio choice), and only it reacts to keyboard shortcuts (`keyboard`); clicking a pane focuses it. Panes keep their DOM position and are placed with CSS `order`, since moving a player's element stalls its stream.
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
- Player extras (all in `ui/scene`): the settings gear (`PlayerSettings`) holds speed (0.25× to 2×), quality, the A–B loop, zoom to fill and the picture-in-picture privacy; `PlayerOptions` keeps them (loop per page, zoom and PiP privacy in the `player` preferences, pure helpers in `PlayerMath`). Frame step shows beside play while paused (`Player.stepFrame`, exact seek, frame rate from the track). `MarkerChips` (under the player, above the seek bar in fullscreen) and notches on the bar come from `scene_marker` rows, lazy so hundreds of markers are fine. A vertical video stays in portrait when going fullscreen (`rememberFullscreenState`). `PictureInPicture.kt` enters PiP on leaving the app (auto-enter from Android 12) and by button; in PiP only the video shows, blurred by default (`PictureInPicturePrivacy`), and closing the window pauses. Headphone unplug pauses via `setHandleAudioBecomingNoisy`. `data/PlaybackQueue` is the in-memory queue (play next, add, whole collections via Play all / Shuffle play, auto-advance at the end of a video in `SceneScreen`, `QueueSheet`).
- Replay data: `PlaybackHeatReporting` sends watched spans and scrub targets (`recordSceneHeat`, table `scene_heat`, 100 buckets); `sceneHeat` returns the replay graph drawn above the seek bar, `bestMomentSeconds` (most-scrubbed bucket, else a marker) and `thumbnailSeconds`. When the best moment is more than 30 s from the thumbnail's, the scene page grabs a frame (`ThumbnailCapture`, `MediaMetadataRetriever`, progressive sources) and uploads it (`setSceneThumbnail`); `scene.thumbnail_path` then wins over `poster_path` in `sceneFromMap` and media cards. The backend has no ffmpeg, so frames only come from the app.
- Discovery screens: `ui/previews` is the Shorts-style feed (`AppNavigator.openPreviewFeed()`, route `previews`): a `VerticalPager` over `recommendations` filtered to scenes with a preview clip, only the visible page plays (`ScenePreview`, which draws nothing when the clip fails to load), tap or Watch opens the scene, the tag button hides tags (`excludeTagIds`). `ui/taste` (`openTaste()`) renders `tasteProfile`. Performer and studio pages use `ui/entity/ChannelHeader` (banner, follow bell, "N new" from the subscription, bio sheet) with a Newest/Top rated switch (`EntityFilter.sort`). `data/TagFilter` + `ui/components/TagFilterSheet` hold include/exclude tags (Search applies both, live results are checked client-side). `SceneQuickActionsSheet` offers "Don't recommend" for the studio, performers and tags through the backend blocklist, which also hides blocked performers, studios and tags from the index lists.

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
