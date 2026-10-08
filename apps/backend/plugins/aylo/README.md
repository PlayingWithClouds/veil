# aylo plugin

One plugin for the four Aylo tube sites: pornhub.com, redtube.com, youporn.com
and tube8.com. Per-site differences (URL shapes, selectors) live in
`src/sites.ts`; everything else is shared. The site is derived from the URL, so
`find`, `resolve` and page listings work on all four.

## Capabilities

- **scene:list** — a query searches every enabled site in parallel and
  interleaves the results; without one the newest feeds are merged. The
  `AYLO_SITES` setting (comma-separated keys) narrows the sites.
- **scene:list:page** — `url` lists a channel, pornstar, category or tag page
  (any page of the four sites, paginated with `page=`).
- **scene:find** — title, duration, date, poster, performers, channel/uploader
  as `StudioRef`, categories + tags, and the related-videos grid.
- **stream:resolve** — tallest quality; HLS (RedTube, YouPorn, Tube8) or MP4
  (Pornhub, see below). Streams are IP-bound, so fetch from the resolving host.
- **tag:list** — each site's category index, deduplicated by name.

## Notes

- `external_id` is `<site>-<id>`: Pornhub viewkey, RedTube numeric id, YouPorn
  `/watch/<id>/`, Tube8 `/porn-video/<id>/`.
- RedTube, YouPorn and Tube8 list quality sets behind a "remote" URL that
  returns JSON. Pornhub's inline HLS URLs are randomly served from a CDN that
  answers 410 to non-browsers, so Pornhub resolves through `get_media` (MP4),
  which needs the page's `ss` session cookie.
- Pornhub rejects Go >= 1.26's default TLS handshake (post-quantum key share)
  with a 403 maintenance page; the embedded runtime needs a client config
  without it (`GODEBUG=tlsmlkem=0` or `CurvePreferences` without X25519MLKEM768).
