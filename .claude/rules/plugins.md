---
paths:
  - "apps/backend/plugins/**"
  - "apps/backend/internal/plugins/**"
  - "apps/backend/internal/pluginstore/**"
  - "apps/backend/internal/face/**"
  - "apps/backend/internal/mediakeep/**"
  - "apps/backend/scripts/*plugins*"
  - "packages/veil-sdk/**"
---

# Plugins

TypeScript packages in `apps/backend/plugins/`, importing `@playingwithclouds/veil-sdk` (`runPlugin` + shared types). The backend never runs source: `scripts/bundle-plugins.ts` bundles each to CommonJS at `DATA_DIR/plugins/<name>/plugin.js`, run in goja (`internal/plugins/embedded*.go`), fresh runtime per call, hot-reloaded via fsnotify. `bun dev` bundles in watch mode (reload within a second). The Bun runner remains only for the parity test.

## Rules

- Runtime: web APIs (`fetch`, `URL`, `URLSearchParams`, `AbortSignal.timeout`, `JSON`, promises), the SDK, pure-JS npm packages. No Node/Bun built-ins, files, processes, native addons. Also available for uploads: `setTimeout`, `Blob`/`FormData` (bytes and string parts), `Uint8Array` fetch bodies; SDK `base64ToBytes`.
- `node-html-parser` stays external; Go answers `parse()` natively (x/net/html + cascadia, ~60× faster). Only `querySelector(All)`, `getAttribute`/`hasAttribute`, `closest`, `text`/`rawText`, `innerHTML`/`outerHTML`/`toString`, `tagName`, `id`, `classList`, `parentNode` exist (`embedded_html.go`). Add anything else there before relying on it.
- After changing parsing, check parity (in `apps/backend`; hits live sites; xhamster tag order varies): `bun scripts/bundle-plugins.ts --out /tmp/bundles && VEIL_EMBEDDED_BUNDLES=/tmp/bundles go test ./internal/plugins -run EmbeddedBundlesLive -v`
- A plugin serving several sites lists them in `meta.sites` (name, host, icon, content, `off_by_default`); the search filter offers each as `plugin/host` and passes the picked hosts as `ListArgs.sites`, which the plugin must honour (even for sites off by default). The library matches `plugin/host` by the source URL's host.
- Set `meta.content` when sites carry more than plain videos (missav `jav`, showcamrips `cam-shows`, kvs from its sites). A multi-kind plugin must honour `ListArgs.content`.
- `scene:find` fills `Scene.related` with the page's related videos as `DiscoveredItem`s, same `external_id` scheme as listings (reuse the card parser).
- Cloudflare-challenged sites: `meta.requires_solver: true`, fetch via `fetchHtmlSmart`. Without `FLARESOLVERR_URL` (e.g. phone) they're unavailable: out of search/browse, stubs hidden, visits skipped.
- Sites signing poster/preview URLs with an expiry: `meta.expiring_media: true`. `internal/mediakeep` then copies listed stubs' posters to `posters/` (unless already a blob or from another site) and swaps in fresh `preview_video` on relisting; expired stubs repair on next listing. Clips aren't copied.
- A site mirrored across domains is one plugin claiming all domains (e.g. `noodlemagazine`, `src/mirrors.ts`: tries noodlemagazine.com, ukdevilz.com, mat6tube.com in order, skips Cloudflare ones without FlareSolverr, names videos by the noodlemagazine.com page). Retired packages go in `pluginstore.RetiredPackages` (update installs the successor, removes them).
- Stripchat isn't supported: open room list, but HLS is Mouflon-gated and serves an ad loop without the player's key.

## Distribution

GitHub only, never npm. Rolling `plugins` release: per version `<folder>-<version>.tgz` (`package/package.json` + `package/plugin.js`, keyword `veil-plugin`) plus `index.json` (latest version, tarball URL, sha512). Package `@playingwithclouds/veil-plugin-<name>`; folder = name minus scope and `veil-plugin-`.

- `internal/pluginstore` reads the index (`DefaultIndex`, `PLUGIN_INDEX`), installs with integrity check, installs the default set on fresh backends, updates at start and every 6 h, seeds from `PLUGIN_SEED_DIR` without undoing updates/uninstalls (`.store.json`). Dev bundles (`localBuild: true`) are never replaced.
- Ship a fix: bump `version` in the plugin's `package.json` (meta reads it), `bun scripts/publish-plugins.ts [plugin …]` in `apps/backend` (`gh` as PlayingWithClouds; `--dry-run`; skips published versions; uploads `index.json` last). Backends update within 6 h or on "Check for updates"; the phone on every launch.

## Capabilities

