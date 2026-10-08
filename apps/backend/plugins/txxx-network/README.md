# txxx-network plugin

Scraper for the **TxxxNetwork** family of adult sites, which all share one backend and JSON API:
`txxx.com`, `upornia.com`, `hclips.com`, `hdzog.com`, `vjav.com`, `voyeurhit.com`, `thegay.com`, `shemalez.com`, `fetishshrine.com`, `hotmovs.com`, `vxxx.com`, `inporn.com`, `abxxx.com`, `01tube.com`, `fuxxx.com`, `fufap.com`, `porntop.com`, `fullvideosporn.com` (formerly `sextu.com`, whose links are mapped to it).

The API is identical everywhere; only the SPA's video page route differs: `/videos/<id>/<slug>/` by default, `/video-<id>/` on vxxx, `/video/<id>/<slug>/` on inporn, abxxx and porntop, `/en/video/<id>/<slug>/` on fullvideosporn (`VIDEO_PAGE_SHAPES` in `src/http.ts`).

Capabilities: `scene:find` (fetch one full scene by URL), `scene:list` (latest-updates catalog, or keyword search with a query), `stream:resolve`.

## Setting

- `TXXX_HOST` (default `txxx.com`) — which network site `scene:list` browses/searches. `scene:find`/`stream:resolve` derive the host from the video URL, so they work across every network domain regardless of this setting.

## API

The site is a Vue SPA; all content comes from a JSON API (no HTML scraping). Requests must carry `X-Requested-With: XMLHttpRequest` and a same-origin `Referer`.

- **Browse**: `GET /api/json/videos2/14400/str/latest-updates/<count>/..<page>.all...jsond`
- **Search**: `GET /api/videos2.php?params=14400/str/relevance/<count>/search..<page>.all..&s=<query>`
- **Video detail**: `GET /api/json/video/8640000/<1e6-bucket>/<1e3-bucket>/<id>.json` where the buckets are `floor(id/1e6)*1e6` and `floor(id/1e3)*1e3`.
- **Related videos**: `GET /api/json/videos_related2/<lifetime>/<count>/<1e6-bucket>/<1e3-bucket>/<id>.all.1.json` → videos2-shaped `{ videos: [...] }`; `scene:find` maps it to `scene.related` (one extra request).
- **Stream**: `GET /api/videofile.php?video_id=<id>&lifetime=8640000` → array of `{ format, video_url, is_default }`. `format` is `_hq.mp4` / `_sd.mp4` / `_lq.mp4` / `_tr.mp4` (trailer), or an untagged `.mp4` on sites that keep one rendition (labelled "Source"). `video_url` is obfuscated.

## Stream deobfuscation (`base164`)

`video_url` is encoded with the site's own base64 variant. Its 65-char alphabet swaps five Latin capitals for Cyrillic homoglyphs (`А В С Е М` → `A B C E M`) and uses `~` as the padding character. Decoding yields a `/get_file/…mp4/?d=…&br=…&ti=…` path on the site host, which 302-redirects to the CDN mp4 (no Referer needed on the CDN). Implemented in `src/http.ts:base164Decode`, mirroring the site's `base164_decode`.

## Video model

One video → one Scene. `models` → performers; `categories` + `tags` → tags; `thumb` → poster; `pv` → preview clip; `post_date` → date; `duration` (`MM:SS`/`HH:MM:SS`) → seconds.

## Streams (all qualities)

`scene:find` calls `videofile.php` and emits **every** available format (usually `_hq`, sometimes also `_lq`) as a download — each a decoded `/get_file/` handle, presented for the user to choose. Because those handles end in `.mp4/` (trailing slash) they are not direct-media URLs, so playing or downloading one routes back through `stream:resolve`, which re-derives a **freshly signed** source for the same quality (matched by the format tag in the handle). This keeps stored links working after their signature expires. `stream:resolve` also accepts a plain video-page URL and returns the best quality.