`meta`, `discover`, `scrape`, `search`, `enrich`, `resolve` (embed/page URL → playable mp4/HLS + headers; routed by `meta.domains`, `"*"` = wildcard), `browse:categories`, `browse:actors`, `scene:list:page` (`scene:list` with `url` lists a channel/performer/category page), `cam:list` (live rooms: query/tag/gender/limit/offset → `CamRoom`s + total), `cam:stream` (username → live HLS master as `ResolveResult`), `face:identify`.

Resolution chain (`internal/resolve`): cache → domain plugin → wildcard → direct.

## Protocol

stdin/stdout JSON: `{ "capability", "args" }` → `{ "result" } | { "error" }`.

- `scrape` streams NDJSON: one `ScrapeResult` per line, root entity (scene/gallery/image/collection) first, then related performers/studios/tags.
- Lists may write `{ "items": [...] }` lines before the result so live search shows them early (SDK `ItemSink` for page crawls, `emitItems` for fan-outs). The result still carries every item; `Runner.ListStream` forwards each once.
- `Scene.studio` is `StudioRef { name, external_id?, source_url? }`.
- `Collection` with ordered `member_urls`; unresolved members are queued for scraping.
- `Scene.markers` (tag/label + seconds/end_seconds) ingest as global markers.
- Optional `rating` (0..10) and `view_count` on `Scene` and listing `DiscoveredItem`s; cards show them.
- `scene:find` may fill `Scene.comments` (plain text, author, ISO date via SDK `relativeDate`, likes, `reply_to` index). The backend doesn't store them yet (scope issue "Site comments stay, per source").

## Accounts

Sites behind a login (`internal/accounts`, SDK `account.ts`). One account per plugin and device; session and form values sealed with `DATA_DIR/account.key`, never through the API.

- `meta.login: { fields, cookies }`: `fields` is the form `account:login` turns into a session (SDK `jarFetch` on `newJar()`, return `loggedIn(jar, label)`); `cookies: true` lets the user import browser cookies (cookies.txt, extension JSON, `Cookie:` line), kept only for `meta.domains`. Both give a `CookieJar` session.
- Later calls fetch with `sessionFetch`: it sends the jar's cookies and, when the site set new ones, the SDK writes a `{"session"}` line the runner stores (rotating cookies stay logged in).
- Throw `NeedsLoginError` when the site wants a login: the runner has the account log in again with the stored form values and retries once; without them (imported cookies) the account turns `expired` and the call fails with `plugins.ErrNeedsLogin`.

## Identification plugins

No backend feature consumes these yet; they return as lookups that answer with matches and evidence (scope issue "Lookups return matches with evidence").

- `face:identify` (args `image_base64`, `mime_type`; result `candidates[{ name, similarity 0..1, image_url?, profile_urls }]`, `videos`, `posts`): facedigger (job API: create, poll, fetch; no key or captcha; `cams` titles → candidates, `porn` → `videos`) and pornstarbyface.
- namethatporn: reverse image search (multipart `imgfile` POST to `/search/images.html`, reads the accepted answers of the first 10 solved posts), `post:find` (one thread: question, comments, `names`, `links` unwrapped from `redir.me`), `post:search`. Cloudflare 403s Bun's fetch but not Go's, so check it live via `TestEmbeddedBundlesLive`.

## Site quirks

- showcamrips: recorded cam shows; searches only whole model names, cam site or category; its mp4 needs `Referer`.
- kvs: one config list of Kernel Video Sharing sites, including cam-show archives cloudbate, camstreams, streamleak and camgirlsleak (the last two are separate catalogs of one operator, not mirrors); `camPlatforms` credits the cam site as studio, also for batch categories like "Chaturbate-2025-01". `enabledByDefault: false` (scatporno: mostly download-only pages) keeps a site out of search and browse unless `KVS_SITES` names it. stream-leak.com 403s Bun's fetch but not Go's, so check it with `EmbeddedBundlesLive` (`VEIL_EMBEDDED_ENV="KVS_SITES=streamleak"` limits the plugin's sites, `VEIL_EMBEDDED_QUERY` picks the search).
- chaturbate: `/api/ts/roomlist/room-list/` and `get_edge_hls_url_ajax`, both need `X-Requested-With: XMLHttpRequest`.
- xvideos: also serves xnxx.com pages (find/resolve only; listings stay on xvideos). Ask `Accept-Language: en`, else titles come machine-translated. Streams are signed for a few hours (`html5player.setVideoHLS`), so downloads carry the page URL and resolve refetches it. Pages are 0-based (`/`, `/new/1`, `/tags/x/1`, `?k=x&p=1`); pornstar/channel/profile listings and comments are XHR JSON (`/<kind>/<name>/videos/new/<n>`, `/threads/video-comments/get-posts/top/<id_video>/0/0`); channel JSON gives click-tracking paths, rebuilt from id and slug.
